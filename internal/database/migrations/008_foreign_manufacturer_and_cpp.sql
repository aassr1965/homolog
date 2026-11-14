-- ============================================================================
-- Migration 008 : Traçabilité des Producteurs Étrangers et Certificats CPP (ANPP 1.H.1 & 1.H.8)
-- Conforme au Décret exécutif n° 20-325 et au Guide d'enregistrement CTD Juillet 2025
-- ============================================================================

ALTER TABLE medicines
    ADD COLUMN IF NOT EXISTS foreign_mah_name VARCHAR(255), -- Laboratoire Titulaire / Demandeur Étranger (MAH)
    ADD COLUMN IF NOT EXISTS foreign_mah_country VARCHAR(2), -- Pays du titulaire étranger (ISO-3166-1 alpha-2)
    ADD COLUMN IF NOT EXISTS cpp_number VARCHAR(100), -- N° Certificat CPP Modèle OMS (Item 1.H.1)
    ADD COLUMN IF NOT EXISTS cpp_issuing_authority VARCHAR(255), -- Autorité sanitaire étrangère émettrice (ANSM, EMA, FDA, etc.)
    ADD COLUMN IF NOT EXISTS cpp_issue_date DATE, -- Date d'octroi du CPP
    ADD COLUMN IF NOT EXISTS manufacturer_site_id UUID REFERENCES manufacturing_sites(id) ON DELETE SET NULL; -- Site Producteur / Fabricant PF (1.M.P / 1.H.8)

CREATE INDEX IF NOT EXISTS idx_medicines_manufacturer_site ON medicines(manufacturer_site_id);

-- Mise à jour des spécialités importées du seed avec leurs producteurs étrangers de référence
UPDATE medicines 
SET 
    foreign_mah_name = 'Sanofi Winthrop Industrie',
    foreign_mah_country = 'FR',
    cpp_number = 'CPP-FR-ANSM-2024-8841',
    cpp_issuing_authority = 'Agence Nationale de Sécurité du Médicament (ANSM) - France',
    cpp_issue_date = '2024-02-15',
    manufacturer_site_id = (SELECT id FROM manufacturing_sites WHERE country_code = 'FR' LIMIT 1)
WHERE origin_type = 'princeps' AND foreign_mah_name IS NULL;

UPDATE medicines 
SET 
    foreign_mah_name = 'Cipla Global Pharmaceuticals',
    foreign_mah_country = 'IN',
    cpp_number = 'CPP-IN-CDSCO-2023-4102',
    cpp_issuing_authority = 'Central Drugs Standard Control Organization (CDSCO) / WHO',
    cpp_issue_date = '2023-08-20',
    manufacturer_site_id = (SELECT id FROM manufacturing_sites WHERE country_code = 'IN' LIMIT 1)
WHERE origin_type = 'generique_importe' AND foreign_mah_name IS NULL;
