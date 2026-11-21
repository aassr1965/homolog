// ============================================================================
// CONSOLE D'ADMINISTRATION LABORATOIRE & MATRICE DES HABILITATIONS RBAC
// Conformité PSSI, BPF Algérie (Décret 21-82) & 21 CFR Part 11
// ============================================================================

let currentAdminSection = 'users'; // 'users' | 'matrix' | 'policies'
let adminUsers = [];
let adminRoles = [];
let adminPermissions = [];
let selectedAdminUserIndex = 0;

// Métadonnées exhaustives et descriptives pour les 14 permissions unitaires (CDC Section 7.1)
const RBAC_PERM_METADATA = {
    "REF_VIEW": { icon: "🏢", label: "Voir Réf.", fullTitle: "Consulter Référentiel", desc: "Consultation du Référentiel MIPH, Établissements, Sites et Produits", cat: "ref", isLast: false },
    "REF_EDIT": { icon: "✏️", label: "Édit Réf.", fullTitle: "Éditer Référentiel", desc: "Modification et création de Produits, Établissements et Sites", cat: "ref", isLast: true },

    "DOSSIER_VIEW": { icon: "📁", label: "Voir CTD", fullTitle: "Consulter Dossiers", desc: "Consultation des dossiers d'enregistrement CTD et homologation DM", cat: "dossier", isLast: false },
    "DOSSIER_EDIT": { icon: "📝", label: "Rédig CTD", fullTitle: "Rédiger Dossiers", desc: "Édition, composition des modules CTD et rattachement des pièces", cat: "dossier", isLast: false },
    "DOSSIER_SUBMIT": { icon: "🚀", label: "Dépôt ANPP", fullTitle: "Soumettre Dépôt", desc: "Verrouillage et soumission officielle du dossier pour dépôt ANPP", cat: "dossier", isLast: false },
    "VARIATION_MANAGE": { icon: "🔄", label: "Variations", fullTitle: "Gérer Variations", desc: "Instruction et gestion des 12 types de variations post-enregistrement", cat: "dossier", isLast: true },

    "DOC_READ": { icon: "📖", label: "Lire GED", fullTitle: "Lire / Télécharger GED", desc: "Lecture et téléchargement des pièces GED (bloqué aux profils IT par Zero-Trust)", cat: "ged", isLast: false },
    "DOC_WRITE": { icon: "📤", label: "Dépôt GED", fullTitle: "Déposer en GED", desc: "Téléversement de pièces, nouvelles versions et check-in/out", cat: "ged", isLast: false },
    "DOC_REVIEW": { icon: "🔍", label: "Revue BPF", fullTitle: "Revue Qualité BPF", desc: "Revue technique contradictoire et vérification qualité BPF (Double Contrôle)", cat: "ged", isLast: false },
    "DOC_APPROVE_PDT": { icon: "✍️", label: "Visa PDT", fullTitle: "Visa Légal DT (1.A.7)", desc: "Signature électronique légale exclusive du Pharmacien Directeur Technique", cat: "ged", isLast: false },
    "EXPORT_LEGAL": { icon: "📦", label: "Export ANPP", fullTitle: "Export Légal Souverain", desc: "Génération de l'archive officielle scellée ANPP et bordereaux BV", cat: "ged", isLast: true },

    "AUDIT_VIEW": { icon: "🛡️", label: "Piste Audit", fullTitle: "Audit Trail 21 CFR", desc: "Consultation des journaux d'audit immuables ALCOA+ (21 CFR Part 11)", cat: "sys", isLast: false },
    "USER_MANAGE": { icon: "👥", label: "Comptes RH", fullTitle: "Gestion Utilisateurs", desc: "Gestion des comptes collaborateurs, attribution des rôles et dérogations", cat: "sys", isLast: false },
    "SYS_MAINT": { icon: "⚙️", label: "Maint. IT", fullTitle: "Maintenance Système", desc: "Administration système, sauvegardes et certificats TLS", cat: "sys", isLast: true }
};

const RBAC_PERMS_LIST = [
    { code: "REF_VIEW", ...RBAC_PERM_METADATA["REF_VIEW"] },
    { code: "REF_EDIT", ...RBAC_PERM_METADATA["REF_EDIT"] },
    { code: "DOSSIER_VIEW", ...RBAC_PERM_METADATA["DOSSIER_VIEW"] },
    { code: "DOSSIER_EDIT", ...RBAC_PERM_METADATA["DOSSIER_EDIT"] },
    { code: "DOSSIER_SUBMIT", ...RBAC_PERM_METADATA["DOSSIER_SUBMIT"] },
    { code: "VARIATION_MANAGE", ...RBAC_PERM_METADATA["VARIATION_MANAGE"] },
    { code: "DOC_READ", ...RBAC_PERM_METADATA["DOC_READ"] },
    { code: "DOC_WRITE", ...RBAC_PERM_METADATA["DOC_WRITE"] },
    { code: "DOC_REVIEW", ...RBAC_PERM_METADATA["DOC_REVIEW"] },
    { code: "DOC_APPROVE_PDT", ...RBAC_PERM_METADATA["DOC_APPROVE_PDT"] },
    { code: "EXPORT_LEGAL", ...RBAC_PERM_METADATA["EXPORT_LEGAL"] },
    { code: "AUDIT_VIEW", ...RBAC_PERM_METADATA["AUDIT_VIEW"] },
    { code: "USER_MANAGE", ...RBAC_PERM_METADATA["USER_MANAGE"] },
    { code: "SYS_MAINT", ...RBAC_PERM_METADATA["SYS_MAINT"] }
];

// Rendu de l'arborescence dédiée du domaine Administration
function renderAdminTree() {
    const container = document.getElementById("tree-root");
    if (!container) return;
    const canManageUsers = typeof hasPermission === 'function' && hasPermission('USER_MANAGE');
    const usersCount = adminUsers.length > 0 ? `${adminUsers.length} actifs` : '9 actifs';

    let html = '';
    if (canManageUsers) {
        html += `
            <div class="tree-node ${currentAdminSection === 'users' ? 'selected' : ''}" onclick="loadAdminSection('users')">
                <span class="chevron">👥</span>
                <span>Collaborateurs & Rôles</span>
                <span class="count-badge ok">${usersCount}</span>
            </div>
            <div class="tree-node ${currentAdminSection === 'matrix' ? 'selected' : ''}" onclick="loadAdminSection('matrix')">
                <span class="chevron">🔐</span>
                <span>Matrice 14 Permissions</span>
                <span class="count-badge ok">Tableau BPF</span>
            </div>
            <div class="tree-node ${currentAdminSection === 'policies' ? 'selected' : ''}" onclick="loadAdminSection('policies')">
                <span class="chevron">🛡️</span>
                <span>Règles SoD & Zero-Trust</span>
                <span class="count-badge req">4 Politiques</span>
            </div>
        `;
    }

    html += `
        <div class="tree-node ${currentAdminSection === 'settings' ? 'selected' : ''}" onclick="loadAdminSection('settings')">
            <span class="chevron">⚙️</span>
            <span>Paramètres Réglementaires & Redevances</span>
            <span class="count-badge ok">ALCOA+</span>
        </div>
    `;
    container.innerHTML = html;
}

// Chargement principal de la section administration
async function loadAdminSection(section) {
    const canManageUsers = typeof hasPermission === 'function' && hasPermission('USER_MANAGE');
    if (!canManageUsers) {
        currentAdminSection = 'settings';
    } else {
        currentAdminSection = section || currentAdminSection || 'users';
    }
    renderAdminTree();

    if (window.innerWidth < 768) {
        toggleMobileTree(false);
    }

    const rootEl = document.getElementById("bc-root");
    const secEl = document.getElementById("bc-sec");
    const prodEl = document.getElementById("bc-prod");
    const modEl = document.getElementById("bc-mod");

    if (rootEl) rootEl.innerText = "ADMINISTRATION";
    if (secEl) secEl.innerText = "Habilitations_RBAC";

    if (currentAdminSection === 'settings') {
        if (prodEl) prodEl.innerText = "Paramètres Réglementaires";
        if (modEl) modEl.innerText = "Paramètres & Redevances";
        await renderAdminSettingsView();
        return;
    }

    if (currentAdminSection === 'users') {
        if (prodEl) prodEl.innerText = "Collaborateurs";
        if (modEl) modEl.innerText = "Gestion des Profils";
    } else if (currentAdminSection === 'matrix') {
        if (prodEl) prodEl.innerText = "Matrice des Droits";
        if (modEl) modEl.innerText = "14 Permissions Unitaires";
    } else if (currentAdminSection === 'policies') {
        if (prodEl) prodEl.innerText = "Garde-Fous PSSI";
        if (modEl) modEl.innerText = "SoD & Zero-Trust BPF";
    }

    const tbody = document.getElementById("doc-tbody");
    if (tbody) {
        tbody.innerHTML = `<tr><td colspan="7" style="text-align:center; padding:25px; color:var(--text-muted);">Chargement de la console d'administration sécurisée...</td></tr>`;
    }

    try {
        const [usersRes, rolesRes, permsRes] = await Promise.all([
            fetch('/api/v1/admin/users'),
            fetch('/api/v1/admin/roles'),
            fetch('/api/v1/admin/permissions')
        ]);

        if (usersRes.status === 401 || rolesRes.status === 401 || permsRes.status === 401) {
            handleAuthExpired();
            return;
        }

        if (usersRes.status === 403 || rolesRes.status === 403 || permsRes.status === 403) {
            if (tbody) {
                tbody.innerHTML = `<tr><td colspan="7" style="text-align:center; padding:25px; color:var(--accent-red); font-weight:600;">🚫 Accès Refusé : Habilitation USER_MANAGE requise par la PSSI.</td></tr>`;
            }
            return;
        }

        if (usersRes.ok) {
            const data = await usersRes.json();
            adminUsers = data.users || [];
        }
        if (rolesRes.ok) {
            const data = await rolesRes.json();
            adminRoles = data.roles || [];
        }
        if (permsRes.ok) {
            const data = await permsRes.json();
            adminPermissions = data.permissions || [];
        }

        renderAdminTree();

        if (currentAdminSection === 'users') {
            renderAdminUsersTable();
        } else if (currentAdminSection === 'matrix') {
            renderAdminMatrixTable();
        } else if (currentAdminSection === 'policies') {
            renderAdminPoliciesView();
        } else if (currentAdminSection === 'settings') {
            await renderAdminSettingsView();
        }
    } catch (e) {
        console.error("Erreur chargement admin:", e);
        if (tbody) {
            tbody.innerHTML = `<tr><td colspan="7" style="text-align:center; padding:25px; color:var(--accent-red);">Erreur de communication : ${e.message}</td></tr>`;
        }
    }
}

