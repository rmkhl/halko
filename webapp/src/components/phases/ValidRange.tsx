import React from "react";
import { Phase } from "../../types/api";
import { useTranslation } from "react-i18next";
import { Stack, Typography } from "@mui/material";
import { TemperatureRangeSlider } from "../form/TemperatureRangeSlider";

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

const ValidRangeComponent: React.FC<ValidRangeComponentProps> = (props) => {
  const { editing, sensor, above, below, onChange } = props;
  const { t } = useTranslation();

  const handleChange = (low: number, high: number) => {
    onChange(sensor, low, high);
  };

  return (
    <TemperatureRangeSlider
      editing={editing}
      title={t(`phases.validRange.${sensor}`) || sensor}
      low={above}
      high={below}
      onChange={handleChange}
    />
  );
};
