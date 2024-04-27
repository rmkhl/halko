import React, { useEffect, useMemo, useState } from "react";
import { useNavigate, useParams } from "react-router-dom";
import { useGetPhasesQuery } from "../../store/services";
import { Phase as ApiPhase } from "../../types/api";
import {
  FormControl,
  FormControlLabel,
  Radio,
  RadioGroup,
  Stack,
  Typography,
} from "@mui/material";
import { FormMode } from "../../types";
import { useDispatch, useSelector } from "react-redux";
import { RootState } from "../../store/store";
import { setEditPhase } from "../../store/features/phasesSlice";
import { NameComponent } from "../form";
import { ValidRange } from "./ValidRange";
import { useTranslation } from "react-i18next";
import { Cycles } from "./Cycles";

const emptyPhase: ApiPhase = {
  id: "",
  name: "",
  validRange: [
    {
      sensor: "oven",
      above: 20,
      below: 100,
    },
    {
      sensor: "material",
      above: 20,
      below: 100,
    },
  ],
  cycleMode: "constant",
};

export const Phase: React.FC = () => {
  const [mode, setMode] = useState<FormMode>("view");
  const [phase, setPhase] = useState<ApiPhase>({ ...emptyPhase });

  const { id } = useParams();
  const { data: phases, isFetching } = useGetPhasesQuery();
  const editPhase = useSelector((state: RootState) => state.phases.edit);
  const navigate = useNavigate();
  const { t } = useTranslation();

  const dispatch = useDispatch();

  const updateEdited =
    (field: keyof ApiPhase) => (event: React.ChangeEvent<HTMLInputElement>) => {
      if (editPhase) {
        dispatch(
          setEditPhase({ ...editPhase, [field]: event.currentTarget.value })
        );
      }
    };

  const updateValidRange = (validRange: ApiPhase["validRange"]) => {
    if (editPhase) {
      dispatch(
        setEditPhase({
          ...editPhase,
          validRange,
        })
      );
    }
  };

  useEffect(() => {
    if (id === "new") {
      setMode("edit");

      if (editPhase && !editPhase.id) {
        return;
      }

      dispatch(setEditPhase(emptyPhase));

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

  const editingThis = useMemo(() => mode === "edit", [mode]);

  if (!id) {
    navigate("/phases");
  }

  return (
    <Stack direction="column" gap={6}>
      <NameComponent
        editing={editingThis}
        name={editingThis ? editPhase?.name : phase.name}
        handleChange={updateEdited("name")}
      />

      <ValidRange
        editing={editingThis}
        validRange={
          editingThis && editPhase ? editPhase?.validRange : phase.validRange
        }
        handleChange={updateValidRange}
      />

      <Cycles
        editing={editingThis}
        cycleMode={editPhase?.cycleMode}
        onChange={updateEdited("cycleMode")}
      />
    </Stack>
  );
};
