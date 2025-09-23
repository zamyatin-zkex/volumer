package web

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
)

func (s *Server) router() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		s.keeper.addConn(conn)
		go s.keeper.keep(conn)
	})

	mux.HandleFunc("/stats", func(w http.ResponseWriter, r *http.Request) {
		token := r.URL.Query().Get("token")
		if token == "" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		stats := s.state.get(token)
		js, err := json.MarshalIndent(stats, "", "  ")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(js)
	})

	return mux
}
