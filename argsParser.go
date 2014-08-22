//http.Request解析
package tgw

import (
	"encoding/json"
	"encoding/xml"
	"github.com/qiniu/log"
	"reflect"
	"strconv"
	"strings"
)

type RegisterParser interface {
	Parse(env *ReqEnv, typ reflect.Type) (reflect.Value, bool)
}

var parserCache = map[reflect.Type]RegisterParser{}

type EnvParse struct{}

func (e *EnvParse) Parse(env *ReqEnv, typ reflect.Type) (vl reflect.Value, parsed bool) {

	if typ != reflect.TypeOf(env).Elem() {
		return
	}
	parsed = true
	vl = reflect.ValueOf(*env)
	return
}

//解释参数经由request.Body传过来的Json形式的参数
type ArgsJson struct {
}

func (r *ArgsJson) Parse(env *ReqEnv, typ reflect.Type) (vl reflect.Value, parsed bool) {

	parsed = strings.HasSuffix(typ.Name(), "ArgsJson")
	if !parsed {
		return
	}
	if _, ok := parserCache[typ]; !ok {
		parserCache[typ] = r
	}
	vl2 := reflect.New(typ)

	decoder := json.NewDecoder(env.Req.Body)

	if err := decoder.Decode(vl2.Interface()); err != nil {
		vl = reflect.ValueOf(vl2.Interface()).Elem()
		log.Error(err)
		return
	}
	vl = reflect.ValueOf(vl2.Interface()).Elem()
	return
}

//解释参数经由request.Body传过来的Xml形式的参数
type ArgsXml struct {
}

func (r *ArgsXml) Parse(env *ReqEnv, typ reflect.Type) (vl reflect.Value, parsed bool) {

	parsed = strings.HasSuffix(typ.Name(), "ArgsXml")
	if !parsed {
		return
	}
	if _, ok := parserCache[typ]; !ok {
		parserCache[typ] = r
	}
	vl2 := reflect.New(typ)

	decoder := xml.NewDecoder(env.Req.Body)

	if err := decoder.Decode(vl2.Interface()); err != nil {
		vl = reflect.ValueOf(vl2.Interface()).Elem()
		log.Error(err)
		return
	}
	vl = reflect.ValueOf(vl2.Interface()).Elem()
	return
}

type RESTFullArgsParse struct {
}

func parseQuery(url string) (data map[string]string) {

	data = map[string]string{}
	str := strings.Split(url, "/")
	for i := 0; i+1 < len(str); i++ {
		data[strings.ToLower(str[i])] = str[i+1]
	}
	return
}

func (r *RESTFullArgsParse) Parse(env *ReqEnv, typ reflect.Type) (vl reflect.Value, parsed bool) {

	parsed = strings.HasSuffix(typ.Name(), "ArgsRest")
	if !parsed {
		return
	}

	if _, ok := parserCache[typ]; !ok {
		parserCache[typ] = r
	}
	vl = reflect.New(typ).Elem()
	qrys := parseQuery(env.Req.URL.Path)

	if env.Req.Method == "POST" {
		env.Req.ParseForm()
	}

	post := func(key string) string {
		return env.Req.FormValue(key)
	}
	fun := func(qryName string) string {
		if qry, ok := qrys[qryName]; ok {
			return qry
		}
		return post(qryName)
	}

	for i := 0; i < typ.NumField(); i++ {
		qryName := typ.Field(i).Name
		qryName = strings.ToLower(qryName)

		qry := fun(qryName)
		log.Info(qryName, qry)
		tn := typ.Field(i).Type.Name()
		switch tn {
		case "string":
			vl.Field(i).SetString(qry)
		case "int":
			if ival, err := strconv.Atoi(qry); err == nil {
				vl.Field(i).SetInt(int64(ival))
			}
		case "int64":
			if ival, err := strconv.ParseInt(qry, 10, 64); err == nil {
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

//解析名称符合`*Args`的结构体
type ArgsParse struct{}

func (a *ArgsParse) Parse(env *ReqEnv, typ reflect.Type) (vl reflect.Value, parsed bool) {

	parsed = strings.HasSuffix(typ.Name(), "Args")
	if !parsed {
		return
	}
	if _, ok := parserCache[typ]; !ok {
		parserCache[typ] = a
	}

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
		case "int64":
			if ival, err := strconv.ParseInt(qry, 10, 64); err == nil {
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
