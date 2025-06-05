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

char *sensorName[3] = {"OvenPrimary", "OvenSecondary", "Wood"};

// LCD
int LCD_RS = 14;
int LCD_EN = 15;
int LCD_D4 = 16;
int LCD_D5 = 17;
int LCD_D6 = 18;
int LCD_D7 = 19;

LiquidCrystal lcd(LCD_RS, LCD_EN, LCD_D4, LCD_D5, LCD_D6, LCD_D7);

void displayTemperature(int sensor, float temperature)
{
    int col = sensor * 5;
    char buffer[6];

    lcd.setCursor(col, 1);
    if (isnan(temperature)) // No valid temperature measurement
    {
        lcd.print(" NaN ");
    }
    else
    {
        sprintf(buffer, "%3dC ", int(temperature));
        lcd.print(buffer);
    }
}

void displayStatus(char *status)
{
    char buffer[16];

    sprintf(buffer, "%-15s", status);
    lcd.setCursor(0, 0);
    lcd.print(buffer);
}

char *ticker[] = {"*", "+"};

void displayRunning()
{
    static int i = 0;
    lcd.setCursor(15, 0);
    lcd.print(ticker[i]);
    i = (i + 1) & 1;
}

float temperature[3] = {0.0, 0.0, 0.0};

float previousCommandMillis = 0.0;
#define DISCONNECTED_INTERVAL 30000
#define MAX_COMMAND_LENGTH 24

char status_text[16] = "";

void processSerial()
{
    static char buffer[MAX_COMMAND_LENGTH + 1];
    static int bufferIndex = 0;

    while (Serial.available())
    {
        char c = Serial.read();
        if (c == ';')
        {
            buffer[bufferIndex] = '\0';
            bufferIndex = 0;
            break;
        }
        if (bufferIndex < MAX_COMMAND_LENGTH)
        {
            if (c == '\n' || c == '\r')
            {
                continue;
            }
            buffer[bufferIndex++] = c;
        }
    }
    Serial.println(bufferIndex);
    Serial.println(buffer);
    // Command received
    if (bufferIndex == 0)
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
                if (isnan(temperature[i]))
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
    }
    previousCommandMillis = millis();
    displayStatus(status_text);
}

unsigned long previousMillis = millis();
#define INTERVAL 1000

void setup()
{
    Serial.begin(9600);
    lcd.begin(16, 2);
    lcd.clear();
    displayStatus("Initializing");
    for (int i = 0; i < 3; i++)
    {
        temperature[i] = sensor[i].readCelsius();
        delay(INTERVAL);
    }
    displayStatus("Disconnected");
    displayRunning();
}

void loop()
{
    unsigned long currentMillis = millis();
    if (Serial.available())
    {
        processSerial();
    }
    // Read the sensors every second and keep running average
    if (currentMillis - previousMillis >= INTERVAL)
    {
        for (int i = 0; i < 3; i++)
        {
            temperature[i] = (temperature[i] + sensor[i].readCelsius()) / 2.0;
            displayTemperature(i, temperature[i]);
        }
        displayRunning();
        previousMillis = currentMillis;

        if (currentMillis - previousCommandMillis >= DISCONNECTED_INTERVAL)
        {
            displayStatus("Disconnected");
        }
    }
}
