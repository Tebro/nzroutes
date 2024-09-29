package main

import (
	"fmt"
	"log"

	"github.com/Tebro/efinmet/pkg/vatsimdata"
)


var routesDataUrl = "https://raw.githubusercontent.com/vatnz-dev/std-rte-public/refs/heads/main/stdRtes.json"



func main() {
	data, err := vatsimdata.GetData()
	if err != nil {
		log.Fatal(err)
	}

	relevant := vatsimdata.PilotsForIcaoPrefix("NZ", data.Pilots)

	var callsigns []string
	for _, pilot := range relevant {
		callsigns = append(callsigns, pilot.Callsign)
	}

	fmt.Printf("Pilots for NZ: %v\n", callsigns)

}	

