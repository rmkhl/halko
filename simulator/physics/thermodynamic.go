package physics

import (
	"errors"
	"fmt"
	"math"

	"github.com/rmkhl/halko/types/log"
)

// Configuration section and field names shared by Initialize and ValidateConfig.
const (
	sectionKiln        = "kiln"
	sectionAir         = "air"
	sectionMaterial    = "material"
	sectionHeater      = "heater"
	sectionConvection  = "convection"
	sectionEnvironment = "environment"
	sectionPhysics     = "physics"

	keyMass         = "mass"
	keySpecificHeat = "specific_heat"
	keySurfaceArea  = "surface_area"
	keyTimeStep     = "time_step"
)

// ThermodynamicSimulation implements physics-based heat transfer using real material properties
// and thermodynamic principles (conduction, convection, radiation)
type ThermodynamicSimulation struct {
	// Kiln properties
	kilnMass         float32 // kg - steel walls
	kilnSpecificHeat float32 // J/kg·K - steel
	kilnSurfaceArea  float32 // m² - exterior surface
	wallUValue       float32 // W/m²·K - overall heat transfer coefficient (includes insulation)
	kilnEmissivity   float32 // 0-1 - for radiation

	// Air properties
	airVolume       float32 // m³
	airSpecificHeat float32 // J/kg·K

	// Material (wood) properties
	materialMass         float32 // kg
	materialSpecificHeat float32 // J/kg·K
	materialSurfaceArea  float32 // m²

	// Heater properties
	heaterWattage    float32 // W - electrical power
	heaterEfficiency float32 // 0-1 - electrical to thermal conversion

	// Convection coefficients
	convectionNatural float32 // W/m²·K - fan off
	convectionForced  float32 // W/m²·K - fan on
	fanWasteHeat      float32 // W - fan motor heat

	// Environment
	ambientTemp float32 // °C

	// Constants
	stefanBoltzmann float32 // W/m²·K⁴
	timeStep        float32 // seconds per tick
}

func (t *ThermodynamicSimulation) Name() string {
	return engineThermodynamic
}

func (t *ThermodynamicSimulation) Initialize(config map[string]interface{}) error {
	// Kiln properties
	kiln := config[sectionKiln].(map[string]interface{})
	t.kilnMass = float32(kiln[keyMass].(float64))
	t.kilnSpecificHeat = float32(kiln[keySpecificHeat].(float64))
	t.kilnSurfaceArea = float32(kiln[keySurfaceArea].(float64))
	t.wallUValue = float32(kiln["wall_u_value"].(float64))
	t.kilnEmissivity = float32(kiln["emissivity"].(float64))

	// Air properties
	air := config[sectionAir].(map[string]interface{})
	t.airVolume = float32(air["volume"].(float64))
	t.airSpecificHeat = float32(air[keySpecificHeat].(float64))

	// Material properties
	material := config[sectionMaterial].(map[string]interface{})
	t.materialMass = float32(material[keyMass].(float64))
	t.materialSpecificHeat = float32(material[keySpecificHeat].(float64))
	t.materialSurfaceArea = float32(material[keySurfaceArea].(float64))

	// Heater properties
	heater := config[sectionHeater].(map[string]interface{})
	t.heaterWattage = float32(heater["wattage"].(float64))
	t.heaterEfficiency = float32(heater["efficiency"].(float64))

	// Convection properties
	convection := config[sectionConvection].(map[string]interface{})
	t.convectionNatural = float32(convection["natural"].(float64))
	t.convectionForced = float32(convection["forced"].(float64))
	t.fanWasteHeat = float32(convection["fan_waste_heat"].(float64))

	// Environment
	env := config[sectionEnvironment].(map[string]interface{})
	t.ambientTemp = float32(env["temperature"].(float64))

	// Physical constants
	physics := config[sectionPhysics].(map[string]interface{})
	t.stefanBoltzmann = float32(physics["stefan_boltzmann"].(float64))
	t.timeStep = float32(physics[keyTimeStep].(float64))

	log.Info("Thermodynamic simulation initialized:")
	log.Info("  Kiln: %.0f kg steel, %.1f m² surface, U=%.2f W/m²·K", t.kilnMass, t.kilnSurfaceArea, t.wallUValue)
	log.Info("  Material: %.0f kg @ %.0f J/kg·K, %.1f m² surface", t.materialMass, t.materialSpecificHeat, t.materialSurfaceArea)
	log.Info("  Heater: %.0f W @ %.0f%% efficiency", t.heaterWattage, t.heaterEfficiency*100)
	log.Info("  Convection: natural=%.1f, forced=%.1f W/m²·K", t.convectionNatural, t.convectionForced)

	return nil
}

