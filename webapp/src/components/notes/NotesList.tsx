import React from "react";
import { Paper, Stack, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";
import { RunNote } from "../../types/api";

interface Props {
  notes?: RunNote[];
}

const formatTime = (epochSeconds: number): string =>
  new Date(epochSeconds * 1000).toLocaleString();

// Newest first: during a run the note just taken is the one being looked for,
// and in history the last thing that happened is the usual question.
export const NotesList: React.FC<Props> = ({ notes }) => {
  const { t } = useTranslation();

  if (!notes || notes.length === 0) {
    return (
      <Typography variant="body2" color="text.secondary">
        {t("notes.none")}
      </Typography>
    );
  }

  const newestFirst = [...notes].reverse();

  return (
    <Stack gap={1}>
      {newestFirst.map((note, index) => (
        <Paper key={index} variant="outlined" sx={{ padding: 1.5 }}>
          <Typography variant="caption" color="text.secondary">
            {t("notes.at", { time: formatTime(note.time), step: note.step })}
          </Typography>
          <Typography variant="caption" color="text.secondary" display="block">
            {t("notes.temperatures", {
              kiln: note.temperatures.kiln.toFixed(1),
              material: note.temperatures.material.toFixed(1),
            })}
          </Typography>
          <Typography variant="body2" sx={{ whiteSpace: "pre-wrap", marginTop: 0.5 }}>
            {note.text}
          </Typography>
        </Paper>
      ))}
    </Stack>
  );
};
