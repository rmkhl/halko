import { Input, Typography } from "@mui/material";
import React from "react";
import { useTranslation } from "react-i18next";

interface Props {
  name?: string;
  editing: boolean;
  placeholder?: string;
  handleChange: (event: React.ChangeEvent<HTMLInputElement>) => void;
}

export const NameComponent: React.FC<Props> = (props) => {
  const { editing, name, handleChange, placeholder } = props;
  const { t } = useTranslation();

  const ph = placeholder || t("common.name");

  return editing ? (
    <Input
      fullWidth
      style={{ fontSize: "2em" }}
      value={name}
      onChange={handleChange}
      placeholder={ph}
    />
  ) : (
    <Typography variant="h4">{name}</Typography>
  );
};
