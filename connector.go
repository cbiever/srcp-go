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
	sessionID             int
	subscriptionChannel   chan interface{}
	infoChannel           chan interface{}
	protocolVersionRegexp *regexp.Regexp
	connectionModeRegexp  *regexp.Regexp
	connectionStatus      ConnectionStatus
	connectionMode        ConnectionMode
	commandTranslator     *CommandTranslator
}

var sessionID = 1

func NewTcpConnector(connection net.Conn, subscriptionChannel chan interface{}) *TcpConnector {
	tcpConnector := new(TcpConnector)
	tcpConnector.connection = connection
	tcpConnector.subscriptionChannel = subscriptionChannel
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
	defer func() {
		if tcpConnector.infoChannel != nil {
			tcpConnector.subscriptionChannel <- UnsubscribeInfo{tcpConnector.sessionID}
		}
	}()

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
					if tcpConnector.sendReply("201 OK PROTOCOL SRCP") != nil {
						return
					}
				} else {
					if tcpConnector.sendReply("400 ERROR unsupported protocol") != nil {
						return
					}
				}
				handled = true
			}

			cm := tcpConnector.connectionModeRegexp.FindStringSubmatch(data)
			if cm != nil && len(cm) == 2 {
				switch cm[1] {
				case "COMMAND":
					tcpConnector.connectionMode = Command
					if tcpConnector.sendReply("202 OK CONNECTIONMODE") != nil {
						return
					}
				case "INFO":
					tcpConnector.connectionMode = Info
					if tcpConnector.sendReply("202 OK CONNECTIONMODE") != nil {
						return
					}
				default:
					if tcpConnector.sendReply("401 ERROR unsupported connection mode") != nil {
						return
					}
				}
				handled = true
			}

			if "GO" == data {
				if tcpConnector.sendReply(fmt.Sprintf("200 OK GO %d", sessionID)) != nil {
					return
				}
				tcpConnector.sessionID = sessionID
				sessionID++
				switch tcpConnector.connectionMode {
				case Command:
					tcpConnector.connectionStatus = CommandMode
					tcpConnector.commandTranslator = NewCommandTranslator()
				case Info:
					tcpConnector.connectionStatus = InformationMode
					tcpConnector.infoChannel = make(chan interface{})
					tcpConnector.subscriptionChannel <- SubscribeInfo{tcpConnector.sessionID, tcpConnector.infoChannel}
					go func() {
						for {
							info := <-tcpConnector.infoChannel
							writer := bufio.NewWriter(tcpConnector.connection)
							writer.WriteString(fmt.Sprintf("%s\n", info))
							writer.Flush()
						}
					}()
				}
				handled = true
			}

			if !handled {
				if tcpConnector.sendReply("410 ERROR unknown command") != nil {
					return
				}
			}
		case CommandMode:
			command := tcpConnector.commandTranslator.Translate(data)
			switch command.(type) {
			default:
				if tcpConnector.sendReply("200 OK") != nil {
					return
				}
				tcpConnector.infoChannel <- command
			case UnrecognizedCommand:
				if tcpConnector.sendReply("410 ERROR unknown command") != nil {
					return
				}
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

func (tcpConnector *TcpConnector) sendReply(message string) (error error) {
	_, err := tcpConnector.connection.Write([]byte(fmt.Sprintf("%.3f %s\n", float64(time.Now().UnixNano())/1e9, message)))
	if err != nil {
		log.Println("Error writing socket", err)
	}
	return err
}
