-- ============================================================================
-- Migration 027 : Homologation Approfondie & Régularisation Complète des Dispositifs Médicaux (DDM)
-- Conforme au Décret exécutif n° 20-324 du 22 novembre 2020 et à l'Arrêté du 10 mai 2021
-- Traçabilité BPF / 21 CFR Part 11 / Mesures Transitoires
-- ============================================================================

-- 1. Enrichissement de la table medical_devices pour les 6 Sections Techniques de l'Homologation
ALTER TABLE medical_devices
    ADD COLUMN IF NOT EXISTS composition_details TEXT,
    ADD COLUMN IF NOT EXISTS has_latex BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS has_phthalates BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS has_active_pharma BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS has_biological_origin BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS has_cmr_substances BOOLEAN DEFAULT FALSE,
    ADD COLUMN IF NOT EXISTS risk_management_iso14971_ref VARCHAR(255) DEFAULT 'PGR-ISO14971-2025-V01',
    ADD COLUMN IF NOT EXISTS risk_management_status VARCHAR(64) DEFAULT 'conforme',
    ADD COLUMN IF NOT EXISTS biocompatibility_iso10993_ref VARCHAR(255) DEFAULT 'BIO-ISO10993-GLP-2025-08',
    ADD COLUMN IF NOT EXISTS biocompatibility_status VARCHAR(64) DEFAULT 'conforme',
    ADD COLUMN IF NOT EXISTS clinical_eval_iso14155_ref VARCHAR(255) DEFAULT 'CER-ISO14155-2025-CLIN',
    ADD COLUMN IF NOT EXISTS clinical_eval_status VARCHAR(64) DEFAULT 'conforme',
    ADD COLUMN IF NOT EXISTS sterilization_validation_ref VARCHAR(255) DEFAULT 'VAL-STERIL-ISO11135-2024',
    ADD COLUMN IF NOT EXISTS packaging_stability_iso11607_ref VARCHAR(255) DEFAULT 'STAB-ISO11607-36M-2025',
    ADD COLUMN IF NOT EXISTS sample_batch_pf VARCHAR(100) DEFAULT 'LOT-PF-2026-DM01',
    ADD COLUMN IF NOT EXISTS sample_batch_mp VARCHAR(100) DEFAULT 'LOT-MP-PUR-098',
    ADD COLUMN IF NOT EXISTS sample_units_count INT DEFAULT 3;

-- 2. Table d'Évaluation de la Recevabilité en 13 Points de Contrôle (ES-REG-03-DM)
CREATE TABLE IF NOT EXISTS dm_recevabilite_evaluations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    dossier_id UUID NOT NULL REFERENCES dossiers(id) ON DELETE CASCADE,
    medical_device_id UUID NOT NULL REFERENCES medical_devices(id) ON DELETE CASCADE,
    deposit_order_number VARCHAR(100) NOT NULL, -- ex: ANPP/DDM/REC/2026/0148
    deposit_date DATE NOT NULL DEFAULT CURRENT_DATE,
    electronic_copies_count INT DEFAULT 4,
    electronic_media_type VARCHAR(64) DEFAULT 'usb_sealed_sha256',
    
    -- Les 13 points de contrôle de l'Arrêté du 10 mai 2021
    item_1_status VARCHAR(32) DEFAULT 'conforme' NOT NULL, -- Lettre de demande 2 ex. + descriptif
    item_2_status VARCHAR(32) DEFAULT 'conforme' NOT NULL, -- Formulaire 6 sections (ES-REG-02-DM)
    item_3_status VARCHAR(32) DEFAULT 'conforme' NOT NULL, -- Agrément MIPH fabricant/exploitant
    item_4_status VARCHAR(32) DEFAULT 'conforme' NOT NULL, -- Justification redevance BNA (300k DA)
    item_5_status VARCHAR(32) DEFAULT 'conforme' NOT NULL, -- CLV / FSC pays d'origine
    item_6_status VARCHAR(32) DEFAULT 'conforme' NOT NULL, -- Certificat CE & Déclaration
    item_7_status VARCHAR(32) DEFAULT 'conforme' NOT NULL, -- BPF / ISO 13485 PF & LCQ
    item_8_status VARCHAR(32) DEFAULT 'conforme' NOT NULL, -- Certifications matières premières
    item_9_status VARCHAR(32) DEFAULT 'conforme' NOT NULL, -- Matériovigilance & PGR
    item_10_status VARCHAR(32) DEFAULT 'conforme' NOT NULL, -- Partie III Stabilité & Stérilisation
    item_11_status VARCHAR(32) DEFAULT 'conforme' NOT NULL, -- Partie IV ISO 14971 & ISO 10993
    item_12_status VARCHAR(32) DEFAULT 'conforme' NOT NULL, -- Partie V Évaluation clinique ISO 14155
    item_13_status VARCHAR(32) DEFAULT 'conforme' NOT NULL, -- 04 Exemplaires électroniques scellés
    
    conformity_score INT DEFAULT 13 NOT NULL,
    decision VARCHAR(64) DEFAULT 'complet_et_recevable' NOT NULL, -- 'complet_et_recevable', 'incomplet_avec_reserves', 'irrecevable'
    reserves_notes TEXT,
    anpp_agent_name VARCHAR(255) DEFAULT 'Dr. Amel Bouzid (Sous-Direction de l''Homologation DDM)',
    pdt_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    pdt_signed_at TIMESTAMPTZ,
    pdt_signature_seal VARCHAR(64),
    sha256_hash VARCHAR(64),
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT uq_dm_eval_dossier UNIQUE (dossier_id)
);

