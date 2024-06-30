package main

import (
	"log"
	"net"
)

const PORT = "6969"

type (
	MessageType int
	Message     struct {
		Type    MessageType
		Conn    net.Conn
		Message string
	}
)

const (
	ClientConnected MessageType = iota + 1
	ClientDisconnected
	NewMessage
)

type Client struct {
	Conn     net.Conn
	Outgoing chan string
}

func server(messages chan Message) {
	clients := map[string]*Client{}

	for {
		msg := <-messages

		switch msg.Type {
		case ClientConnected:
			addr := msg.Conn.RemoteAddr().String()

			clients[addr] = &Client{
				Conn: msg.Conn,
			}

			log.Printf("INFO: Client %v connected", addr)
		case ClientDisconnected:
			addr := msg.Conn.RemoteAddr().String()

			delete(clients, addr)

			log.Printf("INFO: Client %v disconnected", addr)
		case NewMessage:
			for _, client := range clients {
				if client.Conn.RemoteAddr().String() != msg.Conn.RemoteAddr().String() {
					client.Conn.Write([]byte(msg.Message))
				}
			}

		}
	}
}

func client(conn net.Conn, messages chan Message) {
	buffer := make([]byte, 64)

	for {
		n, err := conn.Read(buffer)
		if err != nil {
			conn.Close()
			messages <- Message{
				Type: ClientDisconnected,
				Conn: conn,
			}
			return
		}

		text := string(buffer[0:n])
		messages <- Message{
			Type:    NewMessage,
			Conn:    conn,
			Message: text,
		}
	}
}

func main() {
	listener, err := net.Listen("tcp", ":"+PORT)
	if err != nil {
		log.Fatalf("error: could not listen to port: %v\n", PORT)
	}

	log.Printf("INFO: Listening to port: %v\n", PORT)

	messages := make(chan Message)
	go server(messages)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Could not accept a connection: %v\n", err.Error())
			continue
		}

		log.Printf("INFO: Accepted conection from: %v\n", conn.RemoteAddr().String())

		messages <- Message{
			Type: ClientConnected,
			Conn: conn,
		}

		go client(conn, messages)
	}
}
