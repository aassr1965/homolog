-- =====================================================================================
-- Migration 037 : Libération de Lots Importés DM & PP (DCE / Arrêtés 14/02/2021 & 20/02/2025)
-- =====================================================================================

CREATE TABLE IF NOT EXISTS import_lot_releases (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    product_kind VARCHAR(8) NOT NULL CHECK (product_kind IN ('dm', 'pp')),
    medical_device_id UUID REFERENCES medical_devices(id) ON DELETE SET NULL,
    medicine_id UUID REFERENCES medicines(id) ON DELETE SET NULL,
    commercial_name VARCHAR(255) NOT NULL,
    designation TEXT NOT NULL DEFAULT '',
    form_aspect TEXT NOT NULL DEFAULT '',
    lot_number VARCHAR(120) NOT NULL,
    serial_number VARCHAR(120) NOT NULL DEFAULT '',
    ddf DATE,
    ddp DATE,
    sales_units VARCHAR(120) NOT NULL DEFAULT '',
    homologation_or_de_ref VARCHAR(150) NOT NULL DEFAULT '',
    dm_class VARCHAR(16) NOT NULL DEFAULT '',
    holder_or_exploiter VARCHAR(255) NOT NULL DEFAULT '',
    import_program_attached BOOLEAN NOT NULL DEFAULT FALSE,
    production_site VARCHAR(255) NOT NULL DEFAULT '',
    qc_site VARCHAR(255) NOT NULL DEFAULT '',
    release_site VARCHAR(255) NOT NULL DEFAULT '',
    customs_decl_number VARCHAR(150) NOT NULL DEFAULT '',
    invoice_number VARCHAR(150) NOT NULL DEFAULT '',
    origin_cert_attached BOOLEAN NOT NULL DEFAULT FALSE,
    tech_form_attached BOOLEAN NOT NULL DEFAULT FALSE,
    sampling_plan_attached BOOLEAN NOT NULL DEFAULT FALSE,
    status VARCHAR(32) NOT NULL DEFAULT 'brouillon',
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    signed_by_pdt_id UUID REFERENCES users(id) ON DELETE SET NULL,
    signed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_import_lot_releases_org ON import_lot_releases(organization_id, product_kind, created_at DESC);
