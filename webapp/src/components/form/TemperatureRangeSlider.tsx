import { Slider, Typography } from "@mui/material";
import React from "react";
import { celsius, celsiusRange } from "../../util";

interface Props {
  editing?: boolean;
  title: string;
  low: number;
  high: number;
  onChange: (low: number, high: number) => void;
}

export const TemperatureRangeSlider: React.FC<Props> = (props) => {
  const { editing, title, low, high, onChange } = props;

  const handleChange = (
    _: Event,
    newValue: number | number[],
    activeThumb: number
  ) => {
    if (!Array.isArray(newValue)) {
      return;
    }

    let [newLow, newHigh] = newValue as number[];
    const aboveChanged = activeThumb === 0;

    if (aboveChanged) {
      newLow = Math.min(newLow, high - minDistance);
    } else {
      newHigh = Math.max(newHigh, low + minDistance);
    }

    onChange(newLow, newHigh);
  };

  return (
    <>
      <Typography>
        {title}: {celsiusRange(low, high)}
      </Typography>

      <Slider
        value={[low, high]}
        step={1}
        getAriaValueText={celsius}
        marks={rangeMarks}
        max={200}
        min={0}
        valueLabelDisplay="auto"
        onChange={handleChange}
        disableSwap
        disabled={!editing}
      />
    </>
  );
};

const rangeMarks = [
  {
    value: 0,
    label: "0°C",
  },
  {
    value: 25,
    label: "25°C",
  },
  {
    value: 50,
    label: "50°C",
  },
  {
    value: 75,
    label: "75°C",
  },
  {
    value: 100,
    label: "100°C",
  },
  {
    value: 125,
    label: "125°C",
  },
  {
    value: 150,
    label: "150°C",
  },
  {
    value: 175,
    label: "175°C",
  },
  {
    value: 200,
    label: "200°C",
  },
];

const minDistance = 5;
