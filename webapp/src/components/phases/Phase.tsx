import React, { useEffect, useMemo } from "react";
import { useGetPhasesQuery, useSavePhaseMutation } from "../../store/services";
import { Phase as ApiPhase, DeltaCycle } from "../../types/api";
import { useDispatch, useSelector } from "react-redux";
import { RootState } from "../../store/store";
import { setEditPhase } from "../../store/features/phasesSlice";
import { NameComponent } from "../form";
import { Cycles } from "./Cycles";
import {
  defaultConstant,
  defaultDeltaCycles,
  emptyConstantPhase,
} from "./templates";
import { validName } from "../../util";
import { useFormData } from "../../hooks/useFormData";
import { DataForm } from "../form/DataForm";

const normalize = (phase: ApiPhase): ApiPhase => {
  const cpy = { ...phase };
  cpy.name = cpy.name.trim();

  return cpy;
};

export const Phase: React.FC = () => {
  const { data } = useGetPhasesQuery();
  const [savePhase, { isSuccess }] = useSavePhaseMutation();
  const editPhase = useSelector((state: RootState) => state.phases.editRecord);

  const phases = useMemo(() => data as ApiPhase[], [data]);

  const {
    editing,
    formData: phase,
    nameUsed,
    handleCancel,
    handleEdit,
    handleSave,
  } = useFormData({
    allData: phases,
    defaultData: emptyConstantPhase(),
    editData: editPhase,
    rootPath: "/phases",
    saveSuccess: isSuccess,
    normalizeData: normalize,
    saveData: savePhase,
    setEditData: setEditPhase,
  });

  const dispatch = useDispatch();

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

  return (
    <DataForm
      editing={editing}
      isValid={isValid}
      handleCancel={handleCancel}
      handleEdit={handleEdit}
      handleSave={handleSave}
    >
      <NameComponent
        editing={editing}
        name={editing ? editPhase?.name : phase.name}
        handleChange={updateEdited("name")}
      />

      <Cycles
        editing={editing}
        phase={editing && editPhase ? editPhase : phase}
        onChangeCycleMode={updateEdited("cycleMode")}
        onChangeConstantCycle={updateConstantCycle}
        onChangeDeltaCycles={updateDeltaCycles}
      />
    </DataForm>
  );
};
