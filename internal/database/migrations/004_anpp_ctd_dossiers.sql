-- ============================================================================
-- Migration 004 : Moteur de Dossiers ANPP & Checklists Dynamiques (Guide CTD Juillet 2025)
-- Conforme Décrets Exécutifs 20-325, 20-324, Arrêté du 10 mai 2021 & Lignes Directrices
-- Typologie officielle : NI, RI, MI, NF, RF, MF et Homologation DM
-- ============================================================================

-- 1. Enrichissement de la table dossiers pour les spécifications ANPP 2025
ALTER TABLE dossiers
    ADD COLUMN IF NOT EXISTS medicine_id UUID REFERENCES medicines(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS medical_device_id UUID REFERENCES medical_devices(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS product_type VARCHAR(32) DEFAULT 'medicament' NOT NULL,
    ADD COLUMN IF NOT EXISTS dossier_code VARCHAR(100),
    ADD COLUMN IF NOT EXISTS procedure_code VARCHAR(16) DEFAULT 'NI' NOT NULL,
    ADD COLUMN IF NOT EXISTS target_deposit_date DATE,
    ADD COLUMN IF NOT EXISTS tracking_number VARCHAR(100),
    ADD COLUMN IF NOT EXISTS deposit_receipt_ref VARCHAR(100),
    ADD COLUMN IF NOT EXISTS deposit_receipt_date DATE,
    ADD COLUMN IF NOT EXISTS submission_notes TEXT,
    ADD COLUMN IF NOT EXISTS pdt_signed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS pdt_signed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS pdt_signature_seal VARCHAR(64),
    ADD COLUMN IF NOT EXISTS pdt_signature_role VARCHAR(100);

-- Mettre à jour product_id comme optionnel si medicine_id ou medical_device_id est renseigné
ALTER TABLE dossiers ALTER COLUMN product_id DROP NOT NULL;
ALTER TABLE dossiers ALTER COLUMN procedure_type DROP NOT NULL;
ALTER TABLE dossiers ALTER COLUMN model_template_id DROP NOT NULL;
ALTER TABLE dossiers ALTER COLUMN model_version DROP NOT NULL;

-- 2. Table Relais Relationnelle pour les Éléments de Checklist CTD & DM
CREATE TABLE IF NOT EXISTS dossier_checklist_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dossier_id UUID NOT NULL REFERENCES dossiers(id) ON DELETE CASCADE,
    module_number INT NOT NULL, -- 1 à 5 pour CTD Médicament, 0 pour Dispositif Médical
    sub_module VARCHAR(32) NOT NULL, -- ex: '1.A', '1.H', '1.M', '1.P', '3.2.S', '3.2.P', '3.2.R', '5.3.1.2', 'DM_TECH'
    item_code VARCHAR(32) NOT NULL, -- ex: '1.A.1', '1.H.4', '1.P.2.2', '1.P.8', '1.P.9', '3.2.P.3.5', '5.3.1.2'
    title_fr VARCHAR(255) NOT NULL,
    title_ar VARCHAR(255),
    description_fr TEXT,
    requirement_level VARCHAR(10) DEFAULT 'E' NOT NULL, -- 'E' (Exigé), 'F' (Facultatif), 'P' (Partiellement requis), 'NA' (Non applicable)
    is_mandatory BOOLEAN DEFAULT TRUE NOT NULL,
    ged_document_id UUID REFERENCES ged_documents(id) ON DELETE SET NULL,
    status VARCHAR(32) DEFAULT 'manquant' NOT NULL, -- 'manquant', 'rattache', 'valide_qualite', 'approuve_dt', 'rejete'
    validation_notes TEXT,
    sort_order INT DEFAULT 0 NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT uq_dossier_item_code UNIQUE (dossier_id, item_code)
);

-- 3. Index optimisés pour les requêtes sur les dossiers et checklists
CREATE INDEX IF NOT EXISTS idx_dossiers_med ON dossiers(medicine_id);
CREATE INDEX IF NOT EXISTS idx_dossiers_dm ON dossiers(medical_device_id);
CREATE INDEX IF NOT EXISTS idx_dossiers_proc_code ON dossiers(organization_id, procedure_code);
CREATE INDEX IF NOT EXISTS idx_dossiers_code ON dossiers(organization_id, dossier_code);
CREATE INDEX IF NOT EXISTS idx_dossier_items_dossier ON dossier_checklist_items(dossier_id);
CREATE INDEX IF NOT EXISTS idx_dossier_items_ged ON dossier_checklist_items(ged_document_id);
CREATE INDEX IF NOT EXISTS idx_dossier_items_status ON dossier_checklist_items(dossier_id, status);
CREATE INDEX IF NOT EXISTS idx_dossier_items_module ON dossier_checklist_items(dossier_id, module_number, sub_module);
