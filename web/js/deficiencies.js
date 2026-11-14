/**
 * Studio de Compléments ANPP & Traçabilité des Réponses Réglementaires
 * Conforme au Guide CTD ANPP 2025 et aux exigences BPF / 21 CFR Part 11
 */

let activeDeficienciesDashboard = null;
let currentDeficiencyLetterDetail = null;

if (typeof escapeHTML !== 'function') {
    function escapeHTML(str) {
        if (!str) return '';
        return String(str)
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#039;');
    }
}

// Chargement et affichage du tableau de bord global des compléments
async function loadDeficienciesDashboard() {
    try {
        const resp = await fetch('/api/v1/deficiencies/dashboard', {
            headers: { 'Accept': 'application/json' }
        });
        if (!resp.ok) {
            throw new Error(`Erreur HTTP ${resp.status}`);
        }
        const data = await resp.json();
        activeDeficienciesDashboard = data;
        renderDeficienciesDashboardView(data);
        updateDeficiencyBadgeCount(data);
    } catch (err) {
        console.error("Erreur chargement compléments:", err);
        const centerView = document.querySelector(".table-container");
        if (centerView) {
            centerView.innerHTML = `
                <div style="padding:30px; text-align:center; color:var(--accent-red);">
                    <h3>⚠️ Impossible de charger les compléments ANPP</h3>
                    <p style="font-size:12px; color:var(--text-muted);">${escapeHTML(err.message)}</p>
                    <button class="btn-pill" onclick="loadDeficienciesDashboard()" style="margin-top:10px;">🔄 Réessayer</button>
                </div>
            `;
        }
    }
}

// Mise à jour de la pastille de notification sur l'Activity Bar
function updateDeficiencyBadgeCount(data) {
    const badge = document.getElementById("badge-deficiencies");
    if (!badge) return;
    const criticalCount = data ? (data.critical_letters || 0) : 0;
    if (criticalCount > 0) {
        badge.innerText = criticalCount;
        badge.style.display = "inline-block";
        badge.style.background = "var(--accent-red)";
    } else {
        const total = data ? (data.total_active_letters || 0) : 0;
        if (total > 0) {
            badge.innerText = total;
            badge.style.display = "inline-block";
            badge.style.background = "var(--accent-blue)";
        } else {
            badge.style.display = "none";
        }
    }
}

// Badge visuel du compte à rebours légal (J-X)
function getUrgencyBadgeHTML(days, urgencyBadge) {
    if (days < 0) {
        return `<span class="badge" style="background:#fee2e2; color:#991b1b; font-weight:700; border:1px solid #f87171;">⛔ DÉPASSÉ (${Math.abs(days)}j)</span>`;
    }
    if (days <= 15) {
        return `<span class="badge" style="background:#fee2e2; color:#dc2626; font-weight:700; border:1px solid #fca5a5; animation:pulse 2s infinite;">🚨 CRITIQUE J-${days}</span>`;
    }
    if (days <= 30) {
        return `<span class="badge" style="background:#ffedd5; color:#c2410c; font-weight:700; border:1px solid #fdba74;">⚠️ URGENT J-${days}</span>`;
    }
    if (days <= 60) {
        return `<span class="badge" style="background:#fef3c7; color:#b45309; font-weight:600; border:1px solid #fde68a;">⏳ Alerte J-${days}</span>`;
    }
    return `<span class="badge" style="background:#ecfdf5; color:#047857; font-weight:600; border:1px solid #a7f3d0;">✅ Dans les délais (J-${days})</span>`;
}

