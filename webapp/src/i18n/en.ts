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
    cycles: "Cycles",
    phases: "Phases",
    programs: "Programs",
  },
  cycle: {
    ...common,
  },
  phases: {
    ...common,
    validRange: {
      title: "Valid ranges",
      material: "Material",
      above: "Above",
      below: "Below",
    },
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
