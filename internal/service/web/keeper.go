package web

import (
	"github.com/gorilla/websocket"
	"sync"
	"time"
)

type keeper struct {
	mx     sync.RWMutex
	active map[*websocket.Conn]struct{}
	subs   map[*websocket.Conn]map[string]struct{}
}

func newKeeper() *keeper {
	return &keeper{
		active: make(map[*websocket.Conn]struct{}),
		subs:   make(map[*websocket.Conn]map[string]struct{}),
	}
}

func (k *keeper) addConn(conn *websocket.Conn) {
	k.mx.Lock()
	defer k.mx.Unlock()
	k.active[conn] = struct{}{}
	k.subs[conn] = make(map[string]struct{})
}
func (k *keeper) walkSubs(fn func(conn *websocket.Conn, tokens map[string]struct{}) error) error {
	k.mx.RLock()
	defer k.mx.RUnlock()

	for conn, tokens := range k.subs {
		if err := fn(conn, tokens); err != nil {
			return err
		}
	}

	return nil
}

func (k *keeper) close(conn *websocket.Conn) {
	k.mx.Lock()
	defer k.mx.Unlock()

	_ = conn.Close()
	delete(k.active, conn)
	delete(k.subs, conn)
}

func (k *keeper) keep(conn *websocket.Conn) {
	pinger := time.NewTicker(time.Second)
	defer pinger.Stop()

	lastAlive := time.Now()
	const deadlineSeconds = 5
	read := make(chan msg)
	defer k.close(conn)

	ponger := conn.PongHandler()
	conn.SetPongHandler(func(appData string) error {
		lastAlive = time.Now()
		return ponger(appData)
	})

	go func() {
		for {
			mt, data, err := conn.ReadMessage()
			read <- msg{
				mType: mt,
				data:  data,
				err:   err,
			}
			if err != nil {
				close(read)
				return
			}
		}
	}()

	for {
		select {
		case <-pinger.C:
			if err := conn.WriteControl(websocket.PingMessage, []byte{}, time.Now().Add(time.Second)); err != nil {
				return
			}
			if time.Since(lastAlive).Seconds() > deadlineSeconds {
				return
			}
		case msg, ok := <-read:
			if !ok {
				return
			}

			if msg.err != nil {
				return
			}

			switch msg.mType {
			case websocket.CloseMessage:
				return
			case websocket.TextMessage:
				token := string(msg.data)
				if token == "" {
					continue
				}
				k.mx.Lock()
				k.subs[conn][token] = struct{}{}
				k.mx.Unlock()
			}

			lastAlive = time.Now()
		}
	}
}

//
//func (k *keeper) run(ctx context.Context) error {
//	pinger := time.NewTicker(time.Second)
//	defer pinger.Stop()
//
//	for {
//		select {
//		case <-pinger.C:
//			err := k.walk(func(conn *websocket.Conn) error {
//				return conn.WriteMessage(websocket.PingMessage, nil)
//			})
//			if err != nil {
//				return fmt.Errorf("ping: %w", err)
//			}
//		case <-ctx.Done():
//			return ctx.Err()
//		}
//	}
//}
