package forecastio

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"
)

// The base part of any constructed URL to request weather data from forecast.io
const (
	baseURL = "https://api.forecast.io/forecast"
)

// excludesSet and unitsSet are effectively sets used to verify options passed
// as excludes or units are valid
var (
	excludesSet map[string]struct{} = map[string]struct{}{
		"currently": struct{}{},
		"minutely":  struct{}{},
		"hourly":    struct{}{},
		"daily":     struct{}{},
		"alerts":    struct{}{},
		"flags":     struct{}{},
	}
	unitsSet map[string]struct{} = map[string]struct{}{
		"us":   struct{}{},
		"si":   struct{}{},
		"ca":   struct{}{},
		"uk":   struct{}{},
		"auto": struct{}{},
	}
)

// APIConn represents a connection to the forecast.io API, each APIConn has
// it's own API key and units settings, it also contains a counter for the
// number of API calls that day, this value is not populated until a Forecast()
// or ForecastAtTime() call is made with this APIConn.
type APIConn struct {
	sync.RWMutex
	apiKey   string
	apiCalls int
	units    string
}

// NewConnection returns a new *APIConn setting the APIkey and the units
// property to "auto", this means that by default the units of values returned
// will be in the standard units for that country, eg; imperial in the US, a
// mix in the UK, metric in France.
func NewConnection(key string) *APIConn {
	return &APIConn{apiKey: key, units: "auto"}
}

// APICalls returns an integer of the current number of APICalls made that
// day by the API key contained in the APIConn, this value will be 0 until a
// Forecast() or ForecastAtTime() call has been made.
func (a *APIConn) APICalls() int {
	a.RLock()
	defer a.RUnlock()
	return a.apiCalls
}

// Returns the current units setting for this APIConn
func (a *APIConn) Units() string {
	a.RLock()
	defer a.RUnlock()
	return a.units
}

// SetUnits set the units property for the APIConn to the value passed in
// unless the units value requested is not known.
// Returns a nil error if successful
func (a *APIConn) SetUnits(units string) error {
	a.Lock()
	defer a.Unlock()
	if _, ok := unitsSet[units]; ok {
		a.units = units
		return nil
	} else {
		return errors.New("Invalid units requested")
	}
}

type currently struct {
	TimeUnix                 int64 `json:"time"`
	Time                     time.Time
	Summary                  string  `json:"summary"`
	Icon                     string  `json:"icon"`
	NearestStormDistance     int     `json:"nearestStormDistance"`
	NearestStormBearing      int     `json:"nearestStormBearing"`
	PrecipitationIntensity   float64 `json:"precipIntensity"`
	PrecipitationProbability float64 `json:"precipProbability"`
	Temperature              float64 `json:"temperature"`
	ApparentTemperature      float64 `json:"apparentTemperature"`
	DewPoint                 float64 `json:"dewPoint"`
	Humidity                 float64 `json:"humidity"`
	WindSpeed                float64 `json:"windSpeed"`
	WindBearing              float64 `json:"windBearing"`
	Visibility               float64 `json:"visibility"`
	CloudCover               float64 `json:"cloudCover"`
	Pressure                 float64 `json:"pressure"`
	Ozone                    float64 `json:"ozone"`
}

type minuteData struct {
	TimeUnix                 int64 `json:"time"`
	Time                     time.Time
	PrecipitationIntensity   float64 `json:"precipIntensity"`
	PrecipitationProbability float64 `json:"precipProbability"`
}

type minutely struct {
	Summary string        `json:"summary"`
	Icon    string        `json:"icon"`
	Data    []*minuteData `json:"data"`
}

type hourData struct {
	TimeUnix                 int64 `json:"time"`
	Time                     time.Time
	Summary                  string  `json:"summary"`
	Icon                     string  `json:"icon"`
	PrecipitationIntensity   float64 `json:"precipIntensity"`
	PrecipitationProbability float64 `json:"precipProbability"`
	PrecipitationType        string  `json:"precipType"`
	Temperature              float64 `json:"temperature"`
	ApparentTemperature      float64 `json:"apparentTemperature"`
	DewPoint                 float64 `json:"dewPoint"`
	Humidity                 float64 `json:"humidity"`
	WindSpeed                float64 `json:"windSpeed"`
	WindBearing              float64 `json:"windBearing"`
	Visibility               float64 `json:"visibility"`
	CloudCover               float64 `json:"cloudCover"`
	Pressure                 float64 `json:"pressure"`
	Ozone                    float64 `json:"ozone"`
}

