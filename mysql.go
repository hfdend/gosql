package gosql

import (
	_ "github.com/go-sql-driver/mysql"
	"database/sql"
	"fmt"
	"log"
	"strings"
	"strconv"
	"errors"
	"time"
)

var MysqlDbMap map[string]*DB

type DB struct {
	*sql.DB
	LastUseTime		time.Time
    IsClose         bool
    AutoCloseTime   int
}

type Pager struct {
	Limit 			int // 每页条数
	Count			int // 总条数
	Offset			int // 偏移量
	IsSubqueries	bool // 是否使用子查询查询总数
}

// 最后一次执行的SQL
type LastSql struct {
	SQL		string
	Args	interface{}
}

/**
 * Mysql数据库操作.
 */
type DbMysql struct {
	DB				*DB
	LastSql			*LastSql
	Info			map[int]interface{}
	tableName		string
	wTableName		string
	selectFields	string
	where 			map[string]string
	condition		*Condition
	group			string
	having			map[string]string
	order			string
	limit			string
	join			string
	noClear			bool
}

func (this *DB) FreshLastUseTime() {
	this.LastUseTime = time.Now()
}

const (
    G_Host = iota
    G_User
    G_Password
    G_Port
    G_DbName
    G_AutoCloseTime
    G_MaxOpenConns
    G_MaxIdleConns
)

func NewDbMysql(host string, port int, user string, password string, dbname string) *DbMysql {
    this := new(DbMysql)
    this.Info = make(map[int]interface{})
    this.Info[G_Host] = host
    this.Info[G_User] = user
    this.Info[G_Password] = password
    this.Info[G_Port] = port
    this.Info[G_DbName] = dbname
    this.Info[G_AutoCloseTime] = 100
    this.Info[G_MaxOpenConns] = 30
    this.Info[G_MaxIdleConns] = 0
    return this
}

// 数据库连接
func (this *DbMysql) Connect() error {
	if MysqlDbMap == nil {
		MysqlDbMap = make(map[string]*DB)
	}
	if this.DB != nil && !this.DB.IsClose {
		return nil
	}
	dataSourceName := fmt.Sprintf("%s:%s@(%s:%d)/%s", this.Info[G_User], this.Info[G_Password], this.Info[G_Host], this.Info[G_Port], this.Info[G_DbName])
	if DB, ok := MysqlDbMap[dataSourceName]; ok {
		DB.FreshLastUseTime()
		this.DB = DB
		return nil
	}
	db, err := sql.Open("mysql", dataSourceName)
	if err != nil {
		return nil
	}
	err = db.Ping()
	if err != nil {
		return err
	}
	DB := &DB{DB: db, LastUseTime: time.Now(), IsClose: false}
	DB.SetMaxOpenConns(this.Info[G_MaxOpenConns].(int))
    DB.SetMaxIdleConns(this.Info[G_MaxIdleConns].(int))
    DB.AutoCloseTime = this.Info[G_AutoCloseTime].(int)
	MysqlDbMap[dataSourceName] = DB
	this.DB = DB
    return nil
}

func (this *DbMysql) Ping() error {
    return this.DB.Ping()
}

// 设置连接自动关闭时间,如果一个连接在t秒内没有任何操作将会被自动关闭掉
func (this *DbMysql) SetAutoCloseTime(t int) {
    this.Info[G_AutoCloseTime] = t
}

func (this *DbMysql) SetMaxIdleConns(t int) {
    this.Info[G_MaxIdleConns] = t
}

func (this *DbMysql) SetMaxOpenConns(t int) {
    this.Info[G_MaxOpenConns] = t
}

func (this *DbMysql) Conf(host, port int, user, password string, dbname string) {
	this.Info = make(map[int]interface{})
	this.Info[G_Host] = host
	this.Info[G_User] = user
	this.Info[G_Password] = password
	this.Info[G_Port] = port
	this.Info[G_DbName] = dbname
}

/**
 * 设置表名
 */
func (this *DbMysql) SetTableName(tableName string) *DbMysql {
	if strings.Index(tableName, "`") == -1 && strings.Index(tableName, ".") == -1 && strings.Index(tableName, " ") == -1 {
		tableName = "`" + tableName + "`"
	}
	this.tableName = tableName
	this.wTableName = tableName
	return this
}

func (this *DbMysql) From(tableName string) *DbMysql {
	if strings.Index(tableName, "`") == -1 && strings.Index(tableName, ".") == -1 && strings.Index(tableName, " ") == -1 {
		tableName = "`" + tableName + "`"
	}
	this.wTableName = tableName
	return this
}

func (this *DbMysql) InnerJoin(tableName, on string) *DbMysql {
	this.join += " inner join " + tableName + " on " + on
	return this
}

func (this *DbMysql) LeftJoin(tableName, on string) *DbMysql {
	this.join += " left join " + tableName + " on " + on
	return this
}

