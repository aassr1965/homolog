-- =========================================================================
-- Migration 034: Découpage granulaire unitaire (1 PDF = 1 point) pour toutes
-- les catégories Décret 21-82 (Préalable Fabrication, Importation,
-- Distribution, Exploitation, Exportation)
-- =========================================================================

DO $$
DECLARE
    app_rec RECORD;
BEGIN
    -- 1. Fabrication - Agrément Préalable de Réalisation (10 points unitaires)
    FOR app_rec IN
        SELECT id FROM corporate_applications
        WHERE category = 'FABRICATION' AND nature = 'PREALABLE_REALISATION'
          AND NOT EXISTS (SELECT 1 FROM corporate_checklist_items WHERE application_id = corporate_applications.id AND item_code = 'FAB.PR.10')
    LOOP
        DELETE FROM corporate_checklist_items WHERE application_id = app_rec.id AND ged_document_id IS NULL;

        INSERT INTO corporate_checklist_items (application_id, item_code, title, section, requirement_level, is_mandatory, sort_order, status)
        VALUES
        (app_rec.id, 'FAB.PR.01', 'Statuts notariés de la société enregistrés et certifiés conformes', 'ADMINISTRATIF', 'E', true, 10, 'manquant'),
        (app_rec.id, 'FAB.PR.02', 'Registre du Commerce électronique CNRC (code activité fabrication pharmaceutique)', 'ADMINISTRATIF', 'E', true, 20, 'manquant'),
        (app_rec.id, 'FAB.PR.03', 'Acte de propriété notarié ou contrat de bail notarié d''une durée minimale de 10 ans enregistré', 'ADMINISTRATIF', 'E', true, 30, 'manquant'),
        (app_rec.id, 'FAB.PR.04', 'Certificat d''urbanisme en cours de validité', 'ADMINISTRATIF', 'E', true, 40, 'manquant'),
        (app_rec.id, 'FAB.PR.05', 'Permis de construire délivré par l''autorité compétente', 'ADMINISTRATIF', 'E', true, 50, 'manquant'),
        (app_rec.id, 'FAB.PR.06', 'Étude d''impact sur l''environnement approuvée par les services compétents', 'TECHNIQUE_LOCAUX', 'E', true, 60, 'manquant'),
        (app_rec.id, 'FAB.PR.07', 'Plans de masse et plans de zonage (niveaux de propreté ISO 5 à ISO 8 / Grades A à D)', 'TECHNIQUE_LOCAUX', 'E', true, 70, 'manquant'),
        (app_rec.id, 'FAB.PR.08', 'Schémas des flux matières, personnels et déchets (élimination stricte des flux croisés)', 'TECHNIQUE_LOCAUX', 'E', true, 80, 'manquant'),
        (app_rec.id, 'FAB.PR.09', 'Liste des formes galéniques envisagées et procédés de fabrication prévisionnels', 'ORGANISATION_BPF', 'E', true, 90, 'manquant'),
        (app_rec.id, 'FAB.PR.10', 'Étude technico-économique et plan de financement validé', 'ADMINISTRATIF', 'E', true, 100, 'manquant')
        ON CONFLICT (application_id, item_code) DO UPDATE SET title = EXCLUDED.title, sort_order = EXCLUDED.sort_order;

        UPDATE corporate_applications SET completeness_percent = (
            SELECT COALESCE(ROUND(COUNT(ged_document_id)::numeric / NULLIF(COUNT(*), 0)::numeric * 100, 2), 0)
            FROM corporate_checklist_items WHERE application_id = app_rec.id
        ) WHERE id = app_rec.id;
    END LOOP;

    -- 2. Importation (15 points unitaires)
    FOR app_rec IN
        SELECT id FROM corporate_applications
        WHERE category = 'IMPORTATION'
          AND NOT EXISTS (SELECT 1 FROM corporate_checklist_items WHERE application_id = corporate_applications.id AND item_code = 'IMP.PDT.04')
    LOOP
        DELETE FROM corporate_checklist_items WHERE application_id = app_rec.id AND ged_document_id IS NULL;

        INSERT INTO corporate_checklist_items (application_id, item_code, title, section, requirement_level, is_mandatory, sort_order, status)
        VALUES
        (app_rec.id, 'IMP.ADM.01', 'Statuts notariés de la société enregistrés et certifiés conformes', 'ADMINISTRATIF', 'E', true, 10, 'manquant'),
        (app_rec.id, 'IMP.ADM.02', 'Registre du Commerce électronique CNRC (code activité importation pharmaceutique / DM)', 'ADMINISTRATIF', 'E', true, 20, 'manquant'),
        (app_rec.id, 'IMP.ADM.03', 'Extrait de rôle fiscal apuré (validité < 3 mois, sans échéancier ni dette contestée)', 'ADMINISTRATIF', 'E', true, 30, 'manquant'),
        (app_rec.id, 'IMP.ADM.04', 'Carte d''immatriculation fiscale (NIF) et attestation d''identification statistique (NIS)', 'ADMINISTRATIF', 'E', true, 40, 'manquant'),
        (app_rec.id, 'IMP.ADM.05', 'Attestations de mise à jour des cotisations sociales CNAS et CASNOS (< 3 mois)', 'ADMINISTRATIF', 'E', true, 50, 'manquant'),
        (app_rec.id, 'IMP.ADM.06', 'Titre de propriété notarié ou contrat de bail notarié enregistré des locaux et entrepôts de stockage (300 m² min. médicaments / 90 m² min. DM / 390 m² mixte)', 'ADMINISTRATIF', 'E', true, 60, 'manquant'),
        (app_rec.id, 'IMP.ADM.07', 'Programme prévisionnel d''importation annuel visé par la Direction de la Pharmacie MIPH', 'ADMINISTRATIF', 'E', true, 70, 'manquant'),
        (app_rec.id, 'IMP.PDT.01', 'Diplôme d''État de Docteur en Pharmacie du Pharmacien Directeur Technique (certifié conforme)', 'PHARMACIEN_DT', 'E', true, 80, 'manquant'),
        (app_rec.id, 'IMP.PDT.02', 'Attestation d''inscription au tableau de l''Ordre des Pharmaciens (CNOP) du Pharmacien DT', 'PHARMACIEN_DT', 'E', true, 90, 'manquant'),
        (app_rec.id, 'IMP.PDT.03', 'Contrat de travail à durée indéterminée exclusif à temps plein du Pharmacien DT (Art. 18 Décret 21-82)', 'PHARMACIEN_DT', 'E', true, 100, 'manquant'),
        (app_rec.id, 'IMP.PDT.04', 'Décision nominative d''exercice du Pharmacien Directeur Technique délivrée par le MIPH', 'PHARMACIEN_DT', 'E', true, 110, 'manquant'),
        (app_rec.id, 'IMP.TEC.01', 'Descriptif et plans des magasins de stockage sous température contrôlée et zones sécurisées (quarantaine/rejet sous clé)', 'CONDITIONS_TECHNIQUES', 'E', true, 120, 'manquant'),
        (app_rec.id, 'IMP.TEC.02', 'Rapports de qualification et cartographies des chambres froides (+2°C à +8°C) et zones climatisées (+15°C à +25°C)', 'CONDITIONS_TECHNIQUES', 'E', true, 130, 'manquant'),
        (app_rec.id, 'IMP.BPF.01', 'Procédure de contrôle documentaire, échantillonnage et libération des lots importés par le Pharmacien DT', 'ORGANISATION_BPF', 'E', true, 140, 'manquant'),
        (app_rec.id, 'IMP.LOG.01', 'Conventions d''assurance qualité avec les transitaires et transporteurs sous chaîne du froid validée', 'CONDITIONS_TECHNIQUES', 'E', true, 150, 'manquant')
        ON CONFLICT (application_id, item_code) DO UPDATE SET title = EXCLUDED.title, sort_order = EXCLUDED.sort_order;

        UPDATE corporate_applications SET completeness_percent = (
            SELECT COALESCE(ROUND(COUNT(ged_document_id)::numeric / NULLIF(COUNT(*), 0)::numeric * 100, 2), 0)
            FROM corporate_checklist_items WHERE application_id = app_rec.id
        ) WHERE id = app_rec.id;
    END LOOP;

    -- 3. Distribution en gros (13 points unitaires)
    FOR app_rec IN
        SELECT id FROM corporate_applications
        WHERE category = 'DISTRIBUTION'
          AND NOT EXISTS (SELECT 1 FROM corporate_checklist_items WHERE application_id = corporate_applications.id AND item_code = 'DIS.TEC.03')
    LOOP
        DELETE FROM corporate_checklist_items WHERE application_id = app_rec.id AND ged_document_id IS NULL;

        INSERT INTO corporate_checklist_items (application_id, item_code, title, section, requirement_level, is_mandatory, sort_order, status)
        VALUES
        (app_rec.id, 'DIS.ADM.01', 'Statuts notariés de la société enregistrés et certifiés conformes', 'ADMINISTRATIF', 'E', true, 10, 'manquant'),
        (app_rec.id, 'DIS.ADM.02', 'Registre du Commerce électronique CNRC (code activité distribution en gros)', 'ADMINISTRATIF', 'E', true, 20, 'manquant'),
        (app_rec.id, 'DIS.ADM.03', 'Extrait de rôle fiscal apuré (validité < 3 mois, sans échéancier ni dette contestée)', 'ADMINISTRATIF', 'E', true, 30, 'manquant'),
        (app_rec.id, 'DIS.ADM.04', 'Attestations de mise à jour des cotisations sociales CNAS et CASNOS (< 3 mois)', 'ADMINISTRATIF', 'E', true, 40, 'manquant'),
        (app_rec.id, 'DIS.ADM.05', 'Cahier des charges de distribution en gros (BPDG) paraphé et engagement d''approvisionnement continu du marché', 'ADMINISTRATIF', 'E', true, 50, 'manquant'),
        (app_rec.id, 'DIS.ADM.06', 'Titre de propriété notarié ou contrat de bail notarié enregistré des locaux et entrepôts de distribution', 'ADMINISTRATIF', 'E', true, 60, 'manquant'),
        (app_rec.id, 'DIS.PDT.01', 'Diplôme d''État de Docteur en Pharmacie du Pharmacien Directeur Technique (certifié conforme)', 'PHARMACIEN_DT', 'E', true, 70, 'manquant'),
        (app_rec.id, 'DIS.PDT.02', 'Attestation d''inscription au tableau de l''Ordre des Pharmaciens (CNOP) du Pharmacien DT', 'PHARMACIEN_DT', 'E', true, 80, 'manquant'),
        (app_rec.id, 'DIS.PDT.03', 'Contrat de travail à durée indéterminée exclusif à temps plein du Pharmacien DT', 'PHARMACIEN_DT', 'E', true, 90, 'manquant'),
        (app_rec.id, 'DIS.TEC.01', 'Plans et descriptif des entrepôts au 1/100e selon les Bonnes Pratiques de Distribution en Gros (BPDG)', 'CONDITIONS_TECHNIQUES', 'E', true, 100, 'manquant'),
        (app_rec.id, 'DIS.TEC.02', 'Inventaire de la flotte logistique isotherme/frigorifique dédiée à la distribution', 'CONDITIONS_TECHNIQUES', 'E', true, 110, 'manquant'),
        (app_rec.id, 'DIS.TEC.03', 'Certificats d''étalonnage périodique des enregistreurs de température et sondes embarquées', 'CONDITIONS_TECHNIQUES', 'E', true, 120, 'manquant'),
        (app_rec.id, 'DIS.BPF.01', 'Système informatique validé de traçabilité des lots et procédure de rappel d''urgence (ICH Q10)', 'ORGANISATION_BPF', 'E', true, 130, 'manquant')
        ON CONFLICT (application_id, item_code) DO UPDATE SET title = EXCLUDED.title, sort_order = EXCLUDED.sort_order;

        UPDATE corporate_applications SET completeness_percent = (
            SELECT COALESCE(ROUND(COUNT(ged_document_id)::numeric / NULLIF(COUNT(*), 0)::numeric * 100, 2), 0)
            FROM corporate_checklist_items WHERE application_id = app_rec.id
        ) WHERE id = app_rec.id;
    END LOOP;

    -- 4. Exploitation (11 points unitaires)
    FOR app_rec IN
        SELECT id FROM corporate_applications
        WHERE category = 'EXPLOITATION'
          AND NOT EXISTS (SELECT 1 FROM corporate_checklist_items WHERE application_id = corporate_applications.id AND item_code = 'EXP.ADM.06')
    LOOP
        DELETE FROM corporate_checklist_items WHERE application_id = app_rec.id AND ged_document_id IS NULL;

        INSERT INTO corporate_checklist_items (application_id, item_code, title, section, requirement_level, is_mandatory, sort_order, status)
        VALUES
        (app_rec.id, 'EXP.ADM.01', 'Statuts notariés de la société enregistrés et certifiés conformes', 'ADMINISTRATIF', 'E', true, 10, 'manquant'),
        (app_rec.id, 'EXP.ADM.02', 'Registre du Commerce électronique CNRC d''exploitant titulaire de décisions d''enregistrement ANPP et/ou d''homologation DM', 'ADMINISTRATIF', 'E', true, 20, 'manquant'),
        (app_rec.id, 'EXP.ADM.03', 'Extrait de rôle fiscal apuré (validité < 3 mois, sans échéancier ni dette contestée)', 'ADMINISTRATIF', 'E', true, 30, 'manquant'),
        (app_rec.id, 'EXP.ADM.04', 'Attestations de mise à jour des cotisations sociales CNAS et CASNOS (< 3 mois)', 'ADMINISTRATIF', 'E', true, 40, 'manquant'),
        (app_rec.id, 'EXP.ADM.05', 'Cahier des conditions techniques à l''exploitation paraphé avec engagement déontologique', 'ADMINISTRATIF', 'E', true, 50, 'manquant'),
        (app_rec.id, 'EXP.ADM.06', 'Titre de propriété notarié ou contrat de bail notarié enregistré du siège et des installations de l''exploitant', 'ADMINISTRATIF', 'E', true, 60, 'manquant'),
        (app_rec.id, 'EXP.PDT.01', 'Diplôme d''État de Docteur en Pharmacie et attestation d''inscription CNOP du Pharmacien DT', 'PHARMACIEN_DT', 'E', true, 70, 'manquant'),
        (app_rec.id, 'EXP.PDT.02', 'Contrat de travail exclusif à temps plein et décision d''exercice nominative du Pharmacien DT', 'PHARMACIEN_DT', 'E', true, 80, 'manquant'),
        (app_rec.id, 'EXP.PV.01', 'Désignation officielle de la Personne Responsable de la Pharmacovigilance (QPPV) et système d''alerte 24/7', 'ORGANISATION_BPF', 'E', true, 90, 'manquant'),
        (app_rec.id, 'EXP.BPF.01', 'Procédure de contrôle et validation de l''information médicale/scientifique par le Pharmacien DT', 'ORGANISATION_BPF', 'E', true, 100, 'manquant'),
        (app_rec.id, 'EXP.TRC.01', 'Système de suivi de la traçabilité des lots sur le marché national et procédure de rappel de lots', 'ORGANISATION_BPF', 'E', true, 110, 'manquant')
        ON CONFLICT (application_id, item_code) DO UPDATE SET title = EXCLUDED.title, sort_order = EXCLUDED.sort_order;

        UPDATE corporate_applications SET completeness_percent = (
            SELECT COALESCE(ROUND(COUNT(ged_document_id)::numeric / NULLIF(COUNT(*), 0)::numeric * 100, 2), 0)
            FROM corporate_checklist_items WHERE application_id = app_rec.id
        ) WHERE id = app_rec.id;
    END LOOP;

    -- 5. Exportation (12 points unitaires)
    FOR app_rec IN
        SELECT id FROM corporate_applications
        WHERE category = 'EXPORTATION'
          AND NOT EXISTS (SELECT 1 FROM corporate_checklist_items WHERE application_id = corporate_applications.id AND item_code = 'EXT.ADM.06')
    LOOP
        DELETE FROM corporate_checklist_items WHERE application_id = app_rec.id AND ged_document_id IS NULL;

        INSERT INTO corporate_checklist_items (application_id, item_code, title, section, requirement_level, is_mandatory, sort_order, status)
        VALUES
        (app_rec.id, 'EXT.ADM.01', 'Statuts notariés de la société enregistrés et certifiés conformes', 'ADMINISTRATIF', 'E', true, 10, 'manquant'),
        (app_rec.id, 'EXT.ADM.02', 'Registre du Commerce électronique CNRC (codes d''activité 702103 médicaments / 702106 DM)', 'ADMINISTRATIF', 'E', true, 20, 'manquant'),
        (app_rec.id, 'EXT.ADM.03', 'Extrait de rôle fiscal apuré (validité < 3 mois, sans échéancier ni dette contestée)', 'ADMINISTRATIF', 'E', true, 30, 'manquant'),
        (app_rec.id, 'EXT.ADM.04', 'Attestations de mise à jour des cotisations sociales CNAS et CASNOS (< 3 mois)', 'ADMINISTRATIF', 'E', true, 40, 'manquant'),
        (app_rec.id, 'EXT.ADM.05', 'Cahier des charges à l''exportation paraphé et engagement de conformité réglementaire (Décret 21-82 Art. 12)', 'ADMINISTRATIF', 'E', true, 50, 'manquant'),
        (app_rec.id, 'EXT.ADM.06', 'Titre de propriété notarié ou contrat de bail notarié enregistré des locaux et aires d''entreposage sous douane/export', 'ADMINISTRATIF', 'E', true, 60, 'manquant'),
        (app_rec.id, 'EXT.PDT.01', 'Diplôme d''État de Docteur en Pharmacie du Pharmacien Directeur Technique (certifié conforme)', 'PHARMACIEN_DT', 'E', true, 70, 'manquant'),
        (app_rec.id, 'EXT.PDT.02', 'Attestation d''inscription au tableau de l''Ordre des Pharmaciens (CNOP) du Pharmacien DT', 'PHARMACIEN_DT', 'E', true, 80, 'manquant'),
        (app_rec.id, 'EXT.PDT.03', 'Contrat de travail à durée indéterminée exclusif à temps plein du Pharmacien DT', 'PHARMACIEN_DT', 'E', true, 90, 'manquant'),
        (app_rec.id, 'EXT.TEC.01', 'Descriptif et plans des magasins de stockage sous température contrôlée destinés à l''export', 'CONDITIONS_TECHNIQUES', 'E', true, 100, 'manquant'),
        (app_rec.id, 'EXT.TEC.02', 'Conventions de sourcing et contrats qualité avec les fabricants locaux ou importateurs agréés', 'CONDITIONS_TECHNIQUES', 'E', true, 110, 'manquant'),
        (app_rec.id, 'EXT.TRC.01', 'Système informatisé de traçabilité des lots exportés et procédure de délivrance des certificats d''exportation', 'ORGANISATION_BPF', 'E', true, 120, 'manquant')
        ON CONFLICT (application_id, item_code) DO UPDATE SET title = EXCLUDED.title, sort_order = EXCLUDED.sort_order;

        UPDATE corporate_applications SET completeness_percent = (
            SELECT COALESCE(ROUND(COUNT(ged_document_id)::numeric / NULLIF(COUNT(*), 0)::numeric * 100, 2), 0)
            FROM corporate_checklist_items WHERE application_id = app_rec.id
        ) WHERE id = app_rec.id;
    END LOOP;
END $$;
