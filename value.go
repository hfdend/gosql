package gosql

import (
    "reflect"
    "strconv"
)

type Value map[string]string

type Values []Value

func (this *Values) Scan(v interface{})  {
    rvs := reflect.ValueOf(v)
    if rvs.Kind() != reflect.Ptr || !rvs.IsValid() {
        return
    }
    slice := reflect.MakeSlice(rvs.Type().Elem(), len(*this), len(*this))
    for i,  value := range *this {
        rv := slice.Index(i)
        value.SetValueOf(rv)
    }
    rvs.Elem().Set(slice)
}

func (this *Value) Scan(v interface{})  {
    rvs := reflect.ValueOf(v)
    rts := rvs.Type()
    if rts.Kind() != reflect.Ptr || rvs.IsNil() {
        return
    }
    this.SetValueOf(rvs)
}

func (this *Value) SetValueOf(rvs reflect.Value) {
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
                case reflect.Bool:
                    v, _ := strconv.ParseBool(valueString)
                    rv.Set(reflect.ValueOf(v))
                }
            }
        }
    }
}