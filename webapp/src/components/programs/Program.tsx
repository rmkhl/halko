import React from "react";
import { Program as ApiProgram } from "../../types/api";
import { Typography } from "@mui/material";

interface Props {
  program: ApiProgram;
}

export const Program: React.FC<Props> = (props) => {
  const { program } = props;

  return <Typography>{program.name}</Typography>;
};
