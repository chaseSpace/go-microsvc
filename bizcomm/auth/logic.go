package auth

import (
	"context"
	"fmt"
	"microsvc/bizcomm/commcache"
	"microsvc/deploy"
	"microsvc/model/svc/micro_svc"

	"github.com/golang-jwt/jwt/v5"
)

func GenerateJwT(claims jwt.Claims, signKey string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	ss, err := token.SignedString([]byte(signKey))
	return ss, err
}

// IsUIDErased UID是否被抹除（删除账号）
func IsUIDErased(ctx context.Context, uid int64) (bool, error) {
	r := micro_svc.R.Exists(ctx, fmt.Sprintf(commcache.CacheKeyUIDErased, uid))
	return r.Val() == 1, r.Err()
}

// EraseUID 抹除UID（删除账号）
func EraseUID(ctx context.Context, uid int64) error {
	exp, _ := deploy.XConf.GetTokenExpiry()
	r := micro_svc.R.Set(ctx, fmt.Sprintf(commcache.CacheKeyUIDErased, uid), 1, exp)
	return r.Err()
}
