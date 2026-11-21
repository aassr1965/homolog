package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"homolog/internal/audit"
	"homolog/internal/domain"
	"homolog/internal/repository"
)

// FeeHandler gère les barèmes des redevances ANPP et l'émission des bordereaux BV1/BV2
type FeeHandler struct {
	feeRepo      *repository.FeeRepository
	dossierRepo  *repository.DossierRepository
	auditService *audit.Service
	pool         *pgxpool.Pool
}

// NewFeeHandler instancie le gestionnaire d'API des redevances
func NewFeeHandler(
	feeRepo *repository.FeeRepository,
	dossierRepo *repository.DossierRepository,
	auditService *audit.Service,
	pool *pgxpool.Pool,
) *FeeHandler {
	return &FeeHandler{
		feeRepo:      feeRepo,
		dossierRepo:  dossierRepo,
		auditService: auditService,
		pool:         pool,
	}
}

// ListFeeSchedules retourne les barèmes de redevances applicables
func (h *FeeHandler) ListFeeSchedules(w http.ResponseWriter, r *http.Request) {
	onlyActive := r.URL.Query().Get("active") == "true"
	schedules, err := h.feeRepo.ListFeeSchedules(r.Context(), onlyActive)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Erreur récupération barèmes: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"schedules": schedules,
		"total":     len(schedules),
	})
}

// CreateOrUpdateFeeSchedule enregistre un barème mis à jour (nouvelle loi de finances / arrêté)
func (h *FeeHandler) CreateOrUpdateFeeSchedule(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	if !user.HasPermission(string(domain.PermUserManage)) && !user.HasPermission(string(domain.PermRefEdit)) {
		writeJSONError(w, http.StatusForbidden, "Habilitation insuffisante : la mise à jour des barèmes légaux est réservée aux administrateurs et responsables réglementaires")
		return
	}

	var req domain.CreateFeeScheduleRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	req.ProcedureCode = strings.TrimSpace(req.ProcedureCode)
	req.LegalReference = strings.TrimSpace(req.LegalReference)
	req.LabelFR = strings.TrimSpace(req.LabelFR)

	if req.ProcedureCode == "" || req.LegalReference == "" || req.LabelFR == "" {
		writeJSONError(w, http.StatusBadRequest, "Le code de procédure, la référence légale (JORADP) et le libellé sont obligatoires")
		return
	}

	if req.AmountDZD <= 0 {
		writeJSONError(w, http.StatusBadRequest, "Le montant de la redevance en DZD doit être supérieur à zéro")
		return
	}

	var validFrom time.Time
	var err error
	if req.ValidFrom != "" {
		validFrom, err = time.Parse("2006-01-02", req.ValidFrom)
		if err != nil {
			writeJSONError(w, http.StatusBadRequest, "Format de date d'effet invalide (attendu: AAAA-MM-JJ)")
			return
		}
	} else {
		validFrom = time.Now().UTC()
	}

	schedule := domain.ANPPFeeSchedule{
		Category:       strings.TrimSpace(req.Category),
		ProcedureCode:  req.ProcedureCode,
		LabelFR:        req.LabelFR,
		LabelAR:        strings.TrimSpace(req.LabelAR),
		LegalReference: req.LegalReference,
		AmountDZD:      req.AmountDZD,
		ValidFrom:      validFrom,
		Notes:          strings.TrimSpace(req.Notes),
	}
	if schedule.Category == "" {
		schedule.Category = "medicament"
	}

	created, err := h.feeRepo.CreateOrUpdateFeeSchedule(r.Context(), &schedule)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Erreur enregistrement barème: %v", err))
		return
	}

	// Audit Trail ALCOA+
	if h.auditService != nil {
		_, _ = h.auditService.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         audit.ActionFeeScheduleUpdate,
			EntityType:     "anpp_fee_schedules",
			EntityID:       created.ID,
			Details: map[string]any{
				"procedure_code":  created.ProcedureCode,
				"amount_dzd":      created.AmountDZD,
				"legal_reference": created.LegalReference,
				"valid_from":      created.ValidFrom.Format("2006-01-02"),
			},
			IPAddress: r.RemoteAddr,
			UserAgent: r.UserAgent(),
		})
	}

	writeJSON(w, http.StatusCreated, created)
}

// ListDossierPaymentSlips liste les bordereaux émis pour un dossier
func (h *FeeHandler) ListDossierPaymentSlips(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	dossierIDStr := chi.URLParam(r, "id")
	if _, err := uuid.Parse(dossierIDStr); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant dossier invalide")
		return
	}

	slips, err := h.feeRepo.ListPaymentSlipsByDossier(r.Context(), user.OrganizationID.String(), dossierIDStr)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Erreur récupération bordereaux: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"dossier_id": dossierIDStr,
		"slips":      slips,
		"total":      len(slips),
	})
}

