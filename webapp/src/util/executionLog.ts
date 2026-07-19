// Parsing and step-grouping of controlunit execution logs (CSV format:
// time,step,steptime,material,kiln,heater,fan,humidifier).

export interface LogRow {
  time: number;
  step: string;
  steptime: number;
  material: number;
  kiln: number;
  heater: number;
  fan: number;
  humidifier: number;
}

export interface StepSegment {
  step: string;
  rows: LogRow[];
}

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
      humidifier: parseFloat(values[7]),
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
