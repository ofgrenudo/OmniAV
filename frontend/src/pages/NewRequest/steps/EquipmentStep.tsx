import React from 'react';
import { EquipmentItem } from '../../../types/EquipmentItem';
import { SelectedEquipment } from '../../../types/SelectedEquipment';

interface EquipmentStepProps {
  items: EquipmentItem[];
  buildingName: string;
  selected: SelectedEquipment[];
  onChange: (selected: SelectedEquipment[]) => void;
  onBack: () => void;
  onContinue: () => void;
}

const EquipmentStep: React.FC<EquipmentStepProps> = ({
  items,
  buildingName,
  selected,
  onChange,
  onBack,
  onContinue,
}) => {
  const quantityFor = (itemId: string): number => selected.find((s) => s.itemId === itemId)?.quantity ?? 0;

  const setQuantity = (item: EquipmentItem, quantity: number) => {
    const clamped = Math.max(0, Math.min(quantity, item.available));
    const withoutItem = selected.filter((s) => s.itemId !== item.id);
    onChange(clamped > 0 ? [...withoutItem, { itemId: item.id, name: item.name, quantity: clamped }] : withoutItem);
  };

  return (
    <div className="request-step">
      <h2 className="request-step__title">Select Equipment</h2>
      <p className="request-step__subtitle">
        Showing available inventory for {buildingName ? <strong>{buildingName}</strong> : 'the selected building'}.
      </p>

      <div className="equipment-grid">
        {items.map((item) => {
          const quantity = quantityFor(item.id);
          return (
            <div key={item.id} className="equipment-card">
              <div className="equipment-card__info">
                <span className="equipment-card__name">{item.name}</span>
                <span className="equipment-card__available">Available: {item.available}</span>
              </div>
              <div className="equipment-card__quantity">
                <button
                  type="button"
                  className="quantity-btn"
                  onClick={() => setQuantity(item, quantity - 1)}
                  disabled={quantity <= 0}
                  aria-label={`Decrease ${item.name} quantity`}
                >
                  −
                </button>
                <span className="quantity-value">{quantity}</span>
                <button
                  type="button"
                  className="quantity-btn"
                  onClick={() => setQuantity(item, quantity + 1)}
                  disabled={quantity >= item.available}
                  aria-label={`Increase ${item.name} quantity`}
                >
                  +
                </button>
              </div>
            </div>
          );
        })}
      </div>

      <div className="request-step__actions">
        <button type="button" className="btn btn--secondary" onClick={onBack}>
          Back
        </button>
        <button type="button" className="btn btn--primary" onClick={onContinue}>
          Continue
        </button>
      </div>
    </div>
  );
};

export default EquipmentStep;
