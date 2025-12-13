import { Button, Stack, Typography } from "@mui/material";
import React from "react";
import { useTranslation } from "react-i18next";

interface Props extends React.PropsWithChildren {
  editing: boolean;
  isValid: boolean;
  programName?: string;

  handleCancel: () => void;
  handleEdit: () => void;
  handleSave: () => void;
  handleRun?: () => void;
  handleBack?: () => void;
}

export const DataForm: React.FC<Props> = (props) => {
  const { children, editing, handleEdit, handleSave, isValid, handleCancel, handleRun, handleBack, programName } =
    props;
  const { t } = useTranslation();

  return (
    <Stack direction="column" gap={6} width="60rem" sx={{ height: "100%", overflow: "hidden", padding: 4 }}>
      {!editing && (
        <Stack direction="row" justifyContent="space-between" alignItems="center">
          <Typography variant="h4">{programName}</Typography>
          <Stack direction="row" gap={2}>
            {handleBack && (
              <Button color="inherit" onClick={handleBack}>
                {t("common.back")}
              </Button>
            )}
            <Button color="primary" onClick={handleEdit}>
              {t("common.edit")}
            </Button>
            {handleRun && (
              <Button color="success" onClick={handleRun}>
                Run
              </Button>
            )}
          </Stack>
        </Stack>
      )}

      <Stack sx={{ flex: 1, overflowY: "auto", minHeight: 0 }} gap={6}>
        {children}
      </Stack>

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
