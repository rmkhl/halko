// Control software for the sensor unit
//
// Read temperatures from 3 MAX6675 thermocouples every second and display
// them on a 16x2 LCD. The values can also be queried via serial port.
// In addition the text on the status line can be updated via command over
// the serial port. If no commands are received for 30 seconds the status
// will indicate the unit is disconnected.
//
// Temperature are displayed on the LCD regardless of the serial port connection
// status.
//
// Commands:
// - "show <text>" - Set the status line text
// - "read" - Read the current temperature values
// - "helo" - Respond with "helo" (initial handshake)
//
// Arduino board: Arduino Nano (ok, chinese clone)
// Sensors: 3x MAX6675 thermocouple
// Display: 16x2 LCD LCM 1602C
//
// Libraries:
// - MAX6675: AdaFruit MAX6675 library
// - LiquidCrystal: LiquidCrystal library 1.0.7
//
#include <max6675.h>
#include <LiquidCrystal.h>
#include <math.h>

// Pins used:
// MAX6675
int SPI_CLK = 2;
int SPI_MISO = 3;
int OVEN_PRIMARY_TEMPERATURE_SENSOR_CS = 4;
int OVEN_SECONDARY_TEMPERATURE_SENSOR_CS = 5;
int WOOD_TEMPERATURE_SENSOR_CS = 6;

// Sensor unit identifiers
#define OVEN_PRIMARY 0
#define OVEN_SECONDARY 1
#define WOOD 2

MAX6675 sensor[3] = {
    MAX6675(SPI_CLK, OVEN_PRIMARY_TEMPERATURE_SENSOR_CS, SPI_MISO),
    MAX6675(SPI_CLK, OVEN_SECONDARY_TEMPERATURE_SENSOR_CS, SPI_MISO),
    MAX6675(SPI_CLK, WOOD_TEMPERATURE_SENSOR_CS, SPI_MISO)};

const char * const sensorName[3] = {"OvenPrimary", "OvenSecondary", "Wood"};

// LCD
int LCD_RS = 14;
int LCD_EN = 15;
int LCD_D4 = 16;
int LCD_D5 = 17;
int LCD_D6 = 18;
int LCD_D7 = 19;

LiquidCrystal lcd(LCD_RS, LCD_EN, LCD_D4, LCD_D5, LCD_D6, LCD_D7);

float temperature[3] = {0.0, 0.0, 0.0};
bool is_valid[3] = {false, false, false};

void displayTemperature(int sensor, float temperature)
{
    int col = sensor * 5;
    char buffer[6];

    lcd.setCursor(col, 1);
    if (!is_valid[sensor]) // No valid temperature measurement
    {
        lcd.print(" NaN ");
    }
    else
    {
        sprintf(buffer, "%3dC ", (int)round(temperature));
        lcd.print(buffer);
    }
}

void displayStatus(const char * const status)
{
    char buffer[16];

    sprintf(buffer, "%-15s", status);
    lcd.setCursor(0, 0);
    lcd.print(buffer);
}

const char * const ticker[] = {"*", "+"};

void displayRunning()
{
    static int i = 0;
    lcd.setCursor(15, 0);
    lcd.print(ticker[i]);
    i = (i + 1) & 1;
}

float previousCommandMillis = 0.0;
#define DISCONNECTED_INTERVAL 30000
#define MAX_COMMAND_LENGTH 24

char status_text[16] = "";

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
        char *command = strtok(buffer, " ");
        if (strcmp(command, "show") == 0)
        {
            strcpy(status_text, &buffer[5]);
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
    displayStatus(status_text);
}

unsigned long previousMillis = millis();
#define INTERVAL 500

void setup()
{
    Serial.begin(9600);
    lcd.begin(16, 2);
    lcd.clear();
    displayStatus("Initializing");
    displayRunning();
}

float measurement[3][4] = {
  {0.0, 0.0, 0.0, 0.0},
  {0.0, 0.0, 0.0, 0.0},
  {0.0, 0.0, 0.0, 0.0}};

int n_measure = 0;

void loop()
{
    static int current_sensor = 0;  // Track which sensor to read this cycle

    unsigned long currentMillis = millis();
    if (Serial.available())
    {
        processSerial();
    }
    // Read one sensor per cycle for better responsiveness
    if (currentMillis - previousMillis >= INTERVAL)
    {
        // Read only the current sensor this cycle
        float sensor_temperature = sensor[current_sensor].readCelsius();
        if (isnan(sensor_temperature)) {
            is_valid[current_sensor] = false;
            displayTemperature(current_sensor, sensor_temperature);
        } else {
            is_valid[current_sensor] = true;
            measurement[current_sensor][n_measure] = sensor_temperature;
            temperature[current_sensor] = 0.0;
            for (int j = 0; j < 4; j++) {
                temperature[current_sensor] += measurement[current_sensor][j];
            }
            temperature[current_sensor] = temperature[current_sensor] / 4.0;
            displayTemperature(current_sensor, temperature[current_sensor]);
        }

        // Move to next sensor for next cycle
        current_sensor = (current_sensor + 1) % 3;

        // Only advance measurement index after all 3 sensors have been read
        if (current_sensor == 0) {
            n_measure = (n_measure + 1) % 4;
        }

        displayRunning();
        previousMillis = currentMillis;

        if (currentMillis - previousCommandMillis >= DISCONNECTED_INTERVAL)
        {
            displayStatus("Disconnected");
        }
    }
}
