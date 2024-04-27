import {
  FormControl,
  FormControlLabel,
  Radio,
  RadioGroup,
  Stack,
  Typography,
} from "@mui/material";
import React from "react";
import { useTranslation } from "react-i18next";
import { ConstantCycle } from "./ConstantCycle";
import { Cycle, Phase } from "../../types/api";

interface Props {
  editing: boolean;
  phase: Phase;
  onChangeCycleMode: (event: React.ChangeEvent<HTMLInputElement>) => void;
  onChangeConstantCycle: (cycle: Cycle) => void;
}

export const Cycles: React.FC<Props> = (props) => {
  const { editing, phase, onChangeCycleMode, onChangeConstantCycle } = props;

  return (
    <Stack gap={3}>
      {editing && (
        <CycleTypeSwitch
          cycleMode={phase.cycleMode}
          onChangeCycleMode={onChangeCycleMode}
        />
      )}

      {phase.cycleMode === "constant" && (
        <ConstantCycle
          editing={editing}
          cycle={phase.constantCycle}
          onChange={onChangeConstantCycle}
        />
      )}
    </Stack>
  );
};

type CycleTypeSwitchProps = { cycleMode: string } & Pick<
  Props,
  "onChangeCycleMode"
>;

const CycleTypeSwitch: React.FC<CycleTypeSwitchProps> = (props) => {
  const { cycleMode, onChangeCycleMode } = props;
  const { t } = useTranslation();

  return (
    <Stack>
      <FormControl>
        <Typography variant="h5" paddingBottom={2}>
          {t("phases.cycles.title")}
        </Typography>

        <RadioGroup
          defaultValue={"constant"}
          value={cycleMode}
          onChange={onChangeCycleMode}
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
    </Stack>
  );
};
