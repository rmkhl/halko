import React, { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useGetPhasesQuery, useSavePhaseMutation } from "../../store/services";
import { Phase as ApiPhase, DeltaCycle } from "../../types/api";
import { Button, Stack } from "@mui/material";
import { FormMode } from "../../types";
import { useDispatch, useSelector } from "react-redux";
import { RootState } from "../../store/store";
import { setEditPhase } from "../../store/features/phasesSlice";
import { NameComponent } from "../form";
import { useTranslation } from "react-i18next";
import { Cycles } from "./Cycles";
import {
  defaultConstant,
  defaultDeltaCycles,
  emptyConstantPhase,
} from "./templates";
import { validName } from "../../util";

export const Phase: React.FC = () => {
  const [mode, setMode] = useState<FormMode>("view");
  const [phase, setPhase] = useState<ApiPhase>(emptyConstantPhase());

  const { name } = useParams();
  const { data } = useGetPhasesQuery();
  const [savePhase, { isSuccess }] = useSavePhaseMutation();
  const editPhase = useSelector((state: RootState) => state.phases.editRecord);
  const navigate = useNavigate();
  const { t } = useTranslation();

  const phases = useMemo(() => data as ApiPhase[], [data]);

  const dispatch = useDispatch();

  useEffect(() => {
    if (name === "new") {
      setMode("edit");

      if (editPhase && !editPhase.name) {
        return;
      }

      dispatch(setEditPhase(emptyConstantPhase()));

      return;
    }

    if (!name || !phases) {
      return;
    }

    const phase = phases.find((p) => p.name === name);

    if (!phase) {
      navigate("/phases");
      return;
    }

    setPhase(phase);
  }, [name, phases]);

  useEffect(() => {
    if (!isSuccess) {
      return;
    }

    const editName = editPhase?.name;
    dispatch(setEditPhase(undefined));

    if (editName === "") {
      navigate("/phases");
    } else {
      setMode("view");
    }
  }, [isSuccess]);

  useEffect(() => {
    if (!editPhase) return;

    let constantCycle: number | undefined;
    let deltaCycles: DeltaCycle[] | undefined;

    switch (editPhase?.cycleMode) {
      case "constant":
        constantCycle = defaultConstant;
        break;
      case "delta":
        deltaCycles = editPhase.deltaCycles || defaultDeltaCycles();
        break;
    }

    dispatch(
      setEditPhase({
        ...editPhase,
        deltaCycles,
        constantCycle,
      })
    );
  }, [editPhase?.cycleMode]);

  const nameUsed = useMemo(() => {
    if (!editPhase || !phases?.length) {
      return false;
    }

    for (const p of phases) {
      if (p.name === phase.name) {
        continue;
      }

      if (p.name.trim() === editPhase.name.trim()) {
        return true;
      }
    }

    return false;
  }, [phases, editPhase]);

  const editingThis = useMemo(() => mode === "edit", [mode]);

  const isValid = useMemo(() => {
    if (!editPhase) {
      return false;
    }

    const { name, cycleMode, constantCycle, deltaCycles } = editPhase;

    return (
      !nameUsed &&
      validName(name, ["new"]) &&
      ((cycleMode === "constant" && !!constantCycle) ||
        (cycleMode === "delta" && !!deltaCycles?.length))
    );
  }, [editPhase]);

  const updateEdited =
    (field: keyof ApiPhase) => (event: React.ChangeEvent<HTMLInputElement>) => {
      if (editPhase) {
        dispatch(
          setEditPhase({ ...editPhase, [field]: event.currentTarget.value })
        );
      }
    };

  const updateConstantCycle = (constantCycle: ApiPhase["constantCycle"]) => {
    if (editPhase) {
      dispatch(
        setEditPhase({
          ...editPhase,
          constantCycle,
          deltaCycles: undefined,
        })
      );
    }
  };

  const updateDeltaCycles = (deltaCycles?: ApiPhase["deltaCycles"]) => {
    if (editPhase) {
      dispatch(
        setEditPhase({
          ...editPhase,
          deltaCycles,
          constantCycle: undefined,
        })
      );
    }
  };

  const handleEdit = () => {
    dispatch(setEditPhase(phase));
    setMode("edit");
  };

  const normalize = (phase: ApiPhase): ApiPhase => {
    const cpy = { ...phase };
    cpy.name = cpy.name.trim();

    return cpy;
  };

  const handleSave = () => {
    if (!editPhase) {
      return;
    }

    const normalized = normalize(editPhase);
    savePhase(normalized);
  };

  const handleCancel = () => {
    const editName = editPhase?.name;
    dispatch(setEditPhase(undefined));

    if (editName === "") {
      navigate("/phases");
    } else {
      setMode("view");
    }
  };

  if (!name) {
    navigate("/phases");
  }

  return (
    <Stack direction="column" gap={6} width="60rem">
      {!editingThis && (
        <Stack direction="row" justifyContent="end" gap={6}>
          <Button color="primary" onClick={handleEdit}>
            {t("phases.edit")}
          </Button>
        </Stack>
      )}

      <NameComponent
        editing={editingThis}
        name={editingThis ? editPhase?.name : phase.name}
        handleChange={updateEdited("name")}
      />

      <Cycles
        editing={editingThis}
        phase={editingThis && editPhase ? editPhase : phase}
        onChangeCycleMode={updateEdited("cycleMode")}
        onChangeConstantCycle={updateConstantCycle}
        onChangeDeltaCycles={updateDeltaCycles}
      />

      {editingThis && (
        <Stack direction="row" gap="3em" justifyContent="flex-end">
          <Button onClick={handleSave} disabled={!isValid} color="success">
            {t("phases.save")}
          </Button>

          <Button onClick={handleCancel} color="warning">
            {t("phases.cancel")}
          </Button>
        </Stack>
      )}
    </Stack>
  );
};