// GenerateDossierPaymentSlip génère un bordereau officiel BV1 ou BV2 selon le barème applicable
func (h *FeeHandler) GenerateDossierPaymentSlip(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	dossierIDStr := chi.URLParam(r, "id")
	dossierUUID, err := uuid.Parse(dossierIDStr)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant dossier invalide")
		return
	}

	var req domain.GeneratePaymentSlipRequest
	_ = json.NewDecoder(r.Body).Decode(&req)

	slipType := strings.ToUpper(strings.TrimSpace(req.SlipType))
	if slipType == "" {
		slipType = domain.SlipTypeBV1
	}
	if slipType != domain.SlipTypeBV1 && slipType != domain.SlipTypeBV2 {
		writeJSONError(w, http.StatusBadRequest, "Type de bordereau invalide (attendu: BV1 ou BV2)")
		return
	}

	// Récupérer le dossier pour identifier la procédure et le type de produit
	dossier, err := h.dossierRepo.GetDossierByID(r.Context(), user.OrganizationID, dossierUUID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Dossier réglementaire introuvable")
		return
	}

	// Déterminer le code de barème applicable
	targetProcedure := "generique_local"
	paymentPurpose := req.PaymentPurpose

	if dossier.ProductType == "dispositif_medical" || dossier.ProductType == "medical_device" {
		dmClass := "IIa"
		if dossier.MedicalDevice != nil && dossier.MedicalDevice.DMClass != "" {
			dmClass = dossier.MedicalDevice.DMClass
		}
		switch dmClass {
		case "I":
			targetProcedure = "dm_classe_I"
		case "IIa":
			targetProcedure = "dm_classe_IIa"
		case "IIb", "III":
			targetProcedure = "dm_classe_IIb_III"
		case "DMDIV":
			targetProcedure = "dm_div"
		default:
			targetProcedure = "dm_classe_IIa"
		}
		if paymentPurpose == "" {
			paymentPurpose = fmt.Sprintf("Redevance d'homologation Dispositif Médical Classe %s (Décret 20-324)", dmClass)
		}
	} else {
		// Médicaments
		procCode := strings.ToUpper(string(dossier.ProcedureCode))
		switch procCode {
		case "PI", "PRINCEPS":
			targetProcedure = "princeps"
		case "NF", "NL":
			targetProcedure = "generique_local"
		case "NI":
			targetProcedure = "generique_importe"
		case "RI", "RF", "RENOUVELLEMENT":
			targetProcedure = "renouvellement"
		case "VAR_MAJ":
			targetProcedure = "variation_majeure"
		case "VAR_MIN":
			targetProcedure = "variation_mineure"
		default:
			targetProcedure = "generique_local"
		}
		if paymentPurpose == "" {
			paymentPurpose = fmt.Sprintf("Redevance d'enregistrement ANPP pour dossier %s (Procédure %s)", dossier.DossierCode, dossier.ProcedureCode)
		}
	}

	// Récupération du barème valide à la date du jour
	feeSchedule, err := h.feeRepo.GetFeeScheduleForProcedure(r.Context(), targetProcedure, time.Now().UTC())
	if err != nil {
		// Repli sur le premier barème actif
		schedules, _ := h.feeRepo.ListFeeSchedules(r.Context(), true)
		if len(schedules) > 0 {
			feeSchedule = &schedules[0]
		} else {
			writeJSONError(w, http.StatusInternalServerError, "Aucun barème de redevance actif configuré")
			return
		}
	}

	amount := feeSchedule.AmountDZD
	if req.CustomAmount != nil && *req.CustomAmount > 0 {
		amount = *req.CustomAmount
	}

	// Conversion en toutes lettres français & arabe
	wordsFR, wordsAR := domain.ConvertAmountDZDToWords(amount)

	// Numéro de bordereau séquentiel horodaté
	slipNumber := fmt.Sprintf("%s-%d-%s", slipType, time.Now().Year(), uuid.New().String()[:8])
	userIDStr := user.UserID.String()

	slip := domain.ANPPPaymentSlip{
		OrganizationID: user.OrganizationID.String(),
		DossierID:      dossierIDStr,
		FeeScheduleID:  feeSchedule.ID,
		SlipType:       slipType,
		SlipNumber:     slipNumber,
		PaymentPurpose: paymentPurpose,
		AmountDZD:      amount,
		AmountWordsFR:  wordsFR,
		AmountWordsAR:  wordsAR,
		BankAccountRIB: "00100624030000001234 (Compte Trésorerie Centrale ANPP Alger)",
		Status:         domain.PaymentStatusEmis,
		GeneratedBy:    &userIDStr,
		GeneratedAt:    time.Now().UTC(),
	}

	created, err := h.feeRepo.CreatePaymentSlip(r.Context(), &slip)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Génération bordereau impossible: %v", err))
		return
	}

	// Audit Trail ALCOA+
	if h.auditService != nil {
		_, _ = h.auditService.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         audit.ActionFeeSlipGenerated,
			EntityType:     "anpp_payment_slips",
			EntityID:       created.ID,
			Details: map[string]any{
				"slip_type":   created.SlipType,
				"slip_number": created.SlipNumber,
				"amount_dzd":  created.AmountDZD,
				"dossier_id":  dossierIDStr,
			},
			IPAddress: r.RemoteAddr,
			UserAgent: r.UserAgent(),
		})
	}

	enriched, _ := h.feeRepo.GetPaymentSlip(r.Context(), user.OrganizationID.String(), created.ID)
	writeJSON(w, http.StatusCreated, enriched)
}

// GetPaymentSlip retourne un bordereau par son ID
func (h *FeeHandler) GetPaymentSlip(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	slipIDStr := chi.URLParam(r, "slipId")
	if _, err := uuid.Parse(slipIDStr); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant bordereau invalide")
		return
	}

	slip, err := h.feeRepo.GetPaymentSlip(r.Context(), user.OrganizationID.String(), slipIDStr)
	if err != nil {
		if errors.Is(err, repository.ErrPaymentSlipNotFound) {
			writeJSONError(w, http.StatusNotFound, "Bordereau de versement introuvable")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Erreur récupération bordereau: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, slip)
}

// UpdatePaymentSlipStatus met à jour le statut du bordereau (marquage payé, quittance)
func (h *FeeHandler) UpdatePaymentSlipStatus(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	slipIDStr := chi.URLParam(r, "slipId")
	if _, err := uuid.Parse(slipIDStr); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant bordereau invalide")
		return
	}

	var req domain.UpdatePaymentSlipStatusRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	var paidAt *time.Time
	if req.PaidAt != "" {
		if t, err := time.Parse("2006-01-02", req.PaidAt); err == nil {
			paidAt = &t
		}
	} else if req.Status == domain.PaymentStatusPaye || req.Status == domain.PaymentStatusQuittanceRecue {
		now := time.Now().UTC()
		paidAt = &now
	}

	err := h.feeRepo.UpdatePaymentSlipStatus(
		r.Context(),
		user.OrganizationID.String(),
		slipIDStr,
		req.Status,
		strings.TrimSpace(req.PaymentReference),
		paidAt,
		req.ReceiptDocumentID,
	)
	if err != nil {
		if errors.Is(err, repository.ErrPaymentSlipNotFound) {
			writeJSONError(w, http.StatusNotFound, "Bordereau introuvable")
			return
		}
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Erreur mise à jour bordereau: %v", err))
		return
	}

	// Audit Trail ALCOA+
	if h.auditService != nil {
		_, _ = h.auditService.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         audit.ActionFeeSlipStatusUpdate,
			EntityType:     "anpp_payment_slips",
			EntityID:       slipIDStr,
			Details: map[string]any{
				"status":            req.Status,
				"payment_reference": req.PaymentReference,
			},
			IPAddress: r.RemoteAddr,
			UserAgent: r.UserAgent(),
		})
	}

	updated, _ := h.feeRepo.GetPaymentSlip(r.Context(), user.OrganizationID.String(), slipIDStr)
	writeJSON(w, http.StatusOK, updated)
}

