-- =====================================================================================
-- Migration 035 : Restriction de AUDIT_VIEW du Pharmacien Assistant à l'intérim DT
-- Décret exécutif n° 21-82 (Art. 24) & EU GMP Annexe 11 § 12 (Moindre privilège)
-- =====================================================================================

DELETE FROM role_permissions
WHERE role_id = 'pharmacien_assistant'
  AND permission_id = 'AUDIT_VIEW';
