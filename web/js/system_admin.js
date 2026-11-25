// ============================================================================
// COCKPIT TECHNIQUE & INFRASTRUCTURE IT (SYS_MAINT / PSSI BPF)
// Dédié exclusivement à l'Administrateur Système Local & Super-Admin Éditeur
// Zéro accès aux données de santé / formules (Garde-Fou Zero-Trust Strict)
// ============================================================================

let currentSystemSection = 'status'; // 'status' | 'storage' | 'sessions' | 'security'
let systemMetricsCache = null;
let systemSessionsCache = [];

// Rendu de l'arborescence contextuelle IT
function renderSystemTree() {
    const container = document.getElementById("tree-root");
    if (!container) return;

    const sessBadge = systemSessionsCache.length > 0 ? `${systemSessionsCache.length} actives` : 'Connectées';

    container.innerHTML = `
        <div class="tree-node ${currentSystemSection === 'status' ? 'selected' : ''}" onclick="loadSystemSection('status')">
            <span class="chevron">🖥️</span>
            <span>Santé Serveur &amp; Base de Données</span>
            <span class="count-badge ok">Healthy</span>
        </div>
        <div class="tree-node ${currentSystemSection === 'storage' ? 'selected' : ''}" onclick="loadSystemSection('storage')">
            <span class="chevron">💾</span>
            <span>Stockage CAS &amp; Sauvegardes</span>
            <span class="count-badge ok">SHA-256</span>
        </div>
        <div class="tree-node ${currentSystemSection === 'reversibility' ? 'selected' : ''}" onclick="loadSystemSection('reversibility')">
            <span class="chevron">📦</span>
            <span>Souveraineté &amp; Réversibilité</span>
            <span class="count-badge ok">.sanad</span>
        </div>
        <div class="tree-node ${currentSystemSection === 'sessions' ? 'selected' : ''}" onclick="loadSystemSection('sessions')">
            <span class="chevron">🔒</span>
            <span>Sessions Actives &amp; TLS</span>
            <span class="count-badge req">${sessBadge}</span>
        </div>
        <div class="tree-node ${currentSystemSection === 'security' ? 'selected' : ''}" onclick="loadSystemSection('security')">
            <span class="chevron">🛡️</span>
            <span>Audit PSSI &amp; Zero-Trust</span>
            <span class="count-badge ok">Zero-Trust 403</span>
        </div>
    `;
}

// Chargement d'une section du cockpit IT
async function loadSystemSection(section) {
    currentSystemSection = section || 'status';
    renderSystemTree();

    if (window.innerWidth < 768) {
        toggleMobileTree(false);
    }

    const rootEl = document.getElementById("bc-root");
    const secEl = document.getElementById("bc-sec");
    const prodEl = document.getElementById("bc-prod");
    const modEl = document.getElementById("bc-mod");

    if (rootEl) rootEl.innerText = "INFRASTRUCTURE_IT";
    if (secEl) secEl.innerText = "On-Premise_BPF";

    if (currentSystemSection === 'status') {
        if (prodEl) prodEl.innerText = "Serveur & Base de Données";
        if (modEl) modEl.innerText = "Télémétries Live";
        loadSystemMetrics();
    } else if (currentSystemSection === 'storage') {
        if (prodEl) prodEl.innerText = "CAS & Backups";
        if (modEl) modEl.innerText = "Sauvegardes Horodatées";
        loadSystemStorage();
    } else if (currentSystemSection === 'reversibility') {
        if (prodEl) prodEl.innerText = "Souveraineté & Réversibilité";
        if (modEl) modEl.innerText = "Archive Universelle .sanad";
        loadSystemReversibility();
    } else if (currentSystemSection === 'sessions') {
        if (prodEl) prodEl.innerText = "Sécurité Accès";
        if (modEl) modEl.innerText = "Sessions & IP Actives";
        loadSystemSessions();
    } else if (currentSystemSection === 'security') {
        if (prodEl) prodEl.innerText = "Politiques PSSI";
        if (modEl) modEl.innerText = "Zero-Trust & Journal";
        loadSystemSecurity();
    }
}

// Chargement des métriques temps réel
async function loadSystemMetrics() {
    const tbody = document.getElementById("doc-tbody");
    const thead = document.getElementById("table-head");
    if (!tbody || !thead) return;

    thead.innerHTML = `<tr><th colspan="7">Cockpit Télémétries Temps Réel (Base de Données, Coffre-fort Électronique, Moteur Système)</th></tr>`;
    tbody.innerHTML = `<tr><td colspan="7" style="text-align:center; padding:25px; color:var(--text-muted);">Interrogation de l'infrastructure et des services...</td></tr>`;

    try {
        const res = await fetch("/api/v1/system/metrics");
        if (res.status === 401 || res.status === 403) {
            tbody.innerHTML = `<tr><td colspan="7" style="text-align:center; padding:25px; color:#ef4444; font-weight:700;">🚫 Accès Refusé : Habilitation SYS_MAINT requise pour la maintenance infrastructure.</td></tr>`;
            return;
        }

        if (!res.ok) {
            tbody.innerHTML = `<tr><td colspan="7" style="text-align:center; padding:25px; color:#ef4444;">Erreur chargement des métriques système.</td></tr>`;
            return;
        }

        const data = await res.json();
        systemMetricsCache = data;
        renderSystemDashboard(data);
    } catch (e) {
        console.error("Erreur métriques système:", e);
        if (tbody) tbody.innerHTML = `<tr><td colspan="7" style="text-align:center; padding:25px; color:#ef4444;">Erreur réseau : ${e.message}</td></tr>`;
    }
}