// Rendu HTML du Cockpit Principal Studio de Compléments ANPP
function renderDeficienciesDashboardView(data) {
    const tableHead = document.getElementById("table-head");
    const centerView = document.querySelector(".table-container");
    if (!centerView) return;

    if (tableHead) tableHead.innerHTML = '';

    const letters = data.recent_letters || [];

    let html = `
        <div style="padding:16px; display:flex; flex-direction:column; gap:16px; overflow-y:auto; height:100%;">
            <!-- En-tête réglementaire et KPI cards -->
            <div style="background:var(--bg-panel); border:1px solid var(--border-color); border-radius:8px; padding:14px;">
                <div style="display:flex; justify-content:space-between; align-items:center; flex-wrap:wrap; gap:10px; margin-bottom:12px;">
                    <div>
                        <div style="font-size:15px; font-weight:700; color:var(--text-main); display:flex; align-items:center; gap:8px;">
                            <span>⏱️ Studio de Compléments ANPP & Compte à Rebours Légal</span>
                            <span style="font-size:11px; background:var(--badge-bg); color:var(--accent-blue); padding:2px 8px; border-radius:4px; font-weight:600;">Décret 20-325 & Guide CTD 2025</span>
                        </div>
                        <div style="font-size:11px; color:var(--text-dim); margin-top:2px;">
                            Traçabilité ALCOA+ des notifications officielles de l'ANPP, décomposition par question/module et visa légal Pharmacien Directeur Technique (1.A.7).
                        </div>
                    </div>
                    <div style="display:flex; gap:8px; align-items:center;">
                        <button class="btn-pill" style="background:var(--accent-blue); color:white; font-weight:700; border:none; padding:6px 14px; font-size:11.5px;" onclick="openNewDeficiencyLetterModal()">➕ Enregistrer une Notification de Compléments ANPP</button>
                        <button class="btn-pill" onclick="loadDeficienciesDashboard()" title="Actualiser les compteurs">🔄 Actualiser</button>
                    </div>
                </div>

                <!-- 4 KPI Cards -->
                <div style="display:grid; grid-template-columns:repeat(auto-fit, minmax(180px, 1fr)); gap:10px;">
                    <div class="admin-preview-card" style="border-left:4px solid #dc2626; padding:10px;">
                        <div style="font-size:10.5px; color:var(--text-dim); text-transform:uppercase; font-weight:700;">🚨 Échéances Critiques (J ≤ 15)</div>
                        <div style="font-size:22px; font-weight:800; color:#dc2626; margin-top:4px;">${data.critical_letters || 0}</div>
                        <div style="font-size:10px; color:var(--text-muted);">Risque de caducité du dossier</div>
                    </div>

                    <div class="admin-preview-card" style="border-left:4px solid #ea580c; padding:10px;">
                        <div style="font-size:10.5px; color:var(--text-dim); text-transform:uppercase; font-weight:700;">⚠️ Échéances Urgentes (J ≤ 30)</div>
                        <div style="font-size:22px; font-weight:800; color:#ea580c; margin-top:4px;">${data.urgent_letters || 0}</div>
                        <div style="font-size:10px; color:var(--text-muted);">Réponses en cours de finalisation</div>
                    </div>

                    <div class="admin-preview-card" style="border-left:4px solid var(--accent-blue); padding:10px;">
                        <div style="font-size:10.5px; color:var(--text-dim); text-transform:uppercase; font-weight:700;">📋 Questions à Répondre</div>
                        <div style="font-size:22px; font-weight:800; color:var(--accent-blue); margin-top:4px;">${data.pending_tasks_count || 0}</div>
                        <div style="font-size:10px; color:var(--text-muted);">Découpées par module CTD</div>
                    </div>

                    <div class="admin-preview-card" style="border-left:4px solid #16a34a; padding:10px;">
                        <div style="font-size:10.5px; color:var(--text-dim); text-transform:uppercase; font-weight:700;">✍️ Visées Pharmacien DT (1.A.7)</div>
                        <div style="font-size:22px; font-weight:800; color:#16a34a; margin-top:4px;">${data.validated_tasks_count || 0}</div>
                        <div style="font-size:10px; color:var(--text-muted);">Scellées et prêtes au dépôt</div>
                    </div>
                </div>
            </div>

            <!-- Tableau des Lettres Actives -->
            <div style="background:var(--bg-panel); border:1px solid var(--border-color); border-radius:8px; overflow:hidden;">
                <div style="padding:10px 14px; background:var(--bg-surface); border-bottom:1px solid var(--border-color); display:flex; justify-content:space-between; align-items:center;">
                    <span style="font-weight:700; font-size:12px; color:var(--text-main);">📋 Notifications Officielles de Compléments en Instance (${letters.length})</span>
                    <button class="btn-pill" style="background:var(--accent-blue); color:white; font-size:11px; font-weight:700; border:none; padding:4px 12px;" onclick="openNewDeficiencyLetterModal()">➕ Nouvelle Notification ANPP</button>
                </div>
    `;

    if (letters.length === 0) {
        html += `
            <div style="padding:40px; text-align:center; color:var(--text-muted);">
                <div style="font-size:32px; margin-bottom:8px;">🎉</div>
                <div style="font-weight:600; font-size:13px; color:var(--text-main);">Aucun complément ANPP en instance</div>
                <div style="font-size:11px; margin-top:4px; margin-bottom:16px;">Tous les dossiers d'enregistrement sont à jour ou en attente d'évaluation initiale.</div>
                <button class="btn-pill" style="background:var(--accent-blue); color:white; font-weight:700; border:none; padding:8px 18px; font-size:12px; cursor:pointer;" onclick="openNewDeficiencyLetterModal()">➕ Enregistrer une Notification de Compléments ANPP</button>
            </div>
        `;
    } else {
        html += `
            <div style="overflow-x:auto;">
                <table style="width:100%; border-collapse:collapse; font-size:11.5px; text-align:left;">
                    <thead>
                        <tr style="background:var(--bg-sub); border-bottom:1px solid var(--border-color); color:var(--text-dim); font-size:10.5px; text-transform:uppercase;">
                            <th style="padding:8px 12px;">Réf. Notification ANPP</th>
                            <th style="padding:8px 12px;">Dossier &amp; Produit</th>
                            <th style="padding:8px 12px;">Date Réception</th>
                            <th style="padding:8px 12px;">Délai Légal</th>
                            <th style="padding:8px 12px;">Compte à Rebours</th>
                            <th style="padding:8px 12px;">Progression</th>
                            <th style="padding:8px 12px; text-align:right;">Actions</th>
                        </tr>
                    </thead>
                    <tbody>
        `;

        letters.forEach((l, idx) => {
            const rowBg = idx % 2 === 0 ? 'var(--bg-panel)' : 'var(--bg-surface)';
            const totalQ = l.total_tasks_count || 0;
            const ansQ = l.answered_tasks_count || 0;
            const valQ = l.validated_tasks_count || 0;
            const pct = totalQ > 0 ? Math.round((valQ / totalQ) * 100) : 0;

            html += `
                <tr style="background:${rowBg}; border-bottom:1px solid var(--border-color); cursor:pointer;" onclick="openDeficiencyLetterDetail('${l.id}')">
                    <td style="padding:10px 12px;">
                        <div style="font-weight:700; color:var(--accent-blue);">${escapeHTML(l.reference_number)}</div>
                        <div style="font-size:10px; color:var(--text-muted);">${l.notes ? escapeHTML(l.notes).substring(0, 45) + '...' : 'Notification officielle'}</div>
                    </td>
                    <td style="padding:10px 12px;">
                        <div style="font-weight:600; color:var(--text-main);">${escapeHTML(l.product_brand_name || l.dossier_reference || 'Dossier ANPP')}</div>
                        <div style="font-size:10px; color:var(--text-dim); font-family:monospace;">${escapeHTML(l.dossier_reference || '')}</div>
                    </td>
                    <td style="padding:10px 12px; white-space:nowrap;">
                        ${(l.received_date || '').substring(0, 10)}
                    </td>
                    <td style="padding:10px 12px; white-space:nowrap;">
                        <span class="badge" style="background:var(--bg-sub); color:var(--text-main); border:1px solid var(--border-color);">${l.legal_deadline_days} jours</span>
                    </td>
                    <td style="padding:10px 12px; white-space:nowrap;">
                        ${getUrgencyBadgeHTML(l.days_remaining, l.urgency_badge)}
                    </td>
                    <td style="padding:10px 12px; min-width:140px;">
                        <div style="display:flex; justify-content:space-between; font-size:10px; margin-bottom:3px;">
                            <span>${valQ}/${totalQ} visés PDT</span>
                            <span style="font-weight:700;">${pct}%</span>
                        </div>
                        <div style="width:100%; height:6px; background:var(--bg-sub); border-radius:3px; overflow:hidden;">
                            <div style="width:${pct}%; height:100%; background:${pct === 100 ? '#16a34a' : 'var(--accent-blue)'}; border-radius:3px;"></div>
                        </div>
                    </td>
                    <td style="padding:10px 12px; text-align:right; white-space:nowrap;" onclick="event.stopPropagation();">
                        <button class="btn-pill" style="font-size:10.5px; padding:3px 8px;" onclick="openDeficiencyLetterDetail('${l.id}')">🔍 Ouvrir Studio</button>
                    </td>
                </tr>
            `;
        });

        html += `
                    </tbody>
                </table>
            </div>
        `;
    }

    html += `
            </div>
        </div>
    `;

    centerView.innerHTML = html;
}