// section returns a named configuration section, which has to be an object.
func section(config map[string]interface{}, name string) (map[string]interface{}, error) {
	raw, exists := config[name]
	if !exists {
		return nil, fmt.Errorf("required configuration section missing: %s", name)
	}

	values, ok := raw.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("configuration section %s must be an object", name)
	}
	return values, nil
}

// number checks that a section holds a numeric value under key. Initialize
// type-asserts these without checking, so everything it reads has to be
// validated here or a bad config panics the simulator at startup.
func number(values map[string]interface{}, sectionName, key string, mustBePositive bool) error {
	raw, exists := values[key]
	if !exists {
		return fmt.Errorf("%s.%s is required", sectionName, key)
	}

	value, ok := raw.(float64)
	if !ok {
		return fmt.Errorf("%s.%s must be a number", sectionName, key)
	}
	if mustBePositive && value <= 0 {
		return fmt.Errorf("%s.%s must be positive, got: %f", sectionName, key, value)
	}
	return nil
}

func (t *ThermodynamicSimulation) ValidateConfig(config map[string]interface{}) error {
	if config == nil {
		return errors.New("config cannot be nil")
	}

	// The fields Initialize reads, and whether a zero would leave the model
	// dividing by a thermal capacity of zero.
	required := []struct {
		section  string
		key      string
		positive bool
	}{
		{sectionKiln, keyMass, true},
		{sectionKiln, keySpecificHeat, true},
		{sectionKiln, keySurfaceArea, false},
		{sectionKiln, "wall_u_value", false},
		{sectionKiln, "emissivity", false},
		{sectionAir, "volume", false},
		{sectionAir, keySpecificHeat, false},
		{sectionMaterial, keyMass, true},
		{sectionMaterial, keySpecificHeat, true},
		{sectionMaterial, keySurfaceArea, false},
		{sectionHeater, "wattage", false},
		{sectionHeater, "efficiency", false},
		{sectionConvection, "natural", false},
		{sectionConvection, "forced", false},
		{sectionConvection, "fan_waste_heat", false},
		{sectionEnvironment, "temperature", false},
		{sectionPhysics, "stefan_boltzmann", false},
		{sectionPhysics, keyTimeStep, true},
	}

	sections := make(map[string]map[string]interface{})
	for _, field := range required {
		values, ok := sections[field.section]
		if !ok {
			var err error
			if values, err = section(config, field.section); err != nil {
				return err
			}
			sections[field.section] = values
		}

		if err := number(values, field.section, field.key, field.positive); err != nil {
			return err
		}
	}

	return nil
}

