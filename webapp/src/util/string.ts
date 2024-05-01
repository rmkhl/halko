const degreesCelsius = "°C";

export const celsius = (value: number) => {
  return `${value}${degreesCelsius}`;
};

export const celsiusRange = (a: number, b: number) => {
  return `${a}-${b}${degreesCelsius}`;
};
