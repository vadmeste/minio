// Copyright (c) 2015-2021 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package main

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/minio/cli"
	"github.com/tinylib/msgp/msgp"
)

var (
	nodeName        string
	waitFactor      int
	rotateGZIPAfter int
	re              = regexp.MustCompile(`\d->(\d*)`)
)

type scanResult struct {
	path       string
	etag       string
	mtime      int64
	size       int64
	zeroParity bool
	mismatchEC bool
	ecM        int
	ecN        int
}

type writtenBytes struct {
	w io.Writer
	n int
}

func (wb *writtenBytes) Write(p []byte) (int, error) {
	n, e := wb.w.Write(p)
	wb.n += n
	return n, e
}

type logger struct {
	results chan scanResult
	wg      sync.WaitGroup
}

func (l *logger) start() {
	l.results = make(chan scanResult)
	l.wg.Add(1)
	go func() {
		l.work()
		l.wg.Done()
	}()
}

func (l *logger) stop() {
	close(l.results)
	l.wg.Wait()
}

func (l logger) work() {
	var (
		totalScanned    uint64
		mismatchEC      uint64
		zeroParityFound uint64
		rotation        int
		step            = uint64(1)
	)

rotateArchive:
	rotation++
	f, err := os.Create(fmt.Sprintf("corrupted-scan-%s-%06d.txt.gz", nodeName, rotation))
	if err != nil {
		log.Fatal("Unable to create the output file:", err)
	}
	wb := &writtenBytes{w: f}
	zw := gzip.NewWriter(wb)

	for e := range l.results {
		if totalScanned%step == 0 {
			fmt.Printf("scanned=%d, mismatch-ec=%d, zero-parity=%d\n", totalScanned, mismatchEC, zeroParityFound)
			if step < 100000 {
				step *= 10
				if step > 100000 {
					step = 100000
				}
			}
		}
		totalScanned++
		if e.zeroParity {
			zeroParityFound++
		}
		if e.mismatchEC {
			mismatchEC++
		}

		if !e.zeroParity && !e.mismatchEC {
			continue
		}

		_, err = zw.Write([]byte(fmt.Sprintf("%s etag=%s mtime=%s size=%d zeroParity=%t mismatchEC=%t, ecM=%d, ecN=%d\n",
			e.path,
			e.etag,
			time.Unix(0, e.mtime).Format(time.RFC3339),
			e.size,
			e.zeroParity,
			e.mismatchEC,
			e.ecM,
			e.ecN,
		)))
		if err != nil {
			log.Println("ERR:", err)
		}

		if wb.n > rotateGZIPAfter {
			zw.Close()
			f.Close()
			goto rotateArchive
		}
	}

	zw.Close()
	f.Close()

	fmt.Printf("scanned=%d, mismatch-ec=%d, zero-parity=%d\n", totalScanned, mismatchEC, zeroParityFound)
}

func checkErasureUpgradeParityMismatch(r io.Reader) (scanResult, error) {
	b, err := io.ReadAll(r)
	if err != nil {
		return scanResult{}, err
	}
	b, _, minor, err := checkXL2V1(b)
	if err != nil {
		return scanResult{}, err
	}

	if minor != 3 {
		return scanResult{}, err
	}
	v, b, err := msgp.ReadBytesZC(b)
	if err != nil {
		return scanResult{}, err
	}
	if _, nbuf, err := msgp.ReadUint32Bytes(b); err == nil {
		// Read metadata CRC (added in v2, ignore if not found)
		b = nbuf
	}

	_, v, err = decodeXLHeaders(v)
	if err != nil {
		return scanResult{}, err
	}

	var result scanResult

	err = decodeVersions(v, 1, func(idx int, hdr, meta []byte) error {
		// var header xlMetaV2VersionHeaderV2
		// if _, err := header.UnmarshalMsg(hdr); err != nil {
		//	return err
		// }

		var buf bytes.Buffer
		if _, err := msgp.UnmarshalAsJSON(&buf, meta); err != nil {
			return err
		}
		type erasureInfo struct {
			V2Obj *struct {
				EcDist  []int
				EcIndex int
				MTime   int64
				EcM     int
				EcN     int
				Size    int64
				MetaSys map[string]string
				MetaUsr map[string]string
			}
		}
		var ei erasureInfo
		if err := json.Unmarshal(buf.Bytes(), &ei); err == nil && ei.V2Obj != nil {
			result.ecM = ei.V2Obj.EcM
			result.ecN = ei.V2Obj.EcN
			result.size = ei.V2Obj.Size
			result.mtime = ei.V2Obj.MTime
			if ei.V2Obj.EcN == 0 {
				result.zeroParity = true
			}
			if ei.V2Obj.MetaUsr != nil {
				result.etag = ei.V2Obj.MetaUsr["etag"]
			}
			if ei.V2Obj.MetaSys != nil {
				if upgrade := ei.V2Obj.MetaSys["x-minio-internal-erasure-upgraded"]; upgrade != "" {
					decoded, err := base64.StdEncoding.DecodeString(upgrade)
					if err != nil {
						return nil
					}
					matches := re.FindAllSubmatch(decoded, -1)
					if len(matches) > 0 && len(matches[0]) > 1 && len(matches[0][1]) > 0 {
						if foundEC, err := strconv.Atoi(string(matches[0][1])); err == nil {
							if ei.V2Obj.EcN != foundEC {
								result.mismatchEC = true
							}
						}
					}
				}
			}
		}
		return nil
	})

	return result, err
}