func (t *ThermodynamicSimulation) Tick(state *SimulationState) {
	// Calculate air density at current temperature (ideal gas approximation)
	// ρ(T) = ρ₀ × (T₀ / T) where T in Kelvin
	tempKelvin := state.KilnTemp + 273.15
	airDensity := 1.2 * (293.15 / tempKelvin) // kg/m³
	airMass := airDensity * t.airVolume

	// Total thermal mass of kiln air space (steel + air)
	kilnThermalCapacity := t.kilnMass*t.kilnSpecificHeat + airMass*t.airSpecificHeat // J/K

	// Heater energy input (Watts × seconds = Joules)
	var heaterEnergy float32
	if state.HeaterIsOn {
		heaterEnergy = t.heaterWattage * t.heaterEfficiency * t.timeStep
	}

	// Fan adds waste heat if running
	if state.FanIsOn {
		heaterEnergy += t.fanWasteHeat * t.timeStep
	}

	// Conduction heat loss through walls (W = U × A × ΔT)
	wallDeltaT := state.KilnTemp - t.ambientTemp
	conductionLoss := t.wallUValue * t.kilnSurfaceArea * wallDeltaT * t.timeStep // Joules

	// Radiation heat loss (Stefan-Boltzmann law)
	// Q = ε × σ × A × (T⁴_hot - T⁴_cold)
	kilnTempK := state.KilnTemp + 273.15
	ambientTempK := t.ambientTemp + 273.15
	radiationPower := t.kilnEmissivity * t.stefanBoltzmann * t.kilnSurfaceArea *
		(float32(math.Pow(float64(kilnTempK), 4)) - float32(math.Pow(float64(ambientTempK), 4)))
	radiationLoss := radiationPower * t.timeStep // Joules

	// Convection between kiln air and material
	// h depends on fan state
	var convectionCoeff float32
	if state.FanIsOn {
		convectionCoeff = t.convectionForced
	} else {
		convectionCoeff = t.convectionNatural
	}

	// Heat transfer from air to material (W = h × A × ΔT)
	airMaterialDelta := state.KilnTemp - state.MaterialTemp
	convectionPower := convectionCoeff * t.materialSurfaceArea * airMaterialDelta
	convectionEnergy := convectionPower * t.timeStep // Joules

	// Energy balance for kiln/air
	kilnNetEnergy := heaterEnergy - conductionLoss - radiationLoss - convectionEnergy
	kilnTempChange := kilnNetEnergy / kilnThermalCapacity

	oldKilnTemp := state.KilnTemp
	state.KilnTemp = max(t.ambientTemp, state.KilnTemp+kilnTempChange)

	// Energy balance for material
	// Material also loses some heat to environment via radiation/convection (less than kiln)
	materialWallDelta := state.MaterialTemp - t.ambientTemp
	materialLossPower := 0.1 * convectionCoeff * t.materialSurfaceArea * materialWallDelta // reduced coefficient
	materialLossEnergy := materialLossPower * t.timeStep

	materialNetEnergy := convectionEnergy - materialLossEnergy
	materialThermalCapacity := t.materialMass * t.materialSpecificHeat
	materialTempChange := materialNetEnergy / materialThermalCapacity

	oldMaterialTemp := state.MaterialTemp
	state.MaterialTemp = max(t.ambientTemp, state.MaterialTemp+materialTempChange)

	// Detailed logging
	if state.HeaterIsOn || state.KilnTemp != oldKilnTemp {
		log.Debug("Simulation[thermodynamic]: Kiln - heater=%.0fJ, loss(cond=%.0fJ, rad=%.0fJ), transfer=%.0fJ → %.1f°C → %.1f°C (Δ%.2f°C)",
			heaterEnergy, conductionLoss, radiationLoss, convectionEnergy,
			oldKilnTemp, state.KilnTemp, kilnTempChange)
	}

	if state.MaterialTemp != oldMaterialTemp {
		log.Debug("Simulation[thermodynamic]: Material - received=%.0fJ, loss=%.0fJ → %.1f°C → %.1f°C (Δ%.2f°C)",
			convectionEnergy, materialLossEnergy,
			oldMaterialTemp, state.MaterialTemp, materialTempChange)
	}

	// Log energy efficiency summary periodically (every 10th tick would be in main loop)
	if state.HeaterIsOn {
		efficiency := (convectionEnergy / (heaterEnergy + 0.001)) * 100 // % of heater energy going to material
		log.Debug("Simulation[thermodynamic]: Energy flow - heater→material: %.0f%%, losses: %.0f%%",
			efficiency, 100-efficiency)
	}
}
