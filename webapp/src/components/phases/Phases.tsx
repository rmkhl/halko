import React from "react";
import { useGetPhasesQuery } from "../../store/services";
import { Button, Stack } from "@mui/material";
import { PhaseRow } from "./PhaseRow";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";

export const Phases: React.FC = () => {
  const { data: phases } = useGetPhasesQuery();

  const { t } = useTranslation();
  const navigate = useNavigate();

  const addNew = () => {
    navigate("/phases/new");
  };

  return (
    <Stack>
      <Stack direction="row" justifyContent="end" gap={6}>
        <Button color="success" onClick={addNew}>
          {t("phases.new")}
        </Button>
      </Stack>

      <Stack direction="column" width="60rem">
        {[...(phases || [])]
          .sort((a, b) => a.name.localeCompare(b.name))
          .map((p) => (
            <PhaseRow key={p.id} phase={p} />
          ))}
      </Stack>
    </Stack>
  );
};
