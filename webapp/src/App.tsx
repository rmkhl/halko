import { Paper, ThemeProvider, Typography } from "@mui/material";
import React, { useMemo } from "react";
import { useTranslation } from "react-i18next";
import {
  Navigate,
  RouterProvider,
  createBrowserRouter,
} from "react-router-dom";
import { Route } from "./types";
import { Navigation } from "./components/Navigation";
import { theme } from "./material-ui/theme";

const getRouter = (routes: Route[]) =>
  createBrowserRouter([
    {
      path: "/",
      element: <Navigation routes={routes} />,
      children: [
        ...routes,
        { path: "/", element: <Navigate to="/current" /> },
        { path: "*", element: <Navigate to="/current" /> },
      ],
    },
  ]);

export const App: React.FC = () => {
  const { t } = useTranslation();

  const routes: Route[] = useMemo(
    () => [
      {
        name: t("tabs.current"),
        path: "current",
        element: <Typography>TODO CURRENT</Typography>,
      },
      {
        name: t("tabs.programs"),
        path: "programs",
        element: <Typography>TODO PROGRAMS</Typography>,
      },
      {
        name: t("tabs.phases"),
        path: "phases",
        element: <Typography>TODO PHASES</Typography>,
      },
      {
        name: t("tabs.cycles"),
        path: "cycles",
        element: <Typography>TODO CYCLES</Typography>,
      },
    ],
    [t]
  );

  return (
    <React.StrictMode>
      <ThemeProvider theme={theme}>
        <Paper sx={{ height: "100%", width: "100%", borderRadius: 0 }}>
          <RouterProvider router={getRouter(routes)} />
        </Paper>
      </ThemeProvider>
    </React.StrictMode>
  );
};
