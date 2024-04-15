# Halko simulator

This is just a simple simulator that can be used to test the frontend side
and executor without a real risk of setting the world on fire. 

This in no way intends to simulate the actual physics involved.

It provides emulated interfaces for the sensors and controlling the power

## The parts

### main

Setup the whole thing, start the "simulation" and starts the gin-gonic
server and spawns a ticker into the background that basically advances the 
clock for each power (heater, fan, humidifier)

### oven

On each tick updates the oven temperature and also sets temperature
for the material (via calling its AmbientTemperature)

oven will never go over the maximum or below the minimum temperature 
regardless the power.

### wood 

Again this does not in any way simulate how the wood would heat up in real life.

Implements the wood temperature sensor.

On each oven temperature tick the wood will heat up or cool down

### fan

Does not affect "simulation" is there only to give something for the power sensor and
control to act on

### humidifier

Does not affect "simulation" is there only to give something for the power sensor and
control to act on
