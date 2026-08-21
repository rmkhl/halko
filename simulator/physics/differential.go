package physics

import (
	"errors"
	"fmt"

	"github.com/rmkhl/halko/types/log"
)

// DifferentialSimulation implements temperature-differential physics where heat transfer
// rates are proportional to temperature differences (Newton's Law of Cooling)
const (
	keyHeaterPower             = "heater_power"
	keyHeatLossCoefficient     = "heat_loss_coefficient"
	keyHeatTransferCoefficient = "heat_transfer_coefficient"
	keyKilnThermalMass         = "kiln_thermal_mass"
	keyMaterialThermalMass     = "material_thermal_mass"
	keySteamPower              = "steam_power"
	keySteamCutoffTemp         = "steam_cutoff_temp"
)

type DifferentialSimulation struct {
	heaterPower             float32 // Energy per tick when heater is on
	heatLossCoefficient     float32 // Proportional heat loss from the kiln to the environment
	heatTransferCoefficient float32 // Heat transfer rate between kiln and material
	kilnThermalMass         float32 // Heat capacity of kiln (energy needed to raise 1°C)
	materialThermalMass     float32 // Heat capacity of material
	steamPower              float32 // Energy per tick when steam is on and the kiln is below the cutoff
	steamCutoffTemp         float32 // Kiln temperature at which steam stops contributing heat
}

func (d *DifferentialSimulation) Name() string {
	return engineDifferential
}

func (d *DifferentialSimulation) Initialize(config map[string]interface{}) error {
	// Extract and validate configuration
	heaterPower, ok := config[keyHeaterPower].(float64)
	if !ok {
		return errors.New("heater_power must be specified as a number")
	}
	d.heaterPower = float32(heaterPower)

	heatLossCoeff, ok := config[keyHeatLossCoefficient].(float64)
	if !ok {
		return errors.New("heat_loss_coefficient must be specified as a number")
	}
	d.heatLossCoefficient = float32(heatLossCoeff)

	heatTransferCoeff, ok := config[keyHeatTransferCoefficient].(float64)
	if !ok {
		return errors.New("heat_transfer_coefficient must be specified as a number")
	}
	d.heatTransferCoefficient = float32(heatTransferCoeff)

	kilnMass, ok := config[keyKilnThermalMass].(float64)
	if !ok {
		return errors.New("kiln_thermal_mass must be specified as a number")
	}
	d.kilnThermalMass = float32(kilnMass)

	materialMass, ok := config[keyMaterialThermalMass].(float64)
	if !ok {
		return errors.New("material_thermal_mass must be specified as a number")
	}
	d.materialThermalMass = float32(materialMass)

	steamPower, ok := config[keySteamPower].(float64)
	if !ok {
		return errors.New("steam_power must be specified as a number")
	}
	d.steamPower = float32(steamPower)

	steamCutoff, ok := config[keySteamCutoffTemp].(float64)
	if !ok {
		return errors.New("steam_cutoff_temp must be specified as a number")
	}
	d.steamCutoffTemp = float32(steamCutoff)

	log.Info("Differential simulation initialized: heater_power=%.2f, heat_loss=%.5f, heat_transfer=%.3f, kiln_mass=%.1f, material_mass=%.2f, steam_power=%.2f below %.0f°C",
		d.heaterPower, d.heatLossCoefficient, d.heatTransferCoefficient, d.kilnThermalMass, d.materialThermalMass,
		d.steamPower, d.steamCutoffTemp)

	return nil
}

