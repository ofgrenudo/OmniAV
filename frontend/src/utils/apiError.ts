import { ApiError } from '../services/apiClient';

export const errorMessage = (err: unknown, fallback: string): string => (err instanceof ApiError ? err.message : fallback);
