package nexmo

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

// Numbers represents the number management API functions
type Numbers struct {
	client *Client
}

// Type NumberSearchOptions defines options for filtering when searching for available numbers to purchase
type NumberSearchOptions struct {
	Pattern       string
	SearchPattern string
}

// Type NumberSearchResponse represents a set of phone number available for purchase, and their count
type NumberSearchResponse struct {
	Count   int64
	Numbers []AvailableNumber
}

// Type AvailableNumber represents a phone number available for purchase
type AvailableNumber struct {
	Country  string
	MSISDN   string
	Type     string
	Features []string
	Cost     float64 `json:",string"`
}

/*
	GET /number/search?api_key={api_key}&api_secret={api_secret}&country={country}&pattern={pattern}&search_pattern={search_pattern}&features={features}&index={index}&size={size}
	{"count":count,"numbers":[{"country":"country-code","msisdn":"phone number","type":"type of number","features":["feature"],"cost":"number cost"}]}
*/

// Search for available phone numbers in a given country
func (c *Numbers) SearchAvailable(countryCode string) (response NumberSearchResponse, err error) {
	return c.SearchAvailableWithOptions(countryCode, NumberSearchOptions{})
}

// Search for available phone numbers in a given country, filtering by a pattern
func (c *Numbers) SearchAvailableWithOptions(countryCode string, opts NumberSearchOptions) (response NumberSearchResponse, err error) {
	if len(countryCode) <= 0 {
		err = errors.New("Invalid country code field specified")
		return
	}

	client := &http.Client{}

	requestUrl := apiRoot + "/number/search/" + c.client.apiKey + "/" + c.client.apiSecret + "/" + countryCode
	if opts.Pattern != "" && opts.SearchPattern != "" {
		requestUrl += "?pattern=" + url.QueryEscape(opts.Pattern)
		if opts.SearchPattern != "" {
			requestUrl += "&search_pattern=" + url.QueryEscape(opts.SearchPattern)
		}
	}

	r, _ := http.NewRequest("GET", requestUrl, nil)
	r.Header.Add("Accept", "application/json")

	resp, err := client.Do(r)
	defer resp.Body.Close()

	if err != nil {
		return
	}

	body, _ := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(body, &response)
	return

}

/*
	POST /number/buy/{api_key}/{api_secret}/{country}/{msisdn}
	POST /number/buy?api_key={api_key}&api_secret={api_secret}&country={country}&msisdn={msisdn}
*/

// Buy a phone number
func (c *Numbers) BuyPhoneNumber(countryCode, number string) (bool, error) {
	if len(countryCode) <= 0 {
		return false, errors.New("Invalid country code field specified")
	}

	if len(number) <= 0 {
		return false, errors.New("Invalid number field specified")
	}

	client := &http.Client{}

	requestUrl := apiRoot + "/number/buy/" + c.client.apiKey + "/" +
		c.client.apiSecret + "/" + countryCode + "/" + number
	r, _ := http.NewRequest("POST", requestUrl, nil)
	r.Header.Add("Accept", "application/json")

	resp, err := client.Do(r)
	defer resp.Body.Close()

	if err != nil {
		return false, err
	}

	switch resp.StatusCode {
	case 200:
		return true, nil
	case 401:
		return false, errors.New("Wrong credentials")
	case 420:
		return false, errors.New("Bad parameters")
	default:
		return false, errors.New("Other error")
	}
}

/*
	POST /number/cancel/{api_key}/{api_secret}/{country}/{msisdn}
	POST /number/cancel?api_key={api_key}&api_secret={api_secret}&country={country}&msisdn={msisdn}
*/

// Cancel a phone number
func (c *Numbers) CancelPhoneNumber(countryCode, number string) (bool, error) {
	if len(countryCode) <= 0 {
		return false, errors.New("Invalid country code field specified")
	}

	if len(number) <= 0 {
		return false, errors.New("Invalid number field specified")
	}

	client := &http.Client{}

	requestUrl := apiRoot + "/number/cancel/" + c.client.apiKey + "/" +
		c.client.apiSecret + "/" + countryCode + "/" + number
	r, _ := http.NewRequest("POST", requestUrl, nil)
	r.Header.Add("Accept", "application/json")

	resp, err := client.Do(r)
	defer resp.Body.Close()

	if err != nil {
		return false, err
	}

	switch resp.StatusCode {
	case 200:
		return true, nil
	case 401:
		return false, errors.New("Wrong credentials")
	case 420:
		return false, errors.New("Bad parameters")
	default:
		return false, errors.New("Other error")
	}
}