function escapeAdminHtml(str) {
    if (!str) return '';
    return String(str)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
}

// 1. Tableau des Collaborateurs du Laboratoire
function renderAdminUsersTable() {
    const thead = document.getElementById("table-head");
    const tbody = document.getElementById("doc-tbody");
    if (!thead || !tbody) return;
    thead.classList.remove("two-tier");

    thead.innerHTML = `
        <tr>
            <th style="width: 100px;">Statut</th>
            <th>Collaborateur</th>
            <th>Email</th>
            <th>Rôle Métier Attribué</th>
            <th>Droits Effectifs</th>
            <th style="width: 90px;">2FA BPF</th>
            <th style="width: 220px; text-align:center;">Actions Administration</th>
        </tr>
    `;

    tbody.innerHTML = "";

    if (adminUsers.length === 0) {
        tbody.innerHTML = `<tr><td colspan="7" style="text-align:center; padding:20px; color:var(--text-dim);">Aucun collaborateur trouvé pour ce laboratoire.</td></tr>`;
        return;
    }

    adminUsers.forEach((u, idx) => {
        const tr = document.createElement("tr");
        if (idx === selectedAdminUserIndex) tr.classList.add("selected");
        
        tr.onclick = (e) => {
            // Ne pas déclencher si on clique directement sur un select, bouton ou badge interactif
            if (e.target.tagName === 'SELECT' || e.target.tagName === 'BUTTON' || e.target.closest('button') || e.target.closest('.status-tag-interactive')) return;
            selectedAdminUserIndex = idx;
            document.querySelectorAll("#doc-tbody tr").forEach(r => r.classList.remove("selected"));
            tr.classList.add("selected");
            renderAdminUserPreview(u);
            if (window.innerWidth < 992) togglePreviewPane(true);
        };

        const permsCount = (u.permissions || []).length;
        const roleOptions = adminRoles.map(r => `
            <option value="${r.id}" ${r.id === u.role ? 'selected' : ''}>
                ${r.name_fr}
            </option>
        `).join('');

        const isCurrentActiveUser = currentUser && currentUser.id === u.id;
        const isPDT = u.role === 'directeur_technique';
        const isZeroTrust = u.role === 'admin_systeme_local' || u.role === 'super_admin_editeur';

        const fullNameEsc = escapeAdminHtml(`${u.first_name} ${u.last_name}`);
        const emailEsc = escapeAdminHtml(u.email);

        let badgeStatus = u.is_active 
            ? `<span class="status-tag status-ok status-tag-interactive" style="cursor:pointer;" onclick="openUserStatusModal('${u.id}', false, '${fullNameEsc}', '${emailEsc}')" title="Compte Actif. Cliquer pour suspendre / désactiver">ACTIF 🟢</span>`
            : `<span class="status-tag status-tag-interactive" style="background:#ef4444; color:white; cursor:pointer;" onclick="openUserStatusModal('${u.id}', true, '${fullNameEsc}', '${emailEsc}')" title="Compte Suspendu. Cliquer pour réactiver">SUSPENDU 🔴</span>`;

        if (isCurrentActiveUser) {
            badgeStatus = `<span class="status-tag" style="background:#0284c7; color:white; font-weight:700;">VOUS ${u.is_active ? '🟢' : '🔴'}</span>`;
        }

        tr.innerHTML = `
            <td>${badgeStatus}</td>
            <td style="font-weight:600; color:var(--text-main);">
                ${u.first_name} ${u.last_name}
                ${isPDT ? '<span title="Pharmacien Directeur Technique titulaire de l\'agrément nominatif (1.A.7)" style="margin-left:4px; font-size:12px;">💊</span>' : ''}
                ${isZeroTrust ? '<span title="Profil Technique soumis au Zero-Trust strict" style="margin-left:4px; font-size:12px;">🛡️</span>' : ''}
                ${!u.is_active ? '<span title="Accès suspendu" style="margin-left:4px; font-size:11px; color:#ef4444;">(Désactivé)</span>' : ''}
            </td>
            <td style="font-family:monospace; font-size:11px; color:var(--text-muted);">${u.email}</td>
            <td>
                <select class="btn-pill" style="font-size:11px; padding:3px 6px; background:var(--bg-surface); color:var(--text-main); border:1px solid var(--border-color); cursor:pointer;"
                        onchange="handleUserRoleChange('${u.id}', this.value, '${u.email}')">
                    ${roleOptions}
                </select>
            </td>
            <td>
                <span class="status-tag" style="background:var(--badge-bg); color:var(--accent-blue); font-weight:600; cursor:pointer;" 
                      onclick="openUserPermissionsModal('${u.id}')"
                      title="Cliquer pour gérer les dérogations de permissions : ${(u.permissions || []).join(', ')}">
                    ${permsCount}/14 droits ${(u.overrides && u.overrides.length > 0) ? `<strong style="color:var(--accent-amber);">(${u.overrides.length} dérog.)</strong>` : ''} (v${u.permissions_version || 1})
                </span>
            </td>
            <td>
                ${u.totp_enabled 
                    ? '<span style="color:var(--accent-emerald); font-weight:600; font-size:11px;">✓ TOTP Actif</span>' 
                    : '<span style="color:var(--text-dim); font-size:11px;">Non activé</span>'}
            </td>
            <td style="text-align:center; white-space:nowrap;">
                <div style="display:inline-flex; gap:3px;">
                    <button class="btn-pill" style="font-size:10px; padding:2px 6px; border-color:var(--accent-blue); color:var(--accent-blue);" 
                            onclick="openUserPermissionsModal('${u.id}')" 
                            title="Gérer les dérogations unitaires (Table user_permissions)">
                        🔐 Droits
                    </button>
                    <button class="btn-pill" style="font-size:10px; padding:2px 6px; border-color:var(--accent-amber); color:var(--accent-amber);" 
                            onclick="openResetPasswordModal('${u.id}', '${emailEsc}', '${fullNameEsc}')" 
                            title="Réinitialiser le mot de passe (21 CFR Part 11)">
                        🔑 Pwd
                    </button>
                    ${u.totp_enabled ? `
                    <button class="btn-pill" style="font-size:10px; padding:2px 6px; border-color:var(--accent-emerald); color:var(--accent-emerald);" 
                            onclick="handleResetTOTP('${u.id}', '${emailEsc}')" 
                            title="Réinitialiser le 2FA TOTP (Perte de terminal)">
                        📱 2FA
                    </button>` : ''}
                    <button class="btn-pill" style="font-size:10px; padding:2px 6px; border-color:#64748b; color:#64748b;" 
                            onclick="devSwitchRole('${u.email}')" 
                            title="Basculer instantanément sur ce compte pour vérifier les masquages">
                        🧪 Test
                    </button>
                </div>
            </td>
        `;

        tbody.appendChild(tr);
    });

    // Aperçu par défaut du premier collaborateur
    if (adminUsers.length > 0 && adminUsers[selectedAdminUserIndex]) {
        renderAdminUserPreview(adminUsers[selectedAdminUserIndex]);
    }
}

