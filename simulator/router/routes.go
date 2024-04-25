package router

import (
	"github.com/gin-gonic/gin"
	"github.com/rmkhl/halko/simulator/types"
)

func SetupRoutes(r *gin.Engine, temperature_sensors map[string]types.TemperatureSensor, power_sensors map[string]types.PowerSensor, power_controls map[string]types.PowerManager) {
	sensor_api := r.Group("sensors/api")
	control_api := r.Group("controls/api")

	sensor_api_v1 := sensor_api.Group("v1")
	control_api_v1 := control_api.Group("v1")

	temp_sensors := sensor_api_v1.Group("temperatures")
	temp_sensors.GET("", readAllTemperatureSensors(temperature_sensors))
	temp_sensors.GET(":sensor", readTemperatureSensor(temperature_sensors))

	psu_sensors := sensor_api_v1.Group("powers")
	psu_sensors.GET("", statusAllPowers(power_sensors))
	psu_sensors.GET(":power", statusPower(power_sensors))

	controls := control_api_v1.Group("powers")
	controls.POST(":power", operatePower(power_controls))
	controls.PUT(":power", operatePower(power_controls))
	controls.PATCH(":power", operatePower(power_controls))
}
