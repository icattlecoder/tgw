package main

import (
	"github.com/icattlecoder/tgw"
	"github.com/icattlecoder/tgw/example/controllers"
	"log"
	"net/http"
)

func main() {
	ser := controllers.NewServer()
	mux := tgw.Register(&ser)
	log.Fatal(http.ListenAndServe(":8080", mux))

}
