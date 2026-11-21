-- Migration 022 : Durcissement du Cloisonnement BPF & Principe Need-to-Know pour l'Administrateur Laboratoire
-- Décret exécutif 20-325 & 21 CFR Part 11 : L'Administrateur Laboratoire est un profil de gouvernance organisationnelle/RH/Accès
-- Il n'a aucun besoin légitime d'accéder aux formules de fabrication, études de stabilités et dossiers techniques CTD/DM.

DELETE FROM role_permissions WHERE role_id = 'admin_laboratoire' AND permission_id = 'DOSSIER_VIEW';

UPDATE roles 
SET description = 'Gestion des collaborateurs du laboratoire, attribution des profils, quotas et configuration des accès (Cloisonné Need-to-Know : sans accès aux dossiers techniques CTD/DM)'
WHERE id = 'admin_laboratoire';
