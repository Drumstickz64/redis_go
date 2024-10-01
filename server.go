package main

import (
	"errors"
	"fmt"
	"io"
	"log"
	"log/slog"
	"net"
	"os"
	"strings"

	"github.com/Drumstickz64/redis-go/data"
	"github.com/Drumstickz64/redis-go/resp"
)

type ConnId uint

const LogLevelUserError slog.Level = slog.LevelDebug + 1

func main() {
	setLogLevelFromEnv()

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

			slog.Error("error while reading message", "id", id, "err", err)
			continue
		}

		msg := string(buf[:bytesRead])

		slog.Info("received message", "id", id, "message", msg)

		decoded, err := resp.Decode(msg)
		if err != nil {
			slog.Error("error while decoding message", "id", id, "err", err)
			continue
		}

		stringArr, ok := decoded.(data.Array)
		if !ok {
			slog.Error("error while interpreting message: expected message to be an array", "id", id, "message", decoded)
			continue
		}

		command, ok := stringArr[0].(data.String)
		if !ok {
			slog.Error("error while interpreting message: expected command name to be a string", "id", id, "message", decoded, "command_name", stringArr[0])
			continue
		}

		switch strings.ToUpper(string(command)) {
		case "PING":
			slog.Debug("executing command", "id", id, "command", "PING")
			conn.Write([]byte(resp.EncodeSimpleString("PONG")))
		case "ECHO":
			if err := commandEcho(conn, stringArr[1:], id); err != nil {
				slog.Error("error while executing command", "id", id, "command", "ECHO", "err", err)
			}
		case "COMMAND":
		default:
			slog.Error("recieved unkown command", "id", id, "command", command)
		}
	}
}

func commandEcho(conn net.Conn, args []data.Data, id ConnId) error {
	slog.Debug("executing command", "id", id, "command", "ECHO")

	if len(args) != 1 {
		return fmt.Errorf("wrong number of arguments for 'echo' command, expected 1, but got %d", len(args))
	}

	command, ok := args[0].(data.String)
	if !ok {
		return fmt.Errorf("expected message to be a string, but was %v", args[0])
	}

	conn.Write([]byte(resp.EncodeBulkString(command)))

	return nil
}

func setLogLevelFromEnv() {
	logLevel := os.Getenv("LOG_LEVEL")
	switch logLevel {
	case "DEBUG":
		slog.SetLogLoggerLevel(slog.LevelDebug)
	case "INFO":
		slog.SetLogLoggerLevel(slog.LevelInfo)
	case "WARNING", "WARN":
		slog.SetLogLoggerLevel(slog.LevelWarn)
	case "ERROR":
		slog.SetLogLoggerLevel(slog.LevelError)
	}
}

func reportUserError(msg string, conn net.Conn, args ...any) {
	// TODO
}
