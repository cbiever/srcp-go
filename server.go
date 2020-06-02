package main

import (
	"fmt"
	"log"
	"net"
	. "srcpd-go/bus"
	. "srcpd-go/command"
	"srcpd-go/configuration"
	. "srcpd-go/connector"
	"sync"
	"time"
)

var subscriptions sync.Map
var busses []Bus

func runServer(srcpConfiguration configuration.Configuration) {
	var server *configuration.Server
	for _, c := range srcpConfiguration.Bus {
		if c.Server != nil {
			server = c.Server
		}
	}

	var port = 4303
	if server != nil {
		port = server.TcpPort
	}

	busses = ConfigureBusses(srcpConfiguration)

	runTcpServer(port)
}

func runTcpServer(port int) {
	serverSocket, err := net.Listen("tcp4", fmt.Sprintf(":%d", port))
	if err != nil {
		log.Fatalf("Unable to open open listening socket (%s)", err.Error())
	}
	defer serverSocket.Close()

	subscriptionChannel := make(chan interface{})

	go handleRegisterListener(subscriptionChannel)

	commandChannel := make(chan RSVP)

	go handleCommand(commandChannel)

	go broadcast()

	for {
		connection, err := serverSocket.Accept()
		if err != nil {
			log.Fatalf("Error accepting incoming socket connection (%s)", err.Error())
		}
		NewTcpConnector(connection, subscriptionChannel, commandChannel).Start()
	}
}

func handleRegisterListener(subscriptionChannel chan interface{}) {
	for {
		switch request := (<-subscriptionChannel).(type) {
		case SubscribeInfo:
			subscriptions.Store(request.SessionID, request.InfoChannel)
			log.Printf("Subscription added for session %d", request.SessionID)
		case UnsubscribeInfo:
			subscriptions.Delete(request.SessionID)
			log.Printf("Subscription removed for session %d", request.SessionID)
		}
	}
}

func broadcast() {
	duration, _ := time.ParseDuration("1s")
	for {
		time.Sleep(duration)
		subscriptions.Range(func(key, value interface{}) bool {
			subscriptor := value.(chan interface{})
			subscriptor <- GLInfo{Init, 1, 123}
			return true
		})
	}
}

func handleCommand(commandChannel chan RSVP) {
	for {
		rsvp := <-commandChannel
		bus := busses[rsvp.Command.Bus()]
		if bus != nil {
			rsvp.ReplyChannel <- bus.HandleCommand(rsvp.Command)
		} else {
			rsvp.ReplyChannel <- Reply{412, "ERROR wrong value"}
		}
	}
}
