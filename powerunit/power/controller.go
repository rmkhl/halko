package power

import (
	"context"
	"sync"
	"time"

	"github.com/rmkhl/halko/powerunit/shelly"
	"github.com/rmkhl/halko/types/log"
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
		isIdle       bool           // Tracks whether we're currently in idle state
		shelly       *shelly.Shelly // Shelly controller for device communication
	}
)

func New(maxIdleTimeS int, cycleLengthS int, shellyCtrl *shelly.Shelly) *Controller {
	log.Trace("Creating new power controller")
	ctx, cancel := context.WithCancel(context.Background())

	cycleLength := time.Duration(cycleLengthS) * time.Second
	maxIdleTime := time.Duration(maxIdleTimeS) * time.Second
	tickDuration := cycleLength / 100

	log.Debug("Power controller config: cycle=%v, maxIdle=%v, tick=%v",
		cycleLength, maxIdleTime, tickDuration)

	var powerStates [shelly.NumberOfDevices]*powerTracker
	for i := range shelly.NumberOfDevices {
		powerStates[i] = &powerTracker{percentage: 0, currentState: shelly.On}
	}
	log.Trace("Initialized %d power trackers", shelly.NumberOfDevices)

	controller := &Controller{
		powerStates:  powerStates,
		ctx:          ctx,
		cancel:       cancel,
		cycleLength:  cycleLength,
		maxIdleTime:  maxIdleTime,
		tickDuration: tickDuration,
		tickCount:    0,
		lastCommand:  time.Now(),
		isIdle:       true,
		shelly:       shellyCtrl,
	}
	log.Debug("Power controller created successfully")
	return controller
}

func (c *Controller) Start() error {
	log.Info("Starting power controller with tick duration %v", c.tickDuration)
	defer c.cancel()

	// Turn off all devices on startup to ensure clean initial state
	// Use retry mechanism to handle race condition shelly startup
	log.Info("Turning off all devices on startup")
	maxRetries := 5
	retryDelay := 500 * time.Millisecond

	for i := range shelly.NumberOfDevices {
		success := false
		for attempt := 0; attempt < maxRetries; attempt++ {
			if _, err := c.shelly.SetState(shelly.Off, i); err != nil {
				if attempt < maxRetries-1 {
					log.Debug("Failed to turn off device %d (attempt %d/%d), retrying in %v: %v",
						i, attempt+1, maxRetries, retryDelay, err)
					time.Sleep(retryDelay)
					retryDelay *= 2 // Exponential backoff
				} else {
					log.Error("Error turning off device %d after %d attempts: %v", i, maxRetries, err)
				}
			} else {
				success = true
				if attempt > 0 {
					log.Debug("Successfully turned off device %d on attempt %d", i, attempt+1)
				}
				break
			}
		}
		if success {
			retryDelay = 500 * time.Millisecond // Reset delay for next device
		}
	}

	ticker := time.NewTicker(c.tickDuration)
	defer ticker.Stop()
	log.Debug("Created ticker for power control loop")

	c.mu.Lock()
	for i := range shelly.NumberOfDevices {
		c.powerStates[i].percentage = 0
		c.powerStates[i].currentState = shelly.Off
	}
	c.lastCommand = time.Now()
	c.mu.Unlock()
	log.Debug("Initialized power states to 0%% and Off")

	for {
		select {
		case <-c.ctx.Done():
			log.Info("Power controller stopped")
			return nil
		case <-ticker.C:
			log.Trace("Processing tick %d", c.tickCount)
			if err := c.processTick(); err != nil {
				log.Error("Error processing tick: %v", err)
			}
		}
	}
}

func (c *Controller) processTick() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.tickCount == 0 {
		log.Trace("Beginning of cycle (tick 0) - checking devices to turn on")
		for id := range shelly.NumberOfDevices {
			tracker := c.powerStates[id]
			if tracker.percentage > 0 && tracker.currentState == shelly.Off {
				log.Debug("Turning on device %d (percentage: %d%%)", id, tracker.percentage)
				if _, err := c.shelly.SetState(shelly.On, id); err != nil {
					log.Error("Error turning on %d: %v", id, err)
					continue
				}
				tracker.currentState = shelly.On
				c.lastCommand = time.Now()
			}
		}
	}

	for id := range shelly.NumberOfDevices {
		tracker := c.powerStates[id]
		if c.tickCount >= int(tracker.percentage) && tracker.currentState == shelly.On && tracker.percentage > 0 {
			log.Debug("Turning off device %d at tick %d (percentage: %d%%)", id, c.tickCount, tracker.percentage)
			if _, err := c.shelly.SetState(shelly.Off, id); err != nil {
				log.Error("Error turning off %d at tick %d: %v", id, c.tickCount, err)
				continue
			}
			tracker.currentState = shelly.Off
			c.lastCommand = time.Now()
		}
	}

	// If more than max idle time has passed since the last command, set all percentages to 0
	timeSinceLastCommand := time.Since(c.lastCommand)
	if timeSinceLastCommand > c.maxIdleTime {
		if !c.isIdle {
			log.Warning("Idle for %v (max: %v), resetting all percentages to 0", timeSinceLastCommand, c.maxIdleTime)
			c.isIdle = true
		}
		for id := range shelly.NumberOfDevices {
			if c.powerStates[id].percentage > 0 {
				log.Debug("Resetting device %d percentage from %d%% to 0%%", id, c.powerStates[id].percentage)
			}
			c.powerStates[id].percentage = 0
		}
		c.lastCommand = time.Now()
	}

	c.tickCount = (c.tickCount + 1) % 100

	return nil
}

// Stop halts the power controller and turns off all devices
func (c *Controller) Stop() {
	log.Info("Stopping power controller and shutting down all devices")
	c.cancel()

	if err := c.shelly.Shutdown(); err != nil {
		log.Error("Error shutting down devices: %v", err)
	} else {
		log.Debug("All devices shut down successfully")
	}
}

// GetAllPercentages returns the current power percentages of all devices
func (c *Controller) GetAllPercentages() [shelly.NumberOfDevices]uint8 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	c.lastCommand = time.Now()

	var percentages [shelly.NumberOfDevices]uint8
	for id := range shelly.NumberOfDevices {
		percentages[id] = c.powerStates[id].percentage
	}

	log.Trace("Retrieved percentages: %v", percentages)
	return percentages
}

// SetAllPercentages updates all power percentages at once for the next cycle
func (c *Controller) SetAllPercentages(percentages [shelly.NumberOfDevices]uint8) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.lastCommand = time.Now()
	c.isIdle = false // Reset idle state when receiving new command

	log.Debug("Setting new percentages: %v", percentages)
	for id := range shelly.NumberOfDevices {
		oldPercentage := c.powerStates[id].percentage
		c.powerStates[id].percentage = percentages[id]
		if oldPercentage != percentages[id] {
			log.Debug("Device %d percentage changed: %d%% -> %d%%", id, oldPercentage, percentages[id])
		}
	}
}

// IsIdle returns whether the controller is currently in idle state
func (c *Controller) IsIdle() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.isIdle
}
