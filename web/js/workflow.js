// ============================================================================
// Jalon 2.1 : Atelier Rédactionnel BPF, Maker-Checker & Moteur d'États de Sortie
// Conformité BPF / 21 CFR Part 11 / ALCOA+ / Guide ANPP 2025
// ============================================================================

// Fallback de sécurité globale pour les notifications Toast
if (typeof window.showToast !== 'function') {
    window.showToast = function(msg, type) {
        console.log(`[TOAST ${type || 'info'}] ${msg}`);
        if (typeof alert === 'function' && type === 'error') alert(msg);
    };
}

let currentWfDossierId = null;
let currentWfItemId = null;
let currentReportCode = 'ES-REG-01';

// Helpers de formatage de badges
function formatWorkflowBadge(status) {
    switch (status) {
        case 'draft_pending':
            return '<span class="badge-wf-draft_pending" title="En attente de rédaction (Maker)">📝 À Rédiger</span>';
        case 'drafting':
            return '<span class="badge-wf-drafting" title="En cours de rédaction (Maker)">✏️ En Rédaction</span>';
        case 'under_review':
            return '<span class="badge-wf-under_review" title="Soumis pour relecture Qualité/Pair (Checker)">🔍 En Revue AQ</span>';
        case 'pending_approval':
            return '<span class="badge-wf-pending_approval" title="Revue validée - En attente Visa Pharmacien Directeur Technique">⚖️ Visa DT Requis</span>';
        case 'approved_sealed':
            return '<span class="badge-wf-approved_sealed" title="Approuvé &amp; Scellé v1.0 par le Pharmacien DT">🛡️ Scellé v1.0</span>';
        default:
            return '<span class="badge-wf-draft_pending">📝 À Rédiger</span>';
    }
}

function formatDepartmentBadge(dept) {
    if (!dept) return '';
    const d = dept.toLowerCase();
    const mapLabels = {
        'ra': 'RA (Réglementaire)',
        'qc': 'QC (Contrôle Qualité)',
        'rd': 'RD (R&amp;D / Galénique)',
        'prod': 'PROD (Production)',
        'qa': 'QA (Assurance Qualité)',
        'external': 'EXT (CRO / Tiers)'
    };
    return `<span class="badge-dept badge-dept-${d}" title="Département responsable : ${mapLabels[d] || dept}">${dept}</span>`;
}

function formatRevisionBadge(rev, isSealed) {
    if (!rev) rev = 'v0.1';
    if (isSealed || rev.startsWith('v1.')) {
        return `<span class="badge-rev-sealed" title="Version formelle scellée">${rev}</span>`;
    }
    return `<span class="badge-rev" title="Version de travail mineure">${rev}</span>`;
}

// ----------------------------------------------------------------------------
// Check-out (Verrouillage en édition)
// ----------------------------------------------------------------------------
function openCheckoutModal(dossierId, itemId, itemCode, itemTitle) {
    currentWfDossierId = dossierId;
    currentWfItemId = itemId;

    const modal = document.getElementById('modal-workflow-checkout');
    if (!modal) return;

    document.getElementById('wf-checkout-code').textContent = itemCode;
    document.getElementById('wf-checkout-title').textContent = itemTitle;
    document.getElementById('wf-checkout-reason').value = 'Élaboration du contenu et consolidation des données analytiques';
    document.getElementById('wf-checkout-err').style.display = 'none';

    openModal('modal-workflow-checkout');
}

async function submitWorkflowCheckout(e) {
    if (e) e.preventDefault();
    const reason = document.getElementById('wf-checkout-reason').value.trim();
    const errBox = document.getElementById('wf-checkout-err');

    try {
        const resp = await fetch(`/api/v1/dossiers/${currentWfDossierId}/items/${currentWfItemId}/checkout`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ reason: reason })
        });
        const data = await resp.json();
        if (!resp.ok) {
            errBox.textContent = data.message || "Erreur lors du Check-out";
            errBox.style.display = 'block';
            return;
        }

        closeModal('modal-workflow-checkout');
        showToast("🔒 Section verrouillée avec succès en Check-out pour édition.", "success");
        if (typeof switchPreviewSubTab === 'function') switchPreviewSubTab('wf');
        if (typeof selectDossier === 'function') {
            await selectDossier(currentWfDossierId);
            if (typeof selectDossierItem === 'function' && currentWfItemId) {
                selectDossierItem(currentWfItemId);
            }
        }
    } catch (err) {
        errBox.textContent = err.message || "Erreur réseau";
        errBox.style.display = 'block';
    }
}

// ----------------------------------------------------------------------------
// Check-in (Restitution et incrément de révision)
// ----------------------------------------------------------------------------
async function openCheckinModal(dossierId, itemId, itemCode, currentRev) {
    currentWfDossierId = dossierId;
    currentWfItemId = itemId;

    const modal = document.getElementById('modal-workflow-checkin');
    if (!modal) return;

    document.getElementById('wf-checkin-code').textContent = itemCode;
    document.getElementById('wf-checkin-rev').textContent = currentRev || 'v0.1';
    document.getElementById('wf-checkin-summary').value = '';
    document.getElementById('wf-checkin-err').style.display = 'none';

    const gedSelect = document.getElementById('wf-checkin-ged-select');
    if (gedSelect) {
        gedSelect.innerHTML = '<option value="">-- Conserver la pièce actuelle ou rattacher ultérieurement --</option>';
        try {
            const res = await fetch('/api/v1/ged/documents?limit=100');
            if (res.ok) {
                const data = await res.json();
                const docs = data.documents || data.items || [];
                docs.forEach(d => {
                    const opt = document.createElement('option');
                    opt.value = d.id;
                    opt.textContent = `[${d.document_code || 'GED'}] ${d.title_fr} (${d.current_version || 'v1.0'})`;
                    gedSelect.appendChild(opt);
                });
            }
        } catch (_) {}
    }

    openModal('modal-workflow-checkin');
}

async function submitWorkflowCheckin(e) {
    if (e) e.preventDefault();
    const summary = document.getElementById('wf-checkin-summary').value.trim();
    const errBox = document.getElementById('wf-checkin-err');

    if (!summary) {
        errBox.textContent = "Le sommaire des modifications (justification ALCOA+) est obligatoire.";
        errBox.style.display = 'block';
        return;
    }

    const gedDocId = document.getElementById('wf-checkin-ged-select') ? document.getElementById('wf-checkin-ged-select').value : null;
    const payload = {
        revision_summary: summary,
        comment: summary
    };
    if (gedDocId) {
        payload.ged_document_id = gedDocId;
    }

    try {
        const resp = await fetch(`/api/v1/dossiers/${currentWfDossierId}/items/${currentWfItemId}/checkin`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });
        const data = await resp.json();
        if (!resp.ok) {
            errBox.textContent = data.message || "Erreur lors du Check-in";
            errBox.style.display = 'block';
            return;
        }

        closeModal('modal-workflow-checkin');
        showToast(`🔓 Check-in validé. Nouvelle révision : ${data.new_revision}`, "success");
        if (typeof switchPreviewSubTab === 'function') switchPreviewSubTab('wf');
        if (typeof selectDossier === 'function') {
            await selectDossier(currentWfDossierId);
            if (typeof selectDossierItem === 'function' && currentWfItemId) {
                selectDossierItem(currentWfItemId);
            }
        }
    } catch (err) {
        errBox.textContent = err.message || "Erreur réseau";
        errBox.style.display = 'block';
    }
}

// ----------------------------------------------------------------------------
// Assignation Départementale & Nominative (BPF / CDC §7.4)
// ----------------------------------------------------------------------------
let assignableUsersCache = [];