// Aperçu détaillé d'un collaborateur dans le volet latéral (Pane 3)
function renderAdminUserPreview(user) {
    const preview = document.getElementById("preview-body");
    if (!preview || !user) return;

    const tabsContainer = document.getElementById("preview-tabs-container");
    if (tabsContainer) {
        tabsContainer.innerHTML = `<span style="font-weight:700; font-size:12px; color:var(--accent-blue); display:flex; align-items:center; gap:6px;">👤 Fiche Collaborateur BPF</span>`;
    }

    const roleObj = adminRoles.find(r => r.id === user.role);
    const roleLabel = roleObj ? roleObj.name_fr : user.role;
    const isPDT = user.role === 'directeur_technique';
    const isZeroTrust = user.role === 'admin_systeme_local' || user.role === 'super_admin_editeur';
    const userPerms = new Set(user.permissions || []);
    const initials = (((user.first_name || 'U').charAt(0)) + ((user.last_name || 'L').charAt(0))).toUpperCase();

    const categories = [
        { name: "🏢 Référentiel MIPH", perms: ["REF_VIEW", "REF_EDIT"] },
        { name: "📋 Dossiers CTD & Soumission", perms: ["DOSSIER_VIEW", "DOSSIER_EDIT", "DOSSIER_SUBMIT", "VARIATION_MANAGE"] },
        { name: "📄 Coffre-Fort GED & Visa", perms: ["DOC_READ", "DOC_WRITE", "DOC_REVIEW", "DOC_APPROVE_PDT", "EXPORT_LEGAL"] },
        { name: "🔒 Sécurité & Administration", perms: ["AUDIT_VIEW", "USER_MANAGE", "SYS_MAINT"] }
    ];

    let permsBreakdownHtml = '';
    categories.forEach(cat => {
        permsBreakdownHtml += `
            <div style="margin-bottom:6px;">
                <div style="font-size:10px; font-weight:700; color:var(--text-muted); text-transform:uppercase; margin-bottom:3px; letter-spacing:0.3px;">${cat.name}</div>
                <div style="display:flex; flex-direction:column; gap:2px;">
                    ${cat.perms.map(p => {
                        const has = userPerms.has(p);
                        const meta = RBAC_PERM_METADATA[p] || { icon: "🔹", label: p, fullTitle: p };
                        const isSpecialPDT = p === 'DOC_APPROVE_PDT';
                        const isDocReadBlocked = p === 'DOC_READ' && isZeroTrust;

                        let badge = has 
                            ? `<span style="color:var(--accent-emerald); font-weight:700; font-size:10px;">✓ Accordé</span>`
                            : `<span style="color:var(--text-dim); font-size:10px;">—</span>`;

                        if (isSpecialPDT && has) {
                            badge = `<span style="color:#d97706; font-weight:700; font-size:10px;">✍️ Visa 1.A.7</span>`;
                        } else if (isDocReadBlocked) {
                            badge = `<span style="color:var(--accent-red); font-weight:700; font-size:10px;">🚫 Zero-Trust</span>`;
                        }

                        return `
                            <div style="display:flex; justify-content:space-between; align-items:center; background:var(--bg-panel); padding:3px 6px; border-radius:3px; font-size:11px;">
                                <div style="display:flex; align-items:center; gap:5px; min-width:0; overflow:hidden;">
                                    <span style="font-size:12px; flex-shrink:0;">${meta.icon}</span>
                                    <span style="font-weight:600; color:${has ? 'var(--text-main)' : 'var(--text-dim)'}; white-space:nowrap; overflow:hidden; text-overflow:ellipsis;">${meta.label}</span>
                                    <span style="font-family:monospace; font-size:8px; color:var(--text-dim); opacity:0.6; flex-shrink:0;">${p}</span>
                                </div>
                                <div style="flex-shrink:0; margin-left:6px;">${badge}</div>
                            </div>
                        `;
                    }).join('')}
                </div>
            </div>
        `;
    });

    preview.innerHTML = `
        <!-- Card 1 : Identité & Actions (hauteur minimale requise par son contenu) -->
        <div class="admin-preview-card">
            <div class="admin-preview-header">
                <div class="admin-user-avatar">${initials}</div>
                <div style="min-width:0; flex:1;">
                    <div style="font-weight:700; font-size:13px; color:var(--text-main); white-space:nowrap; overflow:hidden; text-overflow:ellipsis;">
                        ${escapeAdminHtml(user.first_name)} ${escapeAdminHtml(user.last_name)}
                    </div>
                    <div style="font-family:monospace; font-size:10px; color:var(--text-dim); white-space:nowrap; overflow:hidden; text-overflow:ellipsis;">
                        ${escapeAdminHtml(user.email)}
                    </div>
                </div>
            </div>

            <!-- Badges statut -->
            <div style="display:flex; flex-wrap:wrap; gap:4px; align-items:center;">
                <span class="status-tag ${user.is_active ? 'status-ok' : 'status-err'}" style="font-size:9.5px; padding:2px 6px;">
                    ${user.is_active ? '🟢 Actif' : '🔴 Suspendu'}
                </span>
                <span class="status-tag ${user.totp_enabled ? 'status-ok' : 'status-req'}" style="font-size:9.5px; padding:2px 6px;">
                    ${user.totp_enabled ? '📱 2FA BPF' : '📱 2FA Non Configuré'}
                </span>
                <span class="status-tag" style="background:var(--bg-panel); color:var(--text-muted); font-size:9.5px; padding:2px 6px;">
                    v${user.permissions_version || 1}
                </span>
            </div>

            <!-- Rôle BPF -->
            <div style="background:var(--bg-panel); border:1px solid var(--border-color); border-radius:5px; padding:6px 8px;">
                <div style="font-size:9px; font-weight:700; color:var(--text-dim); text-transform:uppercase;">Rôle Métier BPF</div>
                <div style="font-size:11.5px; font-weight:700; color:var(--accent-blue); margin-top:1px;">${escapeAdminHtml(roleLabel)}</div>
                ${user.contract_expires_at ? `<div style="font-size:10px; color:var(--accent-amber); margin-top:2px;">⏳ Échéance : <strong>${new Date(user.contract_expires_at).toLocaleDateString()}</strong></div>` : ''}
            </div>

            <!-- Boutons d'Action (pleine largeur avec marges nettes) -->
            <div style="display:flex; gap:6px; margin-top:2px;">
                <button class="btn-pill" style="flex:1; background:var(--accent-blue); color:white; border:none; padding:6px 8px; font-size:11px; font-weight:700; text-align:center;" onclick="openUserPermissionsModal('${user.id}')">
                    🔐 Dérogations
                </button>
                <button class="btn-pill" style="flex:1; background:var(--bg-surface); border:1px solid var(--border-color); color:var(--text-main); padding:6px 8px; font-size:11px; font-weight:600; text-align:center;" onclick="openResetPasswordModal('${user.id}', '${escapeAdminHtml(user.email)}', '${escapeAdminHtml(user.first_name + ' ' + user.last_name)}')">
                    🔑 Mot de Passe
                </button>
            </div>
        </div>

        <!-- Card 2 (si applicable) : Mises en garde réglementaires (hauteur minimale requise) -->
        ${isPDT ? `
            <div class="admin-preview-card" style="background:rgba(217, 119, 6, 0.08); border-color:#d97706; padding:8px 10px;">
                <div style="font-size:10.5px; color:var(--text-main); line-height:1.4;">
                    <strong style="color:#d97706;">✍️ Titulaire du Visa Légal (1.A.7) :</strong> Seul compte habilité à signer et sceller les dossiers d'AMM opposables à l'ANPP (Annexe 11 BPF).
                </div>
            </div>
        ` : ''}

        ${isZeroTrust ? `
            <div class="admin-preview-card" style="background:rgba(239, 68, 68, 0.08); border-color:#ef4444; padding:8px 10px;">
                <div style="font-size:10.5px; color:var(--text-main); line-height:1.4;">
                    <strong style="color:var(--accent-red);">🛡️ Isolation Zero-Trust :</strong> Accès interdit aux formules secrètes et dossiers cliniques (Blocage 403 <code>ERR_ZERO_TRUST_DENIED</code>).
                </div>
            </div>
        ` : ''}

        ${user.role === 'consultant_externe' ? `
            <div class="admin-preview-card" style="background:rgba(2, 132, 199, 0.08); border-color:#0284c7; padding:8px 10px;">
                <div style="font-size:10.5px; color:var(--text-main); line-height:1.4;">
                    <strong style="color:var(--accent-blue);">🤝 Accès Tiers Cloisonné :</strong> Confiné au seul dossier assigné, sans accès transverse ni visibilité des prix.
                </div>
            </div>
        ` : ''}

        <!-- Card 3 : Détail des Habilitations Classées par Domaine (hauteur minimale requise) -->
        <div class="admin-preview-card">
            <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:4px; border-bottom:1px solid var(--border-color); padding-bottom:4px;">
                <span style="font-size:11px; font-weight:700; color:var(--text-main);">Habilitations Effectives</span>
                <span style="font-size:10px; color:var(--accent-blue); font-weight:700;">${(user.permissions || []).length} / 14</span>
            </div>
            ${permsBreakdownHtml}
        </div>
    `;
}

// 2. Tableau de la Matrice des 14 Permissions x Rôles
function renderAdminMatrixTable() {
    const thead = document.getElementById("table-head");
    const tbody = document.getElementById("doc-tbody");
    if (!thead || !tbody) return;
    thead.classList.add("two-tier");

    const permsList = RBAC_PERMS_LIST;

    thead.innerHTML = `
        <tr style="border-bottom: 1px solid var(--border-color);">
            <th rowspan="2" style="min-width: 200px; vertical-align: middle; padding: 8px 10px; background: var(--bg-th); border-right: 2px solid var(--border-color);">
                <div style="font-size: 11px; font-weight: 700; color: var(--text-main); display:flex; align-items:center; gap:5px;">
                    <span>🛡️</span> Rôle Réglementaire BPF
                </div>
                <div style="font-size: 10px; color: var(--text-dim); font-weight: normal; margin-top:2px;">10 profils pharmaceutiques habilités</div>
            </th>
            <th colspan="2" style="text-align: center; font-size: 10px; font-weight: 700; background: rgba(3, 105, 161, 0.12); color: #0284c7; padding: 6px 4px; border-right: 2px solid var(--border-color); text-transform: uppercase; letter-spacing: 0.5px;">
                🏢 Référentiel MIPH
            </th>
            <th colspan="4" style="text-align: center; font-size: 10px; font-weight: 700; background: rgba(16, 185, 129, 0.12); color: #059669; padding: 6px 4px; border-right: 2px solid var(--border-color); text-transform: uppercase; letter-spacing: 0.5px;">
                📋 Dossiers CTD & Soumission ANPP
            </th>
            <th colspan="5" style="text-align: center; font-size: 10px; font-weight: 700; background: rgba(217, 119, 6, 0.12); color: #d97706; padding: 6px 4px; border-right: 2px solid var(--border-color); text-transform: uppercase; letter-spacing: 0.5px;">
                📄 Coffre-Fort GED & Approbation BPF
            </th>
            <th colspan="3" style="text-align: center; font-size: 10px; font-weight: 700; background: rgba(99, 102, 241, 0.12); color: #6366f1; padding: 6px 4px; text-transform: uppercase; letter-spacing: 0.5px;">
                🔒 PSSI, Sécurité & Système
            </th>
        </tr>
        <tr>
            ${permsList.map(p => `
                <th style="text-align: center; padding: 6px 4px; font-size: 10px; min-width: 72px; vertical-align: bottom; ${p.code === 'DOC_APPROVE_PDT' ? 'background: rgba(217,119,6,0.18);' : ''} ${p.isLast ? 'border-right: 2px solid var(--border-color);' : ''}" title="${p.fullTitle} (${p.code}) : ${p.desc}">
                    <div style="font-size: 13px; line-height: 1; margin-bottom: 2px;">${p.icon}</div>
                    <div style="font-weight: 700; color: ${p.code === 'DOC_APPROVE_PDT' ? '#d97706' : 'var(--text-main)'}; line-height: 1.2; white-space: nowrap;">${p.label}</div>
                    <div style="font-family: monospace; font-size: 8px; color: var(--text-dim); margin-top: 2px; letter-spacing: -0.3px; opacity: 0.8;">${p.code}</div>
                </th>
            `).join('')}
        </tr>
    `;

    tbody.innerHTML = "";

    adminRoles.forEach((role, idx) => {
        const tr = document.createElement("tr");
        const rolePerms = new Set(role.permissions || []);
        const isPDT = role.id === 'directeur_technique';
        const isZeroTrust = role.id === 'admin_systeme_local' || role.id === 'super_admin_editeur';
        const isConsultantRole = role.id === 'consultant_externe';

        let cellsHtml = '';
        permsList.forEach(p => {
            const has = rolePerms.has(p.code);
            const borderSep = p.isLast ? 'border-right: 2px solid var(--border-color);' : '';

            if (p.code === 'DOC_APPROVE_PDT') {
                if (isPDT) {
                    cellsHtml += `<td style="text-align:center; background:rgba(217,119,6,0.1); ${borderSep}"><span style="color:#d97706; font-weight:700;" title="Exclusivité Nominative 1.A.7">✓ DT</span></td>`;
                } else {
                    cellsHtml += `<td style="text-align:center; color:var(--text-dim); opacity:0.3; ${borderSep}">—</td>`;
                }
            } else if (p.code === 'DOC_READ' && isZeroTrust) {
                cellsHtml += `<td style="text-align:center; background:rgba(239,68,68,0.1); ${borderSep}"><span style="color:var(--accent-red); font-weight:700; font-size:9px;" title="Bloqué par Zero-Trust">🚫 ZT</span></td>`;
            } else if (isConsultantRole && p.code === 'DOSSIER_VIEW') {
                cellsHtml += `<td style="text-align:center; background:rgba(2,132,199,0.1); ${borderSep}"><span style="color:var(--accent-blue); font-weight:700; font-size:9px;" title="Cloisonné au seul dossier assigné">✓ Dossier</span></td>`;
            } else {
                if (has) {
                    cellsHtml += `<td style="text-align:center; ${borderSep}"><span style="color:var(--accent-emerald); font-weight:700;">✓</span></td>`;
                } else {
                    cellsHtml += `<td style="text-align:center; color:var(--text-dim); opacity:0.3; ${borderSep}">—</td>`;
                }
            }
        });

        tr.innerHTML = `
            <td style="font-weight:600; color:var(--text-main); border-right: 2px solid var(--border-color);">
                ${role.name_fr}
                <div style="font-size:10px; color:var(--text-dim); font-weight:normal; font-family:monospace;">${role.id}</div>
            </td>
            ${cellsHtml}
        `;

        tr.onclick = () => {
            renderRolePreview(role, permsList);
            if (window.innerWidth < 992) togglePreviewPane(true);
        };

        tbody.appendChild(tr);
    });

    if (adminRoles.length > 0) {
        renderRolePreview(adminRoles[0], permsList);
    }
}

