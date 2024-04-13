import { Container, Tab, Tabs, Typography } from "@mui/material";
import React from "react";
import { useTranslation } from "react-i18next";
import { BrowserRouter } from "react-router-dom";

export const App: React.FC = () => {
  const { t } = useTranslation();

  return (
    <Container>
      <Typography variant="h2">{t("header.title")}</Typography>

      <Tabs>
        <Tab label={t("tabs.current")} />

        <Tab label={t("tabs.programs")} />

        <Tab label={t("tabs.phases")} />

        <Tab label={t("tabs.cycles")} />
      </Tabs>
    </Container>
  );
};
