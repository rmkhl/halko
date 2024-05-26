import React, { useEffect, useMemo } from "react";
import { useGetRunningProgramQuery } from "../../store/services/executorApi";
import { Program as ApiProgram } from "../../types/api";
import { Program } from "./Program";

export const Current: React.FC = () => {
  const { data } = useGetRunningProgramQuery();

  const currentProgram = useMemo(
    () => (data ? (data as ApiProgram) : undefined),
    [data]
  );

  useEffect(() => {
    console.log("data", data);
  }, [data]);

  return currentProgram ? <Program program={currentProgram} /> : null;
};
