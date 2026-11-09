package repository

import (
	"fmt"
	"sync"
	"testing"

	"github.com/ArtemOlyinik/client_server_Practichna_04/internal/domain"
)

func TestMemoryRepositoryCRUD(t *testing.T) {
	repo := NewMemoryDeviceRepository()

	created, err := repo.Create(domain.Device{
		Name:      "Main Thermal Sensor",
		IPAddress: "192.168.1.50",
		Type:      "sensor",
		Status:    "online",
	})
	if err != nil {
		t.Fatalf("не вдалося створити пристрій: %v", err)
	}
	if created.ID == "" {
		t.Fatal("репозиторій не згенерував ID")
	}

	found, exists := repo.GetByID(created.ID)
	if !exists || found.Name != "Main Thermal Sensor" {
		t.Fatal("створений пристрій не знайдено")
	}

	updated, exists := repo.Update(created.ID, domain.Device{
		Name:      "Main Thermal Sensor 2",
		IPAddress: "192.168.1.51",
		Type:      "sensor",
		Status:    "offline",
	})
	if !exists || updated.Name != "Main Thermal Sensor 2" || updated.Status != "offline" {
		t.Fatal("пристрій не оновився")
	}

	if !repo.Delete(created.ID) {
		t.Fatal("пристрій не видалився")
	}
	if _, exists := repo.GetByID(created.ID); exists {
		t.Fatal("видалений пристрій залишився в сховищі")
	}
}

func TestMemoryRepositoryFilter(t *testing.T) {
	repo := NewMemoryDeviceRepository()

	_, _ = repo.Create(domain.Device{Name: "Sensor A", IPAddress: "10.0.0.1", Type: "sensor", Status: "online"})
	_, _ = repo.Create(domain.Device{Name: "Sensor B", IPAddress: "10.0.0.2", Type: "sensor", Status: "offline"})
	_, _ = repo.Create(domain.Device{Name: "Gateway A", IPAddress: "10.0.0.3", Type: "gateway", Status: "online"})

	online := repo.GetAll(domain.DeviceFilter{Status: "online"})
	if len(online) != 2 {
		t.Fatalf("очікували 2 online-пристрої, отримали %d", len(online))
	}

	sensors := repo.GetAll(domain.DeviceFilter{Type: "sensor", Status: "online"})
	if len(sensors) != 1 || sensors[0].Name != "Sensor A" {
		t.Fatalf("фільтр за типом і статусом працює некоректно: %+v", sensors)
	}
}

func TestMemoryRepositoryConcurrentAccess(t *testing.T) {
	repo := NewMemoryDeviceRepository()
	var waitGroup sync.WaitGroup

	for i := 0; i < 100; i++ {
		waitGroup.Add(1)
		go func(index int) {
			defer waitGroup.Done()
			_, _ = repo.Create(domain.Device{
				Name:      fmt.Sprintf("Device %d", index),
				IPAddress: "10.0.0.1",
				Type:      "sensor",
				Status:    "online",
			})
			_ = repo.GetAll(domain.DeviceFilter{})
		}(i)
	}

	waitGroup.Wait()
	if got := len(repo.GetAll(domain.DeviceFilter{})); got != 100 {
		t.Fatalf("очікували 100 пристроїв, отримали %d", got)
	}
}
