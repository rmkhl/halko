import { Stack, styled } from "@mui/material";

export const ClickableStack = styled(Stack)(() => ({
  cursor: "pointer",
  padding: "1em",
  borderRadius: "1em",
  alignItems: "start",
  "&:hover": {
    backgroundColor: "#666",
  },
}));
