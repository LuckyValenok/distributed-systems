package endpoints

import (
	"distributedsystems/crud"
	"distributedsystems/server"
	"github.com/valyala/fasthttp"
	"strconv"
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
	if link := c.PostArgs().Peek("link"); len(link) == 0 {
		c.SetStatusCode(fasthttp.StatusBadRequest)
	} else {
		if i, err := crud.AddLink(s.DB, string(link)); err != nil {
			c.SetStatusCode(fasthttp.StatusInternalServerError)
		} else {
			_, _ = c.WriteString(strconv.Itoa(i))
		}
	}
}
