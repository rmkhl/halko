import { configureStore } from "@reduxjs/toolkit";
import phasesReducer from "./features/phasesSlice";
import { configuratorApi } from "./services";
import { executorApi } from "./services/executorApi";
import programsSlice from "./features/programsSlice";

export const store = configureStore({
  reducer: {
    phases: phasesReducer,
    programs: programsSlice,
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