async function openAssignModal(dossierId, itemId, itemCode, currentDept, currentDueDate, currentUserId) {
    currentWfDossierId = dossierId;
    currentWfItemId = itemId;

    const modal = document.getElementById('modal-workflow-assign');
    if (!modal) return;

    document.getElementById('wf-assign-code').textContent = itemCode;
    document.getElementById('wf-assign-dept').value = currentDept || 'RA';
    document.getElementById('wf-assign-due').value = currentDueDate ? currentDueDate.substring(0, 10) : '';
    document.getElementById('wf-assign-notes').value = '';
    document.getElementById('wf-assign-err').style.display = 'none';

    // Charger et peupler les collaborateurs selon le pôle métier
    await loadAssignableUsersForWorkflow(dossierId, currentDept || 'RA', currentUserId);

    openModal('modal-workflow-assign');
}

async function loadAssignableUsersForWorkflow(dossierId, dept, selectedUserId) {
    const userSelect = document.getElementById('wf-assign-user');
    if (!userSelect) return;

    try {
        if (!assignableUsersCache || assignableUsersCache.length === 0) {
            // 1. Tenter les membres du dossier
            let res = await fetch(`/api/v1/dossiers/${dossierId}/members`);
            if (res.ok) {
                const data = await res.json();
                if (data.members && data.members.length > 0) {
                    assignableUsersCache = data.members.map(m => ({
                        id: m.user_id,
                        name: `${m.user_first_name || ''} ${m.user_last_name || ''}`.trim() || m.user_email,
                        email: m.user_email,
                        role: m.user_role
                    }));
                }
            }
            // 2. Si aucun membre spécifique, récupérer tous les utilisateurs du laboratoire
            if (!assignableUsersCache || assignableUsersCache.length === 0) {
                res = await fetch('/api/v1/admin/users');
                if (res.ok) {
                    const data = await res.json();
                    if (data.users && data.users.length > 0) {
                        assignableUsersCache = data.users.map(u => ({
                            id: u.id,
                            name: `${u.first_name || ''} ${u.last_name || ''}`.trim() || u.email,
                            email: u.email,
                            role: u.role
                        }));
                    }
                }
            }
        }

        renderAssignUsersOptions(dept, selectedUserId);
    } catch(err) {
        console.error("loadAssignableUsersForWorkflow", err);
        userSelect.innerHTML = '<option value="">-- Non assigné (Pôle global) --</option>';
    }
}

function filterAssignUsersByDept() {
    const dept = document.getElementById('wf-assign-dept').value;
    const currentSelected = document.getElementById('wf-assign-user').value;
    renderAssignUsersOptions(dept, currentSelected);
}

function renderAssignUsersOptions(dept, selectedUserId) {
    const userSelect = document.getElementById('wf-assign-user');
    if (!userSelect) return;

    let html = '<option value="">-- Non assigné (Pôle global) --</option>';
    if (assignableUsersCache && assignableUsersCache.length > 0) {
        html += assignableUsersCache.map(u => {
            const isSel = (selectedUserId && (u.id === selectedUserId || String(u.id) === String(selectedUserId))) ? 'selected' : '';
            return `<option value="${u.id}" ${isSel}>${escapeHTML(u.name)} (${escapeHTML(u.role || '')})</option>`;
        }).join('');
    }
    userSelect.innerHTML = html;
}

async function submitWorkflowAssign(e) {
    if (e) e.preventDefault();
    const dept = document.getElementById('wf-assign-dept').value;
    const userSelect = document.getElementById('wf-assign-user');
    const assignedUserId = userSelect && userSelect.value ? userSelect.value : null;
    let due = document.getElementById('wf-assign-due').value;
    if (!due) {
        // Fixe par défaut une échéance à +14 jours si non renseignée pour planifier la mission
        const targetD = new Date();
        targetD.setDate(targetD.getDate() + 14);
        due = targetD.toISOString().split('T')[0];
    }
    const notes = document.getElementById('wf-assign-notes').value.trim();
    const errBox = document.getElementById('wf-assign-err');

    try {
        const resp = await fetch(`/api/v1/dossiers/${currentWfDossierId}/items/${currentWfItemId}/assign`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                assigned_department: dept,
                assigned_user_id: assignedUserId,
                due_date: due ? due : null,
                notes: notes
            })
        });
        const data = await resp.json();
        if (!resp.ok) {
            errBox.textContent = data.message || "Erreur lors de l'assignation";
            errBox.style.display = 'block';
            return;
        }

        closeModal('modal-workflow-assign');
        showToast("👥 Section assignée avec succès au pôle métier.", "success");
        if (typeof switchPreviewSubTab === 'function') switchPreviewSubTab('wf');
        if (typeof selectDossier === 'function') {
            await selectDossier(currentWfDossierId);
            if (typeof selectDossierItem === 'function' && currentWfItemId) {
                selectDossierItem(currentWfItemId);
            }
        }
    } catch (err) {
        errBox.textContent = err.message || "Erreur réseau";
        errBox.style.display = 'block';
    }
}

// ----------------------------------------------------------------------------
// Soumission à Relecture Qualité (Maker -> Checker)
// ----------------------------------------------------------------------------
async function submitItemForReview(dossierId, itemId, itemCode) {
    if (!confirm(`Soumettre la section [${itemCode}] à la relecture Qualité/Pair (Checker BPF) ?\nElle basculera en statut 'En revue AQ'.`)) {
        return;
    }

    try {
        const resp = await fetch(`/api/v1/dossiers/${dossierId}/items/${itemId}/submit-review`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ notes: "Soumission pour relecture BPF contradictoire" })
        });
        const data = await resp.json();
        if (!resp.ok) {
            if (resp.status === 422 && data.code === 'GATING_BLOCKED') {
                showGatingBlockedAlert(data);
                return;
            }
            alert(data.message || data.error || "Erreur lors de la soumission à revue");
            return;
        }

        showToast("📤 Section soumise avec succès pour revue Qualité (Checker).", "success");
        if (typeof switchPreviewSubTab === 'function') switchPreviewSubTab('wf');
        if (typeof selectDossier === 'function') {
            await selectDossier(dossierId);
            if (typeof selectDossierItem === 'function' && itemId) {
                selectDossierItem(itemId);
            }
        }
    } catch (err) {
        alert(err.message || "Erreur réseau");
    }
}

