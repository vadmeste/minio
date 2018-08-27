// +build ignore

/*
 * Minio Cloud Storage, (C) 2017 Minio, Inc.
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 *
 */

package main

import (
	"io"
	"io/ioutil"
	"log"
	"time"

	"github.com/minio/minio/pkg/madmin"
)

func main() {
	// Note: YOUR-ACCESSKEYID, YOUR-SECRETACCESSKEY are
	// dummy values, please replace them with original values.

	// Note: YOUR-ACCESSKEYID, YOUR-SECRETACCESSKEY are
	// dummy values, please replace them with original values.

	// API requests are secure (HTTPS) if secure=true and insecure (HTTPS) otherwise.
	// New returns an Minio Admin client object.
	madmClnt, err := madmin.New("localhost:9000", "minio", "minio123", false)
	if err != nil {
		log.Fatalln(err)
	}

	profiler := "cpu"
	log.Println("Starting " + profiler + " profiling..")

	err = madmClnt.StartProfiling(profiler)
	if err != nil {
		log.Fatalln(err)
	}

	sleep := time.Duration(60)
	time.Sleep(time.Second * sleep)

	log.Println("Stopping profiling..")

	err = madmClnt.StopProfiling()
	if err != nil {
		log.Fatalln(err)
	}

	log.Println("Download profiling..")

	profilingData, err := madmClnt.DownloadProfilingData()
	if err != nil {
		log.Fatalln(err)
	}

	profilingFile, err := ioutil.TempFile("", profiler)
	if err != nil {
		log.Fatal(err)
	}

	if _, err := io.Copy(profilingFile, profilingData); err != nil {
		log.Fatal(err)
	}

	if err := profilingFile.Close(); err != nil {
		log.Fatal(err)
	}

	if err := profilingData.Close(); err != nil {
		log.Fatal(err)
	}

	log.Println("File " + profilingFile.Name() + " successfully generated.")
}
