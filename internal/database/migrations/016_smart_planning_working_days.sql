-- Migration 016 : Paramètre des Jours Ouvrables pour le Smart Planning & Rétroplanning Inversé
-- Conformité Section 6.4.1 & 7.4.3.6 du Cahier des Charges

ALTER TABLE organizations
    ADD COLUMN IF NOT EXISTS working_days_per_week INT DEFAULT 5;

ALTER TABLE dossiers
    ADD COLUMN IF NOT EXISTS working_days_per_week INT DEFAULT 5;
