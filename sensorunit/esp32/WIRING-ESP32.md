# ESP32 Wiring Diagram - MAX31855 + OLED Display

## Overview

This document describes the wiring for 3x MAX31855 thermocouple amplifier modules and an I2C OLED display connected to an ESP32 DevKit board using **hardware SPI** and **I2C**.

## Hardware Required

- 1x ESP32 DevKit V1 (30-pin, Micro-USB)
- 3x MAX31855 thermocouple amplifier breakout boards
- 3x K-type thermocouples
- 1x I2C OLED display (SSD1306 or SH1106, 0.96" or 1.3")
- Breadboard or prototyping shield
- Jumper wires
- Micro-USB cable (USB-A to Micro-USB for Raspberry Pi connection)

## Recommended OLED Display

### **Best Choice: 0.96" I2C OLED (SSD1306) - 128x64**

**Specifications:**

- Driver IC: SSD1306
- Resolution: 128x64 pixels
- Interface: I2C (4 pins: VCC, GND, SDA, SCL)
- Voltage: 3.3V-5V compatible
- Color: White, Blue, or Yellow/Blue (dual color)
- Viewing angle: >160°
- Cost: $3-5

**Why this one:**

- ✅ Perfect size for showing 3 temperatures + status
- ✅ Very readable in various lighting conditions
- ✅ Low power consumption (~15-20mA)
- ✅ Excellent library support (Adafruit SSD1306)
- ✅ Widely available
- ✅ Only 4 wires needed

**Search terms:**

- "0.96 inch OLED I2C 128x64 SSD1306"
- "OLED Display Module I2C IIC 0.96"

**Common brands:**

- Adafruit (premium, $17-20)
- Generic Chinese modules (good quality, $3-5)

### **Alternative: 1.3" I2C OLED (SH1106) - 128x64**

Same specs as 0.96" but larger physical size:

- Better visibility from distance
- Slightly higher power (20-25mA)
- $5-7
- Uses Adafruit_SH1106 library (or SSD1306 library works too)

**Recommendation:** Start with 0.96" version. Upgrade to 1.3" only if visibility is an issue.

## Pin Connections

### ESP32 DevKit V1 30-Pin Pinout Reference

```text
                         ESP32 DevKit V1
                    ┌─────────────────────┐
                    │                     │
            3V3  ───┤ 3V3             GND ├───  GND
            EN   ───┤ EN              D23 ├───  GPIO23 (VSPI MOSI - unused)
            VP   ───┤ VP(36)          D22 ├───  GPIO22 (I2C SCL)
            VN   ───┤ VN(39)          TX0 ├───  GPIO1 (TX)
            D34  ───┤ D34             RX0 ├───  GPIO3 (RX)
            D35  ───┤ D35             D21 ├───  GPIO21 (I2C SDA)
            D32  ───┤ D32             GND ├───  GND
            D33  ───┤ D33             D19 ├───  GPIO19 (VSPI MISO) ★
            D25  ───┤ D25             D18 ├───  GPIO18 (VSPI SCK)  ★
            D26  ───┤ D26              D5 ├───  GPIO5  (CS #1)     ★
            D27  ───┤ D27             D17 ├───  GPIO17 (CS #2)     ★
            D14  ───┤ D14             D16 ├───  GPIO16 (CS #3)     ★
            D12  ───┤ D12              D4 ├───  GPIO4
            GND  ───┤ GND              D0 ├───  GPIO0
            D13  ───┤ D13              D2 ├───  GPIO2
            D9   ───┤ D9              D15 ├───  GPIO15
            D10  ───┤ D10              D8 ├───  GPIO8
            D11  ───┤ D11              D7 ├───  GPIO7
            VIN  ───┤ VIN              D6 ├───  GPIO6
                    │     [USB]           │
                    └─────────────────────┘

★ = Used in this project
```

### Connection Tables

#### Hardware SPI Bus (Shared - All MAX31855 Modules)

