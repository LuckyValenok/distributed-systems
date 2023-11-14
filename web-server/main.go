package main

import (
	"database/sql"
	"fmt"
	"github.com/fasthttp/router"
	_ "github.com/lib/pq"
	"github.com/valyala/fasthttp"
	config2 "internal/config"
	"log"
	"web-server/endpoints"
	"web-server/rabbit"
	"web-server/server"
)

type Config struct {
	Port int `json:"port"`

	Database struct {
		User     string `json:"user"`
		Password string `json:"password"`
		Host     string `json:"host"`
		Port     int    `json:"port"`
		DBName   string `json:"db_name"`
	} `json:"database"`

	RabbitMQ struct {
		Host     string `json:"host"`
		User     string `json:"user"`
		Password string `json:"password"`
	} `json:"rabbit_mq"`
}

type Server server.Server

func main() {
	var config Config
	if err := config2.LoadFromJson("config.json", &config); err != nil {
		log.Fatalf("Load config from json: %v", err)
	}

	connStr := fmt.Sprintf("user=%v password=%v host=%v port=%v dbname=%v sslmode=disable", config.Database.User, config.Database.Password, config.Database.Host, config.Database.Port, config.Database.DBName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	rabbitConf := config.RabbitMQ
	rabbitConn, err := rabbit.ConnectRabbitMQ(rabbitConf.Host, rabbitConf.User, rabbitConf.Password)
	if err != nil {
		log.Fatal(err)
	}

	webServer := Server{
		DB:       db,
		RabbitMQ: rabbitConn,
	}
	webServer.setupRouteAndRun(fmt.Sprintf(":%v", config.Port))
}

func (s Server) setupRouteAndRun(addr string) {
	r := router.New()

	endpointsServer := (endpoints.Server)(s)
	linksGroup := r.Group("/links")
	{
		linksGroup.POST("/", endpointsServer.AddLinkHandler)
		linksGroup.GET("/{id}", endpointsServer.GetLinkHandler)
		linksGroup.PUT("/", endpointsServer.UpdateLinkStatusHandler)
	}

	if err := fasthttp.ListenAndServe(addr, r.Handler); err != nil {
		log.Fatalf("Listen http server except: %v", err)
	}
}
