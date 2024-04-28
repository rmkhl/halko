import React, { useMemo, useState } from "react";
import { Cycle as ApiCycle } from "../../types/api";
import { useGetCyclesQuery } from "../../store/services";
import { useTranslation } from "react-i18next";
import { Box, Button, styled } from "@mui/material";
import { Cycle } from "../cycles/Cycle";
import { Dialog } from "../form/Dialog";

interface Props {
  onSelect: (cycle: ApiCycle) => void;
}

export const CycleSelector: React.FC<Props> = (props) => {
  const { onSelect } = props;
  const { data: cycles, isFetching } = useGetCyclesQuery();
  const { t } = useTranslation();
  const selectStr = useMemo(() => t("phases.cycles.select"), [t]);

  const [show, setShow] = useState(false);

  const selectCycle = (c: ApiCycle) => {
    onSelect(c);
    setShow(false);
  };

  return (
    <>
      <Button color="primary" onClick={() => setShow(true)}>
        {selectStr}
      </Button>

      <Dialog title={selectStr} handleClose={() => setShow(false)} open={show}>
        {cycles?.map((c) => (
          <CycleBox
            key={`phase-constant-cycle-${c.name}`}
            onClick={() => selectCycle(c)}
          >
            <Cycle cycle={c} />
          </CycleBox>
        ))}
      </Dialog>
    </>
  );
};

const CycleBox = styled(Box)(() => ({
  "&:hover": {
    backgroundColor: "#666",
  },
}));