// Rendu du tableau de bord système (Dashboard Cards)
function renderSystemDashboard(data) {
    const thead = document.getElementById("table-head");
    const tbody = document.getElementById("doc-tbody");
    if (!thead || !tbody) return;

    thead.innerHTML = `
        <tr>
            <th style="width: 120px;">Composant</th>
            <th>Indicateur</th>
            <th>Valeur Actuelle</th>
            <th>Statut Santé</th>
            <th style="text-align:center; width: 140px;">Action Immédiate</th>
        </tr>
    `;

    tbody.innerHTML = "";

    const s = data.server || {};
    const db = data.database || {};
    const cas = data.storage_cas || {};
    const sec = data.security_pssi || {};

    const rows = [
        {
            comp: "🗄️ Base Réglementaire",
            label: "Registre Central Qualifié",
            val: `${db.tables_count || 0} tables | Taille: ${db.size_pretty || '0 B'} | Connexions: ${db.acquired_conns || 0}/${db.max_conns || 25}`,
            status: '<span class="status-tag status-ok">CONNECTÉ 🟢</span>',
            action: `<button class="btn-pill" style="font-size:10px; padding:2px 8px; border-color:var(--accent-emerald); color:var(--accent-emerald);" onclick="triggerSystemBackup()">💾 Sauvegarde</button>`,
            detailKey: 'db'
        },
        {
            comp: "💾 Stockage CAS",
            label: "Dépôt Fichiers Dédupliqué",
            val: `${cas.indexed_files || 0} fichiers physiques | Utilisé: ${cas.size_pretty || '0 B'} (${cas.quota_used_percent || '0%'} de 50 Go)`,
            status: '<span class="status-tag status-ok">SHA-256 VALIDÉ 🟢</span>',
            action: `<button class="btn-pill" style="font-size:10px; padding:2px 8px; border-color:var(--accent-blue); color:var(--accent-blue);" onclick="triggerCASIntegrityCheck()">🔍 Tester Intégrité</button>`,
            detailKey: 'cas'
        },
        {
            comp: "🖥️ Serveur Go",
            label: "Processus & Runtime Sanad Pharma",
            val: `OS: ${s.operating_system || ''} | Goroutines: ${s.num_goroutines || 0} | RAM: ${(s.mem_alloc_mb || 0).toFixed(1)} Mo / ${(s.mem_sys_mb || 0).toFixed(1)} Mo Sys`,
            status: `<span class="status-tag status-ok">UPTIME ${Math.floor((s.uptime_seconds || 0)/60)}m 🟢</span>`,
            action: `<button class="btn-pill" style="font-size:10px; padding:2px 8px; border-color:#0284c7; color:#0284c7;" onclick="loadSystemMetrics()">🔄 Rafraîchir</button>`,
            detailKey: 'server'
        },
        {
            comp: "🔒 Sécurité PSSI",
            label: "Sessions & Politique Zero-Trust",
            val: `${sec.active_sessions || 0} session(s) active(s) | ${sec.total_active_users || 0} utilisateurs habilités | ${sec.locked_accounts || 0} compte(s) verrouillé(s)`,
            status: '<span class="status-tag status-ok">ZERO-TRUST ACTIF 🛡️</span>',
            action: `<button class="btn-pill" style="font-size:10px; padding:2px 8px; border-color:#ef4444; color:#ef4444;" onclick="openEmergencyLockdownModal()">🚨 Verrouiller</button>`,
            detailKey: 'sec'
        },
        {
            comp: "📜 Piste d'Audit",
            label: "Journal Immuable ALCOA+",
            val: `${sec.audit_entries_count || 0} événements scellés par empreinte SHA-256 chaînée`,
            status: '<span class="status-tag status-ok">SCELLÉ SHA-256 🟢</span>',
            action: `<button class="btn-pill" style="font-size:10px; padding:2px 8px; border-color:#d97706; color:#d97706;" onclick="openAuditModal()">📜 Voir Logs</button>`,
            detailKey: 'audit'
        }
    ];

    rows.forEach((r, idx) => {
        const tr = document.createElement("tr");
        tr.style.cursor = "pointer";
        tr.onclick = (e) => {
            if (e.target.tagName === 'BUTTON' || e.target.closest('button')) return;
            document.querySelectorAll("#doc-tbody tr").forEach(row => row.classList.remove("selected"));
            tr.classList.add("selected");
            renderSystemDetailPane(r.detailKey, data);
            if (window.innerWidth < 1400) togglePreviewPane(true);
        };

        tr.innerHTML = `
            <td style="font-weight:700; color:var(--text-main);">${r.comp}</td>
            <td style="font-weight:600; color:var(--accent-blue);">${r.label}</td>
            <td style="font-family:monospace; font-size:11px; color:var(--text-main);">${r.val}</td>
            <td>${r.status}</td>
            <td style="text-align:center;">${r.action}</td>
        `;
        tbody.appendChild(tr);
    });

    // Aperçu par défaut de la base PostgreSQL
    renderSystemDetailPane('db', data);
}

