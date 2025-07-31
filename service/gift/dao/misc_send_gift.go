package dao

import (
	"context"
	"microsvc/consts"
	"microsvc/model"
	"microsvc/model/svc/currency"
	"microsvc/model/svc/gift"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/giftpb"
	"microsvc/util"
	"microsvc/util/db"
	"microsvc/util/umath"
	"time"
	"unicode/utf8"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type giftTxCtrlT struct {
}

var GiftTxCtrl = giftTxCtrlT{}

// GiftTxParams 交易参数（赠送）
type GiftTxParams struct {
	FromUID int64
	ToUID   int64
	GiftID  int64
	Delta   int64
	Remark  string
	giftpb.TxType
	giftpb.GiftScene
	G                 *gift.GiftConf
	__fpTxType        giftpb.FirstPersonalTxType
	__fpTxTypeReverse giftpb.FirstPersonalTxType
}

// GiftTransaction 执行Gift双人交易
// -- fromUID 扣减，toUID 增加
func (g giftTxCtrlT) GiftTransaction(ctx context.Context, params *GiftTxParams) error {
	if err := g.checkParams(params); err != nil {
		return err
	}
	return currency.Q.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return g.txLogic(ctx, tx, params)
	})
}

func (g giftTxCtrlT) txLogic(ctx context.Context, tx *gorm.DB, params *GiftTxParams) error {
	// 单人交易（购买、管理员发放）
	if params.FromUID == params.ToUID {
		return g.personalTxLogic(ctx, tx, params)
	}

	// 双人交易（赠送）
	// from-
	do := tx.WithContext(ctx).Model(&gift.GiftAccount{}).
		Where("uid = ? and gift_id = ?", params.FromUID, params.GiftID).
		Where("amount >= ?", params.Delta).
		Update("amount", gorm.Expr("amount - ?", params.Delta))
	if do.Error != nil {
		return xerr.WrapMySQL(do.Error)
	}
	if do.RowsAffected != 1 {
		return xerr.ErrGiftTxBalanceNotEnough
	}
	// to+
	err := g.saveGiftAccount(ctx, tx, params.ToUID, params.GiftID, params.Delta)
	if err != nil {
		return err
	}

	// 取出双人余额，写入记录表
	_, gmap, err := GetAccountSingleGift(ctx, tx, params.GiftID, params.FromUID, params.ToUID)
	if err != nil {
		return xerr.WrapMySQL(err)
	}

	// 写入交易记录
	fromBalance, toBalance := gmap[params.FromUID].Amount, gmap[params.ToUID].Amount
	return g.addTxLog(ctx, tx, params, fromBalance, toBalance)
}

// 个人交易（FromUID == ToUID）
func (g giftTxCtrlT) personalTxLogic(ctx context.Context, tx *gorm.DB, params *GiftTxParams) (err error) {
	if params.Delta > 0 { // ++
		// insert or update
		err = g.saveGiftAccount(ctx, tx, params.FromUID, params.GiftID, params.Delta)
		if err != nil {
			return err
		}
	} else {
		// update (--)
		do := tx.WithContext(ctx).
			Model(&gift.GiftAccount{}).
			Where("uid = ? and gift_id = ?", params.FromUID, params.GiftID).
			Update("amount", gorm.Expr("amount + ?", params.Delta))
		if do.Error != nil {
			return xerr.WrapMySQL(do.Error)
		}
		if do.RowsAffected != 1 {
			return xerr.ErrGiftTxBalanceNotEnough
		}
	}
	_, gmap, err := GetAccountSingleGift(ctx, tx, params.GiftID, params.FromUID)
	if err != nil {
		return err
	}
	fromBalance := gmap[params.FromUID].Amount
	return g.addTxLog(ctx, tx, params, fromBalance, 0)
}

