import React, { useState } from 'react';
import { Building } from '../../../types/Building';
import { WEEKDAYS, Weekday } from '../../../types/Weekday';
import { WhenWhereData } from '../../../types/WhenWhereData';
import { ROOM_NUMBER_HINT, isValidRoomNumber, sanitizeRoomNumber } from '../../../utils/room';
import {
  generateTimeOptions,
  isAtLeast24HoursOut,
  isWithinServiceHours,
  minSelectableDate,
  serviceHoursLabel,
  toDateInputValue,
} from '../../../utils/time';

interface WhenWhereStepProps {
  data: WhenWhereData;
  buildings: Building[];
  onChange: (data: WhenWhereData) => void;
  onContinue: () => void;
  onCancel: () => void;
}

type Errors = Partial<Record<keyof WhenWhereData, string>>;

const timeOptions = generateTimeOptions();

export const validateWhenWhere = (data: WhenWhereData): Errors => {
  const errors: Errors = {};

  if (!data.requestName.trim()) errors.requestName = 'Request name is required.';

  if (!data.firstDate) {
    errors.firstDate = 'First date needed is required.';
  } else if (data.firstDate < toDateInputValue(minSelectableDate())) {
    errors.firstDate = 'Requests must be made at least 24 hours in advance.';
  }

  if (!data.startTime) {
    errors.startTime = 'Start time is required.';
  } else if (!isWithinServiceHours(data.startTime)) {
    errors.startTime = `Start time must be between ${serviceHoursLabel()}.`;
  } else if (data.firstDate && !isAtLeast24HoursOut(data.firstDate, data.startTime)) {
    errors.startTime = 'Requests must be made at least 24 hours in advance.';
  }

  if (!data.endTime) {
    errors.endTime = 'End time is required.';
  } else if (!isWithinServiceHours(data.endTime)) {
    errors.endTime = `End time must be between ${serviceHoursLabel()}.`;
  } else if (data.startTime && data.endTime <= data.startTime) {
    errors.endTime = 'End time must be after start time.';
  }

  if (!data.weeks || data.weeks < 1) errors.weeks = 'Enter at least 1 week.';
  if (data.weeks > 52) errors.weeks = 'Enter no more than 52 weeks.';
  if (data.days.length === 0) errors.days = 'Select at least one day of the week.';
  if (!data.buildingId) errors.buildingId = 'Building is required.';

  if (!data.roomNumber.trim()) {
    errors.roomNumber = 'Room number is required.';
  } else if (!isValidRoomNumber(data.roomNumber)) {
    errors.roomNumber = ROOM_NUMBER_HINT;
  }

  return errors;
};

const WhenWhereStep: React.FC<WhenWhereStepProps> = ({ data, buildings, onChange, onContinue, onCancel }) => {
  const [touched, setTouched] = useState<Partial<Record<keyof WhenWhereData, boolean>>>({});
  const [submitAttempted, setSubmitAttempted] = useState(false);
  const minDate = toDateInputValue(minSelectableDate());

  const errors = validateWhenWhere(data);
  // A field's error stays hidden until the user has touched it (or tried to continue), so the form
  // doesn't open covered in complaints about fields they haven't reached yet.
  const errorFor = (field: keyof WhenWhereData): string | undefined =>
    submitAttempted || touched[field] ? errors[field] : undefined;

  const update = <K extends keyof WhenWhereData>(field: K, value: WhenWhereData[K]) => {
    setTouched((prev) => ({ ...prev, [field]: true }));
    onChange({ ...data, [field]: value });
  };

  const toggleDay = (day: Weekday) => {
    const days = data.days.includes(day) ? data.days.filter((d) => d !== day) : [...data.days, day];
    update('days', days);
  };

  const handleContinue = () => {
    setSubmitAttempted(true);
    if (Object.keys(errors).length === 0) onContinue();
  };

  // Times that fall inside the 24-hour notice window on the chosen date can't be honored, so they
  // aren't selectable at all rather than being rejected after the fact.
  const startDisabled = (time: string): boolean => Boolean(data.firstDate) && !isAtLeast24HoursOut(data.firstDate, time);

  return (
    <div className="request-step">
      <h2 className="request-step__title">When &amp; Where</h2>
      <p className="request-step__subtitle">
        Requests must be made at least 24 hours in advance, for times between {serviceHoursLabel()}.
      </p>

      <div className="form-field">
        <label htmlFor="requestName">Request Name *</label>
        <input
          id="requestName"
          type="text"
          value={data.requestName}
          onChange={(e) => update('requestName', e.target.value)}
          onBlur={() => setTouched((prev) => ({ ...prev, requestName: true }))}
          placeholder="e.g. BIO 101 Lecture, Fall Career Fair"
        />
        <p className="form-field__hint">Tip: use the course name or event name.</p>
        {errorFor('requestName') && <p className="form-field__error">{errorFor('requestName')}</p>}
      </div>

      <div className="form-field">
        <label htmlFor="firstDate">First Date Needed *</label>
        <input
          id="firstDate"
          type="date"
          min={minDate}
          value={data.firstDate}
          onChange={(e) => update('firstDate', e.target.value)}
          onBlur={() => setTouched((prev) => ({ ...prev, firstDate: true }))}
        />
        <p className="form-field__hint">Earliest bookable date: {minDate}.</p>
        {errorFor('firstDate') && <p className="form-field__error">{errorFor('firstDate')}</p>}
      </div>

      <div className="form-row">
        <div className="form-field">
          <label htmlFor="startTime">From *</label>
          <select id="startTime" value={data.startTime} onChange={(e) => update('startTime', e.target.value)}>
            <option value="">Select time</option>
            {timeOptions.map((opt) => (
              <option key={opt.value} value={opt.value} disabled={startDisabled(opt.value)}>
                {opt.label}
              </option>
            ))}
          </select>
          {errorFor('startTime') && <p className="form-field__error">{errorFor('startTime')}</p>}
        </div>

        <div className="form-field">
          <label htmlFor="endTime">To *</label>
          <select id="endTime" value={data.endTime} onChange={(e) => update('endTime', e.target.value)}>
            <option value="">Select time</option>
            {timeOptions.map((opt) => (
              <option key={opt.value} value={opt.value} disabled={Boolean(data.startTime) && opt.value <= data.startTime}>
                {opt.label}
              </option>
            ))}
          </select>
          {errorFor('endTime') && <p className="form-field__error">{errorFor('endTime')}</p>}
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
            onBlur={() => setTouched((prev) => ({ ...prev, weeks: true }))}
          />
          {errorFor('weeks') && <p className="form-field__error">{errorFor('weeks')}</p>}
        </div>
      </div>
      <p className="form-field__hint">Times available {serviceHoursLabel()} in 10-minute intervals.</p>

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
        {errorFor('days') && <p className="form-field__error">{errorFor('days')}</p>}
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
          {errorFor('buildingId') && <p className="form-field__error">{errorFor('buildingId')}</p>}
        </div>

        <div className="form-field">
          <label htmlFor="roomNumber">Room Number *</label>
          <input
            id="roomNumber"
            type="text"
            inputMode="text"
            value={data.roomNumber}
            onChange={(e) => update('roomNumber', sanitizeRoomNumber(e.target.value))}
            onBlur={() => setTouched((prev) => ({ ...prev, roomNumber: true }))}
            placeholder="123a"
          />
          <p className="form-field__hint">{ROOM_NUMBER_HINT}</p>
          {errorFor('roomNumber') && <p className="form-field__error">{errorFor('roomNumber')}</p>}
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
