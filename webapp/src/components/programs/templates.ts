import { Program } from "../../types/api";
import { v4 as uuidv4 } from "uuid";

export const emptyProgram = (): Program => ({
  name: "",
  steps: [
    {
      id: uuidv4(),
      name: "Heat",
      type: "heating",
      targetTemperature: 100,
      heater: {
        pid: {},
        power: 100,
      },
      humidifier: {
        power: 50,
      },
      fan: {
        power: 0,
      },
    },
    {
      id: uuidv4(),
      name: "Cool down",
      type: "cooling",
      targetTemperature: 30,
      heater: {
        pid: {},
        power: 0,
      },
      humidifier: {
        power: 0,
      },
      fan: {
        power: 0,
      },
    },
  ],
});
