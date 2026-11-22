package api

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"homolog/internal/domain"
	"homolog/internal/storage"
)

// makeDMSeedPDF génère un fichier PDF-1.4 synthétique valide pour le CAS storage
func makeDMSeedPDF(title, docCode, issuer string) string {
	return fmt.Sprintf(`%%PDF-1.4
1 0 obj
<< /Title (%s) /Author (%s) /Subject (Homologation Dispositifs Medicaux - Arrete du 10 mai 2021) /Keywords (%s) >>
endobj
2 0 obj
<< /Type /Catalog /Pages 3 0 R >>
endobj
3 0 obj
<< /Type /Pages /Kids [4 0 R] /Count 1 >>
endobj
4 0 obj
<< /Type /Page /Parent 3 0 R /MediaBox [0 0 595 842] /Contents 5 0 R >>
endobj
5 0 obj
<< /Length 120 >>
stream
BT
/F1 12 Tf
72 750 Td
(%s - Code: %s) Tj
ET
endstream
endobj
xref
0 6
0000000000 65535 f 
0000000009 00000 n 
0000000140 00000 n 
0000000188 00000 n 
0000000247 00000 n 
0000000332 00000 n 
trailer
<< /Size 6 /Root 2 0 R >>
startxref
506
%%%%EOF`, title, issuer, docCode, title, docCode)
}

// ensureDMGEDDoc vérifie ou crée un document officiel dans le Coffre-Fort GED et le CAS
func ensureDMGEDDoc(
	ctx context.Context,
	pool *pgxpool.Pool,
	storageProvider storage.StorageProvider,
	orgID, dtUserID uuid.UUID,
	code, titleFR, titleAR, taxonomy, docType, dept, filename, summary string,
	tags []string,
	isSigned bool,
) (uuid.UUID, error) {
	var docID uuid.UUID
	err := pool.QueryRow(ctx, `SELECT id FROM ged_documents WHERE organization_id = $1 AND document_code = $2`, orgID, code).Scan(&docID)
	if err == nil && docID != uuid.Nil {
		return docID, nil
	}

	content := makeDMSeedPDF(titleFR, code, "Laboratoires Saidal / ANPP DDM")
	r := bytes.NewReader([]byte(content))
	fileSHA, sizeBytes, storagePath, err := storageProvider.Save(ctx, r)
	if err != nil {
		return uuid.Nil, fmt.Errorf("sauvegarde CAS %s: %w", code, err)
	}

	docID = uuid.New()
	insertDoc := `
		INSERT INTO ged_documents (
			id, organization_id, taxonomy_path, title_fr, title_ar,
			document_code, document_type, department, tags, status,
			current_version, is_locked, created_by, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9, $10,
			'v1.0', FALSE, $11, NOW(), NOW()
		)
	`
	_, err = pool.Exec(ctx, insertDoc,
		docID, orgID, taxonomy, titleFR, titleAR,
		code, docType, dept, tags, domain.GEDStatusEnVigueur,
		dtUserID,
	)
	if err != nil {
		log.Printf("[DM Seed] Erreur insertion ged_documents %s: %v", code, err)
		return uuid.Nil, fmt.Errorf("insertion ged_documents %s: %w", code, err)
	}

	var signedBy *uuid.UUID
	var signedAt *time.Time
	var sigDecl, sigSeal, signedIP string
	if isSigned && dtUserID != uuid.Nil {
		signedBy = &dtUserID
		tNow := time.Now().UTC()
		signedAt = &tNow
		sigDecl = domain.SolemnPDTDeclarationGED
		signedIP = "127.0.0.1"
		sigSeal = domain.ComputeSignatureSeal(fileSHA, dtUserID, string(domain.RoleDirecteurTechnique), tNow, signedIP)
	}

	versionID := uuid.New()
	insertVersion := `
		INSERT INTO ged_versions (
			id, document_id, version_label, is_major_version, revision_summary,
			file_sha256, file_size_bytes, mime_type, original_filename,
			storage_path, is_signed_dt, signed_by_dt_id, signed_at,
			signature_declaration, signature_sha256, signed_ip,
			author_id, created_at
		) VALUES (
			$1, $2, 'v1.0', TRUE, $3,
			$4, $5, 'application/pdf', $6,
			$7, $8, $9, $10,
			$11, $12, $13,
			$14, NOW()
		)
		ON CONFLICT DO NOTHING
	`
	_, _ = pool.Exec(ctx, insertVersion,
		versionID, docID, summary,
		fileSHA, sizeBytes, filename,
		storagePath, isSigned, signedBy, signedAt,
		sigDecl, sigSeal, signedIP,
		dtUserID,
	)

	return docID, nil
}

