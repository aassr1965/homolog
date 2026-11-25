-- =========================================================================
-- Migration 032: Jalon 2.7bis — Harmonisation Décret 21-82 & Veille Juridique
-- 1. Généralisation du titre de propriété / bail notarié ([E]) sur Imp/Dist/Exp
-- 2. Nettoyage des doublons de veille juridique et injection des 9 Piliers
-- =========================================================================

-- 1. Backfill des baux notariés / titres de propriété sur les dossiers existants
DO $$
DECLARE
    app_rec RECORD;
BEGIN
    -- Importation : IMP.ADM.04
    FOR app_rec IN SELECT id FROM corporate_applications WHERE category = 'IMPORTATION' LOOP
        INSERT INTO corporate_checklist_items (application_id, item_code, title, section, requirement_level, is_mandatory, sort_order, status)
        VALUES (app_rec.id, 'IMP.ADM.04', 'Titre de propriété notarié ou contrat de bail notarié enregistré des locaux et entrepôts de stockage (300 m² min. médicaments / 90 m² min. DM / 390 m² mixte)', 'ADMINISTRATIF', 'E', true, 35, 'manquant')
        ON CONFLICT (application_id, item_code) DO NOTHING;

        UPDATE corporate_applications
        SET completeness_percent = (
            SELECT COALESCE(ROUND(COUNT(ged_document_id)::numeric / NULLIF(COUNT(*), 0)::numeric * 100, 2), 0)
            FROM corporate_checklist_items WHERE application_id = app_rec.id
        )
        WHERE id = app_rec.id;
    END LOOP;

    -- Distribution en gros : DIS.ADM.04
    FOR app_rec IN SELECT id FROM corporate_applications WHERE category = 'DISTRIBUTION' LOOP
        INSERT INTO corporate_checklist_items (application_id, item_code, title, section, requirement_level, is_mandatory, sort_order, status)
        VALUES (app_rec.id, 'DIS.ADM.04', 'Titre de propriété notarié ou contrat de bail notarié enregistré des locaux et entrepôts de distribution', 'ADMINISTRATIF', 'E', true, 35, 'manquant')
        ON CONFLICT (application_id, item_code) DO NOTHING;

        UPDATE corporate_applications
        SET completeness_percent = (
            SELECT COALESCE(ROUND(COUNT(ged_document_id)::numeric / NULLIF(COUNT(*), 0)::numeric * 100, 2), 0)
            FROM corporate_checklist_items WHERE application_id = app_rec.id
        )
        WHERE id = app_rec.id;
    END LOOP;

    -- Exploitation : EXP.ADM.04
    FOR app_rec IN SELECT id FROM corporate_applications WHERE category = 'EXPLOITATION' LOOP
        INSERT INTO corporate_checklist_items (application_id, item_code, title, section, requirement_level, is_mandatory, sort_order, status)
        VALUES (app_rec.id, 'EXP.ADM.04', 'Titre de propriété notarié ou contrat de bail notarié enregistré du siège et des installations de l''exploitant', 'ADMINISTRATIF', 'E', true, 35, 'manquant')
        ON CONFLICT (application_id, item_code) DO NOTHING;

        UPDATE corporate_applications
        SET completeness_percent = (
            SELECT COALESCE(ROUND(COUNT(ged_document_id)::numeric / NULLIF(COUNT(*), 0)::numeric * 100, 2), 0)
            FROM corporate_checklist_items WHERE application_id = app_rec.id
        )
        WHERE id = app_rec.id;
    END LOOP;
END $$;

-- 2. Déduplication propre de regulatory_watch_items (conserver 1 exemplaire par org/réf/titre)
DELETE FROM regulatory_watch_items a
USING (
    SELECT id,
           ROW_NUMBER() OVER (PARTITION BY organization_id, legal_reference, title_fr ORDER BY created_at ASC, id ASC) AS rn
    FROM regulatory_watch_items
) b
WHERE a.id = b.id AND b.rn > 1;

-- 3. Alimentation exhaustive des 9 Piliers Réglementaires dans regulatory_watch_items
DO $$
DECLARE
    org_rec RECORD;
