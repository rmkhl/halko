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
    display.setTextColor(SSD1306_WHITE);

    // Column x positions for 3 sensors — fits 3×3-digit temps in 128px at size 2
    const int colX[3] = {0, 44, 88};
    const char* labels[3] = {"K1", "K2", "Wd"};

    // Row 1: temperatures in large font (size 2, 16px tall)
    display.setTextSize(2);
    for (int i = 0; i < 3; i++) {
        display.setCursor(colX[i], 0);
        if (is_valid[i]) {
            display.print((int)temperature[i]);
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

    // Wait for sensors to stabilize
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
        double sensor_temperature = sensor[current_sensor].readCelsius();
        uint8_t fault = sensor[current_sensor].readError();

        if (isnan(sensor_temperature) || fault != 0)
        {
            is_valid[current_sensor] = false;
        }
        else
        {
            is_valid[current_sensor] = true;
            measurement[current_sensor][n_measure] = sensor_temperature;
            temperature[current_sensor] = 0.0;
            for (int j = 0; j < 4; j++)
            {
                temperature[current_sensor] += measurement[current_sensor][j];
            }
            temperature[current_sensor] = temperature[current_sensor] / 4.0;
        }

        // Sensor counter drives display refresh timing
        current_sensor = (current_sensor + 1) % 3;

        if (current_sensor == 0)
        {
            n_measure = (n_measure + 1) % 4;
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
