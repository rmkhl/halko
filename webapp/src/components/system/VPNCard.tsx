import React from "react";
import {
  Paper,
  Typography,
  Box,
  List,
  ListItem,
  ListItemText,
  Chip,
  Button,
  CircularProgress,
  Alert,
  Stack,
} from "@mui/material";
import { useGetVPNListQuery, useStartVPNMutation, useStopVPNMutation } from "../../store/services/dbusunitApi";

export const VPNCard: React.FC = () => {
  const { data: vpnData, isLoading, error } = useGetVPNListQuery(undefined, {
    pollingInterval: 5000, // Poll every 5 seconds
    skipPollingIfUnfocused: true,
  });

  const [startVPN, { isLoading: isStarting }] = useStartVPNMutation();
  const [stopVPN, { isLoading: isStopping }] = useStopVPNMutation();

  const handleStart = async (name: string) => {
    try {
      await startVPN(name).unwrap();
    } catch (err) {
      console.error("Failed to start VPN:", err);
    }
  };

  const handleStop = async (name: string) => {
    try {
      await stopVPN(name).unwrap();
    } catch (err) {
      console.error("Failed to stop VPN:", err);
    }
  };

  const getStatusColor = (status: string): "success" | "error" | "warning" | "default" => {
    switch (status) {
      case "active":
        return "success";
      case "failed":
        return "error";
      case "inactive":
        return "default";
      default:
        return "warning";
    }
  };

  if (isLoading) {
    return (
      <Paper sx={{ padding: 3 }}>
        <Typography variant="h6" sx={{ marginBottom: 2 }}>
          VPN Connections
        </Typography>
        <Box sx={{ display: "flex", justifyContent: "center", padding: 2 }}>
          <CircularProgress size={24} />
        </Box>
      </Paper>
    );
  }

  if (error) {
    return (
      <Paper sx={{ padding: 3 }}>
        <Typography variant="h6" sx={{ marginBottom: 2 }}>
          VPN Connections
        </Typography>
        <Alert severity="error">Failed to load VPN status</Alert>
      </Paper>
    );
  }

  const vpns = vpnData?.data || [];

  return (
    <Paper sx={{ padding: 3 }}>
      <Typography variant="h6" sx={{ marginBottom: 2 }}>
        VPN Connections
      </Typography>

      {vpns.length === 0 ? (
        <Alert severity="info">No VPN connections configured</Alert>
      ) : (
        <List disablePadding>
          {vpns.map((vpn) => (
            <ListItem
              key={vpn.name}
              disablePadding
              sx={{
                paddingY: 1.5,
                borderBottom: "1px solid",
                borderColor: "divider",
                "&:last-child": { borderBottom: "none" },
              }}
            >
              <ListItemText
                primary={
                  <Stack direction="row" alignItems="center" spacing={1}>
                    <Typography variant="body1">{vpn.name}</Typography>
                    <Chip
                      label={vpn.status}
                      color={getStatusColor(vpn.status)}
                      size="small"
                      sx={{ textTransform: "capitalize" }}
                    />
                    {vpn.enabled && (
                      <Chip
                        label="Auto-start"
                        size="small"
                        variant="outlined"
                        sx={{ fontSize: "0.7rem" }}
                      />
                    )}
                  </Stack>
                }
                secondary={
                  vpn.tunnel_ip ? (
                    <Typography variant="caption" color="text.secondary">
                      Tunnel IP: {vpn.tunnel_ip}
                    </Typography>
                  ) : null
                }
              />
              <Stack direction="row" spacing={1}>
                {vpn.status === "active" ? (
                  <Button
                    variant="outlined"
                    color="error"
                    size="small"
                    onClick={() => handleStop(vpn.name)}
                    disabled={isStopping}
                  >
                    {isStopping ? <CircularProgress size={16} /> : "Stop"}
                  </Button>
                ) : (
                  <Button
                    variant="contained"
                    color="primary"
                    size="small"
                    onClick={() => handleStart(vpn.name)}
                    disabled={isStarting}
                  >
                    {isStarting ? <CircularProgress size={16} /> : "Start"}
                  </Button>
                )}
              </Stack>
            </ListItem>
          ))}
        </List>
      )}
    </Paper>
  );
};