// ----------------------------------------------------------------------------
// Décision de Relecture Checker (Segregation of Duties / Double Regard)
// ----------------------------------------------------------------------------
function openReviewDecisionModal(dossierId, itemId, itemCode, title) {
    currentWfDossierId = dossierId;
    currentWfItemId = itemId;

    // Résolution robuste de l'item depuis currentDossierDetail
    let itemObj = null;
    if (typeof currentDossierDetail === 'object' && currentDossierDetail && Array.isArray(currentDossierDetail.checklist)) {
        itemObj = currentDossierDetail.checklist.find(i => i.id === itemId);
    }
    if (itemObj) {
        if (!itemCode) itemCode = itemObj.item_code;
        if (!title) title = itemObj.title_fr;
    }

    // Garde-fou BPF côté client : Ségrégation des Fonctions (Maker ≠ Checker)
    const user = (typeof window.getCurrentUser === 'function' ? window.getCurrentUser() : window.currentUser);
    if (user && itemObj && itemObj.maker_user_id && itemObj.maker_user_id === user.id) {
        if (typeof showNotifyModal === 'function') {
            showNotifyModal(
                "Ségrégation des Fonctions (BPF / ALCOA+)",
                `<div style="display:flex; flex-direction:column; gap:10px;">
                    <div style="color:var(--accent-amber); font-weight:700; font-size:13px;">
                        ⚖️ Règle du Double Regard Obligatoire (Maker ≠ Checker)
                    </div>
                    <div style="font-size:12px; line-height:1.5; color:var(--text-main);">
                        Vous êtes enregistré comme le <strong>Rédacteur (Maker)</strong> de la section <code>[${escapeHTML(itemCode || '')}]</code>.
                    </div>
                    <div style="background:var(--bg-sub); border:1px solid var(--border-color); border-radius:4px; padding:10px; font-size:11.5px; color:var(--text-muted); line-height:1.4;">
                        Conformément à l'Annexe 11 des BPF et au 21 CFR Part 11, un contributeur ne peut en aucun cas relire, valider ou approuver son propre document.<br><br>
                        👉 Cette revue contradictoire doit être effectuée par un vérificateur indépendant (ex: <strong>Responsable Assurance Qualité</strong> ou pair technique).
                    </div>
                    <div style="font-size:11px; color:var(--text-dim); margin-top:4px;">
                        <em>💡 Astuce de test : Utilisez le commutateur de rôle <strong>(Dev-Switcher en haut à droite)</strong> pour basculer sur un profil indépendant tel que 'Responsable AQ' ou 'Pharmacien Directeur Technique'.</em>
                    </div>
                </div>`,
                "⚖️",
                "Compris"
            );
            return;
        }
    }

    const modal = document.getElementById('modal-workflow-review-decision');
    if (!modal) return;

    document.getElementById('wf-review-code').textContent = itemCode || '';
    document.getElementById('wf-review-title').textContent = title || '';
    document.getElementById('wf-review-comments').value = '';
    document.getElementById('wf-review-actions').value = '';
    document.getElementById('wf-review-err').style.display = 'none';

    openModal('modal-workflow-review-decision');
}

async function submitWorkflowReviewDecision(decision) {
    const comments = document.getElementById('wf-review-comments').value.trim();
    const reqActions = document.getElementById('wf-review-actions').value.trim();
    const errBox = document.getElementById('wf-review-err');

    if (!comments) {
        errBox.textContent = "Les commentaires détaillés de relecture sont obligatoires (Audit Trail BPF).";
        errBox.style.display = 'block';
        return;
    }

    if (decision === 'rejected' && !reqActions) {
        errBox.textContent = "En cas de rejet / demande de reprise, les actions correctives attendues sont obligatoires.";
        errBox.style.display = 'block';
        return;
    }

    try {
        const resp = await fetch(`/api/v1/dossiers/${currentWfDossierId}/items/${currentWfItemId}/review-decision`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                decision: decision,
                notes: comments,
                comments: comments,
                required_actions: reqActions
            })
        });
        const data = await resp.json();
        if (!resp.ok) {
            const errorMsg = data.message || data.error || "Erreur lors de l'enregistrement de la décision";
            errBox.textContent = errorMsg;
            errBox.style.display = 'block';
            if (resp.status === 403 && typeof showNotifyModal === 'function') {
                showNotifyModal(
                    "Ségrégation des Fonctions (SoD BPF)",
                    `<div style="display:flex; flex-direction:column; gap:10px;">
                        <div style="color:var(--accent-red); font-weight:700; font-size:13px;">
                            🚫 Action Réglementaire Interdite (21 CFR Part 11)
                        </div>
                        <div style="font-size:12px; line-height:1.5; color:var(--text-main);">
                            ${escapeHTML(errorMsg)}
                        </div>
                        <div style="background:var(--bg-sub); border:1px solid var(--border-color); border-radius:4px; padding:10px; font-size:11.5px; color:var(--text-muted); line-height:1.4;">
                            <strong>Règle Maker ≠ Checker :</strong> Conformément à l'Annexe 11 BPF et aux normes ALCOA+, le rédacteur d'une section ne peut pas valider ou approuver son propre document.<br><br>
                            👉 <strong>Pour valider la revue :</strong> Utilisez le commutateur de rôle en haut à droite <em>(Dev-Switcher)</em> pour basculer sur le profil d'un vérificateur indépendant tel que <strong>Karim Belkacem (Pharmacien Responsable Qualité / Checker)</strong> ou <strong>Directeur Technique</strong>.
                        </div>
                    </div>`,
                    "🚫",
                    "Compris"
                );
            }
            return;
        }

        closeModal('modal-workflow-review-decision');
        if (decision === 'approved') {
            showToast("✅ Revue Qualité validée avec succès. La section est prête pour le Visa du PDT.", "success");
        } else {
            showToast("↩️ Relecture défavorable : section renvoyée au Maker pour corrections.", "info");
        }

        if (typeof switchPreviewSubTab === 'function') switchPreviewSubTab('wf');
        if (typeof selectDossier === 'function') {
            await selectDossier(currentWfDossierId);
            if (typeof selectDossierItem === 'function' && currentWfItemId) {
                selectDossierItem(currentWfItemId);
            }
        }
    } catch (err) {
        errBox.textContent = err.message || "Erreur réseau";
        errBox.style.display = 'block';
    }
}

// ----------------------------------------------------------------------------
// Visa Pharmacien Directeur Technique (PDT - 21 CFR Part 11 & Sceau Inaltérable)
// ----------------------------------------------------------------------------
function openPdtSealModal(dossierId, itemId, itemCode, title) {
    currentWfDossierId = dossierId;
    currentWfItemId = itemId;

    let itemObj = null;
    if (typeof currentDossierDetail === 'object' && currentDossierDetail && Array.isArray(currentDossierDetail.checklist)) {
        itemObj = currentDossierDetail.checklist.find(i => i.id === itemId);
    }
    if (itemObj) {
        if (!itemCode) itemCode = itemObj.item_code;
        if (!title) title = itemObj.title_fr;
    }

    const modal = document.getElementById('modal-workflow-pdt-seal');
    if (!modal) return;

    document.getElementById('wf-pdt-code').textContent = itemCode || '';
    document.getElementById('wf-pdt-title').textContent = title || '';
    document.getElementById('wf-pdt-totp').value = '';
    document.getElementById('wf-pdt-solemn-check').checked = false;
    document.getElementById('wf-pdt-err').style.display = 'none';

    openModal('modal-workflow-pdt-seal');
}

async function submitWorkflowPdtApproval(e) {
    if (e) e.preventDefault();
    const totp = document.getElementById('wf-pdt-totp').value.trim();
    const solemnCheck = document.getElementById('wf-pdt-solemn-check').checked;
    const errBox = document.getElementById('wf-pdt-err');

    if (!solemnCheck) {
        errBox.textContent = "Vous devez accepter formellement la déclaration solennelle de responsabilité pharmaceutique.";
        errBox.style.display = 'block';
        return;
    }

    try {
        const resp = await fetch(`/api/v1/dossiers/${currentWfDossierId}/items/${currentWfItemId}/approve-pdt`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                totp_code: totp,
                password: totp,
                solemn_declaration: true,
                solemn_declaration_accepted: true
            })
        });
        const data = await resp.json();
        if (!resp.ok) {
            if (resp.status === 422 && data.code === 'GATING_BLOCKED') {
                closeModal('modal-workflow-pdt-seal');
                showGatingBlockedAlert(data);
                return;
            }
            errBox.textContent = data.message || data.error || "Erreur lors du visa PDT";
            errBox.style.display = 'block';
            return;
        }

        closeModal('modal-workflow-pdt-seal');
        showToast(`✍️ Section approuvée et scellée v1.0 avec empreinte SHA-256 : ${data.signature_seal ? data.signature_seal.substring(0, 16) + '...' : ''}`, "success");

        if (typeof switchPreviewSubTab === 'function') switchPreviewSubTab('wf');
        if (typeof selectDossier === 'function') {
            await selectDossier(currentWfDossierId);
            if (typeof selectDossierItem === 'function' && currentWfItemId) {
                selectDossierItem(currentWfItemId);
            }
        }
    } catch (err) {
        errBox.textContent = err.message || "Erreur réseau";
        errBox.style.display = 'block';
    }
}

