package handlers

import (
	"fmt"
	"io"
	"log"
	"net/http"
)

func StartServer() {
	http.HandleFunc("/request-data", requestDataHandler)
	http.HandleFunc("/receive-data", receiveDataHandler)

	fmt.Println("Web Service A running on http://localhost:8081")
	log.Fatal(http.ListenAndServe(":8081", nil))
}

func requestDataHandler(w http.ResponseWriter, r *http.Request) {
	// Send request to Web Service B to show the form
	resp, err := http.Get("http://localhost:8082/show-form")
	if err != nil {
		http.Error(w, "Failed to request data from Web Service B", http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	io.Copy(w, resp.Body)
}

func receiveDataHandler(w http.ResponseWriter, r *http.Request) {
	// Receive completed form data from Web Service B
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read data", http.StatusInternalServerError)
		return
	}
	defer r.Body.Close()

	// Process the received data (log it for now)
	fmt.Printf("Received data from Web Service B: %s\n", body)
	w.Write([]byte("Data received successfully by Web Service A"))
}
