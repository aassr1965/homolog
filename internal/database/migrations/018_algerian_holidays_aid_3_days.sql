-- Migration 018 : Extension à 3 jours chômés payés pour l'Aïd El-Fitr et l'Aïd El-Adha
-- Conformément à la Loi n° 23-07 du 3 mai 2023 modifiant la Loi n° 63-278 relative aux fêtes légales

INSERT INTO algerian_holidays (holiday_date, name_fr, holiday_type) VALUES
    -- 2026 : 3ème jour Aïd El-Fitr et Aïd El-Adha
    ('2026-03-22', 'Aïd El-Fitr (Jour 3)', 'RELIGIEUX'),
    ('2026-05-29', 'Aïd El-Adha (Jour 3)', 'RELIGIEUX'),

    -- 2027 : 3ème jour Aïd El-Fitr et Aïd El-Adha
    ('2027-03-12', 'Aïd El-Fitr (Jour 3)', 'RELIGIEUX'),
    ('2027-05-19', 'Aïd El-Adha (Jour 3)', 'RELIGIEUX'),

    -- 2028 : 3ème jour Aïd El-Fitr et Aïd El-Adha
    ('2028-02-29', 'Aïd El-Fitr (Jour 3)', 'RELIGIEUX'),
    ('2028-05-07', 'Aïd El-Adha (Jour 3)', 'RELIGIEUX')
ON CONFLICT (holiday_date) DO NOTHING;
