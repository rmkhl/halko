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
import { Steps } from "./Steps";
import { useParams, useNavigate } from "react-router-dom";
import { Stack, Paper, Typography, Button, Divider, Alert, Box } from "@mui/material";
import SaveIcon from "@mui/icons-material/Save";
import CancelIcon from "@mui/icons-material/Cancel";

const normalize = (program: ApiProgram): ApiProgram => {
  const cpy = { ...program };
  cpy.name = cpy.name.trim();

  return cpy;
};

const getValidationErrors = (editProgram: ApiProgram | null, nameUsed: boolean): string[] => {
  const errors: string[] = [];

  if (!editProgram) return errors;

  const { name, steps } = editProgram;

  if (!name || name.trim() === "") {
    errors.push("Program name is required");
  } else if (nameUsed) {
    errors.push("Program name already exists");
  } else if (!validName(name, ["new", "latest", "current"])) {
    errors.push("Invalid program name (avoid reserved words: new, latest, current)");
  }

  if (!steps || steps.length === 0) {
    errors.push("At least one step is required");
  } else {
    steps.forEach((step, idx) => {
      if (!step.name || step.name.trim() === "") {
        errors.push(`Step ${idx + 1}: Name is required`);
      }
      if (!step.type) {
        errors.push(`Step ${idx + 1}: Type is required`);
      }
      if (!step.temperature_target || step.temperature_target <= 0) {
        errors.push(`Step ${idx + 1}: Valid temperature target is required`);
      }
    });
  }

  return errors;
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

  const validationErrors = useMemo(() => {
    return getValidationErrors(editProgram, nameUsed);
  }, [editProgram, nameUsed]);

  return (
    <Box
      sx={{
        display: "flex",
        width: "100%",
        height: "calc(100vh - 120px)",
        padding: 2,
        justifyContent: "center",
      }}
    >
      <Paper
        sx={{
          width: "100%",
          maxWidth: "1200px",
          padding: 3,
          display: "flex",
          flexDirection: "column",
          overflow: "hidden",
        }}
      >
        {/* Header */}
        <Box sx={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: 2 }}>
          <Typography variant="h5">
            {name === "new" ? "New Program" : "Edit Program"}
          </Typography>
          <Box sx={{ display: "flex", gap: 2 }}>
            <Button
              variant="outlined"
              startIcon={<CancelIcon />}
              onClick={handleCancel}
              color="inherit"
            >
              Cancel
            </Button>
            <Button
              variant="contained"
              startIcon={<SaveIcon />}
              onClick={handleSave}
              disabled={!isValid}
              color="primary"
            >
              Save
            </Button>
          </Box>
        </Box>

        <Divider sx={{ marginBottom: 2 }} />

        {/* Validation Errors */}
        {editing && validationErrors.length > 0 && (
          <Alert severity="error" sx={{ marginBottom: 2 }}>
            <Typography variant="subtitle2" sx={{ fontWeight: "bold", marginBottom: 1 }}>
              Please fix the following issues:
            </Typography>
            <ul style={{ margin: 0, paddingLeft: 20 }}>
              {validationErrors.map((error, idx) => (
                <li key={idx}>{error}</li>
              ))}
            </ul>
          </Alert>
        )}

        {/* Content */}
        <Box sx={{ flex: 1, overflow: "auto", paddingRight: 1 }}>
          <Stack gap={3} sx={{ paddingBottom: 4 }}>
            {editing && (
              <Paper variant="outlined" sx={{ padding: 2 }}>
                <Typography variant="subtitle1" sx={{ marginBottom: 2, fontWeight: "bold" }}>
                  Program Name
                </Typography>
                <NameComponent
                  editing={editing}
                  name={editProgram?.name}
                  handleChange={updateName}
                />
              </Paper>
            )}

            <Box>
              <Typography variant="h6" sx={{ marginBottom: 2 }}>
                Steps ({editProgram?.steps?.length || 0})
              </Typography>
              <Steps
                editing={true}
                steps={editProgram?.steps}
                onChange={updateEdited("steps")}
              />
            </Box>
          </Stack>
        </Box>
      </Paper>
    </Box>
  );
};
