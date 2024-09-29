package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Tebro/efinmet/pkg/utils"
	"github.com/Tebro/efinmet/pkg/vatsimdata"
)

var routesDataUrl = "https://raw.githubusercontent.com/vatnz-dev/std-rte-public/refs/heads/main/stdRtes.json"

type AirfieldRouteRemarks struct {
	AircraftType  string `json:"ac_type"`
	AltitudeLimit string `json:"alt_limit"`
}

func (r AirfieldRouteRemarks) String() string {
	var remarks []string
	if r.AircraftType != "" {
		remarks = append(remarks, fmt.Sprintf("Aircraft type: %s", r.AircraftType))
	}
	if r.AltitudeLimit != "" {
		remarks = append(remarks, fmt.Sprintf("Altitude limit: %s", r.AltitudeLimit))
	}

	return strings.Join(remarks, ", ")
}

type RunwayDependantRoutePoints struct {
	Runway string   `json:"runway"`
	Points []string `json:"points"`
}

type AirfieldRoute struct {
	Id                 string               `json:"id"`
	Departure          string               `json:"departureAerodrome"`
	Destination        string               `json:"destinationAerodrome"`
	RunwayDependant    bool                 `json:"runwayDependant"`
	RoutePointsRaw     any                  `json:"routePoints"`
	Remarks            AirfieldRouteRemarks `json:"remarks"`
	RoutePointsStrings []string
	RoutePointsDetails []RunwayDependantRoutePoints
}

func (r *AirfieldRoute) RoutePoints() string {
	if !r.RunwayDependant {
		return fmt.Sprintf("%s", strings.Join(r.RoutePointsStrings, " "))
	}

	return fmt.Sprintf("%s: %s", r.RoutePointsDetails[0].Runway, strings.Join(r.RoutePointsDetails[0].Points, " "))
}

type Data struct {
	Routes map[string][]*AirfieldRoute `json:"routes"`
}

func GetRoutes() (*Data, error) {
	res, err := http.Get(routesDataUrl)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	var d Data
	err = json.NewDecoder(res.Body).Decode(&d)
	if err != nil {
		return nil, err
	}
	for _, routes := range d.Routes {
		for _, route := range routes {
			if route.RunwayDependant {
				for _, point := range route.RoutePointsRaw.([]interface{}) {
					p := point.(map[string]interface{})
					runwayDependantRoutePoints := RunwayDependantRoutePoints{
						Runway: p["runway"].(string),
					}
					for _, point := range p["points"].([]interface{}) {
						runwayDependantRoutePoints.Points = append(runwayDependantRoutePoints.Points, point.(string))
					}

					route.RoutePointsDetails = append(route.RoutePointsDetails, runwayDependantRoutePoints)
				}
			} else {
				for _, point := range route.RoutePointsRaw.([]interface{}) {
					route.RoutePointsStrings = append(route.RoutePointsStrings, point.(string))
				}
			}
		}
	}

	return &d, nil
}

func main() {
	routesData, err := GetRoutes()
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
			var relevantRoutes []*AirfieldRoute
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
