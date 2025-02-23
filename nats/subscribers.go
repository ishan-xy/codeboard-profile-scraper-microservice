package nats

import (
	"log"
	"scraper/worker"

	"github.com/nats-io/nats.go"
)

func SubscribeNewUser() {
	if natConn.IsConnected(){
		natConn.Subscribe("user.created", func(m *nats.Msg) {
			log.Println(string(m.Data))
			worker.AddUsernameToCache(string(m.Data))
		})
	} else {
		log.Println("NATS connection is not established")
	}
	select {}
}