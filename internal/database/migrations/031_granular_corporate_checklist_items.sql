-- =========================================================================
-- Migration 031: Découpage granulaire des exigences Décret 21-82
-- Séparation des dossiers administratifs, techniques et DT en sous-rubriques
-- =========================================================================

-- Mise à niveau des dossiers existants de catégorie FABRICATION (OUVERTURE_INITIALE ou AGREMENT_INITIAL)
DO $$
DECLARE
    app_rec RECORD;
    v_statuts_ged_id UUID;
BEGIN
    FOR app_rec IN 
        SELECT id FROM corporate_applications 
        WHERE category = 'FABRICATION' 
          AND nature IN ('OUVERTURE_INITIALE', 'AGREMENT_INITIAL')
    LOOP
        -- Sauvegarder l'ID du document GED rattaché aux statuts notariés si présent
        SELECT ged_document_id INTO v_statuts_ged_id 
        FROM corporate_checklist_items 
        WHERE application_id = app_rec.id 
          AND (item_code = 'FAB.OUV.01' OR item_code = 'FAB.ADM.01')
          AND ged_document_id IS NOT NULL 
        LIMIT 1;

        -- Supprimer les anciens items de checklist composites
        DELETE FROM corporate_checklist_items WHERE application_id = app_rec.id;

        -- Insérer les 22 exigences granulaires
        -- 1. Dossier Administratif & Juridique
        INSERT INTO corporate_checklist_items (application_id, item_code, title, section, requirement_level, is_mandatory, sort_order, ged_document_id, status)
        VALUES
        (app_rec.id, 'FAB.ADM.01', 'Statuts notariés de la société enregistrés et certifiés conformes', 'ADMINISTRATIF', 'E', true, 10, v_statuts_ged_id, CASE WHEN v_statuts_ged_id IS NOT NULL THEN 'rattache' ELSE 'manquant' END),
        (app_rec.id, 'FAB.ADM.02', 'Registre du Commerce électronique CNRC (code activité fabrication pharmaceutique)', 'ADMINISTRATIF', 'E', true, 20, NULL, 'manquant'),
        (app_rec.id, 'FAB.ADM.03', 'Extrait de rôle fiscal apuré (validité < 3 mois, sans échéancier ni dette contestée)', 'ADMINISTRATIF', 'E', true, 30, NULL, 'manquant'),
        (app_rec.id, 'FAB.ADM.04', 'Carte d''immatriculation fiscale (NIF) et attestation d''identification statistique (NIS)', 'ADMINISTRATIF', 'E', true, 40, NULL, 'manquant'),
        (app_rec.id, 'FAB.ADM.05', 'Attestations de mise à jour des cotisations sociales CNAS et CASNOS (< 3 mois)', 'ADMINISTRATIF', 'E', true, 50, NULL, 'manquant'),
        (app_rec.id, 'FAB.ADM.06', 'Titre de propriété foncière notarié ou contrat de bail notarié d''une durée minimale de 10 ans enregistré', 'ADMINISTRATIF', 'E', true, 60, NULL, 'manquant'),
        (app_rec.id, 'FAB.ADM.07', 'Police d''assurance en responsabilité civile professionnelle couvrant les risques industriels et d''exploitation', 'ADMINISTRATIF', 'E', true, 70, NULL, 'manquant'),

        -- 2. Conception Architecturale & Zonage BPF
        (app_rec.id, 'FAB.PLN.01', 'Plan de masse général côté et plans de coupes architecturales as-built certifiés conformes', 'TECHNIQUE_LOCAUX', 'E', true, 80, NULL, 'manquant'),
        (app_rec.id, 'FAB.PLN.02', 'Plan de zonage des classes de propreté particulaire et microbiologique (ISO 5 à 8 / Grades A à D)', 'TECHNIQUE_LOCAUX', 'E', true, 90, NULL, 'manquant'),
        (app_rec.id, 'FAB.FLX.01', 'Synoptique des flux des matières premières (API, excipients) et articles de conditionnement', 'TECHNIQUE_LOCAUX', 'E', true, 100, NULL, 'manquant'),
        (app_rec.id, 'FAB.FLX.02', 'Synoptique des flux du personnel et gestion des sas vestiaires (habillage/déshabillage)', 'TECHNIQUE_LOCAUX', 'E', true, 110, NULL, 'manquant'),
        (app_rec.id, 'FAB.FLX.03', 'Synoptique d''évacuation des déchets industriels, effluents et produits sous quarantaine/rejet', 'TECHNIQUE_LOCAUX', 'E', true, 120, NULL, 'manquant'),

        -- 3. Utilités Critiques & Équipements Industriels
        (app_rec.id, 'FAB.UTL.01', 'Rapport de qualification des systèmes de traitement d''air CVC/HVAC (surpressions en cascade et filtration HEPA H14)', 'TECHNIQUE_LOCAUX', 'E', true, 130, NULL, 'manquant'),
        (app_rec.id, 'FAB.UTL.02', 'Rapports de qualification et de validation des boucles d''eau purifiée et d''eau PPI (Pharmacopée Européenne)', 'TECHNIQUE_LOCAUX', 'E', true, 140, NULL, 'manquant'),
        (app_rec.id, 'FAB.EQP.01', 'Dossiers de Qualification QI/QO/QP des équipements majeurs de production et de conditionnement', 'TECHNIQUE_LOCAUX', 'E', true, 150, NULL, 'manquant'),

        -- 4. Système Qualité Pharmaceutique & BPF
        (app_rec.id, 'FAB.PQS.01', 'Manuel du Système de Gestion de la Qualité Pharmaceutique (PQS) et procédures maîtresses de libération des lots', 'ORGANISATION_BPF', 'E', true, 160, NULL, 'manquant'),
        (app_rec.id, 'FAB.PQS.02', 'Organigramme BPF nominatif garantissant l''indépendance hiérarchique du Contrôle Qualité vis-à-vis de la Production', 'ORGANISATION_BPF', 'E', true, 170, NULL, 'manquant'),

        -- 5. Direction Technique & Personnel Clé (Art. 16 à 18 Décret 21-82)
        (app_rec.id, 'FAB.PDT.01', 'Diplôme d''État de Docteur en Pharmacie du Pharmacien Directeur Technique (certifié conforme)', 'PHARMACIEN_DT', 'E', true, 180, NULL, 'manquant'),
        (app_rec.id, 'FAB.PDT.02', 'Attestation d''inscription au tableau de l''Ordre des Pharmaciens (Conseil National de l''Ordre CNOP)', 'PHARMACIEN_DT', 'E', true, 190, NULL, 'manquant'),
        (app_rec.id, 'FAB.PDT.03', 'Attestation d''expérience professionnelle préalable minimale en milieu industriel pharmaceutique (Art. 17 Décret 21-82)', 'PHARMACIEN_DT', 'E', true, 200, NULL, 'manquant'),
        (app_rec.id, 'FAB.PDT.04', 'Contrat de travail à durée indéterminée exclusif à temps plein du Pharmacien DT (Art. 18 Décret 21-82)', 'PHARMACIEN_DT', 'E', true, 210, NULL, 'manquant'),
        (app_rec.id, 'FAB.PDT.05', 'Contrats de travail et diplômes des Pharmaciens Assistants / Adjoints avec inscription à l''Ordre', 'PHARMACIEN_DT', 'C', true, 220, NULL, 'manquant');

        -- Recalculer le pourcentage de complétude
        UPDATE corporate_applications
        SET completeness_percent = (
            SELECT ROUND(COUNT(ged_document_id)::numeric / COUNT(*)::numeric * 100, 2)
            FROM corporate_checklist_items
            WHERE application_id = app_rec.id
        )
        WHERE id = app_rec.id;
    END LOOP;
END $$;
