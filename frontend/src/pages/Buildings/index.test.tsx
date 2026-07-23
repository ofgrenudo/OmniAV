import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import Buildings from './index';
import * as buildingsApi from '../../services/buildingsApi';
import { Building } from '../../types/Building';

jest.mock('../../services/buildingsApi', () => ({
  ...jest.requireActual('../../services/buildingsApi'),
  listBuildings: jest.fn(),
  getBuilding: jest.fn(),
  createBuilding: jest.fn(),
  updateBuilding: jest.fn(),
  archiveBuilding: jest.fn(),
}));

const mockedApi = buildingsApi as jest.Mocked<typeof buildingsApi>;

const makeBuilding = (overrides: Partial<Building> = {}): Building => ({
  id: 1,
  name: 'Anna Whitten Hall',
  archived: false,
  description: null,
  address: null,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
  ...overrides,
});

const listResult = (buildings: Building[], overrides: Partial<buildingsApi.PaginationMeta> = {}) => ({
  data: buildings,
  meta: { page: 1, pageSize: 10, totalItems: buildings.length, totalPages: 1, ...overrides },
});

beforeEach(() => {
  jest.clearAllMocks();
});

test('renders the fetched buildings using the default active filter', async () => {
  mockedApi.listBuildings.mockResolvedValue(listResult([makeBuilding()]));

  render(<Buildings />);

  expect(screen.getByText(/loading buildings/i)).toBeInTheDocument();
  expect(await screen.findByText('Anna Whitten Hall')).toBeInTheDocument();
  expect(mockedApi.listBuildings).toHaveBeenCalledWith(
    expect.objectContaining({ page: 1, pageSize: 10, sort: 'name', order: 'asc', archived: false })
  );
});

test('shows an error message when the request fails', async () => {
  mockedApi.listBuildings.mockRejectedValue(new buildingsApi.ApiError('boom', 500));

  render(<Buildings />);

  expect(await screen.findByText('boom')).toBeInTheDocument();
});

test('shows an empty state when no buildings match', async () => {
  mockedApi.listBuildings.mockResolvedValue(listResult([]));

  render(<Buildings />);

  expect(await screen.findByText(/no buildings match/i)).toBeInTheDocument();
});

test('searches by name after the user stops typing', async () => {
  mockedApi.listBuildings.mockResolvedValue(listResult([makeBuilding()]));

  render(<Buildings />);
  await screen.findByText('Anna Whitten Hall');

  userEvent.type(screen.getByLabelText(/search/i), 'whit');

  await waitFor(
    () => expect(mockedApi.listBuildings).toHaveBeenLastCalledWith(expect.objectContaining({ q: 'whit' })),
    { timeout: 2000 }
  );
});

test('creates a new building and refreshes the list', async () => {
  mockedApi.listBuildings.mockResolvedValue(listResult([]));
  mockedApi.createBuilding.mockResolvedValue(makeBuilding({ name: 'New Hall' }));

  render(<Buildings />);
  await screen.findByText(/no buildings match/i);

  userEvent.click(screen.getByRole('button', { name: /add building/i }));
  userEvent.type(screen.getByLabelText(/^name/i), 'New Hall');

  mockedApi.listBuildings.mockResolvedValue(listResult([makeBuilding({ name: 'New Hall' })]));
  userEvent.click(screen.getByRole('button', { name: /save/i }));

  await waitFor(() =>
    expect(mockedApi.createBuilding).toHaveBeenCalledWith(expect.objectContaining({ name: 'New Hall' }))
  );
  expect(await screen.findByText('New Hall')).toBeInTheDocument();
});

test('rejects submitting a blank name without calling the API', async () => {
  mockedApi.listBuildings.mockResolvedValue(listResult([]));

  render(<Buildings />);
  await screen.findByText(/no buildings match/i);

  userEvent.click(screen.getByRole('button', { name: /add building/i }));
  userEvent.click(screen.getByRole('button', { name: /save/i }));

  expect(await screen.findByText(/name is required/i)).toBeInTheDocument();
  expect(mockedApi.createBuilding).not.toHaveBeenCalled();
});

test('archives an active building', async () => {
  mockedApi.listBuildings.mockResolvedValue(listResult([makeBuilding()]));
  mockedApi.archiveBuilding.mockResolvedValue(undefined);

  render(<Buildings />);
  await screen.findByText('Anna Whitten Hall');

  userEvent.click(screen.getByRole('button', { name: /^archive$/i }));

  await waitFor(() => expect(mockedApi.archiveBuilding).toHaveBeenCalledWith(1));
});

test('restores an archived building via update', async () => {
  mockedApi.listBuildings.mockResolvedValue(listResult([makeBuilding({ archived: true })]));
  mockedApi.updateBuilding.mockResolvedValue(makeBuilding({ archived: false }));

  render(<Buildings />);
  await screen.findByText('Anna Whitten Hall');

  userEvent.click(screen.getByRole('button', { name: /^restore$/i }));

  await waitFor(() =>
    expect(mockedApi.updateBuilding).toHaveBeenCalledWith(1, expect.objectContaining({ archived: false }))
  );
});
