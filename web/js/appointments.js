// ============================================================================
// GESTION DES RENDEZ-VOUS & SÉANCES DE DÉPÔT ANPP
// Planification des dépôts physiques, commissions et expertises LNCM
// ============================================================================

let activeAppointmentsDossierId = null;
let activeAppointmentsOnlyCurrent = false;

async function openAppointmentsModal(forDossierId, options = {}) {
    const modal = document.getElementById("modal-appointments");
    if (!modal) return;

    activeAppointmentsDossierId = forDossierId || (typeof selectedDossierId !== 'undefined' ? selectedDossierId : null);
    activeAppointmentsOnlyCurrent = !!options.onlyCurrent;

    openModal("modal-appointments");
    renderAppointmentsDossierContext(activeAppointmentsDossierId, options);
    await loadAppointmentsList();

    if (options.prefill || options.fromSmartPlanning) {
        setTimeout(() => prefillAppointmentFormForDossier(), 120);
    }
}

function renderAppointmentsDossierContext(dossierId, options = {}) {
    const container = document.getElementById("appointments-dossier-context");
    if (!container) return;

    if (!dossierId || typeof currentDossierDetail === 'undefined' || !currentDossierDetail || currentDossierDetail.id !== dossierId) {
        container.innerHTML = '';
        container.style.display = 'none';
        return;
    }

    const d = currentDossierDetail;
    const code = d.dossier_code || d.reference || 'Dossier en cours';
    const name = d.commercial_name || d.name || 'Produit';
    const isDM = (d.items || []).some(it => (it.item_code || '').startsWith('DM.')) || (d.procedure_type || '').includes('dm');
    const targetPiece = isDM ? 'DM.A.1' : '1.A.1';
    const portalUrl = isDM ? 'rdvdm.anpp.dz' : 'rdvpp.anpp.dz';
    const portalWindow = isDM ? 'Mardi 10h00 au Jeudi 10h00' : 'Mercredi 10h00 au Jeudi 10h00';
    const depositRef = d.deposit_receipt_ref ? ` • Réf. Récépissé : <strong style="color:var(--accent-emerald);">${escapeHTML(d.deposit_receipt_ref)}</strong>` : '';

    container.style.display = 'block';
    container.innerHTML = `
        <div style="background:linear-gradient(135deg, rgba(37,99,235,0.08), rgba(16,185,129,0.08)); border:1px solid rgba(37,99,235,0.25); border-radius:6px; padding:10px 14px; display:flex; justify-content:space-between; align-items:center; flex-wrap:wrap; gap:10px;">
            <div>
                <div style="font-size:12px; font-weight:700; color:var(--text-main); display:flex; align-items:center; gap:6px; flex-wrap:wrap;">
                    <span>📁</span>
                    <span>Dossier actif : <strong>${escapeHTML(code)}</strong> — ${escapeHTML(name)}</span>
                    <span style="font-size:10px; background:rgba(37,99,235,0.15); color:var(--accent-blue); padding:1px 6px; border-radius:4px; font-weight:700;">${isDM ? 'Dispositif Médical' : 'Médicament CTD'}</span>
                    <span style="font-size:10px; background:rgba(16,185,129,0.15); color:var(--accent-emerald); padding:1px 6px; border-radius:4px; font-weight:700;">J-60 Réservation Créneau</span>
                </div>
                <div style="font-size:11px; color:var(--text-dim); margin-top:3px;">
                    Portail officiel : <strong style="color:var(--accent-blue);">${portalUrl}</strong> (Guichet : ${portalWindow})${depositRef}
                </div>
            </div>
            <div style="display:flex; align-items:center; gap:8px; flex-wrap:wrap;">
                <button type="button" class="btn-pill" style="background:var(--accent-blue); color:white; border:none; font-size:11px; font-weight:700; padding:4px 10px; cursor:pointer;" onclick="prefillAppointmentFormForDossier()" title="Pré-remplir le formulaire de créneau pour ce dossier">
                    ➕ Saisir Créneau Dépôt Initial
                </button>
                <button type="button" class="btn-pill" style="font-size:11px; padding:4px 10px; cursor:pointer; background:var(--bg-panel); border:1px solid var(--border-color);" onclick="jumpToAppointmentChecklistItem('${targetPiece}')" title="Fermer cette fenêtre et naviguer vers la pièce justificative ${targetPiece} dans la checklist">
                    📄 Aller à la pièce ${targetPiece} ➔
                </button>
                <button type="button" class="btn-pill" id="btn-toggle-appointments-filter" style="font-size:10.5px; padding:4px 8px; cursor:pointer;" onclick="toggleAppointmentsFilter()" title="Basculer entre les créneaux de ce dossier et tous les créneaux">
                    ${activeAppointmentsOnlyCurrent ? '🌐 Voir tous les dossiers' : '🎯 Ce dossier uniquement'}
                </button>
            </div>
        </div>
    `;
}

