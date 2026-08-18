package main

import (
	"strings"
	"testing"

	"github.com/rmkhl/halko/types"
)

// describeConfig exists so `halkoctl config` can show what a config actually
// resolves to, not just that it parsed. Every value the control unit falls back
// to should appear, since those are the ones nothing else surfaces.
func TestDescribeConfigShowsResolvedDefaults(t *testing.T) {
	config, err := types.LoadConfig("../templates/halko.cfg")
	if err != nil {
		t.Fatalf("load template config: %v", err)
	}

	out := describeConfig(config)

	for _, want := range []string{
		"heating", "acclimate",
		"fan_power", "steam_power", "preheat_fan_power",
		"max_target_temperature", "steam_ceiling",
		"sensor_timeout", "execution_log_interval",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("description does not mention %q:\n%s", want, out)
		}
	}
}

func TestDescribeConfigShowsDeltaBandValues(t *testing.T) {
	config, err := types.LoadConfig("../templates/halko.cfg")
	if err != nil {
		t.Fatalf("load template config: %v", err)
	}

	out := describeConfig(config)
	acclimate := config.ControlUnitConfig.Defaults.Deltas[types.StepTypeAcclimate]
	if !strings.Contains(out, "-1") || !strings.Contains(out, "3") {
		t.Errorf("acclimate band %v/%v not shown:\n%s", acclimate.MinDelta, acclimate.MaxDelta, out)
	}
}
