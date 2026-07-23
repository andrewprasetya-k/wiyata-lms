package service

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

var ErrWSTicketInvalid = errors.New("ws ticket is invalid, expired, or already used")

// wsTicketTTL is deliberately very short — a ticket only needs to survive
// the moment between "frontend fetches it" and "frontend opens the
// WebSocket," not any longer.
const wsTicketTTL = 60 * time.Second

type WSTicketService interface {
	// Issue mints a new single-use ticket for userID and returns the raw
	// (unhashed) value — only its hash is ever stored.
	Issue(userID string) (string, error)
	// Consume validates+deletes the ticket in one atomic step (so it can
	// never be used twice) and returns the userID it was issued for.
	Consume(rawTicket string) (string, error)
}

type wsTicketEntry struct {
	userID    string
	expiresAt time.Time
}

// wsTicketService is a deliberately simple in-memory, single-use ticket
// store — not a DB table like email_verifications/password_reset_tokens/
// refresh_tokens. A WS ticket is extremely short-lived (60s) and low-value
// (it only ever grants one WebSocket handshake, nothing more), so a
// migration + DB round-trip for something that's gone within a minute
// either way isn't worth it, and it doesn't need to survive a server
// restart — a ticket lost on restart just means the client's next connect
// attempt fetches a new one, same as any other in-flight request would be
// interrupted by a restart. If this app is ever scaled horizontally, this
// would need to move to a shared store (Redis, or a real table) — same
// caveat already noted for InMemoryRateLimiterStore.
type wsTicketService struct {
	mu      sync.Mutex
	tickets map[string]wsTicketEntry
}

func NewWSTicketService() WSTicketService {
	store := &wsTicketService{tickets: make(map[string]wsTicketEntry)}
	go store.sweepLoop()
	return store
}

func (s *wsTicketService) Issue(userID string) (string, error) {
	rawTicket, hash, err := generateWSTicket()
	if err != nil {
		return "", err
	}

	s.mu.Lock()
	s.tickets[hash] = wsTicketEntry{userID: userID, expiresAt: time.Now().Add(wsTicketTTL)}
	s.mu.Unlock()

	return rawTicket, nil
}

func (s *wsTicketService) Consume(rawTicket string) (string, error) {
	hash, err := hashWSTicket(rawTicket)
	if err != nil {
		return "", ErrWSTicketInvalid
	}

	s.mu.Lock()
	entry, exists := s.tickets[hash]
	if exists {
		delete(s.tickets, hash) // single-use: removed on first read regardless of outcome below
	}
	s.mu.Unlock()

	if !exists || time.Now().After(entry.expiresAt) {
		return "", ErrWSTicketInvalid
	}
	return entry.userID, nil
}

func (s *wsTicketService) sweepLoop() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()
	for now := range ticker.C {
		s.mu.Lock()
		for hash, entry := range s.tickets {
			if now.After(entry.expiresAt) {
				delete(s.tickets, hash)
			}
		}
		s.mu.Unlock()
	}
}

func generateWSTicket() (string, string, error) {
	ticketBytes := make([]byte, 32)
	if _, err := rand.Read(ticketBytes); err != nil {
		return "", "", err
	}
	rawTicket := base64.RawURLEncoding.EncodeToString(ticketBytes)
	sum := sha256.Sum256([]byte(rawTicket))
	return rawTicket, hex.EncodeToString(sum[:]), nil
}

func hashWSTicket(ticket string) (string, error) {
	if ticket == "" {
		return "", ErrWSTicketInvalid
	}
	sum := sha256.Sum256([]byte(ticket))
	return hex.EncodeToString(sum[:]), nil
}
