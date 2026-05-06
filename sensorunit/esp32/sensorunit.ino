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
// - "read" - Read the current temperature values
// - "helo" - Respond with "helo" (initial handshake)
//
// Hardware: ESP32 DevKit (Micro-USB)
// Sensors: 3x MAX31855 thermocouple amplifiers (K-type thermocouples)
// Display: 0.96" or 1.3" I2C OLED (SSD1306 or SH1106)
//
// Libraries required:
// - Adafruit MAX31855 library
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

#include <Adafruit_MAX31855.h>
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

// Initialize MAX31855 sensors using hardware SPI
Adafruit_MAX31855 sensor[3] = {
    Adafruit_MAX31855(KILN_PRIMARY_CS),
    Adafruit_MAX31855(KILN_SECONDARY_CS),
    Adafruit_MAX31855(WOOD_CS)
};

const char * const sensorName[3] = {"KilnPrimary", "KilnSecondary", "Wood"};

float temperature[3] = {0.0, 0.0, 0.0};
bool is_valid[3] = {false, false, false};

// Status and timing
char status_text[32] = "";
bool shown_disconnect = false;
unsigned long previousCommandMillis = 0;
unsigned long previousMillis = 0;

#define DISCONNECTED_INTERVAL 30000
#define INTERVAL 500
#define MAX_COMMAND_LENGTH 32

// Moving average for temperature readings
float measurement[3][4] = {
  {0.0, 0.0, 0.0, 0.0},
  {0.0, 0.0, 0.0, 0.0},
  {0.0, 0.0, 0.0, 0.0}
};
int n_measure = 0;

void displayTemperatures()
{
    display.clearDisplay();

    // Draw status line at top
    display.setTextSize(1);
    display.setTextColor(SSD1306_WHITE);
    display.setCursor(0, 0);
    if (strlen(status_text) > 0) {
        display.print(status_text);
    } else {
        display.print("Halko Sensor Unit");
    }

    // Draw separator line
    display.drawLine(0, 10, SCREEN_WIDTH, 10, SSD1306_WHITE);

    // Draw temperatures in larger font
    display.setTextSize(1);

    // Kiln Primary (line 2)
    display.setCursor(0, 18);
    display.print("K1:");
    if (is_valid[KILN_PRIMARY]) {
        display.setCursor(24, 18);
        display.print(temperature[KILN_PRIMARY], 1);
        display.print("C");
    } else {
        display.setCursor(30, 14);
        display.print("NaN");
    }

    // Kiln Secondary (line 3)
    display.setCursor(0, 28);
    display.print("K2:");
    if (is_valid[KILN_SECONDARY]) {
        display.setCursor(24, 28);
        display.print(temperature[KILN_SECONDARY], 1);
        display.print("C");
    } else {
        display.setCursor(30, 28);
        display.print("NaN");
    }

    // Wood (line 4)
    display.setCursor(0, 42);
    display.print("Wod:");
    if (is_valid[WOOD]) {
        display.setCursor(30, 42);
        display.print(temperature[WOOD], 1);
        display.print("C");
    } else {
        display.setCursor(30, 42);
        display.print("NaN");
    }

    // Activity indicator (bottom right corner)
    static bool blink = false;
    if (blink) {
        display.fillCircle(SCREEN_WIDTH - 4, SCREEN_HEIGHT - 4, 2, SSD1306_WHITE);
    }
    blink = !blink;

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

    // Initialize SPI
    SPI.begin();

    // Wait for MAX31855 to stabilize
    delay(500);

    // Update display
    strncpy(status_text, "Ready", sizeof(status_text));
    displayTemperatures();

    Serial.println("Initialization complete");
    Serial.println("Commands: helo; read; show TEXT;");
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
        // Read the current sensor
        double sensor_temperature = sensor[current_sensor].readCelsius();

        // MAX31855 has built-in error detection
        uint8_t fault = sensor[current_sensor].readError();

        if (isnan(sensor_temperature) || fault != 0)
        {
            // Sensor read failed or has a fault
            is_valid[current_sensor] = false;

            // Optional: Log fault codes for debugging via serial
            // Uncomment to enable fault reporting:
            // if (fault != 0) {
            //     Serial.print("Sensor ");
            //     Serial.print(current_sensor);
            //     Serial.print(" fault: 0x");
            //     Serial.println(fault, HEX);
            // }
        }
        else
        {
            // Valid reading - update moving average
            is_valid[current_sensor] = true;
            measurement[current_sensor][n_measure] = sensor_temperature;

            // Calculate average of last 4 readings
            temperature[current_sensor] = 0.0;
            for (int j = 0; j < 4; j++)
            {
                temperature[current_sensor] += measurement[current_sensor][j];
            }
            temperature[current_sensor] = temperature[current_sensor] / 4.0;
        }

        // Move to next sensor for next cycle (0→1→2→0...)
        current_sensor = (current_sensor + 1) % 3;

        // Advance measurement slot when wrapping back to sensor 0
        if (current_sensor == 0)
        {
            n_measure = (n_measure + 1) % 4;
            // Update display after reading all sensors
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
