package main

import (
	"fmt"
	"log"
	"net"
	"srcpd-go/bus"
	"srcpd-go/command"
	"srcpd-go/configuration"
	"srcpd-go/connector"
	. "srcpd-go/model"
	"sync"
)

var subscriptions sync.Map
var busses []bus.Bus

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

	busses = bus.ConfigureBusses(srcpConfiguration)

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

	commandChannel := make(chan command.RSVP)

	go handleCommand(commandChannel)

	for {
		connection, err := serverSocket.Accept()
		if err != nil {
			log.Fatalf("Error accepting incoming socket connection (%s)", err.Error())
		}
		connector.NewTcpConnector(connection, subscriptionChannel, commandChannel).Start()
	}
}

func handleRegisterListener(subscriptionChannel chan interface{}) {
	for {
		switch request := (<-subscriptionChannel).(type) {
		case command.SubscribeInfo:
			subscriptions.Store(request.SessionID, request.InfoChannel)
			log.Printf("Subscription added for session %d", request.SessionID)
		case command.UnsubscribeInfo:
			subscriptions.Delete(request.SessionID)
			log.Printf("Subscription removed for session %d", request.SessionID)
		}
	}
}

func handleCommand(commandChannel chan command.RSVP) {
	for {
		rsvp := <-commandChannel
		if 0 <= rsvp.Command.Bus() && rsvp.Command.Bus() < len(busses) {
			bus := busses[rsvp.Command.Bus()]
			reply := bus.HandleCommand(rsvp.Command)
			rsvp.ReplyChannel <- reply
			switch reply.InfoType {
			default:
				subscriptions.Range(func(key, value interface{}) bool {
					subscriptor := value.(chan interface{})
					subscriptor <- command.Info{command.Init, 1, 123, GL{}}
					return true
				})
			case command.Error:
				// Errors are not being broadcast
			}
		} else {
			rsvp.ReplyChannel <- command.Reply{command.Error, nil, "ERROR wrong value", 412}
		}
	}
}
