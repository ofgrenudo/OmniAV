import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import Stepper from './Stepper';
import WhenWhereStep from './steps/WhenWhereStep';
import EquipmentStep from './steps/EquipmentStep';
import ReviewStep from './steps/ReviewStep';
import { listBuildings } from '../../services/buildingsApi';
import { getEquipmentGroupAvailability, listEquipmentGroups } from '../../services/equipmentGroupsApi';
import { assignEquipment, createRequest, getRequest } from '../../services/requestsApi';
import { errorMessage } from '../../utils/apiError';
import { Building } from '../../types/Building';
import { EquipmentItem } from '../../types/EquipmentItem';
import { Request } from '../../types/Request';
import { SelectedEquipment } from '../../types/SelectedEquipment';
import { createEmptyWhenWhere } from '../../types/WhenWhereData';
import './NewRequest.css';

const STEPS = ['When & Where', 'Equipment', 'Review & Submit'];

const NewRequest: React.FC = () => {
  const navigate = useNavigate();
  const [currentStep, setCurrentStep] = useState(1);
  const [whenWhere, setWhenWhere] = useState(createEmptyWhenWhere());
  const [selectedEquipment, setSelectedEquipment] = useState<SelectedEquipment[]>([]);
  const [submittedRequest, setSubmittedRequest] = useState<Request | null>(null);
  const [submitting, setSubmitting] = useState(false);
  const [submitError, setSubmitError] = useState<string | null>(null);
  const [buildings, setBuildings] = useState<Building[]>([]);
  const [equipment, setEquipment] = useState<EquipmentItem[]>([]);
  const [equipmentLoading, setEquipmentLoading] = useState(false);
  const [equipmentError, setEquipmentError] = useState<string | null>(null);

  useEffect(() => {
    listBuildings({ pageSize: 100, sort: 'name', order: 'asc', archived: false })
      .then((result) => setBuildings(result.data))
      .catch(() => setBuildings([]));
  }, []);

  useEffect(() => {
    if (currentStep !== 2) return;
    if (!whenWhere.firstDate || !whenWhere.startTime || !whenWhere.endTime) return;
    if (!whenWhere.buildingId) return;

    const buildingId = Number(whenWhere.buildingId);
    let cancelled = false;
    setEquipmentLoading(true);
    setEquipmentError(null);

    // Both calls are scoped to the selected building: the list drops groups with no units stocked
    // here, and availability counts only the units this room could actually receive.
    listEquipmentGroups({ pageSize: 100, sort: 'name', order: 'asc', archived: false, disabled: false, buildingId })
      .then((result) =>
        Promise.all(
          result.data.map(async (group) => {
            const available = await getEquipmentGroupAvailability(group.id, {
              firstDate: whenWhere.firstDate,
              startTime: whenWhere.startTime,
              endTime: whenWhere.endTime,
              weeks: whenWhere.weeks,
              buildingId,
            });
            return { id: group.id, name: group.name, available };
          })
        )
      )
      .then((items) => {
        if (cancelled) return;
        setEquipment(items);
        // Going back and switching buildings can leave selections for groups that aren't stocked
        // in the new one; drop those instead of carrying them into the submitted request.
        const availableIds = new Set(items.map((item) => item.id));
        setSelectedEquipment((prev) => prev.filter((s) => availableIds.has(s.itemId)));
      })
      .catch((err) => {
        if (cancelled) return;
        setEquipment([]);
        setEquipmentError(errorMessage(err, 'Failed to load equipment availability.'));
      })
      .finally(() => {
        if (!cancelled) setEquipmentLoading(false);
      });

    return () => {
      cancelled = true;
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [currentStep]);

  const building = buildings.find((b) => String(b.id) === whenWhere.buildingId);

  const handleSubmit = async () => {
    setSubmitting(true);
    setSubmitError(null);
    try {
      const created = await createRequest({
        name: whenWhere.requestName,
        firstDateNeeded: whenWhere.firstDate,
        startTime: whenWhere.startTime,
        endTime: whenWhere.endTime,
        numberOfWeeks: whenWhere.weeks,
        buildingId: Number(whenWhere.buildingId),
        room: whenWhere.roomNumber,
        comments: whenWhere.comments.trim() ? whenWhere.comments.trim() : null,
      });

      try {
        for (const item of selectedEquipment) {
          for (let i = 0; i < item.quantity; i += 1) {
            // eslint-disable-next-line no-await-in-loop
            await assignEquipment(created.id, item.itemId);
          }
        }
        setSubmittedRequest(await getRequest(created.id));
      } catch (assignErr) {
        setSubmittedRequest(await getRequest(created.id).catch(() => created));
        setSubmitError(
          `Request #${created.id} was created, but not all equipment could be assigned: ` +
            `${errorMessage(assignErr, 'unknown error')}. Contact AV support if you still need those items.`
        );
      }
    } catch (err) {
      setSubmitError(errorMessage(err, 'Failed to submit request.'));
    } finally {
      setSubmitting(false);
    }
  };

  const handleStartNew = () => {
    setWhenWhere(createEmptyWhenWhere());
    setSelectedEquipment([]);
    setSubmittedRequest(null);
    setSubmitError(null);
    setCurrentStep(1);
  };

  return (
    <div className="new-request">
      <Stepper steps={STEPS} currentStep={currentStep} />

      <div className="new-request__card">
        {currentStep === 1 && (
          <WhenWhereStep
            data={whenWhere}
            buildings={buildings}
            onChange={setWhenWhere}
            onContinue={() => setCurrentStep(2)}
            onCancel={() => navigate('/')}
          />
        )}

        {currentStep === 2 && (
          <EquipmentStep
            items={equipment}
            buildingName={building?.name ?? ''}
            selected={selectedEquipment}
            loading={equipmentLoading}
            error={equipmentError}
            onChange={setSelectedEquipment}
            onBack={() => setCurrentStep(1)}
            onContinue={() => setCurrentStep(3)}
          />
        )}

        {currentStep === 3 && (
          <ReviewStep
            whenWhere={whenWhere}
            buildingName={building?.name ?? ''}
            equipment={selectedEquipment}
            submittedRequest={submittedRequest}
            submitting={submitting}
            submitError={submitError}
            onBack={() => setCurrentStep(2)}
            onSubmit={handleSubmit}
            onStartNew={handleStartNew}
          />
        )}
      </div>
    </div>
  );
};

export default NewRequest;
