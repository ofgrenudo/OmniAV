import React from 'react';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import Inventory from './index';
import * as equipmentGroupsApi from '../../services/equipmentGroupsApi';
import * as equipmentApi from '../../services/equipmentApi';
import { EquipmentGroup } from '../../types/EquipmentGroup';
import { Equipment } from '../../types/Equipment';

jest.mock('../../services/equipmentGroupsApi', () => ({
  ...jest.requireActual('../../services/equipmentGroupsApi'),
  listEquipmentGroups: jest.fn(),
  getEquipmentGroup: jest.fn(),
  createEquipmentGroup: jest.fn(),
  updateEquipmentGroup: jest.fn(),
  archiveEquipmentGroup: jest.fn(),
}));

jest.mock('../../services/equipmentApi', () => ({
  ...jest.requireActual('../../services/equipmentApi'),
  listEquipment: jest.fn(),
  createEquipment: jest.fn(),
  updateEquipment: jest.fn(),
  archiveEquipment: jest.fn(),
}));

const mockedGroupsApi = equipmentGroupsApi as jest.Mocked<typeof equipmentGroupsApi>;
const mockedEquipmentApi = equipmentApi as jest.Mocked<typeof equipmentApi>;

const makeGroup = (overrides: Partial<EquipmentGroup> = {}): EquipmentGroup => ({
  id: 1,
  name: 'Cow Cart',
  description: null,
  disabled: false,
  archived: false,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
  ...overrides,
});

const makeUnit = (overrides: Partial<Equipment> = {}): Equipment => ({
  id: 1,
  name: 'Cow Cart A',
  description: null,
  disabled: false,
  archived: false,
  groupId: 1,
  createdAt: '2026-01-01T00:00:00Z',
  updatedAt: '2026-01-01T00:00:00Z',
  ...overrides,
});

const groupListResult = (groups: EquipmentGroup[]) => ({
  data: groups,
  meta: { page: 1, pageSize: 10, totalItems: groups.length, totalPages: 1 },
});

const unitListResult = (units: Equipment[]) => ({
  data: units,
  meta: { page: 1, pageSize: 100, totalItems: units.length, totalPages: 1 },
});

beforeEach(() => {
  jest.clearAllMocks();
});

test('renders the fetched equipment groups', async () => {
  mockedGroupsApi.listEquipmentGroups.mockResolvedValue(groupListResult([makeGroup()]));

  render(<Inventory />);

  expect(screen.getByText(/loading equipment groups/i)).toBeInTheDocument();
  expect(await screen.findByText('Cow Cart')).toBeInTheDocument();
  expect(mockedGroupsApi.listEquipmentGroups).toHaveBeenCalledWith(
    expect.objectContaining({ sort: 'name', order: 'asc', archived: false })
  );
});

test('shows an empty state when no groups match', async () => {
  mockedGroupsApi.listEquipmentGroups.mockResolvedValue(groupListResult([]));

  render(<Inventory />);

  expect(await screen.findByText(/no equipment groups match/i)).toBeInTheDocument();
});

test('creates a new equipment group and refreshes the list', async () => {
  mockedGroupsApi.listEquipmentGroups.mockResolvedValue(groupListResult([]));
  mockedGroupsApi.createEquipmentGroup.mockResolvedValue(makeGroup({ name: 'Projector Kit' }));

  render(<Inventory />);
  await screen.findByText(/no equipment groups match/i);

  userEvent.click(screen.getByRole('button', { name: /add equipment group/i }));
  userEvent.type(screen.getByLabelText(/^name/i), 'Projector Kit');

  mockedGroupsApi.listEquipmentGroups.mockResolvedValue(groupListResult([makeGroup({ name: 'Projector Kit' })]));
  userEvent.click(screen.getByRole('button', { name: /save/i }));

  await waitFor(() =>
    expect(mockedGroupsApi.createEquipmentGroup).toHaveBeenCalledWith(
      expect.objectContaining({ name: 'Projector Kit' })
    )
  );
  expect(await screen.findByText('Projector Kit')).toBeInTheDocument();
});

test('archives an active group', async () => {
  mockedGroupsApi.listEquipmentGroups.mockResolvedValue(groupListResult([makeGroup()]));
  mockedGroupsApi.archiveEquipmentGroup.mockResolvedValue(undefined);

  render(<Inventory />);
  await screen.findByText('Cow Cart');

  userEvent.click(screen.getByRole('button', { name: /^archive$/i }));

  await waitFor(() => expect(mockedGroupsApi.archiveEquipmentGroup).toHaveBeenCalledWith(1));
});

test('expands a group to show and add its units', async () => {
  mockedGroupsApi.listEquipmentGroups.mockResolvedValue(groupListResult([makeGroup()]));
  mockedEquipmentApi.listEquipment.mockResolvedValue(unitListResult([makeUnit()]));
  mockedEquipmentApi.createEquipment.mockResolvedValue(makeUnit({ id: 2, name: 'Cow Cart B' }));

  render(<Inventory />);
  await screen.findByText('Cow Cart');

  userEvent.click(screen.getByRole('button', { name: /^units$/i }));

  expect(await screen.findByText('Cow Cart A')).toBeInTheDocument();
  expect(mockedEquipmentApi.listEquipment).toHaveBeenCalledWith(
    expect.objectContaining({ groupId: 1, sort: 'name', order: 'asc' })
  );

  userEvent.click(screen.getByRole('button', { name: /add unit/i }));
  userEvent.type(screen.getByPlaceholderText(/e\.g\. cow cart a/i), 'Cow Cart B');

  mockedEquipmentApi.listEquipment.mockResolvedValue(unitListResult([makeUnit(), makeUnit({ id: 2, name: 'Cow Cart B' })]));
  userEvent.click(screen.getByRole('button', { name: /^save$/i }));

  await waitFor(() =>
    expect(mockedEquipmentApi.createEquipment).toHaveBeenCalledWith(
      expect.objectContaining({ name: 'Cow Cart B', groupId: 1 })
    )
  );
  expect(await screen.findByText('Cow Cart B')).toBeInTheDocument();
});
