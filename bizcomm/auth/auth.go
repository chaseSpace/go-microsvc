package auth

import (
	"context"
	"fmt"
	"microsvc/consts"
	"microsvc/enums"
	"microsvc/pkg/xlog"
	"microsvc/util"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/spf13/cast"
	"go.uber.org/zap"
)

const (
	HeaderKey           = "Authorization"
	TokenIssuer         = "x.microsvc"
	TokenIssuerExtAdmin = "x.external.admin"
)

type Method struct {
}

var NoAuthMethods = map[string]*Method{
	// user
	"/svc.user.userExt/SignInAll":           {},
	"/svc.user.userExt/SignUpAll":           {},
	"/svc.user.userExt/ResetPassword":       {},
	"/svc.user.userExt/SendEmailCodeOnUser": {},

	// thirdparty
	"/svc.thirdparty.thirdpartyExt/VerifyEmailCode": {},

	// barbase
	"/svc.barbase.barbaseExt/ListRecommendBars":        {},
	"/svc.barbase.barbaseExt/SearchBars":               {},
	"/svc.barbase.barbaseExt/GetBarInfo":               {},
	"/svc.barbase.barbaseExt/ListSearchCandidateWords": {},
	"/svc.barbase.barbaseExt/ListEventsForUser":        {},
	"/svc.barbase.barbaseExt/GetEventInfo":             {},
	"/svc.barbase.barbaseExt/ListReviewsForWine":       {},
}

type CtxAuthenticated struct{}

func ExtractSvcUser(ctx context.Context) *SvcCaller {
	caller := ctx.Value(CtxAuthenticated{}).(*SvcClaims).SvcCaller
	return &caller
}

func ExtractAdminUser(ctx context.Context) *AdminCaller {
	caller := ctx.Value(CtxAuthenticated{}).(*AdminClaims).AdminCaller
	return &caller
}

func GetAuthUserAPI(ctx context.Context) AuthenticateUserAPI {
	obj, _ := ctx.Value(CtxAuthenticated{}).(AuthenticateUserAPI)
	return obj
}

func GetAuthUID(ctx context.Context) int64 {
	obj := GetAuthUserAPI(ctx)
	if obj != nil {
		return obj.GetCredentialUID()
	}
	return 0
}

func HasAuth(ctx context.Context) bool {
	return ctx.Value(CtxAuthenticated{}) != nil
}

type AuthenticateUserAPI interface {
	jwt.Claims
	IsValidCredential() bool
	GetCredentialUID() int64
	GetCredentialLoginAt() time.Time
	GetCredentialRegAt() time.Time
}

var _ AuthenticateUserAPI = new(SvcClaims)
var _ AuthenticateUserAPI = new(AdminClaims)

type AdminCaller struct {
	Credential
	IsSuper bool `json:"is_super"` // 是否超级管理员，超管是普通管理员的超集！
}

type SvcCaller struct {
	Credential
}

type SvcClaims struct {
	SvcCaller
	jwt.RegisteredClaims
}

type AdminClaims struct {
	AdminCaller
	jwt.RegisteredClaims
}

func (a *SvcCaller) IsValidCredential() bool {
	if a.Credential.IsValidCredential() {
		return a.Uid > 0
	}
	return false
}

// Credential 基础凭据
type Credential struct {
	Uid      int64     `json:"uid"`
	Nickname string    `json:"nickname"`
	Sex      enums.Sex `json:"sex"`
	LoginAt  string    `json:"login_at"`
	RegAt    string    `json:"reg_at"`
	_LoginAt time.Time
	_RegAt   time.Time
}

func (a *Credential) GetCredentialUID() int64 {
	return a.Uid
}

func (a *Credential) IsValidCredential() (valid bool) {
	defer func() {
		if !valid {
			xlog.Error("Credential is not valid", zap.Any("cre", a))
		}
	}()
	var err1, err2 error
	a._LoginAt, err1 = consts.Datetime(a.LoginAt).Time()
	a._RegAt, err2 = consts.Datetime(a.RegAt).Time()
	if err1 != nil || err2 != nil {
		fmt.Printf("")
		return
	}
	if a.Uid > 0 {
		return true
	}
	// 注意：这里的条件判断必须和 GenerateJwT 方法同步，否则颁发的token无法调用接口
	return
}

func (a *Credential) GetCredentialLoginAt() time.Time {
	return a._LoginAt
}

func (a *Credential) GetCredentialRegAt() time.Time {
	return a._RegAt
}

func NewTestSvcUser(uid int64, sex enums.Sex) *SvcClaims {
	return &SvcClaims{
		SvcCaller: SvcCaller{Credential: Credential{
			Uid:      uid,
			Nickname: "",
			Sex:      sex,
			RegAt:    "2024-01-01 00:00:00",
			LoginAt:  time.Now().Format(time.DateTime),
		}},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: nil, // never expire
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			NotBefore: jwt.NewNumericDate(time.Now()),
			Issuer:    TokenIssuer,
			Subject:   cast.ToString(uid),
			ID:        util.NewKsuid(),
		},
	}
}

func NewTestAdminUser(uid int64, sex enums.Sex) *AdminClaims {
	now := time.Now()
	return &AdminClaims{
		AdminCaller: AdminCaller{
			Credential: Credential{
				Uid:      uid,
				Nickname: "haha",
				Sex:      sex,
				RegAt:    "2024-01-01 00:00:00",
				LoginAt:  time.Now().Format(time.DateTime),
			},
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: nil, // never expire
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    TokenIssuer,
			Subject:   cast.ToString(uid),
			ID:        util.NewKsuid(),
		},
	}
}

func NewFakeAdminCaller() *AdminCaller {
	return &AdminCaller{
		Credential: Credential{},
		IsSuper:    false,
	}
}
