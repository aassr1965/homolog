-- ============================================================================
-- Migration 002 : Référentiel Établissements & Produits (Guide ANPP 2025)
-- Harmonisation Réglementaire : Princeps, Générique Importé, Générique Local
-- Conformité Décret 20-325, Décret 20-324, Arrêté 10/05/2021 & Lignes Directrices
-- ============================================================================

-- 1. Table des Établissements Pharmaceutiques (Demandeur / Détenteur / Titulaire)
CREATE TABLE IF NOT EXISTS establishments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    name_fr VARCHAR(255) NOT NULL,
    name_ar VARCHAR(255),
    legal_status VARCHAR(100), -- 'SARL', 'SPA', 'EURL', 'SNC', etc.
    miph_license_number VARCHAR(100), -- Agrément ministériel MIPH
    miph_license_grant_date DATE,
    miph_license_expiry_date DATE,
    technical_director_name VARCHAR(255) NOT NULL, -- Pharmacien Directeur Technique (PDT - 1.A.7)
    technical_director_decision_ref VARCHAR(100), -- Décision ministérielle nominative d'exercice
    technical_director_decision_date DATE,
    address_fr TEXT,
    address_ar TEXT,
    wilaya VARCHAR(100),
    phone VARCHAR(50),
    email VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- 2. Table des Sites de Fabrication, Conditionnement et Contrôle (1.H.8 & 1.M)
