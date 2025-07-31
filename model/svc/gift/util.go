package gift

import (
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/giftpb"
)

// 第三人称交易类型 => 第一人称交易类型（所有Tx类型都有一个对应的FPTx类型）
var txType2FPTxTypeMap = map[giftpb.TxType]giftpb.FirstPersonalTxType{
	giftpb.TxType_TT_Purchase:  giftpb.FirstPersonalTxType_FPTT_Purchase,
	giftpb.TxType_TT_Send:      giftpb.FirstPersonalTxType_FPTT_Send,
	giftpb.TxType_TT_AdminIncr: giftpb.FirstPersonalTxType_FPTT_AdminIncr,
	giftpb.TxType_TT_AdminDecr: giftpb.FirstPersonalTxType_FPTT_AdminDecr,
}

// 第一人称交易类型【反转】（作为ToUID记录的交易类型）
var fpTxTypeReverseMap = map[giftpb.FirstPersonalTxType]giftpb.FirstPersonalTxType{
	giftpb.FirstPersonalTxType_FPTT_Send: giftpb.FirstPersonalTxType_FPTT_Receive,
}

func TxType2FPTxType(txType giftpb.TxType) (giftpb.FirstPersonalTxType, error) {
	tp := txType2FPTxTypeMap[txType]
	if tp == 0 {
		return 0, xerr.ErrGiftTxTypeConvertFailed
	}
	return tp, nil
}

func FPTxTypeReverse(fpTxType giftpb.FirstPersonalTxType) giftpb.FirstPersonalTxType {
	tp := fpTxTypeReverseMap[fpTxType]
	if tp == 0 {
		tp = fpTxType
	}
	return tp
}
