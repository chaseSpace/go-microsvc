package gateway

import (
	"microsvc/pkg/xerr"
)

func GetOpenedAPIRateLimitConf() (list []*APIRateLimitConf, err error) {
	err = Q.Find(&list, "state=?", APIRateLimitConfStateEnabled).Error
	return list, xerr.WrapMySQL(err)
}