func (this *DbMysql) Where(condition map[string] string) *DbMysql {
	this.where = condition
	return this
}

func (this *DbMysql) SetCondition(condition *Condition) *DbMysql {
	this.condition = condition
	return this
}

func (this *DbMysql) Close()  {
	this.DB = nil
}

/**
 * 设置查询字段
 */
func (this *DbMysql) Select(args... string) *DbMysql {
	fields := []string{}
	for _, v := range(args) {
		ary := strings.Split(v, ",")
		fields = append(fields, ary...)
	}
	for k, v := range(fields) {
		v = strings.Trim(v, " ")
		fields[k] = v
	}
	selectFields := strings.Join(fields, ",")
	this.selectFields = selectFields
	return this
}

func (this *DbMysql) Group(field string) *DbMysql {
	this.group = field
	return this
}

func (this *DbMysql) Having(condition map[string]string) *DbMysql {
	this.having = condition
	return this
}

func (this *DbMysql) Order(order string) *DbMysql {
	this.order = order
	return this
}

func (this *DbMysql) Limit(limit... int) *DbMysql {
	tmp := make([]string, len(limit))
	for i, v := range(limit) {
		tmp[i] = strconv.Itoa(v)
	}
	this.limit = strings.Join(tmp, ",")
	return this
}

func (this *DbMysql) PagerFindAll(p *Pager) (Values, error) {
	this.noClear = true
    var err error
	if p.IsSubqueries {
		p.Count, err = this.CountUseSubqueries()
	} else {
		p.Count, err = this.Count()
	}
	this.noClear = false
	this.Limit(p.Offset, p.Limit)
    if err != nil {
        return nil, err
    }
	return this.FindAll()
}

func (this *DbMysql) FindAll() (Values, error) {
	sql, exeArgs := this.getSelectSql(false)
	result, err := this.Query(sql, exeArgs...)
    if err != nil {
        return nil, err
    }
	return result, nil
}

// 使用子查询统计长度
func (this *DbMysql) CountUseSubqueries() (int, error) {
	num := 0
	sql, exeArgs := this.getCountSql(true)
	result, err := this.QueryOne(sql, exeArgs...)
    if err != nil {
        return num, err
    }
	if v, ok := result["num"]; ok {
		num, _ = strconv.Atoi(v)
	}
	return num, nil
}

func (this *DbMysql) Count() (int, error) {
	num := 0
	sql, exeArgs := this.getCountSql(false)
	result, err := this.QueryOne(sql, exeArgs...)
    if err != nil {
        return num, err
    }
	if v, ok := result["num"]; ok {
		num, _ = strconv.Atoi(v)
	}
	return num, nil
}

func (this *DbMysql) FindOne() (Value, error) {
    this.Limit(1)
	sql, exeArgs := this.getSelectSql(false)
	result, err := this.QueryOne(sql, exeArgs...)
    if err != nil {
        return nil, err
    }
	return result, nil
}

func (this *DbMysql) QueryOne(sql string, args ...interface{}) (Value, error) {
	ary, err := this.Query(sql, args...)
    if err != nil {
        return nil, err
    }
	result := Value{}
	if len(ary) > 0 {
		result = ary[0]
        return result, nil
	} else {
        return nil, nil
    }
}

func (this *DbMysql) Update(data map[string]interface{}) (int, error) {
	sql, exeArgs := this.getUpdateSql(toString(data))
	return this.Exec(sql, exeArgs...)
}

func (this *DbMysql) Insert(data map[string]interface{}) (int, error) {
	sql, exeArgs := this.getInsertSql(toString(data))
	return this.Exec(sql, exeArgs...)
}

func (this *DbMysql) InsertIgnore(data map[string]interface{}) (int, error) {
	sql, exeArgs := this.getInsertSql(toString(data))
	sql = strings.Replace(sql, "insert into", "insert ignore into", -1)
	return this.Exec(sql, exeArgs...)
}

func (this *DbMysql) Del() (int, error) {
	sql, exeArgs := this.getDelSql()
	return this.Exec(sql, exeArgs...)
}

func (this *DbMysql) LockWriteTable(table string) error {
	sql := "lock table `" + table + "` write"
	_, err := this.Exec(sql)
    return err
}

func (this *DbMysql) LockReadTable(table string) error {
	sql := "lock table `" + table + "` read"
    _, err := this.Exec(sql)
    return err
}

func (this *DbMysql) UnLockTable() error {
	sql := "unlock tables"
    _, err := this.Exec(sql)
    return err
}

func (this *DbMysql) Query(sql string, args ...interface{}) (Values, error) {
	if err := this.Connect(); err != nil {
        return nil, err
    }
	this.DB.FreshLastUseTime()
	this.setLastSql(sql, args)
	rows, err := this.DB.Query(sql, args...)
	defer func() {
		if rows != nil {
			rows.Close()
		}
	}()
	if err != nil {
		return nil, err
	}
	ary, err := this.rowsToAry(rows)
	this.clear()
	return ary, nil
}

