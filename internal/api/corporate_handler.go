package api

import (
	"archive/zip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html"
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

	// 3. Bordereau récapitulatif officiel d'accompagnement (pour émargement au Bureau d'Ordre Central du MIPH)
	bordereauHTML := renderBordereauMIPHHTML(app)
	fBordereau, err := zw.Create("BORDEREAU_ACCOMPAGNEMENT_MIPH.html")
	if err == nil {
		_, _ = fBordereau.Write([]byte(bordereauHTML))
	}
}

// GetMIPHBordereau renvoie le bordereau récapitulatif officiel d'accompagnement en HTML imprimable
func (h *CorporateHandler) GetMIPHBordereau(w http.ResponseWriter, r *http.Request) {
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

	htmlContent := renderBordereauMIPHHTML(app)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(htmlContent))
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

// renderBordereauMIPHHTML génère le bordereau officiel d'accompagnement pour émargement avec décharge au bureau d'ordre central MIPH
func renderBordereauMIPHHTML(app *domain.CorporateApplication) string {
	dateStr := time.Now().Format("02/01/2006 à 15:04")
	pdtVisa := "EN ATTENTE DE VISA DU PHARMACIEN DIRECTEUR TECHNIQUE"
	pdtSealInfo := "Le visa formel n'a pas encore été apposé."
	if app.PDTSignedAt != nil {
		pdtVisa = "DOCUMENT CERTIFIÉ CONFORME PAR LE PHARMACIEN DT"
		pdtSealInfo = fmt.Sprintf("Visa 21 CFR Part 11 apposé le %s<br>Empreinte cryptographique : <code>%s</code>",
			app.PDTSignedAt.Format("02/01/2006 15:04:05 MST"),
			html.EscapeString(app.PDTSignatureSeal),
		)
	}

	tabadolInfo := "Non transmis sur la passerelle Tabadol"
	if app.TabadolSubmissionRef != "" {
		tabadolDate := ""
		if app.TabadolSubmittedAt != nil {
			tabadolDate = " le " + app.TabadolSubmittedAt.Format("02/01/2006 à 15:04")
		}
		tabadolInfo = fmt.Sprintf("Télésoumis avec succès — Réf. <strong>%s</strong>%s",
			html.EscapeString(app.TabadolSubmissionRef), tabadolDate,
		)
	}

	rowsHTML := strings.Builder{}
	for _, it := range app.ChecklistItems {
		badgeClass := "status-missing"
		badgeLabel := "Manquant"
		if it.Status == "rattache" || it.Status == "valide_qualite" {
			badgeClass = "status-attached"
			badgeLabel = "Rattaché"
		}

		shaDisplay := "-"
		if it.FileSHA256 != "" {
			if len(it.FileSHA256) > 16 {
				shaDisplay = it.FileSHA256[:16] + "…"
			} else {
				shaDisplay = it.FileSHA256
			}
		}

		docCode := "-"
		if it.DocumentCode != "" {
			docCode = it.DocumentCode
		}

		rowsHTML.WriteString(fmt.Sprintf(`
			<tr>
				<td style="font-family:monospace; font-weight:700;">%s</td>
				<td><strong>%s</strong></td>
				<td style="text-align:center;"><span class="badge-req req-%s">[%s]</span></td>
				<td>%s</td>
				<td style="text-align:center;"><span class="%s">%s</span></td>
				<td style="font-family:monospace; font-size:7.5pt; text-align:center;">%s</td>
			</tr>
		`,
			html.EscapeString(it.ItemCode),
			html.EscapeString(it.Title),
			strings.ToLower(it.RequirementLevel),
			html.EscapeString(it.RequirementLevel),
			html.EscapeString(docCode),
			badgeClass, badgeLabel,
			html.EscapeString(shaDisplay),
		))
	}

	return fmt.Sprintf(`<!DOCTYPE html>
<html lang="fr">
<head>
    <meta charset="UTF-8">
    <title>Bordereau Récapitulatif d'Accompagnement — MIPH / Tabadol (%s)</title>
    <style>
        @page { size: A4; margin: 12mm 15mm 15mm 15mm; }
        * { box-sizing: border-box; }
        body { font-family: 'Segoe UI', Arial, sans-serif; font-size: 8.5pt; line-height: 1.35; color: #0f172a; margin: 0; padding: 15px; }
        .header-table { width: 100%%; border-collapse: collapse; margin-bottom: 12px; }
        .header-title { text-align: center; }
        .header-title h1 { font-size: 11pt; font-weight: 800; margin: 0 0 2px 0; text-transform: uppercase; letter-spacing: 0.5px; }
        .header-title h2 { font-size: 9pt; font-weight: 700; color: #1e3a8a; margin: 0 0 2px 0; }
        .header-title h3 { font-size: 8pt; font-weight: 600; color: #475569; margin: 0; }
        .doc-banner { background: #0f2942; color: white; padding: 8px 12px; border-radius: 4px; text-align: center; margin: 10px 0 14px 0; }
        .doc-banner h4 { margin: 0; font-size: 10.5pt; letter-spacing: 0.8px; text-transform: uppercase; }
        .doc-banner p { margin: 2px 0 0 0; font-size: 7.5pt; opacity: 0.9; }
        .meta-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 8px; margin-bottom: 14px; }
        .meta-card { border: 1px solid #cbd5e1; border-radius: 4px; padding: 8px 10px; background: #f8fafc; }
        .meta-card table { width: 100%%; font-size: 8pt; border-collapse: collapse; }
        .meta-card td { padding: 2px 4px; vertical-align: top; }
        .meta-card td.lbl { font-weight: 600; color: #475569; width: 40%%; }
        .items-table { width: 100%%; border-collapse: collapse; font-size: 7.8pt; margin-bottom: 16px; }
        .items-table th { background: #1e293b; color: white; padding: 6px 6px; text-align: left; font-size: 7.5pt; text-transform: uppercase; }
        .items-table td { border-bottom: 1px solid #e2e8f0; padding: 5px 6px; }
        .items-table tr:nth-child(even) { background: #f8fafc; }
        .badge-req { font-weight: 700; padding: 1px 4px; border-radius: 3px; font-size: 7pt; }
        .req-e { background: #fee2e2; color: #b91c1c; }
        .req-c { background: #fef3c7; color: #b45309; }
        .req-f { background: #e0f2fe; color: #0369a1; }
        .req-s { background: #f3e8ff; color: #6b21a8; }
        .status-attached { background: #dcfce7; color: #15803d; font-weight: 700; padding: 2px 6px; border-radius: 3px; font-size: 7pt; }
        .status-missing { background: #fee2e2; color: #b91c1c; font-weight: 700; padding: 2px 6px; border-radius: 3px; font-size: 7pt; }
        .sign-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 14px; margin-top: 14px; page-break-inside: avoid; }
        .sign-box { border: 1.5px solid #0f2942; border-radius: 4px; padding: 10px; min-height: 120px; position: relative; }
        .sign-header { font-weight: 700; font-size: 8pt; color: #0f2942; border-bottom: 1px solid #cbd5e1; padding-bottom: 4px; margin-bottom: 6px; text-transform: uppercase; }
        .sign-space { height: 60px; display: flex; align-items: center; justify-content: center; color: #94a3b8; font-style: italic; font-size: 7.5pt; }
        .footer-note { font-size: 7pt; color: #64748b; text-align: center; margin-top: 15px; border-top: 1px solid #e2e8f0; padding-top: 6px; }
        @media print {
            body { padding: 0; }
            .no-print { display: none; }
        }
    </style>
</head>
<body>
    <div class="no-print" style="margin-bottom: 12px; text-align: right;">
        <button onclick="window.print()" style="background:#0f2942; color:white; border:none; padding:6px 14px; border-radius:4px; font-weight:700; cursor:pointer;">🖨️ Imprimer le Bordereau</button>
    </div>

    <table class="header-table">
        <tr>
            <td style="width: 20%%; text-align: left; font-size: 7.5pt; color: #64748b;">
                RÉPUBLIQUE ALGÉRIENNE DÉMOCRATIQUE ET POPULAIRE<br>
                MINISTÈRE DE L'INDUSTRIE PHARMACEUTIQUE (MIPH)<br>
                DIRECTION DES AGRÉMENTS
            </td>
            <td class="header-title">
                <h1>Direction Générale de la Réglementation</h1>
                <h2>Guichet de Télésoumission Tabadol (gate.miph.gov.dz)</h2>
                <h3>Décret exécutif n° 21-82 du 23 février 2021</h3>
            </td>
            <td style="width: 20%%; text-align: right; font-size: 7.5pt; color: #64748b;">
                BUREAU D'ORDRE CENTRAL<br>
                Gouvernance BPF / ALCOA+<br>
                Date d'édition : %s
            </td>
        </tr>
    </table>

    <div class="doc-banner">
        <h4>Bordereau Récapitulatif d'Accompagnement du Dossier d'Agrément</h4>
        <p>Document officiel à joindre au dépôt physique pour émargement avec décharge au Bureau d'Ordre Central du Ministère</p>
    </div>

    <div class="meta-grid">
        <div class="meta-card">
            <table>
                <tr><td class="lbl">Réf. Demande :</td><td><strong>%s</strong></td></tr>
                <tr><td class="lbl">Intitulé :</td><td>%s</td></tr>
                <tr><td class="lbl">Catégorie :</td><td><strong>%s</strong></td></tr>
                <tr><td class="lbl">Nature Démarche :</td><td>%s</td></tr>
            </table>
        </div>
        <div class="meta-card">
            <table>
                <tr><td class="lbl">Établissement :</td><td><strong>%s</strong></td></tr>
                <tr><td class="lbl">Plateforme Tabadol :</td><td>%s</td></tr>
                <tr><td class="lbl">Complétude [E] :</td><td><strong>%.1f %%%%</strong></td></tr>
                <tr><td class="lbl">Statut Dossier :</td><td><strong>%s</strong></td></tr>
            </table>
        </div>
    </div>

    <table class="items-table">
        <thead>
            <tr>
                <th style="width: 12%%%%;">Code</th>
                <th style="width: 44%%%%;">Pièce Justificative Exigée (Décret 21-82)</th>
                <th style="width: 8%%%%; text-align:center;">Niv.</th>
                <th style="width: 14%%%%;">Réf. GED</th>
                <th style="width: 10%%%%; text-align:center;">État</th>
                <th style="width: 12%%%%; text-align:center;">Empreinte SHA-256</th>
            </tr>
        </thead>
        <tbody>
            %s
        </tbody>
    </table>

    <div class="sign-grid">
        <div class="sign-box">
            <div class="sign-header">Visa & Certification du Pharmacien Directeur Technique</div>
            <div style="font-size:7pt; color:#475569; margin-bottom:4px;">
                Je soussigné, Pharmacien Directeur Technique, certifie sur l'honneur l'exactitude des informations et pièces rattachées à la présente demande.
            </div>
            <div class="sign-space">
                <div>
                    <strong style="color:#0f2942;">%s</strong><br>
                    <span style="font-size:6.8pt;">%s</span>
                </div>
            </div>
        </div>
        <div class="sign-box">
            <div class="sign-header">Cadre Réservé au Bureau d'Ordre Central du MIPH</div>
            <div style="font-size:7pt; color:#475569; margin-bottom:4px;">
                Réservé à l'émargement physique, enregistrement et apposition de la décharge officielle :
            </div>
            <div class="sign-space">
                <div>
                    Date de dépôt : ____ / ____ / 2026<br>
                    N° d'enregistrement BO : ____________________<br>
                    <span style="font-size:6.8pt;">Cachet humide & visa de l'agent récepteur</span>
                </div>
            </div>
        </div>
    </div>

    <div class="footer-note">
        SANAD PHARMA • Plateforme Réglementaire Souveraine Homolog • Intégrité ALCOA+ & 21 CFR Part 11 • Document scellé cryptographiquement
    </div>
</body>
</html>`,
		html.EscapeString(app.ApplicationNumber),
		dateStr,
		html.EscapeString(app.ApplicationNumber),
		html.EscapeString(app.Title),
		html.EscapeString(string(app.Category)),
		html.EscapeString(string(app.Nature)),
		html.EscapeString(app.EstablishmentName),
		tabadolInfo,
		app.CompletenessPercent,
		html.EscapeString(string(app.Status)),
		rowsHTML.String(),
		pdtVisa,
		pdtSealInfo,
	)
}
