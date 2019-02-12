package main

import (
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	_ "net/http/pprof"

	"github.com/satori/go.uuid"
	"github.com/valyala/fasthttp"
)

var s = &StorageMu{
	data: make(map[uuid.UUID]Model),
}

func main() {
	// go http.ListenAndServe(":6060", nil)
	log.Fatal(fasthttp.ListenAndServe(":8080", Handler))
}

var (
	notFoundMsg   = []byte(`{"error":"not found"}`)
	errorMsg      = []byte(`{"error":"invalid data"}`)
	invalidMethod = []byte(`{"error":"invalid method"}`)
)

func Handler(ctx *fasthttp.RequestCtx) {
	if ctx.IsGet() {
		id, err := uuid.FromString(string(ctx.RequestURI()[1:]))
		if err != nil {
			ctx.Error("invalid UUID format", http.StatusBadRequest)
			return
		}
		GetHandler(ctx, id)
		return
	}
	if ctx.IsPost() {
		PostHandler(ctx)
		return
	}
	ctx.SetBody(invalidMethod)
	ctx.SetStatusCode(fasthttp.StatusMethodNotAllowed)
}

func PostHandler(ctx *fasthttp.RequestCtx) {
	var m Model
	if err := json.Unmarshal(ctx.Request.Body(), &m); err != nil {
		ctx.SetBody(errorMsg)
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		return
	}
	s.Set(m)
	ctx.SetStatusCode(fasthttp.StatusNoContent)
}

func GetHandler(ctx *fasthttp.RequestCtx, id uuid.UUID) {
	ctx.SetContentType("application/json")
	if m, ok := s.Get(id); ok {
		b, _ := json.Marshal(m)
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.SetBody(b)
		return
	}
	ctx.SetBody(notFoundMsg)
	ctx.SetStatusCode(fasthttp.StatusNotFound)
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
