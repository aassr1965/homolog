// ============================================================================
// CONSOLE D'ADMINISTRATION LABORATOIRE & MATRICE DES HABILITATIONS RBAC (Jalon 1.2)
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
    "DOC_REVIEW": { icon: "🔍", label: "Revue BPF", fullTitle: "Revue Qualité BPF", desc: "Revue technique et vérification qualité BPF (Maker-Checker)", cat: "ged", isLast: false },
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
    const usersCount = adminUsers.length > 0 ? `${adminUsers.length} actifs` : '9 actifs';
    container.innerHTML = `
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

// Chargement principal de la section administration
async function loadAdminSection(section) {
    currentAdminSection = section || 'users';
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
            if (window.innerWidth < 1400) togglePreviewPane(true);
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

    const roleObj = adminRoles.find(r => r.id === user.role);
    const roleLabel = roleObj ? roleObj.name_fr : user.role;
    const isPDT = user.role === 'directeur_technique';
    const isZeroTrust = user.role === 'admin_systeme_local' || user.role === 'super_admin_editeur';

    const userPerms = new Set(user.permissions || []);

    const categories = {
        "Référentiel": ["REF_VIEW", "REF_EDIT"],
        "Dossiers ANPP (CTD)": ["DOSSIER_VIEW", "DOSSIER_EDIT", "DOSSIER_SUBMIT", "VARIATION_MANAGE"],
        "Coffre-Fort GED & Approbation": ["DOC_READ", "DOC_WRITE", "DOC_REVIEW", "DOC_APPROVE_PDT", "EXPORT_LEGAL"],
        "Système, PSSI & Audit": ["AUDIT_VIEW", "USER_MANAGE", "SYS_MAINT"]
    };

    let permsBreakdownHtml = '';
    for (const [cat, perms] of Object.entries(categories)) {
        permsBreakdownHtml += `
            <div style="margin-top:10px;">
                <div style="font-size:10px; font-weight:700; color:var(--text-muted); text-transform:uppercase; margin-bottom:4px;">${cat}</div>
                <div style="display:flex; flex-direction:column; gap:3px;">
                    ${perms.map(p => {
                        const has = userPerms.has(p);
                        const meta = RBAC_PERM_METADATA[p] || { icon: "🔹", label: p, fullTitle: p };
                        const isSpecialPDT = p === 'DOC_APPROVE_PDT';
                        const isDocReadBlocked = p === 'DOC_READ' && isZeroTrust;

                        let statusBadge = has 
                            ? `<span style="color:var(--accent-emerald); font-weight:700;">✓ Habilité</span>`
                            : `<span style="color:var(--text-dim);">✗ Non attribué</span>`;

                        if (isSpecialPDT && has) {
                            statusBadge = `<span style="color:#d97706; font-weight:700;">✍️ Visa Exclusif 1.A.7</span>`;
                        } else if (isDocReadBlocked) {
                            statusBadge = `<span style="color:var(--accent-red); font-weight:700;">🚫 Bloqué Zero-Trust</span>`;
                        }

                        return `
                            <div style="display:flex; justify-content:space-between; align-items:center; background:var(--bg-surface); padding:4px 8px; border-radius:3px; font-size:11px;">
                                <div style="display:flex; align-items:center; gap:6px;">
                                    <span>${meta.icon}</span>
                                    <span style="font-weight:600; color:${has ? 'var(--text-main)' : 'var(--text-dim)'};">${meta.fullTitle}</span>
                                    <span style="font-family:monospace; font-size:9px; color:var(--text-dim); opacity:0.7;">(${p})</span>
                                </div>
                                <div>${statusBadge}</div>
                            </div>
                        `;
                    }).join('')}
                </div>
            </div>
        `;
    }

    preview.innerHTML = `
        <div class="pdf-sim-page">
            <div class="pdf-sim-header">
                RÉPUBLIQUE ALGÉRIENNE DÉMOCRATIQUE ET POPULAIRE<br>
                SYSTÈME D'INFORMATION PHARMACEUTIQUE SOUTERRAIN SANAD PHARMA
            </div>
            <div style="font-size:12px; font-weight:700; color:#0369a1; text-align:center; margin-bottom:10px;">
                FICHE D'HABILITATION ET DE CONTRÔLE D'ACCÈS (BPF / PSSI)
            </div>
            <div style="font-size:11px; line-height:1.6;">
                <strong>Collaborateur :</strong> ${user.first_name} ${user.last_name}<br>
                <strong>Identifiant Email :</strong> <code>${user.email}</code><br>
                <strong>Rôle Officiel BPF :</strong> <span style="font-weight:700; color:var(--accent-blue);">${roleLabel}</span><br>
                <strong>Version des Permissions :</strong> <code>v${user.permissions_version || 1}</code> (Incrémentation automatique sur modification)<br>
                <strong>Authentification 2FA TOTP :</strong> ${user.totp_enabled ? '<span style="color:var(--accent-emerald); font-weight:600;">✓ Activée (21 CFR Part 11)</span>' : 'Non activée'}<br>
                <strong>Statut du Compte :</strong> ${user.is_active ? '<span style="color:var(--accent-emerald); font-weight:700;">🟢 Actif</span>' : '<span style="color:#ef4444; font-weight:700;">🔴 Suspendu / Inactif</span>'}<br>
                ${user.contract_expires_at ? `<strong>Échéance Contrat Expert :</strong> <code>${new Date(user.contract_expires_at).toLocaleDateString()}</code><br>` : ''}
            </div>
        </div>

        <div style="display:flex; gap:6px; margin: 10px 0;">
            <button class="btn-pill" style="flex:1; background:var(--accent-blue); color:white; border:none; padding:6px; font-size:11px; font-weight:600;" onclick="openUserPermissionsModal('${user.id}')">
                🔐 Gérer Dérogations
            </button>
            <button class="btn-pill" style="flex:1; background:var(--bg-surface); border:1px solid var(--accent-amber); color:var(--accent-amber); padding:6px; font-size:11px; font-weight:600;" onclick="openResetPasswordModal('${user.id}', '${escapeAdminHtml(user.email)}', '${escapeAdminHtml(user.first_name + ' ' + user.last_name)}')">
                🔑 Mot de Passe
            </button>
        </div>

        ${isPDT ? `
            <div style="background:rgba(217, 119, 6, 0.1); border:1px solid #d97706; border-radius:4px; padding:10px; font-size:11px; color:var(--text-main); margin-bottom:12px;">
                <strong>💊 Titulaire du Visa Réglementaire (1.A.7) :</strong> Ce compte est le seul habilité par la réglementation algérienne à sceller les dossiers d'AMM et les pièces opposables à l'ANPP (Annexe 11 BPF).
            </div>
        ` : ''}

        ${isZeroTrust ? `
            <div style="background:rgba(239, 68, 68, 0.1); border:1px solid #ef4444; border-radius:4px; padding:10px; font-size:11px; color:var(--text-main); margin-bottom:12px;">
                <strong>🛡️ Isolation Zero-Trust Obligatoire :</strong> En vertu du principe de moindre privilège pharmaceutique, ce compte d'administration système ne peut en aucun cas télécharger les fichiers contenant les formules secrètes ou les données cliniques (Blocage 403 <code>ERR_ZERO_TRUST_DENIED</code>).
            </div>
        ` : ''}

        ${user.role === 'consultant_externe' ? `
            <div style="background:rgba(2, 132, 199, 0.1); border:1px solid #0284c7; border-radius:4px; padding:10px; font-size:11px; color:var(--text-main); margin-bottom:12px;">
                <strong>🤝 Accès Tiers Cloisonné (CRO / Expert) :</strong> Ce compte invité ne possède aucun accès transverse (GED et Référentiel masqués). Il est strictement confiné au dossier d'enregistrement assigné par la table <code>dossier_consultants</code> avec masquage automatique des rubriques de prix et données financières (1.A.4, 1.H.4, 1.H.5).
            </div>
        ` : ''}

        <div class="meta-box">
            <div style="font-size:11px; font-weight:700; color:var(--text-main); margin-bottom:6px;">
                Détail des 14 Habilitations Unitaires RBAC
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
                <div style="font-size: 10px; color: var(--text-dim); font-weight: normal; margin-top:2px;">9 profils métier (CDC 7.1)</div>
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
            if (window.innerWidth < 1400) togglePreviewPane(true);
        };

        tbody.appendChild(tr);
    });

    if (adminRoles.length > 0) {
        renderRolePreview(adminRoles[0], permsList);
    }
}

