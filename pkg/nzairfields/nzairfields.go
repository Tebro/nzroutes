package nzairfields

import (
	"encoding/json"
	"fmt"
)

type LatLong struct {
	Lat  float64 `json:"lat"`
	Long float64 `json:"long"`
}

type AirfieldData struct {
	ICAO     string  `json:"icao"`
	Name     string  `json:"name"`
	Location LatLong `json:"location"`
}

func ParseAirfields(data []byte) (map[string]AirfieldData, error) {
	var airfields []AirfieldData
	err := json.Unmarshal(data, &airfields)
	if err != nil {
		return nil, fmt.Errorf("could not parse airfields data: %w", err)
	}
	res := make(map[string]AirfieldData)
	for _, airfield := range airfields {
		res[airfield.ICAO] = airfield
	}

	return res, nil
}
