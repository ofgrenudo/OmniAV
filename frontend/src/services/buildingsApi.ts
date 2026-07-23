import { Building } from '../types/Building';
import { PaginationMeta, apiRequest, buildQueryString } from './apiClient';

export { ApiError } from './apiClient';
export type { PaginationMeta } from './apiClient';

export const BUILDING_SORT_COLUMNS = ['id', 'name', 'archived', 'created_at', 'updated_at'] as const;
export type BuildingSortColumn = (typeof BUILDING_SORT_COLUMNS)[number];

export interface BuildingListResult {
  data: Building[];
  meta: PaginationMeta;
}

export type BuildingListParams = {
  page?: number;
  pageSize?: number;
  sort?: BuildingSortColumn;
  order?: 'asc' | 'desc';
  q?: string;
  archived?: boolean;
};

export interface BuildingInput {
  name: string;
  archived?: boolean;
  description?: string | null;
  address?: string | null;
}

export const listBuildings = (params: BuildingListParams = {}): Promise<BuildingListResult> =>
  apiRequest<BuildingListResult>(`/buildings?${buildQueryString(params)}`);

export const getBuilding = (id: number): Promise<Building> => apiRequest<Building>(`/buildings/${id}`);

export const createBuilding = (input: BuildingInput): Promise<Building> =>
  apiRequest<Building>('/buildings', { method: 'POST', body: JSON.stringify(input) });

export const updateBuilding = (id: number, input: BuildingInput): Promise<Building> =>
  apiRequest<Building>(`/buildings/${id}`, { method: 'PUT', body: JSON.stringify(input) });

export const archiveBuilding = (id: number): Promise<void> =>
  apiRequest<void>(`/buildings/${id}`, { method: 'DELETE' });
