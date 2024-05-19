import React from "react";
import { Cycle } from "../cycles/Cycle";

interface Props {
  editing: boolean;
  onChange: (percentage: number) => void;
  percentage?: number;
}

export const ConstantCycle: React.FC<Props> = (props) => {
  const { editing, percentage, onChange } = props;

  return (
    <>
      {percentage !== undefined && (
        <Cycle percentage={percentage} handleChange={onChange} />
      )}
    </>
  );
};
