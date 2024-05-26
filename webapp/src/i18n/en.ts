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
  },
};
