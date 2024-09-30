package routevalidation

import (
	"strings"

	"github.com/Tebro/efinmet/pkg/vatsimdata"
	"github.com/Tebro/nzroutes/pkg/nzroutes"
)

type ValidationResult int

const (
	Valid ValidationResult = iota
	InvalidACType
	InvalidAltitude
	InvalidFlightPlanRoute
)

var validationResultNames = []string{
	"Valid",
	"InvalidACType",
	"InvalidAltitude",
	"InvalidFlightPlanRoute",
}

func (v ValidationResult) String() string {
	return validationResultNames[v]
}

type ValidatedRoute struct {
	Route            *nzroutes.AirfieldRoute
	ValidationResult ValidationResult
}

// TODO: Find a better way
var jetTypesPrefixes = []string{
	"A31",
	"B73",
	"B73",
	"B73",
	"B74",
	"B77",
	"B78",
}

func cleanFlightPlanRoute(raw string) []string {
	steps := strings.Split(raw, " ")
	for i, step := range steps {
		steps[i] = strings.Split(step, "/")[0]
	}
	return steps
}

// containsSlice checks if the larger slice contains the smaller slice
func containsSlice(larger, smaller []string) bool {
	if len(smaller) == 0 {
		return true
	}
	if len(larger) < len(smaller) {
		return false
	}

	for i := 0; i <= len(larger)-len(smaller); i++ {
		match := true
		for j := 0; j < len(smaller); j++ {
			if larger[i+j] != smaller[j] {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// Check if the pilot has filed a valid route
// Currently only checks the flight plan route
// Does not check aircraft type or altitude at this time, but may do that in the future.
// TODO: The logic needs to be validated, currently it is a close guess based on seen scenarios
func ValidatePilotRoute(plan vatsimdata.DataFlightPlan, route *nzroutes.AirfieldRoute) bool {
	filedRoute := cleanFlightPlanRoute(plan.Route)
	if len(filedRoute) == 0 {
		return false
	}
	if route.RunwayDependant {
		for _, runwayDependantRoutePoints := range route.RoutePointsDetails {
			if containsSlice(filedRoute, runwayDependantRoutePoints.Points) || containsSlice(runwayDependantRoutePoints.Points, filedRoute) {
				return true
			}
		}
		return false
	}
	return containsSlice(filedRoute, route.RoutePointsStrings) || containsSlice(route.RoutePointsStrings, filedRoute)
}
