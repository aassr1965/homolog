-- ============================================================================
-- Migration 038 : Stockage des PDF complets des 19 Guidelines ANPP en base
-- (regulatory_watch_items.document_pdf BYTEA + SHA-256 + indexation corpus)
-- ============================================================================

ALTER TABLE regulatory_watch_items
    ADD COLUMN IF NOT EXISTS document_pdf BYTEA,
    ADD COLUMN IF NOT EXISTS document_filename VARCHAR(255),
    ADD COLUMN IF NOT EXISTS document_sha256 VARCHAR(64),
    ADD COLUMN IF NOT EXISTS document_size_bytes BIGINT NOT NULL DEFAULT 0;

CREATE TABLE IF NOT EXISTS regulatory_watch_documents (
    document_filename VARCHAR(255) PRIMARY KEY,
    document_pdf BYTEA NOT NULL,
    document_sha256 VARCHAR(64) NOT NULL,
    document_size_bytes BIGINT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Associer le fichier PDF officiel à l'entrée LD-DEPP07/V01 existante si présente
UPDATE regulatory_watch_items
SET document_filename = 'ld_depp07_modifications.pdf'
WHERE legal_reference = 'LD-DEPP07/V01'
  AND (document_filename IS NULL OR document_filename = '');

-- Insérer les 18 autres Guidelines officielles ANPP pour toutes les organisations
INSERT INTO regulatory_watch_items (
    organization_id, title_fr, legal_reference, joradp_number,
    published_date, effective_date, category, impact_level,
    summary_fr, impact_analysis, compliance_status, source_url, document_filename
)
SELECT
    o.id,
    v.title_fr,
    v.legal_reference,
    'Guideline Officielle ANPP',
    v.pub_date::date,
    v.pub_date::date,
    v.category,
    v.impact_level,
    v.summary_fr,
    v.impact_analysis,
    'en_vigueur',
    v.source_url,
    v.document_filename
FROM organizations o
CROSS JOIN (VALUES
    (
        'Guide Algérien de Bonnes Pratiques de Pharmacovigilance (v2)',
        'ANPP-BPPV-V2',
        '2024-06-01',
        'pharmacovigilance',
        'critique',
        'Référentiel national de pharmacovigilance : organisation QPPV/RLPV, PSMF, recueil et transmission des cas d''effets indésirables au CNPM/ANPP.',
        'Couvert par GED 04.01, EXP.PV.01, engagement 1.P.9 et fiche DVTIV PDT/QPPV.',
        'https://anpp.dz/download/304/guidelines/21774/guide-algerien-de-bonnes-pratiques-de-pharmacovigilance-v2.pdf',
        'guide_bppv_v2.pdf'
    ),
    (
        'Guide des Bonnes Pratiques de Distribution et de Stockage des PP et DM (BPDG)',
        'ANPP-BPDG',
        '2024-06-01',
        'bpf_inspection',
        'majeur',
        'Exigences de stockage sous température contrôlée, traçabilité des lots, gestion des retours et rappels en distribution en gros.',
        'Couvert par la checklist eForm-MIPH-DISTRIBUTION.',
        'https://anpp.dz/download/304/guidelines/21837/guide-des-bonnes-pratiques-de-distribution-et-de-stockage-des-produits-pharmaceutiques-et-des-dispositifs-medicaux-a-usage-humain.pdf',
        'guide_bpdg.pdf'
    ),
    (
        'Guide relatif au rappel des lots et au retrait des PP et DM non conformes (v2)',
        'ANPP-RAPPEL-LOTS-V2',
        '2024-06-01',
        'pharmacovigilance',
        'critique',
        'Procédure d''alerte, de rappel de lots sous 24h et de retrait/destruction des produits pharmaceutiques et DM non conformes.',
        'Workflow dédié Rappel de lots 24h planifié (Phase 3).',
        'https://anpp.dz/download/304/guidelines/21772/guide-relatif-au-rappel-des-lots-et-au-retrait-des-produits-pharmaceutiques-et-des-dispositifs-medicaux-non-conformes.pdf',
        'guide_rappel_lots_v2.pdf'
    ),
    (
        'Guide relatif au retrait de lot et au rappel des PP et DM non conformes (v1)',
        'ANPP-RETRAIT-LOT-V1',
        '2023-06-01',
        'pharmacovigilance',
        'majeur',
        'Version initiale des modalités de retrait de lot et de notification aux autorités sanitaires.',
        'Archivé en référence historique avec la v2.',
        'https://anpp.dz/download/304/guidelines/21767/guide-relatif-au-retrait-de-lot-et-au-rappel-des-pp-et-dm-non-conforme.pdf',
        'guide_retrait_lot_v1.pdf'
    ),
    (
        'Ligne Directrice Inspections Réglementaires (LD-DVIV04/V01)',
        'LD-DVIV04/V01',
        '2024-09-01',
        'bpf_inspection',
        'critique',
        'Cadre officiel de conduite des inspections réglementaires BPF/BPDG/PV par les inspecteurs de l''ANPP (DVTIV).',
        'Workflow Inspections & CAPA planifié (Phase 4).',
        'https://anpp.dz/download/304/guidelines/22241/ld-dviv04v01-ligne-directrice-inspections-reglementaires.pdf',
        'ld_dviv04_inspections.pdf'
    ),
    (
        'Ligne Directrice Classification des observations et écarts d''inspections BPF',
        'ANPP-LD-ECARTS-BPF',
        '2024-06-01',
        'bpf_inspection',
        'critique',
        'Grille de classification des écarts BPF (Critique, Majeur, Autre) en fonction du risque qualité produit et patient.',
        'Cotation des écarts BPF planifiée (Phase 4).',
        'https://anpp.dz/download/304/guidelines/21775/ligne-directrice-relative-a-la-classification-des-observations-ecarts-liees-aux-inspections-des-bonnes-pratiques-de-fabrication-bpf-des-medicaments-a-usage-de-la-medecine-humaine-en-fonction-du-r.pdf',
        'ld_classification_ecarts_bpf.pdf'
    ),
    (
        'Ligne Directrice Préparation et soumission des demandes d''enregistrement, modification et renouvellement PP',
        'ANPP-LD-SOUMISSION-PP',
        '2024-06-01',
        'enregistrement_amm',
        'critique',
        'Instructions détaillées de préparation et soumission des dossiers CTD M1-M5 et des formulaires DEPP02.',
        'Couvert par le moteur CTD M1-M5 et les 9 formulaires DEPP .docx.',
        'https://anpp.dz/download/304/guidelines/21798/ligne-directrice-pour-la-preparation-et-la-soumission-des-demandes-denregistrement-de-modification-et-de-renouvellement-des-produits-pharmaceutiques-a-usage-de-la-medecine-humaine-en-algerie.pdf',
        'ld_soumission_pp.pdf'
    ),
    (
        'Ligne Directrice Prérequis du renouvellement de la décision d''enregistrement',
        'ANPP-LD-RENOUVELLEMENT-DE',
        '2024-06-01',
        'enregistrement_amm',
        'majeur',
        'Conditions et pièces requises pour le renouvellement quinquennal de la DE et la déclaration 1.H.10.',
        'Couvert par les procédures RI/RF et l''état officiel ES-ENG-1H10.',
        'https://anpp.dz/download/304/guidelines/21796/ligne-directrice-relative-aux-prerequis-du-renouvellement-de-la-decision-denregistrement.pdf',
        'ld_renouvellement_de.pdf'
    ),
    (
        'Ligne Directrice Exigences des études de bioéquivalence et d''équivalence thérapeutique',
        'ANPP-LD-BIOEQUIVALENCE',
        '2024-06-01',
        'essais_cliniques_be',
        'majeur',
        'Exigences cliniques, bioanalytiques BPL et statistiques applicables aux dossiers génériques soumis à bioéquivalence.',
        'Couvert par le module CTD M5 et le tracker BE Study (EI-CHR-02).',
        'https://anpp.dz/download/304/guidelines/21793/ligne-directrice-relative-aux-exigences-des-etudes-de-bioequivalence-et-dequivalence-therapeutique.pdf',
        'ld_bioequivalence.pdf'
    ),
    (
        'Règles de rédaction de l''information produit (RCP, notice et conditionnement)',
        'ANPP-LD-RCP-NOTICE',
        '2024-06-01',
        'enregistrement_amm',
        'majeur',
        'Normes rédactionnelles bilingues FR/AR pour le Résumé des Caractéristiques du Produit, la notice et les maquettes.',
        'Couvert par le dossier GED 01.05 Notices & Articles de Conditionnement.',
        'https://anpp.dz/download/304/guidelines/21797/regles-de-redaction-de-linformation-produit-rcp-notice-et-conditionnement-des-medicaments-destines-a-lusage-de-la-medecine-humaine.pdf',
        'ld_rcp_notice_cond.pdf'
    ),
    (
        'Ligne Directrice Refus, retrait temporaire et retrait définitif de la décision d''enregistrement',
        'ANPP-LD-REFUS-RETRAIT-DE',
        '2024-06-01',
        'enregistrement_amm',
        'critique',
        'Motifs et conséquences réglementaires d''un refus, d''une suspension ou d''un retrait définitif de DE.',
        'Gating contournable par le PDT sur les lots DCE planifié (Phase 5).',
        'https://anpp.dz/download/304/guidelines/21839/ligne-directrice-relative-au-refus-retrait-temporaire-et-retrait-definitif-de-la-decision-denregistrement.pdf',
        'ld_refus_retrait_de.pdf'
    ),
    (
        'Ligne Directrice Définition des rôles et responsabilités dans la procédure d''enregistrement ANPP',
        'ANPP-LD-ROLES-ENREG',
        '2024-06-01',
        'enregistrement_amm',
        'mineur',
        'Répartition des responsabilités entre demandeur, PDT, rapporteurs et commissions d''enregistrement.',
        'Couvert par le RBAC Maker-Checker et le Parapheur DT.',
        'https://anpp.dz/download/304/guidelines/22012/ligne-directrice-relative-a-la-definition-des-roles-et-des-responsabilites-des-entites-impliquees-dans-la-procedure-denregistrement-des-produits-pharmaceutiques-au-sein-de-lanpp.pdf',
        'ld_roles_enregistrement.pdf'
    ),
    (
        'Ligne Directrice Confiance Réglementaire (Reliance) — Enregistrement et Modification DE',
        'ANPP-LD-RELIANCE-ENREG',
        '2024-09-01',
        'enregistrement_amm',
        'majeur',
        'Recours à la confiance réglementaire et reconnaissance des autorités strictes (EMA, FDA, OMS, etc.) pour l''AMM.',
        'Parcours Reliance sur dossiers CTD planifié (Phase 2).',
        'https://anpp.dz/download/304/guidelines/22208/ligne-directrice-relative-a-la-confiance-reglementaire-et-a-la-reconnaissance-dans-le-cadre-de-lenregistrement-et-de-la-modification-des-decision-denregistrement.pdf',
        'ld_reliance_enreg.pdf'
    ),
    (
        'Ligne Directrice Confiance Réglementaire (Reliance) — Tests de laboratoire',
        'ANPP-LD-RELIANCE-LABO',
        '2024-09-01',
        'enregistrement_amm',
        'majeur',
        'Modalités de recours à la reconnaissance réglementaire pour les essais et contrôles de laboratoire.',
        'Parcours Reliance Labo sur libération de lots planifié (Phase 2).',
        'https://anpp.dz/download/304/guidelines/22174/ligne-directrice-fixant-les-modalites-de-recours-a-la-confiance-reglementaire-et-a-la-reconnaissance-pour-les-tests-de-laboratoire-des-produits-pharmaceutiques-a-usage-de-la-medecine-humaine.pdf',
        'ld_reliance_labo.pdf'
    ),
    (
        'Ligne Directrice Confiance Réglementaire (Reliance) — Inspections BPF',
        'ANPP-LD-RELIANCE-BPF',
        '2024-09-01',
        'bpf_inspection',
        'majeur',
        'Pratiques de reliance et reconnaissance des rapports d''inspection BPF d''autorités réglementaires strictes.',
        'Parcours Reliance BPF planifié (Phase 2).',
        'https://anpp.dz/download/304/guidelines/22227/ligne-directrice-des-pratiques-de-la-confiance-reglementaire-reliance-et-de-reconnaisance-en-matiere-dinspection-des-bonnes-pratiques-de-fabrication.pdf',
        'ld_reliance_bpf.pdf'
    ),
    (
        'Ligne Directrice Confiance Réglementaire (Reliance) — Pharmacovigilance',
        'ANPP-LD-RELIANCE-PV',
        '2024-09-01',
        'pharmacovigilance',
        'majeur',
        'Pratiques de confiance réglementaire et partage d''informations de sécurité en pharmacovigilance.',
        'Référentiel PV intégré dans la veille réglementaire.',
        'https://anpp.dz/download/304/guidelines/22207/ligne-directrice-relative-aux-pratiques-de-confiance-reglementaire-en-matiere-de-pharmacovigilance.pdf',
        'ld_reliance_pv.pdf'
    ),
    (
        'Ligne Directrice relative à la communication (ANPP)',
        'ANPP-LD-COMMUNICATION',
        '2024-06-01',
        'enregistrement_amm',
        'mineur',
        'Principes et canaux officiels de communication réglementaire avec l''ANPP.',
        'Référentiel documentaire intégré en base.',
        'https://anpp.dz/download/304/guidelines/21838/ligne-directrice-relative-a-la-communication.pdf',
        'ld_communication.pdf'
    ),
    (
        'Ligne Directrice relative au recrutement et à la formation (ANPP)',
        'ANPP-LD-FORMATION',
        '2024-06-01',
        'bpf_inspection',
        'mineur',
        'Exigences de qualification, recrutement et formation continue des personnels et évaluateurs.',
        'Référentiel documentaire intégré en base.',
        'https://anpp.dz/download/304/guidelines/21792/ligne-directrice-relative-au-recrutement-et-a-la-formation.pdf',
        'ld_recrutement_formation.pdf'
    )
) AS v(title_fr, legal_reference, pub_date, category, impact_level, summary_fr, impact_analysis, source_url, document_filename)
WHERE NOT EXISTS (
    SELECT 1 FROM regulatory_watch_items r
    WHERE r.organization_id = o.id
      AND (r.document_filename = v.document_filename OR r.legal_reference = v.legal_reference)
);
