package power

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/rmkhl/halko/powerunit/shelly"
)

type (
	powerTracker struct {
		currentState shelly.PowerState // Current power state (on/off)
		percentage   uint8             // 0-100 percentage of cycle to be powered on
	}

	Controller struct {
		powerStates  map[shelly.ID]*powerTracker
		mu           sync.RWMutex // mutex for thread-safe access to the powerStates map
		ctx          context.Context
		cancel       context.CancelFunc
		cycleLength  time.Duration  // Total duration of a power cycle
		tickDuration time.Duration  // Duration of a single tick (1% of cycle)
		tickCount    int            // Current tick count (0-99)
		shelly       *shelly.Shelly // Shelly controller for device communication
	}
)

// New creates a new power controller with the specified cycle length in milliseconds
func New(cycleLengthMs int, shellyCtrl *shelly.Shelly) *Controller {
	if cycleLengthMs <= 0 {
		cycleLengthMs = 60000 // Default to 1 minute if not specified
	}

	ctx, cancel := context.WithCancel(context.Background())

	cycleLength := time.Duration(cycleLengthMs) * time.Millisecond
	tickDuration := cycleLength / 100 // Divide cycle into 100 ticks

	powerStates := make(map[shelly.ID]*powerTracker)
	powerStates[shelly.Fan] = &powerTracker{percentage: 0, currentState: shelly.Off}
	powerStates[shelly.Heater] = &powerTracker{percentage: 0, currentState: shelly.Off}
	powerStates[shelly.Humidifier] = &powerTracker{percentage: 0, currentState: shelly.Off}

	return &Controller{
		powerStates:  powerStates,
		ctx:          ctx,
		cancel:       cancel,
		cycleLength:  cycleLength,
		tickDuration: tickDuration,
		tickCount:    0,
		shelly:       shellyCtrl,
	}
}

// Start begins the power cycling process for all devices in a single goroutine
func (c *Controller) Start() error {
	defer c.cancel()

	ticker := time.NewTicker(c.tickDuration)
	defer ticker.Stop()

	// Initialize power states at startup
	if err := c.updateAllPowerStates(); err != nil {
		return fmt.Errorf("failed to initialize power states: %w", err)
	}

	// Run the power cycle loop
	for {
		// Check if context is done
		select {
		case <-c.ctx.Done():
			log.Println("Power controller stopped")
			return nil
		case <-ticker.C:
			// Process current tick
			if err := c.processTick(); err != nil {
				log.Printf("Error processing tick: %v", err)
			}
		}
	}
}

// processTick handles power state updates for the current tick
func (c *Controller) processTick() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// At the start of a new cycle
	if c.tickCount == 0 {
		// Reset power states based on percentages
		for id, tracker := range c.powerStates {
			if tracker.percentage == 0 && tracker.currentState == shelly.On {
				// Turn off if percentage is 0
				if _, err := c.shelly.SetState(shelly.Off, id); err != nil {
					log.Printf("Error turning off %s: %v", id, err)
					continue
				}
				tracker.currentState = shelly.Off
			} else if tracker.percentage > 0 && tracker.currentState == shelly.Off {
				// Turn on if percentage > 0
				if _, err := c.shelly.SetState(shelly.On, id); err != nil {
					log.Printf("Error turning on %s: %v", id, err)
					continue
				}
				tracker.currentState = shelly.On
			}
		}
	} else {
		// During the cycle, check if any powers need to be turned off
		for id, tracker := range c.powerStates {
			if tracker.percentage > 0 && tracker.currentState == shelly.On {
				// If the tick count exceeds the percentage, turn off the power
				if c.tickCount >= int(tracker.percentage) {
					if _, err := c.shelly.SetState(shelly.Off, id); err != nil {
						log.Printf("Error turning off %s at tick %d: %v", id, c.tickCount, err)
						continue
					}
					tracker.currentState = shelly.Off
				}
			}
		}
	}

	// Increment tick count and reset at the end of a cycle
	c.tickCount = (c.tickCount + 1) % 100

	return nil
}

// updateAllPowerStates fetches the current state of all devices from the shelly interface
func (c *Controller) updateAllPowerStates() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	for id, tracker := range c.powerStates {
		state, err := c.shelly.GetState(id)
		if err != nil {
			return fmt.Errorf("failed to get state for %s: %w", id, err)
		}
		tracker.currentState = state
	}

	return nil
}

// Stop halts the power controller and turns off all devices
func (c *Controller) Stop() {
	// Signal all goroutines to stop
	c.cancel()

	// Ensure all powers are turned off
	for id := range c.powerStates {
		if _, err := c.shelly.SetState(shelly.Off, id); err != nil {
			log.Printf("Error turning off %s during shutdown: %v", id, err)
		}
	}
}

// GetAllCycles returns the current power cycle percentages of all devices
func (c *Controller) GetAllCycles() (map[shelly.ID]uint8, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	cycles := make(map[shelly.ID]uint8)
	for id, tracker := range c.powerStates {
		cycles[id] = tracker.percentage
	}

	return cycles, nil
}

// SetAllCycles updates all power cycle percentages at once
func (c *Controller) SetAllCycles(cycles map[shelly.ID]uint8) error {
	// Validate all percentages first
	for id, percentage := range cycles {
		if percentage > 100 {
			return fmt.Errorf("percentage for %s must be between 0 and 100", id.String())
		}
		if _, exists := c.powerStates[id]; !exists {
			return fmt.Errorf("unknown power ID: %d", id)
		}
	}

	// Apply all changes
	c.mu.Lock()
	defer c.mu.Unlock()

	for id, percentage := range cycles {
		tracker := c.powerStates[id]
		tracker.percentage = percentage

		// If we're currently in the off portion of the cycle but the new percentage
		// would make it on, turn it on immediately
		if tracker.currentState == shelly.Off && percentage > 0 && c.tickCount < int(percentage) {
			if _, err := c.shelly.SetState(shelly.On, id); err != nil {
				return fmt.Errorf("error turning on %s after setting cycle: %w", id, err)
			}
			tracker.currentState = shelly.On
		}

		// If we're currently in the on portion of the cycle but the new percentage
		// would make it off, turn it off immediately
		if tracker.currentState == shelly.On && (percentage == 0 || c.tickCount >= int(percentage)) {
			if _, err := c.shelly.SetState(shelly.Off, id); err != nil {
				return fmt.Errorf("error turning off %s after setting cycle: %w", id, err)
			}
			tracker.currentState = shelly.Off
		}
	}

	return nil
}
