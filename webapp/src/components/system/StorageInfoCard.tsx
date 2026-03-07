import React from "react";
import { Paper, Typography, Box, List, ListItem, ListItemText } from "@mui/material";
import { StorageStatus } from "../../store/services/systemApi";

interface Props {
  storage: StorageStatus;
}

export const StorageInfoCard: React.FC<Props> = ({ storage }) => {
  return (
    <Paper sx={{ padding: 3, height: "100%" }}>
      <Typography variant="h6" gutterBottom>
        Storage
      </Typography>

      <Typography variant="subtitle2" color="text.secondary" sx={{ marginBottom: 1 }}>
        Programs:
      </Typography>

      <List dense disablePadding>
        <ListItem disablePadding sx={{ paddingY: 0.5 }}>
          <ListItemText
            primary={
              <Box sx={{ display: "flex", justifyContent: "space-between" }}>
                <Typography variant="body2" color="text.secondary">
                  Stored:
                </Typography>
                <Typography variant="body2">{storage.stored_programs}</Typography>
              </Box>
            }
          />
        </ListItem>

        <ListItem disablePadding sx={{ paddingY: 0.5 }}>
          <ListItemText
            primary={
              <Box sx={{ display: "flex", justifyContent: "space-between" }}>
                <Typography variant="body2" color="text.secondary">
                  Running:
                </Typography>
                <Typography variant="body2">{storage.running_programs}</Typography>
              </Box>
            }
          />
        </ListItem>

        <ListItem disablePadding sx={{ paddingY: 0.5 }}>
          <ListItemText
            primary={
              <Box sx={{ display: "flex", justifyContent: "space-between" }}>
                <Typography variant="body2" color="text.secondary">
                  History:
                </Typography>
                <Typography variant="body2">{storage.history_count}</Typography>
              </Box>
            }
          />
        </ListItem>
      </List>

      <List dense disablePadding sx={{ marginTop: 2 }}>
        <ListItem disablePadding sx={{ paddingY: 0.5 }}>
          <ListItemText
            primary={
              <Box sx={{ display: "flex", justifyContent: "space-between" }}>
                <Typography variant="body2" color="text.secondary">
                  Available Space:
                </Typography>
                <Typography variant="body2">{storage.disk_space_mb} MB</Typography>
              </Box>
            }
          />
        </ListItem>
      </List>
    </Paper>
  );
};
