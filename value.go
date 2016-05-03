package gosql

import (
    "reflect"
    "strconv"
)

type Values interface {
    Scan(interface{}, ...func(interface{}, Value))
    Result()        []map[string]string
    ResultValue()   []Value
}

type Value interface {
    Scan(interface{}, ...func(interface{}, Value))
    Result()    map[string]string
    Val(string) (string, bool)
    MustVal(string) (string)
}

type Row map[string]string

type Rows []*Row

func (this *Rows) Scan(v interface{}, funcs ...func(interface{}, Value))  {
    rvs := reflect.ValueOf(v)
    if rvs.Kind() != reflect.Ptr || !rvs.IsValid() {
        return
    }
    slice := reflect.MakeSlice(rvs.Elem().Type(), len(*this), len(*this))
    for i,  value := range *this {
        rv := slice.Index(i)
        if rv.Kind() == reflect.Ptr {
            s := reflect.New(rv.Type().Elem())
            value.setValueOf(s)
            rv.Set(s)
        } else {
            value.setValueOf(rv)
        }
        for _, f := range funcs {
            f(rv.Interface(), value)
        }
    }
    rvs.Elem().Set(slice)
}

func (this *Rows) Result() []map[string]string {
    result := make([]map[string]string, len(*this))
    for i, value := range *this {
        result[i] = value.Result()
    }
    return result
}

func (this *Rows) ResultValue() []Value {
    valueAry := make([]Value, len(*this))
    for i, v := range *this {
        valueAry[i] = v
    }
    return valueAry
}

func (this *Row) Scan(v interface{}, funcs ...func(interface{}, Value))  {
    rvs := reflect.ValueOf(v)
    rts := rvs.Type()
    if rts.Kind() != reflect.Ptr || rvs.IsNil() {
        return
    }
    this.setValueOf(rvs)
    for _, f := range funcs {
        f(v, this)
    }
}

func (this *Row) Result() map[string]string {
    return map[string]string(*this)
}

func (this *Row) Val(key string) (string, bool) {
    val, ok := (*this)[key]
    return val, ok
}

func (this *Row) MustVal(key string) string {
    val, _ := (*this)[key]
    return val
}

func (this *Row) setValueOf(rvs reflect.Value) {
    rts := rvs.Type()
    if rvs.Kind() == reflect.Ptr {
        rvs = rvs.Elem()
        rts = rts.Elem()
    }
    for i := 0; i < rvs.NumField(); i++ {
        rv := rvs.Field(i)
        rt := rts.Field(i)
        if rv.CanSet() {
            fieldName := rt.Name
            if tagName := rt.Tag.Get("field"); tagName != "" {
                if tagName == "-" {
                    continue
                }
                fieldName = tagName
            }
            if valueString, ok := (*this)[fieldName]; ok {
                switch rv.Kind() {
                case reflect.String:
                    rv.SetString(valueString)
                case reflect.Int8:
                    v, _ := strconv.Atoi(valueString)
                    rv.Set(reflect.ValueOf(int8(v)))
                case reflect.Int16:
                    v, _ := strconv.Atoi(valueString)
                    rv.Set(reflect.ValueOf(int16(v)))
                case reflect.Int32:
                    v, _ := strconv.Atoi(valueString)
                    rv.Set(reflect.ValueOf(int32(v)))
                case reflect.Int:
                    v, _ := strconv.Atoi(valueString)
                    rv.Set(reflect.ValueOf(v))
                case reflect.Int64:
                    v, _ := strconv.ParseInt(valueString, 10, 64)
                    rv.SetInt(v)
                case reflect.Uint:
                    v, _ := strconv.ParseUint(valueString, 10, 64)
                    rv.Set(reflect.ValueOf(uint(v)))
                case reflect.Uint8:
                    v, _ := strconv.ParseUint(valueString, 10, 8)
                    rv.Set(reflect.ValueOf(uint8(v)))
                case reflect.Uint16:
                    v, _ := strconv.ParseUint(valueString, 10, 16)
                    rv.Set(reflect.ValueOf(uint16(v)))
                case reflect.Uint32:
                    v, _ := strconv.ParseUint(valueString, 10, 32)
                    rv.Set(reflect.ValueOf(uint32(v)))
                case reflect.Uint64:
                    v, _ := strconv.ParseUint(valueString, 10, 64)
                    rv.Set(reflect.ValueOf(uint64(v)))
                case reflect.Bool:
                    v, _ := strconv.ParseBool(valueString)
                    rv.Set(reflect.ValueOf(v))
                case reflect.Float64:
                    v, _ := strconv.ParseFloat(valueString, 64)
                    rv.Set(reflect.ValueOf(v))
                case reflect.Float32:
                    v, _ := strconv.ParseFloat(valueString, 32)
                    rv.Set(reflect.ValueOf(float32(v)))
                }

            }
        }
    }
}