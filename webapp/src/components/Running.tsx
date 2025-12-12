import React from "react";
import { Box } from "@mui/material";
import { RunningProgram } from "./programs/RunningProgram";
import { ExecutionChart } from "./ExecutionChart";

export const Running: React.FC = () => {
  return (
    <Box
      sx={{
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        width: "100%",
        padding: 4,
      }}
    >
      <RunningProgram />
      <ExecutionChart title="Sample Execution Data" />
    </Box>
  );
};