function toggleAppointmentsFilter() {
    activeAppointmentsOnlyCurrent = !activeAppointmentsOnlyCurrent;
    const btn = document.getElementById("btn-toggle-appointments-filter");
    if (btn) {
        btn.innerText = activeAppointmentsOnlyCurrent ? '🌐 Voir tous les dossiers' : '🎯 Ce dossier uniquement';
    }
    loadAppointmentsList();
}

function prefillAppointmentFormForDossier() {
    if (typeof currentDossierDetail === 'undefined' || !currentDossierDetail) return;
    const d = currentDossierDetail;
    const code = d.dossier_code || d.reference || '';
    const name = d.commercial_name || d.name || '';
    const isDM = (d.items || []).some(it => (it.item_code || '').startsWith('DM.')) || (d.procedure_type || '').includes('dm');

    const typeSelect = document.getElementById("app-type-select");
    if (typeSelect) typeSelect.value = "depot_initial";

    const notesInput = document.getElementById("app-notes");
    if (notesInput) {
        notesInput.value = `Dépôt initial ${isDM ? 'Homologation DM' : 'AMM CTD'} - ${code} (${name})`;
    }

    const slotInput = document.getElementById("app-slot-ref");
    if (slotInput) {
        if (d.deposit_receipt_ref) {
            slotInput.value = d.deposit_receipt_ref;
        } else if (!slotInput.value) {
            slotInput.placeholder = isDM ? "ex: RDVDM-2026-N°..." : "ex: RDVPP-2026-N°...";
        }
    }

    const schedInput = document.getElementById("app-scheduled-at");
    if (schedInput && !schedInput.value && d.target_deposit_date) {
        try {
            const dt = new Date(d.target_deposit_date);
            if (!isNaN(dt.getTime())) {
                dt.setHours(10, 0, 0, 0);
                const localISO = new Date(dt.getTime() - (dt.getTimezoneOffset() * 60000)).toISOString().slice(0, 16);
                schedInput.value = localISO;
            }
        } catch (_) {}
    }

    const form = document.getElementById("form-create-appointment");
    if (form) {
        form.scrollIntoView({ behavior: 'smooth', block: 'center' });
        if (slotInput) slotInput.focus();
    }
}

function jumpToAppointmentChecklistItem(itemCode) {
    closeModal('modal-appointments');
    if (typeof navigateToSmartPlanningItem === 'function') {
        navigateToSmartPlanningItem(itemCode, 'Récépissé / Convocation ANPP (J-60)', 'J_60_ETASJIL_APPOINTMENT');
    }
}

async function loadAppointmentsList() {
    const tbody = document.getElementById("appointments-tbody");
    if (!tbody) return;

    tbody.innerHTML = `<tr><td colspan="6" style="padding:20px; text-align:center; color:var(--text-dim);">Chargement des séances ANPP...</td></tr>`;

    try {
        const queryParams = new URLSearchParams();
        if (activeAppointmentsOnlyCurrent && activeAppointmentsDossierId) {
            queryParams.set("dossier_id", activeAppointmentsDossierId);
        }
        const url = "/api/v1/appointments" + (queryParams.toString() ? "?" + queryParams.toString() : "");
        const resp = await fetch(url, {
            headers: { 'Accept': 'application/json' }
        });
        if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
        const data = await resp.json();

        renderAppointmentsTable(data.appointments || []);
        
        // Mettre à jour le badge d'urgence si présent
        const urgentBadge = document.getElementById("badge-urgent-appointments");
        if (urgentBadge) {
            if (data.urgent_count > 0) {
                urgentBadge.innerText = `${data.urgent_count} RDV (J ≤ 7)`;
                urgentBadge.style.display = "inline-flex";
            } else {
                urgentBadge.style.display = "none";
            }
        }
    } catch (err) {
        tbody.innerHTML = `<tr><td colspan="6" style="padding:20px; text-align:center; color:#dc2626;">Erreur de chargement : ${escapeHTML(err.message)}</td></tr>`;
    }
}

