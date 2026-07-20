import React, { useState } from "react";
import {
  Box,
  Typography,
  List,
  ListItem,
  ListItemButton,
  ListItemText,
  Paper,
  Chip,
  CircularProgress,
  Alert,
  IconButton,
  Divider,
} from "@mui/material";
import DeleteIcon from "@mui/icons-material/Delete";
import FileDownloadIcon from "@mui/icons-material/FileDownload";
import PictureAsPdfIcon from "@mui/icons-material/PictureAsPdf";
import { useGetExecutionHistoryQuery, useGetExecutionLogQuery, useDeleteExecutionMutation, useLazyGetRunQuery } from "../store/services/controlunitApi";
import { generateRunReportPdf, NoStepDataError } from "../util/runReportPdf";
import { ExecutionChart } from "./ExecutionChart";

const formatTimestamp = (timestamp?: number): string => {
  if (!timestamp) return "N/A";
  return new Date(timestamp * 1000).toLocaleString();
};

const getStateColor = (state: string): "success" | "error" | "warning" | "default" => {
  switch (state) {
    case "completed":
      return "success";
    case "failed":
      return "error";
    case "canceled":
      return "warning";
    default:
      return "default";
  }
};

export const History: React.FC = () => {
  const [selectedProgram, setSelectedProgram] = useState<string | null>(null);
  const { data: history, isLoading, error } = useGetExecutionHistoryQuery();
  const { data: logData, isLoading: isLoadingLog } = useGetExecutionLogQuery(selectedProgram || "", {
    skip: !selectedProgram,
  });
  const [deleteExecution] = useDeleteExecutionMutation();
  const [isGeneratingPdf, setIsGeneratingPdf] = useState(false);
  const [fetchRun] = useLazyGetRunQuery();

  const handleDownloadCsv = () => {
    if (!selectedProgram || !logData) return;
    const blob = new Blob([logData], { type: "text/csv" });
    const url = URL.createObjectURL(blob);
    const link = document.createElement("a");
    link.href = url;
    link.download = `${selectedProgram}.csv`;
    link.click();
    URL.revokeObjectURL(url);
  };

  const handleDownloadPdf = async () => {
    if (!selectedProgram || !logData) return;
    setIsGeneratingPdf(true);
    try {
      // If this fails, executed stays undefined and the report is
      // generated without the program appendix.
      const { data: executed } = await fetchRun(selectedProgram);
      const doc = generateRunReportPdf({
        runName: selectedProgram,
        csv: logData,
        executed,
      });
      doc.save(`${selectedProgram}.pdf`);
    } catch (error) {
      if (error instanceof NoStepDataError) {
        window.alert(error.message);
      } else {
        console.error(error);
        window.alert("Failed to generate PDF report");
      }
    } finally {
      setIsGeneratingPdf(false);
    }
  };

  const handleDelete = async (name: string, event: React.MouseEvent) => {
    event.stopPropagation();
    if (window.confirm(`Delete execution "${name}"?`)) {
      await deleteExecution(name);
      if (selectedProgram === name) {
        setSelectedProgram(null);
      }
    }
  };

  if (isLoading) {
    return (
      <Box
        sx={{
          display: "flex",
          justifyContent: "center",
          alignItems: "center",
          minHeight: "400px",
          padding: 4,
        }}
      >
        <CircularProgress />
      </Box>
    );
  }

  if (error) {
    return (
      <Box sx={{ padding: 4 }}>
        <Alert severity="error">Failed to load execution history</Alert>
      </Box>
    );
  }

  return (
    <Box
      sx={{
        display: "flex",
        width: "100%",
        height: "calc(100vh - 120px)",
        padding: 2,
        gap: 2,
      }}
    >
      {/* Left panel - List of executions */}
      <Paper
        sx={{
          width: "400px",
          flexShrink: 0,
          overflow: "auto",
        }}
      >
        <Box sx={{ padding: 2 }}>
          <Typography variant="h6" gutterBottom>
            Execution History
          </Typography>
          <Typography variant="body2" color="text.secondary">
            {history?.length || 0} execution{history?.length !== 1 ? "s" : ""}
          </Typography>
        </Box>
        <Divider />
        {!history || history.length === 0 ? (
          <Box sx={{ padding: 3, textAlign: "center" }}>
            <Typography color="text.secondary">No execution history</Typography>
          </Box>
        ) : (
          <List sx={{ padding: 0 }}>
            {history.map((item) => (
              <React.Fragment key={item.name}>
                <ListItem
                  disablePadding
                  secondaryAction={
                    <IconButton
                      edge="end"
                      aria-label="delete"
                      onClick={(e) => handleDelete(item.name, e)}
                      size="small"
                    >
                      <DeleteIcon />
                    </IconButton>
                  }
                >
                  <ListItemButton
                    selected={selectedProgram === item.name}
                    onClick={() => setSelectedProgram(item.name)}
                  >
                    <ListItemText
                      primary={
                        <Box sx={{ display: "flex", alignItems: "center", gap: 1 }}>
                          <Typography variant="body2" sx={{ flexGrow: 1, fontSize: "0.9rem" }}>
                            {item.name.split("@")[0]}
                          </Typography>
                          <Chip label={item.state} color={getStateColor(item.state)} size="small" />
                        </Box>
                      }
                      secondary={
                        <Box sx={{ fontSize: "0.75rem" }}>
                          <div>Started: {formatTimestamp(item.started_at)}</div>
                          {item.completed_at && <div>Ended: {formatTimestamp(item.completed_at)}</div>}
                        </Box>
                      }
                    />
                  </ListItemButton>
                </ListItem>
                <Divider />
              </React.Fragment>
            ))}
          </List>
        )}
      </Paper>

      {/* Right panel - Chart */}
      <Box sx={{ flexGrow: 1, overflow: "auto" }}>
        {selectedProgram ? (
          <ExecutionChart
            csvData={logData}
            title={`${selectedProgram.split("@")[0]} - Execution Chart`}
            isLoading={isLoadingLog}
            headerAction={
              logData ? (
                <Box>
                  <IconButton aria-label="download csv" onClick={handleDownloadCsv}>
                    <FileDownloadIcon />
                  </IconButton>
                  <IconButton
                    aria-label="download pdf report"
                    onClick={handleDownloadPdf}
                    disabled={isGeneratingPdf}
                  >
                    {isGeneratingPdf ? <CircularProgress size={20} /> : <PictureAsPdfIcon />}
                  </IconButton>
                </Box>
              ) : undefined
            }
          />
        ) : (
          <Paper
            sx={{
              padding: 4,
              height: "100%",
              display: "flex",
              alignItems: "center",
              justifyContent: "center",
            }}
          >
            <Typography variant="h6" color="text.secondary">
              Select a program to view its execution chart
            </Typography>
          </Paper>
        )}
      </Box>
    </Box>
  );
};
