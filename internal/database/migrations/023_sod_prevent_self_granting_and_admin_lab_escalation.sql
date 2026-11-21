-- Migration 023 : Ségrégation des Tâches (SoD) & Garde-Fou Anti-Élévation de Privilèges
-- Décret exécutif 20-325 & 21 CFR Part 11 :
-- 1. Interdiction absolue d'auto-attribution de permissions (No Self-Granting).
-- 2. Incompatibilité stricte de fonctions pour l'Administrateur Laboratoire (rejet de DOSSIER_*, DOC_* et SYS_MAINT).

CREATE OR REPLACE FUNCTION check_user_permission_sod_guard()
RETURNS TRIGGER AS $$
DECLARE
    target_role VARCHAR;
    prohibited_perms TEXT[] := ARRAY['DOSSIER_VIEW', 'DOSSIER_CREATE', 'DOC_READ', 'DOC_DRAFT_MAKER', 'DOC_CHECK_REVIEW', 'DOC_APPROVE_PDT', 'SYS_MAINT'];
BEGIN
    -- 1. Anti-Auto-Élévation : assigned_by ne peut pas être égal à user_id
    IF NEW.assigned_by IS NOT NULL AND NEW.assigned_by = NEW.user_id THEN
        RAISE EXCEPTION 'Garde-fou BPF / 21 CFR Part 11 : Auto-attribution de permissions strictement interdite (Self-Granting).';
    END IF;

    -- 2. Incompatibilité SoD pour admin_laboratoire
    SELECT role INTO target_role FROM users WHERE id = NEW.user_id;
    IF target_role = 'admin_laboratoire' AND NEW.granted = TRUE AND NEW.permission_id = ANY(prohibited_perms) THEN
        RAISE EXCEPTION 'Garde-fou SoD BPF : La permission % est strictement incompatible avec le mandat de l''Administrateur Laboratoire.', NEW.permission_id;
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_check_user_permission_sod ON user_permissions;
CREATE TRIGGER trg_check_user_permission_sod
BEFORE INSERT OR UPDATE ON user_permissions
FOR EACH ROW
EXECUTE FUNCTION check_user_permission_sod_guard();
