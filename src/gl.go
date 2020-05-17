package main

type GL struct {
	Name                     string  `json:"name" yaml:"name"`
	Bus                      int     `json:"bus" yaml:"-"`
	Address                  int     `json:"address" yaml:"-"`
	Protocol                 string  `json:"protocol" yaml:"protocol"`
	ProtocolVersion          int     `json:"protocol-version" yaml:"protocol-version"`
	DecoderSpeedSteps        int     `json:"decoder-speed-steps" yaml:"decoder-speed-steps"`
	NumberOfDecoderFunctions int     `json:"number-of-decoder-functions" yaml:"number-of-decoder-functions"`
	Drivemode                int     `json:"drivemode" yaml:"-"`
	V                        int     `json:"v" yaml:"-"`
	Vmax                     int     `json:"v-max" yaml:"v-max"`
	Function                 []int   `json:"functions" yaml:"-"`
	LastTimestamp            float64 `json:"-" yaml:"-"`
}
