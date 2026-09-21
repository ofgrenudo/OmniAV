import React, { useEffect, useMemo, useState } from 'react';
import { ScheduleEntry, getEquipmentGroupSchedule } from '../../services/equipmentGroupsApi';
import { errorMessage } from '../../utils/apiError';
import { SERVICE_END_TIME, SERVICE_START_TIME, minutesOfDay, to12Hour, toDateInputValue } from '../../utils/time';

interface EquipmentTrackPanelProps {
  groupId: number;
  groupName: string;
}

// The timeline spans service hours, but a booking made before those hours were tightened could
// still fall outside them — widen the window rather than draw a bar off the edge of the chart.
const timelineBounds = (entries: ScheduleEntry[]): { start: number; end: number } => {
  const starts = [minutesOfDay(SERVICE_START_TIME), ...entries.map((e) => minutesOfDay(e.startTime))];
  const ends = [minutesOfDay(SERVICE_END_TIME), ...entries.map((e) => minutesOfDay(e.endTime))];
  return { start: Math.min(...starts), end: Math.max(...ends) };
};

const hourTicks = (start: number, end: number): number[] => {
  const ticks: number[] = [];
  for (let h = Math.ceil(start / 60); h * 60 <= end; h += 1) ticks.push(h * 60);
  return ticks;
};

const pad = (n: number): string => String(n).padStart(2, '0');
const hourLabel = (minutes: number): string => to12Hour(`${pad(Math.floor(minutes / 60))}:${pad(minutes % 60)}`);

// Shared building → unit ordering for both the chart rows and the summary table below it, so the
// two views can't drift apart on how they order units.
const compareByBuildingThenUnit = (aBuilding: string, aUnit: string, bBuilding: string, bUnit: string): number => {
  const byBuilding = aBuilding.localeCompare(bBuilding);
  return byBuilding !== 0 ? byBuilding : aUnit.localeCompare(bUnit);
};

