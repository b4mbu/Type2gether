package main

func main() {
    server := NewServer("youshallnotpass")
    server.SetupRoutes()
    server.Start()
}
