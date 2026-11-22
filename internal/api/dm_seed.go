package api

import (
	"context"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SeedInitialDMData provisionne les données de démonstration pour le Jalon 2.6
func SeedInitialDMData(ctx context.Context, pool *pgxpool.Pool, orgSlug string) {
	var orgID uuid.UUID
	err := pool.QueryRow(ctx, `SELECT id FROM organizations WHERE slug = $1`, orgSlug).Scan(&orgID)
	if err != nil {
		log.Printf("[DM Seed] Organisation %s introuvable: %v", orgSlug, err)
		return
	}

	// 1. Récupérer un dispositif médical existant ou l'initialiser
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
		if err != nil {
			log.Printf("[DM Seed] Impossible d'insérer le cathéter: %v", err)
			return
		}
	}

	// 2. Mettre à jour les 6 sections techniques du cathéter
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

	// 3. Vérifier s'il existe un dossier d'homologation DM pour ce DM, sinon le créer
	var dossierID uuid.UUID
	err = pool.QueryRow(ctx, `
		SELECT id FROM dossiers 
		WHERE organization_id = $1 AND medical_device_id = $2
		LIMIT 1
	`, orgID, catheterID).Scan(&dossierID)

	if err != nil {
		dossierID = uuid.New()
		_, err = pool.Exec(ctx, `
			INSERT INTO dossiers (
				id, organization_id, internal_reference, dossier_code, procedure_type, product_type,
				status, medical_device_id, procedure_code, created_at, updated_at
			) VALUES (
				$1, $2, 'HOM-DM-2026-0042', 'DOS-DM-2026-0042', 'HOMOLOGATION_DM', 'dispositif_medical',
				'RECEVABLE', $3, 'HOMOLOG_DM', NOW(), NOW()
			)
		`, dossierID, orgID, catheterID)
		if err != nil {
			log.Printf("[DM Seed] Impossible de créer le dossier DM: %v", err)
		}
	}

	if dossierID != uuid.Nil {
		// Insérer ou mettre à jour l'évaluation de recevabilité en 13 points
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
				'ANPP/DDM/REC/2026/0148', CURRENT_DATE, 4, 'usb_sealed_sha256',
				'conforme', 'conforme', 'conforme', 'conforme',
				'conforme', 'conforme', 'conforme', 'conforme',
				'conforme', 'conforme', 'conforme', 'conforme',
				'conforme', 13, 'complet_et_recevable', 'Dossier complet et exemplaire conformément à l''Arrêté du 10 mai 2021',
				'Dr. Amel Bouzid (Sous-Direction de l''Homologation DDM)', NOW(), NOW()
			)
			ON CONFLICT (dossier_id) DO UPDATE SET
				deposit_order_number = EXCLUDED.deposit_order_number,
				conformity_score = 13,
				decision = 'complet_et_recevable',
				updated_at = NOW()
		`
		_, _ = pool.Exec(ctx, upsertEvalQuery, orgID, dossierID, catheterID)
	}

	// 4. Seeder les DM sous mesures transitoires (Régularisation Décret 20-324)
	// Vérifier si des enregistrements existent déjà
	var regCount int
	_ = pool.QueryRow(ctx, `SELECT COUNT(*) FROM dm_regularizations WHERE organization_id = $1`, orgID).Scan(&regCount)
	if regCount == 0 {
		// Créer un second DM pour avoir plusieurs cas représentatifs
		var syringeID uuid.UUID
		err = pool.QueryRow(ctx, `
			SELECT id FROM medical_devices 
			WHERE organization_id = $1 AND (commercial_name_fr ILIKE '%Seringue%' OR manufacturer_reference ILIKE '%SYR%')
			LIMIT 1
		`, orgID).Scan(&syringeID)

		if err != nil {
			// Insérer le DM Seringue
			syringeID = uuid.New()
			insertSyringeQuery := `
				INSERT INTO medical_devices (
					id, organization_id, commercial_name_fr, commercial_name_ar,
					manufacturer_reference, dm_class, classification_rule, is_sterile, sterilization_method,
					is_single_use, intended_use_fr, gmdn_code, origin_type,
					composition_details, has_latex, has_phthalates, has_active_pharma, has_biological_origin, has_cmr_substances,
					risk_management_iso14971_ref, risk_management_status,
					biocompatibility_iso10993_ref, biocompatibility_status,
					clinical_eval_iso14155_ref, clinical_eval_status,
					sterilization_validation_ref, packaging_stability_iso11607_ref,
					sample_batch_pf, sample_batch_mp, sample_units_count,
					status, created_at, updated_at
				) VALUES (
					$1, $2, 'Seringue 3 pièces stérile 5ml Luer-Lock', 'محقنة معقمة 5 مل ثلاث قطع ذات استعمال وحيد',
					'SYR-5ML-LL-01', 'IIa', 'Règle 2 (Dispositif médical invasif chirurgical)', TRUE, 'Rayonnement Gamma (25 kGy)',
					TRUE, 'Administration et prélèvement parentéral de solutés et médicaments.', '47017', 'fabrication_locale',
					'Corps en Polypropylène (PP) médical transparent 75%, Piston en Polyéthylène haute densité (PEHD) 20%, Joint d''étanchéité élastomère sans latex 5%',
					FALSE, FALSE, FALSE, FALSE, FALSE,
					'PGR-SYR-2025-V02', 'conforme',
					'BIO-SYR-ISO10993-2025', 'conforme',
					'CER-SYR-CLIN-2025', 'conforme',
					'VAL-STERIL-ISO11137-GAMMA', 'STAB-TYVEK-60M',
					'LOT-SYR-2026-B1', 'LOT-PP-MED-44', 3,
					'en_cours', NOW(), NOW()
				)
			`
			_, _ = pool.Exec(ctx, insertSyringeQuery, syringeID, orgID)
		}

		// Insérer les 3 cas de suivi de régularisation transitoire
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
				marketSub:   -18 * 30 * 24 * time.Hour, // 18 mois avant
				status:      "commercialise_derogatoire",
				ddmAck:      "ACCUSE-DDM-PROV-4412",
				depositSub:  -15 * 24 * time.Hour,
				deadlineAdd: 22 * 24 * time.Hour, // J-22 (Critique !)
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
				deadlineAdd: 52 * 24 * time.Hour, // J-52 (Urgent !)
				imported:    120000.00,
				distrib:     115000.00,
				hospShare:   60,
				retShare:    40,
				obs:         "Dossier technique 6 sections déposé au guichet DDM le 22/10/2026 sous récépissé officiel.",
			},
			{
				code:        "REG-DM-2026-003",
				deviceID:    catheterID,
				derogNum:    "DEROG-MIPH-2023-7115",
				marketSub:   -24 * 30 * 24 * time.Hour,
				status:      "en_instruction",
				ddmAck:      "ANPP/DDM/INST/2026/0034",
				depositSub:  -60 * 24 * time.Hour,
				deadlineAdd: 84 * 24 * time.Hour, // J-84 (Alerte !)
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

	// 5. Seeder les modifications post-homologation (Décret 20-324 Articles 22 à 25)
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
			orgID, catheterID, dossierID,
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
			orgID, catheterID, dossierID,
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
			orgID, catheterID, dossierID,
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

	log.Println("[DM Seed] Initialisation des données DM Homologation, Régularisation & Modifications terminée avec succès.")
}
