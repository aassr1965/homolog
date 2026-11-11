package domain

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Déclaration solennelle officielle du Pharmacien Directeur Technique (PDT) pour scellement GED BPF
const SolemnPDTDeclarationGED = "En qualité de Pharmacien Directeur Technique, j'atteste sur l'honneur l'exactitude, la véracité et la complétude des données du présent document conformément aux règles de Bonnes Pratiques de Fabrication (BPF) et à la réglementation pharmaceutique en vigueur."

// GEDFolder représente un sous-dossier standardisé de la GED pharmaceutique (Section 8 du CDC)
type GEDFolder struct {
	Path              string   `json:"path"`
	Code              string   `json:"code"`
	NameFR            string   `json:"name_fr"`
	NameAR            string   `json:"name_ar"`
	Description       string   `json:"description"`
	AllowedExtensions []string `json:"allowed_extensions"`
	DocumentCount     int      `json:"document_count,omitempty"`
}

// GEDBranch représente une des 5 branches maîtresses de la GED pharmaceutique
type GEDBranch struct {
	Code        string      `json:"code"`
	NameFR      string      `json:"name_fr"`
	NameAR      string      `json:"name_ar"`
	Description string      `json:"description"`
	Folders     []GEDFolder `json:"folders"`
	TotalDocs   int         `json:"total_docs,omitempty"`
}

// GEDDocumentDetailed représente un document complet avec métadonnées ALCOA+ et état de verrouillage
type GEDDocumentDetailed struct {
	ID                uuid.UUID         `json:"id"`
	OrganizationID    uuid.UUID         `json:"organization_id"`
	TaxonomyPath      string            `json:"taxonomy_path"`
	TitleFR           string            `json:"title_fr"`
	TitleAR           string            `json:"title_ar"`
	DocumentCode      string            `json:"document_code"`
	DocumentType      string            `json:"document_type"`
	Department        string            `json:"department"`
	Tags              []string          `json:"tags"`
	Status            GEDDocumentStatus `json:"status"`
	CurrentVersion    string            `json:"current_version"`
	ExpiryDate        *time.Time        `json:"expiry_date,omitempty"`
	IsLocked          bool              `json:"is_locked"`
	LockedBy          *uuid.UUID        `json:"locked_by,omitempty"`
	LockedByName      string            `json:"locked_by_name,omitempty"`
	LockedAt          *time.Time        `json:"locked_at,omitempty"`
	LockReason        string            `json:"lock_reason,omitempty"`
	EstablishmentID   *uuid.UUID        `json:"establishment_id,omitempty"`
	EstablishmentName string            `json:"establishment_name,omitempty"`
	SiteID            *uuid.UUID        `json:"site_id,omitempty"`
	SiteName          string            `json:"site_name,omitempty"`
	MedicineID        *uuid.UUID        `json:"medicine_id,omitempty"`
	MedicineBrandName string            `json:"medicine_brand_name,omitempty"`
	MedicalDeviceID   *uuid.UUID        `json:"medical_device_id,omitempty"`
	MedicalDeviceName string            `json:"medical_device_name,omitempty"`
	CreatedBy         *uuid.UUID        `json:"created_by,omitempty"`
	CreatedByName     string            `json:"created_by_name,omitempty"`
	CreatedAt         time.Time         `json:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"`
	LatestVersion     *GEDVersionDetail `json:"latest_version,omitempty"`
	VersionsCount     int               `json:"versions_count,omitempty"`
}

// GEDVersionDetail représente une révision physique immuable avec scellement ALCOA+
type GEDVersionDetail struct {
	ID                   uuid.UUID  `json:"id"`
	DocumentID           uuid.UUID  `json:"document_id"`
	VersionLabel         string     `json:"version_label"`
	IsMajorVersion       bool       `json:"is_major_version"`
	RevisionSummary      string     `json:"revision_summary"`
	FileSHA256           string     `json:"file_sha256"`
	FileSizeBytes        int64      `json:"file_size_bytes"`
	MimeType             string     `json:"mime_type"`
	OriginalFilename     string     `json:"original_filename"`
	StoragePath          string     `json:"storage_path"`
	IsSignedDT           bool       `json:"is_signed_dt"`
	SignedByDTID         *uuid.UUID `json:"signed_by_dt_id,omitempty"`
	SignedByDTName       string     `json:"signed_by_dt_name,omitempty"`
	SignedAt             *time.Time `json:"signed_at,omitempty"`
	SignatureDeclaration string     `json:"signature_declaration,omitempty"`
	SignatureSHA256      string     `json:"signature_sha256,omitempty"`
	SignedIP             string     `json:"signed_ip,omitempty"`
	AuthorID             *uuid.UUID `json:"author_id,omitempty"`
	AuthorName           string     `json:"author_name,omitempty"`
	CreatedAt            time.Time  `json:"created_at"`
}

