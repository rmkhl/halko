import React from "react";
import { Route } from "../types";
import { Tab, Tabs } from "@mui/material";
import { useNavigate } from "react-router-dom";

type Props = {
  routes: Route[];
  idx: number;
};

export const RouteTabs: React.FC<Props> = (props: Props) => {
  const { idx, routes } = props;

  const navigate = useNavigate();

  const handleChange = (_: React.SyntheticEvent, newValue: number) => {
    const path = routes[newValue]?.path;

    if (!path) {
      return;
    }

    navigate(path);
  };

  return (
    <Tabs value={idx > -1 ? idx : false} onChange={handleChange}>
      {routes.map((route) => (
        <Tab key={`route-${route.name}`} label={route.name} />
      ))}
    </Tabs>
  );
};
