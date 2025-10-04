package climate

type Stat struct {
	UsageHeating    float64 `json:"usage_heating"`     // total consumption for heating in kWh
	UsageHotWater   float64 `json:"usage_hot_water"`   // total consumption for hot water in kWh
	CurrentHeatLoss float64 `json:"current_heat_loss"` // current heat loss in W
	WaterPressure   float64 `json:"water_pressure"`    // current water pressure in bar
}