// ----------------------------------------------------------------------------
// Historique des Revues & Audit Trail de l'item
// ----------------------------------------------------------------------------
async function openItemReviewsModal(dossierId, itemId, itemCode) {
    const modal = document.getElementById('modal-workflow-reviews-history');
    if (!modal) return;

    document.getElementById('wf-hist-code').textContent = itemCode;
    const body = document.getElementById('wf-hist-body');
    body.innerHTML = `<div style="text-align:center; padding:20px; color:var(--text-dim);">Chargement du journal d'audit...</div>`;
    openModal('modal-workflow-reviews-history');

    try {
        const resp = await fetch(`/api/v1/dossiers/${dossierId}/items/${itemId}/reviews`);
        const data = await resp.json();
        if (!resp.ok) {
            body.innerHTML = `<div style="color:var(--accent-red); padding:10px;">${data.message || 'Erreur'}</div>`;
            return;
        }

        const reviews = Array.isArray(data) ? data : (data.reviews || []);
        if (!reviews || reviews.length === 0) {
            body.innerHTML = `<div style="text-align:center; padding:30px; color:var(--text-muted);">
                <div style="font-size:24px; margin-bottom:8px;">📜</div>
                <div>Aucune revue enregistrée pour cette section.</div>
            </div>`;
            return;
        }

        body.innerHTML = reviews.map(r => {
            const isApproved = r.review_type === 'approval' || r.decision === 'approved';
            const badge = isApproved 
                ? `<span class="status-tag status-ok">APPROUVÉ</span>` 
                : `<span class="status-tag status-rej">REJETÉ / REPRISE</span>`;
            
            const reviewerName = r.author_name || (r.reviewer ? (r.reviewer.first_name + ' ' + r.reviewer.last_name) : 'Réviseur');
            const reviewerRole = r.author_role || r.reviewer_role || 'AQ';
            const commentText = r.comment || r.comments || '';

            return `
                <div style="border-left:3px solid ${isApproved ? 'var(--accent-emerald)' : 'var(--accent-red)'}; background:var(--bg-panel); border-radius:4px; padding:10px 14px; margin-bottom:10px;">
                    <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:6px;">
                        <div style="display:flex; align-items:center; gap:8px;">
                            ${badge}
                            <strong style="color:var(--text-main); font-size:12px;">${escapeHTML(reviewerName)}</strong>
                            <span style="font-size:10px; color:var(--text-dim); background:var(--bg-sub); padding:1px 6px; border-radius:3px;">${escapeHTML(reviewerRole)}</span>
                        </div>
                        <div style="font-size:10.5px; color:var(--text-dim); font-family:monospace;">
                            ${new Date(r.created_at).toLocaleString()}
                        </div>
                    </div>
                    <div style="font-size:11.5px; color:var(--text-main); margin-bottom:4px; line-height:1.4;">
                        ${escapeHTML(commentText)}
                    </div>
                    ${r.required_actions ? `
                        <div style="font-size:11px; color:var(--accent-red); background:rgba(239,68,68,0.06); padding:4px 8px; border-radius:4px; margin-top:4px;">
                            <strong>Actions requises :</strong> ${escapeHTML(r.required_actions)}
                        </div>
                    ` : ''}
                </div>
            `;
        }).join('');
    } catch (err) {
        body.innerHTML = `<div style="color:var(--accent-red); padding:10px;">${err.message || 'Erreur réseau'}</div>`;
    }
}

// ----------------------------------------------------------------------------
// Templates Maîtres ANPP (Catalogue & Consultation)
// ----------------------------------------------------------------------------
// Master Document Templates BPF & ICH (Section 7.4.4 du CDC)
// ----------------------------------------------------------------------------
let cachedMasterTemplates = [];
let currentTplCategory = 'ALL';

function filterMasterTemplates(cat) {
    currentTplCategory = cat;
    document.querySelectorAll('.btn-tpl-cat').forEach(b => b.classList.remove('active'));
    const btn = document.getElementById(`btn-tpl-cat-${cat}`);
    if (btn) btn.classList.add('active');
    renderMasterTemplatesCards();
}

function renderMasterTemplatesCards() {
    const listContainer = document.getElementById('master-templates-cards') || document.getElementById('master-templates-list');
    if (!listContainer) return;

    let filtered = cachedMasterTemplates;
    if (currentTplCategory && currentTplCategory !== 'ALL') {
        filtered = cachedMasterTemplates.filter(t => t.category === currentTplCategory);
    }

    if (filtered.length === 0) {
        listContainer.innerHTML = `<div style="text-align:center; padding:30px; color:var(--text-muted); font-size:12px;">Aucun modèle maître dans cette catégorie.</div>`;
        return;
    }

    listContainer.innerHTML = filtered.map((t) => {
        const sectionCode = t.sub_module || t.section_code || '';
        const title = t.title_fr || t.template_name || '';
        const refCode = t.code || t.template_code || '';
        let dept = t.target_department;
        if (!dept) {
            if (t.category === 'qualite_cmc' || sectionCode === '1.P.9' || sectionCode === '3.2.P.8' || sectionCode === '3.2.P.5') dept = 'QC';
            else if (t.category === 'production' || sectionCode === '1.P.8' || sectionCode === '3.2.P.3.5') dept = 'PROD';
            else if (sectionCode === '1.H.10' || sectionCode === '1.P.1' || sectionCode === '1.P.2') dept = 'RA';
            else dept = 'QA';
        }

        return `
        <div style="background:var(--bg-panel); border:1px solid var(--border-color); border-radius:6px; padding:12px; margin-bottom:12px;">
            <div style="display:flex; justify-content:space-between; align-items:flex-start; margin-bottom:6px;">
                <div>
                    <span class="badge" style="background:var(--accent-blue); color:white; font-family:monospace; margin-right:6px;">${escapeHTML(sectionCode)}</span>
                    <strong style="color:var(--text-main); font-size:13px;">${escapeHTML(title)}</strong>
                    <span style="font-size:10px; color:var(--text-dim); margin-left:6px;">(Réf: ${escapeHTML(refCode)} | Guide CTD 2025)</span>
                    ${t.title_ar ? `<div style="font-size:11.5px; color:var(--text-muted); direction:rtl; margin-top:3px; font-weight:600;">${escapeHTML(t.title_ar)}</div>` : ''}
                </div>
                <div>
                    ${formatDepartmentBadge(dept)}
                </div>
            </div>
            <div style="font-size:11.5px; color:var(--text-muted); margin-bottom:8px;">
                Trame institutionnelle BPF conforme aux exigences du Guide CTD ANPP Édition 2025 et référentiels internationaux ICH.
            </div>
            <div style="display:flex; justify-content:space-between; align-items:center;">
                <span style="font-size:10.5px; color:var(--text-dim);">Standard opposable ANPP / ICH • Catégorie : <code>${escapeHTML(t.category)}</code></span>
                <button class="btn-pill" style="font-size:11px; padding:3px 10px; background:var(--accent-emerald); color:white; border:none;" onclick="previewMasterTemplate('${t.id}')">
                    👁️ Consulter la Trame BPF
                </button>
            </div>
        </div>
        `;
    }).join('');
}

