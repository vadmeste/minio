// Code generated by "stringer -type=APIErrorCode -trimprefix=Err api-errors.go"; DO NOT EDIT.

package cmd

import "strconv"

func _() {
	// An "invalid array index" compiler error signifies that the constant values have changed.
	// Re-run the stringer command to generate them again.
	var x [1]struct{}
	_ = x[ErrNone-0]
	_ = x[ErrAccessDenied-1]
	_ = x[ErrBadDigest-2]
	_ = x[ErrEntityTooSmall-3]
	_ = x[ErrEntityTooLarge-4]
	_ = x[ErrPolicyTooLarge-5]
	_ = x[ErrIncompleteBody-6]
	_ = x[ErrInternalError-7]
	_ = x[ErrInvalidAccessKeyID-8]
	_ = x[ErrAccessKeyDisabled-9]
	_ = x[ErrInvalidBucketName-10]
	_ = x[ErrInvalidDigest-11]
	_ = x[ErrInvalidRange-12]
	_ = x[ErrInvalidRangePartNumber-13]
	_ = x[ErrInvalidCopyPartRange-14]
	_ = x[ErrInvalidCopyPartRangeSource-15]
	_ = x[ErrInvalidMaxKeys-16]
	_ = x[ErrInvalidEncodingMethod-17]
	_ = x[ErrInvalidMaxUploads-18]
	_ = x[ErrInvalidMaxParts-19]
	_ = x[ErrInvalidPartNumberMarker-20]
	_ = x[ErrInvalidPartNumber-21]
	_ = x[ErrInvalidRequestBody-22]
	_ = x[ErrInvalidCopySource-23]
	_ = x[ErrInvalidMetadataDirective-24]
	_ = x[ErrInvalidCopyDest-25]
	_ = x[ErrInvalidPolicyDocument-26]
	_ = x[ErrInvalidObjectState-27]
	_ = x[ErrMalformedXML-28]
	_ = x[ErrMissingContentLength-29]
	_ = x[ErrMissingContentMD5-30]
	_ = x[ErrMissingRequestBodyError-31]
	_ = x[ErrMissingSecurityHeader-32]
	_ = x[ErrNoSuchBucket-33]
	_ = x[ErrNoSuchBucketPolicy-34]
	_ = x[ErrNoSuchBucketLifecycle-35]
	_ = x[ErrNoSuchLifecycleConfiguration-36]
	_ = x[ErrInvalidLifecycleWithObjectLock-37]
	_ = x[ErrNoSuchBucketSSEConfig-38]
	_ = x[ErrNoSuchCORSConfiguration-39]
	_ = x[ErrNoSuchWebsiteConfiguration-40]
	_ = x[ErrReplicationConfigurationNotFoundError-41]
	_ = x[ErrRemoteDestinationNotFoundError-42]
	_ = x[ErrReplicationDestinationMissingLock-43]
	_ = x[ErrRemoteTargetNotFoundError-44]
	_ = x[ErrReplicationRemoteConnectionError-45]
	_ = x[ErrReplicationBandwidthLimitError-46]
	_ = x[ErrBucketRemoteIdenticalToSource-47]
	_ = x[ErrBucketRemoteAlreadyExists-48]
	_ = x[ErrBucketRemoteLabelInUse-49]
	_ = x[ErrBucketRemoteArnTypeInvalid-50]
	_ = x[ErrBucketRemoteArnInvalid-51]
	_ = x[ErrBucketRemoteRemoveDisallowed-52]
	_ = x[ErrRemoteTargetNotVersionedError-53]
	_ = x[ErrReplicationSourceNotVersionedError-54]
	_ = x[ErrReplicationNeedsVersioningError-55]
	_ = x[ErrReplicationBucketNeedsVersioningError-56]
	_ = x[ErrReplicationDenyEditError-57]
	_ = x[ErrReplicationNoMatchingRuleError-58]
	_ = x[ErrObjectRestoreAlreadyInProgress-59]
	_ = x[ErrNoSuchKey-60]
	_ = x[ErrNoSuchUpload-61]
	_ = x[ErrInvalidVersionID-62]
	_ = x[ErrNoSuchVersion-63]
	_ = x[ErrNotImplemented-64]
	_ = x[ErrPreconditionFailed-65]
	_ = x[ErrRequestTimeTooSkewed-66]
	_ = x[ErrSignatureDoesNotMatch-67]
	_ = x[ErrMethodNotAllowed-68]
	_ = x[ErrInvalidPart-69]
	_ = x[ErrInvalidPartOrder-70]
	_ = x[ErrAuthorizationHeaderMalformed-71]
	_ = x[ErrMalformedPOSTRequest-72]
	_ = x[ErrPOSTFileRequired-73]
	_ = x[ErrSignatureVersionNotSupported-74]
	_ = x[ErrBucketNotEmpty-75]
	_ = x[ErrAllAccessDisabled-76]
	_ = x[ErrMalformedPolicy-77]
	_ = x[ErrMissingFields-78]
	_ = x[ErrMissingCredTag-79]
	_ = x[ErrCredMalformed-80]
	_ = x[ErrInvalidRegion-81]
	_ = x[ErrInvalidServiceS3-82]
	_ = x[ErrInvalidServiceSTS-83]
	_ = x[ErrInvalidRequestVersion-84]
	_ = x[ErrMissingSignTag-85]
	_ = x[ErrMissingSignHeadersTag-86]
	_ = x[ErrMalformedDate-87]
	_ = x[ErrMalformedPresignedDate-88]
	_ = x[ErrMalformedCredentialDate-89]
	_ = x[ErrMalformedCredentialRegion-90]
	_ = x[ErrMalformedExpires-91]
	_ = x[ErrNegativeExpires-92]
	_ = x[ErrAuthHeaderEmpty-93]
	_ = x[ErrExpiredPresignRequest-94]
	_ = x[ErrRequestNotReadyYet-95]
	_ = x[ErrUnsignedHeaders-96]
	_ = x[ErrMissingDateHeader-97]
	_ = x[ErrInvalidQuerySignatureAlgo-98]
	_ = x[ErrInvalidQueryParams-99]
	_ = x[ErrBucketAlreadyOwnedByYou-100]
	_ = x[ErrInvalidDuration-101]
	_ = x[ErrBucketAlreadyExists-102]
	_ = x[ErrMetadataTooLarge-103]
	_ = x[ErrUnsupportedMetadata-104]
	_ = x[ErrMaximumExpires-105]
	_ = x[ErrSlowDown-106]
	_ = x[ErrInvalidPrefixMarker-107]
	_ = x[ErrBadRequest-108]
	_ = x[ErrKeyTooLongError-109]
	_ = x[ErrInvalidBucketObjectLockConfiguration-110]
	_ = x[ErrObjectLockConfigurationNotFound-111]
	_ = x[ErrObjectLockConfigurationNotAllowed-112]
	_ = x[ErrNoSuchObjectLockConfiguration-113]
	_ = x[ErrObjectLocked-114]
	_ = x[ErrInvalidRetentionDate-115]
	_ = x[ErrPastObjectLockRetainDate-116]
	_ = x[ErrUnknownWORMModeDirective-117]
	_ = x[ErrBucketTaggingNotFound-118]
	_ = x[ErrObjectLockInvalidHeaders-119]
	_ = x[ErrInvalidTagDirective-120]
	_ = x[ErrInvalidEncryptionMethod-121]
	_ = x[ErrInvalidEncryptionKeyID-122]
	_ = x[ErrInsecureSSECustomerRequest-123]
	_ = x[ErrSSEMultipartEncrypted-124]
	_ = x[ErrSSEEncryptedObject-125]
	_ = x[ErrInvalidEncryptionParameters-126]
	_ = x[ErrInvalidSSECustomerAlgorithm-127]
	_ = x[ErrInvalidSSECustomerKey-128]
	_ = x[ErrMissingSSECustomerKey-129]
	_ = x[ErrMissingSSECustomerKeyMD5-130]
	_ = x[ErrSSECustomerKeyMD5Mismatch-131]
	_ = x[ErrInvalidSSECustomerParameters-132]
	_ = x[ErrIncompatibleEncryptionMethod-133]
	_ = x[ErrKMSNotConfigured-134]
	_ = x[ErrKMSKeyNotFoundException-135]
	_ = x[ErrNoAccessKey-136]
	_ = x[ErrInvalidToken-137]
	_ = x[ErrEventNotification-138]
	_ = x[ErrARNNotification-139]
	_ = x[ErrRegionNotification-140]
	_ = x[ErrOverlappingFilterNotification-141]
	_ = x[ErrFilterNameInvalid-142]
	_ = x[ErrFilterNamePrefix-143]
	_ = x[ErrFilterNameSuffix-144]
	_ = x[ErrFilterValueInvalid-145]
	_ = x[ErrOverlappingConfigs-146]
	_ = x[ErrUnsupportedNotification-147]
	_ = x[ErrContentSHA256Mismatch-148]
	_ = x[ErrReadQuorum-149]
	_ = x[ErrWriteQuorum-150]
	_ = x[ErrStorageFull-151]
	_ = x[ErrRequestBodyParse-152]
	_ = x[ErrObjectExistsAsDirectory-153]
	_ = x[ErrInvalidObjectName-154]
	_ = x[ErrInvalidObjectNamePrefixSlash-155]
	_ = x[ErrInvalidResourceName-156]
	_ = x[ErrServerNotInitialized-157]
	_ = x[ErrOperationTimedOut-158]
	_ = x[ErrClientDisconnected-159]
	_ = x[ErrOperationMaxedOut-160]
	_ = x[ErrInvalidRequest-161]
	_ = x[ErrTransitionStorageClassNotFoundError-162]
	_ = x[ErrInvalidStorageClass-163]
	_ = x[ErrBackendDown-164]
	_ = x[ErrMalformedJSON-165]
	_ = x[ErrAdminNoSuchUser-166]
	_ = x[ErrAdminNoSuchGroup-167]
	_ = x[ErrAdminGroupNotEmpty-168]
	_ = x[ErrAdminNoSuchPolicy-169]
	_ = x[ErrAdminInvalidArgument-170]
	_ = x[ErrAdminInvalidAccessKey-171]
	_ = x[ErrAdminInvalidSecretKey-172]
	_ = x[ErrAdminConfigNoQuorum-173]
	_ = x[ErrAdminConfigTooLarge-174]
	_ = x[ErrAdminConfigBadJSON-175]
	_ = x[ErrAdminConfigDuplicateKeys-176]
	_ = x[ErrAdminCredentialsMismatch-177]
	_ = x[ErrInsecureClientRequest-178]
	_ = x[ErrObjectTampered-179]
	_ = x[ErrSiteReplicationInvalidRequest-180]
	_ = x[ErrSiteReplicationPeerResp-181]
	_ = x[ErrSiteReplicationBackendIssue-182]
	_ = x[ErrSiteReplicationServiceAccountError-183]
	_ = x[ErrSiteReplicationBucketConfigError-184]
	_ = x[ErrSiteReplicationBucketMetaError-185]
	_ = x[ErrSiteReplicationIAMError-186]
	_ = x[ErrAdminBucketQuotaExceeded-187]
	_ = x[ErrAdminNoSuchQuotaConfiguration-188]
	_ = x[ErrHealNotImplemented-189]
	_ = x[ErrHealNoSuchProcess-190]
	_ = x[ErrHealInvalidClientToken-191]
	_ = x[ErrHealMissingBucket-192]
	_ = x[ErrHealAlreadyRunning-193]
	_ = x[ErrHealOverlappingPaths-194]
	_ = x[ErrIncorrectContinuationToken-195]
	_ = x[ErrEmptyRequestBody-196]
	_ = x[ErrUnsupportedFunction-197]
	_ = x[ErrInvalidExpressionType-198]
	_ = x[ErrBusy-199]
	_ = x[ErrUnauthorizedAccess-200]
	_ = x[ErrExpressionTooLong-201]
	_ = x[ErrIllegalSQLFunctionArgument-202]
	_ = x[ErrInvalidKeyPath-203]
	_ = x[ErrInvalidCompressionFormat-204]
	_ = x[ErrInvalidFileHeaderInfo-205]
	_ = x[ErrInvalidJSONType-206]
	_ = x[ErrInvalidQuoteFields-207]
	_ = x[ErrInvalidRequestParameter-208]
	_ = x[ErrInvalidDataType-209]
	_ = x[ErrInvalidTextEncoding-210]
	_ = x[ErrInvalidDataSource-211]
	_ = x[ErrInvalidTableAlias-212]
	_ = x[ErrMissingRequiredParameter-213]
	_ = x[ErrObjectSerializationConflict-214]
	_ = x[ErrUnsupportedSQLOperation-215]
	_ = x[ErrUnsupportedSQLStructure-216]
	_ = x[ErrUnsupportedSyntax-217]
	_ = x[ErrUnsupportedRangeHeader-218]
	_ = x[ErrLexerInvalidChar-219]
	_ = x[ErrLexerInvalidOperator-220]
	_ = x[ErrLexerInvalidLiteral-221]
	_ = x[ErrLexerInvalidIONLiteral-222]
	_ = x[ErrParseExpectedDatePart-223]
	_ = x[ErrParseExpectedKeyword-224]
	_ = x[ErrParseExpectedTokenType-225]
	_ = x[ErrParseExpected2TokenTypes-226]
	_ = x[ErrParseExpectedNumber-227]
	_ = x[ErrParseExpectedRightParenBuiltinFunctionCall-228]
	_ = x[ErrParseExpectedTypeName-229]
	_ = x[ErrParseExpectedWhenClause-230]
	_ = x[ErrParseUnsupportedToken-231]
	_ = x[ErrParseUnsupportedLiteralsGroupBy-232]
	_ = x[ErrParseExpectedMember-233]
	_ = x[ErrParseUnsupportedSelect-234]
	_ = x[ErrParseUnsupportedCase-235]
	_ = x[ErrParseUnsupportedCaseClause-236]
	_ = x[ErrParseUnsupportedAlias-237]
	_ = x[ErrParseUnsupportedSyntax-238]
	_ = x[ErrParseUnknownOperator-239]
	_ = x[ErrParseMissingIdentAfterAt-240]
	_ = x[ErrParseUnexpectedOperator-241]
	_ = x[ErrParseUnexpectedTerm-242]
	_ = x[ErrParseUnexpectedToken-243]
	_ = x[ErrParseUnexpectedKeyword-244]
	_ = x[ErrParseExpectedExpression-245]
	_ = x[ErrParseExpectedLeftParenAfterCast-246]
	_ = x[ErrParseExpectedLeftParenValueConstructor-247]
	_ = x[ErrParseExpectedLeftParenBuiltinFunctionCall-248]
	_ = x[ErrParseExpectedArgumentDelimiter-249]
	_ = x[ErrParseCastArity-250]
	_ = x[ErrParseInvalidTypeParam-251]
	_ = x[ErrParseEmptySelect-252]
	_ = x[ErrParseSelectMissingFrom-253]
	_ = x[ErrParseExpectedIdentForGroupName-254]
	_ = x[ErrParseExpectedIdentForAlias-255]
	_ = x[ErrParseUnsupportedCallWithStar-256]
	_ = x[ErrParseNonUnaryAgregateFunctionCall-257]
	_ = x[ErrParseMalformedJoin-258]
	_ = x[ErrParseExpectedIdentForAt-259]
	_ = x[ErrParseAsteriskIsNotAloneInSelectList-260]
	_ = x[ErrParseCannotMixSqbAndWildcardInSelectList-261]
	_ = x[ErrParseInvalidContextForWildcardInSelectList-262]
	_ = x[ErrIncorrectSQLFunctionArgumentType-263]
	_ = x[ErrValueParseFailure-264]
	_ = x[ErrEvaluatorInvalidArguments-265]
	_ = x[ErrIntegerOverflow-266]
	_ = x[ErrLikeInvalidInputs-267]
	_ = x[ErrCastFailed-268]
	_ = x[ErrInvalidCast-269]
	_ = x[ErrEvaluatorInvalidTimestampFormatPattern-270]
	_ = x[ErrEvaluatorInvalidTimestampFormatPatternSymbolForParsing-271]
	_ = x[ErrEvaluatorTimestampFormatPatternDuplicateFields-272]
	_ = x[ErrEvaluatorTimestampFormatPatternHourClockAmPmMismatch-273]
	_ = x[ErrEvaluatorUnterminatedTimestampFormatPatternToken-274]
	_ = x[ErrEvaluatorInvalidTimestampFormatPatternToken-275]
	_ = x[ErrEvaluatorInvalidTimestampFormatPatternSymbol-276]
	_ = x[ErrEvaluatorBindingDoesNotExist-277]
	_ = x[ErrMissingHeaders-278]
	_ = x[ErrInvalidColumnIndex-279]
	_ = x[ErrAdminConfigNotificationTargetsFailed-280]
	_ = x[ErrAdminProfilerNotEnabled-281]
	_ = x[ErrInvalidDecompressedSize-282]
	_ = x[ErrAddUserInvalidArgument-283]
	_ = x[ErrAdminAccountNotEligible-284]
	_ = x[ErrAccountNotEligible-285]
	_ = x[ErrAdminServiceAccountNotFound-286]
	_ = x[ErrPostPolicyConditionInvalidFormat-287]
}

