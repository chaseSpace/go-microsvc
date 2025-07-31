package langimpl

import (
	"fmt"
	"microsvc/protocol/svc/commonpb"
)

type English struct {
	meta map[Scene]MsgMap
}

func (e *English) Lang() commonpb.Lang {
	return commonpb.Lang_CL_EN
}
func (e *English) Translate(scene Scene, code Code, subId SubId) string {
	if m1 := e.meta[scene]; m1 != nil {
		m2, ok := m1[code]
		if ok {
			if subId != "" {
				return m2[subId]
			}
			return m2["default"]
		}
	}
	return ""
}

var EN = &English{
	meta: map[Scene]MsgMap{
		SceneError: map[Code]map[SubId]string{},
	},
}

func init() {
	EN.registerSceneErrorForCommon()
	EN.registerSceneErrorForFriendSvc()
	EN.registerSceneErrorGiftSvc()
	EN.registerSceneErrorForCurrencySvc()
	EN.registerSceneErrorForMomentSvc()
	EN.registerSceneErrorForThirdPartySvc()
	EN.registerSceneErrorForUserSvc()
}

func (e *English) registerSceneErrorForCommon() {
	add := map[Code]string{
		ErrorCodeNil:              "OK",
		ErrorCodeParams:           "Invalid parameters provided",
		ErrorCodeUnauthorized:     "Authentication required",
		ErrorCodeForbidden:        "Access denied",
		ErrorCodeNotFound:         "Resource not found",
		ErrorCodeMethodNotAllowed: "Method not supported",
		ErrorCodeReqTimeout:       "Request timed out",
		ErrorCodeTooManyRequests:  "Too many requests, please try again later",

		ErrorCodeInternal:           "Internal server error",
		ErrorCodeGateway:            "Gateway error",
		ErrorCodeServiceUnavailable: "Target micro-service unavailable",
		ErrorCodeAPIUnavailable:     "API service unavailable",
		ErrorCodeMySQL:              "MySQL error",
		ErrorCodeRedis:              "Redis error",

		ErrorCodeBizTimeout:                "Business Timeout",
		ErrorCodeThirdParty:                "Exception From Third-party Service",
		ErrorCodeInvalidRegisterInfo:       "Invalid Registration Details",
		ErrorCodeUserNotFound:              "User Account Not Found",
		ErrorCodeRepeatedOperation:         "Operation Already Performed",
		ErrorCodeJSONMarshal:               "Failed to Marshal JSON",
		ErrorCodeJSONUnmarshal:             "Failed to Unmarshal JSON",
		ErrorCodeIDShouldBeZeroOnAdd:       "New Records Must Have Zero ID",
		ErrorCodeIDShouldNotBeZeroOnUpdate: "Update Requires a Non-Zero ID",
		ErrorCodeNoRowAffectedOnUpdate:     "No Changes Detected for Update",
		ErrorCodeInvalidID:                 "ID Provided Is Invalid",
		ErrorCodeDataNotExist:              "Requested Data Is Non-Existent",
		ErrorCodeBaseReq:                   "Error in Basic Request Parameters",
		ErrorCodeReviewResultReject:        "Content Review Denied",
		ErrorCodeDataDuplicate:             "Duplicate Data try Detected",
		ErrorCodeHTTPRequest:               "Error in HTTP Request",
		ErrorCodeRPC:                       "RPC failed",
		ErrorCodeCurrencyNotSupported:      "Currency Not Supported",
	}

	for k, v := range add {
		if _, ok := e.meta[SceneError][k]; ok {
			panic(fmt.Sprintf("duplicate key: (%d: %s) on register code", k, v))
		}
		e.meta[SceneError][k] = map[SubId]string{"default": v} // default msg of this code
	}
}

func (e *English) addSceneError(add MsgMap) {
	for k, newSubMap := range add {
		if oldSubMap, ok := e.meta[SceneError][k]; ok && len(newSubMap) == 0 {
			panic(fmt.Sprintf("duplicate key: (%d) on register code", k))
		} else if len(oldSubMap) > 0 {
			for k2, v2 := range newSubMap {
				if _, ok = oldSubMap[k2]; ok {
					panic(fmt.Sprintf("duplicate key: (code=%d: sub=%s) on register sub identity of code", k, k2))
				}
				oldSubMap[k2] = v2
			}
			continue
		}
		e.meta[SceneError][k] = newSubMap
	}
}

