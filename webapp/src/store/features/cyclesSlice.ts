import { PayloadAction, createSlice } from "@reduxjs/toolkit";
import { Cycle } from "../../types/api";
import {
  getJSONFromSessionStorage,
  removeFromSessionStorage,
  setJSONToSessionStorage,
} from "../../util";

const editKey = "editCycle";

const initialState = {
  cycles: [] as Cycle[],
  edit: getJSONFromSessionStorage<Cycle>(editKey),
};

export const cyclesSlice = createSlice({
  name: "cycles",
  initialState,
  reducers: {
    setCycles: (
      state: typeof initialState,
      action: PayloadAction<typeof initialState.cycles>
    ) => ({
      ...state,
      cycles: action.payload.sort((a, b) => a.name.localeCompare(b.name)),
    }),
    setEditCycle: (
      state: typeof initialState,
      action: PayloadAction<typeof initialState.edit>
    ) => {
      const { payload: cycle } = action;

      if (!cycle) {
        removeFromSessionStorage(editKey);
      } else {
        setJSONToSessionStorage(editKey, cycle);
      }

      return {
        ...state,
        edit: cycle,
      };
    },
  },
});

export const { setCycles, setEditCycle } = cyclesSlice.actions;

export default cyclesSlice.reducer;
