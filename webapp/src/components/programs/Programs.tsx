import React, { useState, useMemo } from "react";
import { useGetProgramsQuery, useGetProgramQuery, useDeleteProgramMutation } from "../../store/services";
import { useNavigate } from "react-router-dom";
import {
  Box,
  Typography,
  List,
  ListItem,
  ListItemButton,
  ListItemText,
  Paper,
  CircularProgress,
  Alert,
  IconButton,
  Divider,
  Button,
  Stack,
  ButtonGroup,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogContentText,
  DialogActions,
} from "@mui/material";
import DeleteIcon from "@mui/icons-material/Delete";
import EditIcon from "@mui/icons-material/Edit";
import AddIcon from "@mui/icons-material/Add";
import SortByAlphaIcon from "@mui/icons-material/SortByAlpha";
import AccessTimeIcon from "@mui/icons-material/AccessTime";
import ArrowUpwardIcon from "@mui/icons-material/ArrowUpward";
import ArrowDownwardIcon from "@mui/icons-material/ArrowDownward";
import { Program as ApiProgram } from "../../types/api";

interface StoredProgramInfo {
  name: string;
  last_modified: string;
}

type SortBy = "name" | "modified";
type SortOrder = "asc" | "desc";

const formatTimestamp = (timestamp: string): string => {
  return new Date(timestamp).toLocaleString();
};