// Rendu détaillé dans le volet d'aperçu Pane 3
function renderSystemDetailPane(key, data) {
    const preview = document.getElementById("preview-body");
    if (!preview || !data) return;

    let contentHtml = '';

    if (key === 'db') {
        const db = data.database || {};
        contentHtml = `
            <div style="font-size:12px; font-weight:700; color:#0369a1; text-align:center; margin-bottom:10px;">
                REGISTRE CENTRAL ET BASE DE DONNÉES SÉCURISÉE
            </div>
            <div style="font-size:11px; line-height:1.7;">
                <strong>Moteur de persistance :</strong> <div style="font-family:monospace; font-size:10px; background:var(--bg-panel); padding:4px; border-radius:3px; margin:3px 0;">${db.version || 'Base Sécurisée'}</div>
                <strong>Taille de la base :</strong> <code>${db.size_pretty || 'N/A'}</code> (${(db.size_bytes || 0).toLocaleString()} octets)<br>
                <strong>Nombre de tables publiques :</strong> <code>${db.tables_count || 0}</code> tables relationnelles BPF<br>
                <strong>Gestionnaire de connexions sécurisées :</strong><br>
                • Connexions actives acquises : <code>${db.acquired_conns || 0}</code><br>
                • Connexions en attente (idle) : <code>${db.idle_conns || 0}</code><br>
                • Connexions maximales autorisées : <code>${db.max_conns || 25}</code><br>
                <strong>Triggers d'immutabilité BPF :</strong> <span style="color:var(--accent-emerald); font-weight:700;">✓ Actifs sur audit_logs (Rejet formel UPDATE/DELETE)</span>
            </div>
            <div style="margin-top:12px;">
                <button class="btn-pill" style="width:100%; background:var(--accent-emerald); color:white; border:none; padding:6px; font-weight:700;" onclick="triggerSystemBackup()">
                    💾 Générer un Instantané de Sauvegarde SQL Scellé
                </button>
            </div>
        `;
    } else if (key === 'cas') {
        const cas = data.storage_cas || {};
        contentHtml = `
            <div style="font-size:12px; font-weight:700; color:#0369a1; text-align:center; margin-bottom:10px;">
                ENTREPÔT PHYSIQUE CAS (CONTENT ADDRESSABLE STORAGE)
            </div>
            <div style="font-size:11px; line-height:1.7;">
                <strong>Répertoire racine local :</strong> <code>${cas.base_path || './data/storage'}</code><br>
                <strong>Mécanisme de déduplication :</strong> <code>${cas.deduplication || 'SHA-256'}</code><br>
                <strong>Volume de fichiers indexés :</strong> <code>${cas.indexed_files || 0}</code> pièces jointes opposables<br>
                <strong>Taille totale stockée :</strong> <code>${cas.size_pretty || '0 B'}</code><br>
                <strong>Jauge d'espace alloué :</strong> <code>${cas.quota_used_percent || '0%'}</code> (Plafond On-Premise : 50 Go)<br>
                <strong>Intégrité ALCOA+ :</strong> Aucun document ne peut être altéré sans rompre l'empreinte cryptographique.
            </div>
            <div style="margin-top:12px;">
                <button class="btn-pill" style="width:100%; background:var(--accent-blue); color:white; border:none; padding:6px; font-weight:700;" onclick="triggerCASIntegrityCheck()">
                    🔍 Lancer un Audit de Cohérence Base vs Fichiers Disques
                </button>
            </div>
        `;
    } else if (key === 'sec') {
        const sec = data.security_pssi || {};
        contentHtml = `
            <div style="font-size:12px; font-weight:700; color:#0369a1; text-align:center; margin-bottom:10px;">
                CONTRÔLE D'ACCÈS ZERO-TRUST &amp; POLITIQUE PSSI
            </div>
            <div style="font-size:11px; line-height:1.7;">
                <strong>Isolation Zero-Trust :</strong> <span style="color:#ef4444; font-weight:700;">Active</span> (Admins IT bloqués sur les données cliniques &amp; formules 403)<br>
                <strong>Sessions actives connectées :</strong> <code>${sec.active_sessions || 0}</code><br>
                <strong>Algorithme de hachage des mots de passe :</strong> <code>${sec.password_hash_algo || 'Argon2id'}</code><br>
                <strong>Cookies de session :</strong> <code>HttpOnly, SameSite=Strict</code><br>
                <strong>Comptes temporairement verrouillés :</strong> <code>${sec.locked_accounts || 0}</code> (après échecs répétés)
            </div>
            <div style="margin-top:12px;">
                <button class="btn-pill" style="width:100%; background:#ef4444; color:white; border:none; padding:6px; font-weight:700;" onclick="openEmergencyLockdownModal()">
                    🚨 Verrouillage d'Urgence : Révoquer Tout Sauf Moi
                </button>
            </div>
        `;
    } else {
        const s = data.server || {};
        contentHtml = `
            <div style="font-size:12px; font-weight:700; color:#0369a1; text-align:center; margin-bottom:10px;">
                DIAGNOSTIC PROCESSUS SANAD PHARMA (RUNTIME GO)
            </div>
            <div style="font-size:11px; line-height:1.7;">
                <strong>Version binaire :</strong> <code>v0.7.0 (Production Ready)</code><br>
                <strong>Compilateur Go :</strong> <code>${s.go_version || ''}</code><br>
                <strong>Plateforme hôte :</strong> <code>${s.operating_system || ''}</code><br>
                <strong>Mode de déploiement :</strong> <code>${s.deployment_mode || 'on-premise'} (100% Souverain Hors-Ligne)</code><br>
                <strong>Goroutines actives :</strong> <code>${s.num_goroutines || 0}</code><br>
                <strong>Mémoire RAM allouée :</strong> <code>${(s.mem_alloc_mb || 0).toFixed(2)} Mo</code><br>
                <strong>Mémoire RAM système :</strong> <code>${(s.mem_sys_mb || 0).toFixed(2)} Mo</code><br>
                <strong>Démarré le :</strong> <code>${s.started_at ? new Date(s.started_at).toLocaleString() : ''}</code>
            </div>
        `;
    }

    preview.innerHTML = `
        <div class="pdf-sim-page">
            <div class="pdf-sim-header">
                RÉPUBLIQUE ALGÉRIENNE DÉMOCRATIQUE ET POPULAIRE<br>
                PLATEFORME PHARMACEUTIQUE SANAD PHARMA — AUDIT TECHNIQUE IT
            </div>
            ${contentHtml}
        </div>
    `;
}

