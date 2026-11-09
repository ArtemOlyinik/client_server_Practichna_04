package handler

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/ArtemOlyinik/client_server_Practichna_04/internal/apierror"
	"github.com/ArtemOlyinik/client_server_Practichna_04/internal/domain"
	"github.com/ArtemOlyinik/client_server_Practichna_04/internal/repository"
)

type DeviceHandler struct {
	repo repository.DeviceRepository
}

func NewDeviceHandler(repo repository.DeviceRepository) *DeviceHandler {
	return &DeviceHandler{repo: repo}
}

func (h *DeviceHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := domain.DeviceFilter{
		Status: r.URL.Query().Get("status"),
		Type:   r.URL.Query().Get("type"),
	}
	writeJSON(w, http.StatusOK, h.repo.GetAll(filter))
}

func (h *DeviceHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input domain.DeviceInput
	if err := decodeJSON(r, &input); err != nil {
		apierror.Write(w, http.StatusBadRequest, "Bad Request", "Некоректний JSON: "+err.Error())
		return
	}

	if err := validateInput(input); err != nil {
		apierror.Write(w, http.StatusBadRequest, "Validation Error", err.Error())
		return
	}

	device := domain.Device{
		Name:      strings.TrimSpace(input.Name),
		IPAddress: strings.TrimSpace(input.IPAddress),
		Type:      strings.ToLower(strings.TrimSpace(input.Type)),
		Status:    normalizedStatus(input.Status),
	}

	created, err := h.repo.Create(device)
	if err != nil {
		apierror.Write(w, http.StatusInternalServerError, "Internal Server Error", "Не вдалося зберегти пристрій")
		return
	}

	w.Header().Set("Location", "/api/v1/devices/"+created.ID)
	writeJSON(w, http.StatusCreated, created)
}

func (h *DeviceHandler) Get(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	device, exists := h.repo.GetByID(id)
	if !exists {
		apierror.Write(w, http.StatusNotFound, "Resource Not Found", fmt.Sprintf("Device with ID '%s' does not exist in registry", id))
		return
	}
	writeJSON(w, http.StatusOK, device)
}

func (h *DeviceHandler) Update(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	var input domain.DeviceInput
	if err := decodeJSON(r, &input); err != nil {
		apierror.Write(w, http.StatusBadRequest, "Bad Request", "Некоректний JSON: "+err.Error())
		return
	}

	if err := validateInput(input); err != nil {
		apierror.Write(w, http.StatusBadRequest, "Validation Error", err.Error())
		return
	}

	updated, exists := h.repo.Update(id, domain.Device{
		Name:      strings.TrimSpace(input.Name),
		IPAddress: strings.TrimSpace(input.IPAddress),
		Type:      strings.ToLower(strings.TrimSpace(input.Type)),
		Status:    normalizedStatus(input.Status),
	})
	if !exists {
		apierror.Write(w, http.StatusNotFound, "Resource Not Found", fmt.Sprintf("Device with ID '%s' does not exist in registry", id))
		return
	}

	writeJSON(w, http.StatusOK, updated)
}

func (h *DeviceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	if !h.repo.Delete(id) {
		apierror.Write(w, http.StatusNotFound, "Resource Not Found", fmt.Sprintf("Device with ID '%s' does not exist in registry", id))
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func decodeJSON(r *http.Request, target any) error {
	if r.Body == nil {
		return fmt.Errorf("тіло запиту порожнє")
	}
	decoder := json.NewDecoder(r.Body)
	return decoder.Decode(target)
}

func validateInput(input domain.DeviceInput) error {
	if strings.TrimSpace(input.Name) == "" {
		return fmt.Errorf("поле 'name' є обов'язковим")
	}
	if strings.TrimSpace(input.IPAddress) == "" {
		return fmt.Errorf("поле 'ip_address' є обов'язковим")
	}
	if net.ParseIP(strings.TrimSpace(input.IPAddress)) == nil {
		return fmt.Errorf("поле 'ip_address' має містити коректну IP-адресу")
	}
	if strings.TrimSpace(input.Type) == "" {
		return fmt.Errorf("поле 'type' є обов'язковим")
	}
	if input.Status != "" && input.Status != "online" && input.Status != "offline" {
		return fmt.Errorf("поле 'status' має бути 'online' або 'offline'")
	}
	return nil
}

func normalizedStatus(status string) string {
	if strings.TrimSpace(status) == "" {
		return "offline"
	}
	return strings.ToLower(strings.TrimSpace(status))
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
