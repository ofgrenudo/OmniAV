import React from 'react';
import { useNavigate } from 'react-router-dom';
import { Request } from '../../../types/Request';
import { SelectedEquipment } from '../../../types/SelectedEquipment';
import { WhenWhereData } from '../../../types/WhenWhereData';
import { to12Hour } from '../../../utils/time';

interface ReviewStepProps {
  whenWhere: WhenWhereData;
  buildingName: string;
  equipment: SelectedEquipment[];
  submittedRequest: Request | null;
  submitting: boolean;
  submitError: string | null;
  onBack: () => void;
  onSubmit: () => void;
  onStartNew: () => void;
}

const ReviewStep: React.FC<ReviewStepProps> = ({
  whenWhere,
  buildingName,
  equipment,
  submittedRequest,
  submitting,
  submitError,
  onBack,
  onSubmit,
  onStartNew,
}) => {
  const navigate = useNavigate();

  if (submittedRequest) {
    return (
      <div className="request-step">
        <h2 className="request-step__title">Request Submitted</h2>
        <p className="request-step__subtitle">
          Your request has been submitted and is awaiting review.
        </p>
        <div className="review-panel">
          <p>
            Confirmation number: <strong>{submittedRequest.id}</strong>
          </p>
          {submittedRequest.requestedEquipment && submittedRequest.requestedEquipment.length > 0 && (
            <>
              <p>
                <strong>Assigned equipment:</strong>
              </p>
              <ul className="review-equipment-list">
                {submittedRequest.requestedEquipment.map((entry) => (
                  <li key={entry.id}>
                    <span>{entry.equipment?.name ?? `Unit #${entry.equipmentId}`}</span>
                  </li>
                ))}
              </ul>
            </>
          )}
        </div>
        {submitError && <p className="form-field__error">{submitError}</p>}
        <div className="request-step__actions">
          <button type="button" className="btn btn--secondary" onClick={onStartNew}>
            Submit Another Request
          </button>
          <button type="button" className="btn btn--primary" onClick={() => navigate('/requests')}>
            View All Requests
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="request-step">
      <h2 className="request-step__title">Review &amp; Submit</h2>
      <p className="request-step__subtitle">Confirm your request details below.</p>

      <div className="review-panel">
        <h3 className="review-panel__title">Request Details</h3>
        <dl className="review-list">
          <div className="review-list__row">
            <dt>Request Name</dt>
            <dd>{whenWhere.requestName}</dd>
          </div>
          <div className="review-list__row">
            <dt>Date &amp; Time</dt>
            <dd>
              {whenWhere.firstDate}
              <br />
              {to12Hour(whenWhere.startTime)} to {to12Hour(whenWhere.endTime)} · {whenWhere.weeks} week(s)
              <br />
              {whenWhere.days.join(', ')}
            </dd>
          </div>
          <div className="review-list__row">
            <dt>Location</dt>
            <dd>
              {buildingName}
              <br />
              Room {whenWhere.roomNumber}
            </dd>
          </div>
        </dl>
      </div>

      <div className="review-panel">
        <h3 className="review-panel__title">Equipment &amp; Notes</h3>
        {equipment.length > 0 ? (
          <ul className="review-equipment-list">
            {equipment.map((item) => (
              <li key={item.itemId}>
                <span>{item.name}</span>
                <span className="review-equipment-list__qty">Qty: {item.quantity}</span>
              </li>
            ))}
          </ul>
        ) : (
          <p className="request-step__subtitle">No equipment requested.</p>
        )}
        {whenWhere.comments && (
          <p className="review-panel__comments">
            <strong>Comments:</strong> {whenWhere.comments}
          </p>
        )}
      </div>

      <div className="review-panel review-panel--ready">
        <h3 className="review-panel__title">Ready to submit</h3>
        <p className="request-step__subtitle">Review the details above, then submit your request.</p>
      </div>

      {submitError && <p className="form-field__error">{submitError}</p>}

      <div className="request-step__actions">
        <button type="button" className="btn btn--secondary" onClick={onBack} disabled={submitting}>
          Back
        </button>
        <button type="button" className="btn btn--primary" onClick={onSubmit} disabled={submitting}>
          {submitting ? 'Submitting…' : 'Submit Request'}
        </button>
      </div>
    </div>
  );
};

export default ReviewStep;
