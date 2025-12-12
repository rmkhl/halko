import React, { useMemo } from "react";
import { Program as ApiProgram } from "../../types/api";
import { setEditProgram } from "../../store/features/programsSlice";
import { useDispatch, useSelector } from "react-redux";
import { NameComponent } from "../form";
import { RootState } from "../../store/store";
import {
  useGetProgramQuery,
  useSaveProgramMutation,
} from "../../store/services";
import { validName } from "../../util";
import { emptyProgram } from "./templates";
import { useFormData } from "../../hooks/useFormData";
import { DataForm } from "../form/DataForm";
import { Steps } from "./Steps";
import { useParams, useNavigate } from "react-router-dom";
import { Stack } from "@mui/material";

const normalize = (program: ApiProgram): ApiProgram => {
  const cpy = { ...program };
  cpy.name = cpy.name.trim();

  return cpy;
};

export const Program: React.FC = () => {
  const { name } = useParams();
  const navigate = useNavigate();
  const { data } = useGetProgramQuery(name || "", { skip: !name || name === "new" });
  const [saveProgram, { isSuccess }] = useSaveProgramMutation();
  const editProgram = useSelector(
    (state: RootState) => state.programs.editRecord
  );

  const program = useMemo(() => {
    if (!data) return undefined;
    const responseData = data as any;
    if (responseData.data) {
      return responseData.data as ApiProgram;
    }
    return data as ApiProgram;
  }, [data]);

  const {
    editing,
    formData: displayProgram,
    nameUsed,
    handleCancel,
    handleEdit,
    handleSave,
  } = useFormData({
    allData: program ? [program] : [],
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

    const { name, steps } = editProgram;

    if (nameUsed || !validName(name, ["new", "latest", "current"]))
      return false;

    if (!steps.length) return false;

    for (const step of steps) {
      if (!step.name || !step.type) {
        return false;
      }
      // Power settings are optional - backend applies defaults
    }

    return true;
  }, [editProgram, nameUsed]);

  const handleRun = () => {
    // TODO: Implement run functionality
    console.log("Run program:", displayProgram?.name);
  };

  const handleBack = () => {
    navigate("/programs");
  };

  return (
    <Stack alignItems="center" sx={{ width: "100%", height: "100%", overflow: "hidden" }}>
      <DataForm
        editing={editing}
        isValid={isValid}
        programName={displayProgram?.name}
        handleCancel={handleCancel}
        handleEdit={handleEdit}
        handleSave={handleSave}
        handleRun={handleRun}
        handleBack={handleBack}
      >
        {editing && (
          <NameComponent
            editing={editing}
            name={editProgram?.name}
            handleChange={updateName}
          />
        )}

        <Steps
          editing={editing}
          steps={editing ? editProgram?.steps : displayProgram?.steps}
          onChange={updateEdited("steps")}
        />
      </DataForm>
    </Stack>
  );
};
