package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"homolog/internal/audit"
	"homolog/internal/auth"
	"homolog/internal/domain"
	"homolog/internal/repository"
)

// DossierHandler gère les endpoints REST du moteur de dossiers ANPP et checklists CTD
type DossierHandler struct {
	repo         *repository.DossierRepository
	pharmaRepo   *repository.PharmaRepository
	gedRepo      *repository.GEDRepository
	auditService *audit.Service
	pool         *pgxpool.Pool
}

// NewDossierHandler initialise le gestionnaire d'API des dossiers
func NewDossierHandler(
	repo *repository.DossierRepository,
	pharmaRepo *repository.PharmaRepository,
	gedRepo *repository.GEDRepository,
	auditService *audit.Service,
	pool *pgxpool.Pool,
) *DossierHandler {
	return &DossierHandler{
		repo:         repo,
		pharmaRepo:   pharmaRepo,
		gedRepo:      gedRepo,
		auditService: auditService,
		pool:         pool,
	}
}

// ListDossiers retourne tous les dossiers réglementaires de l'organisation
func (h *DossierHandler) ListDossiers(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	procFilter := r.URL.Query().Get("procedure_code")
	statusFilter := r.URL.Query().Get("status")

	dossiers, err := h.repo.ListDossiers(r.Context(), user.OrganizationID, procFilter, statusFilter)
	if err != nil {
		log.Printf("[ERROR] ListDossiers: %v", err)
		writeJSONError(w, http.StatusInternalServerError, "Impossible de charger les dossiers")
		return
	}

	if dossiers == nil {
		dossiers = []domain.Dossier{}
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"dossiers": dossiers,
		"total":    len(dossiers),
	})
}

// GetDossier retourne un dossier avec l'ensemble de ses exigences de checklist et documents GED associés
func (h *DossierHandler) GetDossier(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	dossierIDStr := chi.URLParam(r, "id")
	dossierID, err := uuid.Parse(dossierIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant de dossier invalide")
		return
	}

	dossier, err := h.repo.GetDossierByID(r.Context(), user.OrganizationID, dossierID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			writeJSONError(w, http.StatusNotFound, "Dossier réglementaire introuvable")
			return
		}
		log.Printf("[ERROR] GetDossierByID %s: %v", dossierID, err)
		writeJSONError(w, http.StatusInternalServerError, "Impossible de charger le dossier")
		return
	}

	writeJSON(w, http.StatusOK, dossier)
}

// CreateDossierRequest payload pour la création d'un dossier
type CreateDossierRequest struct {
	DossierCode       string               `json:"dossier_code"`
	ProcedureCode     domain.ProcedureCode `json:"procedure_code"`
	ProductType       domain.ProductType   `json:"product_type"` // "medicament" ou "dispositif_medical"
	MedicineID        *uuid.UUID           `json:"medicine_id,omitempty"`
	MedicalDeviceID   *uuid.UUID           `json:"medical_device_id,omitempty"`
	InternalReference string               `json:"internal_reference"`
	TargetDepositDate *string              `json:"target_deposit_date,omitempty"` // YYYY-MM-DD
	SubmissionNotes   *string              `json:"submission_notes,omitempty"`
}

