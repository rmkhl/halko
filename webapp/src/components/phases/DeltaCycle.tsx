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
  size?: "sm" | "lg";
}

export const DeltaCycle: React.FC<Props> = (props) => {
  const { prev, curr, next, onChangeAbove, onChangeBelow, size = "lg" } = props;

  const { above, below, delta } = curr;

  let nextAbove = below;

  if (next) {
    nextAbove = next.above;
  }

  const arrowTransform = (direction: "up" | "down"): React.CSSProperties => {
    const transforms: string[] = [];

    if (!prev) {
      if (direction === "up") {
        transforms.push("rotate(45deg)");
      } else {
        transforms.push("rotate(-45deg)");
      }
    } else if (!next) {
      if (direction === "up") {
        transforms.push("rotate(-45deg)");
      } else {
        transforms.push("rotate(45deg)");
      }
    }

    if (size === "sm") {
      transforms.push("scale(0.5)");
    }

    if (!transforms.length) {
      return {};
    }

    return { transform: transforms.join(" ") };
  };

  return (
    <Stack gap={size === "lg" ? 2 : undefined}>
      {!prev && (
        <Stack direction="row">
          <Stack flex={1} alignItems="center">
            <Cycle
              percentage={above}
              showInfo={false}
              size={size === "lg" ? "sm" : "xs"}
            />
          </Stack>
        </Stack>
      )}

      <Stack direction="row">
        <Stack flex={1} alignItems="center">
          <ArrowUpwardRoundedIcon style={arrowTransform("up")} />
        </Stack>

        <Stack flex={0.25} alignItems="center">
          <Typography fontSize={size === "sm" ? ".75em" : undefined}>
            {celsius(delta)}
          </Typography>
        </Stack>

        <Stack flex={1} alignItems="center">
          <ArrowDownwardRoundedIcon style={arrowTransform("down")} />
        </Stack>
      </Stack>

      <Stack direction="row">
        {!!next ? (
          <>
            <Stack flex={1} alignItems="center">
              <Cycle
                percentage={nextAbove}
                showInfo={false}
                size={size === "lg" ? "sm" : "xs"}
                handleChange={onChangeAbove}
              />
            </Stack>

            <Stack flex={0.25} />

            <Stack flex={1} alignItems="center">
              <Cycle
                percentage={below}
                showInfo={false}
                size={size === "lg" ? "sm" : "xs"}
                handleChange={onChangeBelow}
              />
            </Stack>
          </>
        ) : (
          <Stack flex={1} alignItems="center">
            <Cycle
              percentage={below}
              showInfo={false}
              size={size === "lg" ? "sm" : "xs"}
              handleChange={onChangeBelow}
            />
          </Stack>
        )}
      </Stack>
    </Stack>
  );
};
