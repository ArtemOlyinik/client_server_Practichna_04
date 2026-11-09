package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/ArtemOlyinik/client_server_Practichna_04/internal/domain"
)

var client = &http.Client{Timeout: 5 * time.Second}

func main() {
	baseURL := strings.TrimRight(env("BASE_URL", "http://localhost:8080"), "/")

	fmt.Println("=== Перевірка Device Registry API ===")

	_, createdBody := do(baseURL, http.MethodPost, "/api/v1/devices", map[string]string{
		"name":       "Main Thermal Sensor",
		"ip_address": "192.168.1.50",
		"type":       "sensor",
		"status":     "online",
	})

	var created domain.Device
	if err := json.Unmarshal(createdBody, &created); err != nil {
		log.Fatalf("не вдалося розібрати створений пристрій: %v", err)
	}

	do(baseURL, http.MethodGet, "/api/v1/devices", nil)
	do(baseURL, http.MethodGet, "/api/v1/devices/"+created.ID, nil)
	do(baseURL, http.MethodPut, "/api/v1/devices/"+created.ID, map[string]string{
		"name":       "Main Thermal Sensor 2",
		"ip_address": "192.168.1.51",
		"type":       "sensor",
		"status":     "offline",
	})
	do(baseURL, http.MethodGet, "/api/v1/devices?status=offline&type=sensor", nil)
	do(baseURL, http.MethodDelete, "/api/v1/devices/"+created.ID, nil)
	do(baseURL, http.MethodGet, "/api/v1/devices/"+created.ID, nil)
	do(baseURL, http.MethodPost, "/api/v1/devices", map[string]string{
		"name":       "",
		"ip_address": "192.168.1.80",
		"type":       "sensor",
	})

	response, _ := do(baseURL, http.MethodOptions, "/api/v1/devices", nil)
	fmt.Printf("[CORS] Access-Control-Allow-Origin: %s\n", response.Header.Get("Access-Control-Allow-Origin"))
	fmt.Printf("[CORS] Access-Control-Allow-Methods: %s\n", response.Header.Get("Access-Control-Allow-Methods"))

	fmt.Println("\nПеревірку завершено.")
}

func do(baseURL, method, path string, payload any) (*http.Response, []byte) {
	var body io.Reader
	if payload != nil {
		encoded, err := json.Marshal(payload)
		if err != nil {
			log.Fatalf("не вдалося закодувати JSON: %v", err)
		}
		body = bytes.NewReader(encoded)
	}

	request, err := http.NewRequest(method, baseURL+path, body)
	if err != nil {
		log.Fatalf("не вдалося створити запит: %v", err)
	}
	if payload != nil {
		request.Header.Set("Content-Type", "application/json")
	}

	response, err := client.Do(request)
	if err != nil {
		log.Fatalf("помилка виконання запиту %s %s: %v", method, path, err)
	}
	defer response.Body.Close()

	data, err := io.ReadAll(response.Body)
	if err != nil {
		log.Fatalf("не вдалося прочитати відповідь: %v", err)
	}

	fmt.Printf("\n[HTTP] %s %s -> %d %s\n", method, path, response.StatusCode, http.StatusText(response.StatusCode))
	if len(data) > 0 {
		var pretty bytes.Buffer
		if err := json.Indent(&pretty, data, "", "  "); err == nil {
			fmt.Println(pretty.String())
		} else {
			fmt.Println(string(data))
		}
	}

	return response, data
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
