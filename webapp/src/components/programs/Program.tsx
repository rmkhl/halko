import React, { useMemo } from "react";
import {
  defaultHeatingStep,
  defaultProgram,
  Step as ApiStep,
  defaultAcclimateStep,
  defaultCoolingStep,
  UIProgram,
} from "../../types/api";
import { setEditProgram } from "../../store/features/programsSlice";
import { useDispatch, useSelector } from "react-redux";
import { NameComponent } from "../form";
import { RootState } from "../../store/store";
import {
  useGetProgramsQuery,
  useSaveProgramMutation,
} from "../../store/services";
import { validName } from "../../util";
import { useFormData } from "../../hooks/useFormData";
import { DataForm } from "../form/DataForm";
import { useTranslation } from "react-i18next";
import { HeatingStep } from "./HeatingStep";
import { AcclimateStep } from "./AcclimateStep";
import { CoolingStep } from "./CoolingStep";

const normalize = (program: UIProgram): UIProgram => {
  const cpy = { ...program };
  cpy.name = cpy.name.trim();

  return cpy;
};

export const Program: React.FC = () => {
  const { data } = useGetProgramsQuery();
  const [saveProgram, { isSuccess }] = useSaveProgramMutation();
  const editProgram =
    useSelector((state: RootState) => state.programs.editRecord) ||
    defaultProgram();

  const programs = useMemo(() => data as UIProgram[], [data]);

  const { t } = useTranslation();

  const {
    editing,
    formData: program,
    nameUsed,
    handleCancel,
    handleEdit,
    handleSave,
  } = useFormData({
    allData: programs,
    defaultData: defaultProgram(),
    editData: editProgram,
    rootPath: "/programs",
    normalizeData: normalize,
    saveSuccess: isSuccess,
    saveData: saveProgram,
    setEditData: setEditProgram,
  });

  const dispatch = useDispatch();

  const updateEdited =
    <Key extends keyof UIProgram, Value extends UIProgram[Key]>(field: Key) =>
    (value: Value) => {
      if (editProgram) {
        dispatch(setEditProgram({ ...editProgram, [field]: value }));
      }
    };

  const updateName = (e: React.ChangeEvent<HTMLInputElement>) =>
    updateEdited("name")(e.currentTarget.value);

  const isValid = useMemo(() => {
    if (!editProgram) {
      return false;
    }

    const { name } = editProgram;

    if (nameUsed || !validName(name, ["new", "latest", "current"]))
      return false;

    return true;
  }, [editProgram]);

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
        name={editing ? editProgram?.name : program.name}
        handleChange={updateName}
      />

      <HeatingStep
        step={editing ? editProgram?.heatingStep : program.heatingStep}
        onChange={updateEdited("heatingStep")}
      />

      <AcclimateStep
        step={editing ? editProgram?.acclimateStep : program.acclimateStep}
        onChange={updateEdited("acclimateStep")}
      />

      <CoolingStep
        step={editing ? editProgram?.coolingStep : program.coolingStep}
        onChange={updateEdited("coolingStep")}
      />
    </DataForm>
  );
};
