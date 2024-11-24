import { Slider, Typography } from "@mui/material";
import React from "react";
import { celsius } from "../../util";

interface Props {
  editing?: boolean;
  title: string;
  value: number;
  onChange: (value: number) => void;
  range?: [number, number];
}

export const TemperatureRangeSlider: React.FC<Props> = (props) => {
  const { editing, title, value, onChange, range = [100, 250] } = props;
  const [min, max] = range;

  const rangeMarks = new Array(Math.floor((max - min) / 25))
    .fill(min)
    .map((v, i) => {
      const value = v + i * 25;
      return { value, label: celsius(value) };
    });

  const handleChange = (
    _: Event,
    newValue: number | number[],
    activeThumb: number
  ) => {
    if (Array.isArray(newValue)) {
      return;
    }

    let newVal = newValue as number;

    onChange(newVal);
  };

  return (
    <>
      <Typography>
        {title}: {celsius(value)}
      </Typography>

      <Slider
        value={value}
        step={1}
        getAriaValueText={celsius}
        marks={rangeMarks}
        max={max}
        min={min}
        valueLabelDisplay="auto"
        onChange={handleChange}
        disableSwap
        disabled={!editing}
      />
    </>
  );
};
