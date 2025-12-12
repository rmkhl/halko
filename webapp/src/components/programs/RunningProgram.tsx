import React, { useMemo } from "react";
import {
  useGetRunningProgramQuery,
  useStopRunningProgramMutation,
} from "../../store/services/controlunitApi";
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

export const RunningProgram: React.FC = () => {
  const { t } = useTranslation();
  const { data: runningProgramData } = useGetRunningProgramQuery(undefined, {
    pollingInterval: 30000,
    skipPollingIfUnfocused: true,
  });
  const { data: sensorData } = useGetTemperaturesQuery(undefined, {
    pollingInterval: 5000,
    skipPollingIfUnfocused: true,
  });
  const [stopProgram] = useStopRunningProgramMutation();

  const runningProgram = useMemo(() => {
    return runningProgramData
      ? (runningProgramData as Response<RunningProgram>)
      : undefined;
  }, [runningProgramData]);

  const temperatures = useMemo(() => {
    return sensorData
      ? (sensorData as Response<Omit<Temperatures, "delta">>)
      : undefined;
  }, [sensorData]);

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
        <Stack>
          <Typography>
            {runningProgram
              ? runningProgram.data.program.name
              : t("programs.noRunning")}
          </Typography>
          {runningProgram && (runningProgramData as any)?.data?.current_step ? (
            <>
              <Typography variant="subtitle2" color="text.secondary">
                Step: {(runningProgramData as any).data.current_step}
              </Typography>
              {(() => {
                // Find the current step object to get its target temperature
                const stepName = (runningProgramData as any).data.current_step;
                const steps = (runningProgramData as any).data.program?.steps;
                if (Array.isArray(steps)) {
                  const step = steps.find((s: any) => s.name === stepName);
                  if (step && step.temperature_target !== undefined) {
                    return (
                      <Typography variant="subtitle2" color="text.secondary">
                        Target: {step.temperature_target}°C
                      </Typography>
                    );
                  }
                }
                return null;
              })()}
            </>
          ) : null}
        </Stack>
        {runningProgram && (
          <Button onClick={() => stopProgram("")}>Stop</Button>
        )}
      </Stack>
    </Stack>
  );
};
