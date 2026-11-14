// ============================================================================
// MODULE VEILLE RÉGLEMENTAIRE & CORPUS JURIDIQUE ALGÉRIEN (JORADP / ANPP)
// Jalon 1.7 - Phase 4 : Suivi des textes légaux, décrets et conformité BPF
// ============================================================================

async function openRegulatoryWatchModal() {
    const modal = document.getElementById("modal-regulatory-watch");
    if (!modal) return;

    openModal("modal-regulatory-watch");
    await loadRegulatoryWatchList();
}

async function loadRegulatoryWatchList() {
    const tbody = document.getElementById("reg-watch-tbody");
    if (!tbody) return;

    const cat = document.getElementById("reg-filter-cat") ? document.getElementById("reg-filter-cat").value : "";
    const impact = document.getElementById("reg-filter-impact") ? document.getElementById("reg-filter-impact").value : "";

    tbody.innerHTML = `<tr><td colspan="6" style="padding:20px; text-align:center; color:var(--text-dim);">Chargement du corpus juridique...</td></tr>`;

    try {
        let url = "/api/v1/regulatory-watch?";
        if (cat) url += `category=${encodeURIComponent(cat)}&`;
        if (impact) url += `impact=${encodeURIComponent(impact)}&`;

        const resp = await fetch(url, {
            headers: { 'Accept': 'application/json' }
        });
        if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
        const data = await resp.json();

        renderRegulatoryWatchTable(data.items || []);

        // Mettre à jour les compteurs
        const kpiCrit = document.getElementById("reg-kpi-critique");
        if (kpiCrit) kpiCrit.innerText = data.critique_count || 0;
        const kpiMaj = document.getElementById("reg-kpi-majeur");
        if (kpiMaj) kpiMaj.innerText = data.majeur_count || 0;
        const kpiTotal = document.getElementById("reg-kpi-total");
        if (kpiTotal) kpiTotal.innerText = data.total || 0;
    } catch (err) {
        tbody.innerHTML = `<tr><td colspan="6" style="padding:20px; text-align:center; color:#dc2626;">Erreur de chargement : ${escapeHTML(err.message)}</td></tr>`;
    }
}

function renderRegulatoryWatchTable(items) {
    const tbody = document.getElementById("reg-watch-tbody");
    if (!tbody) return;

    if (items.length === 0) {
        tbody.innerHTML = `<tr><td colspan="6" style="padding:25px; text-align:center; color:var(--text-dim);">Aucun texte juridique ne correspond aux filtres</td></tr>`;
        return;
    }

    let html = '';
    items.forEach(it => {
        let impactBadge = '';
        if (it.impact_level === 'critique') {
            impactBadge = `<span style="background:#fee2e2; color:#991b1b; font-weight:800; padding:2px 8px; border-radius:10px; font-size:10px;">CRITIQUE</span>`;
        } else if (it.impact_level === 'majeur') {
            impactBadge = `<span style="background:#fef3c7; color:#b45309; font-weight:700; padding:2px 8px; border-radius:10px; font-size:10px;">MAJEUR</span>`;
        } else {
            impactBadge = `<span style="background:var(--bg-sub); color:var(--text-dim); font-weight:600; padding:2px 8px; border-radius:10px; font-size:10px;">MINEUR</span>`;
        }

        let compBadge = '';
        if (it.compliance_status === 'conforme') {
            compBadge = `<span style="background:#dcfce7; color:#15803d; font-weight:700; padding:2px 8px; border-radius:10px; font-size:10px;">CONFORME</span>`;
        } else if (it.compliance_status === 'plan_action_requis') {
            compBadge = `<span style="background:#fee2e2; color:#dc2626; font-weight:800; padding:2px 8px; border-radius:10px; font-size:10px;">PLAN REQUIS</span>`;
        } else {
            compBadge = `<span style="background:#eff6ff; color:#2563eb; font-weight:600; padding:2px 8px; border-radius:10px; font-size:10px;">EN VIGUEUR</span>`;
        }

        const pubDateStr = (it.published_date || '').substring(0, 10);
        const effDateStr = (it.effective_date || '').substring(0, 10);

        html += `
            <tr style="border-bottom:1px solid var(--border-color);">
                <td style="padding:10px;">
                    <div style="font-weight:700; font-size:12px; color:var(--text-main);">${escapeHTML(it.title_fr)}</div>
                    ${it.title_ar ? `<div style="font-size:11px; color:var(--text-dim); direction:rtl; margin-top:2px;">${escapeHTML(it.title_ar)}</div>` : ''}
                    <div style="font-size:10px; color:var(--accent-blue); font-weight:600; margin-top:3px; font-family:monospace;">${escapeHTML(it.legal_reference)}</div>
                </td>
                <td style="padding:10px; font-size:11px;">
                    <div style="font-weight:600; color:var(--text-main);">${escapeHTML(it.joradp_number || 'Bulletin ANPP')}</div>
                    <div style="font-size:10px; color:var(--text-dim); margin-top:2px;">Publié : ${pubDateStr}</div>
                    <div style="font-size:10px; color:var(--text-dim);">Effet : ${effDateStr}</div>
                </td>
                <td style="padding:10px; font-size:11px; max-width:280px;">
                    <div style="color:var(--text-main); line-height:1.4;">${escapeHTML(it.summary_fr)}</div>
                    ${it.impact_analysis ? `<div style="font-size:10px; color:var(--accent-emerald); font-weight:600; margin-top:4px;">🔍 Analyse DT : ${escapeHTML(it.impact_analysis)}</div>` : ''}
                </td>
                <td style="padding:10px; text-align:center;">
                    ${impactBadge}
                </td>
                <td style="padding:10px; text-align:center;">
                    ${compBadge}
                </td>
            </tr>
        `;
    });

    tbody.innerHTML = html;
}

// Enregistrement d'un nouveau texte juridique
async function executeCreateRegulatoryWatch(e) {
    if (e) e.preventDefault();

    const title = document.getElementById("reg-title").value.trim();
    const titleAr = document.getElementById("reg-title-ar").value.trim();
    const ref = document.getElementById("reg-ref").value.trim();
    const joradp = document.getElementById("reg-joradp").value.trim();
    const pub = document.getElementById("reg-pub-date").value;
    const eff = document.getElementById("reg-eff-date").value;
    const cat = document.getElementById("reg-cat").value;
    const impact = document.getElementById("reg-impact").value;
    const summary = document.getElementById("reg-summary").value.trim();
    const analysis = document.getElementById("reg-analysis").value.trim();

    if (!title || !ref || !pub || !summary) {
        alert("Veuillez remplir tous les champs obligatoires (*).");
        return;
    }

    try {
        const payload = {
            title_fr: title,
            title_ar: titleAr || undefined,
            legal_reference: ref,
            joradp_number: joradp || undefined,
            published_date: pub,
            effective_date: eff || pub,
            category: cat,
            impact_level: impact,
            summary_fr: summary,
            impact_analysis: analysis || undefined,
            compliance_status: "conforme"
        };

        const resp = await fetch("/api/v1/regulatory-watch", {
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

        alert("Texte de veille réglementaire enregistré avec succès dans le corpus.");
        document.getElementById("form-create-reg-watch").reset();
        await loadRegulatoryWatchList();
    } catch (err) {
        alert("Erreur lors de l'enregistrement : " + err.message);
    }
}
