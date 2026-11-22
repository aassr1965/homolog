package api

import (
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
	"homolog/internal/domain"
	"homolog/internal/reporting"
	"homolog/internal/repository"
)

// VariationHandler orchestre l'API des modifications post-AMM selon LD-DEPP07/V01
type VariationHandler struct {
	varRepo      *repository.VariationRepository
	dossierRepo  *repository.DossierRepository
	auditService *audit.Service
}

// NewVariationHandler instancie le gestionnaire d'API des variations
func NewVariationHandler(
	varRepo *repository.VariationRepository,
	dossierRepo *repository.DossierRepository,
	auditService *audit.Service,
) *VariationHandler {
	return &VariationHandler{
		varRepo:      varRepo,
		dossierRepo:  dossierRepo,
		auditService: auditService,
	}
}

// ListRubriques retourne le catalogue complet des 34 rubriques officielles algériennes
func (h *VariationHandler) ListRubriques(w http.ResponseWriter, r *http.Request) {
	category := strings.TrimSpace(r.URL.Query().Get("category"))
	classification := strings.TrimSpace(r.URL.Query().Get("classification"))

	result := make([]domain.RubriqueDefinition, 0)
	for _, rub := range domain.MasterRubriquesRegistry {
		if category != "" && rub.Category != category {
			continue
		}
		if classification != "" && string(rub.DefaultClassification) != classification {
			continue
		}
		result = append(result, rub)
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"total":     len(result),
		"rubriques": result,
	})
}

// GetRubrique retourne une rubrique spécifique avec ses conditions et ses impacts CTD
func (h *VariationHandler) GetRubrique(w http.ResponseWriter, r *http.Request) {
	code := chi.URLParam(r, "code")
	rub := domain.FindRubriqueDefinition(code)
	if rub == nil {
		writeJSONError(w, http.StatusNotFound, "Rubrique officielle introuvable")
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":  true,
		"rubrique": rub,
	})
}

// ListVariations retourne les demandes de variations de l'organisation
func (h *VariationHandler) ListVariations(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	var medID *uuid.UUID
	medIDStr := r.URL.Query().Get("medicine_id")
	if medIDStr != "" {
		if parsed, err := uuid.Parse(medIDStr); err == nil {
			medID = &parsed
		}
	}
	status := r.URL.Query().Get("status")

	variations, err := h.varRepo.ListVariationRequests(r.Context(), user.OrganizationID, medID, status)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Erreur lors de la récupération des variations: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":    true,
		"count":      len(variations),
		"variations": variations,
	})
}

