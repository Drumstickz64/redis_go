package main

import (
	"log"
	"net"
	"os"
)

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		log.Println("Failed to bind to port 6379")
		os.Exit(1)
	}

	log.Println("Listening on port 6379")

	conn, err := l.Accept()
	if err != nil {
		log.Println("Error accepting connection: ", err.Error())
		os.Exit(1)
	}
	defer conn.Close()

	for {
		buf := []byte{}
		_, err := conn.Read(buf)
		if err != nil {
			log.Println("Error reader message:", err)
			os.Exit(1)
		}

		log.Println("Received message:", string(buf))

		conn.Write([]byte("+PONG\r\n"))
	}

}