// GetPaymentSlipPrintHTML génère la version imprimable officielle du bordereau BV1 ou BV2
func (h *FeeHandler) GetPaymentSlipPrintHTML(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		http.Error(w, "Authentification requise", http.StatusUnauthorized)
		return
	}

	slipIDStr := chi.URLParam(r, "slipId")
	slip, err := h.feeRepo.GetPaymentSlip(r.Context(), user.OrganizationID.String(), slipIDStr)
	if err != nil {
		http.Error(w, "Bordereau introuvable", http.StatusNotFound)
		return
	}

	// Format procedure label
	procLabel := slip.ProcedureCode
	switch slip.ProcedureCode {
	case "NI":
		procLabel = "NI — Nouvelle Demande d'Enregistrement (Spécialité Importée)"
	case "NF":
		procLabel = "NF — Nouvelle Demande d'Enregistrement (Fabrication Locale)"
	case "MI":
		procLabel = "MI — Modification Majeure Post-Enregistrement (Spécialité Importée)"
	case "MF":
		procLabel = "MF — Modification Majeure Post-Enregistrement (Fabrication Locale)"
	case "RI":
		procLabel = "RI — Renouvellement Quinquennal (Spécialité Importée)"
	case "RF":
		procLabel = "RF — Renouvellement Quinquennal (Fabrication Locale)"
	default:
		if procLabel == "" {
			procLabel = "Enregistrement Initial (Décret Exécutif 20-325)"
		}
	}

	// Format slip title & labels
	slipTypeLabel := "BV1 — DÉPÔT INITIAL"
	slipTypeLabelAR := "الجدول رقم 1 (إيداع أولي)"
	if slip.SlipType == "BV2" {
		slipTypeLabel = "BV2 — COMPLÉMENT & EXPERTISE LNCPP"
		slipTypeLabelAR = "الجدول رقم 2 (تكميلي / تحاليل مخبرية)"
	}

	// Watermark text & class
	wmText := "DOCUMENT DE LIQUIDATION OFFICIEL"
	wmClass := "watermark-neutral"
	if slip.VerificationStatus == "conforme" {
		wmText = "ACQUITTÉ & CONFORME ANPP (VISÉ DT)"
		wmClass = "watermark-success"
	} else if slip.Status == "paye" || slip.Status == "quittance_recue" {
		wmText = "QUITTANCE TRÉSOR ENREGISTRÉE"
		wmClass = "watermark-info"
	}

	// Formats pharmaceutiques enrichis
	dciText := slip.ProductDCI
	if dciText == "" {
		dciText = "Substance active déclarée au dossier technique"
	}

	dosageText := slip.ProductDosage
	if dosageText == "" {
		dosageText = "Dosage standard"
	}

	formText := slip.ProductForm
	if formText == "" {
		formText = "Forme pharmaceutique conventionnelle"
	}

	routeText := slip.ProductRoute
	if routeText == "" {
		routeText = "Selon RCP approuvé"
	}

	miphLic := slip.MIPHLicenseNumber
	if miphLic == "" {
		miphLic = "Agrément MIPH conforme Décret Exécutif 20-325"
	}

	dtName := slip.TechnicalDirectorName
	if dtName == "" {
		dtName = "Pharmacien Directeur Technique Titulaire (Mandat 1.A.7)"
	}

	estName := slip.EstablishmentName
	if estName == "" {
		estName = slip.OrganizationName
	}

	genByName := slip.GeneratedByName
	if genByName == "" {
		genByName = "Département Affaires Réglementaires & Financières"
	}

	// Construction du bloc de change BA si applicable
	fxBlockHTML := ""
	if slip.RequiresFXAttestation || slip.Currency != "DZD" {
		fxRateStr := "En attente du cours officiel"
		if slip.FXRate != nil && *slip.FXRate > 0 {
			rateDateStr := ""
			if slip.FXRateDate != nil {
				rateDateStr = " au " + slip.FXRateDate.Format("02/01/2006")
			}
			fxRateStr = fmt.Sprintf("1 %s = %.4f DZD%s", slip.Currency, *slip.FXRate, rateDateStr)
		}

		currAmtStr := "—"
		if slip.CurrencyAmount != nil && *slip.CurrencyAmount > 0 {
			currAmtStr = fmt.Sprintf("%.2f %s", *slip.CurrencyAmount, slip.Currency)
		}

		attestRef := slip.FXAttestationRef
		if attestRef == "" {
			attestRef = "Attestation de cours officiel requise (1.A.4)"
		}

		fxBlockHTML = fmt.Sprintf(`
        <div style="background: #fdf4ff; border: 1.5px solid #d946ef; border-radius: 4px; padding: 8px 12px; margin-bottom: 12px; font-size: 8pt;">
            <div style="font-weight: 800; color: #86198f; margin-bottom: 3px; display: flex; justify-content: space-between; align-items: center;">
                <span>💱 Contrôle du Taux de Change Officiel — Banque d'Algérie (Dossier Importé)</span>
                <span style="font-size: 7.5pt; background: #fae8ff; padding: 2px 6px; border-radius: 3px; color: #a21caf;">Volet 1.A.4</span>
            </div>
            <div style="display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 8px; color: #701a75;">
                <div><strong>Devise &amp; Montant :</strong> %s</div>
                <div><strong>Cours Banque d'Algérie :</strong> %s</div>
                <div><strong>Réf Attestation BA :</strong> %s</div>
            </div>
        </div>
		`, currAmtStr, fxRateStr, escapeHTMLStr(attestRef))
	}

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="fr" dir="ltr">
<head>
    <meta charset="UTF-8">
    <title>Bordereau de Versement %s — %s — ANPP</title>
    <style>
        @page { size: A4 portrait; margin: 0; }
        *, *::before, *::after { box-sizing: border-box; }
        body { 
            font-family: 'Segoe UI', -apple-system, BlinkMacSystemFont, Roboto, sans-serif; 
            color: #0f172a; 
            background-color: #f1f5f9; 
            margin: 0; 
            padding: 24px 0; 
            font-size: 8.8pt; 
            line-height: 1.42; 
            -webkit-print-color-adjust: exact; 
            print-color-adjust: exact; 
        }
        .no-print-toolbar { 
            position: fixed; 
            top: 0; 
            left: 0; 
            right: 0; 
            height: 52px; 
            background: #091e36; 
            color: #ffffff; 
            display: flex; 
            align-items: center; 
            justify-content: space-between; 
            padding: 0 24px; 
            z-index: 9999; 
            box-shadow: 0 4px 12px rgba(0,0,0,0.15); 
            font-size: 13px; 
        }
        .no-print-toolbar .btn { 
            display: inline-flex; 
            align-items: center; 
            gap: 8px; 
            padding: 7px 16px; 
            border-radius: 6px; 
            font-weight: 700; 
            font-size: 12px; 
            cursor: pointer; 
            border: none; 
            text-decoration: none; 
        }
        .btn-primary { background: #0284c7; color: #ffffff; }
        .btn-outline { background: rgba(255,255,255,0.12); color: #f8fafc; border: 1px solid rgba(255,255,255,0.25); }
        .sheet { 
            width: 210mm; 
            min-height: 297mm; 
            margin: 48px auto 24px auto; 
            background: #ffffff; 
            box-shadow: 0 8px 30px rgba(15,23,42,0.12); 
            padding: 12mm 15mm; 
            position: relative; 
            overflow: hidden; 
        }
        @media print { 
            body { background: #ffffff; padding: 0; margin: 0; } 
            .no-print-toolbar { display: none !important; } 
            .sheet { width: 100%%; margin: 0; padding: 10mm 14mm; box-shadow: none; border-radius: 0; } 
        }
        .watermark { 
            position: absolute; 
            top: 48%%; 
            left: 50%%; 
            transform: translate(-50%%, -50%%) rotate(-35deg); 
            font-size: 38pt; 
            font-weight: 900; 
            letter-spacing: 4px; 
            pointer-events: none; 
            user-select: none; 
            z-index: 1; 
            white-space: nowrap; 
            text-transform: uppercase; 
        }
        .watermark-neutral { color: rgba(100,116,139,0.06); }
        .watermark-success { color: rgba(16,185,129,0.08); }
        .watermark-info { color: rgba(37,99,235,0.07); }

        .header-table { 
            width: 100%%; 
            border-collapse: collapse; 
            margin-bottom: 10px; 
            border-bottom: 2px solid #091e36; 
            padding-bottom: 8px; 
        }
        .title-banner { 
            background: linear-gradient(135deg, #091e36 0%%, #0f3b60 100%%); 
            color: #ffffff; 
            padding: 10px 14px; 
            border-radius: 4px; 
            margin-bottom: 12px; 
            box-shadow: 0 2px 4px rgba(0,0,0,0.1); 
        }
        .banner-title { 
            font-size: 11pt; 
            font-weight: 900; 
            letter-spacing: 0.5px; 
            text-transform: uppercase; 
            margin: 0; 
            color: #ffffff; 
        }
        .banner-sub { 
            font-size: 8.5pt; 
            color: #93c5fd; 
            margin-top: 3px; 
            font-weight: 600; 
        }
        .banner-legal { 
            font-size: 7.2pt; 
            color: #cbd5e1; 
            margin-top: 4px; 
            border-top: 1px dashed rgba(255,255,255,0.2); 
            padding-top: 3px; 
        }
        .info-card { 
            border: 1px solid #cbd5e1; 
            border-radius: 4px; 
            padding: 8px 12px; 
            background: #ffffff; 
            margin-bottom: 10px; 
        }
        .info-card-header { 
            font-size: 8.2pt; 
            font-weight: 800; 
            color: #091e36; 
            text-transform: uppercase; 
            letter-spacing: 0.5px; 
            border-bottom: 1px solid #e2e8f0; 
            padding-bottom: 4px; 
            margin-bottom: 6px; 
            display: flex; 
            align-items: center; 
            justify-content: space-between; 
        }
        .field-grid { 
            display: grid; 
            grid-template-columns: 1fr 1fr; 
            gap: 4px 16px; 
        }
        .field-row { 
            display: flex; 
            font-size: 8.2pt; 
            line-height: 1.35; 
        }
        .field-label { 
            width: 145px; 
            font-weight: 700; 
            color: #475569; 
            flex-shrink: 0; 
        }
        .field-value { 
            color: #091e36; 
            font-weight: 600; 
            flex-grow: 1; 
        }
        .data-table { 
            width: 100%%; 
            border-collapse: collapse; 
            margin-bottom: 10px; 
            font-size: 8.2pt; 
        }
        .data-table th { 
            background: #0f3b60; 
            color: #ffffff; 
            padding: 6px 10px; 
            text-align: left; 
            font-weight: 700; 
            font-size: 8pt; 
            border: 1px solid #0f3b60; 
        }
        .data-table td { 
            padding: 6px 10px; 
            border: 1px solid #cbd5e1; 
        }
        .amount-in-words-box { 
            border: 1.5px solid #047857; 
            background: #f0fdf4; 
            border-radius: 4px; 
            padding: 9px 12px; 
            margin-bottom: 10px; 
            color: #065f46; 
            font-size: 8.5pt; 
        }
        .rib-box { 
            border: 1.5px solid #091e36; 
            background: #f8fafc; 
            border-radius: 4px; 
            padding: 9px 12px; 
            margin-bottom: 10px; 
            font-size: 8.2pt; 
        }
        .rib-number { 
            font-family: monospace; 
            font-size: 11pt; 
            font-weight: 900; 
            letter-spacing: 2px; 
            color: #091e36; 
            background: #ffffff; 
            padding: 4px 10px; 
            border: 1.5px solid #0284c7; 
            border-radius: 3px; 
            display: inline-block; 
            margin: 3px 0; 
        }
        .sign-grid-3 { 
            display: grid; 
            grid-template-columns: 1fr 1fr 1fr; 
            gap: 10px; 
            margin-top: 10px; 
            page-break-inside: avoid; 
        }
        .sign-card { 
            border: 1px solid #091e36; 
            border-radius: 4px; 
            padding: 8px; 
            background: #ffffff; 
        }
        .sign-header { 
            font-weight: 800; 
            font-size: 7.5pt; 
            color: #091e36; 
            text-transform: uppercase; 
            margin-bottom: 3px; 
        }
        .sign-space { 
            height: 62px; 
            border: 1px dashed #cbd5e1; 
            background: #f8fafc; 
            border-radius: 3px; 
            display: flex; 
            align-items: center; 
            justify-content: center; 
            font-size: 7.2pt; 
            color: #64748b; 
            margin-top: 4px; 
            text-align: center; 
            padding: 4px; 
        }
        .footer-legal { 
            margin-top: 12px; 
            padding-top: 6px; 
            border-top: 1px solid #cbd5e1; 
            display: flex; 
            align-items: center; 
            justify-content: space-between; 
            font-size: 6.8pt; 
            color: #64748b; 
            page-break-inside: avoid; 
        }
    </style>
</head>
<body>
    <div class="no-print-toolbar">
        <div>
            <strong>Homolog Reporting BPF</strong> • Bordereau Officiel de Versement (%s) — Réf : %s
        </div>
        <div style="display:flex; gap:10px;">
            <button class="btn btn-primary" onclick="window.print()">🖨️ Imprimer / Export PDF (A4 300 DPI)</button>
            <button type="button" class="btn btn-outline" onclick="window.close()">✕ Fermer</button>
        </div>
    </div>

    <div class="sheet">
        <div class="watermark %s">%s</div>

        <table class="header-table">
            <tr>
                <td style="width:55px; vertical-align:middle;">
                    <!-- Emblème Officiel National / Caducée ANPP -->
                    <svg width="48" height="48" viewBox="0 0 100 100" fill="none">
                        <circle cx="50" cy="50" r="46" stroke="#091e36" stroke-width="4" fill="#f8fafc"/>
                        <path d="M50 18v64M30 38h40M30 38l-10 20h20l-10-20zM70 38l-10 20h20l-10-20zM40 82h20" stroke="#047857" stroke-width="3.5" stroke-linecap="round" stroke-linejoin="round"/>
                        <circle cx="50" cy="24" r="5" fill="#dc2626"/>
                    </svg>
                </td>
                <td style="vertical-align:middle; padding-left:12px;">
                    <div style="font-size:7.2pt; font-weight:700; color:#64748b; text-transform:uppercase;">
                        RÉPUBLIQUE ALGÉRIENNE DÉMOCRATIQUE ET POPULAIRE
                    </div>
                    <div style="font-size:7.8pt; font-weight:700; color:#334155;" dir="rtl">
                        الجمهورية الجزائرية الديمقراطية الشعبية
                    </div>
                    <div style="font-size:8.2pt; font-weight:800; color:#091e36; margin-top:1px;">
                        MINISTÈRE DE L'INDUSTRIE ET DE LA PRODUCTION PHARMACEUTIQUE
                    </div>
                    <div style="font-size:8pt; font-weight:800; color:#047857;">
                        AGENCE NATIONALE DES PRODUITS PHARMACEUTIQUES (ANPP)
                    </div>
                    <div style="font-size:6.8pt; color:#475569;">
                        Direction de l'Enregistrement des Produits Pharmaceutiques (DEPP) • Sous-Direction des Redevances &amp; Tarifications
                    </div>
                </td>
                <td style="text-align:right; vertical-align:middle; width:190px;">
                    <div style="font-size:6.8pt; color:#64748b; text-transform:uppercase;">BORDEREAU OFFICIEL D'ÉTAT</div>
                    <div style="font-family:monospace; font-size:11pt; font-weight:900; color:#091e36; letter-spacing:1.5px;">*%s*</div>
                    <div style="font-size:7.5pt; color:#047857; font-weight:800;">N° : %s</div>
                    <div style="font-size:6.8pt; color:#64748b;">Date d'Émission : %s</div>
                </td>
            </tr>
        </table>

        <div class="title-banner">
            <div style="display:flex; justify-content:space-between; align-items:flex-start;">
                <div>
                    <h1 class="banner-title">BORDEREAU DE LIQUIDATION &amp; DE PAIEMENT DE LA REDEVANCE — %s</h1>
                    <div class="banner-sub" dir="rtl">%s</div>
                </div>
                <div style="text-align:right;">
                    <span style="font-size:8pt; font-weight:800; background:rgba(255,255,255,0.2); padding:3px 8px; border-radius:3px;">
                        DOSSIER : %s
                    </span>
                </div>
            </div>
            <div class="banner-legal">
                Base Réglementaire : Décret Exécutif n° 20-325 du 22 Novembre 2020 &amp; Arrêté Interministériel fixant le barème officiel des redevances dues à l'ANPP (Journal Officiel).
            </div>
        </div>

        <!-- 1. Identification de l'Établissement Déposant -->
        <div class="info-card">
            <div class="info-card-header">
                <span>1. Identification de l'Établissement Pharmaceutique Déposant (Assujetti)</span>
                <span style="font-size:7.5pt; font-weight:700; color:#047857;">Section 1.A.4 &amp; 1.A.7</span>
            </div>
            <div class="field-grid">
                <div class="field-row">
                    <span class="field-label">Raison Sociale :</span>
                    <span class="field-value">%s</span>
                </div>
                <div class="field-row">
                    <span class="field-label">N° Agrément MIPH :</span>
                    <span class="field-value">%s</span>
                </div>
                <div class="field-row">
                    <span class="field-label">Pharmacien DT Déposant :</span>
                    <span class="field-value">%s</span>
                </div>
                <div class="field-row">
                    <span class="field-label">Établissement / Site :</span>
                    <span class="field-value">%s</span>
                </div>
                <div class="field-row">
                    <span class="field-label">Agent Déposant :</span>
                    <span class="field-value">%s</span>
                </div>
                <div class="field-row">
                    <span class="field-label">Statut d'Encaissement :</span>
                    <span class="field-value" style="color:%s; font-weight:800;">%s</span>
                </div>
            </div>
        </div>

        <!-- 2. Identification de la Spécialité Pharmaceutique -->
        <div class="info-card">
            <div class="info-card-header">
                <span>2. Identification Détaillée de la Spécialité Pharmaceutique</span>
                <span style="font-size:7.5pt; font-weight:700; color:#0f3b60;">Module 1 CTD</span>
            </div>
            <div class="field-grid">
                <div class="field-row">
                    <span class="field-label">Dénomination Commerciale :</span>
                    <span class="field-value" style="font-size:9pt; font-weight:800; color:#091e36;">%s</span>
                </div>
                <div class="field-row">
                    <span class="field-label">Substance Active (DCI) :</span>
                    <span class="field-value">%s</span>
                </div>
                <div class="field-row">
                    <span class="field-label">Dosage &amp; Forme Galénique :</span>
                    <span class="field-value">%s • %s</span>
                </div>
                <div class="field-row">
                    <span class="field-label">Voie d'Administration :</span>
                    <span class="field-value">%s</span>
                </div>
                <div class="field-row">
                    <span class="field-label">Procédure Réglementaire :</span>
                    <span class="field-value" style="color:#0284c7; font-weight:700;">%s</span>
                </div>
                <div class="field-row">
                    <span class="field-label">N° Dossier CTD :</span>
                    <span class="field-value" style="font-family:monospace; font-weight:800;">%s</span>
                </div>
            </div>
        </div>

        <!-- 3. Liquidation Détaillée de la Redevance Légale -->
        <table class="data-table">
            <thead>
                <tr>
                    <th style="width:14%%; text-align:center;">Code Barème</th>
                    <th style="width:48%%;">Désignation Légale de la Redevance ANPP</th>
                    <th style="width:18%%; text-align:center;">Base Juridique (JORADP)</th>
                    <th style="width:20%%; text-align:right;">Montant Net Exigible (DZD)</th>
                </tr>
            </thead>
            <tbody>
                <tr>
                    <td style="text-align:center; font-family:monospace; font-weight:700;">%s</td>
                    <td>
                        <strong style="color:#091e36;">%s</strong><br/>
                        <span style="font-size:7.5pt; color:#64748b;">Objet : %s</span>
                    </td>
                    <td style="text-align:center; font-size:7.5pt;">
                        %s
                    </td>
                    <td style="text-align:right; font-size:10pt; font-weight:800; color:#091e36;">
                        %.2f DZD
                    </td>
                </tr>
                <tr style="background:#f8fafc; font-weight:800; font-size:8.8pt;">
                    <td colspan="3" style="text-align:right; color:#091e36;">
                        TOTAL LIQUIDÉ À ACQUITTER AU TRÉSOR PUBLIC :
                    </td>
                    <td style="text-align:right; color:#047857; font-size:10pt;">
                        %.2f DZD
                    </td>
                </tr>
            </tbody>
        </table>

        <!-- 4. Arrêté du Montant en Toutes Lettres -->
        <div class="amount-in-words-box">
            <div style="display:flex; justify-content:space-between; align-items:flex-start;">
                <div>
                    <div style="font-size:7.5pt; font-weight:700; text-transform:uppercase; color:#047857;">Arrêté du présent bordereau à la somme nette de :</div>
                    <div style="font-size:9.5pt; font-weight:800; color:#065f46; margin-top:2px;">« %s »</div>
                </div>
                <div style="text-align:right;" dir="rtl">
                    <div style="font-size:7.5pt; font-weight:700; color:#047857;">أوقف جدول الدفع الحالي عند مبلغ صافي قدره :</div>
                    <div style="font-size:10pt; font-weight:800; color:#065f46; margin-top:2px;">« %s »</div>
                </div>
            </div>
        </div>

        %s

        <!-- 5. Modalités de Règlement & Compte Trésorerie ANPP -->
        <div class="rib-box">
            <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:3px;">
                <div style="font-weight:800; color:#091e36; font-size:8.2pt;">Coordonnées Bancaires Officielles du Trésor Public (ANPP) :</div>
                <span style="font-size:7.2pt; font-weight:700; color:#047857; background:#dcfce7; padding:2px 6px; border-radius:3px;">Compte Trésorerie Centrale</span>
            </div>
            <div style="color:#475569; font-size:7.8pt;">
                Bénéficiaire : <strong>Agent Comptable de l'Agence Nationale des Produits Pharmaceutiques</strong><br/>
                Banque Domiciliataire : <strong>Banque Nationale d'Algérie (BNA) • Agence Trésor Alger</strong>
            </div>
            <div>
                <span style="font-size:7.5pt; color:#475569;">Relevé d'Identité Bancaire (RIB) officiel à créditer obligatoirement :</span><br/>
                <span class="rib-number">%s</span>
            </div>
            <div style="font-size:7.2pt; color:#64748b; margin-top:2px;">
                <em>Mode alternatif admis (Décret exécutif n° 21-551) : Chèque de banque certifié libellé à l'ordre du « Régisseur des Recettes de l'ANPP ».</em>
            </div>
            <div style="font-size:7.8pt; color:#b91c1c; font-weight:800; margin-top:3px; background:#fef2f2; padding:3px 8px; border-radius:3px; border:1px solid #fecaca; display:inline-block;">
                Motif impératif à porter sur l'avis d'opéré : DEP-%s-%s
            </div>
        </div>

        <!-- 6. Triple Cartouche Officiel des 3 Volets de Validation & Signatures -->
        <div class="sign-grid-3">
            <div class="sign-card">
                <div class="sign-header">Volet 1 : Établissement Assujetti</div>
                <div style="font-size:6.8pt; color:#475569;">Visa Pharmacien DT &amp; Direction Financière</div>
                <div class="sign-space">
                    <div>
                        <strong>%s</strong><br/>
                        <span style="font-size:6.2pt; color:#64748b;">Mandat Légal 1.A.7<br/>Date, Cachet Émetteur &amp; Signature</span>
                    </div>
                </div>
            </div>

            <div class="sign-card">
                <div class="sign-header">Volet 2 : Établissement Bancaire</div>
                <div style="font-size:6.8pt; color:#475569;">Attestation de virement exécuté</div>
                <div class="sign-space">
                    <div>
                        <span style="font-size:6.2pt; color:#64748b;">Avis d'opéré bancaire &amp; Date de valeur<br/>Cachet guichet &amp; Référence transaction</span>
                    </div>
                </div>
            </div>

            <div class="sign-card">
                <div class="sign-header">Volet 3 : Régie des Recettes ANPP</div>
                <div style="font-size:6.8pt; color:#475569;">Validation &amp; Quittance délivrée</div>
                <div class="sign-space">
                    <div>
                        <span style="font-size:6.2pt; color:#64748b;">Quittance N° : .............................<br/>Date &amp; Timbre humide Régisseur ANPP</span>
                    </div>
                </div>
            </div>
        </div>

        <!-- 7. Pied de page de Sécurité & ALCOA+ -->
        <div class="footer-legal">
            <div>
                <strong>SANAD PHARMA</strong> • Document Officiel d'État • Conforme Guide CTD ANPP Juillet 2025 &amp; GAMP 5
            </div>
            <div style="font-family:monospace; font-size:6.5pt;">
                CAS-ID : %s • Horodatage certifié UTC : %s
            </div>
        </div>
    </div>
</body>
</html>`,
		slip.SlipType,
		escapeHTMLStr(slip.ProductBrandName),
		slip.SlipType,
		slip.SlipNumber,
		wmClass, wmText,
		slip.SlipNumber,
		slip.SlipNumber,
		slip.GeneratedAt.Format("02/01/2006"),
		slipTypeLabel,
		slipTypeLabelAR,
		escapeHTMLStr(slip.DossierCode),
		escapeHTMLStr(slip.OrganizationName),
		escapeHTMLStr(miphLic),
		escapeHTMLStr(dtName),
		escapeHTMLStr(estName),
		escapeHTMLStr(genByName),
		func() string {
			if slip.VerificationStatus == "conforme" {
				return "#10b981"
			} else if slip.Status == "paye" || slip.Status == "quittance_recue" {
				return "#2563eb"
			}
			return "#b45309"
		}(),
		func() string {
			if slip.VerificationStatus == "conforme" {
				return "✅ Acquitté & Conforme (Visé DT)"
			} else if slip.Status == "paye" || slip.Status == "quittance_recue" {
				return "🏛️ Quittance Enregistrée"
			}
			return "⏳ Émis (En attente d'encaissement)"
		}(),
		escapeHTMLStr(slip.ProductBrandName),
		escapeHTMLStr(dciText),
		escapeHTMLStr(dosageText),
		escapeHTMLStr(formText),
		escapeHTMLStr(routeText),
		escapeHTMLStr(procLabel),
		escapeHTMLStr(slip.DossierCode),
		slip.SlipType,
		escapeHTMLStr(slip.PaymentPurpose),
		escapeHTMLStr(slip.ProductBrandName),
		escapeHTMLStr(slip.FeeScheduleReference),
		slip.AmountDZD,
		slip.AmountDZD,
		slip.AmountWordsFR,
		slip.AmountWordsAR,
		fxBlockHTML,
		escapeHTMLStr(slip.BankAccountRIB),
		slip.SlipType,
		escapeHTMLStr(slip.DossierCode),
		escapeHTMLStr(dtName),
		fmt.Sprintf("CAS-SHA256-%s", slip.ID[:8]),
		time.Now().UTC().Format("02/01/2006 15:04:05 UTC"),
	)

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(html))
}


// VerifyPaymentSlip1A4 applique le contrôle officiel 1.A.4 sur un bordereau (quittance, concordance, change BA)
func (h *FeeHandler) VerifyPaymentSlip1A4(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	dossierID := chi.URLParam(r, "id")
	slipID := chi.URLParam(r, "slipId")
	if dossierID == "" || slipID == "" {
		writeJSONError(w, http.StatusBadRequest, "Identifiants dossier et bordereau requis")
		return
	}

	// Contrôle SoD : la validation financière 1.A.4 est réservée aux profils réglementaires et DT
	normRole := strings.ToLower(user.Role)
	canVerify := user.HasPermission(string(domain.PermDocApprovePDT)) ||
		user.HasPermission(string(domain.PermDocCheckReview)) ||
		normRole == "directeur_technique" ||
		normRole == "pharmacien_responsable" ||
		normRole == "responsable_reglementaire" ||
		normRole == "admin"
	if !canVerify {
		writeJSONError(w, http.StatusForbidden, "Ségrégation des tâches : seul le Responsable Réglementaire ou le Pharmacien DT peut viser la conformité financière 1.A.4")
		return
	}

	var req domain.VerifyPaymentSlip1A4Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	req.BankName = strings.TrimSpace(req.BankName)
	req.BankReceiptRef = strings.TrimSpace(req.BankReceiptRef)
	req.BankReceiptDate = strings.TrimSpace(req.BankReceiptDate)
	if req.BankReceiptRef == "" || req.BankReceiptDate == "" {
		writeJSONError(w, http.StatusBadRequest, "La référence de quittance/avis bancaire et la date de paiement sont obligatoires")
		return
	}

	dossUUID, err := uuid.Parse(dossierID)
	if err != nil {
		writeJSONError(w, http.StatusBadRequest, "Identifiant dossier invalide")
		return
	}

	// Récupérer le dossier pour connaître le type de procédure
	dossier, err := h.dossierRepo.GetDossierByID(r.Context(), user.OrganizationID, dossUUID)
	if err != nil {
		writeJSONError(w, http.StatusNotFound, "Dossier réglementaire introuvable")
		return
	}

	isImported := (dossier.ProcedureCode == "NI" || dossier.ProcedureCode == "RI" || dossier.ProcedureCode == "MI" || dossier.ProcedureCode == "dm_importe")

	// Récupérer le barème légal attendu
	expectedAmt := 150000.00 // Par défaut générique local
	sched, err := h.feeRepo.GetFeeScheduleForProcedure(r.Context(), string(dossier.ProcedureCode), time.Now())
	if err == nil && sched != nil {
		expectedAmt = sched.AmountDZD
	} else if isImported {
		expectedAmt = 250000.00
	}

	verifierName := strings.TrimSpace(user.FirstName + " " + user.LastName)
	if verifierName == "" {
		verifierName = user.Email
	}

	updatedSlip, err := h.feeRepo.VerifyPaymentSlip1A4(
		r.Context(),
		user.OrganizationID.String(),
		slipID,
		user.UserID.String(),
		verifierName,
		req,
		expectedAmt,
		isImported,
	)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Erreur vérification 1.A.4: %v", err))
		return
	}

	// Audit Trail ALCOA+
	if h.auditService != nil {
		h.auditService.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "FEE_VERIFY_1A4",
			EntityType:     "anpp_payment_slips",
			EntityID:       slipID,
			Details: map[string]any{
				"dossier_id":            dossierID,
				"dossier_code":          dossier.DossierCode,
				"slip_number":           updatedSlip.SlipNumber,
				"amount_dzd":            updatedSlip.AmountDZD,
				"expected_amount_dzd":   expectedAmt,
				"bank_name":             req.BankName,
				"bank_receipt_ref":      req.BankReceiptRef,
				"verification_status":   updatedSlip.VerificationStatus,
				"amount_concordance_ok": updatedSlip.AmountConcordanceOK,
				"requires_fx":           updatedSlip.RequiresFXAttestation,
				"fx_rate":               updatedSlip.FXRate,
			},
		})
	}

	summary, _ := h.feeRepo.GetSection1A4Summary(r.Context(), user.OrganizationID.String(), dossierID)

	writeJSON(w, http.StatusOK, map[string]any{
		"message": "Contrôle réglementaire 1.A.4 consigné avec succès",
		"slip":    updatedSlip,
		"summary": summary,
	})
}

// GetSection1A4Summary retourne l'analyse de conformité financière pour la section 1.A.4
func (h *FeeHandler) GetSection1A4Summary(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	dossierID := chi.URLParam(r, "id")
	if dossierID == "" {
		writeJSONError(w, http.StatusBadRequest, "Identifiant dossier requis")
		return
	}

	summary, err := h.feeRepo.GetSection1A4Summary(r.Context(), user.OrganizationID.String(), dossierID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Erreur récupération synthèse 1.A.4: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

// VerifyPreSubmission1A5 enregistre et valide la décision préalable et les frais de pré-soumission ANPP
func (h *FeeHandler) VerifyPreSubmission1A5(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	dossierID := chi.URLParam(r, "id")
	if dossierID == "" {
		writeJSONError(w, http.StatusBadRequest, "Identifiant dossier requis")
		return
	}

	// Contrôle SoD
	normRole := strings.ToLower(user.Role)
	canVerify := user.HasPermission(string(domain.PermDocApprovePDT)) ||
		user.HasPermission(string(domain.PermDocCheckReview)) ||
		normRole == "directeur_technique" ||
		normRole == "pharmacien_responsable" ||
		normRole == "responsable_reglementaire" ||
		normRole == "admin"
	if !canVerify {
		writeJSONError(w, http.StatusForbidden, "Ségrégation des tâches : seul le Responsable Réglementaire ou le Pharmacien DT peut viser la pré-soumission 1.A.5")
		return
	}

	var req domain.VerifyPreSubmission1A5Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeJSONError(w, http.StatusBadRequest, "Corps JSON invalide")
		return
	}

	req.ANPPReference = strings.TrimSpace(req.ANPPReference)
	req.SlipReference = strings.TrimSpace(req.SlipReference)
	req.ANPPNotificationDate = strings.TrimSpace(req.ANPPNotificationDate)
	req.PaymentDate = strings.TrimSpace(req.PaymentDate)

	verifierName := strings.TrimSpace(user.FirstName + " " + user.LastName)
	if verifierName == "" {
		verifierName = user.Email
	}

	preSub, err := h.feeRepo.VerifyPreSubmission1A5(
		r.Context(),
		user.OrganizationID.String(),
		dossierID,
		user.UserID.String(),
		verifierName,
		req,
	)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Erreur enregistrement pré-soumission 1.A.5: %v", err))
		return
	}

	// Audit Trail ALCOA+
	if h.auditService != nil {
		h.auditService.Log(r.Context(), &audit.Entry{
			OrganizationID: user.OrganizationID,
			UserID:         &user.UserID,
			UserEmail:      user.Email,
			UserRole:       user.Role,
			Action:         "FEE_VERIFY_1A5",
			EntityType:     "anpp_pre_submissions",
			EntityID:       preSub.ID,
			Details: map[string]any{
				"dossier_id":           dossierID,
				"anpp_reference":       req.ANPPReference,
				"decision_type":        preSub.DecisionType,
				"fee_amount_dzd":       preSub.FeeAmountDZD,
				"slip_reference":       preSub.SlipReference,
				"verification_status":  preSub.VerificationStatus,
			},
		})
	}

	summary, _ := h.feeRepo.GetSection1A5Summary(r.Context(), user.OrganizationID.String(), dossierID)

	writeJSON(w, http.StatusOK, map[string]any{
		"message":        "Accord et redevance de pré-soumission enregistrés avec succès",
		"pre_submission": preSub,
		"summary":        summary,
	})
}

// GetSection1A5Summary retourne l'analyse de conformité pour la section 1.A.5
func (h *FeeHandler) GetSection1A5Summary(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	dossierID := chi.URLParam(r, "id")
	if dossierID == "" {
		writeJSONError(w, http.StatusBadRequest, "Identifiant dossier requis")
		return
	}

	summary, err := h.feeRepo.GetSection1A5Summary(r.Context(), user.OrganizationID.String(), dossierID)
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Erreur récupération synthèse 1.A.5: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

// GetCostConsolidationDashboard retourne la synthèse consolidée des dépenses réglementaires
func (h *FeeHandler) GetCostConsolidationDashboard(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	summary, err := h.feeRepo.GetCostConsolidationSummary(r.Context(), user.OrganizationID.String())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Erreur consolidation coûts: %v", err))
		return
	}

	writeJSON(w, http.StatusOK, summary)
}

// ExportCostConsolidationCSV génère un export CSV officiel pour la comptabilité et la DAF
func (h *FeeHandler) ExportCostConsolidationCSV(w http.ResponseWriter, r *http.Request) {
	user := GetSessionUser(r.Context())
	if user == nil {
		writeJSONError(w, http.StatusUnauthorized, "Authentification requise")
		return
	}

	summary, err := h.feeRepo.GetCostConsolidationSummary(r.Context(), user.OrganizationID.String())
	if err != nil {
		writeJSONError(w, http.StatusInternalServerError, fmt.Sprintf("Erreur export coûts: %v", err))
		return
	}

	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", "attachment; filename=\"sanad_couts_reglementaires_consolides.csv\"")

	// UTF-8 BOM
	w.Write([]byte{0xEF, 0xBB, 0xBF})

	b := &strings.Builder{}
	b.WriteString("# SANAD PHARMA — ÉTAT RÉCAPITULATIF CONSOLIDÉ DES REDEVANCES ANPP (ALCOA+)\n")
	b.WriteString(fmt.Sprintf("# Établissement : %s\n", summary.OrganizationName))
	b.WriteString(fmt.Sprintf("# Date d'extraction : %s\n", time.Now().Format("02/01/2006 15:04:05")))
	b.WriteString(fmt.Sprintf("# Total Émis : %.2f DZD | Total Acquitté : %.2f DZD | En Attente : %.2f DZD\n\n",
		summary.TotalFeesEmittedDZD, summary.TotalFeesPaidDZD, summary.TotalFeesPendingDZD))

	b.WriteString("SECTION 1 : COÛTS RÉGLEMENTAIRES PAR PRODUIT & DOSSIER\n")
	b.WriteString("Produit;Type;Dossier;Procédure;Pré-soumission (1.A.5) DZD;Dépôt Initial (1.A.4) DZD;Analyses LNCM (BV2) DZD;Variations DZD;Total Engagé DZD;Total Acquitté DZD;Statut Financier\n")

	for _, p := range summary.ProductBreakdown {
		b.WriteString(fmt.Sprintf("%s;%s;%s;%s;%.2f;%.2f;%.2f;%.2f;%.2f;%.2f;%s\n",
			p.ProductName, p.ProductType, p.DossierCode, p.ProcedureCode,
			p.PreSubmissionDZD, p.InitialDepositDZD, p.ComplementaryDZD, p.VariationsDZD,
			p.TotalCommittedDZD, p.TotalPaidDZD, p.FinancialStatus))
	}

	b.WriteString("\nSECTION 2 : CONSOLIDATION PAR EXERCICE ANNUEL\n")
	b.WriteString("Année;Nombre de Dossiers;Total Émis DZD;Total Acquitté DZD;Reste à Décaisser DZD;Pré-soumission DZD;Dépôt Initial DZD;Analyses LNCM DZD\n")

	for _, y := range summary.YearlyBreakdown {
		b.WriteString(fmt.Sprintf("%d;%d;%.2f;%.2f;%.2f;%.2f;%.2f;%.2f\n",
			y.Year, y.DossierCount, y.TotalEmittedDZD, y.TotalPaidDZD, y.TotalPendingDZD,
			y.PreSubmissionDZD, y.InitialDepositDZD, y.ComplementaryDZD))
	}

	w.Write([]byte(b.String()))
}

func escapeHTMLStr(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}