async function openMasterTemplatesModal() {
    const modal = document.getElementById('modal-master-templates');
    if (!modal) return;

    currentTplCategory = 'ALL';
    openModal('modal-master-templates');

    const cardsContainer = document.getElementById('master-templates-cards') || document.getElementById('master-templates-list');
    if (cardsContainer) {
        cardsContainer.innerHTML = `<div style="text-align:center; padding:20px; color:var(--text-dim);">Chargement du catalogue des templates maîtres...</div>`;
    }

    try {
        const resp = await fetch('/api/v1/master-templates');
        const data = await resp.json();
        if (!resp.ok) {
            if (cardsContainer) cardsContainer.innerHTML = `<div style="color:var(--accent-red); padding:10px;">${data.message || 'Erreur'}</div>`;
            return;
        }

        cachedMasterTemplates = Array.isArray(data) ? data : (data.templates || []);
        renderMasterTemplatesCards();
    } catch (err) {
        if (cardsContainer) cardsContainer.innerHTML = `<div style="color:var(--accent-red); padding:10px;">${err.message || 'Erreur réseau'}</div>`;
    }
}

let currentActiveTemplateRaw = '';

function substituteDossierVariables(text) {
    if (!text) return '';
    let res = text;
    const d = (typeof currentDossierDetail !== 'undefined') ? currentDossierDetail : null;
    if (d) {
        const med = d.medicine || {};
        const dm = d.medical_device || {};
        const prodName = med.brand_name_fr || dm.commercial_name_fr || 'Produit de Référence';
        const inn = med.dci_fr || dm.generic_name_fr || 'Substance Active';
        const dosage = med.dosage_fr || '';
        const form = med.pharmaceutical_form_fr || 'Forme galénique standard';
        const dCode = d.dossier_code || 'DOS-2026-CTD';
        const today = new Date().toLocaleDateString('fr-FR');

        res = res
            .replace(/\{\{PRODUCT_NAME\}\}/g, prodName)
            .replace(/\{\{INN\}\}/g, inn)
            .replace(/\{\{DOSAGE\}\}/g, dosage)
            .replace(/\{\{FORM\}\}/g, form)
            .replace(/\{\{DOSSIER_CODE\}\}/g, dCode)
            .replace(/\{\{ESTABLISHMENT_NAME\}\}/g, 'Laboratoire SANAD PHARMA Algérie')
            .replace(/\{\{AGREEMENT_NUM\}\}/g, 'MIPH-DZ-2024-0089')
            .replace(/\{\{SITE_NAME\}\}/g, 'Site Industriel SANAD Oued Smar')
            .replace(/\{\{SITE_ADDRESS\}\}/g, 'Zone Industrielle Oued Smar, Alger')
            .replace(/\{\{PDT_NAME\}\}/g, 'Dr. Yacine Benkhalid')
            .replace(/\{\{REVISION_DATE\}\}/g, today)
            .replace(/\{\{EVALUATION_DATE\}\}/g, today)
            .replace(/\{\{ROUTE_OF_ADMINISTRATION\}\}/g, 'orale')
            .replace(/\{\{SHELF_LIFE_MONTHS\}\}/g, '36')
            .replace(/\{\{SHELF_LIFE_PROPOSED\}\}/g, '36')
            .replace(/\{\{SHELF_LIFE_YEARS\}\}/g, '3')
            .replace(/\{\{BATCH_1\}\}/g, 'LOT-2026-001')
            .replace(/\{\{BATCH_2\}\}/g, 'LOT-2026-002')
            .replace(/\{\{BATCH_3\}\}/g, 'LOT-2026-003')
            .replace(/\{\{BATCH_SIZE\}\}/g, '100 000')
            .replace(/\{\{INDICATIONS\}\}/g, 'Infections sensibles aux principes actifs selon l\'AMM')
            .replace(/\{\{PHARMACO_CLASS\}\}/g, 'Médicament conforme aux normes ANPP')
            .replace(/\{\{BLISTER_TYPE\}\}/g, 'Alu/PVDC haute barrière');
    }
    return res;
}

function toggleTemplateSubstitution() {
    const chk = document.getElementById('chk-auto-substitute');
    const textarea = document.getElementById('tpl-detail-content');
    if (!textarea) return;
    if (chk && chk.checked) {
        textarea.value = substituteDossierVariables(currentActiveTemplateRaw);
    } else {
        textarea.value = currentActiveTemplateRaw;
    }
}

function previewMasterTemplate(tplId) {
    const t = cachedMasterTemplates.find(x => x.id === tplId);
    if (!t) return;

    const modal = document.getElementById('modal-master-template-detail');
    if (!modal) return;

    const sectionCode = t.sub_module || t.section_code || '';
    const title = t.title_fr || t.template_name || '';

    currentActiveTemplateRaw = t.content_template || '';

    document.getElementById('tpl-detail-code').textContent = sectionCode;
    document.getElementById('tpl-detail-name').textContent = title;

    const chk = document.getElementById('chk-auto-substitute');
    if (chk && chk.checked) {
        document.getElementById('tpl-detail-content').value = substituteDossierVariables(currentActiveTemplateRaw);
    } else {
        document.getElementById('tpl-detail-content').value = currentActiveTemplateRaw;
    }

    openModal('modal-master-template-detail');
}

function copyTemplateContentToClipboard() {
    const textarea = document.getElementById('tpl-detail-content');
    if (!textarea) return;
    textarea.select();
    navigator.clipboard.writeText(textarea.value);
    showToast("📋 Trame BPF copiée dans le presse-papier.", "success");
}

// ----------------------------------------------------------------------------
// Moteur Universel des États Documentaires BPF (ES 1-20 & EI 1-8)
// ----------------------------------------------------------------------------
let currentReportMode = 'ES';

function setReportMode(mode) {
    currentReportMode = mode;
    const btnES = document.getElementById('report-mode-es');
    const btnEI = document.getElementById('report-mode-ei');
    const desc = document.getElementById('report-category-description');
    const sealBtn = document.getElementById('btn-seal-active-report');

    if (btnES) btnES.classList.toggle('active', mode === 'ES');
    if (btnEI) btnEI.classList.toggle('active', mode === 'EI');

    if (mode === 'ES') {
        if (desc) desc.textContent = "États réglementaires opposables ANPP avec cachet 2FA et horodatage indélébile.";
        if (sealBtn) sealBtn.style.display = 'inline-block';
        currentReportCode = 'ES-REG-01';
    } else {
        if (desc) desc.textContent = "Documents internes de travail et matrices de pilotage BPF (non opposables).";
        if (sealBtn) sealBtn.style.display = 'none';
        currentReportCode = 'EI-AVT-01';
    }

    renderReportTabs();
    switchReportTab(currentReportCode);
}