| Function | ESP32 GPIO | ESP32 Pin Label | All 3 MAX31855 Modules |
|----------|------------|-----------------|------------------------|
| **SCK** (Clock) | GPIO18 | D18 | Connect to **SCK** pin on all 3 modules |
| **MISO** (Data) | GPIO19 | D19 | Connect to **DO/SO** pin on all 3 modules |
| **VCC** | 3.3V | 3V3 | Connect to **VIN/VCC** on all 3 modules |
| **GND** | Ground | GND | Connect to **GND** on all 3 modules |

**Note:** MOSI (GPIO23) is part of VSPI but not used by MAX31855 (read-only devices).

#### Individual Chip Select Pins

| Sensor | ESP32 GPIO | ESP32 Pin Label | MAX31855 Module | Purpose |
|--------|------------|-----------------|-----------------|---------|
| Kiln Primary | GPIO5 | D5 | Module 1 CS pin | Kiln thermocouple #1 |
| Kiln Secondary | GPIO17 | D17 | Module 2 CS pin | Kiln thermocouple #2 |
| Wood Material | GPIO16 | D16 | Module 3 CS pin | Wood thermocouple |

#### I2C OLED Display Connections

| Function | ESP32 GPIO | ESP32 Pin Label | OLED Display Pin |
|----------|------------|-----------------|------------------|
| **SDA** (Data) | GPIO21 | D21 | SDA |
| **SCL** (Clock) | GPIO22 | D22 | SCL |
| **VCC** | 3.3V | 3V3 | VCC/VDD |
| **GND** | Ground | GND | GND |

**Note:** Most OLED modules work with both 3.3V and 5V. Use 3.3V from ESP32 for consistency.

#### USB Connection

| Function | ESP32 | Raspberry Pi B+ |
|----------|-------|-----------------|
| **USB Data** | Micro-USB port | USB-A port (any) |
| **Purpose** | Programming + Serial communication | |
| **Cable** | Standard USB-A to Micro-USB | |

## Visual Wiring Diagram

```
                                   ESP32 DevKit V1
                              ┌───────────────────────┐
                              │                       │
                  ┌───────────┤ GPIO18 (SCK)          │
                  │     ┌─────┤ GPIO19 (MISO)         │
                  │     │  ┌──┤ GPIO5  (CS #1)        │
                  │     │  │┌─┤ GPIO17 (CS #2)        │
                  │     │  ││┌┤ GPIO16 (CS #3)        │
                  │     │  │││ │                       │
                  │     │  │││ │ GPIO21 (SDA) ────────┼──┐
                  │     │  │││ │ GPIO22 (SCL) ────────┼──┼──► OLED Display
                  │     │  │││ │                       │  │    ┌──────────┐
                  │     │  │││ │ 3.3V ────────────────┼──┼────┤ VCC  SDA │
                  │     │  │││ │ GND ─────────────────┼──┼────┤ GND  SCL │
                  │     │  │││ │                       │  │    └──────────┘
                  │     │  │││ └───────────────────────┘  │
                  │     │  │││                            │
                  ▼     ▼  ▼││                            ▼
        ┌─────────────────┐││                    (All share 3.3V & GND)
        │ MAX31855 #1     │││
        │ (Kiln Primary)  │││
        ├─────────────────┤││
        │ VIN  ◄── 3.3V   │││
        │ GND  ◄── GND    │││
        │ SCK  ◄──────────┘││
        │ DO   ◄───────────┘│
        │ CS   ◄── GPIO5    │
        │                   │
        │ T+  T- (K-Type)   │
        └───┬───┬───────────┘
            │   │
            └───┤ Thermocouple #1
                └─ Kiln Location

                  ▼              ▼
        ┌─────────────────────────┐
        │ MAX31855 #2             │
        │ (Kiln Secondary)        │
        ├─────────────────────────┤
        │ VIN  ◄── 3.3V (shared)  │
        │ GND  ◄── GND  (shared)  │
        │ SCK  ◄── GPIO18 (shared)│
        │ DO   ◄── GPIO19 (shared)│
        │ CS   ◄── GPIO17         │
        │                         │
        │ T+  T- (K-Type)         │
        └───┬───┬─────────────────┘
            │   │
            └───┤ Thermocouple #2
                └─ Kiln Location

                              ▼
        ┌─────────────────────────┐
        │ MAX31855 #3             │
        │ (Wood Material)         │
        ├─────────────────────────┤
        │ VIN  ◄── 3.3V (shared)  │
        │ GND  ◄── GND  (shared)  │
        │ SCK  ◄── GPIO18 (shared)│
        │ DO   ◄── GPIO19 (shared)│
        │ CS   ◄── GPIO16         │
        │                         │
        │ T+  T- (K-Type)         │
        └───┬───┬─────────────────┘
            │   │
            └───┤ Thermocouple #3
                └─ Wood Material
```

