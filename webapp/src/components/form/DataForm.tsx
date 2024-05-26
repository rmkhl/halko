import { Button, Stack } from "@mui/material";
import React from "react";
import { useTranslation } from "react-i18next";

interface Props extends React.PropsWithChildren {
  editing: boolean;
  isValid: boolean;

  handleCancel: () => void;
  handleEdit: () => void;
  handleSave: () => void;
}

export const DataForm: React.FC<Props> = (props) => {
  const { children, editing, handleEdit, handleSave, isValid, handleCancel } =
    props;
  const { t } = useTranslation();

  return (
    <Stack direction="column" gap={6} width="60rem">
      {!editing && (
        <Stack direction="row" justifyContent="end" gap={6}>
          <Button color="primary" onClick={handleEdit}>
            {t("common.edit")}
          </Button>
        </Stack>
      )}

      {children}

      {editing && (
        <Stack direction="row" gap="3em" justifyContent="flex-end">
          <Button onClick={handleSave} disabled={!isValid} color="success">
            {t("common.save")}
          </Button>

          <Button onClick={handleCancel} color="warning">
            {t("common.cancel")}
          </Button>
        </Stack>
      )}
    </Stack>
  );
};
