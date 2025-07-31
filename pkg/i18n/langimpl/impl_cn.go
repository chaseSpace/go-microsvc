package langimpl

import (
	"fmt"
	"microsvc/protocol/svc/commonpb"
)

type Chinese struct {
	meta map[Scene]MsgMap
}

func (e *Chinese) Lang() commonpb.Lang {
	return commonpb.Lang_CL_CN
}

func (e *Chinese) Translate(scene Scene, id Code, subId SubId) string {
	if m1 := e.meta[scene]; m1 != nil {
		m2, ok := m1[id]
		if ok {
			if subId != "" {
				return m2[subId]
			}
			return m2["default"]
		}
	}
	return ""
}

var CN = &Chinese{
	meta: map[Scene]MsgMap{
		SceneError: map[Code]map[SubId]string{},
	},
}

func init() {
	CN.registerSceneErrorForCommon()
	CN.registerSceneErrorForFriendSvc()
	CN.registerSceneErrorGiftSvc()
	CN.registerSceneErrorForCurrencySvc()
	CN.registerSceneErrorForMomentSvc()
	CN.registerSceneErrorForThirdPartySvc()
	CN.registerSceneErrorForUserSvc()
}

func (e *Chinese) registerSceneErrorForCommon() {
	add := map[Code]string{
		ErrorCodeNil:              "成功",
		ErrorCodeParams:           "提供的参数无效",
		ErrorCodeUnauthorized:     "需要认证",
		ErrorCodeForbidden:        "访问被拒绝",
		ErrorCodeNotFound:         "资源未找到",
		ErrorCodeMethodNotAllowed: "不支持的方法",
		ErrorCodeReqTimeout:       "请求超时",
		ErrorCodeTooManyRequests:  "请求过多，请稍后再试",

		ErrorCodeInternal:           "内部服务器错误",
		ErrorCodeGateway:            "网关错误",
		ErrorCodeServiceUnavailable: "目标微服务不可用",
		ErrorCodeAPIUnavailable:     "API不可用",
		ErrorCodeMySQL:              "MySQL错误",
		ErrorCodeRedis:              "Redis错误",

		ErrorCodeBizTimeout:                "业务超时",
		ErrorCodeThirdParty:                "第三方服务异常",
		ErrorCodeInvalidRegisterInfo:       "无效的注册信息",
		ErrorCodeUserNotFound:              "用户不存在",
		ErrorCodeRepeatedOperation:         "操作重复执行",
		ErrorCodeJSONMarshal:               "JSON序列化失败",
		ErrorCodeJSONUnmarshal:             "JSON反序列化失败",
		ErrorCodeIDShouldBeZeroOnAdd:       "新增记录必须有零ID",
		ErrorCodeIDShouldNotBeZeroOnUpdate: "更新需要非零ID",
		ErrorCodeNoRowAffectedOnUpdate:     "数据无变化或不存在",
		ErrorCodeInvalidID:                 "提供的ID无效",
		ErrorCodeDataNotExist:              "目标数据不存在",
		ErrorCodeBaseReq:                   "基本请求参数错误",
		ErrorCodeReviewResultReject:        "内容审核拒绝",
		ErrorCodeDataDuplicate:             "检测到重复数据",
		ErrorCodeHTTPRequest:               "HTTP请求错误",
		ErrorCodeRPC:                       "RPC调用错误",
		ErrorCodeCurrencyNotSupported:      "不支持的币种",
	}

	for k, v := range add {
		if _, ok := e.meta[SceneError][k]; ok {
			panic(fmt.Sprintf("duplicate key: (%d: %s) on register code", k, v))
		}
		e.meta[SceneError][k] = map[SubId]string{"default": v} // default msg of this code
	}
}

func (e *Chinese) addSceneError(add MsgMap) {
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

	//fmt.Printf("33333 %+v\n", e.meta)
}

func (e *Chinese) registerSceneErrorForCurrencySvc() {
	add := map[Code]map[SubId]string{
		ErrorCodeParams: {
			SubIdTxFromAndToCannotBeSame: "交易发起人和接收人不能是同一人",
			SubIdTxAmountShouldNotBeZero: "交易金额不能为零",
			SubIdTxRemarkTooLong:         "交易备注不能超过100个字符",
			SubIdTxBalanceNotEnough:      "余额不足",
			SubIdTxInvalidTxType:         "无效的交易类型",
			SubIdTxInvalidSingleTxType:   "无效的单个交易类型",
		},
	}
	e.addSceneError(add)
}

