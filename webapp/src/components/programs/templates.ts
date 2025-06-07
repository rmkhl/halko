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
    },
    {
      id: uuidv4(),
      name: "Cool down",
      type: "cooling",
      targetTemperature: 30,
    },
  ],
});
