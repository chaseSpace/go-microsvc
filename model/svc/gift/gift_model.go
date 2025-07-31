package gift

import (
	"fmt"
	"microsvc/consts"
	"microsvc/model"
	"microsvc/model/modelsql"
	"microsvc/protocol/svc/giftpb"
	"strings"
	"time"
)

const (
	TableUKGiftAccount  = "'uk_uid_giftId'"
	TableUKGiftConfName = "'uk_name'"
)

type GiftConf struct {
	model.TableBase
	Name            string             `gorm:"column:name" json:"name"`
	Icon            string             `gorm:"column:icon" json:"icon"`
	Price           int64              `gorm:"column:price" json:"price"` // 金币价格
	Type            giftpb.GiftType    `gorm:"column:type" json:"type"`
	State           giftpb.GiftState   `gorm:"column:state" json:"state"`
	SupportedScenes []giftpb.GiftScene `gorm:"column:supported_scenes;serializer:json" json:"supported_scenes"` // 不含0
}

func (*GiftConf) TableName() string {
	return "gift_conf"
}

func (g *GiftConf) ToPB(amount int64) *giftpb.Gift {
	return &giftpb.Gift{
		Id:     g.Id,
		Name:   g.Name,
		Icon:   g.Icon,
		Price:  g.Price,
		Type:   g.Type,
		Amount: amount,
	}
}

func (g *GiftConf) Check() error {
	if g.Name == "" {
		return fmt.Errorf("名称不能为空")
	}
	if giftpb.GiftType_name[int32(g.Type)] == "" {
		return fmt.Errorf("无效的礼物类型")
	}
	if giftpb.GiftState_name[int32(g.State)] == "" {
		return fmt.Errorf("无效的礼物状态")
	}
	if g.Icon == "" {
		return fmt.Errorf("图标不能为空")
	}
	if g.Price < 1 {
		return fmt.Errorf("价格不能小于1")
	}
	if len(g.SupportedScenes) == 0 {
		return fmt.Errorf("至少支持一种场景")
	}
	for _, v := range g.SupportedScenes {
		if v == giftpb.GiftScene_GS_Unknown || giftpb.GiftScene_name[int32(v)] == "" {
			return fmt.Errorf("无效的礼物场景")
		}
	}
	return nil
}

func (g *GiftConf) ToGiftItem() *giftpb.GiftItem {
	return &giftpb.GiftItem{
		Meta:      g.ToPB(0),
		State:     g.State,
		CreatedAt: g.CreatedAt.Unix(),
		UpdatedAt: g.UpdatedAt.Unix(),
	}
}

// GiftAccount 礼物账户表（先购买，再赠送，不支持直接通过金币赠送）
type GiftAccount struct {
	model.FieldUID
	model.FieldAt
	GiftID int64 `gorm:"column:gift_id" json:"gift_id"`
	Amount int64 `gorm:"column:amount" json:"amount"`
}

func (GiftAccount) TableName() string {
	return "gift_account"
}

// GiftTxLog 礼物交易记录表（总表，用于统计）
type GiftTxLog struct {
	model.TableBase
	TxId       string           `gorm:"column:tx_id" json:"tx_id"`
	FromUID    int64            `gorm:"column:from_uid" json:"from_uid"`
	ToUID      int64            `gorm:"column:to_uid" json:"to_uid"`
	GiftID     int64            `gorm:"column:gift_id" json:"gift_id"`
	GiftName   string           `gorm:"column:gift_name" json:"gift_name"`
	Price      int64            `gorm:"column:price" json:"price"`
	Amount     int64            `gorm:"column:amount" json:"amount"`
	TotalValue int64            `gorm:"column:total_value" json:"total_value"` // 总价值：price * amount
	TxType     giftpb.TxType    `gorm:"column:tx_type" json:"tx_type"`
	GiftScene  giftpb.GiftScene `gorm:"column:gift_scene" json:"gift_scene"`
	GiftType   giftpb.GiftType  `gorm:"column:gift_type" json:"gift_type"`
	Remark     string           `gorm:"column:remark" json:"remark"`
}

func (*GiftTxLog) TableName() string {
	return "gift_tx_log"
}

func (t *GiftTxLog) ToIntPB() *giftpb.GiftTxLogInt {
	return &giftpb.GiftTxLogInt{
		Base: &giftpb.GiftTxLogBase{
			TxId:       t.TxId,
			GiftId:     t.GiftID,
			GiftName:   t.GiftName,
			Price:      t.Price,
			Amount:     t.Amount,
			TotalValue: t.TotalValue,
			TxScene:    t.GiftScene,
			GiftType:   t.GiftType,
			CreatedAt:  t.CreatedAt.Unix(),
		},
		FromUid: t.FromUID,
		ToUid:   t.ToUID,
		TxType:  t.TxType,
	}
}

// GiftTxLogPersonal 个人礼物交易记录表（月表，用于查询）
type GiftTxLogPersonal struct {
	*model.TableBase
	suffix            string
	TxId              string                     `gorm:"column:tx_id" json:"tx_id"`
	UID               int64                      `gorm:"column:uid" json:"uid"`
	GiftId            int64                      `gorm:"column:gift_id" json:"gift_id"`
	GiftName          string                     `gorm:"column:gift_name" json:"gift_name"`
	Price             int64                      `gorm:"column:price" json:"price"`
	RelatedUID        int64                      `gorm:"column:related_uid" json:"related_uid"`
	Delta             int64                      `gorm:"column:delta" json:"delta"`
	TotalValue        int64                      `gorm:"column:total_value" json:"total_value"` // 总价值：price * amount
	Balance           int64                      `gorm:"column:balance" json:"balance"`         // 变更后
	Remark            string                     `gorm:"column:remark" json:"remark"`
	FirstPersonTxType giftpb.FirstPersonalTxType `gorm:"column:first_person_tx_type" json:"first_person_tx_type"`
	GiftScene         giftpb.GiftScene           `gorm:"column:gift_scene" json:"gift_scene"`
	GiftType          giftpb.GiftType            `gorm:"column:gift_type" json:"gift_type"`
}

// TableName 注意：这个方法读取了内部变量
// gorm 不能通过 Model() 读取到 TableName（成员变量永远为空），只能通过 gorm.DB.Table(v.TableName()) 设置
func (t *GiftTxLogPersonal) TableName() string {
	return "gift_tx_log_personal_" + t.suffix
}

func (t *GiftTxLogPersonal) SetSuffix(suffix string) {
	t.suffix = suffix
}

func (t *GiftTxLogPersonal) SetSuffixByTime(ti time.Time) {
	t.suffix = ti.Format("200601")
}

func (t *GiftTxLogPersonal) GetSuffix() string {
	return t.suffix
}

func (t *GiftTxLogPersonal) DDLSql() string {
	return fmt.Sprintf(strings.Replace(modelsql.GiftTxLogPersonalMonthTable, consts.YearMonth, t.suffix, 1))
}

func (t *GiftTxLogPersonal) ToPB() *giftpb.GiftPersonalTxLog {
	return &giftpb.GiftPersonalTxLog{
		Base: &giftpb.GiftTxLogBase{
			TxId:       t.TxId,
			GiftId:     t.GiftId,
			GiftName:   t.GiftName,
			Price:      t.Price,
			Amount:     t.Delta,
			TotalValue: t.TotalValue,
			TxScene:    t.GiftScene,
			GiftType:   t.GiftType,
			CreatedAt:  t.CreatedAt.Unix(),
		},
		RelatedUid: t.RelatedUID,
		Balance:    t.Balance,
		TxType:     t.FirstPersonTxType,
	}
}
