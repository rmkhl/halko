import React, { useMemo, useState } from "react";
import { useGetCyclesQuery } from "../../store/services";
import { Button, Stack } from "@mui/material";
import { Cycle } from "./Cycle";
import { Cycle as ApiCycle } from "../../types/api";
import { useTranslation } from "react-i18next";

type FormMode = "create" | "edit" | "view";

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
  const [edit, setEdit] = useState("");
  const [mode, setMode] = useState<FormMode>("view");
  const [newCycle, setNewCycle] = useState<ApiCycle>({ ...emptyCycle });

  const { t } = useTranslation();

  const handleEdit = (id: string) => () => {
    setEdit(id);
  };

  const handleSave = () => {
    setEdit("");
    setMode("view");
  };

  const allCycles = useMemo(() => {
    const allCycles: ApiCycle[] = [];

    if (mode === "create") {
      allCycles.push(newCycle);
    }

    if (cycles) {
      allCycles.push(
        ...[...cycles].sort((a, b) => a.name.localeCompare(b.name))
      );
    }

    return allCycles;
  }, [cycles, newCycle, mode]);

  const addNew = () => {
    setMode("create");
  };

  const cancelEdit = () => {
    setEdit("");
    setMode("view");
  };

  return (
    <Stack direction="column" gap={6}>
      <Stack direction="row" justifyContent="end" gap={6}>
        <Button
          color="success"
          onClick={addNew}
          disabled={mode === "create" || edit !== ""}
        >
          {t("cycle.new")}
        </Button>
      </Stack>

      {allCycles?.map((c) => (
        <Cycle
          key={`cycle-${c.id}`}
          cycle={c}
          mode={edit === c.id ? "edit" : mode}
          handleEdit={handleEdit(c.id)}
          onSave={handleSave}
          canEdit={!edit}
          handleCancelEdit={cancelEdit}
        />
      ))}
    </Stack>
  );
};
