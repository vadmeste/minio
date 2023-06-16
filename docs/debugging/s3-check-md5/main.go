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
	"context"
	"crypto/md5"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var (
	endpoint, accessKey, secretKey string
	minModTimeStr                  string
	bucket, prefix                 string
	onlyErrs, debug                bool
	versions                       bool
	insecure                       bool
)

// getMD5Sum returns MD5 sum of given data.
func getMD5Sum(data []byte) []byte {
	hash := md5.New()
	hash.Write(data)
	return hash.Sum(nil)
}

// DefaultTransport - this default transport is similar to
// http.DefaultTransport but with additional param  DisableCompression
// is set to true to avoid decompressing content with 'gzip' encoding.
var defaultTransport = func(secure bool) (*http.Transport, error) {
	tr := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   30 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext,
		MaxIdleConns:          256,
		MaxIdleConnsPerHost:   16,
		ResponseHeaderTimeout: 30 * time.Minute,
		IdleConnTimeout:       time.Minute,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 10 * time.Second,
		// Set this value so that the underlying transport round-tripper
		// doesn't try to auto decode the body of objects with
		// content-encoding set to `gzip`.
		//
		// Refer:
		//    https://golang.org/src/net/http/transport.go?h=roundTrip#L1843
		DisableCompression: true,
	}

	if secure {
		tr.TLSClientConfig = &tls.Config{
			// Can't use SSLv3 because of POODLE and BEAST
			// Can't use TLSv1.0 because of POODLE and BEAST using CBC cipher
			// Can't use TLSv1.1 because of RC4 cipher usage
			MinVersion: tls.VersionTLS12,
		}
	}
	return tr, nil
}

