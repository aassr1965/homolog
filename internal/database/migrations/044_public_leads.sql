-- Migration 044 : Table des demandes publiques (Démo, Contact & Pré-Onboarding) depuis le portail SANAD Pharma
CREATE TABLE IF NOT EXISTS public_leads (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    lead_type VARCHAR(32) NOT NULL CHECK (lead_type IN ('demo', 'contact', 'onboarding')),
    full_name VARCHAR(200) NOT NULL,
    email VARCHAR(255) NOT NULL,
    organization_name VARCHAR(255) NOT NULL DEFAULT '',
    role_title VARCHAR(150) NOT NULL DEFAULT '',
    phone VARCHAR(60) NOT NULL DEFAULT '',
    deployment_interest VARCHAR(32) NOT NULL DEFAULT 'saas',
    language VARCHAR(8) NOT NULL DEFAULT 'fr',
    message TEXT NOT NULL DEFAULT '',
    ip_address VARCHAR(64) NOT NULL DEFAULT '',
    user_agent TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_public_leads_created_at ON public_leads (created_at DESC);
CREATE INDEX IF NOT EXISTS idx_public_leads_type ON public_leads (lead_type);
