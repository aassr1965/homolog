package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

// ProcedureCode represents the 6 official ANPP submission procedures + DM homologation
type ProcedureCode string

const (
	ProcedureNI ProcedureCode = "NI"         // Enregistrement Initial d'un produit Importé
	ProcedureRI ProcedureCode = "RI"         // Renouvellement de la DE d'un produit Importé
	ProcedureMI ProcedureCode = "MI"         // Modification relative à un produit Importé
	ProcedureNF ProcedureCode = "NF"         // Enregistrement Initial d'un produit Fabriqué localement
	ProcedureRF ProcedureCode = "RF"         // Renouvellement de la DE d'un produit Fabriqué localement
	ProcedureMF ProcedureCode = "MF"         // Modification relative à un produit Fabriqué localement
	ProcedureDM ProcedureCode = "HOMOLOG_DM" // Homologation Dispositif Médical
)

// DossierStatus represents the BPF lifecycle stages of an ANPP registration dossier
type DossierStatus string

const (
	DossierStatusDraft             DossierStatus = "brouillon"
	DossierStatusInPreparation     DossierStatus = "en_preparation"
	DossierStatusUnderReview       DossierStatus = "en_revue"
	DossierStatusValidatedInternal DossierStatus = "valide_interne" // Exige 100% complétude + signature PDT
	DossierStatusSubmittedANPP     DossierStatus = "depose_anpp"
	DossierStatusApprovedANPP      DossierStatus = "approuve_anpp"
	DossierStatusRejectedANPP      DossierStatus = "rejete_anpp"
)

// RequirementLevel represents the ANPP regulatory necessity of a document
type RequirementLevel string

const (
	RequirementExiged        RequirementLevel = "E"  // Exigé obligatoirement
	RequirementFacultative   RequirementLevel = "F"  // Facultatif (l'ANPP peut exiger à tout moment)
	RequirementPartial       RequirementLevel = "P"  // Partiellement requis (ex: bioéquivalence M5 ou Biowaiver)
	RequirementNotApplicable RequirementLevel = "NA" // Non applicable pour cette procédure
)

// ChecklistItemStatus represents the attachment and validation state of a requirement
type ChecklistItemStatus string

const (
	ItemStatusMissing    ChecklistItemStatus = "manquant"
	ItemStatusAttached   ChecklistItemStatus = "rattache"
	ItemStatusValidated  ChecklistItemStatus = "valide_qualite"
	ItemStatusApprovedDT ChecklistItemStatus = "approuve_dt"
	ItemStatusRejected   ChecklistItemStatus = "rejete"
)

// Dossier represents an official ANPP regulatory dossier instance
type Dossier struct {
	ID                    uuid.UUID       `json:"id"`
	OrganizationID        uuid.UUID       `json:"organization_id"`
	MedicineID            *uuid.UUID      `json:"medicine_id,omitempty"`
	MedicalDeviceID       *uuid.UUID      `json:"medical_device_id,omitempty"`
	ProductType           ProductType     `json:"product_type"`
	DossierCode           string          `json:"dossier_code"`
	ProcedureCode         ProcedureCode   `json:"procedure_code"`
	InternalReference     string          `json:"internal_reference"`
	ANPPOfficialReference *string         `json:"anpp_official_reference,omitempty"`
	ANPPDepositDate       *time.Time      `json:"anpp_deposit_date,omitempty"`
	ANPPDecisionDate      *time.Time      `json:"anpp_decision_date,omitempty"`
	ANPPValidityEndDate   *time.Time      `json:"anpp_validity_end_date,omitempty"`
	TargetDepositDate     *time.Time      `json:"target_deposit_date,omitempty"`
	TrackingNumber        *string         `json:"tracking_number,omitempty"`
	DepositReceiptRef     *string         `json:"deposit_receipt_ref,omitempty"`
	DepositReceiptDate    *time.Time      `json:"deposit_receipt_date,omitempty"`
	SubmissionNotes       *string         `json:"submission_notes,omitempty"`
	Status                DossierStatus   `json:"status"`
	ResponsibleID         *uuid.UUID      `json:"responsible_id,omitempty"`
	CompletenessPercent   int             `json:"completeness_percent"`
	IsLockedForDeposit    bool            `json:"is_locked_for_deposit"`
	LockedByDTID          *uuid.UUID      `json:"locked_by_dt_id,omitempty"`
	LockedAt              *time.Time      `json:"locked_at,omitempty"`
	PDTSignedAt           *time.Time      `json:"pdt_signed_at,omitempty"`
	PDTSignedBy           *uuid.UUID      `json:"pdt_signed_by,omitempty"`
	PDTSignatureSeal      *string         `json:"pdt_signature_seal,omitempty"`
	PDTSignatureRole      *string         `json:"pdt_signature_role,omitempty"`
	ChecklistData         []ChecklistItem `json:"checklist_data,omitempty"`
	Items                 []ChecklistItem `json:"items,omitempty"`
	Medicine              *Medicine       `json:"medicine,omitempty"`
	MedicalDevice         *MedicalDevice  `json:"medical_device,omitempty"`
	CreatedAt             time.Time       `json:"created_at"`
	UpdatedAt             time.Time       `json:"updated_at"`
}

