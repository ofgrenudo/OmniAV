export interface Equipment {
  id: number;
  name: string;
  description: string | null;
  disabled: boolean;
  archived: boolean;
  groupId: number;
  buildingId: number;
  createdAt: string;
  updatedAt: string;
}
