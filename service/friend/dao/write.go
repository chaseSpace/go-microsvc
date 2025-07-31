package dao

import (
	"context"
	"microsvc/model/svc/friend"
	"microsvc/pkg/xerr"
	"microsvc/util/db"

	"gorm.io/gorm"
)

func FollowOne(ctx context.Context, tx *gorm.DB, uid, beFollowUid int64) (err error) {
	err = tx.WithContext(ctx).Create(&friend.Friend{
		UID: uid,
		FID: beFollowUid,
	}).Error
	if err == nil {
		return
	}

	var conflictIdx string
	if db.IsMysqlDuplicateErr(err, &conflictIdx) {
		if conflictIdx == friend.TableFriendUKMain {
			return xerr.ErrFriendAlreadyFollow
		}
	}
	return xerr.WrapMySQL(err)
}

func UnFollowOne(ctx context.Context, tx *gorm.DB, uid, unFollowUid int64) error {
	err := tx.WithContext(ctx).Delete(&friend.Friend{}, "uid =? and fid =?", uid, unFollowUid).Error
	return xerr.WrapMySQL(err)
}

func BlockOne(ctx context.Context, tx *gorm.DB, uid, bid int64) error {
	err := tx.WithContext(ctx).Create(&friend.Block{
		UID: uid,
		BID: bid,
	}).Error
	if err == nil {
		return nil
	}
	conflictIdx := ""
	if db.IsMysqlDuplicateErr(err, &conflictIdx) && conflictIdx == friend.TableBlockUKMain {
		return xerr.ErrRepeatedOperation
	}
	return xerr.WrapMySQL(err)
}

func UnBlockOne(ctx context.Context, uid, bid int64) error {
	err := friend.Q.WithContext(ctx).Delete(&friend.Block{}, "uid = ? and bid = ?", uid, bid).Error
	return xerr.WrapMySQL(err)
}

func SaveVisitor(ctx context.Context, uid, vid, seconds int64) error {
	sql := `INSERT INTO visitor (uid, vid, day_visit_times, day_visit_duration, date, created_at, updated_at)
			VALUES (?, ?, 1, ?, REPLACE(DATE(NOW()), '-', ''), NOW(), NOW())
			ON DUPLICATE KEY UPDATE day_visit_times    = day_visit_times + 1,
									day_visit_duration = day_visit_duration + ?,
									updated_at         = NOW();`
	err := friend.Q.WithContext(ctx).Exec(sql, uid, vid, seconds, seconds).Error
	return xerr.WrapMySQL(err)
}
