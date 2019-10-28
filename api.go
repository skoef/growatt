package growatt

import (
	"bytes"
	"crypto/md5"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"
)

const (
	apiBasePath        = "https://server.growatt.com"
	apiLoginPath       = "/LoginAPI.do"
	apiCookieSessionID = "JSESSIONID"
	apiCookieServerID  = "SERVERID"
)

var (
	// ErrAPILogin is the error representing API login failures
	ErrAPILogin = errors.New("API Error: login failed")
)

// API represents the API client structure
type API struct {
	username  string
	password  string
	sessionID string
	serverID  string
	UserID    int
	UserLevel int
}

// NewAPI returns a new API struct configured with authentication details
func NewAPI(username, password string) *API {
	return &API{
		username: username,
		password: password,
	}
}

// GetPlantEnergy returns detailed energy information for plant with given PlantID
// for given period of time
// Based on the format of the given date, the timespan will be either:
// - total (empty string, "")
// - year (eg. "2019")
// - month (eg. "2019-01")
// - day (eg. "2019-01-01")
// any other format will result in error
func (a API) GetPlantEnergy(plantID int, date string) ([]TimeEnergy, error) {
	var timespan int
	var dateLayout string
	dateFormat := "%s"
	switch {
	case regexp.MustCompile(`^\d+$`).MatchString(date): // year
		timespan = 3
		dateFormat = fmt.Sprintf("%s-%%s", date)
		dateLayout = "2006-01"
	case regexp.MustCompile(`^\d+-\d+$`).MatchString(date): // month
		timespan = 2
		dateFormat = fmt.Sprintf("%s-%%s", date)
		dateLayout = "2006-01-02"
	case regexp.MustCompile(`^\d+-\d+-\d+$`).MatchString(date): // day
		timespan = 1
		dateLayout = "2006-01-02 15:04"
	case date == "":
		timespan = 4
		dateLayout = "2006"
	default:
		return nil, fmt.Errorf("could not parse timespan %s", date)
	}

	val := url.Values{}
	val.Add("plantId", fmt.Sprintf("%d", plantID))
	val.Add("type", fmt.Sprintf("%d", timespan))
	val.Add("date", date)

	data, err := a.call("GET", "newPlantDetailAPI.do", val.Encode())
	if err != nil {
		return nil, err
	}

	var r struct {
		X map[string]interface{} `json:"-"` // we don't know all the keys upfront
	}

	if err := json.Unmarshal(data, &r.X); err != nil {
		return nil, err
	}

	if _, ok := r.X["back"]; !ok {
		return nil, errors.New("could not unmarshal energy data")
	}

	energyData, ok := r.X["back"].(map[string]interface{})
	if !ok {
		return nil, errors.New("could not convert energy data")
	}

	list, ok := energyData["data"].(map[string]interface{})
	if !ok {
		return nil, errors.New("could not convert energy data")
	}

	var result []TimeEnergy
	for k, v := range list {
		normalizedDate := fmt.Sprintf(dateFormat, k)
		ts, err := time.Parse(dateLayout, normalizedDate)
		if err != nil {
			continue
		}

		pwr, err := strconv.ParseFloat(v.(string), 64)
		if err != nil {
			continue
		}

		result = append(result, TimeEnergy{
			Timestamp: ts,
			Power:     pwr,
		})
	}

	return result, nil
}

// GetPlantList get's a list of plants from the API
func (a API) GetPlantList() ([]Plant, error) {
	data, err := a.call("GET", "/PlantListAPI.do", "")
	if err != nil {
		return nil, err
	}

	var r struct {
		Back struct {
			Data []plantData `json:"data"`
			/*
				There is more JSON data being returned, but we're currently not interested
				in that data in this context
					TotalData struct {
						CurrentPowerSum string `json:"currentPowerSum"`
						CO2Sum          string `json:"CO2Sum"`
						IsHaveStorage   string `json:"isHaveStorage"`
						ETotalMoneyText string `json:"eTotalMoneyText"`
						TodayEnergySum  string `json:"todayEnergySum"`
						TotalEnergySum  string `json:"totalEnergySum"`
					} `json:"totalData"`
				Success bool `json:"success"`
			*/
		} `json:"back"`
	}

	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}

	var plants []Plant

	for _, p := range r.Back.Data {
		plants = append(plants, parsePlantData(p))
	}

	return plants, nil
}

// GetPlantInverterList returns a list of Inverter objects for given plant
func (a API) GetPlantInverterList(plantID int) ([]Inverter, error) {
	val := url.Values{}
	val.Add("op", "getAllDeviceListThree")
	val.Add("plantId", fmt.Sprintf("%d", plantID))
	val.Add("pageNum", "1")
	val.Add("pageSize", "1")

	data, err := a.call("GET", "/newPlantAPI.do", val.Encode())
	if err != nil {
		return nil, err
	}

	var r struct {
		Devices []inverterData `json:"deviceList"`
	}

	if err := json.Unmarshal(data, &r); err != nil {
		return nil, err
	}

	var result []Inverter
	for _, d := range r.Devices {
		result = append(result, parseInvertedData(d))
	}

	return result, nil
}

