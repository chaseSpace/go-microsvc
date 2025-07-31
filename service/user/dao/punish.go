package dao

import (
	"context"
	"errors"
	"microsvc/model"
	"microsvc/model/svc/user"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/commonpb"
	"microsvc/protocol/svc/userpb"
	"microsvc/util/db"
	"strings"
	"unicode/utf8"

	"gorm.io/gorm"
)

type punishT struct {
}

var Punish = &punishT{}

func (p *punishT) New(ctx context.Context, params *user.Punish) error {
	err := user.Q.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row := &user.Punish{}
		err := tx.Where("uid=? AND type=? and state=? AND ADDDATE(created_at, interval duration second) > now()",
			params.UID, params.Type, commonpb.PunishState_PS_InProgress).
			Take(row).Error
		if db.IsMysqlErr(err) {
			return err
		}
		if row.Id != 0 { // 自动续期
			err = p.IncrDuration(ctx, row.Id, params.Duration, params.CreatedBy, params.Reason)
			return err
		}
		err = tx.Create(params).Error
		if err != nil {
			return err
		}
		err = tx.Create(&user.PunishLog{
			FieldUID:       model.FieldUID{UID: params.UID},
			PunishType:     params.Type,
			OpType:         commonpb.PunishOpType_POT_New,
			Reason:         params.Reason,
			Duration:       params.Duration,
			FieldCreatedBy: model.FieldCreatedBy{CreatedBy: params.CreatedBy},
		}).Error
		return err
	})
	return xerr.WrapMySQL(err)
}

func (*punishT) IncrDuration(ctx context.Context, id int64, duration, adminUID int64, reason string) error {
	if id == 0 || duration < 1 || adminUID < 1 {
		return xerr.ErrParams
	}
	if reason == "" {
		return xerr.ErrParams.New("Field `reason` is required")
	}
	err := user.Q.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row := new(user.Punish)
		err := tx.Take(row, "id=?", id).Error
		if db.IsMysqlErr(err) {
			return xerr.WrapMySQL(err)
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return xerr.ErrDataNotExist.New("No record could be update")
		}
		newReason := strings.Join([]string{row.Reason, reason}, ";")
		if utf8.RuneCountInString(newReason) > 200 {
			return xerr.ErrParams.New("Total reason exceeds 200 chars")
		}
		// 只能操作正在惩罚中的记录
		v := tx.Model(&user.Punish{}).
			Where("id=? and (state=? AND ADDDATE(created_at, interval duration second) > now())", id, commonpb.PunishState_PS_InProgress).
			Update("duration", gorm.Expr("duration + ?", duration)).
			Update("reason", newReason).
			Update("updated_by", adminUID)
		if v.Error != nil {
			return v.Error
		}
		if v.RowsAffected != 1 {
			return xerr.ErrDataNotExist.New("No record could be update")
		}

		// add log
		err = tx.Create(&user.PunishLog{
			FieldUID:       model.FieldUID{UID: row.UID},
			OpType:         commonpb.PunishOpType_POT_IncrDuration,
			PunishType:     row.Type,
			Reason:         reason,
			Duration:       duration,
			FieldCreatedBy: model.FieldCreatedBy{CreatedBy: adminUID},
		}).Error
		return xerr.WrapMySQL(err)
	})
	return err
}

// Dismiss 提前解除惩罚
func (*punishT) Dismiss(ctx context.Context, id, adminUID int64, reason string) error {
	if reason == "" {
		return xerr.ErrParams.New("Field `reason` is required")
	}
	err := user.Q.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		row := new(user.Punish)
		err := tx.Take(row, "id=? and state=?", id, commonpb.PunishState_PS_InProgress).Error
		if db.IsMysqlErr(err) {
			return err
		}
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return xerr.ErrDataNotExist.New("No record could be operate")
		}
		// 解除时不要覆盖 reason，写入下面的操作记录即可
		err = tx.Model(&user.Punish{}).Where("id=?", id).
			Updates(map[string]interface{}{
				"state":      commonpb.PunishState_PS_Dismissed,
				"updated_by": adminUID,
				"reason":     strings.Join([]string{row.Reason, "（解除）" + reason}, ";"),
			}).Error
		if err != nil {
			return xerr.WrapMySQL(err)
		}
		// add log
		err = tx.Create(&user.PunishLog{
			FieldUID:       model.FieldUID{UID: row.UID},
			PunishType:     commonpb.PunishType_PT_None,
			OpType:         commonpb.PunishOpType_POT_Dismiss,
			Reason:         reason,
			Duration:       0,
			FieldCreatedBy: model.FieldCreatedBy{CreatedBy: adminUID},
		}).Error
		return xerr.WrapMySQL(err)
	})
	return err
}

// List 分页查询惩罚记录
func (*punishT) List(ctx context.Context, req *userpb.PunishListReq) (list []*user.PunishRPC, total int64, err error) {
	q := user.Q.WithContext(ctx).Table(new(user.Punish).TableAs("p"))
	if len(req.SearchUid) > 0 {
		q = q.Where("p.uid in (?)", req.SearchUid)
	}
	if req.SearchAdminUid > 0 {
		q = q.Where("p.created_by=? or p.updated_by=?", req.SearchAdminUid, req.SearchAdminUid)
	}
	if len(req.SearchType) > 0 {
		q = q.Where("p.type in (?)", req.SearchType)
	}
	if req.SearchState > commonpb.PunishState_PS_None {
		if req.SearchState == commonpb.PunishState_PS_NaturalEnd {
			q = q.Where("ADDDATE(p.created_at, interval p.duration second) <= now()")
		} else if req.SearchState == commonpb.PunishState_PS_InProgress {
			q = q.Where("p.state=?", req.SearchState).Where("ADDDATE(p.created_at, interval p.duration second) > now()")
		} else {
			q = q.Where("p.state=?", req.SearchState)
		}
	}
	q = q.Joins("LEFT JOIN user AS u ON p.uid=u.uid").Select("p.*, u.nickname")
	err = db.PageQuery(q, req.Page, "p.updated_at desc,p.id desc", &total, &list)
	err = xerr.WrapMySQL(err)
	return
}

// ListPunishLog 查询单个用户所有惩罚记录
func (*punishT) ListPunishLog(ctx context.Context, uid int64) (list []*user.PunishLogRPC, err error) {
	err = user.Q.WithContext(ctx).Table(new(user.PunishLog).TableAs("p")).
		Joins("LEFT JOIN user AS u ON p.uid=u.uid").
		Where("p.uid=?", uid).
		Order("p.created_at desc,p.id desc").
		Select("p.*, u.nickname").
		Scan(&list).Error
	err = xerr.WrapMySQL(err)
	return
}

func (*punishT) GetUserPunish(ctx context.Context, req *userpb.GetUserPunishReq) (list []*user.PunishRPC, err error) {
	q := user.Q.WithContext(ctx).
		Where("uid = ?", req.Uid).
		Where("state = ?", commonpb.PunishState_PS_InProgress).
		Where("ADDDATE(created_at, interval duration second) > now()")
	if req.Type > commonpb.PunishType_PT_None {
		q = q.Where("type = ?", req.Type)
	}
	err = q.Find(&list).Error
	return list, xerr.WrapMySQL(err)
}
