const commonTemplate = {
  edit: "Edit",
  name: "Name",
  new: "New",
  save: "Save",
  cancel: "Cancel",
  back: "Back",
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
    running: "Status",
    history: "History",
    programs: "Programs",
    system: "System",
  },
  programs: {
    ...common(),
    noRunning: "No currently running program",
    stop: "Stop",
    stopping: "Stopping...",
    confirmStopTitle: "Confirm Stop",
    confirmStopBody:
      "Are you sure you want to stop the running program? The run will be marked as canceled and all power switched off. This cannot be undone.",
    equalize: {
      title: "Startup",
      delta: "Equalize delta (°C)",
      deltaHelp:
        "How close the kiln and the material must be before the first step begins.",
      deltaHelpDefault: "Leave empty to use the configured default of {{delta}} °C.",
      steamPrewarm: "Warm the steam generator before the run",
      steamPrewarmHelp:
        "Runs the steam generator until the kiln shows it has reached boiling. Fails the run if it never does.",
    },
    steps: {
      ...common(),
      add: "Add",
      title: "Steps",
      heater: "Heater",
      fan: "Fan",
      steam: "Steam",
    },
  },
  sensors: {
    material: "Material",
    kiln: "Kiln",
  },
  system: {
    title: "System Status",
    services: "Services",
    storage: "Storage",
    hardware: "Hardware",
    systemInfo: "System Information",
  },
};