// Ouvrir les détails et la gestion d'une lettre de compléments
async function openDeficiencyLetterDetail(letterId) {
    try {
        const resp = await fetch(`/api/v1/deficiencies/letters/${letterId}`, {
            headers: { 'Accept': 'application/json' }
        });
        if (!resp.ok) throw new Error(`Erreur HTTP ${resp.status}`);
        const letter = await resp.json();
        currentDeficiencyLetterDetail = letter;
        renderDeficiencyLetterDetailModal(letter);
    } catch (err) {
        alert("Impossible de charger la lettre: " + err.message);
    }
}

// Rendu modal interactif de la lettre avec les questions unitaires et visa PDT
function renderDeficiencyLetterDetailModal(letter) {
    const modal = document.getElementById("modal-deficiency-detail");
    if (!modal) return;

    document.getElementById("m-def-ref").innerText = letter.reference_number;
    document.getElementById("m-def-product").innerText = `${letter.product_brand_name || ''} (${letter.dossier_reference || ''})`;
    document.getElementById("m-def-countdown").innerHTML = getUrgencyBadgeHTML(letter.days_remaining, letter.urgency_badge);
    document.getElementById("m-def-received").innerText = (letter.received_date || '').substring(0, 10);
    document.getElementById("m-def-due").innerText = (letter.due_date || '').substring(0, 10);
    document.getElementById("m-def-notes").innerText = letter.notes || 'Aucune note particulière.';

    const tasksList = document.getElementById("m-def-tasks-list");
    if (!tasksList) return;

    const isPDT = typeof hasRole === 'function' && (hasRole('directeur_technique', 'admin_laboratoire') || (typeof hasPermission === 'function' && hasPermission('DOC_APPROVE_PDT')));

    let tasksHTML = '';
    (letter.tasks || []).forEach(t => {
        const isAnswered = t.status === 'repondu' || t.status === 'valide_dt';
        const isValidated = t.status === 'valide_dt';

        let statusBadge = `<span class="badge" style="background:#fee2e2; color:#b91c1c;">En attente de réponse</span>`;
        if (isValidated) {
            statusBadge = `<span class="badge" style="background:#ecfdf5; color:#047857;">✍️ Visé par le Pharmacien DT</span>`;
        } else if (isAnswered) {
            statusBadge = `<span class="badge" style="background:#fef3c7; color:#b45309;">📝 Réponse rédigée (En attente visa PDT)</span>`;
        }

        tasksHTML += `
            <div class="admin-preview-card" style="padding:12px; border-left:4px solid ${isValidated ? '#16a34a' : (isAnswered ? '#f59e0b' : 'var(--accent-blue)')};">
                <div style="display:flex; justify-content:space-between; align-items:flex-start; margin-bottom:8px; gap:8px;">
                    <div>
                        <div style="font-weight:700; color:var(--text-main); font-size:12px;">
                            Question #${t.question_number} — <span style="color:var(--accent-blue);">${escapeHTML(t.module_target || 'Général')}</span>
                        </div>
                        <div style="font-size:11px; color:var(--text-muted); margin-top:4px; line-height:1.4;">
                            ${escapeHTML(t.question_text)}
                        </div>
                    </div>
                    <div style="text-align:right; white-space:nowrap;">
                        ${statusBadge}
                    </div>
                </div>

                <!-- Zone de réponse -->
                <div style="margin-top:10px; background:var(--bg-sub); padding:10px; border-radius:6px; border:1px solid var(--border-color);">
                    <div style="font-size:11px; font-weight:600; color:var(--text-main); margin-bottom:4px;">
                        Argumentaire / Résumé de la réponse réglementaire :
                    </div>
                    <textarea id="task-resp-${t.id}" rows="2" ${isValidated ? 'disabled' : ''} style="width:100%; padding:6px; font-size:11.5px; background:var(--bg-panel); color:var(--text-main); border:1px solid var(--border-color); border-radius:4px;" placeholder="Détailler la justification scientifique, joindre les références d'essais ou certificats...">${escapeHTML(t.response_summary || '')}</textarea>
                    
                    ${t.validated_by_name ? `
                        <div style="margin-top:6px; font-size:10.5px; color:#16a34a; font-weight:600;">
                            ✅ Visé électroniquement par ${escapeHTML(t.validated_by_name)} le ${(t.validated_at || '').substring(0, 10)}
                        </div>
                    ` : ''}

                    <div style="display:flex; justify-content:flex-end; gap:8px; margin-top:8px;">
                        ${!isValidated ? `
                            <button class="btn-pill" style="font-size:11px;" onclick="saveTaskAnswer('${t.id}')">💾 Enregistrer Brouillon</button>
                        ` : ''}
                        ${(!isValidated && isPDT) ? `
                            <button class="btn-pill" style="font-size:11px; background:#16a34a; color:white; border:none;" onclick="validateTaskDT('${t.id}')">✍️ Apposer Visa PDT (1.A.7)</button>
                        ` : ''}
                    </div>
                </div>
            </div>
        `;
    });

    tasksList.innerHTML = tasksHTML;
    openModal("modal-deficiency-detail");
}

