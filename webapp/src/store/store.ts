import { configureStore } from "@reduxjs/toolkit";
import phasesReducer from "./features/phasesSlice";
import { configuratorApi } from "./services";

export const store = configureStore({
  reducer: {
    phases: phasesReducer,
    [configuratorApi.reducerPath]: configuratorApi.reducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(configuratorApi.middleware),
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