function renderReportTabs() {
    const tabsBar = document.getElementById('report-tabs-bar');
    if (!tabsBar) return;

    let isDM = false;
    if (typeof currentDossierDetail !== 'undefined' && currentDossierDetail && currentDossierDetail.id === currentWfDossierId) {
        if (currentDossierDetail.product_type === 'dispositif_medical' || currentDossierDetail.procedure_code === 'HOMOLOG_DM' || currentDossierDetail.medical_device) {
            isDM = true;
        }
    } else if (typeof currentDossiers !== 'undefined' && currentDossiers) {
        const found = currentDossiers.find(d => d.id === currentWfDossierId);
        if (found && (found.product_type === 'dispositif_medical' || found.procedure_code === 'HOMOLOG_DM')) {
            isDM = true;
        }
    }

    if (currentReportMode === 'ES') {
        if (isDM) {
            tabsBar.innerHTML = `
                <button class="btn-pill report-tab-btn" data-code="ES-REG-01" onclick="switchReportTab('ES-REG-01')">ES-REG-01 : Bordereau Dépôt</button>
                <button class="btn-pill report-tab-btn" data-code="ES-REG-02" onclick="switchReportTab('ES-REG-02')">ES-REG-02-DM : Demande (6 sections)</button>
                <button class="btn-pill report-tab-btn" data-code="ES-REG-03-DM" onclick="switchReportTab('ES-REG-03-DM')">ES-REG-03-DM : Récépissé Dépôt</button>
                <button class="btn-pill report-tab-btn" data-code="ES-REG-05" onclick="switchReportTab('ES-REG-05')">ES-REG-05 : Échantillons</button>
                <button class="btn-pill report-tab-btn" data-code="ES-FIN-01" onclick="switchReportTab('ES-FIN-01')">ES-FIN-01 : BV Initial BNA</button>
                <button class="btn-pill report-tab-btn" data-code="ES-FIN-02" onclick="switchReportTab('ES-FIN-02')">ES-FIN-02 : BV Modif/Renouv</button>
            `;
        } else {
            tabsBar.innerHTML = `
                <button class="btn-pill report-tab-btn" data-code="ES-REG-01" onclick="switchReportTab('ES-REG-01')">ES-REG-01 : Bordereau CTD</button>
                <button class="btn-pill report-tab-btn" data-code="ES-REG-02" onclick="switchReportTab('ES-REG-02')">ES-REG-02 : Formulaire 1.A.3</button>
                <button class="btn-pill report-tab-btn" data-code="ES-REG-03" onclick="switchReportTab('ES-REG-03')">ES-REG-03 : Index CTD (M1-M5)</button>
                <button class="btn-pill report-tab-btn" data-code="ES-REG-04" onclick="switchReportTab('ES-REG-04')">ES-REG-04 : Variation LD</button>
                <button class="btn-pill report-tab-btn" data-code="ES-REG-05" onclick="switchReportTab('ES-REG-05')">ES-REG-05 : Échantillons LNCPP</button>
                <button class="btn-pill report-tab-btn" data-code="ES-FIN-01" onclick="switchReportTab('ES-FIN-01')">ES-FIN-01 : BV1 Initial</button>
                <button class="btn-pill report-tab-btn" data-code="ES-FIN-02" onclick="switchReportTab('ES-FIN-02')">ES-FIN-02 : BV2 Variations</button>
                <button class="btn-pill report-tab-btn" data-code="1.P.8" onclick="switchReportTab('1.P.8')">1.P.8 : Stock 3 mois</button>
                <button class="btn-pill report-tab-btn" data-code="1.P.9" onclick="switchReportTab('1.P.9')">1.P.9 : Vigilance</button>
                <button class="btn-pill report-tab-btn" data-code="1.H.10" onclick="switchReportTab('1.H.10')">1.H.10 : Non-Modif.</button>
            `;
        }
    } else {
        tabsBar.innerHTML = `
            <button class="btn-pill report-tab-btn" data-code="EI-AVT-01" onclick="switchReportTab('EI-AVT-01')">EI-AVT-01 : Navette CTD/DM</button>
            <button class="btn-pill report-tab-btn" data-code="EI-AVT-02" onclick="switchReportTab('EI-AVT-02')">EI-AVT-02 : Relevé Écarts</button>
            <button class="btn-pill report-tab-btn" data-code="EI-AVT-03" onclick="switchReportTab('EI-AVT-03')">EI-AVT-03 : Fiche Checker</button>
            <button class="btn-pill report-tab-btn" data-code="EI-CHR-01" onclick="switchReportTab('EI-CHR-01')">EI-CHR-01 : Stabilités Zone IVb</button>
            <button class="btn-pill report-tab-btn" data-code="EI-CHR-02" onclick="switchReportTab('EI-CHR-02')">EI-CHR-02 : Bioéquivalence</button>
            <button class="btn-pill report-tab-btn" data-code="EI-VIG-01" onclick="switchReportTab('EI-VIG-01')">EI-VIG-01 : Plan d'Action Carence</button>
            <button class="btn-pill report-tab-btn" data-code="EI-CHG-01" onclick="switchReportTab('EI-CHG-01')">EI-CHG-01 : Change Control</button>
            <button class="btn-pill report-tab-btn" data-code="EI-PIP-01" onclick="switchReportTab('EI-PIP-01')">EI-PIP-01 : Pipeline CoDir</button>
        `;
    }
}

function openDossierReportsModal(dossierId) {
    currentWfDossierId = dossierId;
    currentReportMode = 'ES';
    currentReportCode = 'ES-REG-01';

    const modal = document.getElementById('modal-dossier-reports');
    if (!modal) return;

    const btnES = document.getElementById('report-mode-es');
    const btnEI = document.getElementById('report-mode-ei');
    const desc = document.getElementById('report-category-description');
    const sealBtn = document.getElementById('btn-seal-active-report');
    if (btnES) btnES.classList.add('active');
    if (btnEI) btnEI.classList.remove('active');
    if (desc) desc.textContent = "États réglementaires opposables ANPP avec cachet 2FA et horodatage indélébile.";
    if (sealBtn) sealBtn.style.display = 'inline-block';

    renderReportTabs();
    openModal('modal-dossier-reports');
    switchReportTab(currentReportCode);
}

function switchReportTab(code) {
    currentReportCode = code;
    document.querySelectorAll('.report-tab-btn').forEach(b => {
        b.classList.toggle('active', b.dataset.code === code);
    });
    loadReportView(code);
}

async function loadReportView(reportCode) {
    const frameContainer = document.getElementById('report-iframe-container');
    const sealStatusBox = document.getElementById('report-seal-status-box');
    if (!frameContainer) return;

    frameContainer.innerHTML = `<div style="text-align:center; padding:40px; color:var(--text-dim);">
        <div style="font-size:24px; margin-bottom:10px;">⏳</div>
        <div>Compilation certifiée du document BPF / ANPP...</div>
    </div>`;

    try {
        const url = `/api/v1/dossiers/${currentWfDossierId}/reports/${reportCode}?format=html`;
        frameContainer.innerHTML = `
            <iframe id="report-active-iframe" src="${url}" style="width:100%; height:550px; border:1px solid var(--border-color); border-radius:4px; background:white;"></iframe>
        `;

        if (sealStatusBox) {
            if (currentReportMode === 'EI') {
                sealStatusBox.innerHTML = `
                    <span style="color:#0284c7; font-size:11px; font-weight:600;">ℹ️ Document de travail interne — Filigrane de confidentialité BPF appliqué (Pas de visa PDT requis)</span>
                `;
            } else {
                sealStatusBox.innerHTML = `
                    <span style="color:var(--text-dim); font-size:11px;">État : Non scellé formellement (Aperçu de travail opposable)</span>
                `;
            }
        }
    } catch (err) {
        frameContainer.innerHTML = `<div style="color:var(--accent-red); padding:20px;">${err.message || 'Erreur'}</div>`;
    }
}

function printActiveReport() {
    const iframe = document.getElementById('report-active-iframe');
    if (iframe && iframe.contentWindow) {
        iframe.contentWindow.focus();
        iframe.contentWindow.print();
    } else {
        window.open(`/api/v1/dossiers/${currentWfDossierId}/reports/${currentReportCode}?format=html`, '_blank');
    }
}

function downloadActiveReportJSON() {
    window.open(`/api/v1/dossiers/${currentWfDossierId}/reports/${currentReportCode}?format=json`, '_blank');
}

async function sealActiveReport() {
    if (!confirm(`Confirmer le scellement formel de l'état ${currentReportCode} avec signature numérique Pharmacien DT et horodatage certifié ?`)) {
        return;
    }

    try {
        const resp = await fetch(`/api/v1/dossiers/${currentWfDossierId}/reports/${currentReportCode}/seal`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ solemn_declaration: "Conformité ANPP garantie" })
        });
        const data = await resp.json();
        if (!resp.ok) {
            alert(data.message || "Erreur lors du scellement");
            return;
        }

        showToast(`🔏 Rapport scellé avec succès ! Empreinte SHA-256 : ${data.sha256_hash ? data.sha256_hash.substring(0, 16) + '...' : ''}`, "success");
        loadReportView(currentReportCode);
    } catch (err) {
        alert(err.message || "Erreur réseau");
    }
}

