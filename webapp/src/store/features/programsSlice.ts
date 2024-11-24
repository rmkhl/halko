import { UIProgram } from "../../types/api";
import { getJSONFromSessionStorage } from "../../util";
import { createEntitySlice } from "./entitySlice";

const editKey = "editProgram";

export const programsSlice = createEntitySlice({
  sliceName: "programs",
  editRecordSessionStorageKey: editKey,
  initialRecords: [] as UIProgram[],
  initialEditRecord: getJSONFromSessionStorage<UIProgram>(editKey),
  reducers: {},
});

export const { setRecords: setPrograms, setEditRecord: setEditProgram } =
  programsSlice.actions;

export default programsSlice.reducer;
