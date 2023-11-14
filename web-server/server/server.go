package server

import (
	"database/sql"
	"github.com/streadway/amqp"
)

type Server struct {
	DB       *sql.DB
	RabbitMQ *amqp.Connection
}
