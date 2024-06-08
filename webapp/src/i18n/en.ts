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
    phases: "Phases",
    programs: "Programs",
  },
  phases: {
    ...common(),
    cycles: {
      title: "Cycles",
      constant: "Constant",
      delta: "Delta",
      select: "Select cycle",
      addDeltaCycle: "Add delta cycle",
      range: "Range",
    },
  },
  programs: {
    ...common(),
    noRunning: "No currently running program",
    defaultStepRuntime: "Default step runtime",
    preheatTo: "Preheat oven to",
    steps: {
      ...common(),
      add: "Add",
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
      selectPhase: "Select",
      noPhaseSelected: "No phase selected",
    },
  },
};
