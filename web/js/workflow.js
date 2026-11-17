// ============================================================================
// Jalon 2.1 : Atelier Rédactionnel BPF, Maker-Checker & Moteur d'États de Sortie
// Conformité BPF / 21 CFR Part 11 / ALCOA+ / Guide ANPP 2025
// ============================================================================

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
    const due = document.getElementById('wf-assign-due').value;
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
            alert(data.message || "Erreur lors de la soumission à revue");
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

    const modal = document.getElementById('modal-workflow-review-decision');
    if (!modal) return;

    document.getElementById('wf-review-code').textContent = itemCode;
    document.getElementById('wf-review-title').textContent = title;
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
                comments: comments,
                required_actions: reqActions
            })
        });
        const data = await resp.json();
        if (!resp.ok) {
            errBox.textContent = data.message || "Erreur lors de l'enregistrement de la décision";
            errBox.style.display = 'block';
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

    const modal = document.getElementById('modal-workflow-pdt-seal');
    if (!modal) return;

    document.getElementById('wf-pdt-code').textContent = itemCode;
    document.getElementById('wf-pdt-title').textContent = title;
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
                solemn_declaration_accepted: true
            })
        });
        const data = await resp.json();
        if (!resp.ok) {
            errBox.textContent = data.message || "Erreur lors du visa PDT";
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
let cachedMasterTemplates = [];

async function openMasterTemplatesModal() {
    const modal = document.getElementById('modal-master-templates');
    if (!modal) return;

    const listContainer = document.getElementById('master-templates-list');
    listContainer.innerHTML = `<div style="text-align:center; padding:20px; color:var(--text-dim);">Chargement du catalogue des templates maîtres...</div>`;
    openModal('modal-master-templates');

    try {
        const resp = await fetch('/api/v1/master-templates');
        const data = await resp.json();
        if (!resp.ok) {
            listContainer.innerHTML = `<div style="color:var(--accent-red); padding:10px;">${data.message || 'Erreur'}</div>`;
            return;
        }

        cachedMasterTemplates = Array.isArray(data) ? data : (data.templates || []);
        if (cachedMasterTemplates.length === 0) {
            listContainer.innerHTML = `<div style="text-align:center; padding:20px; color:var(--text-muted);">Aucun template maître actif.</div>`;
            return;
        }

        listContainer.innerHTML = cachedMasterTemplates.map((t, idx) => {
            const sectionCode = t.sub_module || t.section_code || '';
            const title = t.title_fr || t.template_name || '';
            const refCode = t.code || t.template_code || '';
            let dept = t.target_department;
            if (!dept) {
                if (t.category === 'qualite_cmc' || sectionCode === '1.P.9') dept = 'QC';
                else if (t.category === 'production' || sectionCode === '1.P.8') dept = 'PROD';
                else if (sectionCode === '1.H.10') dept = 'RA';
                else dept = 'QA';
            }

            return `
            <div style="background:var(--bg-panel); border:1px solid var(--border-color); border-radius:6px; padding:12px; margin-bottom:12px;">
                <div style="display:flex; justify-content:space-between; align-items:flex-start; margin-bottom:6px;">
                    <div>
                        <span class="badge" style="background:var(--accent-blue); color:white; font-family:monospace; margin-right:6px;">${escapeHTML(sectionCode)}</span>
                        <strong style="color:var(--text-main); font-size:13px;">${escapeHTML(title)}</strong>
                        <span style="font-size:10px; color:var(--text-dim); margin-left:6px;">(Réf: ${escapeHTML(refCode)} | Guide CTD 2025)</span>
                    </div>
                    <div>
                        ${formatDepartmentBadge(dept)}
                    </div>
                </div>
                <div style="font-size:11.5px; color:var(--text-muted); margin-bottom:8px;">
                    Trame institutionnelle BPF conforme aux exigences du Guide CTD ANPP Édition 2025.
                </div>
                <div style="display:flex; justify-content:space-between; align-items:center;">
                    <span style="font-size:10.5px; color:var(--text-dim);">Standard opposable ANPP / ICH</span>
                    <button class="btn-pill" style="font-size:11px; padding:3px 10px; background:var(--accent-emerald); color:white; border:none;" onclick="previewMasterTemplate('${t.id}')">
                        👁️ Consulter la Trame BPF
                    </button>
                </div>
            </div>
            `;
        }).join('');
    } catch (err) {
        listContainer.innerHTML = `<div style="color:var(--accent-red); padding:10px;">${err.message || 'Erreur réseau'}</div>`;
    }
}

function previewMasterTemplate(tplId) {
    const t = cachedMasterTemplates.find(x => x.id === tplId);
    if (!t) return;

    const modal = document.getElementById('modal-master-template-detail');
    if (!modal) return;

    const sectionCode = t.sub_module || t.section_code || '';
    const title = t.title_fr || t.template_name || '';

    document.getElementById('tpl-detail-code').textContent = sectionCode;
    document.getElementById('tpl-detail-name').textContent = title;
    document.getElementById('tpl-detail-content').value = t.content_template || '';

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
// Moteur Dynamique des États de Sortie Réglementaires (ES-REG-01, 02, ES-FIN-01)
// ----------------------------------------------------------------------------
function openDossierReportsModal(dossierId) {
    currentWfDossierId = dossierId;
    currentReportCode = 'ES-REG-01';

    const modal = document.getElementById('modal-dossier-reports');
    if (!modal) return;

    // Détection si Dispositif Médical
    let isDM = false;
    if (typeof currentDossierDetail !== 'undefined' && currentDossierDetail && currentDossierDetail.id === dossierId) {
        if (currentDossierDetail.product_type === 'dispositif_medical' || currentDossierDetail.procedure_code === 'HOMOLOG_DM' || currentDossierDetail.medical_device) {
            isDM = true;
        }
    } else if (typeof currentDossiers !== 'undefined' && currentDossiers) {
        const found = currentDossiers.find(d => d.id === dossierId);
        if (found && (found.product_type === 'dispositif_medical' || found.procedure_code === 'HOMOLOG_DM')) {
            isDM = true;
        }
    }

    const tabsBar = document.getElementById('report-tabs-bar');
    if (tabsBar) {
        if (isDM) {
            tabsBar.innerHTML = `
                <button class="btn-pill report-tab-btn active" data-code="ES-REG-01" onclick="switchReportTab('ES-REG-01')">ES-REG-01 : Bordereau Dépôt</button>
                <button class="btn-pill report-tab-btn" data-code="ES-REG-02" onclick="switchReportTab('ES-REG-02')">ES-REG-02-DM : Demande Homologation (6 sections)</button>
                <button class="btn-pill report-tab-btn" data-code="ES-REG-03-DM" onclick="switchReportTab('ES-REG-03-DM')">ES-REG-03-DM : Récépissé Dépôt (13 points)</button>
                <button class="btn-pill report-tab-btn" data-code="ES-FIN-01" onclick="switchReportTab('ES-FIN-01')">ES-FIN-01 : Bordereau BV (BNA Chéraga)</button>
            `;
        } else {
            tabsBar.innerHTML = `
                <button class="btn-pill report-tab-btn active" data-code="ES-REG-01" onclick="switchReportTab('ES-REG-01')">ES-REG-01 : Bordereau CTD</button>
                <button class="btn-pill report-tab-btn" data-code="ES-REG-02" onclick="switchReportTab('ES-REG-02')">ES-REG-02 : Formulaire 1.A.3</button>
                <button class="btn-pill report-tab-btn" data-code="ES-FIN-01" onclick="switchReportTab('ES-FIN-01')">ES-FIN-01 : Bordereau BV1 (Trésor Public)</button>
            `;
        }
    } else {
        // Fallback
        document.querySelectorAll('.report-tab-btn').forEach(b => {
            b.classList.toggle('active', b.dataset.code === 'ES-REG-01');
        });
    }

    openModal('modal-dossier-reports');
    loadReportView(currentReportCode);
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
        <div>Compilation certifiée de l'état de sortie officiel ANPP...</div>
    </div>`;

    try {
        const url = `/api/v1/dossiers/${currentWfDossierId}/reports/${reportCode}?format=html`;
        frameContainer.innerHTML = `
            <iframe id="report-active-iframe" src="${url}" style="width:100%; height:550px; border:1px solid var(--border-color); border-radius:4px; background:white;"></iframe>
        `;

        if (sealStatusBox) {
            sealStatusBox.innerHTML = `
                <span style="color:var(--text-dim); font-size:11px;">État : Non scellé formellement (Aperçu de travail)</span>
            `;
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
