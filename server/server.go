package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"

	"github.com/gorilla/websocket"

	"server/logger"
)

type Message struct {
    AuthorUsername string 
    Message        string `json:"message"`
}

type Client struct {
    /* Обращаю внимание, что это структура клиента именно внутри сервера */
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

    logger       *logger.Logger
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
        logger: logger.NewLogger(os.Stdout),
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
        
        if password == s.password && !ok {
            next.ServeHTTP(w, r)
            return
        }

        w.Header().Set("WWW-Authenticate", `Basic realm="restricted", charset="UTF-8"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
        s.logger.Error(fmt.Sprintf("User %s: access denied", username))
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

    s.logger.Success(fmt.Sprintf("User %s: connected", username))

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
            s.logger.Error(err.Error())
            return
        }

        s.logger.Info(fmt.Sprintf("User %s(id: %d) sent message: %s", client.Username, client.Id, string(p)))

        if messageType == websocket.CloseMessage {
            s.logger.Info(fmt.Sprintf("User %s(id: %d) closed connection", client.Username, client.Id))
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
    s.logger.Success("Server started")
    go s.RunWriter()
    err := http.ListenAndServe(":8080", nil)
    if err != nil {
        s.logger.Error(err.Error())
    }
}