BEGIN
    FOR org_rec IN SELECT id FROM organizations LOOP

        -- Pilier 1 : Loi Santé 18-11 & ANPP 19-190
        INSERT INTO regulatory_watch_items (organization_id, title_fr, legal_reference, joradp_number, published_date, effective_date, category, impact_level, summary_fr, impact_analysis, compliance_status)
        SELECT org_rec.id,
               'Loi n° 18-11 du 18 Chaoual 1439 correspondant au 2 juillet 2018 relative à la santé',
               'Loi n° 18-11',
               'JORADP N° 46 du 29/07/2018',
               '2018-07-29', '2018-07-29',
               'enregistrement_amm', 'critique',
               'Loi-cadre sanitaire consacrant l''enregistrement obligatoire des médicaments, l''homologation des dispositifs médicaux, l''agrément des établissements et la responsabilité pénale du Pharmacien DT.',
               'Socle juridique du référentiel produits et du parapheur légal 1.A.7.',
               'conforme'
        WHERE NOT EXISTS (SELECT 1 FROM regulatory_watch_items WHERE organization_id = org_rec.id AND legal_reference = 'Loi n° 18-11');

        INSERT INTO regulatory_watch_items (organization_id, title_fr, legal_reference, joradp_number, published_date, effective_date, category, impact_level, summary_fr, impact_analysis, compliance_status)
        SELECT org_rec.id,
               'Décret exécutif n° 19-190 du 30 Chaoual 1440 correspondant au 3 juillet 2019 fixant les missions, l''organisation et le fonctionnement de l''Agence nationale des produits pharmaceutiques (ANPP)',
               'Décret Exécutif N° 19-190',
               'JORADP N° 43 du 07/07/2019',
               '2019-07-07', '2019-07-07',
               'enregistrement_amm', 'majeur',
               'Institue l''ANPP comme autorité indépendante d''évaluation, d''enregistrement, d''homologation, d''inspection et de contrôle qualité (LNCPP).',
               'Intégré dans le module E-TASJIL / Bordereaux BV1-BV2 et RDV ANPP.',
               'conforme'
        WHERE NOT EXISTS (SELECT 1 FROM regulatory_watch_items WHERE organization_id = org_rec.id AND legal_reference = 'Décret Exécutif N° 19-190');

        -- Pilier 2 : Établissements Pharmaceutiques (Décret 21-82 & Arrêtés 2021-2026)
        INSERT INTO regulatory_watch_items (organization_id, title_fr, legal_reference, joradp_number, published_date, effective_date, category, impact_level, summary_fr, impact_analysis, compliance_status)
        SELECT org_rec.id,
               'Décret exécutif n° 21-82 du 11 Rajab 1442 correspondant au 23 février 2021 relatif aux établissements pharmaceutiques et aux conditions de leur agrément',
               'Décret Exécutif N° 21-82',
               'JORADP N° 14 du 28/02/2021',
               '2021-02-28', '2021-02-28',
               'bpf_inspection', 'critique',
               'Définit les 5 catégories d''établissements (Fabrication, Exploitation, Importation, Distribution en gros, Exportation), l''exclusivité du PDT, l''intérim plafonné à 3 mois / 92 jours (Art. 24) et la caducité à 2 ans (Art. 16).',
               'Module Sanad Corporate complet avec checklists granulaires, intérim DT et passerelle Tabadol.',
               'conforme'
        WHERE NOT EXISTS (SELECT 1 FROM regulatory_watch_items WHERE organization_id = org_rec.id AND legal_reference = 'Décret Exécutif N° 21-82');

        INSERT INTO regulatory_watch_items (organization_id, title_fr, legal_reference, joradp_number, published_date, effective_date, category, impact_level, summary_fr, impact_analysis, compliance_status)
        SELECT org_rec.id,
               'Arrêté du 27 Safar 1443 correspondant au 5 octobre 2021 fixant les éléments du dossier d''agrément de l''établissement pharmaceutique de distribution en gros et le cahier des charges BPDG',
               'Arrêté du 5 Octobre 2021',
               'JORADP N° 08 du 31/01/2022',
               '2022-01-31', '2022-01-31',
               'bpf_inspection', 'majeur',
               'Fixe la composition du dossier d''agrément de distribution en gros, les plans 1/100e, le titre foncier/bail notarié et le cahier des charges des Bonnes Pratiques de Distribution en Gros (BPDG).',
               'Checklist DIS.ADM / DIS.TEC / DIS.BPF alignée avec exigence DIS.ADM.04.',
               'conforme'
        WHERE NOT EXISTS (SELECT 1 FROM regulatory_watch_items WHERE organization_id = org_rec.id AND legal_reference = 'Arrêté du 5 Octobre 2021');

        INSERT INTO regulatory_watch_items (organization_id, title_fr, legal_reference, joradp_number, published_date, effective_date, category, impact_level, summary_fr, impact_analysis, compliance_status)
        SELECT org_rec.id,
               'Arrêté du 15 Joumada El Oula 1443 correspondant au 20 décembre 2021 (complété le 17 septembre 2022) fixant les éléments du dossier d''agrément de l''établissement d''importation',
               'Arrêté du 20 Décembre 2021',
               'JORADP N° 04 du 16/01/2022',
               '2022-01-16', '2022-01-16',
               'bpf_inspection', 'majeur',
               'Impose les superficies minimales de stockage (300 m² médicaments dont 200 m² d''un seul tenant au sol, 90 m² DM, 390 m² mixte), la chaîne du froid qualifiée et le bail notarié.',
               'Checklist IMP.ADM.04 et contrôles de cartographie thermique intégrés.',
               'conforme'
        WHERE NOT EXISTS (SELECT 1 FROM regulatory_watch_items WHERE organization_id = org_rec.id AND legal_reference = 'Arrêté du 20 Décembre 2021');

        INSERT INTO regulatory_watch_items (organization_id, title_fr, legal_reference, joradp_number, published_date, effective_date, category, impact_level, summary_fr, impact_analysis, compliance_status)
        SELECT org_rec.id,
               'Arrêté du 22 septembre 2025 fixant les éléments du dossier d''agrément de l''établissement pharmaceutique de fabrication, son traitement et les modifications substantielles',
               'Arrêté du 22 Septembre 2025',
               'JORADP N° 01/2026',
               '2026-01-05', '2026-01-05',
               'bpf_inspection', 'critique',
               'Formalise la procédure en 2 étapes (Agrément préalable de réalisation puis Agrément d''ouverture avec Site Master File, levée des réserves et télétransmission Tabadol Intaj).',
               'Parcours Fabrication scindé en 2 étapes dans Sanad Corporate.',
               'conforme'
        WHERE NOT EXISTS (SELECT 1 FROM regulatory_watch_items WHERE organization_id = org_rec.id AND legal_reference = 'Arrêté du 22 Septembre 2025');

        INSERT INTO regulatory_watch_items (organization_id, title_fr, legal_reference, joradp_number, published_date, effective_date, category, impact_level, summary_fr, impact_analysis, compliance_status)
        SELECT org_rec.id,
               'Arrêté du 9 Rabie Ethani 1447 correspondant au 1er octobre 2025 fixant les éléments du dossier d''agrément d''un établissement pharmaceutique d''exploitation des médicaments et DM',
               'Arrêté du 1er Octobre 2025',
               'JORADP Fin 2025',
               '2025-10-15', '2025-10-15',
               'bpf_inspection', 'majeur',
               'Actualise le dossier d''agrément d''exploitation couvrant les décisions d''enregistrement et d''homologation DM, la pharmacovigilance 24/7 et les modifications substantielles.',
               'Checklist Exploitation mise à jour avec EXP.ADM.04 et QPPV.',
               'conforme'
        WHERE NOT EXISTS (SELECT 1 FROM regulatory_watch_items WHERE organization_id = org_rec.id AND legal_reference = 'Arrêté du 1er Octobre 2025');

        -- Pilier 3 : Médicaments CTD & Variations
        INSERT INTO regulatory_watch_items (organization_id, title_fr, legal_reference, joradp_number, published_date, effective_date, category, impact_level, summary_fr, impact_analysis, compliance_status)
        SELECT org_rec.id,
               'Arrêté du 12 Joumada El Oula 1442 correspondant au 27 décembre 2020 fixant la composition du dossier d''enregistrement des médicaments au format CTD',
               'Arrêté CTD du 27 Décembre 2020',
               'JORADP N° 80 du 30/12/2020',
               '2020-12-30', '2020-12-30',
               'enregistrement_amm', 'critique',
               'Rend obligatoire la structuration en 5 Modules CTD (M1 Administratif, M2 Résumés, M3 Qualité, M4 Non-Clinique, M5 Clinique & Bioéquivalence).',
               'Arborescence native Sanad CTD Studio alignée à 100%.',
               'conforme'
        WHERE NOT EXISTS (SELECT 1 FROM regulatory_watch_items WHERE organization_id = org_rec.id AND legal_reference = 'Arrêté CTD du 27 Décembre 2020');

        INSERT INTO regulatory_watch_items (organization_id, title_fr, legal_reference, joradp_number, published_date, effective_date, category, impact_level, summary_fr, impact_analysis, compliance_status)
        SELECT org_rec.id,
               'Ligne directrice relative aux modifications des décisions d''enregistrement des produits pharmaceutiques à usage humain (ANPP LD-DEPP07/V01)',
               'LD-DEPP07/V01',
               'Recueil ANPP Avril 2026',
               '2026-04-22', '2026-04-22',
               'enregistrement_amm', 'critique',
               'Établit les 34 rubriques algériennes de modifications post-AMM (Mmin, Mmod, Mmaj), le reclassement dynamique et le regroupement (Grouping Art. 09).',
               'Moteur de Variations Sanad entièrement conforme.',
               'conforme'
        WHERE NOT EXISTS (SELECT 1 FROM regulatory_watch_items WHERE organization_id = org_rec.id AND legal_reference = 'LD-DEPP07/V01');

        -- Pilier 4 : DM Modification
        INSERT INTO regulatory_watch_items (organization_id, title_fr, legal_reference, joradp_number, published_date, effective_date, category, impact_level, summary_fr, impact_analysis, compliance_status)
        SELECT org_rec.id,
               'Arrêté du 12 Dhou El Kaâda 1442 correspondant au 23 juin 2021 fixant les modalités de modification de la décision d''homologation des dispositifs médicaux',
               'Arrêté du 23 Juin 2021',
               'JORADP N° 62 du 15/08/2021',
               '2021-08-15', '2021-08-15',
               'dispositifs_medicaux', 'majeur',
               'Encadre les modifications substantielles et non-substantielles post-homologation des dispositifs médicaux (conception, matériaux, stérilisation).',
               'Géré dans le module Dispositifs Médicaux (Jalon 2.6).',
               'conforme'
        WHERE NOT EXISTS (SELECT 1 FROM regulatory_watch_items WHERE organization_id = org_rec.id AND legal_reference = 'Arrêté du 23 Juin 2021');

        -- Pilier 5 : BPF Décret 22-247
        INSERT INTO regulatory_watch_items (organization_id, title_fr, legal_reference, joradp_number, published_date, effective_date, category, impact_level, summary_fr, impact_analysis, compliance_status)
        SELECT org_rec.id,
               'Décret exécutif n° 22-247 du 26 Dhou El Kaâda 1443 correspondant au 26 juin 2022 fixant les règles de bonnes pratiques de fabrication des produits pharmaceutiques',
               'Décret Exécutif N° 22-247',
               'JORADP N° 44 du 29/06/2022',
               '2022-06-29', '2022-06-29',
               'bpf_inspection', 'critique',
               'Adopte le Guide algérien des Bonnes Pratiques de Fabrication (BPF) harmonisé PIC/S et OMS (PQS, zones ISO/Grades A-D, HVAC, eau PPI, validation des procédés).',
               'Référentiel maître de la GED Coffre-Fort Sanad Vault.',
               'conforme'
        WHERE NOT EXISTS (SELECT 1 FROM regulatory_watch_items WHERE organization_id = org_rec.id AND legal_reference = 'Décret Exécutif N° 22-247');

        -- Pilier 6 : Prix & Marges
        INSERT INTO regulatory_watch_items (organization_id, title_fr, legal_reference, joradp_number, published_date, effective_date, category, impact_level, summary_fr, impact_analysis, compliance_status)
        SELECT org_rec.id,
               'Décret exécutif n° 20-326 du 22 novembre 2020 (modifié par Décret n° 23-105) relatif au Comité économique intersectoriel des médicaments & Décret n° 98-44 (Marges)',
               'Décret Exécutif N° 20-326 / 23-105',
               'JORADP N° 78 & N° 16/2023',
               '2023-03-12', '2023-03-12',
               'redevances_tarifs', 'majeur',
               'Fixe les attributions du CEIM pour la tarification des médicaments et plafonne les marges réglementaires (production 20%, conditionnement 10%, gros 10-20%, détail 20-50%).',
               'Intégré au module Paramètres Financiers & Pricing.',
               'conforme'
        WHERE NOT EXISTS (SELECT 1 FROM regulatory_watch_items WHERE organization_id = org_rec.id AND legal_reference = 'Décret Exécutif N° 20-326 / 23-105');

        -- Pilier 7 : Pharmacovigilance & Psychotropes
        INSERT INTO regulatory_watch_items (organization_id, title_fr, legal_reference, joradp_number, published_date, effective_date, category, impact_level, summary_fr, impact_analysis, compliance_status)
        SELECT org_rec.id,
               'Décret exécutif n° 24-112 du 13 mars 2024 & Décret n° 19-379 relatifs au contrôle administratif, technique et à la classification des stupéfiants et psychotropes',
               'Décret Exécutif N° 24-112 / 19-379',
               'JORADP N° 20 du 20/03/2024',
               '2024-03-20', '2024-03-20',
               'pharmacovigilance', 'critique',
               'Encadre la classification des psychotropes (Loi n° 04-18 modifiée), les registres cotés et paraphés, le stockage sécurisé et le contrôle strict des flux.',
               'Suivi réglementaire actif dans la veille et préparation Phase 3.',
               'en_vigueur'
        WHERE NOT EXISTS (SELECT 1 FROM regulatory_watch_items WHERE organization_id = org_rec.id AND legal_reference = 'Décret Exécutif N° 24-112 / 19-379');

        -- Pilier 8 : Redevances ANPP
        INSERT INTO regulatory_watch_items (organization_id, title_fr, legal_reference, joradp_number, published_date, effective_date, category, impact_level, summary_fr, impact_analysis, compliance_status)
        SELECT org_rec.id,
               'Décret exécutif n° 21-551 du 25 Joumada El Oula 1443 correspondant au 30 décembre 2021 fixant les modalités de recouvrement de la redevance des produits pharmaceutiques et DM',
               'Décret Exécutif N° 21-551',
               'JORADP N° 02 du 09/01/2022',
               '2022-01-09', '2022-01-09',
               'redevances_tarifs', 'majeur',
               'Encadre le recouvrement et l''affectation des redevances versées à l''ANPP pour l''instruction des dossiers, expertises et contrôles.',
               'Module Liquidation Redevances & Bordereaux BV1/BV2 conforme.',
               'conforme'
        WHERE NOT EXISTS (SELECT 1 FROM regulatory_watch_items WHERE organization_id = org_rec.id AND legal_reference = 'Décret Exécutif N° 21-551');

        -- Pilier 9 : Signature Électronique & Data Integrity
        INSERT INTO regulatory_watch_items (organization_id, title_fr, legal_reference, joradp_number, published_date, effective_date, category, impact_level, summary_fr, impact_analysis, compliance_status)
        SELECT org_rec.id,
               'Loi n° 15-04 du 1er février 2015 (Signature électronique) & Loi n° 18-07 du 10 juin 2018 (Protection des données personnelles ANPDP) / 21 CFR Part 11',
               'Loi 15-04 / Loi 18-07 / 21 CFR Part 11',
               'JORADP N° 06/2015 & N° 38/2018',
               '2018-06-10', '2018-06-10',
               'bpf_inspection', 'critique',
               'Garantit la valeur probante des signatures électroniques qualifiées, la protection des données et l''intégrité ALCOA+ (Audit Trail SHA-256, 2FA, timeout 15 min).',
               'Architecture native Sanad Audit & Compliance validée.',
               'conforme'
        WHERE NOT EXISTS (SELECT 1 FROM regulatory_watch_items WHERE organization_id = org_rec.id AND legal_reference = 'Loi 15-04 / Loi 18-07 / 21 CFR Part 11');

    END LOOP;
END $$;
