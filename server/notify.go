package main

import (
	"net"
	"net/http"

	"github.com/gobwas/ws"
	"github.com/gobwas/ws/wsutil"
)

var connPool = map[*net.Conn]bool{}

// NotifyListeners notifies all the clients that have subscribed for PushNotifs
// that the data in the database has changed.
func NotifyListeners() {
	for conn, active := range connPool {
		if active {
			wsutil.WriteServerMessage(*conn, ws.OpText, []byte("Data changed."))
		}
	}
}

// PushNotifs establishes a socket to notify clients that the data has changed.
func PushNotifs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	conn, _, _, err := ws.UpgradeHTTP(r, w)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	connPool[&conn] = true

	go func() {
		defer func() {
			connPool[&conn] = false
			conn.Close()
		}()

		for {
			_, op, err := wsutil.ReadClientData(conn)
			if err != nil {
				if _, ok := err.(net.Error); ok {
					break
				}

				continue
			}

			if op == ws.OpClose {
				break
			}
		}
	}()
}