## Breadboard Layout Example

```
Power Rails:        ESP32 DevKit V1        MAX31855 Modules
════════════        ═══════════════        ════════════════

  + 3.3V ───────────────┐                  VIN VIN VIN
                        │                   │   │   │
  - GND ────────────────┼───────────────────┼───┼───┼─── GND GND GND
                        │                   │   │   │
                    ┌───┴───┐              │   │   │
                    │ ESP32 │              │   │   │
                    │DevKit │              │   │   │
                    └───┬───┘              │   │   │
                        │                  │   │   │
         ┌──────────────┼──────────────────┘   │   │
         │  ┌───────────┼──────────────────────┘   │
         │  │  ┌────────┼──────────────────────────┘
         │  │  │        │
    GPIO 18 19 5       17 16
     SCK MISO CS1     CS2 CS3
         │  │  │        │  │
         └──┼──┼────────┼──┼── SCK (All modules)
            └──┼────────┼──┼── DO  (All modules)
               └────────┘  └───── CS pins (individual)

    GPIO 21 22
     SDA SCL
      │   │
      └───┼──────► OLED SDA
          └──────► OLED SCL
```

## Complete Pin Summary Table

| Component | Pin | ESP32 GPIO | ESP32 Pin Label | Notes |
|-----------|-----|------------|-----------------|-------|
| **MAX31855 (all 3)** | SCK | GPIO18 | D18 | Shared SPI clock |
| **MAX31855 (all 3)** | DO/MISO | GPIO19 | D19 | Shared SPI data |
| **MAX31855 (all 3)** | VIN | 3.3V | 3V3 | Shared power |
| **MAX31855 (all 3)** | GND | GND | GND | Shared ground |
| **MAX31855 #1** | CS | GPIO5 | D5 | Individual chip select |
| **MAX31855 #2** | CS | GPIO17 | D17 | Individual chip select |
| **MAX31855 #3** | CS | GPIO16 | D16 | Individual chip select |
| **OLED Display** | SDA | GPIO21 | D21 | I2C data |
| **OLED Display** | SCL | GPIO22 | D22 | I2C clock |
| **OLED Display** | VCC | 3.3V | 3V3 | Power (3.3V or 5V) |
| **OLED Display** | GND | GND | GND | Ground |
| **USB Serial** | D+/D- | Built-in USB | Micro-USB port | Raspberry Pi connection |

## OLED I2C Address

Most OLED displays use one of two I2C addresses:
- **0x3C** (most common)
- **0x3D** (less common)

If the display doesn't work, try changing the address in the firmware:
```cpp
#define SCREEN_ADDRESS 0x3C  // Try 0x3D if this doesn't work
```

## Power Budget

| Device | Current Draw @ 3.3V |
|--------|---------------------|
| ESP32 (WiFi disabled) | ~40mA |
| OLED Display 0.96" | 15-20mA |
| MAX31855 #1 | 1.5mA |
| MAX31855 #2 | 1.5mA |
| MAX31855 #3 | 1.5mA |
| **Total** | **~60-65mA** |

USB 2.0 provides 500mA @ 5V, ESP32 3.3V regulator handles 600-800mA.
**Result:** Plenty of headroom, no power issues.