// SecurityClassification représente le niveau de secret industriel ou commercial d'une exigence
type SecurityClassification string

const (
	ClassificationStandard         SecurityClassification = "STANDARD"
	ClassificationSecretCommercial SecurityClassification = "SECRET_COMMERCIAL"
	ClassificationSecretIndustriel SecurityClassification = "SECRET_INDUSTRIEL"
)

// ChecklistItem represents an atomic documentary requirement of an ANPP dossier
type ChecklistItem struct {
	ID                     uuid.UUID              `json:"id"`
	DossierID              uuid.UUID              `json:"dossier_id"`
	ModuleNumber           int                    `json:"module_number"` // 1-5 pour CTD Médicament, 0 pour Dispositif Médical
	SubModule              string                 `json:"sub_module"`    // '1.A', '1.H', '1.M', '1.P', '3.2.S', '5.3.1.2', etc.
	ItemCode               string                 `json:"item_code"`     // '1.A.1', '1.H.4', '1.P.8', '5.3.1.2'
	TitleFR                string                 `json:"title_fr"`
	TitleAR                string                 `json:"title_ar,omitempty"`
	DescriptionFR          string                 `json:"description_fr,omitempty"`
	RequirementLevel       RequirementLevel       `json:"requirement_level"` // E, F, P, NA
	IsMandatory            bool                   `json:"is_mandatory"`
	GEDDocumentID          *uuid.UUID             `json:"ged_document_id,omitempty"`
	GEDDocument            *GEDDocument           `json:"ged_document,omitempty"`
	Status                 ChecklistItemStatus    `json:"status"` // manquant, rattache, valide_qualite, approuve_dt, rejete
	ValidationNotes        *string                `json:"validation_notes,omitempty"`
	SortOrder              int                    `json:"sort_order"`
	SecurityClassification SecurityClassification `json:"security_classification"`
	// Enrichissements Workflow Maker-Checker-Approver (Jalon 2.1)
	WorkflowStatus        WorkflowStatus `json:"workflow_status"` // draft_pending, drafting, under_review, pending_approval, approved_sealed
	AssignedDepartment    string         `json:"assigned_department"`
	AssignedUserID        *uuid.UUID     `json:"assigned_user_id,omitempty"`
	AssignedUserName      string         `json:"assigned_user_name,omitempty"`
	DueDate               *time.Time     `json:"due_date,omitempty"`
	IsLocked              bool           `json:"is_locked"`
	LockedBy              *uuid.UUID     `json:"locked_by,omitempty"`
	LockedByName          string         `json:"locked_by_name"`
	LockedAt              *time.Time     `json:"locked_at,omitempty"`
	LockReason            *string        `json:"lock_reason,omitempty"`
	CurrentRevision       string         `json:"current_revision"`
	MakerUserID           *uuid.UUID     `json:"maker_user_id,omitempty"`
	MakerUserName         string         `json:"maker_user_name,omitempty"`
	MakerSubmittedAt      *time.Time     `json:"maker_submitted_at,omitempty"`
	CheckerUserID         *uuid.UUID     `json:"checker_user_id,omitempty"`
	CheckerUserName       string         `json:"checker_user_name,omitempty"`
	CheckerReviewedAt     *time.Time     `json:"checker_reviewed_at,omitempty"`
	CheckerDecision       *string        `json:"checker_decision,omitempty"`
	CheckerNotes          *string        `json:"checker_notes,omitempty"`
	ApproverUserID        *uuid.UUID     `json:"approver_user_id,omitempty"`
	ApproverUserName      string         `json:"approver_user_name,omitempty"`
	ApproverSignedAt      *time.Time     `json:"approver_signed_at,omitempty"`
	ApproverSignatureSeal *string        `json:"approver_signature_seal,omitempty"`
	AuthoredContent       *string        `json:"authored_content,omitempty"`
	TemplateCode          *string        `json:"template_code,omitempty"`
	CreatedAt             time.Time      `json:"created_at"`
	UpdatedAt             time.Time      `json:"updated_at"`
}

// DossierAssignment représente l'affectation nominative d'un collaborateur ou consultant à un dossier (Need-to-Know)
type DossierAssignment struct {
	ID                      uuid.UUID  `json:"id"`
	DossierID               uuid.UUID  `json:"dossier_id"`
	UserID                  uuid.UUID  `json:"user_id"`
	UserEmail               string     `json:"user_email,omitempty"`
	UserFirstName           string     `json:"user_first_name,omitempty"`
	UserLastName            string     `json:"user_last_name,omitempty"`
	UserRole                string     `json:"user_role,omitempty"`
	RoleInDossier           string     `json:"role_in_dossier"`
	CanViewFinancials       bool       `json:"can_view_financials"`
	CanViewIndustrialSecret bool       `json:"can_view_industrial_secret"`
	AssignedAt              time.Time  `json:"assigned_at"`
	ExpiresAt               *time.Time `json:"expires_at,omitempty"`
	Notes                   *string    `json:"notes,omitempty"`
	CreatedAt               time.Time  `json:"created_at"`
}

