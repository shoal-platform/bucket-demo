package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"

	"cloud.google.com/go/storage"
	"google.golang.org/api/iterator"
)

type FileItem struct {
	Name string `json:"name"`
	Size int64 `json:"size"`
	Updated string `json:"updated"`
	ContentType string `json:"contentType"`
}

func main() {
	bucketName := os.Getenv("BUCKET_NAME")
	if bucketName == "" {
		log.Fatal("BUCKET_NAME is required")
	}

	ctx := context.Background()
	client, err := storage.NewClient(ctx)
	if err != nil { log.Fatal(err) }
	defer client.Close()

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) { writeJSON(w, map[string]string{"status": "ok"}) })
	http.HandleFunc("/items", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
		query := client.Bucket(bucketName).Objects(ctx, nil)
		var items []FileItem
		it := query
		for {
			obj, err := it.Next()
			if err == iterator.Done { break }
			if err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
			items = append(items, FileItem{Name: obj.Name, Size: obj.Size, Updated: obj.Updated.Format("2006-01-02T15:04:05Z07:00"), ContentType: obj.ContentType})
		}
		if items == nil { items = []FileItem{} }
		writeJSON(w, items)
	})
	http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
		if err := r.ParseMultipartForm(10 << 20); err != nil { http.Error(w, "invalid upload: "+err.Error(), http.StatusBadRequest); return }
		file, header, err := r.FormFile("file")
		if err != nil { http.Error(w, "choose a text file", http.StatusBadRequest); return }
		defer file.Close()
		name := safeName(header)
		obj := client.Bucket(bucketName).Object(name)
		writer := obj.NewWriter(ctx)
		writer.ContentType = "text/plain; charset=utf-8"
		if _, err = io.Copy(writer, file); err != nil { writer.Close(); http.Error(w, err.Error(), http.StatusInternalServerError); return }
		if err = writer.Close(); err != nil { http.Error(w, err.Error(), http.StatusInternalServerError); return }
		writeJSON(w, map[string]string{"name": name})
	})
	http.HandleFunc("/download", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet { http.Error(w, "method not allowed", http.StatusMethodNotAllowed); return }
		name, err := url.QueryUnescape(r.URL.Query().Get("name"))
		if err != nil || name == "" || strings.Contains(name, "..") { http.Error(w, "invalid file name", http.StatusBadRequest); return }
		reader, err := client.Bucket(bucketName).Object(name).NewReader(ctx)
		if err != nil { http.Error(w, "file not found", http.StatusNotFound); return }
		defer reader.Close()
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", path.Base(name)))
		io.Copy(w, reader)
	})

	port := os.Getenv("PORT")
	if port == "" { port = "8080" }
	log.Printf("text vault API listening on %s", port)
	log.Fatal(http.ListenAndServe(":"+port, withCORS(http.DefaultServeMux)))
}

func safeName(header *multipart.FileHeader) string {
	name := path.Base(strings.TrimSpace(header.Filename))
	if name == "." || name == "/" || name == "" { return "untitled.txt" }
	return name
}

func writeJSON(w http.ResponseWriter, value any) { w.Header().Set("Content-Type", "application/json"); json.NewEncoder(w).Encode(value) }

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		if r.Method == http.MethodOptions { w.WriteHeader(http.StatusNoContent); return }
		next.ServeHTTP(w, r)
	})
}
