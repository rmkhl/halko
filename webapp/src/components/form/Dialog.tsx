import {
  Dialog as MuiDialog,
  DialogProps,
  DialogTitle,
  IconButton,
  DialogContent,
} from "@mui/material";
import CloseIcon from "@mui/icons-material/Close";
import React from "react";

interface Props extends DialogProps {
  title: string;
  handleClose: () => void;
}

export const Dialog: React.FC<Props> = (props) => {
  const { children, handleClose, title, fullScreen = true, ...rest } = props;

  return (
    <MuiDialog onClose={handleClose} fullScreen={fullScreen} {...rest}>
      <DialogTitle>{title}</DialogTitle>

      <IconButton
        aria-label="close"
        onClick={handleClose}
        sx={{
          position: "absolute",
          right: 8,
          top: 8,
          color: (theme) => theme.palette.grey[500],
        }}
      >
        <CloseIcon />
      </IconButton>

      <DialogContent dividers>{children}</DialogContent>
    </MuiDialog>
  );
};
