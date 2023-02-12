package main

import (
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
	AuthorUsername string `json:"AuthorUsername"`
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
		Host:   host,
		Path:   "ws",
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
		Conn:     conn,
		Host:     host,
		Username: username,
		Password: password,
	}, nil
}

func (client *Client) Start(min chan *Message, mout chan string) {
	log.Success("Client started")
	go client.startReader(min)
	client.startWriter(mout)

	client.Conn.Close()
}

func (client *Client) startWriter(messages chan string) {
	for {
		select {
		case input := <-messages:
			if input == ":exit" {
				cm := websocket.FormatCloseMessage(websocket.CloseNormalClosure,
					fmt.Sprintf("username %s: closed", client.Username))

				if err := client.Conn.WriteMessage(websocket.CloseMessage, cm); err != nil {
					log.Error(err.Error())
				}
				return
			}

			if err := client.Conn.WriteJSON(Message{AuthorUsername: client.Username, Message: input}); err != nil {
				log.Error(err.Error())
				continue
			}
		}
	}
}

func (client *Client) startReader(min chan *Message) {
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
			// TODO: send to application
			log.Info("Server closed connection")
			return
		}
		min <- &message
	}
}
