-- ============================================================================
-- Migration 005 : PSSI, RBAC Dynamique 14 Permissions & Matrice des 9 Rôles BPF
-- Conforme Section 7.1, 7.2 & 7.3 du Cahier des Charges (v1.0 ANPP 2025/2026)
-- 21 CFR Part 11, ALCOA+, Défense en Profondeur et Ségrégation des Tâches (SoD)
-- ============================================================================

-- 1. Table des Rôles (Système & Personnalisés)
CREATE TABLE IF NOT EXISTS roles (
    id VARCHAR(64) PRIMARY KEY,
    name_fr VARCHAR(100) NOT NULL,
    name_ar VARCHAR(100),
    description TEXT,
    is_system BOOLEAN DEFAULT TRUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- 2. Table du Référentiel des 14 Permissions Unitaires
CREATE TABLE IF NOT EXISTS permissions (
    id VARCHAR(64) PRIMARY KEY,
    name_fr VARCHAR(100) NOT NULL,
    description TEXT,
    category VARCHAR(64) NOT NULL, -- 'SYSTEM', 'REFERENTIEL', 'DOSSIER', 'GED', 'AUDIT'
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL
);

-- 3. Table de Jointure Rôle <-> Permissions (Matrice par défaut)
CREATE TABLE IF NOT EXISTS role_permissions (
    role_id VARCHAR(64) NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id VARCHAR(64) NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    PRIMARY KEY (role_id, permission_id)
);

-- 4. Table des Surcharges Nominatives (Dérogations et permissions spécifiques par utilisateur)
CREATE TABLE IF NOT EXISTS user_permissions (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    permission_id VARCHAR(64) NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    granted BOOLEAN DEFAULT TRUE NOT NULL, -- TRUE: octroi exceptionnel, FALSE: révocation explicite
    assigned_by UUID REFERENCES users(id) ON DELETE SET NULL,
    created_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    PRIMARY KEY (user_id, permission_id)
);

-- 5. Versioning de permissions pour invalidation instantanée du cache en mémoire Go
ALTER TABLE users ADD COLUMN IF NOT EXISTS permissions_version INT DEFAULT 1 NOT NULL;
ALTER TABLE organizations ADD COLUMN IF NOT EXISTS permissions_version INT DEFAULT 1 NOT NULL;

-- 6. Insertion des 14 Permissions Unitaires Officielles (CDC Section 7.2)
INSERT INTO permissions (id, name_fr, description, category) VALUES
    ('SYS_MAINT', 'Maintenance Système', 'Maintenance système, sauvegardes, configuration TLS et provisioning', 'SYSTEM'),
    ('USER_MANAGE', 'Gestion Utilisateurs', 'Administration des comptes utilisateurs, assignation des rôles internes', 'SYSTEM'),
    ('REF_VIEW', 'Consultation Référentiel', 'Consultation du référentiel des médicaments, dispositifs médicaux et établissements', 'REFERENTIEL'),
    ('REF_EDIT', 'Édition Référentiel', 'Création et mise à jour des produits, sites de fabrication et établissements', 'REFERENTIEL'),
    ('DOSSIER_VIEW', 'Consultation Dossiers', 'Consultation de l arborescence des dossiers d enregistrement CTD et homologation DM', 'DOSSIER'),
    ('DOSSIER_CREATE', 'Création Dossiers', 'Création et instanciation d un nouveau dossier réglementaire ANPP', 'DOSSIER'),
    ('DOC_READ', 'Lecture Documents GED', 'Visualisation et téléchargement des documents GED et pièces de soumission CTD', 'GED'),
    ('DOC_DRAFT_MAKER', 'Rédaction Document (Maker)', 'Création de brouillon, édition collaborative et import de nouvelles versions', 'GED'),
    ('DOC_CHECK_REVIEW', 'Relecture Qualité (Checker)', 'Vérification qualité, formulation d observations et validation intermédiaire', 'GED'),
    ('DOC_LOCK_CHECKOUT', 'Verrouillage Check-out', 'Verrouillage exclusif temporaire d un document en cours de révision', 'GED'),
    ('DOC_APPROVE_PDT', 'Visa Légal PDT (Approver)', 'Approbation finale réglementaire et signature électronique 21 CFR Part 11 du Pharmacien Directeur Technique', 'GED'),
    ('DOC_ARCHIVE', 'Archivage Documents', 'Déclaration d obsolescence et archivage définitif de documents ou dossiers', 'GED'),
    ('EXPORT_PACKAGE', 'Export Package Soumission', 'Génération du dossier de soumission E-TASJIL et exports de conformité ANPP', 'DOSSIER'),
    ('AUDIT_VIEW', 'Consultation Piste d Audit', 'Consultation du journal d audit immuable audit_logs et rapports médico-légaux', 'AUDIT')
ON CONFLICT (id) DO UPDATE SET
    name_fr = EXCLUDED.name_fr,
    description = EXCLUDED.description,
    category = EXCLUDED.category;

-- 7. Insertion des 9 Rôles Réglementaires Stricts (CDC Section 7.1)
INSERT INTO roles (id, name_fr, name_ar, description, is_system) VALUES
    ('super_admin_editeur', 'Super Administrateur Éditeur SaaS', 'المشرف العام لمنصة البرمجيات', 'Administration de la plateforme SaaS multi-tenant, maintenance globale et provisioning', TRUE),
    ('admin_systeme_local', 'Administrateur Système Local', 'مدير النظام المحلي', 'Administration technique locale, gestion des serveurs, backups et certificats TLS (Zero-Trust sur données santé)', TRUE),
    ('admin_laboratoire', 'Administrateur Laboratoire', 'مسؤول إدارة المؤسسة الصيدلانية', 'Gestion des collaborateurs du laboratoire, attribution des profils, quotas et configuration des accès', TRUE),
    ('responsable_affaires_reglementaires', 'Responsable Affaires Réglementaires (RA)', 'مسؤول الشؤون التنظيمية', 'Coordination réglementaire, montage des dossiers CTD/DM, contrôle qualité des pièces (Checker)', TRUE),
    ('directeur_technique', 'Pharmacien Directeur Technique (PDT)', 'الصيدلي المدير التقني', 'Responsable légal et ordinal titulaire de la décision ministérielle (1.A.7). Détenteur exclusif du visa légal (Approver)', TRUE),
    ('contributeur_qualite_bpf', 'Contributeur Qualité BPF', 'محرر ضمان ومراقبة الجودة', 'Assurance et Contrôle Qualité, rédaction et relecture des modules 3, procédures SOP et stabilités Zone IVb', TRUE),
    ('contributeur_rd_production', 'Contributeur R&D / Production', 'محرر البحث والتطوير والإنتاج', 'Développement galénique, validation de procédé 3 lots, dossiers de lots et données techniques', TRUE),
    ('consultant_externe', 'Consultant Externe (CRO / Expert)', 'خبير استشاري خارجي', 'Expert tiers intervenant sur un dossier spécifique (accès cloisonné au seul périmètre assigné)', TRUE),
    ('auditeur_inspecteur', 'Auditeur / Inspecteur ANPP', 'مفتش / مدقق معتمد', 'Inspecteur de l autorité sanitaire ou auditeur qualité (accès en lecture seule avec piste d audit)', TRUE)
ON CONFLICT (id) DO UPDATE SET
    name_fr = EXCLUDED.name_fr,
    name_ar = EXCLUDED.name_ar,
    description = EXCLUDED.description;

-- 8. Attribution de la Matrice Complète Rôles <-> Permissions (CDC Section 7.2)

-- Nettoyage préventif des assignations système par défaut
DELETE FROM role_permissions WHERE role_id IN (
    'super_admin_editeur', 'admin_systeme_local', 'admin_laboratoire',
    'responsable_affaires_reglementaires', 'directeur_technique',
    'contributeur_qualite_bpf', 'contributeur_rd_production',
    'consultant_externe', 'auditeur_inspecteur'
);

-- (1) Super Administrateur SaaS
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('super_admin_editeur', 'SYS_MAINT');

-- (2) Administrateur Système Local
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('admin_systeme_local', 'SYS_MAINT');

-- (3) Administrateur Laboratoire
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('admin_laboratoire', 'USER_MANAGE'),
    ('admin_laboratoire', 'REF_VIEW'),
    ('admin_laboratoire', 'DOSSIER_VIEW'),
    ('admin_laboratoire', 'AUDIT_VIEW');

-- (4) Responsable Affaires Réglementaires (RA)
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('responsable_affaires_reglementaires', 'REF_VIEW'),
    ('responsable_affaires_reglementaires', 'REF_EDIT'),
    ('responsable_affaires_reglementaires', 'DOSSIER_VIEW'),
    ('responsable_affaires_reglementaires', 'DOSSIER_CREATE'),
    ('responsable_affaires_reglementaires', 'DOC_READ'),
    ('responsable_affaires_reglementaires', 'DOC_DRAFT_MAKER'),
    ('responsable_affaires_reglementaires', 'DOC_CHECK_REVIEW'),
    ('responsable_affaires_reglementaires', 'DOC_LOCK_CHECKOUT'),
    ('responsable_affaires_reglementaires', 'DOC_ARCHIVE'),
    ('responsable_affaires_reglementaires', 'EXPORT_PACKAGE');

-- (5) Pharmacien Directeur Technique (PDT)
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('directeur_technique', 'REF_VIEW'),
    ('directeur_technique', 'REF_EDIT'),
    ('directeur_technique', 'DOSSIER_VIEW'),
    ('directeur_technique', 'DOSSIER_CREATE'),
    ('directeur_technique', 'DOC_READ'),
    ('directeur_technique', 'DOC_DRAFT_MAKER'),
    ('directeur_technique', 'DOC_CHECK_REVIEW'),
    ('directeur_technique', 'DOC_LOCK_CHECKOUT'),
    ('directeur_technique', 'DOC_APPROVE_PDT'),
    ('directeur_technique', 'DOC_ARCHIVE'),
    ('directeur_technique', 'EXPORT_PACKAGE'),
    ('directeur_technique', 'AUDIT_VIEW');

-- (6) Contributeur Qualité BPF
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('contributeur_qualite_bpf', 'REF_VIEW'),
    ('contributeur_qualite_bpf', 'DOSSIER_VIEW'),
    ('contributeur_qualite_bpf', 'DOC_READ'),
    ('contributeur_qualite_bpf', 'DOC_DRAFT_MAKER'),
    ('contributeur_qualite_bpf', 'DOC_CHECK_REVIEW'),
    ('contributeur_qualite_bpf', 'DOC_LOCK_CHECKOUT');

-- (7) Contributeur R&D / Production
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('contributeur_rd_production', 'REF_VIEW'),
    ('contributeur_rd_production', 'DOSSIER_VIEW'),
    ('contributeur_rd_production', 'DOC_READ'),
    ('contributeur_rd_production', 'DOC_DRAFT_MAKER'),
    ('contributeur_rd_production', 'DOC_LOCK_CHECKOUT');

-- (8) Consultant Externe (Périmètre cloisonné par assignation de dossier)
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('consultant_externe', 'REF_VIEW'),
    ('consultant_externe', 'DOSSIER_VIEW'),
    ('consultant_externe', 'DOC_READ'),
    ('consultant_externe', 'DOC_DRAFT_MAKER'),
    ('consultant_externe', 'DOC_LOCK_CHECKOUT');

-- (9) Auditeur / Inspecteur
INSERT INTO role_permissions (role_id, permission_id) VALUES
    ('auditeur_inspecteur', 'REF_VIEW'),
    ('auditeur_inspecteur', 'DOSSIER_VIEW'),
    ('auditeur_inspecteur', 'DOC_READ'),
    ('auditeur_inspecteur', 'AUDIT_VIEW');

-- 9. Index de performance pour les contrôles d accès à haute fréquence
CREATE INDEX IF NOT EXISTS idx_role_perms_role ON role_permissions(role_id);
CREATE INDEX IF NOT EXISTS idx_role_perms_perm ON role_permissions(permission_id);
CREATE INDEX IF NOT EXISTS idx_user_perms_user ON user_permissions(user_id);
CREATE INDEX IF NOT EXISTS idx_users_perm_version ON users(permissions_version);