// CreateVariation crée une nouvelle demande de variation
func (h *VariationHandler) CreateVariation(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	var payload struct {
		MedicineID               string `json:"medicine_id"`
		DossierID                string `json:"dossier_id"`
		TrackingNumber           string `json:"tracking_number"`
		Title                    string `json:"title"`
		Description              string `json:"description"`
		IsGrouping               bool   `json:"is_grouping"`
		GroupingJustification    string `json:"grouping_justification"`
		CurrentApprovedSituation string `json:"current_approved_situation"`
		ProposedSituation        string `json:"proposed_situation"`
		TechnicalJustification   string `json:"technical_justification"`
		NonModificationAttested  bool   `json:"non_modification_attested"`
		NonModificationDetails   string `json:"non_modification_details"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Payload JSON invalide")
		return
	}

	medUUID, err := uuid.Parse(payload.MedicineID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant produit (medicine_id) invalide")
		return
	}

	if strings.TrimSpace(payload.Title) == "" {
		writeJSONError(w, http.StatusBadRequest, "Le libellé de la variation est requis")
		return
	}
	if strings.TrimSpace(payload.CurrentApprovedSituation) == "" {
		writeJSONError(w, http.StatusBadRequest, "L'état actuel approuvé est obligatoire (Section 5.3 LD-DEPP07)")
		return
	}
	if strings.TrimSpace(payload.ProposedSituation) == "" {
		writeJSONError(w, http.StatusBadRequest, "La situation nouvelle proposée est obligatoire (Section 5.3 LD-DEPP07)")
		return
	}
	if strings.TrimSpace(payload.TechnicalJustification) == "" {
		writeJSONError(w, http.StatusBadRequest, "Le justificatif technico-réglementaire motivant le changement est obligatoire")
		return
	}

	tracking := strings.TrimSpace(payload.TrackingNumber)
	if tracking == "" {
		tracking = fmt.Sprintf("VAR-%d-%04d", time.Now().Year(), time.Now().Nanosecond()%10000)
	}

	var dossierUUID *uuid.UUID
	if payload.DossierID != "" {
		if dID, err := uuid.Parse(payload.DossierID); err == nil {
			dossierUUID = &dID
		}
	}

	req := domain.VariationRequest{
		OrganizationID:           user.OrganizationID,
		MedicineID:               medUUID,
		DossierID:                dossierUUID,
		TrackingNumber:           tracking,
		Title:                    payload.Title,
		Description:              payload.Description,
		IsGrouping:               payload.IsGrouping,
		GroupingJustification:    payload.GroupingJustification,
		OverallClassification:   domain.ClassMmin,
		NominalClassification:   domain.ClassMmin,
		CurrentApprovedSituation: payload.CurrentApprovedSituation,
		ProposedSituation:        payload.ProposedSituation,
		TechnicalJustification:   payload.TechnicalJustification,
		NonModificationAttested:  payload.NonModificationAttested,
		NonModificationDetails:   payload.NonModificationDetails,
		Status:                   "brouillon",
		CreatedBy:                &user.UserID,
	}

	if err := h.varRepo.CreateVariationRequest(r.Context(), &req); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Échec création demande variation: "+err.Error())
		return
	}

	if h.auditService != nil {
		_, _ = h.auditService.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "VARIATION_REQUEST_CREATED",
			EntityType:     "anpp_variation_requests",
			EntityID:       req.ID.String(),
			Details: map[string]any{
				"tracking_number": req.TrackingNumber,
				"medicine_id":     req.MedicineID.String(),
				"is_grouping":     req.IsGrouping,
			},
		})
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success":   true,
		"variation": req,
		"message":   "Demande de variation initialisée avec les 12 pièces obligatoires de la LD-DEPP07",
	})
}

// GetVariation retourne les détails complets d'une variation
func (h *VariationHandler) GetVariation(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	varUUID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant de variation invalide")
		return
	}

	variation, err := h.varRepo.GetVariationRequestByID(r.Context(), varUUID)
	if err != nil {
		if err == repository.ErrVariationNotFound {
			writeJSONError(w, http.StatusNotFound, "Variation introuvable")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, "Erreur lecture variation: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"variation": variation,
	})
}

// AddVariationItem ajoute une rubrique et réévalue l'escalade automatique
func (h *VariationHandler) AddVariationItem(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	idStr := chi.URLParam(r, "id")
	varUUID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant de variation invalide")
		return
	}

	var payload struct {
		RubriqueCode        string          `json:"rubrique_code"`
		ConditionsEvaluated map[string]bool `json:"conditions_evaluated"`
		CurrentState        string          `json:"current_state"`
		ProposedState       string          `json:"proposed_state"`
		Notes               string          `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Payload JSON invalide")
		return
	}

	rubDef := domain.FindRubriqueDefinition(payload.RubriqueCode)
	if rubDef == nil {
		writeJSONError(w, http.StatusBadRequest, "Rubrique officielle introuvable parmi les 34 rubriques LD-DEPP07")
		return
	}

	item := domain.VariationItem{
		VariationRequestID:  varUUID,
		RubriqueCode:        payload.RubriqueCode,
		ConditionsEvaluated: payload.ConditionsEvaluated,
		CurrentState:        payload.CurrentState,
		ProposedState:       payload.ProposedState,
		Notes:               payload.Notes,
	}

	if err := h.varRepo.AddVariationItem(r.Context(), &item); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Échec ajout rubrique: "+err.Error())
		return
	}

	// Récupérer la variation actualisée
	updatedVar, _ := h.varRepo.GetVariationRequestByID(r.Context(), varUUID)

	if h.auditService != nil {
		_, _ = h.auditService.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "VARIATION_ITEM_ADDED",
			EntityType:     "anpp_variation_items",
			EntityID:       item.ID.String(),
			Details: map[string]any{
				"rubrique_code": payload.RubriqueCode,
				"effective":     item.EffectiveClassification,
				"is_escalated":  item.IsEscalated,
			},
		})
	}

	writeJSON(w, http.StatusCreated, map[string]interface{}{
		"success":       true,
		"item":          item,
		"variation":     updatedVar,
		"is_escalated":  item.IsEscalated,
		"escalation":    item.EscalationReason,
	})
}

