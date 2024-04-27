import { PayloadAction, createSlice } from "@reduxjs/toolkit";
import { Phase } from "../../types/api";
import {
  getJSONFromSessionStorage,
  removeFromSessionStorage,
  setJSONToSessionStorage,
} from "../../util";

const editKey = "editPhase";

const initialState = {
  cycles: [] as Phase[],
  edit: getJSONFromSessionStorage<Phase>(editKey),
};

export const phasesSlice = createSlice({
  name: "phases",
  initialState,
  reducers: {
    setPhases: (
      state: typeof initialState,
      action: PayloadAction<typeof initialState.cycles>
    ) => ({
      ...state,
      cycles: action.payload.sort((a, b) => a.name.localeCompare(b.name)),
    }),
    setEditPhase: (
      state: typeof initialState,
      action: PayloadAction<typeof initialState.edit>
    ) => {
      const { payload: phase } = action;

      if (!phase) {
        removeFromSessionStorage(editKey);
      } else {
        setJSONToSessionStorage(editKey, phase);
      }

      return {
        ...state,
        edit: phase,
      };
    },
  },
});

export const { setPhases, setEditPhase } = phasesSlice.actions;

export default phasesSlice.reducer;
