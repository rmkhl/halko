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
    (field: keyof ApiProgram) =>
    (event: React.ChangeEvent<HTMLInputElement>) => {
      if (editProgram) {
        dispatch(
          setEditProgram({ ...editProgram, [field]: event.currentTarget.value })
        );
      }
    };

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
        handleChange={updateEdited("name")}
      />
    </DataForm>
  );
};
