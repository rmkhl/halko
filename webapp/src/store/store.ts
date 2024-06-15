import { configureStore } from "@reduxjs/toolkit";
import phasesReducer from "./features/phasesSlice";
import { configuratorApi } from "./services";
import { executorApi } from "./services/executorApi";
import programsSlice from "./features/programsSlice";
import { sensorApi } from "./services/sensorsApi";

export const store = configureStore({
  reducer: {
    phases: phasesReducer,
    programs: programsSlice,
    [configuratorApi.reducerPath]: configuratorApi.reducer,
    [executorApi.reducerPath]: executorApi.reducer,
    [sensorApi.reducerPath]: sensorApi.reducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware()
      .concat(configuratorApi.middleware)
      .concat(executorApi.middleware)
      .concat(sensorApi.middleware),
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
