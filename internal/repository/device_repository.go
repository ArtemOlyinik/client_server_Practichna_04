package repository

import (
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/ArtemOlyinik/client_server_Practichna_04/internal/domain"
)

// DeviceRepository описує контракт сховища пристроїв.
type DeviceRepository interface {
	GetAll(filter domain.DeviceFilter) []domain.Device
	GetByID(id string) (domain.Device, bool)
	Create(device domain.Device) (domain.Device, error)
	Update(id string, device domain.Device) (domain.Device, bool)
	Delete(id string) bool
}

// MemoryDeviceRepository зберігає пристрої в пам'яті.
// RWMutex дозволяє багатьом запитам читати дані одночасно.
type MemoryDeviceRepository struct {
	mu      sync.RWMutex
	storage map[string]domain.Device
	nextID  uint64
}

func NewMemoryDeviceRepository() *MemoryDeviceRepository {
	return &MemoryDeviceRepository{
		storage: make(map[string]domain.Device),
	}
}

func (r *MemoryDeviceRepository) GetAll(filter domain.DeviceFilter) []domain.Device {
	r.mu.RLock()
	defer r.mu.RUnlock()

	devices := make([]domain.Device, 0, len(r.storage))
	for _, device := range r.storage {
		if filter.Status != "" && !strings.EqualFold(device.Status, filter.Status) {
			continue
		}
		if filter.Type != "" && !strings.EqualFold(device.Type, filter.Type) {
			continue
		}
		devices = append(devices, device)
	}

	sort.Slice(devices, func(i, j int) bool {
		return devices[i].ID < devices[j].ID
	})
	return devices
}

func (r *MemoryDeviceRepository) GetByID(id string) (domain.Device, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	device, exists := r.storage[id]
	return device, exists
}

func (r *MemoryDeviceRepository) Create(device domain.Device) (domain.Device, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.nextID++
	device.ID = fmt.Sprintf("dev-%03d", r.nextID)
	device.UpdatedAt = time.Now().UTC()
	r.storage[device.ID] = device
	return device, nil
}

func (r *MemoryDeviceRepository) Update(id string, device domain.Device) (domain.Device, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, exists := r.storage[id]
	if !exists {
		return domain.Device{}, false
	}

	existing.Name = device.Name
	existing.IPAddress = device.IPAddress
	existing.Type = device.Type
	existing.Status = device.Status
	existing.UpdatedAt = time.Now().UTC()
	r.storage[id] = existing

	return existing, true
}

func (r *MemoryDeviceRepository) Delete(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.storage[id]; !exists {
		return false
	}
	delete(r.storage, id)
	return true
}
