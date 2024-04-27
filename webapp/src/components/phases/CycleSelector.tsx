import React, { useState } from "react";
import { Cycle as ApiCycle } from "../../types/api";
import { useGetCyclesQuery } from "../../store/services";
import { useTranslation } from "react-i18next";
import {
  Box,
  Button,
  Dialog,
  DialogContent,
  DialogTitle,
  IconButton,
  styled,
} from "@mui/material";
import CloseIcon from "@mui/icons-material/Close";
import { Cycle } from "../cycles/Cycle";

interface Props {
  onSelect: (cycle: ApiCycle) => void;
}

export const CycleSelector: React.FC<Props> = (props) => {
  const { onSelect } = props;
  const { data: cycles, isFetching } = useGetCyclesQuery();
  const { t } = useTranslation();

  const [show, setShow] = useState(false);

  const selectCycle = (c: ApiCycle) => {
    onSelect(c);
    setShow(false);
  };

  return (
    <>
      <Button color="primary" onClick={() => setShow(true)}>
        {t("phases.cycles.select")}
      </Button>

      <Dialog open={show} onClose={() => setShow(false)}>
        <DialogTitle>{t("phases.cycles.select")}</DialogTitle>

        <IconButton
          aria-label="close"
          onClick={() => setShow(false)}
          sx={{
            position: "absolute",
            right: 8,
            top: 8,
            color: (theme) => theme.palette.grey[500],
          }}
        >
          <CloseIcon />
        </IconButton>

        <DialogContent dividers>
          {cycles?.map((c) => (
            <CycleBox
              key={`phase-constant-cycle-${c.name}`}
              onClick={() => selectCycle(c)}
            >
              <Cycle cycle={c} />
            </CycleBox>
          ))}
        </DialogContent>
      </Dialog>
    </>
  );
};

const CycleBox = styled(Box)(({ theme }) => ({
  "&:hover": {
    backgroundColor: "#666",
  },
}));
