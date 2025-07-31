package comminfra

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"sync"
)

var monthTableMap sync.Map

func HasTable(db *gorm.DB, table string) bool {
	k := "is_table_exists_" + db.Dialector.(*mysql.Dialector).DSNConfig.DBName + "_" + table
	if val, ok := monthTableMap.Load(k); ok && val.(bool) {
		return true
	}
	if db.Migrator().HasTable(table) {
		monthTableMap.Store(k, true)
		return true
	}
	return false
}
