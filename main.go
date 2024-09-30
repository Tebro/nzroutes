package main

import (
	_ "embed"
	"fmt"
	"log"
	"sort"
	"strings"
	"time"

	"github.com/Tebro/efinmet/pkg/utils"
	"github.com/Tebro/efinmet/pkg/vatsimdata"
	"github.com/Tebro/nzroutes/pkg/nzairfields"
	"github.com/Tebro/nzroutes/pkg/nzroutes"
	"github.com/Tebro/nzroutes/pkg/routevalidation"
	"github.com/Tebro/nzroutes/pkg/termtable"
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
		table := termtable.New("Callsign", "Departure", "Arrival", "Route ID", "Route Points", "Remarks", "Flight plan matches? (WIP)")

		for _, pilot := range routeRelevantPilots {
			depFieldRoutes, ok := routesData.Routes[pilot.FlightPlan.Departure]
			if !ok {
				fmt.Printf("Flight %s has a departure airfield with no routes: %s\n", pilot.Callsign, pilot.FlightPlan.Departure)
				continue
			}
			var relevantRoutes []*nzroutes.AirfieldRoute
			for _, route := range depFieldRoutes {
				if route.Destination == pilot.FlightPlan.Arrival {
					relevantRoutes = append(relevantRoutes, route)
				}
			}

			if len(relevantRoutes) > 1 {
				table.AddRow(pilot.Callsign, pilot.FlightPlan.Departure, pilot.FlightPlan.Arrival)
				for _, route := range relevantRoutes {
					validationResult := "Valid"
					if !routevalidation.ValidatePilotRoute(pilot.FlightPlan, route) {
						validationResult = "Invalid"
					}
					table.AddRow("", "", "", route.Id, route.RoutePoints(), route.Remarks.String(), validationResult)
				}
			} else {
				validationResult := "Valid"
				if !routevalidation.ValidatePilotRoute(pilot.FlightPlan, relevantRoutes[0]) {
					validationResult = "Invalid"
				}
				table.AddRow(pilot.Callsign, pilot.FlightPlan.Departure, pilot.FlightPlan.Arrival, relevantRoutes[0].Id, relevantRoutes[0].RoutePoints(), relevantRoutes[0].Remarks.String(), validationResult)
			}
		}
		table.Print()

		metarArrivalRelevantPilots := filterCloseToArrivalPilots(relevant, airfieldsData)
		metarDepartureRelevantPilots := filterCloseToDepPilots(relevant, airfieldsData)
		metarRelevantPilots := append(metarDepartureRelevantPilots, metarArrivalRelevantPilots...)
		uniqueIcaos := make(map[string]bool)
		for _, pilot := range metarRelevantPilots {
			uniqueIcaos[pilot.FlightPlan.Departure] = true
		}

		fmt.Println("\n\nCurrently relevant METARs")
		uniqueIcaosSlice := make([]string, 0, len(uniqueIcaos))
		for icao := range uniqueIcaos {
			uniqueIcaosSlice = append(uniqueIcaosSlice, icao)
		}
		sort.Strings(uniqueIcaosSlice)

		for _, icao := range uniqueIcaosSlice {
			metar, err := vatsimmetar.GetMetar(icao)
			if err != nil {
				log.Printf("Could not get METAR for %s: %v\n", icao, err)
			}
			fmt.Println(metar)
		}

		time.Sleep(30 * time.Second)
	}

}
