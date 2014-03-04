package tgw

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

var (
	viewDir = "view"
	DEBUG   = true
)

type ReqEnv struct {
	RW  http.ResponseWriter
	Req *http.Request
}

func staticDirHandler(mux *http.ServeMux, prefix string, staticDir string, flags int) {
	mux.HandleFunc(prefix, func(w http.ResponseWriter, r *http.Request) {

		file := staticDir
		if prefix == "/" && r.URL.Path == "/" {
			file += "/index.html"
		} else {
			file += r.URL.Path[len(staticDir)+1:]
		}
		if (flags) == 0 {
			fi, err := os.Stat(file)
			if err != nil || fi.IsDir() {
				http.NotFound(w, r)
				return
			}
		}
		http.ServeFile(w, r, file)
	})
}

func Register(controller interface{}) (mux *http.ServeMux) {

	mux = http.NewServeMux()
	_type := reflect.TypeOf(controller).Elem()
	_value := reflect.ValueOf(controller).Elem()

	view, err := NewView(viewDir)
	if err != nil {
		log.Println("NewView err:", err)
	}
	env := ReqEnv{}
	env_type := reflect.TypeOf(env)

	//auto register routers based on reflect
	for i := 0; i < _type.NumMethod(); i++ {

		funName := _type.Method(i).Name
		router := fun_router(funName)
		method := _value.Method(i)
		methodTyp := method.Type()

		log.Println("Register ", router, "===>", funName)
		mux.HandleFunc(router, func(rw http.ResponseWriter, req *http.Request) {

			args := []reflect.Value{}
			for i := 0; i < methodTyp.NumIn(); i++ {
				//解析第i个参数
				arg_t := methodTyp.In(i)
				arg_v := reflect.New(arg_t).Elem()
				if arg_t == env_type {
					arg_v = reflect.ValueOf(newReqEnv(rw, req))
				} else {
					requestQueryParse(req, arg_t, &arg_v)
				}
				args = append(args, arg_v)
			}
			start := time.Now()
			callRet := method.Call(args)

			if len(callRet) > 0 {
				if tpl, err := view.Get(router); err != nil {
					if bytes, err := json.Marshal(callRet[0].Interface()); err == nil {
						rw.Write(bytes)
					}
				} else {
					tpl.Execute(rw, callRet[0].Interface())
				}
			}
			end := time.Now()
			lgr := Logger{start: start.Unix(), method: req.Method, url: req.URL.RawQuery, host: req.Host, taken: end.Sub(start).Nanoseconds()}
			lgr.INFO()
		})
	}

	//static file server
	staticDirHandler(mux, "/static/", "static", 0)
	return
}

func newReqEnv(rw http.ResponseWriter, req *http.Request) (reqEnv ReqEnv) {
	reqEnv = ReqEnv{RW: rw, Req: req}
	return
}

// 通过Request参数解析形参
func requestQueryParse(req *http.Request, typ reflect.Type, vl *reflect.Value) {

	get := func(key string) string {
		return req.URL.Query().Get(key)
	}

	post := func(key string) string {
		return req.FormValue(key)
	}

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

func fun_router(funName string) string {
	paths := [][]byte{}
	for s, i := 0, 0; i < len(funName); i++ {
		if funName[i] >= 'A' && funName[i] <= 'Z' {
			paths = append(paths, []byte(funName[s:i]))
			s = i
		}
		if i == len(funName)-1 {
			paths = append(paths, []byte(funName[s:]))
		}
	}
	return strings.ToLower(string(bytes.Join(paths, []byte{'/'})))
}
