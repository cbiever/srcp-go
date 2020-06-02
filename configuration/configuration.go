package configuration

import (
	"encoding/xml"
	"strings"
)

type Configuration struct {
	XMLName xml.Name `xml:"srcpd"`
	Version string   `xml:"version,attr"`
	Bus     []Bus    `xml:"bus"`
}

type Bus struct {
	XMLName xml.Name `xml:"bus"`

	Server     *Server     `xml:"server"`
	DDL        *DDL        `xml:"ddl"`
	M605X      *M605X      `xml:"m605x"`
	Intellibox *Intellibox `xml:"intellibox"`
	Li100      *Li100      `xml:"li100"`
	Loopback   *Loopback   `xml:"loopback"`
	DdlS88     *DdlS88     `xml:"ddl-s88"`
	Hsi88      *Hsi88      `xml:"hsi-88"`
	I2cDev     *I2cDev     `xml:"i2c-dev"`
	Zimo       *Zimo       `xml:"zimo"`
	Selectrix  *Selectrix  `xml:"selectrix"`
	Loconet    *Loconet    `xml:"loconet"`

	AutoPowerOn YesOrNo `xml:"auto_power_on"`
	Device      string  `xml:"device"`
	Speed       int     `xml:"speed"`
	Verbosity   int     `xml:"verbosity"`
}

type Server struct {
	TcpPort   int    `xml:"tcp-port"`
	PidFile   string `xml:"pid-file"`
	Username  string `xml:"username"`
	Groupname string `xml:"groupname"`
}

type DDL struct {
	NumberGA                    int     `xml:"number_ga"`
	NumberGL                    int     `xml:"number_gl"`
	EnableMaerklin              YesOrNo `xml:"enable_maerklin"`
	EnableNmradcc               YesOrNo `xml:"enable_nmradcc"`
	EnableRingindicatorChecking YesOrNo `xml:"enable_ringindicator_checking"`
	EnableCheckshortChecking    YesOrNo `xml:"enable_checkshort_checking"`
	InverseDsrHandling          YesOrNo `xml:"inverse_dsr_handling"`
	ShortcutFailureDelay        string  `xml:"shortcut_failure_delay"`
	NmradccTranslationRoutine   int     `xml:"nmradcc_translation_routine"`
	ImproveNmradccTiming        YesOrNo `xml:"improve_nmradcc_timing"`
	EnableUsleepPatch           YesOrNo `xml:"enable_usleep_patch"`
	UsleepUsec                  int     `xml:"usleep_usec"`
}

type M605X struct {
	NumberFb             int     `xml:"number_fb"`
	GaMinActivetime      int     `xml:"ga_min_activetime"`
	ModeM6020            YesOrNo `xml:"mode_m6020"`
	FbDelayTime0         int     `xml:"fb_delay_time_0"`
	PauseBetweenBytes    int     `xml:"pause_between_bytes"`
	PauseBetweenCommands int     `xml:"pause_between_commands"`
}

type Intellibox struct {
	NumberFb     int `xml:"number_fb"`
	NumberGa     int `xml:"number_ga"`
	NumberGl     int `xml:"number_gl"`
	FbDelayTime0 int `xml:"fb_delay_time_0"`
}

type Li100 struct {
	NumberGa     int `xml:"number_ga"`
	NumberGl     int `xml:"number_gl"`
	NumberFb     int `xml:"number_fb"`
	NumberSm     int `xml:"number_sm"`
	FbDelayTime0 int `xml:"fb_delay_time_0"`
}

type Loopback struct {
	NumberFb int `xml:"number_fb"`
	NumberGa int `xml:"number_ga"`
	NumberGl int `xml:"number_gl"`
}

type DdlS88 struct {
	Ioport       string `xml:"ioport"`
	Clockscale   int    `xml:"clockscale"`
	Refresh      int    `xml:"refresh"`
	FbDelayTime0 int    `xml:"fb_delay_time_0"`
	NumberFb1    int    `xml:"number_fb_1"`
	NumberFb2    int    `xml:"number_fb_2"`
	NumberFb3    int    `xml:"number_fb_3"`
	NumberFb4    int    `xml:"number_fb_4"`
}

type Hsi88 struct {
	NumberFbLeft   int `xml:"number_fb_left"`
	NumberFbCenter int `xml:"number_fb_center"`
	NumberFbRight  int `xml:"number_fb_right"`
	Refresh        int `xml:"refresh"`
	FbDelayTime0   int `xml:"fb_delay_time_0"`
}

type I2cDev struct {
	MultiplexBuses      int `xml:"multiplex_buses"`
	GaHardwareInverters int `xml:"ga_hardware_inverters"`
	GaResetDevices      int `xml:"ga_reset_devices"`
}

type Zimo struct {
	NumberGa int `xml:"number_ga"`
	NumberGl int `xml:"number_gl"`
	NumberFb int `xml:"number_fb"`
}

type Selectrix struct {
	NumberGa   int     `xml:"number_ga"`
	NumberGl   int     `xml:"number_gl"`
	NumberFb   int     `xml:"number_fb"`
	ModeCc2000 YesOrNo `xml:"mode_cc2000"`
}

type Loconet struct {
	LoconetId           int     `xml:"loconet-id"`
	SyncTimeFromLoconet YesOrNo `xml:"sync-time-from-loconet"`
	Ms100               YesOrNo `xml:"ms100"`
}

type YesOrNo bool

func (yesOrNo *YesOrNo) UnmarshalText(text []byte) error {
	switch strings.ToLower(string(text)) {
	case "yes":
		*yesOrNo = true
	case "no":
		*yesOrNo = false
	default:
		*yesOrNo = false
	}
	return nil
}
