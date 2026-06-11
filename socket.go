package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"errors"
)

func CreateSocket() (*net.Listener, error) {
	socketPath := "/tmp/alug_learning.sock"

	// Remove in case it exists
	os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, errors.New("Cannot get a Socket path")
	}

	return &listener, nil
}

func RunSocket(podman *Connection) error {
	listener, err := CreateSocket()
	if err != nil {
		return errors.New(fmt.Sprintf("Failed to create socket: %s", err))
	}
	defer (*listener).Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/inspect/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/inspect/")
		if name == "" {
			http.Error(w, "missing container name", http.StatusBadRequest)
			return
		}

		data, err := podman.Inspect(name)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(data)
	})

	return http.Serve(*listener, mux)
}