// Enregistrement d'une réponse à une question
async function saveTaskAnswer(taskId) {
    const summaryEl = document.getElementById(`task-resp-${taskId}`);
    if (!summaryEl) return;
    const summary = summaryEl.value.trim();

    try {
        const resp = await fetch(`/api/v1/deficiencies/tasks/${taskId}`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            },
            body: JSON.stringify({
                response_summary: summary,
                status: summary ? 'repondu' : 'a_traiter'
            })
        });

        if (!resp.ok) {
            const err = await resp.json();
            throw new Error(err.error || `Erreur HTTP ${resp.status}`);
        }

        if (typeof showToast === 'function') {
            showToast("Réponse enregistrée avec succès");
        } else {
            alert("Réponse enregistrée avec succès");
        }

        if (currentDeficiencyLetterDetail) {
            openDeficiencyLetterDetail(currentDeficiencyLetterDetail.id);
        }
    } catch (err) {
        alert("Erreur lors de l'enregistrement: " + err.message);
    }
}

// Apposition du visa légal du Pharmacien Directeur Technique (PDT - 1.A.7)
async function validateTaskDT(taskId) {
    const comment = prompt("Visa du Pharmacien Directeur Technique (Mandat Légal 1.A.7) :\nSaisir un commentaire de validation pour l'Audit Trail BPF (optionnel) :", "Réponse approuvée conforme aux exigences du Guide CTD ANPP 2025");
    if (comment === null) return;

    try {
        const resp = await fetch(`/api/v1/deficiencies/tasks/${taskId}/validate-dt`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            },
            body: JSON.stringify({ comment: comment })
        });

        if (!resp.ok) {
            const err = await resp.json();
            throw new Error(err.error || `Erreur HTTP ${resp.status}`);
        }

        if (typeof showToast === 'function') {
            showToast("Visa légal PDT apposé avec succès (Audit Trail scellé)");
        } else {
            alert("Visa légal PDT apposé avec succès (Audit Trail scellé)");
        }

        if (currentDeficiencyLetterDetail) {
            openDeficiencyLetterDetail(currentDeficiencyLetterDetail.id);
        }
    } catch (err) {
        alert("Erreur lors de l'apposition du visa PDT: " + err.message);
    }
}

