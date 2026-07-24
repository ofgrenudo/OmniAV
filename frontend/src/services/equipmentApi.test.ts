import {
  ApiError,
  archiveEquipment,
  createEquipment,
  getEquipment,
  listEquipment,
  updateEquipment,
} from './equipmentApi';
import { Equipment } from '../types/Equipment';

const sampleEquipment: Equipment = {
  id: 1,
  name: 'Cow Cart A',
  description: null,
  disabled: false,
  archived: false,
  groupId: 5,
  buildingId: null,
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

describe('equipmentApi', () => {
  beforeEach(() => {
    global.fetch = jest.fn() as unknown as typeof fetch;
  });

  it('builds a query string for listEquipment, including groupId', async () => {
    const result = { data: [sampleEquipment], meta: { page: 1, pageSize: 20, totalItems: 1, totalPages: 1 } };
    mockFetchResponse(result);

    const got = await listEquipment({ groupId: 5, disabled: false });

    expect(got).toEqual(result);
    const [url] = (global.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe('/api/equipment?groupId=5&disabled=false');
  });

  it('fetches a single unit by id', async () => {
    mockFetchResponse(sampleEquipment);

    const got = await getEquipment(1);

    expect(got).toEqual(sampleEquipment);
    const [url] = (global.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe('/api/equipment/1');
  });

  it('posts a new unit on createEquipment', async () => {
    mockFetchResponse(sampleEquipment, 201);

    const got = await createEquipment({ name: 'Cow Cart A', groupId: 5 });

    expect(got).toEqual(sampleEquipment);
    const [url, init] = (global.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe('/api/equipment');
    expect(init.method).toBe('POST');
    expect(JSON.parse(init.body)).toEqual({ name: 'Cow Cart A', groupId: 5 });
  });

  it('puts an update on updateEquipment', async () => {
    mockFetchResponse({ ...sampleEquipment, groupId: 9 });

    const got = await updateEquipment(1, { name: 'Cow Cart A', groupId: 9 });

    expect(got.groupId).toBe(9);
    const [url, init] = (global.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe('/api/equipment/1');
    expect(init.method).toBe('PUT');
  });

  it('resolves with no body for archiveEquipment on a 204', async () => {
    mockFetchResponse(null, 204);

    await expect(archiveEquipment(1)).resolves.toBeUndefined();

    const [url, init] = (global.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe('/api/equipment/1');
    expect(init.method).toBe('DELETE');
  });

  it('throws an ApiError carrying the server message and status on failure', async () => {
    expect.assertions(3);
    mockFetchResponse({ error: 'equipment group not found' }, 400);

    try {
      await createEquipment({ name: 'Orphan', groupId: 999999 });
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError);
      expect((err as ApiError).status).toBe(400);
      expect((err as ApiError).message).toBe('equipment group not found');
    }
  });
});
