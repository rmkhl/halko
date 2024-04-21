import { PayloadAction, createSlice } from "@reduxjs/toolkit";
import { Cycle } from "../../types/api";

const initialState = {
  value: [] as Cycle[],
};

export const cyclesSlice = createSlice({
  name: "cycles",
  initialState,
  reducers: {
    setCycles: (state: typeof initialState, action: PayloadAction<Cycle[]>) => {
      state.value = action.payload.sort((a, b) => a.name.localeCompare(b.name));
    },
  },
});

export const { setCycles } = cyclesSlice.actions;

export default cyclesSlice.reducer;
