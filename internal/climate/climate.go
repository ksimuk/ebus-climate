package climate

type Mode struct {
	heating bool
	// hotWater bool
}

type Sensor struct {
	temperature float64
	humidity    float64
}

type ClimateConfig struct {
	minTemperature         float64
	maxTemperature         float64
	hotWaterTemperature    int
	desiredFlowTemperature int
	isHotWater             bool
	isHeating              bool

	mode Mode

	roomSensor        Sensor
	heatingFire       bool
	flowTemperature   float64
	returnTemperature float64
}

type Info struct {
	Signal      string
	Scan        string
	Debug       string
	EbusVersion string
}

type Climate interface {
	Info() ([]string, error)
}
