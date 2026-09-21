import { Equipment } from '../types/Equipment';
import { PaginationMeta, apiRequest, buildQueryString } from './apiClient';

export { ApiError } from './apiClient';
export type { PaginationMeta } from './apiClient';

export const EQUIPMENT_SORT_COLUMNS = ['id', 'name', 'disabled', 'archived', 'created_at', 'updated_at'] as const;
export type EquipmentSortColumn = (typeof EQUIPMENT_SORT_COLUMNS)[number];

export interface EquipmentListResult {
  data: Equipment[];
  meta: PaginationMeta;
}

export type EquipmentListParams = {
  page?: number;
  pageSize?: number;
  sort?: EquipmentSortColumn;
  order?: 'asc' | 'desc';
  q?: string;
  groupId?: number;
  buildingId?: number;
  archived?: boolean;
  disabled?: boolean;
};

export interface EquipmentInput {
  name: string;
  description?: string | null;
  disabled?: boolean;
  archived?: boolean;
  groupId: number;
  /**
   * Every unit has a permanent home building. There is no "unassigned / central storage" state,
   * so buildingId is required on create and update — a move is an admin edit that swaps one
   * building for another.
   */
  buildingId: number;
}

export const listEquipment = (params: EquipmentListParams = {}): Promise<EquipmentListResult> =>
  apiRequest<EquipmentListResult>(`/equipment?${buildQueryString(params)}`);

export const getEquipment = (id: number): Promise<Equipment> => apiRequest<Equipment>(`/equipment/${id}`);

/** One request a unit is assigned to — the unit's movement history, oldest first. */
export interface EquipmentBooking {
  requestedEquipmentId: number;
  requestId: number;
  requestName: string;
  groupId: number;
  buildingId: number;
  buildingName: string;
  room: string;
  firstDateNeeded: string;
  daysOfWeek: string;
  numberOfWeeks: number;
  startTime: string;
  endTime: string;
}

export interface EquipmentBookings {
  equipmentId: number;
  equipmentName: string;
  data: EquipmentBooking[];
}

export const listEquipmentBookings = (id: number): Promise<EquipmentBookings> =>
  apiRequest<EquipmentBookings>(`/equipment/${id}/bookings`);

export const createEquipment = (input: EquipmentInput): Promise<Equipment> =>
  apiRequest<Equipment>('/equipment', { method: 'POST', body: JSON.stringify(input) });

export const updateEquipment = (id: number, input: EquipmentInput): Promise<Equipment> =>
  apiRequest<Equipment>(`/equipment/${id}`, { method: 'PUT', body: JSON.stringify(input) });

export const archiveEquipment = (id: number): Promise<void> =>
  apiRequest<void>(`/equipment/${id}`, { method: 'DELETE' });
