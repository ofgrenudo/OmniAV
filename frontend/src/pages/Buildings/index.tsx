import React, { useEffect, useState } from 'react';
import { Building } from '../../types/Building';
import {
  BuildingInput,
  BuildingSortColumn,
  PaginationMeta,
  archiveBuilding,
  createBuilding,
  listBuildings,
  updateBuilding,
} from '../../services/buildingsApi';
import { errorMessage } from '../../utils/apiError';

type ArchivedFilter = 'active' | 'archived' | 'all';

interface SortState {
  column: BuildingSortColumn;
  order: 'asc' | 'desc';
}

interface BuildingFormValues {
  name: string;
  address: string;
  description: string;
  archived: boolean;
}

const PAGE_SIZE_OPTIONS = [10, 20, 50];

const emptyMeta: PaginationMeta = { page: 1, pageSize: 10, totalItems: 0, totalPages: 1 };
const emptyFormValues: BuildingFormValues = { name: '', address: '', description: '', archived: false };

const toFormValues = (building: Building): BuildingFormValues => ({
  name: building.name,
  address: building.address ?? '',
  description: building.description ?? '',
  archived: building.archived,
});

const toInput = (values: BuildingFormValues): BuildingInput => ({
  name: values.name.trim(),
  archived: values.archived,
  address: values.address.trim() ? values.address.trim() : null,
  description: values.description.trim() ? values.description.trim() : null,
});

