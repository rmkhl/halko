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
  },
  programs: {
    ...common(),
    noRunning: "No currently running program",
    steps: {
      ...common(),
      add: "Add",
      title: "Steps",
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
