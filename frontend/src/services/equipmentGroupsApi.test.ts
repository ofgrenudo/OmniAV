import {
  ApiError,
  archiveEquipmentGroup,
  createEquipmentGroup,
  getEquipmentGroup,
  getEquipmentGroupAvailability,
  getEquipmentGroupSchedule,
  listEquipmentGroups,
  updateEquipmentGroup,
} from './equipmentGroupsApi';
import { EquipmentGroup } from '../types/EquipmentGroup';

const sampleGroup: EquipmentGroup = {
  id: 1,
  name: 'Cow Cart',
  description: null,
  disabled: false,
  archived: false,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
};

const mockFetchResponse = (body: unknown, status = 200) => {
  (global.fetch as jest.Mock).mockResolvedValueOnce({
    ok: status >= 200 && status < 300,
    status,
    json: () => Promise.resolve(body),
  });
};

describe('equipmentGroupsApi', () => {
  beforeEach(() => {
    global.fetch = jest.fn() as unknown as typeof fetch;
  });

  it('builds a query string for listEquipmentGroups', async () => {
    const result = { data: [sampleGroup], meta: { page: 1, pageSize: 20, totalItems: 1, totalPages: 1 } };
    mockFetchResponse(result);

    const got = await listEquipmentGroups({ q: 'cow', archived: false, disabled: false });

    expect(got).toEqual(result);
    const [url] = (global.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe('/api/equipment-groups?q=cow&archived=false&disabled=false');
  });

  it('fetches a single group by id', async () => {
    mockFetchResponse(sampleGroup);

    const got = await getEquipmentGroup(1);

    expect(got).toEqual(sampleGroup);
    const [url] = (global.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe('/api/equipment-groups/1');
  });

  it('posts a new group on createEquipmentGroup', async () => {
    mockFetchResponse(sampleGroup, 201);

    const got = await createEquipmentGroup({ name: 'Cow Cart' });

    expect(got).toEqual(sampleGroup);
    const [url, init] = (global.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe('/api/equipment-groups');
    expect(init.method).toBe('POST');
  });

  it('puts an update on updateEquipmentGroup', async () => {
    mockFetchResponse({ ...sampleGroup, disabled: true });

    const got = await updateEquipmentGroup(1, { name: 'Cow Cart', disabled: true });

    expect(got.disabled).toBe(true);
    const [url, init] = (global.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe('/api/equipment-groups/1');
    expect(init.method).toBe('PUT');
  });

  it('resolves with no body for archiveEquipmentGroup on a 204', async () => {
    mockFetchResponse(null, 204);

    await expect(archiveEquipmentGroup(1)).resolves.toBeUndefined();

    const [url, init] = (global.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe('/api/equipment-groups/1');
    expect(init.method).toBe('DELETE');
  });

  it('fetches availability and unwraps the count', async () => {
    mockFetchResponse({ available: 3 });

    const got = await getEquipmentGroupAvailability(1, {
      firstDate: '2026-08-02',
      startTime: '09:00',
      endTime: '10:00',
      weeks: 4,
      buildingId: 7,
    });

    expect(got).toBe(3);
    const [url] = (global.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe(
      '/api/equipment-groups/1/availability?firstDate=2026-08-02&startTime=09%3A00&endTime=10%3A00&weeks=4&buildingId=7'
    );
  });

  it('throws an ApiError carrying the server message and status on failure', async () => {
    expect.assertions(3);
    mockFetchResponse({ error: 'equipment group not found' }, 404);

    try {
      await getEquipmentGroup(999);
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError);
      expect((err as ApiError).status).toBe(404);
      expect((err as ApiError).message).toBe('equipment group not found');
    }
  });
});

describe('getEquipmentGroupSchedule', () => {
  it('requests the given day and unwraps the entries', async () => {
    const entry = {
      equipmentId: 10,
      equipmentName: 'Cow Cart A',
      requestId: 42,
      requestName: 'Bio Lecture',
      buildingId: 7,
      buildingName: 'Anna Whitten Hall',
      room: '204',
      startTime: '09:00',
      endTime: '10:00',
      comments: null,
    };
    mockFetchResponse({ date: '2026-08-02', data: [entry] });

    const got = await getEquipmentGroupSchedule(3, '2026-08-02');

    expect(got).toEqual([entry]);
    const [url] = (global.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe('/api/equipment-groups/3/schedule?date=2026-08-02');
  });
});
