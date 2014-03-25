package main

import (
	"github.com/icattlecoder/tgw"
	"github.com/icattlecoder/tgw/example/controllers"
	"log"
)

func main() {
	ser := controllers.NewServer()
	t := tgw.NewTGW()

	store := tgw.NewMemcachedSessionStore("127.0.0.1:11211")
	//设置session存储介质为mc
	log.Fatal(t.SetSessionStore(store).Register(&ser).Run(":8080"))
}
