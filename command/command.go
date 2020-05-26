package command

import . "srcpd-go/model"

type UnrecognizedCommand struct {
}

type InitGLCommand struct {
	GL GL
}

type GetGLCommand struct {
	GL GL
}

type SetGLCommand struct {
	GL GL
}

type TermGLCommand struct {
	GL GL
}

type Command struct {
	Command      interface{}
	ReplyChannel chan Reply
}

type Reply struct {
	Message string
}

type InfoType int

const (
	Get = iota + 100
	Init
	Term
)

type GLInfo struct {
	InfoType InfoType
	Bus      int
	Address  int
}

type SubscribeInfo struct {
	SessionID   uint32
	InfoChannel chan interface{}
}

type UnsubscribeInfo struct {
	SessionID uint32
}
