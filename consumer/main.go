package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
	"github.com/valyala/fasthttp"
	config2 "internal/config"
	"internal/rabbit"
	"log"
	"mime/multipart"
	"strconv"
)

type Config struct {
	Host     string `json:"host"`
	User     string `json:"user"`
	Password string `json:"password"`

	WebServer string `json:"web_server"`

	Redis struct {
		Addr     string `json:"addr"`
		Password string `json:"password"`
		DB       int    `json:"db"`
	} `json:"redis"`
}

func main() {
	var config Config
	if err := config2.LoadFromJson("config.json", &config); err != nil {
		log.Fatalf("Load config from json: %v", err)
	}

	redisCfg := config.Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     redisCfg.Addr,
		Password: redisCfg.Password,
		DB:       redisCfg.DB,
	})

	conn, err := amqp.Dial(fmt.Sprintf("amqp://%v:%v@%v/", config.User, config.Password, config.Host))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Fatal(err)
	}
	defer ch.Close()

	q, err := ch.QueueDeclare("linkQueue", true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Fatal(err)
	}

	forever := make(chan bool)

	go func() {
		for d := range msgs {
			var linkData rabbit.LinkData
			if err := json.Unmarshal(d.Body, &linkData); err != nil {
				log.Printf("Ошибка при декодировании JSON: %s", err)
				continue
			}
			log.Printf("Получена ссылка: %s", linkData.URL)

			if status, err := makeHttpRequest(rdb, linkData.URL); err != nil {
				log.Printf("Ошибка при получении HTTP-статуса: %v", err)
			} else {
				log.Printf("HTTP-статус для %s: %d", linkData.URL, status)

				updateLinkStatus(rdb, config.WebServer, linkData.ID, status)
			}
		}
	}()

	log.Printf(" [*] Ожидание сообщений. Для выхода нажмите CTRL+C")
	<-forever
}

func makeHttpRequest(rdb *redis.Client, link string) (int, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod(fasthttp.MethodGet)

	req.SetRequestURI(link)

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	if err := fasthttp.Do(req, res); err != nil {
		return 0, err
	}

	statusCode := res.StatusCode()

	err := rdb.Set(context.Background(), link, statusCode, 0).Err()
	if err != nil {
		log.Printf("Ошибка при сохранении статуса в Redis: %v", err)
	}

	return statusCode, nil
}

func updateLinkStatus(rdb *redis.Client, url string, id, status int) {
	fullURL := fmt.Sprintf("%s/links/", url)

	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)
	req.Header.SetMethod("PUT")
	req.SetRequestURI(fullURL)

	formData := &bytes.Buffer{}
	writer := multipart.NewWriter(formData)
	_ = writer.WriteField("id", strconv.Itoa(id))
	_ = writer.WriteField("status", strconv.Itoa(status))
	writer.Close()

	req.Header.SetContentType(writer.FormDataContentType())
	req.SetBody(formData.Bytes())

	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(resp)
	err := fasthttp.Do(req, resp)
	if err != nil {
		log.Fatalf("Ошибка при отправке запроса: %s", err)
	}

	if resp.StatusCode() != fasthttp.StatusOK {
		log.Fatalf("Ошибка при обновлении статуса: статус %d", resp.StatusCode())
	} else {
		log.Printf("Статус успешно обновлён для ID %d", id)
	}

	if err := rdb.Set(context.Background(), fmt.Sprintf("status:%d", id), status, 0).Err(); err != nil {
		log.Fatalf("Ошибка при сохранении статуса в Redis: %s", err)
	} else {
		log.Printf("Статус для ID %d успешно сохранен в Redis", id)
	}
}
