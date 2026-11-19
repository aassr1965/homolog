-- Migration 017 : Table des jours fériés légaux et chômés en Algérie (Loi n° 63-278 modifiée)
-- Utilisée par le Smart Planning ANPP pour exclure les jours fériés du calcul des délais incompressibles

CREATE TABLE IF NOT EXISTS algerian_holidays (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    holiday_date DATE NOT NULL UNIQUE,
    name_fr VARCHAR(150) NOT NULL,
    holiday_type VARCHAR(50) NOT NULL DEFAULT 'NATIONAL', -- 'NATIONAL', 'RELIGIEUX'
    is_observed BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_algerian_holidays_date ON algerian_holidays(holiday_date);

-- Insertion des jours fériés légaux fixes et religieux chômés payés (2026, 2027, 2028)
INSERT INTO algerian_holidays (holiday_date, name_fr, holiday_type) VALUES
    -- 2026 : Fêtes nationales fixes
    ('2026-01-01', 'Jour de l''An', 'NATIONAL'),
    ('2026-01-12', 'Yennayer (Nouvel An Amazigh)', 'NATIONAL'),
    ('2026-05-01', 'Fête du Travail', 'NATIONAL'),
    ('2026-07-05', 'Fête de l''Indépendance et de la Jeunesse', 'NATIONAL'),
    ('2026-11-01', 'Fête de la Révolution (1er Novembre)', 'NATIONAL'),
    -- 2026 : Fêtes religieuses chômées payées (Loi 23-07 : 3 jours pour les 2 Aïds)
    ('2026-03-20', 'Aïd El-Fitr (Jour 1)', 'RELIGIEUX'),
    ('2026-03-21', 'Aïd El-Fitr (Jour 2)', 'RELIGIEUX'),
    ('2026-03-22', 'Aïd El-Fitr (Jour 3)', 'RELIGIEUX'),
    ('2026-05-27', 'Aïd El-Adha (Jour 1)', 'RELIGIEUX'),
    ('2026-05-28', 'Aïd El-Adha (Jour 2)', 'RELIGIEUX'),
    ('2026-05-29', 'Aïd El-Adha (Jour 3)', 'RELIGIEUX'),
    ('2026-06-16', '1er Moharram (Nouvel An de l''Hégire)', 'RELIGIEUX'),
    ('2026-06-25', 'Achoura (10 Moharram)', 'RELIGIEUX'),
    ('2026-08-25', 'El Mawlid Ennabawi Echarif', 'RELIGIEUX'),

    -- 2027 : Fêtes nationales fixes
    ('2027-01-01', 'Jour de l''An', 'NATIONAL'),
    ('2027-01-12', 'Yennayer (Nouvel An Amazigh)', 'NATIONAL'),
    ('2027-05-01', 'Fête du Travail', 'NATIONAL'),
    ('2027-07-05', 'Fête de l''Indépendance et de la Jeunesse', 'NATIONAL'),
    ('2027-11-01', 'Fête de la Révolution (1er Novembre)', 'NATIONAL'),
    -- 2027 : Fêtes religieuses chômées payées (Loi 23-07 : 3 jours pour les 2 Aïds)
    ('2027-03-10', 'Aïd El-Fitr (Jour 1)', 'RELIGIEUX'),
    ('2027-03-11', 'Aïd El-Fitr (Jour 2)', 'RELIGIEUX'),
    ('2027-03-12', 'Aïd El-Fitr (Jour 3)', 'RELIGIEUX'),
    ('2027-05-17', 'Aïd El-Adha (Jour 1)', 'RELIGIEUX'),
    ('2027-05-18', 'Aïd El-Adha (Jour 2)', 'RELIGIEUX'),
    ('2027-05-19', 'Aïd El-Adha (Jour 3)', 'RELIGIEUX'),
    ('2027-06-06', '1er Moharram (Nouvel An de l''Hégire)', 'RELIGIEUX'),
    ('2027-06-15', 'Achoura (10 Moharram)', 'RELIGIEUX'),
    ('2027-08-15', 'El Mawlid Ennabawi Echarif', 'RELIGIEUX'),

    -- 2028 : Fêtes nationales fixes
    ('2028-01-01', 'Jour de l''An', 'NATIONAL'),
    ('2028-01-12', 'Yennayer (Nouvel An Amazigh)', 'NATIONAL'),
    ('2028-05-01', 'Fête du Travail', 'NATIONAL'),
    ('2028-07-05', 'Fête de l''Indépendance et de la Jeunesse', 'NATIONAL'),
    ('2028-11-01', 'Fête de la Révolution (1er Novembre)', 'NATIONAL'),
    -- 2028 : Fêtes religieuses chômées payées (Loi 23-07 : 3 jours pour les 2 Aïds)
    ('2028-02-27', 'Aïd El-Fitr (Jour 1)', 'RELIGIEUX'),
    ('2028-02-28', 'Aïd El-Fitr (Jour 2)', 'RELIGIEUX'),
    ('2028-02-29', 'Aïd El-Fitr (Jour 3)', 'RELIGIEUX'),
    ('2028-05-05', 'Aïd El-Adha (Jour 1)', 'RELIGIEUX'),
    ('2028-05-06', 'Aïd El-Adha (Jour 2)', 'RELIGIEUX'),
    ('2028-05-07', 'Aïd El-Adha (Jour 3)', 'RELIGIEUX'),
    ('2028-05-25', '1er Moharram (Nouvel An de l''Hégire)', 'RELIGIEUX'),
    ('2028-06-03', 'Achoura (10 Moharram)', 'RELIGIEUX'),
    ('2028-08-03', 'El Mawlid Ennabawi Echarif', 'RELIGIEUX')
ON CONFLICT (holiday_date) DO NOTHING;
