-- ==============================================================================
-- Migration 011: Barème Réglementaire Versionné & Bordereaux de Versement BV1/BV2 ANPP
-- Conforme aux décrets d'homologation et arrêtés interministériels des redevances
-- ==============================================================================

-- 1. Table des Barèmes Officiels des Redevances ANPP (Historique & Prise d'Effet)
CREATE TABLE IF NOT EXISTS anpp_fee_schedules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    category VARCHAR(64) NOT NULL, -- 'medicament', 'dispositif_medical', 'variation', 'autre'
    procedure_code VARCHAR(64) NOT NULL, -- 'princeps', 'generique_local', 'generique_importe', 'renouvellement', 'variation_majeure', 'variation_mineure', 'dm_classe_I', 'dm_classe_IIa', 'dm_classe_IIb_III', 'dm_div'
    label_fr VARCHAR(255) NOT NULL,
    label_ar VARCHAR(255),
    legal_reference VARCHAR(255) NOT NULL, -- ex: Arrêté Interministériel JORADP N°44
    amount_dzd NUMERIC(14, 2) NOT NULL,
    valid_from DATE NOT NULL DEFAULT CURRENT_DATE,
    valid_to DATE, -- NULL si tarif en vigueur actuellement
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Index pour recherche rapide du barème applicable à une date donnée
CREATE INDEX IF NOT EXISTS idx_fee_schedules_proc_date ON anpp_fee_schedules(procedure_code, valid_from, valid_to);
CREATE INDEX IF NOT EXISTS idx_fee_schedules_cat ON anpp_fee_schedules(category);

-- 2. Injection des barèmes officiels de référence en Algérie (Idempotent)
INSERT INTO anpp_fee_schedules (category, procedure_code, label_fr, label_ar, legal_reference, amount_dzd, valid_from, is_active)
SELECT 'medicament', 'princeps', 'Enregistrement Médicament Princeps / Nouvelle Molécule', 'تسجيل دواء أصلي / جزيء جديد', 'Arrêté Interministériel fixant les redevances dues à l''ANPP (JORADP)', 350000.00, '2020-01-01', TRUE
WHERE NOT EXISTS (SELECT 1 FROM anpp_fee_schedules WHERE procedure_code = 'princeps' AND valid_to IS NULL);

INSERT INTO anpp_fee_schedules (category, procedure_code, label_fr, label_ar, legal_reference, amount_dzd, valid_from, is_active)
SELECT 'medicament', 'generique_local', 'Enregistrement Médicament Générique (Fabrication Locale)', 'تسجيل دواء جنيس (صناعة محلية)', 'Décret Exécutif 20-325 & Arrêté des Redevances ANPP', 150000.00, '2020-01-01', TRUE
WHERE NOT EXISTS (SELECT 1 FROM anpp_fee_schedules WHERE procedure_code = 'generique_local' AND valid_to IS NULL);

INSERT INTO anpp_fee_schedules (category, procedure_code, label_fr, label_ar, legal_reference, amount_dzd, valid_from, is_active)
SELECT 'medicament', 'generique_importe', 'Enregistrement Médicament Générique (Importation Conditionnée)', 'تسجيل دواء جنيس (مستورد)', 'Décret Exécutif 20-325 & Arrêté des Redevances ANPP', 250000.00, '2020-01-01', TRUE
WHERE NOT EXISTS (SELECT 1 FROM anpp_fee_schedules WHERE procedure_code = 'generique_importe' AND valid_to IS NULL);

INSERT INTO anpp_fee_schedules (category, procedure_code, label_fr, label_ar, legal_reference, amount_dzd, valid_from, is_active)
SELECT 'medicament', 'renouvellement', 'Renouvellement Quinquennal de la Décision d''Enregistrement', 'تجديد خماسي لمقرر التسجيل', 'Arrêté Interministériel fixant les redevances dues à l''ANPP', 100000.00, '2020-01-01', TRUE
WHERE NOT EXISTS (SELECT 1 FROM anpp_fee_schedules WHERE procedure_code = 'renouvellement' AND valid_to IS NULL);

INSERT INTO anpp_fee_schedules (category, procedure_code, label_fr, label_ar, legal_reference, amount_dzd, valid_from, is_active)
SELECT 'variation', 'variation_majeure', 'Modification Majeure Post-Enregistrement (Variation Mmaj)', 'تعديل جوهري بعد التسجيل', 'Directive ANPP LD-DEPP07/V01 & Arrêté Redevances', 100000.00, '2020-01-01', TRUE
WHERE NOT EXISTS (SELECT 1 FROM anpp_fee_schedules WHERE procedure_code = 'variation_majeure' AND valid_to IS NULL);

INSERT INTO anpp_fee_schedules (category, procedure_code, label_fr, label_ar, legal_reference, amount_dzd, valid_from, is_active)
SELECT 'variation', 'variation_mineure', 'Modification Mineure / Modérée (Variation Mmin / Mmod)', 'تعديل طفيف أو متوسط بعد التسجيل', 'Directive ANPP LD-DEPP07/V01 & Arrêté Redevances', 50000.00, '2020-01-01', TRUE
WHERE NOT EXISTS (SELECT 1 FROM anpp_fee_schedules WHERE procedure_code = 'variation_mineure' AND valid_to IS NULL);

INSERT INTO anpp_fee_schedules (category, procedure_code, label_fr, label_ar, legal_reference, amount_dzd, valid_from, is_active)
SELECT 'dispositif_medical', 'dm_classe_I', 'Homologation Dispositif Médical - Classe I (Non Invasif)', 'مصادقة مستلزم طبي - الصنف الأول', 'Décret Exécutif 20-324 & Arrêté du 10 Mai 2021', 50000.00, '2021-05-10', TRUE
WHERE NOT EXISTS (SELECT 1 FROM anpp_fee_schedules WHERE procedure_code = 'dm_classe_I' AND valid_to IS NULL);

INSERT INTO anpp_fee_schedules (category, procedure_code, label_fr, label_ar, legal_reference, amount_dzd, valid_from, is_active)
SELECT 'dispositif_medical', 'dm_classe_IIa', 'Homologation Dispositif Médical - Classe IIa (Usage Court)', 'مصادقة مستلزم طبي - الصنف الثاني أ', 'Décret Exécutif 20-324 & Arrêté du 10 Mai 2021', 100000.00, '2021-05-10', TRUE
WHERE NOT EXISTS (SELECT 1 FROM anpp_fee_schedules WHERE procedure_code = 'dm_classe_IIa' AND valid_to IS NULL);

INSERT INTO anpp_fee_schedules (category, procedure_code, label_fr, label_ar, legal_reference, amount_dzd, valid_from, is_active)
SELECT 'dispositif_medical', 'dm_classe_IIb_III', 'Homologation Dispositif Médical - Classe IIb & III (Implantable)', 'مصادقة مستلزم طبي - الصنف الثاني ب والثالث (قابل للزرع)', 'Décret Exécutif 20-324 & Arrêté du 10 Mai 2021', 150000.00, '2021-05-10', TRUE
WHERE NOT EXISTS (SELECT 1 FROM anpp_fee_schedules WHERE procedure_code = 'dm_classe_IIb_III' AND valid_to IS NULL);

INSERT INTO anpp_fee_schedules (category, procedure_code, label_fr, label_ar, legal_reference, amount_dzd, valid_from, is_active)
SELECT 'dispositif_medical', 'dm_div', 'Homologation Réactifs et Dispositifs de Diagnostic In Vitro (DMDIV)', 'مصادقة مستلزمات التشخيص المخبري في المختبر', 'Décret Exécutif 20-324 & Arrêté du 10 Mai 2021', 80000.00, '2021-05-10', TRUE
WHERE NOT EXISTS (SELECT 1 FROM anpp_fee_schedules WHERE procedure_code = 'dm_div' AND valid_to IS NULL);

-- 3. Table des Bordereaux de Versement BV1 / BV2 Générés & Scellés
CREATE TABLE IF NOT EXISTS anpp_payment_slips (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    dossier_id UUID NOT NULL REFERENCES dossiers(id) ON DELETE CASCADE,
    fee_schedule_id UUID NOT NULL REFERENCES anpp_fee_schedules(id) ON DELETE RESTRICT,
    slip_type VARCHAR(16) NOT NULL, -- 'BV1' (Dépôt initial), 'BV2' (Complément / Expertise)
    slip_number VARCHAR(64) NOT NULL, -- ex: BV1-2026-AMOX-001
    payment_purpose VARCHAR(255) NOT NULL,
    amount_dzd NUMERIC(14, 2) NOT NULL,
    amount_words_fr VARCHAR(255) NOT NULL,
    amount_words_ar VARCHAR(255) NOT NULL,
    bank_account_rib VARCHAR(64) NOT NULL DEFAULT '00100624030000001234 (Compte Trésorerie ANPP Alger)',
    status VARCHAR(32) NOT NULL DEFAULT 'emis', -- 'emis', 'paye', 'quittance_recue', 'annule'
    payment_reference VARCHAR(120), -- N° quittance trésor ou avis d'opéré bancaire
    paid_at DATE,
    receipt_document_id UUID REFERENCES ged_documents(id) ON DELETE SET NULL, -- Scan quittance dans GED
    generated_by UUID REFERENCES users(id) ON DELETE SET NULL,
    generated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_slip_number UNIQUE (organization_id, slip_number)
);

-- Index pour requêtes bordereaux
CREATE INDEX IF NOT EXISTS idx_payment_slips_dossier ON anpp_payment_slips(dossier_id);
CREATE INDEX IF NOT EXISTS idx_payment_slips_org ON anpp_payment_slips(organization_id);
CREATE INDEX IF NOT EXISTS idx_payment_slips_status ON anpp_payment_slips(status);
