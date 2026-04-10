package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"

	"github.com/prasenjit-net/openid-golang/pkg/models"
)

// ─── Helper used by all handlers ─────────────────────────────────────────────

// logAudit creates and persists an AuditLog entry. Errors are silently dropped
// so that an audit-write failure never interrupts the primary request flow.
func (h *Handlers) logAudit(
	action models.AuditAction,
	actorType models.AuditActorType,
	actor string,
	resource, resourceID string,
	status models.AuditStatus,
	ip, ua string,
	details map[string]interface{},
) {
	entry := &models.AuditLog{
		ID:         uuid.NewString(),
		Timestamp:  time.Now().UTC(),
		Action:     action,
		Actor:      actor,
		ActorType:  actorType,
		Resource:   resource,
		ResourceID: resourceID,
		IPAddress:  ip,
		UserAgent:  ua,
		Status:     status,
		Details:    details,
	}
	_ = h.storage.CreateAuditLog(entry)
}

// logAdminAudit is the same but is called from AdminHandler which holds a
// different storage reference (h.store vs h.storage).
func (h *AdminHandler) logAdminAudit(
	action models.AuditAction,
	actorType models.AuditActorType,
	actor string,
	resource, resourceID string,
	status models.AuditStatus,
	ip, ua string,
	details map[string]interface{},
) {
	entry := &models.AuditLog{
		ID:         uuid.NewString(),
		Timestamp:  time.Now().UTC(),
		Action:     action,
		Actor:      actor,
		ActorType:  actorType,
		Resource:   resource,
		ResourceID: resourceID,
		IPAddress:  ip,
		UserAgent:  ua,
		Status:     status,
		Details:    details,
	}
	_ = h.store.CreateAuditLog(entry)
}

// ─── Admin API: GET /api/admin/audit ─────────────────────────────────────────

// GetAuditLogs returns a paginated, optionally filtered list of audit log entries.
// Query params:
//
//	limit  (int, default 50, max 200)
//	offset (int, default 0)
//	action (AuditAction string, optional)
//	actor  (username/client_id, optional)
func (h *AdminHandler) GetAuditLogs(c echo.Context) error {
	limit := 50
	if l := c.QueryParam("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	if limit > 200 {
		limit = 200
	}

	offset := 0
	if o := c.QueryParam("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}

	filter := models.AuditFilter{
		Action: models.AuditAction(c.QueryParam("action")),
		Actor:  c.QueryParam("actor"),
		Limit:  limit,
		Offset: offset,
	}

	entries, err := h.store.GetAuditLogs(filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to retrieve audit logs: " + err.Error(),
		})
	}

	// Ensure nil slice serialises as [] not null
	if entries == nil {
		entries = []*models.AuditLog{}
	}

	total := h.store.GetAuditLogsCount(models.AuditFilter{
		Action: filter.Action,
		Actor:  filter.Actor,
	})

	return c.JSON(http.StatusOK, map[string]interface{}{
		"entries": entries,
		"total":   total,
		"limit":   limit,
		"offset":  offset,
	})
}
