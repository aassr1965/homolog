-- ============================================================================
-- Migration 001 : Schéma Initial Homolog - Réglementaire ANPP & GED Pharmaceutique
-- Conforme BPF, GxP, ALCOA+ et 21 CFR Part 11
-- ============================================================================

-- Extensions nécessaires
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Table des Organisations (Laboratoires / Tenants)
CREATE TABLE IF NOT EXISTS organizations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug VARCHAR(64) UNIQUE NOT NULL,
    legal_name VARCHAR(255) NOT NULL,
    miph_license_number VARCHAR(100),
    technical_director_name VARCHAR(255),
    address TEXT,
    phone VARCHAR(50),
    email VARCHAR(100),
    storage_quota_bytes BIGINT DEFAULT 53687091200, -- 50 Go par défaut
    storage_used_bytes BIGINT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- Table des Utilisateurs nominatifs
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100) NOT NULL,
    last_name VARCHAR(100) NOT NULL,
    role VARCHAR(64) NOT NULL,
    totp_secret VARCHAR(128),
    totp_enabled BOOLEAN DEFAULT FALSE,
    recovery_codes TEXT[],
    failed_login_attempts INT DEFAULT 0,
    locked_until TIMESTAMPTZ,
    contract_expires_at TIMESTAMPTZ, -- Obligatoire pour consultants externes
    is_active BOOLEAN DEFAULT TRUE,
    last_login_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT uq_org_email UNIQUE (organization_id, email)
);

-- Table des Sessions sécurisées (HttpOnly Cookie avec révocation instantanée)
CREATE TABLE IF NOT EXISTS sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,
    ip_address VARCHAR(45),
    user_agent TEXT,
    expires_at TIMESTAMPTZ NOT NULL,
    last_activity_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- Table du Journal d'Audit Immuable (Audit Trail Append-Only)
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    user_email VARCHAR(255),
    user_role VARCHAR(64),
    action VARCHAR(100) NOT NULL,
    entity_type VARCHAR(64) NOT NULL,
    entity_id VARCHAR(128) NOT NULL,
    details JSONB DEFAULT '{}'::jsonb NOT NULL,
    ip_address VARCHAR(45),
    user_agent TEXT,
    sha256_hash VARCHAR(64) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- Trigger pour empêcher strictement toute modification (UPDATE) ou suppression (DELETE) sur l'Audit Trail
CREATE OR REPLACE FUNCTION prevent_audit_log_tampering()
RETURNS TRIGGER AS $$
BEGIN
    RAISE EXCEPTION 'ALERTE CONFORMITÉ BPF : Le Journal d Audit est strictement immuable (interdiction formelle de UPDATE ou DELETE)';
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_audit_logs_immutable ON audit_logs;
CREATE TRIGGER trg_audit_logs_immutable
BEFORE UPDATE OR DELETE ON audit_logs
FOR EACH ROW EXECUTE FUNCTION prevent_audit_log_tampering();

-- Table des Produits (Médicaments et Dispositifs Médicaux)
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    product_type VARCHAR(32) NOT NULL, -- 'medicament' ou 'dispositif_medical'
    brand_name_fr VARCHAR(255) NOT NULL,
    brand_name_ar VARCHAR(255),
    dci VARCHAR(255),
    dosage VARCHAR(100),
    pharmaceutical_form VARCHAR(100),
    administration_route VARCHAR(100),
    dm_class VARCHAR(20), -- 'I', 'IIa', 'IIb', 'III', 'DMDIV'
    is_sterile BOOLEAN DEFAULT FALSE,
    dm_classification_rule VARCHAR(50),
    manufacturer_name VARCHAR(255) NOT NULL,
    marketing_status VARCHAR(50) DEFAULT 'en_preparation',
    is_hors_nomenclature BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- Table des Documents GED Pharmaceutique
