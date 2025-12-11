import { configureStore } from "@reduxjs/toolkit";
import phasesReducer from "./features/phasesSlice";
import { configuratorApi } from "./services";
import { controlunitApi } from "./services/controlunitApi";
import programsSlice from "./features/programsSlice";
import { sensorApi } from "./services/sensorsApi";

export const store = configureStore({
  reducer: {
    phases: phasesReducer,
    programs: programsSlice,
    [configuratorApi.reducerPath]: configuratorApi.reducer,
    [controlunitApi.reducerPath]: controlunitApi.reducer,
    [sensorApi.reducerPath]: sensorApi.reducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware()
      .concat(configuratorApi.middleware)
      .concat(controlunitApi.middleware)
      .concat(sensorApi.middleware),
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
