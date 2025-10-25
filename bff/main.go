package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
)

type Event struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

var (
	clients    = make(map[chan Event]bool)
	clientsMu  sync.Mutex
	lastEvent  *Event       // <-- ultimo evento ricevuto
	lastEventM sync.RWMutex // mutex per accesso concorrente
)

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	http.HandleFunc("/events", handleIncomingEvent)
	http.HandleFunc("/stream", handleStream)

	log.Printf("Service A listening on :%s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}

func handleIncomingEvent(w http.ResponseWriter, r *http.Request) {
	var e Event
	if err := json.NewDecoder(r.Body).Decode(&e); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// salva l'ultimo evento ricevuto
	lastEventM.Lock()
	lastEvent = &e
	lastEventM.Unlock()

	broadcast(e)
	w.WriteHeader(http.StatusAccepted)
}

// Funzione SSE con supporto CORS + invio ultimo evento
func handleStream(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	eventChan := make(chan Event, 1)
	clientsMu.Lock()
	clients[eventChan] = true
	clientsMu.Unlock()

	defer func() {
		clientsMu.Lock()
		delete(clients, eventChan)
		clientsMu.Unlock()
		close(eventChan)
	}()

	// ✅ Invia subito l’ultimo evento, se esiste
	lastEventM.RLock()
	if lastEvent != nil {
		data, _ := json.Marshal(lastEvent)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}
	lastEventM.RUnlock()

	log.Printf("New SSE client connected (%d total)", len(clients))
	for event := range eventChan {
		data, _ := json.Marshal(event)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}
}

// broadcast manda un evento a tutti i client
func broadcast(event Event) {
	clientsMu.Lock()
	defer clientsMu.Unlock()

	for c := range clients {
		select {
		case c <- event:
		default:
			close(c)
			delete(clients, c)
		}
	}
}
