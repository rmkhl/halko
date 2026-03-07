import React from "react";
import { Paper, Typography, Chip, Box } from "@mui/material";

interface Props {
  name: string;
  isHealthy: boolean;
}

const formatName = (name: string): string => {
  // Convert snake_case to Title Case
  return name
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
};

export const SubsystemStatusCard: React.FC<Props> = ({ name, isHealthy }) => {
  return (
    <Paper sx={{ padding: 3, height: "100%" }}>
      <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center" }}>
        <Typography variant="h6">{formatName(name)}</Typography>
        <Chip label={isHealthy ? "Healthy" : "Unavailable"} color={isHealthy ? "success" : "error"} size="small" />
      </Box>
    </Paper>
  );
};
