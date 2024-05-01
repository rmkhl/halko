import React from "react";
import { Cycle as ApiCycle } from "../../types/api";
import { Cycle } from "../cycles/Cycle";
import { CycleSelector } from "./CycleSelector";

interface Props {
  editing: boolean;
  cycle?: ApiCycle;
  onChange: (cycle: ApiCycle) => void;
}

export const ConstantCycle: React.FC<Props> = (props) => {
  const { editing, cycle, onChange } = props;

  return (
    <>
      {editing && <CycleSelector onSelect={onChange} />}

      {!!cycle && <Cycle cycle={cycle} />}
    </>
  );
};
