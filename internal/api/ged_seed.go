package api

import (
	"bytes"
	"context"
	"log"
	"time"

	"homolog/internal/domain"
	"homolog/internal/storage"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SeedInitialGEDDocuments injecte des documents pharmaceutiques de référence réalistes dans le CAS et la base
func SeedInitialGEDDocuments(ctx context.Context, pool *pgxpool.Pool, storageProvider storage.StorageProvider, orgSlug string) {
	var orgID uuid.UUID
	err := pool.QueryRow(ctx, "SELECT id FROM organizations WHERE slug = $1", orgSlug).Scan(&orgID)
	if err != nil {
		return
	}

	// Récupérer l'ID du Directeur Technique pour signature
	var dtUserID uuid.UUID
	_ = pool.QueryRow(ctx, "SELECT id FROM users WHERE organization_id = $1 AND role = 'directeur_technique' LIMIT 1", orgID).Scan(&dtUserID)

	// Documents modèles à injecter
	seedDocs := []struct {
		TaxonomyPath    string
		TitleFR         string
		TitleAR         string
		DocumentCode    string
		DocumentType    string
		Department      string
		Tags            []string
		Status          domain.GEDDocumentStatus
		VersionLabel    string
		IsMajor         bool
		RevisionSummary string
		Filename        string
		DummyContent    string
		IsSigned        bool
	}{
		{
			TaxonomyPath:    "01_Affaires_Reglementaires/01_01_Agrements_Etablissement",
			TitleFR:         "Agrément d'Exploitation Laboratoires Saidal SPA (Site El Harrach)",
			TitleAR:         "اعتماد استغلال المخابر صيدال م.ذ.أ (موقع الحراش)",
			DocumentCode:    "MIPH-AGR-2024-0089",
			DocumentType:    "agrement_miph",
			Department:      "affaires_reglementaires",
			Tags:            []string{"MIPH", "Agrément", "Saidal", "El Harrach", "1.A.6"},
			Status:          domain.GEDStatusEnVigueur,
			VersionLabel:    "v1.0",
			IsMajor:         true,
			RevisionSummary: "Enregistrement de l'arrêté ministériel portant agrément d'exploitation pharmaceutique",
			Filename:        "MIPH-AGR-2024-0089_Saidal_El_Harrach.pdf",
			DummyContent:    "%PDF-1.4\n1 0 obj\n<< /Title (Agrement d'Exploitation MIPH Saidal El Harrach) /Author (Ministere de l'Industrie Pharmaceutique) >>\nendobj\ntrailer\n<< /Root 1 0 R >>\n%%EOF",
			IsSigned:        true,
		},
		{
			TaxonomyPath:    "01_Affaires_Reglementaires/01_02_Decisions_Enregistrement",
			TitleFR:         "Décision d'Enregistrement Ministérielle - Ciprofloxacine Saidal 500 mg",
			TitleAR:         "قرار تسجيل وزاري - سيبروفلوكساسين صيدال 500 ملغ",
			DocumentCode:    "DE-CIPRO-500-2021",
			DocumentType:    "decision_enregistrement",
			Department:      "affaires_reglementaires",
			Tags:            []string{"DE", "ANPP", "Ciprofloxacine", "500mg", "1.A.9"},
			Status:          domain.GEDStatusEnVigueur,
			VersionLabel:    "v1.0",
			IsMajor:         true,
			RevisionSummary: "Décision d'enregistrement quinquennale initiale délivrée par le Ministère de la Santé",
			Filename:        "DE-CIPRO-500-2021_Decision_Officielle.pdf",
			DummyContent:    "%PDF-1.4\n% Decision d'Enregistrement Ciprofloxacine 500 mg - Numero 923/04 R 089/21\n%%EOF",
			IsSigned:        true,
		},
		{
			TaxonomyPath:    "02_Assurance_Qualite_BPF/02_01_Manuel_Qualite_Procedures",
			TitleFR:         "Procédure Opératoire Normalisée : Libération et Traçabilité des Lots de Produits Finis",
			TitleAR:         "إجراء عمل موحد: الإفراج وتتبع حصص المنتجات الصيدلانية النهائية",
			DocumentCode:    "POS-QA-001-V2",
			DocumentType:    "sop_qualite",
			Department:      "qualite",
			Tags:            []string{"SOP", "Libération", "BPF", "Directeur Technique", "ALCOA+"},
			Status:          domain.GEDStatusEnVigueur,
			VersionLabel:    "v2.0",
			IsMajor:         true,
			RevisionSummary: "Révision majeure intégrant les exigences du nouveau Guide BPF Algérie 2025",
			Filename:        "POS-QA-001-V2_Liberation_Lots.pdf",
			DummyContent:    "%PDF-1.4\n% SOP QA-001 : Processus de Revue de Dossier de Lot et Liberation PDT\n%%EOF",
			IsSigned:        true,
		},
		{
			TaxonomyPath:    "02_Assurance_Qualite_BPF/02_01_Manuel_Qualite_Procedures",
			TitleFR:         "Procédure d'Échantillonnage et Contrôle Réception des Matières Premières Actives",
			TitleAR:         "إجراء أخذ العينات ومراقبة استلام المواد الأولية الفعالة",
			DocumentCode:    "POS-QC-014",
			DocumentType:    "sop_qualite",
			Department:      "qualite",
			Tags:            []string{"SOP", "Contrôle", "Échantillonnage", "Matières Premières"},
			Status:          domain.GEDStatusEnRevue,
			VersionLabel:    "v0.2",
			IsMajor:         false,
			RevisionSummary: "Mise à jour des critères de prélèvement sous flux laminaire",
			Filename:        "POS-QC-014_Echantillonnage_MP.docx",
			DummyContent:    "PK\x03\x04 SOP QC-014 Echantillonnage Matieres Premieres Saidal",
			IsSigned:        false,
		},
		{
			TaxonomyPath:    "02_Assurance_Qualite_BPF/02_02_Certificats_BPF_ISO",
			TitleFR:         "Certificat de Bonnes Pratiques de Fabrication (BPF/GMP) - Site Saidal Médéa",
			TitleAR:         "شهادة حسن ممارسات الصنع (GMP) - موقع صيدال المدية",
			DocumentCode:    "GMP-DZ-2023-018",
			DocumentType:    "certificat_bpf",
			Department:      "qualite",
			Tags:            []string{"BPF", "GMP", "Médéa", "MIPH", "1.M.P.2"},
			Status:          domain.GEDStatusEnVigueur,
			VersionLabel:    "v1.0",
			IsMajor:         true,
			RevisionSummary: "Certificat de conformité BPF délivré par l'Inspection Générale de la Pharmacie",
			Filename:        "GMP-DZ-2023-018_Saidal_Medea.pdf",
			DummyContent:    "%PDF-1.4\n% Certificat GMP Site Saidal Medea - Formes Seches et Liquides\n%%EOF",
			IsSigned:        true,
		},
		{
			TaxonomyPath:    "03_Developpement_Production/03_01_Formules_Fabrication_DMF",
			TitleFR:         "Drug Master File (Partie Ouverte / DPSA) - Ciprofloxacine HCl Monohydrate",
			TitleAR:         "الملف الفني للمادة الفعالة (القسم المفتوح) - سيبروفلوكساسين كلوريدات",
			DocumentCode:    "DMF-CIPRO-API-2023",
			DocumentType:    "dmf_substance_active",
			Department:      "production",
			Tags:            []string{"DMF", "DPSA", "Substance Active", "Aurobindo", "3.2.S"},
			Status:          domain.GEDStatusEnVigueur,
			VersionLabel:    "v1.0",
			IsMajor:         true,
			RevisionSummary: "Dossier technique de substance active déposé par le fabricant de matière première",
			Filename:        "DMF-CIPRO-API-2023_Open_Part.pdf",
			DummyContent:    "%PDF-1.4\n% Drug Master File Ciprofloxacin Hydrochloride - DMF No. 2023/DZ/API/441\n%%EOF",
			IsSigned:        false,
		},
		{
			TaxonomyPath:    "03_Developpement_Production/03_03_Etudes_Stabilite",
			TitleFR:         "Rapport Final de Stabilité en Temps Réel 36 Mois (Zone Climatique IVb : 30°C / 75% HR)",
			TitleAR:         "تقرير نهائي لدراسات الثبات بالزمن الفعلي 36 شهرا (المنطقة المناخية IVb)",
			DocumentCode:    "STAB-CIPRO-IVB-36M",
			DocumentType:    "rapport_stabilite",
			Department:      "production",
			Tags:            []string{"Stabilité", "Zone IVb", "36 Mois", "Temps Réel", "3.2.P.8"},
			Status:          domain.GEDStatusEnApprobationDT,
			VersionLabel:    "v0.9",
			IsMajor:         false,
			RevisionSummary: "Rapport complet d'étude de stabilité sous climatisation tropicale pour soumission ANPP",
			Filename:        "STAB-CIPRO-IVB-36M_Rapport_Final.pdf",
			DummyContent:    "%PDF-1.4\n% Rapport de Stabilite Zone IVb - 36 Mois - Lots industriels 21CIP001, 21CIP002, 21CIP003\n%%EOF",
			IsSigned:        false,
		},
		{
			TaxonomyPath:    "01_Affaires_Reglementaires/01_05_Notices_Articles_Conditionnement",
			TitleFR:         "Notice Patient Bilingue Arabe/Français et Maquette Étui - Amoxicilline Saidal 1g",
			TitleAR:         "نشرة المريض ثنائية اللغة عربية/فرنسية ونموذج علبة التوضيب - أموكسيسيلين صيدال 1 غ",
			DocumentCode:    "NOT-AMOX-1G-BILINGUE",
			DocumentType:    "notice_conditionnement",
			Department:      "affaires_reglementaires",
			Tags:            []string{"Notice", "Bilingue", "Packaging", "Amoxicilline", "1.P.2"},
			Status:          domain.GEDStatusEnVigueur,
			VersionLabel:    "v1.0",
			IsMajor:         true,
			RevisionSummary: "Maquettes d'impression à l'échelle 100% avec mentions obligatoires décret 20-324",
			Filename:        "NOT-AMOX-1G-BILINGUE_Maquettes_Couleur.pdf",
			DummyContent:    "%PDF-1.4\n% Notice d'emballage bilingue Saidal Amoxicilline 1g comprimes pellicules\n%%EOF",
			IsSigned:        true,
		},
		{
			TaxonomyPath:    "05_Contrats_Tiers/05_01_Accords_Qualite_QTA",
			TitleFR:         "Quality Technical Agreement (QTA) de Façonnage - Saidal / Sanofi Algérie",
			TitleAR:         "اتفاقية الجودة التقنية للمناولة والتصنيع - صيدال / سانوفي الجزائر",
			DocumentCode:    "QTA-SAIDAL-SANOFI-2024",
			DocumentType:    "accord_qualite",
			Department:      "qualite",
			Tags:            []string{"QTA", "Façonnage", "Sanofi", "Saidal", "BPF", "1.A.8"},
			Status:          domain.GEDStatusEnVigueur,
			VersionLabel:    "v1.0",
			IsMajor:         true,
			RevisionSummary: "Convention bipartite de délégation de fabrication et libération de lots selon Annexe 16 BPF",
			Filename:        "QTA-SAIDAL-SANOFI-2024_Approuve.pdf",
			DummyContent:    "%PDF-1.4\n% Quality Technical Agreement Saidal SPA & Sanofi Aventis Algerie SPA\n%%EOF",
			IsSigned:        true,
		},
	}

	for _, d := range seedDocs {
		var docID uuid.UUID
		err := pool.QueryRow(ctx, "SELECT id FROM ged_documents WHERE organization_id = $1 AND document_code = $2", orgID, d.DocumentCode).Scan(&docID)
		if err != nil {
			// Sauvegarde du fichier physique dans le CAS
			r := bytes.NewReader([]byte(d.DummyContent))
			fileSHA, sizeBytes, storagePath, err := storageProvider.Save(ctx, r)
			if err != nil {
				log.Printf("[ERREUR SEED GED] Sauvegarde CAS échouée pour %s: %v", d.DocumentCode, err)
				continue
			}

			newDocID := uuid.New()
			insertDoc := `
				INSERT INTO ged_documents (
					id, organization_id, taxonomy_path, title_fr, title_ar,
					document_code, document_type, department, tags, status,
					current_version, is_locked, created_by, created_at, updated_at
				) VALUES (
					$1, $2, $3, $4, $5,
					$6, $7, $8, $9, $10,
					$11, FALSE, $12, NOW(), NOW()
				)
			`
			_, err = pool.Exec(ctx, insertDoc,
				newDocID, orgID, d.TaxonomyPath, d.TitleFR, d.TitleAR,
				d.DocumentCode, d.DocumentType, d.Department, d.Tags, d.Status,
				d.VersionLabel, dtUserID,
			)
			if err != nil {
				log.Printf("[ERREUR SEED GED] Insertion document %s: %v", d.DocumentCode, err)
				continue
			}

			// Signature PDT si requise
			var signedBy *uuid.UUID
			var signedAt *time.Time
			var sigDecl, sigSeal, signedIP string
			if d.IsSigned && dtUserID != uuid.Nil {
				signedBy = &dtUserID
				tNow := time.Now().UTC()
				signedAt = &tNow
				sigDecl = domain.SolemnPDTDeclarationGED
				signedIP = "127.0.0.1"
				sigSeal = domain.ComputeSignatureSeal(fileSHA, dtUserID, string(domain.RoleDirecteurTechnique), tNow, signedIP)
			}

			insertVersion := `
				INSERT INTO ged_versions (
					id, document_id, version_label, is_major_version, revision_summary,
					file_sha256, file_size_bytes, mime_type, original_filename,
					storage_path, is_signed_dt, signed_by_dt_id, signed_at,
					signature_declaration, signature_sha256, signed_ip,
					author_id, created_at
				) VALUES (
					$1, $2, $3, $4, $5,
					$6, $7, 'application/pdf', $8,
					$9, $10, $11, $12,
					$13, $14, $15,
					$16, NOW()
				)
			`
			_, err = pool.Exec(ctx, insertVersion,
				uuid.New(), newDocID, d.VersionLabel, d.IsMajor, d.RevisionSummary,
				fileSHA, sizeBytes, d.Filename,
				storagePath, d.IsSigned, signedBy, signedAt,
				sigDecl, sigSeal, signedIP,
				dtUserID,
			)
			if err == nil {
				log.Printf("[Seed GED] Document créé dans le coffre-fort : [%s] %s (%s)", d.DocumentCode, d.TitleFR, d.VersionLabel)
			}
		}
	}
	log.Println("[Seed GED] Documents modèles du coffre-fort initialisés avec succès.")
}
