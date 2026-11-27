package api

import (
	"encoding/json"
	"fmt"
	"html"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"homolog/internal/audit"
	"homolog/internal/domain"
	"homolog/internal/repository"
	"homolog/internal/storage"
)

// RecallHandler gère le registre des rappels de lots et retraits PP/DM non conformes (délai 24h ANPP).
type RecallHandler struct {
	pool            *pgxpool.Pool
	auditSvc        *audit.Service
	gedRepo         *repository.GEDRepository
	storageProvider storage.StorageProvider
}

// NewRecallHandler initialise le handler des rappels de lots.
func NewRecallHandler(pool *pgxpool.Pool, auditSvc *audit.Service, gedRepo *repository.GEDRepository, storageProvider storage.StorageProvider) *RecallHandler {
	return &RecallHandler{
		pool:            pool,
		auditSvc:        auditSvc,
		gedRepo:         gedRepo,
		storageProvider: storageProvider,
	}
}

// ListRecalls retourne les rappels de lots et exercices à blanc de l'organisation.
func (h *RecallHandler) ListRecalls(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	rows, err := h.pool.Query(r.Context(), `
		SELECT id, organization_id, recall_reference, product_kind, product_name, lot_numbers,
		       severity_class, is_mock_drill, COALESCE(reason, ''), detected_at, notified_anpp_at,
		       status, ged_pv_document_id, created_by, signed_by_pdt_id, signed_at, created_at, updated_at
		FROM lot_recalls
		WHERE organization_id = $1
		ORDER BY created_at DESC
		LIMIT 100
	`, user.OrganizationID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur chargement rappels de lots")
		return
	}
	defer rows.Close()

	recalls := make([]domain.LotRecall, 0)
	for rows.Next() {
		var it domain.LotRecall
		if rows.Scan(
			&it.ID, &it.OrganizationID, &it.RecallReference, &it.ProductKind, &it.ProductName, &it.LotNumbers,
			&it.SeverityClass, &it.IsMockDrill, &it.Reason, &it.DetectedAt, &it.NotifiedANPPAt,
			&it.Status, &it.GEDPVDocumentID, &it.CreatedBy, &it.SignedByPDTID, &it.SignedAt, &it.CreatedAt, &it.UpdatedAt,
		) == nil {
			it.Within24hSLA = it.ComputeSLA24h()
			recalls = append(recalls, it)
		}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"recalls": recalls,
		"count":   len(recalls),
	})
}

