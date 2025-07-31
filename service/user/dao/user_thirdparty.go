package dao

import (
	"microsvc/model/svc/user"
	"time"

	"gorm.io/gorm"
)

func CreateUserTh(tx *gorm.DB, ent *user.UserRegisterTh) error {
	ent.CreatedAt = time.Now()
	ent.UpdatedAt = ent.CreatedAt
	return tx.Create(ent).Error
}
