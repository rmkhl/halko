import { configureStore } from "@reduxjs/toolkit";
import phasesReducer from "./features/phasesSlice";
import { configuratorApi } from "./services";
import { executorApi } from "./services/executorApi";

export const store = configureStore({
  reducer: {
    phases: phasesReducer,
    [configuratorApi.reducerPath]: configuratorApi.reducer,
    [executorApi.reducerPath]: executorApi.reducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware()
      .concat(configuratorApi.middleware)
      .concat(executorApi.middleware),
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