// ValidateDossier checks core integrity of the dossier before creation
func (d *Dossier) ValidateDossier() error {
	if strings.TrimSpace(d.DossierCode) == "" {
		return errors.New("le code du dossier réglementaire est obligatoire")
	}
	if strings.TrimSpace(d.InternalReference) == "" {
		return errors.New("la référence interne du dossier est obligatoire")
	}

	validProcedures := map[ProcedureCode]bool{
		ProcedureNI: true, ProcedureRI: true, ProcedureMI: true,
		ProcedureNF: true, ProcedureRF: true, ProcedureMF: true,
		ProcedureDM: true,
	}
	if !validProcedures[d.ProcedureCode] {
		return fmt.Errorf("procédure réglementaire invalide: %s", d.ProcedureCode)
	}

	if d.ProductType == ProductTypeMedicine && d.MedicineID == nil {
		return errors.New("l'identifiant du médicament est obligatoire pour un dossier CTD médicament")
	}
	if d.ProductType == ProductTypeMedicalDevice && d.MedicalDeviceID == nil {
		return errors.New("l'identifiant du dispositif médical est obligatoire pour un dossier d'homologation DM")
	}

	return nil
}

// CalculateCompleteness evaluates real-time completion percentage based on mandatory items
func CalculateCompleteness(items []ChecklistItem) int {
	totalMandatory := 0
	fulfilledMandatory := 0

	for _, item := range items {
		if item.IsMandatory && item.RequirementLevel == RequirementExiged {
			totalMandatory++
			if item.Status == ItemStatusAttached || item.Status == ItemStatusValidated || item.Status == ItemStatusApprovedDT {
				fulfilledMandatory++
			}
		}
	}

	if totalMandatory == 0 {
		return 100
	}

	return (fulfilledMandatory * 100) / totalMandatory
}

// CanValidateInternal enforces strict 21 CFR Part 11 & BPF integrity before internal validation or deposit
func CanValidateInternal(d *Dossier, items []ChecklistItem) error {
	completeness := CalculateCompleteness(items)
	if completeness < 100 {
		return fmt.Errorf("verrouillage de conformité BPF: le dossier est incomplet (%d%% / 100%% exigés). Toutes les pièces obligatoires [E] doivent être rattachées", completeness)
	}

	// Vérification de la présence effective de chaque pièce obligatoire [E]
	for _, item := range items {
		if item.IsMandatory && item.RequirementLevel == RequirementExiged {
			if item.GEDDocumentID == nil || item.Status == ItemStatusMissing || item.Status == ItemStatusRejected {
				return fmt.Errorf("pièce obligatoire manquante ou rejetée: [%s] %s", item.ItemCode, item.TitleFR)
			}
		}
	}

	// Vérification du visa légal du Pharmacien Directeur Technique
	if d.PDTSignedAt == nil || d.PDTSignatureSeal == nil || strings.TrimSpace(*d.PDTSignatureSeal) == "" {
		return errors.New("verrouillage réglementaire légal: le visa électronique 21 CFR Part 11 du Pharmacien Directeur Technique est obligatoire pour valider le dossier avant dépôt ANPP")
	}

	return nil
}

// IsLocalManufacturing returns true if procedure is national production (NF, RF, MF)
func (p ProcedureCode) IsLocalManufacturing() bool {
	return p == ProcedureNF || p == ProcedureRF || p == ProcedureMF
}

// IsImport returns true if procedure is imported product (NI, RI, MI)
func (p ProcedureCode) IsImport() bool {
	return p == ProcedureNI || p == ProcedureRI || p == ProcedureMI
}

// IsRenewal returns true if procedure is 5-year renewal (RI, RF)
func (p ProcedureCode) IsRenewal() bool {
	return p == ProcedureRI || p == ProcedureRF
}

// IsVariation returns true if procedure is variation/modification (MI, MF)
func (p ProcedureCode) IsVariation() bool {
	return p == ProcedureMI || p == ProcedureMF
}

