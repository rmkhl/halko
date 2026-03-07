import React from "react";
import { Paper, Typography, Box, List, ListItem, ListItemText } from "@mui/material";
import { SystemInfo } from "../../store/services/systemApi";

interface Props {
  system: SystemInfo;
}

const formatMemory = (usedMB: number, totalMB: number): string => {
  if (totalMB === 0) return "N/A";
  return `${usedMB} / ${totalMB} MB`;
};

export const SystemInfoCard: React.FC<Props> = ({ system }) => {
  return (
    <Paper sx={{ padding: 3, height: "100%" }}>
      <Typography variant="h6" gutterBottom>
        System Resources
      </Typography>

      <List dense disablePadding>
        <ListItem disablePadding sx={{ paddingY: 0.5 }}>
          <ListItemText
            primary={
              <Box sx={{ display: "flex", justifyContent: "space-between" }}>
                <Typography variant="body2" color="text.secondary">
                  Memory:
                </Typography>
                <Typography variant="body2">
                  {formatMemory(system.memory_used_mb, system.memory_total_mb)}
                </Typography>
              </Box>
            }
          />
        </ListItem>

        <ListItem disablePadding sx={{ paddingY: 0.5 }}>
          <ListItemText
            primary={
              <Box sx={{ display: "flex", justifyContent: "space-between" }}>
                <Typography variant="body2" color="text.secondary">
                  Swap:
                </Typography>
                <Typography variant="body2">
                  {formatMemory(system.swap_used_mb, system.swap_total_mb)}
                </Typography>
              </Box>
            }
          />
        </ListItem>

        <ListItem disablePadding sx={{ paddingY: 0.5 }}>
          <ListItemText
            primary={
              <Box sx={{ display: "flex", justifyContent: "space-between" }}>
                <Typography variant="body2" color="text.secondary">
                  Available Space:
                </Typography>
                <Typography variant="body2">{system.disk_space_mb} MB</Typography>
              </Box>
            }
          />
        </ListItem>
      </List>
    </Paper>
  );
};
