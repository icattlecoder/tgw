//http.Request解析
package tgw

import (
	"reflect"
	"strconv"
	"strings"
)

type RegisterParse interface {
	Parse(env *ReqEnv, typ reflect.Type) (reflect.Value, bool)
}

type EnvParse struct{}

func (e *EnvParse) Parse(env *ReqEnv, typ reflect.Type) (vl reflect.Value, parsed bool) {

	if typ != reflect.TypeOf(env).Elem() {
		return
	}
	parsed = true
	vl = reflect.ValueOf(*env)
	return
}

//解析名称符合`*Args`的结构体

type ArgsParse struct{}

func (d *ArgsParse) Parse(env *ReqEnv, typ reflect.Type) (vl reflect.Value, parsed bool) {

	i := strings.LastIndex(typ.Name(), "Args")
	if len(typ.Name())-i != 4 {
		return
	}
	parsed = true
	vl = reflect.New(typ).Elem()

	req := env.Req
	get := func(key string) string { return req.URL.Query().Get(key) }
	post := func(key string) string { return req.FormValue(key) }
	fun := (func() func(string) string {
		if req.Method == "POST" {
			req.ParseForm()
			return post
		}
		return get
	})()

	for i := 0; i < typ.NumField(); i++ {
		qryName := typ.Field(i).Name
		qryName = strings.ToLower(qryName)
		qry := fun(qryName)
		tn := typ.Field(i).Type.Name()
		switch tn {
		case "string":
			vl.Field(i).SetString(qry)
		case "int":
			if ival, err := strconv.Atoi(qry); err == nil {
				vl.Field(i).SetInt(int64(ival))
			}
		case "bool":
			if ival, err := strconv.ParseBool(qry); err == nil {
				vl.Field(i).SetBool(ival)
			}
		case "float64":
			if ival, err := strconv.ParseFloat(qry, 64); err == nil {
				vl.Field(i).SetFloat(ival)
			}
		}
	}
	return
}
