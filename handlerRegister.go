package tgw

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"reflect"
	"strings"
)

var (
	ViewDir   = "view"
	StaticDir = "static"
)

type ReqEnv struct {
	RW      http.ResponseWriter
	Req     *http.Request
	Session SessionInterface
}

type tgw struct {
	parsesr      []RegisterParser
	mux          *http.ServeMux
	sessionStore SessionStoreInterface
	index        string
}

func NewTGW() *tgw {

	mux := http.NewServeMux()
	t := tgw{mux: mux, index: "/index"}
	argsParse := ArgsParse{}
	envParser := EnvParse{}

	return t.AddParser(&argsParse).AddParser(&envParser)
}

//增加url=>args的解析接口。内置两种，EnvParse和ArgsParse，请参考args_parser.go
//用户也可以根据自己的业务逻辑实现Parse接口
func (t *tgw) AddParser(parser RegisterParser) *tgw {
	t.parsesr = append(t.parsesr, parser)
	return t
}

// 设置默认页面，默认为"/"=>"/index",即访问localhost实现路由至localhost/index
func (t *tgw) SetIndexPage(prefix string) *tgw {
	t.index = prefix
	return t
}

// 设置Session的存储介质，内置memecached
func (t *tgw) SetSessionStore(store SessionStoreInterface) *tgw {
	t.sessionStore = store
	return t
}

// 注册路由，此函数自动将入参的对外方法注册为路由，成员方法以驼峰式命名,其规则如下：
// AxxxBxxCxx() => /axxx/bxx/cx
func (t *tgw) Register(controller interface{}) *tgw {

	if t.mux == nil {
		t.mux = http.NewServeMux()
	}
	_type := reflect.TypeOf(controller).Elem()
	_value := reflect.ValueOf(controller).Elem()

	view, err := NewView(ViewDir)
	if err != nil {
		log.Println("NewView err:", err)
	}
	if t.sessionStore == nil {
		t.sessionStore = NewMemcachedSessionStore()
	}

	//auto register routers based on reflect
	for i := 0; i < _type.NumMethod(); i++ {

		funName := _type.Method(i).Name
		// only register public func
		if !(funName[0] >= 'A' && funName[0] <= 'Z') {
			continue
		}

		router := fun_router(funName)
		method := _value.Method(i)
		methodTyp := method.Type()

		viewName := router
		//set defalut page
		if router == t.index {
			router = "/"
		}
		log.Println("Register ", router, "===>", funName)

		t.mux.HandleFunc(router, func(rw http.ResponseWriter, req *http.Request) {
			session := NewSimpleSession(rw, req, t.sessionStore)
			args := []reflect.Value{}
			env := newReqEnv(rw, req, session)
			for i := 0; i < methodTyp.NumIn(); i++ {
				arg_t := methodTyp.In(i)
				for _, v := range t.parsesr {
					if arg_v, ok := v.Parse(&env, arg_t); ok {
						args = append(args, arg_v)
						break
					}
				}
			}
			callRet := method.Call(args)

			if len(callRet) > 0 {
				if tpl, err := view.Get(viewName); err != nil {
					if bytes, err := json.Marshal(callRet[0].Interface()); err == nil {
						rw.Write(bytes)
					}
				} else {
					tpl.Execute(rw, callRet[0].Interface())
				}
			} else {
				if r, err := view.GetHtml(viewName); err == nil {
					io.Copy(rw, r)
				}
			}
		})
	}
	return t
}

func (t *tgw) RegisterREST(controller interface{}) *tgw {
	if t.mux == nil {
		t.mux = http.NewServeMux()
	}
	_type := reflect.TypeOf(controller).Elem()
	_value := reflect.ValueOf(controller).Elem()

	//auto register routers based on reflect
	for i := 0; i < _type.NumMethod(); i++ {

		funName := _type.Method(i).Name
		// only register public func
		if !(funName[0] >= 'A' && funName[0] <= 'Z') {
			continue
		}

		router := fun_router(funName)
		method := _value.Method(i)
		methodTyp := method.Type()

		//set defalut page
		if router == t.index {
			router = "/"
		}
		log.Println("Register ", router, "===>", funName)

		t.mux.HandleFunc(router, func(rw http.ResponseWriter, req *http.Request) {
			session := NewSimpleSession(rw, req, t.sessionStore)
			args := []reflect.Value{}
			env := newReqEnv(rw, req, session)
			for i := 0; i < methodTyp.NumIn(); i++ {
				arg_t := methodTyp.In(i)
				for _, v := range t.parsesr {
					if arg_v, ok := v.Parse(&env, arg_t); ok {
						args = append(args, arg_v)
						break
					}
				}
			}
			callRet := method.Call(args)

			if len(callRet) > 0 {
				if bytes, err := json.Marshal(callRet[0].Interface()); err == nil {
					rw.Write(bytes)
				}
			}
		})
	}
	return t
}

func (t *tgw) Run(addr string) (err error) {
	if t.mux == nil {
		err = errors.New("mux is nil")
		return
	}
	//static file server
	staticDirHandler(t.mux, "/static/", StaticDir)
	log.Println("Listen:", addr)
	return http.ListenAndServe(addr, t.mux)
}

func staticDirHandler(mux *http.ServeMux, prefix string, staticDir string) {
	mux.HandleFunc(prefix, func(w http.ResponseWriter, r *http.Request) {
		file := staticDir + r.URL.Path[len(staticDir)+1:]
		http.ServeFile(w, r, file)
	})
}

func newReqEnv(rw http.ResponseWriter, req *http.Request, session SessionInterface) ReqEnv {
	return ReqEnv{RW: rw, Req: req, Session: session}
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
