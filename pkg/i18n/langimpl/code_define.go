package langimpl

// 200-400 series （前端错误）
const (
	ErrorCodeNil              Code = 200
	ErrorCodeParams           Code = 400
	ErrorCodeUnauthorized     Code = 401
	ErrorCodeForbidden        Code = 403
	ErrorCodeNotFound         Code = 404
	ErrorCodeMethodNotAllowed Code = 405
	ErrorCodeReqTimeout       Code = 408
	ErrorCodeTooManyRequests  Code = 409
)

// 500 series (内部错误)
const (
	ErrorCodeInternal           Code = 500
	ErrorCodeGateway            Code = 501
	ErrorCodeServiceUnavailable Code = 502
	ErrorCodeAPIUnavailable     Code = 510
	ErrorCodeMySQL              Code = 511
	ErrorCodeRedis              Code = 512
	ErrorCodeUnknown            Code = 520
)

// 自定义 1000+ （业务错误）
const (
	ErrorCodeBizTimeout                Code = 1000
	ErrorCodeThirdParty                Code = 1001
	ErrorCodeInvalidRegisterInfo       Code = 1002
	ErrorCodeUserNotFound              Code = 1003
	ErrorCodeRepeatedOperation         Code = 1004
	ErrorCodeJSONMarshal               Code = 1005
	ErrorCodeJSONUnmarshal             Code = 1006
	ErrorCodeIDShouldBeZeroOnAdd       Code = 1007
	ErrorCodeIDShouldNotBeZeroOnUpdate Code = 1008
	ErrorCodeNoRowAffectedOnUpdate     Code = 1009
	ErrorCodeInvalidID                 Code = 1010
	ErrorCodeDataNotExist              Code = 1011
	ErrorCodeBaseReq                   Code = 1012
	ErrorCodeReviewResultReject        Code = 1013
	ErrorCodeDataDuplicate             Code = 1014
	ErrorCodeHTTPRequest               Code = 1015
	ErrorCodeRPC                       Code = 1016
	ErrorCodeCurrencyNotSupported      Code = 1017
)

// 自定义SubId：Currency svc
const (
	SubIdTxFromAndToCannotBeSame SubId = "TxFromAndToCannotBeSame"
	SubIdTxAmountShouldNotBeZero SubId = "TxAmountShouldNotBeZero"
	SubIdTxRemarkTooLong         SubId = "TxRemarkTooLong"
	SubIdTxBalanceNotEnough      SubId = "TxBalanceNotEnough"
	SubIdTxInvalidTxType         SubId = "TxInvalidTxType"
	SubIdTxInvalidSingleTxType   SubId = "TxInvalidSingleTxType"
)

// 自定义SubId：Friend svc
const (
	SubIdFriendAlreadyFollow    SubId = "ErrFriendAlreadyFollow"
	SubIdFriendCountUpToMax     SubId = "ErrFriendCountUpToMax"
	SubIdFriendPeerCountUpToMax SubId = "ErrFriendPeerCountUpToMax"
)

// 自定义SubId：Gift svc
const (
	SubIdGiftTxFromAndToCannotBeSame      SubId = "GiftTxFromAndToCannotBeSame"
	SubIdGiftTxAmountMustBePositive       SubId = "GiftTxAmountMustBePositive"
	SubIdGiftTxAmountMustBePositiveOnSent SubId = "GiftTxAmountMustBePositiveOnSent"
	SubIdGiftTxInvalidTxType              SubId = "GiftTxInvalidTxType"
	SubIdGiftTxInvalidTxScene             SubId = "GiftTxInvalidTxScene"
	SubIdGiftTxInvalidFirstPersonTxType   SubId = "GiftTxInvalidFirstPersonTxType"
	SubIdGiftTxTypeConvertFailed          SubId = "GiftTxTypeConvertFailed"
	SubIdGiftTxBalanceNotEnough           SubId = "GiftTxBalanceNotEnough"
	SubIdGiftNotFound                     SubId = "GiftNotFound"
)

// 自定义SubId：Moment svc
const (
	SubIdMomentTextTooLong  SubId = "MomentTextTooLong"
	SubIdMomentTypeNotFound SubId = "MomentTypeNotFound"
	SubIdMomentNotFound     SubId = "MomentNotFound"
	SubIdCommentNotFound    SubId = "CommentNotFound"
	SubIdCommentTextTooLong SubId = "CommentTextTooLong"
)

// 自定义SubId ：Thirdparty svc
const (
	SubIdInvalidFileType               SubId = "InvalidFileType"
	SubIdThirdPartyServiceNameNotMatch SubId = "ThirdPartyServiceNameNotMatch"
	SubIdEmailCodeNeedSendFirst        SubId = "EmailCodeNeedSendFirst"
	SubIdWriteFileFailed               SubId = "WriteFileFailed"
)

// 自定义SubId ：User svc
const (
	SubIdPasswdFormat           SubId = "PasswdFormat"
	SubIdPasswdNoChange         SubId = "PasswdNoChange"
	SubIdInvalidPhoneNo         SubId = "InvalidPhoneNo"
	SubIdInvalidVerifyCode      SubId = "InvalidVerifyCode"
	SubIdInvalidLenPhoneNo      SubId = "InvalidLenPhoneNo"
	SubIdNotSupportedPhoneArea  SubId = "NotSupportedPhoneArea"
	SubIdChangePasswdFrequently SubId = "ChangePasswdFrequently"
	SubIdUnRegisteredPhone      SubId = "UnRegisteredPhone"
	SubIdUnSupportedSignInType  SubId = "UnSupportedSignInType"
	SubIdIncorrectPassword      SubId = "IncorrectPassword"
	SubIdSignInBanned           SubId = "SignInBanned"
	SubIdAccountBanned          SubId = "AccountBanned"
	SubIdSignInFailed           SubId = "SignInFailed"
	SubIdAccountAlreadyExists   SubId = "AccountAlreadyExists"
	SubIdLoginFrequently        SubId = "LoginFrequently"
	SubIdLoginFrequentlyLong    SubId = "LoginFrequentlyLong"
	SubIdPasswordNotSetOnLogin  SubId = "PasswordNotSetOnLogin"
	SubIdPasswordNotGiven       SubId = "PasswordNotGiven"
)

// 自定义SubId：Barbase svc

const (
	SubIdOrderStatusCannotBeCancelled SubId = "OrderStatusCannotBeCancelled"
)
