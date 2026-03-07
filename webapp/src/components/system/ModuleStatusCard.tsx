import React from "react";
import { Paper, Typography, Chip, Box } from "@mui/material";

interface Props {
  name: string;
  isConnected: boolean;
}

export const ModuleStatusCard: React.FC<Props> = ({ name, isConnected }) => {
  return (
    <Paper sx={{ padding: 3, height: "100%" }}>
      <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
        <Typography variant="h6">{name}</Typography>
        <Chip label={isConnected ? "Connected" : "Disconnected"} color={isConnected ? "success" : "error"} size="small" />
      </Box>
    </Paper>
  );
};