// Ouverture de la modale de création d'une nouvelle notification ANPP
async function openNewDeficiencyLetterModal(dossierId) {
    const modal = document.getElementById("modal-new-deficiency-letter");
    if (!modal) return;

    const selectEl = document.getElementById("def-dossier-select");
    if (selectEl) {
        selectEl.innerHTML = '<option value="">-- Sélectionner un dossier d\'enregistrement --</option>';
        try {
            let dossiers = typeof currentDossiers !== 'undefined' && currentDossiers && currentDossiers.length > 0 ? currentDossiers : [];
            if (dossiers.length === 0) {
                const resp = await fetch('/api/v1/dossiers');
                if (resp.ok) {
                    const data = await resp.json();
                    dossiers = data.dossiers || [];
                }
            }
            dossiers.forEach(d => {
                const opt = document.createElement("option");
                opt.value = d.id;
                const prodName = d.medicine ? d.medicine.brand_name_fr : (d.medical_device ? d.medical_device.commercial_name_fr : '');
                opt.text = `${d.dossier_code} — ${prodName || d.internal_reference || 'Dossier'}`;
                if (dossierId && d.id === dossierId) {
                    opt.selected = true;
                }
                selectEl.appendChild(opt);
            });
        } catch (e) {
            console.error("Erreur chargement dossiers pour modal:", e);
        }
    }

    if (dossierId) {
        document.getElementById("def-dossier-id").value = dossierId;
        if (selectEl) selectEl.value = dossierId;
    } else {
        document.getElementById("def-dossier-id").value = '';
    }

    document.getElementById("def-ref-number").value = '';
    document.getElementById("def-received-date").value = new Date().toISOString().substring(0, 10);
    document.getElementById("def-deadline-days").value = "90";
    document.getElementById("def-notes").value = '';

    // Réinitialiser les questions à 1 question par défaut
    const questionsContainer = document.getElementById("def-questions-rows");
    if (questionsContainer) {
        questionsContainer.innerHTML = '';
        addQuestionRowToLetterModal();
    }

    openModal("modal-new-deficiency-letter");
}

