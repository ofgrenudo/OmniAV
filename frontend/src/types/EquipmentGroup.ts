export interface EquipmentGroup {
  id: number;
  name: string;
  description: string | null;
  disabled: boolean;
  archived: boolean;
  createdAt: string;
  updatedAt: string;
}
