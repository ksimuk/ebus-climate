package climate

type Climate interface {
	Info() ([]string, error)
	SetInsideOverride(temp float64)
	SetOutsideOverride(temp float64)
	GetInsideTemp() float64
	GetOutsideTemp() float64
	GetMode() string
	GetTargetTemperature() float64
	GetHWTargetTemp() int

	GetFlowTemp() float64
	GetReturnTemp() float64
	GetDesiredFlowTemp() int
	GetPower() int

	IsGasActive() bool
	IsPumpActive() bool
	IsConnected() bool
	GetError() string

	GetBoilerInfo() BoilerInfo

	SetMode(mode string) error
	SetTargetTemperature(temp float64) error
	SetHWTargetTemp(temp int) error
	StartHeating()
	StopHeating()

	GetConsumption() float64
}

type BoilerInfo struct {
	Model    string `json:"model"`
	Firmware string `json:"firmware"`
}
