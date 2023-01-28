package main

func main() {
    server := server.NewServer()
    server.SetupRoutes()
    server.Start()
}
