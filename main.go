package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/ArtemOlyinik/client_server_Practichna_04/internal/handler"
	appmiddleware "github.com/ArtemOlyinik/client_server_Practichna_04/internal/middleware"
	"github.com/ArtemOlyinik/client_server_Practichna_04/internal/repository"
)

func main() {
	log.SetFlags(0)
	log.Println("=== Мережевий реєстр пристроїв (Device Registry API) ===")

	repo := repository.NewMemoryDeviceRepository()
	deviceHandler := handler.NewDeviceHandler(repo)

	router := chi.NewRouter()
	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(appmiddleware.Logger)
	router.Use(appmiddleware.Recoverer)
	router.Use(appmiddleware.CORS)

	router.Route("/api/v1/devices", func(r chi.Router) {
		r.Get("/", deviceHandler.List)
		r.Post("/", deviceHandler.Create)

		r.Route("/{id}", func(r chi.Router) {
			r.Get("/", deviceHandler.Get)
			r.Put("/", deviceHandler.Update)
			r.Delete("/", deviceHandler.Delete)
		})
	})

	server := &http.Server{
		Addr:              ":8080",
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Println("[СЕРВЕР] Ініціалізація сховища даних (sync.RWMutex: active)")
	log.Println("[СЕРВЕР] Конвеєр Middleware підключено: Logger, Recoverer, CORS")
	log.Println("[СЕРВЕР] Маршрути Chi змонтовано на префіксі /api/v1/devices")
	log.Println("[СЕРВЕР] HTTP-сервер запущено на порту :8080...")

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("[СЕРВЕР] Помилка HTTP-сервера: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	log.Println("[СЕРВЕР] Отримано сигнал завершення. Зупинка HTTP-сервера...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Printf("[СЕРВЕР] Помилка плавної зупинки: %v", err)
	}
	log.Println("[СЕРВЕР] Сервер успішно зупинено.")
}
