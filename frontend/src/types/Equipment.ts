export interface Equipment {
  id: number;
  name: string;
  description: string | null;
  disabled: boolean;
  archived: boolean;
  groupId: number;
  createdAt: string;
  updatedAt: string;
}
