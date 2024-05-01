import React, { useMemo, useState } from "react";
import { useGetCyclesQuery } from "../../store/services";
import { Button, Stack } from "@mui/material";
import { Cycle } from "./Cycle";
import { Cycle as ApiCycle } from "../../types/api";
import { useTranslation } from "react-i18next";
import { useDispatch, useSelector } from "react-redux";
import { RootState } from "../../store/store";
import { setEditCycle } from "../../store/features/cyclesSlice";
import { FormMode } from "../../types";

const emptyCycle: ApiCycle = {
  id: "",
  name: "",
  states: [
    false,
    false,
    false,
    false,
    false,
    false,
    false,
    false,
    false,
    false,
  ],
};

export const Cycles: React.FC = () => {
  const { data: cycles, isFetching } = useGetCyclesQuery();
  const [mode, setMode] = useState<FormMode>("view");

  const editCycle = useSelector((state: RootState) => state.cycles.edit);
  const dispatch = useDispatch();

  const { t } = useTranslation();

  const handleEdit = (cycle: ApiCycle) => {
    dispatch(setEditCycle(cycle));
  };

  const handleSave = () => {
    dispatch(setEditCycle(undefined));
    setMode("view");
  };

  const allCycles = useMemo(() => {
    const allCycles: ApiCycle[] = [];

    if (editCycle && !editCycle.id) {
      allCycles.push(editCycle);
    }

    if (cycles) {
      allCycles.push(
        ...[...cycles].sort((a, b) => a.name.localeCompare(b.name))
      );
    }

    return allCycles;
  }, [cycles, editCycle, mode]);

  const addNew = () => {
    dispatch(setEditCycle({ ...emptyCycle }));
  };

  const cancelEdit = () => {
    dispatch(setEditCycle(undefined));
    setMode("view");
  };

  return (
    <Stack direction="column" gap={6} width="60rem">
      <Stack direction="row" justifyContent="end" gap={6}>
        <Button color="success" onClick={addNew} disabled={!!editCycle}>
          {t("cycle.new")}
        </Button>
      </Stack>

      {allCycles?.map((c) => (
        <Cycle
          key={`cycle-${c.id}`}
          cycle={c}
          mode={editCycle?.id === c.id ? "edit" : mode}
          handleEdit={handleEdit}
          onSave={handleSave}
          canEdit={!editCycle}
          handleCancelEdit={cancelEdit}
        />
      ))}
    </Stack>
  );
};
