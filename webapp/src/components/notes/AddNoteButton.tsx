import React, { useState } from "react";
import { Button } from "@mui/material";
import AddCommentIcon from "@mui/icons-material/AddComment";
import { useTranslation } from "react-i18next";
import { useAddRunningNoteMutation } from "../../store/services/controlunitApi";
import { AddNoteDialog } from "./AddNoteDialog";

// Sits next to Stop, and is the one of the pair that a stray Enter may
// activate: adding a note is harmless, stopping the run is not. The caller
// renders it only while something is running.
export const AddNoteButton: React.FC = () => {
  const { t } = useTranslation();
  const [open, setOpen] = useState(false);
  const [failed, setFailed] = useState(false);

  const [addNote, { isLoading }] = useAddRunningNoteMutation();

  const handleSubmit = async (text: string): Promise<boolean> => {
    try {
      await addNote(text).unwrap();
      setFailed(false);
      setOpen(false);
      return true;
    } catch {
      // Leave the dialog open with the text still in it.
      setFailed(true);
      return false;
    }
  };

  return (
    <>
      <Button
        variant="contained"
        startIcon={<AddCommentIcon />}
        onClick={() => {
          setFailed(false);
          setOpen(true);
        }}
      >
        {t("notes.add")}
      </Button>

      <AddNoteDialog
        open={open}
        submitting={isLoading}
        failed={failed}
        onCancel={() => setOpen(false)}
        onSubmit={handleSubmit}
      />
    </>
  );
};
