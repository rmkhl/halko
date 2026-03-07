import React from "react";
import { Box, CircularProgress, Alert, Typography, Grid } from "@mui/material";
import { useGetSystemStatusQuery, useGetHardwareStatusQuery } from "../store/services/systemApi";
import { SubsystemStatusCard } from "./system/SubsystemStatusCard";
import { ModuleStatusCard } from "./system/ModuleStatusCard";
import { SystemInfoCard } from "./system/SystemInfoCard";
import { VPNCard } from "./system/VPNCard";
import { PowerCard } from "./system/PowerCard";

export const System: React.FC = () => {
  const { data: systemStatus, isLoading, error } = useGetSystemStatusQuery(undefined, {
    pollingInterval: 10000, // Poll every 10 seconds
    skipPollingIfUnfocused: true, // Stop when browser tab inactive
  });

  const { data: hardwareStatus, isLoading: isLoadingHardware } = useGetHardwareStatusQuery(undefined, {
    pollingInterval: 5000, // Poll every 5 seconds for real-time hardware data
    skipPollingIfUnfocused: true,
  });

  if (isLoading) {
    return (
      <Box
        sx={{
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          minHeight: "400px",
          padding: 4,
        }}
      >
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Box sx={{ padding: 4 }}>
        <Alert severity="error">Failed to load system status</Alert>
      </Box>
    );
  }

  if (!systemStatus) {
    return (
      <Box sx={{ padding: 4 }}>
        <Alert severity="warning">No system status data available</Alert>
      </Box>
    );
  }

  return (
    <Box sx={{ padding: 4, width: "100%" }}>
      <Typography variant="h4" sx={{ marginBottom: 3 }}>
        System Status
      </Typography>

      {/* Subsystems & External Modules */}
      <Box
        sx={{
          display: "grid",
          gridTemplateColumns: "repeat(auto-fit, minmax(200px, 1fr))",
          gap: 3,
          marginBottom: 3,
        }}
      >
        <SubsystemStatusCard name="controlunit" isHealthy={systemStatus.services.controlunit.status === "healthy"} />
        <SubsystemStatusCard name="powerunit" isHealthy={systemStatus.services.powerunit.status === "healthy"} />
        <SubsystemStatusCard name="sensorunit" isHealthy={systemStatus.services.sensorunit.status === "healthy"} />
        {hardwareStatus && !isLoadingHardware && (
          <>
            <ModuleStatusCard name="Shelly" isConnected={hardwareStatus.shelly.reachable} />
            <ModuleStatusCard
              name="Arduino"
              isConnected={
                systemStatus.services.sensorunit.details?.arduino_connected === true ||
                systemStatus.services.sensorunit.details?.arduino_connected === "true"
              }
            />
          </>
        )}
      </Box>

      {/* System Resources, VPN, and Power Management */}
      <Grid container spacing={3}>
        <Grid item xs={12} md={4}>
          <SystemInfoCard system={systemStatus.system} />
        </Grid>
        <Grid item xs={12} md={4}>
          <VPNCard />
        </Grid>
        <Grid item xs={12} md={4}>
          <PowerCard uptimeSeconds={systemStatus.system.uptime_seconds} />
        </Grid>
      </Grid>
    </Box>
  );
};