CREATE TABLE IF NOT EXISTS manufacturing_sites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    establishment_id UUID REFERENCES establishments(id) ON DELETE SET NULL,
    site_name_fr VARCHAR(255) NOT NULL,
    site_name_ar VARCHAR(255),
    country_code VARCHAR(2) DEFAULT 'DZ' NOT NULL, -- Code ISO-3166-1 alpha-2 (DZ, FR, IN, etc.)
    address TEXT,
    site_role VARCHAR(64) NOT NULL, -- 'synthese_sa', 'intermediaire', 'formulateur_pf', 'conditionneur_primaire', 'conditionneur_secondaire', 'controle_analytique', 'stockage'
    manufacturing_license_number VARCHAR(100), -- Licence ML (1.M.S.1, 1.M.P.1)
    gmp_certificate_number VARCHAR(100), -- Certificat BPF/GMP (1.M.S.2, 1.M.P.2)
    gmp_grant_date DATE,
    gmp_expiry_date DATE,
    gmp_issuing_authority VARCHAR(255), -- ex: ANPP, ANSM, FDA, EDQM, WHO
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- 3. Table Détaillée des Médicaments (Conforme Typologie ANPP Juillet 2025)
CREATE TABLE IF NOT EXISTS medicines (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    product_id UUID UNIQUE REFERENCES products(id) ON DELETE CASCADE,
    origin_type VARCHAR(32) NOT NULL, -- 'princeps', 'generique_importe', 'generique_local'
    brand_name_fr VARCHAR(255) NOT NULL,
    brand_name_ar VARCHAR(255),
    dci_fr VARCHAR(255) NOT NULL,
    dci_ar VARCHAR(255),
    dosage_fr VARCHAR(100) NOT NULL,
    dosage_ar VARCHAR(100),
    pharmaceutical_form_fr VARCHAR(100) NOT NULL,
    pharmaceutical_form_ar VARCHAR(100),
    administration_route_fr VARCHAR(100) NOT NULL,
    administration_route_ar VARCHAR(100),
    atc_code VARCHAR(20), -- Code anatomique, thérapeutique et chimique OMS
    
    -- Attributs Spécifiques Substance Active (3.2.S)
    has_pharmacopoeia_monograph BOOLEAN DEFAULT TRUE,
    pharmacopoeia_name VARCHAR(100) DEFAULT 'Ph. Eur.',
    has_cep BOOLEAN DEFAULT FALSE,
    cep_number VARCHAR(100), -- CEP EDQM (1.M.S.3 / 3.2.S.3)
    cep_access_letter_held BOOLEAN DEFAULT FALSE, -- Lettre d'accès nominative FPP
    apimf_dmf_number VARCHAR(100), -- Dossier Permanent de la Substance Active (DPSA)
    polymorphic_form VARCHAR(255), -- Forme cristalline / polymorphisme (1.P.4 / 3.2.S.1.3)
    particle_size_criticality BOOLEAN DEFAULT FALSE, -- Criticité granulométrie (1.P.4 / 3.2.S.4)
    
    -- Attributs Spécifiques Produit Fini (3.2.P)
    storage_conditions_fr VARCHAR(255), -- ex: "A conserver à une température ne dépassant pas 25°C"
    storage_conditions_ar VARCHAR(255),
    shelf_life_months INT DEFAULT 36,
    in_use_shelf_life VARCHAR(100), -- Péremption après première ouverture ou reconstitution
    validation_batch_files_held BOOLEAN DEFAULT FALSE, -- Dossiers lots validation 3.2.P.3.5 / 1.P.8
    
    -- Données Économiques Réglementaires (1.H.4 & 1.H.5)
    economic_file_type VARCHAR(32) DEFAULT 'pcsu', -- 'pcsu' (fabrication locale) ou 'fob_benchmark' (importation)
    pcsu_amount NUMERIC(12,2), -- Prix de Revient / Sortie Usine (DZD)
    fob_price NUMERIC(12,2), -- Prix FOB
    fob_currency VARCHAR(3) DEFAULT 'EUR',
    has_chamber_of_commerce_visa BOOLEAN DEFAULT FALSE, -- Visa Chambre de Commerce prix FOB
    has_benchmark_dossier BOOLEAN DEFAULT FALSE, -- Dossier Benchmark prix international (1.H.5)
    
    -- Références Réglementaires ANPP
    registration_decision_number VARCHAR(100), -- N° DE
    registration_date DATE,
    expiry_date DATE, -- Échéance quinquennale
    status VARCHAR(32) DEFAULT 'en_preparation' NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- 4. Table Détaillée des Dispositifs Médicaux (Décret 20-324 & Arrêté 10/05/2021)
CREATE TABLE IF NOT EXISTS medical_devices (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    product_id UUID UNIQUE REFERENCES products(id) ON DELETE CASCADE,
    commercial_name_fr VARCHAR(255) NOT NULL,
    commercial_name_ar VARCHAR(255),
    manufacturer_reference VARCHAR(100) NOT NULL,
    dm_class VARCHAR(20) NOT NULL, -- 'I', 'IIa', 'IIb', 'III', 'DMDIV'
    classification_rule VARCHAR(100), -- Règle selon l'annexe de l'arrêté du 10 mai 2021
    is_sterile BOOLEAN DEFAULT FALSE,
    is_single_use BOOLEAN DEFAULT TRUE,
    intended_use_fr TEXT,
    intended_use_ar TEXT,
    homologation_decision_number VARCHAR(100), -- N° Décision d'Homologation
    homologation_date DATE,
    homologation_expiry_date DATE,
    status VARCHAR(32) DEFAULT 'en_cours' NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- 5. Table de liaison Médicaments <-> Sites Intervenants (Cartographie 1.H.8)
CREATE TABLE IF NOT EXISTS medicine_manufacturing_sites (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    medicine_id UUID NOT NULL REFERENCES medicines(id) ON DELETE CASCADE,
    site_id UUID NOT NULL REFERENCES manufacturing_sites(id) ON DELETE CASCADE,
    role_in_product VARCHAR(64) NOT NULL, -- 'fabricant_sa', 'fabricant_pf', 'conditionneur_primaire', 'conditionneur_secondaire', 'controle_liberation'
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    CONSTRAINT uq_med_site_role UNIQUE (medicine_id, site_id, role_in_product)
);

-- Index d'optimisation
CREATE INDEX IF NOT EXISTS idx_establishments_org ON establishments(organization_id);
CREATE INDEX IF NOT EXISTS idx_manufacturing_sites_org ON manufacturing_sites(organization_id);
CREATE INDEX IF NOT EXISTS idx_manufacturing_sites_est ON manufacturing_sites(establishment_id);
CREATE INDEX IF NOT EXISTS idx_manufacturing_sites_gmp_expiry ON manufacturing_sites(gmp_expiry_date);
CREATE INDEX IF NOT EXISTS idx_medicines_org ON medicines(organization_id);
CREATE INDEX IF NOT EXISTS idx_medicines_product ON medicines(product_id);
CREATE INDEX IF NOT EXISTS idx_medicines_origin ON medicines(origin_type);
CREATE INDEX IF NOT EXISTS idx_medicines_atc ON medicines(atc_code);
CREATE INDEX IF NOT EXISTS idx_medicines_expiry ON medicines(expiry_date);
CREATE INDEX IF NOT EXISTS idx_medical_devices_org ON medical_devices(organization_id);
CREATE INDEX IF NOT EXISTS idx_medical_devices_product ON medical_devices(product_id);
CREATE INDEX IF NOT EXISTS idx_medical_devices_class ON medical_devices(dm_class);
CREATE INDEX IF NOT EXISTS idx_medical_devices_expiry ON medical_devices(homologation_expiry_date);
CREATE INDEX IF NOT EXISTS idx_med_sites_medicine ON medicine_manufacturing_sites(medicine_id);
CREATE INDEX IF NOT EXISTS idx_med_sites_site ON medicine_manufacturing_sites(site_id);
