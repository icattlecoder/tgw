package tgw

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"reflect"
	"strings"
	"time"
)

var (
	viewDir   = "view"
	staticDir = "static"
	DEBUG     = true
)

type ReqEnv struct {
	RW      http.ResponseWriter
	Req     *http.Request
	Session SessionInterface
}

type tgw struct {
	parses       []RegisterParse
	mux          *http.ServeMux
	sessionStore SessionStoreInterface
	index 		string
}

func NewTGW() *tgw {
	mux := http.NewServeMux()
	t := tgw{mux: mux,index:"/index"}
	argsParse := ArgsParse{}
	envParser := EnvParse{}

	return t.AddParser(&argsParse).AddParser(&envParser)
}

func (t *tgw) AddParser(parser RegisterParse) *tgw {
	t.parses = append(t.parses, parser)
	return t
}

// eg . "/default"
func (t *tgw) SetIndexPage(prefix string) *tgw {
	t.index = prefix
	return t
}

func (t *tgw) SetSessionStore(store SessionStoreInterface) *tgw {
	t.sessionStore = store
	return t
}

func (t *tgw) Register(controller interface{}) *tgw {

	if t.mux == nil {
		t.mux = http.NewServeMux()
	}
	_type := reflect.TypeOf(controller).Elem()
	_value := reflect.ValueOf(controller).Elem()

	view, err := NewView(viewDir)
	if err != nil {
		log.Println("NewView err:", err)
	}
	if t.sessionStore == nil {
		t.sessionStore = NewSimpleSessionStore()
	}

	//auto register routers based on reflect
	for i := 0; i < _type.NumMethod(); i++ {

		funName := _type.Method(i).Name
		router := fun_router(funName)
		method := _value.Method(i)
		methodTyp := method.Type()

		viewName := router
		if router == t.index {
			router ="/"
		}
		log.Println("Register ", router, "===>", funName)

		t.mux.HandleFunc(router, func(rw http.ResponseWriter, req *http.Request) {
			session := NewSimpleSession(rw, req, t.sessionStore)
			args := []reflect.Value{}
			env := newReqEnv(rw, req, session)
			for i := 0; i < methodTyp.NumIn(); i++ {
				//解析第i个参数
				arg_t := methodTyp.In(i)
				for _, v := range t.parses {
					if arg_v, ok := v.Parse(&env, arg_t); ok {
						args = append(args, arg_v)
						break
					}
				}
			}
			start := time.Now()
			callRet := method.Call(args)

			if len(callRet) > 0 {
				if tpl, err := view.Get(viewName); err != nil {
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
	return t
}

func (t *tgw) Run(addr string) (err error) {
	if t.mux == nil {
		err = errors.New("mux is nil")
		return
	}
	//static file server
	staticDirHandler(t.mux, "/static/", staticDir)
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