// GetStandardGEDTaxonomy retourne la taxonomie normalisée en 5 branches et 20 sous-dossiers
func GetStandardGEDTaxonomy() []GEDBranch {
	defaultExtensions := []string{".pdf", ".docx", ".doc", ".xlsx", ".xls", ".zip"}

	return []GEDBranch{
		{
			Code:        "01",
			NameFR:      "01. Affaires Réglementaires",
			NameAR:      "01. الشؤون التنظيمية والصيدلانية",
			Description: "Agréments d'exploitation, décisions d'enregistrement/homologation, dossiers CTD et correspondances officielles avec l'ANPP",
			Folders: []GEDFolder{
				{
					Path:              "01_Affaires_Reglementaires/01_01_Agrements_Etablissement",
					Code:              "01_01",
					NameFR:            "01.01 Agréments Établissement",
					NameAR:            "اعتمادات المؤسسة الصيدلانية",
					Description:       "Agréments ministériels MIPH de l'établissement demandeur, des façonniers et sous-traitants agréés",
					AllowedExtensions: defaultExtensions,
				},
				{
					Path:              "01_Affaires_Reglementaires/01_02_Decisions_Enregistrement",
					Code:              "01_02",
					NameFR:            "01.02 Décisions d'Enregistrement",
					NameAR:            "قرارات تسجيل الأدوية",
					Description:       "Décisions d'enregistrement (DE) initiales, renouvellements quinquennaux et avenants ministériels",
					AllowedExtensions: defaultExtensions,
				},
				{
					Path:              "01_Affaires_Reglementaires/01_03_Decisions_Homologation",
					Code:              "01_03",
					NameFR:            "01.03 Décisions d'Homologation",
					NameAR:            "قرارات المصادقة على المستلزمات الطبية",
					Description:       "Décisions d'homologation des dispositifs médicaux (DM et DMDIV) selon le décret exécutif 20-324",
					AllowedExtensions: defaultExtensions,
				},
				{
					Path:              "01_Affaires_Reglementaires/01_04_Dossiers_CTD_Reference",
					Code:              "01_04",
					NameFR:            "01.04 Dossiers CTD de Référence",
					NameAR:            "ملفات التسجيل التقنية المرجعية CTD",
					Description:       "Dossiers maîtres CTD archivés, dossiers de variations majeures/mineures et fiches de synthèse",
					AllowedExtensions: defaultExtensions,
				},
				{
					Path:              "01_Affaires_Reglementaires/01_05_Notices_Articles_Conditionnement",
					Code:              "01_05",
					NameFR:            "01.05 Notices & Articles de Conditionnement",
					NameAR:            "النشرات ومواد التغليف والتوضيب",
					Description:       "Notices bilingues FR/AR approuvées, étiquetages, maquettes couleurs à l'échelle 100% et braille",
					AllowedExtensions: defaultExtensions,
				},
				{
					Path:              "01_Affaires_Reglementaires/01_06_Correspondances_ANPP",
					Code:              "01_06",
					NameFR:            "01.06 Correspondances & Notifications ANPP",
					NameAR:            "المراسلات والإخطارات مع الوكالة الوطنية",
					Description:       "Courriers de compléments, notifications de recevabilité, accusés de réception et convocations aux commissions",
					AllowedExtensions: defaultExtensions,
				},
			},
		},
		{
			Code:        "02",
			NameFR:      "02. Assurance Qualité & BPF",
			NameAR:      "02. ضمان الجودة وحسن ممارسات الصنع",
			Description: "Manuel qualité, procédures opératoires normalisées (SOP), certificats BPF/GMP, audits et déviations CAPA",
			Folders: []GEDFolder{
				{
					Path:              "02_Assurance_Qualite_BPF/02_01_Manuel_Qualite_Procedures",
					Code:              "02_01",
					NameFR:            "02.01 Manuel Qualité & Procédures (SOP)",
					NameAR:            "دليل الجودة وإجراءات العمل الموحدة",
					Description:       "Manuel qualité du laboratoire, procédures opératoires normalisées (POS/SOP) et instructions de travail",
					AllowedExtensions: defaultExtensions,
				},
				{
					Path:              "02_Assurance_Qualite_BPF/02_02_Certificats_BPF_ISO",
					Code:              "02_02",
					NameFR:            "02.02 Certificats BPF & ISO",
					NameAR:            "شهادات حسن ممارسات الصنع وISO",
					Description:       "Certificats BPF/GMP des fabricants de substances actives et de produits finis, normes ISO 9001 et ISO 13485",
					AllowedExtensions: defaultExtensions,
				},
				{
					Path:              "02_Assurance_Qualite_BPF/02_03_Rapports_Audit_Inspections",
					Code:              "02_03",
					NameFR:            "02.03 Rapports d'Audit & Inspections",
					NameAR:            "تقارير التدقيق والتفتيش الصيدلاني",
					Description:       "Rapports d'inspection de l'ANPP ou du MIPH, audits qualité internes et audits fournisseurs de matières premières",
					AllowedExtensions: defaultExtensions,
				},
				{
					Path:              "02_Assurance_Qualite_BPF/02_04_CAPA_Deviations",
					Code:              "02_04",
					NameFR:            "02.04 CAPA & Gestion des Déviations",
					NameAR:            "الإجراءات التصحيحية والوقائية والانحرافات",
					Description:       "Fiches d'investigation des déviations, actions correctives et préventives (CAPA) et revues périodiques de conformité",
					AllowedExtensions: defaultExtensions,
				},
			},
		},
		{
			Code:        "03",
			NameFR:      "03. Développement & Production",
			NameAR:      "03. التطوير والإنتاج الصيدلاني",
			Description: "Formules cadres, dossiers de lots maîtres, validations analytiques, études de stabilité et dossiers techniques DM",
			Folders: []GEDFolder{
				{
					Path:              "03_Developpement_Production/03_01_Formules_Fabrication_DMF",
					Code:              "03_01",
					NameFR:            "03.01 Formules de Fabrication & DMF",
					NameAR:            "صيغ التصنيع وملفات المواد الفعالة",
					Description:       "Formules cadres, dossiers de lots industriels de validation (3.2.P.3.5) et dossiers Drug Master File (DMF/DPSA)",
					AllowedExtensions: defaultExtensions,
				},
				{
					Path:              "03_Developpement_Production/03_02_Methodes_Controle_Validation",
					Code:              "03_02",
					NameFR:            "03.02 Méthodes de Contrôle & Validations",
					NameAR:            "طرق المراقبة والتحقق التحليلي",
					Description:       "Procédures d'analyse physico-chimique et microbiologique, rapports de validation analytique selon l'ICH Q2",
					AllowedExtensions: defaultExtensions,
				},
				{
					Path:              "03_Developpement_Production/03_03_Etudes_Stabilite",
					Code:              "03_03",
					NameFR:            "03.03 Études de Stabilité (Zone IVa/IVb)",
					NameAR:            "دراسات الثبات وحفظ الأدوية",
					Description:       "Protocoles et rapports de stabilité en temps réel et accéléré en conditions climatiques Zone IVa/IVb",
					AllowedExtensions: defaultExtensions,
				},
				{
					Path:              "03_Developpement_Production/03_04_Specifications_Matieres_PF",
					Code:              "03_04",
					NameFR:            "03.04 Spécifications Matières & Produit Fini",
					NameAR:            "دفاتر مواصفات المواد والمنتج النهائي",
					Description:       "Spécifications de libération des substances actives, excipients, articles de conditionnement et produit fini",
					AllowedExtensions: defaultExtensions,
				},
				{
					Path:              "03_Developpement_Production/03_05_Dossiers_Techniques_DM",
					Code:              "03_05",
					NameFR:            "03.05 Dossiers Techniques DM",
					NameAR:            "الملفات التقنية للمستلزمات الطبية",
					Description:       "Dossiers techniques de marquage et de conformité des dispositifs médicaux selon l'arrêté du 10 mai 2021",
					AllowedExtensions: defaultExtensions,
				},
			},
		},
		{
			Code:        "04",
			NameFR:      "04. Vigilance & Post-Market",
			NameAR:      "04. اليقظة الصيدلانية ومتابعة ما بعد التسويق",
			Description: "Pharmacovigilance (PSUR), matériovigilance des DM et procédures de gestion des retraits ou rappels de lots",
			Folders: []GEDFolder{
				{
					Path:              "04_Vigilance_Post_Market/04_01_Pharmacovigilance",
					Code:              "04_01",
					NameFR:            "04.01 Pharmacovigilance & PSUR",
					NameAR:            "اليقظة الدوائية وتقارير السلامة الدورية",
					Description:       "Rapports périodiques actualisés de sécurité (RPS/PSUR), fiches de notification d'effets indésirables et plans de gestion des risques (PGR)",
					AllowedExtensions: defaultExtensions,
				},
				{
					Path:              "04_Vigilance_Post_Market/04_02_Materiovigilance",
					Code:              "04_02",
					NameFR:            "04.02 Matériovigilance",
					NameAR:            "اليقظة الخاصة بالمستلزمات الطبية",
					Description:       "Signalements d'incidents ou de risques d'incidents graves liés à l'utilisation des dispositifs médicaux",
					AllowedExtensions: defaultExtensions,
				},
				{
					Path:              "04_Vigilance_Post_Market/04_03_Rappels_Lots",
					Code:              "04_03",
					NameFR:            "04.03 Gestion des Rappels & Retraits de Lots",
					NameAR:            "إجراءات سحب واسترجاع الحصص الدوائية",
					Description:       "Procédures d'urgence de retrait de lots, procès-verbaux de destruction et rapports d'investigation transmis à l'ANPP",
					AllowedExtensions: defaultExtensions,
				},
			},
		},
		{
			Code:        "05",
			NameFR:      "05. Contrats Tiers & Façonnage",
			NameAR:      "05. العقود مع الغير والمناولة",
			Description: "Accords qualité (QTA), contrats de fabrication ou conditionnement sous-traités et attestations d'API (CEP/APIMF)",
			Folders: []GEDFolder{
				{
					Path:              "05_Contrats_Tiers/05_01_Accords_Qualite_QTA",
					Code:              "05_01",
					NameFR:            "05.01 Accords Qualité (QTA)",
					NameAR:            "اتفاقيات الجودة التقنية مع الشركاء",
					Description:       "Quality Technical Agreements définissant les responsabilités BPF entre titulaire, fabricant et sous-traitants",
					AllowedExtensions: defaultExtensions,
				},
				{
					Path:              "05_Contrats_Tiers/05_02_Contrats_Faconnage_SousTraitance",
					Code:              "05_02",
					NameFR:            "05.02 Contrats de Façonnage & Sous-Traitance",
					NameAR:            "عقود التصنيع بالمناولة والتحاليل الخارجية",
					Description:       "Contrats juridiques d'exploitation, de fabrication, de conditionnement et de contrôle sous-traité approuvés",
					AllowedExtensions: defaultExtensions,
				},
				{
					Path:              "05_Contrats_Tiers/05_03_Certificats_Fournisseurs_APIs",
					Code:              "05_03",
					NameFR:            "05.03 Certificats CEP & Accès APIs",
					NameAR:            "شهادات المطابقة الأوروبية ورسائل الوصول للمواد",
					Description:       "Certificats de conformité à la Pharmacopée Européenne (CEP EDQM), lettres d'accès APIMF et déclarations EST/OGM",
					AllowedExtensions: defaultExtensions,
				},
			},
		},
	}
}

