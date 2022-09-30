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

package cmd

import (
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio/internal/auth"
	xhttp "github.com/minio/minio/internal/http"
)

// credentialHeader data type represents structured form of Credential
// string from authorization header.
type credentialHeader struct {
	accessKey string
	scope     struct {
		date    time.Time
		region  string
		service string
		request string
	}

	multiRegion bool
}

// Return scope string.
func (c credentialHeader) getScope() string {
	var elems []string
	if c.multiRegion {
		elems = []string{
			c.scope.date.Format(yyyymmdd),
			c.scope.service,
			c.scope.request,
		}
	} else {
		elems = []string{
			c.scope.date.Format(yyyymmdd),
			c.scope.region,
			c.scope.service,
			c.scope.request,
		}
	}

	return strings.Join(elems, SlashSeparator)
}

func getReqAccessKeyV4(r *http.Request, region string, stype serviceType, v4a bool) (auth.Credentials, bool, APIErrorCode) {
	var (
		ch    credentialHeader
		s3Err APIErrorCode
	)

	if credsForm := r.Form.Get(xhttp.AmzCredential); credsForm != "" {
		ch, s3Err = parseCredentialHeader("Credential="+credsForm, region, stype, v4a)
	} else {
		v4Auth := strings.TrimSpace(r.Header.Get("Authorization"))
		if v4a {
			v4Auth = strings.TrimPrefix(v4Auth, signV4AAlgorithm)
		} else {
			v4Auth = strings.TrimPrefix(v4Auth, signV4Algorithm)
		}
		authFields := strings.Split(v4Auth, ",")
		if len(authFields) != 3 {
			return auth.Credentials{}, false, ErrMissingFields
		}
		ch, s3Err = parseCredentialHeader(authFields[0], region, stype, v4a)
	}
	if s3Err != ErrNone {
		return auth.Credentials{}, false, s3Err
	}
	return checkKeyValid(r, ch.accessKey)
}

// parse credentialHeader string into its structured form.
func parseCredentialHeader(credElement string, systemRegion string, stype serviceType, multiRegion bool) (ch credentialHeader, aec APIErrorCode) {
	creds := strings.SplitN(strings.TrimSpace(credElement), "=", 2)
	if len(creds) != 2 {
		return ch, ErrMissingFields
	}
	if creds[0] != "Credential" {
		return ch, ErrMissingCredTag
	}
	credElements := strings.Split(strings.TrimSpace(creds[1]), SlashSeparator)

	var (
		accessKey, region, date     string
		serviceType, requestVersion string
	)
	if multiRegion {
		if len(credElements) < 4 {
			return ch, ErrCredMalformed
		}
		accessKey = strings.Join(credElements[:len(credElements)-3], SlashSeparator) // The access key may contain one or more `/`
		credElements = credElements[len(credElements)-3:]
		date = credElements[0]
		serviceType = credElements[1]
		requestVersion = credElements[2]

	} else {
		if len(credElements) < 5 {
			return ch, ErrCredMalformed
		}
		accessKey = strings.Join(credElements[:len(credElements)-4], SlashSeparator) // The access key may contain one or more `/`
		credElements = credElements[len(credElements)-4:]
		date = credElements[0]
		region = credElements[1]
		serviceType = credElements[2]
		requestVersion = credElements[3]
	}

	// Validate and isave access key id.
	if !auth.IsAccessKeyValid(accessKey) {
		return ch, ErrInvalidAccessKeyID
	}
	ch.accessKey = accessKey
	// Parse and save the date
	var e error
	ch.scope.date, e = time.Parse(yyyymmdd, date)
	if e != nil {
		return ch, ErrMalformedCredentialDate
	}
	if multiRegion {
		ch.multiRegion = true
	} else {
		// Region is set to be empty, we use whatever was sent by the
		// request and proceed further. This is a work-around to address
		// an important problem for ListBuckets() getting signed with
		// different regions.
		if systemRegion == "" {
			systemRegion = region
		}
		// Should validate region, only if region is set.
		if !isValidRegion(region, systemRegion) {
			return ch, ErrAuthorizationHeaderMalformed
		}
		ch.scope.region = region
	}
	// Validate and save service type
	if serviceType != string(stype) {
		switch stype {
		case serviceSTS:
			return ch, ErrInvalidServiceSTS
		}
		return ch, ErrInvalidServiceS3
	}
	ch.scope.service = serviceType
	// Validate and save request version
	if requestVersion != "aws4_request" {
		return ch, ErrInvalidRequestVersion
	}
	ch.scope.request = requestVersion

	return ch, ErrNone
}

// Parse signature from signature tag.
func parseSignature(signElement string) (string, APIErrorCode) {
	signFields := strings.Split(strings.TrimSpace(signElement), "=")
	if len(signFields) != 2 {
		return "", ErrMissingFields
	}
	if signFields[0] != "Signature" {
		return "", ErrMissingSignTag
	}
	if signFields[1] == "" {
		return "", ErrMissingFields
	}
	signature := signFields[1]
	return signature, ErrNone
}

// Parse slice of signed headers from signed headers tag.
func parseSignedHeader(signedHdrElement string) ([]string, APIErrorCode) {
	signedHdrFields := strings.Split(strings.TrimSpace(signedHdrElement), "=")
	if len(signedHdrFields) != 2 {
		return nil, ErrMissingFields
	}
	if signedHdrFields[0] != "SignedHeaders" {
		return nil, ErrMissingSignHeadersTag
	}
	if signedHdrFields[1] == "" {
		return nil, ErrMissingFields
	}
	signedHeaders := strings.Split(signedHdrFields[1], ";")
	return signedHeaders, ErrNone
}

