package unlock

import (
	"encoding/json"
	"io"
	"net/http"
)

type IPSBResponse struct {
	Country string `json:"country"`
	City    string `json:"ip"`
}

func GetLocation(client *http.Client) (string, error) {
	req, err := http.NewRequest("GET", "https://api.ip.sb/geoip", nil)
	if err != nil {
		return "", err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")

	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var result IPSBResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return "", err
	}

	if result.Country != "" {
		return result.Country, nil
	}
	return result.City, nil
}
