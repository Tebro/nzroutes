package vatsimmetar

import (
	"fmt"
	"io"
	"net/http"
)

func GetMetar(icao string) (string, error) {
	res, err := http.Get(fmt.Sprintf("https://metar.vatsim.net/%s", icao))
	if err != nil {
		return "", fmt.Errorf("could not get METAR data: %w", err)
	}
	defer res.Body.Close()
	if res.StatusCode == 404 {
		return "", fmt.Errorf("METAR data not found for %s", icao)
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", fmt.Errorf("could not read METAR data: %w", err)
	}

	return string(body), nil
}
