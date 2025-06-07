const commonTemplate = {
  edit: "Edit",
  name: "Name",
  new: "New",
  save: "Save",
  cancel: "Cancel",
};

const common = (
  args?: Partial<typeof commonTemplate>
): typeof commonTemplate => ({
  ...commonTemplate,
  ...args,
});

export const en = {
  common: { ...common() },
  time: {
    seconds: "seconds",
  },
  header: {
    title: "Halko",
  },
  tabs: {
    current: "Current program",
    programs: "Programs",
  },
  programs: {
    ...common(),
    noRunning: "No currently running program",
    defaultStepRuntime: "Default step runtime",
    preheatTo: "Preheat oven to",
    steps: {
      ...common(),
      add: "Add step",
      title: "Steps",
      timeConstraint: "Time constraint",
      temperatureConstraint: {
        title: "Temperature constraint",
        minimum: "Minimum",
        maximum: "Maximum",
      },
      heater: "Heater",
      fan: "Fan",
      humidifier: "Humidifier",
    },
  },
  sensors: {
    material: "Material",
    oven: "Oven",
  },
};
