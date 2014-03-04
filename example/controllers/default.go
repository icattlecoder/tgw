package controllers

import (
	"github.com/icattlecoder/tgw"
	. "github.com/icattlecoder/tgw/example/models"
)

type Server struct {
	//
}

func NewServer( /**/) *Server {
	return &Server{}
}

type TestArgs struct {
	Msg string
}

func (s *Server) Hello(args TestArgs, env tgw.ReqEnv) {
	env.RW.Write([]byte(args.Msg))
}

func (s *Server) Index() (data map[string]interface{}) {
	data = map[string]interface{}{}
	author := Author{
		Name:  "icattlecoder",
		Email: []string{"icattlecoder@gmail.com", "iwangming@hotmail.com"},
		QQ:    "405283013",
		Blog:  "http://blog.segmentfault.com/icattlecoder",
	}
	data["author"] = author
	return
}

func (s *Server) Json() (data map[string]interface{}) {
	data = map[string]interface{}{}
	author := Author{
		Name:  "icattlecoder",
		Email: []string{"icattlecoder@gmail.com", "iwangming@hotmail.com"},
		QQ:    "405283013",
		Blog:  "http://blog.segmentfault.com/icattlecoder",
	}
	data["author"] = author
	return
}