// Ajout dynamique d'une ligne de question dans la modale
function addQuestionRowToLetterModal() {
    const container = document.getElementById("def-questions-rows");
    if (!container) return;

    const count = container.children.length + 1;
    const row = document.createElement("div");
    row.className = "question-input-row";
    row.style = "display:flex; flex-direction:column; gap:6px; background:var(--bg-sub); padding:8px; border-radius:6px; border:1px solid var(--border-color); margin-bottom:8px;";

    row.innerHTML = `
        <div style="display:flex; justify-content:space-between; align-items:center;">
            <span style="font-weight:700; font-size:11px; color:var(--accent-blue);">Question #${count}</span>
            ${count > 1 ? `<button type="button" class="btn-pill" style="font-size:10px; padding:2px 6px; color:var(--accent-red);" onclick="this.closest('.question-input-row').remove()">✕ Supprimer</button>` : ''}
        </div>
        <textarea class="q-text" rows="3" required placeholder="Libellé exact ou extrait officiel de la question/réserve notifiée par l'ANPP..." style="width:100%; padding:8px; font-size:12px; background:var(--bg-panel); color:var(--text-main); border:1px solid var(--border-color); border-radius:4px; resize:vertical; min-height:64px; font-family:inherit; line-height:1.4;"></textarea>
        <div style="display:grid; grid-template-columns:1fr 1fr; gap:8px;">
            <div>
                <label style="font-size:10px; font-weight:600;">Module Cible :</label>
                <select class="q-module" style="width:100%; padding:6px; font-size:11px; background:var(--bg-panel); color:var(--text-main); border:1px solid var(--border-color); border-radius:4px;">
                    <option value="Module 1">Module 1 (Admin & Spécificités Algérie)</option>
                    <option value="Module 2">Module 2 (Résumés & Overviews)</option>
                    <option value="Module 3" selected>Module 3 (Qualité / CMC)</option>
                    <option value="Module 4">Module 4 (Non-Clinique)</option>
                    <option value="Module 5">Module 5 (Clinique & Bioéquivalence)</option>
                    <option value="Notice/Étiquetage">Notice & Étiquetage (1.M.1)</option>
                    <option value="Dossier Technique DM">Dossier Technique DM (Arrêté 10/05/2021)</option>
                    <option value="Variation Post-AMM">Variation Post-AMM (LD-DEPP07/V01)</option>
                </select>
            </div>
            <div>
                <label style="font-size:10px; font-weight:600;">Priorité :</label>
                <select class="q-priority" style="width:100%; padding:6px; font-size:11px; background:var(--bg-panel); color:var(--text-main); border:1px solid var(--border-color); border-radius:4px;">
                    <option value="normale">Normale</option>
                    <option value="haute">Haute</option>
                    <option value="urgente">Urgente</option>
                </select>
            </div>
        </div>
    `;

    container.appendChild(row);
}

