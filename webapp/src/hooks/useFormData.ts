import { useEffect, useMemo, useState } from "react";
import { FormMode } from "../types";
import { useNavigate, useParams } from "react-router-dom";
import { useDispatch } from "react-redux";
import { UnknownAction } from "@reduxjs/toolkit";

interface Named {
  name: string;
}

interface Props<T extends Named> {
  allData: T[];
  defaultData: T;
  editData?: T;
  rootPath: string;
  saveSuccess: boolean;

  normalizeData?: (data: T) => T;
  saveData: (data: T) => void;
  setEditData: (editData?: T) => UnknownAction;
}

export const useFormData = <T extends Named>(props: Props<T>) => {
  const {
    allData,
    defaultData,
    editData,
    rootPath,
    saveSuccess,
    normalizeData,
    saveData,
    setEditData,
  } = props;

  const [mode, setMode] = useState<FormMode>("view");
  const [formData, setFormData] = useState<T>(defaultData);

  const { name } = useParams();
  const navigate = useNavigate();

  const dispatch = useDispatch();


  useEffect(() => {
    if (name === "new") {
      // eslint-disable-next-line react-hooks/set-state-in-effect
      setMode("edit");

      if (editData && !editData.name) {
        return;
      }

      dispatch(setEditData(defaultData));

      return;
    }

    if (!name || !allData || allData.length === 0) {
      return;
    }

    // Don't reset if already editing
    if (editData) {
      return;
    }

    const data = allData.find((p) => p.name === name);

    if (!data) {
      navigate(rootPath);
      return;
    }

    // When loading an existing program, start in edit mode
    // eslint-disable-next-line react-hooks/set-state-in-effect
    setMode("edit");
    setFormData(data);
    dispatch(setEditData(data));
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [name, allData]);


  useEffect(() => {
    if (!saveSuccess) {
      return;
    }

    const editName = editData?.name;
    dispatch(setEditData(undefined));

    // eslint-disable-next-line react-hooks/set-state-in-effect
    setMode("view");

    if (name === "new") {
      navigate(`${rootPath}/${editName}`);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [saveSuccess]);

  const handleEdit = () => {
    dispatch(setEditData(formData));
    setMode("edit");
  };

  const handleSave = () => {
    if (!editData) {
      return;
    }

    const normalized = normalizeData?.(editData) || editData;
    saveData({ id: name !== "new" ? name : undefined, ...normalized });

    const { name: editName } = editData;

    dispatch(setEditData(undefined));

    if (name === "new") {
      navigate(`${rootPath}${editName.length ? `/${editName}` : ""}`);
    }

    setMode("view");
  };

  const handleCancel = () => {
    dispatch(setEditData(undefined));
    navigate(rootPath);
  };


  useEffect(() => {
    if (!name) {
      navigate(rootPath);
    }
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [name]);


  const nameUsed = useMemo(() => {
    if (!editData || !allData?.length) {
      return false;
    }

    for (const p of allData) {
      if (p.name === formData.name) {
        continue;
      }

      if (p.name.trim() === editData.name.trim()) {
        return true;
      }
    }

    return false;
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [allData, editData]);

  return {
    editing: useMemo(() => mode === "edit", [mode]),
    formData,
    nameUsed,
    handleCancel,
    handleEdit,
    handleSave,
  };
};
