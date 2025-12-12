import React, { useEffect, useState } from "react";
import { Route } from "../types";
import { Outlet, useLocation } from "react-router-dom";
import { Stack, Typography } from "@mui/material";
import { RouteTabs } from "./RouteTabs";
import { useTranslation } from "react-i18next";

interface Props {
  routes: Route[];
}

export const Navigation: React.FC<Props> = (props) => {
  const { routes } = props;
  const location = useLocation();
  const [idx, setIdx] = useState(0);

  const { t } = useTranslation();

  useEffect(() => {
    const newIdx = routes.findIndex(
      (r) => r.path && location.pathname === `/${r.path}`
    );

    // eslint-disable-next-line react-hooks/set-state-in-effect
    setIdx(newIdx);
  }, [location, routes]);

  return (
    <Stack alignItems="center" sx={{ height: "100vh", overflow: "hidden" }}>
      <Stack alignItems="center" paddingTop={2}>
        <Typography variant="h2">{t("header.title")}</Typography>
      </Stack>

      <Stack paddingTop={2} paddingBottom={3}>
        <RouteTabs routes={routes} idx={idx} />
      </Stack>

      <Stack flex={1} width="100%" sx={{ minHeight: 0, overflow: "hidden" }}>
        <Outlet />
      </Stack>
    </Stack>
  );
};
