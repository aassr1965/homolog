-- ============================================================================
-- Migration 013 : Atelier d'Élaboration Documentaire, Workflow Maker-Checker-Approver
-- & Moteur de Reporting Réglementaire BPF (Jalon 2.1)
-- Conforme : Guide CTD ANPP 2025, Décret 20-325, BPF & ALCOA+
-- ============================================================================

-- 1. Enrichissement de dossier_checklist_items pour le workflow Maker-Checker-Approver
ALTER TABLE dossier_checklist_items
    ADD COLUMN IF NOT EXISTS workflow_status VARCHAR(32) DEFAULT 'draft_pending' NOT NULL,
    -- 'draft_pending' (À rédiger), 'drafting' (En cours), 'under_review' (En revue Checker/AQ),
    -- 'pending_approval' (Prêt pour visa DT), 'approved_sealed' (Approuvé & Scellé)
    ADD COLUMN IF NOT EXISTS assigned_department VARCHAR(32) DEFAULT 'RA',
    -- 'RA' (Affaires Réglem.), 'QC' (Contrôle Qualité), 'RD' (R&D / Galénique),
    -- 'PROD' (Production Industrielle), 'QA' (Assurance Qualité), 'EXTERNAL' (Tiers/CRO)
    ADD COLUMN IF NOT EXISTS assigned_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS due_date DATE,
    -- Verrouillage collaboratif Check-out / Check-in
    ADD COLUMN IF NOT EXISTS is_locked BOOLEAN DEFAULT FALSE NOT NULL,
    ADD COLUMN IF NOT EXISTS locked_by UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS locked_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS lock_reason VARCHAR(255),
    -- Versionnage granulaire de travail (v0.1, v0.2... jusqu'à v1.0 scellée)
    ADD COLUMN IF NOT EXISTS current_revision VARCHAR(20) DEFAULT 'v0.1' NOT NULL,
    -- Rédacteur désigné (Maker)
    ADD COLUMN IF NOT EXISTS maker_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS maker_submitted_at TIMESTAMPTZ,
    -- Relecteur Qualité (Checker - AQ)
    ADD COLUMN IF NOT EXISTS checker_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS checker_reviewed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS checker_decision VARCHAR(32), -- 'approved', 'rejected'
    ADD COLUMN IF NOT EXISTS checker_notes TEXT,
    -- Approbateur Légal (Approver - Pharmacien DT titulaire 1.A.7)
    ADD COLUMN IF NOT EXISTS approver_user_id UUID REFERENCES users(id) ON DELETE SET NULL,
    ADD COLUMN IF NOT EXISTS approver_signed_at TIMESTAMPTZ,
    ADD COLUMN IF NOT EXISTS approver_signature_seal VARCHAR(64),
    -- Contenu texte dynamique rédigé (Markdown / HTML) si élaboré in-app
    ADD COLUMN IF NOT EXISTS authored_content TEXT,
    ADD COLUMN IF NOT EXISTS template_code VARCHAR(64);

-- 2. Table des Annotations et Historique de Relecture Qualité (Audit Trail Checker)
CREATE TABLE IF NOT EXISTS dossier_item_reviews (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    dossier_item_id UUID NOT NULL REFERENCES dossier_checklist_items(id) ON DELETE CASCADE,
    author_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    author_role VARCHAR(64) NOT NULL,
    review_type VARCHAR(32) NOT NULL, -- 'annotation', 'demande_modification', 'validation_qualite', 'rejet_motive'
    comment TEXT NOT NULL,
    revision_version VARCHAR(20) DEFAULT 'v0.1' NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- 3. Table des Modèles Maîtres Pharmaceutiques (Master Templates ANPP / ICH)
CREATE TABLE IF NOT EXISTS master_document_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(64) NOT NULL UNIQUE,
    title_fr VARCHAR(255) NOT NULL,
    title_ar VARCHAR(255),
    category VARCHAR(64) NOT NULL, -- 'ctd_template', 'lettre_engagement', 'sop_qualite', 'rapport_technique'
    module_number INT NOT NULL,
    sub_module VARCHAR(32) NOT NULL,
    target_procedure VARCHAR(32) DEFAULT 'ALL' NOT NULL,
    content_template TEXT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    updated_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- 4. Table des États de Sortie Générés & Certifiés (Reporting BPF ALCOA+)
CREATE TABLE IF NOT EXISTS generated_reports (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    organization_id UUID NOT NULL REFERENCES organizations(id) ON DELETE CASCADE,
    dossier_id UUID NOT NULL REFERENCES dossiers(id) ON DELETE CASCADE,
    report_code VARCHAR(32) NOT NULL, -- 'ES-REG-01', 'ES-REG-02', 'ES-FIN-01', 'EI-AVT-01', etc.
    report_title VARCHAR(255) NOT NULL,
    format VARCHAR(16) DEFAULT 'pdf' NOT NULL, -- 'pdf', 'xlsx', 'html'
    watermark_status VARCHAR(32) DEFAULT 'BROUILLON' NOT NULL, -- 'BROUILLON', 'OFFICIEL', 'ARCHIVE'
    sha256_hash VARCHAR(64) NOT NULL,
    file_size_bytes BIGINT DEFAULT 0 NOT NULL,
    signed_by_pdt_id UUID REFERENCES users(id) ON DELETE SET NULL,
    signed_at TIMESTAMPTZ,
    qr_verification_url VARCHAR(512),
    report_metadata JSONB DEFAULT '{}'::jsonb NOT NULL,
    created_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- 5. Index pour la performance des requêtes du workflow et du reporting
CREATE INDEX IF NOT EXISTS idx_dossier_items_workflow ON dossier_checklist_items(dossier_id, workflow_status);
CREATE INDEX IF NOT EXISTS idx_dossier_items_department ON dossier_checklist_items(assigned_department);
CREATE INDEX IF NOT EXISTS idx_dossier_items_assigned_user ON dossier_checklist_items(assigned_user_id);
CREATE INDEX IF NOT EXISTS idx_dossier_items_locked ON dossier_checklist_items(dossier_id, is_locked) WHERE is_locked = TRUE;
CREATE INDEX IF NOT EXISTS idx_item_reviews_item ON dossier_item_reviews(dossier_item_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_master_templates_code ON master_document_templates(code);
CREATE INDEX IF NOT EXISTS idx_generated_reports_dossier ON generated_reports(organization_id, dossier_id, report_code);
CREATE INDEX IF NOT EXISTS idx_generated_reports_sha ON generated_reports(sha256_hash);

-- 6. Injection des Modèles Maîtres par défaut (Engagements et Templates BPF)
INSERT INTO master_document_templates (code, title_fr, category, module_number, sub_module, content_template)
VALUES 
(
    'TEMPLATE-1.P.8-ENGAGEMENT-3LOTS',
    'Lettre d''engagement de validation industrielle sur les trois (03) premiers lots commerciaux',
    'lettre_engagement',
    1,
    '1.P.8',
    '# ENGAGEMENT DE VALIDATION DE PROCÉDÉ SUR LES TROIS PREMIERS LOTS COMMERCIAUX\n\n**Produit :** {{PRODUCT_NAME}}\n**DCI :** {{INN}}\n**Dosage :** {{DOSAGE}} • **Forme :** {{FORM}}\n**Site de Fabrication :** {{SITE_NAME}}\n\nEn ma qualité de **Pharmacien Directeur Technique** de l''établissement {{ESTABLISHMENT_NAME}} (Agrément MIPH n° {{AGREEMENT_NUM}}), je soussigné(e) m''engage solennellement auprès de l''Agence Nationale des Produits Pharmaceutiques (ANPP) à :\n\n1. Réaliser la validation prospective de procédé de fabrication selon le protocole approuvé sur les trois (03) premiers lots consécutifs à l''échelle industrielle commerciale.\n2. Soumettre à l''ANPP les rapports de validation complets dès achèvement des contrôles de conformité.\n3. Ne libérer aucun lot pour la distribution commerciale avant validation totale et visa qualité.'
),
(
    'TEMPLATE-1.P.9-SUIVI-STABILITE',
    'Lettre d''engagement de suivi de stabilité continue post-commercialisation en Zone IVb',
    'lettre_engagement',
    1,
    '1.P.9',
    '# ENGAGEMENT DE SUIVI DE STABILITÉ CONTINUE EN CONDITIONS ZONE IVb\n\n**Produit :** {{PRODUCT_NAME}}\n**DCI :** {{INN}}\n\nLe laboratoire {{ESTABLISHMENT_NAME}} s''engage à soumettre un lot commercial annuel de {{PRODUCT_NAME}} au programme d''études de stabilité continue en étuves régulées sous conditions climatiques Zone IVb (30°C ± 2°C / 75% HR ± 5% HR) conformément aux exigences de l''ANPP et aux directives ICH Q1A(R2).\n\nTout résultat hors spécification (OOS) ou dérive analytique constatée sera immédiatement notifié à la Direction de l''Enregistrement de l''ANPP.'
),
(
    'TEMPLATE-1.H.10-ATTESTATION-PDT',
    'Attestation d''évaluation et de conformité réglementaire par le Pharmacien Directeur Technique',
    'lettre_engagement',
    1,
    '1.H.10',
    '# ATTESTATION D''ÉVALUATION DU PHARMACIEN DIRECTEUR TECHNIQUE (ART. 1.A.7 / 1.H.10)\n\nJe soussigné(e), {{PDT_NAME}}, Pharmacien Directeur Technique désigné sous décision ministérielle n° {{PDT_DECISION_NUM}}, atteste sur l''honneur avoir procédé à l''évaluation technique, scientifique et réglementaire complète du dossier CTD du produit {{PRODUCT_NAME}} (Dossier {{DOSSIER_CODE}}).\n\nJ''atteste que les données présentées dans les cinq (05) modules sont véridiques, exactes, conformes aux Bonnes Pratiques de Fabrication algériennes et aux normes internationales en vigueur.'
)
ON CONFLICT (code) DO NOTHING;
