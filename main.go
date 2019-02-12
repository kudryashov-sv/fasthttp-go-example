package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/buaazp/fasthttprouter"

	"github.com/satori/go.uuid"
	"github.com/valyala/fasthttp"
)

var s = &StorageMu{
	data: make(map[uuid.UUID]Model),
}

func main() {
	router := fasthttprouter.New()
	router.GET("/:id", func(ctx *fasthttp.RequestCtx) {
		id, err := uuid.FromString(ctx.UserValue("id").(string))
		if err != nil {
			ctx.Error(`{"error":"invalid data"}`, http.StatusBadRequest)
			return
		}
		ctx.SetContentType("application/json")
		if m, ok := s.Get(id); ok {
			b, _ := json.Marshal(m)
			ctx.SetStatusCode(fasthttp.StatusOK)
			ctx.SetBody(b)
			return
		}
		ctx.Error(`{"error":"not found"}`, http.StatusNotFound)
	})
	router.POST("/", func(ctx *fasthttp.RequestCtx) {
		var m Model
		if err := json.Unmarshal(ctx.Request.Body(), &m); err != nil {
			ctx.SetBody([]byte(`{"error":"invalid data"}`))
			ctx.SetStatusCode(fasthttp.StatusBadRequest)
			return
		}
		s.Set(m)
		ctx.SetStatusCode(fasthttp.StatusNoContent)
	})
	log.Fatal(fasthttp.ListenAndServe(":8080", router.Handler))
}

type Model struct {
	Id        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Timestamp time.Time `json:"timestamp"`
}

type StorageMu struct {
	mu   sync.RWMutex
	data map[uuid.UUID]Model
}

func (s *StorageMu) Set(m Model) {
	s.mu.Lock()
	s.data[m.Id] = m
	s.mu.Unlock()
}

func (s *StorageMu) Get(id uuid.UUID) (m Model, found bool) {
	s.mu.RLock()
	m, found = s.data[id]
	s.mu.RUnlock()
	return
}
