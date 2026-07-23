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

export const generateTimeOptions = (
  startHour = 7,
  endHour = 23,
  endMinute = 30,
  stepMinutes = 10
): TimeOption[] => {
  const options: TimeOption[] = [];
  let h = startHour;
  let m = 0;
  while (h < endHour || (h === endHour && m <= endMinute)) {
    const value = `${pad(h)}:${pad(m)}`;
    options.push({ value, label: to12Hour(value) });
    m += stepMinutes;
    if (m >= 60) {
      m -= 60;
      h += 1;
    }
  }
  return options;
};

export const minAdvanceDate = (): Date => {
  const now = new Date();
  now.setHours(now.getHours() + 24);
  return now;
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