// Aperçu d'un rôle dans le volet latéral (cards avec hauteur minimale naturelle)
function renderRolePreview(role, permsList) {
    const preview = document.getElementById("preview-body");
    if (!preview || !role) return;

    const tabsContainer = document.getElementById("preview-tabs-container");
    if (tabsContainer) {
        tabsContainer.innerHTML = `<span style="font-weight:700; font-size:12px; color:#d97706; display:flex; align-items:center; gap:6px;">🔐 Fiche Profil de Rôle BPF</span>`;
    }

    const rolePerms = new Set(role.permissions || []);
    const isPDT = role.id === 'directeur_technique';
    const isZeroTrust = role.id === 'admin_systeme_local' || role.id === 'super_admin_editeur';
    const permCount = (role.permissions || []).length;
    const permPct = Math.round((permCount / 14) * 100);

    const categories = [
        { name: "🏢 Référentiel MIPH", perms: ["REF_VIEW", "REF_EDIT"] },
        { name: "📋 Dossiers CTD & Soumission", perms: ["DOSSIER_VIEW", "DOSSIER_EDIT", "DOSSIER_SUBMIT", "VARIATION_MANAGE"] },
        { name: "📄 Coffre-Fort GED & Visa", perms: ["DOC_READ", "DOC_WRITE", "DOC_REVIEW", "DOC_APPROVE_PDT", "EXPORT_LEGAL"] },
        { name: "🔒 Sécurité & Administration", perms: ["AUDIT_VIEW", "USER_MANAGE", "SYS_MAINT"] }
    ];

    let permsBreakdownHtml = '';
    categories.forEach(cat => {
        permsBreakdownHtml += `
            <div style="margin-bottom:6px;">
                <div style="font-size:10px; font-weight:700; color:var(--text-muted); text-transform:uppercase; margin-bottom:3px; letter-spacing:0.3px;">${cat.name}</div>
                <div style="display:flex; flex-direction:column; gap:2px;">
                    ${cat.perms.map(p => {
                        const has = rolePerms.has(p);
                        const meta = RBAC_PERM_METADATA[p] || { icon: "🔹", label: p, fullTitle: p };
                        const isSpecialPDT = p === 'DOC_APPROVE_PDT';
                        const isDocReadBlocked = p === 'DOC_READ' && isZeroTrust;

                        let badge = has 
                            ? `<span style="color:var(--accent-emerald); font-weight:700; font-size:10px;">✓ Accordée</span>`
                            : `<span style="color:var(--text-dim); font-size:10px;">—</span>`;

                        if (isSpecialPDT && has) {
                            badge = `<span style="color:#d97706; font-weight:700; font-size:10px;">✍️ Visa 1.A.7</span>`;
                        } else if (isDocReadBlocked) {
                            badge = `<span style="color:var(--accent-red); font-weight:700; font-size:10px;">🚫 Zero-Trust</span>`;
                        }

                        return `
                            <div style="display:flex; justify-content:space-between; align-items:center; background:var(--bg-panel); padding:3px 6px; border-radius:3px; font-size:11px;">
                                <div style="display:flex; align-items:center; gap:5px; min-width:0; overflow:hidden;">
                                    <span style="font-size:12px; flex-shrink:0;">${meta.icon}</span>
                                    <span style="font-weight:600; color:${has ? 'var(--text-main)' : 'var(--text-dim)'}; white-space:nowrap; overflow:hidden; text-overflow:ellipsis;">${meta.label}</span>
                                    <span style="font-family:monospace; font-size:8px; color:var(--text-dim); opacity:0.6; flex-shrink:0;">${p}</span>
                                </div>
                                <div style="flex-shrink:0; margin-left:6px;">${badge}</div>
                            </div>
                        `;
                    }).join('')}
                </div>
            </div>
        `;
    });

    preview.innerHTML = `
        <!-- Card 1 : Définition du Rôle (hauteur minimale requise par son contenu) -->
        <div class="admin-preview-card">
            <div>
                <div style="font-size:9.5px; font-weight:700; color:var(--text-dim); text-transform:uppercase;">Profil Réglementaire BPF</div>
                <div style="font-size:13.5px; font-weight:800; color:${isPDT ? '#d97706' : 'var(--accent-blue)'}; margin-top:2px;">
                    ${role.name_fr}
                </div>
                ${role.name_ar ? `
                    <div dir="rtl" style="font-size:12px; font-weight:600; color:var(--text-muted); margin-top:2px; font-family:'Segoe UI', Tahoma, sans-serif;">
                        ${role.name_ar}
                    </div>
                ` : ''}
                <div style="font-family:monospace; font-size:10px; color:var(--text-dim); margin-top:2px;">
                    <code>${role.id}</code>
                </div>
            </div>

            <!-- Jauge d'habilitations -->
            <div style="background:var(--bg-panel); border:1px solid var(--border-color); border-radius:5px; padding:6px 8px;">
                <div style="display:flex; justify-content:space-between; font-size:10.5px; margin-bottom:3px;">
                    <span style="color:var(--text-muted);">Périmètre accordé :</span>
                    <strong style="color:var(--text-main);">${permCount} / 14 droits (${permPct}%)</strong>
                </div>
                <div style="height:5px; background:var(--border-color); border-radius:3px; overflow:hidden;">
                    <div style="width:${permPct}%; height:100%; background:${isPDT ? '#d97706' : 'var(--accent-blue)'}; border-radius:3px;"></div>
                </div>
            </div>

            <!-- Description Métier -->
            <div style="font-size:11px; color:var(--text-main); line-height:1.4; background:var(--bg-panel); padding:6px 8px; border-radius:5px;">
                ${role.description || 'Rôle standard du référentiel pharmaceutique.'}
            </div>
        </div>

        <!-- Card 2 (si applicable) : Mises en garde spécifiques (hauteur minimale requise) -->
        ${isPDT ? `
            <div class="admin-preview-card" style="background:rgba(217, 119, 6, 0.08); border-color:#d97706; padding:8px 10px;">
                <div style="font-size:10.5px; color:var(--text-main); line-height:1.4;">
                    <strong style="color:#d97706;">✍️ Exclusivité Nominative 1.A.7 :</strong> Ce rôle est le seul à posséder la permission <code>DOC_APPROVE_PDT</code>.
                </div>
            </div>
        ` : ''}

        ${isZeroTrust ? `
            <div class="admin-preview-card" style="background:rgba(239, 68, 68, 0.08); border-color:#ef4444; padding:8px 10px;">
                <div style="font-size:10.5px; color:var(--text-main); line-height:1.4;">
                    <strong style="color:var(--accent-red);">🛡️ Zero-Trust Actif :</strong> Ce rôle technique a la permission <code>DOC_READ</code> strictement verrouillée et bloquée.
                </div>
            </div>
        ` : ''}

        <!-- Card 3 : Matrice des Droits par Défaut (hauteur minimale requise) -->
        <div class="admin-preview-card">
            <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:4px; border-bottom:1px solid var(--border-color); padding-bottom:4px;">
                <span style="font-size:11px; font-weight:700; color:var(--text-main);">Droits Attribués au Rôle</span>
                <span style="font-size:10px; color:${isPDT ? '#d97706' : 'var(--accent-blue)'}; font-weight:700;">${permCount} / 14</span>
            </div>
            ${permsBreakdownHtml}
        </div>
    `;
}

