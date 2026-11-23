package api

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"

	"homolog/internal/audit"
	"homolog/internal/auth"
	"homolog/internal/domain"
	"homolog/internal/repository"
)

// CorporateHandler gère les endpoints REST du module Sanad Corporate & MIPH
type CorporateHandler struct {
	corpRepo *repository.CorporateRepository
	auditSvc *audit.Service
}

// NewCorporateHandler instancie le CorporateHandler
func NewCorporateHandler(corpRepo *repository.CorporateRepository, auditSvc *audit.Service) *CorporateHandler {
	return &CorporateHandler{
		corpRepo: corpRepo,
		auditSvc: auditSvc,
	}
}

func (h *CorporateHandler) logAction(ctx context.Context, user *auth.SessionUser, action, entityType, entityID string, details map[string]interface{}) {
	if h.auditSvc != nil && user != nil {
		_, _ = h.auditSvc.Log(ctx, &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         action,
			EntityType:     entityType,
			EntityID:       entityID,
			Details:        details,
		})
	}
}

// ListApplications renvoie la liste des demandes d'agrément filtrées
func (h *CorporateHandler) ListApplications(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Non authentifié")
		return
	}

	cat := r.URL.Query().Get("category")
	status := r.URL.Query().Get("status")

	apps, err := h.corpRepo.ListApplications(r.Context(), user.OrganizationID, cat, status)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":      true,
		"applications": apps,
		"count":        len(apps),
	})
}

// GetApplication renvoie le détail d'une demande avec ses pièces de checklist
func (h *CorporateHandler) GetApplication(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Non authentifié")
		return
	}

	appID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "ID de demande invalide")
		return
	}

	app, err := h.corpRepo.GetApplication(r.Context(), user.OrganizationID, appID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"application": app,
	})
}

// CreateApplicationRequest payload de création d'une demande
type CreateApplicationRequest struct {
	EstablishmentID   *uuid.UUID                   `json:"establishment_id,omitempty"`
	Category          domain.EstablishmentCategory `json:"category"`
	Nature            domain.ApplicationNature     `json:"nature"`
	ApplicationNumber string                       `json:"application_number"`
	InternalReference string                       `json:"internal_reference,omitempty"`
	Title             string                       `json:"title"`
	FormData          map[string]interface{}       `json:"form_data,omitempty"`
}

// CreateApplication initialise un nouveau dossier d'agrément ministériel
func (h *CorporateHandler) CreateApplication(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Non authentifié")
		return
	}

	var req CreateApplicationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Payload invalide")
		return
	}

	if req.Category == "" || req.Nature == "" || strings.TrimSpace(req.Title) == "" {
		writeJSONError(w, http.StatusBadRequest, "Catégorie, nature et titre requis")
		return
	}

	if strings.TrimSpace(req.ApplicationNumber) == "" {
		req.ApplicationNumber = fmt.Sprintf("MIPH-%d-%s-%03d", time.Now().Year(), req.Category[:3], time.Now().Unix()%1000)
	}

	app := &domain.CorporateApplication{
		ID:                uuid.New(),
		OrganizationID:    user.OrganizationID,
		EstablishmentID:   req.EstablishmentID,
		Category:          req.Category,
		Nature:            req.Nature,
		ApplicationNumber: req.ApplicationNumber,
		InternalReference: req.InternalReference,
		Title:             req.Title,
		FormData:          req.FormData,
		CreatedBy:         &user.UserID,
	}

	if err := h.corpRepo.CreateApplication(r.Context(), app); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.logAction(r.Context(), user, "CORP_CREATE_APP", "corporate_applications", app.ID.String(), map[string]interface{}{
		"application_number": app.ApplicationNumber,
		"category":           app.Category,
		"nature":             app.Nature,
	})

	created, err := h.corpRepo.GetApplication(r.Context(), user.OrganizationID, app.ID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "erreur récupération demande créée: "+err.Error())
		return
	}
	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success":     true,
		"application": created,
	})
}

