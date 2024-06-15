import React from "react";
import { Translation } from "../../i18n";
import {
  NumberComponent,
  Props as NumberComponentProps,
} from "./NumberComponent";
import { Stack, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";

interface Props extends NumberComponentProps {
  timeUnit?: keyof Translation["time"];
}

export const TimeComponent: React.FC<Props> = (props) => {
  const { timeUnit = "seconds", ...rest } = props;
  const { t } = useTranslation();

  return (
    <Stack direction="row" alignItems="center" gap={3}>
      <NumberComponent {...rest}>
        <Typography>{t(`time.${timeUnit}`)}</Typography>
      </NumberComponent>
    </Stack>
  );
};
