import React, { useMemo, useState, useEffect } from "react";
import {
  useGetRunningProgramQuery,
  useStopRunningProgramMutation,
} from "../../store/services/controlunitApi";
import { Button, Stack, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";
import { useGetTemperaturesQuery } from "../../store/services/sensorsApi";
import { useGetPowerStatusQuery } from "../../store/services/powerunitApi";
import { celsius } from "../../util";
import { RunningProgramResponse, TemperatureStatus, APIResponse, Step } from "../../types/api";

// Format duration in seconds to human-readable string
const formatDuration = (seconds: number): string => {
  const hours = Math.floor(seconds / 3600);
  const minutes = Math.floor((seconds % 3600) / 60);
  const secs = seconds % 60;

  if (hours > 0) {
    return `${hours}h ${minutes}m`;
  } else if (minutes > 0) {
    return `${minutes}m ${secs}s`;
  } else {
    return `${secs}s`;
  }
};

export const RunningProgram: React.FC = () => {
  const { t } = useTranslation();
  const [currentTime, setCurrentTime] = useState(() => Math.floor(Date.now() / 1000));
  const [isStoppingLocally, setIsStoppingLocally] = useState(false);

  // Update current time every second for duration display
  useEffect(() => {
    const interval = setInterval(() => {
      setCurrentTime(Math.floor(Date.now() / 1000));
    }, 1000);
    return () => clearInterval(interval);
  }, []);

  const { data: runningProgramData, refetch: refetchRunningProgram } = useGetRunningProgramQuery(undefined, {
    pollingInterval: 5000,
    skipPollingIfUnfocused: true,
  });
  const { data: sensorData } = useGetTemperaturesQuery(undefined, {
    pollingInterval: 5000,
    skipPollingIfUnfocused: true,
  });
  const { data: powerData } = useGetPowerStatusQuery(undefined, {
    pollingInterval: 5000,
    skipPollingIfUnfocused: true,
  });
  const [stopProgram, { isLoading: isStopping }] = useStopRunningProgramMutation();

  const handleStop = async () => {
    try {
      setIsStoppingLocally(true);
      await stopProgram("").unwrap();
      // Force immediate refetch after stop to clear stale data
      refetchRunningProgram();
    } catch (error) {
      console.error("Failed to stop program:", error);
      setIsStoppingLocally(false);
    }
  };

  const runningProgram = useMemo(() => {
    return runningProgramData
      ? (runningProgramData as RunningProgramResponse)
      : undefined;
  }, [runningProgramData]);

  // Reset stopping state when program actually stops
  useEffect(() => {
    if (isStoppingLocally && !runningProgram) {
      setIsStoppingLocally(false);
    }
  }, [isStoppingLocally, runningProgram]);

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
          {powerData?.data && (
            <>
              <Stack direction="row" justifyContent="space-between" gap={2} sx={{ mt: 1 }}>
                <Typography>Heater:</Typography>
                <Typography>{powerData.data.heater?.percent ?? 0}%</Typography>
              </Stack>
              <Stack direction="row" justifyContent="space-between" gap={2}>
                <Typography>Fan:</Typography>
                <Typography>{powerData.data.fan?.percent ?? 0}%</Typography>
              </Stack>
              <Stack direction="row" justifyContent="space-between" gap={2}>
                <Typography>Humidifier:</Typography>
                <Typography>{powerData.data.humidifier?.percent ?? 0}%</Typography>
              </Stack>
            </>
          )}
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
              Started: {new Date(runningProgram.data.started_at * 1000).toLocaleString()} ({formatDuration(currentTime - runningProgram.data.started_at)})
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
                  Step started: {new Date(runningProgram.data.current_step_started_at * 1000).toLocaleString()} ({formatDuration(currentTime - runningProgram.data.current_step_started_at)})
                </Typography>
              )}
            </>
          ) : null}
        </Stack>
        {runningProgram && (
          <Button onClick={handleStop} disabled={isStopping || isStoppingLocally}>
            {(isStopping || isStoppingLocally) ? "Stopping..." : "Stop"}
          </Button>
        )}
      </Stack>
    </Stack>
  );
};
