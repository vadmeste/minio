// Code generated by "stringer -type=authType -trimprefix=authType auth-handler.go"; DO NOT EDIT.

package cmd

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[authTypeUnknown-0]
	_ = x[authTypeAnonymous-1]
	_ = x[authTypePresigned-2]
	_ = x[authTypePresignedV2-3]
	_ = x[authTypePostPolicy-4]
	_ = x[authTypeStreamingSigned-5]
	_ = x[authTypeSigned-6]
	_ = x[authTypeSignedV2-7]
	_ = x[authTypeJWT-8]
	_ = x[authTypeSTS-9]
}

const _authType_name = "UnknownAnonymousPresignedPresignedV2PostPolicyStreamingSignedSignedSignedV2JWTSTS"

var _authType_index = [...]uint8{0, 7, 16, 25, 36, 46, 61, 67, 75, 78, 81}

func (i authType) String() string {
	if i < 0 || i >= authType(len(_authType_index)-1) {
		return "authType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _authType_name[_authType_index[i]:_authType_index[i+1]]
}
