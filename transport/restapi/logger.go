package restapi

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/middleware"
	"github.com/go-kit/kit/log"
)

type KitLoggerWrapper struct {
	log log.Logger
}

func (l *KitLoggerWrapper) NewLogEntry(r *http.Request) middleware.LogEntry {
	return &KitLogEntry{l.log, r}
}

type KitLogEntry struct {
	log log.Logger
	r   *http.Request
}

func (k *KitLogEntry) Write(status, bytes int, elapsed time.Duration) {
	k.log.Log("time", time.Now().Format(time.RFC3339), "method", k.r.Method, "uri", k.r.RequestURI, "status", status, "bytes", bytes, "elapsed", elapsed)
}

func (k KitLogEntry) Panic(v interface{}, stack []byte) {
	k.log.Log("panic", fmt.Sprintf("%+v", v), "stack", string(stack))
}