type hourly struct {
	Summary string      `json:"summary"`
	Icon    string      `json:"icon"`
	Data    []*hourData `json:"data"`
}

type dayData struct {
	TimeUnix                          int64 `json:"time"`
	Time                              time.Time
	Summary                           string `json:"summary"`
	Icon                              string `json:"icon"`
	SunriseUnix                       int64  `json:"sunriseTime"`
	Sunrise                           time.Time
	SunsetUnix                        int64 `json:"sunsetTime"`
	Sunset                            time.Time
	MoonPhase                         float64 `json:"moonPhase"`
	PrecipitationIntensity            float64 `json:"precipIntensity"`
	PrecipitationIntensityMax         float64 `json:"precipIntensityMax"`
	PrecipitationIntensityMaxTimeUnix int64   `json:"precipIntensityMaxTime"`
	PrecipitationIntensityMaxTime     time.Time
	PrecipitationProbability          float64 `json:"precipProbability"`
	PrecipitationType                 string  `json:"precipType"`
	TemperatureMin                    float64 `json:"temperatureMin"`
	TemperatureMinTimeUnix            int64   `json:"temperatureMinTime"`
	TemperatureMinTime                time.Time
	TemperatureMax                    float64 `json:"temperatureMax"`
	TemperatureMaxTimeUnix            int64   `json:"temperatureMaxTime"`
	TemperatureMaxTime                time.Time
	ApparentTemperatureMin            float64 `json:"apparentTemperatureMin"`
	ApparentTemperatureMinTimeUnix    int64   `json:"apparentTemperatureMinTime"`
	ApparentTemperatureMinTime        time.Time
	ApparentTemperatureMax            float64 `json:"apparentTemperatureMax"`
	ApparentTemperatureMaxTimeUnix    int64   `json:"apparentTemperatureMaxTime"`
	ApparentTemperatureMaxTime        time.Time
	DewPoint                          float64 `json:"dewPoint"`
	Humidity                          float64 `json:"humidity"`
	WindSpeed                         float64 `json:"windSpeed"`
	WindBearing                       float64 `json:"windBearing"`
	Visibility                        float64 `json:"visibility"`
	CloudCover                        float64 `json:"cloudCover"`
	Pressure                          float64 `json:"pressure"`
	Ozone                             float64 `json:"ozone"`
}

type daily struct {
	Summary string     `json:"summary"`
	Icon    string     `json:"icon"`
	Data    []*dayData `json:"data"`
}

type flags struct {
	Sources           []string `json:"sources"`
	ISDStations       []string `json:"isd-stations"`
	MadisStations     []string `json:"madis-stations"`
	DatapointStations []string `json:"datapoint-stations"`
	DarkskyStations   []string `json:"darksky-stations"`
	Units             string
}

type alert struct {
	Title       string `json:"title"`
	ExpiresUnix int64  `json:"expires"`
	Expires     time.Time
	Description string `json:"description"`
	URI         string `json:"uri"`
}

type Forecast struct {
	Latitude   float64   `json:"latitude"`
	Longitude  float64   `json:"longitude"`
	Timezone   string    `json:"timezone"`
	TimeOffset int       `json:"offset"`
	Currently  currently `json:"currently"`
	Minutely   minutely  `json:"minutely"`
	Hourly     hourly    `json:"hourly"`
	Daily      daily     `json:"daily"`
	Alerts     []*alert  `json:"alerts"`
	Flags      flags     `json:"flags"`
}

