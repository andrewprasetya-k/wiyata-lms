package realtime

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

const (
	pongWait     = 60 * time.Second
	pingInterval = 45 * time.Second
)

type Client struct {
	UserID   string
	SchoolID string
	hub      *Hub
	conn     *websocket.Conn
	writeMu  sync.Mutex
}

func NewClient(hub *Hub, conn *websocket.Conn, userID string, schoolID string) *Client {
	return &Client{
		UserID:   userID,
		SchoolID: schoolID,
		hub:      hub,
		conn:     conn,
	}
}

func (c *Client) ReadLoop() {
	defer c.hub.Unregister(c)

	c.conn.SetReadLimit(1024)
	_ = c.conn.SetReadDeadline(time.Now().Add(pongWait))
	c.conn.SetPongHandler(func(string) error {
		return c.conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			if _, _, err := c.conn.NextReader(); err != nil {
				return
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			c.writeMu.Lock()
			err := c.conn.WriteControl(websocket.PingMessage, nil, time.Now().Add(10*time.Second))
			c.writeMu.Unlock()
			if err != nil {
				return
			}
		}
	}
}

func (c *Client) WriteEvent(event Event) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return c.conn.WriteJSON(event)
}

func (c *Client) WriteJSON(payload any) error {
	c.writeMu.Lock()
	defer c.writeMu.Unlock()
	_ = c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	return c.conn.WriteJSON(payload)
}

func (c *Client) Close() {
	_ = c.conn.Close()
}
