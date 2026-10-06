package websocket

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/application/port/out"
	"github.com/carlosmorales-dev-mx/ticketing-system/api/internal/domain/shared"
)

// client envuelve la conexión con su propio candado de escritura:
// gorilla/websocket NO permite escrituras concurrentes sobre una misma
// conexión, y ahora un reinicio de mapa emite decenas de mensajes en
// ráfaga mientras otras reservas también publican.
type client struct {
	conn *websocket.Conn
	wmu  sync.Mutex
}

const writeWait = 5 * time.Second

func (c *client) send(payload []byte) error {
	c.wmu.Lock()
	defer c.wmu.Unlock()
	_ = c.conn.SetWriteDeadline(time.Now().Add(writeWait))
	return c.conn.WriteMessage(websocket.TextMessage, payload)
}

type seatUpdateMessage struct {
	EventID string `json:"event_id"`
	SeatID  string `json:"seat_id"`
	Status  string `json:"status"`
}

// Hub mantiene las conexiones activas agrupadas por evento (así solo
// se notifica a quien está viendo ese evento, no a todo el mundo).
type Hub struct {
	mu    sync.RWMutex
	conns map[string]map[*client]struct{} // eventID -> set de conexiones
}

func NewHub() *Hub {
	return &Hub{conns: make(map[string]map[*client]struct{})}
}

// upgrader controla la actualización HTTP -> WebSocket.
//
// CheckOrigin: en producción esto debe validar contra una lista de
// orígenes permitidos (el dominio del frontend Nuxt). Por defecto,
// aquí se restringe a localhost de desarrollo; nunca se debe dejar
// devolviendo "true" sin condiciones en un despliegue real, porque
// abriría el WebSocket a cualquier sitio (CSWSH).
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return origin == "" || origin == "http://localhost:3000" || origin == "http://localhost:8080"
	},
}

// ServeWS registra una nueva conexión para un evento concreto.
func (h *Hub) ServeWS(w http.ResponseWriter, r *http.Request, eventID string) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("error actualizando a websocket: %v", err)
		return
	}

	cl := &client{conn: conn}
	h.mu.Lock()
	if h.conns[eventID] == nil {
		h.conns[eventID] = make(map[*client]struct{})
	}
	h.conns[eventID][cl] = struct{}{}
	h.mu.Unlock()

	// Bucle de lectura: no esperamos mensajes del cliente, pero hay
	// que leer igualmente para detectar cierre de conexión (si no,
	// la conexión se queda "zombie" consumiendo memoria).
	go func() {
		defer func() {
			h.mu.Lock()
			delete(h.conns[eventID], cl)
			h.mu.Unlock()
			conn.Close()
		}()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				return
			}
		}
	}()
}

func (h *Hub) BroadcastSeatUpdate(ctx context.Context, eventID, seatID shared.ID, status string) error {
	payload, err := json.Marshal(seatUpdateMessage{
		EventID: eventID.String(),
		SeatID:  seatID.String(),
		Status:  status,
	})
	if err != nil {
		return err
	}

	h.mu.RLock()
	defer h.mu.RUnlock()

	for cl := range h.conns[eventID.String()] {
		// Best-effort: si un cliente concreto falla, no interrumpimos
		// la notificación al resto.
		if err := cl.send(payload); err != nil {
			log.Printf("aviso: error enviando a un cliente websocket: %v", err)
		}
	}
	return nil
}

var _ out.RealtimeNotifier = (*Hub)(nil)
