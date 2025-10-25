package main

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"
)

type Event struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

func main() {
	serviceAURL := os.Getenv("SERVICE_A_URL")
	if serviceAURL == "" {
		serviceAURL = "http://localhost:8080/events"
	}

	for {
		time.Sleep(5 * time.Second)
		event := Event{
			ID:      time.Now().Format("150405"),
			Status:  "completed",
			Message: "Process finished successfully",
		}
		payload, _ := json.Marshal(event)
		resp, err := http.Post(serviceAURL, "application/json", bytes.NewBuffer(payload))
		if err != nil {
			log.Printf("Error sending event: %v", err)
			continue
		}
		resp.Body.Close()
		log.Printf("Sent event to %s", serviceAURL)
	}
}