// ParseTimes will fill out all time.Time variables in a Forecast by
// converting the Unix time values returned by forecast.io
func (f *Forecast) ParseTimes() {
	f.Currently.Time = time.Unix(f.Currently.TimeUnix, 0)
	for _, m := range f.Minutely.Data {
		m.Time = time.Unix(m.TimeUnix, 0)
	}
	for _, h := range f.Hourly.Data {
		h.Time = time.Unix(h.TimeUnix, 0)
	}
	for _, d := range f.Daily.Data {
		d.Time = time.Unix(d.TimeUnix, 0)
		d.Sunrise = time.Unix(d.SunriseUnix, 0)
		d.Sunset = time.Unix(d.SunsetUnix, 0)
		d.PrecipitationIntensityMaxTime = time.Unix(d.PrecipitationIntensityMaxTimeUnix, 0)
		d.TemperatureMinTime = time.Unix(d.TemperatureMinTimeUnix, 0)
		d.TemperatureMaxTime = time.Unix(d.TemperatureMaxTimeUnix, 0)
		d.ApparentTemperatureMinTime = time.Unix(d.ApparentTemperatureMinTimeUnix, 0)
		d.ApparentTemperatureMaxTime = time.Unix(d.ApparentTemperatureMaxTimeUnix, 0)
	}
	for _, a := range f.Alerts {
		a.Expires = time.Unix(a.ExpiresUnix, 0)
	}
}

// Forecast requests a forecast from forecastio using the APIConn a.
// lat and lon are float64s representing the latitude and longitude of
// the location the forecast is for.
// excludes is an array of strings for fields that are to be excluded
// from the forecast, valid exludes are;
//     currently, minutely, hourly, daily, alerts, flags
// extendHourly is a boolean flag indicating whether to return hourly
// data for 7 days rather than the default of 2 days.
// Returns a pointer to a Forecast and an error. The two are mutually
// exclusive in that one will always be nil.
func (a *APIConn) Forecast(lat, lon float64, excludes []string, extendHourly bool) (*Forecast, error) {
	var forecast Forecast

	for _, ex := range excludes {
		if ex == "" {
			continue
		}
		if _, ok := excludesSet[ex]; !ok {
			return nil, errors.New("Invalid exclude requested")
		}
	}
	query := fmt.Sprintf("%s/%s/%f,%f?units=%s&exclude=%s", baseURL, a.apiKey, lat, lon, a.units, strings.Join(excludes, ","))
	if extendHourly {
		query += "&extend=hourly"
	}

	a.Lock()
	defer a.Unlock()
	response, err := http.Get(query)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	response.Body.Close()
	err = json.Unmarshal(body, &forecast)
	a.apiCalls, _ = strconv.Atoi(response.Header.Get("X-Forecast-API-Calls"))
	return &forecast, err
}

// ForecastAtTime requests a forecast from forecastio using the APIConn a.
// for a specific point in time.
// lat and lon are float64s representing the latitude and longitude of
// the location the forecast is for.
// date should be either a time.Time, an int/int64 of the unix time for
// formatted as follows: [YYYY]-[MM]-[DD]T[HH]:[MM]:[SS] (with an optional
// time zone formatted as Z for GMT time or {+,-}[HH][MM], for example;
//     2013-05-06T12:00:00-0400
// the point in time requested or a string. If a string it must be either
// a string representation of the unix time or
// excludes is an array of strings for fields that are to be excluded
// from the forecast, valid exludes are;
//     currently, minutely, hourly, daily, alerts, flags
// Returns a pointer to a Forecast and an error. The two are mutually
// exclusive in that one will always be nil.
func (a *APIConn) ForecastAtTime(lat, lon float64, date interface{}, excludes []string) (*Forecast, error) {
	var forecast Forecast
	var query string

	for _, ex := range excludes {
		if _, ok := excludesSet[ex]; !ok {
			return nil, errors.New("Invalid exclude requested")
		}
	}
	switch date.(type) {
	case time.Time:
		query = fmt.Sprintf("%s/%s/%f,%f,%d?units=%s&exclude=%s", baseURL, a.apiKey, lat, lon, date.(time.Time).Unix(), a.units, strings.Join(excludes, ","))
	case int, int64:
		query = fmt.Sprintf("%s/%s/%f,%f,%d?units=%s&exclude=%s", baseURL, a.apiKey, lat, lon, date.(int64), a.units, strings.Join(excludes, ","))
	case string:
		query = fmt.Sprintf("%s/%s/%f,%f,%s?units=%s&exclude=%s", baseURL, a.apiKey, lat, lon, date.(string), a.units, strings.Join(excludes, ","))
	}

	a.Lock()
	defer a.Unlock()
	response, err := http.Get(query)
	if err != nil {
		return nil, err
	}
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	response.Body.Close()
	err = json.Unmarshal(body, &forecast)
	a.apiCalls, _ = strconv.Atoi(response.Header.Get("X-Forecast-API-Calls"))
	return &forecast, err
}
