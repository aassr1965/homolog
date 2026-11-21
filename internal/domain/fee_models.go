package domain

import (
	"fmt"
	"strings"
	"time"
)

// Constantes pour les types de bordereaux de versement
const (
	SlipTypeBV1 = "BV1" // Bordereau Dépôt Initial de Dossier (Enregistrement / Homologation)
	SlipTypeBV2 = "BV2" // Bordereau Complémentaire (Analyses LNCM, Expertises, Variations)
)

// Constantes pour le statut du paiement
const (
	PaymentStatusEmis           = "emis"
	PaymentStatusPaye           = "paye"
	PaymentStatusQuittanceRecue = "quittance_recue"
	PaymentStatusAnnule         = "annule"
)

// ANPPFeeSchedule représente une ligne de barème officiel avec sa durée de validité juridique
type ANPPFeeSchedule struct {
	ID             string     `json:"id"`
	Category       string     `json:"category"`       // medicament, dispositif_medical, variation, autre
	ProcedureCode  string     `json:"procedure_code"` // princeps, generique_local, generique_importe, etc.
	LabelFR        string     `json:"label_fr"`
	LabelAR        string     `json:"label_ar,omitempty"`
	LegalReference string     `json:"legal_reference"` // Journal Officiel / Arrêté interministériel
	AmountDZD      float64    `json:"amount_dzd"`
	ValidFrom      time.Time  `json:"valid_from"`
	ValidTo        *time.Time `json:"valid_to,omitempty"`
	IsActive       bool       `json:"is_active"`
	Notes          string     `json:"notes,omitempty"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

// ANPPPaymentSlip représente un bordereau de versement BV1 ou BV2 scellé pour le Trésor Public
type ANPPPaymentSlip struct {
	ID                   string     `json:"id"`
	OrganizationID       string     `json:"organization_id"`
	OrganizationName     string     `json:"organization_name,omitempty"`
	DossierID            string     `json:"dossier_id"`
	DossierCode          string     `json:"dossier_code,omitempty"`
	ProductBrandName     string     `json:"product_brand_name,omitempty"`
	EstablishmentName    string     `json:"establishment_name,omitempty"`
	FeeScheduleID        string     `json:"fee_schedule_id"`
	FeeScheduleReference string     `json:"fee_schedule_reference,omitempty"`
	SlipType             string     `json:"slip_type"`   // BV1, BV2
	SlipNumber           string     `json:"slip_number"` // BV1-2026-AMOX-001
	PaymentPurpose       string     `json:"payment_purpose"`
	AmountDZD            float64    `json:"amount_dzd"`
	AmountWordsFR        string     `json:"amount_words_fr"`
	AmountWordsAR        string     `json:"amount_words_ar"`
	BankAccountRIB       string     `json:"bank_account_rib"`
	Status               string     `json:"status"` // emis, paye, quittance_recue, annule
	PaymentReference     string     `json:"payment_reference,omitempty"`
	PaidAt               *time.Time `json:"paid_at,omitempty"`
	ReceiptDocumentID    *string    `json:"receipt_document_id,omitempty"`
	ReceiptDocumentTitle string     `json:"receipt_document_title,omitempty"`
	GeneratedBy          *string    `json:"generated_by,omitempty"`
	GeneratedByName      string     `json:"generated_by_name,omitempty"`
	GeneratedAt          time.Time  `json:"generated_at"`
	CreatedAt            time.Time  `json:"created_at"`
	UpdatedAt            time.Time  `json:"updated_at"`

	// Volet 1.A.4 : Traçabilité quittance bancaire, conformité financière & Attestation Banque d'Algérie
	Currency              string     `json:"currency"`                         // DZD, EUR, USD...
	CurrencyAmount        *float64   `json:"currency_amount,omitempty"`
	BankName              string     `json:"bank_name,omitempty"`             // BNA, CPA, BEA...
	PaymentMode           string     `json:"payment_mode"`                    // virement, cheque_certifie, versement_especes
	BankReceiptRef        string     `json:"bank_receipt_ref,omitempty"`      // Réf avis de débit ou quittance
	BankReceiptDate       *time.Time `json:"bank_receipt_date,omitempty"`
	RequiresFXAttestation bool       `json:"requires_fx_attestation"`         // Obligatoire pour importé si devises
	FXRate                *float64   `json:"fx_rate,omitempty"`               // Taux officiel Banque d'Algérie
	FXRateDate            *time.Time `json:"fx_rate_date,omitempty"`          // Date cours Banque d'Algérie
	FXAttestationRef      string     `json:"fx_attestation_ref,omitempty"`   // N° attestation Banque d'Algérie
	FXAttestationDocID    *string    `json:"fx_attestation_doc_id,omitempty"` // Scan GED attestation de cours
	FXAttestationDocTitle string     `json:"fx_attestation_doc_title,omitempty"`
	VerificationStatus    string     `json:"verification_status"`             // pending, conforme, non_conforme
	VerifiedBy            *string    `json:"verified_by,omitempty"`
	VerifiedByName        string     `json:"verified_by_name,omitempty"`
	VerifiedAt            *time.Time `json:"verified_at,omitempty"`
	VerificationNotes     string     `json:"verification_notes,omitempty"`
	AmountConcordanceOK   bool       `json:"amount_concordance_ok"`
	SoDCheckPassed        bool       `json:"sod_check_passed"`
}

// VerifyPaymentSlip1A4Request représente la demande de vérification réglementaire et financière 1.A.4
type VerifyPaymentSlip1A4Request struct {
	BankName              string   `json:"bank_name"`
	PaymentMode           string   `json:"payment_mode"` // virement, cheque_certifie, versement_especes
	BankReceiptRef        string   `json:"bank_receipt_ref"`
	BankReceiptDate       string   `json:"bank_receipt_date"` // YYYY-MM-DD
	ReceiptDocumentID     *string  `json:"receipt_document_id,omitempty"`
	Currency              string   `json:"currency"` // DZD, EUR, USD
	CurrencyAmount        *float64 `json:"currency_amount,omitempty"`
	RequiresFXAttestation bool     `json:"requires_fx_attestation"`
	FXRate                *float64 `json:"fx_rate,omitempty"`
	FXRateDate            *string  `json:"fx_rate_date,omitempty"` // YYYY-MM-DD
	FXAttestationRef      string   `json:"fx_attestation_ref,omitempty"`
	FXAttestationDocID    *string  `json:"fx_attestation_doc_id,omitempty"`
	VerificationNotes     string   `json:"verification_notes,omitempty"`
}

// Section1A4Summary représente le statut de conformité et l'ensemble des pièces pour la section 1.A.4
type Section1A4Summary struct {
	DossierID            string           `json:"dossier_id"`
	DossierCode          string           `json:"dossier_code"`
	ProcedureCode        string           `json:"procedure_code"` // NI, RI, MI, NF, RF, MF
	IsImported           bool             `json:"is_imported"`
	Section1A4Status     string           `json:"section_1a4_status"` // manquant, rattache, valide_qualite, approuve_dt
	ActiveSlip           *ANPPPaymentSlip `json:"active_slip,omitempty"`
	ExpectedFeeAmountDZD float64          `json:"expected_fee_amount_dzd"`
	FeeScheduleLabel     string           `json:"fee_schedule_label"`
	LegalReference       string           `json:"legal_reference"`
	IsFeeSatisfied       bool             `json:"is_fee_satisfied"`
	IsReceiptAttached    bool             `json:"is_receipt_attached"`
	IsFXRequired         bool             `json:"is_fx_required"`
	IsFXSatisfied        bool             `json:"is_fx_satisfied"`
	OverallCompliant     bool             `json:"overall_compliant"`
	ComplianceChecklist  []string         `json:"compliance_checklist"`
	BlockingIssues       []string         `json:"blocking_issues"`
}

// CreateFeeScheduleRequest représente la demande de mise à jour ou création d'un barème
type CreateFeeScheduleRequest struct {
	Category       string  `json:"category"`
	ProcedureCode  string  `json:"procedure_code"`
	LabelFR        string  `json:"label_fr"`
	LabelAR        string  `json:"label_ar,omitempty"`
	LegalReference string  `json:"legal_reference"`
	AmountDZD      float64 `json:"amount_dzd"`
	ValidFrom      string  `json:"valid_from"` // YYYY-MM-DD
	Notes          string  `json:"notes,omitempty"`
}

// GeneratePaymentSlipRequest représente la demande de génération d'un bordereau BV1/BV2
type GeneratePaymentSlipRequest struct {
	SlipType       string   `json:"slip_type"` // BV1 ou BV2
	PaymentPurpose string   `json:"payment_purpose,omitempty"`
	CustomAmount   *float64 `json:"custom_amount,omitempty"` // Optionnel : surcharge manuelle si dérogation légale
}

// UpdatePaymentSlipStatusRequest représente la saisie de quittance ou confirmation de paiement
type UpdatePaymentSlipStatusRequest struct {
	Status            string  `json:"status"` // paye, quittance_recue, annule
	PaymentReference  string  `json:"payment_reference"`
	PaidAt            string  `json:"paid_at"` // YYYY-MM-DD
	ReceiptDocumentID *string `json:"receipt_document_id,omitempty"`
}

// Tableaux de conversion pour les nombres en français
var unitsFR = []string{"", "un", "deux", "trois", "quatre", "cinq", "six", "sept", "huit", "neuf", "dix", "onze", "douze", "treize", "quatorze", "quinze", "seize", "dix-sept", "dix-huit", "dix-neuf"}
var tensFR = []string{"", "", "vingt", "trente", "quarante", "cinquante", "soixante", "soixante-dix", "quatre-vingts", "quatre-vingt-dix"}

// NumberToWordsFR convertit un entier (0 à 999 999 999) en lettres françaises
func NumberToWordsFR(n int64) string {
	if n == 0 {
		return "zéro"
	}
	if n < 0 {
		return "moins " + NumberToWordsFR(-n)
	}

	var parts []string

	// Millions
	if n >= 1000000 {
		millions := n / 1000000
		if millions == 1 {
			parts = append(parts, "un million")
		} else {
			parts = append(parts, NumberToWordsFR(millions)+" millions")
		}
		n %= 1000000
	}

	// Milliers
	if n >= 1000 {
		thousands := n / 1000
		if thousands == 1 {
			parts = append(parts, "mille")
		} else {
			parts = append(parts, NumberToWordsFR(thousands)+" mille")
		}
		n %= 1000
	}

	// Centaines
	if n >= 100 {
		hundreds := n / 100
		if hundreds == 1 {
			parts = append(parts, "cent")
		} else {
			rem := n % 100
			if rem == 0 {
				parts = append(parts, unitsFR[hundreds]+" cents")
			} else {
				parts = append(parts, unitsFR[hundreds]+" cent")
			}
		}
		n %= 100
	}

	// Dizaines et unités (1 à 99)
	if n > 0 {
		if n < 20 {
			parts = append(parts, unitsFR[n])
		} else if n < 70 {
			t := n / 10
			u := n % 10
			if u == 1 {
				parts = append(parts, tensFR[t]+" et un")
			} else if u > 1 {
				parts = append(parts, tensFR[t]+"-"+unitsFR[u])
			} else {
				parts = append(parts, tensFR[t])
			}
		} else if n < 80 {
			u := n - 60
			if u == 11 {
				parts = append(parts, "soixante et onze")
			} else {
				parts = append(parts, "soixante-"+unitsFR[u])
			}
		} else if n < 90 {
			u := n - 80
			if u == 0 {
				parts = append(parts, "quatre-vingts")
			} else {
				parts = append(parts, "quatre-vingt-"+unitsFR[u])
			}
		} else {
			u := n - 80
			parts = append(parts, "quatre-vingt-"+unitsFR[u])
		}
	}

	res := strings.Join(parts, " ")
	return strings.TrimSpace(res)
}

// NumberToWordsAR convertit un entier (0 à 999 999 999) en lettres arabes
func NumberToWordsAR(n int64) string {
	if n == 0 {
		return "صفر"
	}

	unitsAR := []string{"", "واحد", "اثنان", "ثلاثة", "أربعة", "خمسة", "ستة", "سبعة", "ثمانية", "تسعة", "عشرة"}
	teensAR := []string{"عشرة", "أحد عشر", "اثنا عشر", "ثلاثة عشر", "أربعة عشر", "خمسة عشر", "ستة عشر", "سبعة عشر", "ثمانية عشر", "تسعة عشر"}
	tensAR := []string{"", "", "عشرون", "ثلاثون", "أربعون", "خمسون", "ستون", "سبعون", "ثمانون", "تسعون"}
	hundredsAR := []string{"", "مائة", "مائتان", "ثلاثمائة", "أربعمائة", "خمسمائة", "ستمائة", "سبعمائة", "ثمانمائة", "تسعمائة"}

	var parts []string

	if n >= 1000000 {
		millions := n / 1000000
		if millions == 1 {
			parts = append(parts, "مليون")
		} else if millions == 2 {
			parts = append(parts, "مليونان")
		} else if millions >= 3 && millions <= 10 {
			parts = append(parts, NumberToWordsAR(millions)+" ملايين")
		} else {
			parts = append(parts, NumberToWordsAR(millions)+" مليون")
		}
		n %= 1000000
	}

	if n >= 1000 {
		thousands := n / 1000
		if thousands == 1 {
			parts = append(parts, "ألف")
		} else if thousands == 2 {
			parts = append(parts, "ألفان")
		} else if thousands >= 3 && thousands <= 10 {
			parts = append(parts, unitsAR[thousands]+" آلاف")
		} else {
			parts = append(parts, NumberToWordsAR(thousands)+" ألف")
		}
		n %= 1000
	}

	if n >= 100 {
		h := n / 100
		parts = append(parts, hundredsAR[h])
		n %= 100
	}

	if n > 0 {
		if n <= 10 {
			parts = append(parts, unitsAR[n])
		} else if n < 20 {
			parts = append(parts, teensAR[n-10])
		} else {
			t := n / 10
			u := n % 10
			if u == 0 {
				parts = append(parts, tensAR[t])
			} else {
				parts = append(parts, unitsAR[u]+" و "+tensAR[t])
			}
		}
	}

	return strings.Join(parts, " و ")
}

// ConvertAmountDZDToWords convertit un montant DZD en prose juridique française et arabe
func ConvertAmountDZDToWords(amount float64) (string, string) {
	intPart := int64(amount)
	centPart := int64(amount*100) % 100

	wordsFR := NumberToWordsFR(intPart)
	// Capitaliser la première lettre
	if len(wordsFR) > 0 {
		wordsFR = strings.ToUpper(wordsFR[:1]) + wordsFR[1:]
	}
	wordsFR += " dinars algériens"
	if centPart > 0 {
		wordsFR += fmt.Sprintf(" et %s centimes", NumberToWordsFR(centPart))
	}

	wordsAR := NumberToWordsAR(intPart) + " دينار جزائري"
	if centPart > 0 {
		wordsAR += " و " + NumberToWordsAR(centPart) + " سنتيم"
	}

	return wordsFR, wordsAR
}