func mainAction(c *cli.Context) error {

	waitFactor = c.Int("wait-factor")
	nodeName = c.String("node-name")
	rotateGZIPAfter = c.Int("rotate-gzip-after")

	if len(c.Args()) == 0 {
		log.Fatalln("specify at least one disk mount")
	}

	logger := logger{}
	logger.start()

	var wg sync.WaitGroup
	for _, arg := range c.Args() {
		wg.Add(1)
		go func(arg string) {
			defer wg.Done()

			err := filepath.Walk(arg, func(path string, info fs.FileInfo, err error) error {
				if err != nil {
					return nil
				}
				if info.IsDir() || info.Name() != "xl.meta" {
					return nil
				}

				if strings.Contains(path, ".minio.sys") {
					return nil
				}

				start := time.Now()
				// Parse xl.meta content and check if all versions
				// are older than the specified --older-than argument
				f, e := os.Open(path)
				if e != nil {
					return nil
				}
				result, e := checkErasureUpgradeParityMismatch(f)
				f.Close()
				if e != nil {
					return nil
				}
				end := time.Now()

				result.path = path
				logger.results <- result
				if waitFactor > 0 {
					// Slow down
					time.Sleep(time.Duration(waitFactor) * end.Sub(start))
				}
				return nil
			})
			if err != nil {
				log.Println("ERROR when walking the path %q: %v\n", arg, err)
			}
		}(arg)
	}

	wg.Wait()
	logger.stop()

	return nil
}

