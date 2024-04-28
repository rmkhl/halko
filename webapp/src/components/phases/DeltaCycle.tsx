import React from "react";
import { DeltaCycle as ApiDeltaCycle } from "../../types/api";
import { Cycle } from "../cycles/Cycle";
import { Stack, Typography } from "@mui/material";
import ArrowDownwardRoundedIcon from "@mui/icons-material/ArrowDownwardRounded";
import ArrowUpwardRoundedIcon from "@mui/icons-material/ArrowUpwardRounded";
import { celsius } from "../../util";

interface Props {
  editing?: boolean;
  deltaCycle: ApiDeltaCycle;
}

export const DeltaCycle: React.FC<Props> = (props) => {
  const { editing, deltaCycle } = props;
  const { above, below, delta } = deltaCycle;

  return (
    <Stack>
      <Stack direction="row">
        <Stack flex={1} alignItems="center">
          <Cycle cycle={below} showInfo={false} size="sm" />
        </Stack>

        <Stack flex={0.25} />

        <Stack flex={1} />
      </Stack>

      <Stack direction="row">
        <Stack flex={1} alignItems="center">
          <ArrowUpwardRoundedIcon />
        </Stack>

        <Stack flex={0.25} alignItems="center">
          <Typography>{celsius(delta)}</Typography>
        </Stack>

        <Stack flex={1} alignItems="center">
          <ArrowDownwardRoundedIcon />
        </Stack>
      </Stack>

      <Stack direction="row">
        <Stack flex={1} />

        <Stack flex={0.25} />

        <Stack flex={1} alignItems="center">
          <Cycle cycle={above} showInfo={false} size="sm" />
        </Stack>
      </Stack>
    </Stack>
  );
};
