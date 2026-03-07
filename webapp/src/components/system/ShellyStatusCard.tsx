import React from "react";
import { Paper, Typography, Box, Chip } from "@mui/material";
import { ShellyStatus } from "../../store/services/systemApi";

interface Props {
  status: ShellyStatus;
}

export const ShellyStatusCard: React.FC<Props> = ({ status }) => {
  return (
    <Paper sx={{ padding: 3, height: "100%" }}>
      <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 2 }}>
        <Typography variant="h6">Shelly Device</Typography>
        <Chip
          label={status.reachable ? "Connected" : "Disconnected"}
          color={status.reachable ? "success" : "error"}
          size="small"
        />
      </Box>

      {status.last_communication && (
        <Box>
          <Typography variant="body2" color="text.secondary">
            Last Communication:
          </Typography>
          <Typography variant="body2" sx={{ marginTop: 0.5 }}>
            {new Date(status.last_communication).toLocaleString()}
          </Typography>
        </Box>
      )}
    </Paper>
  );
};
