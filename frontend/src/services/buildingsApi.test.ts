import { ApiError, archiveBuilding, createBuilding, getBuilding, listBuildings, updateBuilding } from './buildingsApi';
import { Building } from '../types/Building';

const sampleBuilding: Building = {
  id: 1,
  name: 'Anna Whitten Hall',
  archived: false,
  description: null,
  address: null,
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

describe('buildingsApi', () => {
  beforeEach(() => {
    global.fetch = jest.fn() as unknown as typeof fetch;
  });

  it('builds a query string for listBuildings and returns the parsed result', async () => {
    const result = { data: [sampleBuilding], meta: { page: 1, pageSize: 20, totalItems: 1, totalPages: 1 } };
    mockFetchResponse(result);

    const got = await listBuildings({
      page: 2,
      pageSize: 10,
      sort: 'name',
      order: 'desc',
      q: 'whitten',
      archived: false,
    });

    expect(got).toEqual(result);
    const [url] = (global.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe('/api/buildings?page=2&pageSize=10&sort=name&order=desc&q=whitten&archived=false');
  });

  it('omits unset params from the query string', async () => {
    mockFetchResponse({ data: [], meta: { page: 1, pageSize: 20, totalItems: 0, totalPages: 0 } });

    await listBuildings();

    const [url] = (global.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe('/api/buildings?');
  });

  it('fetches a single building by id', async () => {
    mockFetchResponse(sampleBuilding);

    const got = await getBuilding(1);

    expect(got).toEqual(sampleBuilding);
    const [url] = (global.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe('/api/buildings/1');
  });

  it('posts a new building on createBuilding', async () => {
    mockFetchResponse(sampleBuilding, 201);

    const got = await createBuilding({ name: 'Anna Whitten Hall' });

    expect(got).toEqual(sampleBuilding);
    const [url, init] = (global.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe('/api/buildings');
    expect(init.method).toBe('POST');
    expect(JSON.parse(init.body)).toEqual({ name: 'Anna Whitten Hall' });
  });

  it('puts an update on updateBuilding', async () => {
    mockFetchResponse({ ...sampleBuilding, name: 'Renamed' });

    const got = await updateBuilding(1, { name: 'Renamed' });

    expect(got.name).toBe('Renamed');
    const [url, init] = (global.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe('/api/buildings/1');
    expect(init.method).toBe('PUT');
  });

  it('resolves with no body for archiveBuilding on a 204', async () => {
    mockFetchResponse(null, 204);

    await expect(archiveBuilding(1)).resolves.toBeUndefined();

    const [url, init] = (global.fetch as jest.Mock).mock.calls[0];
    expect(url).toBe('/api/buildings/1');
    expect(init.method).toBe('DELETE');
  });

  it('throws an ApiError carrying the server message and status on failure', async () => {
    expect.assertions(3);
    mockFetchResponse({ error: 'building not found' }, 404);

    try {
      await getBuilding(999);
    } catch (err) {
      expect(err).toBeInstanceOf(ApiError);
      expect((err as ApiError).status).toBe(404);
      expect((err as ApiError).message).toBe('building not found');
    }
  });
});
