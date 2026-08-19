package serial

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/rmkhl/halko/types"
	"github.com/rmkhl/halko/types/log"
)

// expectedReadings is how many values the unit reports per read: three
// thermocouples, then the cold junction each one is referenced to.
const expectedReadings = 6

// parseTemperatureResponse turns a `read` response of the form
// `KilnPrimary=XX.XC,KilnSecondary=XX.XC,Wood=XX.XC,KilnPrimaryDie=XX.XC,
// KilnSecondaryDie=XX.XC,WoodDie=XX.XC` into readings. A sensor that failed
// reports `NaN`, which becomes types.InvalidTemperatureReading. Anything
// else that will not parse rejects the whole response: a partially read line
// would silently lose a sensor, and the caller can retry.
func parseTemperatureResponse(response string) ([]Temperature, error) {
	readings := strings.Split(response, ",")
	if len(readings) != expectedReadings {
		return nil, fmt.Errorf("invalid temperature reading format, expected %d readings: %q", expectedReadings, response)
	}

	temperatures := make([]Temperature, 0, expectedReadings)
	for _, reading := range readings {
		name, valueStr, found := strings.Cut(reading, "=")
		if !found {
			return nil, fmt.Errorf("malformed temperature reading %q in response %q", reading, response)
		}

		if valueStr == "NaN" {
			log.Debug("Sensor %q has invalid reading (NaN)", name)
			temperatures = append(temperatures, Temperature{
				Name:  name,
				Value: types.InvalidTemperatureReading,
				Unit:  "C",
			})
			continue
		}

		if len(valueStr) < 2 {
			return nil, fmt.Errorf("malformed temperature value %q for sensor %q", valueStr, name)
		}
		unit := valueStr[len(valueStr)-1:]

		valueF64, err := strconv.ParseFloat(valueStr[:len(valueStr)-1], 32)
		if err != nil {
			return nil, fmt.Errorf("cannot parse temperature %q for sensor %q: %w", valueStr, name, err)
		}
		value := float32(valueF64)

		temperatures = append(temperatures, Temperature{Name: name, Value: value, Unit: unit})
	}

	return temperatures, nil
}
