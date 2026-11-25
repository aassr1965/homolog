// ============================================================================
// MODULE VEILLE RÉGLEMENTAIRE & CORPUS JURIDIQUE ALGÉRIEN (JORADP / ANPP)
// Suivi des textes légaux, décrets, intitulés complets au survol & conformité BPF
// ============================================================================

const SANAD_REGULATORY_DICTIONARY = [
    {
        patterns: [/Décret\s+(?:Exécutif\s+)?(?:N°\s*)?21-82/i, /\b21-82\b/],
        fullTitle: "Décret exécutif n° 21-82 du 11 Rajab 1442 correspondant au 23 février 2021 relatif aux établissements pharmaceutiques et aux conditions de leur agrément (JORADP n° 14 du 28/02/2021)."
    },
    {
        patterns: [/Décret\s+(?:Exécutif\s+)?(?:N°\s*)?20-325/i, /\b20-325\b/],
        fullTitle: "Décret exécutif n° 20-325 du 6 Rabie Ethani 1442 correspondant au 22 novembre 2020 relatif aux modalités d'enregistrement des produits pharmaceutiques (JORADP n° 71 du 25/11/2020)."
    },
    {
        patterns: [/Décret\s+(?:Exécutif\s+)?(?:N°\s*)?20-324/i, /\b20-324\b/],
        fullTitle: "Décret exécutif n° 20-324 du 6 Rabie Ethani 1442 correspondant au 22 novembre 2020 relatif aux modalités d'homologation des dispositifs médicaux à usage de la médecine humaine (JORADP n° 71 du 25/11/2020)."
    },
    {
        patterns: [/Décret\s+(?:Exécutif\s+)?(?:N°\s*)?22-247/i, /\b22-247\b/],
        fullTitle: "Décret exécutif n° 22-247 du 26 Dhou El Kaâda 1443 correspondant au 26 juin 2022 fixant les règles de bonnes pratiques de fabrication des produits pharmaceutiques à usage de la médecine humaine (JORADP n° 44 du 29/06/2022)."
    },
    {
        patterns: [/Décret\s+(?:Exécutif\s+)?(?:N°\s*)?21-551/i, /\b21-551\b/],
        fullTitle: "Décret exécutif n° 21-551 du 25 Joumada El Oula 1443 correspondant au 30 décembre 2021 fixant les modalités de recouvrement et d'affectation du produit de la redevance des produits pharmaceutiques et des dispositifs médicaux (JORADP n° 02/2022)."
    },
    {
        patterns: [/Décret\s+(?:Exécutif\s+)?(?:N°\s*)?20-326/i, /Décret\s+(?:Exécutif\s+)?(?:N°\s*)?23-105/i],
        fullTitle: "Décret exécutif n° 20-326 du 22 novembre 2020 (modifié par le Décret exécutif n° 23-105 du 7 mars 2023) fixant les missions, la composition, l'organisation et le fonctionnement du comité économique intersectoriel des médicaments (CEIM)."
    },
    {
        patterns: [/Décret\s+(?:Exécutif\s+)?(?:N°\s*)?19-190/i],
        fullTitle: "Décret exécutif n° 19-190 du 30 Chaoual 1440 correspondant au 3 juillet 2019 fixant les missions, l'organisation et le fonctionnement de l'Agence nationale des produits pharmaceutiques (ANPP)."
    },
    {
        patterns: [/Décret\s+(?:Exécutif\s+)?(?:N°\s*)?24-112/i, /Décret\s+(?:Exécutif\s+)?(?:N°\s*)?19-379/i],
        fullTitle: "Décret exécutif n° 19-379 du 31 décembre 2019 et Décret exécutif n° 24-112 du 13 mars 2024 relatifs au contrôle administratif, technique et de sécurité ainsi qu'à la classification des stupéfiants, substances psychotropes et précurseurs."
    },
    {
        patterns: [/Arrêté\s+(?:Ministériel\s+)?(?:du\s+)?10\s+Mai(?:\s+2021)?/i],
        fullTitle: "Arrêté du 28 Ramadhan 1442 correspondant au 10 mai 2021 fixant la composition du dossier technique d'homologation des dispositifs médicaux à usage de la médecine humaine (JORADP n° 48 du 20/06/2021)."
    },
    {
        patterns: [/Arrêté\s+(?:du\s+)?23\s+Juin\s+2021/i],
        fullTitle: "Arrêté du 12 Dhou El Kaâda 1442 correspondant au 23 juin 2021 fixant les modalités de modification de la décision d'homologation des dispositifs médicaux à usage de la médecine humaine (JORADP n° 62 du 15/08/2021)."
    },
    {
        patterns: [/Arrêté\s+(?:du\s+)?22\s*(?:\/09\/|septembre\s+)2025/i],
        fullTitle: "Arrêté du 22 septembre 2025 fixant les éléments du dossier de demande d'agrément de l'établissement pharmaceutique de fabrication, les modalités de son traitement ainsi que la liste limitative des modifications à caractère substantiel."
    },
    {
        patterns: [/Arrêté\s+(?:du\s+)?30\s*(?:\/09\/|septembre\s+)2025/i],
        fullTitle: "Arrêté du 30 septembre 2025 fixant les éléments du dossier de demande d'agrément d'un établissement pharmaceutique de distribution en gros."
    },
    {
        patterns: [/Arrêté\s+(?:du\s+)?(?:01|1er)\s*(?:\/10\/|octobre\s+)2025/i],
        fullTitle: "Arrêté du 9 Rabie Ethani 1447 correspondant au 1er octobre 2025 fixant les éléments du dossier de demande d'agrément d'un établissement pharmaceutique d'exploitation des décisions d'enregistrement des produits pharmaceutiques et des décisions d'homologation des dispositifs médicaux."
    },
    {
        patterns: [/LD-DEPP07\/V01/i],
        fullTitle: "Ligne directrice relative aux modifications des décisions d'enregistrement des produits pharmaceutiques à usage de la médecine humaine (ANPP Code LD-DEPP07/V01, Version 01, 22 avril 2026)."
    },
    {
        patterns: [/Loi\s+(?:n°\s*)?18-11/i],
        fullTitle: "Loi n° 18-11 du 18 Chaoual 1439 correspondant au 2 juillet 2018 relative à la santé (JORADP n° 46 du 29/07/2018)."
    },
    {
        patterns: [/Loi\s+(?:n°\s*)?15-04/i],
        fullTitle: "Loi n° 15-04 du 11 Rabie Ethani 1436 correspondant au 1er février 2015 fixant les règles générales relatives à la signature et à la certification électroniques (JORADP n° 06)."
    },
    {
        patterns: [/Loi\s+(?:n°\s*)?18-07/i],
        fullTitle: "Loi n° 18-07 du 25 Ramadhan 1439 correspondant au 10 juin 2018 relative à la protection des personnes physiques dans le traitement des données à caractère personnel (JORADP n° 38)."
    },
    {
        patterns: [/21\s*CFR\s*Part\s*11/i],
        fullTitle: "FDA 21 CFR Part 11 : Electronic Records; Electronic Signatures — Intégrité des données, piste d'audit immuable, horodatage et signatures électroniques à double authentification."
    },
    {
        patterns: [/Annexe\s*11/i],
        fullTitle: "EU GMP Annexe 11 : Systèmes informatisés — Exigences de validation, sécurité des données ALCOA+ et gestion des risques."
    },
    {
        patterns: [/GAMP\s*5/i],
        fullTitle: "ISPE GAMP 5 (2e édition) : A Risk-Based Approach to Compliant GxP Computerized Systems."
    }
];