func main() {
	app := cli.NewApp()
	app.CustomAppHelpTemplate = `NAME:
  {{.Name}} - {{.Usage}}

USAGE:
  {{.Name}} {{if .VisibleFlags}}[FLAGS]{{end}} [DIRS]...

Pass multiple mount points to scan for xl.meta that have all versions
older than to --older-than e.g. 120h (5 days)

FLAGS:
  {{range .VisibleFlags}}{{.}}
  {{end}}
`

	app.Flags = []cli.Flag{
		cli.IntFlag{
			Usage: "A wait factor between xl.meta scans",
			Name:  "wait-factor",
		},
		cli.IntFlag{
			Usage: "rotate gzip archive after accumulating the specified size",
			Name:  "rotate-gzip-after",
			Value: 100 * 1024 * 1024,
		},
		cli.StringFlag{
			Usage: "The current node name (optional)",
			Name:  "node-name",
		},
	}

	app.Action = mainAction

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

var (
	// XL header specifies the format
	xlHeader = [4]byte{'X', 'L', '2', ' '}

	// Current version being written.
	xlVersionCurrent [4]byte
)

const (
	// Breaking changes.
	// Newer versions cannot be read by older software.
	// This will prevent downgrades to incompatible versions.
	xlVersionMajor = 1

	// Non breaking changes.
	// Bumping this is informational, but should be done
	// if any change is made to the data stored, bumping this
	// will allow to detect the exact version later.
	xlVersionMinor = 1
)

func init() {
	binary.LittleEndian.PutUint16(xlVersionCurrent[0:2], xlVersionMajor)
	binary.LittleEndian.PutUint16(xlVersionCurrent[2:4], xlVersionMinor)
}

// checkXL2V1 will check if the metadata has correct header and is a known major version.
// The remaining payload and versions are returned.
func checkXL2V1(buf []byte) (payload []byte, major, minor uint16, err error) {
	if len(buf) <= 8 {
		return payload, 0, 0, fmt.Errorf("xlMeta: no data")
	}

	if !bytes.Equal(buf[:4], xlHeader[:]) {
		return payload, 0, 0, fmt.Errorf("xlMeta: unknown XLv2 header, expected %v, got %v", xlHeader[:4], buf[:4])
	}

	if bytes.Equal(buf[4:8], []byte("1   ")) {
		// Set as 1,0.
		major, minor = 1, 0
	} else {
		major, minor = binary.LittleEndian.Uint16(buf[4:6]), binary.LittleEndian.Uint16(buf[6:8])
	}
	if major > xlVersionMajor {
		return buf[8:], major, minor, fmt.Errorf("xlMeta: unknown major version %d found", major)
	}

	return buf[8:], major, minor, nil
}

const xlMetaInlineDataVer = 1

type xlMetaInlineData []byte

// afterVersion returns the payload after the version, if any.
func (x xlMetaInlineData) afterVersion() []byte {
	if len(x) == 0 {
		return x
	}
	return x[1:]
}

// versionOK returns whether the version is ok.
func (x xlMetaInlineData) versionOK() bool {
	if len(x) == 0 {
		return true
	}
	return x[0] > 0 && x[0] <= xlMetaInlineDataVer
}

const (
	xlHeaderVersion = 2
	xlMetaVersion   = 2
)

func decodeXLHeaders(buf []byte) (versions int, b []byte, err error) {
	hdrVer, buf, err := msgp.ReadUintBytes(buf)
	if err != nil {
		return 0, buf, err
	}
	metaVer, buf, err := msgp.ReadUintBytes(buf)
	if err != nil {
		return 0, buf, err
	}
	if hdrVer > xlHeaderVersion {
		return 0, buf, fmt.Errorf("decodeXLHeaders: Unknown xl header version %d", metaVer)
	}
	if metaVer > xlMetaVersion {
		return 0, buf, fmt.Errorf("decodeXLHeaders: Unknown xl meta version %d", metaVer)
	}
	versions, buf, err = msgp.ReadIntBytes(buf)
	if err != nil {
		return 0, buf, err
	}
	if versions < 0 {
		return 0, buf, fmt.Errorf("decodeXLHeaders: Negative version count %d", versions)
	}
	return versions, buf, nil
}

// decodeVersions will decode a number of versions from a buffer
// and perform a callback for each version in order, newest first.
// Any non-nil error is returned.
func decodeVersions(buf []byte, versions int, fn func(idx int, hdr, meta []byte) error) (err error) {
	var tHdr, tMeta []byte // Zero copy bytes
	for i := 0; i < versions; i++ {
		tHdr, buf, err = msgp.ReadBytesZC(buf)
		if err != nil {
			return err
		}
		tMeta, buf, err = msgp.ReadBytesZC(buf)
		if err != nil {
			return err
		}
		if err = fn(i, tHdr, tMeta); err != nil {
			return err
		}
	}
	return nil
}

type xlMetaV2VersionHeaderV2 struct {
	VersionID [16]byte
	ModTime   int64
	Signature [4]byte
	Type      uint8
	Flags     uint8
}

// UnmarshalMsg implements msgp.Unmarshaler
func (z *xlMetaV2VersionHeaderV2) UnmarshalMsg(bts []byte) (o []byte, err error) {
	var zb0001 uint32
	zb0001, bts, err = msgp.ReadArrayHeaderBytes(bts)
	if err != nil {
		err = msgp.WrapError(err)
		return
	}
	if zb0001 != 5 {
		err = msgp.ArrayError{Wanted: 5, Got: zb0001}
		return
	}
	bts, err = msgp.ReadExactBytes(bts, (z.VersionID)[:])
	if err != nil {
		err = msgp.WrapError(err, "VersionID")
		return
	}
	z.ModTime, bts, err = msgp.ReadInt64Bytes(bts)
	if err != nil {
		err = msgp.WrapError(err, "ModTime")
		return
	}
	bts, err = msgp.ReadExactBytes(bts, (z.Signature)[:])
	if err != nil {
		err = msgp.WrapError(err, "Signature")
		return
	}
	{
		var zb0002 uint8
		zb0002, bts, err = msgp.ReadUint8Bytes(bts)
		if err != nil {
			err = msgp.WrapError(err, "Type")
			return
		}
		z.Type = zb0002
	}
	{
		var zb0003 uint8
		zb0003, bts, err = msgp.ReadUint8Bytes(bts)
		if err != nil {
			err = msgp.WrapError(err, "Flags")
			return
		}
		z.Flags = zb0003
	}
	o = bts
	return
}
