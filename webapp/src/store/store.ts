import { configureStore } from "@reduxjs/toolkit";
import { setupListeners } from "@reduxjs/toolkit/query";
import { configuratorApi } from "./services";
import { controlunitApi } from "./services/controlunitApi";
import programsSlice from "./features/programsSlice";
import { sensorApi } from "./services/sensorsApi";
import { systemApi } from "./services/systemApi";
import { powerunitApi } from "./services/powerunitApi";
import { dbusunitApi } from "./services/dbusunitApi";

export const store = configureStore({
  reducer: {
    programs: programsSlice,
    [configuratorApi.reducerPath]: configuratorApi.reducer,
    [controlunitApi.reducerPath]: controlunitApi.reducer,
    [sensorApi.reducerPath]: sensorApi.reducer,
    [systemApi.reducerPath]: systemApi.reducer,
    [powerunitApi.reducerPath]: powerunitApi.reducer,
    [dbusunitApi.reducerPath]: dbusunitApi.reducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware()
      .concat(configuratorApi.middleware)
      .concat(controlunitApi.middleware)
      .concat(sensorApi.middleware)
      .concat(systemApi.middleware)
      .concat(powerunitApi.middleware)
      .concat(dbusunitApi.middleware),
});

// Every polled query asks to pause while the tab is hidden. RTK Query decides
// that from its own "focused" flag, which is initialized from
// document.visibilityState at store creation and only ever changes when these
// listeners dispatch. Without them a page that loads hidden - a reload in a
// background tab - stays unfocused for good and never polls again: the initial
// fetch on mount is the last data it ever shows, while the clocks keep ticking
// and the chart's websocket keeps it looking live.
setupListeners(store.dispatch);

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
