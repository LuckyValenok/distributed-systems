package main

import (
	"database/sql"
	config2 "distributedsystems/config"
	"distributedsystems/endpoints"
	"distributedsystems/server"
	"fmt"
	"github.com/fasthttp/router"
	_ "github.com/lib/pq"
	"github.com/valyala/fasthttp"
	"log"
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
}

type Server server.Server

func main() {
	var config Config
	if err := config2.LoadFromJson("config.json", &config); err != nil {
		log.Fatalf("Load config from json except: %v", err)
	}

	connStr := fmt.Sprintf("user=%v password=%v host=%v port=%v dbname=%v sslmode=disable", config.Database.User, config.Database.Password, config.Database.Host, config.Database.Port, config.Database.DBName)
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	webServer := Server{
		DB: db,
	}
	webServer.setupRouteAndRun(fmt.Sprintf(":%v", config.Port))
}

func (s Server) setupRouteAndRun(addr string) {
	r := router.New()

	endpointsServer := (endpoints.Server)(s)
	r.POST("/links", endpointsServer.AddLinkHandler)
	r.GET("/links/{id}", endpointsServer.GetLinkHandler)

	if err := fasthttp.ListenAndServe(addr, r.Handler); err != nil {
		log.Fatalf("Listen http server except: %v", err)
	}
}
