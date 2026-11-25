// ============================================================================
// ASSISTANT DE SOUMISSION E-SERVICES ANPP & VALIDATION PRÉ-DÉPÔT
// Contrôle de complétude réglementaire, générateur de manifeste & verrouillage
// ============================================================================

let currentAuditDossierId = null;

// Ouverture de la modale d'audit pré-dépôt et validation E-Services ANPP
async function openPreSubmissionAuditModal(dossierId) {
    if (!dossierId) {
        if (typeof selectedDossierId !== 'undefined' && selectedDossierId) {
            dossierId = selectedDossierId;
        } else {
            showToast("Veuillez sélectionner un dossier d'enregistrement", "warning");
            return;
        }
    }

    currentAuditDossierId = dossierId;
    const modal = document.getElementById("modal-pre-submission-audit");
    if (!modal) return;

    // Reset UI
    document.getElementById("audit-summary-banner").innerHTML = '<div style="padding:15px; text-align:center; color:var(--text-dim);">Analyse de complétude réglementaire en cours...</div>';
    document.getElementById("audit-checks-tbody").innerHTML = '';
    document.getElementById("btn-lock-submission").disabled = true;

    openModal("modal-pre-submission-audit");
    await loadPreSubmissionAudit(dossierId);
}

// Chargement des données d'audit pré-dépôt depuis l'API
async function loadPreSubmissionAudit(dossierId) {
    try {
        const resp = await fetch(`/api/v1/dossiers/${dossierId}/pre-submission-audit`, {
            headers: { 'Accept': 'application/json' }
        });

        if (!resp.ok) {
            const err = await resp.json();
            throw new Error(err.error || `Erreur HTTP ${resp.status}`);
        }

        const report = await resp.json();
        renderPreSubmissionReport(report);
    } catch (err) {
        document.getElementById("audit-summary-banner").innerHTML = `
            <div style="background:#fee2e2; border:1px solid #f87171; color:#991b1b; padding:12px; border-radius:6px; font-size:12px;">
                ⚠️ Erreur lors de l'audit pré-dépôt : ${escapeHTML(err.message)}
            </div>
        `;
    }
}

