const degreesCelsius = "°C";

// The sensor unit reports a failed probe as absolute zero. Nothing at or below
// it is a real reading, so show it as NaN rather than a believable -273.15°C.
const invalidTemperatureReading = -273.15;

export const celsius = (value?: number) => {
  if (value === undefined) {
    return `${degreesCelsius}`;
  }
  if (value <= invalidTemperatureReading) {
    return "NaN";
  }
  return `${value.toFixed(1)}${degreesCelsius}`;
};

export const celsiusRange = (a: number, b: number) => {
  return `${a}-${b}${degreesCelsius}`;
};

export const validName = (name: string, forbiddenNames: string[] = []) => {
  return (
    !!name.match(/^[\wäöÄÖ\-, ]+$/) && !forbiddenNames.includes(name.trim())
  );
};
