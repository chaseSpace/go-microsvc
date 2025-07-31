package dao

import (
	"context"
	"microsvc/model/svc/micro_svc"
	"microsvc/model/svc/user"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/adminpb"
	"microsvc/util/db"
	"strings"
)

type bizUser struct {
}

var BizUser bizUser

func (*bizUser) orderFields() db.OrderFieldMap {
	return map[string]*struct{}{
		"uid":        {},
		"nid":        {},
		"sex":        {},
		"created_at": {},
		"updated_at": {},
	}
}

func (b *bizUser) ListUser(ctx context.Context, req *adminpb.ListUserReq) (list []*user.User, total int64, err error) {
	q := user.Q.WithContext(ctx).Model(&user.User{})
	if req.SearchUid != 0 {
		q = q.Where("uid = ?", req.SearchUid)
	}
	if req.SearchNid != 0 {
		q = q.Where("nid = ?", req.SearchNid)
	}
	if req.SearchNickname != "" {
		q = q.Where("nickname LIKE ?", "%"+req.SearchNickname+"%")
	}
	if req.SearchPhone != "" {
		q = q.Where("phone LIKE ?", "%"+req.SearchPhone+"%")
	}
	var orderBy string
	orderBy, err = db.GenSortClause(req.Sort, b.orderFields(), true)
	if err != nil {
		return
	}
	err = db.PageQuery(q, req.Page, orderBy, &total, &list)
	err = xerr.WrapMySQL(err)
	return
}

func (b *bizUser) GetLatestSignInLog(ctx context.Context, uids []int64) (list []*user.SignInLog, err error) {
	//tabName := new(user.SignInLog).TableName()
	// 这里按照将自增id替代时间戳来优化查询性能
	err = user.QLog.WithContext(ctx).Raw(`SELECT s1.*
			FROM sign_in_log s1
					 INNER JOIN (SELECT uid, MAX(id) as id
								 FROM sign_in_log
								 where uid in (?)
								 GROUP BY uid) s2 ON s1.uid = s2.uid AND s1.id = s2.id
			where s1.uid in (?)`, uids, uids).Scan(&list).Error

	return
}

func (b *bizUser) ListUserAPICallLog(ctx context.Context, req *adminpb.ListUserAPICallLogReq) (list []*micro_svc.APICallLog, total int64, err error) {
	row := &micro_svc.APICallLog{}
	row.SetSuffix(strings.ReplaceAll(req.TimeRange.StartDt[:7], "-", ""))
	q := user.Q.WithContext(ctx).Table(row.TableName())
	if req.Uid != 0 {
		q = q.Where("uid = ?", req.Uid)
	}
	q = q.Where("created_at between ? and ?", req.TimeRange.StartDt, req.TimeRange.EndDt)

	err = db.PageQuery(q, req.Page, "created_at desc", &total, &list)
	err = xerr.WrapMySQL(err)
	return
}

func (b *bizUser) GetLastSignInLogs(ctx context.Context, uid, limit int64) (list []*user.SignInLog, err error) {
	err = user.QLog.WithContext(ctx).Order("created_at DESC").Limit(int(limit)).Find(&list, "uid = ?", uid).Error
	return
}