// Soumission de la nouvelle notification officielle ANPP
async function executeCreateDeficiencyLetter(e) {
    if (e) e.preventDefault();

    let dossierId = document.getElementById("def-dossier-id").value;
    const selectEl = document.getElementById("def-dossier-select");
    if (!dossierId && selectEl) {
        dossierId = selectEl.value;
    }
    if (!dossierId) {
        alert("Veuillez sélectionner un dossier d'enregistrement");
        return;
    }

    const refNumber = document.getElementById("def-ref-number").value.trim();
    const receivedDate = document.getElementById("def-received-date").value;
    const deadlineDays = parseInt(document.getElementById("def-deadline-days").value, 10) || 90;
    const notes = document.getElementById("def-notes").value.trim();

    if (!refNumber) {
        alert("La référence officielle de la lettre ANPP est obligatoire");
        return;
    }

    const questionRows = document.querySelectorAll("#def-questions-rows .question-input-row");
    const tasks = [];
    questionRows.forEach((row, i) => {
        const text = row.querySelector(".q-text").value.trim();
        const mod = row.querySelector(".q-module").value;
        const prio = row.querySelector(".q-priority").value;
        if (text) {
            tasks.push({
                question_number: i + 1,
                question_text: text,
                module_target: mod,
                priority: prio
            });
        }
    });

    if (tasks.length === 0) {
        alert("Veuillez saisir au moins une question notifiée par l'ANPP");
        return;
    }

    try {
        const resp = await fetch(`/api/v1/dossiers/${dossierId}/deficiencies`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            },
            body: JSON.stringify({
                reference_number: refNumber,
                received_date: receivedDate,
                legal_deadline_days: deadlineDays,
                notes: notes,
                tasks: tasks
            })
        });

        if (!resp.ok) {
            const err = await resp.json();
            throw new Error(err.error || `Erreur HTTP ${resp.status}`);
        }

        const created = await resp.json();
        closeModal("modal-new-deficiency-letter");
        
        if (typeof showToast === 'function') {
            showToast(`Lettre ANPP enregistrée (${tasks.length} questions décomposées)`);
        } else {
            alert(`Lettre ANPP enregistrée (${tasks.length} questions décomposées)`);
        }

        // Si nous sommes dans la vue compléments, rafraîchir le dashboard
        if (typeof currentDomain !== 'undefined' && currentDomain === 'deficiencies') {
            loadDeficienciesDashboard();
        } else if (typeof loadDossiers === 'function') {
            loadDossiers();
        }
    } catch (err) {
        alert("Erreur lors de la création de la lettre: " + err.message);
    }
}
