import {
  PayloadAction,
  createSlice,
  SliceCaseReducers,
  ValidateSliceCaseReducers,
} from "@reduxjs/toolkit";
import { removeFromSessionStorage, setJSONToSessionStorage } from "../../util";

export interface EditableRecord<T> {
  records: T[];
  editRecord: T;
}

interface StoreSliceProps<T, R extends SliceCaseReducers<EditableRecord<T>>> {
  sliceName: string;
  editRecordSessionStorageKey: string;
  initialRecords: T[];
  initialEditRecord: T;
  reducers: ValidateSliceCaseReducers<EditableRecord<T>, R>;
}

export function createEntitySlice<
  T,
  R extends SliceCaseReducers<EditableRecord<T>>
>(props: StoreSliceProps<T, R>) {
  const {
    sliceName,
    editRecordSessionStorageKey,
    initialEditRecord,
    initialRecords,
    reducers,
  } = props;

  const initialState: EditableRecord<T> = {
    records: initialRecords,
    editRecord: initialEditRecord,
  };

  return createSlice({
    name: sliceName,
    initialState,
    reducers: {
      ...reducers,
      setRecords: (
        state: typeof initialState,
        action: PayloadAction<typeof initialState.records>
      ) => {
        const { payload } = action;

        return {
          ...state,
          records: [...payload],
        };
      },
      setEditRecord: (
        state: typeof initialState,
        action: PayloadAction<typeof initialState.editRecord>
      ) => {
        const { payload } = action;

        if (!payload) {
          removeFromSessionStorage(editRecordSessionStorageKey);
        } else {
          setJSONToSessionStorage(editRecordSessionStorageKey, payload);
        }

        return {
          ...state,
          editRecord: payload,
        };
      },
    },
  });
}
