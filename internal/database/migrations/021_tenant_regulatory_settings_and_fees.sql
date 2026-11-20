-- ==============================================================================
-- Migration 021: Settings Réglementaires, Institutionnels & Barèmes Découplés par Tenant
-- Conforme au Chapitre 6 (Section 6.4) et Chapitre 11 du CDC
-- Découplage strict par organization_id, versioning temporel ALCOA+ (valid_from/valid_to)
-- ==============================================================================

CREATE TABLE IF NOT EXISTS tenant_regulatory_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    setting_key VARCHAR(64) NOT NULL,
    setting_value TEXT NOT NULL,
    data_type VARCHAR(32) NOT NULL DEFAULT 'string', -- 'string', 'number', 'integer', 'boolean', 'json'
    category VARCHAR(64) NOT NULL, -- 'finance_medicine', 'finance_dm', 'regulatory_deadlines', 'institutions_anpp', 'fee_schedule'
    label_fr VARCHAR(255) NOT NULL,
    label_ar VARCHAR(255),
    description TEXT,
    legal_reference VARCHAR(255),
    valid_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    valid_to TIMESTAMPTZ, -- NULL si version en vigueur active
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    pending_value TEXT,
    pending_reason TEXT,
    pending_proposed_by UUID REFERENCES users(id) ON DELETE SET NULL,
    status VARCHAR(32) NOT NULL DEFAULT 'ACTIVE', -- 'ACTIVE', 'PENDING_APPROVAL_DT', 'HISTORICAL'
    approved_by UUID REFERENCES users(id) ON DELETE SET NULL,
    approved_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_tenant_setting_version UNIQUE (organization_id, setting_key, valid_from)
);

CREATE INDEX IF NOT EXISTS idx_tenant_reg_settings_lookup 
ON tenant_regulatory_settings(organization_id, setting_key, is_active, valid_to);

CREATE INDEX IF NOT EXISTS idx_tenant_reg_settings_category 
ON tenant_regulatory_settings(organization_id, category);