func (e *English) registerSceneErrorForCurrencySvc() {
	add := map[Code]map[SubId]string{
		ErrorCodeParams: {
			SubIdTxFromAndToCannotBeSame: "The transaction initiator and receiver cannot be the same",
			SubIdTxAmountShouldNotBeZero: "The transaction amount cannot be zero",
			SubIdTxRemarkTooLong:         "The transaction remark cannot exceed 100 characters",
			SubIdTxBalanceNotEnough:      "Insufficient balance",
			SubIdTxInvalidTxType:         "Invalid transaction type",
			SubIdTxInvalidSingleTxType:   "Invalid individual transaction type",
		},
	}
	e.addSceneError(add)
}

func (e *English) registerSceneErrorForFriendSvc() {
	add := map[Code]map[SubId]string{
		ErrorCodeParams: {
			SubIdFriendAlreadyFollow:    "Already followed",
			SubIdFriendCountUpToMax:     "Friend limit reached",
			SubIdFriendPeerCountUpToMax: "Peer friend limit reached",
		},
	}
	e.addSceneError(add)
}

func (e *English) registerSceneErrorGiftSvc() {
	add := map[Code]map[SubId]string{
		ErrorCodeParams: {
			SubIdGiftTxFromAndToCannotBeSame:      "Sender and receiver cannot be the same",
			SubIdGiftTxAmountMustBePositive:       "Gift quantity must be greater than 0",
			SubIdGiftTxAmountMustBePositiveOnSent: "Gift quantity on send must be greater than 0",
			SubIdGiftTxInvalidTxType:              "Invalid transaction type",
			SubIdGiftTxInvalidTxScene:             "Invalid gift scene",
			SubIdGiftTxInvalidFirstPersonTxType:   "Invalid first-person transaction type",
			SubIdGiftTxTypeConvertFailed:          "Transaction type conversion failed",
			SubIdGiftTxBalanceNotEnough:           "Insufficient gift balance",
			SubIdGiftNotFound:                     "Gift not found",
		},
	}
	e.addSceneError(add)
}

func (e *English) registerSceneErrorForMomentSvc() {
	add := map[Code]map[SubId]string{
		ErrorCodeParams: {
			SubIdMomentTextTooLong:  "Text cannot exceed 200 characters",
			SubIdMomentTypeNotFound: "Unsupported moment type",
			SubIdMomentNotFound:     "Moment does not exist",
			SubIdCommentNotFound:    "Comment does not exist",
			SubIdCommentTextTooLong: "Comment cannot exceed 100 characters",
		},
	}
	e.addSceneError(add)
}

func (e *English) registerSceneErrorForThirdPartySvc() {
	add := map[Code]map[SubId]string{
		ErrorCodeParams: {
			SubIdInvalidFileType:               "Invalid file type",
			SubIdThirdPartyServiceNameNotMatch: "Third party service name does not match",
			SubIdEmailCodeNeedSendFirst:        "Please send email code firstly",
			SubIdWriteFileFailed:               "Write file failed",
		},
	}
	e.addSceneError(add)
}

func (e *English) registerSceneErrorForUserSvc() {
	add := map[Code]map[SubId]string{
		ErrorCodeParams: {
			SubIdPasswdFormat:           "Please enter a valid password parameter",
			SubIdPasswdNoChange:         "The new password cannot be the same as the old",
			SubIdInvalidPhoneNo:         "Please provide a valid phone number",
			SubIdInvalidVerifyCode:      "Verification code is incorrect",
			SubIdInvalidLenPhoneNo:      "Incorrect phone number length",
			SubIdNotSupportedPhoneArea:  "Unsupported phone area code",
			SubIdChangePasswdFrequently: "Password changed too frequently",
			SubIdUnRegisteredPhone:      "Phone number is not registered",
			SubIdUnSupportedSignInType:  "Unsupported sign-in method",
			SubIdIncorrectPassword:      "Incorrect password",
			SubIdSignInBanned:           "Account has been banned from logging in",
			SubIdAccountBanned:          "Account has been banned",
			SubIdSignInFailed:           "Account does not exist or password is incorrect",
			SubIdAccountAlreadyExists:   "Account already exists",
			SubIdPasswordNotSetOnLogin:  "No password set. Please use another login method",
			SubIdPasswordNotGiven:       "Password not given",
		},
		ErrorCodeTooManyRequests: {
			SubIdLoginFrequently:     "Login too frequently",
			SubIdLoginFrequentlyLong: "Login restricted. Please take a break and try again later",
		},
	}
	e.addSceneError(add)
}
