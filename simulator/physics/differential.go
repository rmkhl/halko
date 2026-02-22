package physics

import (
	"errors"
	"fmt"

	"github.com/rmkhl/halko/types/log"
)

// DifferentialSimulation implements temperature-differential physics where heat transfer
// rates are proportional to temperature differences (Newton's Law of Cooling)
type DifferentialSimulation struct {
	heaterPower            float32 // Energy per tick when heater is on
	heatLossCoefficient    float32 // Proportional heat loss to environment
	heatTransferCoefficient float32 // Heat transfer rate between oven and material
	ovenThermalMass        float32 // Heat capacity of oven (energy needed to raise 1°C)
	materialThermalMass    float32 // Heat capacity of material
}

func (d *DifferentialSimulation) Name() string {
	return "differential"
}

func (d *DifferentialSimulation) Initialize(config map[string]interface{}) error {
	// Extract and validate configuration
	heaterPower, ok := config["heater_power"].(float64)
	if !ok {
		return errors.New("heater_power must be specified as a number")
	}
	d.heaterPower = float32(heaterPower)

	heatLossCoeff, ok := config["heat_loss_coefficient"].(float64)
	if !ok {
		return errors.New("heat_loss_coefficient must be specified as a number")
	}
	d.heatLossCoefficient = float32(heatLossCoeff)

	heatTransferCoeff, ok := config["heat_transfer_coefficient"].(float64)
	if !ok {
		return errors.New("heat_transfer_coefficient must be specified as a number")
	}
	d.heatTransferCoefficient = float32(heatTransferCoeff)

	ovenMass, ok := config["oven_thermal_mass"].(float64)
	if !ok {
		return errors.New("oven_thermal_mass must be specified as a number")
	}
	d.ovenThermalMass = float32(ovenMass)

	materialMass, ok := config["material_thermal_mass"].(float64)
	if !ok {
		return errors.New("material_thermal_mass must be specified as a number")
	}
	d.materialThermalMass = float32(materialMass)

	log.Info("Differential simulation initialized: heater_power=%.2f, heat_loss=%.3f, heat_transfer=%.3f, oven_mass=%.1f, material_mass=%.1f",
		d.heaterPower, d.heatLossCoefficient, d.heatTransferCoefficient, d.ovenThermalMass, d.materialThermalMass)

	return nil
}

func (d *DifferentialSimulation) ValidateConfig(config map[string]interface{}) error {
	required := []string{"heater_power", "heat_loss_coefficient", "heat_transfer_coefficient", "oven_thermal_mass", "material_thermal_mass"}
	for _, key := range required {
		if _, exists := config[key]; !exists {
			return fmt.Errorf("required configuration parameter missing: %s", key)
		}
		if _, ok := config[key].(float64); !ok {
			return fmt.Errorf("configuration parameter %s must be a number", key)
		}
	}

	// Validate positive values
	if heaterPower := config["heater_power"].(float64); heaterPower <= 0 {
		return errors.New("heater_power must be positive")
	}
	if heatLoss := config["heat_loss_coefficient"].(float64); heatLoss <= 0 {
		return errors.New("heat_loss_coefficient must be positive")
	}
	if heatTransfer := config["heat_transfer_coefficient"].(float64); heatTransfer <= 0 {
		return errors.New("heat_transfer_coefficient must be positive")
	}
	if ovenMass := config["oven_thermal_mass"].(float64); ovenMass <= 0 {
		return errors.New("oven_thermal_mass must be positive")
	}
	if materialMass := config["material_thermal_mass"].(float64); materialMass <= 0 {
		return errors.New("material_thermal_mass must be positive")
	}

	return nil
}

func (d *DifferentialSimulation) Tick(state *SimulationState) {
	oldOvenTemp := state.OvenTemp
	oldMaterialTemp := state.MaterialTemp

	// Calculate energy inputs/outputs for oven
	var heaterEnergy float32
	if state.HeaterIsOn {
		heaterEnergy = d.heaterPower
	}

	// Heat loss to environment (Newton's Law of Cooling)
	// Higher temperature difference → faster heat loss
	ovenHeatLoss := d.heatLossCoefficient * (state.OvenTemp - state.EnvironmentTemp)

	// Heat transfer between oven and material (proportional to temp difference)
	// Positive = heat flows from oven to material
	// Negative = heat flows from material to oven
	ovenMaterialTransfer := d.heatTransferCoefficient * (state.OvenTemp - state.MaterialTemp)

	// Net energy change for oven
	ovenNetEnergy := heaterEnergy - ovenHeatLoss - ovenMaterialTransfer

	// Temperature change = energy / thermal mass
	ovenTempChange := ovenNetEnergy / d.ovenThermalMass
	state.OvenTemp = max(state.EnvironmentTemp, state.OvenTemp+ovenTempChange)

	// Material receives energy from oven and loses to environment
	materialHeatLoss := d.heatLossCoefficient * (state.MaterialTemp - state.EnvironmentTemp)
	materialNetEnergy := ovenMaterialTransfer - materialHeatLoss
	materialTempChange := materialNetEnergy / d.materialThermalMass
	state.MaterialTemp = max(state.EnvironmentTemp, state.MaterialTemp+materialTempChange)

	// Log energy flows and temperature changes
	if state.HeaterIsOn {
		log.Debug("Simulation[differential]: Heater ON - energy=%.3f, oven_loss=%.3f, transfer=%.3f → oven: %.1f°C → %.1f°C (Δ%.2f°C)",
			heaterEnergy, ovenHeatLoss, ovenMaterialTransfer, oldOvenTemp, state.OvenTemp, ovenTempChange)
	} else if state.OvenTemp != oldOvenTemp {
		log.Debug("Simulation[differential]: Heater OFF - oven_loss=%.3f, transfer=%.3f → oven: %.1f°C → %.1f°C (Δ%.2f°C)",
			ovenHeatLoss, ovenMaterialTransfer, oldOvenTemp, state.OvenTemp, ovenTempChange)
	}

	if state.MaterialTemp != oldMaterialTemp {
		log.Debug("Simulation[differential]: Material - received=%.3f, mat_loss=%.3f → material: %.1f°C → %.1f°C (Δ%.2f°C)",
			ovenMaterialTransfer, materialHeatLoss, oldMaterialTemp, state.MaterialTemp, materialTempChange)
	}
}
