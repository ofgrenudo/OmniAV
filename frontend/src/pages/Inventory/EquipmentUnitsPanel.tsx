import React, { useEffect, useState } from 'react';
import { Building } from '../../types/Building';
import { Equipment } from '../../types/Equipment';
import {
  EquipmentInput,
  archiveEquipment,
  createEquipment,
  listEquipment,
  updateEquipment,
} from '../../services/equipmentApi';
import { listBuildings } from '../../services/buildingsApi';
import { errorMessage } from '../../utils/apiError';

interface EquipmentUnitsPanelProps {
  groupId: number;
  groupName: string;
}

interface UnitFormValues {
  name: string;
  description: string;
  disabled: boolean;
  archived: boolean;
  buildingId: string;
}

const emptyFormValues: UnitFormValues = {
  name: '',
  description: '',
  disabled: false,
  archived: false,
  buildingId: '',
};

const toFormValues = (unit: Equipment): UnitFormValues => ({
  name: unit.name,
  description: unit.description ?? '',
  disabled: unit.disabled,
  archived: unit.archived,
  buildingId: unit.buildingId !== null ? String(unit.buildingId) : '',
});

const toInput = (groupId: number, values: UnitFormValues): EquipmentInput => ({
  name: values.name.trim(),
  description: values.description.trim() ? values.description.trim() : null,
  disabled: values.disabled,
  archived: values.archived,
  groupId,
  buildingId: values.buildingId ? Number(values.buildingId) : null,
});

const unitStatus = (unit: Equipment): { label: string; badgeClass: string } => {
  if (unit.archived) return { label: 'Archived', badgeClass: 'badge--archived' };
  if (unit.disabled) return { label: 'Disabled', badgeClass: 'badge--disabled' };
  return { label: 'Active', badgeClass: 'badge--active' };
};

