package web

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/zamyatin-zkex/volumer/internal/event"
	"net/http"
)

type Server struct {
	web    *http.Server
	keeper *keeper
	state  *state
}

func New(addr string) *Server {
	serv := &Server{
		web: &http.Server{
			Addr: addr,
		},
		keeper: newKeeper(),
		state:  newState(),
	}
	serv.web.Handler = serv.router()
	return serv
}

func (s *Server) Run(ctx context.Context) error {
	closed := make(chan error, 1)

	go func() {
		closed <- s.web.ListenAndServe()
	}()

	select {
	case err := <-closed:
		return err
	case <-ctx.Done():
		_ = s.web.Shutdown(ctx)
		return ctx.Err()
	}
}

func (s *Server) UpdateStats(ctx context.Context, stats event.StatsUpdated) error {
	s.state.update(stats.Tokens)

	err := s.keeper.walkSubs(func(conn *websocket.Conn, subs map[string]struct{}) error {
		for sub := range subs {
			frames := make([]Frame, 0)
			for i, v := range stats.Tokens[sub] {
				frames = append(frames, Frame{
					Interval: i,
					Volume:   v,
				})
			}

			msg := NewMessage(TokenStats{Token: sub, Frames: frames})
			js, err := json.MarshalIndent(msg, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal json: %w", err)
			}

			return conn.WriteMessage(websocket.TextMessage, js)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("walk subs: %w", err)
	}

	return nil
}
