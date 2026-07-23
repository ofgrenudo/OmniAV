import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import AllRequests from './index';
import * as requestsApi from '../../services/requestsApi';
import { Request } from '../../types/Request';

jest.mock('../../services/requestsApi', () => ({
  ...jest.requireActual('../../services/requestsApi'),
  listRequests: jest.fn(),
  getRequest: jest.fn(),
  deleteRequest: jest.fn(),
  unassignEquipment: jest.fn(),
}));

const mockedApi = requestsApi as jest.Mocked<typeof requestsApi>;

const makeRequest = (overrides: Partial<Request> = {}): Request => ({
  id: 1,
  name: 'Bio Lecture',
  firstDateNeeded: '2026-08-02T00:00:00Z',
  startTime: '0000-01-01T09:00:00Z',
  endTime: '0000-01-01T10:00:00Z',
  numberOfWeeks: 1,
  daysOfWeek: 'Sunday',
  buildingId: 1,
  room: '101',
  building: { id: 1, name: 'Anna Whitten Hall', archived: false, description: null, address: null, createdAt: '', updatedAt: '' },
  comments: null,
  attachmentId: null,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
  ...overrides,
});

const listResult = (requests: Request[]) => ({
  data: requests,
  meta: { page: 1, pageSize: 10, totalItems: requests.length, totalPages: 1 },
});

beforeEach(() => {
  jest.clearAllMocks();
});

test('renders the fetched requests', async () => {
  mockedApi.listRequests.mockResolvedValue(listResult([makeRequest()]));

  render(<AllRequests />);

  expect(screen.getByText(/loading requests/i)).toBeInTheDocument();
  expect(await screen.findByText('Bio Lecture')).toBeInTheDocument();
  expect(screen.getByText('Anna Whitten Hall')).toBeInTheDocument();
});

test('shows an empty state when no requests match', async () => {
  mockedApi.listRequests.mockResolvedValue(listResult([]));

  render(<AllRequests />);

  expect(await screen.findByText(/no requests match/i)).toBeInTheDocument();
});

test('searches by name after the user stops typing', async () => {
  mockedApi.listRequests.mockResolvedValue(listResult([makeRequest()]));

  render(<AllRequests />);
  await screen.findByText('Bio Lecture');

  userEvent.type(screen.getByLabelText(/search/i), 'bio');

  await waitFor(
    () => expect(mockedApi.listRequests).toHaveBeenLastCalledWith(expect.objectContaining({ q: 'bio' })),
    { timeout: 2000 }
  );
});

test('expands a request to show its assigned equipment', async () => {
  mockedApi.listRequests.mockResolvedValue(listResult([makeRequest()]));
  mockedApi.getRequest.mockResolvedValue(
    makeRequest({
      requestedEquipment: [
        { id: 1, groupId: 1, group: { id: 1, name: 'Cow Cart', description: null, disabled: false, archived: false, createdAt: '', updatedAt: '' }, equipmentId: 1, equipment: { id: 1, name: 'Cow Cart A', description: null, disabled: false, archived: false, groupId: 1, createdAt: '', updatedAt: '' }, requestId: 1 },
      ],
    })
  );

  render(<AllRequests />);
  await screen.findByText('Bio Lecture');

  userEvent.click(screen.getByRole('button', { name: /^equipment$/i }));

  expect(await screen.findByText('Cow Cart A')).toBeInTheDocument();
  expect(mockedApi.getRequest).toHaveBeenCalledWith(1);
});

test('removing an assignment calls unassignEquipment and reloads', async () => {
  mockedApi.listRequests.mockResolvedValue(listResult([makeRequest()]));
  mockedApi.getRequest.mockResolvedValue(
    makeRequest({
      requestedEquipment: [
        { id: 42, groupId: 1, equipmentId: 1, equipment: { id: 1, name: 'Cow Cart A', description: null, disabled: false, archived: false, groupId: 1, createdAt: '', updatedAt: '' }, requestId: 1 },
      ],
    })
  );
  mockedApi.unassignEquipment.mockResolvedValue(undefined);

  render(<AllRequests />);
  await screen.findByText('Bio Lecture');
  userEvent.click(screen.getByRole('button', { name: /^equipment$/i }));
  await screen.findByText('Cow Cart A');

  userEvent.click(screen.getByRole('button', { name: /remove/i }));

  await waitFor(() => expect(mockedApi.unassignEquipment).toHaveBeenCalledWith(1, 42));
});

test('cancelling a request requires confirmation', async () => {
  mockedApi.listRequests.mockResolvedValue(listResult([makeRequest()]));
  mockedApi.deleteRequest.mockResolvedValue(undefined);

  render(<AllRequests />);
  await screen.findByText('Bio Lecture');

  userEvent.click(screen.getByRole('button', { name: /cancel request/i }));
  expect(mockedApi.deleteRequest).not.toHaveBeenCalled();

  mockedApi.listRequests.mockResolvedValue(listResult([]));
  userEvent.click(screen.getByRole('button', { name: /confirm cancel/i }));

  await waitFor(() => expect(mockedApi.deleteRequest).toHaveBeenCalledWith(1));
});
