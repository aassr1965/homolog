package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"homolog/internal/audit"
	"homolog/internal/domain"
)

// ANPPDraftHandler gère la sauvegarde et la réédition en mode brouillon des formulaires complémentaires ANPP.
type ANPPDraftHandler struct {
	pool     *pgxpool.Pool
	auditSvc *audit.Service
}

// NewANPPDraftHandler initialise le handler des brouillons ANPP.
func NewANPPDraftHandler(pool *pgxpool.Pool, auditSvc *audit.Service) *ANPPDraftHandler {
	return &ANPPDraftHandler{pool: pool, auditSvc: auditSvc}
}

func normalizeDraftKind(kind string) (string, bool) {
	k := strings.ToLower(strings.TrimSpace(kind))
	switch k {
	case "structure_prix", "ca_sr_secondaire", "dvtiv_qppv", "dce_sa_complement":
		return k, true
	default:
		return "", false
	}
}

func resolveDraftEntityID(r *http.Request, orgID uuid.UUID) uuid.UUID {
	if raw := strings.TrimSpace(r.URL.Query().Get("entity_id")); raw != "" {
		if parsed, err := uuid.Parse(raw); err == nil {
			return parsed
		}
	}
	return orgID
}

// GetDraft récupère un brouillon éditable existant (ou un objet vide au statut 'brouillon').
func (h *ANPPDraftHandler) GetDraft(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}
	kind, ok := normalizeDraftKind(chi.URLParam(r, "kind"))
	if !ok {
		writeJSONError(w, http.StatusBadRequest, "Type de brouillon inconnu")
		return
	}
	entityID := resolveDraftEntityID(r, user.OrganizationID)

	var d domain.ANPPFormDraft
	err := h.pool.QueryRow(r.Context(), `
		SELECT id, organization_id, form_kind, entity_id, payload, status, updated_by, created_at, updated_at
		FROM anpp_form_drafts
		WHERE organization_id = $1 AND form_kind = $2 AND entity_id = $3
	`, user.OrganizationID, kind, entityID).Scan(
		&d.ID, &d.OrganizationID, &d.FormKind, &d.EntityID, &d.Payload, &d.Status, &d.UpdatedBy, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			writeJSON(w, http.StatusOK, map[string]any{
				"form_kind": kind,
				"entity_id": entityID,
				"status":    "brouillon",
				"payload":   map[string]any{},
			})
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Erreur lecture brouillon ANPP")
		return
	}

	writeJSON(w, http.StatusOK, d)
}

// SaveDraft enregistre ou met à jour un brouillon ANPP en conservant toujours le statut éditable 'brouillon'.
func (h *ANPPDraftHandler) SaveDraft(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}
	kind, ok := normalizeDraftKind(chi.URLParam(r, "kind"))
	if !ok {
		writeJSONError(w, http.StatusBadRequest, "Type de brouillon inconnu")
		return
	}
	entityID := resolveDraftEntityID(r, user.OrganizationID)

	var rawPayload json.RawMessage
	if err := json.NewDecoder(r.Body).Decode(&rawPayload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}
	if len(rawPayload) == 0 {
		rawPayload = json.RawMessage(`{}`)
	}

	var d domain.ANPPFormDraft
	err := h.pool.QueryRow(r.Context(), `
		INSERT INTO anpp_form_drafts (organization_id, form_kind, entity_id, payload, status, updated_by)
		VALUES ($1, $2, $3, $4, 'brouillon', $5)
		ON CONFLICT (organization_id, form_kind, entity_id) DO UPDATE SET
			payload = EXCLUDED.payload,
			status = 'brouillon',
			updated_by = EXCLUDED.updated_by,
			updated_at = NOW()
		RETURNING id, organization_id, form_kind, entity_id, payload, status, updated_by, created_at, updated_at
	`, user.OrganizationID, kind, entityID, rawPayload, user.UserID).Scan(
		&d.ID, &d.OrganizationID, &d.FormKind, &d.EntityID, &d.Payload, &d.Status, &d.UpdatedBy, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur sauvegarde brouillon ANPP")
		return
	}

	if h.auditSvc != nil {
		_, _ = h.auditSvc.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "ANPP_FORM_DRAFT_SAVED",
			EntityType:     "anpp_form_draft",
			EntityID:       kind + ":" + entityID.String(),
			Details: map[string]any{
				"draft_id":  d.ID.String(),
				"form_kind": kind,
				"entity_id": entityID.String(),
				"status":    "brouillon",
			},
			IPAddress: getClientIP(r),
			UserAgent: r.UserAgent(),
		})
	}

	writeJSON(w, http.StatusOK, d)
}

func loadANPPDraft(ctx context.Context, pool *pgxpool.Pool, orgID uuid.UUID, kind string, entityID uuid.UUID, out any) bool {
	if pool == nil || out == nil {
		return false
	}
	var payload []byte
	err := pool.QueryRow(ctx, `
		SELECT payload FROM anpp_form_drafts
		WHERE organization_id = $1 AND form_kind = $2 AND entity_id = $3
	`, orgID, kind, entityID).Scan(&payload)
	if err != nil || len(payload) == 0 {
		return false
	}
	return json.Unmarshal(payload, out) == nil
}
