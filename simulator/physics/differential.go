package physics

import (
	"errors"
	"fmt"

	"github.com/rmkhl/halko/types/log"
)

// DifferentialSimulation implements temperature-differential physics where heat transfer
// rates are proportional to temperature differences (Newton's Law of Cooling)
type DifferentialSimulation struct {
	heaterPower             float32 // Energy per tick when heater is on
	heatLossCoefficient     float32 // Proportional heat loss to environment
	heatTransferCoefficient float32 // Heat transfer rate between kiln and material
	kilnThermalMass         float32 // Heat capacity of kiln (energy needed to raise 1°C)
	materialThermalMass     float32 // Heat capacity of material
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

	kilnMass, ok := config["kiln_thermal_mass"].(float64)
	if !ok {
		return errors.New("kiln_thermal_mass must be specified as a number")
	}
	d.kilnThermalMass = float32(kilnMass)

	materialMass, ok := config["material_thermal_mass"].(float64)
	if !ok {
		return errors.New("material_thermal_mass must be specified as a number")
	}
	d.materialThermalMass = float32(materialMass)

	log.Info("Differential simulation initialized: heater_power=%.2f, heat_loss=%.3f, heat_transfer=%.3f, kiln_mass=%.1f, material_mass=%.1f",
		d.heaterPower, d.heatLossCoefficient, d.heatTransferCoefficient, d.kilnThermalMass, d.materialThermalMass)

	return nil
}

func (d *DifferentialSimulation) ValidateConfig(config map[string]interface{}) error {
	required := []string{"heater_power", "heat_loss_coefficient", "heat_transfer_coefficient", "kiln_thermal_mass", "material_thermal_mass"}
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
	if kilnMass := config["kiln_thermal_mass"].(float64); kilnMass <= 0 {
		return errors.New("kiln_thermal_mass must be positive")
	}
	if materialMass := config["material_thermal_mass"].(float64); materialMass <= 0 {
		return errors.New("material_thermal_mass must be positive")
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

	// Heat loss to environment (Newton's Law of Cooling)
	// Higher temperature difference → faster heat loss
	kilnHeatLoss := d.heatLossCoefficient * (state.KilnTemp - state.EnvironmentTemp)

	// Heat transfer between kiln and material (proportional to temp difference)
	// Positive = heat flows from kiln to material
	// Negative = heat flows from material to kiln
	kilnMaterialTransfer := d.heatTransferCoefficient * (state.KilnTemp - state.MaterialTemp)

	// Net energy change for kiln
	kilnNetEnergy := heaterEnergy - kilnHeatLoss - kilnMaterialTransfer

	// Temperature change = energy / thermal mass
	kilnTempChange := kilnNetEnergy / d.kilnThermalMass
	state.KilnTemp = max(state.EnvironmentTemp, state.KilnTemp+kilnTempChange)

	// Material receives energy from kiln and loses to environment
	materialHeatLoss := d.heatLossCoefficient * (state.MaterialTemp - state.EnvironmentTemp)
	materialNetEnergy := kilnMaterialTransfer - materialHeatLoss
	materialTempChange := materialNetEnergy / d.materialThermalMass
	state.MaterialTemp = max(state.EnvironmentTemp, state.MaterialTemp+materialTempChange)

	// Log energy flows and temperature changes
	if state.HeaterIsOn {
		log.Debug("Simulation[differential]: Heater ON - energy=%.3f, kiln_loss=%.3f, transfer=%.3f → kiln: %.1f°C → %.1f°C (Δ%.2f°C)",
			heaterEnergy, kilnHeatLoss, kilnMaterialTransfer, oldKilnTemp, state.KilnTemp, kilnTempChange)
	} else if state.KilnTemp != oldKilnTemp {
		log.Debug("Simulation[differential]: Heater OFF - kiln_loss=%.3f, transfer=%.3f → kiln: %.1f°C → %.1f°C (Δ%.2f°C)",
			kilnHeatLoss, kilnMaterialTransfer, oldKilnTemp, state.KilnTemp, kilnTempChange)
	}

	if state.MaterialTemp != oldMaterialTemp {
		log.Debug("Simulation[differential]: Material - received=%.3f, mat_loss=%.3f → material: %.1f°C → %.1f°C (Δ%.2f°C)",
			kilnMaterialTransfer, materialHeatLoss, oldMaterialTemp, state.MaterialTemp, materialTempChange)
	}
}
