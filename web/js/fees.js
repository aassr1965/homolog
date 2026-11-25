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
        showToast("Veuillez sélectionner un dossier réglementaire", "warning");
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

        showToast(`Bordereau ${slip.slip_type} généré : ${slip.slip_number} (${formatDZD(slip.amount_dzd)})`, "success");

        // Ouvrir automatiquement la vue imprimable officielle
        openPrintSlipWindow(slip.id);

        // Rafraîchir la vue du dossier
        if (typeof invalidateDossierCache === 'function') invalidateDossierCache(dossierId);
        if (typeof selectDossier === 'function') {
            selectDossier(dossierId, true);
        }
    } catch (err) {
        showToast("Erreur lors de la génération du bordereau: " + err.message, "error");
    }
}

// Ouverture de la version imprimable officielle dans un nouvel onglet
function openPrintSlipWindow(slipId) {
    if (!slipId) return;
    window.open(`/api/v1/fees/slips/${slipId}/print-html`, '_blank');
}

// Affiche/Masque les champs d'attestation de taux de change Banque d'Algérie
function toggleFXFields(show) {
    const box = document.getElementById("pay-slip-fx-fields");
    if (box) box.style.display = show ? "block" : "none";
}

// Ouverture de la modale de contrôle 1.A.4 & quittance bancaire
async function openRecordSlipPaymentModal(slipId, slipNumber, amountDZD, dossierId) {
    const modal = document.getElementById("modal-record-slip-payment");
    if (!modal) return;

    const dId = dossierId || (typeof selectedDossierId !== 'undefined' ? selectedDossierId : null);
    document.getElementById("pay-slip-id").value = slipId;
    document.getElementById("pay-slip-dossier-id").value = dId || "";
    document.getElementById("pay-slip-number-disp").innerText = slipNumber;
    document.getElementById("pay-slip-amount-disp").innerText = formatDZD(amountDZD);
    document.getElementById("pay-slip-ref").value = "";
    document.getElementById("pay-slip-date").value = new Date().toISOString().substring(0, 10);
    document.getElementById("pay-slip-notes").value = "";

    // Charger les documents GED du dossier pour les sélecteurs de fichiers
    loadGEDDocsForSelect("pay-slip-receipt-doc", dId);
    loadGEDDocsForSelect("pay-slip-fx-doc", dId);

    // Charger le résumé 1.A.4 et les détails du bordereau si dispo
    if (dId) {
        try {
            const summaryResp = await fetch(`/api/v1/dossiers/${dId}/fees/1a4-summary`);
            if (summaryResp.ok) {
                const summary = await summaryResp.json();
                const badge = document.getElementById("pay-slip-concordance-badge");
                const text = document.getElementById("pay-slip-concordance-text");
                const fxCheck = document.getElementById("pay-slip-requires-fx");
                
                if (summary.expected_fee_amount_dzd) {
                    if (amountDZD >= summary.expected_fee_amount_dzd) {
                        badge.style.background = "rgba(16,185,129,0.12)";
                        badge.style.borderColor = "rgba(16,185,129,0.3)";
                        badge.style.color = "#10b981";
                        text.innerHTML = `<strong>Concordance Conforme :</strong> Montant du bordereau (${formatDZD(amountDZD)}) &ge; Barème officiel (${formatDZD(summary.expected_fee_amount_dzd)} - ${escapeHTML(summary.fee_schedule_label)})`;
                    } else {
                        badge.style.background = "rgba(239,68,68,0.12)";
                        badge.style.borderColor = "rgba(239,68,68,0.3)";
                        badge.style.color = "#ef4444";
                        text.innerHTML = `<strong>Écart Financier Détecté :</strong> Montant du bordereau (${formatDZD(amountDZD)}) &lt; Barème officiel exigé (${formatDZD(summary.expected_fee_amount_dzd)}). Justification requise.`;
                    }
                }

                if (summary.is_imported) {
                    fxCheck.checked = true;
                    toggleFXFields(true);
                } else {
                    fxCheck.checked = false;
                    toggleFXFields(false);
                }
            }
        } catch (e) {
            console.warn("Impossible de charger la synthèse 1.A.4:", e);
        }
    }

    openModal("modal-record-slip-payment");
}

