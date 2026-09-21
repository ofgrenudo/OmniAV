import {
  SERVICE_END_TIME,
  SERVICE_START_TIME,
  generateTimeOptions,
  isAtLeast24HoursOut,
  isWithinServiceHours,
  minSelectableDate,
  toDateInputValue,
} from './time';

describe('generateTimeOptions', () => {
  const options = generateTimeOptions();

  it('starts at 7:30 AM and ends at 10:00 PM', () => {
    expect(options[0]).toEqual({ value: '07:30', label: '7:30 AM' });
    expect(options[options.length - 1]).toEqual({ value: '22:00', label: '10:00 PM' });
  });

  it('steps every 10 minutes with nothing outside service hours', () => {
    expect(options).toHaveLength((22 * 60 - (7 * 60 + 30)) / 10 + 1);
    expect(options.every((opt) => isWithinServiceHours(opt.value))).toBe(true);
  });
});

describe('isWithinServiceHours', () => {
  it.each([SERVICE_START_TIME, '12:00', SERVICE_END_TIME])('accepts %s', (time) => {
    expect(isWithinServiceHours(time)).toBe(true);
  });

  it.each(['07:00', '07:20', '22:10', '23:30', ''])('rejects %s', (time) => {
    expect(isWithinServiceHours(time)).toBe(false);
  });
});

describe('the 24-hour notice window', () => {
  const at = (offsetHours: number): { date: string; time: string } => {
    const d = new Date();
    d.setHours(d.getHours() + offsetHours);
    const pad = (n: number) => String(n).padStart(2, '0');
    return { date: toDateInputValue(d), time: `${pad(d.getHours())}:${pad(d.getMinutes())}` };
  };

  it('rejects a time less than 24 hours out', () => {
    const { date, time } = at(23);
    expect(isAtLeast24HoursOut(date, time)).toBe(false);
  });

  it('accepts a time more than 24 hours out', () => {
    const { date, time } = at(25);
    expect(isAtLeast24HoursOut(date, time)).toBe(true);
  });

  it('never offers a date earlier than tomorrow', () => {
    const today = toDateInputValue(new Date());
    expect(toDateInputValue(minSelectableDate()) > today).toBe(true);
  });
});
