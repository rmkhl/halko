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
    description: {
      title: "Description",
      label: "Description",
      help:
        "Free notes about this program - what charge it suits, why the steps are shaped the way they are. Shown wherever the program is, and kept with the run in its history.",
    },
    equalize: {
      title: "Startup",
      delta: "Equalize delta (°C)",
      deltaHelp:
        "How close the kiln and the material must be before the first step begins.",
      steamPrewarm: "Warm the steam generator before the run",
      steamPrewarmHelp:
        "Runs the steam generator until the kiln shows it has reached boiling. Fails the run if it never does.",
      valueDelta: "Equalize delta: {{value}} °C",
      valuePrewarm: "Steam warm-up: {{value}}",
      on: "on",
      off: "off",
      usingDefault: "(default)",
      defaultsUnavailable:
        "Could not load the engine defaults, so these fields start empty. Anything left unset falls back to the configured default when the run starts.",
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
