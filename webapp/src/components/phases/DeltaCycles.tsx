import React, { useMemo, useState } from "react";
import { DeltaCycle as ApiDeltaCycle } from "../../types/api";
import { DeltaCycle } from "./DeltaCycle";
import { Button, Stack } from "@mui/material";
import { useTranslation } from "react-i18next";
import { Dialog } from "../form/Dialog";
import { AddDeltaCycle } from "./AddDeltaCycle";

interface Props {
  editing?: boolean;
  deltaCycles?: ApiDeltaCycle[];
  onChange: (cycles: ApiDeltaCycle[]) => void;
}

export const DeltaCycles: React.FC<Props> = (props) => {
  const { editing, deltaCycles, onChange } = props;
  const { t } = useTranslation();
  const addDeltaCycleStr = useMemo(() => t("phases.cycles.addDeltaCycle"), [t]);

  const [showAddDeltaCycleDialog, setShowAddDeltaCycleDialog] = useState(false);

  const addDeltaCycle = (deltaCycle: ApiDeltaCycle) => {
    const updatedDeltaCycles = [...(deltaCycles || []), deltaCycle].sort(
      (a, b) => b.delta - a.delta
    );

    onChange(updatedDeltaCycles);
    setShowAddDeltaCycleDialog(false);
  };

  return (
    <>
      {editing && (
        <Stack alignItems="center">
          <Button
            onClick={() => setShowAddDeltaCycleDialog(true)}
            style={{ width: "fit-content" }}
          >
            {addDeltaCycleStr}
          </Button>
        </Stack>
      )}

      <Stack>
        {deltaCycles?.map((d) => (
          <DeltaCycle key={`deltaCycle-${d.delta}`} deltaCycle={d} />
        ))}
      </Stack>

      <Dialog
        open={showAddDeltaCycleDialog}
        title={addDeltaCycleStr}
        handleClose={() => setShowAddDeltaCycleDialog(false)}
      >
        <AddDeltaCycle
          existingDeltaCycles={deltaCycles}
          onSelect={addDeltaCycle}
        />
      </Dialog>
    </>
  );
};
