import React, { useMemo } from "react";
import { Program as ApiProgram } from "../../types/api";
import { setEditProgram } from "../../store/features/programsSlice";
import { useDispatch, useSelector } from "react-redux";
import { NameComponent } from "../form";
import { RootState } from "../../store/store";
import {
  useGetProgramsQuery,
  useSaveProgramMutation,
} from "../../store/services";
import { validName } from "../../util";
import { emptyProgram } from "./templates";
import { useFormData } from "../../hooks/useFormData";
import { DataForm } from "../form/DataForm";
import { useTranslation } from "react-i18next";
import { Steps } from "./Steps";
import { TimeComponent } from "../form/TimeComponent";

const normalize = (program: ApiProgram): ApiProgram => {
  const cpy = { ...program };
  cpy.name = cpy.name.trim();

  return cpy;
};

export const Program: React.FC = () => {
  const { data } = useGetProgramsQuery();
  const [saveProgram, { isSuccess }] = useSaveProgramMutation();
  const editProgram = useSelector(
    (state: RootState) => state.programs.editRecord
  );

  const programs = useMemo(() => data as ApiProgram[], [data]);

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
    defaultData: emptyProgram(),
    editData: editProgram,
    rootPath: "/programs",
    normalizeData: normalize,
    saveSuccess: isSuccess,
    saveData: saveProgram,
    setEditData: setEditProgram,
  });

  const dispatch = useDispatch();

  const updateEdited =
    <Key extends keyof ApiProgram, Value extends ApiProgram[Key]>(field: Key) =>
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

    return !nameUsed && validName(name, ["new", "latest", "current"]);
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

      <TimeComponent
        editing={editing}
        title={t("programs.defaultStepRuntime")}
        value={editProgram?.defaultStepRuntime || 0}
        onChange={updateEdited("defaultStepRuntime")}
      />

      <TimeComponent
        editing={editing}
        title={t("programs.preheatTo")}
        value={editProgram?.preheatTo || 0}
        onChange={updateEdited("preheatTo")}
      />

      <Steps
        editing={true}
        steps={editProgram?.steps}
        onChange={updateEdited("steps")}
      />
    </DataForm>
  );
};