// ValidateTaxonomyPath vérifie si le chemin de taxonomie appartient bien à la nomenclature officielle
func ValidateTaxonomyPath(path string) bool {
	for _, b := range GetStandardGEDTaxonomy() {
		for _, f := range b.Folders {
			if f.Path == path {
				return true
			}
		}
	}
	return false
}

// GetFolderByPath retourne les métadonnées d'un dossier de taxonomie par son chemin
func GetFolderByPath(path string) *GEDFolder {
	for _, b := range GetStandardGEDTaxonomy() {
		for _, f := range b.Folders {
			if f.Path == path {
				folder := f
				return &folder
			}
		}
	}
	return nil
}

// ValidateDoc valide les contraintes fondamentales d'un document GED
func (d *GEDDocumentDetailed) Validate() error {
	if strings.TrimSpace(d.TitleFR) == "" {
		return errors.New("le titre français du document est obligatoire")
	}
	if !ValidateTaxonomyPath(d.TaxonomyPath) {
		return fmt.Errorf("chemin de taxonomie invalide : %s", d.TaxonomyPath)
	}
	if strings.TrimSpace(d.DocumentCode) == "" {
		return errors.New("le code d'identification documentaire est obligatoire")
	}
	return nil
}

// CanTransitionTo vérifie si le changement d'état ALCOA+ est autorisé selon le rôle utilisateur
func (d *GEDDocumentDetailed) CanTransitionTo(target GEDDocumentStatus, role UserRole) (bool, string) {
	if d.Status == target {
		return false, "le document est déjà dans cet état"
	}

	switch d.Status {
	case GEDStatusBrouillon:
		if target == GEDStatusEnRevue {
			return true, ""
		}
		return false, "un brouillon ne peut passer qu'à l'état 'en_revue'"

	case GEDStatusEnRevue:
		if target == GEDStatusBrouillon {
			return true, "" // rejet de la revue
		}
		if target == GEDStatusEnApprobationDT {
			return true, ""
		}
		return false, "un document en revue ne peut passer qu'à 'en_approbation_dt' ou revenir en 'brouillon'"

	case GEDStatusEnApprobationDT:
		if target == GEDStatusEnVigueur {
			if role != RoleDirecteurTechnique && role != RoleAdminLaboratoire {
				return false, "seul le Pharmacien Directeur Technique peut approuver et mettre en vigueur un document officiel"
			}
			return true, ""
		}
		if target == GEDStatusEnRevue || target == GEDStatusBrouillon {
			return true, "" // rejet par le DT
		}
		return false, "un document en approbation ne peut passer qu'à 'en_vigueur' (sur signature PDT) ou être renvoyé en revue"

	case GEDStatusEnVigueur:
		if target == GEDStatusObsolete || target == GEDStatusArchive {
			if role != RoleDirecteurTechnique && role != RoleResponsableAffairesReg && role != RoleAdminLaboratoire {
				return false, "seul le Directeur Technique ou le Responsable Réglementaire peut rendre obsolète ou archiver un document en vigueur"
			}
			return true, ""
		}
		return false, "un document en vigueur ne peut passer qu'à l'état 'obsolete' ou 'archive'"

	case GEDStatusObsolete:
		if target == GEDStatusArchive {
			return true, ""
		}
		return false, "un document obsolète ne peut qu'être archivé"

	case GEDStatusArchive:
		return false, "un document archivé ne peut plus changer d'état (ALCOA+ Immuable)"

	default:
		return false, "transition d'état inconnue"
	}
}

