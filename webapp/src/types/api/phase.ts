export interface DeltaCycle {
  delta: number;
  above: number;
  below: number;
}

export interface Phase {
  name: string;
  cycleMode: string;
  constantCycle?: number;
  deltaCycles?: DeltaCycle[];
}
