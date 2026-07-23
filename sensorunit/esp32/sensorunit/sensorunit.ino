// Control software for the sensor unit with ESP32 + MAX31855
//
// ESP32 replacement for Arduino Nano with improved SPI handling
// WiFi and Bluetooth disabled for power savings and simplicity
//
// Read temperatures from 3 MAX31855 thermocouples and display
// them on an I2C OLED display. Values can be queried via USB serial port.
// Status line can be updated via serial command.
//
// Commands (identical to Arduino version):
// - "show <text>" - Set the status line text
// - "addr <text>" - Set the IP address line text
// - "read" - Read the current temperature values
// - "helo" - Respond with "helo" (initial handshake)
//
// Hardware: ESP32 DevKit (Micro-USB)
// Sensors: 3x MAX31855 thermocouple amplifiers (K-type thermocouples)
// Display: 0.96" or 1.3" I2C OLED (SSD1306 or SH1106)
//
// Libraries required:
// - Adafruit SSD1306 library
// - Adafruit GFX library
// - Wire (built-in for ESP32)
//
// Hardware SPI Pins (ESP32 VSPI):
// - SCK:  GPIO18
// - MISO: GPIO19
// - MOSI: GPIO23 (not used by MAX31855, but part of SPI bus)
//
// I2C Pins (ESP32 default):
// - SDA: GPIO21
// - SCL: GPIO22
//

#include <Adafruit_GFX.h>
#include <Adafruit_SSD1306.h>
#include <Wire.h>
#include <SPI.h>

// Disable WiFi and Bluetooth to save power
#include <WiFi.h>
#include "esp_bt.h"

// OLED Display Configuration
#define SCREEN_WIDTH 128
#define SCREEN_HEIGHT 64
#define OLED_RESET    -1  // Reset pin # (or -1 if sharing ESP32 reset pin)
#define SCREEN_ADDRESS 0x3C  // Common I2C address (try 0x3D if 0x3C doesn't work)

Adafruit_SSD1306 display(SCREEN_WIDTH, SCREEN_HEIGHT, &Wire, OLED_RESET);

// MAX31855 Chip Select pins
#define KILN_PRIMARY_CS   5
#define KILN_SECONDARY_CS 17
#define WOOD_CS           16

// Sensor unit identifiers
#define KILN_PRIMARY   0
#define KILN_SECONDARY 1
#define WOOD           2

const int cs_pin[3] = {KILN_PRIMARY_CS, KILN_SECONDARY_CS, WOOD_CS};

// MAX31855 fault bits (D2..D0 of the data frame)
#define FAULT_OPEN 0x1  // thermocouple circuit broken
#define FAULT_GND  0x2  // thermocouple shorted/leaking to ground
#define FAULT_VCC  0x4  // thermocouple shorted to supply

const char * const sensorName[3] = {"KilnPrimary", "KilnSecondary", "Wood"};

float temperature[3] = {0.0, 0.0, 0.0};
bool is_valid[3] = {false, false, false};
uint8_t last_fault[3] = {0, 0, 0};
char addr_text[32] = "";
uint16_t fault_total[3] = {0, 0, 0};

// One SPI transaction returns the whole 32-bit MAX31855 frame, so the
// temperature and the fault bits always come from the same conversion.
uint32_t readRawFrame(int pin)
{
    SPI.beginTransaction(SPISettings(1000000, MSBFIRST, SPI_MODE0));
    digitalWrite(pin, LOW);
    uint32_t raw = 0;
    for (int i = 0; i < 4; i++)
    {
        raw = (raw << 8) | SPI.transfer(0);
    }
    digitalWrite(pin, HIGH);
    SPI.endTransaction();
    return raw;
}

// Returns the thermocouple temperature, or NAN on fault after recording
// which fault type fired so the display can show it.
float parseFrame(int idx, uint32_t raw)
{
    // An all-zero frame cannot come from a working chip (the internal
    // temperature bits are never all zero); it means the module is not
    // answering at all. No fault type in that case, just NaN.
    if (raw == 0)
    {
        last_fault[idx] = 0;
        return NAN;
    }

    uint8_t faults = raw & 0x7;
    if (faults != 0)
    {
        if (fault_total[idx] < 0xFFFF)
        {
            fault_total[idx]++;
        }
        last_fault[idx] = faults;
        return NAN;
    }

    // D[31:18] is the 14-bit signed thermocouple value, 0.25 °C per LSB
    int32_t v = (int32_t)raw >> 18;
    return v * 0.25f;
}

// Status and timing
char status_text[32] = "";
bool shown_disconnect = false;
unsigned long previousCommandMillis = 0;
unsigned long previousMillis = 0;