CREATE INDEX IF NOT EXISTS idx_dm_eval_org ON dm_recevabilite_evaluations(organization_id);
CREATE INDEX IF NOT EXISTS idx_dm_eval_dossier ON dm_recevabilite_evaluations(dossier_id);
CREATE INDEX IF NOT EXISTS idx_dm_eval_device ON dm_recevabilite_evaluations(medical_device_id);

-- 3. Table de Suivi des Mesures Transitoires & Régularisation des DM non homologués
CREATE TABLE IF NOT EXISTS dm_regularizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    medical_device_id UUID NOT NULL REFERENCES medical_devices(id) ON DELETE CASCADE,
    regularization_code VARCHAR(100) NOT NULL, -- ex: REG-DM-2026-001
    derogation_number VARCHAR(100), -- ex: DEROG-MIPH-2024-8841
    initial_market_date DATE NOT NULL DEFAULT (CURRENT_DATE - INTERVAL '18 months'),
    status VARCHAR(64) DEFAULT 'commercialise_derogatoire' NOT NULL, 
    -- 'commercialise_derogatoire', 'dossier_depose_ddm', 'en_instruction', 'homologue_regularise', 'retire'
    ddm_acknowledgement_number VARCHAR(100), -- N° Accusé de dépôt DDM
    ddm_deposit_date DATE,
    legal_deadline DATE NOT NULL DEFAULT (CURRENT_DATE + INTERVAL '90 days'), -- Échéance légale
    quarterly_volume_imported NUMERIC(15, 2) DEFAULT 25000.00 NOT NULL,
    quarterly_volume_distributed NUMERIC(15, 2) DEFAULT 22400.00 NOT NULL,
    hospital_distribution_share INT DEFAULT 70 NOT NULL, -- % Hôpital
    retail_distribution_share INT DEFAULT 30 NOT NULL,   -- % Officine
    last_declaration_date DATE DEFAULT CURRENT_DATE NOT NULL,
    observations TEXT,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT uq_dm_reg_org_code UNIQUE (organization_id, regularization_code)
);

CREATE INDEX IF NOT EXISTS idx_dm_reg_org ON dm_regularizations(organization_id);
CREATE INDEX IF NOT EXISTS idx_dm_reg_device ON dm_regularizations(medical_device_id);
CREATE INDEX IF NOT EXISTS idx_dm_reg_status ON dm_regularizations(status);
CREATE INDEX IF NOT EXISTS idx_dm_reg_deadline ON dm_regularizations(legal_deadline);

-- 4. Initialisation des données pour les dispositifs médicaux existants
UPDATE medical_devices
SET
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
    sample_units_count = 3
WHERE commercial_name_fr ILIKE '%Cathéter%' OR manufacturer_reference ILIKE '%IVC%';
