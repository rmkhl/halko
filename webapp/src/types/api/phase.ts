export interface DeltaCycle {
  delta: number;
  above: number;
  below: number;
}

export interface Phase {
  id: string;
  name: string;
  cycleMode: string;
  constantCycle?: number;
  deltaCycles?: DeltaCycle[];
}
