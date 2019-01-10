package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gbrlsnchs/websocket"
)

const (
	agent   = "websocket"
	timeout = 15 * time.Second
)

var logger = log.New(os.Stderr, "", log.Lshortfile)

func main() {
	count, err := testCount()
	if err != nil {
		logger.Print(err)
		os.Exit(1)
	}

	for i := 1; i <= count; i++ {
		uri := fmt.Sprintf("ws://localhost:9001/runCase?case=%d&agent=%s", i, agent)
		ws, err := websocket.Open(uri, timeout)
		if err != nil {
			logger.Print(err)
			os.Exit(1)
		}
		for ws.Next() {
			payload, opcode := ws.Message()
			ws.SetOpcode(opcode)
			ws.Write(payload)
		}
		if err := ws.Err(); err != nil {
			logger.Print(err)
		}
		logger.Print(ws.CloseCode())
	}
	uri := fmt.Sprintf("ws://localhost:9001/updateReports?agent=%s", agent)
	websocket.Open(uri, timeout)
}

func testCount() (count int, err error) {
	ws, err := websocket.Open("ws://localhost:9001/getCaseCount", timeout)
	if err != nil {
		logger.Print("x")
		return
	}
	ws.Next()
	if err = ws.Err(); err != nil {
		logger.Print(".")
		return
	}
	payload, _ := ws.Message()
	if count, err = strconv.Atoi(string(payload)); err != nil {
		logger.Print("!")
		return
	}
	if count == 0 {
		logger.Print("no tests available")
		os.Exit(1)
	}
	return
}
