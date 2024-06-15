import { Input, Stack, Typography } from "@mui/material";
import React from "react";

interface Props {
  title: string;
  editing?: boolean;
  value: string;
  onChange: (value: string) => void;
}

export const TextComponent: React.FC<Props> = (props) => {
  const { title, editing, value, onChange } = props;

  return (
    <Stack direction="row" alignItems="center" justifyContent="space-between">
      <Stack flex={1}>
        <Typography>{title}</Typography>
      </Stack>

      <Stack flex={3}>
        {editing ? (
          <Input
            value={value}
            onChange={(e) => onChange(e.currentTarget.value)}
          />
        ) : (
          <Typography>{value}</Typography>
        )}
      </Stack>
    </Stack>
  );
};
