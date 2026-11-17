-- ============================================================================
-- Migration 015 : Rétro-alignement des Départements BPF sur les Checklists Existantes
-- Conforme Section 7.4.1 du Cahier des Charges (v1.0 Novembre 2026)
-- Priorité des spécificités techniques (1.P.8, 1.P.4/5, 1.M.P.3, 1.M.S, 3.2.S, DM)
-- ============================================================================

-- 1. Réinitialisation / Fallbacks généraux par module CTD
UPDATE dossier_checklist_items SET assigned_department = 'RA' WHERE module_number = 1;
UPDATE dossier_checklist_items SET assigned_department = 'RD' WHERE module_number = 2;
UPDATE dossier_checklist_items SET assigned_department = 'QC' WHERE module_number = 3;
UPDATE dossier_checklist_items SET assigned_department = 'RD' WHERE module_number = 4;
UPDATE dossier_checklist_items SET assigned_department = 'EXTERNAL' WHERE module_number = 5;

-- 2. Règles Partenaires Externes / Tiers / CRO & API Holder (ASMF/DMF 3.2.S, CEP/Accès 1.M.S, Bioéquivalence)
UPDATE dossier_checklist_items SET assigned_department = 'EXTERNAL' 
WHERE item_code LIKE '1.M.S%' 
   OR item_code LIKE '3.2.S%' 
   OR item_code LIKE '5.3.1%';

-- 3. Règles Recherche & Développement / Galénique (Formulation 3.2.P.2, Biowaiver 1.M.P.3)
UPDATE dossier_checklist_items SET assigned_department = 'RD' 
WHERE item_code = '1.M.P.3' 
   OR item_code LIKE '3.2.P.2%';

-- 4. Règles Contrôle Qualité & Stabilités (CoA 1.P.4, 1.P.5, spécifications 3.2.S.4, méthodes 3.2.P.5, stabilités 3.2.P.8)
UPDATE dossier_checklist_items SET assigned_department = 'QC' 
WHERE item_code IN ('1.P.4', '1.P.5') 
   OR item_code LIKE '3.2.S.4%' 
   OR item_code LIKE '3.2.P.5%' 
   OR item_code LIKE '3.2.P.8%';

-- 5. Règles Production Industrielle & Procédés (3.2.P.3 et engagement 1.P.8 sur 3 lots industriels)
UPDATE dossier_checklist_items SET assigned_department = 'PROD' 
WHERE item_code LIKE '3.2.P.3%' 
   OR item_code = '1.P.8';

-- 6. Règles Dispositifs Médicaux (DM - Décret 20-324 / Arrêté 10 mai 2021)
UPDATE dossier_checklist_items SET assigned_department = 'RA' WHERE item_code LIKE 'DM.A%' OR item_code LIKE 'DM.I%';
UPDATE dossier_checklist_items SET assigned_department = 'QA' WHERE item_code LIKE 'DM.Q%';
UPDATE dossier_checklist_items SET assigned_department = 'RD' WHERE item_code LIKE 'DM.T%' OR item_code LIKE 'DM.C%';
UPDATE dossier_checklist_items SET assigned_department = 'PROD' WHERE item_code LIKE 'DM.T.3%' OR item_code LIKE 'DM.S%';
