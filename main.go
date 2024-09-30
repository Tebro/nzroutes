package main

import (
	_ "embed"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Tebro/efinmet/pkg/utils"
	"github.com/Tebro/efinmet/pkg/vatsimdata"
	"github.com/Tebro/nzroutes/pkg/nzairfields"
	"github.com/Tebro/nzroutes/pkg/nzroutes"
	"github.com/Tebro/nzroutes/pkg/vatsimmetar"
)

//go:embed aerodromes.json
var airfieldsJson []byte

func isCloseToLocation(pilot vatsimdata.DataPilot, location nzairfields.LatLong, distanceLimitNm float64) bool {
	distance := utils.DistanceBetween(location.Lat, location.Long, pilot.Latitude, pilot.Longitude) * 1.852
	return distance < distanceLimitNm
}

func filterCloseToDepPilots(pilots []vatsimdata.DataPilot, airfields map[string]nzairfields.AirfieldData) []vatsimdata.DataPilot {
	var relevantPilots []vatsimdata.DataPilot
	for _, pilot := range pilots {
		depField, ok := airfields[pilot.FlightPlan.Departure]
		if !ok {
			continue
		}

		if isCloseToLocation(pilot, depField.Location, 10) {
			relevantPilots = append(relevantPilots, pilot)
		}
	}
	return relevantPilots
}

func filterCloseToArrivalPilots(pilots []vatsimdata.DataPilot, airfields map[string]nzairfields.AirfieldData) []vatsimdata.DataPilot {
	var relevantPilots []vatsimdata.DataPilot
	for _, pilot := range pilots {
		arrField, ok := airfields[pilot.FlightPlan.Arrival]
		if !ok {
			continue
		}
		if isCloseToLocation(pilot, arrField.Location, 200) {
			relevantPilots = append(relevantPilots, pilot)
		}
	}
	return relevantPilots
}

func main() {
	airfieldsData, err := nzairfields.ParseAirfields(airfieldsJson)
	if err != nil {
		log.Fatal(err)
	}

	routesData, err := nzroutes.GetRoutes()
	if err != nil {
		log.Fatal(err)
	}

	for {

		pilotsData, err := vatsimdata.GetData()
		if err != nil {
			log.Println("Could not get pilots data", err)
		}

		relevant := vatsimdata.PilotsForIcaoPrefix("NZ", pilotsData.Pilots)

		var internalPilots []vatsimdata.DataPilot
		for _, pilot := range relevant {
			if strings.HasPrefix(pilot.FlightPlan.Departure, "NZ") && strings.HasPrefix(pilot.FlightPlan.Arrival, "NZ") {
				internalPilots = append(internalPilots, pilot)
			}
		}

		routeRelevantPilots := filterCloseToDepPilots(internalPilots, airfieldsData)

		utils.ClearTerm()
		fmt.Println("NZRoutes, relevant routes")
		fmt.Println("Callsign | Departure -> Arrival | Route ID | Route Points | Remarks")
		fmt.Println("=====================================================================")

		for _, pilot := range routeRelevantPilots {
			depFieldRoutes, ok := routesData.Routes[pilot.FlightPlan.Departure]
			if !ok {
				log.Printf("No routes for departure field %s\n", pilot.FlightPlan.Departure)
				continue
			}
			var relevantRoutes []*nzroutes.AirfieldRoute
			for _, route := range depFieldRoutes {
				if route.Destination == pilot.FlightPlan.Arrival {
					relevantRoutes = append(relevantRoutes, route)
				}
			}

			if len(relevantRoutes) > 1 {
				fmt.Printf("%s", pilot.Callsign)
				for _, route := range relevantRoutes {
					fmt.Printf("		| %s -> %s | %s | %s | %s\n", pilot.FlightPlan.Departure, pilot.FlightPlan.Arrival, route.Id, route.RoutePoints(), route.Remarks)
				}
			} else {
				fmt.Printf("%s | %s -> %s | %s | %s | %s\n", pilot.Callsign, pilot.FlightPlan.Departure, pilot.FlightPlan.Arrival, relevantRoutes[0].Id, relevantRoutes[0].RoutePoints(), relevantRoutes[0].Remarks)
			}
		}

		arrivalRelevantPilots := filterCloseToArrivalPilots(internalPilots, airfieldsData)
		metarRelevantPilots := append(routeRelevantPilots, arrivalRelevantPilots...)
		uniqueIcaos := make(map[string]bool)
		for _, pilot := range metarRelevantPilots {
			uniqueIcaos[pilot.FlightPlan.Departure] = true
		}

		fmt.Println("\n\nCurrently relevant METARs")
		// TODO: Sort by ICAO
		for icao := range uniqueIcaos {
			metar, err := vatsimmetar.GetMetar(icao)
			if err != nil {
				log.Printf("Could not get METAR for %s: %v\n", icao, err)
			}
			fmt.Println(metar)
		}

		time.Sleep(30 * time.Second)
	}

}