// GenerateCTDChecklist generates the complete official ANPP CTD checklist according to Guide Juillet 2025
func GenerateCTDChecklist(proc ProcedureCode, origin MedicineOriginType, isPrinceps bool) []ChecklistItem {
	isLocal := proc.IsLocalManufacturing()
	isRenewal := proc.IsRenewal()
	isVariation := proc.IsVariation()

	var items []ChecklistItem
	order := 10

	// =========================================================================
	// MODULE 1 : SPÉCIFICITÉS RÉGIONALES ANPP (ALGÉRIE)
	// =========================================================================

	// --- Sous-Module 1.A : Demandeur & Établissement (Applicant) ---
	add1A := func(code, titleFR, titleAR, desc string, req RequirementLevel) {
		items = append(items, ChecklistItem{
			ModuleNumber:     1,
			SubModule:        "1.A",
			ItemCode:         code,
			TitleFR:          titleFR,
			TitleAR:          titleAR,
			DescriptionFR:    desc,
			RequirementLevel: req,
			IsMandatory:      req == RequirementExiged,
			Status:           ItemStatusMissing,
			SortOrder:        order,
		})
		order += 10
	}

	add1A("1.A.1", "Récépissé de dépôt du dossier", "وصل إيداع الملف", "Document scanné avec numéro d'enregistrement ANPP et horodatage", RequirementExiged)
	add1A("1.A.2", "Courrier de demande officiel", "رسالة الطلب الرسمية", "À en-tête du laboratoire, daté et signé par le Pharmacien Directeur Technique", RequirementExiged)
	add1A("1.A.3", "Formulaire officiel paraphé, daté, signé et cacheté par le PDT", "الاستمارة الرسمية مؤشرة وموقعة من الصيدلي المدير التقني", "Formulaire officiel ANPP, paraphes sur chaque page, signature et cachet légal DT", RequirementExiged)
	add1A("1.A.4", "Bordereau de versement + justificatif bancaire / quittance", "جدول الدفع + الوثيقة البنكية / وصل السداد", "Preuve de paiement des redevances ANPP + attestation de change Banque d'Algérie si devises", RequirementExiged)

	// 1.A.5 : Accord de pré-soumission (Obligatoire pour NI, MI, NF, MF ; Non requis pour renouvellements purs)
	req1A5 := RequirementExiged
	if isRenewal {
		req1A5 = RequirementNotApplicable
	}
	add1A("1.A.5", "Copie de la réponse à la demande de pré-soumission (+ BV)", "نسخة من الرد على طلب الإيداع الأولي + جدول الدفع", "Accord formel de pré-soumission délivré par l'ANPP", req1A5)

	add1A("1.A.6", "Agrément d'exploitation / fabrication / sous-traitance", "اعتماد الاستغلال / الصنع / المناولة", "Agrément ministériel en cours de validité délivré par le MIPH", RequirementExiged)
	add1A("1.A.7", "Décision d'exercice du Pharmacien Directeur Technique (PDT)", "قرار ممارسة الصيدلي المدير التقني", "Décision ministérielle nominative nommant le DT responsable", RequirementExiged)
	add1A("1.A.8", "Copie des Contrats d'exploitation, façonnage et accords qualité (QTA)", "نسخ من عقود الاستغلال والمناولة واتفاقيات الجودة", "Contrats juridiques et Quality Technical Agreements liant le demandeur aux sites", RequirementExiged)

	// 1.A.9 : Historique DE et anciens dépôts (Obligatoire pour RI, RF, MI, MF)
	req1A9 := RequirementNotApplicable
	if isRenewal || isVariation {
		req1A9 = RequirementExiged
	}
	add1A("1.A.9", "Copie de la DE, approbations et accusés des anciens dépôts", "نسخة من قرار التسجيل والاعتمادات والوصولات السابقة", "Historique réglementaire indispensable pour renouvellements et variations", req1A9)

	// --- Sous-Module 1.H : Détenteur (Holder) ---
	add1H := func(code, titleFR, titleAR, desc string, req RequirementLevel) {
		items = append(items, ChecklistItem{
			ModuleNumber:     1,
			SubModule:        "1.H",
			ItemCode:         code,
			TitleFR:          titleFR,
			TitleAR:          titleAR,
			DescriptionFR:    desc,
			RequirementLevel: req,
			IsMandatory:      req == RequirementExiged,
			Status:           ItemStatusMissing,
			SortOrder:        order,
		})
		order += 10
	}

	add1H("1.H.1", "Courrier de demande du détenteur", "رسالة طلب الحائز", "Mandat ou courrier officiel du détenteur autorisant le demandeur local à déposer", RequirementExiged)
	add1H("1.H.2", "Autorisation de Mise sur le Marché (AMM pays d'origine)", "رخصة التسويق في بلد المنشأ", "AMM délivrée dans le pays d'origine par l'autorité sanitaire compétente", RequirementExiged)
	add1H("1.H.3", "Certificat de Produit Pharmaceutique (CPP - Modèle OMS)", "شهادة المنتوج الصيدلاني (نموذج منظمة الصحة العالمية)", "CPP valide, légalisé/apostillé conforme aux directives OMS", RequirementExiged)

	// 1.H.4 : Structure PCSU (Local) ou Attestation FOB Chambre de Commerce (Import)
	desc1H4 := "Structure PCSU détaillée (Prix de Revient et Sortie Usine)"
	if !isLocal {
		desc1H4 = "Attestation de prix FOB Algérie visée par la Chambre de Commerce du pays d'origine"
	}
	add1H("1.H.4", "Structure de prix PCSU / Attestation de prix FOB visée", "هيكل الأسعار عند الخروج من المصنع / شهادة سعر فوب مؤشرة", desc1H4, RequirementExiged)

	// 1.H.5 : Benchmark de prix international (Exigé à l'import NI/RI/MI)
	req1H5 := RequirementNotApplicable
	if !isLocal {
		req1H5 = RequirementExiged
	}
	add1H("1.H.5", "Dossier BENCHMARK de prix international", "ملف المقارنة المرجعية الدولية للأسعار", "Comparatif de prix dans les pays de référence (obligatoire pour importations)", req1H5)

	// 1.H.6 & 1.H.7 : Spécifique variations (MI/MF)
	reqVar := RequirementNotApplicable
	if isVariation {
		reqVar = RequirementExiged
	}
	add1H("1.H.6", "Approbation de la variation par le pays d'origine", "موافقة بلد المنشأ على التعديل", "Preuve que la modification a été approuvée dans le pays d'origine", reqVar)
	add1H("1.H.7", "Descriptif de la mise à jour signé", "الوصف التفصيلي للتحديث موقع", "Document de synthèse expliquant le rationnel technico-réglementaire", reqVar)

	add1H("1.H.8", "Liste des sites intervenants (SA et Produit Fini) visée", "قائمة المواقع المتدخلة في الصنع والمراقبة مؤشرة", "Cartographie exhaustive des sites signée et visée par le détenteur", RequirementExiged)

	// 1.H.9 : Déclaration de similarité (Initial NI/NF)
	reqSim := RequirementNotApplicable
	if !isRenewal && !isVariation {
		reqSim = RequirementExiged
	}
	add1H("1.H.9", "Déclaration de similarité", "تصريح بالمماثلة والتطابق", "Attestation certifiant la similarité de formule et procédé", reqSim)

	// 1.H.10 : Déclaration de non-modification (Renouvellement RI/RF)
	reqNonMod := RequirementNotApplicable
	if isRenewal {
		reqNonMod = RequirementExiged
	}
	add1H("1.H.10", "Déclaration de non-modification", "تصريح بعدم إجراء أي تعديل غير مرخص", "Attestation formelle certifiant qu'aucune modification non autorisée n'a été opérée", reqNonMod)

	add1H("1.H.11", "Accord de transfert de DE le cas échéant", "اتفاق التنازل أو تحويل قرار التسجيل", "Requis uniquement si la variation implique un changement de titulaire de la DE", RequirementFacultative)
	add1H("1.H.12", "Attestation de date d’implémentation de la variation", "شهادة تاريخ تطبيق التعديل في بلد المنشأ", "Date d'effet industriel de la modification", RequirementFacultative)

	// --- Sous-Module 1.M : Fabricants SA et Produit Fini (Manufacturers) ---
	add1M := func(code, titleFR, titleAR, desc string, req RequirementLevel) {
		items = append(items, ChecklistItem{
			ModuleNumber:     1,
			SubModule:        "1.M",
			ItemCode:         code,
			TitleFR:          titleFR,
			TitleAR:          titleAR,
			DescriptionFR:    desc,
			RequirementLevel: req,
			IsMandatory:      req == RequirementExiged,
			Status:           ItemStatusMissing,
			SortOrder:        order,
		})
		order += 10
	}

	add1M("1.M.S.1", "Agrément du site de fabrication de la Substance Active (ML)", "اعتماد موقع تصنيع المادة الفعالة", "Licence de fabrication officielle du site de synthèse API", RequirementExiged)
	add1M("1.M.S.2", "Certificat BPF/GMP de la Substance Active", "شهادة حسن ممارسات التصنيع للمادة الفعالة", "Certificat GMP valide émis par autorité compétente pour l'API", RequirementExiged)
	add1M("1.M.S.3", "Lettre d’accès / Certificat de Conformité CEP EDQM", "رسالة النفاذ / شهادة مطابقة الفارماكوبيا الأوروبية", "CEP valide avec déclaration d'accès nominative ou DPSA", RequirementExiged)
	add1M("1.M.P.1", "Agrément du site de fabrication du Produit Fini (ML)", "اعتماد موقع تصنيع المنتوج النهائي", "Licence d'établissement pharmaceutique pour le façonnier/fabricant PF", RequirementExiged)
	add1M("1.M.P.2", "Certificat de conformité BPF/GMP du Produit Fini", "شهادة حسن ممارسات التصنيع للمنتوج النهائي", "Certificat BPF valide pour toutes les étapes de fabrication/conditionnement", RequirementExiged)

	// 1.M.P.3 : Étude de bioéquivalence ou justification Biowaiver
	reqBE := RequirementExiged
	if isPrinceps || isRenewal {
		reqBE = RequirementNotApplicable
	}
	add1M("1.M.P.3", "Étude de bioéquivalence ou justification Biowaiver", "دراسة التكافؤ الحيوي أو تبرير الإعفاء", "Rapport d'étude BE in vivo ou justification d'exonération BCS classe I/III", reqBE)

	// --- Sous-Module 1.P : Informations Produit & Contrôle (Product Info) ---
	add1P := func(code, titleFR, titleAR, desc string, req RequirementLevel) {
		items = append(items, ChecklistItem{
			ModuleNumber:     1,
			SubModule:        "1.P",
			ItemCode:         code,
			TitleFR:          titleFR,
			TitleAR:          titleAR,
			DescriptionFR:    desc,
			RequirementLevel: req,
			IsMandatory:      req == RequirementExiged,
			Status:           ItemStatusMissing,
			SortOrder:        order,
		})
		order += 10
	}

	add1P("1.P.1.1", "RCP officiel pays d'origine", "ملخص خصائص المنتوج في بلد المنشأ", "Résumé des Caractéristiques du Produit approuvé à l'origine", RequirementExiged)
	add1P("1.P.1.2", "Notice officielle pays d'origine", "النشرة الرسمية في بلد المنشأ", "Notice patient approuvée dans le pays exportateur", RequirementExiged)
	add1P("1.P.1.3", "Étiquetage officiel pays d'origine", "الوسم واللصاقات الرسمية في بلد المنشأ", "Étiquettes de conditionnement primaire et secondaire d'origine", RequirementExiged)

	add1P("1.P.2.1", "RCP Modèle Algérie (visé par le PDT)", "ملخص خصائص المنتوج - نموذج الجزائر مؤشر من المدير التقني", "Projet de RCP conforme aux exigences médicales et réglementaires nationales", RequirementExiged)
	add1P("1.P.2.2", "Notice Modèle Algérie (Bilingue FR/AR visée PDT)", "نشرة المريض نموذج الجزائر (ثنائية اللغة فرنسي/عربي مؤشرة)", "Notice patient obligatoire en arabe et en français, relue et visée par le DT", RequirementExiged)
	add1P("1.P.2.3", "Étiquette Modèle Algérie", "لصاقة نموذج الجزائر", "Projet d'étiquetage conforme aux mentions légales algériennes", RequirementExiged)
	add1P("1.P.2.4", "Maquettes Modèle Algérie (Couleurs échelle 100%)", "عينات العلب واللصاقات بالألوان وبالمقاس الحقيقي 100%", "Conditionnements primaire et secondaire à plat en taille réelle 100%", RequirementExiged)

	add1P("1.P.3", "Enregistrement des Échantillons du Produit Fini", "تسجيل إيداع عينات المنتوج النهائي لدى مخبر الرقابة", "Preuve de remise physique des échantillons au laboratoire de l'ANPP", RequirementExiged)
	add1P("1.P.4", "Bulletin d'analyse Substance Active et recontrôle", "شهادة تحليل المادة الفعالة وإعادة المراقبة", "CoA fabricant SA + bulletin de recontrôle complet avec granulométrie et polymorphisme", RequirementExiged)
	add1P("1.P.5", "Bulletin d’analyse du Produit Fini libératoire", "شهادة تحرير وتحليل المنتوج النهائي", "Certificat d'analyse libératoire complet conforme aux spécifications", RequirementExiged)
	add1P("1.P.6", "Attestation de correspondance entre les lots", "شهادة مطابقة وتوافق الحصص", "Traçabilité entre le lot soumis, le lot de bioéquivalence et les lots de stabilité", RequirementExiged)
	add1P("1.P.7", "Certificat officiel de libération de lot", "شهادة تحرير الحصة الرسمية", "Certificat émis par la personne qualifiée (QP/PDT)", RequirementExiged)

	// 1.P.8 & 1.P.9 : Exigences exclusives Fabrication Locale (NF/MF)
	reqLocal := RequirementNotApplicable
	if isLocal && (proc == ProcedureNF || proc == ProcedureMF) {
		reqLocal = RequirementExiged
	}
	add1P("1.P.8", "Lettre d’engagement de non-commercialisation des lots de validation", "تعهد بعدم تسويق حصص التحقق من الصلاحية قبل نيل المقرر", "Engagement formel signé du DT à ne pas commercialiser les lots de validation avant DE", reqLocal)
	add1P("1.P.9", "Lettre d’engagement de suivi des stabilités en temps réel Zone IVa/IVb", "تعهد بمواصلة دراسات الاستقرار في الزمن الحقيقي للمنطقة الرابعة", "Engagement à poursuivre les stabilités en temps réel et transmettre les résultats annuels", reqLocal)

	// =========================================================================
	// MODULES TECHNIQUES CTD (2, 3, 4 ET 5)
	// =========================================================================

	// Module 2 : Résumés
	reqM2 := RequirementExiged
	descM2 := "Classeurs M2.1 à M2.7 obligatoires pour Princeps"
	if !isPrinceps {
		reqM2 = RequirementFacultative
		descM2 = "Incorporé dans les modules 3, 4 et 5 pour les produits génériques"
	}
	items = append(items, ChecklistItem{
		ModuleNumber:     2,
		SubModule:        "2.CTD",
		ItemCode:         "2.Overview",
		TitleFR:          "Module 2 : Résumés globaux et synthèses de qualité",
		TitleAR:          "الوحدة 2 : الملخصات العامة والتقييمات",
		DescriptionFR:    descM2,
		RequirementLevel: reqM2,
		IsMandatory:      reqM2 == RequirementExiged,
		Status:           ItemStatusMissing,
		SortOrder:        order,
	})
	order += 10

	// Module 3 : Qualité Pharmaceutique, Chimique et Biologique (CMC) - Toujours exigé [E]
	addM3 := func(code, titleFR, titleAR, desc string, req RequirementLevel) {
		items = append(items, ChecklistItem{
			ModuleNumber:     3,
			SubModule:        "3.CMC",
			ItemCode:         code,
			TitleFR:          titleFR,
			TitleAR:          titleAR,
			DescriptionFR:    desc,
			RequirementLevel: req,
			IsMandatory:      req == RequirementExiged,
			Status:           ItemStatusMissing,
			SortOrder:        order,
		})
		order += 10
	}

	addM3("3.2.S", "Module 3.2.S : Substance Active (Caractérisation, Synthèse, Contrôle, Stabilité)", "الوحدة 3.2.S : المادة الفعالة (التصنيع، الخصائص، المراقبة والاستقرار)", "Données complètes sur l'API, profil d'impuretés, granulométrie et stabilités", RequirementExiged)
	addM3("3.2.P", "Module 3.2.P : Produit Fini (Formule, Développement galénique, Procédé, Stabilité Zone IVb)", "الوحدة 3.2.P : المنتوج النهائي (التركيبة، التطوير، طريقة الصنع، الاستقرار)", "Dossier galénique complet, spécifications à libération et péremption, stabilités 30°C/65% RH ou 40°C/75% RH", RequirementExiged)

	// 3.2.P.3.5 : Lots de validation (Spécifique fabrication locale NF/MF)
	req32P35 := RequirementNotApplicable
	if isLocal {
		req32P35 = RequirementExiged
	}
	addM3("3.2.P.3.5", "Dossiers de lot maîtres et fabrication des lots de validation sur site national", "سجلات تصنيع حصص التحقق الصناعية المنفذة بالمصنع الوطني", "Dossiers complets de fabrication et conditionnement des lots de validation industriels", req32P35)

	addM3("3.2.R", "Module 3.2.R : Informations Régionales Algérie (Master Batch Records & Validations)", "الوحدة 3.2.R : المعلومات الإقليمية الخاصة بالجزائر", "Formules cadres de fabrication/conditionnement et méthodes de validation analytique sur site", RequirementExiged)

	// Module 4 : Données Non-Cliniques (Toxicologie / Pharmacologie)
	reqM4 := RequirementExiged
	descM4 := "Études pharmacologiques et toxicologiques obligatoires pour Princeps"
	if !isPrinceps {
		reqM4 = RequirementFacultative
		descM4 = "Facultatif pour Génériques selon tableau officiel ANPP (peut être requis pour impuretés nouvelles)"
	}
	items = append(items, ChecklistItem{
		ModuleNumber:     4,
		SubModule:        "4.TOX",
		ItemCode:         "4.Safety",
		TitleFR:          "Module 4 : Rapports d’études non cliniques (Sécurité et Toxicologie)",
		TitleAR:          "الوحدة 4 : دراسات السلامة ما قبل السريرية وعلم السموم",
		DescriptionFR:    descM4,
		RequirementLevel: reqM4,
		IsMandatory:      reqM4 == RequirementExiged,
		Status:           ItemStatusMissing,
		SortOrder:        order,
	})
	order += 10

	// Module 5 : Données Cliniques & Bioéquivalence
	reqM5 := RequirementExiged
	descM5 := "Dossier d'essais cliniques complets (Phases I à III) pour Princeps"
	if !isPrinceps {
		reqM5 = RequirementPartial
		descM5 = "Partiellement requis : focalisé impérativement sur la bioéquivalence (5.3.1.2) ou Biowaiver"
	}
	items = append(items, ChecklistItem{
		ModuleNumber:     5,
		SubModule:        "5.CLIN",
		ItemCode:         "5.Clinical",
		TitleFR:          "Module 5 : Rapports d'études cliniques & Bioéquivalence in vivo",
		TitleAR:          "الوحدة 5 : تقارير التجارب السريرية ودراسات التكافؤ الحيوي",
		DescriptionFR:    descM5,
		RequirementLevel: reqM5,
		IsMandatory:      reqM5 == RequirementExiged || (!isPrinceps && reqBE == RequirementExiged),
		Status:           ItemStatusMissing,
		SortOrder:        order,
	})
	order += 10

	// Spécifique 5.3.1.2 : Bioéquivalence in vivo stricte pour Génériques
	if !isPrinceps && reqBE == RequirementExiged {
		items = append(items, ChecklistItem{
			ModuleNumber:     5,
			SubModule:        "5.3.1.2",
			ItemCode:         "5.3.1.2",
			TitleFR:          "Étude comparative de Biodisponibilité (BA) et Bioéquivalence (BE) in vivo",
			TitleAR:          "الدراسة المقارنة للتوافر الحيوي والتكافؤ الحيوي في جسم الإنسان",
			DescriptionFR:    "Rapport complet de l'étude clinique in vivo (IC 90% sur Cmax et AUC compris entre 80,00% et 125,00%)",
			RequirementLevel: RequirementExiged,
			IsMandatory:      true,
			Status:           ItemStatusMissing,
			SortOrder:        order,
		})
		order += 10

		items = append(items, ChecklistItem{
			ModuleNumber:     5,
			SubModule:        "5.3.1.4",
			ItemCode:         "5.3.1.4",
			TitleFR:          "Validation des méthodes bioanalytiques humaines (BPL/GCP)",
			TitleAR:          "التحقق من صحة طرق التحليل البيولوجي البشري",
			DescriptionFR:    "Validation de la méthode de dosage bioanalytique (LC-MS/MS) conforme aux BPL",
			RequirementLevel: RequirementExiged,
			IsMandatory:      true,
			Status:           ItemStatusMissing,
			SortOrder:        order,
		})
		order += 10
	}

	return items
}

