/**
 * Calculateur Officiel des Redevances ANPP & Générateur de Bordereaux BV1/BV2
 * Conforme aux décrets d'homologation et arrêtés interministériels des redevances
 */

let currentDossierPaymentSlips = [];
let allFeeSchedules = [];

// Chargement des bordereaux d'un dossier
async function loadDossierPaymentSlips(dossierId) {
    if (!dossierId) return [];
    try {
        const resp = await fetch(`/api/v1/dossiers/${dossierId}/fees`, {
            headers: { 'Accept': 'application/json' }
        });
        if (!resp.ok) throw new Error(`Erreur HTTP ${resp.status}`);
        const data = await resp.json();
        currentDossierPaymentSlips = data.slips || [];
        return currentDossierPaymentSlips;
    } catch (e) {
        console.error("Erreur chargement bordereaux:", e);
        return [];
    }
}

// Chargement des barèmes officiels (pour affichage et mise à jour)
async function loadFeeSchedules() {
    try {
        const resp = await fetch('/api/v1/fees/schedules', {
            headers: { 'Accept': 'application/json' }
        });
        if (!resp.ok) throw new Error(`Erreur HTTP ${resp.status}`);
        const data = await resp.json();
        allFeeSchedules = data.schedules || [];
        renderFeeSchedulesTable(allFeeSchedules);
    } catch (e) {
        console.error("Erreur chargement barèmes:", e);
    }
}

// Ouverture de la modale de génération de bordereau BV1/BV2
async function openGenerateFeeSlipModal(dossierId) {
    const modal = document.getElementById("modal-generate-slip");
    if (!modal) return;

    const id = dossierId || (typeof selectedDossierId !== 'undefined' ? selectedDossierId : null);
    if (!id) {
        alert("Veuillez sélectionner un dossier réglementaire");
        return;
    }

    document.getElementById("slip-dossier-id").value = id;
    document.getElementById("slip-type-select").value = "BV1";
    document.getElementById("slip-purpose-input").value = "";
    document.getElementById("slip-custom-amount").value = "";

    // Affichage des informations du dossier
    if (typeof currentDossierDetail !== 'undefined' && currentDossierDetail) {
        document.getElementById("slip-target-dossier-info").innerText = `${currentDossierDetail.dossier_code} — ${currentDossierDetail.medicine ? currentDossierDetail.medicine.brand_name_fr : (currentDossierDetail.medical_device ? currentDossierDetail.medical_device.commercial_name_fr : '')}`;
    }

    openModal("modal-generate-slip");
}

