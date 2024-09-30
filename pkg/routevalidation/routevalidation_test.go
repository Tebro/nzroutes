package routevalidation

import (
	"testing"

	"github.com/Tebro/efinmet/pkg/vatsimdata"
	"github.com/Tebro/nzroutes/pkg/nzroutes"
)

func TestValidatePilotRoute(t *testing.T) {
	tests := []struct {
		name     string
		plan     vatsimdata.DataFlightPlan
		route    *nzroutes.AirfieldRoute
		expected bool
	}{
		{
			name: "Empty filed route",
			plan: vatsimdata.DataFlightPlan{Route: ""},
			route: &nzroutes.AirfieldRoute{
				RunwayDependant:    false,
				RoutePointsStrings: []string{"POINTA", "POINTB"},
			},
			expected: false,
		},
		{
			name: "Non-matching route points",
			plan: vatsimdata.DataFlightPlan{Route: "POINTX POINTY"},
			route: &nzroutes.AirfieldRoute{
				RunwayDependant:    false,
				RoutePointsStrings: []string{"POINTA", "POINTB"},
			},
			expected: false,
		},
		{
			name: "Partially matching route points",
			plan: vatsimdata.DataFlightPlan{Route: "POINTX POINTY"},
			route: &nzroutes.AirfieldRoute{
				RunwayDependant:    false,
				RoutePointsStrings: []string{"POINTX", "POINTB"},
			},
			expected: false,
		},
		{
			name: "Matching route points",
			plan: vatsimdata.DataFlightPlan{Route: "POINTA POINTB"},
			route: &nzroutes.AirfieldRoute{
				RunwayDependant:    false,
				RoutePointsStrings: []string{"POINTA", "POINTB"},
			},
			expected: true,
		},
		{
			name: "Runway dependant matching route points",
			plan: vatsimdata.DataFlightPlan{Route: "POINTA POINTB"},
			route: &nzroutes.AirfieldRoute{
				RunwayDependant: true,
				RoutePointsDetails: []nzroutes.RunwayDependantRoutePoints{
					{
						Points: []string{"POINTA", "POINTB"},
					},
				},
			},
			expected: true,
		},
		{
			name: "Runway dependant non-matching route points",
			plan: vatsimdata.DataFlightPlan{Route: "POINTX POINTY"},
			route: &nzroutes.AirfieldRoute{
				RunwayDependant: true,
				RoutePointsDetails: []nzroutes.RunwayDependantRoutePoints{
					{
						Points: []string{"POINTA", "POINTB"},
					},
				},
			},
			expected: false,
		},
		{
			name: "Runway dependant partial matching route points",
			plan: vatsimdata.DataFlightPlan{Route: "POINTX POINTY"},
			route: &nzroutes.AirfieldRoute{
				RunwayDependant: true,
				RoutePointsDetails: []nzroutes.RunwayDependantRoutePoints{
					{
						Points: []string{"POINTA", "POINTB"},
					},
					{
						Points: []string{"POINTX", "POINTY"},
					},
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ValidatePilotRoute(tt.plan, tt.route)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v", tt.expected, result)
			}
		})
	}
}
