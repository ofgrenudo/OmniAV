import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import Stepper from './Stepper';
import WhenWhereStep from './steps/WhenWhereStep';
import EquipmentStep from './steps/EquipmentStep';
import ReviewStep from './steps/ReviewStep';
import { getBuildings } from '../../data/buildings';
import { getEquipment } from '../../data/inventory';
import { submitRequest } from '../../services/requestsStore';
import { AVRequest } from '../../types/AVRequest';
import { Building } from '../../types/Building';
import { EquipmentItem } from '../../types/EquipmentItem';
import { SelectedEquipment } from '../../types/SelectedEquipment';
import { createEmptyWhenWhere } from '../../types/WhenWhereData';
import './NewRequest.css';

const STEPS = ['When & Where', 'Equipment', 'Review & Submit'];

const NewRequest: React.FC = () => {
  const navigate = useNavigate();
  const [currentStep, setCurrentStep] = useState(1);
  const [whenWhere, setWhenWhere] = useState(createEmptyWhenWhere());
  const [selectedEquipment, setSelectedEquipment] = useState<SelectedEquipment[]>([]);
  const [submittedRequest, setSubmittedRequest] = useState<AVRequest | null>(null);
  const [buildings, setBuildings] = useState<Building[]>([]);
  const [equipment, setEquipment] = useState<EquipmentItem[]>([]);

  useEffect(() => {
    getBuildings().then(setBuildings);
    getEquipment().then(setEquipment);
  }, []);

  const building = buildings.find((b) => b.id === whenWhere.buildingId);

  const handleSubmit = () => {
    const result = submitRequest({ whenWhere, equipment: selectedEquipment });
    setSubmittedRequest(result);
  };

  const handleStartNew = () => {
    setWhenWhere(createEmptyWhenWhere());
    setSelectedEquipment([]);
    setSubmittedRequest(null);
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
