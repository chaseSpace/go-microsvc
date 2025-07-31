package db

import (
	"fmt"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func JSONSetRaw(col, path string, val interface{}) clause.Expr {
	if val == nil {
		val = "{}"
	} else {
		val = fmt.Sprintf(`'%v'`, val)
	}
	return gorm.Expr(fmt.Sprintf(`JSON_SET(%s, '%s', %v)`, col, path, val))
}

// JSONSet key 只能是字符串、数字
func JSONSet(col string, key interface{}, val interface{}) clause.Expr {
	return JSONSetRaw(col, fmt.Sprintf(`$."%v"`, key), val)
}

func JSONRemoveRaw(col, path string) clause.Expr {
	return gorm.Expr(fmt.Sprintf(`JSON_REMOVE(%s, '%s')`, col, path))
}

func JSONRemove(col string, key interface{}) clause.Expr {
	return JSONRemoveRaw(col, fmt.Sprintf(`$."%v"`, key))
}

func JSONContainsRaw(col, val string) clause.Expr {
	return gorm.Expr(fmt.Sprintf(`JSON_CONTAINS(%s, '%s')`, col, val))
}
func JSONContains(col string, val interface{}) clause.Expr {
	return JSONContainsRaw(col, fmt.Sprintf(`"%v"`, val))
}
