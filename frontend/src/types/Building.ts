export interface Building {
  id: number;
  name: string;
  archived: boolean;
  description: string | null;
  address: string | null;
  createdAt: string;
  updatedAt: string;
}
