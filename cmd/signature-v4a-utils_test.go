package cmd

import (
	"encoding/hex"
	"fmt"
	"net/http"
	"net/url"
	"testing"
	"time"
)

func TestDoesV4ASignatureMatch(t *testing.T) {
	credentialTemplate := "%s/%s/s3/aws4_request"
	now := UTCNow()
	accessKey := "minioadmin"
	secretKey := "minioadmin"

	testCases := []struct {
		form     http.Header
		expected APIErrorCode
	}{
		// (0) It should fail if 'X-Amz-Credential' is missing.
		{
			form:     http.Header{},
			expected: ErrCredMalformed,
		},
		// (1) It should fail if the access key is incorrect.
		{
			form: http.Header{
				"X-Amz-Credential": []string{fmt.Sprintf(credentialTemplate, "EXAMPLEINVALIDEXAMPL", now.Format(yyyymmdd))},
			},
			expected: ErrInvalidAccessKeyID,
		},
		// (2) It should fail without signature
		{
			form: http.Header{
				"X-Amz-Credential": []string{fmt.Sprintf(credentialTemplate, accessKey, now.Format(yyyymmdd))},
				"X-Amz-Date":       []string{now.Format(iso8601Format)},
			},
			expected: ErrSignatureDoesNotMatch,
		},
		// (2) It should fail with empty signature
		{
			form: http.Header{
				"X-Amz-Credential": []string{fmt.Sprintf(credentialTemplate, accessKey, now.Format(yyyymmdd))},
				"X-Amz-Date":       []string{now.Format(iso8601Format)},
				"X-Amz-Signature":  []string{""},
			},
			expected: ErrSignatureDoesNotMatch,
		},
		// (2) It should fail with a bad signature.
		{
			form: http.Header{
				"X-Amz-Credential": []string{fmt.Sprintf(credentialTemplate, accessKey, now.Format(yyyymmdd))},
				"X-Amz-Date":       []string{now.Format(iso8601Format)},
				"X-Amz-Signature":  []string{"invalidsignature"},
			},
			expected: ErrSignatureDoesNotMatch,
		},
		// (3) It should succeed if everything is correct.
		{
			form: http.Header{
				"X-Amz-Credential": []string{
					fmt.Sprintf(credentialTemplate, accessKey, now.Format(yyyymmdd)),
				},
				"X-Amz-Date": []string{now.Format(iso8601Format)},
				"X-Amz-Signature": []string{
					getSignature(getSigningKey(globalActiveCred.SecretKey, now,
						globalMinioDefaultRegion, serviceS3), "policy"),
				},
			},
			expected: ErrNone,
		},
	}

	// Run each test case individually.
	for i, testCase := range testCases {
		_, code := doesSignatureMatch(testCase.form)
		if code != testCase.expected {
			t.Errorf("(%d) expected to get %s, instead got %s", i, niceError(testCase.expected), niceError(code))
		}
	}
}

