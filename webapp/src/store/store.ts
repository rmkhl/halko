import { configureStore } from "@reduxjs/toolkit";
import cyclesReducer from "./features/cyclesSlice";
import { configuratorApi } from "./services";

export const store = configureStore({
  reducer: {
    cycles: cyclesReducer,
    [configuratorApi.reducerPath]: configuratorApi.reducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware().concat(configuratorApi.middleware),
});

export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;
