import React, { useEffect, useMemo, useState } from "react";
import { Program as ApiProgram } from "../../types/api";
import { Button, Stack } from "@mui/material";
import { useTranslation } from "react-i18next";
import { FormMode } from "../../types";
import { setEditProgram } from "../../store/features/programsSlice";
import { useDispatch, useSelector } from "react-redux";
import { NameComponent } from "../form";
import { RootState } from "../../store/store";
import { useNavigate, useParams } from "react-router-dom";
import {
  useGetProgramsQuery,
  useSaveProgramMutation,
} from "../../store/services";
import { validName } from "../../util";
import { emptyProgram } from "./templates";

export const Program: React.FC = () => {
  const [mode, setMode] = useState<FormMode>("view");
  const { name } = useParams();
  const [program, setProgram] = useState<ApiProgram>(emptyProgram());
  const { t } = useTranslation();
  const navigate = useNavigate();
  const { data } = useGetProgramsQuery();
  const [saveProgram, { isSuccess }] = useSaveProgramMutation();
  const dispatch = useDispatch();
  const editProgram = useSelector(
    (state: RootState) => state.programs.editRecord
  );

  const programs = useMemo(() => data as ApiProgram[], [data]);
  const handleEdit = () => {
    dispatch(setEditProgram(program));
    setMode("edit");
  };

  const updateEdited =
    (field: keyof ApiProgram) =>
    (event: React.ChangeEvent<HTMLInputElement>) => {
      if (editProgram) {
        dispatch(
          setEditProgram({ ...editProgram, [field]: event.currentTarget.value })
        );
      }
    };

  useEffect(() => {
    if (name === "new") {
      setMode("edit");

      if (editProgram && !editProgram.name) {
        return;
      }

      dispatch(setEditProgram(emptyProgram()));

      return;
    }

    if (!name || !programs) {
      return;
    }

    const program = programs.find((p) => p.name === name);

    if (!program) {
      navigate("/phases");
      return;
    }

    setProgram(program);
  }, [name, programs]);

  useEffect(() => {
    if (!isSuccess) {
      return;
    }

    const editName = editProgram?.name;
    dispatch(setEditProgram(undefined));

    if (editName === "") {
      navigate("/phases");
    } else {
      setMode("view");
    }
  }, [isSuccess]);

  const editingThis = useMemo(() => mode === "edit", [mode]);

  const normalize = (program: ApiProgram): ApiProgram => {
    const cpy = { ...program };
    cpy.name = cpy.name.trim();

    return cpy;
  };

  const handleSave = () => {
    if (!editProgram) {
      return;
    }

    const normalized = normalize(editProgram);
    saveProgram(normalized);
  };

  const handleCancel = () => {
    const editName = editProgram?.name;
    dispatch(setEditProgram(undefined));

    if (editName === "") {
      navigate("/programs");
    } else {
      setMode("view");
    }
  };

  const nameUsed = useMemo(() => {
    if (!editProgram || !programs?.length) {
      return false;
    }

    for (const p of programs) {
      if (p.name === program.name) {
        continue;
      }

      if (p.name.trim() === editProgram.name.trim()) {
        return true;
      }
    }

    return false;
  }, [programs, editProgram]);

  const isValid = useMemo(() => {
    if (!editProgram) {
      return false;
    }

    const { name } = editProgram;

    return !nameUsed && validName(name, ["new", "latest", "current"]);
  }, [editProgram]);

  return (
    <Stack direction="column" gap={6} width="60rem">
      {!editingThis && (
        <Stack direction="row" justifyContent="end" gap={6}>
          <Button color="primary" onClick={handleEdit}>
            {t("programs.edit")}
          </Button>
        </Stack>
      )}

      <NameComponent
        editing={editingThis}
        name={editingThis ? editProgram?.name : program.name}
        handleChange={updateEdited("name")}
      />

      {editingThis && (
        <Stack direction="row" gap="3em" justifyContent="flex-end">
          <Button onClick={handleSave} disabled={!isValid} color="success">
            {t("programs.save")}
          </Button>

          <Button onClick={handleCancel} color="warning">
            {t("programs.cancel")}
          </Button>
        </Stack>
      )}
    </Stack>
  );
};