func TestDoesPresignedSignatureMatch(t *testing.T) {
	// sha256 hash of "payload"
	payloadSHA256 := "239f59ed55e737c77147cf55ad0c1b030b6d7ee748a7426952f9b852d5a935e5"
	now := UTCNow()
	credentialTemplate := "%s/%s/s3/aws4_request"

	region := globalSite.Region
	accessKeyID := globalActiveCred.AccessKey
	testCases := []struct {
		queryParams map[string]string
		headers     map[string]string
		region      string
		expected    APIErrorCode
	}{
		// (0) Should error without a set URL query.
		{
			expected: ErrInvalidQueryParams,
		},
		// (1) Should error on an invalid access key.
		{
			queryParams: map[string]string{
				"X-Amz-Algorithm":     signV4AAlgorithm,
				"X-Amz-Date":          now.Format(iso8601Format),
				"X-Amz-Expires":       "60",
				"X-Amz-Signature":     "badsignature",
				"X-Amz-SignedHeaders": "host;x-amz-content-sha256;x-amz-date",
				"X-Amz-Credential":    fmt.Sprintf(credentialTemplate, "Z7IXGOO6BZ0REAN1Q26I", now.Format(yyyymmdd)),
			},
			expected: ErrInvalidAccessKeyID,
		},
		{
			queryParams: map[string]string{
				"X-Amz-Algorithm":      signV4AAlgorithm,
				"X-Amz-Date":           now.Format(iso8601Format),
				"X-Amz-Expires":        "60",
				"X-Amz-Signature":      "",
				"X-Amz-SignedHeaders":  "host;x-amz-content-sha256;x-amz-date",
				"X-Amz-Credential":     fmt.Sprintf(credentialTemplate, accessKeyID, now.Format(yyyymmdd)),
				"X-Amz-Content-Sha256": payloadSHA256,
			},
			expected: ErrUnsignedHeaders,
		},
		// (3) Should fail to extract headers if the host header is not signed.
		{
			queryParams: map[string]string{
				"X-Amz-Algorithm":      signV4AAlgorithm,
				"X-Amz-Date":           now.Format(iso8601Format),
				"X-Amz-Expires":        "60",
				"X-Amz-Signature":      "badsignature",
				"X-Amz-SignedHeaders":  "x-amz-content-sha256;x-amz-date",
				"X-Amz-Credential":     fmt.Sprintf(credentialTemplate, accessKeyID, now.Format(yyyymmdd)),
				"X-Amz-Content-Sha256": payloadSHA256,
			},
			expected: ErrUnsignedHeaders,
		},
		// (4) Should give an expired request if it has expired.
		{
			queryParams: map[string]string{
				"X-Amz-Algorithm":      signV4AAlgorithm,
				"X-Amz-Date":           now.AddDate(0, 0, -2).Format(iso8601Format),
				"X-Amz-Expires":        "60",
				"X-Amz-Signature":      "badsignature",
				"X-Amz-SignedHeaders":  "host;x-amz-content-sha256;x-amz-date",
				"X-Amz-Credential":     fmt.Sprintf(credentialTemplate, accessKeyID, now.Format(yyyymmdd)),
				"X-Amz-Content-Sha256": payloadSHA256,
			},
			headers: map[string]string{
				"X-Amz-Date":           now.AddDate(0, 0, -2).Format(iso8601Format),
				"X-Amz-Content-Sha256": payloadSHA256,
			},
			expected: ErrExpiredPresignRequest,
		},
		// (5) Should error if the signature is incorrect.
		{
			queryParams: map[string]string{
				"X-Amz-Algorithm":      signV4AAlgorithm,
				"X-Amz-Date":           now.Format(iso8601Format),
				"X-Amz-Expires":        "60",
				"X-Amz-Signature":      "badsignature",
				"X-Amz-SignedHeaders":  "host;x-amz-content-sha256;x-amz-date",
				"X-Amz-Credential":     fmt.Sprintf(credentialTemplate, accessKeyID, now.Format(yyyymmdd)),
				"X-Amz-Content-Sha256": payloadSHA256,
			},
			headers: map[string]string{
				"X-Amz-Date":           now.Format(iso8601Format),
				"X-Amz-Content-Sha256": payloadSHA256,
			},
			expected: ErrSignatureDoesNotMatch,
		},
		// (6) Should error if the request is not ready yet, ie X-Amz-Date is in the future.
		{
			queryParams: map[string]string{
				"X-Amz-Algorithm":      signV4AAlgorithm,
				"X-Amz-Date":           now.Add(1 * time.Hour).Format(iso8601Format),
				"X-Amz-Expires":        "60",
				"X-Amz-Signature":      "badsignature",
				"X-Amz-SignedHeaders":  "host;x-amz-content-sha256;x-amz-date",
				"X-Amz-Credential":     fmt.Sprintf(credentialTemplate, accessKeyID, now.Format(yyyymmdd), region),
				"X-Amz-Content-Sha256": payloadSHA256,
			},
			headers: map[string]string{
				"X-Amz-Date":           now.Format(iso8601Format),
				"X-Amz-Content-Sha256": payloadSHA256,
			},
			expected: ErrRequestNotReadyYet,
		},
		// (7) Should not error with invalid region instead, call should proceed
		// with sigature does not match.
		{
			queryParams: map[string]string{
				"X-Amz-Algorithm":      signV4AAlgorithm,
				"X-Amz-Date":           now.Format(iso8601Format),
				"X-Amz-Expires":        "60",
				"X-Amz-Signature":      "badsignature",
				"X-Amz-SignedHeaders":  "host;x-amz-content-sha256;x-amz-date",
				"X-Amz-Credential":     fmt.Sprintf(credentialTemplate, accessKeyID, now.Format(yyyymmdd), region),
				"X-Amz-Content-Sha256": payloadSHA256,
			},
			headers: map[string]string{
				"X-Amz-Date":           now.Format(iso8601Format),
				"X-Amz-Content-Sha256": payloadSHA256,
			},
			expected: ErrSignatureDoesNotMatch,
		},
		// (8) Should error with signature does not match. But handles
		// query params which do not precede with "x-amz-" header.
		{
			queryParams: map[string]string{
				"X-Amz-Algorithm":       signV4AAlgorithm,
				"X-Amz-Date":            now.Format(iso8601Format),
				"X-Amz-Expires":         "60",
				"X-Amz-Signature":       "badsignature",
				"X-Amz-SignedHeaders":   "host;x-amz-content-sha256;x-amz-date",
				"X-Amz-Credential":      fmt.Sprintf(credentialTemplate, accessKeyID, now.Format(yyyymmdd), region),
				"X-Amz-Content-Sha256":  payloadSHA256,
				"response-content-type": "application/json",
			},
			headers: map[string]string{
				"X-Amz-Date":           now.Format(iso8601Format),
				"X-Amz-Content-Sha256": payloadSHA256,
			},
			expected: ErrSignatureDoesNotMatch,
		},
		// (9) Should error with unsigned headers.
		{
			queryParams: map[string]string{
				"X-Amz-Algorithm":       signV4AAlgorithm,
				"X-Amz-Date":            now.Format(iso8601Format),
				"X-Amz-Expires":         "60",
				"X-Amz-Signature":       "badsignature",
				"X-Amz-SignedHeaders":   "host;x-amz-content-sha256;x-amz-date",
				"X-Amz-Credential":      fmt.Sprintf(credentialTemplate, accessKeyID, now.Format(yyyymmdd)),
				"X-Amz-Content-Sha256":  payloadSHA256,
				"response-content-type": "application/json",
			},
			headers: map[string]string{
				"X-Amz-Date": now.Format(iso8601Format),
			},
			expected: ErrUnsignedHeaders,
		},
	}

	// Run each test case individually.
	for i, testCase := range testCases {
		// Turn the map[string]string into map[string][]string, because Go.
		query := url.Values{}
		for key, value := range testCase.queryParams {
			query.Set(key, value)
		}

		// Create a request to use.
		req, e := http.NewRequest(http.MethodGet, "http://host/a/b?"+query.Encode(), nil)
		if e != nil {
			t.Errorf("(%d) failed to create http.Request, got %v", i, e)
		}

		// Do the same for the headers.
		for key, value := range testCase.headers {
			req.Header.Set(key, value)
		}

		// parse form.
		req.ParseForm()

		// Check if it matches!
		err := doesPresignedSignatureMatch(payloadSHA256, req, "", serviceS3, true)
		if err != testCase.expected {
			t.Errorf("(%d) expected to get %s, instead got %s", i, niceError(testCase.expected), niceError(err))
		}
	}
}

