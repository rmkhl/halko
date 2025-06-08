import React, { useEffect, useState } from "react";
import {
  NumberComponent,
  Props as NumberComponentProps,
} from "./NumberComponent";
import { Stack, Typography } from "@mui/material";
import { v4 as uuidv4 } from "uuid";

export const TimeComponent: React.FC<NumberComponentProps> = (props) => {
  const { editing, title, value = 0, onChange, ...rest } = props;
  const [timeParts, setTimeParts] = useState<TimeMap>();

  const handleChange = (field: Field) => (val: number) => {
    let tp = timeParts;
    if (tp === undefined) {
      tp = { h: 0, m: 0 };
    }
    setTimeParts({ ...tp, [field]: val });
  };

  useEffect(() => {
    const tp = Object.keys(timeDefs).reduce(
      (acc: TimeMap, field: string) => {
        const timeDef = timeDefs[field as Field];
        let val = value % timeDef.modulo;
        val = val / timeDef.multiplier;
        val = Math.floor(val);
        acc[field as Field] = val;
        return acc;
      },
      { h: 0, m: 0 } as TimeMap
    );
    setTimeParts(tp);
  }, []);

  useEffect(() => {
    if (timeParts === undefined) return;
    const newVal = Object.entries(timeParts).reduce(
      (acc: number, [field, val]: [string, number]) => {
        acc += timeDefs[field as Field].multiplier * val;
        return acc;
      },
      0
    );
    onChange(newVal);
  }, [timeParts]);

  return (
    <Stack
      direction="row"
      justifyContent="space-between"
      alignItems="center"
      width="100%"
    >
      <Typography variant="h6">{title}:</Typography>

      <Stack direction="row" alignItems="center" gap={3}>
        {Fields.map((f) => {
          const timeDef = timeDefs[f];
          const { unit, max } = timeDef;
          const timeVal = timeParts?.[f] ?? 0;

          return (
            <Time
              key={`timepart-${uuidv4()}`}
              value={timeVal}
              field={f}
              unit={unit}
              max={max}
              onChange={handleChange(f)}
              editing={editing}
            />
          );
        })}
      </Stack>
    </Stack>
  );
};

interface TimeProps {
  value: number;
  field: Field;
  unit: string;
  max: number;
  onChange: (value: number) => void;
  editing?: boolean;
}

const Time: React.FC<TimeProps> = (props) => {
  const { value, unit, max, onChange, editing } = props;
  return (
    <NumberComponent
      value={value}
      onChange={onChange}
      max={max}
      min={0}
      editing={editing}
    >
      {unit}
    </NumberComponent>
  );
};

const Fields = ["h", "m"] as const;
export type Field = (typeof Fields)[number];

interface TimeDef {
  unit: string;
  max: number;
  multiplier: number;
  modulo: number;
}

const hourSeconds = 3600;

const timeDefs: Record<Field, TimeDef> = {
  h: {
    unit: "hours",
    max: 240,
    multiplier: hourSeconds,
    modulo: Number.MAX_SAFE_INTEGER,
  },
  m: {
    unit: "minutes",
    max: 59,
    multiplier: 60,
    modulo: hourSeconds,
  },
};

type TimeMap = Record<Field, number>;
