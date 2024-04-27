import React from "react";
import { Phase } from "../../types/api";
import { useTranslation } from "react-i18next";
import { Slider, Stack, Typography } from "@mui/material";

interface Props {
  editing: boolean;
  validRange: Phase["validRange"];
  handleChange: (validRange: Phase["validRange"]) => void;
}

export const ValidRange: React.FC<Props> = (props) => {
  const { editing, validRange, handleChange } = props;
  const { t } = useTranslation();

  const handleVrChange = (sensor: string, above: number, below: number) => {
    const updated: Phase["validRange"] = validRange?.map((vr) =>
      vr.sensor === sensor ? { sensor, above, below } : { ...vr }
    );

    handleChange(updated);
  };

  return (
    <Stack gap={3}>
      <Typography variant="h5">{t("phases.validRange.title")}</Typography>

      {validRange?.map((vr, i) => (
        <ValidRangeComponent
          key={vr.sensor}
          editing={editing}
          sensor={vr.sensor}
          above={vr.above}
          below={vr.below}
          onChange={handleVrChange}
        />
      ))}
    </Stack>
  );
};

interface ValidRangeComponentProps {
  editing: boolean;
  sensor: string;
  above: number;
  below: number;
  onChange: (sensor: string, above: number, below: number) => void;
}

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

const minDistance = 1;

const ValidRangeComponent: React.FC<ValidRangeComponentProps> = (props) => {
  const { editing, sensor, above, below, onChange } = props;
  const { t } = useTranslation();

  const celsius = (value: number) => {
    return `${value}°C`;
  };

  const handleChange = (
    event: Event,
    newValue: number | number[],
    activeThumb: number
  ) => {
    if (!Array.isArray(newValue)) {
      return;
    }

    let [newAbove, newBelow] = newValue as number[];
    const aboveChanged = activeThumb === 0;

    if (aboveChanged) {
      newAbove = Math.min(newAbove, below - minDistance);
    } else {
      newBelow = Math.max(newBelow, above + minDistance);
    }

    onChange(sensor, newAbove, newBelow);
  };

  return (
    <>
      <Typography>
        {t(`phases.validRange.${sensor}`) || sensor}: {above}-{below}°C
      </Typography>

      <Slider
        value={[above, below]}
        step={1}
        getAriaValueText={celsius}
        marks={rangeMarks}
        max={200}
        min={0}
        valueLabelDisplay="auto"
        onChange={handleChange}
        disableSwap
      />
    </>
  );
};