CREATE TABLE IF NOT EXISTS ged_documents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    taxonomy_path VARCHAR(255) NOT NULL, -- ex: '01_Affaires_Reglementaires/01_04_Dossiers_CTD_Reference'
    title_fr VARCHAR(255) NOT NULL,
    title_ar VARCHAR(255),
    document_code VARCHAR(100),
    status VARCHAR(32) DEFAULT 'brouillon' NOT NULL,
    current_version VARCHAR(20) DEFAULT 'v0.1' NOT NULL,
    expiry_date DATE,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- Table des Versions de Documents GED (Inaltérables)
CREATE TABLE IF NOT EXISTS ged_versions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    document_id UUID NOT NULL REFERENCES ged_documents(id) ON DELETE CASCADE,
    version_label VARCHAR(20) NOT NULL, -- 'v0.1', 'v1.0'
    revision_summary TEXT NOT NULL,
    file_sha256 VARCHAR(64) NOT NULL,
    file_size_bytes BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    original_filename VARCHAR(255) NOT NULL,
    storage_path VARCHAR(512) NOT NULL,
    is_signed_dt BOOLEAN DEFAULT FALSE,
    signed_by_dt_id UUID REFERENCES users(id) ON DELETE SET NULL,
    signed_at TIMESTAMPTZ,
    signature_declaration TEXT,
    author_id UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- Table des Dossiers Réglementaires ANPP
CREATE TABLE IF NOT EXISTS dossiers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE CASCADE,
    procedure_type VARCHAR(64) NOT NULL,
    internal_reference VARCHAR(100) NOT NULL,
    anpp_official_reference VARCHAR(100),
    anpp_deposit_date DATE,
    anpp_decision_date DATE,
    anpp_validity_end_date DATE,
    status VARCHAR(32) DEFAULT 'brouillon' NOT NULL,
    responsible_id UUID REFERENCES users(id) ON DELETE SET NULL,
    model_template_id VARCHAR(100) NOT NULL,
    model_version VARCHAR(20) NOT NULL,
    checklist_data JSONB DEFAULT '[]'::jsonb NOT NULL,
    completeness_percent INT DEFAULT 0,
    is_locked_for_deposit BOOLEAN DEFAULT FALSE,
    locked_by_dt_id UUID REFERENCES users(id) ON DELETE SET NULL,
    locked_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT uq_org_internal_ref UNIQUE (organization_id, internal_reference)
);

-- Table des Tâches Unitaires de Compléments ANPP
CREATE TABLE IF NOT EXISTS anpp_deficiency_tasks (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dossier_id UUID NOT NULL REFERENCES dossiers(id) ON DELETE CASCADE,
    question_number INT NOT NULL,
    question_text TEXT NOT NULL,
    assigned_to UUID REFERENCES users(id) ON DELETE SET NULL,
    due_date DATE NOT NULL,
    status VARCHAR(32) DEFAULT 'a_traiter' NOT NULL,
    response_document_id UUID REFERENCES ged_documents(id) ON DELETE SET NULL,
    response_summary TEXT,
    validated_by_dt_id UUID REFERENCES users(id) ON DELETE SET NULL,
    validated_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- Index d'optimisation
CREATE INDEX IF NOT EXISTS idx_users_org ON users(organization_id);
CREATE INDEX IF NOT EXISTS idx_sessions_token ON sessions(token_hash);
CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_org_created ON audit_logs(organization_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_products_org ON products(organization_id);
CREATE INDEX IF NOT EXISTS idx_ged_docs_org_tax ON ged_documents(organization_id, taxonomy_path);
CREATE INDEX IF NOT EXISTS idx_ged_versions_doc ON ged_versions(document_id);
CREATE INDEX IF NOT EXISTS idx_dossiers_org_status ON dossiers(organization_id, status);
CREATE INDEX IF NOT EXISTS idx_dossiers_product ON dossiers(product_id);
CREATE INDEX IF NOT EXISTS idx_dossiers_checklist ON dossiers USING GIN (checklist_data);
