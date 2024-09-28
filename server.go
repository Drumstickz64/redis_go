package main

import (
	"errors"
	"io"
	"log"
	"log/slog"
	"net"
)

type ConnId uint

func main() {
	l, err := net.Listen("tcp", "0.0.0.0:6379")
	if err != nil {
		log.Fatalln("Failed to bind to port 6379")
	}
	defer l.Close()

	slog.Info("listening on port 6379")

	connId := ConnId(0)
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatalln("error accepting connection: ", err)
		}

		go handleConnection(conn, connId)
		connId++
	}
}

func handleConnection(conn net.Conn, id ConnId) {
	defer conn.Close()

	slog.Info("accepted connection", "id", id)

	for {
		buf := make([]byte, 1024)
		bytesRead, err := conn.Read(buf)
		if err != nil {
			if errors.Is(err, io.EOF) {
				slog.Info("connection closed", "id", id)
				return
			}

			slog.Warn("error while reading message", "id", id, "err", err)
			continue
		}

		slog.Info("received message", "id", id, "message", string(buf[:bytesRead]))

		conn.Write([]byte("+PONG\r\n"))
	}
}
