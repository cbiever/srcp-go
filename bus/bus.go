package bus

import (
	. "srcpd-go/command"
	"srcpd-go/configuration"
)

type Bus interface {
	HandleCommand(command Command) Reply
}

func ConfigureBusses(configuration configuration.Configuration) []Bus {
	busses := make([]Bus, len(configuration.Bus))
	for i, bus := range configuration.Bus {
		if bus.DDL != nil {
			busses[i] = NewDDLBus(i, bus.DDL)
		}
	}
	return busses
}
