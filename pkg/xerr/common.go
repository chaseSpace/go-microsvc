package xerr

import "microsvc/pkg/i18n/langimpl"

// 大部分遵循 HTTP 状态码语义

// 200-400 series （前端错误）
var (
	ErrNil              = XErr{Code: langimpl.ErrorCodeNil.Int32(), Msg: "OK"}
	ErrParams           = XErr{Code: langimpl.ErrorCodeParams.Int32(), Msg: "ErrParams"}
	ErrUnauthorized     = XErr{Code: langimpl.ErrorCodeUnauthorized.Int32(), Msg: "ErrUnauthorized"}
	ErrForbidden        = XErr{Code: langimpl.ErrorCodeForbidden.Int32(), Msg: "ErrForbidden"}
	ErrNotFound         = XErr{Code: langimpl.ErrorCodeNotFound.Int32(), Msg: "ErrNotFound"}
	ErrMethodNotAllowed = XErr{Code: langimpl.ErrorCodeMethodNotAllowed.Int32(), Msg: "ErrMethodNotAllowed"}
	ErrReqTimeout       = XErr{Code: langimpl.ErrorCodeReqTimeout.Int32(), Msg: "ErrReqTimeout"}
	ErrTooManyRequests  = XErr{Code: langimpl.ErrorCodeTooManyRequests.Int32(), Msg: "ErrTooManyRequests"}
)

// 500 series (internal errors)
var (
	ErrInternal           = XErr{Code: langimpl.ErrorCodeInternal.Int32(), Msg: "ErrInternal"}
	ErrGateway            = XErr{Code: langimpl.ErrorCodeGateway.Int32(), Msg: "ErrGateway"}
	ErrServiceUnavailable = XErr{Code: langimpl.ErrorCodeServiceUnavailable.Int32(), Msg: "ErrServiceUnavailable"}
	ErrAPIUnavailable     = XErr{Code: langimpl.ErrorCodeAPIUnavailable.Int32(), Msg: "ErrAPIUnavailable"}
	ErrMySQL              = XErr{Code: langimpl.ErrorCodeMySQL.Int32(), Msg: "ErrMySQL"}
	ErrRedis              = XErr{Code: langimpl.ErrorCodeRedis.Int32(), Msg: "ErrRedis"}
	ErrUnknown            = XErr{Code: langimpl.ErrorCodeUnknown.Int32(), Msg: "ErrUnknown"}
)

// Customized errors (business errors)
var (
	ErrBizTimeout                = XErr{Code: langimpl.ErrorCodeBizTimeout.Int32(), Msg: "ErrBizTimeout"}
	ErrThirdParty                = XErr{Code: langimpl.ErrorCodeThirdParty.Int32(), Msg: "ErrThirdParty"}
	ErrInvalidRegisterInfo       = XErr{Code: langimpl.ErrorCodeInvalidRegisterInfo.Int32(), Msg: "ErrInvalidRegisterInfo"}
	ErrUserNotFound              = XErr{Code: langimpl.ErrorCodeUserNotFound.Int32(), Msg: "ErrUserNotFound"}
	ErrRepeatedOperation         = XErr{Code: langimpl.ErrorCodeRepeatedOperation.Int32(), Msg: "ErrRepeatedOperation"}
	ErrJSONMarshal               = XErr{Code: langimpl.ErrorCodeJSONMarshal.Int32(), Msg: "ErrJSONMarshal"}
	ErrJSONUnmarshal             = XErr{Code: langimpl.ErrorCodeJSONUnmarshal.Int32(), Msg: "ErrJSONUnmarshal"}
	ErrIDShouldBeZeroOnAdd       = XErr{Code: langimpl.ErrorCodeIDShouldBeZeroOnAdd.Int32(), Msg: "ErrIDShouldBeZeroOnAdd"}
	ErrIDShouldNotBeZeroOnUpdate = XErr{Code: langimpl.ErrorCodeIDShouldNotBeZeroOnUpdate.Int32(), Msg: "ErrIDShouldNotBeZeroOnUpdate"}
	ErrNoRowAffectedOnUpdate     = XErr{Code: langimpl.ErrorCodeNoRowAffectedOnUpdate.Int32(), Msg: "ErrNoRowAffectedOnUpdate"}
	ErrInvalidID                 = XErr{Code: langimpl.ErrorCodeInvalidID.Int32(), Msg: "ErrInvalidID"}
	ErrDataNotExist              = XErr{Code: langimpl.ErrorCodeDataNotExist.Int32(), Msg: "ErrDataNotExist"}
	ErrBaseReq                   = XErr{Code: langimpl.ErrorCodeBaseReq.Int32(), Msg: "ErrBaseReq"}
	ErrReviewResultReject        = XErr{Code: langimpl.ErrorCodeReviewResultReject.Int32(), Msg: "ErrReviewResultReject"}
	ErrDataDuplicate             = XErr{Code: langimpl.ErrorCodeDataDuplicate.Int32(), Msg: "ErrDataDuplicate"}
	ErrHTTPRequest               = XErr{Code: langimpl.ErrorCodeHTTPRequest.Int32(), Msg: "ErrHTTPRequest"}
	ErrRPC                       = XErr{Code: langimpl.ErrorCodeRPC.Int32(), Msg: "ErrRPC"} // 子服务之间的调用失败
	ErrCurrencyNotSupported      = XErr{Code: langimpl.ErrorCodeCurrencyNotSupported.Int32(), Msg: "ErrCurrencyNotSupported"}
)