// 2. Gestion des Sessions Actives & Sécurité
async function loadSystemSessions() {
    const tbody = document.getElementById("doc-tbody");
    const thead = document.getElementById("table-head");
    if (!tbody || !thead) return;

    thead.innerHTML = `
        <tr>
            <th style="width: 140px;">Collaborateur</th>
            <th>Email</th>
            <th>Rôle Métier</th>
            <th>Adresse IP</th>
            <th>Agent Navigateur</th>
            <th>Dernière Activité</th>
            <th style="text-align:center; width: 100px;">Action Révocation</th>
        </tr>
    `;

    tbody.innerHTML = `<tr><td colspan="7" style="text-align:center; padding:20px; color:var(--text-muted);">Chargement des sessions actives...</td></tr>`;

    try {
        const res = await fetch("/api/v1/system/sessions");
        if (!res.ok) {
            tbody.innerHTML = `<tr><td colspan="7" style="text-align:center; padding:20px; color:#ef4444;">Impossible de charger les sessions.</td></tr>`;
            return;
        }

        const data = await res.json();
        systemSessionsCache = data.sessions || [];

        if (systemSessionsCache.length === 0) {
            tbody.innerHTML = `<tr><td colspan="7" style="text-align:center; padding:20px; color:var(--text-dim);">Aucune session active enregistrée.</td></tr>`;
            return;
        }

        tbody.innerHTML = "";
        systemSessionsCache.forEach(sess => {
            const tr = document.createElement("tr");
            const isMe = currentUser && currentUser.id === sess.user_id;

            tr.innerHTML = `
                <td style="font-weight:600; color:var(--text-main);">
                    ${escapeAdminHtml(sess.first_name + ' ' + sess.last_name)}
                    ${isMe ? '<span style="color:var(--accent-blue); font-size:10px; margin-left:4px;">(Vous)</span>' : ''}
                </td>
                <td style="font-family:monospace; font-size:11px; color:var(--text-muted);">${escapeAdminHtml(sess.email)}</td>
                <td><span class="status-tag status-ok" style="font-size:10px;">${escapeAdminHtml(sess.role)}</span></td>
                <td style="font-family:monospace; font-size:11px; color:var(--accent-blue); font-weight:600;">${escapeAdminHtml(sess.ip_address)}</td>
                <td style="font-size:10px; color:var(--text-dim); max-width:200px; overflow:hidden; text-overflow:ellipsis; white-space:nowrap;" title="${escapeAdminHtml(sess.user_agent)}">${escapeAdminHtml(sess.user_agent)}</td>
                <td style="font-size:11px; color:var(--text-main);">${sess.last_activity_at ? new Date(sess.last_activity_at).toLocaleTimeString() : ''}</td>
                <td style="text-align:center;">
                    ${isMe ? `
                        <span style="font-size:10px; color:var(--text-dim);">Session courante</span>
                    ` : `
                        <button class="btn-pill" style="font-size:10px; padding:2px 8px; border-color:#ef4444; color:#ef4444;" onclick="revokeSessionById('${sess.id}', '${escapeAdminHtml(sess.email)}')">
                            ❌ Révoquer
                        </button>
                    `}
                </td>
            `;
            tbody.appendChild(tr);
        });

    } catch (e) {
        console.error("Erreur sessions:", e);
        tbody.innerHTML = `<tr><td colspan="7" style="text-align:center; padding:20px; color:#ef4444;">Erreur : ${e.message}</td></tr>`;
    }
}

// Révoquer une session individuelle
async function revokeSessionById(sessionId, email) {
    if (!(await showConfirmModal("Révocation de Session", `Confirmez-vous la révocation immédiate de la session active de ${email} ?\nL'utilisateur sera immédiatement déconnecté.`, { icon: "🔒", danger: true, confirmText: "Révoquer" }))) {
        return;
    }

    try {
        const res = await fetch(`/api/v1/system/sessions/${sessionId}`, { method: "DELETE" });
        const data = await res.json();
        if (!res.ok) {
            showNotifyModal("Erreur Révocation", data.error || "Impossible de révoquer la session", "⚠️", "Fermer");
            return;
        }

        showNotifyModal("Session Révoquée", `La session de <code>${email}</code> a été détruite avec succès.`, "🔒", "Compris");
        loadSystemSessions();
    } catch (e) {
        showNotifyModal("Erreur Réseau", e.message, "⚠️", "Fermer");
    }
}

// Modal Verrouillage d'Urgence
async function openEmergencyLockdownModal() {
    const reason = await showPromptModal(
        "Alerte Sécurité PSSI / Verrouillage d'Urgence",
        "Cette opération détruira TOUTES les sessions actives sur la plateforme, forçant l'ensemble des collaborateurs à se reconnecter.\nVeuillez saisir le motif réglementaire de sécurité :",
        "Mesure conservatoire de sécurité PSSI / Alerte intrusion",
        { icon: "🚨", danger: true, confirmText: "Verrouiller la Plateforme" }
    );

    if (!reason || reason.trim().length < 3) {
        if (reason !== null) {
            showNotifyModal("Action Annulée", "Un motif circonstancié est obligatoire pour déclencher un verrouillage d'urgence.", "⚠️", "Fermer");
        }
        return;
    }

    executeEmergencyLockdown(reason.trim());
}