const EquipmentTrackPanel: React.FC<EquipmentTrackPanelProps> = ({ groupId, groupName }) => {
  const [date, setDate] = useState(() => toDateInputValue(new Date()));
  const [entries, setEntries] = useState<ScheduleEntry[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);

    getEquipmentGroupSchedule(groupId, date)
      .then((data) => {
        if (!cancelled) setEntries(data);
      })
      .catch((err) => {
        if (cancelled) return;
        setEntries([]);
        setError(errorMessage(err, 'Failed to load the schedule for this group.'));
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [groupId, date]);

  const { start, end } = useMemo(() => timelineBounds(entries), [entries]);
  const span = Math.max(end - start, 1);
  const ticks = useMemo(() => hourTicks(start, end), [start, end]);
  const offset = (minutes: number): number => ((minutes - start) / span) * 100;

  // One row per unit, so a single cart's whole day reads left to right. Rows are grouped by the
  // unit's home building (every booking for a unit is in that unit's building, since a request
  // can only draw from units stocked there) and sorted building → unit alphabetically.
  const rows = useMemo(() => {
    const byUnit = new Map<number, { name: string; buildingName: string; bookings: ScheduleEntry[] }>();
    entries.forEach((entry) => {
      const row =
        byUnit.get(entry.equipmentId) ??
        { name: entry.equipmentName, buildingName: entry.buildingName, bookings: [] };
      row.bookings.push(entry);
      byUnit.set(entry.equipmentId, row);
    });
    return Array.from(byUnit.entries())
      .map(([equipmentId, row]) => ({ equipmentId, ...row }))
      .sort((a, b) => compareByBuildingThenUnit(a.buildingName, a.name, b.buildingName, b.name));
  }, [entries]);

  // Bottom summary table lists every booking, sorted the same way the chart rows are (building →
  // unit) with start time as a further tiebreaker within a unit's own bookings, so the two views
  // scan in the same order.
  const sortedEntries = useMemo(
    () =>
      [...entries].sort((a, b) => {
        const byBuildingThenUnit = compareByBuildingThenUnit(
          a.buildingName,
          a.equipmentName,
          b.buildingName,
          b.equipmentName
        );
        return byBuildingThenUnit !== 0 ? byBuildingThenUnit : a.startTime.localeCompare(b.startTime);
      }),
    [entries]
  );

  const shiftDate = (days: number) => {
    const [y, m, d] = date.split('-').map(Number);
    const next = new Date(y, m - 1, d);
    next.setDate(next.getDate() + days);
    setDate(toDateInputValue(next));
  };

  return (
    <div className="track-panel">
      <div className="track-panel__header">
        <h4 className="units-panel__title">Where {groupName} units are on a given day</h4>
        <div className="track-panel__date">
          <button type="button" className="btn btn--secondary btn--small" onClick={() => shiftDate(-1)}>
            ‹ Previous day
          </button>
          <label className="track-panel__date-label" htmlFor={`track-date-${groupId}`}>
            Date
          </label>
          <input
            id={`track-date-${groupId}`}
            type="date"
            value={date}
            onChange={(e) => setDate(e.target.value)}
          />
          <button type="button" className="btn btn--secondary btn--small" onClick={() => shiftDate(1)}>
            Next day ›
          </button>
          <button
            type="button"
            className="btn btn--secondary btn--small"
            onClick={() => setDate(toDateInputValue(new Date()))}
          >
            Today
          </button>
        </div>
      </div>

      {error !== null && <p className="admin-page__status admin-page__status--error">{error}</p>}
      {error === null && loading && <p className="admin-page__status">Loading schedule…</p>}
      {error === null && !loading && entries.length === 0 && (
        <p className="admin-page__status">No {groupName} units are booked on this date.</p>
      )}

      {error === null && !loading && entries.length > 0 && (
        <>
          <div className="track-chart">
            <div className="track-chart__ruler">
              <span className="track-chart__unit-label" />
              <div className="track-chart__track">
                {ticks.map((tick) => (
                  <span key={tick} className="track-chart__tick" style={{ left: `${offset(tick)}%` }}>
                    {hourLabel(tick)}
                  </span>
                ))}
              </div>
            </div>

            {rows.map((row, i) => {
              // Insert a building header whenever the building changes, so the eye can chunk the
              // chart by physical location instead of scanning a flat list of unit names.
              const showBuildingHeader = i === 0 || rows[i - 1].buildingName !== row.buildingName;
              return (
                <React.Fragment key={row.equipmentId}>
                  {showBuildingHeader && (
                    <div className="track-chart__building-header">
                      <span className="track-chart__unit-label track-chart__building-name">
                        {row.buildingName}
                      </span>
                      <div className="track-chart__track" />
                    </div>
                  )}
                  <div className="track-chart__row">
                    <span className="track-chart__unit-label">{row.name}</span>
                    <div className="track-chart__track">
                      {ticks.map((tick) => (
                        <span key={tick} className="track-chart__gridline" style={{ left: `${offset(tick)}%` }} />
                      ))}
                      {row.bookings.map((booking) => (
                        <div
                          key={booking.requestId}
                          className="track-chart__bar"
                          style={{
                            left: `${offset(minutesOfDay(booking.startTime))}%`,
                            width: `${Math.max(
                              ((minutesOfDay(booking.endTime) - minutesOfDay(booking.startTime)) / span) * 100,
                              2
                            )}%`,
                          }}
                          title={
                            `${booking.requestName}\n` +
                            `${booking.buildingName} · Room ${booking.room}\n` +
                            `${to12Hour(booking.startTime)} – ${to12Hour(booking.endTime)}`
                          }
                        >
                          <span className="track-chart__bar-label">Room {booking.room}</span>
                        </div>
                      ))}
                    </div>
                  </div>
                </React.Fragment>
              );
            })}
          </div>

          <table className="data-table track-panel__table">
            <thead>
              <tr>
                <th>Building</th>
                <th>Unit</th>
                <th>Time</th>
                <th>Room</th>
                <th className="data-table__cell--wrap">Request</th>
              </tr>
            </thead>
            <tbody>
              {sortedEntries.map((entry) => (
                <tr key={`${entry.requestId}-${entry.equipmentId}`}>
                  <td>{entry.buildingName}</td>
                  <td>{entry.equipmentName}</td>
                  <td>
                    {to12Hour(entry.startTime)} – {to12Hour(entry.endTime)}
                  </td>
                  <td>{entry.room}</td>
                  <td className="data-table__cell--wrap">{entry.requestName}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </>
      )}
    </div>
  );
};

export default EquipmentTrackPanel;
