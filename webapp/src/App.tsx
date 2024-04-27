import { CssBaseline, Paper, ThemeProvider, Typography } from "@mui/material";
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
import { Cycles } from "./components/cycles";
import { Provider } from "react-redux";
import { store } from "./store/store";
import { Phases } from "./components/phases/Phases";
import { Phase } from "./components/phases/Phase";

const getRouter = (routes: Route[]) =>
  createBrowserRouter([
    {
      path: "/",
      element: <Navigation routes={routes} />,
      children: [
        ...routes,
        { path: "phases/:id", element: <Phase /> },
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
        name: t("tabs.programs"),
        path: "programs",
        element: <Typography>TODO PROGRAMS</Typography>,
      },
      {
        name: t("tabs.phases"),
        path: "phases",
        element: <Phases />,
      },
      {
        name: t("tabs.cycles"),
        path: "cycles",
        element: <Cycles />,
      },
    ],
    [t]
  );

  return (
    <React.StrictMode>
      <ThemeProvider theme={theme}>
        <CssBaseline />

        <Paper sx={{ height: "100%", width: "100%", borderRadius: 0 }}>
          <Provider store={store}>
            <RouterProvider router={getRouter(routes)} />
          </Provider>
        </Paper>
      </ThemeProvider>
    </React.StrictMode>
  );
};