// ============================================================================
// Jalon 2.1b : Fonctions UI Gating & Chemin Critique Réglementaire
// ============================================================================

function showGatingBlockedAlert(data) {
    const titleEl = document.getElementById('gating-modal-sub');
    const contentEl = document.getElementById('gating-modal-content');
    if (!contentEl) {
        alert(data.error || "Action bloquée par verrou réglementaire");
        return;
    }

    if (titleEl) {
        titleEl.textContent = "Tentative de transition rejetée par le contrôle des dépendances BPF";
    }

    const unfulfilled = data.unmet_prerequisites || (data.evaluation && data.evaluation.unmet_prerequisites) || [];

    contentEl.innerHTML = `
        <div style="background:rgba(239,68,68,0.08); border:1.5px solid #f87171; border-radius:6px; padding:12px; margin-bottom:14px;">
            <div style="font-size:13px; font-weight:800; color:#b91c1c; margin-bottom:6px; display:flex; align-items:center; gap:6px;">
                <span>⛔</span>
                <span>TRANSITION BLOQUÉE PAR RÈGLE RÉGLEMENTAIRE (GATING)</span>
            </div>
            <div style="font-size:12px; color:var(--text-main); line-height:1.5;">
                ${escapeHTML(data.error || "Cette pièce ne peut franchir cette étape tant que ses dépendances critiques ne sont pas résolues.")}
            </div>
        </div>

        <div style="font-size:11.5px; font-weight:700; text-transform:uppercase; color:var(--text-dim); margin-bottom:8px;">
            Prérequis Réglementaires Non Satisfaits (${unfulfilled.length}) :
        </div>

        <div style="display:flex; flex-direction:column; gap:8px;">
            ${unfulfilled.map(u => `
                <div style="background:var(--bg-panel); border:1px solid var(--border-color); border-left:3px solid #ef4444; border-radius:4px; padding:10px 12px;">
                    <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:4px;">
                        <span style="font-family:monospace; font-weight:800; color:var(--accent-blue); font-size:12px;">
                            [${escapeHTML(u.required_item_code)}] ${escapeHTML(u.required_title || '')}
                        </span>
                        <span class="status-tag status-rej" style="font-size:10px;">
                            Statut actuel : ${escapeHTML(u.current_status || 'non scellé')}
                        </span>
                    </div>
                    <div style="font-size:11.5px; color:var(--text-main); margin-bottom:6px;">
                        ${escapeHTML(u.blocking_reason)}
                    </div>
                    <div style="font-size:10.5px; color:var(--text-dim); background:var(--bg-sub); padding:4px 8px; border-radius:3px;">
                        ⚖️ <strong>Fondement légal :</strong> ${escapeHTML(u.legal_ref)}
                    </div>
                </div>
            `).join('')}
        </div>

        <div style="margin-top:14px; font-size:11px; color:var(--text-dim); font-style:italic; border-top:1px solid var(--border-color); padding-top:8px;">
            💡 <strong>Action requise :</strong> Complétez, relisez et scellez la ou les pièces prérequises listées ci-dessus avant de renouveler cette soumission.
        </div>
    `;

    openModal('modal-workflow-gating-info');
}

async function showGatingInfoModal(itemId, itemCode) {
    const dossierId = currentWfDossierId || (typeof currentDossierDetail === 'object' && currentDossierDetail ? currentDossierDetail.id : null);
    if (!dossierId) return;

    const titleEl = document.getElementById('gating-modal-sub');
    const contentEl = document.getElementById('gating-modal-content');
    if (!contentEl) return;

    if (titleEl) {
        titleEl.textContent = `Diagnostic du chemin critique pour la section [${itemCode}]`;
    }

    contentEl.innerHTML = `<div style="text-align:center; padding:20px; color:var(--text-dim);">Analyse des dépendances en cours...</div>`;
    openModal('modal-workflow-gating-info');

    try {
        const resp = await fetch(`/api/v1/dossiers/${dossierId}/items/${itemId}/gating`);
        if (!resp.ok) throw new Error("HTTP " + resp.status);
        const data = await resp.json();
        const evalRes = data.evaluation;

        if (evalRes.is_blocked) {
            contentEl.innerHTML = `
                <div style="background:rgba(239,68,68,0.08); border:1.5px solid #f87171; border-radius:6px; padding:12px; margin-bottom:14px;">
                    <div style="font-size:13px; font-weight:800; color:#b91c1c; margin-bottom:6px; display:flex; align-items:center; gap:6px;">
                        <span>🔒</span>
                        <span>SECTION ACTUELLEMENT VERROUILLÉE</span>
                    </div>
                    <div style="font-size:11.5px; color:var(--text-main); line-height:1.5;">
                        ${escapeHTML(evalRes.summary_message)}
                    </div>
                </div>

                <div style="font-size:11.5px; font-weight:700; text-transform:uppercase; color:var(--text-dim); margin-bottom:8px;">
                    Détail des Verrous &amp; Dépendances BPF :
                </div>

                <div style="display:flex; flex-direction:column; gap:8px;">
                    ${(evalRes.unmet_prerequisites || []).map(u => `
                        <div style="background:var(--bg-panel); border:1px solid var(--border-color); border-left:3px solid #ef4444; border-radius:4px; padding:10px 12px;">
                            <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:4px;">
                                <span style="font-family:monospace; font-weight:800; color:var(--accent-blue); font-size:12px;">
                                    [${escapeHTML(u.required_item_code)}]
                                </span>
                                <span class="status-tag status-rej" style="font-size:10px;">
                                    Exigé : ${escapeHTML(u.expected_condition)} (Actuel: ${escapeHTML(u.current_status)})
                                </span>
                            </div>
                            <div style="font-size:11.5px; color:var(--text-main); margin-bottom:6px;">
                                ${escapeHTML(u.blocking_reason)}
                            </div>
                            <div style="font-size:10.5px; color:var(--text-dim); background:var(--bg-sub); padding:4px 8px; border-radius:3px;">
                                ⚖️ <strong>Réf :</strong> ${escapeHTML(u.legal_ref)}
                            </div>
                        </div>
                    `).join('')}
                </div>
            `;
        } else {
            contentEl.innerHTML = `
                <div style="background:rgba(16,185,129,0.08); border:1.5px solid #10b981; border-radius:6px; padding:14px; text-align:center;">
                    <div style="font-size:28px; margin-bottom:8px;">🔓</div>
                    <div style="font-size:13px; font-weight:800; color:var(--accent-emerald); margin-bottom:6px;">
                        SECTION TOTALEMENT ÉLIGIBLE (AUCUN VERROU ACTIF)
                    </div>
                    <div style="font-size:11.5px; color:var(--text-main); line-height:1.5;">
                        Toutes les dépendances réglementaires et chronologiques de la section [${escapeHTML(itemCode)}] sont validées. Le document peut être rédigé, soumis ou approuvé conformément à votre rôle habilité.
                    </div>
                </div>
            `;
        }
    } catch(err) {
        contentEl.innerHTML = `<div style="color:var(--accent-red); padding:16px;">Erreur lors du chargement des verrous : ${escapeHTML(err.message)}</div>`;
    }
}

