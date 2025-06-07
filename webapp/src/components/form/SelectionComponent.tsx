import { MenuItem, Select, Stack, Typography } from "@mui/material";
import React from "react";

interface Props {
  title: string;
  editing?: boolean;
  value: string;
  options: string[];
  onChange: (value: string) => void;
}

export const SelectionComponent: React.FC<Props> = (props) => {
  const { title, value, editing, options, onChange } = props;

  return (
    <Stack direction="row" alignItems="center" justifyContent="space-between">
      <Stack flex={1}>
        <Typography>{title}</Typography>
      </Stack>

      <Stack flex={3}>
        {editing ? (
          <Select
            value={value}
            onChange={(event) => onChange(event.target.value)}
          >
            {options.map((s) => (
              <MenuItem key={`option-${s}`} value={s}>
                {s}
              </MenuItem>
            ))}
          </Select>
        ) : (
          <Typography>{value}</Typography>
        )}
      </Stack>
    </Stack>
  );
};
