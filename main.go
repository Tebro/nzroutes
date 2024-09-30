package main

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/Tebro/efinmet/pkg/utils"
	"github.com/Tebro/efinmet/pkg/vatsimdata"
	"github.com/Tebro/nzroutes/pkg/nzroutes"
)


func main() {
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

		utils.ClearTerm()
		fmt.Println("NZRoutes, relevant routes")
		fmt.Println("Callsign | Departure | Arrival | Route ID | Route Points | Remarks")
		fmt.Println("=====================================================================")

		for _, pilot := range internalPilots {
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
					fmt.Printf("		| %s | %s | %s | %s | %s\n", pilot.FlightPlan.Departure, pilot.FlightPlan.Arrival, route.Id, route.RoutePoints(), route.Remarks)
				}
			} else {
				fmt.Printf("%s | %s | %s | %s | %s | %s\n", pilot.Callsign, pilot.FlightPlan.Departure, pilot.FlightPlan.Arrival, relevantRoutes[0].Id, relevantRoutes[0].RoutePoints(), relevantRoutes[0].Remarks)
			}
		}

		time.Sleep(30 * time.Second)
	}

}
