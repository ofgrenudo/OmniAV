export interface TimeOption {
  value: string;
  label: string;
}

const pad = (n: number): string => String(n).padStart(2, '0');

export const to12Hour = (time: string): string => {
  const [hStr, mStr] = time.split(':');
  const h = Number(hStr);
  const period = h >= 12 ? 'PM' : 'AM';
  const displayHour = h % 12 === 0 ? 12 : h % 12;
  return `${displayHour}:${mStr} ${period}`;
};

/** Equipment can only be requested while AV staff are on site. */
export const SERVICE_START_TIME = '07:30';
export const SERVICE_END_TIME = '22:00';

export const minutesOfDay = (time: string): number => {
  const [h, m] = time.split(':').map(Number);
  return h * 60 + m;
};

export const generateTimeOptions = (
  startTime = SERVICE_START_TIME,
  endTime = SERVICE_END_TIME,
  stepMinutes = 10
): TimeOption[] => {
  const options: TimeOption[] = [];
  const last = minutesOfDay(endTime);
  for (let mins = minutesOfDay(startTime); mins <= last; mins += stepMinutes) {
    const value = `${pad(Math.floor(mins / 60))}:${pad(mins % 60)}`;
    options.push({ value, label: to12Hour(value) });
  }
  return options;
};

export const isWithinServiceHours = (time: string): boolean =>
  Boolean(time) && minutesOfDay(time) >= minutesOfDay(SERVICE_START_TIME) && minutesOfDay(time) <= minutesOfDay(SERVICE_END_TIME);

export const serviceHoursLabel = (): string => `${to12Hour(SERVICE_START_TIME)} to ${to12Hour(SERVICE_END_TIME)}`;

export const minAdvanceDate = (): Date => {
  const now = new Date();
  now.setHours(now.getHours() + 24);
  return now;
};

// minSelectableDate is the earliest date the calendar may offer: the day 24 hours out, or the day
// after it when the 24-hour cutoff already falls past the last bookable time, since every slot on
// that day would be rejected anyway.
export const minSelectableDate = (): Date => {
  const cutoff = minAdvanceDate();
  const cutoffMinutes = cutoff.getHours() * 60 + cutoff.getMinutes();
  if (cutoffMinutes >= minutesOfDay(SERVICE_END_TIME)) {
    cutoff.setDate(cutoff.getDate() + 1);
  }
  return cutoff;
};

export const toDateInputValue = (date: Date): string => {
  return `${date.getFullYear()}-${pad(date.getMonth() + 1)}-${pad(date.getDate())}`;
};

export const isAtLeast24HoursOut = (dateStr: string, timeStr: string): boolean => {
  if (!dateStr || !timeStr) return false;
  const [year, month, day] = dateStr.split('-').map(Number);
  const [hour, minute] = timeStr.split(':').map(Number);
  const target = new Date(year, month - 1, day, hour, minute);
  const minAllowed = new Date();
  minAllowed.setHours(minAllowed.getHours() + 24);
  return target.getTime() >= minAllowed.getTime();
};

export const weekdayFromDateString = (dateStr: string): string | null => {
  if (!dateStr) return null;
  const [year, month, day] = dateStr.split('-').map(Number);
  const date = new Date(year, month - 1, day);
  const labels = ['Sun', 'Mon', 'Tue', 'Wed', 'Thu', 'Fri', 'Sat'];
  return labels[date.getDay()];
};
