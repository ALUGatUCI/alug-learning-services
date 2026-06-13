package main

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
)

func CreateSocket() (*net.Listener, error) {
	socketPath := "/tmp/alug_learning.sock"

	// Remove in case it exists
	os.Remove(socketPath)

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return nil, fmt.Errorf("Cannot get a Socket path: %w", err)
	}

	return &listener, nil
}

func RunSocket(podman *Connection) error {
	listener, err := CreateSocket()
	if err != nil {
		return fmt.Errorf("Failed to create socket: %w", err)
	}
	defer (*listener).Close()

	mux := http.NewServeMux()

	mux.HandleFunc("/inspect/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

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

		json.NewEncoder(w).Encode(data)
	})

	mux.HandleFunc("/provision/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/provision/")
		if name == "" {
			http.Error(w, "missing container name", http.StatusBadRequest)
			return
		}

		imagePath := os.Getenv("IMAGE")
		if imagePath == "" {
			http.Error(w, "IMAGE is not set", http.StatusInternalServerError)
			return
		}

		container, err := podman.CreateContainer(imagePath, name);
		if err != nil {
			http.Error(w, "failed to create container: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"status": "created",
			"name":   container.name,
			"sshPort": strconv.Itoa(int(container.sshPort)),
		})
	})

	return http.Serve(*listener, mux)
}