// GetInverterEnergy returns detailed energy information for an inverter with
// given serial number for the given date. While GetPlantEnergy does support to
// return result for multiple timespans (day, month, year) GetInverterEnergy only
// supports results for a timespan of a day
func (a API) GetInverterEnergy(inverterSerial string, date time.Time) ([]TimeEnergy, error) {
	// prepare request
	val := url.Values{}
	val.Add("op", "getInverterData")
	val.Add("id", inverterSerial)
	// Not really sure what type 1 is, but different types seem to be returning
	// values in different units. It's unlike the type of GetPlantEnergy which
	// determines the timespan over which you want details
	val.Add("type", "1")
	val.Add("date", date.Format("2006-01-02"))

	data, err := a.call("GET", "newInverterAPI.do", val.Encode())
	if err != nil {
		return nil, err
	}

	var r struct {
		X map[string]interface{} `json:"-"` // we don't know the JSON keys
	}

	if err := json.Unmarshal(data, &r.X); err != nil {
		return nil, err
	}

	pacData, ok := r.X["invPacData"].(map[string]interface{})
	if !ok {
		return nil, errors.New("could not convert energy data")
	}

	var result []TimeEnergy
	for k, v := range pacData {
		ts, err := time.Parse("2006-01-02 15:04", k)
		if err != nil {
			continue
		}

		result = append(result, TimeEnergy{
			Timestamp: ts,
			Power:     v.(float64),
		})
	}

	return result, nil
}

func (a *API) call(method, path, data string) ([]byte, error) {
	// if API hasn't logged in yet, do so now
	if !a.isLoggedIn() && path != apiLoginPath {
		if err := a.login(); err != nil {
			return nil, err
		}
	}

	// prepare HTTP request
	var url string
	var req *http.Request
	var err error

	switch method {
	case "POST":
		url = getAPIURL(path, "")
		req, err = http.NewRequest(method, url, strings.NewReader(data))
	case "GET":
		url = getAPIURL(path, data)
		req, err = http.NewRequest(method, url, nil)
	}

	// we couldn't prepare the request
	if err != nil {
		return nil, err
	}

	// make sure to set proper content-type when sending data
	if method == "POST" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	// add SessionID and ServerID cookie if we're not currently trying to login
	if path != apiLoginPath {
		req.AddCookie(&http.Cookie{
			Name:  apiCookieSessionID,
			Value: a.sessionID,
		})
		req.AddCookie(&http.Cookie{
			Name:  apiCookieServerID,
			Value: a.serverID,
		})
	}

	// do HTTP request
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case 200: // do nothing, this is OK
	default:
		return nil, fmt.Errorf("API HTTP error: %s", resp.Status)
	}

	// read entire response body
	b, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var aErr struct {
		Back struct {
			ErrCode string `json:"errCode"`
			Succces bool   `json:"success"`
		} `json:"back"`
	}

	// if we succeed in unmarshalling the response into an API error, we'll just
	// return that error
	if err := json.Unmarshal(b, &aErr); err == nil && aErr.Back.ErrCode != "" {
		switch aErr.Back.ErrCode {
		case "502":
			return nil, ErrAPILogin
		}

		return nil, fmt.Errorf("API Error %s", aErr.Back.ErrCode)
	}

	// if this is a login attempt, record given sessionID and serverID
	if path == apiLoginPath {
		for _, c := range resp.Cookies() {
			switch c.Name {
			case apiCookieSessionID:
				a.sessionID = c.Value
			case apiCookieServerID:
				a.serverID = c.Value
			}
		}
	}

	return b, nil
}

func (a *API) login() error {
	// prepare form data
	val := url.Values{}
	val.Add("userName", a.username)
	val.Add("password", hashPassword(a.password))

	// do login request
	data, err := a.call("POST", apiLoginPath, val.Encode())
	if err != nil {
		return err
	}

	var r struct {
		Back struct {
			UserID    int  `json:"userId"`
			UserLevel int  `json:"userLevel"`
			Success   bool `json:"success"`
		} `json:"back"`
	}

	if err := json.Unmarshal(data, &r); err != nil {
		return err
	}

	// record response meta data
	a.UserID = r.Back.UserID
	a.UserLevel = r.Back.UserLevel

	return nil
}

func (a *API) isLoggedIn() bool {
	return (len(a.sessionID) > 0) && (len(a.serverID) > 0)
}

func getAPIURL(path, query string) string {
	glue := ""
	if !strings.HasPrefix(path, "/") {
		glue = "/"
	}

	if query != "" {
		query = fmt.Sprintf("?%s", query)
	}

	return fmt.Sprintf("%s%s%s%s", apiBasePath, glue, path, query)
}

func hashPassword(passwd string) string {
	var buf bytes.Buffer
	// hash password as MD5
	hash := md5.Sum([]byte(passwd))
	// go over each byte and add 12 (hex C) if the byte is less than 16
	for _, b := range hash {
		if b <= 0xf {
			b += 0xc0
		}

		buf.Write([]byte{b})

	}

	return fmt.Sprintf("%x", buf.String())
}
