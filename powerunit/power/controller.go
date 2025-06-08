package power

import (
	"context"
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
		powerStates  [shelly.NumberOfDevices]*powerTracker // Fixed array for power states with IDs 0, 1, and 2
		mu           sync.RWMutex
		ctx          context.Context
		cancel       context.CancelFunc
		cycleLength  time.Duration  // Total duration of a power cycle
		tickDuration time.Duration  // Duration of a single tick (1% of cycle)
		maxIdleTime  time.Duration  // Maximum idle time before resetting percentages
		tickCount    int            // Current tick count (0-99)
		lastCommand  time.Time      // Timestamp when the last command was applied
		shelly       *shelly.Shelly // Shelly controller for device communication
	}
)

func New(maxIdleTimeS int, cycleLengthS int, shellyCtrl *shelly.Shelly) *Controller {
	ctx, cancel := context.WithCancel(context.Background())

	cycleLength := time.Duration(cycleLengthS) * time.Second
	maxIdleTime := time.Duration(maxIdleTimeS) * time.Second
	tickDuration := cycleLength / 100

	var powerStates [shelly.NumberOfDevices]*powerTracker
	for i := range shelly.NumberOfDevices {
		powerStates[i] = &powerTracker{percentage: 0, currentState: shelly.On}
	}

	return &Controller{
		powerStates:  powerStates,
		ctx:          ctx,
		cancel:       cancel,
		cycleLength:  cycleLength,
		maxIdleTime:  maxIdleTime,
		tickDuration: tickDuration,
		tickCount:    0,
		lastCommand:  time.Now(),
		shelly:       shellyCtrl,
	}
}

func (c *Controller) Start() error {
	defer c.cancel()

	ticker := time.NewTicker(c.tickDuration)
	defer ticker.Stop()

	// Initialize all power states to 0% and assume they are On
	c.mu.Lock()
	for i := range shelly.NumberOfDevices {
		c.powerStates[i].percentage = 0
		c.powerStates[i].currentState = shelly.On
	}
	c.lastCommand = time.Now()
	c.mu.Unlock()

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

func (c *Controller) processTick() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// At the start of a new cycle
	if c.tickCount == 0 {
		// Turn on units that are OFF and have non-zero percentage
		for id := range shelly.NumberOfDevices {
			tracker := c.powerStates[id]
			if tracker.percentage > 0 && tracker.currentState == shelly.Off {
				if _, err := c.shelly.SetState(shelly.On, id); err != nil {
					log.Printf("Error turning on %d: %v", id, err)
					continue
				}
				tracker.currentState = shelly.On
				c.lastCommand = time.Now()
			}
		}
	}

	// During the cycle, check if any powers need to be turned off
	for id := range shelly.NumberOfDevices {
		tracker := c.powerStates[id]
		if c.tickCount >= int(tracker.percentage) && tracker.currentState == shelly.On {
			if _, err := c.shelly.SetState(shelly.Off, id); err != nil {
				log.Printf("Error turning off %d at tick %d: %v", id, c.tickCount, err)
				continue
			}
			tracker.currentState = shelly.Off
			c.lastCommand = time.Now()
		}
	}

	// If more than five minutes have passed since the last command, set all percentages to 0
	if time.Since(c.lastCommand) > c.maxIdleTime {
		log.Printf("Idle for %s, setting all percentages to 0", c.maxIdleTime.String())
		for id := range shelly.NumberOfDevices {
			c.powerStates[id].percentage = 0
		}
		// Update lastCommand timestamp to avoid repeated shutdowns
		c.lastCommand = time.Now()
	}

	c.tickCount = (c.tickCount + 1) % 100

	return nil
}

// Stop halts the power controller and turns off all devices
func (c *Controller) Stop() {
	c.cancel()

	if err := c.shelly.Shutdown(); err != nil {
		log.Printf("Error shutting down devices: %v", err)
	}
}

// GetAllPercentages returns the current power percentages of all devices
func (c *Controller) GetAllPercentages() [shelly.NumberOfDevices]uint8 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	// Update lastCommand timestamp
	c.lastCommand = time.Now()

	var percentages [shelly.NumberOfDevices]uint8
	for id := range shelly.NumberOfDevices {
		percentages[id] = c.powerStates[id].percentage
	}

	return percentages
}

// SetAllPercentages updates all power percentages at once for the next cycle
func (c *Controller) SetAllPercentages(percentages [shelly.NumberOfDevices]uint8) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Update lastCommand timestamp
	c.lastCommand = time.Now()

	for id := range shelly.NumberOfDevices {
		c.powerStates[id].percentage = percentages[id]
	}
}
