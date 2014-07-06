package tgw

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"testing"
	"time"
)

type Service struct {
}

type HelloArgs struct {
	Who  string
	When int64
	What string
}

const (
	HOST  = "localhost:2343"
	HOST2 = "localhost:2344"
)

func (s *Service) Hello(args HelloArgs, env ReqEnv) (data map[string]interface{}) {
	data = map[string]interface{}{}
	data["who"] = args.Who
	data["when"] = args.When
	data["what"] = args.What
	return
}

// mock setup
func BenchmarkA(b *testing.B) {
	go (func() {
		svr := &Service{}
		_tgw := NewTGW()
		err := _tgw.RegisterREST(&svr).Run(HOST)
		log.Fatalln(err)
	})()
	go (func() {
		mux := http.NewServeMux()
		mux.HandleFunc("/hello", func(rw http.ResponseWriter, req *http.Request) {
			vals := req.URL.Query()
			who := vals.Get("who")
			when, _ := strconv.ParseInt(vals.Get("when"), 10, 64)
			what := vals.Get("what")
			data := map[string]interface{}{}
			data["who"] = who
			data["when"] = when
			data["what"] = what
			bs, _ := json.Marshal(data)
			rw.Write(bs)
		})
		err := http.ListenAndServe(HOST2, mux)
		log.Fatalln(err)
	})()
	time.Sleep(time.Second)
}

func get(now int64, host string) {
	resp, err := http.Get(fmt.Sprintf("http://%s/hello?who=icattlecoder&when=%d&what=hello", host, now))
	if err != nil {
		log.Fatalln("http.Get")
	}

	decoder := json.NewDecoder(resp.Body)
	res := HelloArgs{}
	err = decoder.Decode(&res)
	if err != nil || res.When != now {
		log.Fatalln("not equal")
	}
	resp.Body.Close()
}

func BenchmarkRegister(b *testing.B) {

	for i := 0; i < b.N; i++ {
		get(time.Now().Unix(), HOST)
	}
}

func BenchmarkRegister2(b *testing.B) {

	for i := 0; i < b.N; i++ {
		get(time.Now().Unix(), HOST2)
	}
}
