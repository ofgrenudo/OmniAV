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
  buildingId?: number | null;
}

export const listEquipment = (params: EquipmentListParams = {}): Promise<EquipmentListResult> =>
  apiRequest<EquipmentListResult>(`/equipment?${buildQueryString(params)}`);

export const getEquipment = (id: number): Promise<Equipment> => apiRequest<Equipment>(`/equipment/${id}`);

export const createEquipment = (input: EquipmentInput): Promise<Equipment> =>
  apiRequest<Equipment>('/equipment', { method: 'POST', body: JSON.stringify(input) });

export const updateEquipment = (id: number, input: EquipmentInput): Promise<Equipment> =>
  apiRequest<Equipment>(`/equipment/${id}`, { method: 'PUT', body: JSON.stringify(input) });

export const archiveEquipment = (id: number): Promise<void> =>
  apiRequest<void>(`/equipment/${id}`, { method: 'DELETE' });
