import { Building } from './Building';
import { Equipment } from './Equipment';
import { EquipmentGroup } from './EquipmentGroup';

export type DayOfWeek = 'Monday' | 'Tuesday' | 'Wednesday' | 'Thursday' | 'Friday' | 'Saturday' | 'Sunday';

export interface RequestedEquipmentEntry {
  id: number;
  groupId: number;
  group?: EquipmentGroup;
  equipmentId: number;
  equipment?: Equipment;
  requestId: number;
}

export interface Request {
  id: number;
  name: string;
  firstDateNeeded: string;
  startTime: string;
  endTime: string;
  numberOfWeeks: number;
  daysOfWeek: DayOfWeek;
  buildingId: number;
  room: string;
  building?: Building;
  comments: string | null;
  attachmentId: number | null;
  createdAt: string;
  updatedAt: string;
  requestedEquipment?: RequestedEquipmentEntry[];
}