// GenerateDMChecklist generates the checklist for Medical Devices according to Décret 20-324 & Arrêté 10/05/2021
func GenerateDMChecklist(dmClass string, isSterile bool) []ChecklistItem {
	var items []ChecklistItem
	order := 10

	addDM := func(code, titleFR, titleAR, desc string, mandatory bool) {
		req := RequirementExiged
		if !mandatory {
			req = RequirementFacultative
		}
		items = append(items, ChecklistItem{
			ModuleNumber:     0,
			SubModule:        "DM_TECH",
			ItemCode:         code,
			TitleFR:          titleFR,
			TitleAR:          titleAR,
			DescriptionFR:    desc,
			RequirementLevel: req,
			IsMandatory:      mandatory,
			Status:           ItemStatusMissing,
			SortOrder:        order,
		})
		order += 10
	}

	// 1. Partie Administrative & Juridique
	addDM("DM.A.1", "Formulaire officiel de demande d'homologation DM paraphé et signé", "استمارة طلب المصادقة على المستلزم الطبي رسمية وموقعة", "Formulaire officiel ANPP visé par le Pharmacien Directeur Technique", true)
	addDM("DM.A.2", "Agrément de l'établissement importateur ou fabricant MIPH", "اعتماد المؤسسة المستوردة أو المصنعة من وزارة الصناعة الصيدلانية", "Agrément ministériel en cours de validité", true)
	addDM("DM.A.3", "Décision nominative d'exercice du Pharmacien Directeur Technique", "قرار ممارسة الصيدلي المدير التقني المسؤول", "Décision officielle nominative délivrée par le MIPH", true)
	addDM("DM.A.4", "Contrat liant le fabricant au demandeur et accords de vigilance", "عقد التوزيع والوكالة مع المصنع واتفاقية اليقظة", "Mandat exclusif ou contrat commercial en cours de validité", true)
	addDM("DM.A.5", "Quittance de paiement des redevances d'homologation ANPP", "وصل دفع الأتعاب والرسوم الخاصة بالمصادقة", "Justificatif bancaire conforme à la nomenclature financière", true)

	// 2. Conformité Qualité & Certifications (ISO 13485 & CE)
	addDM("DM.Q.1", "Certificat de Système de Management de la Qualité ISO 13485", "شهادة نظام إدارة الجودة إيزو 13485 للمصنع", "Certificat ISO 13485 valide délivré par un organisme notifié reconnu", true)
	addDM("DM.Q.2", "Déclaration de conformité CE / Certificat de conformité CE", "إعلان المطابقة للمواصفات الأوروبية أو الدولية", "Déclaration de conformité selon la directive 93/42/CEE ou règlement 2017/745", true)
	addDM("DM.Q.3", "Certificat de Vente Libre (FSC / Free Sale Certificate)", "شهادة البيع الحر في بلد المنشأ", "Certificat délivré par l'autorité sanitaire du pays fabricant", true)

	// 3. Dossier Technique (Conception & Spécifications)
	addDM("DM.T.1", "Descriptif technique complet, indications et principes de fonctionnement", "الوصف التقني الكامل ودواعي الاستعمال وطريقة العمل", "Fiche technique détaillée, schémas, nomenclature des composants et matériaux", true)
	addDM("DM.T.2", "Dossier d'analyse et gestion des risques (ISO 14971)", "ملف تحليل وتسيير المخاطر وفق إيزو 14971", "Rapport complet de gestion des risques sur le cycle de vie du DM", true)
	addDM("DM.T.3", "Rapport de validation de procédé et contrôles de libération", "تقرير التحقق من طريقة الصنع ومراقبة الجودة عند التحرير", "Spécifications de libération du produit fini", true)

	// 4. Stérilisation (Si DM stérile)
	if isSterile {
		addDM("DM.S.1", "Dossier de validation du procédé de stérilisation (ISO 11135 / ISO 11137)", "ملف التحقق من طريقة التعقيم (أكسيد الإيثيلين أو الإشعاع)", "Méthode de stérilisation, niveau d'assurance de stérilité (SAL 10^-6) et résidus", true)
		addDM("DM.S.2", "Preuve d'intégrité de la barrière stérile et durée de validité (ISO 11607)", "إثبات سلامة غلاف التعقيم ومدة الصلاحية", "Études de vieillissement accéléré et temps réel de l'emballage stérile", true)
	}

	// 5. Informations Produit (Notices & Étiquetage Algérie)
	addDM("DM.I.1", "Notice d'utilisation bilingue (Arabe et Français obligatoire)", "دليل الاستعمال باللغتين العربية والفرنسية إجباريا", "Notice patient ou praticien avec mentions légales algériennes", true)
	addDM("DM.I.2", "Maquettes d'étiquetage primaire et secondaire à l'échelle 100%", "عينات الوسم واللصاقات بالمقاس الحقيقي 100%", "Étiquettes mentionnant le numéro de lot, péremption, symbole stérile et fabricant", true)

	// 6. Évaluation Clinique ou Matériovigilance
	if dmClass == "III" || dmClass == "IIb" || dmClass == "DMDIV" {
		addDM("DM.C.1", "Rapport d'évaluation clinique ou performances diagnostiques (DMDIV)", "تقرير التقييم السريري أو الأداء التشخيصي للمستلزمات المخبرية", "Données cliniques ou études de sensibilité/spécificité analytique et clinique", true)
	}

	return items
}

// GeneratePDTSignatureSeal calculates the 21 CFR Part 11 cryptographic seal for Technical Director validation
func GeneratePDTSignatureSeal(dossierID, userID uuid.UUID, userRole string, signedAt time.Time) string {
	payload := fmt.Sprintf("DOSSIER:%s|PDT:%s|ROLE:%s|TIME:%s|ALCOA_PLUS:VALIDATED",
		dossierID.String(),
		userID.String(),
		userRole,
		signedAt.UTC().Format(time.RFC3339Nano),
	)
	hash := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(hash[:])
}
