import { Button, Stack } from "@mui/material";
import React from "react";
import { useTranslation } from "react-i18next";

interface Props {
  editing: boolean;
  editDisabled: boolean;
  saveDisabled: boolean;
  handleEdit: () => void;
  handleSave: () => void;
  handleCancelEdit: () => void;
}

export const FormButtons: React.FC<Props> = (props) => {
  const {
    editing,
    editDisabled,
    saveDisabled,
    handleEdit,
    handleSave,
    handleCancelEdit,
  } = props;
  const { t } = useTranslation();

  return editing ? (
    <Stack direction="row" gap={3}>
      <Button onClick={handleSave} disabled={saveDisabled}>
        {t("cycle.save")}
      </Button>

      <Button onClick={handleCancelEdit} color="warning">
        {t("cycle.cancel")}
      </Button>
    </Stack>
  ) : (
    <Button disabled={editDisabled} onClick={handleEdit}>
      {t("cycle.edit")}
    </Button>
  );
};