func (d *DifferentialSimulation) ValidateConfig(config map[string]interface{}) error {
	required := []string{
		keyHeaterPower, keyHeatLossCoefficient, keyHeatTransferCoefficient,
		keyKilnThermalMass, keyMaterialThermalMass, keySteamPower, keySteamCutoffTemp,
	}
	for _, key := range required {
		if _, exists := config[key]; !exists {
			return fmt.Errorf("required configuration parameter missing: %s", key)
		}
		if _, ok := config[key].(float64); !ok {
			return fmt.Errorf("configuration parameter %s must be a number", key)
		}
	}

	// Validate positive values
	if heaterPower := config[keyHeaterPower].(float64); heaterPower <= 0 {
		return errors.New("heater_power must be positive")
	}
	if heatLoss := config[keyHeatLossCoefficient].(float64); heatLoss <= 0 {
		return errors.New("heat_loss_coefficient must be positive")
	}
	if heatTransfer := config[keyHeatTransferCoefficient].(float64); heatTransfer <= 0 {
		return errors.New("heat_transfer_coefficient must be positive")
	}
	if kilnMass := config[keyKilnThermalMass].(float64); kilnMass <= 0 {
		return errors.New("kiln_thermal_mass must be positive")
	}
	if materialMass := config[keyMaterialThermalMass].(float64); materialMass <= 0 {
		return errors.New("material_thermal_mass must be positive")
	}
	// Zero is meaningful here: it is how a configuration says steam is inert,
	// which is what the fitted config does until a real run constrains it.
	if steamPower := config[keySteamPower].(float64); steamPower < 0 {
		return errors.New("steam_power must not be negative")
	}
	if steamCutoff := config[keySteamCutoffTemp].(float64); steamCutoff <= 0 {
		return errors.New("steam_cutoff_temp must be positive")
	}

	return nil
}

func (d *DifferentialSimulation) Tick(state *SimulationState) {
	oldKilnTemp := state.KilnTemp
	oldMaterialTemp := state.MaterialTemp

	// Calculate energy inputs/outputs for kiln
	var heaterEnergy float32
	if state.HeaterIsOn {
		heaterEnergy = d.heaterPower
	}

	// Steam only heats a kiln below the cutoff. Above it the steam has to be
	// raised to kiln temperature itself, so it stops contributing - the same
	// reasoning the control unit's steam ceiling rests on, which is where the
	// cutoff temperature comes from.
	var steamEnergy float32
	if state.SteamIsOn && state.KilnTemp < d.steamCutoffTemp {
		steamEnergy = d.steamPower
	}

	// Heat loss to environment (Newton's Law of Cooling)
	// Higher temperature difference → faster heat loss
	kilnHeatLoss := d.heatLossCoefficient * (state.KilnTemp - state.EnvironmentTemp)

	// Heat transfer between kiln and material (proportional to temp difference)
	// Positive = heat flows from kiln to material
	// Negative = heat flows from material to kiln
	kilnMaterialTransfer := d.heatTransferCoefficient * (state.KilnTemp - state.MaterialTemp)

	// Net energy change for kiln
	kilnNetEnergy := heaterEnergy + steamEnergy - kilnHeatLoss - kilnMaterialTransfer

	// Temperature change = energy / thermal mass
	kilnTempChange := kilnNetEnergy / d.kilnThermalMass
	state.KilnTemp = max(state.EnvironmentTemp, state.KilnTemp+kilnTempChange)

	// The material sits inside the kiln, so the kiln is its only thermal
	// contact: whatever the transfer moves is the whole of its energy change.
	// It has no path to the environment of its own.
	materialTempChange := kilnMaterialTransfer / d.materialThermalMass
	state.MaterialTemp = max(state.EnvironmentTemp, state.MaterialTemp+materialTempChange)

	// Log energy flows and temperature changes
	if steamEnergy > 0 {
		log.Debug("Simulation[differential]: Steam contributing %.3f (kiln %.1f°C below cutoff %.0f°C)",
			steamEnergy, oldKilnTemp, d.steamCutoffTemp)
	}
	if state.HeaterIsOn {
		log.Debug("Simulation[differential]: Heater ON - energy=%.3f, steam=%.3f, kiln_loss=%.3f, transfer=%.3f → kiln: %.1f°C → %.1f°C (Δ%.2f°C)",
			heaterEnergy, steamEnergy, kilnHeatLoss, kilnMaterialTransfer, oldKilnTemp, state.KilnTemp, kilnTempChange)
	} else if state.KilnTemp != oldKilnTemp {
		log.Debug("Simulation[differential]: Heater OFF - kiln_loss=%.3f, transfer=%.3f → kiln: %.1f°C → %.1f°C (Δ%.2f°C)",
			kilnHeatLoss, kilnMaterialTransfer, oldKilnTemp, state.KilnTemp, kilnTempChange)
	}

	if state.MaterialTemp != oldMaterialTemp {
		log.Debug("Simulation[differential]: Material - received=%.3f → material: %.1f°C → %.1f°C (Δ%.2f°C)",
			kilnMaterialTransfer, oldMaterialTemp, state.MaterialTemp, materialTempChange)
	}
}
