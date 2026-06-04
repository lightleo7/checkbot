package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type Hub struct {
	viewers    map[*websocket.Conn]bool
	broadcast  chan []byte
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
	mu         sync.Mutex
}

var hub = Hub{
	viewers:    make(map[*websocket.Conn]bool),
	broadcast:  make(chan []byte, 10), // Буфер кадров
	register:   make(chan *websocket.Conn),
	unregister: make(chan *websocket.Conn),
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.viewers[client] = true
			h.mu.Unlock()
			fmt.Println("[WS] Новый зритель подключился к стриму")

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.viewers[client]; ok {
				delete(h.viewers, client)
				client.Close()
				fmt.Println("[WS] Зритель отключился")
			}
			h.mu.Unlock()

		case frameBytes := <-h.broadcast:
			h.mu.Lock()
			for client := range h.viewers {
				err := client.WriteMessage(websocket.BinaryMessage, frameBytes)
				if err != nil {
					log.Println("[WS] Ошибка отправки зрителю, закрываем соединение:", err)
					client.Close()
					delete(h.viewers, client)
				}
			}
			h.mu.Unlock()
		}
	}
}

func handleStreamer(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("[WS] Ошибка апгрейда стримера:", err)
		return
	}
	defer conn.Close()
	fmt.Println("[WS] Стример (камера) успешно подключился!")

	for {
		messageType, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("[WS] Стример отключился:", err)
			break
		}

		if messageType == websocket.BinaryMessage {
			hub.broadcast <- message
		}
	}
}

func handleViewer(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("[WS] Ошибка апгрейда зрителя:", err)
		return
	}

	hub.register <- conn

	go func() {
		defer func() {
			hub.unregister <- conn
		}()
		for {
			if _, _, err := conn.ReadMessage(); err != nil {
				break
			}
		}
	}()
}

func InitStreamRoutes() {
	go hub.run()

	http.HandleFunc("/ws/stream", handleStreamer)
	http.HandleFunc("/ws/view", handleViewer) 
	
	fmt.Println("[WS] Эндпоинты стриминга /ws/stream и /ws/view успешно инициализированы")
}
