# ESP32 Sensor Unit

This directory contains ESP32-based firmware for the Halko sensor unit.

## Overview

The ESP32 is a more powerful alternative to the Arduino Nano ATmega328P, offering:

- Faster processor (240 MHz dual-core vs 16 MHz)
- More memory (520 KB SRAM vs 2 KB)
- Built-in WiFi and Bluetooth (disabled for simplicity and power savings)
- Hardware SPI, I2C, and multiple UARTs
- Better development tools and libraries

## Directory Structure

```text
sensorunit/esp32/
├── README.md              # This file
├── WIRING-ESP32.md        # Hardware wiring guide
└── sensorunit/            # ESP32 sensorunit firmware
    └── sensorunit.ino     # Main firmware (MAX31855 + OLED)
```

## Hardware

### NodeMCU-32S (AZDelivery ESP32 with CP2102)

- **Microcontroller**: ESP32-WROOM-32
- **USB-to-Serial**: CP2102 chip
- **USB**: Micro-USB
- **Built-in LED**: GPIO2 (used by blink test)
- **Operating Voltage**: 3.3V (5V USB power input)
- **Clock Speed**: 240 MHz (dual-core)
- **Board Identifier**: `esp32:esp32:nodemcu-32s`

### Pin Mapping (for future MAX31855 integration)

**Hardware SPI (VSPI):**

- SCK:  GPIO18
- MISO: GPIO19
- MOSI: GPIO23 (not used by MAX31855, but part of SPI bus)
- CS pins: GPIO5, GPIO16, GPIO17 (for three sensors)

**I2C (for OLED display):**

- SDA: GPIO21
- SCL: GPIO22

## Development Workflow

### First-Time Setup

1. **Install ESP32 board support:**

   ```bash
   make prepare-esp32
   ```

   This will:
   - Install Arduino CLI (if not already installed)
   - Add ESP32 board manager URL to configuration
   - Install ESP32 board support package
   - Create firmware-esp32/ directory

### Building and Uploading

1. **Build the sensorunit firmware:**

   ```bash
   make build-esp32
   ```

   Compiles `sensorunit/esp32/sensorunit/sensorunit.ino` to `firmware-esp32/`

2. **Upload to ESP32:**

   ```bash
   make upload-esp32             # Upload to /dev/ttyUSB0
   make upload-esp32 PORT=/dev/ttyUSB1  # Upload to specific port
   ```

### Testing

1. **Monitor serial output:**

   ```bash
   make monitor-esp32            # Connect to /dev/ttyUSB0
   make monitor-esp32 PORT=/dev/ttyUSB1  # Connect to specific port
   ```

   - Baud rate: 115200
   - Displays temperature readings from MAX31855 sensors
   - Shows status on OLED display
   - Press Ctrl+C to exit (may need Ctrl+A then K in screen)

### Maintenance

```bash
make clean-esp32              # Remove build artifacts
make esp32-help               # Show all available commands
```

## Firmware Features

The ESP32 sensorunit firmware provides:

- **Temperature Reading**: 3x MAX31855 thermocouple amplifiers (K-type)
- **Display**: I2C OLED (SSD1306/SH1106) for local temperature display
- **Serial Protocol**: USB serial interface compatible with controlunit
- **Commands**:
  - `read` - Read current temperature values
  - `show <text>` - Set status line on display
  - `helo` - Initial handshake response
- **Power Management**: WiFi and Bluetooth disabled for power savings

## Development Notes

### Serial Port Permissions

If you get permission errors accessing `/dev/ttyUSB0`:

```bash
sudo usermod -a -G dialout $USER
# Then log out and back in
```

### Finding the ESP32 Port

```bash
# Before plugging in ESP32:
ls /dev/ttyUSB*

# Plug in ESP32, then:
ls /dev/ttyUSB*
# The new device is your ESP32
```

### Troubleshooting

**Upload fails with "Failed to connect to ESP32":**

- Hold the BOOT button on ESP32 during upload
- Try a different USB cable (some are power-only)
- Check that port has correct permissions

**Serial monitor shows garbage characters:**

- Verify baud rate is 115200
- Press the EN (reset) button on ESP32 to restart the program

**Build fails with "esp32:esp32:esp32 not found":**

- Run `make prepare-esp32` to install board support
- Check that `.arduino-cli/bin/arduino-cli` exists

## Hardware Requirements

- **NodeMCU-32S** (AZDelivery ESP32 with CP2102)
- **3x MAX31855** thermocouple amplifier modules
- **3x K-type thermocouples**
- **I2C OLED display** (0.96" or 1.3", SSD1306 or SH1106 compatible)
- Jumper wires for connections

See [WIRING-ESP32.md](WIRING-ESP32.md) for complete wiring instructions.

## Future Enhancements

Potential improvements:

- WiFi-based temperature reporting (currently disabled for power savings)
- OTA (Over-The-Air) firmware updates
- Web-based configuration interface
- Data logging to SD card

## References

- [ESP32 Arduino Core Documentation](https://docs.espressif.com/projects/arduino-esp32/en/latest/)
- [ESP32 Technical Reference](https://www.espressif.com/sites/default/files/documentation/esp32_technical_reference_manual_en.pdf)
- [Arduino-CLI Documentation](https://arduino.github.io/arduino-cli/)
- [MAX31855 Datasheet](https://www.maximintegrated.com/en/products/sensors/MAX31855.html)