## Assembly Tips

### 1. **Power Distribution**
   - Use breadboard power rails for 3.3V and GND
   - Connect ESP32 3.3V and GND pins to power rails
   - Connect all modules to power rails

### 2. **Wire Management**
   - Keep SPI wires (SCK, MISO) as short as possible
   - Route SPI wires together, separate from I2C wires
   - Use different colored wires for power (red), ground (black), signals (others)

### 3. **Testing Order**
   1. Connect only ESP32, verify USB serial works
   2. Add OLED display, test with simple sketch
   3. Add one MAX31855, test temperature reading
   4. Add remaining MAX31855 modules
   5. Upload final firmware

### 4. **Prototyping Shield Option**
   - ESP32 prototyping shields available for $4-5
   - Provides solderable proto area for permanent installation
   - Cleaner than breadboard for final assembly

## Common OLED Module Pinouts

Most modules have one of these pin orders:

**Type 1 (Most common):**
```
┌────────────┐
│   OLED     │
│  Display   │
└─┬──┬──┬──┬─┘
  │  │  │  │
 GND VCC SCL SDA
```

**Type 2 (Alternative):**
```
┌────────────┐
│   OLED     │
│  Display   │
└─┬──┬──┬──┬─┘
  │  │  │  │
 VCC GND SCL SDA
```

**Always check your specific module!** Wrong power polarity can damage the display.

## Troubleshooting

### OLED Display Not Working

1. **Check I2C address:**
   ```bash
   # On ESP32, use I2C scanner sketch to find address
   # Usually 0x3C or 0x3D
   ```

2. **Check connections:**
   - SDA to GPIO21
   - SCL to GPIO22
   - VCC to 3.3V (not 5V if module is 3.3V only)
   - GND to GND

3. **Check power:**
   - Measure 3.3V at module VCC pin
   - Ensure module is getting power

### MAX31855 Shows NaN

1. **Check thermocouple connection:**
   - Ensure thermocouple is plugged into T+ and T-
   - Check polarity (red = +, yellow/white = -)

2. **Check SPI wiring:**
   - SCK to GPIO18 on all modules
   - DO/MISO to GPIO19 on all modules
   - CS to correct GPIO for each module

3. **Check power:**
   - All modules need VIN connected to 3.3V
   - All modules need GND connected

### Serial Not Working

1. **Check USB cable:** Some cables are charge-only, not data
2. **Check baud rate:** Must be 9600
3. **Check driver:** ESP32 uses CP2102 or CH340 USB-serial chip
4. **Check /dev/ttyUSB*:** May be /dev/ttyUSB0, /dev/ttyUSB1, etc.

## Library Installation

### Arduino IDE

```
Tools → Manage Libraries
Search and install:
  - "Adafruit MAX31855 library"
  - "Adafruit SSD1306"
  - "Adafruit GFX Library"
```

### PlatformIO

```ini
[env:esp32dev]
platform = espressif32
board = esp32dev
framework = arduino
lib_deps =
    adafruit/Adafruit MAX31855 library
    adafruit/Adafruit SSD1306
    adafruit/Adafruit GFX Library
```

## Next Steps

1. Assemble hardware per this wiring diagram
2. Install required libraries
3. Upload `sensorunit-esp32.ino` firmware
4. Test with serial monitor (9600 baud)
5. Verify OLED shows temperatures
6. Connect to Raspberry Pi via USB
7. Test with existing Go sensorunit service (no changes needed!)

## References

- [ESP32 Pinout Reference](https://randomnerdtutorials.com/esp32-pinout-reference-gpios/)
- [Adafruit MAX31855 Guide](https://learn.adafruit.com/thermocouple/)
- [Adafruit SSD1306 OLED Guide](https://learn.adafruit.com/monochrome-oled-breakouts/)
- [ESP32 SPI Reference](https://docs.espressif.com/projects/esp-idf/en/latest/esp32/api-reference/peripherals/spi_master.html)