func (e *Chinese) registerSceneErrorForFriendSvc() {
	add := map[Code]map[SubId]string{
		ErrorCodeParams: {
			SubIdFriendAlreadyFollow:    "已经关注",
			SubIdFriendCountUpToMax:     "好友数量已达上限",
			SubIdFriendPeerCountUpToMax: "相互好友数量已达上限",
		},
	}
	e.addSceneError(add)
}

func (e *Chinese) registerSceneErrorGiftSvc() {
	add := map[Code]map[SubId]string{
		ErrorCodeParams: {
			SubIdGiftTxFromAndToCannotBeSame:      "发送者和接收者不能是同一人",
			SubIdGiftTxAmountMustBePositive:       "礼物数量必须大于0",
			SubIdGiftTxAmountMustBePositiveOnSent: "发送时礼物数量必须大于0",
			SubIdGiftTxInvalidTxType:              "无效的交易类型",
			SubIdGiftTxInvalidTxScene:             "无效的礼物场景",
			SubIdGiftTxInvalidFirstPersonTxType:   "无效的第一人称交易类型",
			SubIdGiftTxTypeConvertFailed:          "交易类型转换失败",
			SubIdGiftTxBalanceNotEnough:           "礼物余额不足",
			SubIdGiftNotFound:                     "礼物未找到",
		},
	}
	e.addSceneError(add)
}

func (e *Chinese) registerSceneErrorForMomentSvc() {
	add := map[Code]map[SubId]string{
		ErrorCodeParams: {
			SubIdMomentTextTooLong:  "文本不能超过200个字符",
			SubIdMomentTypeNotFound: "不支持的动态类型",
			SubIdMomentNotFound:     "动态不存在",
			SubIdCommentNotFound:    "评论不存在",
			SubIdCommentTextTooLong: "评论不能超过100个字符",
		},
	}
	e.addSceneError(add)
}

func (e *Chinese) registerSceneErrorForThirdPartySvc() {
	add := map[Code]map[SubId]string{
		ErrorCodeParams: {
			SubIdInvalidFileType:               "文件类型无效",
			SubIdThirdPartyServiceNameNotMatch: "第三方服务名称不匹配",
			SubIdEmailCodeNeedSendFirst:        "需要先发送验证码",
			SubIdWriteFileFailed:               "写入文件失败",
		},
	}
	e.addSceneError(add)
}

func (e *Chinese) registerSceneErrorForUserSvc() {
	add := map[Code]map[SubId]string{
		ErrorCodeParams: {
			SubIdPasswdFormat:           "请输入有效的密码参数",
			SubIdPasswdNoChange:         "新旧密码不能一致",
			SubIdInvalidPhoneNo:         "请提供有效的手机号",
			SubIdInvalidVerifyCode:      "验证码不正确",
			SubIdInvalidLenPhoneNo:      "手机号长度不正确",
			SubIdNotSupportedPhoneArea:  "不支持的手机号区域码",
			SubIdChangePasswdFrequently: "密码更改过于频繁",
			SubIdUnRegisteredPhone:      "手机号未注册",
			SubIdUnSupportedSignInType:  "不支持的登录方式",
			SubIdIncorrectPassword:      "密码不正确",
			SubIdSignInBanned:           "账户已被禁止登录",
			SubIdAccountBanned:          "账户已被禁止",
			SubIdSignInFailed:           "账户不存在或密码不正确",
			SubIdAccountAlreadyExists:   "账户已存在",
			SubIdPasswordNotSetOnLogin:  "密码未设置，请更换登录方式",
			SubIdPasswordNotGiven:       "密码未提供",
		},
		ErrorCodeTooManyRequests: {
			SubIdLoginFrequently:     "登录过于频繁，稍后再试",
			SubIdLoginFrequentlyLong: "短时间内登录次数过多，请晚些时候再试",
		},
	}
	e.addSceneError(add)
}
