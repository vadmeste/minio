/*
 * Minio Cloud Storage, (C) 2017, 2018 Minio, Inc.
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

package madmin

import (
	"errors"
	"fmt"
	"io"
	"net/http"
)

func (adm *AdminClient) StartProfiling(profiler string) error {
	path := fmt.Sprintf("/v1/profiling/start/%s", profiler)
	resp, err := adm.executeMethod("POST", requestData{
		relPath: path,
	})
	defer closeResponse(resp)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return httpRespToErrorResponse(resp)
	}

	return nil
}

func (adm *AdminClient) StopProfiling() error {
	path := fmt.Sprintf("/v1/profiling/stop")
	resp, err := adm.executeMethod("POST", requestData{
		relPath: path,
	})
	defer closeResponse(resp)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusNoContent {
		return httpRespToErrorResponse(resp)
	}

	return nil
}

func (adm *AdminClient) DownloadProfilingData() (io.ReadCloser, error) {
	path := fmt.Sprintf("/v1/profiling/download")
	resp, err := adm.executeMethod("GET", requestData{
		relPath: path,
	})

	if err != nil {
		closeResponse(resp)
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		defer closeResponse(resp)
		return nil, httpRespToErrorResponse(resp)
	}

	if resp.Body == nil {
		return nil, errors.New("body is nil")
	}

	return resp.Body, nil
}