// CreateDossier crée un nouveau dossier réglementaire et auto-génère sa checklist complète
func (h *DossierHandler) CreateDossier(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	var req CreateDossierRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	// Normalisation
	req.DossierCode = strings.TrimSpace(req.DossierCode)
	req.InternalReference = strings.TrimSpace(req.InternalReference)

	if req.DossierCode == "" {
		writeJSONError(w, http.StatusBadRequest, "Le code dossier (ex: DOS-2026-AMOX-NI) est obligatoire")
		return
	}
	if req.InternalReference == "" {
		req.InternalReference = req.DossierCode
	}

	if req.ProductType == "" {
		if req.ProcedureCode == domain.ProcedureDM {
			req.ProductType = domain.ProductTypeDispositifMedical
		} else {
			req.ProductType = domain.ProductTypeMedicament
		}
	}

	dossier := &domain.Dossier{
		ID:                uuid.New(),
		OrganizationID:    user.OrganizationID,
		MedicineID:        req.MedicineID,
		MedicalDeviceID:   req.MedicalDeviceID,
		ProductType:       req.ProductType,
		DossierCode:       req.DossierCode,
		ProcedureCode:     req.ProcedureCode,
		InternalReference: req.InternalReference,
		SubmissionNotes:   req.SubmissionNotes,
		ResponsibleID:     &user.UserID,
		Status:            domain.DossierStatusInPreparation,
	}

	if req.TargetDepositDate != nil && strings.TrimSpace(*req.TargetDepositDate) != "" {
		t, err := time.Parse("2006-01-02", strings.TrimSpace(*req.TargetDepositDate))
		if err == nil {
			dossier.TargetDepositDate = &t
		}
	}

	// Génération de la Checklist réglementaire selon la typologie officielle ANPP
	var items []domain.ChecklistItem

	if req.ProductType == domain.ProductTypeMedicament {
		if req.MedicineID == nil {
			writeJSONError(w, http.StatusBadRequest, "L'identifiant du médicament (medicine_id) est obligatoire")
			return
		}

		// Récupération de la fiche médicament pour les règles spécifiques (princeps, origine, forme)
		med, err := h.pharmaRepo.GetMedicineByID(r.Context(), user.OrganizationID, *req.MedicineID)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "Médicament sélectionné introuvable dans le référentiel")
			return
		}

		isPrinceps := med.OriginType == domain.OriginPrinceps
		items = domain.GenerateCTDChecklist(req.ProcedureCode, med.OriginType, isPrinceps)
	} else if req.ProductType == domain.ProductTypeDispositifMedical {
		if req.MedicalDeviceID == nil {
			writeJSONError(w, http.StatusBadRequest, "L'identifiant du dispositif médical (medical_device_id) est obligatoire")
			return
		}

		dm, err := h.pharmaRepo.GetMedicalDeviceByID(r.Context(), user.OrganizationID, *req.MedicalDeviceID)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "Dispositif médical sélectionné introuvable")
			return
		}

		items = domain.GenerateDMChecklist(dm.DMClass, dm.IsSterile)
	} else {
		writeJSONError(w, http.StatusBadRequest, "Type de produit non pris en charge")
		return
	}

	created, err := h.repo.CreateDossier(r.Context(), dossier, items)
	if err != nil {
		log.Printf("[ERROR] CreateDossier: %v", err)
		writeJSONError(w, http.StatusBadRequest, fmt.Sprintf("Création du dossier impossible : %v", err))
		return
	}

	// Audit Trail BPF (21 CFR Part 11)
	_, _ = h.auditService.Log(r.Context(), &audit.Entry{
		OrganizationID: user.OrganizationID,
		UserID:         &user.UserID,
		UserEmail:      user.Email,
		UserRole:       user.Role,
		Action:         audit.ActionDossierCreate,
		EntityType:     "dossier",
		EntityID:       created.ID.String(),
		Details: map[string]any{
			"dossier_code":    created.DossierCode,
			"procedure_code":  created.ProcedureCode,
			"items_generated": len(items),
			"product_type":    created.ProductType,
		},
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	})

	writeJSON(w, http.StatusCreated, created)
}

// AttachDocumentRequest payload pour rattacher un document GED
type AttachDocumentRequest struct {
	GEDDocumentID uuid.UUID `json:"ged_document_id"`
	Notes         *string   `json:"notes,omitempty"`
}

