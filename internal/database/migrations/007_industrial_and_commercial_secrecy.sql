-- ============================================================================
-- Migration 007 : Cloisonnement Hermétique Renforcé, Secret Commercial & Industriel
-- Conformité Section 7.2.2 & 7.2.3 du Cahier des Charges (PSSI, BPF, 21 CFR Part 11)
-- Principe du "Besoin d'en connaître" (Need-to-Know) & Protection Propriété Intellectuelle
-- ============================================================================

-- 1. Table d'affectation universelle aux dossiers (Membres & Habilitations de Sécurité)
CREATE TABLE IF NOT EXISTS dossier_assignments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dossier_id UUID NOT NULL REFERENCES dossiers(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_in_dossier VARCHAR(64) NOT NULL DEFAULT 'contributeur', -- 'lead_ra', 'contributeur_qa', 'contributeur_rd', 'consultant_cro', 'auditeur'
    can_view_financials BOOLEAN NOT NULL DEFAULT FALSE,          -- Droit exceptionnel de lever le Secret Commercial (1.A.4, 1.H.4, 1.H.5)
    can_view_industrial_secret BOOLEAN NOT NULL DEFAULT TRUE,   -- Droit de consulter les formules complètes et procédés de synthèse (3.2.S, 3.2.P)
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(dossier_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_dossier_assignments_user ON dossier_assignments(user_id);
CREATE INDEX IF NOT EXISTS idx_dossier_assignments_dossier ON dossier_assignments(dossier_id);

-- 2. Migration des affectations consultants existantes vers dossier_assignments
INSERT INTO dossier_assignments (dossier_id, user_id, role_in_dossier, can_view_financials, can_view_industrial_secret, assigned_at, expires_at, notes, created_at)
SELECT dossier_id, user_id, 'consultant_cro', FALSE, FALSE, assigned_at, expires_at, notes, created_at
FROM dossier_consultants
ON CONFLICT (dossier_id, user_id) DO UPDATE SET
    can_view_financials = EXCLUDED.can_view_financials,
    can_view_industrial_secret = EXCLUDED.can_view_industrial_secret,
    expires_at = EXCLUDED.expires_at,
    notes = EXCLUDED.notes;

-- 3. Ajout de la classification de sécurité sur les éléments de checklist
ALTER TABLE dossier_checklist_items 
ADD COLUMN IF NOT EXISTS security_classification VARCHAR(32) NOT NULL DEFAULT 'STANDARD';

CREATE INDEX IF NOT EXISTS idx_dossier_checklist_sec_class ON dossier_checklist_items(security_classification);

-- 4. Application des classifications réglementaires strictes
-- A. SECRET COMMERCIAL : Prix PCSU/FOB, benchmark international, quittances bancaires et change
UPDATE dossier_checklist_items
SET security_classification = 'SECRET_COMMERCIAL'
WHERE item_code IN ('1.A.4', '1.H.4', '1.H.5');

-- B. SECRET INDUSTRIEL : Synthèse substance active (ASMF/DMF) et formule de fabrication complète (Procédés 3 lots)
UPDATE dossier_checklist_items
SET security_classification = 'SECRET_INDUSTRIEL'
WHERE item_code LIKE '3.2.S%' OR item_code LIKE '3.2.P%' OR sub_module IN ('3.2.S', '3.2.P');

-- 5. Affectation initiale des contributeurs qualité et R&D sur le dossier démo Sanad Pharma
-- (Conforme au principe du moindre privilège : seuls les collaborateurs affectés voient le dossier)
INSERT INTO dossier_assignments (dossier_id, user_id, role_in_dossier, can_view_financials, can_view_industrial_secret, notes)
SELECT d.id, u.id, 'contributeur_qa', FALSE, TRUE, 'Habilitation Qualité BPF - Modules 1 & 3'
FROM dossiers d
CROSS JOIN users u
WHERE u.email = 'qa.lead@pharma.dz'
ON CONFLICT (dossier_id, user_id) DO NOTHING;

INSERT INTO dossier_assignments (dossier_id, user_id, role_in_dossier, can_view_financials, can_view_industrial_secret, notes)
SELECT d.id, u.id, 'contributeur_rd_production', FALSE, TRUE, 'Habilitation R&D & Validation Procédé 3 lots'
FROM dossiers d
CROSS JOIN users u
WHERE u.email = 'prod.lead@pharma.dz'
ON CONFLICT (dossier_id, user_id) DO NOTHING;

INSERT INTO dossier_assignments (dossier_id, user_id, role_in_dossier, can_view_financials, can_view_industrial_secret, notes)
SELECT d.id, u.id, 'contributeur_rd_production', FALSE, TRUE, 'Habilitation Contrôle Qualité Laboratoire'
FROM dossiers d
CROSS JOIN users u
WHERE u.email = 'qc.lead@pharma.dz'
ON CONFLICT (dossier_id, user_id) DO NOTHING;

-- Invalidation du cache des permissions
UPDATE users SET permissions_version = COALESCE(permissions_version, 1) + 1;