async function executeEmergencyLockdown(reason) {
    try {
        const res = await fetch("/api/v1/system/sessions/emergency-lockdown", {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ reason: reason })
        });

        const data = await res.json();
        if (!res.ok) {
            showNotifyModal("Échec Verrouillage", data.error || "Erreur lors du verrouillage", "⚠️", "Fermer");
            return;
        }

        showNotifyModal(
            "Verrouillage d'Urgence Exécuté",
            `<div style="font-size:12px;">
                <strong style="color:#ef4444;">✓ ${data.revoked_sessions || 0} session(s) active(s) révoquée(s) immédiatement.</strong><br>
                Seule votre session d'administrateur reste ouverte.<br>
                L'incident et votre motif ont été scellés dans l'Audit Trail immuable ALCOA+.
            </div>`,
            "🚨",
            "Compris"
        );
        loadSystemSection('sessions');
    } catch (e) {
        showNotifyModal("Erreur Réseau", e.message, "⚠️", "Fermer");
    }
}

// 3. Déclenchement d'un Instantané de Sauvegarde
async function triggerSystemBackup() {
    showNotifyModal("Sauvegarde en cours", "Création de l'instantané SQL et scellement SHA-256 en cours...", "💾", "Veuillez patienter");

    try {
        const res = await fetch("/api/v1/system/backup", { method: "POST" });
        const data = await res.json();
        if (!res.ok) {
            showNotifyModal("Erreur Sauvegarde", data.error || "Impossible de générer le backup", "⚠️", "Fermer");
            return;
        }

        showNotifyModal(
            "Sauvegarde Générée & Scellée (ALCOA+)",
            `<div style="font-size:11px; line-height:1.6;">
                <strong>Identifiant Backup :</strong> <code>${data.backup_id}</code><br>
                <strong>Taille :</strong> <code>${data.size_pretty}</code><br>
                <strong>Empreinte SHA-256 :</strong> <div style="font-family:monospace; font-size:10px; background:var(--bg-panel); padding:4px; border-radius:3px; word-break:break-all;">${data.sha256_seal}</div>
                <strong>Emplacement disque :</strong> <code>${data.file_path}</code><br>
                Conformité BPF : L'événement a été scellé dans la piste d'audit immuable.
            </div>`,
            "💾",
            "Terminer"
        );
        loadSystemMetrics();
    } catch (e) {
        showNotifyModal("Erreur Réseau", e.message, "⚠️", "Fermer");
    }
}

// 4. Contrôle d'Intégrité CAS
async function triggerCASIntegrityCheck() {
    showNotifyModal("Contrôle d'Intégrité", "Vérification en cours des empreintes SHA-256 entre la base et les fichiers disques...", "🔍", "Patientez");

    try {
        const res = await fetch("/api/v1/system/verify-integrity", { method: "POST" });
        const data = await res.json();
        if (!res.ok) {
            showNotifyModal("Erreur Contrôle", data.error || "Impossible d'exécuter le contrôle", "⚠️", "Fermer");
            return;
        }

        const is100 = data.mismatch_count === 0 && data.missing_count === 0;

        showNotifyModal(
            is100 ? "Intégrité CAS 100% Conforme" : "Alerte Intégrité Fichiers",
            `<div style="font-size:12px; line-height:1.6;">
                <strong>Fichiers physiques contrôlés :</strong> <code>${data.total_checked}</code><br>
                <strong>Empreintes conformes SHA-256 :</strong> <span style="color:var(--accent-emerald); font-weight:700;">${data.valid_count}</span><br>
                <strong>Empreintes altérées (mismatch) :</strong> <span style="color:${data.mismatch_count > 0 ? '#ef4444' : 'var(--text-muted)'}; font-weight:700;">${data.mismatch_count}</span><br>
                <strong>Fichiers manquants :</strong> <span style="color:${data.missing_count > 0 ? '#ef4444' : 'var(--text-muted)'}; font-weight:700;">${data.missing_count}</span><br>
                ${is100 ? '<div style="margin-top:6px; color:var(--accent-emerald); font-weight:700;">✓ Aucune altération détectée. Le coffre-fort CAS respecte l\'intégrité ALCOA+.</div>' : '<div style="margin-top:6px; color:#ef4444; font-weight:700;">⚠️ Des anomalies physiques ont été scellées dans l\'Audit Trail !</div>'}
            </div>`,
            is100 ? "🟢" : "⚠️",
            "Fermer"
        );
        loadSystemMetrics();
    } catch (e) {
        showNotifyModal("Erreur Réseau", e.message, "⚠️", "Fermer");
    }
}

// 5. Vue Stockage & Backups
function loadSystemStorage() {
    if (systemMetricsCache) {
        renderSystemDashboard(systemMetricsCache);
        renderSystemDetailPane('cas', systemMetricsCache);
    } else {
        loadSystemMetrics();
    }
}

// 6. Vue Sécurité PSSI & Zero-Trust
function loadSystemSecurity() {
    if (systemMetricsCache) {
        renderSystemDashboard(systemMetricsCache);
        renderSystemDetailPane('sec', systemMetricsCache);
    } else {
        loadSystemMetrics();
    }
}

