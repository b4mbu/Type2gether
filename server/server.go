package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"

	"github.com/fatih/color"
	"github.com/gorilla/websocket"

	"server/filemanager"
	"server/logger"
	"server/textmanager"
)

const logo string = `     __    __ __            __          
    / /_  / // / ____ ___  / /_  __  __ 
   / __ \/ // /_/ __ ` + "`" + `__ \/ __ \/ / / / 
  / /_/ /__  __/ / / / / / /_/ / /_/ /  
 /_.___/  /_/ /_/ /_/ /_/_.___/\__,_/   `

type Message struct {
	AuthorUsername string
	Id             int64
	Message        string `json:"message"`
}

type Client struct {
	/* Обращаю внимание, что это структура клиента именно внутри сервера */
	Id       int64
	Username string
	Conn     *websocket.Conn
	R        *http.Request
}

func NewClient(id int64, username string, conn *websocket.Conn, r *http.Request) *Client {
	return &Client{Id: id, Conn: conn, R: r, Username: username}
}

type Server struct {
	password string
	upgrader websocket.Upgrader

	clientsMutex sync.Mutex
	clients      map[string]*Client
	messages     chan *Message

	textMutex sync.Mutex
	text      *textmanager.Text
	filename  string

	logger *logger.Logger
}

func NewServer(password, filename string) *Server {
	server := &Server{
		password: password,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
		},
		clients:  make(map[string]*Client),
		messages: make(chan *Message),
		logger:   logger.NewLogger(os.Stdout),
		text:     textmanager.NewText(),
		filename: filename,
	}
	if err := server.LoadTextFromFile(filename); err != nil {
		server.logger.Error(err.Error())
	}
	server.upgrader.CheckOrigin = func(r *http.Request) bool {
		return !server.ConsistsClient(r.Header.Get("username"))
	}
	return server
}

func (s *Server) LoadTextFromFile(filename string) error {
	str, err := filemanager.ReadFromFile(filename)

	if err != nil {
		return err
	}

	var cursorId int64 = 0
	s.text.AddNewCursor(cursorId)

	defer s.text.RemoveCursor(cursorId)

	if err = s.text.Paste(str, cursorId); err != nil {
		return err
	}

	s.text.SetCursorStartPosition(cursorId)
	return nil
}

func (s *Server) SaveTextToFile(filename string) error {
	return filemanager.SaveToFile(filename, s.text.GetFullText())
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

	s.clientsMutex.Lock()
	username := r.Header.Get("username")
	client := NewClient(int64(len(s.clients)), username, ws, r)
	s.clients[username] = client
	s.clientsMutex.Unlock()

	s.textMutex.Lock()
	s.text.AddNewCursor(client.Id)

	//// DEBUG
	// s.text.InsertCharAfter(client.Id, '#')
	////
	s.textMutex.Unlock()

	s.logger.Success(fmt.Sprintf("User %s: connected", username))

	go s.RunReader(client)
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
			return
		}

		var message Message
		json.Unmarshal(p, &message)
		message.AuthorUsername = client.Username
		message.Id = client.Id
		go s.HandleMessage(&message)
		s.messages <- &message
	}
}

func (s *Server) HandleMessage(message *Message) {
	s.text.InsertCharAfter(message.Id, rune(message.Message[0]))
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

	if len(s.clients) == 0 {
		s.SaveTextToFile(s.filename)
	}
}

func (s *Server) SetupRoutes() {
	http.HandleFunc("/ws", s.checkAuth(s.wsEndpoint))
}

func (s *Server) Start() {
	s.printLogo()
	s.logger.Success("Server started")
	go s.RunWriter()
	go func() {
		err := http.ListenAndServe(":8080", nil)
		if err != nil {
			s.logger.Error(err.Error())
		}
	}()
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-sc

	s.SaveTextToFile(s.filename)
}

func (s *Server) printLogo() {
	green := color.New(color.FgGreen).FprintlnFunc()
	green(os.Stdin, logo)
}