#define DISCONNECTED_INTERVAL 30000
#define INTERVAL 500
#define MAX_COMMAND_LENGTH 32

// Median filter for temperature readings: a single noise spike is rejected
// outright instead of being smeared over the whole window as with a mean.
#define SAMPLE_COUNT 5
// Consecutive failed reads before a sensor is reported invalid. Transient
// fault-bit glitches (e.g. from lead noise) keep the last good value.
#define FAULT_LIMIT 4

float measurement[3][SAMPLE_COUNT];
int sample_index[3] = {0, 0, 0};
bool seeded[3] = {false, false, false};
int fault_count[3] = {0, 0, 0};

float medianOfSamples(const float *samples)
{
    float sorted[SAMPLE_COUNT];
    memcpy(sorted, samples, sizeof(sorted));
    for (int i = 1; i < SAMPLE_COUNT; i++)
    {
        float value = sorted[i];
        int j = i - 1;
        while (j >= 0 && sorted[j] > value)
        {
            sorted[j + 1] = sorted[j];
            j--;
        }
        sorted[j + 1] = value;
    }
    return sorted[SAMPLE_COUNT / 2];
}

void displayTemperatures()
{
    display.clearDisplay();
    display.setTextColor(SSD1306_WHITE);

    // Column x positions for 3 sensors — fits 3×3-digit temps in 128px at size 2
    const int colX[3] = {0, 44, 88};
    const char* labels[3] = {"K1", "K2", "Wd"};

    // Row 1: temperatures in large font (size 2, 16px tall)
    display.setTextSize(2);
    for (int i = 0; i < 3; i++) {
        display.setCursor(colX[i], 0);
        if (is_valid[i]) {
            display.print(lroundf(temperature[i]));
        } else if (last_fault[i] & FAULT_GND) {
            display.print("GND");
        } else if (last_fault[i] & FAULT_OPEN) {
            display.print("OPN");
        } else if (last_fault[i] & FAULT_VCC) {
            display.print("VCC");
        } else {
            display.print("NaN");
        }
    }

    // Row 2: labels in small font (size 1, 8px tall) at y=17
    display.setTextSize(1);
    for (int i = 0; i < 3; i++) {
        display.setCursor(colX[i], 17);
        display.print(labels[i]);
    }

    // Separator line at y=27
    display.drawLine(0, 27, SCREEN_WIDTH - 1, 27, SSD1306_WHITE);

    // Row 3: status text at y=31
    display.setCursor(0, 31);
    if (strlen(status_text) > 0) {
        display.print(status_text);
    } else {
        display.print("Halko Sensor");
    }

    // Blinking activity dot (right side of status row)
    static bool blink = false;
    if (blink) {
        display.fillCircle(SCREEN_WIDTH - 4, 35, 3, SSD1306_WHITE);
    }
    blink = !blink;

    // Row 4: IP address (from "addr" command) at y=40
    if (strlen(addr_text) > 0) {
        display.setCursor(0, 40);
        display.print(addr_text);
    }

    // Rows 5+6: per-sensor self-diagnostics, columns aligned under the temps.
    // F = cumulative fault count since boot, L = last fault type.
    const int diagX[3] = {12, 54, 96};

    display.setCursor(0, 48);
    display.print("F");
    for (int i = 0; i < 3; i++) {
        display.setCursor(diagX[i], 48);
        display.print(fault_total[i] > 9999 ? 9999 : fault_total[i]);
    }

    display.setCursor(0, 56);
    display.print("L");
    for (int i = 0; i < 3; i++) {
        display.setCursor(diagX[i], 56);
        if (last_fault[i] & FAULT_GND) {
            display.print("GND");
        } else if (last_fault[i] & FAULT_OPEN) {
            display.print("OPN");
        } else if (last_fault[i] & FAULT_VCC) {
            display.print("VCC");
        } else {
            display.print("--");
        }
    }

    display.display();
}

