import { CssBaseline, Paper, ThemeProvider } from "@mui/material";
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
import { Provider } from "react-redux";
import { store } from "./store/store";
import { Programs } from "./components/programs/Programs";
import { Program } from "./components/programs/Program";
import { Running } from "./components/Running";
import { History } from "./components/History";
import { ErrorBoundary } from "./components/ErrorBoundary";

const getRouter = (routes: Route[]) =>
  createBrowserRouter([
    {
      path: "/",
      element: <Navigation routes={routes} />,
      errorElement: <ErrorBoundary />,
      children: [
        ...routes,
        { path: "programs/:name", element: <Program />, errorElement: <ErrorBoundary /> },
        { path: "/", element: <Navigate to="/running" /> },
        { path: "*", element: <Navigate to="/running" /> },
      ],
    },
  ]);

export const App: React.FC = () => {
  const { t } = useTranslation();

  const routes: Route[] = useMemo(
    () => [
      {
        name: t("tabs.running"),
        path: "running",
        element: <Running />,
      },
      {
        name: t("tabs.programs"),
        path: "programs",
        element: <Programs />,
      },
      {
        name: t("tabs.history"),
        path: "history",
        element: <History />,
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