// 3. Vue des Politiques de Sécurité et Garde-Fous PSSI & BPF
function renderAdminPoliciesView() {
    const thead = document.getElementById("table-head");
    const tbody = document.getElementById("doc-tbody");
    if (!thead || !tbody) return;
    thead.classList.remove("two-tier");

    thead.innerHTML = `
        <tr>
            <th colspan="7" style="background:var(--bg-th); padding:10px; font-size:12px; font-weight:700; color:var(--accent-blue);">
                🛡️ Politiques d'Intégrité BPF & Contrôles d'Accès Qualifiés (Norme 21 CFR Part 11)
            </th>
        </tr>
    `;

    tbody.innerHTML = `
        <tr>
            <td colspan="7" style="padding:20px;">
                <div style="display:grid; grid-template-columns: repeat(auto-fit, minmax(280px, 1fr)); gap:15px;">
                    
                    <!-- Carte 1 : SoD Maker != Approver -->
                    <div style="background:var(--bg-surface); border:1px solid var(--border-color); border-radius:6px; padding:15px; display:flex; flex-direction:column; gap:8px;">
                        <div style="display:flex; align-items:center; gap:8px;">
                            <span style="font-size:20px;">⚖️</span>
                            <span style="font-weight:700; font-size:13px; color:var(--text-main);">1. Séparation des Tâches (SoD)</span>
                        </div>
                        <div style="font-size:11px; color:var(--text-dim); line-height:1.5;">
                            Règle BPF formelle : <strong>Maker ≠ Approver</strong>.<br>
                            Le Pharmacien Directeur Technique ne peut en aucun cas approuver électroniquement (Visa PDT) une version d'un document dont il est lui-même l'auteur initial (<code>author_id == pdt_id</code>).
                        </div>
                        <div style="margin-top:auto; font-size:10px; background:rgba(239,68,68,0.1); color:var(--accent-red); padding:4px 8px; border-radius:4px; font-family:monospace;">
                            Blocage 403 ERR_SOD_VIOLATION scellé dans l'audit trail
                        </div>
                    </div>

                    <!-- Carte 2 : Exclusivité Visa PDT 1.A.7 -->
                    <div style="background:var(--bg-surface); border:1px solid var(--border-color); border-radius:6px; padding:15px; display:flex; flex-direction:column; gap:8px;">
                        <div style="display:flex; align-items:center; gap:8px;">
                            <span style="font-size:20px;">💊</span>
                            <span style="font-weight:700; font-size:13px; color:var(--text-main);">2. Exclusivité Visa PDT (1.A.7)</span>
                        </div>
                        <div style="font-size:11px; color:var(--text-dim); line-height:1.5;">
                            Seul le Pharmacien Directeur Technique titulaire de la décision nominative du Ministère (MIPH) possède l'habilitation <code>DOC_APPROVE_PDT</code>.
                            Aucun administrateur ou délégué ne peut signer à sa place.
                        </div>
                        <div style="margin-top:auto; font-size:10px; background:rgba(217,119,6,0.1); color:#d97706; padding:4px 8px; border-radius:4px; font-family:monospace;">
                            Arrêté MIPH Décembre 2020 &amp; Décret 21-82
                        </div>
                    </div>

                    <!-- Carte 3 : Zero-Trust IT Isolation -->
                    <div style="background:var(--bg-surface); border:1px solid var(--border-color); border-radius:6px; padding:15px; display:flex; flex-direction:column; gap:8px;">
                        <div style="display:flex; align-items:center; gap:8px;">
                            <span style="font-size:20px;">🛡️</span>
                            <span style="font-weight:700; font-size:13px; color:var(--text-main);">3. Isolation Zero-Trust Cloud &amp; IT</span>
                        </div>
                        <div style="font-size:11px; color:var(--text-dim); line-height:1.5;">
                            Les profils d'administration technique (<code>admin_systeme_local</code>, <code>super_admin_editeur</code>) assurent la maintenance (<code>SYS_MAINT</code>) mais sont <strong>strictement interdits</strong> de téléchargement des formules ou pièces pharmaceutiques (<code>DOC_READ</code> bloqué).
                        </div>
                        <div style="margin-top:auto; font-size:10px; background:rgba(2,132,199,0.1); color:var(--accent-blue); padding:4px 8px; border-radius:4px; font-family:monospace;">
                            Protection Secret Pharmaceutique &amp; PSSI
                        </div>
                    </div>

                    <!-- Carte 4 : Audit Trail Inviolable -->
                    <div style="background:var(--bg-surface); border:1px solid var(--border-color); border-radius:6px; padding:15px; display:flex; flex-direction:column; gap:8px;">
                        <div style="display:flex; align-items:center; gap:8px;">
                            <span style="font-size:20px;">⛓️</span>
                            <span style="font-weight:700; font-size:13px; color:var(--text-main);">4. Audit Trail Scellé (21 CFR Part 11)</span>
                        </div>
                        <div style="font-size:11px; color:var(--text-dim); line-height:1.5;">
                            Toute modification d'habilitation, attribution de rôle ou signature est consignée de manière inaltérable dans le registre d'audit ALCOA+ scellé par empreinte cryptographique interdisant toute modification ou suppression.
                        </div>
                        <div style="margin-top:auto; font-size:10px; background:rgba(16,185,129,0.1); color:var(--accent-emerald); padding:4px 8px; border-radius:4px; font-family:monospace;">
                            Horodatage UTC exact + Empreinte SHA-256
                        </div>
                    </div>

                </div>
            </td>
        </tr>
    `;

    const preview = document.getElementById("preview-body");
    if (preview) {
        preview.innerHTML = `
            <div class="pdf-sim-page">
                <div class="pdf-sim-header">
                    RÉPUBLIQUE ALGÉRIENNE DÉMOCRATIQUE ET POPULAIRE<br>
                    RÈGLES DE CONFORMITÉ &amp; POLITIQUES BPF / PSSI
                </div>
                <div style="font-size:12px; font-weight:700; color:#0369a1; text-align:center; margin-bottom:10px;">
                    SYNTHÈSE DU CONTRÔLE D'INTÉGRITÉ ALCOA+
                </div>
                <div style="font-size:11px; line-height:1.6;">
                    Le système SANAD PHARMA implémente nativement les 5 principes d'intégrité des données requises par l'ANPP et la pharmacopée :<br><br>
                    <strong>A - Attribuable :</strong> Chaque action est rattachée à un utilisateur unique avec ré-authentification obligatoire.<br>
                    <strong>L - Lisible :</strong> Traçabilité claire en français et en arabe avec horodatage UTC.<br>
                    <strong>C - Contemporain :</strong> Enregistrement instantané sans décalage temporel.<br>
                    <strong>O - Original :</strong> Conservation de la version source scellée par son hash SHA-256 CAS.<br>
                    <strong>A - Exact (Accurate) :</strong> Contrôles d'accès RBAC et barrières de séparation des pouvoirs (SoD).
                </div>
            </div>
        `;
    }
}

// Mise à jour du rôle d'un collaborateur
async function handleUserRoleChange(userId, newRole, email) {
    if (!confirm(`Confirmez-vous la modification du rôle pour ${email} vers "${formatRole(newRole)}" ?\nCette opération incrémentera la version des permissions et sera inscrite dans l'Audit Trail.`)) {
        loadAdminSection('users');
        return;
    }

    try {
        const res = await fetch(`/api/v1/admin/users/${userId}/role`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ role: newRole })
        });

        const data = await res.json();
        if (res.ok) {
            showNotifyModal(
                "Rôle Collaborateur Mis à Jour",
                `<div style="display:flex; flex-direction:column; gap:8px;">
                    <div style="color:var(--accent-emerald); font-weight:700;">
                        ✓ Habilitation actualisée avec succès
                    </div>
                    <div style="background:var(--bg-surface); border:1px solid var(--border-color); border-radius:4px; padding:10px; font-size:11px; display:flex; flex-direction:column; gap:4px;">
                        <div><strong>Collaborateur :</strong> <code>${email}</code></div>
                        <div><strong>Nouveau Rôle :</strong> <span style="font-weight:600; color:var(--accent-blue);">${formatRole(newRole)}</span></div>
                        <div><strong>Statut Audit :</strong> Événement <code>USER_ROLE_UPDATED</code> scellé</div>
                    </div>
                </div>`,
                "👤",
                "Fermer"
            );

            // Recharger la liste des utilisateurs
            loadAdminSection('users');

            // Si c'est l'utilisateur actif actuel, rafraîchir sa session en direct
            if (currentUser && currentUser.id === userId) {
                checkAuthSession();
            }
        } else {
            showNotifyModal("Erreur de Modification", data.error || "Impossible de mettre à jour le rôle", "⚠️", "Fermer");
            loadAdminSection('users');
        }
    } catch (e) {
        console.error("Erreur mise à jour rôle:", e);
        showNotifyModal("Erreur Réseau", e.message, "⚠️", "Fermer");
        loadAdminSection('users');
    }
}

// Bascule d'affichage mot de passe
function togglePasswordVisibility(inputId) {
    const el = document.getElementById(inputId);
    if (!el) return;
    el.type = el.type === 'password' ? 'text' : 'password';
}

// 1. Modal Nouveau Collaborateur
function openNewUserModal() {
    const roleSelect = document.getElementById("new-user-role");
    if (roleSelect) {
        roleSelect.innerHTML = adminRoles.map(r => `
            <option value="${r.id}">${r.name_fr}</option>
        `).join('');
    }
    const form = document.getElementById("form-new-user");
    if (form) form.reset();

    const expContainer = document.getElementById("new-user-contract-container");
    if (expContainer) expContainer.style.display = 'none';

    const modal = document.getElementById("modal-new-user");
    if (modal) modal.classList.add("open");
}

function handleRoleSelectionChange(roleId) {
    const expContainer = document.getElementById("new-user-contract-container");
    const expInput = document.getElementById("new-user-contract-exp");
    if (expContainer) {
        if (roleId === 'consultant_externe') {
            expContainer.style.display = 'block';
            if (expInput) expInput.required = true;
        } else {
            expContainer.style.display = 'none';
            if (expInput) {
                expInput.required = false;
                expInput.value = '';
            }
        }
    }
}