const _APIErrorCode_name = "NoneAccessDeniedBadDigestEntityTooSmallEntityTooLargePolicyTooLargeIncompleteBodyInternalErrorInvalidAccessKeyIDAccessKeyDisabledInvalidBucketNameInvalidDigestInvalidRangeInvalidRangePartNumberInvalidCopyPartRangeInvalidCopyPartRangeSourceInvalidMaxKeysInvalidEncodingMethodInvalidMaxUploadsInvalidMaxPartsInvalidPartNumberMarkerInvalidPartNumberInvalidRequestBodyInvalidCopySourceInvalidMetadataDirectiveInvalidCopyDestInvalidPolicyDocumentInvalidObjectStateMalformedXMLMissingContentLengthMissingContentMD5MissingRequestBodyErrorMissingSecurityHeaderNoSuchBucketNoSuchBucketPolicyNoSuchBucketLifecycleNoSuchLifecycleConfigurationInvalidLifecycleWithObjectLockNoSuchBucketSSEConfigNoSuchCORSConfigurationNoSuchWebsiteConfigurationReplicationConfigurationNotFoundErrorRemoteDestinationNotFoundErrorReplicationDestinationMissingLockRemoteTargetNotFoundErrorReplicationRemoteConnectionErrorReplicationBandwidthLimitErrorBucketRemoteIdenticalToSourceBucketRemoteAlreadyExistsBucketRemoteLabelInUseBucketRemoteArnTypeInvalidBucketRemoteArnInvalidBucketRemoteRemoveDisallowedRemoteTargetNotVersionedErrorReplicationSourceNotVersionedErrorReplicationNeedsVersioningErrorReplicationBucketNeedsVersioningErrorReplicationDenyEditErrorReplicationNoMatchingRuleErrorObjectRestoreAlreadyInProgressNoSuchKeyNoSuchUploadInvalidVersionIDNoSuchVersionNotImplementedPreconditionFailedRequestTimeTooSkewedSignatureDoesNotMatchMethodNotAllowedInvalidPartInvalidPartOrderAuthorizationHeaderMalformedMalformedPOSTRequestPOSTFileRequiredSignatureVersionNotSupportedBucketNotEmptyAllAccessDisabledMalformedPolicyMissingFieldsMissingCredTagCredMalformedInvalidRegionInvalidServiceS3InvalidServiceSTSInvalidRequestVersionMissingSignTagMissingSignHeadersTagMalformedDateMalformedPresignedDateMalformedCredentialDateMalformedCredentialRegionMalformedExpiresNegativeExpiresAuthHeaderEmptyExpiredPresignRequestRequestNotReadyYetUnsignedHeadersMissingDateHeaderInvalidQuerySignatureAlgoInvalidQueryParamsBucketAlreadyOwnedByYouInvalidDurationBucketAlreadyExistsMetadataTooLargeUnsupportedMetadataMaximumExpiresSlowDownInvalidPrefixMarkerBadRequestKeyTooLongErrorInvalidBucketObjectLockConfigurationObjectLockConfigurationNotFoundObjectLockConfigurationNotAllowedNoSuchObjectLockConfigurationObjectLockedInvalidRetentionDatePastObjectLockRetainDateUnknownWORMModeDirectiveBucketTaggingNotFoundObjectLockInvalidHeadersInvalidTagDirectiveInvalidEncryptionMethodInvalidEncryptionKeyIDInsecureSSECustomerRequestSSEMultipartEncryptedSSEEncryptedObjectInvalidEncryptionParametersInvalidSSECustomerAlgorithmInvalidSSECustomerKeyMissingSSECustomerKeyMissingSSECustomerKeyMD5SSECustomerKeyMD5MismatchInvalidSSECustomerParametersIncompatibleEncryptionMethodKMSNotConfiguredKMSKeyNotFoundExceptionNoAccessKeyInvalidTokenEventNotificationARNNotificationRegionNotificationOverlappingFilterNotificationFilterNameInvalidFilterNamePrefixFilterNameSuffixFilterValueInvalidOverlappingConfigsUnsupportedNotificationContentSHA256MismatchReadQuorumWriteQuorumStorageFullRequestBodyParseObjectExistsAsDirectoryInvalidObjectNameInvalidObjectNamePrefixSlashInvalidResourceNameServerNotInitializedOperationTimedOutClientDisconnectedOperationMaxedOutInvalidRequestTransitionStorageClassNotFoundErrorInvalidStorageClassBackendDownMalformedJSONAdminNoSuchUserAdminNoSuchGroupAdminGroupNotEmptyAdminNoSuchPolicyAdminInvalidArgumentAdminInvalidAccessKeyAdminInvalidSecretKeyAdminConfigNoQuorumAdminConfigTooLargeAdminConfigBadJSONAdminConfigDuplicateKeysAdminCredentialsMismatchInsecureClientRequestObjectTamperedSiteReplicationInvalidRequestSiteReplicationPeerRespSiteReplicationBackendIssueSiteReplicationServiceAccountErrorSiteReplicationBucketConfigErrorSiteReplicationBucketMetaErrorSiteReplicationIAMErrorAdminBucketQuotaExceededAdminNoSuchQuotaConfigurationHealNotImplementedHealNoSuchProcessHealInvalidClientTokenHealMissingBucketHealAlreadyRunningHealOverlappingPathsIncorrectContinuationTokenEmptyRequestBodyUnsupportedFunctionInvalidExpressionTypeBusyUnauthorizedAccessExpressionTooLongIllegalSQLFunctionArgumentInvalidKeyPathInvalidCompressionFormatInvalidFileHeaderInfoInvalidJSONTypeInvalidQuoteFieldsInvalidRequestParameterInvalidDataTypeInvalidTextEncodingInvalidDataSourceInvalidTableAliasMissingRequiredParameterObjectSerializationConflictUnsupportedSQLOperationUnsupportedSQLStructureUnsupportedSyntaxUnsupportedRangeHeaderLexerInvalidCharLexerInvalidOperatorLexerInvalidLiteralLexerInvalidIONLiteralParseExpectedDatePartParseExpectedKeywordParseExpectedTokenTypeParseExpected2TokenTypesParseExpectedNumberParseExpectedRightParenBuiltinFunctionCallParseExpectedTypeNameParseExpectedWhenClauseParseUnsupportedTokenParseUnsupportedLiteralsGroupByParseExpectedMemberParseUnsupportedSelectParseUnsupportedCaseParseUnsupportedCaseClauseParseUnsupportedAliasParseUnsupportedSyntaxParseUnknownOperatorParseMissingIdentAfterAtParseUnexpectedOperatorParseUnexpectedTermParseUnexpectedTokenParseUnexpectedKeywordParseExpectedExpressionParseExpectedLeftParenAfterCastParseExpectedLeftParenValueConstructorParseExpectedLeftParenBuiltinFunctionCallParseExpectedArgumentDelimiterParseCastArityParseInvalidTypeParamParseEmptySelectParseSelectMissingFromParseExpectedIdentForGroupNameParseExpectedIdentForAliasParseUnsupportedCallWithStarParseNonUnaryAgregateFunctionCallParseMalformedJoinParseExpectedIdentForAtParseAsteriskIsNotAloneInSelectListParseCannotMixSqbAndWildcardInSelectListParseInvalidContextForWildcardInSelectListIncorrectSQLFunctionArgumentTypeValueParseFailureEvaluatorInvalidArgumentsIntegerOverflowLikeInvalidInputsCastFailedInvalidCastEvaluatorInvalidTimestampFormatPatternEvaluatorInvalidTimestampFormatPatternSymbolForParsingEvaluatorTimestampFormatPatternDuplicateFieldsEvaluatorTimestampFormatPatternHourClockAmPmMismatchEvaluatorUnterminatedTimestampFormatPatternTokenEvaluatorInvalidTimestampFormatPatternTokenEvaluatorInvalidTimestampFormatPatternSymbolEvaluatorBindingDoesNotExistMissingHeadersInvalidColumnIndexAdminConfigNotificationTargetsFailedAdminProfilerNotEnabledInvalidDecompressedSizeAddUserInvalidArgumentAdminAccountNotEligibleAccountNotEligibleAdminServiceAccountNotFoundPostPolicyConditionInvalidFormat"

