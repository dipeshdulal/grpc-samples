### Console Chat App using grpc.

This is a basic console chat application built using grpc.

When client joins the channel by running 
```
go run client/client.go -name Dipesh -channel default
```

In the server, 
- the request is received `chan` is constructed and put into the map for later use when multiple users are connected in a channel.

- On another user connecting to the channel, the list of channel is appended with another `chan` variable.

- And then the function `JoinChannel` is waiting for message to be present in `chan`

    ```go
    case msg := <-msgChannel:
        fmt.Printf("GO ROUTINE (got message): %v \n", msg)
        msgStream.Send(msg)
    ```

- When one client sends the message, `SendMessage` function receives the message and puts the message in `chan` object present in the map looping over.

    ```go
    streams := s.channel[msg.Channel.Name]
    for _, msgChan := range streams {
        msgChan <- msg
    }
    ```

In this way, server receives and routes the message using `grpc`.