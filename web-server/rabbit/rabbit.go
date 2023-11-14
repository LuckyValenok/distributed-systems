package rabbit

import (
	"encoding/json"
	"fmt"
	"github.com/streadway/amqp"
	"log"
)

func ConnectRabbitMQ(host, user, pass string) (*amqp.Connection, error) {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%v:%v@%v/", user, pass, host))
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func PublishToRabbitMQ(conn *amqp.Connection, message any) {
	ch, err := conn.Channel()
	if err != nil {
		log.Print(err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("linkQueue", true, false, false, false, nil)
	if err != nil {
		log.Print(err)
	}

	marshal, _ := json.Marshal(message)

	err = ch.Publish("", q.Name, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        marshal,
	})
	if err != nil {
		log.Print(err)
	}
}