const Buildings: React.FC = () => {
  const [buildings, setBuildings] = useState<Building[]>([]);
  const [meta, setMeta] = useState<PaginationMeta>(emptyMeta);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [archivedFilter, setArchivedFilter] = useState<ArchivedFilter>('active');
  const [sort, setSort] = useState<SortState>({ column: 'name', order: 'asc' });
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [refreshToken, setRefreshToken] = useState(0);

  const [editingId, setEditingId] = useState<number | 'new' | null>(null);
  const [formValues, setFormValues] = useState<BuildingFormValues>(emptyFormValues);
  const [formError, setFormError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    const timeout = setTimeout(() => setDebouncedSearch(search.trim()), 300);
    return () => clearTimeout(timeout);
  }, [search]);

  useEffect(() => {
    setPage(1);
  }, [debouncedSearch, archivedFilter, sort, pageSize]);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);

    listBuildings({
      page,
      pageSize,
      sort: sort.column,
      order: sort.order,
      q: debouncedSearch || undefined,
      archived: archivedFilter === 'all' ? undefined : archivedFilter === 'archived',
    })
      .then((result) => {
        if (cancelled) return;
        setBuildings(result.data);
        setMeta(result.meta);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(errorMessage(err, 'Failed to load buildings.'));
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [page, pageSize, sort, debouncedSearch, archivedFilter, refreshToken]);

  const toggleSort = (column: BuildingSortColumn) => {
    setSort((prev) =>
      prev.column === column ? { column, order: prev.order === 'asc' ? 'desc' : 'asc' } : { column, order: 'asc' }
    );
  };

  const startCreate = () => {
    setEditingId('new');
    setFormValues(emptyFormValues);
    setFormError(null);
  };

  const startEdit = (building: Building) => {
    setEditingId(building.id);
    setFormValues(toFormValues(building));
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
        await createBuilding(toInput(formValues));
      } else if (editingId !== null) {
        await updateBuilding(editingId, toInput(formValues));
      }
      setEditingId(null);
      setRefreshToken((n) => n + 1);
    } catch (err) {
      setFormError(errorMessage(err, 'Failed to save building.'));
    } finally {
      setSaving(false);
    }
  };

  const handleArchive = async (building: Building) => {
    try {
      if (building.archived) {
        await updateBuilding(building.id, {
          name: building.name,
          address: building.address,
          description: building.description,
          archived: false,
        });
      } else {
        await archiveBuilding(building.id);
      }
      setRefreshToken((n) => n + 1);
    } catch (err) {
      setError(errorMessage(err, 'Failed to update building.'));
    }
  };

  const sortIndicator = (column: BuildingSortColumn) => {
    if (sort.column !== column) return '';
    return sort.order === 'asc' ? ' ▲' : ' ▼';
  };

  return (
    <div className="admin-page">
      <div className="admin-page__header">
        <div>
          <h1>Buildings</h1>
          <p>Manage campus buildings and rooms.</p>
        </div>
        <button type="button" className="btn btn--primary" onClick={startCreate}>
          Add Building
        </button>
      </div>

      {editingId !== null && (
        <div className="admin-page__panel">
          <h2 className="admin-page__panel-title">{editingId === 'new' ? 'Add Building' : 'Edit Building'}</h2>
          <form onSubmit={handleSubmit}>
            <div className="form-field">
              <label htmlFor="building-name">Name *</label>
              <input
                id="building-name"
                type="text"
                value={formValues.name}
                onChange={(e) => setFormValues({ ...formValues, name: e.target.value })}
              />
            </div>
            <div className="form-field">
              <label htmlFor="building-address">Address</label>
              <input
                id="building-address"
                type="text"
                value={formValues.address}
                onChange={(e) => setFormValues({ ...formValues, address: e.target.value })}
              />
            </div>
            <div className="form-field">
              <label htmlFor="building-description">Description</label>
              <textarea
                id="building-description"
                rows={2}
                value={formValues.description}
                onChange={(e) => setFormValues({ ...formValues, description: e.target.value })}
              />
            </div>
            {editingId !== 'new' && (
              <div className="form-field">
                <label htmlFor="building-archived">
                  <input
                    id="building-archived"
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
              <button type="button" className="btn btn--secondary" onClick={cancelForm} disabled={saving}>
                Cancel
              </button>
              <button type="submit" className="btn btn--primary" disabled={saving}>
                {saving ? 'Saving…' : 'Save'}
              </button>
            </div>
          </form>
        </div>
      )}

      <div className="admin-page__toolbar">
        <div className="form-field admin-page__search">
          <label htmlFor="building-search">Search</label>
          <input
            id="building-search"
            type="text"
            placeholder="Search by name, address, or description"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <div className="form-field">
          <label htmlFor="building-archived-filter">Status</label>
          <select
            id="building-archived-filter"
            value={archivedFilter}
            onChange={(e) => setArchivedFilter(e.target.value as ArchivedFilter)}
          >
            <option value="active">Active</option>
            <option value="archived">Archived</option>
            <option value="all">All</option>
          </select>
        </div>
        <div className="form-field">
          <label htmlFor="building-page-size">Per page</label>
          <select id="building-page-size" value={pageSize} onChange={(e) => setPageSize(Number(e.target.value))}>
            {PAGE_SIZE_OPTIONS.map((size) => (
              <option key={size} value={size}>
                {size}
              </option>
            ))}
          </select>
        </div>
      </div>

      {error !== null && <p className="admin-page__status admin-page__status--error">{error}</p>}

      {error === null && loading && <p className="admin-page__status">Loading buildings…</p>}

      {error === null && !loading && buildings.length === 0 && (
        <p className="admin-page__status">No buildings match your filters.</p>
      )}

      {error === null && !loading && buildings.length > 0 && (
        <>
          <div className="data-table-wrap">
            <table className="data-table">
              <thead>
                <tr>
                  <th className="data-table__sortable" onClick={() => toggleSort('name')}>
                    Name{sortIndicator('name')}
                  </th>
                  <th className="data-table__cell--wrap">Address</th>
                  <th className="data-table__cell--wrap">Description</th>
                  <th>Status</th>
                  <th className="data-table__sortable" onClick={() => toggleSort('updated_at')}>
                    Updated{sortIndicator('updated_at')}
                  </th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {buildings.map((building) => (
                  <tr key={building.id}>
                    <td>{building.name}</td>
                    <td className="data-table__cell--wrap">{building.address ?? '—'}</td>
                    <td className="data-table__cell--wrap">{building.description ?? '—'}</td>
                    <td>
                      <span className={`badge ${building.archived ? 'badge--archived' : 'badge--active'}`}>
                        {building.archived ? 'Archived' : 'Active'}
                      </span>
                    </td>
                    <td>{new Date(building.updatedAt).toLocaleDateString()}</td>
                    <td className="data-table__actions">
                      <button
                        type="button"
                        className="btn btn--secondary btn--small"
                        onClick={() => startEdit(building)}
                      >
                        Edit
                      </button>
                      <button
                        type="button"
                        className="btn btn--secondary btn--small"
                        onClick={() => handleArchive(building)}
                      >
                        {building.archived ? 'Restore' : 'Archive'}
                      </button>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>

          <div className="admin-page__pagination">
            <span>
              Page {meta.page} of {Math.max(meta.totalPages, 1)} ({meta.totalItems} total)
            </span>
            <div className="admin-page__pagination-controls">
              <button
                type="button"
                className="btn btn--secondary btn--small"
                onClick={() => setPage((p) => Math.max(1, p - 1))}
                disabled={meta.page <= 1}
              >
                Previous
              </button>
              <button
                type="button"
                className="btn btn--secondary btn--small"
                onClick={() => setPage((p) => Math.min(meta.totalPages, p + 1))}
                disabled={meta.page >= meta.totalPages}
              >
                Next
              </button>
            </div>
          </div>
        </>
      )}
    </div>
  );
};

export default Buildings;
