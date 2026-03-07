import React, { useState } from "react";
import {
  Card,
  CardContent,
  Typography,
  Button,
  Box,
  Alert,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  DialogActions,
  CircularProgress,
} from "@mui/material";
import PowerSettingsNewIcon from "@mui/icons-material/PowerSettingsNew";
import RestartAltIcon from "@mui/icons-material/RestartAlt";
import { useShutdownSystemMutation, useRebootSystemMutation } from "../../store/services/dbusunitApi";

interface PowerCardProps {
  uptimeSeconds: number;
}

export const PowerCard: React.FC<PowerCardProps> = ({ uptimeSeconds }) => {
  const [shutdownSystem, { isLoading: isShuttingDown, error: shutdownError }] = useShutdownSystemMutation();
  const [rebootSystem, { isLoading: isRebooting, error: rebootError }] = useRebootSystemMutation();

  const [shutdownDialogOpen, setShutdownDialogOpen] = useState(false);
  const [rebootDialogOpen, setRebootDialogOpen] = useState(false);

  // Format uptime as "X days, Y hours, Z minutes"
  const formatUptime = (seconds: number): string => {
    const days = Math.floor(seconds / 86400);
    const hours = Math.floor((seconds % 86400) / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);

    const parts: string[] = [];
    if (days > 0) parts.push(`${days} day${days !== 1 ? "s" : ""}`);
    if (hours > 0) parts.push(`${hours} hour${hours !== 1 ? "s" : ""}`);
    if (minutes > 0 || parts.length === 0) parts.push(`${minutes} minute${minutes !== 1 ? "s" : ""}`);

    return parts.join(", ");
  };

  const handleShutdownClick = () => {
    setShutdownDialogOpen(true);
  };

  const handleRebootClick = () => {
    setRebootDialogOpen(true);
  };

  const handleShutdownConfirm = async () => {
    try {
      await shutdownSystem().unwrap();
      setShutdownDialogOpen(false);
    } catch (error) {
      console.error("Shutdown failed:", error);
    }
  };

  const handleRebootConfirm = async () => {
    try {
      await rebootSystem().unwrap();
      setRebootDialogOpen(false);
    } catch (error) {
      console.error("Reboot failed:", error);
    }
  };

  const handleDialogClose = () => {
    setShutdownDialogOpen(false);
    setRebootDialogOpen(false);
  };

  return (
    <>
      <Card sx={{ height: "100%" }}>
        <CardContent>
          <Typography variant="h6" gutterBottom>
            Power Management
          </Typography>

          <Box sx={{ marginBottom: 3 }}>
            <Typography variant="body2" color="text.secondary" gutterBottom>
              System Uptime
            </Typography>
            <Typography variant="h6">{formatUptime(uptimeSeconds)}</Typography>
          </Box>

          {(shutdownError || rebootError) && (
            <Alert severity="error" sx={{ marginBottom: 2 }}>
              {shutdownError ? "Shutdown failed" : "Reboot failed"}
            </Alert>
          )}

          <Box sx={{ display: "flex", gap: 2, flexDirection: "column" }}>
            <Button
              variant="outlined"
              color="warning"
              startIcon={isRebooting ? <CircularProgress size={20} /> : <RestartAltIcon />}
              onClick={handleRebootClick}
              disabled={isRebooting || isShuttingDown}
              fullWidth
            >
              {isRebooting ? "Rebooting..." : "Reboot System"}
            </Button>

            <Button
              variant="outlined"
              color="error"
              startIcon={isShuttingDown ? <CircularProgress size={20} /> : <PowerSettingsNewIcon />}
              onClick={handleShutdownClick}
              disabled={isShuttingDown || isRebooting}
              fullWidth
            >
              {isShuttingDown ? "Shutting Down..." : "Shutdown System"}
            </Button>
          </Box>
        </CardContent>
      </Card>

      {/* Shutdown Confirmation Dialog */}
      <Dialog open={shutdownDialogOpen} onClose={handleDialogClose}>
        <DialogTitle>Confirm Shutdown</DialogTitle>
        <DialogContent>
          <DialogContentText>
            Are you sure you want to shutdown the system? This will stop all running programs and services.
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleDialogClose} color="primary">
            Cancel
          </Button>
          <Button onClick={handleShutdownConfirm} color="error" variant="contained">
            Shutdown
          </Button>
        </DialogActions>
      </Dialog>

      {/* Reboot Confirmation Dialog */}
      <Dialog open={rebootDialogOpen} onClose={handleDialogClose}>
        <DialogTitle>Confirm Reboot</DialogTitle>
        <DialogContent>
          <DialogContentText>
            Are you sure you want to reboot the system? This will stop all running programs and services.
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleDialogClose} color="primary">
            Cancel
          </Button>
          <Button onClick={handleRebootConfirm} color="warning" variant="contained">
            Reboot
          </Button>
        </DialogActions>
      </Dialog>
    </>
  );
};
