package bus

import (
	"github.com/pkg/term"
)

type Device interface {
	Init()
}

type DdlDevice struct {
	address          string
	verbosity        int
	serialConnection *term.Term
}

func newDDLDevice(address string, verbosity int) DdlDevice {
	ddlDevice := DdlDevice{}
	ddlDevice.address = address
	ddlDevice.verbosity = verbosity
	return ddlDevice
}

func (ddlDevice DdlDevice) Init() {
	/*
		var err error

		ddlDevice.serialConnection, err = term.Open(ddlDevice.address, term.Speed(19200), term.RawMode)
		if err != nil {
			log.Fatalf("Error opening serial device %s (%v)", ddlDevice.address, err)
		}

	*/
}
