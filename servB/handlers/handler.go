package handlers

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Active WebSocket connections
var clients = make(map[*websocket.Conn]bool)
var mu sync.Mutex

// Counter to track how many times the form has been shown
var formCounter int
var counterMutex sync.Mutex

func StartServer() {
	http.HandleFunc("/ws", websocketHandler) // WebSocket endpoint
	http.HandleFunc("/show-form", showFormHandler)
	http.HandleFunc("/submit-form", submitFormHandler)

	fmt.Println("Web Service B running on http://localhost:8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}

func websocketHandler(w http.ResponseWriter, r *http.Request) {
	// Upgrade HTTP connection to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket Upgrade Error:", err)
		return
	}
	defer conn.Close()

	mu.Lock()
	clients[conn] = true
	mu.Unlock()

	log.Println("New WebSocket client connected")

	// Listen for disconnect
	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			mu.Lock()
			delete(clients, conn)
			mu.Unlock()
			log.Println("WebSocket client disconnected")
			break
		}
	}
}

func notifyClients(message string) {
	mu.Lock()
	defer mu.Unlock()

	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, []byte(message))
		if err != nil {
			log.Println("WebSocket Write Error:", err)
			client.Close()
			delete(clients, client)
		}
	}
}

func showFormHandler(w http.ResponseWriter, r *http.Request) {
	// Determine which form to show based on the counter
	counterMutex.Lock()
	defer counterMutex.Unlock()

	var form string
	if formCounter%2 == 0 {
		form = `
		<!DOCTYPE html>
		<html>
		<body>
			<script>
				// Connect to WebSocket
				const ws = new WebSocket("ws://localhost:8082/ws");

				ws.onmessage = (event) => {
					alert(event.data); // Show notification
					location.reload(); // Reload page to fetch new form
				};
			</script>
			<h2>Form 1</h2>
			<form action="http://localhost:8082/submit-form" method="POST">
				<label for="name">Name:</label><br>
				<input type="text" id="name" name="name"><br>
				<label for="email">Email:</label><br>
				<input type="email" id="email" name="email"><br><br>
				<input type="submit" value="Submit">
			</form>
		</body>
		</html>`
	} else {
		form = `
		<!DOCTYPE html>
		<html>
		<body>
			<script>
				// Connect to WebSocket
				const ws = new WebSocket("ws://localhost:8082/ws");

				ws.onmessage = (event) => {
					alert(event.data); // Show notification
					location.reload(); // Reload page to fetch new form
				};
			</script>
			<h2>Form 2</h2>
			<form action="http://localhost:8082/submit-form" method="POST">
				<label for="age">Age:</label><br>
				<input type="number" id="age" name="age"><br>
				<label for="country">Country:</label><br>
				<input type="text" id="country" name="country"><br><br>
				<input type="submit" value="Submit">
			</form>
		</body>
		</html>`
	}

	// Write the selected form to the response
	w.Header().Set("Content-Type", "text/html")
	w.Write([]byte(form))
}

func submitFormHandler(w http.ResponseWriter, r *http.Request) {
	// Read the submitted form data
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read form data", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Send the completed form data to Web Service A
	resp, err := http.Post("http://localhost:8081/receive-data", "application/x-www-form-urlencoded", bytes.NewBuffer(body))
	if err != nil {
		http.Error(w, "Failed to send data to Web Service A", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	// Increment the form counter
	counterMutex.Lock()
	formCounter++
	counterMutex.Unlock()

	// Notify all connected WebSocket clients
	go notifyClients("A new form is now available!")

	// Notify the user that the form was submitted successfully
	w.Write([]byte("Form submitted successfully!"))
}
