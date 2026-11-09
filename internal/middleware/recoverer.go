package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/ArtemOlyinik/client_server_Practichna_04/internal/apierror"
)

// Recoverer перехоплює паніку в обробнику та повертає клієнту JSON 500.
func Recoverer(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Printf("[PANIC] %s %s: %v\n%s", r.Method, r.URL.Path, recovered, debug.Stack())
				apierror.Write(w, http.StatusInternalServerError, "Internal Server Error", "Сталася непередбачена помилка сервера")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
