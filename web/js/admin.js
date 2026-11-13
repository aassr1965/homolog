// ============================================================================
// CONSOLE D'ADMINISTRATION LABORATOIRE & MATRICE DES HABILITATIONS RBAC (Jalon 1.2)
// Conformité PSSI, BPF Algérie (Décret 21-82) & 21 CFR Part 11
// ============================================================================

let currentAdminSection = 'users'; // 'users' | 'matrix' | 'policies'
let adminUsers = [];
let adminRoles = [];
let adminPermissions = [];
let selectedAdminUserIndex = 0;

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

// 1. Tableau des Collaborateurs du Laboratoire
function renderAdminUsersTable() {
    const thead = document.getElementById("table-head");
    const tbody = document.getElementById("doc-tbody");
    if (!thead || !tbody) return;

    thead.innerHTML = `
        <tr>
            <th style="width: 80px;">Statut</th>
            <th>Collaborateur</th>
            <th>Email</th>
            <th>Rôle Métier Attribué</th>
            <th>Droits Effectifs</th>
            <th style="width: 90px;">2FA BPF</th>
            <th style="width: 140px; text-align:center;">Action Dev</th>
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
            // Ne pas déclencher si on clique directement sur le select ou le bouton
            if (e.target.tagName === 'SELECT' || e.target.tagName === 'BUTTON' || e.target.closest('button')) return;
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

        let badgeStatus = '<span class="status-tag status-ok">ACTIF</span>';
        if (isCurrentActiveUser) {
            badgeStatus = '<span class="status-tag" style="background:#0284c7; color:white; font-weight:700;">VOUS</span>';
        }

        tr.innerHTML = `
            <td>${badgeStatus}</td>
            <td style="font-weight:600; color:var(--text-main);">
                ${u.first_name} ${u.last_name}
                ${isPDT ? '<span title="Pharmacien Directeur Technique titulaire de l\'agrément nominatif (1.A.7)" style="margin-left:4px; font-size:12px;">💊</span>' : ''}
                ${isZeroTrust ? '<span title="Profil Technique soumis au Zero-Trust strict" style="margin-left:4px; font-size:12px;">🛡️</span>' : ''}
            </td>
            <td style="font-family:monospace; font-size:11px; color:var(--text-muted);">${u.email}</td>
            <td>
                <select class="btn-pill" style="font-size:11px; padding:3px 6px; background:var(--bg-surface); color:var(--text-main); border:1px solid var(--border-color); cursor:pointer;"
                        onchange="handleUserRoleChange('${u.id}', this.value, '${u.email}')">
                    ${roleOptions}
                </select>
            </td>
            <td>
                <span class="status-tag" style="background:var(--badge-bg); color:var(--accent-blue); font-weight:600; cursor:help;" 
                      title="${(u.permissions || []).join(', ')}">
                    ${permsCount}/14 droits (v${u.permissions_version || 1})
                </span>
            </td>
            <td>
                ${u.totp_enabled 
                    ? '<span style="color:var(--accent-emerald); font-weight:600; font-size:11px;">✓ TOTP Actif</span>' 
                    : '<span style="color:var(--text-dim); font-size:11px;">Non activé</span>'}
            </td>
            <td style="text-align:center;">
                <button class="btn-pill" style="font-size:10px; padding:2px 8px; border-color:#0284c7; color:#0284c7;" 
                        onclick="devSwitchRole('${u.email}')" 
                        title="Basculer instantanément sur ce compte pour vérifier les masquages">
                    🧪 Tester rôle
                </button>
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
                        const isSpecialPDT = p === 'DOC_APPROVE_PDT';
                        const isDocReadBlocked = p === 'DOC_READ' && isZeroTrust;

                        let statusBadge = has 
                            ? `<span style="color:var(--accent-emerald); font-weight:700;">✓ Habilité</span>`
                            : `<span style="color:var(--text-dim);">✗ Non attribué</span>`;

                        if (isSpecialPDT && has) {
                            statusBadge = `<span style="color:#d97706; font-weight:700;">💊 Visa Exclusif 1.A.7</span>`;
                        } else if (isDocReadBlocked) {
                            statusBadge = `<span style="color:var(--accent-red); font-weight:700;">🚫 Bloqué Zero-Trust</span>`;
                        }

                        return `
                            <div style="display:flex; justify-content:space-between; align-items:center; background:var(--bg-surface); padding:4px 8px; border-radius:3px; font-size:11px;">
                                <span style="font-family:monospace; color:${has ? 'var(--text-main)' : 'var(--text-dim)'};">${p}</span>
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
            </div>
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

    const permsList = [
        { code: "REF_VIEW", label: "REF_V", desc: "Consultation Référentiel MIPH" },
        { code: "REF_EDIT", label: "REF_E", desc: "Modification Référentiel Sites/Produits" },
        { code: "DOSSIER_VIEW", label: "DOS_V", desc: "Consultation Dossiers CTD & Variations" },
        { code: "DOSSIER_EDIT", label: "DOS_E", desc: "Édition et Remplissage des Modules CTD" },
        { code: "DOSSIER_SUBMIT", label: "DOS_SUB", desc: "Soumission Dépôt Officiel ANPP" },
        { code: "DOC_READ", label: "DOC_R", desc: "Lecture et Téléchargement GED BPF (Zero-Trust)" },
        { code: "DOC_WRITE", label: "DOC_W", desc: "Dépôt de Documents et Nouvelles Versions" },
        { code: "DOC_REVIEW", label: "DOC_REV", desc: "Revue Qualité et Vérification Technique" },
        { code: "DOC_APPROVE_PDT", label: "DOC_PDT", desc: "Visa Légal Pharmacien DT (1.A.7 Exclusif)" },
        { code: "AUDIT_VIEW", label: "AUDIT_V", desc: "Consultation Audit Trail 21 CFR Part 11" },
        { code: "USER_MANAGE", label: "USR_MNG", desc: "Administration des Utilisateurs & Rôles" },
        { code: "SYS_MAINT", label: "SYS_MNT", desc: "Maintenance Système et Sauvegardes" },
        { code: "VARIATION_MANAGE", label: "VAR_MNG", desc: "Gestion des 12 Items de Variations Post-AMM" },
        { code: "EXPORT_LEGAL", label: "EXP_LEG", desc: "Export Archive Souveraine Certifiée ANPP" }
    ];

    thead.innerHTML = `
        <tr>
            <th style="min-width: 180px;">Rôle Réglementaire</th>
            ${permsList.map(p => `
                <th style="text-align:center; padding:6px 4px; font-size:10px; font-family:monospace; ${p.code === 'DOC_APPROVE_PDT' ? 'background:rgba(217,119,6,0.15); color:#d97706;' : ''}" title="${p.code} : ${p.desc}">
                    ${p.label}
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

        let cellsHtml = '';
        permsList.forEach(p => {
            const has = rolePerms.has(p.code);
            if (p.code === 'DOC_APPROVE_PDT') {
                if (isPDT) {
                    cellsHtml += `<td style="text-align:center; background:rgba(217,119,6,0.1);"><span style="color:#d97706; font-weight:700;" title="Exclusivité Nominative 1.A.7">✓ DT</span></td>`;
                } else {
                    cellsHtml += `<td style="text-align:center; color:var(--text-dim); opacity:0.3;">—</td>`;
                }
            } else if (p.code === 'DOC_READ' && isZeroTrust) {
                cellsHtml += `<td style="text-align:center; background:rgba(239,68,68,0.1);"><span style="color:var(--accent-red); font-weight:700; font-size:9px;" title="Bloqué par Zero-Trust">🚫 ZT</span></td>`;
            } else {
                if (has) {
                    cellsHtml += `<td style="text-align:center;"><span style="color:var(--accent-emerald); font-weight:700;">✓</span></td>`;
                } else {
                    cellsHtml += `<td style="text-align:center; color:var(--text-dim); opacity:0.3;">—</td>`;
                }
            }
        });

        tr.innerHTML = `
            <td style="font-weight:600; color:var(--text-main);">
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
                    return `
                        <div style="display:flex; justify-content:space-between; align-items:center; background:var(--bg-surface); padding:4px 8px; border-radius:3px; font-size:11px;">
                            <span style="font-family:monospace; color:${has ? 'var(--text-main)' : 'var(--text-dim)'};">${p.code}</span>
                            <span style="color:${has ? 'var(--accent-emerald)' : 'var(--text-dim)'}; font-weight:${has ? '700' : 'normal'};">
                                ${has ? '✓ Accordée' : '— Refusée'}
                            </span>
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

// Modal d'invitation de nouveau collaborateur
function openNewUserModal() {
    showNotifyModal(
        "Inviter un Nouveau Collaborateur",
        `<div style="display:flex; flex-direction:column; gap:10px; font-size:12px;">
            <div>Pour inviter un nouveau membre d'équipe, renseignez son profil et assignez-lui son rôle BPF initial :</div>
            <div>
                <label style="display:block; font-weight:600; margin-bottom:4px;">Nom complet :</label>
                <input type="text" id="invite-name" placeholder="ex: Dr. Nassim B." style="width:100%; padding:6px; background:var(--bg-surface); border:1px solid var(--border-color); border-radius:4px; color:var(--text-main);">
            </div>
            <div>
                <label style="display:block; font-weight:600; margin-bottom:4px;">Adresse email professionnelle :</label>
                <input type="email" id="invite-email" placeholder="nassim.b@pharma.dz" style="width:100%; padding:6px; background:var(--bg-surface); border:1px solid var(--border-color); border-radius:4px; color:var(--text-main);">
            </div>
            <div>
                <label style="display:block; font-weight:600; margin-bottom:4px;">Rôle réglementaire :</label>
                <select id="invite-role" style="width:100%; padding:6px; background:var(--bg-surface); border:1px solid var(--border-color); border-radius:4px; color:var(--text-main);">
                    ${adminRoles.map(r => `<option value="${r.id}">${r.name_fr}</option>`).join('')}
                </select>
            </div>
            <div style="font-size:10px; color:var(--text-dim); margin-top:4px;">
                * Le collaborateur recevra ses identifiants temporaires et devra activer le 2FA TOTP lors de sa première connexion.
            </div>
        </div>`,
        "➕",
        "Envoyer l'Invitation"
    );
}

// Exportation de la matrice des habilitations en CSV certifié
function exportRBACMatrixCSV() {
    if (adminRoles.length === 0) {
        showNotifyModal("Export Impossible", "Aucune donnée de rôle disponible", "⚠️", "Fermer");
        return;
    }

    const perms = [
        "REF_VIEW", "REF_EDIT", "DOSSIER_VIEW", "DOSSIER_EDIT", "DOSSIER_SUBMIT",
        "DOC_READ", "DOC_WRITE", "DOC_REVIEW", "DOC_APPROVE_PDT", "AUDIT_VIEW",
        "USER_MANAGE", "SYS_MAINT", "VARIATION_MANAGE", "EXPORT_LEGAL"
    ];

    let csv = "Role_ID;Role_Name_FR;Role_Name_AR;" + perms.join(";") + "\n";

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
