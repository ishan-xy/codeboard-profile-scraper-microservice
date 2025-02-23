package worker

import (
	"log"

	go_utils "github.com/ItsMeSamey/go_utils"
	"github.com/nats-io/nats.go"
)

func ConnectNats() (*nats.Conn, error) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		return nil, go_utils.WithStack(err)
	}
	return nc, nil
}

func InitSubs(){
	nc, _ := ConnectNats()
	log.Println("Connected to NATS")
	SubscribeNewUser(nc)
	select {}
}