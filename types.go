package main

import (
	"net/http"
	"time"

	log "github.com/tommzn/go-log"
)

type webServer struct {
	logger     log.Logger
	httpServer *http.Server
	errTimers  map[string]time.Time
	okTimers   map[string]time.Time
}
