package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/mux"

	log "github.com/tommzn/go-log"
)

const DEFAULT_PORT = "8080"

func newServer(logger log.Logger) *webServer {
	return &webServer{
		logger:    logger,
		errTimers: make(map[string]time.Time),
		okTimers:  make(map[string]time.Time),
	}
}

// Run starts a HTTP server to listen for rendering requests.
func (server *webServer) Run(ctx context.Context, waitGroup *sync.WaitGroup) error {

	defer waitGroup.Done()
	defer server.logger.Flush()

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/test", server.handleRequest)
	router.HandleFunc("/health", server.handleHealthCheckRequest).Methods("GET")

	server.logger.Infof("Listen [%s]", DEFAULT_PORT)
	server.logger.Flush()
	server.httpServer = &http.Server{Addr: ":" + DEFAULT_PORT, Handler: router}

	endChan := make(chan error, 1)
	go func() {
		endChan <- server.httpServer.ListenAndServe()
	}()

	select {
	case <-ctx.Done():
		server.stopHttpServer()
	case err := <-endChan:
		return err
	}
	return nil
}

// StopHttpServer will try to sop running HTTP server graceful. Timeout is 3s.
func (server *webServer) stopHttpServer() {
	server.logger.Info("Stopping HTTP server.")
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := server.httpServer.Shutdown(ctx); err != nil {
		server.logger.Error("Unable to stop HTTP server, reason: ", err)
	}
}

func (server *webServer) handleHealthCheckRequest(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}

func (server *webServer) handleRequest(w http.ResponseWriter, r *http.Request) {

	defer server.logger.Flush()
	defer r.Body.Close()

	server.logContextFromRequest(r)
	if buf, err := ioutil.ReadAll(r.Body); err == nil {
		server.logger.Debugf("Payload: %s", buf)
		r.Body = ioutil.NopCloser(bytes.NewBuffer(buf))
	} else {
		server.logger.Error("Unable to read request body, reason: ", err)
	}

	w.WriteHeader(server.getResponseStatus(r))
}

func (server *webServer) logContextFromRequest(r *http.Request) {
	contextValues := make(map[string]string)
	contextValues["requestid"] = uuid.New().String()
	contextValues["http.method"] = r.Method
	contextValues["http.url"] = r.URL.String()
	contextValues["http.uri"] = r.RequestURI
	contextValues["http.headers"] = fmt.Sprintf("%+v", r.Header)
	log.AppendContextValues(server.logger, contextValues)
}

func (server *webServer) getResponseStatus(r *http.Request) int {

	switch r.URL.Query().Get("responsestatusstrategy") {
	case "5xx":

		clientId := r.URL.Query().Get("clientid")
		if clientId == "" {
			clientId = "<default>"
		}

		if oktimer, ok := server.okTimers[clientId]; ok {
			if oktimer.Before(time.Now()) {
				delete(server.okTimers, clientId)
				server.errTimers[clientId] = time.Now().Add(time.Minute * 2)
			}
			return http.StatusNoContent
		}

		if errtimer, ok := server.errTimers[clientId]; ok {
			if errtimer.Before(time.Now()) {
				delete(server.errTimers, clientId)
				server.okTimers[clientId] = time.Now().Add(time.Minute * 2)
			}
			return http.StatusServiceUnavailable
		}
		server.errTimers[clientId] = time.Now().Add(time.Minute * 2)
		return http.StatusServiceUnavailable

	case "4xx":
		return http.StatusBadRequest

	case "429":
		if rand.Intn(100) <= 70 {
			return http.StatusTooManyRequests
		} else {
			return http.StatusNoContent
		}

	default:
		return http.StatusNoContent
	}
}
