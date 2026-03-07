import React from "react";
import { Paper, Typography, Box, List, ListItem, ListItemText } from "@mui/material";
import { HardwareStatusResponse } from "../../store/services/systemApi";

interface Props {
  hardware: HardwareStatusResponse;
}

const celsius = (temp: number): string => {
  return temp > -200 ? `${temp.toFixed(1)}°C` : "N/A";
};

export const HardwareStatusCard: React.FC<Props> = ({ hardware }) => {
  return (
    <Paper sx={{ padding: 3 }}>
      <Typography variant="h6" gutterBottom>
        Hardware Status
      </Typography>

      <Box sx={{ display: "grid", gridTemplateColumns: "repeat(auto-fit, minmax(250px, 1fr))", gap: 3 }}>
        {/* Temperatures Section */}
        <Box>
          <Typography variant="subtitle2" color="text.secondary" sx={{ marginBottom: 1 }}>
            Temperatures:
          </Typography>
          <List dense disablePadding>
            <ListItem disablePadding sx={{ paddingY: 0.5 }}>
              <ListItemText
                primary={
                  <Box sx={{ display: "flex", justifyContent: "space-between" }}>
                    <Typography variant="body2" color="text.secondary">
                      Wood:
                    </Typography>
                    <Typography variant="body2">{celsius(hardware.temperatures.material)}</Typography>
                  </Box>
                }
              />
            </ListItem>

            <ListItem disablePadding sx={{ paddingY: 0.5 }}>
              <ListItemText
                primary={
                  <Box sx={{ display: "flex", justifyContent: "space-between" }}>
                    <Typography variant="body2" color="text.secondary">
                      Oven:
                    </Typography>
                    <Typography variant="body2">{celsius(hardware.temperatures.oven)}</Typography>
                  </Box>
                }
              />
            </ListItem>

            <ListItem disablePadding sx={{ paddingY: 0.5, paddingLeft: 2 }}>
              <ListItemText
                primary={
                  <Box sx={{ display: "flex", justifyContent: "space-between" }}>
                    <Typography variant="body2" color="text.secondary" sx={{ fontSize: "0.85rem" }}>
                      Primary:
                    </Typography>
                    <Typography variant="body2" sx={{ fontSize: "0.85rem" }}>
                      {celsius(hardware.temperatures.oven_primary)}
                    </Typography>
                  </Box>
                }
              />
            </ListItem>

            <ListItem disablePadding sx={{ paddingY: 0.5, paddingLeft: 2 }}>
              <ListItemText
                primary={
                  <Box sx={{ display: "flex", justifyContent: "space-between" }}>
                    <Typography variant="body2" color="text.secondary" sx={{ fontSize: "0.85rem" }}>
                      Secondary:
                    </Typography>
                    <Typography variant="body2" sx={{ fontSize: "0.85rem" }}>
                      {celsius(hardware.temperatures.oven_secondary)}
                    </Typography>
                  </Box>
                }
              />
            </ListItem>
          </List>
        </Box>

        {/* Power Levels Section */}
        <Box>
          <Typography variant="subtitle2" color="text.secondary" sx={{ marginBottom: 1 }}>
            Power Levels:
          </Typography>
          <List dense disablePadding>
            <ListItem disablePadding sx={{ paddingY: 0.5 }}>
              <ListItemText
                primary={
                  <Box sx={{ display: "flex", justifyContent: "space-between" }}>
                    <Typography variant="body2" color="text.secondary">
                      Heater:
                    </Typography>
                    <Typography variant="body2">{hardware.power.heater}%</Typography>
                  </Box>
                }
              />
            </ListItem>

            <ListItem disablePadding sx={{ paddingY: 0.5 }}>
              <ListItemText
                primary={
                  <Box sx={{ display: "flex", justifyContent: "space-between" }}>
                    <Typography variant="body2" color="text.secondary">
                      Fan:
                    </Typography>
                    <Typography variant="body2">{hardware.power.fan}%</Typography>
                  </Box>
                }
              />
            </ListItem>

            <ListItem disablePadding sx={{ paddingY: 0.5 }}>
              <ListItemText
                primary={
                  <Box sx={{ display: "flex", justifyContent: "space-between" }}>
                    <Typography variant="body2" color="text.secondary">
                      Humidifier:
                    </Typography>
                    <Typography variant="body2">{hardware.power.humidifier}%</Typography>
                  </Box>
                }
              />
            </ListItem>
          </List>
        </Box>
      </Box>
    </Paper>
  );
};
