import {
  FormControl,
  FormControlLabel,
  Radio,
  RadioGroup,
  Typography,
} from "@mui/material";
import React from "react";
import { useTranslation } from "react-i18next";

interface Props {
  editing: boolean;
  cycleMode?: string;
  onChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
}

export const Cycles: React.FC<Props> = (props) => {
  const { editing, cycleMode = "constant", onChange } = props;
  const { t } = useTranslation();

  return (
    <FormControl>
      <Typography variant="h5" paddingBottom={2}>
        {t("phases.cycles.title")}
      </Typography>

      <RadioGroup
        defaultValue={"constant"}
        value={cycleMode}
        onChange={onChange}
      >
        <FormControlLabel
          value="constant"
          control={<Radio />}
          label={t("phases.cycles.constant")}
        />

        <FormControlLabel
          value="delta"
          control={<Radio />}
          label={t("phases.cycles.delta")}
        />
      </RadioGroup>
    </FormControl>
  );
};
