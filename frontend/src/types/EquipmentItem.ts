// EquipmentItem is the NewRequest flow's view of an EquipmentGroup: its id is the group's id,
// and `available` is the live count of free units for the schedule the user picked.
export interface EquipmentItem {
  id: number;
  name: string;
  available: number;
}