var _APIErrorCode_index = [...]uint16{0, 4, 16, 25, 39, 53, 67, 81, 94, 112, 129, 146, 159, 171, 193, 213, 239, 253, 274, 291, 306, 329, 346, 364, 381, 405, 420, 441, 459, 471, 491, 508, 531, 552, 564, 582, 603, 631, 661, 682, 705, 731, 768, 798, 831, 856, 888, 918, 947, 972, 994, 1020, 1042, 1070, 1099, 1133, 1164, 1201, 1225, 1255, 1285, 1294, 1306, 1322, 1335, 1349, 1367, 1387, 1408, 1424, 1435, 1451, 1479, 1499, 1515, 1543, 1557, 1574, 1589, 1602, 1616, 1629, 1642, 1658, 1675, 1696, 1710, 1731, 1744, 1766, 1789, 1814, 1830, 1845, 1860, 1881, 1899, 1914, 1931, 1956, 1974, 1997, 2012, 2031, 2047, 2066, 2080, 2088, 2107, 2117, 2132, 2168, 2199, 2232, 2261, 2273, 2293, 2317, 2341, 2362, 2386, 2405, 2428, 2450, 2476, 2497, 2515, 2542, 2569, 2590, 2611, 2635, 2660, 2688, 2716, 2732, 2755, 2766, 2778, 2795, 2810, 2828, 2857, 2874, 2890, 2906, 2924, 2942, 2965, 2986, 2996, 3007, 3018, 3034, 3057, 3074, 3102, 3121, 3141, 3158, 3176, 3193, 3207, 3242, 3261, 3272, 3285, 3300, 3316, 3334, 3351, 3371, 3392, 3413, 3432, 3451, 3469, 3493, 3517, 3538, 3552, 3581, 3604, 3631, 3665, 3697, 3727, 3750, 3774, 3803, 3821, 3838, 3860, 3877, 3895, 3915, 3941, 3957, 3976, 3997, 4001, 4019, 4036, 4062, 4076, 4100, 4121, 4136, 4154, 4177, 4192, 4211, 4228, 4245, 4269, 4296, 4319, 4342, 4359, 4381, 4397, 4417, 4436, 4458, 4479, 4499, 4521, 4545, 4564, 4606, 4627, 4650, 4671, 4702, 4721, 4743, 4763, 4789, 4810, 4832, 4852, 4876, 4899, 4918, 4938, 4960, 4983, 5014, 5052, 5093, 5123, 5137, 5158, 5174, 5196, 5226, 5252, 5280, 5313, 5331, 5354, 5389, 5429, 5471, 5503, 5520, 5545, 5560, 5577, 5587, 5598, 5636, 5690, 5736, 5788, 5836, 5879, 5923, 5951, 5965, 5983, 6019, 6042, 6065, 6087, 6110, 6128, 6155, 6187}

func (i APIErrorCode) String() string {
	if i < 0 || i >= APIErrorCode(len(_APIErrorCode_index)-1) {
		return "APIErrorCode(" + strconv.FormatInt(int64(i), 10) + ")"
	}
	return _APIErrorCode_name[_APIErrorCode_index[i]:_APIErrorCode_index[i+1]]
}