func (this *DbMysql) Exec(sql string, args ...interface{}) (int, error) {
	if err := this.Connect(); err != nil {
        return 0, err
    }
	this.DB.FreshLastUseTime()
	this.setLastSql(sql, args)
	res, err := this.DB.Exec(sql, args...)
	if err != nil {
		return 0, err
	}
	var row int64
	row, err = res.LastInsertId()
	if err != nil {
		return 0, err
	}
	if row == 0 {
		row, _ = res.RowsAffected()
	}
	this.clear()
	return int(row), nil
}

func (this *DbMysql) getUpdateSql(data map[string]string) (string, []interface{}) {
	sql := "update " + this.wTableName + " set"
	exeArgs := []interface{}{}
	for k, v := range(data) {
		sql += " `" + k + "` = ?,"
		exeArgs = append(exeArgs, v)
	}
	sql = string([]byte(sql)[:len(sql) - 1])
	whereSql, whereExeArgs := this.getWhereSql(this.where)
	if whereSql != "" {
		sql += " where " + whereSql
	}
	exeArgs = append(exeArgs, whereExeArgs...)
	return sql, exeArgs
}

func (this *DbMysql) getInsertSql(data map[string]string) (string, []interface{}) {
	exeArgs := []interface{}{}
	fields := []string{}
	values := []string{}
	for k, v := range(data) {
		fields = append(fields, k)
		exeArgs = append(exeArgs, v)
		values = append(values, "?")
	}
	sql := "insert into " + this.wTableName + " (`" + strings.Join(fields, "`,`") + "`) values (" + strings.Join(values, ",") + ")"
	return sql, exeArgs;
}

func (this *DbMysql) getDelSql() (string, []interface{}) {
	whereSql, whereExeArgs := this.getWhereSql(this.where)
	if whereSql == "" {
		log.Println("删除没有条件")
	}
	sql := "delete from " + this.wTableName + " where " + whereSql
	return sql, whereExeArgs
}

// 是否启用子查询查询总条数
func (this *DbMysql) getCountSql(isSubqueries bool) (string, []interface{}) {
	if (isSubqueries) {
		sql, exeArgs := this.getSelectSql(false)
		sql = "select count(1) as num from (" + sql + ") as tb"
		return sql, exeArgs
	}
	sql, exeArgs := this.getSelectSql(true)
	return sql, exeArgs
}

// 得到条件查询sql
// isCount 是否是查询总数的sql
func (this *DbMysql) getSelectSql(isCount bool) (string, []interface{}) {
	s := "*"
	if this.selectFields != "" {
		s = this.selectFields
	}
	if isCount {
		s = "count(1) as num"
	}
	table := this.wTableName
	whereSql, exeArgs := this.getWhereSql(this.where)
	havingWhereSql, havingExeArgs := this.getHavingSql(this.having)
	sql := "select " + s + " from " + table;
	if this.join != "" {
		sql += " " + this.join + " "
	}
	if whereSql != "" {
		sql += " where " + whereSql
	}
	if this.group != "" {
		sql += " group by " + this.group
	}
	if havingWhereSql != "" {
		sql += " having " + havingWhereSql
	}
	if this.order != "" {
		sql += " order by " + this.order
	}
	if this.limit != "" {
		sql += " limit " + this.limit
	}
	return sql, append(exeArgs, havingExeArgs...)
}

func (this *DbMysql) getWhereSql(condition map[string]string) (string, []interface{}) {
	whereSql, args := this.parseSql(condition)
	if this.condition != nil {
		if where, ary := this.condition.GetSql(); where != "" {
			if whereSql == "" {
				whereSql = where
			} else {
				whereSql += " and " + where
			}
			args = append(args, ary...)
		}
	}
	return whereSql, args
}

func (this *DbMysql) getHavingSql(condition map[string]string) (string, []interface{}) {
	whereSql, args := this.parseSql(condition)
	return whereSql, args
}


func (this *DbMysql) parseSql(condition map[string]string) (string, []interface{}) {
	args := []interface{}{}
	whereAry := []string{}
	var tmpAryKey []string
	for k, v := range(condition) {
		k = strings.Trim(k, " ")
		tmpAryKey = []string{}
		for _, vv := range(strings.Split(k, " ")) {
			if vv != "" {
				tmpAryKey = append(tmpAryKey, vv)
			}
		}
		if len(tmpAryKey) == 1 {
			whereAry = append(whereAry, tmpAryKey[0] + " = ?")
		} else {
			whereAry = append(whereAry, tmpAryKey[0] + " " + tmpAryKey[1] + " ?")
		}
		args = append(args, v)
	}
	whereSql := strings.Join(whereAry, " and ")

	return whereSql, args
}

