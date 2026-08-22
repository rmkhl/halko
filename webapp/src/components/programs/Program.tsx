import React, { useMemo, useEffect, useRef } from "react";
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
import {
  Stack,
  Paper,
  Typography,
  Button,
  Divider,
  Alert,
  Box,
  TextField,
  Checkbox,
  FormControlLabel,
  FormHelperText,
} from "@mui/material";
import { useTranslation } from "react-i18next";
import { useGetDefaultsQuery } from "../../store/services/controlunitApi";
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

  const delta = editProgram.equalize?.delta;
  if (delta !== undefined && delta <= 0) {
    errors.push("Equalize delta must be greater than zero");
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
  const { t } = useTranslation();
  const { data: engineDefaults } = useGetDefaultsQuery();

  // All hooks and variables declared once at the top
  const { name } = useParams();
  const { data } = useGetProgramQuery(name || "", { skip: !name || name === "new", refetchOnMountOrArgChange: true });
  const [saveProgram, { isSuccess }] = useSaveProgramMutation();
  const editProgram = useSelector((state: RootState) => state.programs.editRecord);
  const dispatch = useDispatch();
  const navigate = useNavigate();

  // useFormData must be above any useMemo that uses nameUsed
  const {
    editing,
    nameUsed,
    handleCancel,
    handleSave,
  } = useFormData({
    allData: data ? [data as ApiProgram] : [],
    defaultData: emptyProgram(),
    editData: editProgram,
    rootPath: "/programs",
    normalizeData: normalize,
    saveSuccess: isSuccess,
    saveData: saveProgram,
    setEditData: setEditProgram,
  });

  // Place this just before return:
  useEffect(() => {
    if (isSuccess) {
      dispatch(setEditProgram(undefined));
      navigate("/programs");
    }
  }, [isSuccess, dispatch, navigate]);

  const program = useMemo(() => {
    if (!data) return undefined;
    if (typeof data === 'object' && data !== null && 'data' in data) {
      return (data as { data: ApiProgram }).data;
    }
    return data as ApiProgram;
  }, [data]);


  // ...existing code...

  // Ensure editRecord is set from loaded program if missing or mismatched
  useEffect(() => {
    if (typeof program === 'undefined') return;
    if (name && name !== "new") {
      if (!editProgram || editProgram.name !== program.name) {
        dispatch(setEditProgram(program));
      }
    }
    if (name === "new" && (!editProgram || editProgram.name)) {
      // For new, clear editRecord if it has a name (should be empty)
      dispatch(setEditProgram(emptyProgram()));
    }
  }, [name, program, editProgram, dispatch]);





  // Place isValid here, after all hooks/vars
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

  // Ensure editRecord is set from loaded program if missing or mismatched
  useEffect(() => {
    if (typeof program === 'undefined') return;
    if (name && name !== "new") {
      if (!editProgram || editProgram.name !== program.name) {
        dispatch(setEditProgram(program));
      }
    }
    if (name === "new" && (!editProgram || editProgram.name)) {
      // For new, clear editRecord if it has a name (should be empty)
      dispatch(setEditProgram(emptyProgram()));
    }
  }, [name, program, editProgram, dispatch]);


  // Fill in whatever startup settings the program does not name itself, so the
  // editor always shows the values the run will actually use. Programs saved
  // before the startup steps existed carry no equalize block at all, and a
  // checkbox has no way to render "not set" honestly.
  //
  // Seed once per program the editor opens: re-running it would fight the user,
  // putting the default straight back into the delta field the moment they
  // clear it to type a new value.
  const seededFor = useRef<string | null>(null);
  useEffect(() => {
    if (!engineDefaults || !editProgram || !name) return;
    // For an existing program, wait until the fetched record is the one being
    // edited - seeding a leftover record would burn the one seeding pass.
    if (name !== "new" && editProgram.name !== name) return;
    if (seededFor.current === name) return;

    seededFor.current = name;

    const equalize = editProgram.equalize;
    if (equalize?.delta !== undefined && equalize?.steam_prewarm !== undefined) {
      return;
    }

    dispatch(
      setEditProgram({
        ...editProgram,
        equalize: {
          delta: equalize?.delta ?? engineDefaults.equalize.delta,
          steam_prewarm:
            equalize?.steam_prewarm ?? engineDefaults.equalize.steam_prewarm,
        },
      })
    );
  }, [engineDefaults, editProgram, name, dispatch]);

  const updateEdited =
    <Key extends keyof ApiProgram, Value extends ApiProgram[Key]>(field: Key) =>
    (value: Value) => {
      if (editProgram) {
        dispatch(setEditProgram({ ...editProgram, [field]: value }));
      }
    };

  const updateName = (e: React.ChangeEvent<HTMLInputElement>) =>
    updateEdited("name")(e.currentTarget.value);

  const validationErrors = useMemo(() => {
    return getValidationErrors(editProgram ?? null, nameUsed);
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

            {editing && (
              <Paper variant="outlined" sx={{ padding: 2 }}>
                <Typography variant="subtitle1" sx={{ marginBottom: 2, fontWeight: "bold" }}>
                  {t("programs.equalize.title")}
                </Typography>
                <Stack gap={2}>
                  {!engineDefaults && (
                    <Alert severity="warning">
                      {t("programs.equalize.defaultsUnavailable")}
                    </Alert>
                  )}
                  <TextField
                    label={t("programs.equalize.delta")}
                    helperText={t("programs.equalize.deltaHelp")}
                    type="number"
                    size="small"
                    value={editProgram?.equalize?.delta ?? ""}
                    onChange={(e) =>
                      updateEdited("equalize")({
                        ...editProgram?.equalize,
                        delta:
                          e.currentTarget.value === ""
                            ? undefined
                            : Number(e.currentTarget.value),
                      })
                    }
                  />
                  <Box>
                    <FormControlLabel
                      control={
                        <Checkbox
                          checked={editProgram?.equalize?.steam_prewarm ?? false}
                          onChange={(e) =>
                            updateEdited("equalize")({
                              ...editProgram?.equalize,
                              steam_prewarm: e.currentTarget.checked,
                            })
                          }
                        />
                      }
                      label={t("programs.equalize.steamPrewarm")}
                    />
                    <FormHelperText>
                      {t("programs.equalize.steamPrewarmHelp")}
                    </FormHelperText>
                  </Box>
                </Stack>
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
