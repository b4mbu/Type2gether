package main

func main() {
	server := NewServer("youshallnotpass", "test.txt")
	server.SetupRoutes()
	server.Start()
}
