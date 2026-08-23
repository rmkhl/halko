import React, { useState } from "react";
import {
  Alert,
  Button,
  Dialog,
  DialogActions,
  DialogContent,
  DialogTitle,
  TextField,
} from "@mui/material";
import { useTranslation } from "react-i18next";

interface Props {
  open: boolean;
  submitting: boolean;
  failed: boolean;
  onCancel: () => void;
  // Resolves true when the note landed. The dialog clears itself only then: on
  // a failure the operator's words must still be there to retry with.
  onSubmit: (text: string) => Promise<boolean>;
}

// Follows the stop-confirmation dialog in RunningProgram.tsx - a plain MUI
// Dialog with a DialogActions row - rather than components/form/Dialog, which
// offers only a close icon and no actions.
//
// The dialog owns the text and clears it only on cancel, so a failed save
// leaves the operator's words where they typed them.
export const AddNoteDialog: React.FC<Props> = ({
  open,
  submitting,
  failed,
  onCancel,
  onSubmit,
}) => {
  const { t } = useTranslation();
  const [text, setText] = useState("");

  const handleCancel = () => {
    setText("");
    onCancel();
  };

  const handleSubmit = async () => {
    if (await onSubmit(text)) {
      setText("");
    }
  };

  return (
    <Dialog open={open} onClose={handleCancel} fullWidth maxWidth="sm">
      <DialogTitle>{t("notes.dialogTitle")}</DialogTitle>
      <DialogContent>
        {failed && (
          <Alert severity="error" sx={{ marginBottom: 2 }}>
            {t("notes.failed")}
          </Alert>
        )}
        <TextField
          autoFocus
          fullWidth
          multiline
          minRows={3}
          margin="dense"
          label={t("notes.label")}
          helperText={t("notes.help")}
          value={text}
          onChange={(e) => setText(e.currentTarget.value)}
        />
      </DialogContent>
      <DialogActions>
        <Button onClick={handleCancel} color="inherit">
          {t("notes.cancel")}
        </Button>
        <Button
          onClick={handleSubmit}
          variant="contained"
          disabled={submitting || text.trim() === ""}
        >
          {t("notes.save")}
        </Button>
      </DialogActions>
    </Dialog>
  );
};
