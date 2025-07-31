package commpay

import (
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/commonpb"
)

// 货币符号文件下载
//https://www.six-group.com/en/products-services/financial-information/data-standards.html#scrollTo=isin

type __CurrencyISO struct {
	Char                string
	SysSupport          bool
	DefaultExchange2USD float64 // 默认转USD汇率
}

var currencyIsoMap = map[commonpb.CurrencyType]__CurrencyISO{
	commonpb.CurrencyType_CT_CNY: {"CNY", true, 0.138},
	commonpb.CurrencyType_CT_USD: {"USD", true, 1},
}

func GetCurrencyConf(currencyType commonpb.CurrencyType) __CurrencyISO {
	return currencyIsoMap[currencyType]
}

func ToISOCurrencyChar(currencyType commonpb.CurrencyType, systemSupport ...bool) (char string, err error) {
	if iso, ok := currencyIsoMap[currencyType]; ok {
		if len(systemSupport) > 0 && systemSupport[0] {
			if !iso.SysSupport {
				return "", xerr.ErrCurrencyNotSupported
			}
		}
		return iso.Char, nil
	}
	return "", xerr.ErrParams.New("Currency not configured")
}

func CurrencyAmount2USD(amount float64, currencyType commonpb.CurrencyType) (float64, error) {
	if iso, ok := currencyIsoMap[currencyType]; ok {
		if iso.SysSupport {
			if iso.DefaultExchange2USD <= 0 {
				return 0, xerr.ErrParams.New("Currency exchange ratio not configured")
			}
			return amount * iso.DefaultExchange2USD, nil
		} else {
			return 0, xerr.ErrParams.New("Currency not supported")
		}
	}
	return 0, xerr.ErrParams.New("Currency not configured")
}
