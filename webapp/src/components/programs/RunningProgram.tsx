import React, { useEffect, useMemo } from "react";
import {
  useGetRunningProgramQuery,
  useStopRunningProgramMutation,
} from "../../store/services/executorApi";
import { Button, Stack, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";
import { useGetTemperaturesQuery } from "../../store/services/sensorsApi";
import { celsius } from "../../util";

interface PowerStatus {
  fan: number;
  heater: number;
  humidifier: number;
}

interface Temperatures {
  delta: number;
  material: number;
  oven: number;
}

interface Program {
  name: string;
}

interface RunningProgram {
  power_status: PowerStatus;
  temperatures: Temperatures;
  program: Program;
}

interface Response<T> {
  data: T;
}

const pollingInterval = 30000;

export const RunningProgram: React.FC = () => {
  const { t } = useTranslation();
  const { data: runningProgramData } = useGetRunningProgramQuery(undefined, {
    pollingInterval,
    skipPollingIfUnfocused: true,
  });

  const { data: sensorData } = useGetTemperaturesQuery(undefined, {
    pollingInterval: 5000,
    skipPollingIfUnfocused: true,
  });

  const [stopProgram] = useStopRunningProgramMutation();

  const runningProgram = useMemo(
    () =>
      runningProgramData
        ? (runningProgramData as Response<RunningProgram>)
        : undefined,
    [runningProgramData]
  );

  const temperatures = useMemo(
    () =>
      sensorData
        ? (sensorData as Response<Omit<Temperatures, "delta">>)
        : undefined,
    [sensorData]
  );

  return (
    <Stack direction="row" justifyContent="space-between" gap={8}>
      {temperatures && (
        <Stack>
          <Stack direction="row" justifyContent="space-between" gap={2}>
            <Typography>{t("sensors.oven")}:</Typography>

            <Typography>{celsius(temperatures.data.oven)}</Typography>
          </Stack>

          <Stack direction="row" justifyContent="space-between" gap={2}>
            <Typography>{t("sensors.material")}:</Typography>

            <Typography>{celsius(temperatures.data.material)}</Typography>
          </Stack>
        </Stack>
      )}

      <Stack direction="row" justifyContent="space-between" gap={4}>
        <Typography>
          {runningProgram
            ? runningProgram.data.program.name
            : t("programs.noRunning")}
        </Typography>

        {runningProgram && (
          <Button onClick={() => stopProgram("")}>Stop</Button>
        )}
      </Stack>
    </Stack>
  );
};