func (this *DbMysql) rowsToAry(rows *sql.Rows) (Values, error) {
	columns, err := rows.Columns()
	defer rows.Close()
	if err != nil {
		return nil, err
	}
	scanArgs := make([]interface{}, len(columns))
	values := make([][]byte, len(columns))
	for i := range values {
	    scanArgs[i] = &values[i]
	}
	result := Values{}
	for len := 0; rows.Next(); len++ {
		err = rows.Scan(scanArgs...)
		if err != nil {
			return nil, err
		}
		record := make(Value)
		for i, col := range values {
	        if col != nil {
	            record[columns[i]] = string(col)
	        }
	    }
        result = append(result, record)
	}
	return result, nil
}

func (this *DbMysql) setLastSql(sql string, args interface{}) {
	lastSql := &LastSql{sql, args}
	this.LastSql = lastSql
}

func (this *DbMysql) NewCondition() *Condition {
	return NewCondition()
}

func (this *DbMysql) NewPager() *Pager {
	return NewPager()
}

func (this *DbMysql) clear() {
	if this.noClear {
		return
	}
	this.wTableName = this.tableName
	this.selectFields = ""
	this.where = map[string]string{}
	this.group = ""
	this.having = map[string]string{}
	this.join = ""
	this.order = ""
	this.limit = ""
	this.condition = nil
}

type Condition struct {
	whereSql	string
	args		[]interface{}
}

// 设置搜索条件
// key 字段名称
// ex  判断表达式 可以是 = , >, >=, <, <=, !=
// val 值, 如果是int或者string则表示等于; 如果是[]int 或者 []string 则表示in查询 其他类型不支持(如果ex不等于"=" 那么仅仅支持int 和 string)
func (this *Condition) SetFilterEx(key string, ex string, val interface{}) error {
	if no := ex == "=" || ex == ">" || ex == ">=" || ex == "<" || ex == "<=" || ex == "!=" || ex == "like"; !no  {
		err := errors.New("ex 参数错误只能是 = | > | >= | < | <= | in | like")
		return err
	}
	sql := ""
	if strings.Index(key, "`") != -1 || strings.Index(key, ".") != -1 {
		sql += key
	} else {
		sql += "`" + key + "`"
	}
	args := []interface{}{}
	switch val.(type) {
	case int, string:
		sql += " " + ex + " ?"
		args = append(args, val)
	case []string:
		sql += " in ("
		strAry := val.([]string)
		for i, v := range strAry {
			if i == 0 {
				sql += "?"
			} else {
				sql += ",?"
			}
			args = append(args, v)
		}
		sql += ")"
	case []int:
		sql += " in ("
		intAry := val.([]int)
		for i, v := range intAry {
			if i == 0 {
				sql += "?"
			} else {
				sql += ",?"
			}
			args = append(args, v)
		}
		sql += ")"
	default:
		err := errors.New("SetFilter 第二个参数错误,只能是int, string, []int, []string 四种类型")
		return err
	}
	if len(this.whereSql) == 0 {
		this.whereSql = sql
	} else {
		this.whereSql += " and " + sql
	}
	if len(args) != 0 {
		this.args = append(this.args, args...)
	}
    return nil
}

// 设置搜索条件
// key 字段名称
// val 值, 如果是int或者string则表示等于; 如果是[]int 或者 []string 则表示in查询 其他类型不支持
func (this *Condition) SetFilter(key string, val interface{}) *Condition {
	this.SetFilterEx(key, "=", val)
    return this
}

func (this *Condition) SetFilterOr(conditions ...*Condition) {
	for _, condition := range conditions {
		sql, args := condition.GetSql()
		if len(this.whereSql) == 0 {
			this.whereSql = "(" + sql + ")"
		} else {
			this.whereSql += " or (" + sql + ")"
		}
		this.args = append(this.args, args...)
	}
}

func (this *Condition) GetSql() (string, []interface{}) {
	return this.whereSql, this.args
}

func NewCondition () *Condition {
	return new(Condition)
}

func NewPager() *Pager {
	p := new(Pager)
	p.Limit = 20
	return p
}

func toString(data map[string]interface{}) map[string]string {
    val := map[string]string{}
    var f = func(v interface{}) string {
        s := ""
        switch v.(type) {
        case string:
            s = v.(string)
        case int:
            s = strconv.Itoa(v.(int))
        case int8:
            s = strconv.Itoa(int(v.(int8)))
        case int16:
            s = strconv.Itoa(int(v.(int16)))
        case int32:
            s = strconv.Itoa(int(v.(int32)))
        case int64:
            s = strconv.FormatInt(v.(int64), 10)
        }
        return s
    }
    for k, v := range data {
        val[k] = f(v)
    }
    return val
}
