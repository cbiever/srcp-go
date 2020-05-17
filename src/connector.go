package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type ConnectionStatus int

const (
	Handshake = iota
	CommandMode
	InformationMode
	ServiceMode
)

type ConnectionMode int

const (
	Command = iota
	Info
)

type TcpConnector struct {
	connection            net.Conn
	protocolVersionRegexp *regexp.Regexp
	connectionModeRegexp  *regexp.Regexp
	connectionStatus      ConnectionStatus
	connectionMode        ConnectionMode
	commandTranslator     *CommandTranslator
}

var sessionID = 1

func NewTcpConnector(connection net.Conn) *TcpConnector {
	tcpConnector := new(TcpConnector)
	tcpConnector.connection = connection
	tcpConnector.protocolVersionRegexp = regexp.MustCompile("SET PROTOCOL SRCP (\\d\\.\\d\\.\\d)")
	tcpConnector.connectionModeRegexp = regexp.MustCompile("SET CONNECTIONMODE SRCP (INFO|COMMAND|.*)")
	tcpConnector.connectionStatus = Handshake
	tcpConnector.connectionMode = Command
	return tcpConnector
}

func (tcpConnector *TcpConnector) Start() {
	go tcpConnector.handleConnection()
}

func (tcpConnector *TcpConnector) handleConnection() {
	for {
		data, err := bufio.NewReader(tcpConnector.connection).ReadString('\n')
		if err != nil {
			log.Println("An error occured while reading socket", err)
			return
		}
		data = strings.TrimSpace(data)

		switch tcpConnector.connectionStatus {
		case Handshake:
			handled := false

			pv := tcpConnector.protocolVersionRegexp.FindStringSubmatch(data)
			if pv != nil && len(pv) == 2 {
				if pv[1] == "0.8.4" {
					tcpConnector.sendReply("201 OK PROTOCOL SRCP")
				} else {
					tcpConnector.sendReply("400 ERROR unsupported protocol")
				}
				handled = true
			}

			cm := tcpConnector.connectionModeRegexp.FindStringSubmatch(data)
			if cm != nil && len(cm) == 2 {
				switch cm[1] {
				case "COMMAND":
					tcpConnector.connectionMode = Command
					tcpConnector.sendReply("202 OK CONNECTIONMODE")
				case "INFO":
					tcpConnector.connectionMode = Info
					tcpConnector.sendReply("202 OK CONNECTIONMODE")
				default:
					tcpConnector.sendReply("401 ERROR unsupported connection mode")
				}
				handled = true
			}

			if "GO" == data {
				tcpConnector.sendReply(fmt.Sprintf("200 OK GO %d", sessionID))
				switch tcpConnector.connectionMode {
				case Command:
					tcpConnector.connectionStatus = CommandMode
					tcpConnector.commandTranslator = NewCommandTranslator()
				case Info:
					tcpConnector.connectionStatus = InformationMode
				}
				sessionID++
				handled = true
			}

			if !handled {
				tcpConnector.sendReply("410 ERROR unknown command")
			}
		case CommandMode:
			command := tcpConnector.commandTranslator.Translate(data)
			switch command.(type) {
			default:
				tcpConnector.sendReply("200 OK")
			case UnrecognizedCommand:
				tcpConnector.sendReply("410 ERROR unknown command")
			}
		case InformationMode:
			// ignore - one direction, outwards, only
		}
	}

	tcpConnector.connection.Close()
}

func (tcpConnector *TcpConnector) processCommand(commandChannel chan string) {
	for {
		select {
		case line := <-commandChannel:
			log.Println(line)
			commandChannel <- strconv.Itoa(rand.Int()) + "\n"
		}
	}
}

func (tcpConnector *TcpConnector) sendReply(message string) {
	tcpConnector.connection.Write([]byte(fmt.Sprintf("%.3f %s\n", float64(time.Now().UnixNano())/1e9, message)))
}