function resolveFullRegulatoryTitle(text) {
    if (!text) return "";
    const matched = [];
    for (const entry of SANAD_REGULATORY_DICTIONARY) {
        if (entry.patterns.some(rx => rx.test(text))) {
            matched.push(entry.fullTitle);
        }
    }
    return matched.join("\n• ");
}

function applyRegulatoryTooltips(root) {
    const container = root || document.body;
    if (!container || !container.querySelectorAll) return;

    const candidates = container.querySelectorAll(
        '.count-badge, .status-tag, .btn-pill, strong, span, td, th, label, option, div.tree-node'
    );
    candidates.forEach(el => {
        if (el.dataset && el.dataset.regTooltipDone === "1") return;
        // Éviter d'écraser les conteneurs volumineux (garder uniquement les éléments compacts ou feuilles)
        if (el.children.length > 2) return;
        const txt = (el.textContent || "").trim();
        if (!txt || txt.length > 160) return;

        const fullTitle = resolveFullRegulatoryTitle(txt);
        if (fullTitle) {
            const existingTitle = (el.getAttribute("title") || "").trim();
            if (!existingTitle) {
                el.setAttribute("title", fullTitle);
            } else if (!existingTitle.includes(fullTitle.substring(0, 28))) {
                el.setAttribute("title", existingTitle + "\n\n📜 " + fullTitle);
            }
            el.setAttribute("data-reg-tooltip", "true");
            if (el.dataset) el.dataset.regTooltipDone = "1";
        }
    });
}

let _regTooltipTimer = null;
function scheduleRegulatoryTooltips(root) {
    if (_regTooltipTimer) clearTimeout(_regTooltipTimer);
    _regTooltipTimer = setTimeout(() => applyRegulatoryTooltips(root), 80);
}

if (typeof document !== "undefined") {
    document.addEventListener("DOMContentLoaded", () => {
        applyRegulatoryTooltips(document.body);
        if (typeof MutationObserver !== "undefined") {
            const observer = new MutationObserver(() => scheduleRegulatoryTooltips(document.body));
            observer.observe(document.body, { childList: true, subtree: true });
        }
    });
}

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
        const fullRefTitle = resolveFullRegulatoryTitle(it.legal_reference || it.title_fr) || it.title_fr;

        html += `
            <tr style="border-bottom:1px solid var(--border-color);">
                <td style="padding:10px;">
                    <div style="font-weight:700; font-size:12px; color:var(--text-main);" title="${escapeHTML(fullRefTitle)}">${escapeHTML(it.title_fr)}</div>
                    ${it.title_ar ? `<div style="font-size:11px; color:var(--text-dim); direction:rtl; margin-top:2px;">${escapeHTML(it.title_ar)}</div>` : ''}
                    <div style="font-size:10px; color:var(--accent-blue); font-weight:600; margin-top:3px; font-family:monospace;" data-reg-tooltip="true" title="${escapeHTML(fullRefTitle)}">${escapeHTML(it.legal_reference)}</div>
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