const EquipmentUnitsPanel: React.FC<EquipmentUnitsPanelProps> = ({ groupId, groupName }) => {
  const [units, setUnits] = useState<Equipment[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [refreshToken, setRefreshToken] = useState(0);

  const [buildings, setBuildings] = useState<Building[]>([]);

  const [editingId, setEditingId] = useState<number | 'new' | null>(null);
  const [formValues, setFormValues] = useState<UnitFormValues>(emptyFormValues);
  const [formError, setFormError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    let cancelled = false;
    listBuildings({ pageSize: 100, sort: 'name', order: 'asc', archived: false })
      .then((result) => {
        if (!cancelled) setBuildings(result.data);
      })
      .catch(() => undefined);
    return () => {
      cancelled = true;
    };
  }, []);

  const buildingName = (buildingId: number | null): string => {
    if (buildingId === null) return '—';
    return buildings.find((b) => b.id === buildingId)?.name ?? `#${buildingId}`;
  };

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);

    listEquipment({ groupId, pageSize: 100, sort: 'name', order: 'asc' })
      .then((result) => {
        if (cancelled) return;
        setUnits(result.data);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(errorMessage(err, 'Failed to load units.'));
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [groupId, refreshToken]);

  const startCreate = () => {
    setEditingId('new');
    setFormValues(emptyFormValues);
    setFormError(null);
  };

  const startEdit = (unit: Equipment) => {
    setEditingId(unit.id);
    setFormValues(toFormValues(unit));
    setFormError(null);
  };

  const cancelForm = () => {
    setEditingId(null);
    setFormError(null);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!formValues.name.trim()) {
      setFormError('Name is required.');
      return;
    }

    setSaving(true);
    setFormError(null);
    try {
      if (editingId === 'new') {
        await createEquipment(toInput(groupId, formValues));
      } else if (editingId !== null) {
        await updateEquipment(editingId, toInput(groupId, formValues));
      }
      setEditingId(null);
      setRefreshToken((n) => n + 1);
    } catch (err) {
      setFormError(errorMessage(err, 'Failed to save unit.'));
    } finally {
      setSaving(false);
    }
  };

  const handleArchive = async (unit: Equipment) => {
    try {
      if (unit.archived) {
        await updateEquipment(unit.id, { ...toInput(groupId, toFormValues(unit)), archived: false });
      } else {
        await archiveEquipment(unit.id);
      }
      setRefreshToken((n) => n + 1);
    } catch (err) {
      setError(errorMessage(err, 'Failed to update unit.'));
    }
  };

  const handleToggleDisabled = async (unit: Equipment) => {
    try {
      await updateEquipment(unit.id, { ...toInput(groupId, toFormValues(unit)), disabled: !unit.disabled });
      setRefreshToken((n) => n + 1);
    } catch (err) {
      setError(errorMessage(err, 'Failed to update unit.'));
    }
  };

  return (
    <div className="units-panel">
      <div className="units-panel__header">
        <h4 className="units-panel__title">Units in {groupName}</h4>
        <button type="button" className="btn btn--secondary btn--small" onClick={startCreate}>
          Add Unit
        </button>
      </div>

      {editingId !== null && (
        <form className="units-panel__form" onSubmit={handleSubmit}>
          <div className="form-field">
            <label htmlFor={`unit-name-${groupId}`}>Name *</label>
            <input
              id={`unit-name-${groupId}`}
              type="text"
              value={formValues.name}
              onChange={(e) => setFormValues({ ...formValues, name: e.target.value })}
              placeholder={`e.g. ${groupName} A`}
            />
          </div>
          <div className="form-field">
            <label htmlFor={`unit-description-${groupId}`}>Description</label>
            <input
              id={`unit-description-${groupId}`}
              type="text"
              value={formValues.description}
              onChange={(e) => setFormValues({ ...formValues, description: e.target.value })}
            />
          </div>
          <div className="form-field">
            <label htmlFor={`unit-building-${groupId}`}>Building</label>
            <select
              id={`unit-building-${groupId}`}
              value={formValues.buildingId}
              onChange={(e) => setFormValues({ ...formValues, buildingId: e.target.value })}
            >
              <option value="">Unassigned</option>
              {buildings.map((building) => (
                <option key={building.id} value={building.id}>
                  {building.name}
                </option>
              ))}
            </select>
          </div>
          {editingId !== 'new' && (
            <div className="units-panel__checkboxes">
              <label>
                <input
                  type="checkbox"
                  checked={formValues.disabled}
                  onChange={(e) => setFormValues({ ...formValues, disabled: e.target.checked })}
                />{' '}
                Disabled
              </label>
              <label>
                <input
                  type="checkbox"
                  checked={formValues.archived}
                  onChange={(e) => setFormValues({ ...formValues, archived: e.target.checked })}
                />{' '}
                Archived
              </label>
            </div>
          )}
          {formError !== null && <p className="form-field__error">{formError}</p>}
          <div className="admin-page__panel-actions">
            <button type="button" className="btn btn--secondary btn--small" onClick={cancelForm} disabled={saving}>
              Cancel
            </button>
            <button type="submit" className="btn btn--primary btn--small" disabled={saving}>
              {saving ? 'Saving…' : 'Save'}
            </button>
          </div>
        </form>
      )}

      {error !== null && <p className="admin-page__status admin-page__status--error">{error}</p>}
      {error === null && loading && <p className="admin-page__status">Loading units…</p>}
      {error === null && !loading && units.length === 0 && (
        <p className="admin-page__status">No units yet in this group.</p>
      )}

      {error === null && !loading && units.length > 0 && (
        <table className="data-table">
          <thead>
            <tr>
              <th>Name</th>
              <th className="data-table__cell--wrap">Description</th>
              <th>Building</th>
              <th>Status</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {units.map((unit) => {
              const status = unitStatus(unit);
              return (
                <tr key={unit.id}>
                  <td>{unit.name}</td>
                  <td className="data-table__cell--wrap">{unit.description ?? '—'}</td>
                  <td>{buildingName(unit.buildingId)}</td>
                  <td>
                    <span className={`badge ${status.badgeClass}`}>{status.label}</span>
                  </td>
                  <td className="data-table__actions">
                    <button type="button" className="btn btn--secondary btn--small" onClick={() => startEdit(unit)}>
                      Edit
                    </button>
                    <button
                      type="button"
                      className="btn btn--secondary btn--small"
                      onClick={() => handleToggleDisabled(unit)}
                    >
                      {unit.disabled ? 'Enable' : 'Disable'}
                    </button>
                    <button
                      type="button"
                      className="btn btn--secondary btn--small"
                      onClick={() => handleArchive(unit)}
                    >
                      {unit.archived ? 'Restore' : 'Archive'}
                    </button>
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      )}
    </div>
  );
};

export default EquipmentUnitsPanel;