// UpdateVariationItem met à jour les conditions d'une rubrique (déclenche réévaluation)
func (h *VariationHandler) UpdateVariationItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	varUUID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant de variation invalide")
		return
	}

	itemIDStr := chi.URLParam(r, "itemId")
	itemUUID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant d'item invalide")
		return
	}

	var payload struct {
		RubriqueCode        string          `json:"rubrique_code"`
		ConditionsEvaluated map[string]bool `json:"conditions_evaluated"`
		CurrentState        string          `json:"current_state"`
		ProposedState       string          `json:"proposed_state"`
		Notes               string          `json:"notes"`
	}

	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Payload JSON invalide")
		return
	}

	item := domain.VariationItem{
		ID:                  itemUUID,
		VariationRequestID:  varUUID,
		RubriqueCode:        payload.RubriqueCode,
		ConditionsEvaluated: payload.ConditionsEvaluated,
		CurrentState:        payload.CurrentState,
		ProposedState:       payload.ProposedState,
		Notes:               payload.Notes,
	}

	if err := h.varRepo.UpdateVariationItem(r.Context(), &item); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Échec mise à jour rubrique: "+err.Error())
		return
	}

	updatedVar, _ := h.varRepo.GetVariationRequestByID(r.Context(), varUUID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":      true,
		"item":         item,
		"variation":    updatedVar,
		"is_escalated": item.IsEscalated,
		"escalation":   item.EscalationReason,
	})
}

// DeleteVariationItem supprime une rubrique
func (h *VariationHandler) DeleteVariationItem(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	varUUID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant de variation invalide")
		return
	}

	itemIDStr := chi.URLParam(r, "itemId")
	itemUUID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant d'item invalide")
		return
	}

	if err := h.varRepo.DeleteVariationItem(r.Context(), itemUUID, varUUID); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Échec suppression rubrique: "+err.Error())
		return
	}

	updatedVar, _ := h.varRepo.GetVariationRequestByID(r.Context(), varUUID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"variation": updatedVar,
		"message":   "Rubrique retirée et criticité globale recalculée",
	})
}

// AttachDocument rattache un document GED à l'une des 12 pièces requises
func (h *VariationHandler) AttachDocument(w http.ResponseWriter, r *http.Request) {
	docIDStr := chi.URLParam(r, "docId")
	docUUID, err := uuid.Parse(docIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant document invalide")
		return
	}

	var payload struct {
		GEDDocumentID string `json:"ged_document_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Payload JSON invalide")
		return
	}

	gedUUID, err := uuid.Parse(payload.GEDDocumentID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant GED invalide")
		return
	}

	if err := h.varRepo.AttachDocument(r.Context(), docUUID, gedUUID); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Échec rattachement pièce GED: "+err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"message": "Pièce justificative rattachée avec succès",
	})
}

// SealVariationDT applique le visa officiel 21 CFR Part 11 du Pharmacien Directeur Technique
func (h *VariationHandler) SealVariationDT(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	normRole := strings.ToLower(user.Role)
	if normRole != "pharmacien_responsable" && normRole != "directeur_technique" && normRole != "admin" {
		writeJSONError(w, http.StatusForbidden, "Seul le Pharmacien Directeur Technique est habilité à sceller officiellement une demande de modification")
		return
	}

	idStr := chi.URLParam(r, "id")
	varUUID, err := uuid.Parse(idStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant de variation invalide")
		return
	}

	var payload struct {
		SolemnDeclaration bool   `json:"solemn_declaration"`
		Notes             string `json:"notes"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Payload JSON invalide")
		return
	}

	if !payload.SolemnDeclaration {
		writeJSONError(w, http.StatusBadRequest, "L'engagement solennel du Pharmacien Directeur Technique est obligatoire pour le scellement BPF")
		return
	}

	// Générer l'empreinte cryptographique de scellement
	now := time.Now().UTC()
	rawSign := fmt.Sprintf("%s|%s|%s|%s", varUUID.String(), user.UserID.String(), user.Email, now.Format(time.RFC3339))
	hash := sha256.Sum256([]byte(rawSign))
	seal := fmt.Sprintf("DT-SEAL-VAR-%s", hex.EncodeToString(hash[:16]))

	if err := h.varRepo.SignAndSealVariationDT(r.Context(), varUUID, user.UserID, seal, payload.Notes); err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Échec scellement variation: "+err.Error())
		return
	}

	if h.auditService != nil {
		_, _ = h.auditService.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "VARIATION_SEALED_BY_DT",
			EntityType:     "anpp_variation_requests",
			EntityID:       varUUID.String(),
			Details: map[string]any{
				"seal":      seal,
				"sealed_at": now.Format(time.RFC3339),
			},
		})
	}

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"success":   true,
		"seal":      seal,
		"sealed_at": now.Format(time.RFC3339),
		"message":   "Demande de variation validée et scellée avec succès par le Pharmacien DT",
	})
}

