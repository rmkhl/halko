const common = {
  edit: "Edit",
  name: "Name",
  new: "New",
  save: "Save",
  cancel: "Cancel",
};

export const en = {
  common,
  header: {
    title: "Halko",
  },
  tabs: {
    current: "Current program",
    phases: "Phases",
    programs: "Programs",
  },
  phases: {
    ...common,
    cycles: {
      title: "Cycles",
      constant: "Constant",
      delta: "Delta",
      select: "Select cycle",
      addDeltaCycle: "Add delta cycle",
      range: "Range",
    },
  },
};
