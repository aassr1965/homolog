-- ============================================================================
-- Migration 009 : Traçabilité Complète des Dispositifs Médicaux & Cycle de Vie
-- Conforme au Décret exécutif n° 20-324 du 22 novembre 2020 et à l'Arrêté du 10 mai 2021
-- Traçabilité 21 CFR Part 11 / BPF ALCOA+
-- ============================================================================

ALTER TABLE medical_devices
    ADD COLUMN IF NOT EXISTS origin_type VARCHAR(32) DEFAULT 'fabrication_locale' NOT NULL, -- 'fabrication_locale' ou 'importe'
    ADD COLUMN IF NOT EXISTS legal_manufacturer_name VARCHAR(255), -- Fabricant Légal étranger
    ADD COLUMN IF NOT EXISTS legal_manufacturer_country VARCHAR(2), -- Code pays ISO-3166-1 alpha-2 du fabricant légal
    ADD COLUMN IF NOT EXISTS manufacturer_site_id UUID REFERENCES manufacturing_sites(id) ON DELETE SET NULL, -- Site réel de fabrication / assemblage
    ADD COLUMN IF NOT EXISTS establishment_id UUID REFERENCES establishments(id) ON DELETE SET NULL, -- Établissement demandeur / mandataire local MIPH
    ADD COLUMN IF NOT EXISTS ce_certificate_number VARCHAR(100), -- N° Certificat CE (Règlement UE 2017/745 ou Directive 93/42/CEE)
    ADD COLUMN IF NOT EXISTS notified_body_name VARCHAR(255), -- Organisme Notifié (Notified Body)
    ADD COLUMN IF NOT EXISTS notified_body_number VARCHAR(32), -- Numéro Organisme Notifié (ex: 0459 GMED, 0123 TÜV SÜD, 0086 BSI)
    ADD COLUMN IF NOT EXISTS ce_expiry_date DATE, -- Date d'échéance du certificat CE
    ADD COLUMN IF NOT EXISTS iso_13485_certificate_number VARCHAR(100), -- Certificat Système Qualité Dispositifs Médicaux ISO 13485
    ADD COLUMN IF NOT EXISTS clv_fsc_number VARCHAR(100), -- N° Certificat de Libre Vente (CLV / FSC) pays d'origine
    ADD COLUMN IF NOT EXISTS clv_issuing_authority VARCHAR(255), -- Autorité sanitaire émettrice du CLV (ANSM, FDA, BfArM, etc.)
    ADD COLUMN IF NOT EXISTS clv_issue_date DATE, -- Date de délivrance du CLV
    ADD COLUMN IF NOT EXISTS sterilization_method VARCHAR(100), -- Méthode de stérilisation : ETO, Gamma, Vapeur saturée...
    ADD COLUMN IF NOT EXISTS gmdn_code VARCHAR(50), -- Code de nomenclature internationale (GMDN / EMDN / CND)
    ADD COLUMN IF NOT EXISTS economic_file_type VARCHAR(32) DEFAULT 'pcsu', -- 'pcsu' (local) ou 'cif_benchmark' (importé)
    ADD COLUMN IF NOT EXISTS pcsu_amount NUMERIC(15, 2), -- Prix de revient sortie d'usine (DZD) pour DM local
    ADD COLUMN IF NOT EXISTS cif_price NUMERIC(15, 2), -- Prix CIF unitaire pour DM importé
    ADD COLUMN IF NOT EXISTS cif_currency VARCHAR(10) DEFAULT 'EUR'; -- Devise CIF (EUR, USD, etc.)

CREATE INDEX IF NOT EXISTS idx_medical_devices_origin ON medical_devices(origin_type);
CREATE INDEX IF NOT EXISTS idx_medical_devices_manufacturer_site ON medical_devices(manufacturer_site_id);
CREATE INDEX IF NOT EXISTS idx_medical_devices_establishment ON medical_devices(establishment_id);

-- Élargissement de la colonne classification rule sur la table products parent
ALTER TABLE products ALTER COLUMN dm_classification_rule TYPE VARCHAR(255);

-- Mise à jour des dispositifs médicaux initiaux de démonstration
UPDATE medical_devices
SET
    origin_type = 'importe',
    legal_manufacturer_name = 'B. Braun Melsungen AG',
    legal_manufacturer_country = 'DE',
    ce_certificate_number = 'CE-0123-MDR-2023-7749',
    notified_body_name = 'TÜV SÜD Product Service GmbH',
    notified_body_number = '0123',
    ce_expiry_date = '2028-06-30',
    iso_13485_certificate_number = 'ISO13485-DE-TUV-9921',
    clv_fsc_number = 'FSC-DE-BFARM-2023-4410',
    clv_issuing_authority = 'Federal Institute for Drugs and Medical Devices (BfArM) - Germany',
    clv_issue_date = '2023-09-12',
    sterilization_method = 'Oxyde d''éthylène (ETO)',
    gmdn_code = '34978',
    economic_file_type = 'cif_benchmark',
    cif_price = 1.85,
    cif_currency = 'EUR',
    manufacturer_site_id = (SELECT id FROM manufacturing_sites WHERE country_code = 'DE' LIMIT 1),
    establishment_id = (SELECT id FROM establishments LIMIT 1)
WHERE commercial_name_fr ILIKE '%Cathéter%' OR manufacturer_reference ILIKE '%IVC%';

UPDATE medical_devices
SET
    origin_type = 'fabrication_locale',
    economic_file_type = 'pcsu',
    pcsu_amount = 45.00,
    sterilization_method = 'Rayonnement Gamma (25 kGy)',
    establishment_id = (SELECT id FROM establishments LIMIT 1)
WHERE origin_type IS NULL OR origin_type = 'fabrication_locale';