-- Fonction idempotent d'injection des settings par défaut pour une organisation
CREATE OR REPLACE FUNCTION seed_tenant_default_settings(org_id UUID) 
RETURNS VOID AS $$
BEGIN
    -- 1. Circuit Médicaments
    INSERT INTO tenant_regulatory_settings (
        organization_id, setting_key, setting_value, data_type, category, label_fr, description, legal_reference, valid_from, is_active, status
    ) VALUES 
    (org_id, 'FIN_PAYMENT_MODE_MEDICINE', 'TRESOR_PUBLIC_OU_CHEQUE', 'string', 'finance_medicine', 
     'Médicaments : Mode de Règlement', 'Modalité de liquidation des dossiers DEPP (Trésor Public ou Chèque certifié)', 
     'Décret exécutif n° 20-325 & Arrêté des Redevances ANPP', NOW(), TRUE, 'ACTIVE'),
    (org_id, 'FIN_TREASURY_ACCOUNT_MEDICINE', 'CAS N° 302.145 (Fonds Spécial ANPP)', 'string', 'finance_medicine', 
     'Médicaments : Compte Trésor Public', 'Mentions légales Trésor Public pour virement et quittances', 
     'Décret exécutif n° 21-82 & Ordonnance 21-07', NOW(), TRUE, 'ACTIVE'),
    (org_id, 'FIN_BENEFICIARY_ORDER_MEDICINE', 'Régisseur des recettes de l''ANPP', 'string', 'finance_medicine', 
     'Médicaments : Ordre Bénéficiaire Chèque', 'Libellé d''ordre des chèques certifiés et bordereaux BV1/BV2', 
     'Régie des Recettes de l''ANPP Alger', NOW(), TRUE, 'ACTIVE')
    ON CONFLICT (organization_id, setting_key, valid_from) DO NOTHING;

    -- 2. Circuit Dispositifs Médicaux
    INSERT INTO tenant_regulatory_settings (
        organization_id, setting_key, setting_value, data_type, category, label_fr, description, legal_reference, valid_from, is_active, status
    ) VALUES 
    (org_id, 'FIN_BANK_NAME_DM', 'BNA Agence Al Qods — Chéraga (Guichet 00622)', 'string', 'finance_dm', 
     'Dispositifs Médicaux : Banque Domiciliataire', 'Établissement bancaire recevant les redevances DDM', 
     'Ordonnance n° 21-07 du 8 juin 2021', NOW(), TRUE, 'ACTIVE'),
    (org_id, 'FIN_BANK_RIB_DM', '00100622030000064395', 'string', 'finance_dm', 
     'Dispositifs Médicaux : RIB ANPP Officiel', 'Clé RIB officielle pour bordereaux ES-FIN-01 (DM) et formulaires ES-REG-02-DM', 
     'Compte BNA Chéraga DDM ANPP', NOW(), TRUE, 'ACTIVE')
    ON CONFLICT (organization_id, setting_key, valid_from) DO NOTHING;

    -- 3. Barèmes Médicaments Décret 21-82
    INSERT INTO tenant_regulatory_settings (
        organization_id, setting_key, setting_value, data_type, category, label_fr, description, legal_reference, valid_from, is_active, status
    ) VALUES 
    (org_id, 'FEE_MEDICINE_NI_DZD', '355000.00', 'number', 'fee_schedule', 
     'Barème PP : Enregistrement Initial Importé (NI)', 'Montant de redevance pour dépôt initial produit importé', 
     'Décret exécutif 21-82 & Arrêté Redevances', NOW(), TRUE, 'ACTIVE'),
    (org_id, 'FEE_MEDICINE_NF_DZD', '150000.00', 'number', 'fee_schedule', 
     'Barème PP : Enregistrement Fabrication Locale (NF)', 'Montant de redevance pour produit fabriqué localement', 
     'Décret exécutif 21-82 & Arrêté Redevances', NOW(), TRUE, 'ACTIVE'),
    (org_id, 'FEE_MEDICINE_RI_DZD', '205000.00', 'number', 'fee_schedule', 
     'Barème PP : Renouvellement Quinquennal (RI/RF)', 'Montant de redevance pour renouvellement de la DE', 
     'Décret exécutif 21-82 & Arrêté Redevances', NOW(), TRUE, 'ACTIVE'),
    (org_id, 'FEE_MEDICINE_MI_DZD', '115000.00', 'number', 'fee_schedule', 
     'Barème PP : Variation Majeure (Mmaj)', 'Montant unitaire par modification majeure post-enregistrement', 
     'Ligne Directrice LD-DEPP07/V01 & Décret 21-82', NOW(), TRUE, 'ACTIVE'),
    (org_id, 'FEE_MEDICINE_MIN_DZD', '50000.00', 'number', 'fee_schedule', 
     'Barème PP : Variation Mineure / Modérée (Mmin/Mmod)', 'Montant unitaire par modification mineure ou modérée', 
     'Ligne Directrice LD-DEPP07/V01 & Décret 21-82', NOW(), TRUE, 'ACTIVE')
    ON CONFLICT (organization_id, setting_key, valid_from) DO NOTHING;

    -- 4. Barèmes Dispositifs Médicaux Ordonnance 21-07 Art. 31
    INSERT INTO tenant_regulatory_settings (
        organization_id, setting_key, setting_value, data_type, category, label_fr, description, legal_reference, valid_from, is_active, status
    ) VALUES 
    (org_id, 'FEE_DM_HOMOLOGATION_DZD', '300000.00', 'number', 'fee_schedule', 
     'Barème DM : Homologation Initiale', 'Redevance d''homologation d''un dispositif médical', 
     'Ordonnance n° 21-07 Art. 31 & Arrêté 10 Mai 2021', NOW(), TRUE, 'ACTIVE'),
    (org_id, 'FEE_DM_RENEWAL_DZD', '300000.00', 'number', 'fee_schedule', 
     'Barème DM : Renouvellement Quinquennal DM', 'Redevance de renouvellement de décision d''homologation DM', 
     'Ordonnance n° 21-07 Art. 31 & Arrêté 10 Mai 2021', NOW(), TRUE, 'ACTIVE'),
    (org_id, 'FEE_DM_MODIFICATION_DZD', '150000.00', 'number', 'fee_schedule', 
     'Barème DM : Modification Technique DM', 'Redevance de modification de dossier d''homologation DM', 
     'Ordonnance n° 21-07 Art. 31', NOW(), TRUE, 'ACTIVE'),
    (org_id, 'FEE_DM_TRANSFER_DZD', '100000.00', 'number', 'fee_schedule', 
     'Barème DM : Transfert / Cession DM', 'Redevance de transfert d''homologation DM à un tiers titulaire', 
     'Ordonnance n° 21-07 Art. 31', NOW(), TRUE, 'ACTIVE')
    ON CONFLICT (organization_id, setting_key, valid_from) DO NOTHING;

    -- 5. Délais Légaux de Carence Réglementaire
    INSERT INTO tenant_regulatory_settings (
        organization_id, setting_key, setting_value, data_type, category, label_fr, description, legal_reference, valid_from, is_active, status
    ) VALUES 
    (org_id, 'REG_DEFICIENCY_DEADLINE_STANDARD_DAYS', '90', 'integer', 'regulatory_deadlines', 
     'Délai Carence Réglementaire (Standard)', 'Délai légal standard pour lever les compléments ANPP (jours calendaires)', 
     'Lignes Directrices ANPP & Décret 20-325', NOW(), TRUE, 'ACTIVE'),
    (org_id, 'REG_DEFICIENCY_DEADLINE_URGENT_DAYS', '45', 'integer', 'regulatory_deadlines', 
     'Délai Carence Réglementaire (Relance / Urgence)', 'Délai réduit en cas de mise en demeure ou relance formelle', 
     'Procédure d''Urgence ANPP', NOW(), TRUE, 'ACTIVE')
    ON CONFLICT (organization_id, setting_key, valid_from) DO NOTHING;

    -- 6. Adresses & Guichets Institutionnels ANPP
    INSERT INTO tenant_regulatory_settings (
        organization_id, setting_key, setting_value, data_type, category, label_fr, description, legal_reference, valid_from, is_active, status
    ) VALUES 
    (org_id, 'INST_ANPP_DEPP_ADDRESS', 'Direction de l''Enregistrement des Produits Pharmaceutiques, Rue Didouche Mourad, Alger', 'string', 'institutions_anpp', 
     'Guichet Dépôt DEPP', 'Adresse officielle pour dépôts des dossiers CTD Médicaments', 
     'Siège ANPP Alger', NOW(), TRUE, 'ACTIVE'),
    (org_id, 'INST_ANPP_DDM_ADDRESS', 'Direction des Dispositifs Médicaux, Alger', 'string', 'institutions_anpp', 
     'Guichet Dépôt DDM', 'Adresse officielle pour dépôts des dossiers d''Homologation DM', 
     'DDM ANPP Alger', NOW(), TRUE, 'ACTIVE'),
    (org_id, 'INST_ANPP_DCE_ADDRESS', 'Laboratoire National de Contrôle des Produits Pharmaceutiques (LNCPP), Alger', 'string', 'institutions_anpp', 
     'Laboratoire Contrôle DCE / LNCPP', 'Adresse pour envoi des échantillons analytiques officiels', 
     'DCE / LNCPP Alger', NOW(), TRUE, 'ACTIVE')
    ON CONFLICT (organization_id, setting_key, valid_from) DO NOTHING;
END;
$$ LANGUAGE plpgsql;

-- Initialisation des paramètres par défaut pour toutes les organisations existantes
DO $$
DECLARE
    org RECORD;
BEGIN
    FOR org IN SELECT id FROM organizations LOOP
        PERFORM seed_tenant_default_settings(org.id);
    END LOOP;
END;
$$;