// UpdateApplication met à jour le formulaire e-Form et métadonnées
func (h *CorporateHandler) UpdateApplication(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Non authentifié")
		return
	}

	appID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "ID invalide")
		return
	}

	var req struct {
		Title             string                 `json:"title"`
		InternalReference string                 `json:"internal_reference"`
		FormData          map[string]interface{} `json:"form_data"`
		EstablishmentID   *uuid.UUID             `json:"establishment_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Payload invalide")
		return
	}

	app := &domain.CorporateApplication{
		ID:                appID,
		OrganizationID:    user.OrganizationID,
		Title:             req.Title,
		InternalReference: req.InternalReference,
		FormData:          req.FormData,
		EstablishmentID:   req.EstablishmentID,
	}

	if err := h.corpRepo.UpdateApplication(r.Context(), app); err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	h.logAction(r.Context(), user, "CORP_UPDATE_APP", "corporate_applications", appID.String(), map[string]interface{}{
		"title": app.Title,
	})

	updated, _ := h.corpRepo.GetApplication(r.Context(), user.OrganizationID, appID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"application": updated,
	})
}

// AttachDocument rattache une pièce GED à une exigence de checklist
func (h *CorporateHandler) AttachDocument(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Non authentifié")
		return
	}

	appID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "ID demande invalide")
		return
	}
	itemID, err := uuid.Parse(chi.URLParam(r, "itemId"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "ID exigence invalide")
		return
	}

	var req struct {
		GEDDocumentID uuid.UUID `json:"ged_document_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.GEDDocumentID == uuid.Nil {
		writeJSONError(w, http.StatusBadRequest, "ID document GED requis")
		return
	}

	if err := h.corpRepo.AttachDocumentToItem(r.Context(), user.OrganizationID, appID, itemID, req.GEDDocumentID); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.logAction(r.Context(), user, "CORP_ATTACH_DOC", "corporate_checklist_items", itemID.String(), map[string]interface{}{
		"application_id":  appID.String(),
		"ged_document_id": req.GEDDocumentID.String(),
	})

	updated, _ := h.corpRepo.GetApplication(r.Context(), user.OrganizationID, appID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"application": updated,
	})
}

// DetachDocument détache une pièce GED d'une exigence
func (h *CorporateHandler) DetachDocument(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Non authentifié")
		return
	}

	appID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "ID demande invalide")
		return
	}
	itemID, err := uuid.Parse(chi.URLParam(r, "itemId"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "ID exigence invalide")
		return
	}

	if err := h.corpRepo.DetachDocumentFromItem(r.Context(), user.OrganizationID, appID, itemID); err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.logAction(r.Context(), user, "CORP_DETACH_DOC", "corporate_checklist_items", itemID.String(), map[string]interface{}{
		"application_id": appID.String(),
	})

	updated, _ := h.corpRepo.GetApplication(r.Context(), user.OrganizationID, appID)
	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"application": updated,
	})
}

// SignPDT appose le visa formel du Pharmacien Directeur Technique avec scellement SHA-256
func (h *CorporateHandler) SignPDT(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Non authentifié")
		return
	}

	appID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "ID demande invalide")
		return
	}

	rawSeal := fmt.Sprintf("MIPH-PDT-VISA|ORG:%s|APP:%s|PDT:%s|DATE:%s",
		user.OrganizationID, appID, user.UserID, time.Now().UTC().Format(time.RFC3339),
	)
	hash := sha256.Sum256([]byte(rawSeal))
	seal := hex.EncodeToString(hash[:])

	signedApp, err := h.corpRepo.SignApplicationPDT(r.Context(), user.OrganizationID, appID, user.UserID, seal)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.logAction(r.Context(), user, "CORP_SIGN_PDT", "corporate_applications", appID.String(), map[string]interface{}{
		"seal": seal,
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"application": signedApp,
	})
}

// SubmitTabadol enregistre la référence de télésoumission sur la plateforme Tabadol
func (h *CorporateHandler) SubmitTabadol(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Non authentifié")
		return
	}

	appID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "ID demande invalide")
		return
	}

	var req struct {
		TabadolReference string `json:"tabadol_reference"`
		Notes            string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.TabadolReference) == "" {
		writeJSONError(w, http.StatusBadRequest, "Numéro de référence Tabadol requis")
		return
	}

	submitted, err := h.corpRepo.SubmitTabadol(r.Context(), user.OrganizationID, appID, req.TabadolReference, req.Notes)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.logAction(r.Context(), user, "CORP_SUBMIT_TABADOL", "corporate_applications", appID.String(), map[string]interface{}{
		"tabadol_ref": req.TabadolReference,
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"application": submitted,
	})
}