// Rendu du rapport d'audit et de conformité
function renderPreSubmissionReport(report) {
    const banner = document.getElementById("audit-summary-banner");
    const tbody = document.getElementById("audit-checks-tbody");
    const lockBtn = document.getElementById("btn-lock-submission");

    document.getElementById("audit-dossier-code").innerText = report.dossier_code;
    document.getElementById("audit-product-name").innerText = report.product_name;

    // 1. Bannière d'éligibilité Gatekeeper BPF / ANPP
    if (report.is_eligible) {
        banner.innerHTML = `
            <div style="background:#f0fdf4; border:1px solid #86efac; border-radius:8px; padding:14px; display:flex; align-items:center; gap:14px;">
                <div style="font-size:32px;">✅</div>
                <div style="flex:1;">
                    <div style="font-weight:800; font-size:14px; color:#166534;">DOSSIER 100% CONFORME — ÉLIGIBLE AU TÉLÉ-DÉPÔT E-SERVICES ANPP</div>
                    <div style="font-size:11.5px; color:#15803d; margin-top:3px;">
                        Toutes les pièces obligatoires [E] sont validées, le bordereau BV1 est acquitté par le Trésor public, et le visa du Pharmacien Directeur Technique est formellement scellé.
                    </div>
                </div>
                <div style="text-align:right;">
                    <div style="font-size:10px; color:#15803d; font-weight:700; text-transform:uppercase;">Score de Complétude :</div>
                    <div style="font-size:24px; font-weight:900; color:#16a34a;">${report.completeness_score}%</div>
                </div>
            </div>
        `;
        if (lockBtn) lockBtn.disabled = false;
    } else {
        banner.innerHTML = `
            <div style="background:#fef2f2; border:1px solid #fca5a5; border-radius:8px; padding:14px; display:flex; align-items:center; gap:14px;">
                <div style="font-size:32px;">🚫</div>
                <div style="flex:1;">
                    <div style="font-weight:800; font-size:14px; color:#991b1b;">TÉLÉ-DÉPÔT BLOQUÉ — ${report.critical_blockers_count} EXIGENCE(S) CRITIQUE(S) NON CONFORME(S)</div>
                    <div style="font-size:11.5px; color:#b91c1c; margin-top:3px;">
                        Le dossier ne peut pas être soumis sur le portail E-Services de l'ANPP tant que les points bloquants ci-dessous ne sont pas résolus (Article 12 du Décret 20-325).
                    </div>
                </div>
                <div style="text-align:right;">
                    <div style="font-size:10px; color:#991b1b; font-weight:700; text-transform:uppercase;">Score Actuel :</div>
                    <div style="font-size:24px; font-weight:900; color:#dc2626;">${report.completeness_score}%</div>
                </div>
            </div>
        `;
        if (lockBtn) lockBtn.disabled = true;
    }

    // Bloc récapitulatif Smart Planning & Chemin Critique Inversé
    if (report.smart_planning) {
        const sp = report.smart_planning;
        const targetFormatted = sp.target_deposit_date ? new Date(sp.target_deposit_date).toLocaleDateString('fr-FR') : 'Non fixée';
        const recFormatted = sp.recommended_deposit_date ? new Date(sp.recommended_deposit_date).toLocaleDateString('fr-FR') : 'Non fixée';
        const isSufficient = sp.is_target_date_sufficient;

        const spCard = `
            <div style="margin-top:10px; background:${isSufficient ? 'rgba(16,185,129,0.08)' : 'rgba(245,158,11,0.08)'}; border:1px solid ${isSufficient ? '#86efac' : '#fcd34d'}; border-radius:6px; padding:10px 12px; display:flex; justify-content:space-between; align-items:center; flex-wrap:wrap; gap:8px;">
                <div>
                    <div style="font-size:12px; font-weight:800; color:var(--text-main); display:flex; align-items:center; gap:6px;">
                        <span>⏱️</span>
                        <span>Smart Planning ANPP :</span>
                        <span style="color:${isSufficient ? '#15803d' : '#b45309'};">${escapeHTML(sp.overall_health_label)}</span>
                    </div>
                    <div style="font-size:11px; color:var(--text-dim); margin-top:3px;">
                        Date Cible : <strong>${targetFormatted}</strong> • Date Recommandée : <strong>${recFormatted}</strong> (Base : ${sp.working_days_per_week}j ouvrables/semaine, ${sp.holidays_excluded_count || 0} jours fériés déduits)
                    </div>
                    ${sp.recommended_action_fr ? `<div style="font-size:10.5px; color:var(--text-main); margin-top:2px;">💡 ${escapeHTML(sp.recommended_action_fr)}</div>` : ''}
                </div>
            </div>
        `;
        banner.innerHTML += spCard;
    }

    // 2. Rendu de la table détaillée des contrôles
    let html = '';
    report.checks.forEach(c => {
        let statusBadge = '';
        let rowBg = '';

        if (c.status === 'CONFORME') {
            statusBadge = `<span style="background:#dcfce7; color:#15803d; font-weight:700; padding:2px 8px; border-radius:10px; font-size:10px;">CONFORME</span>`;
        } else if (c.status === 'BLOQUANT') {
            statusBadge = `<span style="background:#fee2e2; color:#991b1b; font-weight:800; padding:2px 8px; border-radius:10px; font-size:10px;">BLOQUANT</span>`;
            rowBg = 'background:rgba(239,68,68,0.04);';
        } else {
            statusBadge = `<span style="background:#fef3c7; color:#b45309; font-weight:700; padding:2px 8px; border-radius:10px; font-size:10px;">AVERTISSEMENT</span>`;
        }

        let catIcon = '📋';
        if (c.category === 'ADMINISTRATIF') catIcon = '🏛️';
        else if (c.category === 'FINANCIER_BV1') catIcon = '💰';
        else if (c.category === 'TECHNIQUE_CTD') catIcon = '📚';
        else if (c.category === 'VISA_PDT') catIcon = '✍️';
        else if (c.category === 'CHEMIN_CRITIQUE') catIcon = '⏱️';

        html += `
            <tr style="border-bottom:1px solid var(--border-color); ${rowBg}">
                <td style="padding:10px; font-size:11px;">
                    <span style="font-size:12px; margin-right:4px;">${catIcon}</span>
                    <strong>${escapeHTML(c.category)}</strong>
                </td>
                <td style="padding:10px;">
                    <div style="font-weight:700; font-size:12px; color:var(--text-main);">${escapeHTML(c.label)}</div>
                    <div style="font-size:11px; color:var(--text-dim); margin-top:2px;">${escapeHTML(c.details)}</div>
                    ${c.remediation ? `<div style="font-size:10.5px; color:#dc2626; margin-top:4px; font-weight:600;">➔ Action requise : ${escapeHTML(c.remediation)}</div>` : ''}
                </td>
                <td style="padding:10px; text-align:center;">
                    ${statusBadge}
                </td>
            </tr>
        `;
    });

    tbody.innerHTML = html;
}

// Téléchargement du manifeste JSON E-Services
function downloadEservicesManifest(format) {
    if (!currentAuditDossierId) return;
    const url = `/api/v1/dossiers/${currentAuditDossierId}/eservices-manifest?format=${format}&download=true`;
    window.open(url, '_blank');
}

// Verrouillage et certification officielle pour soumission E-Services ANPP
async function executeLockPreSubmission() {
    if (!currentAuditDossierId) return;

    const notes = await showPromptModal(
        "Certification Télé-Dépôt E-Services ANPP",
        "Notes ou observations pour la certification de télé-dépôt (optionnel) :",
        "Vérifié conforme pour télé-transmission officielle session ANPP 2026",
        { icon: "🔏", confirmText: "Certifier & Sceller" }
    );
    if (notes === null) return; // Annulé

    try {
        const resp = await fetch(`/api/v1/dossiers/${currentAuditDossierId}/lock-pre-submission`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Accept': 'application/json'
            },
            body: JSON.stringify({ submission_notes: notes })
        });

        if (!resp.ok) {
            const err = await resp.json();
            throw new Error(err.error || `Erreur HTTP ${resp.status}`);
        }

        const data = await resp.json();
        showNotifyModal(
            "Télé-Dépôt Certifié & Scellé",
            `<div>${escapeHTML(data.message || '')}</div><div style="margin-top:8px;"><strong>Empreinte d'intégrité scellée (SHA-256) :</strong><br><code style="font-size:10.5px; word-break:break-all;">${escapeHTML(data.manifest_digest || '')}</code></div>`,
            "✅",
            "Fermer"
        );

        closeModal("modal-pre-submission-audit");

        // Rafraîchir les dossiers
        if (typeof loadDossiers === 'function') {
            loadDossiers();
        }
        if (typeof selectDossier === 'function') {
            selectDossier(currentAuditDossierId);
        }
    } catch (err) {
        showToast("Erreur lors de la certification : " + err.message, "error");
    }
}