async function handleCreateUserSubmit(event) {
    event.preventDefault();
    const firstName = document.getElementById("new-user-firstname").value.trim();
    const lastName = document.getElementById("new-user-lastname").value.trim();
    const email = document.getElementById("new-user-email").value.trim();
    const role = document.getElementById("new-user-role").value;
    const password = document.getElementById("new-user-password").value;
    const expInput = document.getElementById("new-user-contract-exp");

    const payload = {
        first_name: firstName,
        last_name: lastName,
        email: email,
        role: role,
        password: password
    };

    if (role === 'consultant_externe' && expInput && expInput.value) {
        payload.contract_expires_at = new Date(expInput.value).toISOString();
    }

    try {
        const res = await fetch("/api/v1/admin/users", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify(payload)
        });

        const data = await res.json();
        if (!res.ok) {
            showNotifyModal("Échec Création Collaborateur", data.error || "Impossible de créer l'utilisateur", "⚠️", "Fermer");
            return;
        }

        closeModal("modal-new-user");
        showNotifyModal(
            "Collaborateur Créé avec Succès",
            `<div style="display:flex; flex-direction:column; gap:8px;">
                <div style="color:var(--accent-emerald); font-weight:700;">
                    ✓ Le compte de ${escapeAdminHtml(firstName + ' ' + lastName)} est actif
                </div>
                <div style="font-size:11px; line-height:1.5;">
                    Identifiant : <code>${escapeAdminHtml(email)}</code><br>
                    Rôle BPF : <strong>${escapeAdminHtml(role)}</strong><br>
                    Mot de passe haché avec <strong>Argon2id</strong> (Conformité OWASP / BPF).
                </div>
            </div>`,
            "👤",
            "Terminer"
        );
        loadAdminSection('users');
    } catch (e) {
        console.error("Erreur création utilisateur:", e);
        showNotifyModal("Erreur Réseau", e.message, "⚠️", "Fermer");
    }
}

// 2. Modal Activation / Suspension Compte
function openUserStatusModal(userId, targetIsActive, name, email) {
    const modal = document.getElementById("modal-user-status");
    if (!modal) return;

    document.getElementById("status-user-id").value = userId;
    document.getElementById("status-target-active").value = targetIsActive ? "true" : "false";
    document.getElementById("status-user-name").value = `${name} (${email})`;
    document.getElementById("status-reason").value = "";

    const descEl = document.getElementById("status-action-desc");
    const submitBtn = document.getElementById("btn-submit-status");
    const titleEl = document.getElementById("user-status-title");
    const iconEl = document.getElementById("user-status-icon");

    if (targetIsActive) {
        titleEl.innerText = "Réactiver le Compte Collaborateur (BPF)";
        iconEl.innerText = "🟢";
        descEl.style.background = "rgba(16, 185, 129, 0.1)";
        descEl.style.color = "var(--accent-emerald)";
        descEl.innerHTML = "✓ Réactivation de l'accès au système avec conservation des droits antérieurs.";
        submitBtn.style.background = "var(--accent-emerald)";
        submitBtn.innerText = "Réactiver le Compte";
    } else {
        titleEl.innerText = "Suspendre / Désactiver le Compte (21 CFR Part 11)";
        iconEl.innerText = "🔴";
        descEl.style.background = "rgba(239, 68, 68, 0.1)";
        descEl.style.color = "#ef4444";
        descEl.innerHTML = "⚠️ Suspension de l'accès et révocation immédiate de toutes les sessions. Zéro suppression physique pour préserver les signatures.";
        submitBtn.style.background = "#ef4444";
        submitBtn.innerText = "Suspendre l'Accès";
    }

    modal.classList.add("open");
}

async function handleUserStatusSubmit(event) {
    event.preventDefault();
    const userId = document.getElementById("status-user-id").value;
    const targetIsActive = document.getElementById("status-target-active").value === "true";
    const reason = document.getElementById("status-reason").value.trim();

    try {
        const res = await fetch(`/api/v1/admin/users/${userId}/status`, {
            method: "PUT",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
                is_active: targetIsActive,
                reason: reason
            })
        });

        const data = await res.json();
        if (!res.ok) {
            showNotifyModal("Erreur Statut Utilisateur", data.error || "Impossible de modifier le statut", "⚠️", "Fermer");
            return;
        }

        closeModal("modal-user-status");
        showNotifyModal(
            "Statut Mis à Jour (Audit Trail Scellé)",
            `<div style="font-size:12px;">
                Le compte a été <strong>${targetIsActive ? 'réactivé' : 'suspendu'}</strong> avec succès.<br>
                Motif réglementaire scellé dans l'Audit Trail immuable ALCOA+.
            </div>`,
            targetIsActive ? "🟢" : "🔒",
            "Compris"
        );
        loadAdminSection('users');
    } catch (e) {
        console.error("Erreur statut utilisateur:", e);
        showNotifyModal("Erreur Réseau", e.message, "⚠️", "Fermer");
    }
}

// 3. Modal Réinitialisation Mot de Passe
function openResetPasswordModal(userId, email, name) {
    const modal = document.getElementById("modal-reset-user-password");
    if (!modal) return;

    document.getElementById("reset-user-id").value = userId;
    document.getElementById("reset-user-name").value = `${name} (${email})`;
    document.getElementById("reset-new-password").value = "";
    document.getElementById("reset-reason").value = "";

    modal.classList.add("open");
}

async function handleResetPasswordSubmit(event) {
    event.preventDefault();
    const userId = document.getElementById("reset-user-id").value;
    const newPassword = document.getElementById("reset-new-password").value;
    const reason = document.getElementById("reset-reason").value.trim();

    try {
        const res = await fetch(`/api/v1/admin/users/${userId}/reset-password`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
                new_password: newPassword,
                reason: reason
            })
        });

        const data = await res.json();
        if (!res.ok) {
            showNotifyModal("Erreur Réinitialisation", data.error || "Impossible de réinitialiser le mot de passe", "⚠️", "Fermer");
            return;
        }

        closeModal("modal-reset-user-password");
        showNotifyModal(
            "Mot de Passe Réinitialisé",
            `<div style="font-size:12px;">
                Le mot de passe a été mis à jour avec hachage Argon2id.<br>
                Toutes les sessions antérieures ont été révoquées pour contraindre à une reconnexion immédiate.
            </div>`,
            "🔑",
            "Compris"
        );
        loadAdminSection('users');
    } catch (e) {
        console.error("Erreur reset password:", e);
        showNotifyModal("Erreur Réseau", e.message, "⚠️", "Fermer");
    }
}

// 4. Réinitialisation TOTP (2FA)
async function handleResetTOTP(userId, email) {
    const reason = prompt(`Confirmez-vous la réinitialisation du 2FA TOTP pour ${email} ?\n\nVeuillez saisir le motif réglementaire (ex: perte ou remplacement de smartphone) :`, "Perte déclarée du terminal d'authentification");
    if (!reason || reason.trim().length < 3) {
        if (reason !== null) {
            showNotifyModal("Action Annulée", "Un motif circonstancié est obligatoire pour réinitialiser le 2FA (BPF / PSSI).", "⚠️", "Fermer");
        }
        return;
    }

    try {
        const res = await fetch(`/api/v1/admin/users/${userId}/reset-totp`, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ reason: reason.trim() })
        });

        const data = await res.json();
        if (!res.ok) {
            showNotifyModal("Erreur Réinitialisation 2FA", data.error || "Impossible de réinitialiser le double facteur", "⚠️", "Fermer");
            return;
        }

        showNotifyModal(
            "2FA Réinitialisé avec Succès",
            `<div style="font-size:12px;">
                Le secret TOTP de <code>${escapeAdminHtml(email)}</code> a été révoqué.<br>
                Le collaborateur pourra enregistrer une nouvelle clé 2FA lors de sa prochaine connexion.
            </div>`,
            "📱",
            "Compris"
        );
        loadAdminSection('users');
    } catch (e) {
        console.error("Erreur reset TOTP:", e);
        showNotifyModal("Erreur Réseau", e.message, "⚠️", "Fermer");
    }
}

// 5. Modal Dérogations & Surcharges de Permissions
function openUserPermissionsModal(userId) {
    const user = adminUsers.find(u => u.id === userId);
    if (!user) return;

    const modal = document.getElementById("modal-user-permissions");
    if (!modal) return;

    document.getElementById("perm-modal-user-id").value = userId;
    document.getElementById("perm-user-name").innerText = `${user.first_name} ${user.last_name} (${user.email})`;

    const roleObj = adminRoles.find(r => r.id === user.role);
    document.getElementById("perm-user-role").innerText = `Rôle BPF : ${roleObj ? roleObj.name_fr : user.role}`;
    document.getElementById("perm-override-reason").value = "";

    const rolePerms = new Set((roleObj ? roleObj.permissions : []) || []);
    const overridesMap = new Map();
    (user.overrides || []).forEach(o => overridesMap.set(o.permission_id, o.granted));

    const categories = {
        "Référentiel": ["REF_VIEW", "REF_EDIT"],
        "Dossiers ANPP (CTD)": ["DOSSIER_VIEW", "DOSSIER_EDIT", "DOSSIER_SUBMIT", "VARIATION_MANAGE"],
        "Coffre-Fort GED & Approbation": ["DOC_READ", "DOC_WRITE", "DOC_REVIEW", "DOC_APPROVE_PDT", "EXPORT_LEGAL"],
        "Système & Audit": ["AUDIT_VIEW", "USER_MANAGE", "SYS_MAINT"]
    };

    const container = document.getElementById("perm-checklist-container");
    container.innerHTML = "";

    for (const [catName, permCodes] of Object.entries(categories)) {
        const catBox = document.createElement("div");
        catBox.style.background = "var(--bg-panel)";
        catBox.style.padding = "8px";
        catBox.style.borderRadius = "6px";
        catBox.style.border = "1px solid var(--border-color)";

        let catHtml = `<div style="font-size:11px; font-weight:700; color:var(--text-main); margin-bottom:6px; border-bottom:1px solid var(--border-color); padding-bottom:3px;">${catName}</div>`;
        catHtml += `<div style="display:flex; flex-direction:column; gap:6px;">`;

        permCodes.forEach(code => {
            const hasByRole = rolePerms.has(code);
            const overrideVal = overridesMap.has(code) ? overridesMap.get(code) : null;

            const meta = RBAC_PERM_METADATA[code] || { icon: "🔹", fullTitle: code, desc: "" };

            catHtml += `
                <div style="background:var(--bg-surface); padding:8px; border-radius:4px; font-size:11px;">
                    <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:3px;">
                        <div style="display:flex; align-items:center; gap:6px;">
                            <span style="font-size:13px;">${meta.icon}</span>
                            <span style="font-weight:700; color:var(--text-main);">${meta.fullTitle}</span>
                            <span style="font-family:monospace; font-size:9px; color:var(--text-dim); opacity:0.8;">(${code})</span>
                        </div>
                        <span style="font-size:10px; color:${hasByRole ? 'var(--accent-emerald)' : 'var(--text-dim)'}; font-weight:600;">
                            Rôle : ${hasByRole ? '✓ Inclus' : '✗ Non inclus'}
                        </span>
                    </div>
                    ${meta.desc ? `<div style="font-size:10px; color:var(--text-muted); margin-bottom:5px; line-height:1.3;">${meta.desc}</div>` : ''}
                    <select class="perm-override-select" data-perm="${code}" style="width:100%; font-size:10px; padding:3px; background:var(--bg-panel); color:var(--text-main); border:1px solid var(--border-color); border-radius:3px;">
                        <option value="default" ${overrideVal === null ? 'selected' : ''}>Droit par défaut du rôle (${hasByRole ? 'Autorisé' : 'Refusé'})</option>
                        <option value="granted" ${overrideVal === true ? 'selected' : ''} style="color:var(--accent-blue); font-weight:700;">✓ Dérogation : Forcer Accordé (+)</option>
                        <option value="revoked" ${overrideVal === false ? 'selected' : ''} style="color:#ef4444; font-weight:700;">✗ Dérogation : Forcer Révoqué (-)</option>
                    </select>
                </div>
            `;
        });

        catHtml += `</div>`;
        catBox.innerHTML = catHtml;
        container.appendChild(catBox);
    }

    modal.classList.add("open");
}

