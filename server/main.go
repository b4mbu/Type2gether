package main

func main() {
	server := NewServer("youshallnotpass", "go.mod")
	server.SetupRoutes()
	server.Start()
}
