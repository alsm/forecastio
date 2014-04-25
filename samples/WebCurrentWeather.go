package main

import (
	"flag"
	"github.com/alsm/forecastio"
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"html/template"
	"net/http"
	"strconv"
)

var rootTemplate string = `
<!DOCTYPE html>
<html>
<head>
	<title>Go Weather!</title>
</head>
<body>
<h1>Go Weather</h1>
<p id="weather"></p>
<script>
var x = document.getElementById("weather");
navigator.geolocation.getCurrentPosition(function (position)
  {
  	var xmlhttp = new XMLHttpRequest();
  	xmlhttp.onreadystatechange = function() {
  		x.innerHTML = xmlhttp.responseText;
  	}
  	xmlhttp.open("GET", "/weather/" + position.coords.latitude + "/" + position.coords.longitude, true);
  	xmlhttp.send();
  });
</script>
</body>
</html>
`

var weatherTemplate string = `
Latitude: {{ .Latitude }} Longitude: {{ .Longitude }}<br>
Current Weather -<br>
Report Time: {{ .Currently.Time.Format "02/Jan/2006 - 15:04" }}  Summary: {{ .Currently.Summary }}<br>
Temperature: {{ printf "%2.0f°" .Currently.Temperature }}  Pressure: {{ printf "%.0fmb" .Currently.Pressure }}  Humidity: {{ printf "%.2f" .Currently.Humidity }}<br>
Summary for tomorrow: {{ .Daily.Summary }}<br>
{{ range .Daily.Data }}
{{ .Time.Format "02/Jan/2006" }} - Temperature (Min/Max): {{ printf "%2.0f/%2.0f°" .TemperatureMin .TemperatureMax }}  Pressure: {{ printf "%4.0fmb" .Pressure }} - {{ .Summary }}<br>
{{ end }}
`

var key string

func init() {
	flag.StringVar(&key, "apikey", "", "API key for forecast.io")
}

func root(c web.C, w http.ResponseWriter, r *http.Request) {
	t := template.Must(template.New("root").Parse(rootTemplate))
	t.Execute(w, nil)
}

func weather(c web.C, w http.ResponseWriter, r *http.Request) {
	fConn := forecastio.NewConnection(key)
	excludes := []string{"hourly", "minutely"}
	lat, _ := strconv.ParseFloat(c.URLParams["lat"], 64)
	lon, _ := strconv.ParseFloat(c.URLParams["lon"], 64)
	t := template.Must(template.New("weather").Parse(weatherTemplate))
	report, _ := fConn.Forecast(lat, lon, excludes, false)
	report.ParseTimes()
	t.Execute(w, report)
}

func main() {
	goji.Get("/", root)
	goji.Get("/weather/:lat/:lon", weather)
	goji.Serve()
}
