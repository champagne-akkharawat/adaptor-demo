package main

import (
	"log"
	"net/http"

	"aura-wellness/line-adaptor/internal/config"
	"aura-wellness/line-adaptor/internal/handler"
	"aura-wellness/line-adaptor/internal/line"
)

func main() {
	cfg := config.Load()

	lineClient := line.NewClient(cfg.ChannelAccessToken)
	webhookHandler := handler.New(cfg.ChannelSecret, lineClient)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("ok"))
	})
	mux.Handle("POST /webhook", webhookHandler)

	log.Printf("LINE adaptor listening on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, mux); err != nil {
		log.Fatal(err)
	}
}
