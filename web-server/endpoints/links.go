package endpoints

import (
	"github.com/valyala/fasthttp"
	rabbit2 "internal/rabbit"
	"strconv"
	"web-server/crud"
	"web-server/rabbit"
	"web-server/server"
)

type Server server.Server

func (s Server) GetLinkHandler(c *fasthttp.RequestCtx) {
	if idStr, ok := c.UserValue("id").(string); !ok {
		c.SetStatusCode(fasthttp.StatusBadRequest)
	} else {
		if id, err := strconv.Atoi(idStr); err != nil {
			c.SetStatusCode(fasthttp.StatusBadRequest)
		} else if link, err := crud.GetLink(s.DB, id); err != nil {
			c.SetStatusCode(fasthttp.StatusNotFound)
		} else {
			_, _ = c.WriteString(link)
		}
	}
}

func (s Server) AddLinkHandler(c *fasthttp.RequestCtx) {
	if link := c.FormValue("link"); len(link) == 0 {
		c.SetStatusCode(fasthttp.StatusBadRequest)
	} else if i, err := crud.AddLink(s.DB, string(link)); err != nil {
		c.SetStatusCode(fasthttp.StatusInternalServerError)
	} else {
		rabbit.PublishToRabbitMQ(s.RabbitMQ, rabbit2.LinkData{
			ID:  i,
			URL: string(link),
		})

		_, _ = c.WriteString(strconv.Itoa(i))
	}
}

func (s Server) UpdateLinkStatusHandler(c *fasthttp.RequestCtx) {
	idStr := c.FormValue("id")
	statusStr := c.FormValue("status")

	if len(idStr) == 0 || len(statusStr) == 0 {
		c.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	id, err := strconv.Atoi(string(idStr))
	if err != nil {
		c.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	status, err := strconv.Atoi(string(statusStr))
	if err != nil {
		c.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}

	if err := crud.UpdateLinkStatus(s.DB, id, status); err != nil {
		c.SetStatusCode(fasthttp.StatusInternalServerError)
	} else {
		c.SetStatusCode(fasthttp.StatusOK)
	}
}