func main() {
	flag.StringVar(&endpoint, "endpoint", "https://play.min.io", "S3 endpoint URL")
	flag.StringVar(&accessKey, "access-key", "Q3AM3UQ867SPQQA43P2F", "S3 Access Key")
	flag.StringVar(&secretKey, "secret-key", "zuf+tfteSlswRu7BJ86wekitnifILbZam1KYY3TG", "S3 Secret Key")
	flag.StringVar(&bucket, "bucket", "", "Select a specific bucket")
	flag.StringVar(&prefix, "prefix", "", "Select a prefix")
	flag.BoolVar(&onlyErrs, "only-errors", false, "Print only errors")
	flag.BoolVar(&debug, "debug", false, "Prints HTTP network calls to S3 endpoint")
	flag.BoolVar(&versions, "versions", false, "Verify all versions")
	flag.BoolVar(&insecure, "insecure", false, "Disable TLS verification")
	flag.StringVar(&minModTimeStr, "modified-since", "", "Specify a minimum object last modified time, e.g.: 2023-01-02T15:04:05Z")
	flag.Parse()

	if endpoint == "" {
		log.Fatalln("Endpoint is not provided")
	}

	if accessKey == "" {
		log.Fatalln("Access key is not provided")
	}

	if secretKey == "" {
		log.Fatalln("Secret key is not provided")
	}

	if bucket == "" && prefix != "" {
		log.Fatalln("--prefix is specified without --bucket.")
	}

	var minModTime time.Time
	if minModTimeStr != "" {
		var e error
		minModTime, e = time.Parse(time.RFC3339, minModTimeStr)
		if e != nil {
			log.Fatalln("Unable to parse --modified-since:", e)
		}
	}

	u, err := url.Parse(endpoint)
	if err != nil {
		log.Fatalln(err)
	}

	secure := strings.EqualFold(u.Scheme, "https")
	transport, err := defaultTransport(secure)
	if err != nil {
		log.Fatalln(err)
	}
	if insecure {
		// skip TLS verification
		transport.TLSClientConfig.InsecureSkipVerify = true
	}

	s3Client, err := minio.New(u.Host, &minio.Options{
		Creds:     credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure:    secure,
		Transport: transport,
	})
	if err != nil {
		log.Fatalln(err)
	}

	if debug {
		s3Client.TraceOn(os.Stderr)
	}

	var buckets []string
	if bucket != "" {
		buckets = append(buckets, bucket)
	} else {
		bucketsInfo, err := s3Client.ListBuckets(context.Background())
		if err != nil {
			log.Fatalln(err)
		}
		for _, b := range bucketsInfo {
			buckets = append(buckets, b.Name)
		}
	}

	for _, bucket := range buckets {
		opts := minio.ListObjectsOptions{
			Recursive:    true,
			Prefix:       prefix,
			WithVersions: versions,
			WithMetadata: true,
		}

		objFullPath := func(obj minio.ObjectInfo) (fpath string) {
			fpath = path.Join(bucket, obj.Key)
			if versions {
				fpath += ":" + obj.VersionID
			}
			return
		}

		// List all objects from a bucket-name with a matching prefix.
		for object := range s3Client.ListObjects(context.Background(), bucket, opts) {
			if object.Err != nil {
				log.Println("FAILED: LIST with error:", object.Err)
				continue
			}
			if !minModTime.IsZero() && object.LastModified.Before(minModTime) {
				continue
			}
			if object.IsDeleteMarker {
				log.Println("SKIPPED: DELETE marker object:", objFullPath(object))
				continue
			}
			if _, ok := object.UserMetadata["X-Amz-Server-Side-Encryption-Customer-Algorithm"]; ok {
				log.Println("SKIPPED: Objects encrypted with SSE-C do not have md5sum as ETag:", objFullPath(object))
				continue
			}
			if v, ok := object.UserMetadata["X-Amz-Server-Side-Encryption"]; ok && v == "aws:kms" {
				log.Println("FAILED: encrypted with SSE-KMS do not have md5sum as ETag:", objFullPath(object))
				continue
			}
			parts := 1
			multipart := false
			s := strings.Split(object.ETag, "-")
			switch len(s) {
			case 1:
				// nothing to do
			case 2:
				if p, err := strconv.Atoi(s[1]); err == nil {
					parts = p
				} else {
					log.Println("FAILED: ETAG of", objFullPath(object), "has a wrong format:", err)
					continue
				}
				multipart = true
			default:
				log.Println("FAILED: Unexpected ETAG", object.ETag, "for object:", objFullPath(object))
				continue
			}

			var partsMD5Sum [][]byte
			var failedMD5 bool
			for p := 1; p <= parts; p++ {
				opts := minio.GetObjectOptions{
					VersionID:  object.VersionID,
					PartNumber: p,
				}
				obj, err := s3Client.GetObject(context.Background(), bucket, object.Key, opts)
				if err != nil {
					log.Println("FAILED: GET", objFullPath(object), "=>", err)
					failedMD5 = true
					break
				}
				h := md5.New()
				if _, err := io.Copy(h, obj); err != nil {
					log.Println("FAILED: MD5 calculation error:", objFullPath(object), "=>", err)
					failedMD5 = true
					break
				}
				partsMD5Sum = append(partsMD5Sum, h.Sum(nil))
			}

			if failedMD5 {
				log.Println("CORRUPTED object:", objFullPath(object))
				continue
			}

			corrupted := false
			if !multipart {
				md5sum := fmt.Sprintf("%x", partsMD5Sum[0])
				if md5sum != object.ETag {
					corrupted = true
				}
			} else {
				var totalMD5SumBytes []byte
				for _, sum := range partsMD5Sum {
					totalMD5SumBytes = append(totalMD5SumBytes, sum...)
				}
				s3MD5 := fmt.Sprintf("%x-%d", getMD5Sum(totalMD5SumBytes), parts)
				if s3MD5 != object.ETag {
					corrupted = true
				}
			}

			if corrupted {
				log.Println("CORRUPTED object:", objFullPath(object))
			} else {
				if !onlyErrs {
					log.Println("INTACT object:", objFullPath(object))
				}
			}
		}
	}
}