func (g giftTxCtrlT) addTxLog(ctx context.Context, tx *gorm.DB, params *GiftTxParams, fromBalance, toBalance int64) error {
	uniqueTxID := util.NewKsuid()
	totalValue := umath.Abs(params.Delta * params.G.Price)
	// 增加交易记录（第三人称）
	err := tx.WithContext(ctx).Create(&gift.GiftTxLog{
		TxId:       uniqueTxID,
		FromUID:    params.FromUID,
		ToUID:      params.ToUID,
		GiftID:     params.GiftID,
		GiftName:   params.G.Name,
		Price:      params.G.Price,
		Amount:     params.Delta,
		TotalValue: totalValue,
		TxType:     params.TxType,
		GiftScene:  params.GiftScene,
		GiftType:   params.G.Type,
		Remark:     params.Remark,
	}).Error
	if err != nil {
		return xerr.WrapMySQL(err)
	}

	// 增加个人交易记录（双人交易时为两条）
	row := &gift.GiftTxLogPersonal{
		TxId:              uniqueTxID,
		UID:               params.FromUID,
		RelatedUID:        params.ToUID,
		GiftId:            params.GiftID,
		GiftName:          params.G.Name,
		Price:             params.G.Price,
		Delta:             params.Delta,
		Balance:           fromBalance,
		TotalValue:        totalValue,
		FirstPersonTxType: params.__fpTxType,
		GiftScene:         params.GiftScene,
		GiftType:          params.G.Type,
		Remark:            params.Remark,
	}
	// 双人交易时，delta一定是正，所以这里要取反（from--  to++）
	if params.FromUID != params.ToUID {
		row.Delta = -row.Delta
	}
	row.SetSuffixByTime(time.Now())
	tx = tx.WithContext(ctx)

	helper := db.NewTableHelper(tx, row.DDLSql())
	err = helper.AutoCreateTable(func(tx2 *gorm.DB) error {
		// 增加交易记录（第一人称）
		return tx2.Table(row.TableName()).Create(row).Error
	})
	// 增加另一人的交易记录（若是双人交易）
	if params.FromUID != params.ToUID {
		row.Id = 0 // RESET
		row.UID = params.ToUID
		row.RelatedUID = params.FromUID
		row.Delta = params.Delta
		row.Balance = toBalance
		row.FirstPersonTxType = params.__fpTxTypeReverse
		err = tx.Table(row.TableName()).Create(row).Error
	}
	return xerr.WrapMySQL(err)
}

func (giftTxCtrlT) checkParams(params *GiftTxParams) (err error) {
	isPersonalTx := params.FromUID == params.ToUID
	if params.FromUID < 1 || params.ToUID < 1 {
		return xerr.ErrUserNotFound.AppendMsg("包含无效的交易UID")
	}
	if params.Delta < 1 && !isPersonalTx {
		return xerr.ErrGiftTxAmountMustBePositiveOnSent
	}
	if utf8.RuneCountInString(params.Remark) > consts.TxRemarkMaxLen {
		return xerr.ErrTxRemarkTooLong
	}
	if params.TxType < 1 || giftpb.TxType_name[int32(params.TxType)] == "" {
		return xerr.ErrTxInvalidTxType
	}
	params.__fpTxType, err = gift.TxType2FPTxType(params.TxType)
	if err != nil {
		return err
	}
	params.__fpTxTypeReverse = gift.FPTxTypeReverse(params.__fpTxType)
	return
}

func (giftTxCtrlT) saveGiftAccount(ctx context.Context, tx *gorm.DB, uid, giftID, amount int64) (err error) {
	err = tx.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "uid"}, {Name: "gift_id"}},
			DoUpdates: clause.Assignments(map[string]interface{}{"amount": gorm.Expr("amount + ?", amount)}),
		}).Create(&gift.GiftAccount{FieldUID: model.FieldUID{UID: uid}, GiftID: giftID, Amount: amount}).Error
	return xerr.WrapMySQL(err)
}
