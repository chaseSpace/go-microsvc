package cache

const CKeyPrefix = "SVC_THIRDPARTY:"

const (
	CKeySmsCode                       = CKeyPrefix + "sms_code:account_%v:scene_%v"
	CKeySmsCodeLimiterKeyAccountScene = CKeyPrefix + "sms_code_limiter:%v:scene_%v"
	CKeySmsCodeLimiterKeyIP           = CKeyPrefix + "sms_code_limiter:ip:%v"

	CKeyEmailCode = CKeyPrefix + "email_code:account_%v:scene_%v"
)
