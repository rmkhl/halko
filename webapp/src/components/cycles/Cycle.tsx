import React, { useMemo } from "react";
import { Cycle as ApiCycle } from "../../types/api";
import { Stack } from "@mui/material";
import { States } from "./States";
import { useSaveCycleMutation } from "../../store/services";
import { useDispatch, useSelector } from "react-redux";
import { setEditCycle } from "../../store/features/cyclesSlice";
import { RootState } from "../../store/store";
import { FormButtons } from "./FormButtons";
import { NameComponent } from "../form";

interface Props {
  canEdit: boolean;
  cycle: ApiCycle;
  mode: "view" | "edit";

  handleCancelEdit: () => void;
  handleEdit: (c: ApiCycle) => void;
  onSave: () => void;
}

export const Cycle: React.FC<Props> = (props) => {
  const { canEdit, cycle, mode, handleCancelEdit, handleEdit, onSave } = props;

  const [saveCycle, { isLoading, error, isSuccess }] = useSaveCycleMutation();
  const editCycle = useSelector((state: RootState) => state.cycles.edit);
  const dispatch = useDispatch();

  const updateEdited =
    (field: keyof ApiCycle) => (event: React.ChangeEvent<HTMLInputElement>) => {
      if (editCycle) {
        dispatch(
          setEditCycle({ ...editCycle, [field]: event.currentTarget.value })
        );
      }
    };

  const handleStatesChange = (states: boolean[]) => {
    if (editCycle) {
      dispatch(setEditCycle({ ...editCycle, states }));
    }
  };

  const editingThis = useMemo(() => mode === "edit", [mode]);

  const handleSave = () => {
    if (editCycle) {
      saveCycle(editCycle);
    }

    onSave();
  };

  const stackStyle = useMemo((): React.CSSProperties => {
    const baseStyle: React.CSSProperties = {
      padding: "1em",
    };

    return editingThis ? { ...baseStyle, ...editStyle } : baseStyle;
  }, [editingThis]);

  return (
    <Stack direction="column" gap={3} style={stackStyle}>
      <NameComponent
        editing={editingThis}
        name={editingThis ? editCycle?.name : cycle.name}
        handleChange={updateEdited("name")}
      />

      <Stack direction="row" gap={3}>
        <States
          cycle={editingThis ? editCycle || cycle : cycle}
          handleChange={editingThis ? handleStatesChange : undefined}
        />

        <FormButtons
          editing={editingThis}
          editDisabled={!canEdit}
          saveDisabled={!editCycle?.name}
          handleEdit={() => handleEdit(cycle)}
          handleSave={handleSave}
          handleCancelEdit={handleCancelEdit}
        />
      </Stack>
    </Stack>
  );
};

const editStyle: React.CSSProperties = {
  borderRadius: "1em",
  backgroundColor: "#333",
};