// signValues data type represents structured form of AWS Signature V4 header.
type signValues struct {
	Credential    credentialHeader
	SignedHeaders []string
	Signature     string
}

// preSignValues data type represents structued form of AWS Signature V4 query string.
type preSignValues struct {
	signValues
	Date    time.Time
	Expires time.Duration
}

// Parses signature version '4' query string of the following form.
//
//	querystring = X-Amz-Algorithm=algorithm
//	querystring += &X-Amz-Credential= urlencode(accessKey + '/' + credential_scope)
//	querystring += &X-Amz-Date=date
//	querystring += &X-Amz-Expires=timeout interval
//	querystring += &X-Amz-SignedHeaders=signed_headers
//	querystring += &X-Amz-Signature=signature
//
// verifies if any of the necessary query params are missing in the presigned request.
func doesV4PresignParamsExist(query url.Values) APIErrorCode {
	v4PresignQueryParams := []string{xhttp.AmzAlgorithm, xhttp.AmzCredential, xhttp.AmzSignature, xhttp.AmzDate, xhttp.AmzSignedHeaders, xhttp.AmzExpires}
	for _, v4PresignQueryParam := range v4PresignQueryParams {
		if _, ok := query[v4PresignQueryParam]; !ok {
			return ErrInvalidQueryParams
		}
	}
	return ErrNone
}

// Parses all the presigned signature values into separate elements.
func parsePreSignV4(query url.Values, region string, stype serviceType, isV4A bool) (psv preSignValues, aec APIErrorCode) {
	// verify whether the required query params exist.
	aec = doesV4PresignParamsExist(query)
	if aec != ErrNone {
		return psv, aec
	}

	// Verify if the query algorithm is supported or not.
	switch query.Get(xhttp.AmzAlgorithm) {
	case signV4Algorithm, signV4AAlgorithm:
	default:
		return psv, ErrInvalidQuerySignatureAlgo
	}

	// Initialize signature version '4' structured header.
	preSignV4Values := preSignValues{}

	// Save credential.
	preSignV4Values.Credential, aec = parseCredentialHeader("Credential="+query.Get(xhttp.AmzCredential), region, stype, isV4A)
	if aec != ErrNone {
		return psv, aec
	}

	var e error
	// Save date in native time.Time.
	preSignV4Values.Date, e = time.Parse(iso8601Format, query.Get(xhttp.AmzDate))
	if e != nil {
		return psv, ErrMalformedPresignedDate
	}

	// Save expires in native time.Duration.
	preSignV4Values.Expires, e = time.ParseDuration(query.Get(xhttp.AmzExpires) + "s")
	if e != nil {
		return psv, ErrMalformedExpires
	}

	if preSignV4Values.Expires < 0 {
		return psv, ErrNegativeExpires
	}

	// Check if Expiry time is less than 7 days (value in seconds).
	if preSignV4Values.Expires.Seconds() > 604800 {
		return psv, ErrMaximumExpires
	}

	// Save signed headers.
	preSignV4Values.SignedHeaders, aec = parseSignedHeader("SignedHeaders=" + query.Get(xhttp.AmzSignedHeaders))
	if aec != ErrNone {
		return psv, aec
	}

	// Save signature.
	preSignV4Values.Signature, aec = parseSignature("Signature=" + query.Get(xhttp.AmzSignature))
	if aec != ErrNone {
		return psv, aec
	}

	// Return structed form of signature query string.
	return preSignV4Values, ErrNone
}

// Parses signature version '4' header of the following form.
//
//	Authorization: algorithm Credential=accessKeyID/credScope, \
//	        SignedHeaders=signedHeaders, Signature=signature
func parseSignV4(v4AuthHdr string, region string, stype serviceType, isV4A bool) (sv signValues, aec APIErrorCode) {

	credElement := strings.Split(strings.TrimSpace(v4AuthHdr), ",")[0]

	// Replace all spaced strings, some clients can send spaced
	// parameters and some won't. So we pro-actively remove any spaces
	// to make parsing easier.
	v4Auth := strings.ReplaceAll(v4AuthHdr, " ", "")
	if v4Auth == "" {
		return sv, ErrAuthHeaderEmpty
	}

	if isV4A {
		if !strings.HasPrefix(v4Auth, signV4AAlgorithm) {
			return sv, ErrSignatureVersionNotSupported
		}
		v4Auth = strings.TrimPrefix(v4Auth, signV4AAlgorithm)
		credElement = strings.TrimPrefix(credElement, signV4AAlgorithm)
	} else {
		if !strings.HasPrefix(v4Auth, signV4Algorithm) {
			return sv, ErrSignatureVersionNotSupported
		}
		v4Auth = strings.TrimPrefix(v4Auth, signV4Algorithm)
		credElement = strings.TrimPrefix(credElement, signV4Algorithm)
	}

	// Strip off the Algorithm prefix.
	authFields := strings.Split(v4Auth, ",")
	if len(authFields) != 3 {
		return sv, ErrMissingFields
	}

	// Initialize signature version '4' structured header.
	signV4Values := signValues{}

	var s3Err APIErrorCode
	// Save credentail values.
	signV4Values.Credential, s3Err = parseCredentialHeader(strings.TrimSpace(credElement), region, stype, isV4A)
	if s3Err != ErrNone {
		return sv, s3Err
	}

	// Save signed headers.
	signV4Values.SignedHeaders, s3Err = parseSignedHeader(authFields[1])
	if s3Err != ErrNone {
		return sv, s3Err
	}

	// Save signature.
	signV4Values.Signature, s3Err = parseSignature(authFields[2])
	if s3Err != ErrNone {
		return sv, s3Err
	}

	// Return the structure here.
	return signV4Values, ErrNone
}
