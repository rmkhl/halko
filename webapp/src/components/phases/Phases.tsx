import React from "react";
import { useGetPhasesQuery } from "../../store/services";
import { Stack, Typography } from "@mui/material";

export const Phases: React.FC = () => {
  const { data: phases, isFetching } = useGetPhasesQuery();

  return (
    <Stack direction="column" gap={6} width="60rem">
      {phases?.map((p) => (
        <Typography>{p.name}</Typography>
      ))}
    </Stack>
  );
};
