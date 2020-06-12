package command

import . "srcpd-go/model"

type Command interface {
	Bus() int
}

type InitGLCommand struct {
	GL GL
}

func (initGLCommand InitGLCommand) Bus() int {
	return initGLCommand.GL.Bus
}

type GetGLCommand struct {
	GL GL
}

func (getGLCommand GetGLCommand) Bus() int {
	return getGLCommand.GL.Bus
}

type SetGLCommand struct {
	GL GL
}

func (setGLCommand SetGLCommand) Bus() int {
	return setGLCommand.GL.Bus
}

type TermGLCommand struct {
	GL GL
}

func (termGLCommand TermGLCommand) Bus() int {
	return termGLCommand.GL.Bus
}

type UnrecognizedCommand struct {
}

func (unrecognizedCommand UnrecognizedCommand) Bus() int {
	return -1
}

type RSVP struct {
	Command      Command
	ReplyChannel chan Reply
}

type Reply struct {
	InfoType  InfoType
	Device    interface{}
	Message   string
	ErrorCode int
}

type InfoType int

const (
	Get = iota + 100
	Init
	Term
	Error
	OK = 200
)

type Info struct {
	InfoType InfoType
	Device   interface{}
}

type SubscribeInfo struct {
	SessionID   uint32
	InfoChannel chan Info
}

type UnsubscribeInfo struct {
	SessionID uint32
}
