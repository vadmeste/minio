package main

import (
	"bufio"
	"encoding/binary"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"hash/crc32"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/dchest/siphash"
	"github.com/google/uuid"
	"github.com/zeebo/xxh3"
)

// hashes the key returning an integer based on the input algorithm.
// This function currently supports
// - SIPMOD
func sipHashMod(key string, cardinality int, id [16]byte) int {
	if cardinality <= 0 {
		return -1
	}
	// use the faster version as per siphash docs
	// https://github.com/dchest/siphash#usage
	k0, k1 := binary.LittleEndian.Uint64(id[0:8]), binary.LittleEndian.Uint64(id[8:16])
	sum64 := siphash.Hash(k0, k1, []byte(key))
	return int(sum64 % uint64(cardinality))
}

// hashOrder - hashes input key to return consistent
// hashed integer slice. Returned integer order is salted
// with an input key. This results in consistent order.
// NOTE: collisions are fine, we are not looking for uniqueness
// in the slices returned.
func hashOrder(key string, cardinality int) []int {
	if cardinality <= 0 {
		// Returns an empty int slice for cardinality < 0.
		return nil
	}

	nums := make([]int, cardinality)
	keyCrc := crc32.Checksum([]byte(key), crc32.IEEETable)

	start := int(keyCrc % uint32(cardinality))
	for i := 1; i <= cardinality; i++ {
		nums[i-1] = 1 + ((start + i) % cardinality)
	}
	return nums
}

const (
	minioMetaBucket   = ".minio.sys"
	prefixCachePrefix = ".pcache"
)

var sep = string(filepath.Separator)

var (
	source      = flag.String("source", "", "list of objects")
	failOnErr   = flag.Bool("fail-on-error", false, "fail on errors")
	bucket      = flag.String("bucket", "", "bucket name")
	minPrefixes = flag.Int("min-prefixes", 3, "bucket name")
	debug       = flag.Bool("debug", false, "enable verbose debugging mode")
)

type walker struct {
	drive     string
	ignoreDir string

	name  string
	count uint64
	set   int

	deploymentID string
	setCount     int
}

type xl struct {
	This string     `json:"this"`
	Sets [][]string `json:"sets"`
}

type format struct {
	ID string `json:"id"`
	XL xl     `json:"xl"`
}

func newWalker(drive string) (walker, error) {
	drive, err := filepath.Abs(drive)
	if err != nil {
		return walker{}, err
	}

	data, err := os.ReadFile(filepath.Join(drive, ".minio.sys", "format.json"))
	if err != nil {
		return walker{}, err
	}

	var f format
	err = json.Unmarshal(data, &f)
	if err != nil {
		return walker{}, err
	}

	setIdx := 0
	for idx := range f.XL.Sets {
		for k := range f.XL.Sets[idx] {
			if f.XL.Sets[idx][k] == f.XL.This {
				setIdx = idx
			}
		}
	}

	w := walker{
		drive:        drive,
		name:         drive,
		set:          setIdx,
		ignoreDir:    filepath.Join(drive, minioMetaBucket) + sep,
		deploymentID: f.ID,
		setCount:     len(f.XL.Sets),
	}

	return w, nil
}

func (w *walker) populateFromFile(source, bucket string, failOnErr bool) error {
	file, err := os.Open(source)
	if err != nil {
		return err
	}
	defer file.Close()

	var id uuid.UUID
	if w.deploymentID != "" {
		id = uuid.MustParse(w.deploymentID)
	}

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		object := scanner.Text()
		err := w.mkLink(bucket, object, w.setCount, id)
		if err != nil && failOnErr {
			log.Println("fail with object:", object, "=> err =", err)
			return err
		}
	}

	if err := scanner.Err(); err != nil {
		return err
	}

	return nil
}

func (w *walker) scan() error {
	fmt.Println("scanning", w.drive)
	return filepath.WalkDir(w.drive, w.walkFunc)
}

