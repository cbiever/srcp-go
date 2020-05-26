package connector

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"net"
	"regexp"
	. "srcpd-go/command"
	"strings"
	"sync/atomic"
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
	Command_ = iota // _ due to conflict with Command type
	Info
)

type TcpConnector struct {
	connection            net.Conn
	sessionID             uint32
	subscriptionChannel   chan interface{}
	commandChannel        chan Command
	replyChannel          chan Reply
	infoChannel           chan interface{}
	protocolVersionRegexp *regexp.Regexp
	connectionModeRegexp  *regexp.Regexp
	connectionStatus      ConnectionStatus
	connectionMode        ConnectionMode
	commandTranslator     *CommandTranslator
}

func NewTcpConnector(connection net.Conn, subscriptionChannel chan interface{}, commandChannel chan Command) *TcpConnector {
	tcpConnector := new(TcpConnector)
	tcpConnector.connection = connection
	tcpConnector.subscriptionChannel = subscriptionChannel
	tcpConnector.commandChannel = commandChannel
	tcpConnector.replyChannel = make(chan Reply)
	tcpConnector.protocolVersionRegexp = regexp.MustCompile("SET PROTOCOL SRCP (\\d\\.\\d\\.\\d)")
	tcpConnector.connectionModeRegexp = regexp.MustCompile("SET CONNECTIONMODE SRCP (INFO|COMMAND|.*)")
	tcpConnector.connectionStatus = Handshake
	tcpConnector.connectionMode = Command_
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
			if err == io.EOF {
				if tcpConnector.sessionID > 0 {
					log.Printf("Socket closed by peer for session %d", tcpConnector.sessionID)
				} else {
					log.Println("Socket closed by peer")
				}
			} else {
				log.Println("An error occurred while reading socket", err)
			}
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

			connectionMode := tcpConnector.connectionModeRegexp.FindStringSubmatch(data)
			if len(connectionMode) == 2 {
				switch connectionMode[1] {
				case "COMMAND":
					tcpConnector.connectionMode = Command_
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
				tcpConnector.sessionID = atomic.AddUint32(&sessionID, 1)
				switch tcpConnector.connectionMode {
				case Command_:
					tcpConnector.connectionStatus = CommandMode
					tcpConnector.commandTranslator = NewCommandTranslator()
				case Info:
					tcpConnector.connectionStatus = InformationMode
					tcpConnector.infoChannel = make(chan interface{})
					tcpConnector.subscriptionChannel <- SubscribeInfo{tcpConnector.sessionID, tcpConnector.infoChannel}
					go func() {
						for {
							switch info := (<-tcpConnector.infoChannel).(type) {
							case GLInfo:
								switch info.InfoType {
								case Get:
									writer.WriteString(fmt.Sprintf("100 %d %d\n", info.Bus, info.Address))
								case Init:
									writer.WriteString(fmt.Sprintf("101 %d %d\n", info.Bus, info.Address))
								case Term:
									writer.WriteString(fmt.Sprintf("102 %d %d\n", info.Bus, info.Address))
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
				tcpConnector.commandChannel <- Command{command, tcpConnector.replyChannel}
				reply := <-tcpConnector.replyChannel
				tcpConnector.sendReply(writer, reply.Message)
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
