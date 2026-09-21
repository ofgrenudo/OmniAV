import { WhenWhereData, createEmptyWhenWhere } from '../../../types/WhenWhereData';
import { toDateInputValue } from '../../../utils/time';
import { validateWhenWhere } from './WhenWhereStep';

const daysOut = (n: number): string => {
  const d = new Date();
  d.setDate(d.getDate() + n);
  return toDateInputValue(d);
};

const validData = (overrides: Partial<WhenWhereData> = {}): WhenWhereData => ({
  ...createEmptyWhenWhere(),
  requestName: 'Bio Lecture',
  firstDate: daysOut(10),
  startTime: '09:00',
  endTime: '10:00',
  weeks: 1,
  days: ['Mon'],
  buildingId: '7',
  roomNumber: '204',
  ...overrides,
});

it('accepts a fully valid form', () => {
  expect(validateWhenWhere(validData())).toEqual({});
});

describe('room number', () => {
  it.each(['204', '123a', '123B', '12c'])('accepts %s', (roomNumber) => {
    expect(validateWhenWhere(validData({ roomNumber })).roomNumber).toBeUndefined();
  });

  it.each(['204d', '12 4', 'lobby', '204-A', '204.1'])('rejects %s', (roomNumber) => {
    expect(validateWhenWhere(validData({ roomNumber })).roomNumber).toBeDefined();
  });

  it('requires a room number', () => {
    expect(validateWhenWhere(validData({ roomNumber: '  ' })).roomNumber).toBe('Room number is required.');
  });
});

describe('24-hour notice', () => {
  it('rejects a date before the earliest bookable day', () => {
    expect(validateWhenWhere(validData({ firstDate: toDateInputValue(new Date()) })).firstDate).toMatch(
      /24 hours in advance/i
    );
  });

  it('rejects a time inside the notice window on an otherwise allowed date', () => {
    const soon = new Date();
    soon.setHours(soon.getHours() + 2);
    const data = validData({ firstDate: toDateInputValue(soon), startTime: '07:30', endTime: '10:00' });
    expect(validateWhenWhere(data).startTime ?? validateWhenWhere(data).firstDate).toMatch(/24 hours in advance/i);
  });
});

describe('service hours', () => {
  it.each(['07:00', '22:30', '23:00'])('rejects a start time of %s', (startTime) => {
    expect(validateWhenWhere(validData({ startTime, endTime: '23:30' })).startTime).toMatch(/7:30 AM to 10:00 PM/);
  });

  it('rejects an end time past 10:00 PM', () => {
    expect(validateWhenWhere(validData({ startTime: '21:00', endTime: '22:30' })).endTime).toMatch(
      /7:30 AM to 10:00 PM/
    );
  });

  it('rejects an end time at or before the start time', () => {
    expect(validateWhenWhere(validData({ startTime: '10:00', endTime: '10:00' })).endTime).toBe(
      'End time must be after start time.'
    );
  });
});
