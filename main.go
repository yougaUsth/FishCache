package main

import (
	"FishCache/service"
	"fmt"
)

func commandHandler(session *service.Session, msg *service.Message) {
	fmt.Print(msg)

}


func main() {
	ss, err := service.NewSocketService(":8889")
	if err != nil {
		panic(err)
	}
	ss.RegMessageHandler(commandHandler)
	ss.Serv()
}
