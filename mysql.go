package gosql

import (
    _ "github.com/go-sql-driver/mysql"
    "database/sql"
    "fmt"
    "log"
    "strings"
    "strconv"
    "time"
    "sync"
    "reflect"
)

var MysqlDbMap map[string]*homeDB
var connectLock sync.Mutex

type homeDB struct {
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
type ExecSql struct {
    SQL		string
    Args	interface{}
}

/**
 * Mysql数据库操作.
 */
type DbMysql struct {
    DB				*homeDB
    LastSql			*ExecSql
    Info			map[int]interface{}
    tableName		string
    wTableName		string
    selectFields	string
    useFieldMap  map[string]bool
    notUseFieldMap  map[string]bool
    where 			map[string]interface{}
    condition		*Condition
    group			string
    having			map[string]interface{}
    order			string
    limit			string
    join			string
    noClear			bool
}

func (this *homeDB) FreshLastUseTime() {
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
    this.Info[G_MaxOpenConns] = 0
    this.Info[G_MaxIdleConns] = 0
    return this
}

// 数据库连接
func (this *DbMysql) Connect() error {
    if MysqlDbMap == nil {
        MysqlDbMap = make(map[string]*homeDB)
    }
    if this.DB != nil && !this.DB.IsClose {
        return nil
    }
    dataSourceName := fmt.Sprintf("%s:%s@(%s:%d)/%s", this.Info[G_User], this.Info[G_Password], this.Info[G_Host], this.Info[G_Port], this.Info[G_DbName])

    if this.connectOnly(dataSourceName) {
        // 用现有资源建立连接成功
        return nil
    }
    // 锁住,然后创建连接
    connectLock.Lock()
    defer connectLock.Unlock()
    if this.connectOnly(dataSourceName) {
        // 如果同事还有其他携程创建连接成功了
        return nil
    }
    db, err := sql.Open("mysql", dataSourceName)
    if err != nil {
        return err
    }
    if err := db.Ping(); err != nil {
        return err
    }
    homeDB := &homeDB{DB: db, LastUseTime: time.Now(), IsClose: false}
    if conns, is := this.Info[G_MaxOpenConns].(int); is && conns > 0 {
        homeDB.SetMaxOpenConns(conns)
    }
    if conns, is := this.Info[G_MaxIdleConns].(int); is && conns > 0 {
        homeDB.SetMaxIdleConns(conns)
    }
    if t, is := this.Info[G_AutoCloseTime].(int); is {
        homeDB.AutoCloseTime = t
    }
    MysqlDbMap[dataSourceName] = homeDB
    this.DB = homeDB
    return nil
}

// 使用有已有的连接资源
func (this *DbMysql) connectOnly(dataSourceName string) bool {
    if homeDB, ok := MysqlDbMap[dataSourceName]; ok {
        homeDB.FreshLastUseTime()
        this.DB = homeDB
        return true
    }
    return false
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

func (this *DbMysql) Where(condition map[string]interface{}) *DbMysql {
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

func (this *DbMysql) Having(condition map[string]interface{}) *DbMysql {
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
    if v, ok := result.Result()["num"]; ok {
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
    if v, ok := result.Result()["num"]; ok {
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
    results := ary.ResultValue()
    if len(results) > 0 {
        result := results[0]
        return result, nil
    } else {
        return nil, nil
    }
}

// 设置更新(插入)的字段 逗号隔开
func (this *DbMysql) SetUseField(s string) *DbMysql {
    if this.useFieldMap == nil {
        this.useFieldMap = map[string]bool{}
    }
    for _, v := range strings.Split(s, ",") {
        if v = strings.TrimSpace(v); v != "" {
            this.useFieldMap[v] = true
        }
    }
    return this
}

// 设置不更新(插入)的字段 逗号隔开
func (this *DbMysql) SetNotUseField(s string) *DbMysql {
    if this.useFieldMap == nil {
        this.notUseFieldMap = map[string]bool{}
    }
    for _, v := range strings.Split(s, ",") {
        if v = strings.TrimSpace(v); v != "" {
            this.notUseFieldMap[v] = true
        }
    }
    return this
}

// data 可以是map[string]string   也可以是一个struct 必须指明tag field 才会更新
// struct {
//     Id      int8         `field:"id"`
//     Name    string       `field:"name"`
//     Name2   float64      `field:"name2"`
// }
func (this *DbMysql) Update(data interface{}) (int, error) {
    sql, exeArgs := this.getUpdateSql(this.GetUseMap(data))
    return this.Exec(sql, exeArgs...)
}

func (this *DbMysql) Insert(data interface{}) (int, error) {
    sql, exeArgs := this.getInsertSql(this.GetUseMap(data))
    return this.Exec(sql, exeArgs...)
}

func (this *DbMysql) InsertIgnore(data interface{}) (int, error) {
    sql, exeArgs := this.getInsertSql(this.GetUseMap(data))
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
    lastSql := &ExecSql{sql, args}
    logWrite(*lastSql)
    this.setLastSql(lastSql)
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
    lastSql := &ExecSql{sql, args}
    logWrite(*lastSql)
    this.setLastSql(lastSql)
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

func (this *DbMysql) GetUseMap(data interface{}) map[string]string {
    dataMap := toString(data)
    for k, _ := range dataMap {
        if this.useFieldMap != nil {
            // 如果不在需要更新的字段中则剔除
            if _, ok := this.useFieldMap[k]; !ok {
                delete(dataMap, k)
            }
        }
        if this.notUseFieldMap != nil {
            // 如果在不需要更新的字段中则剔除
            if _, ok := this.notUseFieldMap[k]; ok {
                delete(dataMap, k)
            }
        }
    }
    return dataMap
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
    if !isCount && this.order != "" {
        sql += " order by " + this.order
    }
    if !isCount && this.limit != "" {
        sql += " limit " + this.limit
    }
    return sql, append(exeArgs, havingExeArgs...)
}

func (this *DbMysql) getWhereSql(condition map[string]interface{}) (string, []interface{}) {
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

func (this *DbMysql) getHavingSql(condition map[string]interface{}) (string, []interface{}) {
    whereSql, args := this.parseSql(condition)
    return whereSql, args
}


func (this *DbMysql) parseSql(condition map[string]interface{}) (string, []interface{}) {
    args := []interface{}{}
    whereAry := []string{}
    var tmpAryKey []string
    for k, v := range condition {
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
    result := &Rows{}
    for len := 0; rows.Next(); len++ {
        err = rows.Scan(scanArgs...)
        if err != nil {
            return nil, err
        }
        record := &Row{}
        for i, col := range values {
            if col != nil {
                (*record)[columns[i]] = string(col)
            }
        }
        *result = append(*result, record)
    }
    return result, nil
}

func (this *DbMysql) setLastSql(lastSql *ExecSql) {
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
    this.where = map[string]interface{}{}
    this.group = ""
    this.having = map[string]interface{}{}
    this.join = ""
    this.order = ""
    this.limit = ""
    this.condition = nil
    this.useFieldMap = nil
    this.notUseFieldMap = nil
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
    sql := ""
    if strings.Index(key, "`") != -1 || strings.Index(key, ".") != -1 {
        sql += key
    } else {
        sql += "`" + key + "`"
    }
    args := []interface{}{}
    switch val.(type) {
    default:
        sql += " " + ex + " ?"
        args = append(args, val)
    case []interface{}:
        sql += " in ("
        strAry := val.([]interface{})
        for i, v := range strAry {
            if i == 0 {
                sql += "?"
            } else {
                sql += ",?"
            }
            args = append(args, v)
        }
        sql += ")"
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
        strAry := val.([]int)
        for i, v := range strAry {
            if i == 0 {
                sql += "?"
            } else {
                sql += ",?"
            }
            args = append(args, v)
        }
        sql += ")"
    case []int64:
        sql += " in ("
        strAry := val.([]int64)
        for i, v := range strAry {
            if i == 0 {
                sql += "?"
            } else {
                sql += ",?"
            }
            args = append(args, v)
        }
        sql += ")"
    case []float64:
        sql += " in ("
        strAry := val.([]float64)
        for i, v := range strAry {
            if i == 0 {
                sql += "?"
            } else {
                sql += ",?"
            }
            args = append(args, v)
        }
        sql += ")"
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

func toString(data interface{}) map[string]string {
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
    if tmp, ok := data.(map[string]string); ok {
        return tmp
    } else if tmp, ok := data.(map[string]interface{}); ok {
        for k, v := range tmp {
            val[k] = f(v)
        }
    } else {
        v := reflect.ValueOf(data)
        if !v.IsValid() {
            return val
        }
        if v.Kind() == reflect.Ptr {
            v = v.Elem()
        }
        if !v.IsValid() {
            return val
        }
        t := v.Type()
        for i := 0; i < v.NumField(); i++ {
            fv := v.Field(i)
            ft := t.Field(i)
            field := ft.Tag.Get("field")
            if field != "" && field != "-"{
                switch ft.Type.Kind() {
                case reflect.Int, reflect.Int64, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
                    v := fv.Int()
                    val[field] = strconv.FormatInt(v, 10)
                case reflect.Float64, reflect.Float32:
                    v := fv.Float()
                    val[field] = strconv.FormatFloat(v, 'f', -1, 64)
                case reflect.String:
                    val[field] = fv.String()
                case reflect.Bool:
                    if fv.Bool() {
                        val[field] = "1"
                    } else {
                        val[field] = "0"
                    }
                }
            }
        }

    }
    return val
}
