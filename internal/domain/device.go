package domain

import "time"

// Device описує вузол мережі в реєстрі.
type Device struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	IPAddress string    `json:"ip_address"`
	Type      string    `json:"type"`
	Status    string    `json:"status"`
	UpdatedAt time.Time `json:"updated_at"`
}

// DeviceInput приймає дані від клієнта під час створення або оновлення.
type DeviceInput struct {
	Name      string `json:"name"`
	IPAddress string `json:"ip_address"`
	Type      string `json:"type"`
	Status    string `json:"status"`
}

// DeviceFilter задає необов'язкові умови фільтрації списку.
type DeviceFilter struct {
	Status string
	Type   string
}
