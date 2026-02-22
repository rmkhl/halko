import React, { useMemo } from "react";
import {
  useGetRunningProgramQuery,
  useStopRunningProgramMutation,
} from "../../store/services/controlunitApi";
import { Button, Stack, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";
import { useGetTemperaturesQuery } from "../../store/services/sensorsApi";
import { celsius } from "../../util";
import { RunningProgramResponse, TemperatureStatus, APIResponse, Step } from "../../types/api";

export const RunningProgram: React.FC = () => {
  const { t } = useTranslation();
  const { data: runningProgramData } = useGetRunningProgramQuery(undefined, {
    pollingInterval: 5000,
    skipPollingIfUnfocused: true,
  });
  const { data: sensorData } = useGetTemperaturesQuery(undefined, {
    pollingInterval: 5000,
    skipPollingIfUnfocused: true,
  });
  const [stopProgram] = useStopRunningProgramMutation();

  const runningProgram = useMemo(() => {
    return runningProgramData
      ? (runningProgramData as RunningProgramResponse)
      : undefined;
  }, [runningProgramData]);

  const temperatures = useMemo(() => {
    return sensorData
      ? (sensorData as APIResponse<Omit<TemperatureStatus, "delta">>)
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
          {runningProgram && runningProgram.data?.started_at && (
            <Typography variant="subtitle2" color="text.secondary">
              Started: {new Date(runningProgram.data.started_at * 1000).toLocaleString()}
            </Typography>
          )}
          {runningProgram && runningProgram.data?.current_step ? (
            <>
              <Typography variant="subtitle2" color="text.secondary">
                Step: {runningProgram.data.current_step}
              </Typography>
              {(() => {
                // Find the current step object to get its target temperature
                const executionStatus = runningProgram.data;
                const stepName = executionStatus.current_step;
                const steps = executionStatus.program?.steps;
                if (Array.isArray(steps)) {
                  const step = steps.find((s: Step) => s.name === stepName);
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
              {runningProgram.data.current_step_started_at && (
                <Typography variant="subtitle2" color="text.secondary">
                  Step started: {new Date(runningProgram.data.current_step_started_at * 1000).toLocaleString()}
                </Typography>
              )}
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