export const Programs: React.FC = () => {
  const [selectedProgram, setSelectedProgram] = useState<string | null>(null);
  const [sortBy, setSortBy] = useState<SortBy>("name");
  const [sortOrder, setSortOrder] = useState<SortOrder>("asc");
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false);
  const [programToDelete, setProgramToDelete] = useState<string | null>(null);
  const { data, isLoading, error } = useGetProgramsQuery();
  const { data: programData, isLoading: isLoadingProgram } = useGetProgramQuery(selectedProgram || "", {
    skip: !selectedProgram,
  });
  const [deleteProgram] = useDeleteProgramMutation();
  const navigate = useNavigate();

  const programInfos = (() => {
    if (!data) return [];
    const responseData = data as any;
    if (responseData.data && Array.isArray(responseData.data)) {
      return responseData.data as StoredProgramInfo[];
    }
    if (Array.isArray(responseData)) {
      return responseData as StoredProgramInfo[];
    }
    return [];
  })();

  const sortedPrograms = useMemo(() => {
    const programs = [...programInfos];

    programs.sort((a, b) => {
      let comparison = 0;

      if (sortBy === "name") {
        comparison = a.name.localeCompare(b.name);
      } else {
        const dateA = new Date(a.last_modified).getTime();
        const dateB = new Date(b.last_modified).getTime();
        comparison = dateA - dateB;
      }

      return sortOrder === "asc" ? comparison : -comparison;
    });

    return programs;
  }, [programInfos, sortBy, sortOrder]);

  const selectedProgramData = (() => {
    if (!programData) return null;
    const responseData = programData as any;
    if (responseData.data) {
      return responseData.data as ApiProgram;
    }
    return programData as ApiProgram;
  })();

  const handleDelete = async (name: string, event: React.MouseEvent) => {
    event.stopPropagation();
    setProgramToDelete(name);
    setDeleteDialogOpen(true);
  };

  const confirmDelete = async () => {
    if (programToDelete) {
      await deleteProgram(programToDelete);
      if (selectedProgram === programToDelete) {
        setSelectedProgram(null);
      }
      setProgramToDelete(null);
    }
    setDeleteDialogOpen(false);
  };

  const cancelDelete = () => {
    setProgramToDelete(null);
    setDeleteDialogOpen(false);
  };

  const handleSortChange = (newSortBy: SortBy) => {
    // Clear selection when sorting changes
    setSelectedProgram(null);

    if (sortBy === newSortBy) {
      // Toggle order if clicking same sort field
      setSortOrder(sortOrder === "asc" ? "desc" : "asc");
    } else {
      // Change sort field and reset to ascending
      setSortBy(newSortBy);
      setSortOrder("asc");
    }
  };

  const handleEdit = (name: string) => {
    navigate(`/programs/${encodeURIComponent(name)}`);
  };

  const handleNew = () => {
    navigate("/programs/new");
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
        <Alert severity="error">Failed to load programs</Alert>
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
      {/* Left panel - List of programs */}
      <Paper
        sx={{
          width: "400px",
          flexShrink: 0,
          overflow: "auto",
        }}
      >
        <Box sx={{ padding: 2, display: "flex", alignItems: "center", justifyContent: "space-between" }}>
          <Box>
            <Typography variant="h6" gutterBottom>
              Programs
            </Typography>
            <Typography variant="body2" color="text.secondary">
              {programInfos.length} program{programInfos.length !== 1 ? "s" : ""}
            </Typography>
          </Box>
          <Button
            variant="contained"
            color="primary"
            startIcon={<AddIcon />}
            onClick={handleNew}
            size="small"
          >
            New
          </Button>
        </Box>
        <Divider />
        <Box sx={{ padding: 1, display: "flex", justifyContent: "center" }}>
          <ButtonGroup size="small" variant="outlined">
            <Button
              onClick={() => handleSortChange("name")}
              variant={sortBy === "name" ? "contained" : "outlined"}
              startIcon={<SortByAlphaIcon />}
              endIcon={sortBy === "name" ? (sortOrder === "asc" ? <ArrowUpwardIcon fontSize="small" /> : <ArrowDownwardIcon fontSize="small" />) : null}
            >
              Name
            </Button>
            <Button
              onClick={() => handleSortChange("modified")}
              variant={sortBy === "modified" ? "contained" : "outlined"}
              startIcon={<AccessTimeIcon />}
              endIcon={sortBy === "modified" ? (sortOrder === "asc" ? <ArrowUpwardIcon fontSize="small" /> : <ArrowDownwardIcon fontSize="small" />) : null}
            >
              Modified
            </Button>
          </ButtonGroup>
        </Box>
        <Divider />
        {programInfos.length === 0 ? (
          <Box sx={{ padding: 3, textAlign: "center" }}>
            <Typography color="text.secondary">No programs found</Typography>
          </Box>
        ) : (
          <List sx={{ padding: 0, paddingBottom: 4 }}>
            {sortedPrograms.map((item) => (
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
                        <Typography variant="body1">{item.name}</Typography>
                      }
                      secondary={
                        <Typography variant="caption" color="text.secondary">
                          Modified: {formatTimestamp(item.last_modified)}
                        </Typography>
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

      {/* Right panel - Program details */}
      <Box sx={{ flexGrow: 1, overflow: "auto" }}>
        {selectedProgram ? (
          isLoadingProgram ? (
            <Paper
              sx={{
                padding: 4,
                height: "100%",
                display: "flex",
                alignItems: "center",
                justifyContent: "center",
              }}
            >
              <CircularProgress />
            </Paper>
          ) : selectedProgramData ? (
            <Paper sx={{ padding: 3, height: "100%", display: "flex", flexDirection: "column", overflow: "hidden" }}>
              <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 3 }}>
                <Typography variant="h5">{selectedProgramData.name}</Typography>
                <Button
                  variant="contained"
                  startIcon={<EditIcon />}
                  onClick={() => handleEdit(selectedProgram)}
                >
                  Edit
                </Button>
              </Box>
              <Divider sx={{ marginBottom: 2 }} />
              <Typography variant="h6" gutterBottom>
                Steps ({selectedProgramData.steps?.length || 0})
              </Typography>
              <Box sx={{ flexGrow: 1, overflow: "auto", paddingRight: 1 }}>
                {selectedProgramData.steps && selectedProgramData.steps.length > 0 ? (
                  <Stack spacing={2} sx={{ paddingBottom: 2 }}>
                    {selectedProgramData.steps.map((step, index) => (
                      <Paper key={index} variant="outlined" sx={{ padding: 2 }}>
                        <Typography variant="subtitle1" fontWeight="bold">
                          {index + 1}. {step.name}
                        </Typography>
                        <Typography variant="body2" color="text.secondary">
                          Type: {step.type}
                        </Typography>
                        <Typography variant="body2" color="text.secondary">
                          Target Temperature: {step.temperature_target}°C
                        </Typography>
                        {step.runtime && (
                          <Typography variant="body2" color="text.secondary">
                            Runtime: {step.runtime}
                          </Typography>
                        )}
                      </Paper>
                    ))}
                  </Stack>
                ) : (
                  <Typography color="text.secondary">No steps defined</Typography>
                )}
              </Box>
            </Paper>
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
                Failed to load program details
              </Typography>
            </Paper>
          )
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
              Select a program to view its details
            </Typography>
          </Paper>
        )}
      </Box>

      {/* Delete Confirmation Dialog */}
      <Dialog
        open={deleteDialogOpen}
        onClose={cancelDelete}
        aria-labelledby="delete-dialog-title"
        aria-describedby="delete-dialog-description"
      >
        <DialogTitle id="delete-dialog-title">Delete Program</DialogTitle>
        <DialogContent>
          <DialogContentText id="delete-dialog-description">
            Are you sure you want to delete &quot;{programToDelete}&quot;? This action cannot be undone.
          </DialogContentText>
        </DialogContent>
        <DialogActions>
          <Button onClick={cancelDelete} color="primary">
            Cancel
          </Button>
          <Button onClick={confirmDelete} color="error" variant="contained" autoFocus>
            Delete
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
};
