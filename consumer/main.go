package main

import (
	"bytes"
	"encoding/json"
	"fmt"
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
}

func main() {
	var config Config
	if err := config2.LoadFromJson("config.json", &config); err != nil {
		log.Fatalf("Load config from json: %v", err)
	}

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

			if status, err := makeHttpRequest(linkData.URL); err != nil {
				log.Printf("Ошибка при получении HTTP-статуса: %v", err)
			} else {
				log.Printf("HTTP-статус для %s: %d", linkData.URL, status)

				updateLinkStatus(config.WebServer, linkData.ID, status)
			}
		}
	}()

	log.Printf(" [*] Ожидание сообщений. Для выхода нажмите CTRL+C")
	<-forever
}

func makeHttpRequest(link string) (int, error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req)

	req.Header.SetMethod(fasthttp.MethodGet)

	req.SetRequestURI(link)

	res := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseResponse(res)

	if err := fasthttp.Do(req, res); err != nil {
		return 0, err
	}

	return res.StatusCode(), nil
}

type LinkStatusUpdate struct {
	ID     int `json:"id"`
	Status int `json:"status"`
}

func updateLinkStatus(url string, id, status int) {
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
}
