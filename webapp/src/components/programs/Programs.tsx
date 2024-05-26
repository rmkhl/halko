import React, { useMemo } from "react";
import { useGetProgramsQuery } from "../../store/services";
import { useTranslation } from "react-i18next";
import { Program } from "../../types/api";
import { useNavigate } from "react-router-dom";
import { Stack } from "@mui/system";
import { Button } from "@mui/material";

export const Programs: React.FC = () => {
  const { data } = useGetProgramsQuery();

  const { t } = useTranslation();
  const navigate = useNavigate();

  const programs = useMemo(() => data as Program[], [data]);

  const addNew = () => {
    navigate("/programs/new");
  };

  return (
    <Stack>
      <Stack direction="row" justifyContent="end" gap={6}>
        <Button color="success" onClick={addNew}>
          {t("programs.new")}
        </Button>
      </Stack>

      <Stack direction="column" width=" 60rem">
        {[...(programs || [])]
          .sort((a, b) => a.name.localeCompare(b.name))
          .map((p) => null)}
      </Stack>
    </Stack>
  );
};
