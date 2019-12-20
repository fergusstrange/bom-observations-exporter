package main

import (
	"fmt"
	"strconv"
	"time"
)

type BomObservationWrapper struct {
	Observations struct {
		Data []BomObservation `json:"data"`
	} `json:"observations"`
}

// BomObservation based on details found here http://www.bom.gov.au/catalogue/observations/about.shtml
type BomObservation struct {
	WMO                 int     `json:"wmo"`
	Name                string  `json:"name"`
	TimeZoneName        string  `json:"time_zone_name"`
	TDZ                 string  `json:"TDZ"`
	TimeUTC             string  `json:"aifstime_utc"`
	TimeLocal           string  `json:"aifstime_local"`
	Lat                 float64 `json:"lat"`
	Lon                 float64 `json:"lon"`
	ApparentT           float64 `json:"apparent_t"`
	DeltaT              float64 `json:"delta_t"`
	AirTemp             float64 `json:"air_temp"`
	GustKmh             float64 `json:"gust_kmh"`
	GustKt              float64 `json:"gust_kt"`
	DewPoint            float64 `json:"dewpt"`
	Press               float64 `json:"press"`
	PressMsl            float64 `json:"press_msl"`
	PressQnh            float64 `json:"press_qnh"`
	PressTend           string  `json:"press_tend"`
	RainHour            float64 `json:"rain_hour"`
	RainTen             float64 `json:"rain_ten"`
	RainTrace           string  `json:"rain_trace"`
	RainTraceTimeUTC    string  `json:"rain_trace_time_utc"`
	Local9AmDateTimeUTC string  `json:"local_9am_date_time_utc"`
	RelHum              float64 `json:"rel_hum"`
	WindDirDeg          float64 `json:"wind_dir_deg"`
	WindSpdKmh          float64 `json:"wind_spd_kmh"`
}

type Location struct {
	Lat float64 `json:"lat"`
	Lon float64 `json:"lon"`
}

type ElasticSearchBomObservation struct {
	WMO                 int       `json:"wmo"`
	Name                string    `json:"name"`
	Timestamp           time.Time `json:"timestamp"`
	Location            Location  `json:"location"`
	TemperatureApparent float64   `json:"temperature_apparent"`
	TemperatureDelta    float64   `json:"temperature_delta"`
	Temperature         float64   `json:"temperature"`
	DewPoint            float64   `json:"dew_point"`
	Humidity            float64   `json:"humidity"`
	PressureMSL         float64   `json:"pressure_msl"`
	PressureQNH         float64   `json:"pressure_qnh"`
	WindSpeed           float64   `json:"wind_speed"`
	WindDirection       float64   `json:"wind_direction"`
	RainfallTrace       float64   `json:"rainfall_trace"`
	RainfallHour        float64   `json:"rainfall_hour"`
	RainfallTen         float64   `json:"rainfall_ten"`
}

func (b BomObservation) ToElasticSearchBomObservation() (ElasticSearchBomObservation, error) {
	timestamp, err := time.Parse(timeFormat, fmt.Sprintf("%s %s", b.TimeLocal, b.TDZ))
	if err != nil {
		return ElasticSearchBomObservation{}, err
	}

	rainTrace, err := strconv.ParseFloat(b.RainTrace, 64)
	if err != nil {
		rainTrace = 0
	}

	return ElasticSearchBomObservation{
		WMO:       b.WMO,
		Name:      b.Name,
		Timestamp: timestamp,
		Location: struct {
			Lat float64 `json:"lat"`
			Lon float64 `json:"lon"`
		}{
			Lat: b.Lat,
			Lon: b.Lon,
		},
		TemperatureApparent: b.ApparentT,
		TemperatureDelta:    b.DeltaT,
		Temperature:         b.AirTemp,
		DewPoint:            b.DewPoint,
		Humidity:            b.RelHum,
		PressureMSL:         b.PressMsl,
		PressureQNH:         b.PressQnh,
		WindSpeed:           b.WindSpdKmh,
		WindDirection:       b.WindDirDeg,
		RainfallTrace:       rainTrace,
		RainfallHour:        b.RainHour,
		RainfallTen:         b.RainTen,
	}, nil
}
