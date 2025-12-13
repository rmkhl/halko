import React from "react";
import { Program as ApiProgram } from "../../types/api";
import { Button, Stack, Typography } from "@mui/material";
import { useNavigate } from "react-router-dom";
import { ClickableStack } from "../ClickableStack";
import { useStartProgramMutation } from "../../store/services/controlunitApi";

interface Props {
  program: ApiProgram;
}

export const ProgramRow: React.FC<Props> = (props) => {
  const { program } = props;
  const navigate = useNavigate();

  const [startProgram, { isLoading }] = useStartProgramMutation();

  return (
    <ClickableStack direction="row" justifyContent="space-between">
      <Stack onClick={() => navigate(`/programs/${program.name}`)}>
        <Typography variant="h6">{program.name}</Typography>
      </Stack>

      <Button onClick={() => startProgram(program)} disabled={isLoading}>
        Start
      </Button>
    </ClickableStack>
  );
};
