import React, { useEffect, useState } from 'react';
import { Request } from '../../types/Request';
import { deleteRequest, getRequest, listRequests, unassignEquipment } from '../../services/requestsApi';
import { errorMessage } from '../../utils/apiError';
import { to12Hour } from '../../utils/time';

const PAGE_SIZE_OPTIONS = [10, 20, 50];

const emptyMeta = { page: 1, pageSize: 10, totalItems: 0, totalPages: 1 };

const dateOnly = (iso: string): string => iso.slice(0, 10);

const timeOfDay = (iso: string): string => {
  const match = iso.match(/T(\d{2}):(\d{2})/);
  return match ? to12Hour(`${match[1]}:${match[2]}`) : iso;
};

const AllRequests: React.FC = () => {
  const [requests, setRequests] = useState<Request[]>([]);
  const [meta, setMeta] = useState(emptyMeta);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const [search, setSearch] = useState('');
  const [debouncedSearch, setDebouncedSearch] = useState('');
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(10);
  const [refreshToken, setRefreshToken] = useState(0);

  const [expandedId, setExpandedId] = useState<number | null>(null);
  const [expandedRequest, setExpandedRequest] = useState<Request | null>(null);
  const [expandedLoading, setExpandedLoading] = useState(false);
  const [expandedError, setExpandedError] = useState<string | null>(null);

  const [confirmingCancelId, setConfirmingCancelId] = useState<number | null>(null);

  useEffect(() => {
    const timeout = setTimeout(() => setDebouncedSearch(search.trim()), 300);
    return () => clearTimeout(timeout);
  }, [search]);

  useEffect(() => {
    setPage(1);
  }, [debouncedSearch, pageSize]);

  useEffect(() => {
    let cancelled = false;
    setLoading(true);
    setError(null);

    listRequests({
      page,
      pageSize,
      sort: 'first_date_needed',
      order: 'desc',
      q: debouncedSearch || undefined,
    })
      .then((result) => {
        if (cancelled) return;
        setRequests(result.data);
        setMeta(result.meta);
      })
      .catch((err) => {
        if (cancelled) return;
        setError(errorMessage(err, 'Failed to load requests.'));
      })
      .finally(() => {
        if (!cancelled) setLoading(false);
      });

    return () => {
      cancelled = true;
    };
  }, [page, pageSize, debouncedSearch, refreshToken]);

  const loadExpanded = (id: number) => {
    setExpandedLoading(true);
    setExpandedError(null);
    getRequest(id)
      .then(setExpandedRequest)
      .catch((err) => setExpandedError(errorMessage(err, 'Failed to load equipment for this request.')))
      .finally(() => setExpandedLoading(false));
  };

  const toggleExpanded = (id: number) => {
    if (expandedId === id) {
      setExpandedId(null);
      setExpandedRequest(null);
      return;
    }
    setExpandedId(id);
    setExpandedRequest(null);
    loadExpanded(id);
  };

  const handleUnassign = async (requestId: number, requestedEquipmentId: number) => {
    try {
      await unassignEquipment(requestId, requestedEquipmentId);
      loadExpanded(requestId);
    } catch (err) {
      setExpandedError(errorMessage(err, 'Failed to remove that assignment.'));
    }
  };

  const handleCancel = async (id: number) => {
    try {
      await deleteRequest(id);
      setConfirmingCancelId(null);
      if (expandedId === id) {
        setExpandedId(null);
        setExpandedRequest(null);
      }
      setRefreshToken((n) => n + 1);
    } catch (err) {
      setError(errorMessage(err, 'Failed to cancel that request.'));
    }
  };

  return (
    <div className="admin-page">
      <div className="admin-page__header">
        <div>
          <h1>All Requests</h1>
          <p>Every request across the college, for technicians and admins.</p>
        </div>
      </div>

      <div className="admin-page__toolbar">
        <div className="form-field admin-page__search">
          <label htmlFor="request-search">Search</label>
          <input
            id="request-search"
            type="text"
            placeholder="Search by request name or room"
            value={search}
            onChange={(e) => setSearch(e.target.value)}
          />
        </div>
        <div className="form-field">
          <label htmlFor="request-page-size">Per page</label>
          <select id="request-page-size" value={pageSize} onChange={(e) => setPageSize(Number(e.target.value))}>
            {PAGE_SIZE_OPTIONS.map((size) => (
              <option key={size} value={size}>
                {size}
              </option>
            ))}
          </select>
        </div>
      </div>

      {error !== null && <p className="admin-page__status admin-page__status--error">{error}</p>}
      {error === null && loading && <p className="admin-page__status">Loading requests…</p>}
      {error === null && !loading && requests.length === 0 && (
        <p className="admin-page__status">No requests match your filters.</p>
      )}

      {error === null && !loading && requests.length > 0 && (
        <>
          <div className="data-table-wrap">
            <table className="data-table">
              <thead>
                <tr>
                  <th>Name</th>
                  <th>Building</th>
                  <th>Room</th>
                  <th>Date</th>
                  <th>Time</th>
                  <th>Weeks</th>
                  <th>Actions</th>
                </tr>
              </thead>
              <tbody>
                {requests.map((request) => {
                  const isExpanded = expandedId === request.id;
                  const isConfirming = confirmingCancelId === request.id;
                  return (
                    <React.Fragment key={request.id}>
                      <tr>
                        <td>{request.name}</td>
                        <td>{request.building?.name ?? `#${request.buildingId}`}</td>
                        <td>{request.room}</td>
                        <td>
                          {dateOnly(request.firstDateNeeded)} ({request.daysOfWeek})
                        </td>
                        <td>
                          {timeOfDay(request.startTime)} – {timeOfDay(request.endTime)}
                        </td>
                        <td>{request.numberOfWeeks}</td>
                        <td className="data-table__actions">
                          <button
                            type="button"
                            className="btn btn--secondary btn--small"
                            onClick={() => toggleExpanded(request.id)}
                          >
                            {isExpanded ? 'Hide Equipment' : 'Equipment'}
                          </button>
                          {isConfirming ? (
                            <>
                              <button
                                type="button"
                                className="btn btn--secondary btn--small"
                                onClick={() => handleCancel(request.id)}
                              >
                                Confirm Cancel
                              </button>
                              <button
                                type="button"
                                className="btn btn--secondary btn--small"
                                onClick={() => setConfirmingCancelId(null)}
                              >
                                Never Mind
                              </button>
                            </>
                          ) : (
                            <button
                              type="button"
                              className="btn btn--secondary btn--small"
                              onClick={() => setConfirmingCancelId(request.id)}
                            >
                              Cancel Request
                            </button>
                          )}
                        </td>
                      </tr>
                      {isExpanded && (
                        <tr>
                          <td colSpan={7} className="units-panel-cell">
                            {expandedError !== null && (
                              <p className="admin-page__status admin-page__status--error">{expandedError}</p>
                            )}
                            {expandedError === null && expandedLoading && (
                              <p className="admin-page__status">Loading equipment…</p>
                            )}
                            {expandedError === null &&
                              !expandedLoading &&
                              expandedRequest &&
                              (expandedRequest.requestedEquipment?.length ? (
                                <table className="data-table">
                                  <thead>
                                    <tr>
                                      <th>Group</th>
                                      <th>Unit</th>
                                      <th>Actions</th>
                                    </tr>
                                  </thead>
                                  <tbody>
                                    {expandedRequest.requestedEquipment.map((entry) => (
                                      <tr key={entry.id}>
                                        <td>{entry.group?.name ?? `#${entry.groupId}`}</td>
                                        <td>{entry.equipment?.name ?? `#${entry.equipmentId}`}</td>
                                        <td>
                                          <button
                                            type="button"
                                            className="btn btn--secondary btn--small"
                                            onClick={() => handleUnassign(request.id, entry.id)}
                                          >
                                            Remove
                                          </button>
                                        </td>
                                      </tr>
                                    ))}
                                  </tbody>
                                </table>
                              ) : (
                                <p className="admin-page__status">No equipment assigned to this request.</p>
                              ))}
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

export default AllRequests;
