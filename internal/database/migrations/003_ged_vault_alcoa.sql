-- ============================================================================
-- Migration 003 : Coffre-Fort GED Pharmaceutique (Standard BPF & ALCOA+)
-- Extension des tables ged_documents et ged_versions
-- ============================================================================

-- 1. Enrichissement de la table ged_documents pour le contrôle d'accès BPF et Check-out/Check-in
ALTER TABLE ged_documents
    ADD COLUMN IF NOT EXISTS document_type VARCHAR(64) DEFAULT 'autre' NOT NULL,
    ADD COLUMN IF NOT EXISTS department VARCHAR(64) DEFAULT 'affaires_reglementaires',
    ADD COLUMN IF NOT EXISTS tags TEXT[] DEFAULT '{}',
    ADD COLUMN IF NOT EXISTS is_locked BOOLEAN DEFAULT FALSE NOT NULL,
    ADD COLUMN IF NOT EXISTS locked_by UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS locked_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS lock_reason VARCHAR(255),
    ADD COLUMN IF NOT EXISTS establishment_id UUID REFERENCES establishments(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS site_id UUID REFERENCES manufacturing_sites(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS medicine_id UUID REFERENCES medicines(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS medical_device_id UUID REFERENCES medical_devices(id) ON DELETE SET NULL;

-- 2. Enrichissement de la table ged_versions pour l'intégrité ALCOA+ et scellement PDT
ALTER TABLE ged_versions
    ADD COLUMN IF NOT EXISTS is_major_version BOOLEAN DEFAULT FALSE NOT NULL,
    ADD COLUMN IF NOT EXISTS signature_sha256 VARCHAR(64),
    ADD COLUMN IF NOT EXISTS signed_ip VARCHAR(64);

-- 3. Index optimisés pour les requêtes GED
CREATE INDEX IF NOT EXISTS idx_ged_docs_org_status ON ged_documents(organization_id, status);
CREATE INDEX IF NOT EXISTS idx_ged_docs_org_code ON ged_documents(organization_id, document_code);
CREATE INDEX IF NOT EXISTS idx_ged_docs_locked ON ged_documents(organization_id, is_locked) WHERE is_locked = TRUE;
CREATE INDEX IF NOT EXISTS idx_ged_docs_establishment ON ged_documents(establishment_id);
CREATE INDEX IF NOT EXISTS idx_ged_docs_medicine ON ged_documents(medicine_id);
CREATE INDEX IF NOT EXISTS idx_ged_docs_medical_device ON ged_documents(medical_device_id);
CREATE INDEX IF NOT EXISTS idx_ged_versions_doc_created ON ged_versions(document_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_ged_versions_file_sha ON ged_versions(file_sha256);
