package web

import (
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

type Web struct {
	hub *Hub
	log *log.Entry
}

func InitWeb(webAddr string) *Web {
	log.Debugf("InitWeb")

	hub := newHub()
	go hub.run()

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "pkg/web/index.html")
	})
	http.HandleFunc("/jquery-1.11.1.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "pkg/web/jquery-1.11.1.js")
	})
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWs(hub, w, r)
	})
	go func() {
		err := http.ListenAndServe(webAddr, nil) // "192.168.1.1:8080"
		if err != nil {
			log.Fatal("ListenAndServe: ", err)
		}
	}()

	return &Web{
		hub: hub,
		log: log.WithFields(log.Fields{
			"web": webAddr,
		}),
	}
}

func (w *Web) Render(msg string) {
	b, err := json.Marshal(msg)
	if err != nil {
		w.log.Error(err)
		return
	}

	go func() { w.hub.render <- b }()
}

func (w *Web) Phone() <-chan PhoneEvent {
	return w.hub.phoneEvents
}
