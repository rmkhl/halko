import React from "react";
import { Paper, Typography, Chip, Box, List, ListItem, ListItemText } from "@mui/material";
import { ServiceStatus } from "../../store/services/systemApi";

interface Props {
  serviceName: string;
  status: ServiceStatus;
}

const getStatusColor = (status: string): "success" | "error" | "warning" | "default" => {
  switch (status) {
    case "healthy":
      return "success";
    case "unavailable":
      return "error";
    case "degraded":
      return "warning";
    default:
      return "default";
  }
};

const formatDetailValue = (value: unknown): string => {
  if (typeof value === "boolean") {
    return value ? "Yes" : "No";
  }
  if (typeof value === "number") {
    return value.toString();
  }
  if (typeof value === "string") {
    return value;
  }
  return JSON.stringify(value);
};

const formatDetailKey = (key: string): string => {
  // Convert snake_case to Title Case
  return key
    .split("_")
    .map((word) => word.charAt(0).toUpperCase() + word.slice(1))
    .join(" ");
};

export const ServiceHealthCard: React.FC<Props> = ({ serviceName, status }) => {
  return (
    <Paper sx={{ padding: 3, height: "100%" }}>
      <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 2 }}>
        <Typography variant="h6">{formatDetailKey(serviceName)}</Typography>
        <Chip label={status.status} color={getStatusColor(status.status)} size="small" />
      </Box>

      {status.details && Object.keys(status.details).length > 0 && (
        <>
          <List dense disablePadding>
            {Object.entries(status.details).map(([key, value]) => (
              <ListItem key={key} disablePadding sx={{ paddingY: 0.5 }}>
                <ListItemText
                  primary={
                    <Box sx={{ display: "flex", justifyContent: "space-between" }}>
                      <Typography variant="body2" color="text.secondary">
                        {formatDetailKey(key)}:
                      </Typography>
                      <Typography variant="body2">{formatDetailValue(value)}</Typography>
                    </Box>
                  }
                />
              </ListItem>
            ))}
          </List>
        </>
      )}
    </Paper>
  );
};
