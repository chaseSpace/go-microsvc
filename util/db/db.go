package db

import (
	"errors"
	"microsvc/pkg/xerr"
	"microsvc/protocol/svc/commonpb"
	"regexp"
	"strings"

	"github.com/go-sql-driver/mysql"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

const (
	CommonOmits = "id,created_at,updated_at,deleted_at"
)

func IsMysqlErr(err error) bool {
	return err != nil && !errors.Is(err, gorm.ErrRecordNotFound)
}
func IsRedisErr(err error) bool {
	return err != nil && !errors.Is(err, redis.Nil)
}
func IgnoreNilErr(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) || errors.Is(err, redis.Nil) {
		return nil
	}
	return err
}

func IgnoreTableNotExist(err error) error {
	if IsMysqlTableNotExistErr(err) {
		return nil
	}
	return err
}

/* mysql 常见错误码
Access Denied (1045): 这是一个常见的错误，表示连接到 MySQL 服务器时权限被拒绝。错误代码为 1045。你可以在 MySQL 官方文档中查找这个错误的信息，了解如何解决权限问题。

Table doesn't exist (1146): 当你尝试查询或操作一个不存在的表时，会遇到这个错误。错误代码为 1146。通常需要检查表名是否正确或确保表已经创建。

Duplicate entry (1062): 这个错误表示尝试插入重复的唯一键值。错误代码为 1062。你可以检查你的数据，或者使用 INSERT IGNORE 或 INSERT ... ON DUPLICATE KEY UPDATE 来处理这种情况。

Syntax error (1064): 这个错误表示 SQL 语法错误。错误代码为 1064。你需要检查 SQL 查询或语句的语法，确保它是有效的。

Lock wait timeout exceeded (1205): 当某个事务等待获取锁的时间超过设置的超时时间时，会发生这个错误。错误代码为 1205。你可以尝试增加超时时间，优化查询，或查看锁的情况。

Lost connection to MySQL server (2013): 当与 MySQL 服务器的连接丢失时，会发生这个错误。错误代码为 2013。这可能是由于网络问题或服务器崩溃引起的。你可以检查网络连接或服务器状态。

Data too long for column (1406): 当尝试插入的数据超过了列的最大长度时，会发生这个错误。错误代码为 1406。你需要检查数据的长度，并根据需要调整列的长度。
*/

var mysqlDupRegex = regexp.MustCompile(`Duplicate entry '(.*?)' for key '(\w+?)'`)

func ParseMysqlDuplicate(errStr string) (entry, uk string) {
	ss := mysqlDupRegex.FindStringSubmatch(errStr)
	if len(ss) > 0 {
		return ss[1], ss[2]
	}
	return
}

// Deprecated: 使用 xerr.ErrDataDuplicate
// 判断Duplicate err并且提取索引名
func IsMysqlDuplicateErr(err error, forIndex *string) bool {
	if err == nil {
		return false
	}
	var err2 *mysql.MySQLError
	if errors.As(err, &err2) && err2.Number == 1062 {
		ss := strings.Split(err.Error(), "for key ")
		if len(ss) == 2 && forIndex != nil {
			*forIndex = ss[1]
		}
		return true
	}
	return false
}

func IsMysqlTableNotExistErr(err error) bool {
	if err == nil {
		return false
	}
	var err2 *mysql.MySQLError
	if errors.As(err, &err2) && err2.Number == 1146 {
		if strings.HasSuffix(err.Error(), "doesn't exist") {
			return true
		}
	}
	return false
}

func __paging(model *gorm.DB, page *commonpb.PageArgs) *gorm.DB {
	return model.Offset(int((page.Pn - 1) * page.Ps)).Limit(int(page.Ps))
}

type PageQueryArgs struct {
	SelectFields string // "a,b"
	OmitFields   string // "a,b"
}

func PageQuery(model *gorm.DB, page *commonpb.PageArgs, orderBy string, total *int64, list interface{}, args ...PageQueryArgs) error {
	err := model.Count(total).Error
	if err != nil {
		return err
	}
	if *total == 0 {
		return nil
	}
	if len(args) > 0 {
		if args[0].SelectFields != "" {
			model = model.Select(args[0].SelectFields)
		}
		if args[0].OmitFields != "" {
			model = model.Omit(args[0].OmitFields)
		}
	}
	// TODO 解决日志输出中打印的代码行不对应“实际”调用行的问题
	err = __paging(model.Order(orderBy), page).Scan(list).Error
	return err
}

// CreateTable 统一管理建表语句
func CreateTable(tx *gorm.DB, sql string) error {
	if !strings.HasPrefix(sql, "create table if not exists") {
		return errors.New("not a create table SQL: " + sql)
	}
	return xerr.WrapMySQL(tx.Exec(sql).Error)
}

type TableHelper struct {
	tx        *gorm.DB
	createSQL string
}

func NewTableHelper(tx *gorm.DB, createSQL string) *TableHelper {
	return &TableHelper{
		tx:        tx,
		createSQL: createSQL,
	}
}

func (t *TableHelper) AutoCreateTable(operation func(tx *gorm.DB) error) error {
	err := operation(t.tx)
	if IsMysqlTableNotExistErr(err) {
		if err := CreateTable(t.tx, t.createSQL); err != nil {
			return err
		}
		return xerr.WrapMySQL(operation(t.tx))
	}
	return err
}

type Sort interface {
	GetOrderField() string
	GetOrderType() commonpb.OrderType
}

type OrderFieldMap map[string]*struct{}

var orderTypMap = map[commonpb.OrderType]string{
	commonpb.OrderType_OT_Asc:  "ASC",
	commonpb.OrderType_OT_Desc: "DESC",
}

func GenSortClause[T Sort](sort []T, fieldMap OrderFieldMap, appendIdDesc ...bool) (clause string, err error) {
	hasId := false
	for _, s := range sort {
		if fieldMap != nil && fieldMap[s.GetOrderField()] == nil {
			return "", xerr.ErrParams.New("Not support order field: " + s.GetOrderField())
		}
		if s.GetOrderField() == "id" {
			hasId = true
		}
		clause += s.GetOrderField() + " " + orderTypMap[s.GetOrderType()] + ", "
	}
	if len(appendIdDesc) > 0 && appendIdDesc[0] && !hasId {
		clause += "id DESC, "
	}
	return strings.TrimRight(clause, ", "), nil
}

func IdAscFn() *commonpb.Sort {
	return &commonpb.Sort{
		OrderField: "id",
		OrderType:  commonpb.OrderType_OT_Asc,
	}
}

func IdDescFn() *commonpb.Sort {
	return &commonpb.Sort{
		OrderField: "id",
		OrderType:  commonpb.OrderType_OT_Desc,
	}
}

func Count(gdb *gorm.DB) (ct int64) {
	gdb.Count(&ct)
	return
}

func LikeWrap(s string) string {
	return "%" + s + "%"
}

func GetAffected(g *gorm.DB) (affected int64, err error) {
	return g.RowsAffected, g.Error
}
func GetBoolAffected(g *gorm.DB) (hits bool, err error) {
	return g.RowsAffected > 0, g.Error
}