// GetComparativeTableHTML génère l'état officiel haute fidélité A4 ES-REG-04
func (h *VariationHandler) GetComparativeTableHTML(w http.ResponseWriter, r *http.Request) {
	idStr := chi.URLParam(r, "id")
	varUUID, err := uuid.Parse(idStr)
	if err != nil {
		http.Error(w, "Identifiant invalide", http.StatusBadRequest)
		return
	}

	variation, err := h.varRepo.GetVariationRequestByID(r.Context(), varUUID)
	if err != nil {
		http.Error(w, "Variation introuvable: "+err.Error(), http.StatusNotFound)
		return
	}

	user := GetSessionUser(r.Context())
	estName := "Établissement Pharmaceutique Déposant"
	estAgreement := "MIPH-AGR-2026"
	estAddress := "Alger, Algérie"
	pdtName := "Dr. Yacine BENALIA"
	pdtDecision := "DEC-MIPH-2024-PDT"
	if user != nil {
		pdtName = user.Email
	}

	// Résumé des rubriques
	rubCodes := make([]string, 0)
	for _, it := range variation.Items {
		rubCodes = append(rubCodes, fmt.Sprintf("%s (%s)", it.RubriqueCode, it.EffectiveClassification))
	}
	rubText := strings.Join(rubCodes, " + ")
	if rubText == "" {
		rubText = "Rubrique 7.d.1 (Mmod)"
	}

	classLabel := fmt.Sprintf("%s (%s)", variation.OverallClassification, mapClassificationLabel(variation.OverallClassification))
	if variation.IsGrouping {
		classLabel += " [Procédure Groupée - Grouping Art. 09]"
	}

	watermark := domain.WatermarkDraft
	if variation.DTVerifiedAt != nil && variation.DTSignatureSeal != "" {
		watermark = domain.WatermarkOfficial
	}

	now := time.Now().UTC()
	rawHash := sha256.Sum256([]byte(variation.ID.String() + variation.TrackingNumber))
	shaStr := hex.EncodeToString(rawHash[:])
	qrURL := fmt.Sprintf("https://sanad.labo.dz/verify/%s", shaStr[:16])

	htmlContent := renderESREG04DynamicHTML(
		variation, estName, estAgreement, estAddress, pdtName, pdtDecision,
		watermark, shaStr, qrURL, now, rubText, classLabel,
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(htmlContent))
}

func mapClassificationLabel(c domain.VariationClassification) string {
	switch c {
	case domain.ClassMmaj:
		return "Variation Majeure — Décret 21-82"
	case domain.ClassMmod:
		return "Variation Modérée — LD-DEPP07/V01"
	case domain.ClassMmin:
		return "Variation Mineure — LD-DEPP07/V01"
	default:
		return "Variation Réglementaire"
	}
}