function renderAppointmentsTable(appointments) {
    const tbody = document.getElementById("appointments-tbody");
    if (!tbody) return;

    if (appointments.length === 0) {
        tbody.innerHTML = `<tr><td colspan="6" style="padding:25px; text-align:center; color:var(--text-dim);">Aucune séance ANPP planifiée${activeAppointmentsOnlyCurrent ? ' pour ce dossier' : ''}</td></tr>`;
        return;
    }

    let html = '';
    const now = new Date();

    appointments.forEach(a => {
        const sched = new Date(a.scheduled_at);
        const diffMs = sched - now;
        const diffDays = Math.ceil(diffMs / (1000 * 60 * 60 * 24));
        const isUrgent = diffDays >= 0 && diffDays <= 7 && (a.status === 'planifie' || a.status === 'confirme_anpp');
        const isPast = diffMs < 0;
        const isCurrentDossier = activeAppointmentsDossierId && a.dossier_id === activeAppointmentsDossierId;

        let statusBadge = '';
        if (a.status === 'confirme_anpp') {
            statusBadge = `<span style="background:#dcfce7; color:#15803d; font-weight:700; padding:2px 8px; border-radius:10px; font-size:10px;">CONFIRMÉ ANPP</span>`;
        } else if (a.status === 'effectue') {
            statusBadge = `<span style="background:#e0e7ff; color:#3730a3; font-weight:700; padding:2px 8px; border-radius:10px; font-size:10px;">EFFECTUÉ</span>`;
        } else if (a.status === 'reporte') {
            statusBadge = `<span style="background:#fef3c7; color:#b45309; font-weight:700; padding:2px 8px; border-radius:10px; font-size:10px;">REPORTÉ</span>`;
        } else if (a.status === 'annule') {
            statusBadge = `<span style="background:#fee2e2; color:#991b1b; font-weight:700; padding:2px 8px; border-radius:10px; font-size:10px;">ANNULÉ</span>`;
        } else {
            statusBadge = `<span style="background:var(--bg-sub); color:var(--text-main); font-weight:700; padding:2px 8px; border-radius:10px; font-size:10px; border:1px solid var(--border-color);">PLANIFIÉ</span>`;
        }

        let typeLabel = a.appointment_type;
        if (a.appointment_type === 'seance_commission') typeLabel = '🏛️ Séance Commission';
        else if (a.appointment_type === 'depot_initial') typeLabel = '📥 Dépôt Initial';
        else if (a.appointment_type === 'expertise_lncm') typeLabel = '🔬 Expertise LNCM';
        else if (a.appointment_type === 'recours_audition') typeLabel = '⚖️ Audition / Recours';
        else if (a.appointment_type === 'depot_complements') typeLabel = '⏱️ Dépôt Compléments';

        let countdownBadge = '';
        if (isPast) {
            countdownBadge = `<span style="font-size:10px; color:var(--text-dim);">Échu</span>`;
        } else if (isUrgent) {
            countdownBadge = `<span style="background:#fee2e2; color:#dc2626; font-weight:800; padding:2px 6px; border-radius:8px; font-size:10px; animation: pulse 1.5s infinite;">J-${diffDays} (URGENT)</span>`;
        } else {
            countdownBadge = `<span style="background:#eff6ff; color:#2563eb; font-weight:600; padding:2px 6px; border-radius:8px; font-size:10px;">Dans ${diffDays}j</span>`;
        }

        const rowStyle = isCurrentDossier 
            ? 'background:rgba(37,99,235,0.06); border-left:3px solid var(--accent-blue);' 
            : (isUrgent ? 'background:rgba(239,68,68,0.03);' : '');

        html += `
            <tr style="border-bottom:1px solid var(--border-color); ${rowStyle}">
                <td style="padding:10px;">
                    <div style="display:flex; align-items:center; gap:6px;">
                        <span style="font-weight:700; font-size:12px; color:var(--accent-blue); font-family:monospace;">${escapeHTML(a.slot_reference || 'N/A')}</span>
                        ${isCurrentDossier ? '<span style="font-size:9.5px; background:var(--accent-blue); color:white; padding:1px 5px; border-radius:3px; font-weight:700;">🎯 Ce dossier</span>' : ''}
                    </div>
                    <div style="font-size:11px; color:var(--text-main); margin-top:2px;">${typeLabel}</div>
                    ${a.dossier_code ? `<div style="font-size:10px; color:var(--text-dim); margin-top:2px;">Dossier : <strong>${escapeHTML(a.dossier_code)}</strong></div>` : ''}
                </td>
                <td style="padding:10px;">
                    <div style="font-weight:700; font-size:11.5px; color:var(--text-main);">${sched.toLocaleDateString('fr-FR', { weekday:'short', day:'2-digit', month:'short', year:'numeric', hour:'2-digit', minute:'2-digit' })}</div>
                    <div style="margin-top:3px;">${countdownBadge}</div>
                </td>
                <td style="padding:10px; font-size:11px;">
                    <div style="color:var(--text-main); font-weight:600;">${escapeHTML(a.location)}</div>
                    ${a.contact_officer ? `<div style="font-size:10px; color:var(--text-dim); margin-top:2px;">Interlocuteur : ${escapeHTML(a.contact_officer)}</div>` : ''}
                </td>
                <td style="padding:10px; font-size:11px;">
                    <div style="color:var(--text-main);">${escapeHTML(a.assigned_delegate_name || 'Pharmacien DT')}</div>
                    ${a.notes ? `<div style="font-size:10px; color:var(--text-dim); margin-top:2px; font-style:italic;">"${escapeHTML(a.notes)}"</div>` : ''}
                </td>
                <td style="padding:10px; text-align:center;">
                    ${statusBadge}
                </td>
                <td style="padding:10px; text-align:right;">
                    <div style="display:flex; justify-content:flex-end; gap:4px;">
                        ${a.status !== 'confirme_anpp' && a.status !== 'effectue' ? `
                            <button class="btn-pill" style="font-size:10px; padding:2px 6px;" title="Confirmer réception convocation" onclick="updateAppointmentState('${a.id}', 'confirme_anpp')">✅ Confirmer</button>
                        ` : ''}
                        ${a.status !== 'effectue' ? `
                            <button class="btn-pill" style="font-size:10px; padding:2px 6px; background:var(--accent-blue); color:white; border:none;" title="Marquer comme réalisé" onclick="updateAppointmentState('${a.id}', 'effectue')">🏛️ Fait</button>
                        ` : ''}
                    </div>
                </td>
            </tr>
        `;
    });

    tbody.innerHTML = html;
}