// AttachDocument rattache un document validé de la GED à un emplacement CTD de la checklist
func (h *DossierHandler) AttachDocument(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	dossierIDStr := chi.URLParam(r, "id")
	dossierID, err := uuid.Parse(dossierIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant dossier invalide")
		return
	}

	itemIDStr := chi.URLParam(r, "itemId")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant item checklist invalide")
		return
	}

	var req AttachDocumentRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}
	if req.GEDDocumentID == uuid.Nil {
		writeJSONError(w, http.StatusBadRequest, "L'identifiant du document GED est obligatoire")
		return
	}

	item, err := h.repo.AttachGEDDocument(r.Context(), user.OrganizationID, dossierID, itemID, req.GEDDocumentID, req.Notes)
	if err != nil {
		log.Printf("[ERROR] AttachGEDDocument: %v", err)
		if errors.Is(err, repository.ErrDossierLocked) {
			writeJSONError(w, http.StatusForbidden, "Le dossier est verrouillé et ne peut plus être modifié")
			return
		}
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Audit Trail
	_, _ = h.auditService.Log(r.Context(), &audit.Entry{
		OrganizationID: user.OrganizationID,
		UserID:         &user.UserID,
		UserEmail:      user.Email,
		UserRole:       user.Role,
		Action:         audit.ActionDossierItemAttach,
		EntityType:     "dossier_item",
		EntityID:       item.ID.String(),
		Details: map[string]any{
			"dossier_id":      dossierID.String(),
			"item_code":       item.ItemCode,
			"ged_document_id": req.GEDDocumentID.String(),
		},
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	})

	writeJSON(w, http.StatusOK, item)
}

// DetachDocument détache un document GED d'un item de checklist
func (h *DossierHandler) DetachDocument(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	dossierIDStr := chi.URLParam(r, "id")
	dossierID, err := uuid.Parse(dossierIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant dossier invalide")
		return
	}

	itemIDStr := chi.URLParam(r, "itemId")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant item checklist invalide")
		return
	}

	item, err := h.repo.DetachGEDDocument(r.Context(), user.OrganizationID, dossierID, itemID)
	if err != nil {
		if errors.Is(err, repository.ErrDossierLocked) {
			writeJSONError(w, http.StatusForbidden, "Le dossier est verrouillé et ne peut plus être modifié")
			return
		}
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Audit Trail
	_, _ = h.auditService.Log(r.Context(), &audit.Entry{
		OrganizationID: user.OrganizationID,
		UserID:         &user.UserID,
		UserEmail:      user.Email,
		UserRole:       user.Role,
		Action:         audit.ActionDossierItemDetach,
		EntityType:     "dossier_item",
		EntityID:       item.ID.String(),
		Details: map[string]any{
			"dossier_id": dossierID.String(),
			"item_code":  item.ItemCode,
		},
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	})

	writeJSON(w, http.StatusOK, item)
}

// UpdateItemStatusRequest payload pour mettre à jour le statut d'un item
type UpdateItemStatusRequest struct {
	Status domain.ChecklistItemStatus `json:"status"` // "valide_qualite", "approuve_dt", "rejete"
	Notes  *string                    `json:"notes,omitempty"`
}

