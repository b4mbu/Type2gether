package server

import "testing"

func TestServer(t *testing.T) {
    server := NewServer("youshallnotpass")
    server.SetupRoutes()
    server.Start()
}
