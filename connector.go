package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"regexp"
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
	commandChannel        chan interface{}
	infoChannel           chan interface{}
	protocolVersionRegexp *regexp.Regexp
	connectionModeRegexp  *regexp.Regexp
	connectionStatus      ConnectionStatus
	connectionMode        ConnectionMode
	commandTranslator     *CommandTranslator
}

var sessionID = 1

func NewTcpConnector(connection net.Conn, subscriptionChannel chan interface{}, commandChannel chan interface{}) *TcpConnector {
	tcpConnector := new(TcpConnector)
	tcpConnector.connection = connection
	tcpConnector.subscriptionChannel = subscriptionChannel
	tcpConnector.commandChannel = commandChannel
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
		tcpConnector.connection.Close()
	}()

	reader := bufio.NewReader(tcpConnector.connection)
	writer := bufio.NewWriter(tcpConnector.connection)

	for {
		data, err := reader.ReadString('\n')
		if err != nil {
			log.Println("An error occured while reading socket", err)
			return
		}
		data = strings.TrimSpace(data)

		switch tcpConnector.connectionStatus {
		case Handshake:
			handled := false

			protocolVersion := tcpConnector.protocolVersionRegexp.FindStringSubmatch(data)
			if len(protocolVersion) == 2 {
				if protocolVersion[1] == "0.8.4" {
					if tcpConnector.sendReply(writer, "201 OK PROTOCOL SRCP") != nil {
						return
					}
				} else {
					if tcpConnector.sendReply(writer, "400 ERROR unsupported protocol") != nil {
						return
					}
				}
				handled = true
			}

			commandMode := tcpConnector.connectionModeRegexp.FindStringSubmatch(data)
			if len(commandMode) == 2 {
				switch commandMode[1] {
				case "COMMAND":
					tcpConnector.connectionMode = Command
					if tcpConnector.sendReply(writer, "202 OK CONNECTIONMODE") != nil {
						return
					}
				case "INFO":
					tcpConnector.connectionMode = Info
					if tcpConnector.sendReply(writer, "202 OK CONNECTIONMODE") != nil {
						return
					}
				default:
					if tcpConnector.sendReply(writer, "401 ERROR unsupported connection mode") != nil {
						return
					}
				}
				handled = true
			}

			if "GO" == data {
				if tcpConnector.sendReply(writer, fmt.Sprintf("200 OK GO %d", sessionID)) != nil {
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
						writer := bufio.NewWriter(tcpConnector.connection)
						for {
							switch info := (<-tcpConnector.infoChannel).(type) {
							case GLInfo:
								switch info.infoType {
								case Get:
									writer.WriteString(fmt.Sprintf("100 %d %d\n", info.bus, info.address))
								case Init:
									writer.WriteString(fmt.Sprintf("101 %d %d\n", info.bus, info.address))
								case Term:
									writer.WriteString(fmt.Sprintf("102 %d %d\n", info.bus, info.address))
								}
							}
							writer.Flush()
						}
					}()
				}
				handled = true
			}

			if !handled {
				if tcpConnector.sendReply(writer, "410 ERROR unknown command") != nil {
					return
				}
			}
		case CommandMode:
			command := tcpConnector.commandTranslator.Translate(data)
			switch command.(type) {
			default:
				if tcpConnector.sendReply(writer, "200 OK") != nil {
					return
				}
				tcpConnector.commandChannel <- command
			case UnrecognizedCommand:
				if tcpConnector.sendReply(writer, "410 ERROR unknown command") != nil {
					return
				}
			}
		case InformationMode:
			// ignore - one direction, outwards, only
		}
	}
}

func (tcpConnector *TcpConnector) sendReply(writer *bufio.Writer, message string) (error error) {
	_, err := writer.WriteString(fmt.Sprintf("%.3f %s\n", float64(time.Now().UnixNano())/1e9, message))
	if err != nil {
		log.Println("Error writing socket", err)
	}
	if err == nil {
		err = writer.Flush()
		if err != nil {
			log.Println("Error flushing socket", err)
		}
	}
	return err
}
