package main

import (
	"log"
	"net/http"

	"line-adaptor/internal/config"
	"line-adaptor/internal/handler"
	"line-adaptor/internal/line/content"
	"line-adaptor/internal/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}

	lgr := logger.New(cfg.LogDir)
	h := handler.New(cfg.ChannelSecret, cfg.ChannelAccessToken, lgr, content.New(cfg.ChannelAccessToken))

	http.HandleFunc("/webhook", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		h.Webhook(w, r)
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	log.Printf("starting server on :%s", cfg.Port)
	log.Fatal(http.ListenAndServe(":"+cfg.Port, nil))

}
