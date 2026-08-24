// Parsing and step-grouping of controlunit execution logs. Columns:
// time,step,steptime,material,kiln,heater,fan,steam,
// material_die,kiln_primary_die,kiln_secondary_die,kiln_primary,kiln_secondary
//
// The last two are appended, so histories written before the kiln pair landed
// have eleven columns and parse without them.

export interface LogRow {
  time: number;
  step: string;
  steptime: number;
  material: number;
  kiln: number;
  heater: number;
  fan: number;
  steam: number;
  kilnPrimary?: number;
  kilnSecondary?: number;
}

export interface StepSegment {
  step: string;
  rows: LogRow[];
}

// A column an older history does not have, or one holding something
// unparseable, reads as absent rather than as NaN in a chart.
const optionalReading = (value?: string): number | undefined => {
  if (value === undefined) {
    return undefined;
  }
  const parsed = parseFloat(value);
  return isNaN(parsed) ? undefined : parsed;
};

export const parseExecutionLog = (csv: string): LogRow[] => {
  const lines = csv.trim().split("\n").filter((line) => line.trim().length > 0);
  const rows: LogRow[] = [];

  for (let i = 1; i < lines.length; i++) {
    const line = lines[i].trim();
    if (!line) continue;

    const values = line.split(",");

    // Skip lines that don't have enough values
    if (values.length < 8) {
      continue;
    }

    const time = parseFloat(values[0]);
    const material = parseFloat(values[3]);
    const kiln = parseFloat(values[4]);

    // Skip lines with invalid numeric values
    if (isNaN(time) || isNaN(material) || isNaN(kiln)) {
      continue;
    }

    rows.push({
      time,
      step: values[1],
      steptime: parseFloat(values[2]),
      material,
      kiln,
      heater: parseFloat(values[5]),
      fan: parseFloat(values[6]),
      steam: parseFloat(values[7]),
      kilnPrimary: optionalReading(values[11]),
      kilnSecondary: optionalReading(values[12]),
    });
  }

  return rows;
};

// Groups consecutive rows that share the same step value, preserving order.
export const segmentBySteps = (rows: LogRow[]): StepSegment[] => {
  const segments: StepSegment[] = [];
  for (const row of rows) {
    const current = segments[segments.length - 1];
    if (current && current.step === row.step) {
      current.rows.push(row);
    } else {
      segments.push({ step: row.step, rows: [row] });
    }
  }
  return segments;
};

// Wall clock time of an execution log row, formatted for a chart axis: local
// hours and minutes, no seconds. Runs spanning midnight are left to read as
// "23:00 ... 01:14" -- the wrap is obvious enough without a date on the tick.
export const formatClock = (epochSeconds: number): string =>
  new Date(epochSeconds * 1000).toLocaleTimeString([], {
    hour: "2-digit",
    minute: "2-digit",
  });

// Unix start time of a run, needed to turn the log's elapsed seconds into wall
// clock time. Prefers the timestamp the API reports and falls back to the one
// the controlunit embeds in the run name ("<program>@<RFC3339>"). Undefined
// when neither is available, in which case charts stay on elapsed minutes.
export const runStartedAt = (startedAt?: number, runName?: string): number | undefined => {
  if (startedAt) {
    return startedAt;
  }
  const stamp = runName?.split("@")[1];
  if (!stamp) {
    return undefined;
  }
  const parsed = Date.parse(stamp);
  return isNaN(parsed) ? undefined : parsed / 1000;
};
