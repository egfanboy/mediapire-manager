package websocket

import (
	"encoding/json"
	"net/http"
	"strings"

	"github.com/egfanboy/mediapire-manager/pkg/types"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
)

var allowedOrigins = make(map[string]struct{})
var allowAllOrigins = true

var upgrader = websocket.Upgrader{
	CheckOrigin: checkOrigin,
}

type Client struct {
	conn *websocket.Conn
	send chan []byte
}

type Hub struct {
	clients    map[*Client]bool
	register   chan *Client
	unregister chan *Client
	broadcast  chan []byte
}

var hub = Hub{
	clients:    make(map[*Client]bool),
	register:   make(chan *Client),
	unregister: make(chan *Client),
	broadcast:  make(chan []byte),
}

func (h *Hub) run() {
	for {
		select {
		case c := <-h.register:
			h.clients[c] = true
		case c := <-h.unregister:
			if _, ok := h.clients[c]; ok {
				delete(h.clients, c)
				close(c.send)
			}
		case msg := <-h.broadcast:
			for c := range h.clients {
				select {
				case c.send <- msg:
				default:
					delete(h.clients, c)
					close(c.send)
				}
			}
		}
	}
}

func (c *Client) readPump() {
	defer func() {
		hub.unregister <- c
		c.conn.Close()
	}()
	for {
		_, _, err := c.conn.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (c *Client) writePump() {
	defer c.conn.Close()
	for msg := range c.send {
		if err := c.conn.WriteMessage(websocket.TextMessage, msg); err != nil {
			break
		}
	}
}

func wsHandler(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	client := &Client{conn: conn, send: make(chan []byte, 256)}
	hub.register <- client

	go client.writePump()
	go client.readPump()
}

func setAllowedOrigins(origins []string) {
	allowed := make(map[string]struct{})

	for _, origin := range origins {
		trimmed := strings.TrimSpace(origin)
		if trimmed == "" {
			continue
		}
		allowed[trimmed] = struct{}{}
	}

	allowedOrigins = allowed
	allowAllOrigins = len(allowedOrigins) == 0
}

func checkOrigin(r *http.Request) bool {
	origin := strings.TrimSpace(r.Header.Get("Origin"))

	if allowAllOrigins {
		return true
	}

	if _, ok := allowedOrigins[origin]; ok {
		return true
	}

	return false
}

func RegisterWebSocketHandler(router *mux.Router, origins []string) {
	setAllowedOrigins(origins)
	go hub.run()
	router.HandleFunc("/ws", wsHandler)
}

func SendMessage(msg []byte) {
	hub.broadcast <- msg
}

type playbackSessionUpdatedEnvelope struct {
	Type  string                     `json:"type"`
	State types.PlaybackSessionState `json:"state"`
}

func SendPlaybackSessionUpdated(state types.PlaybackSessionState) error {
	payload, err := json.Marshal(playbackSessionUpdatedEnvelope{
		Type:  "playback.session.updated",
		State: state,
	})
	if err != nil {
		return err
	}

	SendMessage(payload)

	return nil
}
