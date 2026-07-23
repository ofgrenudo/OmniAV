import React, { useEffect, useState } from 'react';
import { EquipmentGroup } from '../../types/EquipmentGroup';
import {
  EquipmentGroupInput,
  EquipmentGroupSortColumn,
  PaginationMeta,
  archiveEquipmentGroup,
  createEquipmentGroup,
  listEquipmentGroups,
  updateEquipmentGroup,
} from '../../services/equipmentGroupsApi';
import { errorMessage } from '../../utils/apiError';
import EquipmentUnitsPanel from './EquipmentUnitsPanel';
import './Inventory.css';

type ArchivedFilter = 'active' | 'archived' | 'all';

interface SortState {
  column: EquipmentGroupSortColumn;
  order: 'asc' | 'desc';
}

interface GroupFormValues {
  name: string;
  description: string;
  disabled: boolean;
  archived: boolean;
}

const PAGE_SIZE_OPTIONS = [10, 20, 50];

const emptyMeta: PaginationMeta = { page: 1, pageSize: 10, totalItems: 0, totalPages: 1 };
const emptyFormValues: GroupFormValues = { name: '', description: '', disabled: false, archived: false };

const toFormValues = (group: EquipmentGroup): GroupFormValues => ({
  name: group.name,
  description: group.description ?? '',
  disabled: group.disabled,
  archived: group.archived,
});

const toInput = (values: GroupFormValues): EquipmentGroupInput => ({
  name: values.name.trim(),
  description: values.description.trim() ? values.description.trim() : null,
  disabled: values.disabled,
  archived: values.archived,
});

const groupStatus = (group: EquipmentGroup): { label: string; badgeClass: string } => {
  if (group.archived) return { label: 'Archived', badgeClass: 'badge--archived' };
  if (group.disabled) return { label: 'Disabled', badgeClass: 'badge--disabled' };
  return { label: 'Active', badgeClass: 'badge--active' };
};