func BenchmarkVerifyV4ASignature(b *testing.B) {
	accessKey := "minioadmin"
	secretKey := "minioadmin"
	hash, _ := hex.DecodeString("936bed871590d4ddd183ed1e8f55bd9fd04c59ac83a1b29953d260f16de30f73")
	sign, _ := hex.DecodeString("3045022056fb6825eea50293f60e6ab56bd9f35f35b6d7570b6493860c625a1794ff9437022100e03491e568b38d95cf7106bbf68cc13f9dfa5c05b7e74747bda0c9047159781c")

	for i := 0; i < b.N; i++ {
		verifyV4ASignature(accessKey, secretKey, hash, sign)
	}
}

func BenchmarkVerifyECDSA(b *testing.B) {
	accessKey := "minioadmin"
	secretKey := "minioadmin"
	hash, _ := hex.DecodeString("936bed871590d4ddd183ed1e8f55bd9fd04c59ac83a1b29953d260f16de30f73")
	sign, _ := hex.DecodeString("3045022056fb6825eea50293f60e6ab56bd9f35f35b6d7570b6493860c625a1794ff9437022100e03491e568b38d95cf7106bbf68cc13f9dfa5c05b7e74747bda0c9047159781c")

	privKey, _ := deriveKeyFromAccessKeyPair(accessKey, secretKey)

	for i := 0; i < b.N; i++ {
		verifyECDSA(&privKey.PublicKey, hash, sign)
	}
}
