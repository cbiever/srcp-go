package command

import (
	"fmt"
	. "srcpd-go/model"
)

type InfoTranslator struct {
}

func NewInfoTranslator() *InfoTranslator {
	infoTranslator := new(InfoTranslator)
	return infoTranslator
}

func (infoTranslator *InfoTranslator) Info2Text(info Info) string {
	switch device := (info.Device).(type) {
	case GL:
		switch info.InfoType {
		case Get:
			return fmt.Sprintf("100 INFO %d GL %d %d %d %d", device.Bus, device.Address, device.Drivemode, device.V, device.Vmax)
		case Init:
			return fmt.Sprintf("101 INFO %d GL %d %s", device.Bus, device.Address, device.Protocol)
		case Term:
			return fmt.Sprintf("102 INFO %d GL %d", device.Bus, device.Address)
		}
	}
	return ""
}
