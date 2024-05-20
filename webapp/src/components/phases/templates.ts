import { DeltaCycle, Phase } from "../../types/api";

const minDelta = -30;
const maxDelta = 30;
const deltaStep = 5;

const fullCycle = 100;
const halfCycle = 50;
const offCycle = 0;

const defaultAbove = fullCycle;

const defaultBelow = offCycle;

export const defaultConstant = halfCycle;

export const defaultDeltaCycles = (): DeltaCycle[] =>
  Array.from(Array((maxDelta - minDelta) / deltaStep + 1).keys())
    .map((_, i) => maxDelta - i * deltaStep)
    .map((delta, i) => {
      let above = defaultAbove;
      let below = defaultBelow;

      switch (delta) {
        case minDelta:
          below = defaultAbove;
          break;
        case maxDelta:
          above = defaultBelow;
          break;
      }

      return { above, below, delta };
    });

export const emptyConstantPhase = (): Phase => ({
  id: "",
  name: "",
  cycleMode: "constant",
  constantCycle: defaultConstant,
});
