import { AVRequest } from '../types/AVRequest';

const STORAGE_KEY = 'omniav.requests';

const read = (): AVRequest[] => {
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    return raw ? JSON.parse(raw) : [];
  } catch {
    return [];
  }
};

const write = (requests: AVRequest[]): void => {
  window.localStorage.setItem(STORAGE_KEY, JSON.stringify(requests));
};

export const submitRequest = (request: Omit<AVRequest, 'id' | 'submittedAt'>): AVRequest => {
  const saved: AVRequest = {
    ...request,
    id: `req-${Math.random().toString(36).slice(2, 10)}`,
    submittedAt: new Date().toISOString(),
  };
  write([...read(), saved]);
  return saved;
};

export const listRequests = (): AVRequest[] => read();