// populateDossierChecklist injecte ou met à jour les éléments de checklist d'un dossier DM
func populateDossierChecklist(
	ctx context.Context,
	pool *pgxpool.Pool,
	dossierID uuid.UUID,
	items []domain.ChecklistItem,
	docMapping map[string]uuid.UUID,
	customStatus map[string]domain.ChecklistItemStatus,
	customWorkflow map[string]domain.WorkflowStatus,
	customNotes map[string]string,
) error {
	var count int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM dossier_checklist_items WHERE dossier_id = $1`, dossierID).Scan(&count)

	if count == 0 {
		insertQuery := `
			INSERT INTO dossier_checklist_items (
				id, dossier_id, module_number, sub_module, item_code,
				title_fr, title_ar, description_fr, requirement_level,
				is_mandatory, ged_document_id, status, validation_notes,
				sort_order, security_classification, assigned_department, workflow_status,
				created_at, updated_at
			) VALUES (
				$1, $2, $3, $4, $5,
				$6, $7, $8, $9,
				$10, $11, $12, $13,
				$14, $15, $16, $17,
				NOW(), NOW()
			)
		`
		for _, item := range items {
			itemID := uuid.New()
			var docID *uuid.UUID
			if gID, ok := docMapping[item.ItemCode]; ok && gID != uuid.Nil {
				docID = &gID
			}

			st := domain.ItemStatusValidated
			if s, ok := customStatus[item.ItemCode]; ok {
				st = s
			}

			wf := domain.WorkflowStatusApprovedSealed
			if w, ok := customWorkflow[item.ItemCode]; ok {
				wf = w
			}

			notes := "Pièce réglementaire vérifiée et validée conforme Décret 20-324 & Arrêté du 10 mai 2021"
			if n, ok := customNotes[item.ItemCode]; ok {
				notes = n
			}

			dept := "affaires_reglementaires"
			if item.SubModule == "DM_TECH" {
				dept = "qualite"
			}

			_, err := pool.Exec(ctx, insertQuery,
				itemID, dossierID, item.ModuleNumber, item.SubModule, item.ItemCode,
				item.TitleFR, item.TitleAR, item.DescriptionFR, item.RequirementLevel,
				item.IsMandatory, docID, st, &notes,
				item.SortOrder, domain.ClassificationStandard, dept, wf,
			)
			if err != nil {
				log.Printf("[DM Seed] Erreur insertion item %s pour dossier %s: %v", item.ItemCode, dossierID, err)
			}
		}
	} else {
		updateQuery := `
			UPDATE dossier_checklist_items SET
				ged_document_id = CASE WHEN $1::uuid IS NOT NULL THEN $1 ELSE ged_document_id END,
				status = $2,
				workflow_status = $3,
				validation_notes = $4,
				updated_at = NOW()
			WHERE dossier_id = $5 AND item_code = $6
		`
		for _, item := range items {
			var docID *uuid.UUID
			if gID, ok := docMapping[item.ItemCode]; ok && gID != uuid.Nil {
				docID = &gID
			}

			st := domain.ItemStatusValidated
			if s, ok := customStatus[item.ItemCode]; ok {
				st = s
			}

			wf := domain.WorkflowStatusApprovedSealed
			if w, ok := customWorkflow[item.ItemCode]; ok {
				wf = w
			}

			notes := "Pièce réglementaire vérifiée et validée conforme Décret 20-324 & Arrêté du 10 mai 2021"
			if n, ok := customNotes[item.ItemCode]; ok {
				notes = n
			}

			_, _ = pool.Exec(ctx, updateQuery, docID, st, wf, &notes, dossierID, item.ItemCode)
		}
	}

	// Recalculer le taux de complétude
	var totalMandatory, fulfilledMandatory int
	rows, err := pool.Query(ctx, `
		SELECT is_mandatory, requirement_level, status
		FROM dossier_checklist_items
		WHERE dossier_id = $1
	`, dossierID)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var isMan bool
			var reqLvl, st string
			if err := rows.Scan(&isMan, &reqLvl, &st); err == nil {
				if isMan && reqLvl == string(domain.RequirementExiged) {
					totalMandatory++
					if st == string(domain.ItemStatusAttached) || st == string(domain.ItemStatusValidated) || st == string(domain.ItemStatusApprovedDT) {
						fulfilledMandatory++
					}
				}
			}
		}
	}
	completeness := 0
	if totalMandatory > 0 {
		completeness = (fulfilledMandatory * 100) / totalMandatory
	}

	_, _ = pool.Exec(ctx, `UPDATE dossiers SET completeness_percent = $1, updated_at = NOW() WHERE id = $2`, completeness, dossierID)

	return nil
}

// SeedInitialDMData provisionne 3 dossiers DM complets avec documents CAS et checklist Jalon 2.6
func SeedInitialDMData(ctx context.Context, pool *pgxpool.Pool, storageProvider storage.StorageProvider, orgSlug string) {
	var orgID uuid.UUID
	err := pool.QueryRow(ctx, `SELECT id FROM organizations WHERE slug = $1`, orgSlug).Scan(&orgID)
	if err != nil {
		log.Printf("[DM Seed] Organisation %s introuvable: %v", orgSlug, err)
		return
	}

	var dtUserID uuid.UUID
	_ = pool.QueryRow(ctx, `SELECT id FROM users WHERE organization_id = $1 AND role = 'directeur_technique' LIMIT 1`, orgID).Scan(&dtUserID)

	// =========================================================================
	// 1. DISPOSITIFS MÉDICAUX (Cathéter IIa, Stent DES III, Seringue IIa locale)
	// =========================================================================

	// A. Cathéter Intraveineux Sécurisé 18G (Classe IIa, Stérile ETO - Importé)
	var catheterID uuid.UUID
	err = pool.QueryRow(ctx, `
		SELECT id FROM medical_devices 
		WHERE organization_id = $1 AND (commercial_name_fr ILIKE '%Cathéter%' OR manufacturer_reference ILIKE '%IVC%')
		LIMIT 1
	`, orgID).Scan(&catheterID)
	if err != nil {
		catheterID = uuid.New()
		_, err = pool.Exec(ctx, `
			INSERT INTO medical_devices (
				id, organization_id, commercial_name_fr, commercial_name_ar, manufacturer_reference,
				dm_class, classification_rule, origin_type, legal_manufacturer_name, legal_manufacturer_country,
				gmdn_code, is_sterile, is_single_use, sterilization_method,
				created_at, updated_at
			) VALUES (
				$1, $2, 'Cathéter Intraveineux Sécurisé 18G à Voie d''Injection', 'قسطرة وريدية آمنة 18G مع منفذ حقن',
				'IVC-18G-SECURE', 'IIa', 'Règle 2 (Invasif chirurgical temporaire)', 'imported',
				'B. Braun Melsungen AG', 'DE', '34978', TRUE, TRUE, 'Oxyde d''éthylène (ETO)',
				NOW(), NOW()
			)
		`, catheterID, orgID)
	}

	updateCatheterQuery := `
		UPDATE medical_devices SET
			composition_details = 'Corps en Polyuréthane thermoplastique (PUR) biocompatible 85%, Canule interne acier inoxydable 304L 10%, Raccord Luer-Lock Polypropylène (PP) 5%',
			has_latex = FALSE,
			has_phthalates = FALSE,
			has_active_pharma = FALSE,
			has_biological_origin = FALSE,
			has_cmr_substances = FALSE,
			risk_management_iso14971_ref = 'PGR-ISO14971-2025-V01 (Rapport d''acceptabilité des risques résiduels)',
			risk_management_status = 'conforme',
			biocompatibility_iso10993_ref = 'BIO-ISO10993-GLP-2025-08 (Cytotoxicité, Sensibilisation & Irritation)',
			biocompatibility_status = 'conforme',
			clinical_eval_iso14155_ref = 'CER-ISO14155-2025-CLIN (Évaluation de la performance clinique et équivalence)',
			clinical_eval_status = 'conforme',
			sterilization_validation_ref = 'VAL-STERIL-ISO11135-2024 (Cycle Oxyde d''Éthylène SAL 10^-6)',
			packaging_stability_iso11607_ref = 'STAB-ISO11607-36M-2025 (Barrière stérile pelable Tyvek)',
			sample_batch_pf = 'LOT-PF-2026-DM01',
			sample_batch_mp = 'LOT-MP-PUR-098',
			sample_units_count = 3,
			intended_use_fr = 'Cathétérisme veineux périphérique pour perfusion intraveineuse intermittente ou continue et injection de solutés ou médicaments.',
			sterilization_method = 'Oxyde d''éthylène (ETO)',
			is_sterile = TRUE,
			is_single_use = TRUE,
			ce_certificate_number = 'CE-0123-MDR-2023-7749',
			notified_body_name = 'TÜV SÜD Product Service GmbH',
			notified_body_number = '0123',
			iso_13485_certificate_number = 'ISO13485-DE-TUV-9921',
			clv_fsc_number = 'FSC-DE-BFARM-2023-4410',
			clv_issuing_authority = 'BfArM (Allemagne)',
			gmdn_code = '34978'
		WHERE id = $1 AND organization_id = $2
	`
	_, _ = pool.Exec(ctx, updateCatheterQuery, catheterID, orgID)

	// B. Stent Coronarien Actif au Sirolimus (Classe III, Stérile E-Beam - Importé)
	var stentID uuid.UUID
	err = pool.QueryRow(ctx, `
		SELECT id FROM medical_devices 
		WHERE organization_id = $1 AND (commercial_name_fr ILIKE '%Stent%' OR manufacturer_reference ILIKE '%DES%')
		LIMIT 1
	`, orgID).Scan(&stentID)
	if err != nil {
		stentID = uuid.New()
		_, err = pool.Exec(ctx, `
			INSERT INTO medical_devices (
				id, organization_id, commercial_name_fr, commercial_name_ar, manufacturer_reference,
				dm_class, classification_rule, origin_type, legal_manufacturer_name, legal_manufacturer_country,
				gmdn_code, is_sterile, is_single_use, sterilization_method,
				created_at, updated_at
			) VALUES (
				$1, $2, 'Stent Coronarien Actif au Sirolimus Cobalt-Chrome (DES)', 'دعامة تاجية فعالة محررة للدوائر سيموليموس',
				'DES-SIROLIMUS-COCR', 'III', 'Règle 8 (Dispositif implantable chirurgical à long terme)', 'imported',
				'Biosensors International Group / Medtronic', 'CH', '47693', TRUE, TRUE, 'Faisceau d''électrons (E-Beam)',
				NOW(), NOW()
			)
		`, stentID, orgID)
	}

	updateStentQuery := `
		UPDATE medical_devices SET
			composition_details = 'Alliage Cobalt-Chrome (L605) épaisseur 81 µm, Revêtement polymère bio-résorbable PLA/PLGA 3 µm avec Sirolimus actif à libération contrôlée (dose 1.4 µg/mm²), Système de pose cathéter Rapid Exchange (Rx) Pebax/Nylon',
			has_latex = FALSE,
			has_phthalates = FALSE,
			has_active_pharma = TRUE,
			has_biological_origin = FALSE,
			has_cmr_substances = FALSE,
			risk_management_iso14971_ref = 'PGR-DES-ISO14971-2025-V03 (Analyse des risques de thrombose, migration et resténose)',
			risk_management_status = 'conforme',
			biocompatibility_iso10993_ref = 'BIO-DES-GLP-2024-C09 (Hémocompatibilité, Génotoxicité & Thrombogénicité ISO 10993-4/10)',
			biocompatibility_status = 'conforme',
			clinical_eval_iso14155_ref = 'CER-DES-SIROLIMUS-PMCF-2025 (Étude prospective multicentrique 500 patients, suivi PMCF 5 ans)',
			clinical_eval_status = 'conforme',
			sterilization_validation_ref = 'VAL-STERIL-EBEAM-ISO11137-2024 (Faisceau d''électrons 25 kGy SAL 10^-6)',
			packaging_stability_iso11607_ref = 'STAB-ALU-TYVEK-48M (Sachet thermosoudé étanche sous atmosphère inerte d''azote)',
			sample_batch_pf = 'LOT-DES-2026-X01',
			sample_batch_mp = 'LOT-COCR-L605-B9',
			sample_units_count = 3,
			intended_use_fr = 'Implantation coronarienne transluminale pour le traitement des lésions sténotiques des artères coronaires natives de diamètre 2.25 mm à 4.0 mm.',
			sterilization_method = 'Faisceau d''électrons (E-Beam)',
			is_sterile = TRUE,
			is_single_use = TRUE,
			ce_certificate_number = 'CE-0086-MDR-2024-1189',
			notified_body_name = 'BSI Group The Netherlands B.V.',
			notified_body_number = '2797',
			iso_13485_certificate_number = 'ISO13485-BSI-MD-7712',
			clv_fsc_number = 'FSC-CH-SWISSMEDIC-2025-88',
			clv_issuing_authority = 'Swissmedic (Suisse)',
			gmdn_code = '47693'
		WHERE id = $1 AND organization_id = $2
	`
	_, _ = pool.Exec(ctx, updateStentQuery, stentID, orgID)

	// C. Seringue 3 pièces stérile 5ml Luer-Lock (Classe IIa, Stérile Gamma - Fabrication Locale)
	var syringeID uuid.UUID
	err = pool.QueryRow(ctx, `
		SELECT id FROM medical_devices 
		WHERE organization_id = $1 AND (commercial_name_fr ILIKE '%Seringue%' OR manufacturer_reference ILIKE '%SYR%')
		LIMIT 1
	`, orgID).Scan(&syringeID)
	if err != nil {
		syringeID = uuid.New()
		_, err = pool.Exec(ctx, `
			INSERT INTO medical_devices (
				id, organization_id, commercial_name_fr, commercial_name_ar,
				manufacturer_reference, dm_class, classification_rule, is_sterile, sterilization_method,
				is_single_use, intended_use_fr, gmdn_code, origin_type,
				legal_manufacturer_name, legal_manufacturer_country,
				created_at, updated_at
			) VALUES (
				$1, $2, 'Seringue 3 pièces stérile 5ml Luer-Lock', 'محقنة معقمة 5 مل ثلاث قطع ذات استعمال وحيد',
				'SYR-5ML-LL-01', 'IIa', 'Règle 2 (Dispositif médical invasif chirurgical)', TRUE, 'Rayonnement Gamma (25 kGy)',
				TRUE, 'Administration et prélèvement parentéral de solutés et médicaments.', '47017', 'fabrication_locale',
				'Laboratoires Saidal SPA - Unité Matériel Médical', 'DZ',
				NOW(), NOW()
			)
		`, syringeID, orgID)
	}

	updateSyringeQuery := `
		UPDATE medical_devices SET
			composition_details = 'Corps Polypropylène médical 75%, Piston PEHD 20%, Joint élastomère synthétique sans latex ni phtalates 5%, Aiguille acier inox 304L siliconée',
			has_latex = FALSE,
			has_phthalates = FALSE,
			has_active_pharma = FALSE,
			has_biological_origin = FALSE,
			has_cmr_substances = FALSE,
			risk_management_iso14971_ref = 'PGR-SYR-2025-V02 (Analyse AFNOR ISO 14971)',
			risk_management_status = 'conforme',
			biocompatibility_iso10993_ref = 'BIO-SYR-ISO10993-2025 (Essais d''irritation et de cytotoxicité LNCM)',
			biocompatibility_status = 'conforme',
			clinical_eval_iso14155_ref = 'CER-SYR-CLIN-2025 (Revue de données bibliographiques et retour marché)',
			clinical_eval_status = 'conforme',
			sterilization_validation_ref = 'VAL-STERIL-ISO11137-GAMMA (Irradiation Gamma Cobalt-60 dose 25 kGy)',
			packaging_stability_iso11607_ref = 'STAB-TYVEK-60M (Blister médical thermosoudé étanche, conservation 5 ans)',
			sample_batch_pf = 'LOT-SYR-2026-B1',
			sample_batch_mp = 'LOT-PP-MED-44',
			sample_units_count = 3,
			intended_use_fr = 'Administration et prélèvement parentéral de solutés et médicaments.',
			origin_type = 'fabrication_locale',
			ce_certificate_number = 'CE-0459-2024-SYR',
			notified_body_name = 'GMED France',
			notified_body_number = '0459',
			iso_13485_certificate_number = 'ISO13485-DZ-SAIDAL-2024',
			clv_fsc_number = 'MIPH-AUT-LOC-2024-099',
			clv_issuing_authority = 'MIPH (Algérie)',
			gmdn_code = '47017'
		WHERE id = $1 AND organization_id = $2
	`
	_, _ = pool.Exec(ctx, updateSyringeQuery, syringeID, orgID)

	// =========================================================================
	// 2. DOSSIER 1 : DOS-HOM-DM-2026-0042 (Cathéter 18G - 100% - RECEVABLE)
	// =========================================================================
	var dos1ID uuid.UUID
	err = pool.QueryRow(ctx, `
		SELECT id FROM dossiers 
		WHERE organization_id = $1 AND (dossier_code = 'DOS-HOM-DM-2026-0042' OR medical_device_id = $2)
		LIMIT 1
	`, orgID, catheterID).Scan(&dos1ID)

	if err != nil {
		dos1ID = uuid.New()
		_, _ = pool.Exec(ctx, `
			INSERT INTO dossiers (
				id, organization_id, internal_reference, dossier_code, procedure_type, product_type,
				status, medical_device_id, procedure_code, tracking_number, completeness_percent,
				created_at, updated_at
			) VALUES (
				$1, $2, 'HOM-DM-2026-0042', 'DOS-HOM-DM-2026-0042', 'HOMOLOGATION_DM', 'dispositif_medical',
				'RECEVABLE', $3, 'HOMOLOG_DM', 'ANPP/DM/2026/0042', 100, NOW(), NOW()
			)
		`, dos1ID, orgID, catheterID)
	}

	// Sceau PDT pour Cathéter
	seal1 := domain.GeneratePDTSignatureSeal(dos1ID, dtUserID, string(domain.RoleDirecteurTechnique), time.Now().Add(-15*24*time.Hour))
	_, _ = pool.Exec(ctx, `
		UPDATE dossiers SET
			procedure_code = 'HOMOLOG_DM',
			procedure_type = 'HOMOLOGATION_DM',
			product_type = 'dispositif_medical',
			status = 'RECEVABLE',
			tracking_number = 'ANPP/DM/2026/0042',
			completeness_percent = 100,
			is_locked_for_deposit = TRUE,
			pdt_signed_at = NOW() - INTERVAL '15 days',
			pdt_signed_by = $1,
			pdt_signature_seal = $2,
			pdt_signature_role = 'directeur_technique',
			updated_at = NOW()
		WHERE id = $3
	`, dtUserID, seal1, dos1ID)

	// Documents GED spécifiques au Cathéter 18G
	docsCath := []struct {
		ItemCode string
		Code     string
		TitleFR  string
		TitleAR  string
		Taxonomy string
		DocType  string
		Filename string
		Summary  string
	}{
		{"DM.A.1", "ES02-FORM-DM-0042", "Formulaire Officiel ES-02 de Demande d'Homologation DM - Cathéter 18G", "استمارة طلب المصادقة الرسمية على المستلزم الطبي ES-02", "01_Affaires_Reglementaires/01_03_Formulaires_Officiels", "formulaire_es02", "ES02_Homologation_Catheter_18G.pdf", "Formulaire ES-02 d'homologation visé par le Pharmacien Directeur Technique"},
		{"DM.A.2", "MIPH-AGR-SAIDAL-2024", "Agrément d'Établissement Importateur et Exploitant MIPH Saidal", "اعتماد المؤسسة المستوردة والمستغلة صيدال من وزارة الصناعة الصيدلانية", "01_Affaires_Reglementaires/01_01_Agrements_Etablissement", "agrement_miph", "MIPH_Agrement_Saidal_Exploitant.pdf", "Agrément d'exploitation pharmaceutique et importation de dispositifs médicaux"},
		{"DM.A.3", "DEC-PDT-BENKHALID-2024", "Décision Ministérielle Nominative d'Exercice du Pharmacien Directeur Technique", "قرار التعيين الرسمي للصيدلي المدير التقني المسؤول", "01_Affaires_Reglementaires/01_02_Decisions_Enregistrement", "decision_pdt", "Decision_PDT_Dr_Benkhalid.pdf", "Décision ministérielle nominative autorisant le Dr. Yacine Benkhalid à exercer comme PDT"},
		{"DM.A.4", "MANDAT-BBRAUN-SAIDAL-2025", "Contrat de Distribution Exclusive et Accord de Matériovigilance B. Braun / Saidal", "عقد التوزيع الحصري واتفاقية اليقظة مع المصنع ب براون", "05_Contrats_Tiers/05_02_Mandats_Distribution", "mandat_fabricant", "Contrat_Distribution_BBraun_Saidal.pdf", "Mandat exclusif d'importation et convention de partage des vigilances matérielles"},
		{"DM.A.5", "QUIT-BNA-HOM-DM-0042", "Quittance de Paiement BNA des Redevances d'Homologation DM (150 000 DZD)", "وصل دفع الرسوم المصرفية للمصادقة عبر البنك الوطني الجزائري", "01_Affaires_Reglementaires/01_04_Redevances_ANPP", "quittance_bna", "Quittance_BNA_150000_DZD_Catheter.pdf", "Bordereau de versement BNA certifié conforme au barème d'homologation ANPP"},
		{"DM.Q.1", "ISO13485-BBRAUN-TUV-2024", "Certificat ISO 13485:2016 TÜV SÜD Product Service GmbH", "شهادة نظام إدارة الجودة إيزو 13485 لمصانع ب براون", "02_Assurance_Qualite_BPF/02_02_Certificats_Qualite", "certificat_iso", "Certificat_ISO13485_BBraun_TUV.pdf", "Système de management de la qualité certifié pour la conception et fabrication de cathéters"},
		{"DM.Q.2", "CE-MDR-0123-BBRAUN-2023", "Certificat de Conformité CE selon Règlement (UE) 2017/745 (MDR)", "شهادة المطابقة الأوروبية لمستلزمات الحقن الوريدي", "02_Assurance_Qualite_BPF/02_02_Certificats_Qualite", "certificat_ce", "Certificat_CE_MDR_Catheter_18G.pdf", "Certificat d'examen UE du système de gestion de la qualité n° CE 0123 7749"},
		{"DM.Q.3", "FSC-BFARM-DE-2025-441", "Certificat de Vente Libre (Free Sale Certificate) BfArM Allemagne", "شهادة البيع الحر الصادرة عن الوكالة الفيدرالية الألمانية للدواء BfArM", "01_Affaires_Reglementaires/01_05_Certificats_Vente_Libre", "clv_fsc", "FSC_BfArM_Germany_Catheter.pdf", "Attestation de libre circulation et commercialisation dans l'Union Européenne"},
		{"DM.T.1", "SPEC-TECH-IVC18G-2025", "Fiche Technique, Spécifications et Schémas d'Assemblage Cathéter 18G", "الملف التقني والمخطط الهندسي لقسطرة الأوردة 18G", "03_Dossier_Technique_DM/03_01_Specifications", "fiche_technique", "Fiche_Technique_Catheter_18G_PUR.pdf", "Dimensions, débit de perfusion (95 ml/min), biocompatibilité et force de rétention d'aiguille"},
		{"DM.T.2", "ISO14971-RISK-IVC-2025", "Dossier d'Analyse et de Gestion des Risques ISO 14971", "ملف تحليل وتسيير المخاطر وفق المعيار الدولي إيزو 14971", "03_Dossier_Technique_DM/03_02_Gestion_Risques", "analyse_risques", "Rapport_Gestion_Risques_ISO14971_IVC.pdf", "Évaluation des risques résiduels d'embolie gazeuse, perforation veineuse et piqûre accidentelle"},
		{"DM.T.3", "VAL-PROC-PUR-IVC-2025", "Rapport de Qualification et Validation du Procédé d'Extrusion Polyuréthane", "تقرير التأهيل والمصادقة على عملية تصنيع وبثق البولي يوريثان", "03_Dossier_Technique_DM/03_03_Validation_Procedes", "validation_procede", "Validation_Extrusion_Canule_PUR.pdf", "Qualification opérationnelle (QO) et de performance (QP) des lignes d'extrusion en salle propre ISO 7"},
		{"DM.S.1", "VAL-STERIL-ETO-ISO11135", "Rapport de Validation du Procédé de Stérilisation Oxyde d'Éthylène (SAL 10^-6)", "ملف التحقق من طريقة التعقيم بأكسيد الإيثيلين إيزو 11135", "04_Sterilisation_Validation/04_01_Sterilisation", "validation_sterilisation", "Validation_Sterilisation_ETO_ISO11135.pdf", "Rapport de conformité microbienne SAL 10^-6 et résidus d'oxyde d'éthylène et d'éthylène chlorhydrine < 4 mg/j"},
		{"DM.S.2", "STAB-TYVEK-ISO11607-36M", "Validation de l'Intégrité de Barrière Stérile Blister Tyvek (Stabilité 36 mois)", "إثبات سلامة غلاف التعقيم تيفيك والاحتفاظ بالعقامة لمدة 36 شهرا", "04_Sterilisation_Validation/04_02_Barriere_Sterile", "stabilite_emballage", "Validation_Emballage_Tyvek_ISO11607.pdf", "Résistance au pelage, étanchéité par bullage et vieillissement accéléré équivalent 3 ans"},
		{"DM.I.1", "NOTICE-BILINGUE-IVC18G", "Notice d'Instructions et Mode d'Emploi Bilingue Français / Arabe", "دليل الاستعمال والتعليمات باللغتين العربية والفرنسية", "06_Informations_Produit/06_01_Notices", "notice_bilingue", "Notice_Emploi_Catheter_18G_FR_AR.pdf", "Recommandations de pose, précautions d'asepsie, contre-indications et schéma didactique"},
		{"DM.I.2", "ETIQ-100-IVC18G-SECURE", "Maquettes d'Étiquetage Primaire (Blister) et Secondaire (Boîte) à 100%", "عينات الوسم والتغليف بالحجم الحقيقي 100%", "06_Informations_Produit/06_02_Etiquetage", "maquettes_etiquetage", "BAT_Etiquetage_100_Catheter_18G.pdf", "Bon à tirer comportant les symboles normalisés ISO 15223-1, numéro de lot et date d'expiration"},
	}

	mapCath := make(map[string]uuid.UUID)
	for _, doc := range docsCath {
		dID, err := ensureDMGEDDoc(ctx, pool, storageProvider, orgID, dtUserID,
			doc.Code, doc.TitleFR, doc.TitleAR, doc.Taxonomy, doc.DocType, "affaires_reglementaires",
			doc.Filename, doc.Summary, []string{"DM", "Cathéter", "IIa", doc.ItemCode}, true)
		if err != nil {
			log.Printf("[DM Seed] Erreur doc cathéter %s: %v", doc.Code, err)
		} else {
			mapCath[doc.ItemCode] = dID
		}
	}

	itemsCath := domain.GenerateDMChecklist("IIa", true)
	_ = populateDossierChecklist(ctx, pool, dos1ID, itemsCath, mapCath, nil, nil, nil)

	// Évaluation de recevabilité en 13 points scellée pour le Cathéter
	upsertEvalQuery := `
		INSERT INTO dm_recevabilite_evaluations (
			id, organization_id, dossier_id, medical_device_id,
			deposit_order_number, deposit_date, electronic_copies_count, electronic_media_type,
			item_1_status, item_2_status, item_3_status, item_4_status,
			item_5_status, item_6_status, item_7_status, item_8_status,
			item_9_status, item_10_status, item_11_status, item_12_status,
			item_13_status, conformity_score, decision, reserves_notes,
			anpp_agent_name, created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3,
			'ANPP/DDM/REC/2026/0148', CURRENT_DATE - INTERVAL '14 days', 4, 'usb_sealed_sha256',
			'conforme', 'conforme', 'conforme', 'conforme',
			'conforme', 'conforme', 'conforme', 'conforme',
			'conforme', 'conforme', 'conforme', 'conforme',
			'conforme', 13, 'complet_et_recevable', 'Dossier complet et exemplaire conformément à l''Arrêté du 10 mai 2021',
			'Dr. Amel Bouzid (Sous-Direction de l''Homologation DDM)', NOW() - INTERVAL '14 days', NOW()
		)
		ON CONFLICT (dossier_id) DO UPDATE SET
			deposit_order_number = EXCLUDED.deposit_order_number,
			conformity_score = 13,
			decision = 'complet_et_recevable',
			updated_at = NOW()
	`
	_, _ = pool.Exec(ctx, upsertEvalQuery, orgID, dos1ID, catheterID)

	// =========================================================================
	// 3. DOSSIER 2 : DOS-2026-DM-STENT-DES (Stent Classe III - 100% - EN INSTRUCTION)
	// =========================================================================
	var dos2ID uuid.UUID
	err = pool.QueryRow(ctx, `
		SELECT id FROM dossiers 
		WHERE organization_id = $1 AND (dossier_code = 'DOS-2026-DM-STENT-DES' OR medical_device_id = $2)
		LIMIT 1
	`, orgID, stentID).Scan(&dos2ID)

	if err != nil {
		dos2ID = uuid.New()
		_, _ = pool.Exec(ctx, `
			INSERT INTO dossiers (
				id, organization_id, internal_reference, dossier_code, procedure_type, product_type,
				status, medical_device_id, procedure_code, tracking_number, completeness_percent,
				deposit_receipt_ref, deposit_receipt_date, target_deposit_date, created_at, updated_at
			) VALUES (
				$1, $2, 'REF-DES-SIROLIMUS-01', 'DOS-2026-DM-STENT-DES', 'HOMOLOGATION_DM', 'dispositif_medical',
				'en_instruction', $3, 'HOMOLOG_DM', 'ANPP/DDM/INST/2026/0078', 100,
				'REC-DEPOT-DDM-2026-0312', CURRENT_DATE - INTERVAL '10 days', CURRENT_DATE + INTERVAL '180 days',
				NOW() - INTERVAL '30 days', NOW()
			)
		`, dos2ID, orgID, stentID)
	}

	seal2 := domain.GeneratePDTSignatureSeal(dos2ID, dtUserID, string(domain.RoleDirecteurTechnique), time.Now().Add(-12*24*time.Hour))
	_, _ = pool.Exec(ctx, `
		UPDATE dossiers SET
			procedure_code = 'HOMOLOG_DM',
			procedure_type = 'HOMOLOGATION_DM',
			product_type = 'dispositif_medical',
			status = 'en_instruction',
			tracking_number = 'ANPP/DDM/INST/2026/0078',
			deposit_receipt_ref = 'REC-DEPOT-DDM-2026-0312',
			completeness_percent = 100,
			is_locked_for_deposit = TRUE,
			pdt_signed_at = NOW() - INTERVAL '12 days',
			pdt_signed_by = $1,
			pdt_signature_seal = $2,
			pdt_signature_role = 'directeur_technique',
			updated_at = NOW()
		WHERE id = $3
	`, dtUserID, seal2, dos2ID)

	// Documents GED spécifiques au Stent Sirolimus DES (Classe III)
	docsStent := []struct {
		ItemCode string
		Code     string
		TitleFR  string
		TitleAR  string
		Taxonomy string
		DocType  string
		Filename string
		Summary  string
	}{
		{"DM.A.1", "ES02-FORM-STENT-DES-2026", "Formulaire Officiel ES-02 d'Homologation DM Classe III - Stent Sirolimus", "استمارة طلب المصادقة الرسمية على المستلزم الطبي من الصنف الثالث دعامة تاجية", "01_Affaires_Reglementaires/01_03_Formulaires_Officiels", "formulaire_es02", "ES02_Homologation_Stent_DES.pdf", "Demande officielle d'homologation DM Classe III avec engagement de matériovigilance renforcée"},
		{"DM.A.2", "MIPH-AGR-SAIDAL-CARDIO", "Agrément d'Établissement Importateur Dispositifs Médicaux Cardiovasculaires", "اعتماد المؤسسة الصيدلانية لاستيراد المستلزمات الطبية لأمراض القلب والشرايين", "01_Affaires_Reglementaires/01_01_Agrements_Etablissement", "agrement_miph", "MIPH_Agrement_Saidal_Cardiovasculaire.pdf", "Agrément MIPH portant autorisation d'importation de dispositifs médicaux implantables actifs et invasifs"},
		{"DM.A.3", "DEC-PDT-BENKHALID-CARDIO", "Décision Ministérielle d'Habilitation Pharmacien DT pour DM Implantables", "قرار التأهيل الوزاري للصيدلي المدير التقني للمستلزمات المزروعة", "01_Affaires_Reglementaires/01_02_Decisions_Enregistrement", "decision_pdt", "Decision_PDT_Habilitation_Cardio.pdf", "Habilitation réglementaire du PDT pour la libération et traçabilité des implants cardiovasculaires"},
		{"DM.A.4", "MANDAT-DES-EXCLUSIVE-2025", "Mandat Exclusif de Distribution et Accord de Matériovigilance Stent DES", "عقد الوكالة الحصرية وتتبع اليقظة لمستلزمات القسطرة القلبية", "05_Contrats_Tiers/05_02_Mandats_Distribution", "mandat_fabricant", "Mandat_Distribution_Stent_DES.pdf", "Convention de distribution exclusive et clause de rappel d'urgence sous 24h avec traçabilité par numéro de série"},
		{"DM.A.5", "QUIT-BNA-STENT-200K", "Quittance de Paiement BNA Redevance Homologation DM Classe III (200 000 DZD)", "وصل دفع الرسوم البنكية لمصادقة مستلزم طبي صنف ثالث (200 ألف دج)", "01_Affaires_Reglementaires/01_04_Redevances_ANPP", "quittance_bna", "Quittance_BNA_200000_DZD_Stent.pdf", "Récépissé BNA pour instruction approfondie d'un DM de Classe III par la Commission Technique ANPP"},
		{"DM.Q.1", "ISO13485-BIOSENSORS-BSI", "Certificat de Système Qualité ISO 13485:2016 BSI Group Netherlands", "شهادة نظام الجودة إيزو 13485 لمصنع الدعامات القلبية", "02_Assurance_Qualite_BPF/02_02_Certificats_Qualite", "certificat_iso", "ISO13485_BSI_Biosensors_Stent.pdf", "Certification complète de conception et de fabrication stérile de stents coronariens à élution médicamenteuse"},
		{"DM.Q.2", "CE-MDR-0086-DES-DESIGN", "Certificat d'Examen CE de la Conception (Annexe IX Chapitre II MDR 2017/745)", "شهادة فحص التصميم الأوروبي CE للدعامات التاجية الفعالة", "02_Assurance_Qualite_BPF/02_02_Certificats_Qualite", "certificat_ce", "Certificat_CE_Examen_Conception_DES.pdf", "Certificat d'évaluation de la documentation technique Classe III délivré par l'organisme notifié BSI n° 2797"},
		{"DM.Q.3", "FSC-SWISSMEDIC-DES-2025", "Certificat de Vente Libre délivré par Swissmedic (Berne)", "شهادة البيع الحر الصادرة عن الوكالة السويسرية للمنتجات العلاجية سويس ميديك", "01_Affaires_Reglementaires/01_05_Certificats_Vente_Libre", "clv_fsc", "FSC_Swissmedic_Stent_DES.pdf", "Attestation certifiant que le stent DES est commercialisé librement sur le marché suisse et européen"},
		{"DM.T.1", "TECH-DES-SIROLIMUS-SPECS", "Dossier Technique du Stent CoCr, Force Radiale et Cinétique d'Élution", "الملف التقني الهندسي لدعامة الكوبالت كروم وقوة المقاومة وحركية الدواء", "03_Dossier_Technique_DM/03_01_Specifications", "fiche_technique", "Dossier_Technique_Stent_Sirolimus.pdf", "Structure cellulaire fermée/ouverte, profil de franchissement 0.038'', profil de libération in vitro du Sirolimus"},
		{"DM.T.2", "ISO14971-DES-RISK-MGMT", "Matrice d'Analyse des Risques Thrombose et Resténose Intrastent ISO 14971", "مصفوفة تحليل مخاطر التخثر وإعادة التضيق داخل الدعامة إيزو 14971", "03_Dossier_Technique_DM/03_02_Gestion_Risques", "analyse_risques", "Analyse_Risques_ISO14971_Stent_DES.pdf", "FMEA complète sur le risque de fracture des mailles, détachement du polymère et thrombose subaiguë"},
		{"DM.T.3", "VAL-COATING-SIROLIMUS", "Validation du Procédé de Micro-Pulvérisation Polymère Sirolimus et Pureté", "تقرير المصادقة على عملية طلاء الدعامة بالبوليمر والدواء الفعال", "03_Dossier_Technique_DM/03_03_Validation_Procedes", "validation_procede", "Validation_Coating_Sirolimus_PLA.pdf", "Uniformité de distribution de l'IPA, dosage par HPLC-UV (1.4 µg/mm² +/- 10%) et cinétique d'élution 30 jours"},
		{"DM.S.1", "VAL-STERIL-EBEAM-ISO11137", "Validation du Procédé de Stérilisation par Faisceau d'Électrons (E-Beam 25 kGy)", "ملف التحقق من التعقيم بواسطة حزمة الإلكترونات إيزو 11137", "04_Sterilisation_Validation/04_01_Sterilisation", "validation_sterilisation", "Validation_Sterilisation_EBeam_DES.pdf", "Dosimétrie selon ISO 11137-2 Méthode VDmax25 garantissant l'intégrité de la molécule de Sirolimus sans dégradation"},
		{"DM.S.2", "STAB-FOIL-DES-BARRIER", "Étude de Maintien de Barrière Stérile sous Atmosphère d'Azote (48 mois)", "دراسة سلامة غلاف الحماية تحت غاز النيتروجين وثبات التعقيم لمدة 48 شهرا", "04_Sterilisation_Validation/04_02_Barriere_Sterile", "stabilite_emballage", "Stabilite_Emballage_Azote_Stent.pdf", "Pochette aluminium barrière étanche à l'humidité et à l'oxygène avec absorbeur d'oxygène et pastille témoin"},
		{"DM.I.1", "NOTICE-STENT-DES-CARDIO", "Manuel d'Utilisation Cardiologie Interventionnelle Bilingue Français / Arabe", "دليل الاستعمال لأطباء أمراض القلب التداخلية باللغتين", "06_Informations_Produit/06_01_Notices", "notice_bilingue", "Manuel_Utilisation_Stent_DES_FR_AR.pdf", "Guide complet de préparation, sélection du diamètre sous angiographie/OCT, pression de déploiement et bi-thérapie antiagrégante"},
		{"DM.I.2", "ETIQ-CARTE-IMPLANT-STENT", "Carte d'Implant Patient Réglementaire et Maquettes d'Étiquetage Échelle 100%", "بطاقة زرع الدعامة للمريض وعينات الوسم الرسمية 100%", "06_Informations_Produit/06_02_Etiquetage", "maquettes_etiquetage", "Carte_Implant_Patient_Etiquetage_DES.pdf", "Carte portative d'identification d'implant pour le patient (MDR Art. 18) et étiquettes adhésives pour dossier médical"},
		{"DM.C.1", "CER-ISO14155-STENT-DES", "Rapport d'Évaluation Clinique ISO 14155 (Cohorte 500 Patients, Suivi PMCF 5 ans)", "تقرير التقييم السريري السريري إيزو 14155 لدراسة 500 مريض ومتابعة 5 سنوات", "07_Evaluation_Clinique/07_01_Rapports_Cliniques", "evaluation_clinique", "Rapport_Evaluation_Clinique_ISO14155_DES.pdf", "Étude clinique prospective randomisée multicentrique démontrant un taux d'échec de la lésion cible (TLF) < 4.1% à 12 mois"},
	}

	mapStent := make(map[string]uuid.UUID)
	for _, doc := range docsStent {
		dID, err := ensureDMGEDDoc(ctx, pool, storageProvider, orgID, dtUserID,
			doc.Code, doc.TitleFR, doc.TitleAR, doc.Taxonomy, doc.DocType, "affaires_reglementaires",
			doc.Filename, doc.Summary, []string{"DM", "Stent", "DES", "Classe_III", doc.ItemCode}, true)
		if err != nil {
			log.Printf("[DM Seed] Erreur doc stent %s: %v", doc.Code, err)
		} else {
			mapStent[doc.ItemCode] = dID
		}
	}

	itemsStent := domain.GenerateDMChecklist("III", true)
	_ = populateDossierChecklist(ctx, pool, dos2ID, itemsStent, mapStent, nil, nil, nil)

	// Évaluation de recevabilité en 13 points pour le Stent Classe III
	upsertEvalStentQuery := `
		INSERT INTO dm_recevabilite_evaluations (
			id, organization_id, dossier_id, medical_device_id,
			deposit_order_number, deposit_date, electronic_copies_count, electronic_media_type,
			item_1_status, item_2_status, item_3_status, item_4_status,
			item_5_status, item_6_status, item_7_status, item_8_status,
			item_9_status, item_10_status, item_11_status, item_12_status,
			item_13_status, conformity_score, decision, reserves_notes,
			anpp_agent_name, created_at, updated_at
		) VALUES (
			gen_random_uuid(), $1, $2, $3,
			'ANPP/DDM/REC/2026/0155', CURRENT_DATE - INTERVAL '10 days', 4, 'usb_sealed_sha256',
			'conforme', 'conforme', 'conforme', 'conforme',
			'conforme', 'conforme', 'conforme', 'conforme',
			'conforme', 'conforme', 'conforme', 'conforme',
			'conforme', 13, 'complet_et_recevable', 'Dossier technique DM Classe III recevable avec examen favorable des données cliniques',
			'Dr. Karim Cherif (Chef de Département Évaluation DDM)', NOW() - INTERVAL '10 days', NOW()
		)
		ON CONFLICT (dossier_id) DO UPDATE SET
			deposit_order_number = EXCLUDED.deposit_order_number,
			conformity_score = 13,
			decision = 'complet_et_recevable',
			anpp_agent_name = EXCLUDED.anpp_agent_name,
			reserves_notes = EXCLUDED.reserves_notes,
			updated_at = NOW()
	`
	_, _ = pool.Exec(ctx, upsertEvalStentQuery, orgID, dos2ID, stentID)

	// =========================================================================
	// 4. DOSSIER 3 : DOS-2026-SERINGUE-DM (Seringue Locale - EN PRÉPARATION)
	// =========================================================================
	var dos3ID uuid.UUID
	err = pool.QueryRow(ctx, `
		SELECT id FROM dossiers 
		WHERE organization_id = $1 AND (dossier_code = 'DOS-2026-SERINGUE-DM' OR medical_device_id = $2)
		LIMIT 1
	`, orgID, syringeID).Scan(&dos3ID)

	if err != nil {
		dos3ID = uuid.New()
		_, _ = pool.Exec(ctx, `
			INSERT INTO dossiers (
				id, organization_id, internal_reference, dossier_code, procedure_type, product_type,
				status, medical_device_id, procedure_code, tracking_number, completeness_percent,
				created_at, updated_at
			) VALUES (
				$1, $2, 'REF-SANAD-DM-03', 'DOS-2026-SERINGUE-DM', 'HOMOLOGATION_DM', 'dispositif_medical',
				'en_preparation', $3, 'HOMOLOG_DM', 'ANPP/DM/2026/0104', 87, NOW(), NOW()
			)
		`, dos3ID, orgID, syringeID)
	}

	_, _ = pool.Exec(ctx, `
		UPDATE dossiers SET
			procedure_code = 'HOMOLOG_DM',
			procedure_type = 'HOMOLOGATION_DM',
			product_type = 'dispositif_medical',
			status = 'en_preparation',
			tracking_number = 'ANPP/DM/2026/0104',
			is_locked_for_deposit = FALSE,
			updated_at = NOW()
		WHERE id = $1
	`, dos3ID)

	// Documents GED spécifiques à la Seringue 5ml locale
	docsSyr := []struct {
		ItemCode string
		Code     string
		TitleFR  string
		TitleAR  string
		Taxonomy string
		DocType  string
		Filename string
		Summary  string
	}{
		{"DM.A.1", "ES02-FORM-SERINGUE-5ML", "Formulaire ES-02 d'Homologation DM Fabrication Locale - Seringue 5ml", "استمارة طلب المصادقة على المستلزم الطبي إنتاج محلي محقنة 5 مل", "01_Affaires_Reglementaires/01_03_Formulaires_Officiels", "formulaire_es02", "ES02_Homologation_Seringue_5ml_Locale.pdf", "Demande officielle d'homologation pour un dispositif médical fabriqué sur le territoire national"},
		{"DM.A.2", "MIPH-AGR-2024-0089", "Agrément d'Exploitation et Fabrication Industrielle Saidal", "اعتماد استغلال وتصنيع المستلزمات الطبية من وزارة الصناعة الصيدلانية", "01_Affaires_Reglementaires/01_01_Agrements_Etablissement", "agrement_miph", "MIPH_Agrement_Saidal_Fabricant.pdf", "Agrément d'établissement pharmaceutique de fabrication de dispositifs médicaux stériles"},
		{"DM.A.3", "ANPP-DEC-PDT-2023", "Décision Nominative d'Exercice du Pharmacien Directeur Technique Saidal", "قرار التعيين الوزاري للصيدلي المدير التقني المسؤول", "01_Affaires_Reglementaires/01_02_Decisions_Enregistrement", "decision_pdt", "Decision_PDT_Saidal_Exercice.pdf", "Décision officielle nominative délivrée par le Ministère de l'Industrie Pharmaceutique"},
		{"DM.A.4", "ENGAG-MATERIO-SAIDAL-2026", "Charte Interne de Matériovigilance et Traçabilité des DM Saidal", "ميثاق اليقظة والتتبع الداخلي للمستلزمات الطبية صيدال", "05_Contrats_Tiers/05_03_Accords_Qualite", "charte_materiovigilance", "Charte_Materiovigilance_Saidal_2026.pdf", "Système d'assurance qualité et cellule interne d'investigation des incidents et matériovigilance"},
		{"DM.A.5", "QUIT-BNA-SERINGUE-80K", "Quittance de Paiement BNA Redevance Homologation Locale (80 000 DZD)", "وصل دفع الرسوم البنكية للمصادقة على التصنيع المحلي (80 ألف دج)", "01_Affaires_Reglementaires/01_04_Redevances_ANPP", "quittance_bna", "Quittance_BNA_80000_DZD_Seringue.pdf", "Tarif préférentiel d'encouragement de la production nationale selon la réglementation ANPP"},
		{"DM.Q.1", "ISO13485-SAIDAL-MED-2025", "Certificat ISO 13485:2016 Unité Matériel Médical Saidal", "شهادة الجودة العالمية إيزو 13485 لوحدة العتاد الطبي صيدال", "02_Assurance_Qualite_BPF/02_02_Certificats_Qualite", "certificat_iso", "ISO13485_Saidal_Materiel_Medical.pdf", "Certification ISO 13485:2016 délivrée pour les ateliers d'injection thermoplastique et conditionnement stérile"},
		{"DM.Q.2", "DECLARATION-CE-SERINGUE", "Déclaration UE de Conformité Fabricant (Annexe IV Règlement 2017/745)", "إعلان المطابقة الأوروبية لمصنع المحاقن الطبية المعقمة", "02_Assurance_Qualite_BPF/02_02_Certificats_Qualite", "declaration_ce", "Declaration_Conformite_CE_Seringue_Saidal.pdf", "Déclaration solennelle de conformité signée par le Pharmacien Directeur Technique"},
		{"DM.Q.3", "ATTEST-FAB-LOCALE-MIPH", "Attestation de Fabrication Locale et Commercialisation Nationale MIPH", "شهادة الإنتاج الوطني والترخيص بالتسويق المحلي", "01_Affaires_Reglementaires/01_05_Certificats_Vente_Libre", "attestation_nationale", "Attestation_Fabrication_Locale_MIPH.pdf", "Attestation ministérielle attestant d'un taux d'intégration local supérieur à 65%"},
		{"DM.T.1", "DOSSIER-TECH-SERINGUE-5ML", "Descriptif Technique, Caractéristiques de Glissement et Étanchéité Piston", "الملف التقني لمواصفات الانزلاق وإحكام إغلاق مكبس المحقنة", "03_Dossier_Technique_DM/03_01_Specifications", "fiche_technique", "Dossier_Technique_Seringue_5ml_PP.pdf", "Spécifications de force de rupture et de glissement selon ISO 7886-1, tolérance de graduation"},
		{"DM.T.2", "ISO14971-SERINGUE-RISK", "Matrice de Criticité des Risques Piqure et Incompatibilité Chimique", "مصفوفة تحليل مخاطر الوخز وعدم التوافق الكيميائي إيزو 14971", "03_Dossier_Technique_DM/03_02_Gestion_Risques", "analyse_risques", "Analyse_Risques_Seringue_ISO14971.pdf", "Évaluation des risques d'interaction contenant-contenu et d'erreur de dosage volumétrique"},
		{"DM.T.3", "VAL-INJECTION-SALLE-ISO8", "Validation du Procédé d'Injection Plastique en Salle Propre ISO 8", "تقرير التحقق من عملية حقن البلاستيك في غرف بيئة نظيفة إيزو 8", "03_Dossier_Technique_DM/03_03_Validation_Procedes", "validation_procede", "Validation_Injection_Salle_Blanche_ISO8.pdf", "Contrôles particulaires et microbiologiques de l'environnement de production continue"},
		{"DM.S.1", "VAL-GAMMA-ISO11137-25KGY", "Dossier de Dosimétrie et Validation de Stérilisation Gamma (Cobalt-60)", "ملف التحقق وقياس جرعات التعقيم بأشعة غاما (كوبالت 60)", "04_Sterilisation_Validation/04_01_Sterilisation", "validation_sterilisation", "Validation_Sterilisation_Gamma_25kGy.pdf", "Cartographie de dose minimale absorbée de 25 kGy et vérification de non-dégradation du polypropylène"},
		{"DM.S.2", "STAB-BLISTER-PE-TYVEK", "Étude de Vieillissement Accéléré Barrière Stérile (5 ans de péremption)", "دراسة التقادم المتسارع لغلاف التعقيم وصلاحية الحفظ 5 سنوات", "04_Sterilisation_Validation/04_02_Barriere_Sterile", "stabilite_emballage", "Stabilite_Barriere_Sterile_Seringue_5ans.pdf", "Tests d'étanchéité au colorant ASTM F1929 et traction de pelage après vieillissement accéléré à 55°C"},
		{"DM.I.1", "NOTICE-SERINGUE-5ML-BILINGUE", "Projet de Notice d'Emploi Hospitalière et Ambulatoire Bilingue FR/AR", "مشروع دليل الاستعمال الاستشفائي باللغتين العربية والفرنسية", "06_Informations_Produit/06_01_Notices", "notice_bilingue", "Projet_Notice_Seringue_5ml_Bilingue.pdf", "Instructions d'aspiration, connexion Luer-Lock sécurisée et mise au rebut dans collecteur DASRI"},
		{"DM.I.2", "ETIQ-SERINGUE-5ML-SCALE100", "Bon à Tirer (BAT) Étiquetage et Cartouche Stérile à 100%", "نموذج وسم الغلاف الفردي والعلبة الكرتونية بمقياس 100%", "06_Informations_Produit/06_02_Etiquetage", "maquettes_etiquetage", "BAT_Etiquetage_100_Seringue_5ml.pdf", "Épreuves contractuelles d'impression de l'emballage individuel pelable et de la boîte distributrice de 100 unités"},
	}

	mapSyr := make(map[string]uuid.UUID)
	for _, doc := range docsSyr {
		dID, err := ensureDMGEDDoc(ctx, pool, storageProvider, orgID, dtUserID,
			doc.Code, doc.TitleFR, doc.TitleAR, doc.Taxonomy, doc.DocType, "affaires_reglementaires",
			doc.Filename, doc.Summary, []string{"DM", "Seringue", "Fabrication_Locale", doc.ItemCode}, true)
		if err != nil {
			log.Printf("[DM Seed] Erreur doc seringue %s: %v", doc.Code, err)
		} else {
			mapSyr[doc.ItemCode] = dID
		}
	}

	itemsSyr := domain.GenerateDMChecklist("IIa", true)
	// Les items DM.I.1 et DM.I.2 sont rattachés mais laissés en attente de visa au Parapheur DT pour animer le cockpit
	customStatusSyr := map[string]domain.ChecklistItemStatus{
		"DM.I.1": domain.ItemStatusAttached,
		"DM.I.2": domain.ItemStatusAttached,
	}
	customWorkflowSyr := map[string]domain.WorkflowStatus{
		"DM.I.1": domain.WorkflowStatusDraftPending,
		"DM.I.2": domain.WorkflowStatusDraftPending,
	}
	customNotesSyr := map[string]string{
		"DM.I.1": "Notice bilingue téléversée en attente de visa final au Parapheur du Pharmacien DT",
		"DM.I.2": "BAT d'étiquetage échelle 100% en attente d'approbation réglementaire",
	}

	_ = populateDossierChecklist(ctx, pool, dos3ID, itemsSyr, mapSyr, customStatusSyr, customWorkflowSyr, customNotesSyr)

	// =========================================================================
	// 5. SUIVI TRANSITOIRE & MODIFICATIONS (DÉCRET 20-324 ARTICLES 22 À 25)
	// =========================================================================

	// Régularisation
	var regCount int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM dm_regularizations WHERE organization_id = $1`, orgID).Scan(&regCount)
	if regCount == 0 {
		regs := []struct {
			code        string
			deviceID    uuid.UUID
			derogNum    string
			marketSub   time.Duration
			status      string
			ddmAck      string
			depositSub  time.Duration
			deadlineAdd time.Duration
			imported    float64
			distrib     float64
			hospShare   int
			retShare    int
			obs         string
		}{
			{
				code:        "REG-DM-2026-001",
				deviceID:    catheterID,
				derogNum:    "DEROG-MIPH-2024-8841",
				marketSub:   -18 * 30 * 24 * time.Hour,
				status:      "commercialise_derogatoire",
				ddmAck:      "ACCUSE-DDM-PROV-4412",
				depositSub:  -15 * 24 * time.Hour,
				deadlineAdd: 22 * 24 * time.Hour,
				imported:    45000.00,
				distrib:     42800.00,
				hospShare:   80,
				retShare:    20,
				obs:         "Dispositif sous dérogation ministérielle 2024 en cours d'audit technique d'urgence avant échéance J-22.",
			},
			{
				code:        "REG-DM-2026-002",
				deviceID:    syringeID,
				derogNum:    "DEROG-MIPH-2024-9102",
				marketSub:   -12 * 30 * 24 * time.Hour,
				status:      "dossier_depose_ddm",
				ddmAck:      "ANPP/DDM/REC/2026/0092",
				depositSub:  -30 * 24 * time.Hour,
				deadlineAdd: 52 * 24 * time.Hour,
				imported:    120000.00,
				distrib:     115000.00,
				hospShare:   60,
				retShare:    40,
				obs:         "Dossier technique 6 sections déposé au guichet DDM le 22/10/2026 sous récépissé officiel.",
			},
			{
				code:        "REG-DM-2026-003",
				deviceID:    stentID,
				derogNum:    "DEROG-MIPH-2023-7115",
				marketSub:   -24 * 30 * 24 * time.Hour,
				status:      "en_instruction",
				ddmAck:      "ANPP/DDM/INST/2026/0034",
				depositSub:  -60 * 24 * time.Hour,
				deadlineAdd: 84 * 24 * time.Hour,
				imported:    30000.00,
				distrib:     29400.00,
				hospShare:   75,
				retShare:    25,
				obs:         "Commission technique d'évaluation en cours d'examen du rapport de biocompatibilité ISO 10993.",
			},
		}

		for _, reg := range regs {
			initDate := time.Now().Add(reg.marketSub)
			depDate := time.Now().Add(reg.depositSub)
			deadline := time.Now().Add(reg.deadlineAdd)

			insertRegQuery := `
				INSERT INTO dm_regularizations (
					id, organization_id, medical_device_id, regularization_code,
					derogation_number, initial_market_date, status,
					ddm_acknowledgement_number, ddm_deposit_date, legal_deadline,
					quarterly_volume_imported, quarterly_volume_distributed,
					hospital_distribution_share, retail_distribution_share,
					last_declaration_date, observations, created_at, updated_at
				) VALUES (
					gen_random_uuid(), $1, $2, $3,
					$4, $5, $6,
					$7, $8, $9,
					$10, $11,
					$12, $13,
					CURRENT_DATE, $14, NOW(), NOW()
				)
				ON CONFLICT (regularization_code) DO NOTHING
			`
			_, _ = pool.Exec(ctx, insertRegQuery,
				orgID, reg.deviceID, reg.code,
				reg.derogNum, initDate, reg.status,
				reg.ddmAck, depDate, deadline,
				reg.imported, reg.distrib,
				reg.hospShare, reg.retShare,
				reg.obs,
			)
		}
	}

	// Modifications
	var modCount int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM dm_modifications WHERE organization_id = $1`, orgID).Scan(&modCount)
	if modCount == 0 {
		insertModQuery := `
			INSERT INTO dm_modifications (
				id, organization_id, medical_device_id, dossier_id,
				modification_code, modification_type, category_code, category_label,
				title, description, justification, current_state, proposed_state,
				impact_risk_iso14971, impact_biocompatibility_iso10993, impact_sterilization_iso11135,
				impact_packaging_iso11607, impact_clinical_iso14155, impact_quality_iso13485,
				fee_amount, fee_receipt_ref, fee_payment_date,
				status, anpp_submission_date, anpp_decision_date,
				avenant_number, avenant_date, decision_notes,
				pdt_signed_at, pdt_signature_seal, pdt_signature_notes,
				created_at, updated_at
			) VALUES (
				gen_random_uuid(), $1, $2, $3,
				$4, $5, $6, $7,
				$8, $9, $10, $11, $12,
				$13, $14, $15,
				$16, $17, $18,
				$19, $20, CURRENT_DATE - INTERVAL '15 days',
				$21, CURRENT_DATE - INTERVAL '14 days', CURRENT_DATE - INTERVAL '3 days',
				$22, CURRENT_DATE - INTERVAL '3 days', $23,
				NOW() - INTERVAL '14 days', $24, $25,
				NOW() - INTERVAL '20 days', NOW()
			)
			ON CONFLICT (organization_id, modification_code) DO NOTHING
		`

		// Mod 1 : Substantielle approuvée avec Avenant (Stérilisation Gamma)
		_, _ = pool.Exec(ctx, insertModQuery,
			orgID, catheterID, dos1ID,
			"MOD-DM-2026-001", "SUBSTANTIELLE", "DM_SUB_03", "Changement du procédé ou cycle de stérilisation (ISO 11135 / ISO 11137)",
			"Substitution du procédé Oxyde d'éthylène par Rayonnement Gamma 25 kGy",
			"Transition industrielle vers irradiation Gamma Cobalt-60 pour conformité environnementale et accélération des délais de libération.",
			"Rapports de validation ISO 11137 et biocompatibilité ISO 10993 démontrant la non-altération physico-chimique du polyuréthane.",
			"Stérilisation Oxyde d'Éthylène (ETO) cycle standard 55°C, temps de désorption 7 jours en chambre ventilée.",
			"Stérilisation Rayonnement Gamma Cobalt-60 dose minimale certifiée 25 kGy avec dosimètres radiochromiques.",
			true, false, true, false, false, true,
			45000.00, "BNA-CHE-2026-8819",
			"approuve_avenant",
			"AVN-DDM-2026-0042", "Avenant accordé après revue favorable de la Sous-Direction de l'Homologation DDM.",
			"DM-AVN-SEAL-8f4b238719ac6e5d", "Attestation de validation dosimétrique certifiée conforme BPF.",
		)

		// Mod 2 : Substantielle en instruction (Matière première Polyuréthane sans DEHP)
		_, _ = pool.Exec(ctx, insertModQuery,
			orgID, catheterID, dos1ID,
			"MOD-DM-2026-002", "SUBSTANTIELLE", "DM_SUB_02", "Modification de la composition des matières premières en contact direct patient (ISO 10993)",
			"Remplacement du plastifiant de la gaine polyuréthane par un grade médical sans phtalate",
			"Élimination des résidus plastifiants DEHP au profit d'un polymère TPU aliphatic biocompatible conforme REACH/BPF.",
			"Exigence de sécurité toxicologique renforcée pour les patients soumis à perfusion prolongée.",
			"Gaine canuleuse en PUR plastifié au DEHP (teneur inférieure à 0.1% m/m).",
			"Gaine canuleuse en TPU aliphatique médical 100% exempt de phtalates et métaux lourds.",
			true, true, false, false, true, false,
			45000.00, "BNA-CHE-2026-9041",
			"en_instruction",
			"", "Dossier en cours d'évaluation toxicologique au LNCM.",
			"DM-AVN-SEAL-1c9983726ab450fe", "Dossier GLP ISO 10993 annexé sous référence BIO-2026-TPU.",
		)

		// Mod 3 : Non-substantielle approuvée (Mise à jour notice bilingue)
		_, _ = pool.Exec(ctx, insertModQuery,
			orgID, catheterID, dos1ID,
			"MOD-DM-2026-003", "NON_SUBSTANTIELLE", "DM_MIN_03", "Mise à jour mineure de l'étiquetage ou notice d'utilisation bilingue",
			"Actualisation des pictogrammes d'utilisation et ajout des mentions arabes de précaution",
			"Conformité avec les recommandations DDM sur la lisibilité des instructions de pose chez l'adulte.",
			"Amélioration de l'ergonomie d'utilisation et réduction du risque d'erreur d'insertion.",
			"Notice d'emploi version V02 bilingue français/arabe standard.",
			"Notice d'emploi enrichie version V03 avec schéma pas-à-pas et mentions de sécurité renforcées.",
			false, false, false, false, false, false,
			15000.00, "BNA-CHE-2026-9214",
			"approuve_avenant",
			"AVN-DDM-NOTIF-2026-018", "Déclaration prise d'acte par la DDM, mise à jour enregistrée au registre.",
			"DM-AVN-SEAL-77a0bc451f893d12", "Notice d'emploi révisée validée par le Pharmacien DT.",
		)
	}

	log.Println("[DM Seed] Initialisation des 3 dossiers DM avec documents CAS, checklists et conformité terminée avec succès.")
}
