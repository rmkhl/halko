import React from "react";
import { DeltaCycle as ApiDeltaCycle } from "../../types/api";
import { Stack, Typography } from "@mui/material";
import ArrowDownwardRoundedIcon from "@mui/icons-material/ArrowDownwardRounded";
import ArrowUpwardRoundedIcon from "@mui/icons-material/ArrowUpwardRounded";
import { celsius } from "../../util";
import { Cycle } from "../cycles/Cycle";

interface Props {
  prev?: ApiDeltaCycle;
  curr: ApiDeltaCycle;
  next?: ApiDeltaCycle;
  onChangeAbove?: (percentage: number) => void;
  onChangeBelow?: (percentage: number) => void;
}

export const DeltaCycle: React.FC<Props> = (props) => {
  const { prev, curr, next, onChangeAbove, onChangeBelow } = props;

  const { above, below, delta } = curr;

  let nextAbove = below;

  if (next) {
    nextAbove = next.above;
  }

  return (
    <Stack gap={2}>
      {!prev && (
        <Stack direction="row">
          <Stack flex={1} alignItems="center">
            <Cycle percentage={above} showInfo={false} size="sm" />
          </Stack>
        </Stack>
      )}

      <Stack direction="row">
        <Stack flex={1} alignItems="center">
          <ArrowUpwardRoundedIcon
            style={{
              transform: !prev
                ? "rotate(45deg)"
                : !next
                ? "rotate(-45deg)"
                : undefined,
            }}
          />
        </Stack>

        <Stack flex={0.25} alignItems="center">
          <Typography>{celsius(delta)}</Typography>
        </Stack>

        <Stack flex={1} alignItems="center">
          <ArrowDownwardRoundedIcon
            style={{
              transform: !prev
                ? "rotate(-45deg)"
                : !next
                ? "rotate(45deg)"
                : undefined,
            }}
          />
        </Stack>
      </Stack>

      <Stack direction="row">
        {!!next ? (
          <>
            <Stack flex={1} alignItems="center">
              <Cycle
                percentage={nextAbove}
                showInfo={false}
                size="sm"
                handleChange={onChangeAbove}
              />
            </Stack>

            <Stack flex={0.25} />

            <Stack flex={1} alignItems="center">
              <Cycle
                percentage={below}
                showInfo={false}
                size="sm"
                handleChange={onChangeBelow}
              />
            </Stack>
          </>
        ) : (
          <Stack flex={1} alignItems="center">
            <Cycle
              percentage={below}
              showInfo={false}
              size="sm"
              handleChange={onChangeBelow}
            />
          </Stack>
        )}
      </Stack>
    </Stack>
  );
};