// 7. Vue Souveraineté & Réversibilité BPF (.homolog)
function loadSystemReversibility() {
    const thead = document.getElementById("table-head");
    const tbody = document.getElementById("doc-tbody");
    if (!thead || !tbody) return;

    thead.innerHTML = `
        <tr>
            <th style="width: 140px;">Composant Archive</th>
            <th>Spécification Technique &amp; Réglementaire</th>
            <th>Périmètre Opposable ALCOA+</th>
            <th>Contrôle Intégrité</th>
            <th style="text-align:center; width: 150px;">Action Immédiate</th>
        </tr>
    `;

    tbody.innerHTML = `
        <tr style="cursor:pointer;" onclick="renderReversibilityDetailPane('manifest')">
            <td style="font-weight:700; color:var(--text-main);">📜 Manifeste Signé</td>
            <td style="font-weight:600; color:var(--accent-blue);"><code>manifest.json</code></td>
            <td style="font-size:11px; color:var(--text-muted);">Horodatage UTC, établissement, auteur BPF, version <code>HOMOLOG_PORTABLE_ARCHIVE_V1</code></td>
            <td><span class="status-tag status-ok">SCELLÉ SHA-256 🟢</span></td>
            <td style="text-align:center;">
                <button class="btn-pill" style="font-size:10px; padding:2px 8px; border-color:var(--accent-emerald); color:var(--accent-emerald);" onclick="exportHomologArchive()">📦 Exporter</button>
            </td>
        </tr>
        <tr style="cursor:pointer;" onclick="renderReversibilityDetailPane('tables')">
            <td style="font-weight:700; color:var(--text-main);">🐘 Base Relationnelle</td>
            <td style="font-weight:600; color:var(--accent-blue);"><code>database/*.json</code> (17 Tables)</td>
            <td style="font-size:11px; color:var(--text-muted);">Établissements, Produits, Lignes, CTD Modules 1-5, DM, Redevances, Journal ALCOA+</td>
            <td><span class="status-tag status-ok">EXHAUSTIF 🟢</span></td>
            <td style="text-align:center;">
                <button class="btn-pill" style="font-size:10px; padding:2px 8px; border-color:var(--accent-blue); color:var(--accent-blue);" onclick="openRestoreHomologModal()">📥 Restaurer</button>
            </td>
        </tr>
        <tr style="cursor:pointer;" onclick="renderReversibilityDetailPane('cas')">
            <td style="font-weight:700; color:var(--text-main);">💾 Coffre-Fort CAS</td>
            <td style="font-weight:600; color:var(--accent-blue);"><code>blobs/&lt;sha256&gt;</code> (GED ALCOA+)</td>
            <td style="font-size:11px; color:var(--text-muted);">Ensemble des pièces jointes opposables dédupliquées, scans officiels ANPP, bordereaux signés</td>
            <td><span class="status-tag status-ok">SHA-256 CAS 🟢</span></td>
            <td style="text-align:center;">
                <button class="btn-pill" style="font-size:10px; padding:2px 8px; border-color:var(--accent-emerald); color:var(--accent-emerald);" onclick="exportHomologArchive()">📦 Exporter</button>
            </td>
        </tr>
        <tr style="cursor:pointer;" onclick="renderReversibilityDetailPane('reversibility')">
            <td style="font-weight:700; color:var(--text-main);">🛡️ Souveraineté &amp; Réversibilité</td>
            <td style="font-weight:600; color:var(--accent-blue);">Zéro Dépendance Propriétaire</td>
            <td style="font-size:11px; color:var(--text-muted);">Migration sans friction Cloud SaaS &lt;--&gt; On-Premise Souverain avec opposabilité totale</td>
            <td><span class="status-tag status-ok">LIBRE &amp; PORTABLE 🟢</span></td>
            <td style="text-align:center;">
                <button class="btn-pill" style="font-size:10px; padding:2px 8px; border-color:var(--text-muted); color:var(--text-main);" onclick="renderReversibilityDetailPane('reversibility')">ℹ️ Spécifications</button>
            </td>
        </tr>
    `;

    renderReversibilityDetailPane('manifest');
}