func (w *walker) walkFunc(originPath string, entry fs.DirEntry, err error) error {
	if err != nil {
		return nil
	}
	if entry.IsDir() {
		return nil
	}
	if strings.HasPrefix(originPath, w.ignoreDir) {
		return nil
	}
	if !strings.HasSuffix(originPath, "/xl.meta") {
		return nil
	}

	path := strings.TrimPrefix(originPath, w.drive+sep)
	path = strings.TrimSuffix(path, "/xl.meta")

	i := strings.Index(path, sep)
	bucket, object := path[:i], path[i+1:]

	return w.mkLink(bucket, object, 0, uuid.UUID{})
}

func osLink(src, dst string) error {
	err := os.Link(src, dst)
	if err != nil && *debug {
		fmt.Printf("os.Link(%s, %s) = %v\n", src, dst, err)
	}
	return err
}

func (w *walker) mkLink(bucket, object string, setCount int, id uuid.UUID) error {
	if strings.Count(object, sep) < *minPrefixes {
		return nil
	}

	var nohash bool
	if setCount <= 0 {
		nohash = true
	}

	var set int
	if !nohash {
		set = sipHashMod(object, setCount, id)
	}

	ohashUint := xxh3.HashString128(object)
	bhash := strconv.FormatUint(xxh3.HashString(bucket), 16)
	ohash := filepath.Join(strconv.FormatUint(ohashUint.Hi, 16), strconv.FormatUint(ohashUint.Lo, 16)) + ".meta"
	linkDir := filepath.Join(w.drive, minioMetaBucket, prefixCachePrefix, bhash, filepath.Dir(ohash))

	if !nohash && set != w.set {
		// Object is not part of this set, skip it.
		return nil
	}

	// fmt.Println("D mkdir", linkDir, " => link", filepath.Join(linkDir, filepath.Base(ohash)))
	err := os.MkdirAll(linkDir, 0o777)
	if err != nil {
		return err
	}

	atomic.AddUint64(&w.count, 1)

	err = osLink(filepath.Join(w.drive, bucket, object, "xl.meta"), filepath.Join(linkDir, filepath.Base(ohash)))
	if errors.Is(err, os.ErrExist) {
		return nil
	}

	return err
}

func main() {
	flag.Parse()

	walkers := make([]walker, flag.NArg())

	var scanWg, statWg sync.WaitGroup

	if flag.NArg() <= 0 && *source == "" {
		log.Fatal("this tool requires an input either via command args or -source")
	}

	for i := 0; i < flag.NArg(); i++ {
		arg, err := filepath.Abs(flag.Arg(i))
		if err != nil {
			log.Fatal(err)
		}
		w, err := newWalker(arg)
		if err != nil {
			log.Fatal(err)
		}
		walkers[i] = w
		scanWg.Add(1)
		go func(i int) {
			defer scanWg.Done()
			var err error
			if *source != "" {
				err = walkers[i].populateFromFile(*source, *bucket, *failOnErr)
			} else {
				err = walkers[i].scan()
			}
			if err != nil {
				log.Println(arg, "=>", err)
			}
		}(i)
	}

	closed := make(chan struct{})

	timer := time.NewTimer(time.Minute)
	defer timer.Stop()

	statWg.Add(1)
	go func() {
		defer statWg.Done()
		closing := false
		for {
			fmt.Println(time.Now(), "Print hardlink creations stats..")
			for i := range walkers {
				c := atomic.LoadUint64(&walkers[i].count)
				fmt.Printf("%v=%v ", walkers[i].name, c)
			}
			fmt.Println("")

			if closing {
				return
			}

			select {
			case <-closed:
				closing = true
			case <-timer.C:
				timer.Reset(time.Minute)
			}
		}
	}()

	scanWg.Wait()
	close(closed)
	statWg.Wait()

	log.Println("All finished without errors..")
}
