import { Phase } from "../../types/api";
import { getJSONFromSessionStorage } from "../../util";
import { createEntitySlice } from "./entitySlice";

const editKey = "editPhase";

export const phasesSlice = createEntitySlice({
  sliceName: "phases",
  editRecordSessionStorageKey: editKey,
  initialRecords: [] as Phase[],
  initialEditRecord: getJSONFromSessionStorage<Phase>(editKey),
  reducers: {},
});

export const { setRecords: setPhases, setEditRecord: setEditPhase } =
  phasesSlice.actions;

export default phasesSlice.reducer;
