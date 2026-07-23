import { EquipmentGroup } from '../types/EquipmentGroup';
import { PaginationMeta, apiRequest, buildQueryString } from './apiClient';

export { ApiError } from './apiClient';
export type { PaginationMeta } from './apiClient';

export const EQUIPMENT_GROUP_SORT_COLUMNS = ['id', 'name', 'disabled', 'archived', 'created_at', 'updated_at'] as const;
export type EquipmentGroupSortColumn = (typeof EQUIPMENT_GROUP_SORT_COLUMNS)[number];

export interface EquipmentGroupListResult {
  data: EquipmentGroup[];
  meta: PaginationMeta;
}

export type EquipmentGroupListParams = {
  page?: number;
  pageSize?: number;
  sort?: EquipmentGroupSortColumn;
  order?: 'asc' | 'desc';
  q?: string;
  archived?: boolean;
  disabled?: boolean;
};

export interface EquipmentGroupInput {
  name: string;
  description?: string | null;
  disabled?: boolean;
  archived?: boolean;
}

export type AvailabilityParams = {
  firstDate: string;
  startTime: string;
  endTime: string;
  weeks?: number;
};

export const listEquipmentGroups = (params: EquipmentGroupListParams = {}): Promise<EquipmentGroupListResult> =>
  apiRequest<EquipmentGroupListResult>(`/equipment-groups?${buildQueryString(params)}`);

export const getEquipmentGroup = (id: number): Promise<EquipmentGroup> =>
  apiRequest<EquipmentGroup>(`/equipment-groups/${id}`);

export const createEquipmentGroup = (input: EquipmentGroupInput): Promise<EquipmentGroup> =>
  apiRequest<EquipmentGroup>('/equipment-groups', { method: 'POST', body: JSON.stringify(input) });

export const updateEquipmentGroup = (id: number, input: EquipmentGroupInput): Promise<EquipmentGroup> =>
  apiRequest<EquipmentGroup>(`/equipment-groups/${id}`, { method: 'PUT', body: JSON.stringify(input) });

export const archiveEquipmentGroup = (id: number): Promise<void> =>
  apiRequest<void>(`/equipment-groups/${id}`, { method: 'DELETE' });

export const getEquipmentGroupAvailability = (id: number, params: AvailabilityParams): Promise<number> =>
  apiRequest<{ available: number }>(`/equipment-groups/${id}/availability?${buildQueryString(params)}`).then(
    (res) => res.available
  );
