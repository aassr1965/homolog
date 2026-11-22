-- =========================================================================
-- Migration 029: Optimisation des Index & Clés Étrangères (Performance BPF)
-- =========================================================================

-- 1. Index de Tri et Filtrage à Haute Fréquence (Navigation & Cockpit)
-- -------------------------------------------------------------------------
-- Élimine le 'quicksort' en mémoire lors du listing des dossiers par organisation
CREATE INDEX IF NOT EXISTS idx_dossiers_org_created_at 
    ON dossiers (organization_id, created_at DESC);

-- Permet un Index Scan direct sans tri en mémoire pour l'affichage de la checklist CTD (50-200 items)
CREATE INDEX IF NOT EXISTS idx_dossier_items_dossier_sort 
    ON dossier_checklist_items (dossier_id, module_number ASC, sort_order ASC, item_code ASC);

-- Accélération du calcul de complétude BPF
CREATE INDEX IF NOT EXISTS idx_dossier_items_completeness 
    ON dossier_checklist_items (dossier_id, is_mandatory, requirement_level, status);

-- Bordereaux de versement filtrés par dossier et ordonnés par date d'émission
CREATE INDEX IF NOT EXISTS idx_payment_slips_org_dossier_gen 
    ON anpp_payment_slips (organization_id, dossier_id, generated_at DESC);

-- Rendez-vous de dépôt ANPP ordonnés par chronologie
CREATE INDEX IF NOT EXISTS idx_anpp_appointments_org_dossier_sched 
    ON anpp_deposit_appointments (organization_id, dossier_id, scheduled_at ASC);

-- États réglementaires générés par dossier
CREATE INDEX IF NOT EXISTS idx_generated_reports_dossier_created 
    ON generated_reports (dossier_id, created_at DESC);

-- Demandes de variations et modifications DM par dossier
CREATE INDEX IF NOT EXISTS idx_anpp_variations_dossier_created 
    ON anpp_variation_requests (dossier_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_dm_modifications_dossier_created 
    ON dm_modifications (dossier_id, created_at DESC);

-- 2. Couverture des Clés Étrangères (Prévention des Verrous de Table & Lenteurs JOIN)
-- -------------------------------------------------------------------------
-- Workflow Maker-Checker & Verrous BPF sur les pièces
CREATE INDEX IF NOT EXISTS idx_dossier_items_maker_user 
    ON dossier_checklist_items (maker_user_id) WHERE maker_user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_dossier_items_checker_user 
    ON dossier_checklist_items (checker_user_id) WHERE checker_user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_dossier_items_approver_user 
    ON dossier_checklist_items (approver_user_id) WHERE approver_user_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_dossier_items_locked_by 
    ON dossier_checklist_items (locked_by) WHERE locked_by IS NOT NULL;

-- Coffre-fort GED
CREATE INDEX IF NOT EXISTS idx_ged_docs_created_by 
    ON ged_documents (created_by);

CREATE INDEX IF NOT EXISTS idx_ged_docs_locked_by 
    ON ged_documents (locked_by) WHERE locked_by IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_ged_docs_site_id 
    ON ged_documents (site_id) WHERE site_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_ged_versions_author 
    ON ged_versions (author_id);

CREATE INDEX IF NOT EXISTS idx_ged_versions_signed_by_dt 
    ON ged_versions (signed_by_dt_id) WHERE signed_by_dt_id IS NOT NULL;

-- Dossiers & Signatures PDT
CREATE INDEX IF NOT EXISTS idx_dossiers_pdt_signed_by 
    ON dossiers (pdt_signed_by) WHERE pdt_signed_by IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_dossiers_locked_by_dt 
    ON dossiers (locked_by_dt_id) WHERE locked_by_dt_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_dossiers_responsible 
    ON dossiers (responsible_id) WHERE responsible_id IS NOT NULL;

-- Piste d'audit ALCOA+ & Redevances
CREATE INDEX IF NOT EXISTS idx_audit_logs_user 
    ON audit_logs (user_id);

CREATE INDEX IF NOT EXISTS idx_payment_slips_fee_schedule 
    ON anpp_payment_slips (fee_schedule_id);

CREATE INDEX IF NOT EXISTS idx_payment_slips_verified_by 
    ON anpp_payment_slips (verified_by) WHERE verified_by IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_anpp_appointments_delegate 
    ON anpp_deposit_appointments (assigned_delegate_id) WHERE assigned_delegate_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_anpp_deficiency_tasks_assigned_due 
    ON anpp_deficiency_tasks (assigned_to, due_date ASC);

-- 3. Sécurité, Affectations & Cloisonnement Multi-Tenant
-- -------------------------------------------------------------------------
CREATE INDEX IF NOT EXISTS idx_dossier_assignments_user_exp 
    ON dossier_assignments (user_id, expires_at);

CREATE INDEX IF NOT EXISTS idx_dossier_consultants_user_exp 
    ON dossier_consultants (user_id, expires_at);
