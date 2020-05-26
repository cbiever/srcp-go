package command

import (
	"fmt"
	"regexp"
	. "srcpd-go/model"
	"strconv"
	"strings"
)

type CommandTranslator struct {
	initGLRegexp *regexp.Regexp
	getGLRegexp  *regexp.Regexp
	setGLRegexp  *regexp.Regexp
	termGLRegexp *regexp.Regexp
}

func NewCommandTranslator() *CommandTranslator {
	commandTranslator := new(CommandTranslator)
	commandTranslator.initGLRegexp = regexp.MustCompile("INIT (\\d+) GL (\\d+) (A|F|L|M|N|P|S|Z)[ ]?([1|2] \\d+ \\d+)?")
	commandTranslator.getGLRegexp = regexp.MustCompile("GET (\\d+) GL (\\d+)")
	commandTranslator.setGLRegexp = regexp.MustCompile("SET (\\d+) GL (\\d+) (0|1|2) (\\d+) (\\d+)([ \\d+]*)?")
	commandTranslator.termGLRegexp = regexp.MustCompile("TERM (\\d+) GL (\\d+)")
	return commandTranslator
}

func (commandTranslator *CommandTranslator) Translate(data string) interface{} {
	initGL := commandTranslator.initGLRegexp.FindStringSubmatch(data)

	if len(initGL) > 3 {
		gl := GL{}
		gl.Name = fmt.Sprintf("GL-%s-%s", initGL[1], initGL[2])
		gl.Bus, _ = strconv.Atoi(initGL[1])
		gl.Address, _ = strconv.Atoi(initGL[2])
		gl.Protocol = initGL[3]
		if gl.Protocol == "M" || gl.Protocol == "N" {
			s := strings.Split(initGL[4], " ")
			if len(s) == 3 {
				gl.ProtocolVersion, _ = strconv.Atoi(s[0])
				gl.DecoderSpeedSteps, _ = strconv.Atoi(s[1])
				gl.NumberOfDecoderFunctions, _ = strconv.Atoi(s[2])
			} else {
				return UnrecognizedCommand{}
			}
		}
		return InitGLCommand{gl}
	}

	getGL := commandTranslator.getGLRegexp.FindStringSubmatch(data)

	if len(getGL) == 3 {
		gl := GL{}
		gl.Bus, _ = strconv.Atoi(getGL[1])
		gl.Address, _ = strconv.Atoi(getGL[2])
		return GetGLCommand{gl}
	}

	setGL := commandTranslator.setGLRegexp.FindStringSubmatch(data)

	if len(setGL) > 6 {
		gl := GL{}
		gl.Bus, _ = strconv.Atoi(setGL[1])
		gl.Address, _ = strconv.Atoi(setGL[2])
		gl.Drivemode, _ = strconv.Atoi(setGL[3])
		gl.V, _ = strconv.Atoi(setGL[4])
		gl.Vmax, _ = strconv.Atoi(setGL[5])
		s := strings.Split(strings.TrimSpace(setGL[6]), " ")
		if len(s) > 0 {
			gl.Function = make([]int, len(s))
			for i, f := range s {
				gl.Function[i], _ = strconv.Atoi(f)
			}
		}
		return SetGLCommand{gl}
	}

	termGL := commandTranslator.termGLRegexp.FindStringSubmatch(data)

	if len(termGL) == 3 {
		gl := GL{}
		gl.Bus, _ = strconv.Atoi(termGL[1])
		gl.Address, _ = strconv.Atoi(termGL[2])
		return TermGLCommand{gl}
	}

	return UnrecognizedCommand{}
}
