package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
)

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/transcribe", transcribeHandler)
	http.ListenAndServe(":8080", nil)
}

func transcribeHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	file, _, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, "failed to get audio: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	tmp, err := os.CreateTemp("", "recording-*.webm")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer os.Remove(tmp.Name())
	if _, err := io.Copy(tmp, file); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	tmp.Close()

	cmd := exec.Command("whisper", tmp.Name(), "--model", "base")
	out, err := cmd.CombinedOutput()
	if err != nil {
		http.Error(w, "whisper error: "+err.Error()+": "+string(out), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"text": string(out)})
}
