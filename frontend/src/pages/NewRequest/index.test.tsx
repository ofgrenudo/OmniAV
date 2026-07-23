import React from 'react';
import { MemoryRouter } from 'react-router-dom';
import { fireEvent, render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import NewRequest from './index';
import * as buildingsApi from '../../services/buildingsApi';
import * as equipmentGroupsApi from '../../services/equipmentGroupsApi';
import * as requestsApi from '../../services/requestsApi';
import { Building } from '../../types/Building';
import { EquipmentGroup } from '../../types/EquipmentGroup';
import { Request } from '../../types/Request';

jest.mock('../../services/buildingsApi', () => ({
  ...jest.requireActual('../../services/buildingsApi'),
  listBuildings: jest.fn(),
}));

jest.mock('../../services/equipmentGroupsApi', () => ({
  ...jest.requireActual('../../services/equipmentGroupsApi'),
  listEquipmentGroups: jest.fn(),
  getEquipmentGroupAvailability: jest.fn(),
}));

jest.mock('../../services/requestsApi', () => ({
  ...jest.requireActual('../../services/requestsApi'),
  createRequest: jest.fn(),
  assignEquipment: jest.fn(),
  getRequest: jest.fn(),
}));

const mockedBuildingsApi = buildingsApi as jest.Mocked<typeof buildingsApi>;
const mockedGroupsApi = equipmentGroupsApi as jest.Mocked<typeof equipmentGroupsApi>;
const mockedRequestsApi = requestsApi as jest.Mocked<typeof requestsApi>;

const building: Building = {
  id: 7,
  name: 'Anna Whitten Hall',
  archived: false,
  description: null,
  address: null,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
};

const group: EquipmentGroup = {
  id: 3,
  name: 'Cow Cart',
  description: null,
  disabled: false,
  archived: false,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
};

const createdRequest: Request = {
  id: 42,
  name: 'Bio Lecture',
  firstDateNeeded: '2026-08-02T00:00:00Z',
  startTime: '0000-01-01T09:00:00Z',
  endTime: '0000-01-01T10:00:00Z',
  numberOfWeeks: 1,
  daysOfWeek: 'Sunday',
  buildingId: 7,
  room: '204',
  comments: null,
  attachmentId: null,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
};

const futureDateInputValue = (): string => {
  const d = new Date();
  d.setDate(d.getDate() + 10);
  const pad = (n: number) => String(n).padStart(2, '0');
  return `${d.getFullYear()}-${pad(d.getMonth() + 1)}-${pad(d.getDate())}`;
};

const fillWhenWhereStep = async () => {
  userEvent.type(screen.getByLabelText(/request name/i), 'Bio Lecture');
  fireEvent.change(screen.getByLabelText(/first date needed/i), { target: { value: futureDateInputValue() } });
  userEvent.selectOptions(screen.getByLabelText(/^from/i), '09:00');
  userEvent.selectOptions(screen.getByLabelText(/^to/i), '10:00');
  userEvent.click(screen.getByRole('button', { name: 'Mon' }));

  await waitFor(() => expect(screen.getByLabelText(/building/i)).not.toBeDisabled());
  userEvent.selectOptions(screen.getByLabelText(/building/i), '7');
  userEvent.type(screen.getByLabelText(/room number/i), '204');

  userEvent.click(screen.getByRole('button', { name: /continue/i }));
};

beforeEach(() => {
  jest.clearAllMocks();
  mockedBuildingsApi.listBuildings.mockResolvedValue({
    data: [building],
    meta: { page: 1, pageSize: 100, totalItems: 1, totalPages: 1 },
  });
  mockedGroupsApi.listEquipmentGroups.mockResolvedValue({
    data: [group],
    meta: { page: 1, pageSize: 100, totalItems: 1, totalPages: 1 },
  });
  mockedGroupsApi.getEquipmentGroupAvailability.mockResolvedValue(2);
});

test('submits a request and assigns the selected equipment', async () => {
  mockedRequestsApi.createRequest.mockResolvedValue(createdRequest);
  mockedRequestsApi.assignEquipment.mockResolvedValue({
    id: 1,
    groupId: group.id,
    equipmentId: 10,
    requestId: createdRequest.id,
  });
  mockedRequestsApi.getRequest.mockResolvedValue({
    ...createdRequest,
    requestedEquipment: [
      {
        id: 1,
        groupId: group.id,
        equipmentId: 10,
        equipment: {
          id: 10,
          name: 'Cow Cart A',
          description: null,
          disabled: false,
          archived: false,
          groupId: group.id,
          createdAt: '',
          updatedAt: '',
        },
        requestId: createdRequest.id,
      },
    ],
  });

  render(
    <MemoryRouter>
      <NewRequest />
    </MemoryRouter>
  );

  await waitFor(() => expect(mockedBuildingsApi.listBuildings).toHaveBeenCalled());
  await fillWhenWhereStep();

  expect(await screen.findByText('Cow Cart')).toBeInTheDocument();
  expect(mockedGroupsApi.getEquipmentGroupAvailability).toHaveBeenCalledWith(
    group.id,
    expect.objectContaining({ startTime: '09:00', endTime: '10:00' })
  );

  userEvent.click(screen.getByRole('button', { name: /increase cow cart quantity/i }));
  userEvent.click(screen.getByRole('button', { name: /continue/i }));

  userEvent.click(await screen.findByRole('button', { name: /submit request/i }));

  await waitFor(() =>
    expect(mockedRequestsApi.createRequest).toHaveBeenCalledWith(
      expect.objectContaining({ name: 'Bio Lecture', buildingId: 7, room: '204' })
    )
  );
  await waitFor(() => expect(mockedRequestsApi.assignEquipment).toHaveBeenCalledWith(createdRequest.id, group.id));
  expect(mockedRequestsApi.assignEquipment).toHaveBeenCalledTimes(1);

  expect(await screen.findByText(/request submitted/i)).toBeInTheDocument();
  expect(screen.getByText('42')).toBeInTheDocument();
  expect(screen.getByText('Cow Cart A')).toBeInTheDocument();
});

test('shows an error and still confirms the request if equipment assignment fails', async () => {
  mockedRequestsApi.createRequest.mockResolvedValue(createdRequest);
  mockedRequestsApi.assignEquipment.mockRejectedValue(
    new requestsApi.ApiError('no equipment in this group is available for the requested time', 409)
  );
  mockedRequestsApi.getRequest.mockResolvedValue(createdRequest);

  render(
    <MemoryRouter>
      <NewRequest />
    </MemoryRouter>
  );

  await waitFor(() => expect(mockedBuildingsApi.listBuildings).toHaveBeenCalled());
  await fillWhenWhereStep();

  await screen.findByText('Cow Cart');
  userEvent.click(screen.getByRole('button', { name: /increase cow cart quantity/i }));
  userEvent.click(screen.getByRole('button', { name: /continue/i }));
  userEvent.click(await screen.findByRole('button', { name: /submit request/i }));

  expect(await screen.findByText(/request submitted/i)).toBeInTheDocument();
  expect(await screen.findByText(/not all equipment could be assigned/i)).toBeInTheDocument();
});
