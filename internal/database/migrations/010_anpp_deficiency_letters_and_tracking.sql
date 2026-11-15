-- ==============================================================================
-- Migration 010: Studio de Compléments ANPP & Traçabilité des Réponses Réglementaires
-- Conforme au Guide d'Homologation ANPP 2025 et aux exigences BPF / 21 CFR Part 11
-- ==============================================================================

-- 1. Table des Lettres Officielles de Demande de Compléments ANPP
CREATE TABLE IF NOT EXISTS anpp_deficiency_letters (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    dossier_id UUID NOT NULL REFERENCES dossiers(id) ON DELETE CASCADE,
    reference_number VARCHAR(120) NOT NULL, -- Ex: ANPP/DEPP/N°1428/2026
    received_date DATE NOT NULL,
    legal_deadline_days INT NOT NULL DEFAULT 90, -- Échéance légale : 30, 60, 90 ou 180 jours
    due_date DATE NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'open', -- 'open', 'partially_answered', 'submitted', 'overdue'
    document_id UUID REFERENCES ged_documents(id) ON DELETE SET NULL, -- Scan officiel archivé en GED
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- 2. Évolution de la table anpp_deficiency_tasks pour rattachement et granularité
ALTER TABLE anpp_deficiency_tasks ADD COLUMN IF NOT EXISTS letter_id UUID REFERENCES anpp_deficiency_letters(id) ON DELETE CASCADE;
ALTER TABLE anpp_deficiency_tasks ADD COLUMN IF NOT EXISTS organization_id UUID REFERENCES organizations(id) ON DELETE CASCADE;
ALTER TABLE anpp_deficiency_tasks ADD COLUMN IF NOT EXISTS module_target VARCHAR(64) DEFAULT 'General';
ALTER TABLE anpp_deficiency_tasks ADD COLUMN IF NOT EXISTS priority VARCHAR(32) DEFAULT 'normale';

-- Rétro-remplissage organization_id si des tâches existaient sans org_id
UPDATE anpp_deficiency_tasks t 
SET organization_id = d.organization_id 
FROM dossiers d 
WHERE t.dossier_id = d.id AND t.organization_id IS NULL;

-- 3. Index d'optimisation pour requêtes haute performance
CREATE INDEX IF NOT EXISTS idx_deficiency_letters_dossier ON anpp_deficiency_letters(dossier_id);
CREATE INDEX IF NOT EXISTS idx_deficiency_letters_org ON anpp_deficiency_letters(organization_id);
CREATE INDEX IF NOT EXISTS idx_deficiency_letters_due_date ON anpp_deficiency_letters(due_date);
CREATE INDEX IF NOT EXISTS idx_deficiency_letters_status ON anpp_deficiency_letters(status);

CREATE INDEX IF NOT EXISTS idx_deficiency_tasks_letter ON anpp_deficiency_tasks(letter_id);
CREATE INDEX IF NOT EXISTS idx_deficiency_tasks_dossier ON anpp_deficiency_tasks(dossier_id);
CREATE INDEX IF NOT EXISTS idx_deficiency_tasks_org ON anpp_deficiency_tasks(organization_id);
CREATE INDEX IF NOT EXISTS idx_deficiency_tasks_due ON anpp_deficiency_tasks(due_date);
CREATE INDEX IF NOT EXISTS idx_deficiency_tasks_status ON anpp_deficiency_tasks(status);
