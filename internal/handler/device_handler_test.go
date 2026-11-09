package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/ArtemOlyinik/client_server_Practichna_04/internal/apierror"
	"github.com/ArtemOlyinik/client_server_Practichna_04/internal/domain"
	appmiddleware "github.com/ArtemOlyinik/client_server_Practichna_04/internal/middleware"
	"github.com/ArtemOlyinik/client_server_Practichna_04/internal/repository"
)

func newTestRouter() *chi.Mux {
	repo := repository.NewMemoryDeviceRepository()
	deviceHandler := NewDeviceHandler(repo)

	router := chi.NewRouter()
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

	return router
}

func performRequest(t *testing.T, router http.Handler, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()

	request := httptest.NewRequest(method, path, strings.NewReader(body))
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)
	return recorder
}

func createDevice(t *testing.T, router http.Handler) domain.Device {
	t.Helper()

	response := performRequest(t, router, http.MethodPost, "/api/v1/devices",
		`{"name":"Main Thermal Sensor","ip_address":"192.168.1.50","type":"sensor","status":"online"}`)
	if response.Code != http.StatusCreated {
		t.Fatalf("очікували 201, отримали %d: %s", response.Code, response.Body.String())
	}

	var device domain.Device
	if err := json.Unmarshal(response.Body.Bytes(), &device); err != nil {
		t.Fatalf("не вдалося розібрати відповідь: %v", err)
	}
	return device
}

func TestCreateDevice(t *testing.T) {
	router := newTestRouter()
	device := createDevice(t, router)

	if device.ID != "dev-001" {
		t.Fatalf("очікували dev-001, отримали %s", device.ID)
	}
	if device.Status != "online" {
		t.Fatalf("очікували status=online, отримали %s", device.Status)
	}
	if device.UpdatedAt.IsZero() {
		t.Fatal("updated_at не встановлено")
	}
}

func TestCreateDeviceValidation(t *testing.T) {
	router := newTestRouter()
	response := performRequest(t, router, http.MethodPost, "/api/v1/devices",
		`{"name":"","ip_address":"192.168.1.50","type":"sensor"}`)

	if response.Code != http.StatusBadRequest {
		t.Fatalf("очікували 400, отримали %d", response.Code)
	}

	var apiErr apierror.Error
	if err := json.Unmarshal(response.Body.Bytes(), &apiErr); err != nil {
		t.Fatalf("не вдалося розібрати помилку: %v", err)
	}
	if apiErr.Status != http.StatusBadRequest || apiErr.Error != "Validation Error" {
		t.Fatalf("некоректний формат помилки: %+v", apiErr)
	}
	if apiErr.Timestamp.IsZero() {
		t.Fatal("у помилці відсутній timestamp")
	}
}

func TestGetDeviceAndNotFound(t *testing.T) {
	router := newTestRouter()
	device := createDevice(t, router)

	response := performRequest(t, router, http.MethodGet, "/api/v1/devices/"+device.ID, "")
	if response.Code != http.StatusOK {
		t.Fatalf("очікували 200, отримали %d", response.Code)
	}

	response = performRequest(t, router, http.MethodGet, "/api/v1/devices/dev-999", "")
	if response.Code != http.StatusNotFound {
		t.Fatalf("очікували 404, отримали %d", response.Code)
	}
}

func TestUpdateAndDeleteDevice(t *testing.T) {
	router := newTestRouter()
	device := createDevice(t, router)

	response := performRequest(t, router, http.MethodPut, "/api/v1/devices/"+device.ID,
		`{"name":"Updated Sensor","ip_address":"192.168.1.51","type":"sensor","status":"offline"}`)
	if response.Code != http.StatusOK {
		t.Fatalf("очікували 200 після оновлення, отримали %d: %s", response.Code, response.Body.String())
	}

	var updated domain.Device
	if err := json.Unmarshal(response.Body.Bytes(), &updated); err != nil {
		t.Fatalf("не вдалося розібрати оновлений пристрій: %v", err)
	}
	if updated.Name != "Updated Sensor" || updated.Status != "offline" {
		t.Fatalf("пристрій оновився некоректно: %+v", updated)
	}

	response = performRequest(t, router, http.MethodDelete, "/api/v1/devices/"+device.ID, "")
	if response.Code != http.StatusNoContent {
		t.Fatalf("очікували 204, отримали %d", response.Code)
	}

	response = performRequest(t, router, http.MethodGet, "/api/v1/devices/"+device.ID, "")
	if response.Code != http.StatusNotFound {
		t.Fatalf("після видалення очікували 404, отримали %d", response.Code)
	}
}

func TestListDevicesWithFilters(t *testing.T) {
	router := newTestRouter()
	createDevice(t, router)

	_ = performRequest(t, router, http.MethodPost, "/api/v1/devices",
		`{"name":"Offline Sensor","ip_address":"192.168.1.60","type":"sensor","status":"offline"}`)
	_ = performRequest(t, router, http.MethodPost, "/api/v1/devices",
		`{"name":"Online Gateway","ip_address":"192.168.1.70","type":"gateway","status":"online"}`)

	response := performRequest(t, router, http.MethodGet, "/api/v1/devices?status=online&type=sensor", "")
	if response.Code != http.StatusOK {
		t.Fatalf("очікували 200, отримали %d", response.Code)
	}

	var devices []domain.Device
	if err := json.Unmarshal(response.Body.Bytes(), &devices); err != nil {
		t.Fatalf("не вдалося розібрати список: %v", err)
	}
	if len(devices) != 1 || devices[0].Name != "Main Thermal Sensor" {
		t.Fatalf("фільтрація повернула некоректний список: %+v", devices)
	}
}

func TestCORSPreflight(t *testing.T) {
	router := newTestRouter()
	response := performRequest(t, router, http.MethodOptions, "/api/v1/devices", "")

	if response.Code != http.StatusNoContent {
		t.Fatalf("очікували 204, отримали %d", response.Code)
	}
	if response.Header().Get("Access-Control-Allow-Origin") != "*" {
		t.Fatal("відсутній заголовок Access-Control-Allow-Origin")
	}
	if !strings.Contains(response.Header().Get("Access-Control-Allow-Methods"), "DELETE") {
		t.Fatal("у CORS немає методу DELETE")
	}
}

func TestPanicRecovery(t *testing.T) {
	router := chi.NewRouter()
	router.Use(appmiddleware.Recoverer)
	router.Get("/panic", func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	request := httptest.NewRequest(http.MethodGet, "/panic", bytes.NewReader(nil))
	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("очікували 500, отримали %d", recorder.Code)
	}

	var apiErr apierror.Error
	if err := json.Unmarshal(recorder.Body.Bytes(), &apiErr); err != nil {
		t.Fatalf("не вдалося розібрати відповідь про паніку: %v", err)
	}
	if apiErr.Status != http.StatusInternalServerError {
		t.Fatalf("очікували status=500 у JSON, отримали %d", apiErr.Status)
	}
}