async function showDossierGatingSummaryModal() {
    const dossierId = currentWfDossierId || (typeof currentDossierDetail === 'object' && currentDossierDetail ? currentDossierDetail.id : null);
    if (!dossierId) return;

    const titleEl = document.getElementById('gating-modal-sub');
    const contentEl = document.getElementById('gating-modal-content');
    if (!contentEl) return;

    if (titleEl) {
        titleEl.textContent = "Synthèse Globale du Chemin Critique & Progression DAG (5 Phases CTD)";
    }

    contentEl.innerHTML = `<div style="text-align:center; padding:20px; color:var(--text-dim);">Calcul du chemin critique...</div>`;
    openModal('modal-workflow-gating-info');

    try {
        const resp = await fetch(`/api/v1/dossiers/${dossierId}/gating-summary`);
        if (!resp.ok) throw new Error("HTTP " + resp.status);
        const data = await resp.json();
        const s = data.summary;

        const phases = [
            { num: 1, name: "Méthodes Analytiques & Spécifications Matières", desc: "Spécifications API (3.2.S) et Méthodes de contrôle Produit Fini (3.2.P.5)" },
            { num: 2, name: "Validation Procédé & Lots Pilotes", desc: "Validation de fabrication (3.2.P.3.5) et Certificats d'Analyse 3 lots (1.P.4 / 1.P.5)" },
            { num: 3, name: "Stabilités Zone IVb & Bioéquivalence", desc: "Données de stabilité M6 30°C/75% RH (3.2.P.8) et Études cliniques / BE (Module 5.3)" },
            { num: 4, name: "QOS (Module 2.3), Notice & RCP", desc: "Notice Patient FR & AR (1.P.2), RCP (1.P.1) et Résumé Global Qualité (2.3)" },
            { num: 5, name: "Engagements Réglementaires & Dépôt Légal PDT", desc: "Engagements de commercialisation (1.P.3), Quittance redevance et Dépôt officiel" }
        ];

        contentEl.innerHTML = `
            <!-- Frise chronologique des 5 phases DAG -->
            <div style="margin-bottom:16px; background:var(--bg-app); border:1px solid var(--border-color); border-radius:6px; padding:12px;">
                <div style="font-size:11px; font-weight:800; text-transform:uppercase; color:var(--text-dim); margin-bottom:10px;">
                    Progression du Dossier dans le Graphe de Dépendances (DAG) :
                </div>
                <div style="display:flex; flex-direction:column; gap:6px;">
                    ${phases.map(p => {
                        const isDone = p.num < s.active_phase;
                        const isCurrent = p.num === s.active_phase;
                        return `
                            <div style="display:flex; align-items:center; gap:8px; padding:6px 8px; border-radius:4px; ${isCurrent ? 'background:rgba(37,99,235,0.12); border:1px solid var(--accent-blue);' : (isDone ? 'background:rgba(16,185,129,0.08);' : 'background:transparent; opacity:0.6;')}">
                                <span style="font-size:14px;">${isDone ? '✅' : (isCurrent ? '▶️' : '⚪')}</span>
                                <div style="flex:1;">
                                    <div style="font-size:11.5px; font-weight:${isCurrent ? '800' : '600'}; color:${isCurrent ? 'var(--accent-blue)' : (isDone ? 'var(--accent-emerald)' : 'var(--text-muted)')};">
                                        Phase ${p.num} : ${p.name}
                                    </div>
                                    <div style="font-size:10px; color:var(--text-dim);">${p.desc}</div>
                                </div>
                                ${isCurrent ? '<span style="font-size:9.5px; background:var(--accent-blue); color:white; font-weight:800; padding:1px 6px; border-radius:8px;">EN COURS</span>' : ''}
                                ${isDone ? '<span style="font-size:9.5px; color:var(--accent-emerald); font-weight:700;">VALIDÉ</span>' : ''}
                            </div>
                        `;
                    }).join('')}
                </div>
            </div>

            <!-- Liste des sections actuellement bloquées -->
            <div style="font-size:11.5px; font-weight:700; text-transform:uppercase; color:var(--text-dim); margin-bottom:8px;">
                Sections Actuellement Bloquées par Verrous (${s.blocked_count}) :
            </div>

            ${s.blocked_count === 0 ? `
                <div style="background:rgba(16,185,129,0.08); border:1px solid rgba(16,185,129,0.3); border-radius:6px; padding:12px; text-align:center; color:var(--accent-emerald); font-weight:600; font-size:12px;">
                    ✅ Aucun verrou bloquant ! Toutes les pièces requises pour la phase active sont satisfaites.
                </div>
            ` : `
                <div style="display:flex; flex-direction:column; gap:8px;">
                    ${(s.blocked_items || []).map(b => `
                        <div style="background:var(--bg-panel); border:1px solid var(--border-color); border-left:3px solid #ef4444; border-radius:4px; padding:8px 12px;">
                            <div style="font-weight:800; color:var(--accent-blue); font-size:12px; margin-bottom:4px;">
                                [${escapeHTML(b.target_item_code)}] ${escapeHTML(b.target_item_title || '')}
                            </div>
                            <div style="font-size:11px; color:var(--text-main); margin-bottom:4px;">
                                ${escapeHTML(b.summary_message)}
                            </div>
                            <div style="font-size:10px; color:var(--text-dim);">
                                ${(b.unmet_prerequisites || []).map(u => `
                                    <div>• Requiert <strong>[${escapeHTML(u.required_item_code)}]</strong> en statut <code>${escapeHTML(u.expected_condition)}</code></div>
                                `).join('')}
                            </div>
                        </div>
                    `).join('')}
                </div>
            `}
        `;
    } catch(err) {
        contentEl.innerHTML = `<div style="color:var(--accent-red); padding:16px;">Erreur lors du calcul du chemin critique : ${escapeHTML(err.message)}</div>`;
    }
}

function showBlockingPrereqModal(itemId, itemCode, blocksCount, blockedCodesStr) {
    const titleEl = document.getElementById('gating-modal-sub');
    const contentEl = document.getElementById('gating-modal-content');
    if (!contentEl) return;

    if (titleEl) {
        titleEl.textContent = `Tâche Prioritaire & Débloqueur Critique : Section [${itemCode}]`;
    }

    contentEl.innerHTML = `
        <div style="background:rgba(245,158,11,0.08); border:1.5px solid #f59e0b; border-radius:6px; padding:12px; margin-bottom:14px;">
            <div style="font-size:13px; font-weight:800; color:#b45309; margin-bottom:6px; display:flex; align-items:center; gap:6px;">
                <span>⚡</span>
                <span>PIÈCE CLÉ DE VOÛTE (PRÉREQUIS BLOQUANT DU CHEMIN CRITIQUE)</span>
            </div>
            <div style="font-size:12px; color:var(--text-main); line-height:1.5;">
                Cette section constitue un <strong>jalon bloquant amont</strong>. Tant qu'elle n'est pas rédigée, validée par le Checker et scellée en v1.0 par le Pharmacien DT, <strong>${blocksCount} autre(s) section(s)</strong> du dossier restent en attente forcée.
            </div>
        </div>

        <div style="font-size:11.5px; font-weight:700; text-transform:uppercase; color:var(--text-dim); margin-bottom:8px;">
            Rubriques En Attente Du Déverrouillage Par Cette Pièce (${blocksCount}) :
        </div>

        <div style="background:var(--bg-panel); border:1px solid var(--border-color); border-left:3.5px solid #f59e0b; border-radius:4px; padding:10px 12px; margin-bottom:14px;">
            <div style="font-family:monospace; font-size:11.5px; font-weight:700; color:var(--text-main); word-break:break-word; line-height:1.6;">
                ${escapeHTML(blockedCodesStr)}
            </div>
        </div>

        <div style="background:var(--bg-app); border:1px solid var(--border-color); border-radius:6px; padding:10px 12px; font-size:11px; color:var(--text-dim); line-height:1.5;">
            🎯 <strong>Recommandation Chef de Projet BPF :</strong> Traitez en priorité absolue cette section (Check-out ➔ Rédaction ➔ Revue Checker ➔ Visa PDT). Son scellement débloquera automatiquement le travail des équipes sur les pièces dépendantes.
        </div>
    `;

    openModal('modal-workflow-gating-info');
}
