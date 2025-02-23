package nats

import (
	"log"

	go_utils "github.com/ItsMeSamey/go_utils"
	"github.com/nats-io/nats.go"
)
var natConn *nats.Conn
func init() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Println(go_utils.WithStack(err))
	}
	natConn = nc
	log.Println("Connected to NATS")
	SubscribeNewUser()
}