// RecordMIPHReceipt enregistre le récépissé physique du bureau d'ordre central
func (h *CorporateHandler) RecordMIPHReceipt(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Non authentifié")
		return
	}

	appID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "ID demande invalide")
		return
	}

	var req struct {
		MIPHReceiptRef  string `json:"miph_receipt_ref"`
		MIPHReceiptDate string `json:"miph_receipt_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.MIPHReceiptRef) == "" {
		writeJSONError(w, http.StatusBadRequest, "Numéro de récépissé bureau d'ordre requis")
		return
	}

	var parsedDate *time.Time
	if req.MIPHReceiptDate != "" {
		t, err := time.Parse("2006-01-02", req.MIPHReceiptDate)
		if err == nil {
			parsedDate = &t
		}
	}
	if parsedDate == nil {
		now := time.Now()
		parsedDate = &now
	}

	updated, err := h.corpRepo.RecordMIPHReceipt(r.Context(), user.OrganizationID, appID, req.MIPHReceiptRef, parsedDate)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.logAction(r.Context(), user, "CORP_RECORD_RECEIPT", "corporate_applications", appID.String(), map[string]interface{}{
		"receipt_ref": req.MIPHReceiptRef,
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"application": updated,
	})
}

// GrantLicense enregistre l'arrêté ou décision d'agrément octroyée
func (h *CorporateHandler) GrantLicense(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Non authentifié")
		return
	}

	appID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "ID demande invalide")
		return
	}

	var req struct {
		DecisionRef  string `json:"decision_ref"`
		DecisionDate string `json:"decision_date"`
		ExpiryDate   string `json:"expiry_date"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil || strings.TrimSpace(req.DecisionRef) == "" {
		writeJSONError(w, http.StatusBadRequest, "Numéro de décision ministérielle requis")
		return
	}

	decDate, err := time.Parse("2006-01-02", req.DecisionDate)
	if err != nil {
		decDate = time.Now()
	}

	var expDate *time.Time
	if req.ExpiryDate != "" {
		if t, err := time.Parse("2006-01-02", req.ExpiryDate); err == nil {
			expDate = &t
		}
	}
	if expDate == nil {
		fiveYears := decDate.AddDate(5, 0, 0)
		expDate = &fiveYears
	}

	updated, err := h.corpRepo.GrantLicense(r.Context(), user.OrganizationID, appID, req.DecisionRef, &decDate, expDate)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	h.logAction(r.Context(), user, "CORP_GRANT_LICENSE", "corporate_applications", appID.String(), map[string]interface{}{
		"decision_ref": req.DecisionRef,
	})

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":     true,
		"application": updated,
	})
}

// AuditTabadol renvoie le rapport d'audit bloquant de pré-soumission
func (h *CorporateHandler) AuditTabadol(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Non authentifié")
		return
	}

	appID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "ID demande invalide")
		return
	}

	report, err := h.corpRepo.AuditTabadolPreSubmission(r.Context(), user.OrganizationID, appID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"audit":   report,
	})
}

// ExportTabadolPackage génère l'archive .zip officielle pour la plateforme Tabadol
func (h *CorporateHandler) ExportTabadolPackage(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Non authentifié")
		return
	}

	appID, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "ID demande invalide")
		return
	}

	app, err := h.corpRepo.GetApplication(r.Context(), user.OrganizationID, appID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, err.Error())
		return
	}

	w.Header().Set("Content-Type", "application/zip")
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"Submission_Package_Tabadol_%s.zip\"", app.ApplicationNumber))

	zw := zip.NewWriter(w)
	defer zw.Close()

	// 1. Manifeste JSON Tabadol
	manifestData := map[string]interface{}{
		"miph_platform":      "TABADOL",
		"specification":      "DECRET_21_82_V2025",
		"application_number": app.ApplicationNumber,
		"category":           app.Category,
		"nature":             app.Nature,
		"title":              app.Title,
		"form_type":          app.FormType,
		"form_data":          app.FormData,
		"completeness_pct":   app.CompletenessPercent,
		"pdt_signature_seal": app.PDTSignatureSeal,
		"exported_at":        time.Now().UTC().Format(time.RFC3339),
		"checklist_count":    len(app.ChecklistItems),
	}
	manifestBytes, _ := json.MarshalIndent(manifestData, "", "  ")

	fManifest, err := zw.Create("MANIFEST_TABADOL.json")
	if err == nil {
		_, _ = fManifest.Write(manifestBytes)
	}

	// 2. Fichier d'index des pièces jointes
	fIndex, err := zw.Create("CHECKLIST_PIECES_JUSTIFICATIVES.csv")
	if err == nil {
		_, _ = fIndex.Write([]byte("Code;Section;Niveau;Exigence;Statut;DocCode;SHA256\n"))
		for _, it := range app.ChecklistItems {
			line := fmt.Sprintf("%s;%s;%s;%s;%s;%s;%s\n",
				it.ItemCode, it.Section, it.RequirementLevel, it.Title, it.Status, it.DocumentCode, it.FileSHA256,
			)
			_, _ = fIndex.Write([]byte(line))
		}
	}
}

// ListLicenses liste les titres réglementaires et compte à rebours d'expiration
func (h *CorporateHandler) ListLicenses(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Non authentifié")
		return
	}

	status := r.URL.Query().Get("status")
	licenses, err := h.corpRepo.ListLicenses(r.Context(), user.OrganizationID, status)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"licenses": licenses,
		"count":    len(licenses),
	})
}