// Aperçu d'un rôle dans le volet latéral
function renderRolePreview(role, permsList) {
    const preview = document.getElementById("preview-body");
    if (!preview || !role) return;

    const rolePerms = new Set(role.permissions || []);
    const isZeroTrust = role.id === 'admin_systeme_local' || role.id === 'super_admin_editeur';

    preview.innerHTML = `
        <div class="pdf-sim-page">
            <div class="pdf-sim-header">
                RÉPUBLIQUE ALGÉRIENNE DÉMOCRATIQUE ET POPULAIRE<br>
                RÉFÉRENTIEL D'HABILITATION DES MÉTIERS PHARMACEUTIQUES
            </div>
            <div style="font-size:12px; font-weight:700; color:#0369a1; text-align:center; margin-bottom:10px;">
                DÉFINITION DU RÔLE : ${role.name_fr.toUpperCase()}
            </div>
            <div style="font-size:11px; line-height:1.6;">
                <strong>Identifiant Système :</strong> <code>${role.id}</code><br>
                <strong>Nom en Arabe :</strong> ${role.name_ar || '-'}<br>
                <strong>Description :</strong> ${role.description || '-'}<br>
                <strong>Habilitations Attribuées :</strong> ${(role.permissions || []).length} / 14 permissions unitaires<br>
            </div>
        </div>

        <div class="meta-box">
            <div style="font-size:11px; font-weight:700; color:var(--text-main); margin-bottom:8px;">
                Matrice des Habilitations par Défaut
            </div>
            <div style="display:flex; flex-direction:column; gap:4px;">
                ${permsList.map(p => {
                    const has = rolePerms.has(p.code);
                    const isSpecialPDT = p.code === 'DOC_APPROVE_PDT';
                    const isDocReadBlocked = p.code === 'DOC_READ' && isZeroTrust;

                    let badge = has 
                        ? `<span style="color:var(--accent-emerald); font-weight:700;">✓ Accordée</span>`
                        : `<span style="color:var(--text-dim);">— Refusée</span>`;

                    if (isSpecialPDT && has) {
                        badge = `<span style="color:#d97706; font-weight:700;">✍️ Visa Exclusif 1.A.7</span>`;
                    } else if (isDocReadBlocked) {
                        badge = `<span style="color:var(--accent-red); font-weight:700;">🚫 Bloqué Zero-Trust</span>`;
                    }

                    return `
                        <div style="display:flex; justify-content:space-between; align-items:center; background:var(--bg-surface); padding:5px 8px; border-radius:3px; font-size:11px;">
                            <div style="display:flex; align-items:center; gap:6px;">
                                <span style="font-size:13px;">${p.icon}</span>
                                <span style="font-weight:600; color:${has ? 'var(--text-main)' : 'var(--text-dim)'};">${p.fullTitle}</span>
                                <span style="font-family:monospace; font-size:9px; color:var(--text-dim); opacity:0.7;">(${p.code})</span>
                            </div>
                            <div>${badge}</div>
                        </div>
                    `;
                }).join('')}
            </div>
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
                🛡️ Politiques d'Intégrité BPF & Garde-Fous Cryptographiques (Section 7.2 du CDC)
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
                            Toute modification d'habilitation, attribution de rôle ou signature est consignée de manière inaltérable dans la table PostgreSQL <code>audit_logs</code> protégée par un trigger SQL interdisant tout <code>UPDATE</code> et <code>DELETE</code>.
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
                ✓ Matrice 9 Rôles x 14 Permissions générée avec succès
            </div>
            <div style="font-size:11px; color:var(--text-main);">
                Le fichier CSV est conforme aux exigences d'audit PSSI et prêt pour la revue annuelle de gestion des accès.
            </div>
        </div>`,
        "📊",
        "Fermer"
    );
}