const Inventory: React.FC = () => {
  const [groups, setGroups] = useState<EquipmentGroup[]>([]);
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
  const [formValues, setFormValues] = useState<GroupFormValues>(emptyFormValues);
  const [formError, setFormError] = useState<string | null>(null);
  const [saving, setSaving] = useState(false);

  const [expandedId, setExpandedId] = useState<number | null>(null);

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

    listEquipmentGroups({
      page,
      pageSize,
      sort: sort.column,
      order: sort.order,
      q: debouncedSearch || undefined,
      archived: archivedFilter === 'all' ? undefined : archivedFilter === 'archived',
    })
      .then((result) => {
        if (cancelled) return;
        setGroups(result.data);
        setMeta(result.meta);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(errorMessage(err, 'Failed to load equipment groups.'));
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [page, pageSize, sort, debouncedSearch, archivedFilter, refreshToken]);

  const toggleSort = (column: EquipmentGroupSortColumn) => {
    setSort((prev) =>
      prev.column === column ? { column, order: prev.order === 'asc' ? 'desc' : 'asc' } : { column, order: 'asc' }
    );
  };

  const startCreate = () => {
    setEditingId('new');
    setFormValues(emptyFormValues);
    setFormError(null);
  };

  const startEdit = (group: EquipmentGroup) => {
    setEditingId(group.id);
    setFormValues(toFormValues(group));
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
        await createEquipmentGroup(toInput(formValues));
      } else if (editingId !== null) {
        await updateEquipmentGroup(editingId, toInput(formValues));
      }
      setEditingId(null);
      setRefreshToken((n) => n + 1);
    } catch (err) {
      setFormError(errorMessage(err, 'Failed to save equipment group.'));
    } finally {
      setSaving(false);
    }
  };

  const handleArchive = async (group: EquipmentGroup) => {
    try {
      if (group.archived) {
        await updateEquipmentGroup(group.id, { ...toInput(toFormValues(group)), archived: false });
      } else {
        await archiveEquipmentGroup(group.id);
      }
      setRefreshToken((n) => n + 1);
    } catch (err) {
      setError(errorMessage(err, 'Failed to update equipment group.'));
    }
  };

  const toggleExpanded = (groupId: number) => {
    setExpandedId((current) => (current === groupId ? null : groupId));
  };

  const sortIndicator = (column: EquipmentGroupSortColumn) => {
    if (sort.column !== column) return '';
    return sort.order === 'asc' ? ' ▲' : ' ▼';
  };

  return (
    <div className="admin-page">
      <div className="admin-page__header">
        <div>
          <h1>Inventory</h1>
          <p>AV equipment inventory and availability, grouped by type (e.g. &quot;Cow Cart&quot;).</p>
        </div>
        <button type="button" className="btn btn--primary" onClick={startCreate}>
          Add Equipment Group
        </button>
      </div>

      {editingId !== null && (
        <div className="admin-page__panel">
          <h2 className="admin-page__panel-title">
            {editingId === 'new' ? 'Add Equipment Group' : 'Edit Equipment Group'}
          </h2>
          <form onSubmit={handleSubmit}>
            <div className="form-field">
              <label htmlFor="group-name">Name *</label>
              <input
                id="group-name"
                type="text"
                value={formValues.name}
                onChange={(e) => setFormValues({ ...formValues, name: e.target.value })}
                placeholder="e.g. Cow Cart"
              />
            </div>
            <div className="form-field">
              <label htmlFor="group-description">Description</label>
              <textarea
                id="group-description"
                rows={2}
                value={formValues.description}
                onChange={(e) => setFormValues({ ...formValues, description: e.target.value })}
              />
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
          <label htmlFor="group-search">Search</label>
          <input
            id="group-search"
            type="text"
            placeholder="Search by name or description"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <div className="form-field">
          <label htmlFor="group-archived-filter">Status</label>
          <select
            id="group-archived-filter"
            value={archivedFilter}
            onChange={(e) => setArchivedFilter(e.target.value as ArchivedFilter)}
          >
            <option value="active">Active</option>
            <option value="archived">Archived</option>
            <option value="all">All</option>
          </select>
        </div>
        <div className="form-field">
          <label htmlFor="group-page-size">Per page</label>
          <select id="group-page-size" value={pageSize} onChange={(e) => setPageSize(Number(e.target.value))}>
            {PAGE_SIZE_OPTIONS.map((size) => (
              <option key={size} value={size}>
                {size}
              </option>
            ))}
          </select>
        </div>
      </div>

      {error !== null && <p className="admin-page__status admin-page__status--error">{error}</p>}

      {error === null && loading && <p className="admin-page__status">Loading equipment groups…</p>}

      {error === null && !loading && groups.length === 0 && (
        <p className="admin-page__status">No equipment groups match your filters.</p>
      )}

      {error === null && !loading && groups.length > 0 && (
        <>
          <div className="data-table-wrap">
            <table className="data-table">
              <thead>
                <tr>
                  <th className="data-table__sortable" onClick={() => toggleSort('name')}>
                    Name{sortIndicator('name')}
                  </th>
                  <th className="data-table__cell--wrap">Description</th>
                  <th>Status</th>
                  <th className="data-table__sortable" onClick={() => toggleSort('updated_at')}>
                    Updated{sortIndicator('updated_at')}
                  </th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {groups.map((group) => {
                  const status = groupStatus(group);
                  const isExpanded = expandedId === group.id;
                  return (
                    <React.Fragment key={group.id}>
                      <tr>
                        <td>{group.name}</td>
                        <td className="data-table__cell--wrap">{group.description ?? '—'}</td>
                        <td>
                          <span className={`badge ${status.badgeClass}`}>{status.label}</span>
                        </td>
                        <td>{new Date(group.updatedAt).toLocaleDateString()}</td>
                        <td className="data-table__actions">
                          <button
                            type="button"
                            className="btn btn--secondary btn--small"
                            onClick={() => toggleExpanded(group.id)}
                          >
                            {isExpanded ? 'Hide Units' : 'Units'}
                          </button>
                          <button
                            type="button"
                            className="btn btn--secondary btn--small"
                            onClick={() => startEdit(group)}
                          >
                            Edit
                          </button>
                          <button
                            type="button"
                            className="btn btn--secondary btn--small"
                            onClick={() => handleArchive(group)}
                          >
                            {group.archived ? 'Restore' : 'Archive'}
                          </button>
                        </td>
                      </tr>
                      {isExpanded && (
                        <tr>
                          <td colSpan={5} className="units-panel-cell">
                            <EquipmentUnitsPanel groupId={group.id} groupName={group.name} />
                          </td>
                        </tr>
                      )}
                    </React.Fragment>
                  );
                })}
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

export default Inventory;
