// Copyright (c) 2015-2022 MinIO, Inc.
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

package cache

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/minio/minio/internal/config"
	xhttp "github.com/minio/minio/internal/http"
	"github.com/minio/pkg/v2/env"
	"github.com/tinylib/msgp/msgp"
)

// Cache related keys
const (
	Enable      = "enable"
	PrefixDepth = "prefix_depth"
	Endpoint    = "endpoint"
	ObjectSize  = "object_size"

	EnvEnable      = "MINIO_CACHE_ENABLE"
	EnvPrefixDepth = "MINIO_CACHE_PREFIX_DEPTH"
	EnvEndpoint    = "MINIO_CACHE_ENDPOINT"
	EnvObjectSize  = "MINIO_CACHE_OBJECT_SIZE"
)

// DefaultKVS - default KV config for cache settings
var DefaultKVS = config.KVS{
	config.KV{
		Key:   Enable,
		Value: "off",
	},
	config.KV{
		Key:   PrefixDepth,
		Value: "",
	},
	config.KV{
		Key:   Endpoint,
		Value: "",
	},
	config.KV{
		Key:   ObjectSize,
		Value: "",
	},
}

// Config represents the subnet related configuration
type Config struct {
	// Flag indicating whether cache is enabled.
	Enable bool `json:"enable"`

	// Prefix depth indicates the tree depth at which MinIO will
	// create additional hardlinks per xl.meta for lookup()
	// efficiency on reads().
	PrefixDepth int `json:"prefixDepth"`

	// Endpoint for caching uses remote mcache server to
	// store and retrieve pre-condition check entities such as
	// Etag and ModTime of an object + version
	Endpoint string `json:"endpoint"`

	// ObjectSize indicates the maximum object size below which
	// data is cached and fetched remotely from DRAM.
	ObjectSize int64

	// Is the HTTP client used for communicating with mcache server
	clnt *http.Client
}

var configLock sync.RWMutex

// Enabled - indicates if cache is enabled or not
func (c *Config) Enabled() bool {
	configLock.RLock()
	defer configLock.RUnlock()

	return c.Enable && c.PrefixDepth > 0
}

// EnabledRemote - indicates if cache is enabled with remote endpoint
func (c *Config) EnabledRemote() bool {
	configLock.RLock()
	defer configLock.RUnlock()

	return c.Enable && c.Endpoint != ""
}

// MatchesDepth returns true if the input object hierarchy matches the configured prefix depth
func (c Config) MatchesDepth(object string) bool {
	configLock.RLock()
	defer configLock.RUnlock()

	if c.Enable && c.PrefixDepth > 0 {
		return strings.Count(object, "/") >= c.PrefixDepth
	}
	return false
}

// MatchesSize verifies if input 'size' falls under cacheable threshold
func (c Config) MatchesSize(size int64) bool {
	configLock.RLock()
	defer configLock.RUnlock()

	return c.Enable && c.ObjectSize > 0 && size <= c.ObjectSize
}

// Update updates new cache frequency
func (c *Config) Update(ncfg Config) {
	configLock.Lock()
	defer configLock.Unlock()

	c.Enable = ncfg.Enable
	c.Endpoint = ncfg.Endpoint
	c.PrefixDepth = ncfg.PrefixDepth
	c.ObjectSize = ncfg.ObjectSize
	c.clnt = ncfg.clnt
}

// cache related errors
var (
	ErrInvalidArgument = errors.New("invalid argument")
	ErrKeyMissing      = errors.New("key is missing")
)

const (
	mcacheV1Check  = "/_mcache/v1/check"
	mcacheV1Update = "/_mcache/v1/update"
	mcacheV1Delete = "/_mcache/v1/delete"
)

// Get performs conditional check and returns the cached object info if any.
func (c Config) Get(r *CondCheck) (*ObjectInfo, error) {
	configLock.RLock()
	defer configLock.RUnlock()

	if !c.Enable {
		return nil, nil
	}

	if c.Endpoint == "" {
		// Endpoint not set, make this a no-op
		return nil, nil
	}

	buf, err := r.MarshalMsg(nil)
	if err != nil {
		return nil, err
	}

	// We do not want Get's to take so much time, anything
	// beyond 250ms we should cut it, remote cache is too
	// busy already.
	ctx, cancel := context.WithTimeout(context.Background(), 250*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.Endpoint+mcacheV1Check, bytes.NewReader(buf))
	if err != nil {
		return nil, err
	}

	resp, err := c.clnt.Do(req)
	if err != nil {
		return nil, err
	}
	defer xhttp.DrainBody(resp.Body)

	switch resp.StatusCode {
	case http.StatusNotFound:
		return nil, ErrKeyMissing
	case http.StatusOK:
		co := &ObjectInfo{}
		return co, co.DecodeMsg(msgp.NewReader(resp.Body))
	default:
		return nil, ErrInvalidArgument
	}
}

// Set sets the cache object info
func (c Config) Set(ci *ObjectInfo) {
	configLock.RLock()
	defer configLock.RUnlock()

	if !c.Enable {
		return
	}

	if c.Endpoint == "" {
		// Endpoint not set, make this a no-op
		return
	}

	buf, err := ci.MarshalMsg(nil)
	if err != nil {
		return
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPut, c.Endpoint+mcacheV1Update, bytes.NewReader(buf))
	if err != nil {
		return
	}

	resp, err := c.clnt.Do(req)
	if err != nil {
		return
	}
	defer xhttp.DrainBody(resp.Body)
}

// Delete deletes remote cached content for object and its version.
func (c Config) Delete(bucket, key string) {
	configLock.RLock()
	defer configLock.RUnlock()

	if !c.Enable {
		return
	}

	if c.Endpoint == "" {
		return
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodDelete, c.Endpoint+fmt.Sprintf("%s?bucket=%s&key=%s", mcacheV1Delete, bucket, key), nil)
	if err != nil {
		return
	}

	resp, err := c.clnt.Do(req)
	if err != nil {
		return
	}
	defer xhttp.DrainBody(resp.Body)
}

// LookupConfig - lookup config and override with valid environment settings if any.
func LookupConfig(kvs config.KVS, transport http.RoundTripper) (cfg Config, err error) {
	cfg.Enable = env.Get(EnvEnable, kvs.GetWithDefault(Enable, DefaultKVS)) == config.EnableOn

	if d := env.Get(EnvPrefixDepth, kvs.GetWithDefault(PrefixDepth, DefaultKVS)); d != "" {
		prefixDepth, err := strconv.Atoi(d)
		if err != nil {
			return cfg, err
		}
		cfg.PrefixDepth = prefixDepth
	}

	if d := env.Get(EnvObjectSize, kvs.GetWithDefault(ObjectSize, DefaultKVS)); d != "" {
		objectSize, err := humanize.ParseBytes(d)
		if err != nil {
			return cfg, err
		}
		cfg.ObjectSize = int64(objectSize)
	}

	cfg.Endpoint = env.Get(EnvEndpoint, kvs.GetWithDefault(Endpoint, DefaultKVS))
	cfg.clnt = &http.Client{Transport: transport}

	return cfg, nil
}