// Planification d'une nouvelle séance
async function executeCreateAppointment(e) {
    if (e) e.preventDefault();

    const type = document.getElementById("app-type-select").value;
    const sched = document.getElementById("app-scheduled-at").value;
    const slot = document.getElementById("app-slot-ref").value.trim();
    const loc = document.getElementById("app-location").value.trim();
    const officer = document.getElementById("app-officer").value.trim();
    const notes = document.getElementById("app-notes").value.trim();

    if (!sched) {
        alert("Veuillez sélectionner la date et l'heure du créneau.");
        return;
    }

    try {
        const payload = {
            appointment_type: type,
            scheduled_at: sched,
            slot_reference: slot || undefined,
            location: loc || undefined,
            contact_officer: officer || undefined,
            notes: notes || undefined
        };

        if (typeof selectedDossierId !== 'undefined' && selectedDossierId) {
            payload.dossier_id = selectedDossierId;
        }

        const resp = await fetch("/api/v1/appointments", {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            },
            body: JSON.stringify(payload)
        });

        if (!resp.ok) {
            const err = await resp.json();
            throw new Error(err.error || `Erreur HTTP ${resp.status}`);
        }

        alert("Séance / Rendez-vous ANPP planifié avec succès.");
        document.getElementById("form-create-appointment").reset();
        await loadAppointmentsList();

        // Si rattaché au dossier actif, rafraîchir le dossier et le smart planning
        if (typeof selectDossier === 'function' && typeof selectedDossierId !== 'undefined' && selectedDossierId) {
            await selectDossier(selectedDossierId);
        }
    } catch (err) {
        alert("Erreur lors de la planification : " + err.message);
    }
}

// Mise à jour de l'état du créneau
async function updateAppointmentState(appId, newStatus) {
    try {
        const resp = await fetch(`/api/v1/appointments/${appId}/status`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            },
            body: JSON.stringify({ status: newStatus })
        });

        if (!resp.ok) {
            const err = await resp.json();
            throw new Error(err.error || `Erreur HTTP ${resp.status}`);
        }

        await loadAppointmentsList();

        if (typeof selectDossier === 'function' && typeof selectedDossierId !== 'undefined' && selectedDossierId) {
            await selectDossier(selectedDossierId);
        }
    } catch (err) {
        alert("Erreur mise à jour : " + err.message);
    }
}
