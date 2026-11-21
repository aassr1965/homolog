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

	html := fmt.Sprintf(`<!DOCTYPE html>
<html lang="fr" dir="ltr">
<head>
    <meta charset="UTF-8">
    <title>Bordereau de Versement %s — ANPP</title>
    <style>
        @page { size: A4 portrait; margin: 0; }
        body { font-family: 'Segoe UI', Arial, sans-serif; font-size: 12px; color: #111; line-height: 1.4; margin: 0; padding: 15mm; background: white; }
        .header { display: flex; justify-content: space-between; align-items: center; border-bottom: 2px solid #0284c7; padding-bottom: 12px; margin-bottom: 16px; }
        .republic { text-align: center; flex: 1; }
        .republic h4 { margin: 0; font-size: 13px; font-weight: 700; }
        .republic h5 { margin: 2px 0 0; font-size: 11px; font-weight: normal; color: #444; }
        .slip-title-box { background: #f0fdf4; border: 2px solid #16a34a; border-radius: 6px; padding: 12px; text-align: center; margin-bottom: 20px; }
        .slip-title { font-size: 17px; font-weight: 800; color: #166534; text-transform: uppercase; }
        .slip-sub { font-size: 11px; color: #15803d; margin-top: 4px; font-weight: 600; }
        .info-grid { display: grid; grid-template-columns: 1fr 1fr; gap: 12px; margin-bottom: 18px; }
        .info-card { border: 1px solid #cbd5e1; border-radius: 4px; padding: 10px; background: #f8fafc; }
        .info-card h6 { margin: 0 0 6px; font-size: 11px; text-transform: uppercase; color: #64748b; }
        table { width: 100%%; border-collapse: collapse; margin-bottom: 20px; }
        th, td { border: 1px solid #cbd5e1; padding: 8px 10px; text-align: left; }
        th { background: #f1f5f9; font-weight: 700; font-size: 11px; }
        .amount-box { background: #eff6ff; border: 1.5px solid #3b82f6; border-radius: 6px; padding: 12px; margin-bottom: 24px; }
        .signatures { display: grid; grid-template-columns: 1fr 1fr 1fr; gap: 12px; margin-top: 40px; text-align: center; font-size: 11px; }
        .sig-box { border: 1px dashed #94a3b8; border-radius: 4px; height: 90px; padding: 6px; }
        .barcode { text-align: right; font-family: monospace; font-size: 11px; letter-spacing: 2px; }
        @media print {
            .no-print { display: none; }
            body { padding: 0; }
        }
    </style>
</head>
<body>
    <div class="no-print" style="margin-bottom: 15px; text-align: right;">
        <button onclick="window.print()" style="background:#0284c7; color:white; border:none; padding:8px 16px; border-radius:4px; font-weight:bold; cursor:pointer;">🖨️ Imprimer le Bordereau Officiel</button>
    </div>

    <div class="header">
        <div style="font-size: 10px; color: #64748b; width: 160px;">
            <strong>SANAD PHARMA</strong><br>
            Système d'Homologation BPF<br>
            Versionné ALCOA+
        </div>
        <div class="republic">
            <h4>RÉPUBLIQUE ALGÉRIENNE DÉMOCRATIQUE ET POPULAIRE</h4>
            <h5 dir="rtl">الجمهورية الجزائرية الديمقراطية الشعبية</h5>
            <div style="font-weight: 700; margin-top: 4px; color: #0284c7;">
                AGENCE NATIONALE DES PRODUITS PHARMACEUTIQUES (ANPP)<br>
                <span dir="rtl" style="font-size: 11px; color: #334155;">الوكالة الوطنية للمواد الصيدلانية</span>
            </div>
        </div>
        <div class="barcode">
            <div style="font-size: 9px; color: #64748b;">NUMÉRO UNIQUE :</div>
            <strong>*%s*</strong>
        </div>
    </div>

    <div class="slip-title-box">
        <div class="slip-title">BORDEREAU DE VERSEMENT DE REDEVANCE — %s</div>
        <div class="slip-sub" dir="rtl">استمارة دفع الأتعاب والرسوم الرسمية للمصادقة والتسجيل</div>
    </div>

    <div class="info-grid">
        <div class="info-card">
            <h6>Demandeur &amp; Établissement Pharmaceutique</h6>
            <div><strong>Raison Sociale :</strong> %s</div>
            <div><strong>N° Agrément MIPH :</strong> Conforme Décret Exécutif 20-325</div>
            <div><strong>Généré par :</strong> %s</div>
        </div>
        <div class="info-card">
            <h6>Objet Réglementaire &amp; Dossier</h6>
            <div><strong>Dossier N° :</strong> %s</div>
            <div><strong>Produit :</strong> %s</div>
            <div><strong>Date d'Émission :</strong> %s</div>
        </div>
    </div>

    <table>
        <thead>
            <tr>
                <th>Désignation de la Redevance Légale</th>
                <th>Base Juridique (Journal Officiel)</th>
                <th style="text-align: right; width: 140px;">Montant Net (DZD)</th>
            </tr>
        </thead>
        <tbody>
            <tr>
                <td>
                    <strong>%s</strong><br>
                    <span style="font-size: 10.5px; color: #475569;">%s</span>
                </td>
                <td style="font-size: 11px;">
                    %s
                </td>
                <td style="text-align: right; font-size: 14px; font-weight: 800; color: #0f172a;">
                    %.2f DZD
                </td>
            </tr>
        </tbody>
    </table>

    <div class="amount-box">
        <div style="font-size: 11px; color: #1e40af; font-weight: 700; text-transform: uppercase;">Montant en toutes lettres :</div>
        <div style="font-size: 13px; font-weight: 800; color: #1e3a8a; margin-top: 4px;">%s</div>
        <div dir="rtl" style="font-size: 14px; font-weight: 800; color: #1e3a8a; margin-top: 4px;">%s</div>
    </div>

    <div style="background: #f8fafc; border: 1px solid #cbd5e1; border-radius: 4px; padding: 10px; font-size: 11px; margin-bottom: 20px;">
        <strong>Compte Bancaire de Règlement (Trésorerie Centrale ANPP) :</strong><br>
        RIB Officiel : <code>%s</code><br>
        <span style="color: #64748b;">Le versement doit s'effectuer par virement bancaire ou chèque certifié à l'ordre de l'Agent Comptable de l'ANPP. La quittance du Trésor fait foi pour le dépôt officiel.</span>
    </div>

    <div class="signatures">
        <div>
            <strong>Le Déposant / Mandataire</strong>
            <div class="sig-box" style="margin-top:6px;">Cachet &amp; Signature</div>
        </div>
        <div>
            <strong>Visa Pharmacien Directeur Technique</strong>
            <div class="sig-box" style="margin-top:6px;">Mandat Légal 1.A.7</div>
        </div>
        <div>
            <strong>L'Agent Comptable ANPP</strong>
            <div class="sig-box" style="margin-top:6px;">Date &amp; Quittance Trésor</div>
        </div>
    </div>
</body>
</html>`,
		slip.SlipNumber,
		slip.SlipType,
		slip.SlipType,
		escapeHTMLStr(slip.OrganizationName),
		escapeHTMLStr(slip.GeneratedByName),
		escapeHTMLStr(slip.DossierCode),
		escapeHTMLStr(slip.ProductBrandName),
		slip.GeneratedAt.Format("02/01/2006"),
		escapeHTMLStr(slip.PaymentPurpose),
		escapeHTMLStr(slip.FeeScheduleReference),
		escapeHTMLStr(slip.FeeScheduleReference),
		slip.AmountDZD,
		slip.AmountWordsFR,
		slip.AmountWordsAR,
		escapeHTMLStr(slip.BankAccountRIB),
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

func escapeHTMLStr(s string) string {
	s = strings.ReplaceAll(s, "&", "&amp;")
	s = strings.ReplaceAll(s, "<", "&lt;")
	s = strings.ReplaceAll(s, ">", "&gt;")
	s = strings.ReplaceAll(s, "\"", "&quot;")
	return s
}