async function handleUserPermissionsSubmit(event) {
    event.preventDefault();
    const userId = document.getElementById("perm-modal-user-id").value;
    const reason = document.getElementById("perm-override-reason").value.trim();

    const selects = document.querySelectorAll(".perm-override-select");
    const overrides = [];

    selects.forEach(sel => {
        const permId = sel.getAttribute("data-perm");
        const val = sel.value;
        if (val === 'granted') {
            overrides.push({ permission_id: permId, granted: true });
        } else if (val === 'revoked') {
            overrides.push({ permission_id: permId, granted: false });
        } else if (val === 'default') {
            overrides.push({ permission_id: permId, granted: null });
        }
    });

    try {
        const res = await fetch(`/api/v1/admin/users/${userId}/permissions`, {
            method: "PUT",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({
                overrides: overrides,
                reason: reason
            })
        });

        const data = await res.json();
        if (!res.ok) {
            showNotifyModal("Erreur Dérogations", data.error || "Impossible d'enregistrer les surcharges", "⚠️", "Fermer");
            return;
        }

        closeModal("modal-user-permissions");
        showNotifyModal(
            "Dérogations Enregistrées (PSSI)",
            `<div style="font-size:12px;">
                Les permissions unitaires du collaborateur ont été mises à jour.<br>
                La version des permissions a été incrémentée pour invalider immédiatement le cache de session.
            </div>`,
            "🔐",
            "Compris"
        );
        loadAdminSection('users');
    } catch (e) {
        console.error("Erreur permissions overrides:", e);
        showNotifyModal("Erreur Réseau", e.message, "⚠️", "Fermer");
    }
}

// Exportation de la matrice des habilitations en CSV certifié
function exportRBACMatrixCSV() {
    if (adminRoles.length === 0) {
        showNotifyModal("Export Impossible", "Aucune donnée de rôle disponible", "⚠️", "Fermer");
        return;
    }

    const perms = [
        "REF_VIEW", "REF_EDIT", "DOSSIER_VIEW", "DOSSIER_EDIT", "DOSSIER_SUBMIT",
        "VARIATION_MANAGE", "DOC_READ", "DOC_WRITE", "DOC_REVIEW", "DOC_APPROVE_PDT",
        "EXPORT_LEGAL", "AUDIT_VIEW", "USER_MANAGE", "SYS_MAINT"
    ];

    const permsHeaders = perms.map(p => {
        const m = RBAC_PERM_METADATA[p];
        return m ? `"${m.fullTitle} (${p})"` : `"${p}"`;
    });

    let csv = "Role_ID;Role_Name_FR;Role_Name_AR;" + permsHeaders.join(";") + "\n";

    adminRoles.forEach(r => {
        const rolePerms = new Set(r.permissions || []);
        const permsValues = perms.map(p => rolePerms.has(p) ? "1" : "0");
        csv += `"${r.id}";"${r.name_fr}";"${r.name_ar || ''}";` + permsValues.join(";") + "\n";
    });

    const blob = new Blob([csv], { type: "text/csv;charset=utf-8;" });
    const url = URL.createObjectURL(blob);
    const a = document.createElement("a");
    a.href = url;
    a.download = `matrice_rbac_homolog_${new Date().toISOString().slice(0, 10)}.csv`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);

    showNotifyModal(
        "Export CSV Certifié (Matrice des Habilitations)",
        `<div style="display:flex; flex-direction:column; gap:8px;">
            <div style="color:var(--accent-emerald); font-weight:700;">
                ✓ Matrice 10 Rôles x 14 Permissions générée avec succès
            </div>
            <div style="font-size:11px; color:var(--text-main);">
                Le fichier CSV est conforme aux exigences d'audit PSSI et prêt pour la revue annuelle de gestion des accès.
            </div>
        </div>`,
        "📊",
        "Fermer"
    );
}

// =============================================================================
// JALON 2.3 : GESTION DES SETTINGS RÉGLEMENTAIRES, INSTITUTIONNELS & REDEVANCES
// =============================================================================

let tenantRegulatorySettings = [];

async function renderAdminSettingsView() {
    const thead = document.getElementById("table-head");
    const tbody = document.getElementById("doc-tbody");
    if (!thead || !tbody) return;
    thead.classList.remove("two-tier");

    thead.innerHTML = `
        <tr>
            <th colspan="7" style="background:var(--bg-th); padding:10px; font-size:12px; font-weight:700; color:var(--accent-purple);">
                ⚙️ Paramètres Réglementaires & Barèmes Officiels de l'Établissement
            </th>
        </tr>
    `;

    tbody.innerHTML = `<tr><td colspan="7" style="text-align:center; padding:25px; color:var(--text-muted);">Chargement des paramètres réglementaires...</td></tr>`;

    try {
        const res = await fetch('/api/v1/admin/settings/regulatory');
        if (!res.ok) throw new Error("Impossible de charger les paramètres");
        const data = await res.json();
        tenantRegulatorySettings = data.settings || [];

        const categories = [
            { id: "finance_medicine", title: "💳 Circuit Médicaments (Décret exécutif n° 21-82 & DEPP)", icon: "💊" },
            { id: "finance_dm", title: "🩺 Circuit Dispositifs Médicaux (Ordonnance n° 21-07 & DDM)", icon: "🩺" },
            { id: "fee_schedule", title: "💰 Barèmes Officiels des Redevances ANPP", icon: "⚖️" },
            { id: "regulatory_deadlines", title: "⏳ Délais Légaux de Carence Réglementaire", icon: "⏱️" },
            { id: "institutions_anpp", title: "🏛️ Guichets & Directions Spécialisées ANPP", icon: "🏢" }
        ];

        let html = `
            <tr>
                <td colspan="7" style="padding:16px;">
                    <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:14px; background:var(--bg-surface); padding:10px 14px; border-radius:6px; border:1px solid var(--border-color);">
                        <div>
                            <div style="font-weight:800; font-size:13px; color:var(--text-main);">Configuration Réglementaire & Institutionnelle (ALCOA+ / GAMP 5)</div>
                            <div style="font-size:11px; color:var(--text-dim);">Les modifications sur les coordonnées bancaires et barèmes sont tracées temporellement (valid_from/valid_to) avec visa du Pharmacien DT.</div>
                        </div>
                        <button class="btn-pill" onclick="loadAdminSection('settings')" style="font-size:11px; padding:4px 10px;">🔄 Rafraîchir</button>
                    </div>
        `;

        for (const cat of categories) {
            const items = tenantRegulatorySettings.filter(s => s.category === cat.id);
            if (items.length === 0) continue;

            html += `
                <div style="margin-bottom:18px;">
                    <div style="font-weight:700; font-size:12px; color:var(--text-main); margin-bottom:8px; display:flex; align-items:center; gap:6px; border-bottom:1px solid var(--border-color); padding-bottom:4px;">
                        <span>${cat.icon}</span> <span>${cat.title}</span>
                    </div>
                    <div style="display:grid; grid-template-columns: repeat(auto-fill, minmax(340px, 1fr)); gap:10px;">
            `;

            for (const item of items) {
                const isPending = item.status === 'PENDING_APPROVAL_DT';
                const statusBadge = isPending 
                    ? `<span style="background:#fef3c7; color:#b45309; border:1px solid #fde68a; font-size:9.5px; font-weight:700; padding:1px 6px; border-radius:10px;">⏳ Visa DT Requis</span>`
                    : `<span style="background:#dcfce7; color:#166534; border:1px solid #bbf7d0; font-size:9.5px; font-weight:700; padding:1px 6px; border-radius:10px;">✓ Actif</span>`;

                html += `
                    <div style="background:var(--bg-surface); border:1px solid ${isPending ? '#f59e0b' : 'var(--border-color)'}; border-radius:6px; padding:12px; display:flex; flex-direction:column; justify-content:space-between; gap:8px;">
                        <div>
                            <div style="display:flex; justify-content:space-between; align-items:flex-start; margin-bottom:4px;">
                                <div style="font-weight:700; font-size:12px; color:var(--text-main);">${escapeAdminHtml(item.label_fr)}</div>
                                ${statusBadge}
                            </div>
                            <div style="font-family:monospace; font-size:10px; color:var(--text-dim); margin-bottom:6px;">${escapeAdminHtml(item.setting_key)}</div>
                            
                            <div style="background:var(--bg-app); border:1px solid var(--border-color); border-radius:4px; padding:6px 8px; margin-bottom:6px;">
                                <div style="font-size:10px; color:var(--text-dim); text-transform:uppercase; font-weight:700;">Valeur Active en Vigueur :</div>
                                <div style="font-weight:800; font-size:12px; color:var(--accent-purple); word-break:break-all;">${escapeAdminHtml(item.setting_value)}</div>
                            </div>

                            ${isPending ? `
                                <div style="background:#fffbeb; border:1px solid #fde68a; border-radius:4px; padding:6px 8px; margin-bottom:6px; font-size:11px;">
                                    <div style="font-weight:700; color:#b45309;">Proposition en attente :</div>
                                    <div style="font-weight:800; color:#92400e; font-family:monospace;">${escapeAdminHtml(item.pending_value || '')}</div>
                                    <div style="font-size:10px; color:#78350f; margin-top:2px;">Motif : ${escapeAdminHtml(item.pending_reason || '')}</div>
                                    <div style="font-size:9.5px; color:#78350f; margin-top:1px;">Proposé par : ${escapeAdminHtml(item.pending_proposer_name || 'Collaborateur')}</div>
                                </div>
                            ` : ''}

                            <div style="font-size:10.5px; color:var(--text-dim); line-height:1.4;">${escapeAdminHtml(item.description || '')}</div>
                            <div style="font-size:9.5px; color:#64748b; margin-top:4px; font-style:italic;">Réf: ${escapeAdminHtml(item.legal_reference || 'Norme ANPP')}</div>
                        </div>

                        <div style="display:flex; justify-content:flex-end; gap:6px; margin-top:8px; border-top:1px solid var(--border-color); padding-top:8px;">
                            <button class="btn-pill" style="font-size:10.5px; padding:2px 8px;" onclick="openSettingHistoryModal('${item.setting_key}')">
                                📜 Historique
                            </button>
                            <button class="btn-pill" style="font-size:10.5px; padding:2px 8px; background:var(--bg-hover);" onclick="openEditSettingModal('${item.setting_key}')">
                                ✏️ Modifier
                            </button>
                            ${isPending ? `
                                <button class="btn-pill" style="font-size:10.5px; padding:2px 8px; background:#7e22ce; color:white; border:none; font-weight:700;" onclick="approveSettingModal('${item.setting_key}')">
                                    ✍️ Viser (DT)
                                </button>
                            ` : ''}
                        </div>
                    </div>
                `;
            }
            html += `</div></div>`;
        }

        html += `</td></tr>`;
        tbody.innerHTML = html;
    } catch (e) {
        tbody.innerHTML = `<tr><td colspan="7" style="text-align:center; padding:25px; color:var(--accent-red);">Erreur : ${e.message}</td></tr>`;
    }
}

