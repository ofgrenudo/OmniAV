import { Request, RequestedEquipmentEntry } from '../types/Request';
import { PaginationMeta, apiRequest, buildQueryString } from './apiClient';

export { ApiError } from './apiClient';
export type { PaginationMeta } from './apiClient';

export const REQUEST_SORT_COLUMNS = ['id', 'name', 'first_date_needed', 'created_at', 'updated_at'] as const;
export type RequestSortColumn = (typeof REQUEST_SORT_COLUMNS)[number];

export interface RequestListResult {
  data: Request[];
  meta: PaginationMeta;
}

export type RequestListParams = {
  page?: number;
  pageSize?: number;
  sort?: RequestSortColumn;
  order?: 'asc' | 'desc';
  q?: string;
  buildingId?: number;
};

export interface RequestInput {
  name: string;
  firstDateNeeded: string;
  startTime: string;
  endTime: string;
  numberOfWeeks: number;
  buildingId: number;
  room: string;
  comments?: string | null;
}

export const listRequests = (params: RequestListParams = {}): Promise<RequestListResult> =>
  apiRequest<RequestListResult>(`/requests?${buildQueryString(params)}`);

export const getRequest = (id: number): Promise<Request> => apiRequest<Request>(`/requests/${id}`);

export const createRequest = (input: RequestInput): Promise<Request> =>
  apiRequest<Request>('/requests', { method: 'POST', body: JSON.stringify(input) });

export const deleteRequest = (id: number): Promise<void> =>
  apiRequest<void>(`/requests/${id}`, { method: 'DELETE' });

export const listRequestedEquipment = (requestId: number): Promise<RequestedEquipmentEntry[]> =>
  apiRequest<RequestedEquipmentEntry[]>(`/requests/${requestId}/equipment`);

export const assignEquipment = (requestId: number, groupId: number): Promise<RequestedEquipmentEntry> =>
  apiRequest<RequestedEquipmentEntry>(`/requests/${requestId}/equipment`, {
    method: 'POST',
    body: JSON.stringify({ groupId }),
  });

export const unassignEquipment = (requestId: number, requestedEquipmentId: number): Promise<void> =>
  apiRequest<void>(`/requests/${requestId}/equipment/${requestedEquipmentId}`, { method: 'DELETE' });
