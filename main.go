package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

type Message struct {
    AuthorUsername string 
    Message        string `json:"message"`
}

type Client struct {
    Id       int
    Username string
    Conn     *websocket.Conn
    R        *http.Request
}

func NewClient(id int, username string, conn *websocket.Conn, r *http.Request) *Client {
    return &Client{Id: id, Conn: conn, R: r, Username: username}
}

type Server struct {
    password     string
    upgrader     websocket.Upgrader

    clientsMutex sync.Mutex
    clients      map[string]*Client
    messages     chan *Message 
}

func NewServer(password string) *Server {
    server := &Server{
        password: password,
        upgrader: websocket.Upgrader{
            ReadBufferSize:  1024,
            WriteBufferSize: 1024,
        },
        clients: make(map[string]*Client),
        messages: make(chan *Message),
    }
    server.upgrader.CheckOrigin = func(r *http.Request) bool { 
        return !server.ConsistsClient(r.Header.Get("username"))
    }
    return server
}

func (s *Server) checkAuth(next http.HandlerFunc) http.HandlerFunc {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        var (
            username = r.Header.Get("username")
            password = r.Header.Get("password")
        )
        _, ok := s.clients[username]
        log.Println(username, password)

        if password == s.password && !ok {
            next.ServeHTTP(w, r)
            return
        }

        w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
    })
}

func (s *Server) wsEndpoint(w http.ResponseWriter, r *http.Request) {
    ws, err := s.upgrader.Upgrade(w, r, nil)
    if err != nil {
        w.Write([]byte(fmt.Sprintf("websocket connection failed: %s\n", err)))
        return
    }
    username := r.Header.Get("username")
    client := NewClient(len(s.clients), username, ws, r)
    s.clients[username] = client

    s.RunReader(client)
}

func (s *Server) ConsistsClient(username string) bool {
    s.clientsMutex.Lock()
    defer s.clientsMutex.Unlock()

    _, ok := s.clients[username]
    return ok
}

func (s *Server) RunReader(client *Client) {
    defer s.CloseConnection(client)
    for {
        messageType, p, err := client.Conn.ReadMessage()
        if err != nil {
            log.Println(err)
            return
        }

        log.Printf("clientId: %d, username: %s, message: %s\n", client.Id, client.Username, string(p))

        if messageType == websocket.CloseMessage {
            log.Printf("clientId: %d closed connection", client.Id)
            s.CloseConnection(client)
            return
        }
        
        var message Message
        json.Unmarshal(p, &message)
        message.AuthorUsername = client.Username
        s.messages <- &message
    }
}

func (s *Server) RunWriter() {
    for {
        select {
        case message := <-s.messages:
            s.writeMessageToClients(message, message.AuthorUsername)
        }
    }
}

func (s *Server) writeMessageToClients(message *Message, authorUsername string) {
    s.clientsMutex.Lock()
    defer s.clientsMutex.Unlock()

    for username, client := range s.clients {
        if username != authorUsername {
            client.Conn.WriteJSON(message)    
        }
    }
}

func (s *Server) CloseConnection(client *Client) {
    s.clientsMutex.Lock()
    defer s.clientsMutex.Unlock()

    client.Conn.Close()
    delete(s.clients, client.Username)
}

func (s *Server) SetupRoutes() {
    http.HandleFunc("/ws", s.checkAuth(s.wsEndpoint))
}

func (s *Server) Start() {
    go s.RunWriter()
    http.ListenAndServe(":8080", nil)
}

func main() {
    server := NewServer("youshallnotpass")
    server.SetupRoutes()
    server.Start()
}
