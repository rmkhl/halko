import React from "react";
import { useRouteError, isRouteErrorResponse, useNavigate } from "react-router-dom";
import { Stack, Typography, Button, Paper } from "@mui/material";

export const ErrorBoundary: React.FC = () => {
  const error = useRouteError();
  const navigate = useNavigate();

  let errorMessage: string;
  let errorDetails: string | undefined;

  if (isRouteErrorResponse(error)) {
    errorMessage = `${error.status} ${error.statusText}`;
    errorDetails = error.data?.message;
  } else if (error instanceof Error) {
    errorMessage = error.message;
    errorDetails = error.stack;
  } else if (typeof error === "string") {
    errorMessage = error;
  } else {
    errorMessage = "An unknown error occurred";
  }

  const handleGoHome = () => {
    navigate("/");
  };

  const handleGoBack = () => {
    navigate(-1);
  };

  return (
    <Stack
      alignItems="center"
      justifyContent="center"
      sx={{ minHeight: "100vh", padding: 4, backgroundColor: "#f5f5f5" }}
    >
      <Paper sx={{ padding: 4, maxWidth: 600, width: "100%" }}>
        <Stack spacing={3}>
          <Typography variant="h4" color="error" align="center">
            Oops! Something went wrong
          </Typography>

          <Typography variant="body1" color="text.secondary" align="center">
            {errorMessage}
          </Typography>

          {errorDetails && (
            <Paper
              sx={{
                padding: 2,
                backgroundColor: "#f9f9f9",
                maxHeight: 200,
                overflowY: "auto",
              }}
            >
              <Typography
                variant="body2"
                component="pre"
                sx={{ fontSize: "0.75rem", whiteSpace: "pre-wrap", wordBreak: "break-word" }}
              >
                {errorDetails}
              </Typography>
            </Paper>
          )}

          <Stack direction="row" spacing={2} justifyContent="center">
            <Button variant="outlined" onClick={handleGoBack}>
              Go Back
            </Button>
            <Button variant="contained" onClick={handleGoHome}>
              Go Home
            </Button>
          </Stack>
        </Stack>
      </Paper>
    </Stack>
  );
};
