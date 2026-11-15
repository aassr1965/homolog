// ============================================================================
// GESTION DES RENDEZ-VOUS & SÉANCES DE DÉPÔT ANPP
// Planification des dépôts physiques, commissions et expertises LNCM
// ============================================================================

async function openAppointmentsModal() {
    const modal = document.getElementById("modal-appointments");
    if (!modal) return;

    openModal("modal-appointments");
    await loadAppointmentsList();
}

async function loadAppointmentsList() {
    const tbody = document.getElementById("appointments-tbody");
    if (!tbody) return;

    tbody.innerHTML = `<tr><td colspan="6" style="padding:20px; text-align:center; color:var(--text-dim);">Chargement des séances ANPP...</td></tr>`;

    try {
        const resp = await fetch("/api/v1/appointments", {
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
        tbody.innerHTML = `<tr><td colspan="6" style="padding:25px; text-align:center; color:var(--text-dim);">Aucune séance ANPP planifiée</td></tr>`;
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

        html += `
            <tr style="border-bottom:1px solid var(--border-color); ${isUrgent ? 'background:rgba(239,68,68,0.03);' : ''}">
                <td style="padding:10px;">
                    <div style="font-weight:700; font-size:12px; color:var(--accent-blue); font-family:monospace;">${escapeHTML(a.slot_reference)}</div>
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
    } catch (err) {
        alert("Erreur mise à jour : " + err.message);
    }
}
