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
	_ = x[authTypePresignedV4A-3]
	_ = x[authTypePresignedV2-4]
	_ = x[authTypePostPolicy-5]
	_ = x[authTypeStreamingSigned-6]
	_ = x[authTypeSigned-7]
	_ = x[authTypeSignedV4A-8]
	_ = x[authTypeSignedV2-9]
	_ = x[authTypeJWT-10]
	_ = x[authTypeSTS-11]
}

const _authType_name = "UnknownAnonymousPresignedPresignedV4APresignedV2PostPolicyStreamingSignedSignedSignedV4ASignedV2JWTSTS"

var _authType_index = [...]uint8{0, 7, 16, 25, 37, 48, 58, 73, 79, 88, 96, 99, 102}

func (i authType) String() string {
	if i < 0 || i >= authType(len(_authType_index)-1) {
		return "authType(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _authType_name[_authType_index[i]:_authType_index[i+1]]
}
