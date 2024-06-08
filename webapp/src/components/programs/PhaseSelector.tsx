import React, { useEffect, useState } from "react";
import { Phase } from "../../types/api";
import { PhaseRow } from "../phases/PhaseRow";
import { Button, Stack, Typography } from "@mui/material";
import { useTranslation } from "react-i18next";
import { Dialog } from "../form/Dialog";
import { Phases } from "../phases/Phases";

interface Props {
  editing?: boolean;
  title: string;
  phase?: Phase;
  onChange: (phase?: Phase) => void;
}

export const PhaseSelector: React.FC<Props> = (props) => {
  const [modalOpen, setModalOpen] = useState(false);
  const { title, phase, editing, onChange } = props;
  const { t } = useTranslation();

  useEffect(() => {
    setModalOpen(false);
  }, [phase]);

  return (
    <>
      <Dialog
        title={t("programs.steps.selectPhase")}
        handleClose={() => setModalOpen(false)}
        open={modalOpen}
      >
        <Phases canAddNew={false} onSelectRow={onChange} />
      </Dialog>

      <Stack direction="row" alignItems="center" justifyContent="space-between">
        <Stack flex={1}>
          <Typography>{title}</Typography>
        </Stack>

        <Stack flex={3}>
          {editing ? (
            phase ? (
              <PhaseRow phase={phase} onSelectRow={() => setModalOpen(true)} />
            ) : (
              <Button onClick={() => setModalOpen(true)}>
                {t("programs.steps.selectPhase")}
              </Button>
            )
          ) : phase ? (
            <PhaseRow phase={phase} selectable={false} />
          ) : (
            <Typography>{t("programs.steps.noPhaseSelected")}</Typography>
          )}{" "}
        </Stack>
      </Stack>
    </>
  );
};
