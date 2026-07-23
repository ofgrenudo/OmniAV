import {
  ApiError,
  assignEquipment,
  createRequest,
  deleteRequest,
  getRequest,
  listRequestedEquipment,
  listRequests,
  unassignEquipment,
} from './requestsApi';
import { Request, RequestedEquipmentEntry } from '../types/Request';

const sampleRequest: Request = {
  id: 1,
  name: 'Bio Lecture',
  firstDateNeeded: '2026-08-02T00:00:00Z',
  startTime: '0000-01-01T09:00:00Z',
  endTime: '0000-01-01T10:00:00Z',
  numberOfWeeks: 1,
  daysOfWeek: 'Sunday',
  buildingId: 1,
  room: '101',
  comments: null,
  attachmentId: null,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
};

const sampleAssignment: RequestedEquipmentEntry = {
  id: 1,
  groupId: 1,
  equipmentId: 1,
  requestId: 1,
};

const mockFetchResponse = (body: unknown, status = 200) => {
  (global.fetch as jest.Mock).mockResolvedValueOnce({
    ok: status >= 200 && status < 300,
    status,
    json: () => Promise.resolve(body),
  });
};

describe('requestsApi', () => {
  beforeEach(() => {
    global.fetch = jest.fn() as unknown as typeof fetch;
  });

  it('builds a query string for listRequests', async () => {
    const result = { data: [sampleRequest], meta: { page: 1, pageSize: 20, totalItems: 1, totalPages: 1 } };
    mockFetchResponse(result);

    const got = await listRequests({ buildingId: 1, q: 'bio' });

    expect(got).toEqual(result);
    const [url] = (global.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe('/api/requests?buildingId=1&q=bio');
  });

  it('fetches a single request by id', async () => {
    mockFetchResponse(sampleRequest);

    const got = await getRequest(1);

    expect(got).toEqual(sampleRequest);
    const [url] = (global.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe('/api/requests/1');
  });

  it('posts a new request on createRequest', async () => {
    mockFetchResponse(sampleRequest, 201);

    const got = await createRequest({
      name: 'Bio Lecture',
      firstDateNeeded: '2026-08-02',
      startTime: '09:00',
      endTime: '10:00',
      numberOfWeeks: 1,
      buildingId: 1,
      room: '101',
    });

    expect(got).toEqual(sampleRequest);
    const [url, init] = (global.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe('/api/requests');
    expect(init.method).toBe('POST');
  });

  it('resolves with no body for deleteRequest on a 204', async () => {
    mockFetchResponse(null, 204);

    await expect(deleteRequest(1)).resolves.toBeUndefined();

    const [url, init] = (global.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe('/api/requests/1');
    expect(init.method).toBe('DELETE');
  });

  it('lists requested equipment for a request', async () => {
    mockFetchResponse([sampleAssignment]);

    const got = await listRequestedEquipment(1);

    expect(got).toEqual([sampleAssignment]);
    const [url] = (global.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe('/api/requests/1/equipment');
  });

  it('posts a groupId to assign equipment', async () => {
    mockFetchResponse(sampleAssignment, 201);

    const got = await assignEquipment(1, 5);

    expect(got).toEqual(sampleAssignment);
    const [url, init] = (global.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe('/api/requests/1/equipment');
    expect(init.method).toBe('POST');
    expect(JSON.parse(init.body)).toEqual({ groupId: 5 });
  });

  it('deletes a specific assignment on unassignEquipment', async () => {
    mockFetchResponse(null, 204);

    await expect(unassignEquipment(1, 7)).resolves.toBeUndefined();

    const [url, init] = (global.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe('/api/requests/1/equipment/7');
    expect(init.method).toBe('DELETE');
  });

  it('throws an ApiError with a 409 when no equipment is available', async () => {
    expect.assertions(3);
    mockFetchResponse({ error: 'no equipment in this group is available for the requested time' }, 409);

    try {
      await assignEquipment(1, 5);
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError);
      expect((err as ApiError).status).toBe(409);
      expect((err as ApiError).message).toBe('no equipment in this group is available for the requested time');
    }
  });
});