// UpdateItemStatus met à jour l'état de validation d'un élément de checklist
func (h *DossierHandler) UpdateItemStatus(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	dossierIDStr := chi.URLParam(r, "id")
	dossierID, err := uuid.Parse(dossierIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant dossier invalide")
		return
	}

	itemIDStr := chi.URLParam(r, "itemId")
	itemID, err := uuid.Parse(itemIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant item checklist invalide")
		return
	}

	var req UpdateItemStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	item, err := h.repo.UpdateChecklistItemStatus(r.Context(), user.OrganizationID, dossierID, itemID, req.Status, req.Notes)
	if err != nil {
		if errors.Is(err, repository.ErrDossierLocked) {
			writeJSONError(w, http.StatusForbidden, "Le dossier est verrouillé et ne peut plus être modifié")
			return
		}
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Audit Trail
	_, _ = h.auditService.Log(r.Context(), &audit.Entry{
		OrganizationID: user.OrganizationID,
		UserID:         &user.UserID,
		UserEmail:      user.Email,
		UserRole:       user.Role,
		Action:         audit.ActionDossierItemStatus,
		EntityType:     "dossier_item",
		EntityID:       item.ID.String(),
		Details: map[string]any{
			"dossier_id": dossierID.String(),
			"item_code":  item.ItemCode,
			"new_status": item.Status,
		},
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	})

	writeJSON(w, http.StatusOK, item)
}

// SignPDTRequest payload pour le visa électronique du Directeur Technique
type SignPDTRequest struct {
	Password string `json:"password"`
	TOTPCode string `json:"totp_code,omitempty"`
}

// SignPDT applique le visa 21 CFR Part 11 du Pharmacien Directeur Technique et verrouille le dossier
func (h *DossierHandler) SignPDT(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	// Vérification de rôle : Exclusivement réservé au Pharmacien Directeur Technique (PDT - 1.A.7)
	if user.Role != string(domain.RoleDirecteurTechnique) {
		writeJSONError(w, http.StatusForbidden, "Le visa légal et le scellement réglementaire du dossier sont réservés exclusivement au Pharmacien Directeur Technique titulaire de la décision ministérielle (1.A.7)")
		return
	}

	dossierIDStr := chi.URLParam(r, "id")
	dossierID, err := uuid.Parse(dossierIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant dossier invalide")
		return
	}

	var req SignPDTRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	if strings.TrimSpace(req.Password) == "" {
		writeJSONError(w, http.StatusBadRequest, "Le mot de passe de re-authentification 21 CFR Part 11 est obligatoire")
		return
	}

	// Re-authentification de sécurité dans la base de données
	var passwordHash string
	var totpEnabled bool
	var totpSecret *string
	userQuery := `SELECT password_hash, totp_enabled, totp_secret FROM users WHERE id = $1`
	err = h.pool.QueryRow(r.Context(), userQuery, user.UserID).Scan(&passwordHash, &totpEnabled, &totpSecret)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, "Impossible de vérifier l'identité de l'utilisateur")
		return
	}

	validPassword, err := auth.VerifyPassword(req.Password, passwordHash)
	if err != nil || !validPassword {
		_ = h.logPDTSignFail(r, user, dossierID, "mot de passe de signature incorrect")
		writeJSONError(w, http.StatusUnauthorized, "Mot de passe incorrect : signature électronique refusée")
		return
	}

	// Contrôle 2FA si activé
	if totpEnabled && totpSecret != nil {
		if strings.TrimSpace(req.TOTPCode) == "" {
			writeJSONError(w, http.StatusUnauthorized, "Le code à usage unique (2FA TOTP) est obligatoire pour valider la signature électronique")
			return
		}
		if !auth.ValidateTOTPCode(req.TOTPCode, *totpSecret) {
			_ = h.logPDTSignFail(r, user, dossierID, "code 2FA TOTP invalide")
			writeJSONError(w, http.StatusUnauthorized, "Code 2FA invalide : signature électronique refusée")
			return
		}
	}

	// Génération de l'empreinte cryptographique inaltérable SHA-256 (21 CFR Part 11)
	signedAt := time.Now().UTC()
	seal := domain.GeneratePDTSignatureSeal(dossierID, user.UserID, user.Role, signedAt)

	signedDossier, err := h.repo.SignDossierPDT(r.Context(), user.OrganizationID, dossierID, user.UserID, user.Role, seal)
	if err != nil {
		log.Printf("[ERROR] SignDossierPDT: %v", err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Audit Trail ALCOA+
	_, _ = h.auditService.Log(r.Context(), &audit.Entry{
		OrganizationID: user.OrganizationID,
		UserID:         &user.UserID,
		UserEmail:      user.Email,
		UserRole:       user.Role,
		Action:         audit.ActionDossierPDTSign,
		EntityType:     "dossier",
		EntityID:       dossierID.String(),
		Details: map[string]any{
			"dossier_code":    signedDossier.DossierCode,
			"signature_seal":  seal,
			"signed_at":       signedAt.Format(time.RFC3339),
			"pdt_signer_role": user.Role,
			"intent":          "Visa légal de conformité préalable au dépôt ANPP (21 CFR Part 11)",
		},
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	})

	writeJSON(w, http.StatusOK, signedDossier)
}

// UpdateDossierStatusRequest payload pour la transition de statut
type UpdateDossierStatusRequest struct {
	Status                domain.DossierStatus `json:"status"` // "depose_anpp", "approuve_anpp", "rejete_anpp"
	SubmissionNotes       *string              `json:"submission_notes,omitempty"`
	DepositReceiptRef     *string              `json:"deposit_receipt_ref,omitempty"`
	DepositReceiptDate    *string              `json:"deposit_receipt_date,omitempty"` // YYYY-MM-DD
	ANPPOfficialReference *string              `json:"anpp_official_reference,omitempty"`
	ANPPDecisionDate      *string              `json:"anpp_decision_date,omitempty"`
	ANPPValidityEndDate   *string              `json:"anpp_validity_end_date,omitempty"`
}

// UpdateDossierStatus gère les étapes ultérieures du cycle de vie ANPP (dépôt physique/numérique, avis, approbation)
func (h *DossierHandler) UpdateDossierStatus(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	dossierIDStr := chi.URLParam(r, "id")
	dossierID, err := uuid.Parse(dossierIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant dossier invalide")
		return
	}

	var req UpdateDossierStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	var receiptDate *time.Time
	if req.DepositReceiptDate != nil && strings.TrimSpace(*req.DepositReceiptDate) != "" {
		t, err := time.Parse("2006-01-02", strings.TrimSpace(*req.DepositReceiptDate))
		if err == nil {
			receiptDate = &t
		}
	}

	var decisionDate *time.Time
	if req.ANPPDecisionDate != nil && strings.TrimSpace(*req.ANPPDecisionDate) != "" {
		t, err := time.Parse("2006-01-02", strings.TrimSpace(*req.ANPPDecisionDate))
		if err == nil {
			decisionDate = &t
		}
	}

	var validityDate *time.Time
	if req.ANPPValidityEndDate != nil && strings.TrimSpace(*req.ANPPValidityEndDate) != "" {
		t, err := time.Parse("2006-01-02", strings.TrimSpace(*req.ANPPValidityEndDate))
		if err == nil {
			validityDate = &t
		}
	}

	updated, err := h.repo.UpdateDossierStatus(
		r.Context(), user.OrganizationID, dossierID,
		req.Status, req.SubmissionNotes,
		req.DepositReceiptRef, receiptDate,
		req.ANPPOfficialReference, decisionDate, validityDate,
	)
	if err != nil {
		log.Printf("[ERROR] UpdateDossierStatus: %v", err)
		writeJSONError(w, http.StatusBadRequest, err.Error())
		return
	}

	// Audit Trail
	_, _ = h.auditService.Log(r.Context(), &audit.Entry{
		OrganizationID: user.OrganizationID,
		UserID:         &user.UserID,
		UserEmail:      user.Email,
		UserRole:       user.Role,
		Action:         audit.ActionDossierStatus,
		EntityType:     "dossier",
		EntityID:       dossierID.String(),
		Details: map[string]any{
			"new_status":  req.Status,
			"receipt_ref": req.DepositReceiptRef,
		},
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	})

	writeJSON(w, http.StatusOK, updated)
}

func (h *DossierHandler) logPDTSignFail(r *http.Request, user *auth.SessionUser, dossierID uuid.UUID, reason string) error {
	_, err := h.auditService.Log(r.Context(), &audit.Entry{
		OrganizationID: user.OrganizationID,
		UserID:         &user.UserID,
		UserEmail:      user.Email,
		UserRole:       user.Role,
		Action:         audit.ActionDossierPDTSign + "_FAILED",
		EntityType:     "dossier",
		EntityID:       dossierID.String(),
		Details: map[string]any{
			"reason": reason,
		},
		IPAddress: getClientIP(r),
		UserAgent: r.UserAgent(),
	})
	return err
}
