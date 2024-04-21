import React, { useEffect, useMemo, useState } from "react";
import { Cycle as ApiCycle } from "../../types/api";
import { Button, Input, Stack, Typography } from "@mui/material";
import { States } from "./States";
import { useSaveCycleMutation } from "../../store/services";
import { useTranslation } from "react-i18next";

interface Props {
  canEdit: boolean;
  cycle: ApiCycle;
  mode: "create" | "view" | "edit";

  handleCancelEdit: () => void;
  handleEdit: () => void;
  onSave: () => void;
}

export const Cycle: React.FC<Props> = (props) => {
  const { canEdit, cycle, mode, handleCancelEdit, handleEdit, onSave } = props;

  const [edit, setEdit] = useState<ApiCycle>({ ...cycle });

  const [saveCycle, { isLoading, error, isSuccess }] = useSaveCycleMutation();
  const { t } = useTranslation();

  const updateEdited =
    (field: string) => (event: React.ChangeEvent<HTMLInputElement>) => {
      setEdit({ ...edit, [field]: event.currentTarget.value });
    };

  const handleStatesChange = (states: boolean[]) => {
    setEdit({ ...edit, states });
  };

  const editingThis = useMemo(
    () => mode === "edit" || (mode === "create" && cycle.id === ""),
    [mode, cycle.id]
  );

  useEffect(() => {
    if (!editingThis) {
      setEdit({ ...cycle });
    }
  }, [editingThis]);

  const NameComponent = useMemo(
    () =>
      editingThis ? (
        <Input
          style={{ fontSize: "2em" }}
          value={edit.name}
          onChange={updateEdited("name")}
          placeholder={t("cycle.name")}
        />
      ) : (
        <Typography variant="h4">{cycle.name}</Typography>
      ),
    [editingThis, edit.name, cycle.name]
  );

  const handleSave = () => {
    saveCycle(edit);
    onSave();
  };

  const InteractionButtons = useMemo(
    () =>
      editingThis ? (
        <Stack direction="row" gap={3}>
          <Button onClick={handleSave} disabled={!edit.name}>
            {t("cycle.save")}
          </Button>

          <Button onClick={handleCancelEdit} color="warning">
            {t("cycle.cancel")}
          </Button>
        </Stack>
      ) : (
        <Button disabled={!canEdit} onClick={handleEdit}>
          {t("cycle.edit")}
        </Button>
      ),
    [editingThis, handleSave, edit.name, canEdit, handleEdit]
  );

  const stackStyle = useMemo((): React.CSSProperties => {
    const baseStyle: React.CSSProperties = {
      padding: "1em",
    };

    return editingThis ? { ...baseStyle, ...editStyle } : baseStyle;
  }, [editingThis]);

  return (
    <Stack direction="column" gap={3} style={stackStyle}>
      {NameComponent}

      <Stack direction="row" gap={3}>
        <States
          cycle={editingThis ? edit : cycle}
          handleChange={editingThis ? handleStatesChange : undefined}
        />

        {InteractionButtons}
      </Stack>
    </Stack>
  );
};

const editStyle: React.CSSProperties = {
  borderRadius: "1em",
  backgroundColor: "#333",
};
