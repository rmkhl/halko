import React, { useMemo } from "react";
import { useGetPhasesQuery } from "../../store/services";
import { Button, Stack } from "@mui/material";
import { PhaseRow } from "./PhaseRow";
import { useTranslation } from "react-i18next";
import { useNavigate } from "react-router-dom";
import { Phase } from "../../types/api";

export const Phases: React.FC = () => {
  const { data } = useGetPhasesQuery();

  const { t } = useTranslation();
  const navigate = useNavigate();

  const phases = useMemo(() => data as Phase[], [data]);

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
            <PhaseRow key={p.name} phase={p} />
          ))}
      </Stack>
    </Stack>
  );
};
