forecastio
==========

A Go language library for accessing weather data from http://forecast.io

Firstly you will need an API key, sign up for one here https://developer.forecast.io/

To add this library to your app add
```
import "github.com/alsm/forecastio"
```
to your code.

In this library an APIConn is used to request weather data, an APIConn has API key and Units properties. The API key is the key from forecast.io that is to be used for the request and units specifies the units of the measurements that are returned.

An APIConn can have only one API key associated with it and the key cannot be changed once created. If you want to use a different key create a new APIConn. Multiple APIConn's can have the same API key with different Units settings. In this case the apiCalls property will not be uniformally updated, only the APIConn that made the most recent call will have the current correct apiCalls value.

An APIConn is created by calling
```
apiConn := forecastio.NewConnection("apikey")
```

There are two functions for requesting weather data; Forecast() and ForecastAtTime()  
The former will get the current weather data for the location requested, the latter will get either the historical weather data or predicted weather data for the location and the date/time provided.

More detailed API documentation on parameters is provided inline in the code.

There are two samples in the samples directory;  
CurrentWeather.go - CLI tool to get and print weather data.  
WebCurrentWeather.go - Uses http://goji.io to provide a simple example of getting location data from a browser and displaying the weather at that location