package main

import "server/server"

func main() {
    server := server.NewServer("youshallnotpass")
    server.SetupRoutes()
    server.Start()
}
