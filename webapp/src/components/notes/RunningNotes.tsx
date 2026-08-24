import React from "react";
import { Paper, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";
import {
  useGetRunningNotesQuery,
  useGetRunningProgramQuery,
} from "../../store/services/controlunitApi";
import { NotesList } from "./NotesList";

// The list only - taking a note is done from the button beside Stop, at the
// top of the page, so this can sit below the chart where it belongs.
export const RunningNotes: React.FC = () => {
  const { t } = useTranslation();

  const { data: runningProgram } = useGetRunningProgramQuery();
  const { data: notes } = useGetRunningNotesQuery(undefined, {
    pollingInterval: 5000,
    skipPollingIfUnfocused: true,
  });

  // Nothing running, nothing to annotate - the rest of this view hides itself
  // the same way.
  if (!runningProgram) {
    return null;
  }

  return (
    <Paper sx={{ width: "100%", maxWidth: "1200px", padding: 2, marginTop: 2 }}>
      <Typography variant="h6" marginBottom={2}>
        {t("notes.title")}
      </Typography>
      <NotesList notes={notes} />
    </Paper>
  );
};
