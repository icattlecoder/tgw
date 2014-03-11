package tgw

import (
	"log"
)

//废品日志
type Logger struct {
	level  string
	start  int64
	method string
	url    string
	host   string
	taken  int64
}

func (l *Logger) INFO() {
	l.level = "INFO"
	log.Println(l)
}
