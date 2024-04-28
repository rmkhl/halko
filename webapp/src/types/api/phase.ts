import { Cycle } from "./cycle";

interface ValidSensorRange {
  sensor: string;
  above: number;
  below: number;
}

export interface DeltaCycle {
  delta: number;
  above: Cycle;
  below: Cycle;
}

export interface Phase {
  id: string;
  name: string;
  validRange: ValidSensorRange[];
  cycleMode: string;
  constantCycle?: Cycle;
  deltaCycles?: DeltaCycle[];
}
