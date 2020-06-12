package bus

import (
	. "srcpd-go/command"
	"srcpd-go/configuration"
)

type Bus struct {
	number      int
	device      Device
	autoPowerOn bool
	numberGA    int
	numberGL    int
	verbosity   int
}

func ConfigureBusses(configuration configuration.Configuration) []Bus {
	busses := make([]Bus, len(configuration.Bus))
	for i, bus := range configuration.Bus[1:] {
		if bus.DDL != nil {
			busses[i] = newBus(i+1, bus.AutoPowerOn, bus.Device, bus.DDL, bus.Verbosity)
		}
		busses[i].Init()
	}
	return busses
}

func newBus(number int, autoPowerOn configuration.YesOrNo, address string, d interface{}, verbosity int) Bus {
	bus := Bus{}
	bus.number = number
	bus.autoPowerOn = bool(autoPowerOn)
	switch device := d.(type) {
	case *configuration.DDL:
		bus.device = newDDLDevice(address, verbosity)
		bus.numberGA = device.NumberGA
		bus.numberGL = device.NumberGL
	}
	bus.verbosity = verbosity
	return bus
}

func (bus Bus) Init() {
	bus.device.Init()
}

func (bus Bus) HandleCommand(command Command) Reply {
	switch c := command.(type) {
	case InitGLCommand:
		return Reply{Init, c.GL, "", 0}
	case GetGLCommand:
		return Reply{Get, c.GL, "", 0}
	case TermGLCommand:
		return Reply{Term, c.GL, "", 0}
	default:
		return Reply{Error, nil, "ERROR unknown command", 410}
	}
}
