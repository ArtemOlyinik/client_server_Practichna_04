package middleware

import (
	"log"
	"net/http"
	"time"
)

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

func (w *statusWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.ResponseWriter.Write(data)
}

// Logger виводить метод, шлях, статус, тривалість і адресу клієнта.
func Logger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		writer := &statusWriter{ResponseWriter: w}

		next.ServeHTTP(writer, r)

		if writer.status == 0 {
			writer.status = http.StatusOK
		}

		log.Printf("[HTTP] %s - %s %s -> %d %s (%s)",
			r.RemoteAddr,
			r.Method,
			r.URL.RequestURI(),
			writer.status,
			http.StatusText(writer.status),
			time.Since(start).Round(time.Microsecond),
		)
	})
}