// Charge les documents de la GED pour peupler un <select>
async function loadGEDDocsForSelect(selectId, dossierId) {
    const sel = document.getElementById(selectId);
    if (!sel) return;
    sel.innerHTML = '<option value="">-- Aucun document GED sélectionné --</option>';
    if (!dossierId) return;

    try {
        const resp = await fetch(`/api/v1/dossiers/${dossierId}/ged`);
        if (!resp.ok) return;
        const data = await resp.json();
        const docs = data.documents || data || [];
        docs.forEach(d => {
            const opt = document.createElement("option");
            opt.value = d.id;
            opt.textContent = `[${d.document_code || 'GED'}] ${d.title_fr} (${d.file_name || d.current_version || 'v1.0'})`;
            sel.appendChild(opt);
        });
    } catch (e) {
        console.warn("Erreur chargement docs GED pour select:", e);
    }
}

// Exécution du contrôle réglementaire 1.A.4
async function executeVerifySlip1A4(e) {
    if (e) e.preventDefault();

    const slipId = document.getElementById("pay-slip-id").value;
    const dossierId = document.getElementById("pay-slip-dossier-id").value || (typeof selectedDossierId !== 'undefined' ? selectedDossierId : null);
    const mode = document.getElementById("pay-slip-mode").value;
    const bank = document.getElementById("pay-slip-bank").value;
    const ref = document.getElementById("pay-slip-ref").value.trim();
    const payDate = document.getElementById("pay-slip-date").value;
    const receiptDocId = document.getElementById("pay-slip-receipt-doc").value || null;
    const notes = document.getElementById("pay-slip-notes").value.trim();

    const requiresFX = document.getElementById("pay-slip-requires-fx").checked;
    const currency = document.getElementById("pay-slip-currency").value;
    const curAmtStr = document.getElementById("pay-slip-currency-amount").value.trim();
    const fxRateStr = document.getElementById("pay-slip-fx-rate").value.trim();
    const fxDate = document.getElementById("pay-slip-fx-date").value;
    const fxRef = document.getElementById("pay-slip-fx-ref").value.trim();
    const fxDocId = document.getElementById("pay-slip-fx-doc").value || null;

    if (!ref || !payDate) {
        showToast("La référence de quittance/avis bancaire et la date de paiement sont obligatoires.", "warning");
        return;
    }

    const payload = {
        bank_name: bank,
        payment_mode: mode,
        bank_receipt_ref: ref,
        bank_receipt_date: payDate,
        receipt_document_id: receiptDocId,
        currency: currency,
        requires_fx_attestation: requiresFX,
        verification_notes: notes
    };

    if (curAmtStr) {
        payload.currency_amount = parseFloat(curAmtStr);
    }
    if (requiresFX) {
        if (!fxRateStr || !fxRef || !fxDate) {
            showToast("Pour un dossier avec change en devises, le cours officiel BA, la référence de l'attestation et la date sont obligatoires.", "warning");
            return;
        }
        payload.fx_rate = parseFloat(fxRateStr);
        payload.fx_rate_date = fxDate;
        payload.fx_attestation_ref = fxRef;
        payload.fx_attestation_doc_id = fxDocId;
    }

    try {
        const resp = await fetch(`/api/v1/dossiers/${dossierId}/fees/slips/${slipId}/verify-1a4`, {
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

        const data = await resp.json();
        closeModal("modal-record-slip-payment");

        if (data.slip && data.slip.verification_status === 'conforme') {
            showToast("✅ Conformité 1.A.4 validée et scellée avec succès !", "success");
        } else {
            showToast("Attention : Enregistrement effectué mais des non-conformités subsistent (vérifiez l'attestation BA ou la concordance du montant).", "warning", 5500);
        }

        if (dossierId && typeof selectDossier === 'function') {
            if (typeof invalidateDossierCache === 'function') invalidateDossierCache(dossierId);
            selectDossier(dossierId, true);
        }
    } catch (err) {
        showToast("Erreur lors de la vérification 1.A.4: " + err.message, "error");
    }
}

// Fonction de rétrocompatibilité
function executeRecordSlipPayment(e) {
    return executeVerifySlip1A4(e);
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
        showToast("Veuillez renseigner tous les champs obligatoires (code, libellé, référence légale, montant valide)", "warning");
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

        showToast("Nouveau barème officiel enregistré et ancienne grille clôturée automatiquement", "success");

        loadFeeSchedules();
    } catch (err) {
        showToast("Erreur lors de la mise à jour du barème: " + err.message, "error");
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

// =========================================================================
// VOLET 1.A.5 : ACCORD DE PRÉ-SOUMISSION & REDEVANCE ANPP
// =========================================================================

async function openRecordPreSubmissionModal(dossierId) {
    const dId = dossierId || (typeof selectedDossierId !== 'undefined' ? selectedDossierId : null);
    if (!dId) {
        showToast("Veuillez d'abord sélectionner un dossier.", "warning");
        return;
    }

    const modal = document.getElementById("modal-record-presubmission");
    if (!modal) return;

    document.getElementById("presub-dossier-id").value = dId;

    // Réinitialiser les champs
    document.getElementById("presub-anpp-ref").value = "";
    document.getElementById("presub-notif-date").value = new Date().toISOString().split("T")[0];
    document.getElementById("presub-decision-type").value = "favorable";
    document.getElementById("presub-fee-amount").value = "50000.00";
    document.getElementById("presub-slip-ref").value = "";
    document.getElementById("presub-payment-date").value = new Date().toISOString().split("T")[0];
    document.getElementById("presub-verification-notes").value = "";

    const accordSelect = document.getElementById("presub-doc-accord");
    const paymentSelect = document.getElementById("presub-doc-payment");
    accordSelect.innerHTML = '<option value="">Sélectionner un document GED...</option>';
    paymentSelect.innerHTML = '<option value="">Sélectionner un document GED...</option>';

    // Charger les documents de la GED pour ce dossier
    try {
        const gedResp = await fetch(`/api/v1/ged/documents`);
        if (gedResp.ok) {
            const docs = await gedResp.json();
            (docs || []).forEach(doc => {
                const opt1 = document.createElement("option");
                opt1.value = doc.id;
                opt1.textContent = `${doc.document_code || 'GED'} - ${doc.title_fr}`;
                accordSelect.appendChild(opt1);

                const opt2 = document.createElement("option");
                opt2.value = doc.id;
                opt2.textContent = `${doc.document_code || 'GED'} - ${doc.title_fr}`;
                paymentSelect.appendChild(opt2);
            });
        }
    } catch (e) {
        console.warn("Échec chargement liste documents GED:", e);
    }

    // Charger la synthèse 1.A.5 actuelle
    try {
        const resp = await fetch(`/api/v1/dossiers/${dId}/fees/1a5-summary`);
        if (resp.ok) {
            const summary = await resp.json();
            const labelEl = document.getElementById("presub-dossier-label");
            const badgeEl = document.getElementById("presub-status-badge");
            const noticeEl = document.getElementById("presub-procedure-notice");
            const issuesEl = document.getElementById("presub-issues-list");

            labelEl.textContent = `Dossier : ${summary.dossier_code} (Procédure ${summary.procedure_code})`;

            if (!summary.is_pre_submission_required) {
                badgeEl.style.background = "rgba(59,130,246,0.15)";
                badgeEl.style.color = "#2563eb";
                badgeEl.textContent = "ℹ️ Non Applicable (RI/RF)";
                noticeEl.textContent = `Pour la procédure ${summary.procedure_code} (Renouvellement quinquennal), l'accord préalable de pré-soumission n'est pas exigé selon le Décret 20-325.`;
            } else if (summary.overall_compliant) {
                badgeEl.style.background = "rgba(16,185,129,0.15)";
                badgeEl.style.color = "#10b981";
                badgeEl.textContent = "✅ 1.A.5 Conforme & Scellé";
                noticeEl.textContent = "L'accord préalable de l'ANPP et la redevance ont été contrôlés et scellés sous visa du Pharmacien Directeur Technique.";
            } else {
                badgeEl.style.background = "rgba(234,179,8,0.15)";
                badgeEl.style.color = "#b45309";
                badgeEl.textContent = "⚠️ Action Requise";
                noticeEl.textContent = "L'accord formel préalable de pré-soumission et la quittance de redevance doivent être consignés et validés par le Pharmacien Directeur Technique.";
            }

            if (summary.blocking_issues && summary.blocking_issues.length > 0) {
                issuesEl.style.display = "block";
                issuesEl.innerHTML = "<strong>Points à régulariser :</strong><br>• " + summary.blocking_issues.join("<br>• ");
            } else {
                issuesEl.style.display = "none";
            }

            // Pré-remplir si déjà existant
            if (summary.pre_submission) {
                const ps = summary.pre_submission;
                document.getElementById("presub-anpp-ref").value = ps.anpp_reference || "";
                if (ps.anpp_notification_date) {
                    document.getElementById("presub-notif-date").value = ps.anpp_notification_date.split("T")[0];
                }
                document.getElementById("presub-decision-type").value = ps.decision_type || "favorable";
                document.getElementById("presub-fee-amount").value = ps.fee_amount_dzd || 50000.00;
                document.getElementById("presub-slip-ref").value = ps.slip_reference || "";
                if (ps.payment_date) {
                    document.getElementById("presub-payment-date").value = ps.payment_date.split("T")[0];
                }
                document.getElementById("presub-verification-notes").value = ps.verification_notes || "";
                if (ps.response_document_id) {
                    accordSelect.value = ps.response_document_id;
                }
                if (ps.payment_document_id) {
                    paymentSelect.value = ps.payment_document_id;
                }
            }
        }
    } catch (e) {
        console.warn("Échec chargement synthèse 1.A.5:", e);
    }

    if (typeof openModal === 'function') {
        openModal("modal-record-presubmission");
    } else {
        const el = document.getElementById("modal-record-presubmission");
        if (el) { el.classList.add("open"); el.style.display = "flex"; }
    }
}

async function executeVerifyPreSubmission1A5(e) {
    if (e) e.preventDefault();

    const dossierId = document.getElementById("presub-dossier-id").value;
    const anppRef = document.getElementById("presub-anpp-ref").value.trim();
    const notifDate = document.getElementById("presub-notif-date").value;
    const decType = document.getElementById("presub-decision-type").value;
    const feeAmt = parseFloat(document.getElementById("presub-fee-amount").value) || 0;
    const slipRef = document.getElementById("presub-slip-ref").value.trim();
    const payDate = document.getElementById("presub-payment-date").value;
    const accordDocId = document.getElementById("presub-doc-accord").value || null;
    const paymentDocId = document.getElementById("presub-doc-payment").value || null;
    const notes = document.getElementById("presub-verification-notes").value.trim();

    if (!anppRef || !notifDate || !slipRef || !payDate || !accordDocId || !paymentDocId) {
        showToast("Veuillez remplir l'ensemble des champs obligatoires et sélectionner les deux justificatifs GED (Accord préalable & Quittance).", "warning");
        return;
    }

    try {
        const resp = await fetch(`/api/v1/dossiers/${dossierId}/fees/verify-1a5`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            },
            body: JSON.stringify({
                anpp_reference: anppRef,
                anpp_notification_date: notifDate,
                decision_type: decType,
                fee_amount_dzd: feeAmt,
                slip_reference: slipRef,
                payment_date: payDate,
                response_document_id: accordDocId,
                payment_document_id: paymentDocId,
                verification_notes: notes
            })
        });

        if (!resp.ok) {
            const err = await resp.json();
            throw new Error(err.error || `Erreur HTTP ${resp.status}`);
        }

        const data = await resp.json();
        closeModal("modal-record-presubmission");

        showToast("✅ Accord et redevance de pré-soumission 1.A.5 validés et scellés avec succès !", "success");

        if (dossierId && typeof selectDossier === 'function') {
            if (typeof invalidateDossierCache === 'function') invalidateDossierCache(dossierId);
            selectDossier(dossierId, true);
        }
    } catch (err) {
        showToast("Erreur lors de la validation 1.A.5: " + err.message, "error");
    }
}

// =========================================================================
// TABLEAU DE BORD CONSOLIDÉ DES COÛTS D'ENREGISTREMENT (ANPP & ALCOA+)
// =========================================================================

function openCostAnalyticsDashboardModal() {
    if (typeof openModal === 'function') {
        openModal("modal-cost-analytics-dashboard");
    } else {
        const el = document.getElementById("modal-cost-analytics-dashboard");
        if (el) { el.classList.add("open"); el.style.display = "flex"; }
    }
    loadCostAnalyticsDashboard();
}

let costDashboardData = null;

async function loadCostAnalyticsDashboard() {
    try {
        const resp = await fetch('/api/v1/fees/analytics/dashboard');
        if (!resp.ok) {
            const err = await resp.json();
            throw new Error(err.error || `Erreur HTTP ${resp.status}`);
        }

        costDashboardData = await resp.json();
        renderCostDashboard(costDashboardData);
    } catch (err) {
        showToast("Erreur lors du chargement du tableau de bord des coûts: " + err.message, "error");
    }
}

function renderCostDashboard(data) {
    if (!data) return;

    // 1. KPI Cards
    const totalEmittedEl = document.getElementById("kpi-cost-total-emitted");
    const totalPaidEl = document.getElementById("kpi-cost-total-paid");
    const totalPendingEl = document.getElementById("kpi-cost-total-pending");
    const paidRatioEl = document.getElementById("kpi-cost-paid-ratio");
    const settledCountEl = document.getElementById("kpi-cost-settled-count");

    const emitted = data.total_fees_emitted_dzd || 0;
    const paid = data.total_fees_paid_dzd || 0;
    const pending = data.total_fees_pending_dzd || 0;

    totalEmittedEl.textContent = formatDZD(emitted);
    totalPaidEl.textContent = formatDZD(paid);
    totalPendingEl.textContent = formatDZD(pending);

    const ratio = emitted > 0 ? Math.round((paid / emitted) * 100) : 0;
    paidRatioEl.textContent = `${ratio}% du total engagé`;

    const totalDossiers = data.total_dossiers || 0;
    const fullySettled = data.fully_settled_count || 0;
    settledCountEl.textContent = `${fullySettled} / ${totalDossiers}`;

    // 2. Table Produits & Dossiers
    const tbodyProducts = document.getElementById("table-costs-products-body");
    if (data.product_breakdown && data.product_breakdown.length > 0) {
        tbodyProducts.innerHTML = data.product_breakdown.map(p => {
            const statusBadge = p.financial_status === 'acquitte'
                ? '<span class="badge" style="background:rgba(16,185,129,0.15); color:#10b981; font-weight:700;">✅ Acquitté</span>'
                : (p.financial_status === 'partiel'
                    ? '<span class="badge" style="background:rgba(234,179,8,0.15); color:#b45309; font-weight:700;">🟡 Partiel</span>'
                    : '<span class="badge" style="background:rgba(239,68,68,0.15); color:#ef4444; font-weight:700;">⏳ En attente</span>');

            return `
                <tr style="border-bottom:1px solid var(--border-color);">
                    <td style="padding:8px 10px; font-weight:700; color:var(--text-main);">${escapeHTML(p.product_name)}</td>
                    <td style="padding:8px 10px; font-family:monospace; color:var(--accent-blue);">${escapeHTML(p.dossier_code)}</td>
                    <td style="padding:8px 10px;"><span class="badge" style="font-size:10px;">${escapeHTML(p.procedure_code)}</span></td>
                    <td style="padding:8px 10px; text-align:right;">${formatDZD(p.pre_submission_dzd)}</td>
                    <td style="padding:8px 10px; text-align:right;">${formatDZD(p.initial_deposit_dzd)}</td>
                    <td style="padding:8px 10px; text-align:right;">${formatDZD(p.complementary_dzd)}</td>
                    <td style="padding:8px 10px; text-align:right; font-weight:700;">${formatDZD(p.total_committed_dzd)}</td>
                    <td style="padding:8px 10px; text-align:right; font-weight:700; color:#10b981;">${formatDZD(p.total_paid_dzd)}</td>
                    <td style="padding:8px 10px; text-align:center;">${statusBadge}</td>
                </tr>
            `;
        }).join('');
    } else {
        tbodyProducts.innerHTML = '<tr><td colspan="9" style="padding:16px; text-align:center; color:var(--text-dim);">Aucun dossier enregistré.</td></tr>';
    }

    // 3. Table Exercice Annuel
    const tbodyYears = document.getElementById("table-costs-years-body");
    if (data.yearly_breakdown && data.yearly_breakdown.length > 0) {
        tbodyYears.innerHTML = data.yearly_breakdown.map(y => `
            <tr style="border-bottom:1px solid var(--border-color);">
                <td style="padding:8px 10px; font-weight:800; color:var(--text-main);">${y.year}</td>
                <td style="padding:8px 10px; text-align:center; font-weight:700;">${y.dossier_count}</td>
                <td style="padding:8px 10px; text-align:right; font-weight:700;">${formatDZD(y.total_emitted_dzd)}</td>
                <td style="padding:8px 10px; text-align:right; font-weight:700; color:#10b981;">${formatDZD(y.total_paid_dzd)}</td>
                <td style="padding:8px 10px; text-align:right; font-weight:700; color:#b45309;">${formatDZD(y.total_pending_dzd)}</td>
                <td style="padding:8px 10px; text-align:right;">${formatDZD(y.pre_submission_dzd)}</td>
                <td style="padding:8px 10px; text-align:right;">${formatDZD(y.initial_deposit_dzd)}</td>
                <td style="padding:8px 10px; text-align:right;">${formatDZD(y.complementary_dzd)}</td>
            </tr>
        `).join('');
    } else {
        tbodyYears.innerHTML = '<tr><td colspan="8" style="padding:16px; text-align:center; color:var(--text-dim);">Aucun exercice disponible.</td></tr>';
    }

    // 4. Catégories
    const catContainer = document.getElementById("cost-categories-container");
    if (data.category_breakdown && data.category_breakdown.length > 0) {
        catContainer.innerHTML = data.category_breakdown.map(c => `
            <div style="background:var(--bg-sub); border:1px solid var(--border-color); border-radius:6px; padding:10px 14px;">
                <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:6px;">
                    <div>
                        <strong style="font-size:12px; color:var(--text-main);">${escapeHTML(c.label_fr)}</strong>
                        <span style="font-size:10.5px; color:var(--text-dim); margin-left:8px;">(${c.percentage.toFixed(1)}%)</span>
                    </div>
                    <strong style="font-size:13px; color:var(--accent-blue);">${formatDZD(c.amount_dzd)}</strong>
                </div>
                <div style="width:100%; height:6px; background:var(--border-color); border-radius:3px; overflow:hidden;">
                    <div style="width:${Math.min(100, Math.max(0, c.percentage))}%; height:100%; background:var(--accent-blue); border-radius:3px;"></div>
                </div>
            </div>
        `).join('');
    } else {
        catContainer.innerHTML = '<div style="padding:16px; text-align:center; color:var(--text-dim);">Aucune dépense catégorisée.</div>';
    }
}

function switchCostTab(tabName) {
    const tabs = ['products', 'years', 'categories'];
    tabs.forEach(t => {
        const btn = document.getElementById(`tab-btn-costs-${t}`);
        const view = document.getElementById(`cost-view-${t}`);
        if (t === tabName) {
            btn.classList.add("active");
            btn.style.color = "var(--accent-blue)";
            view.style.display = "block";
        } else {
            btn.classList.remove("active");
            btn.style.color = "var(--text-dim)";
            view.style.display = "none";
        }
    });
}

function exportCostConsolidationCSV() {
    window.open('/api/v1/fees/analytics/export', '_blank');
}

