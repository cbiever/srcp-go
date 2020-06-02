package bus

import (
	. "srcpd-go/command"
	"srcpd-go/configuration"
)

type DDLBus struct {
	number   int
	numberGA int
	numberGL int
}

func NewDDLBus(number int, ddl *configuration.DDL) *DDLBus {
	ddlBus := DDLBus{}
	ddlBus.number = number
	ddlBus.numberGA = ddl.NumberGA
	ddlBus.numberGL = ddl.NumberGL
	return &ddlBus
}

func (ddlBus *DDLBus) HandleCommand(command Command) Reply {
	return Reply{0, "reply ddl bus"}
}
