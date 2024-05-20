import { Program } from "../../types/api";
import { getJSONFromSessionStorage } from "../../util";
import { createEntitySlice } from "./entitySlice";

const editKey = "editProgram";

export const programsSlice = createEntitySlice({
  sliceName: "programs",
  editRecordSessionStorageKey: editKey,
  initialRecords: [] as Program[],
  initialEditRecord: getJSONFromSessionStorage<Program>(editKey),
  reducers: {},
});

export const { setRecords: setPrograms, setEditRecord: setEditProgram } =
  programsSlice.actions;

export default programsSlice.reducer;
