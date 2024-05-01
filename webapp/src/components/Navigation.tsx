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

    setIdx(newIdx);
  }, [location]);

  return (
    <Stack alignItems="center">
      <Stack flex={1} alignItems="center">
        <Typography variant="h2">{t("header.title")}</Typography>
      </Stack>

      <Stack paddingBottom={3}>
        <RouteTabs routes={routes} idx={idx} />
      </Stack>

      <Outlet />
    </Stack>
  );
};
