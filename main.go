package main

import (
	"embed"
	"io/fs"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/k/tictactoe-rl/server"
)

//go:embed web/*
var webFS embed.FS

func main() {
	modelDir := "models"
	if d := os.Getenv("MODEL_DIR"); d != "" {
		modelDir = d
	}
	modelDir, _ = filepath.Abs(modelDir)

	addr := ":8080"
	if a := os.Getenv("ADDR"); a != "" {
		addr = a
	}

	handler := server.NewHandler(modelDir)

	webContent, _ := fs.Sub(webFS, "web")
	http.Handle("/", http.FileServer(http.FS(webContent)))
	http.HandleFunc("/ws", handler.ServeWS)

	log.Printf("starting server on %s", addr)
	log.Fatal(http.ListenAndServe(addr, nil))
}