void processSerial()
{
    static char buffer[MAX_COMMAND_LENGTH + 1];
    static int bufferIndex = 0;
    static bool command_ready = false;

    while (Serial.available())
    {
        char c = Serial.read();

        switch (c) {
          case '\n':
          case '\r':
              break;  // ignore end of line characters
          case ';':
              buffer[bufferIndex] = '\0';
              bufferIndex = 0;
              command_ready = true;
              break;
          default:
              buffer[bufferIndex] = c;
              if (bufferIndex < MAX_COMMAND_LENGTH)
              {
                  bufferIndex++;
              }
        }
    }

    if (command_ready)
    {
        shown_disconnect = false;  // Reset flag on any command

        char *command = strtok(buffer, " ");
        if (strcmp(command, "show") == 0)
        {
            // Extract text after "show "
            const char *text = buffer + 5;  // Skip "show "
            strncpy(status_text, text, sizeof(status_text) - 1);
            status_text[sizeof(status_text) - 1] = '\0';
            displayTemperatures();
        }
        else if (strcmp(command, "addr") == 0)
        {
            // Extract text after "addr "
            const char *text = buffer + 5;  // Skip "addr "
            strncpy(addr_text, text, sizeof(addr_text) - 1);
            addr_text[sizeof(addr_text) - 1] = '\0';
            displayTemperatures();
        }
        else if (strcmp(command, "read") == 0)
        {
            for (int i = 0; i < 3; i++)
            {
                Serial.print(sensorName[i]);
                Serial.print("=");
                if (!is_valid[i])
                {
                    Serial.print("NaN");
                }
                else
                {
                    Serial.print(temperature[i]);
                    Serial.print("C");
                }
                if (i < 2)
                {
                    Serial.print(",");
                }
                else
                {
                    Serial.println();
                }
            }
        }
        else if (strcmp(command, "helo") == 0)
        {
            Serial.println("helo");
        }
        command_ready = false;
    }
    previousCommandMillis = millis();
}

void setup()
{
    // Disable WiFi and Bluetooth to save power and reduce complexity
    WiFi.mode(WIFI_OFF);
    btStop();
    esp_bt_controller_disable();

    // Initialize USB serial
    Serial.begin(9600);
    Serial.println("Halko ESP32 Sensor Unit");

    // Initialize I2C for OLED
    Wire.begin();

    // Initialize OLED display
    if (!display.begin(SSD1306_SWITCHCAPVCC, SCREEN_ADDRESS)) {
        Serial.println("Error: OLED display not found at 0x3C");
        Serial.println("Check wiring or try address 0x3D");
        // Continue anyway - serial interface will still work
    } else {
        display.clearDisplay();
        display.setTextSize(1);
        display.setTextColor(SSD1306_WHITE);
        display.setCursor(0, 0);
        display.println("Halko Sensor Unit");
        display.println("Initializing...");
        display.display();
    }

    // Initialize SPI and the MAX31855 chip selects (idle high)
    SPI.begin();
    for (int i = 0; i < 3; i++)
    {
        pinMode(cs_pin[i], OUTPUT);
        digitalWrite(cs_pin[i], HIGH);
    }

    // Wait for sensors to stabilize
    delay(500);

    // Update display
    strncpy(status_text, "Ready", sizeof(status_text));
    displayTemperatures();

    Serial.println("Initialization complete");
    Serial.println("Commands: helo; read; show TEXT; addr TEXT;");
}

void loop()
{
    static int current_sensor = 0;  // Track which sensor to read this cycle

    unsigned long currentMillis = millis();

    // Process serial commands
    if (Serial.available())
    {
        processSerial();
    }

    // Read one sensor per cycle for better responsiveness
    if (currentMillis - previousMillis >= INTERVAL)
    {
        float sensor_temperature = parseFrame(current_sensor, readRawFrame(cs_pin[current_sensor]));

        if (isnan(sensor_temperature))
        {
            if (fault_count[current_sensor] < FAULT_LIMIT)
            {
                fault_count[current_sensor]++;
            }
            if (fault_count[current_sensor] >= FAULT_LIMIT)
            {
                is_valid[current_sensor] = false;
                // Samples go stale during a real outage; re-seed on recovery
                seeded[current_sensor] = false;
            }
        }
        else
        {
            fault_count[current_sensor] = 0;
            if (!seeded[current_sensor])
            {
                // Fill the whole window so the reported value is correct
                // immediately instead of ramping up from empty slots
                for (int j = 0; j < SAMPLE_COUNT; j++)
                {
                    measurement[current_sensor][j] = sensor_temperature;
                }
                seeded[current_sensor] = true;
            }
            else
            {
                measurement[current_sensor][sample_index[current_sensor]] = sensor_temperature;
                sample_index[current_sensor] = (sample_index[current_sensor] + 1) % SAMPLE_COUNT;
            }
            is_valid[current_sensor] = true;
            temperature[current_sensor] = medianOfSamples(measurement[current_sensor]);
        }

        // Sensor counter drives display refresh timing
        current_sensor = (current_sensor + 1) % 3;

        if (current_sensor == 0)
        {
            displayTemperatures();
        }

        previousMillis = currentMillis;

        // Show disconnected status if no serial commands for 30 seconds
        if (currentMillis - previousCommandMillis >= DISCONNECTED_INTERVAL)
        {
            if (!shown_disconnect)
            {
                strncpy(status_text, "Disconnected", sizeof(status_text));
                displayTemperatures();
                shown_disconnect = true;
            }
        }
    }
}
