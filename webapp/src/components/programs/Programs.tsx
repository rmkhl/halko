import React, { useMemo, useState } from "react";
import { useGetProgramsQuery } from "../../store/services";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { Stack, Button, Typography, Box } from "@mui/material";

interface StoredProgramInfo {
  name: string;
  last_modified: string;
}

type SortField = "name" | "last_modified";
type SortOrder = "asc" | "desc";

export const Programs: React.FC = () => {
  const { data, isLoading, error } = useGetProgramsQuery();
  const [sortField, setSortField] = useState<SortField>("name");
  const [sortOrder, setSortOrder] = useState<SortOrder>("asc");

  const { t } = useTranslation();
  const navigate = useNavigate();

  const programInfos = useMemo(() => {
    if (!data) return [];

    const responseData = data as any;

    // Check if we have a data property with an array
    if (responseData.data && Array.isArray(responseData.data)) {
      return responseData.data as StoredProgramInfo[];
    }

    // Fallback: if data is directly an array
    if (Array.isArray(responseData)) {
      return responseData as StoredProgramInfo[];
    }

    console.error("Unexpected response format:", responseData);
    return [];
  }, [data]);

  const sortedPrograms = useMemo(() => {
    const programs = [...programInfos];

    programs.sort((a, b) => {
      let comparison = 0;

      if (sortField === "name") {
        comparison = a.name.localeCompare(b.name);
      } else {
        const dateA = new Date(a.last_modified).getTime();
        const dateB = new Date(b.last_modified).getTime();
        comparison = dateA - dateB;
      }

      return sortOrder === "asc" ? comparison : -comparison;
    });

    return programs;
  }, [programInfos, sortField, sortOrder]);

  const handleSort = (field: SortField) => {
    if (sortField === field) {
      setSortOrder(sortOrder === "asc" ? "desc" : "asc");
    } else {
      setSortField(field);
      setSortOrder("asc");
    }
  };

  const addNew = () => {
    navigate("/programs/new");
  };

  const handleShow = (programName: string) => {
    navigate(`/programs/${encodeURIComponent(programName)}`);
  };

  if (isLoading) {
    return (
      <Stack alignItems="center" padding={4}>
        <Typography>Loading programs...</Typography>
      </Stack>
    );
  }

  if (error) {
    const errorMessage = (() => {
      const err = error as any;
      if (err && typeof err === 'object') {
        if ('status' in err) {
          // FetchBaseQueryError
          if (err.status === 'FETCH_ERROR') {
            return 'Cannot connect to the backend service. Please check if the control unit is running.';
          }
          if (err.status === 'PARSING_ERROR') {
            return 'Failed to parse server response.';
          }
          if (typeof err.status === 'number') {
            return `Server error: HTTP ${err.status}`;
          }
          return `Error: ${err.status}`;
        }
        if ('message' in err) {
          // SerializedError
          return err.message || 'An unknown error occurred';
        }
      }
      return 'Failed to load programs';
    })();

    return (
      <Stack alignItems="center" padding={4} gap={2}>
        <Typography color="error" variant="h6">Error loading programs</Typography>
        <Typography color="text.secondary">{errorMessage}</Typography>
      </Stack>
    );
  }

  return (
    <Stack padding={4} width="100%" sx={{ height: "100%", overflow: "hidden" }}>
      <Stack direction="row" alignItems="center" marginBottom={2} paddingX={2}>
        <Typography variant="subtitle2" color="text.secondary" flex={1} sx={{ cursor: "pointer" }} onClick={() => handleSort("name")}>
          Name {sortField === "name" && (sortOrder === "asc" ? "↑" : "↓")}
        </Typography>
        <Typography variant="subtitle2" color="text.secondary" flex={1} sx={{ cursor: "pointer" }} onClick={() => handleSort("last_modified")}>
          Last modified {sortField === "last_modified" && (sortOrder === "asc" ? "↑" : "↓")}
        </Typography>
        <Box sx={{ width: 160, display: "flex", justifyContent: "flex-end" }}>
          <Button color="success" onClick={addNew}>
            {t("programs.new")}
          </Button>
        </Box>
      </Stack>

      <Stack direction="column" gap={1} sx={{ overflowY: "auto", flex: 1, minHeight: 0 }}>
        {programInfos.length === 0 ? (
          <Typography color="text.secondary" padding={2}>
            No programs found
          </Typography>
        ) : (
          sortedPrograms.map((programInfo) => (
              <Stack
                key={`program-${programInfo.name}`}
                direction="row"
                alignItems="center"
                paddingY={0.5}
                paddingX={2}
              >
                <Typography variant="h6" flex={1}>{programInfo.name}</Typography>
                <Typography variant="body2" color="text.secondary" flex={1}>
                  {new Date(programInfo.last_modified).toLocaleString()}
                </Typography>
                <Box sx={{ width: 160, display: "flex", gap: 1, justifyContent: "flex-end" }}>
                  <Button variant="contained" size="small" onClick={() => handleShow(programInfo.name)}>
                    Show
                  </Button>
                </Box>
              </Stack>
            ))
        )}
      </Stack>
    </Stack>
  );
};
