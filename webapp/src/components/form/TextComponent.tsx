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
    <Stack direction="row" gap={2} alignItems="center">
      <Typography sx={{ minWidth: 100 }}>{title}</Typography>

      <Stack flex={1}>
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
