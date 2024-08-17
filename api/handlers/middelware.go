package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"time"
)

type RecoverMiddleware struct {
	handler http.Handler
}

func NewRecoverMiddleware(handler http.Handler) *RecoverMiddleware {
	return &RecoverMiddleware{
		handler: handler,
	}
}

func (midd *RecoverMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		err := recover()
		if err != nil {
			log.Printf("(!) panic recovery. err: %s", err)

			jsonBody, _ := json.Marshal(map[string]string{
				"error": "internal server error",
			})

			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write(jsonBody)
		}
	}()
	midd.handler.ServeHTTP(w, r)
}

type TimeCounterMiddleware struct {
	f func(w http.ResponseWriter, r *http.Request)
}

func (t *TimeCounterMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	defer func() {
		log.Printf("[INFO] time for %s %s = %fs\n", r.Method, r.URL, time.Since(start).Seconds())
	}()
	t.f(w, r)
}
