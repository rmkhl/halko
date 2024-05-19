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

export const Phase: React.FC = () => {
  const [mode, setMode] = useState<FormMode>("view");
  const [phase, setPhase] = useState<ApiPhase>(emptyConstantPhase());

  const { id } = useParams();
  const { data: phases, isFetching } = useGetPhasesQuery();
  const [savePhase, { isLoading, error, isSuccess }] = useSavePhaseMutation();
  const editPhase = useSelector((state: RootState) => state.phases.edit);
  const navigate = useNavigate();
  const { t } = useTranslation();

  const dispatch = useDispatch();

  useEffect(() => {
    if (id === "new") {
      setMode("edit");

      if (editPhase && !editPhase.id) {
        return;
      }

      dispatch(setEditPhase(emptyConstantPhase()));

      return;
    }

    if (!id || !phases) {
      return;
    }

    const phase = phases.find((p) => p.id === id);

    if (!phase) {
      navigate("/phases");
      return;
    }

    setPhase(phase);
  }, [id, phases]);

  useEffect(() => {
    if (!isSuccess) {
      return;
    }

    const editId = editPhase?.id;
    dispatch(setEditPhase(undefined));

    if (editId === "") {
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
        deltaCycles = defaultDeltaCycles();
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

  const editingThis = useMemo(() => mode === "edit", [mode]);

  const isValid = useMemo(() => {
    if (!editPhase) {
      return false;
    }

    const { name, cycleMode, constantCycle, deltaCycles } = editPhase;

    return (
      !!name &&
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

  const handleSave = () => {
    if (editPhase) {
      savePhase(editPhase);
    }
  };

  const handleCancel = () => {
    const editId = editPhase?.id;
    dispatch(setEditPhase(undefined));

    if (editId === "") {
      navigate("/phases");
    } else {
      setMode("view");
    }
  };

  if (!id) {
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