// GenerateNextVersionLabel incrémente le numéro de version selon la norme BPF
// v0.1 -> v0.2 (brouillon) | v0.x -> v1.0 (première mise en vigueur) | v1.0 -> v1.1 (mineure) | v1.0 -> v2.0 (majeure)
func GenerateNextVersionLabel(current string, isMajor bool, isFirstApproval bool) string {
	if strings.TrimSpace(current) == "" {
		return "v0.1"
	}

	if isFirstApproval {
		return "v1.0"
	}

	re := regexp.MustCompile(`^v?(\d+)\.(\d+)$`)
	matches := re.FindStringSubmatch(current)
	if len(matches) != 3 {
		return "v1.0"
	}

	maj, _ := strconv.Atoi(matches[1])
	min, _ := strconv.Atoi(matches[2])

	if isMajor {
		return fmt.Sprintf("v%d.0", maj+1)
	}

	// Révision mineure
	return fmt.Sprintf("v%d.%d", maj, min+1)
}

// ComputeSignatureSeal calcule le hachage cryptographique SHA-256 de scellement ALCOA+
// liant l'empreinte physique du fichier, l'identité du signataire, le rôle, l'horodatage UTC et l'adresse IP
func ComputeSignatureSeal(fileSHA256 string, userID uuid.UUID, userRole string, timestamp time.Time, ipAddress string) string {
	payload := fmt.Sprintf("%s|%s|%s|%s|%s",
		fileSHA256,
		userID.String(),
		userRole,
		timestamp.UTC().Format(time.RFC3339Nano),
		ipAddress,
	)
	hash := sha256.Sum256([]byte(payload))
	return hex.EncodeToString(hash[:])
}
