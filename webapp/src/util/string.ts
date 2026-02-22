const degreesCelsius = "°C";

export const celsius = (value?: number) => {
  return value !== undefined ? `${value.toFixed(1)}${degreesCelsius}` : `${degreesCelsius}`;
};

export const celsiusRange = (a: number, b: number) => {
  return `${a}-${b}${degreesCelsius}`;
};

export const validName = (name: string, forbiddenNames: string[] = []) => {
  return (
    !!name.match(/^[\wäöÄÖ\-, ]+$/) && !forbiddenNames.includes(name.trim())
  );
};