// CreateRecall enregistre un nouveau rappel de lot ou un exercice annuel à blanc.
func (h *RecallHandler) CreateRecall(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	var req struct {
		ProductKind   string `json:"product_kind"`
		ProductName   string `json:"product_name"`
		LotNumbers    string `json:"lot_numbers"`
		SeverityClass string `json:"severity_class"`
		IsMockDrill   bool   `json:"is_mock_drill"`
		Reason        string `json:"reason"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}
	if strings.TrimSpace(req.ProductName) == "" || strings.TrimSpace(req.LotNumbers) == "" {
		writeJSONError(w, http.StatusBadRequest, "Le nom du produit et les numéros de lots sont obligatoires")
		return
	}

	kind := strings.ToLower(strings.TrimSpace(req.ProductKind))
	if kind != "pp" && kind != "dm" {
		kind = "pp"
	}
	sev := strings.ToUpper(strings.TrimSpace(req.SeverityClass))
	if sev != "CLASSE_I" && sev != "CLASSE_II" && sev != "CLASSE_III" {
		sev = "CLASSE_I"
	}

	id := uuid.New()
	ref := fmt.Sprintf("RCL-%s-%s", time.Now().Format("20060102"), strings.ToUpper(id.String()[:4]))
	now := time.Now().UTC()

	_, err := h.pool.Exec(r.Context(), `
		INSERT INTO lot_recalls (
			id, organization_id, recall_reference, product_kind, product_name, lot_numbers,
			severity_class, is_mock_drill, reason, detected_at, status, created_by, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,'ouvert',$11,$10,$10)
	`, id, user.OrganizationID, ref, kind, strings.TrimSpace(req.ProductName), strings.TrimSpace(req.LotNumbers),
		sev, req.IsMockDrill, strings.TrimSpace(req.Reason), now, user.UserID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur enregistrement rappel de lot")
		return
	}

	if h.auditSvc != nil {
		_, _ = h.auditSvc.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "RECALL_CREATED",
			EntityType:     "lot_recall",
			EntityID:       ref,
			Details: map[string]any{
				"recall_id":      id.String(),
				"product_kind":   kind,
				"product_name":   req.ProductName,
				"lot_numbers":    req.LotNumbers,
				"severity_class": sev,
				"is_mock_drill":  req.IsMockDrill,
			},
			IPAddress: getClientIP(r),
			UserAgent: r.UserAgent(),
		})
	}

	writeJSON(w, http.StatusCreated, map[string]any{
		"id":               id,
		"recall_reference": ref,
		"status":           "ouvert",
		"within_24h_sla":   true,
	})
}

// NotifyANPP horodate la notification officielle à l'ANPP (alerte réglementaire sous 24h).
func (h *RecallHandler) NotifyANPP(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	recallID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant de rappel invalide")
		return
	}

	var detectedAt time.Time
	var ref string
	err = h.pool.QueryRow(r.Context(), `
		UPDATE lot_recalls
		SET notified_anpp_at = COALESCE(notified_anpp_at, NOW()),
		    status = CASE WHEN status = 'ouvert' THEN 'notifie_anpp' ELSE status END,
		    updated_at = NOW()
		WHERE id = $1 AND organization_id = $2
		RETURNING detected_at, recall_reference
	`, recallID, user.OrganizationID).Scan(&detectedAt, &ref)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Rappel de lot introuvable")
		return
	}

	within24h := time.Since(detectedAt) <= 24*time.Hour
	if h.auditSvc != nil {
		_, _ = h.auditSvc.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "RECALL_NOTIFIED_ANPP",
			EntityType:     "lot_recall",
			EntityID:       ref,
			Details: map[string]any{
				"recall_id":      recallID.String(),
				"within_24h_sla": within24h,
			},
			IPAddress: getClientIP(r),
			UserAgent: r.UserAgent(),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":         "notifie_anpp",
		"within_24h_sla": within24h,
	})
}

// CloseRecall clôture le rappel de lot avec archivage du PV de retrait/destruction dans la GED (04_03_Rappels_Lots).
func (h *RecallHandler) CloseRecall(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	recallID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant de rappel invalide")
		return
	}

	var req struct {
		ResolutionNotes string `json:"resolution_notes"`
	}
	_ = json.NewDecoder(r.Body).Decode(&req)

	var it domain.LotRecall
	err = h.pool.QueryRow(r.Context(), `
		SELECT id, organization_id, recall_reference, product_kind, product_name, lot_numbers,
		       severity_class, is_mock_drill, COALESCE(reason, ''), detected_at, notified_anpp_at, status
		FROM lot_recalls
		WHERE id = $1 AND organization_id = $2
	`, recallID, user.OrganizationID).Scan(
		&it.ID, &it.OrganizationID, &it.RecallReference, &it.ProductKind, &it.ProductName, &it.LotNumbers,
		&it.SeverityClass, &it.IsMockDrill, &it.Reason, &it.DetectedAt, &it.NotifiedANPPAt, &it.Status,
	)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Rappel de lot introuvable")
		return
	}

	pvHTML := fmt.Sprintf(`<!DOCTYPE html><html lang="fr"><head><meta charset="UTF-8"><title>PV Clôture Rappel %s</title></head><body>
<h1>Procès-Verbal de Retrait / Rappel de Lot (%s)</h1>
<p><strong>Référence :</strong> %s</p>
<p><strong>Produit (%s) :</strong> %s — <strong>Lots :</strong> %s</p>
<p><strong>Sévérité :</strong> %s | <strong>Exercice à blanc :</strong> %v</p>
<p><strong>Motif initial :</strong> %s</p>
<p><strong>Notes de clôture / destruction :</strong> %s</p>
<p><strong>Clôturé par :</strong> %s (%s) le %s UTC</p>
</body></html>`,
		html.EscapeString(it.RecallReference),
		html.EscapeString(it.RecallReference),
		html.EscapeString(it.RecallReference),
		html.EscapeString(strings.ToUpper(it.ProductKind)),
		html.EscapeString(it.ProductName),
		html.EscapeString(it.LotNumbers),
		html.EscapeString(it.SeverityClass),
		it.IsMockDrill,
		html.EscapeString(it.Reason),
		html.EscapeString(req.ResolutionNotes),
		html.EscapeString(user.Email),
		html.EscapeString(user.Role),
		time.Now().UTC().Format("2006-01-02 15:04:05"),
	)

	var gedDocUUID *uuid.UUID
	var gedDocCode string
	if gedIDStr, code := archiveGeneratedDocInGED(r, user, h.gedRepo, h.storageProvider, h.auditSvc, gedArchiveSpec{
		CodePrefix:      "PV-RCL",
		TitleFR:         fmt.Sprintf("PV de Clôture Rappel %s — %s", it.RecallReference, it.ProductName),
		TaxonomyPath:    "04_Vigilance_Post_Market/04_03_Rappels_Lots",
		DocumentType:    "PV_RAPPEL_LOT",
		Department:      "pharmacovigilance",
		Tags:            []string{"ANPP", "RappelLot", it.RecallReference, it.SeverityClass},
		RevisionSummary: "Procès-verbal de clôture / destruction archivé lors de la clôture du rappel",
		MimeType:        "text/html",
		Filename:        fmt.Sprintf("PV-Rappel-%s.html", it.RecallReference),
		SourceTag:       "recall_pv_auto_archive",
	}, []byte(pvHTML)); gedIDStr != "" {
		if parsed, pErr := uuid.Parse(gedIDStr); pErr == nil {
			gedDocUUID = &parsed
			gedDocCode = code
		}
	}

	_, err = h.pool.Exec(r.Context(), `
		UPDATE lot_recalls
		SET status = 'cloture',
		    ged_pv_document_id = COALESCE($1, ged_pv_document_id),
		    signed_by_pdt_id = $2,
		    signed_at = NOW(),
		    updated_at = NOW()
		WHERE id = $3 AND organization_id = $4
	`, gedDocUUID, user.UserID, recallID, user.OrganizationID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur clôture rappel de lot")
		return
	}

	if h.auditSvc != nil {
		_, _ = h.auditSvc.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "RECALL_CLOSED",
			EntityType:     "lot_recall",
			EntityID:       it.RecallReference,
			Details: map[string]any{
				"recall_id":        recallID.String(),
				"ged_pv_doc_code":  gedDocCode,
				"resolution_notes": req.ResolutionNotes,
			},
			IPAddress: getClientIP(r),
			UserAgent: r.UserAgent(),
		})
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":             "cloture",
		"ged_pv_document_id": gedDocUUID,
		"ged_pv_doc_code":    gedDocCode,
	})
}
