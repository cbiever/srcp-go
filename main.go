package main

import (
	"encoding/xml"
	"flag"
	"io/ioutil"
	"log"
	"os"
	. "srcpd-go/configuration"
)

func main() {
	configurationFilepath := flag.String("f", "srcpd.conf", "-f <configuration file>")

	flag.Parse()

	configurationFile, err := os.Open(*configurationFilepath)
	if err != nil {
		log.Fatalf("Unable to open configuration file (%s)", err.Error())
	}

	configurationXml, err := ioutil.ReadAll(configurationFile)
	if err != nil {
		log.Fatalf("Unable to read configuration file (%s)", err.Error())
	}
	configurationFile.Close()

	var configuration Configuration
	xml.Unmarshal(configurationXml, &configuration)

	runServer(configuration)
}