// Exécution de la génération du bordereau
async function executeGenerateFeeSlip(e) {
    if (e) e.preventDefault();

    const dossierId = document.getElementById("slip-dossier-id").value;
    const slipType = document.getElementById("slip-type-select").value;
    const purpose = document.getElementById("slip-purpose-input").value.trim();
    const customAmtStr = document.getElementById("slip-custom-amount").value.trim();

    const payload = {
        slip_type: slipType,
        payment_purpose: purpose
    };
    if (customAmtStr) {
        const amt = parseFloat(customAmtStr);
        if (!isNaN(amt) && amt > 0) {
            payload.custom_amount = amt;
        }
    }

    try {
        const resp = await fetch(`/api/v1/dossiers/${dossierId}/fees/generate-slip`, {
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

        const slip = await resp.json();
        closeModal("modal-generate-slip");

        if (typeof showToast === 'function') {
            showToast(`Bordereau ${slip.slip_type} généré : ${slip.slip_number} (${formatDZD(slip.amount_dzd)})`);
        } else {
            alert(`Bordereau ${slip.slip_type} généré : ${slip.slip_number} (${formatDZD(slip.amount_dzd)})`);
        }

        // Ouvrir automatiquement la vue imprimable officielle
        openPrintSlipWindow(slip.id);

        // Rafraîchir la vue du dossier
        if (typeof selectDossier === 'function') {
            selectDossier(dossierId);
        }
    } catch (err) {
        alert("Erreur lors de la génération du bordereau: " + err.message);
    }
}

// Ouverture de la version imprimable officielle dans un nouvel onglet
function openPrintSlipWindow(slipId) {
    if (!slipId) return;
    window.open(`/api/v1/fees/slips/${slipId}/print-html`, '_blank');
}

// Ouverture de la modale de saisie de quittance Trésor
function openRecordSlipPaymentModal(slipId, slipNumber, amountDZD) {
    const modal = document.getElementById("modal-record-slip-payment");
    if (!modal) return;

    document.getElementById("pay-slip-id").value = slipId;
    document.getElementById("pay-slip-number-disp").innerText = slipNumber;
    document.getElementById("pay-slip-amount-disp").innerText = formatDZD(amountDZD);
    document.getElementById("pay-slip-ref").value = "";
    document.getElementById("pay-slip-date").value = new Date().toISOString().substring(0, 10);

    openModal("modal-record-slip-payment");
}

async function executeRecordSlipPayment(e) {
    if (e) e.preventDefault();

    const slipId = document.getElementById("pay-slip-id").value;
    const ref = document.getElementById("pay-slip-ref").value.trim();
    const payDate = document.getElementById("pay-slip-date").value;

    if (!ref) {
        alert("Le numéro de quittance ou référence de paiement est obligatoire");
        return;
    }

    try {
        const resp = await fetch(`/api/v1/fees/slips/${slipId}/status`, {
            method: 'PUT',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            },
            body: JSON.stringify({
                status: 'paye',
                payment_reference: ref,
                paid_at: payDate
            })
        });

        if (!resp.ok) {
            const err = await resp.json();
            throw new Error(err.error || `Erreur HTTP ${resp.status}`);
        }

        closeModal("modal-record-slip-payment");

        if (typeof showToast === 'function') {
            showToast("Quittance de versement Trésor enregistrée avec succès");
        } else {
            alert("Quittance de versement Trésor enregistrée avec succès");
        }

        if (typeof selectedDossierId !== 'undefined' && selectedDossierId && typeof selectDossier === 'function') {
            selectDossier(selectedDossierId);
        }
    } catch (err) {
        alert("Erreur lors de l'enregistrement de la quittance: " + err.message);
    }
}

// Ouverture de la modale d'administration des barèmes officiels
function openFeeSchedulesModal() {
    const modal = document.getElementById("modal-fee-schedules");
    if (!modal) return;
    loadFeeSchedules();
    openModal("modal-fee-schedules");
}

// Rendu du tableau des barèmes
function renderFeeSchedulesTable(schedules) {
    const tbody = document.getElementById("fee-schedules-tbody");
    if (!tbody) return;

    if (!schedules || schedules.length === 0) {
        tbody.innerHTML = `<tr><td colspan="6" style="padding:20px; text-align:center; color:var(--text-muted);">Aucun barème configuré</td></tr>`;
        return;
    }

    let html = '';
    schedules.forEach(s => {
        const isActive = s.is_active;
        const validFromStr = (s.valid_from || '').substring(0, 10);
        const validToStr = s.valid_to ? s.valid_to.substring(0, 10) : 'En vigueur';

        html += `
            <tr style="border-bottom:1px solid var(--border-color); ${!isActive ? 'opacity:0.6; background:var(--bg-sub);' : ''}">
                <td style="padding:8px 10px;">
                    <div style="font-weight:700; color:var(--text-main); font-size:11.5px;">${escapeHTML(s.label_fr)}</div>
                    <div style="font-size:10px; color:var(--text-dim); font-family:monospace;">${escapeHTML(s.procedure_code)}</div>
                </td>
                <td style="padding:8px 10px; font-size:11px;">
                    ${escapeHTML(s.legal_reference)}
                </td>
                <td style="padding:8px 10px; text-align:right; font-weight:800; color:var(--accent-blue); font-size:12px;">
                    ${formatDZD(s.amount_dzd)}
                </td>
                <td style="padding:8px 10px; font-size:10.5px; white-space:nowrap;">
                    ${validFromStr} ➔ ${validToStr}
                </td>
                <td style="padding:8px 10px; text-align:center;">
                    <span class="badge" style="background:${isActive ? 'rgba(16,185,129,0.15)' : 'rgba(100,116,139,0.15)'}; color:${isActive ? '#10b981' : '#64748b'};">
                        ${isActive ? 'Actif' : 'Clôturé'}
                    </span>
                </td>
                <td style="padding:8px 10px; text-align:right;">
                    <button class="btn-pill" style="font-size:10.5px; padding:2px 8px;" onclick="prefillReviseSchedule('${s.procedure_code}', '${s.category}', '${escapeHTML(s.label_fr)}', ${s.amount_dzd})">✏️ Réviser</button>
                </td>
            </tr>
        `;
    });

    tbody.innerHTML = html;
}

// Pré-remplir le formulaire de révision d'un barème
function prefillReviseSchedule(procCode, category, labelFR, currentAmt) {
    document.getElementById("sched-proc-code").value = procCode;
    document.getElementById("sched-category").value = category;
    document.getElementById("sched-label").value = labelFR;
    document.getElementById("sched-amount").value = currentAmt;
    document.getElementById("sched-legal-ref").value = "Arrêté Interministériel du " + new Date().toLocaleDateString('fr-FR');
    document.getElementById("sched-valid-from").value = new Date().toISOString().substring(0, 10);
    document.getElementById("form-revise-schedule-container").scrollIntoView({ behavior: 'smooth' });
}

// Enregistrement d'un nouveau tarif / révision de barème
async function executeSaveFeeSchedule(e) {
    if (e) e.preventDefault();

    const procCode = document.getElementById("sched-proc-code").value.trim();
    const category = document.getElementById("sched-category").value;
    const label = document.getElementById("sched-label").value.trim();
    const legalRef = document.getElementById("sched-legal-ref").value.trim();
    const amt = parseFloat(document.getElementById("sched-amount").value);
    const validFrom = document.getElementById("sched-valid-from").value;
    const notes = document.getElementById("sched-notes").value.trim();

    if (!procCode || !label || !legalRef || isNaN(amt) || amt <= 0) {
        alert("Veuillez renseigner tous les champs obligatoires (code, libellé, référence légale, montant valide)");
        return;
    }

    try {
        const resp = await fetch('/api/v1/fees/schedules', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            },
            body: JSON.stringify({
                category: category,
                procedure_code: procCode,
                label_fr: label,
                legal_reference: legalRef,
                amount_dzd: amt,
                valid_from: validFrom,
                notes: notes
            })
        });

        if (!resp.ok) {
            const err = await resp.json();
            throw new Error(err.error || `Erreur HTTP ${resp.status}`);
        }

        if (typeof showToast === 'function') {
            showToast("Nouveau barème officiel enregistré et ancienne grille clôturée automatiquement");
        } else {
            alert("Nouveau barème officiel enregistré et ancienne grille clôturée automatiquement");
        }

        loadFeeSchedules();
    } catch (err) {
        alert("Erreur lors de la mise à jour du barème: " + err.message);
    }
}

// Formatage monétaire DZD
function formatDZD(amount) {
    if (typeof amount !== 'number') amount = parseFloat(amount) || 0;
    return new Intl.NumberFormat('fr-DZ', {
        minimumFractionDigits: 2,
        maximumFractionDigits: 2
    }).format(amount) + " DZD";
}
