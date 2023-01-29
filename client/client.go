package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"client/logger"

	"github.com/gorilla/websocket"
)

var log = logger.NewLogger(os.Stdout)

type Message struct {
    AuthorUsername string
    Message        string `json:"message"`
}

type Client struct {
    Conn     *websocket.Conn
    Host     string
    Username string
    Password string
}

func NewClient(host, username, password string) (*Client, error) {
    u := url.URL{
        Scheme: "ws", 
        Host: host,
        Path: "ws",
    }
    
    header := http.Header{
        "username": {username},
        "password": {password},
    }
    
    conn, _, err := websocket.DefaultDialer.Dial(u.String(), header)

    if err != nil {
        return nil, err
    }

    return &Client{
        Conn: conn,
        Host: host,
        Username: username,
        Password: password,
    }, nil
}

func (client *Client) Start() {
    log.Success("Client started")
    go client.startReader()
    client.startWriter()

    client.Conn.Close()
}

func (client *Client) startWriter() {
    scanner := bufio.NewScanner(os.Stdin)
    for {
        scanner.Scan()
        input := scanner.Text()

        if input == ":exit" {
            cm := websocket.FormatCloseMessage(websocket.CloseNormalClosure,
                                               fmt.Sprintf("username %s: closed", client.Username))

            if err := client.Conn.WriteMessage(websocket.CloseMessage, cm); err != nil {
                log.Error(err.Error())
            }
            return
        }

        if err := client.Conn.WriteJSON(Message{Message: input}); err != nil {
            log.Error(err.Error())
            continue
        }
    }
}

func (client *Client) startReader() {
    for {
        messageType, p, err := client.Conn.ReadMessage()
        if err != nil {
            log.Error(err.Error())
            return
        }

        var message Message
        json.Unmarshal(p, &message)

        fmt.Printf("[%s] %s\n", message.AuthorUsername, message.Message)

        if messageType == websocket.CloseMessage {
            log.Info("Server closed connection")
            return
        }
    }
}