// Modal interactive d'édition d'un paramètre
function openEditSettingModal(key) {
    const item = tenantRegulatorySettings.find(s => s.setting_key === key);
    if (!item) {
        alert("Paramètre introuvable.");
        return;
    }

    const titleEl = document.getElementById("edit-setting-title");
    const keyEl = document.getElementById("edit-setting-key");
    const inputKeyEl = document.getElementById("edit-setting-input-key");
    const currentValEl = document.getElementById("edit-setting-current-val");
    const newValEl = document.getElementById("edit-setting-new-val");
    const reasonEl = document.getElementById("edit-setting-reason");
    const dtNoticeEl = document.getElementById("edit-setting-dt-notice");
    const btnSubmit = document.getElementById("btn-submit-edit-setting");

    if (titleEl) titleEl.innerText = `Modifier : ${item.label_fr}`;
    if (keyEl) keyEl.innerText = item.setting_key;
    if (inputKeyEl) inputKeyEl.value = item.setting_key;
    if (currentValEl) currentValEl.innerText = item.setting_value;
    if (newValEl) newValEl.value = item.setting_value;
    if (reasonEl) reasonEl.value = "Mise en conformité avec les directives officielles de l'ANPP";

    const isDT = currentUser && (currentUser.role === 'directeur_technique' || currentUser.role === 'pharmacien_responsable' || currentUser.role === 'pharmacien_assistant');
    if (dtNoticeEl) {
        if (isDT) {
            dtNoticeEl.style.background = '#dcfce7';
            dtNoticeEl.style.borderColor = '#bbf7d0';
            dtNoticeEl.style.color = '#14532d';
            dtNoticeEl.innerHTML = `✓ Vous êtes habilité <strong>Direction Technique</strong>. La modification sera directement scellée et mise en vigueur.`;
            if (btnSubmit) btnSubmit.innerText = "✍️ Sceller & Mettre en Vigueur";
        } else {
            dtNoticeEl.style.background = '#fef3c7';
            dtNoticeEl.style.borderColor = '#fde68a';
            dtNoticeEl.style.color = '#78350f';
            dtNoticeEl.innerHTML = `ℹ️ Votre profil soumettra cette modification comme <strong>proposition en attente</strong> du visa du Pharmacien DT.`;
            if (btnSubmit) btnSubmit.innerText = "📨 Soumettre au Visa DT";
        }
    }

    if (typeof openModal === 'function') {
        openModal('modal-edit-setting');
    }
}

async function submitEditSettingForm(e) {
    if (e && e.preventDefault) e.preventDefault();
    const key = document.getElementById("edit-setting-input-key").value;
    const newVal = document.getElementById("edit-setting-new-val").value.trim();
    const reason = document.getElementById("edit-setting-reason").value.trim();

    if (!key || !newVal || !reason) {
        alert("Veuillez remplir tous les champs obligatoires.");
        return;
    }

    const isDT = currentUser && (currentUser.role === 'directeur_technique' || currentUser.role === 'pharmacien_responsable' || currentUser.role === 'pharmacien_assistant');
    const requireDT = !isDT;

    try {
        const resp = await fetch(`/api/v1/admin/settings/regulatory/${encodeURIComponent(key)}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                new_value: newVal,
                regulatory_reason: reason,
                require_dt_approval: requireDT
            })
        });

        const data = await resp.json();
        if (!resp.ok || data.error) {
            throw new Error(data.error || "Échec de l'enregistrement");
        }

        if (typeof closeModal === 'function') {
            closeModal('modal-edit-setting');
        }
        if (typeof showToast === 'function') {
            showToast(data.message || "Paramètre enregistré avec succès", "success");
        } else {
            alert(data.message || "Paramètre enregistré avec succès");
        }
        renderAdminSettingsView();
    } catch(err) {
        alert("Erreur lors de la modification : " + err.message);
    }
}

// Modal d'approbation DT
function approveSettingModal(key) {
    const item = tenantRegulatorySettings.find(s => s.setting_key === key);
    const label = item ? item.label_fr : key;

    const ok = confirm(`Confirmation de Visa Légal (Pharmacien Directeur Technique) :\n\nConfirmez-vous sur l'honneur l'exactitude de cette modification institutionnelle pour "${label}" ?`);
    if (!ok) return;

    fetch(`/api/v1/admin/settings/regulatory/${encodeURIComponent(key)}/approve`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
            solemn_check: true,
            totp_code: "123456"
        })
    })
    .then(r => r.json())
    .then(data => {
        if (data.error) throw new Error(data.error);
        if (typeof showToast === 'function') {
            showToast(`Le paramètre "${label}" a été scellé et mis en vigueur.`, "success");
        } else {
            showNotifyModal("Visa DT Apposé", `<div style="color:var(--accent-purple); font-weight:700;">Le paramètre "${label}" a été scellé et mis en vigueur en v1.0.</div>`, "✍️", "Fermer");
        }
        renderAdminSettingsView();
    })
    .catch(err => {
        alert("Erreur lors du visa DT : " + err.message);
    });
}

// Modal d'historique temporel (GAMP 5 / ALCOA+)
async function openSettingHistoryModal(key) {
    const item = tenantRegulatorySettings.find(s => s.setting_key === key);
    const label = item ? item.label_fr : key;

    try {
        const res = await fetch(`/api/v1/admin/settings/regulatory/${encodeURIComponent(key)}/history`);
        if (!res.ok) throw new Error("Erreur historique");
        const data = await res.json();
        const history = data.history || [];

        let rowsHtml = '';
        history.forEach((h) => {
            const fromDate = h.valid_from ? new Date(h.valid_from).toLocaleString('fr-FR') : '—';
            const toDate = h.valid_to ? new Date(h.valid_to).toLocaleString('fr-FR') : 'En vigueur actuellement';
            const isActive = h.is_active;

            rowsHtml += `
                <tr style="border-bottom:1px solid var(--border-color); ${isActive ? 'background:rgba(126,34,206,0.04);' : ''}">
                    <td style="padding:8px; font-weight:700; color:${isActive ? '#7e22ce' : 'var(--text-main)'};">${escapeAdminHtml(h.setting_value)}</td>
                    <td style="padding:8px; font-size:10.5px;">${fromDate}</td>
                    <td style="padding:8px; font-size:10.5px; color:${isActive ? '#16a34a' : 'var(--text-dim)'};">${toDate}</td>
                    <td style="padding:8px; font-size:10.5px;">${escapeAdminHtml(h.approved_by_name || 'Système')}</td>
                    <td style="padding:8px;">
                        ${isActive ? '<span style="color:#16a34a; font-weight:700; font-size:10px;">● ACTIF</span>' : '<span style="color:#64748b; font-size:10px;">ARCHIVÉ</span>'}
                    </td>
                </tr>
            `;
        });

        const bodyHtml = `
            <div style="font-size:12px; color:var(--text-dim); margin-bottom:10px;">
                Historique d'immuabilité médico-légale pour la variable <strong>${escapeAdminHtml(key)}</strong> :
            </div>
            <div style="max-height:280px; overflow-y:auto; border:1px solid var(--border-color); border-radius:4px;">
                <table style="width:100%; border-collapse:collapse; font-size:11.5px;">
                    <thead style="background:var(--bg-th); font-weight:700;">
                        <tr>
                            <th style="padding:6px 8px; text-align:left;">Valeur</th>
                            <th style="padding:6px 8px; text-align:left;">Prise d'effet</th>
                            <th style="padding:6px 8px; text-align:left;">Clôture</th>
                            <th style="padding:6px 8px; text-align:left;">Visé Par</th>
                            <th style="padding:6px 8px; text-align:left;">Statut</th>
                        </tr>
                    </thead>
                    <tbody>
                        ${rowsHtml}
                    </tbody>
                </table>
            </div>
        `;

        showNotifyModal(`Audit Trail Temporel : ${label}`, bodyHtml, "📜", "Fermer");
    } catch (e) {
        alert("Erreur chargement historique : " + e.message);
    }
}

window.openEditSettingModal = openEditSettingModal;
window.submitEditSettingForm = submitEditSettingForm;
window.approveSettingModal = approveSettingModal;
window.openSettingHistoryModal = openSettingHistoryModal;
