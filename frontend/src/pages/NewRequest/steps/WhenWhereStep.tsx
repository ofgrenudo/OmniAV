import React, { useMemo, useState } from 'react';
import { Building } from '../../../types/Building';
import { WEEKDAYS, Weekday } from '../../../types/Weekday';
import { WhenWhereData } from '../../../types/WhenWhereData';
import { generateTimeOptions, isAtLeast24HoursOut, minAdvanceDate, toDateInputValue } from '../../../utils/time';

interface WhenWhereStepProps {
  data: WhenWhereData;
  buildings: Building[];
  onChange: (data: WhenWhereData) => void;
  onContinue: () => void;
  onCancel: () => void;
}

type Errors = Partial<Record<keyof WhenWhereData, string>>;

const timeOptions = generateTimeOptions();

const WhenWhereStep: React.FC<WhenWhereStepProps> = ({ data, buildings, onChange, onContinue, onCancel }) => {
  const [errors, setErrors] = useState<Errors>({});
  const minDate = useMemo(() => toDateInputValue(minAdvanceDate()), []);

  const update = <K extends keyof WhenWhereData>(field: K, value: WhenWhereData[K]) => {
    onChange({ ...data, [field]: value });
  };

  const toggleDay = (day: Weekday) => {
    const days = data.days.includes(day) ? data.days.filter((d) => d !== day) : [...data.days, day];
    update('days', days);
  };

  const validate = (): boolean => {
    const next: Errors = {};
    if (!data.requestName.trim()) next.requestName = 'Request name is required.';
    if (!data.firstDate) next.firstDate = 'First date needed is required.';
    if (!data.startTime) next.startTime = 'Start time is required.';
    if (!data.endTime) next.endTime = 'End time is required.';
    if (data.startTime && data.endTime && data.endTime <= data.startTime) {
      next.endTime = 'End time must be after start time.';
    }
    if (data.firstDate && data.startTime && !isAtLeast24HoursOut(data.firstDate, data.startTime)) {
      next.firstDate = 'Requests must be made at least 24 hours in advance.';
    }
    if (!data.weeks || data.weeks < 1) next.weeks = 'Enter at least 1 week.';
    if (data.days.length === 0) next.days = 'Select at least one day of the week.';
    if (!data.buildingId) next.buildingId = 'Building is required.';
    if (!data.roomNumber.trim()) next.roomNumber = 'Room number is required.';

    setErrors(next);
    return Object.keys(next).length === 0;
  };

  const handleContinue = () => {
    if (validate()) onContinue();
  };

  return (
    <div className="request-step">
      <h2 className="request-step__title">When &amp; Where</h2>
      <p className="request-step__subtitle">Requests must be made at least 24 hours in advance.</p>

      <div className="form-field">
        <label htmlFor="requestName">Request Name *</label>
        <input
          id="requestName"
          type="text"
          value={data.requestName}
          onChange={(e) => update('requestName', e.target.value)}
          placeholder="e.g. BIO 101 Lecture, Fall Career Fair"
        />
        <p className="form-field__hint">Tip: use the course name or event name.</p>
        {errors.requestName && <p className="form-field__error">{errors.requestName}</p>}
      </div>

      <div className="form-field">
        <label htmlFor="firstDate">First Date Needed *</label>
        <input
          id="firstDate"
          type="date"
          min={minDate}
          value={data.firstDate}
          onChange={(e) => update('firstDate', e.target.value)}
        />
        {errors.firstDate && <p className="form-field__error">{errors.firstDate}</p>}
      </div>

      <div className="form-row">
        <div className="form-field">
          <label htmlFor="startTime">From *</label>
          <select id="startTime" value={data.startTime} onChange={(e) => update('startTime', e.target.value)}>
            <option value="">Select time</option>
            {timeOptions.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
          {errors.startTime && <p className="form-field__error">{errors.startTime}</p>}
        </div>

        <div className="form-field">
          <label htmlFor="endTime">To *</label>
          <select id="endTime" value={data.endTime} onChange={(e) => update('endTime', e.target.value)}>
            <option value="">Select time</option>
            {timeOptions.map((opt) => (
              <option key={opt.value} value={opt.value}>
                {opt.label}
              </option>
            ))}
          </select>
          {errors.endTime && <p className="form-field__error">{errors.endTime}</p>}
        </div>

        <div className="form-field">
          <label htmlFor="weeks">Weeks *</label>
          <input
            id="weeks"
            type="number"
            min={1}
            max={52}
            value={data.weeks}
            onChange={(e) => update('weeks', Number(e.target.value))}
          />
          {errors.weeks && <p className="form-field__error">{errors.weeks}</p>}
        </div>
      </div>
      <p className="form-field__hint">Times valid 07:00–23:30 in 10-minute intervals.</p>

      <div className="form-field">
        <span className="form-field__label-text">Days of the Week</span>
        <div className="weekday-toggle">
          {WEEKDAYS.map((day) => (
            <button
              key={day}
              type="button"
              className={`weekday-toggle__day ${data.days.includes(day) ? 'weekday-toggle__day--selected' : ''}`}
              onClick={() => toggleDay(day)}
            >
              {day}
            </button>
          ))}
        </div>
        {errors.days && <p className="form-field__error">{errors.days}</p>}
      </div>

      <div className="form-row">
        <div className="form-field">
          <label htmlFor="buildingId">Building *</label>
          <select id="buildingId" value={data.buildingId} onChange={(e) => update('buildingId', e.target.value)}>
            <option value="">Select a building</option>
            {buildings.map((building) => (
              <option key={building.id} value={building.id}>
                {building.name}
              </option>
            ))}
          </select>
          {errors.buildingId && <p className="form-field__error">{errors.buildingId}</p>}
        </div>

        <div className="form-field">
          <label htmlFor="roomNumber">Room Number *</label>
          <input
            id="roomNumber"
            type="text"
            value={data.roomNumber}
            onChange={(e) => update('roomNumber', e.target.value)}
            placeholder="123a"
          />
          {errors.roomNumber && <p className="form-field__error">{errors.roomNumber}</p>}
        </div>
      </div>

      <div className="form-field">
        <label htmlFor="comments">Comments</label>
        <textarea
          id="comments"
          rows={3}
          value={data.comments}
          onChange={(e) => update('comments', e.target.value)}
          placeholder="Any special instructions..."
        />
      </div>

      <div className="form-field">
        <label htmlFor="attachment">Attachment (disabled)</label>
        <input id="attachment" type="file" disabled />
        <p className="form-field__hint">File upload disabled</p>
      </div>

      <div className="request-step__actions">
        <button type="button" className="btn btn--secondary" onClick={onCancel}>
          Cancel
        </button>
        <button type="button" className="btn btn--primary" onClick={handleContinue}>
          Continue
        </button>
      </div>
    </div>
  );
};

export default WhenWhereStep;