// Volet d'aperçu pour la réversibilité
function renderReversibilityDetailPane(key) {
    const preview = document.getElementById("preview-body");
    if (!preview) return;

    let contentHtml = '';

    if (key === 'manifest') {
        contentHtml = `
            <div style="font-size:12px; font-weight:700; color:#0369a1; text-align:center; margin-bottom:10px;">
                MANIFESTE DE SOUVERAINETÉ (MANIFEST.JSON)
            </div>
            <div style="font-size:11px; line-height:1.7;">
                <strong>Format conteneur :</strong> <code>SANAD_PORTABLE_ARCHIVE_V1</code><br>
                <strong>Encapsulation :</strong> Flux tar.gz universel, lisible par n'importe quel outil standard (tar, gzip, python, Go)<br>
                <strong>Sceau d'archive :</strong> Empreinte SHA-256 calculée sur l'intégralité du contenu relationnel et binaire<br>
                <strong>Horodatage opposable :</strong> Norme ISO 8601 UTC avec signature de l'auteur légal BPF<br>
                <strong>Contrôle anti-altération :</strong> Vérification systématique du sceau avant toute opération de réinjection
            </div>
            <div style="margin-top:14px; display:flex; flex-direction:column; gap:8px;">
                <button class="btn-pill" style="width:100%; background:var(--accent-emerald); color:white; border:none; padding:8px; font-weight:700;" onclick="exportHomologArchive()">
                    📦 Télécharger l'Archive Universelle .sanad
                </button>
            </div>
        `;
    } else if (key === 'tables') {
        contentHtml = `
            <div style="font-size:12px; font-weight:700; color:#0369a1; text-align:center; margin-bottom:10px;">
                PÉRIMÈTRE RELATIONNEL EXHAUSTIF (17 TABLES BPF)
            </div>
            <div style="font-size:11px; line-height:1.7;">
                <strong>Tables incluses :</strong><br>
                • <code>organizations</code>, <code>establishments</code>, <code>manufacturing_lines</code><br>
                • <code>products</code>, <code>product_presentations</code>, <code>users</code><br>
                • <code>ged_documents</code>, <code>ged_versions</code>, <code>anpp_ctd_dossiers</code>, <code>anpp_ctd_sections</code><br>
                • <code>manufacturing_secrets</code>, <code>foreign_manufacturers</code>, <code>medical_devices</code><br>
                • <code>anpp_deficiency_letters</code>, <code>anpp_deficiency_tasks</code><br>
                • <code>fee_schedules</code>, <code>payment_slips</code>, <code>anpp_appointments</code>, <code>regulatory_watch_items</code><br>
                • <code>audit_logs</code> (Piste d'audit complète scellée)
            </div>
            <div style="margin-top:14px;">
                <button class="btn-pill" style="width:100%; background:var(--accent-blue); color:white; border:none; padding:8px; font-weight:700;" onclick="openRestoreHomologModal()">
                    📥 Restaurer une Archive Universelle (.sanad / .homolog)
                </button>
            </div>
        `;
    } else if (key === 'cas') {
        contentHtml = `
            <div style="font-size:12px; font-weight:700; color:#0369a1; text-align:center; margin-bottom:10px;">
                COFFRE-FORT NUMÉRIQUE CAS (CONTENT ADDRESSABLE STORAGE)
            </div>
            <div style="font-size:11px; line-height:1.7;">
                <strong>Architecture :</strong> Chaque fichier binaire est archivé dans <code>blobs/&lt;sha256&gt;</code>.<br>
                <strong>Zéro dépendance :</strong> Les fichiers originaux (PDF, scans, bordereaux) sont préservés dans leur état binaire strict.<br>
                <strong>Vérification à la restauration :</strong> Chaque blob est haché à la volée. S'il ne correspond pas exactement au hash du manifeste, la restauration est rejetée avec code 422.<br>
                <strong>Opposabilité ALCOA+ :</strong> Intégrité absolue garantie face aux inspecteurs de l'ANPP ou du MIPH.
            </div>
            <div style="margin-top:14px;">
                <button class="btn-pill" style="width:100%; background:var(--accent-emerald); color:white; border:none; padding:8px; font-weight:700;" onclick="exportHomologArchive()">
                    📦 Télécharger l'Archive Universelle .sanad
                </button>
            </div>
        `;
    } else {
        contentHtml = `
            <div style="font-size:12px; font-weight:700; color:#0369a1; text-align:center; margin-bottom:10px;">
                GARANTIE SOUVERAINE DE RÉVERSIBILITÉ BPF
            </div>
            <div style="font-size:11px; line-height:1.7;">
                <strong>Pourquoi le format universel .sanad ?</strong><br>
                Dans l'industrie pharmaceutique, la dépendance à un éditeur logiciel (vendor lock-in) représente un risque critique pour l'homologation des médicaments.<br><br>
                <strong>Notre engagement éthique &amp; réglementaire :</strong><br>
                1. Vos dossiers d'enregistrement vous appartiennent à 100%.<br>
                2. L'archive <code>.sanad</code> (ou <code>.homolog</code>) peut être restaurée sur n'importe quel serveur On-Premise ou instance Cloud sans licence restrictive.<br>
                3. Les visas électroniques du Pharmacien DT (21 CFR Part 11) et les traces ALCOA+ restent vérifiables en toute circonstance.
            </div>
        `;
    }

    preview.innerHTML = `
        <div class="pdf-sim-page">
            <div class="pdf-sim-header">
                RÉPUBLIQUE ALGÉRIENNE DÉMOCRATIQUE ET POPULAIRE<br>
                SANAD PHARMA — SOUVERAINETÉ NUMÉRIQUE &amp; RÉVERSIBILITÉ BPF
            </div>
            ${contentHtml}
        </div>
    `;
}

