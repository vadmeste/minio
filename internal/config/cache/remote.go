package cache

import (
	"net/http"
	"regexp"
	"time"

	"github.com/minio/minio/internal/amztime"
	xhttp "github.com/minio/minio/internal/http"
)

// ObjectInfo represents the object information cached remotely
type ObjectInfo struct {
	Key        string    `json:"key"`
	VID        string    `json:"vid"`
	Bucket     string    `json:"bucket"`
	ETag       string    `json:"etag"`
	ModTime    time.Time `json:"modTime"`
	StatusCode int       `json:"-"`
}

// CondCheck represents the conditional request made to the remote cache
// for validation during GET/HEAD object requests.
type CondCheck struct {
	ObjectInfo
	IfMatch           string     `json:"ifMatch,omitempty"`
	IfNoneMatch       string     `json:"ifNoneMatch,omitempty"`
	IfModifiedSince   *time.Time `json:"ifModSince,omitempty"`
	IfUnModifiedSince *time.Time `json:"ifUnmodSince,omitempty"`
}

// IsSet tells the cache lookup to avoid sending a request
func (r *CondCheck) IsSet() bool {
	if r == nil {
		return false
	}
	return r.IfMatch != "" || r.IfNoneMatch != "" || r.IfModifiedSince != nil || r.IfUnModifiedSince != nil
}

var etagRegex = regexp.MustCompile("\"*?([^\"]*?)\"*?$")

// canonicalizeETag returns ETag with leading and trailing double-quotes removed,
// if any present
func canonicalizeETag(etag string) string {
	return etagRegex.ReplaceAllString(etag, "$1")
}

// Init - populates the input values, initializes CondCheck
// before sending the request remotely.
func (r *CondCheck) Init(bucket, object, vid string, header http.Header) {
	r.Key = object
	r.VID = vid
	r.Bucket = bucket

	ifModifiedSinceHeader := header.Get(xhttp.IfModifiedSince)
	if ifModifiedSinceHeader != "" {
		if givenTime, err := amztime.ParseHeader(ifModifiedSinceHeader); err == nil {
			r.IfModifiedSince = &givenTime
		}
	}
	ifUnmodifiedSinceHeader := header.Get(xhttp.IfUnmodifiedSince)
	if ifUnmodifiedSinceHeader != "" {
		if givenTime, err := amztime.ParseHeader(ifUnmodifiedSinceHeader); err == nil {
			r.IfUnModifiedSince = &givenTime
		}
	}
	r.IfMatch = canonicalizeETag(header.Get(xhttp.IfMatch))
	r.IfNoneMatch = canonicalizeETag(header.Get(xhttp.IfNoneMatch))
}
