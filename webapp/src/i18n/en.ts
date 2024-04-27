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
      oven: "Oven",
      material: "Material",
      above: "Above",
      below: "Below",
    },
    cycles: {
      title: "Cycles",
      constant: "Constant",
      delta: "Delta",
    },
  },
};
