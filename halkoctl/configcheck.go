package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/rmkhl/halko/types"
)

// handleConfigCommand loads the configuration and reports what it resolves to.
// It is dispatched before the global config load, so that a config the services
// would refuse to start on can still be diagnosed here rather than producing
// the same bare error every other command gives.
func handleConfigCommand(configPath string) {
	config, err := types.LoadConfig(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Configuration is not usable:\n  %v\n\n", err)
		fmt.Fprintln(os.Stderr, "Every service refuses to start on this, so fix it before deploying.")
		os.Exit(exitError)
	}

	fmt.Println("Configuration is valid.")
	fmt.Println()
	fmt.Print(describeConfig(config))
}

// describeConfig renders the values the control unit falls back to. These are
// the ones nothing else surfaces: a program that names no heater, fan or steam
// gets them from here, and so do the limits it is validated against.
func describeConfig(config *types.HalkoConfig) string {
	var b strings.Builder
	d := config.ControlUnitConfig.Defaults

	fmt.Fprintf(&b, "Control unit\n")
	fmt.Fprintf(&b, "  base_path                 %s\n", config.ControlUnitConfig.BasePath)
	fmt.Fprintf(&b, "  tick_length               %s\n", config.ControlUnitConfig.TickLength)
	fmt.Fprintf(&b, "\nDefaults applied to a program that names none\n")
	for _, stepType := range []types.StepType{types.StepTypeHeating, types.StepTypeAcclimate} {
		band := d.Deltas[stepType]
		fmt.Fprintf(&b, "  deltas.%-19s min_delta %g, max_delta %g\n", stepType, band.MinDelta, band.MaxDelta)
	}
	fmt.Fprintf(&b, "  fan_power                 %d%%\n", *d.FanPower)
	fmt.Fprintf(&b, "  steam_power               %d%%\n", *d.SteamPower)
	fmt.Fprintf(&b, "  equalize\n")
	fmt.Fprintf(&b, "    delta                   %.1f°C\n", *d.Equalize.Delta)
	fmt.Fprintf(&b, "    steam_prewarm           %t\n", *d.Equalize.SteamPrewarm)
	fmt.Fprintf(&b, "    steam_prewarm_timeout   %s\n", d.Equalize.SteamPrewarmTimeout)
	fmt.Fprintf(&b, "\nLimits and timings\n")
	fmt.Fprintf(&b, "  max_target_temperature    %d°C\n", *d.MaxTargetTemperature)
	fmt.Fprintf(&b, "  steam_ceiling             %d°C\n", *d.SteamCeiling)
	fmt.Fprintf(&b, "  sensor_timeout            %s\n", d.SensorTimeout)
	fmt.Fprintf(&b, "  execution_log_interval    %s\n", d.ExecutionLogInterval)

	return b.String()
}

func showConfigHelp() {
	fmt.Println("Usage: halkoctl config")
	fmt.Println()
	fmt.Println("Check the configuration file and show what it resolves to.")
	fmt.Println("Exits non-zero if the configuration is one the services would")
	fmt.Println("refuse to start on.")
	fmt.Println()
	fmt.Println("Use -c/--config to check a file other than the default.")
}