// renderESREG04DynamicHTML produit le rendu A4 officiel du Tableau Comparatif standardisé
func renderESREG04DynamicHTML(
	v *domain.VariationRequest,
	estName, estAgreement, estAddress, pdtName, pdtDecision string,
	watermark domain.WatermarkStatus,
	sha256Hash, qrURL string,
	generatedAt time.Time,
	rubText, classLabel string,
) string {
	wmHTML := reporting.RenderWatermark(watermark)
	headerHTML := reporting.RenderInstitutionalHeader(estName, estAgreement, estAddress, pdtName, pdtDecision, "ES-REG-04 • Ligne Directrice Variations LD-DEPP07/V01")
	footerHTML := reporting.RenderLegalFooter("ES-REG-04", watermark, sha256Hash, qrURL, 1, 1, generatedAt)
	signHTML := reporting.RenderSignaturesBlock(pdtName, pdtDecision, v.DTVerifiedAt, "Sous-Direction de l'Enregistrement (DEPP / ANPP)")

	itemsRows := strings.Builder{}
	for _, it := range v.Items {
		badgeEsc := ""
		if it.IsEscalated {
			badgeEsc = `<span style="background:#fee2e2; color:#b91c1c; font-size:7pt; padding:1px 4px; border-radius:3px; font-weight:700;">Reclassé</span>`
		}
		itemsRows.WriteString(fmt.Sprintf(`
			<tr>
				<td style="font-weight:700; font-family:monospace;">%s</td>
				<td>%s</td>
				<td style="text-align:center;"><strong>%s</strong> %s</td>
				<td style="font-size:7.5pt; color:#475569;">%s</td>
				<td style="font-size:7.5pt; color:#0369a1;">%s</td>
			</tr>
		`, html.EscapeString(it.RubriqueCode), html.EscapeString(it.RubriqueTitle), html.EscapeString(string(it.EffectiveClassification)), badgeEsc, html.EscapeString(it.CurrentState), html.EscapeString(it.ProposedState)))
	}

	docRows := strings.Builder{}
	for _, doc := range v.Documents {
		statusBadge := `<span style="color:#dc2626; font-weight:700;">Attente</span>`
		if doc.IsAttached {
			statusBadge = `<span style="color:#047857; font-weight:700;">✔ Versé</span>`
		}
		docRows.WriteString(fmt.Sprintf(`
			<tr>
				<td style="text-align:center; font-weight:700;">%d</td>
				<td style="font-weight:600;">%s</td>
				<td style="font-size:7pt; color:#64748b;">%s</td>
				<td style="text-align:center;">%s</td>
			</tr>
		`, doc.DocIndex, html.EscapeString(doc.Title), html.EscapeString(doc.Description), statusBadge))
	}

	feeFormatted := fmt.Sprintf("%.2f DZD", v.FeeAmountDZD)

	body := fmt.Sprintf(`
    <div class="no-print-toolbar">
        <div><strong>Homolog Reporting BPF</strong> • ES-REG-04 Tableau Comparatif LD-DEPP07/V01</div>
        <div style="display:flex; gap:10px;">
            <button class="btn btn-primary" onclick="window.print()">🖨️ Imprimer / Export PDF (A4 300 DPI)</button>
            <button type="button" class="btn btn-outline" onclick="if(window.parent && window.parent !== window && typeof window.parent.closeVariationReportModal === 'function'){ window.parent.closeVariationReportModal(); } else { window.close(); }">✕ Fermer</button>
        </div>
    </div>
    <script>
        if (window.parent && window.parent !== window) {
            var tb = document.querySelector('.no-print-toolbar');
            if (tb) tb.style.display = 'none';
            var sh = document.querySelector('.sheet');
            if (sh) sh.style.marginTop = '15px';
        }
    </script>
    <div class="sheet">
        %s
        %s

        <div class="title-banner">
            <div class="banner-title">ES-REG-04 — Tableau Comparatif des Modifications Post-AMM (LD-DEPP07/V01)</div>
            <div class="banner-code">%s • Dossier: %s</div>
        </div>

        <div class="info-grid">
            <div class="info-card">
                <div class="info-card-header">1. Identification de la Spécialité Pharmaceutique</div>
                <div class="field-row"><div class="field-label">Dénomination :</div><div class="field-value">%s</div></div>
                <div class="field-row"><div class="field-label">D.C.I. :</div><div class="field-value">%s</div></div>
                <div class="field-row"><div class="field-label">Forme & Dosage :</div><div class="field-value">%s — %s</div></div>
                <div class="field-row"><div class="field-label">N° Décision AMM :</div><div class="field-value">%s</div></div>
            </div>
            <div class="info-card">
                <div class="info-card-header">2. Classification Réglementaire Officielle ANPP</div>
                <div class="field-row"><div class="field-label">Classification :</div><div class="field-value" style="color:#0284c7; font-weight:800;">%s</div></div>
                <div class="field-row"><div class="field-label">Rubrique(s) :</div><div class="field-value" style="font-weight:700;">%s</div></div>
                <div class="field-row"><div class="field-label">Base Juridique :</div><div class="field-value">Ligne Directrice LD-DEPP07/V01 • Décret 21-82</div></div>
                <div class="field-row"><div class="field-label">Redevance BV2 :</div><div class="field-value" style="font-weight:700; color:#047857;">%s</div></div>
            </div>
        </div>

        <div style="margin-top:10px;">
            <div style="font-size:8.5pt; font-weight:800; color:#091e36; text-transform:uppercase; border-bottom:1.5px solid #0284c7; padding-bottom:3px; margin-bottom:6px;">
                3. Tableau Comparatif Standardisé (Section 5.3 LD-DEPP07/V01)
            </div>
            <table class="data-table" style="width:100%%; font-size:8pt;">
                <thead>
                    <tr style="background:#091e36; color:white;">
                        <th style="width:50px;">Code</th>
                        <th>Intitulé de la Rubrique</th>
                        <th style="width:70px; text-align:center;">Catégorie</th>
                        <th style="width:35%%;">Situation Actuelle Approuvée</th>
                        <th style="width:35%%;">Situation Nouvelle Proposée</th>
                    </tr>
                </thead>
                <tbody>
                    %s
                </tbody>
            </table>
        </div>

        <div style="margin-top:10px; display:grid; grid-template-columns:1fr 1fr; gap:10px;">
            <div style="background:#f8fafc; border:1px solid #cbd5e1; border-radius:4px; padding:8px;">
                <div style="font-size:7.5pt; font-weight:800; color:#091e36; text-transform:uppercase; margin-bottom:4px;">
                    4. Justificatif Technico-Réglementaire
                </div>
                <div style="font-size:7.5pt; line-height:1.4; color:#334155;">
                    %s
                </div>
            </div>
            <div style="background:#f8fafc; border:1px solid #cbd5e1; border-radius:4px; padding:8px;">
                <div style="font-size:7.5pt; font-weight:800; color:#091e36; text-transform:uppercase; margin-bottom:4px;">
                    5. Attestation Formelle de Non-Modification
                </div>
                <div style="font-size:7.5pt; line-height:1.4; color:#334155;">
                    Le Pharmacien Directeur Technique soussigné atteste formellement sur l'honneur qu'aucune section du dossier d'enregistrement n'a été modifiée en dehors de celles déclarées ci-dessus.
                </div>
            </div>
        </div>

        <div style="margin-top:10px;">
            <div style="font-size:8.5pt; font-weight:800; color:#091e36; text-transform:uppercase; border-bottom:1.5px solid #0284c7; padding-bottom:3px; margin-bottom:6px;">
                6. Contrôle des 12 Pièces Obligatoires Exigées pour Toute Variation (Section 5.3 CDC)
            </div>
            <table class="data-table" style="width:100%%; font-size:7.5pt;">
                <thead>
                    <tr style="background:#f1f5f9; color:#091e36;">
                        <th style="width:25px; text-align:center;">N°</th>
                        <th style="width:40%%;">Pièce Justificative Exigée</th>
                        <th>Objet Réglementaire</th>
                        <th style="width:60px; text-align:center;">Statut</th>
                    </tr>
                </thead>
                <tbody>
                    %s
                </tbody>
            </table>
        </div>

        %s
        %s
    </div>
	`,
		wmHTML, headerHTML,
		html.EscapeString(v.TrackingNumber), html.EscapeString(v.TrackingNumber),
		html.EscapeString(v.MedicineName), html.EscapeString(v.MedicineDCI),
		html.EscapeString(v.MedicineForm), html.EscapeString(v.MedicineDosage),
		html.EscapeString(v.RegistrationNumber),
		html.EscapeString(classLabel), html.EscapeString(rubText), feeFormatted,
		itemsRows.String(),
		html.EscapeString(v.TechnicalJustification),
		docRows.String(),
		signHTML, footerHTML,
	)

	return reporting.RenderDocumentWrap("ES-REG-04 • Tableau Comparatif LD-DEPP07/V01", body)
}
