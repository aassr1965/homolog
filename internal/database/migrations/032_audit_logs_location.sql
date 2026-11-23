-- =========================================================================
-- Migration 032: Traçabilité géographique et physique de l'Audit Trail
-- Ajout de la colonne 'location' pour identifier le lieu de connexion ALCOA+
-- =========================================================================

ALTER TABLE audit_logs ADD COLUMN IF NOT EXISTS location VARCHAR(255) DEFAULT '';
