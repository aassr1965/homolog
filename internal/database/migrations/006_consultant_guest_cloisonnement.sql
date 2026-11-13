-- Migration 006 : Cloisonnement Hermétique Guest pour Consultants Externes (CRO / Experts Tiers)
-- Conformité Section 7.2.2 & 7.2.3 du Cahier des Charges (PSSI & BPF)

-- 1. Table d'assignation contractuelle des dossiers aux consultants
CREATE TABLE IF NOT EXISTS dossier_consultants (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dossier_id UUID NOT NULL REFERENCES dossiers(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    assigned_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    expires_at TIMESTAMPTZ,
    notes TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(dossier_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_dossier_consultants_user ON dossier_consultants(user_id);
CREATE INDEX IF NOT EXISTS idx_dossier_consultants_dossier ON dossier_consultants(dossier_id);

-- 2. Durcissement des permissions pour le rôle consultant_externe :
-- Retrait formel de REF_VIEW et DOC_READ transverse.
-- Le consultant n'a pas accès au référentiel d'entreprise ni à la GED transverse (branches 01 à 05).
-- Il n'a accès qu'à son dossier assigné (DOSSIER_VIEW) et peut éditer/verrouiller les pièces assignées (DOC_WRITE, DOC_LOCK_CHECKOUT).
DELETE FROM role_permissions 
WHERE role_id = 'consultant_externe' 
  AND permission_id IN ('REF_VIEW', 'DOC_READ');

-- 3. Harmonisation du rôle et assignation contractuelle de test pour consultant.ext@cro.com
UPDATE users SET role = 'consultant_externe' WHERE email = 'consultant.ext@cro.com';

INSERT INTO dossier_consultants (dossier_id, user_id, expires_at, notes)
SELECT d.id, u.id, CURRENT_TIMESTAMP + INTERVAL '90 days', 'Mission CRO : Étude de Bioéquivalence & Validation Module 5'
FROM dossiers d
CROSS JOIN users u
WHERE u.email = 'consultant.ext@cro.com'
ORDER BY d.created_at ASC
LIMIT 1
ON CONFLICT (dossier_id, user_id) DO UPDATE SET
    expires_at = EXCLUDED.expires_at,
    notes = EXCLUDED.notes;

-- 4. Invalidation instantanée de session pour les consultants
UPDATE users SET permissions_version = COALESCE(permissions_version, 1) + 1 WHERE role = 'consultant_externe';
