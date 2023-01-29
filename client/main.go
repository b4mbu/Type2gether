package main

func main() {
    client, err := NewClient("localhost:8080", "user_client", "youshallnotpass")
    if err != nil {
        log.Error(err.Error())
        return
    }
    client.Start()
}
