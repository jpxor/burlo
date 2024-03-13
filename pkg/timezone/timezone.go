package timezone

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type geoTimezoneResponse struct {
	Location             string `json:"location"`
	IANATimezone         string `json:"iana_timezone"`
	TimezoneAbbreviation string `json:"timezone_abbreviation"`
	DSTAbbreviation      string `json:"dst_abbreviation"`
	ErrorMessage         string `json:"error_message"`
}

func FromCoordinates(latitude, longitude string) (string, error) {
	url := fmt.Sprintf("https://api.geotimezone.com/public/timezone?latitude=%s&longitude=%s", latitude, longitude)

	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var geoResponse geoTimezoneResponse
	err = json.Unmarshal(body, &geoResponse)
	if err != nil {
		return "", err
	}

	if geoResponse.ErrorMessage != "" {
		return "", fmt.Errorf("error response: %s", geoResponse.ErrorMessage)
	}

	return geoResponse.IANATimezone, nil
}