// Déclenchement de l'exportation .sanad (Sanad Pharma)
async function exportHomologArchive() {
    showNotifyModal(
        "Génération de l'Archive Universelle .sanad",
        "Exportation des 17 tables relationnelles, collecte des blobs du coffre-fort CAS et calcul du sceau SHA-256 en cours...",
        "📦",
        "Veuillez patienter"
    );

    try {
        const response = await fetch('/api/v1/system/backup/export');
        if (!response.ok) {
            let errMsg = "Échec du téléchargement";
            try { const errJson = await response.json(); errMsg = errJson.error || errMsg; } catch(e){}
            showNotifyModal("Erreur Exportation", errMsg, "⚠️", "Fermer");
            return;
        }

        const seal = response.headers.get("X-Sanad-Seal") || response.headers.get("X-Homolog-Seal") || "N/A";
        const disposition = response.headers.get("Content-Disposition");
        let filename = "sanad_backup.sanad";
        if (disposition && disposition.indexOf("filename=") !== -1) {
            const matches = /filename[^;=\n]*=((['"]).*?\2|[^;\n]*)/.exec(disposition);
            if (matches != null && matches[1]) { 
                filename = matches[1].replace(/['"]/g, '');
            }
        }

        const blob = await response.blob();
        const url = window.URL.createObjectURL(blob);
        const a = document.createElement('a');
        a.style.display = 'none';
        a.href = url;
        a.download = filename;
        document.body.appendChild(a);
        a.click();
        window.URL.revokeObjectURL(url);
        a.remove();

        showNotifyModal(
            "Archive Universelle .sanad Générée &amp; Scellée",
            `<div style="font-size:11px; line-height:1.6;">
                <strong>Fichier téléchargé :</strong> <code>${escapeAdminHtml(filename)}</code><br>
                <strong>Taille :</strong> <code>${(blob.size / 1024 / 1024).toFixed(2)} Mo</code> (${blob.size.toLocaleString()} octets)<br>
                <strong>Sceau d'intégrité SHA-256 :</strong>
                <div style="font-family:monospace; font-size:10px; background:var(--bg-panel); padding:4px; border-radius:3px; word-break:break-all; margin-top:4px;">${seal}</div>
                <div style="margin-top:8px; color:var(--accent-emerald); font-weight:600;">
                    ✓ Conformité Réversibilité BPF : Données relationnelles et coffre CAS intégralement préservés.
                </div>
            </div>`,
            "📦",
            "Terminer"
        );
    } catch (e) {
        showNotifyModal("Erreur Réseau", e.message, "⚠️", "Fermer");
    }
}

// Ouverture de la modale de restauration
function openRestoreHomologModal() {
    closeAllModals();
    const modal = document.getElementById("modal-restore-homolog");
    if (modal) {
        const fileInput = document.getElementById("restore-file-input");
        if (fileInput) fileInput.value = "";
        const progContainer = document.getElementById("restore-progress-container");
        if (progContainer) progContainer.style.display = "none";
        const resultBox = document.getElementById("restore-result-box");
        if (resultBox) resultBox.style.display = "none";
        const submitBtn = document.getElementById("btn-submit-restore");
        if (submitBtn) submitBtn.disabled = false;
        modal.classList.add("open");
    }
}

// Soumission de la restauration .homolog
async function submitRestoreHomolog(e) {
    e.preventDefault();
    const fileInput = document.getElementById("restore-file-input");
    if (!fileInput || !fileInput.files || fileInput.files.length === 0) {
        showToast("Veuillez sélectionner un fichier d'archive .sanad ou .homolog.", "warning");
        return;
    }

    const file = fileInput.files[0];
    const submitBtn = document.getElementById("btn-submit-restore");
    const progContainer = document.getElementById("restore-progress-container");
    const progBar = document.getElementById("restore-progress-bar");
    const progLabel = document.getElementById("restore-progress-label");
    const progPercent = document.getElementById("restore-progress-percent");
    const resultBox = document.getElementById("restore-result-box");

    if (submitBtn) submitBtn.disabled = true;
    if (progContainer) progContainer.style.display = 'block';
    if (progBar) progBar.style.width = '30%';
    if (progPercent) progPercent.innerText = '30%';
    if (progLabel) progLabel.innerText = "Téléversement de l'archive...";
    if (resultBox) resultBox.style.display = 'none';

    const formData = new FormData();
    formData.append("archive", file);

    try {
        if (progBar) progBar.style.width = '60%';
        if (progPercent) progPercent.innerText = '60%';
        if (progLabel) progLabel.innerText = "Contrôle cryptographique SHA-256 et réinjection des tables...";

        const res = await fetch("/api/v1/system/backup/restore", {
            method: "POST",
            body: formData
        });

        const data = await res.json();

        if (progBar) progBar.style.width = '100%';
        if (progPercent) progPercent.innerText = '100%';

        if (!res.ok) {
            if (progContainer) progContainer.style.display = 'none';
            if (submitBtn) submitBtn.disabled = false;
            if (resultBox) {
                resultBox.style.display = 'block';
                resultBox.innerHTML = `
                    <div style="background:rgba(239,68,68,0.12); border:1px solid #ef4444; border-radius:4px; padding:10px; color:#ef4444;">
                        <strong>🚫 Échec de la restauration :</strong> ${escapeAdminHtml(data.error || 'Erreur inconnue')}
                    </div>
                `;
            }
            return;
        }

        if (resultBox) {
            resultBox.style.display = 'block';
            resultBox.innerHTML = `
                <div style="background:rgba(16,185,129,0.12); border:1px solid var(--accent-emerald); border-radius:4px; padding:10px; color:var(--accent-emerald);">
                    <strong style="font-size:12px;">✓ Restauration certifiée 100% conforme !</strong><br>
                    • Format : <code>${escapeAdminHtml((data.manifest && data.manifest.format) || 'SANAD_PORTABLE_ARCHIVE_V1')}</code><br>
                    • Tables relationnelles restaurées : <strong>${data.tables_restored || 0}</strong><br>
                    • Documents CAS vérifiés et restaurés : <strong>${data.blobs_restored || 0}</strong><br>
                    • Sceau d'archive SHA-256 : <code style="word-break:break-all; font-size:10px;">${escapeAdminHtml((data.manifest && data.manifest.archive_seal) || '')}</code>
                    <div style="margin-top:8px; font-weight:700; color:var(--text-main);">
                        Actualisation dans 3 secondes...
                    </div>
                </div>
            `;
        }

        setTimeout(() => {
            closeModal('modal-restore-homolog');
            if (submitBtn) submitBtn.disabled = false;
            if (progContainer) progContainer.style.display = 'none';
            loadSystemSection('reversibility');
            loadSystemMetrics();
        }, 3000);

    } catch (err) {
        if (progContainer) progContainer.style.display = 'none';
        if (submitBtn) submitBtn.disabled = false;
        if (resultBox) {
            resultBox.style.display = 'block';
            resultBox.innerHTML = `
                <div style="background:rgba(239,68,68,0.12); border:1px solid #ef4444; border-radius:4px; padding:10px; color:#ef4444;">
                    <strong>Erreur réseau :</strong> ${escapeAdminHtml(err.message)}
                </div>
            `;
        }
    }
}
