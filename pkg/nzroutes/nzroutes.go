package nzroutes

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
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

// This builds the parsed RoutePointsStrings and RoutePointsDetails fields from the raw field
// this is because the JSON has a different structure for runway dependant routes
func (route *AirfieldRoute) initRoutePoints() {
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
			route.initRoutePoints()
		}
	}

	return &d, nil
}
