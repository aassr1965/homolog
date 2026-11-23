// ============================================================================
// MODULE SANAD CORPORATE — AGRÉMENTS ÉTABLISSEMENTS PHARMACEUTIQUES MIPH
// Conforme Décret exécutif n° 21-82, Arrêtés sept/oct 2025 & Passerelle Tabadol
// ============================================================================

let currentCorporateApplications = [];
let currentCorporateLicenses = [];
let selectedCorporateAppID = null;
let currentCorporateFilter = 'ALL'; // ALL, FABRICATION, IMPORTATION, DISTRIBUTION, EXPLOITATION, LICENSES
let currentCorporatePreviewTab = 'checklist';

// ============================================================================
// 1. BADGES & COMPTEURS
// ============================================================================

function updateCorporateBadgeCount(count) {
    const badge = document.getElementById("badge-corporate-count");
    if (badge) {
        if (count > 0) {
            badge.textContent = count;
            badge.style.display = "inline-block";
        } else {
            badge.style.display = "none";
        }
    }
}

async function refreshCorporateBadgeCount() {
    try {
        const resp = await fetch('/api/v1/corporate/applications');
        if (resp.ok) {
            const data = await resp.json();
            const list = data.applications || [];
            updateCorporateBadgeCount(list.length);
        }
    } catch (e) {
        console.warn("Échec actualisation badge corporate", e);
    }
}

// ============================================================================
// 2. ARBORESCENCE CONTEXTUELLE CORPORATE
// ============================================================================

function renderCorporateTree() {
    const container = document.getElementById("tree-root");
    if (!container) return;

    container.innerHTML = `
        <div style="padding: 10px 10px 6px;">
            <button class="btn-pill" style="width:100%; background:var(--accent-blue); color:white; border:none; justify-content:center; font-weight:700; height:32px; border-radius:6px; cursor:pointer;" onclick="openNewCorporateAppModal()" title="Initier une nouvelle demande d'agrément d'établissement pharmaceutique (Décret 21-82)">
                ➕ Nouvelle Demande MIPH
            </button>
        </div>
        <div class="tree-node ${currentCorporateFilter === 'ALL' ? 'selected' : ''}" onclick="selectCorporateFilter('ALL')">
            <span class="chevron">📁</span>
            <span>Toutes les Demandes MIPH</span>
            <span class="count-badge ok" id="tree-cnt-corp-all">-</span>
        </div>
        <div class="tree-node ${currentCorporateFilter === 'FABRICATION' ? 'selected' : ''}" onclick="selectCorporateFilter('FABRICATION')">
            <span class="chevron">🏭</span>
            <span>Fabrication Pharmaceutique</span>
            <span class="count-badge">Art. 4</span>
        </div>
        <div class="tree-node ${currentCorporateFilter === 'IMPORTATION' ? 'selected' : ''}" onclick="selectCorporateFilter('IMPORTATION')">
            <span class="chevron">🚢</span>
            <span>Importation Médicaments &amp; DM</span>
            <span class="count-badge">Tabadol</span>
        </div>
        <div class="tree-node ${currentCorporateFilter === 'DISTRIBUTION' ? 'selected' : ''}" onclick="selectCorporateFilter('DISTRIBUTION')">
            <span class="chevron">🚚</span>
            <span>Distribution en Gros</span>
            <span class="count-badge">BPD</span>
        </div>
        <div class="tree-node ${currentCorporateFilter === 'EXPLOITATION' ? 'selected' : ''}" onclick="selectCorporateFilter('EXPLOITATION')">
            <span class="chevron">📑</span>
            <span>Exploitation de Spécialités</span>
            <span class="count-badge">Titulaire</span>
        </div>
        <div style="margin: 12px 8px 6px; border-top: 1px solid var(--border-color); padding-top: 8px;">
            <div style="font-size: 10px; font-weight: 700; color: var(--text-dim); text-transform: uppercase; margin-bottom: 6px;">
                Veille &amp; Titres Délivrés
            </div>
            <div class="tree-node ${currentCorporateFilter === 'LICENSES' ? 'selected' : ''}" onclick="selectCorporateFilter('LICENSES')">
                <span class="chevron">📜</span>
                <span>Registre Agréments Actifs</span>
                <span class="count-badge warn">J-180 / J-90</span>
            </div>
        </div>
    `;

    // Met à jour les compteurs dans l'arbre
    if (currentCorporateApplications.length > 0) {
        const elAll = document.getElementById("tree-cnt-corp-all");
        if (elAll) elAll.textContent = currentCorporateApplications.length;
    }
}

function selectCorporateFilter(filter) {
    currentCorporateFilter = filter;
    renderCorporateTree();
    renderCorporateView(filter);
}

// ============================================================================
// 3. VUE PRINCIPALE CORPORATE (TABLEAU APPLICATIONS OU REGISTRE AGRÉMENTS)
// ============================================================================

function renderCorporateToolbar() {
    const tb = document.getElementById("main-toolbar");
    if (!tb) return;
    tb.innerHTML = `
        <button class="btn-pill" style="background:var(--accent-blue); color:white; border:none; font-weight:700; cursor:pointer;" onclick="openNewCorporateAppModal()" title="Créer une nouvelle demande d'agrément (Décret 21-82)">➕ Nouvelle Demande</button>
        <button class="btn-table-icon" onclick="selectCorporateFilter('ALL')" title="Toutes les Demandes MIPH">📁</button>
        <button class="btn-table-icon" onclick="selectCorporateFilter('LICENSES')" title="Registre des Agréments Ministériels Actifs">📜</button>
        <button class="btn-table-icon" onclick="renderCorporateView()" title="Actualiser la liste">🔄</button>
        <button class="btn-table-icon" onclick="openAuditModal()" title="Journal d'Audit Trail Réglementaire">🛡️</button>
        <button class="btn-table-icon" onclick="togglePreviewPane()" title="Afficher / Masquer le volet latéral (Alt+P)">👁️</button>
        <div style="margin-left:auto; font-size:11px; color:var(--text-muted);" id="corporate-progress">
            Passerelle Tabadol : <strong style="color:var(--accent-emerald);">Connectée &amp; Sécurisée</strong>
        </div>
    `;
}

async function renderCorporateView(filter) {
    if (filter) currentCorporateFilter = filter;
    renderCorporateToolbar();

    const thead = document.getElementById("table-head");
    const tbody = document.getElementById("doc-tbody");
    if (!tbody || !thead) return;

    if (window.innerWidth < 768 && typeof toggleMobileTree === 'function') {
        toggleMobileTree(false);
    }

    if (currentCorporateFilter === 'LICENSES') {
        await renderCorporateLicensesTable();
        return;
    }

    thead.innerHTML = `
        <tr>
            <th style="width: 110px;">Statut Réglem.</th>
            <th>N° Demande &amp; Objet de l'Agrément MIPH</th>
            <th>Catégorie (Décret 21-82)</th>
            <th>Nature de Demande</th>
            <th class="col-hide-md">Établissement</th>
            <th style="width: 140px;">Complétude</th>
            <th class="col-hide-sm">Passerelle Tabadol</th>
            <th style="width: 160px; text-align:center;">Actions Réglementaires</th>
        </tr>
    `;

    tbody.innerHTML = `<tr><td colspan="8" style="text-align:center; padding:30px; color:var(--text-muted);"><span class="spin">⏳</span> Chargement des dossiers d'agrément ministériel MIPH...</td></tr>`;

    try {
        let url = '/api/v1/corporate/applications';
        if (currentCorporateFilter && currentCorporateFilter !== 'ALL') {
            url += '?category=' + encodeURIComponent(currentCorporateFilter);
        }
        const resp = await fetch(url);
        if (!resp.ok) throw new Error("Erreur HTTP " + resp.status);
        const data = await resp.json();
        currentCorporateApplications = data.applications || [];
        updateCorporateBadgeCount(currentCorporateApplications.length);

        const elAll = document.getElementById("tree-cnt-corp-all");
        if (elAll) elAll.textContent = currentCorporateApplications.length;

        if (currentCorporateApplications.length === 0) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="8" style="text-align:center; padding:50px; color:var(--text-muted);">
                        <div style="font-size:32px; margin-bottom:12px;">🏛️</div>
                        <div style="font-weight:700; font-size:14px; color:var(--text-main); margin-bottom:6px;">
                            Aucune demande d'agrément ministériel enregistrée
                        </div>
                        <div style="font-size:12px; color:var(--text-dim); max-width:480px; margin:0 auto 16px;">
                            Conformément au Décret exécutif n° 21-82, constituez le dossier d'agrément de votre établissement pharmaceutique (Fabrication, Importation, Distribution ou Exploitation) avec télétransmission Tabadol.
                        </div>
                        <button class="btn-pill" style="background:var(--accent-blue); color:white; border:none; padding:8px 16px; font-weight:700;" onclick="openNewCorporateAppModal()">
                            ➕ Initier une Demande d'Agrément MIPH
                        </button>
                    </td>
                </tr>
            `;
            const preview = document.getElementById("preview-body");
            if (preview) preview.innerHTML = `<div style="text-align:center; padding:40px; color:var(--text-muted);">Sélectionnez une demande pour consulter la checklist Décret 21-82 et piloter la télésoumission Tabadol.</div>`;
            return;
        }

        tbody.innerHTML = "";
        currentCorporateApplications.forEach((app) => {
            const tr = document.createElement("tr");
            if (app.id === selectedCorporateAppID) tr.classList.add("selected");
            tr.onclick = (e) => {
                if (e.target.closest('button')) return;
                viewCorporateChecklist(app.id);
            };

            // Badge Statut
            let statusBadge = `<span class="status-tag status-req" style="background:#64748b; color:white;">Brouillon</span>`;
            if (app.status === 'VISA_PDT_VALIDE') {
                statusBadge = `<span class="status-tag status-ok" style="background:#0284c7; color:white;">Visa PDT Apposé</span>`;
            } else if (app.status === 'SOUMIS_TABADOL') {
                statusBadge = `<span class="status-tag status-rev" style="background:var(--accent-blue); color:white;">Soumis Tabadol</span>`;
            } else if (app.status === 'RECEPISE_BO_DELIVRE') {
                statusBadge = `<span class="status-tag status-rev" style="background:#d97706; color:white;">Récépissé BO</span>`;
            } else if (app.status === 'INSPECTION_MIPH_PROGRAMMEE') {
                statusBadge = `<span class="status-tag status-rev" style="background:#ea580c; color:white;">Inspection MIPH</span>`;
            } else if (app.status === 'AGREMENT_OCTROYE') {
                statusBadge = `<span class="status-tag status-ok" style="background:#059669; color:white; font-weight:700;">Agrément Octroyé</span>`;
            } else if (app.status === 'REJETE') {
                statusBadge = `<span class="status-tag status-req" style="background:#dc2626; color:white;">Rejeté</span>`;
            }

            // Catégorie
            let catLabel = app.category;
            let catIcon = "🏢";
            if (app.category === 'FABRICATION') { catLabel = "Fabrication"; catIcon = "🏭"; }
            else if (app.category === 'IMPORTATION') { catLabel = "Importation"; catIcon = "🚢"; }
            else if (app.category === 'DISTRIBUTION') { catLabel = "Distribution"; catIcon = "🚚"; }
            else if (app.category === 'EXPLOITATION') { catLabel = "Exploitation"; catIcon = "📑"; }

            // Nature
            let natureLabel = app.nature;
            if (app.nature === 'AGREMENT_INITIAL') natureLabel = "Agrément Initial";
            else if (app.nature === 'RENOUVELLEMENT') natureLabel = "Renouvellement";
            else if (app.nature === 'MODIFICATION_SUBSTANTIELLE') natureLabel = "Modif. Substantielle";
            else if (app.nature === 'EXTENSION_LIGNES') natureLabel = "Extension Lignes";
            else if (app.nature === 'TRANSFERT_SITE') natureLabel = "Transfert de Site";

            // Complétude
            const pct = app.completeness_percent || 0;
            const barColor = pct >= 100 ? "var(--accent-emerald)" : (pct >= 50 ? "var(--accent-blue)" : "var(--accent-orange)");
            const compHtml = `
                <div style="display:flex; align-items:center; gap:8px;">
                    <div style="flex:1; height:6px; background:var(--bg-input); border-radius:3px; overflow:hidden;">
                        <div style="width:${pct}%; height:100%; background:${barColor};"></div>
                    </div>
                    <span style="font-size:11px; font-weight:700; width:34px; text-align:right;">${pct}%</span>
                </div>
            `;

            // Tabadol info
            let tabadolHtml = `<span style="color:var(--text-dim); font-size:10px;">Non transmis</span>`;
            if (app.tabadol_submission_ref) {
                tabadolHtml = `
                    <div style="display:flex; flex-direction:column; gap:2px;">
                        <span style="font-weight:700; color:var(--accent-blue); font-size:10px;">🌐 ${escapeHtml(app.tabadol_submission_ref)}</span>
                        <span style="font-size:9px; color:var(--text-dim);">${formatDate(app.tabadol_submitted_at)}</span>
                    </div>
                `;
            }

            tr.innerHTML = `
                <td>${statusBadge}</td>
                <td>
                    <div style="font-weight:700; color:var(--text-main); font-size:12px;">${escapeHtml(app.title)}</div>
                    <div style="font-size:10px; color:var(--text-dim); font-family:var(--font-mono);">N° ${escapeHtml(app.application_number)}</div>
                </td>
                <td>
                    <span style="display:inline-flex; align-items:center; gap:4px; font-weight:600; font-size:11px;">
                        <span>${catIcon}</span> <span>${catLabel}</span>
                    </span>
                </td>
                <td><span style="font-size:11px; color:var(--text-main);">${natureLabel}</span></td>
                <td class="col-hide-md"><span style="font-size:11px; color:var(--text-dim);">${escapeHtml(app.establishment_name || 'Établissement Principal')}</span></td>
                <td>${compHtml}</td>
                <td class="col-hide-sm">${tabadolHtml}</td>
                <td style="text-align:center;">
                    <div style="display:inline-flex; gap:5px; align-items:center; justify-content:center;">
                        <button class="btn-table-icon btn-table-icon-primary" onclick="event.stopPropagation(); viewCorporateChecklist('${app.id}')" title="Afficher la checklist des pièces réglementaires Décret 21-82">📋</button>
                        <button class="btn-table-icon" onclick="event.stopPropagation(); openTabadolAuditModal('${app.id}')" title="Audit de conformité Tabadol (Décret 21-82)">🔍</button>
                        <button class="btn-table-icon" onclick="event.stopPropagation(); downloadTabadolPackage('${app.id}')" title="Télécharger l'Archive Tabadol (.zip)">📦</button>
                    </div>
                </td>
            `;

            tbody.appendChild(tr);
        });

        // Si une application était sélectionnée, rafraîchir son aperçu, sinon sélectionner la 1ère
        if (selectedCorporateAppID) {
            renderCorporateAppPreview(selectedCorporateAppID);
        } else if (currentCorporateApplications.length > 0) {
            selectedCorporateAppID = currentCorporateApplications[0].id;
            renderCorporateAppPreview(selectedCorporateAppID);
        }

    } catch (e) {
        tbody.innerHTML = `<tr><td colspan="8" style="text-align:center; padding:30px; color:var(--accent-red);">❌ Erreur de chargement: ${escapeHtml(e.message)}</td></tr>`;
    }
}

function selectCorporateApp(appId) {
    viewCorporateChecklist(appId);
}

// ============================================================================
// CONFIGURATION DES VOLETS RÉGLEMENTAIRES DÉCRET EXÉCUTIF N° 21-82
// ============================================================================

const CORPORATE_SECTIONS_ORDER = [
    'ADMINISTRATIF',
    'PLANS_FLUX',
    'UTILITES_EQUIPEMENTS',
    'TECHNIQUE_LOCAUX',
    'ORGANISATION_BPF',
    'PHARMACIEN_DT',
    'CONDITIONS_TECHNIQUES',
    'MODIFICATIONS',
    'AUTRE'
];

const CORPORATE_SECTION_CONFIG = {
    'ADMINISTRATIF': {
        title: "Volet I : Dossier Administratif, Légal & Statutaire (Art. 4 & 5 Décret 21-82)",
        shortTitle: "Volet I : Dossier Administratif",
        badge: "Art. 4-5",
        icon: "🏛️"
    },
    'PLANS_FLUX': {
        title: "Volet II : Conception des Locaux, Zonage BPF & Synoptiques des Flux (Art. 7 à 10)",
        shortTitle: "Volet II : Locaux, Zonage & Flux",
        badge: "Art. 7-10",
        icon: "📐"
    },
    'UTILITES_EQUIPEMENTS': {
        title: "Volet III : Utilités Critiques (CVC/Eau PPI) & Qualification Équipements",
        shortTitle: "Volet III : Utilités & Équipements",
        badge: "BPF / Ph. Eur.",
        icon: "⚙️"
    },
    'TECHNIQUE_LOCAUX': {
        title: "Volet Technique : Aménagement des Locaux & Installations BPF",
        shortTitle: "Volet Technique : Locaux",
        badge: "Locaux",
        icon: "🏭"
    },
    'ORGANISATION_BPF': {
        title: "Volet IV : Système de Gestion de la Qualité (PQS) & Procédures BPF (Art. 11 à 15)",
        shortTitle: "Volet IV : Système Qualité PQS",
        badge: "Art. 11-15",
        icon: "📋"
    },
    'PHARMACIEN_DT': {
        title: "Volet V : Pharmacien Directeur Technique & Équipe Pharmaceutique (Art. 16 à 24)",
        shortTitle: "Volet V : Direction Technique",
        badge: "Art. 16-24",
        icon: "⚖️"
    },
    'CONDITIONS_TECHNIQUES': {
        title: "Volet Technique : Conditions de Stockage & Chaîne du Froid (BPDG)",
        shortTitle: "Volet Technique : BPDG",
        badge: "BPDG",
        icon: "❄️"
    },
    'MODIFICATIONS': {
        title: "Volet Modifications : Éléments Substantiels & Justificatifs (Arrêté 22/09/2025)",
        shortTitle: "Volet Modifications",
        badge: "Arrêté 2025",
        icon: "📑"
    },
    'AUTRE': {
        title: "Volet Complémentaire : Exigences Réglementaires Spécifiques",
        shortTitle: "Volet Complémentaire",
        badge: "Réglementaire",
        icon: "📁"
    }
};

let corporateCollapsedSections = new Set();
let corporateDrawerCollapsedSections = new Set();

function getCorporateSectionKey(item) {
    if (item.section === 'TECHNIQUE_LOCAUX') {
        if (item.item_code.includes('.UTL.') || item.item_code.includes('.EQP.')) {
            return 'UTILITES_EQUIPEMENTS';
        }
        if (item.item_code.includes('.PLN.') || item.item_code.includes('.FLX.')) {
            return 'PLANS_FLUX';
        }
    }
    return item.section || 'AUTRE';
}

function toggleCorporateSectionCollapse(secKey) {
    if (corporateCollapsedSections.has(secKey)) {
        corporateCollapsedSections.delete(secKey);
    } else {
        corporateCollapsedSections.add(secKey);
    }
    const isNowCollapsed = corporateCollapsedSections.has(secKey);
    const rows = document.querySelectorAll(`tr[data-corp-section="${secKey}"]`);
    rows.forEach(r => {
        r.style.display = isNowCollapsed ? 'none' : '';
    });
    const indicator = document.getElementById(`corp-sec-ind-${secKey}`);
    if (indicator) {
        indicator.textContent = isNowCollapsed ? '▶' : '▼';
    }
}

function toggleAllCorporateSections() {
    const allSecKeys = Object.keys(CORPORATE_SECTION_CONFIG);
    const anyExpanded = allSecKeys.some(k => !corporateCollapsedSections.has(k));
    if (anyExpanded) {
        allSecKeys.forEach(k => corporateCollapsedSections.add(k));
    } else {
        corporateCollapsedSections.clear();
    }
    allSecKeys.forEach(secKey => {
        const isNowCollapsed = corporateCollapsedSections.has(secKey);
        const rows = document.querySelectorAll(`tr[data-corp-section="${secKey}"]`);
        rows.forEach(r => {
            r.style.display = isNowCollapsed ? 'none' : '';
        });
        const indicator = document.getElementById(`corp-sec-ind-${secKey}`);
        if (indicator) {
            indicator.textContent = isNowCollapsed ? '▶' : '▼';
        }
    });
}

function toggleDrawerCorporateSection(secKey) {
    if (corporateDrawerCollapsedSections.has(secKey)) {
        corporateDrawerCollapsedSections.delete(secKey);
    } else {
        corporateDrawerCollapsedSections.add(secKey);
    }
    const isCollapsed = corporateDrawerCollapsedSections.has(secKey);
    const body = document.getElementById(`drawer-sec-body-${secKey}`);
    const ind = document.getElementById(`drawer-sec-ind-${secKey}`);
    if (body) body.style.display = isCollapsed ? 'none' : 'flex';
    if (ind) ind.textContent = isCollapsed ? '▶' : '▼';
}

// ============================================================================
// 3.B VUE CHECKLIST DES DOCUMENTS RÉGLEMENTAIRES À JOINDRE (DÉCRET 21-82)
// ============================================================================

async function viewCorporateChecklist(appId) {
    selectedCorporateAppID = appId;
    const thead = document.getElementById("table-head");
    const tbody = document.getElementById("doc-tbody");
    if (!tbody || !thead) return;

    tbody.innerHTML = `<tr><td colspan="6" style="text-align:center; padding:30px; color:var(--text-muted);"><span class="spin">⏳</span> Chargement de la checklist des volets Décret 21-82...</td></tr>`;

    try {
        const resp = await fetch('/api/v1/corporate/applications/' + appId);
        if (!resp.ok) throw new Error("Erreur HTTP " + resp.status);
        const data = await resp.json();
        const app = data.application;
        if (!app) throw new Error("Dossier introuvable");

        const items = app.checklist_items || [];
        const pct = app.completeness_percent || 0;

        // Mise à jour de la barre de titre / fil d'Ariane
        const rootEl = document.getElementById("bc-root");
        const secEl = document.getElementById("bc-sec");
        const prodEl = document.getElementById("bc-prod");
        const modEl = document.getElementById("bc-mod");
        if (rootEl) rootEl.innerText = "SANAD_CORPORATE";
        if (secEl) secEl.innerText = "MIPH_Décret_21-82";
        if (prodEl) prodEl.innerText = app.title;
        if (modEl) modEl.innerText = `Pièces à joindre (${items.length})`;

        // Barre d'outils contextuelle avec bouton de retour
        const tb = document.getElementById("main-toolbar");
        if (tb) {
            tb.innerHTML = `
                <button class="btn-table-icon" onclick="renderCorporateView('ALL')" title="Retour aux Demandes">←</button>
                <button class="btn-table-icon" onclick="openNewCorporateAppModal()" title="Créer une nouvelle demande d'agrément (Décret 21-82)">➕</button>
                <button class="btn-table-icon btn-table-icon-primary" onclick="openGEDPickerForCorporateItem('${app.id}', '')" title="Rattacher un document du coffre-fort GED">📎</button>
                <button class="btn-table-icon" onclick="toggleAllCorporateSections()" title="Tout déplier / Tout replier les volets réglementaires">📑</button>
                <button class="btn-table-icon" onclick="openTabadolAuditModal('${app.id}')" title="Audit de conformité Tabadol (Décret 21-82)">🔍</button>
                <button class="btn-table-icon" onclick="downloadTabadolPackage('${app.id}')" title="Télécharger l'Archive Tabadol (.zip)">📦</button>
                <button class="btn-table-icon" onclick="viewCorporateChecklist('${app.id}')" title="Actualiser la liste">🔄</button>
                <button class="btn-table-icon" onclick="togglePreviewPane()" title="Volet d'aperçu latéral (Alt+P)">👁️</button>
                <div style="margin-left:auto; font-size:11px; color:var(--text-muted);" id="corporate-progress">
                    Complétude Dossier : <strong style="color:var(--accent-emerald);">${pct}% (${items.filter(i => i.ged_document_id).length}/${items.length} pièces rattachées)</strong>
                </div>
            `;
        }

        // En-tête du tableau des documents à joindre
        thead.innerHTML = `
            <tr>
                <th style="width: 55px; text-align:center;" title="Statut de rattachement au coffre-fort GED">GED</th>
                <th style="width: 110px;">Code Exigence</th>
                <th>Intitulé du Document Réglementaire (Décret 21-82)</th>
                <th style="width: 135px; text-align:center;">Niveau d'Exigence</th>
                <th>Document du Coffre-fort GED</th>
                <th style="width: 110px; text-align:center;">Action</th>
            </tr>
        `;

        if (items.length === 0) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="6" style="text-align:center; padding:40px; color:var(--text-muted);">
                        Aucune pièce requise configurée pour cette démarche.
                    </td>
                </tr>
            `;
            return;
        }

        const groupedSections = {};
        items.forEach(it => {
            const secKey = getCorporateSectionKey(it);
            if (!groupedSections[secKey]) groupedSections[secKey] = [];
            groupedSections[secKey].push(it);
        });

        tbody.innerHTML = "";

        CORPORATE_SECTIONS_ORDER.forEach(secKey => {
            const secItems = groupedSections[secKey];
            if (!secItems || secItems.length === 0) return;

            const secCfg = CORPORATE_SECTION_CONFIG[secKey] || {
                title: `Volet : ${secKey}`,
                badge: secKey,
                icon: "📁"
            };

            const isCollapsed = corporateCollapsedSections.has(secKey);
            const totalInSec = secItems.length;
            const linkedInSec = secItems.filter(it => it.ged_document_id).length;
            const mandatoryInSec = secItems.filter(it => it.is_mandatory && it.requirement_level === 'E').length;

            // Ligne d'en-tête de section (accordéon déployable)
            const trHeader = document.createElement("tr");
            trHeader.className = "section-accordion-header";
            trHeader.onclick = () => toggleCorporateSectionCollapse(secKey);
            trHeader.title = isCollapsed ? "Cliquer pour déplier ce volet" : "Cliquer pour replier ce volet";

            trHeader.innerHTML = `
                <td colspan="6">
                    <div class="sec-banner" style="display:flex; align-items:center; gap:8px; padding:2px 4px;">
                        <span id="corp-sec-ind-${secKey}" style="font-size:11px; color:var(--text-dim); width:14px; display:inline-block; text-align:center;">
                            ${isCollapsed ? '▶' : '▼'}
                        </span>
                        <span>${secCfg.icon}</span>
                        <span style="color:var(--text-main); font-weight:800; font-size:12px;">${secCfg.title}</span>
                        <span style="font-size:9.5px; padding:2px 6px; border-radius:3px; background:rgba(2,132,199,0.12); color:var(--accent-blue); font-weight:700;">
                            ${secCfg.badge}
                        </span>
                        <div class="sec-counts" style="margin-left:auto; display:inline-flex; align-items:center; gap:8px;">
                            <span style="font-size:11px; color:${linkedInSec === totalInSec ? 'var(--accent-emerald)' : 'var(--text-muted)'}; font-weight:700;">
                                ${linkedInSec} / ${totalInSec} pièce(s) liée(s)
                            </span>
                            ${mandatoryInSec > 0 ? `<span style="font-size:10px; color:var(--accent-amber); font-weight:800; background:rgba(217,119,6,0.1); border:1px solid rgba(217,119,6,0.3); padding:1px 6px; border-radius:3px;">[${mandatoryInSec} Éliminatoire(s)]</span>` : ''}
                        </div>
                    </div>
                </td>
            `;
            tbody.appendChild(trHeader);

            // Lignes enfants du volet
            secItems.forEach(it => {
                const tr = document.createElement("tr");
                tr.setAttribute("data-corp-section", secKey);
                if (isCollapsed) tr.style.display = "none";

                // Statut GED (icône seule avec tooltip)
                let stBadge = `<span class="badge-icon-status" style="background:rgba(220,38,38,0.12); border:1px solid rgba(220,38,38,0.3);" title="Statut GED : Document manquant (exigence non couverte)">⚠️</span>`;
                if (it.ged_document_id) {
                    stBadge = `<span class="badge-icon-status" style="background:rgba(5,150,105,0.12); border:1px solid rgba(5,150,105,0.3);" title="Statut GED : Document rattaché (${escapeHtml(it.document_name || 'Document GED')})">✅</span>`;
                }

                // Criticité
                let reqBadge = `<span class="status-tag" style="background:#64748b; color:white; font-size:9px;">[S] Standard</span>`;
                if (it.requirement_level === 'ELIMINATOIRE') {
                    reqBadge = `<span class="status-tag status-req" style="background:#dc2626; color:white; font-size:9px; font-weight:800;">[E] Éliminatoire Tabadol</span>`;
                } else if (it.requirement_level === 'COMPLEMENT') {
                    reqBadge = `<span class="status-tag status-rev" style="background:#d97706; color:white; font-size:9px; font-weight:700;">[C] Complément</span>`;
                } else if (it.requirement_level === 'FACULTATIF') {
                    reqBadge = `<span class="status-tag" style="background:#0284c7; color:white; font-size:9px;">[F] Facultatif</span>`;
                }

                // Info GED rattaché
                let docInfo = `<span style="color:var(--text-dim); font-size:11px; font-style:italic;">Aucune pièce rattachée</span>`;
                let actionBtn = `
                    <button class="btn-table-icon" onclick="openGEDPickerForCorporateItem('${app.id}', '${it.id}')" title="Rattacher un document du coffre-fort GED">
                        📎
                    </button>
                `;

                if (it.ged_document_id) {
                    docInfo = `
                        <div style="display:flex; flex-direction:column; gap:2px;">
                            <a href="javascript:void(0)" onclick="openCorporateDocumentViewer('${it.ged_document_id}', '${escapeHtml(it.document_name)}', '${it.item_code}')" style="font-weight:700; color:var(--accent-blue); font-size:11px; text-decoration:underline; display:inline-flex; align-items:center; gap:4px;" title="Cliquer pour visionner la pièce jointe">
                                <span>📄 ${escapeHtml(it.document_name || 'Document GED')}</span>
                            </a>
                            <span style="font-size:9px; color:var(--text-dim); font-family:var(--font-mono);">
                                Réf: ${escapeHtml(it.document_code || '-')} • SHA-256: ${escapeHtml(it.file_sha256 ? it.file_sha256.substring(0, 16) + '...' : '-')}
                            </span>
                        </div>
                    `;
                    actionBtn = `
                        <div style="display:inline-flex; gap:5px; align-items:center; justify-content:center;">
                            <button class="btn-table-icon btn-table-icon-primary" onclick="openCorporateDocumentViewer('${it.ged_document_id}', '${escapeHtml(it.document_name)}', '${it.item_code}')" title="Visionner la pièce jointe">👁️</button>
                            <button class="btn-table-icon" onclick="openGEDPickerForCorporateItem('${app.id}', '${it.id}')" title="Remplacer par un autre document GED">🔄</button>
                            <button class="btn-table-icon btn-table-icon-del" onclick="detachGEDDocumentFromItem('${app.id}', '${it.id}')" title="Détacher la pièce de l'exigence">❌</button>
                        </div>
                    `;
                }

                tr.innerHTML = `
                    <td style="text-align:center;">${stBadge}</td>
                    <td><span style="font-family:var(--font-mono); font-weight:700; color:var(--accent-blue); font-size:11px;">${escapeHtml(it.item_code)}</span></td>
                    <td>
                        <div style="font-weight:700; color:var(--text-main); font-size:12px;">${escapeHtml(it.title)}</div>
                    </td>
                    <td style="text-align:center;">${reqBadge}</td>
                    <td>${docInfo}</td>
                    <td style="text-align:center;">${actionBtn}</td>
                `;

                tbody.appendChild(tr);
            });
        });

        // Met également à jour le volet d'aperçu latéral
        renderCorporateAppPreview(appId);

    } catch (e) {
        tbody.innerHTML = `<tr><td colspan="6" style="text-align:center; padding:30px; color:var(--accent-red);">❌ Erreur: ${escapeHtml(e.message)}</td></tr>`;
    }
}

// ============================================================================
// 4. REGISTRE DES AGRÉMENTS MINISTÉRIELS DÉLIVRÉS (AVEC COUNTDOWN LÉGAL)
// ============================================================================

async function renderCorporateLicensesTable() {
    renderCorporateToolbar();
    const thead = document.getElementById("table-head");
    const tbody = document.getElementById("doc-tbody");
    if (!tbody || !thead) return;

    thead.innerHTML = `
        <tr>
            <th style="width: 100px;">Validité Légale</th>
            <th>N° Agrément Ministériel MIPH</th>
            <th>Catégorie de l'Établissement</th>
            <th>Établissement Titulaire</th>
            <th class="col-hide-md">Pharmacien Directeur Technique</th>
            <th class="col-hide-sm">Date d'Octroi</th>
            <th>Échéance Quinquennale</th>
            <th style="width: 170px; text-align:center;">Compte à Rebours &amp; Renouvellement</th>
        </tr>
    `;

    tbody.innerHTML = `<tr><td colspan="8" style="text-align:center; padding:30px; color:var(--text-muted);"><span class="spin">⏳</span> Chargement du registre des agréments ministériels...</td></tr>`;

    try {
        const resp = await fetch('/api/v1/corporate/licenses');
        if (!resp.ok) throw new Error("Erreur HTTP " + resp.status);
        const data = await resp.json();
        currentCorporateLicenses = data.licenses || [];

        if (currentCorporateLicenses.length === 0) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="8" style="text-align:center; padding:50px; color:var(--text-muted);">
                        <div style="font-size:32px; margin-bottom:12px;">📜</div>
                        <div style="font-weight:700; font-size:14px; color:var(--text-main); margin-bottom:6px;">
                            Aucun agrément ministériel finalisé dans le registre
                        </div>
                        <div style="font-size:12px; color:var(--text-dim); max-width:460px; margin:0 auto;">
                            Dès qu'une demande d'agrément franchit l'instruction MIPH et que l'Arrêté ministériel est enregistré, le titre et son compte à rebours de validité quinquennale (Décret 21-82) s'affichent ici.
                        </div>
                    </td>
                </tr>
            `;
            return;
        }

        tbody.innerHTML = "";
        currentCorporateLicenses.forEach((lic) => {
            const tr = document.createElement("tr");

            // Calcul du badge de compte à rebours
            const days = lic.days_remaining != null ? lic.days_remaining : 999;
            let cdBadge = `<span class="status-tag status-ok" style="background:#059669; color:white; font-weight:700;">🟢 ${days} j restants (Conforme)</span>`;
            let statusTag = `<span class="status-tag status-ok">ACTIF</span>`;

            if (days < 0) {
                cdBadge = `<span class="status-tag status-req" style="background:#1e293b; color:#f87171; font-weight:800;">⚫ EXPIRÉ (${Math.abs(days)} j de retard)</span>`;
                statusTag = `<span class="status-tag status-req">CADUC</span>`;
            } else if (days <= 30) {
                cdBadge = `<span class="status-tag status-req" style="background:#dc2626; color:white; font-weight:800; animation:pulse 1.5s infinite;">🔴 J-${days} (URGENCE ABSOLUE)</span>`;
            } else if (days <= 90) {
                cdBadge = `<span class="status-tag status-rev" style="background:#ea580c; color:white; font-weight:700;">🟠 J-${days} (Dépôt Tabadol Requis)</span>`;
            } else if (days <= 180) {
                cdBadge = `<span class="status-tag status-rev" style="background:#d97706; color:white; font-weight:700;">🟡 J-${days} (Alerte Décret 21-82)</span>`;
            }

            let catIcon = "🏢";
            if (lic.category === 'FABRICATION') catIcon = "🏭";
            else if (lic.category === 'IMPORTATION') catIcon = "🚢";
            else if (lic.category === 'DISTRIBUTION') catIcon = "🚚";
            else if (lic.category === 'EXPLOITATION') catIcon = "📑";

            tr.innerHTML = `
                <td>${statusTag}</td>
                <td>
                    <div style="font-weight:700; color:var(--text-main); font-family:var(--font-mono); font-size:12px;">
                        🏛️ ${escapeHtml(lic.license_number)}
                    </div>
                    <div style="font-size:10px; color:var(--text-dim);">${escapeHtml(lic.decision_ref || 'Arrêté Ministériel')}</div>
                </td>
                <td><span style="font-weight:600; font-size:11px;">${catIcon} ${escapeHtml(lic.category)}</span></td>
                <td><span style="font-weight:600; color:var(--text-main); font-size:11px;">${escapeHtml(lic.establishment_name || 'Établissement Titulaire')}</span></td>
                <td class="col-hide-md"><span style="font-size:11px; color:var(--text-dim);">Dr. ${escapeHtml(lic.technical_director_name || 'Directeur Technique')}</span></td>
                <td class="col-hide-sm"><span style="font-size:11px;">${formatDate(lic.grant_date)}</span></td>
                <td><strong style="font-size:11px; color:var(--text-main);">${formatDate(lic.expiry_date)}</strong></td>
                <td style="text-align:center;">${cdBadge}</td>
            `;

            tbody.appendChild(tr);
        });

    } catch (e) {
        tbody.innerHTML = `<tr><td colspan="8" style="text-align:center; padding:30px; color:var(--accent-red);">❌ Erreur: ${escapeHtml(e.message)}</td></tr>`;
    }
}

// ============================================================================
// 5. VOLET D'APERÇU & DETAIL DEMANDE D'AGRÉMENT (PANE-PREVIEW)
// ============================================================================

async function renderCorporateAppPreview(appId) {
    const preview = document.getElementById("preview-body");
    if (!preview) return;

    preview.innerHTML = `<div style="text-align:center; padding:40px; color:var(--text-muted);"><span class="spin">⏳</span> Chargement du dossier d'agrément...</div>`;

    try {
        const resp = await fetch('/api/v1/corporate/applications/' + appId);
        if (!resp.ok) throw new Error("Erreur HTTP " + resp.status);
        const data = await resp.json();
        const app = data.application;
        if (!app) throw new Error("Dossier introuvable");

        const pct = app.completeness_percent || 0;
        const items = app.checklist_items || [];

        // Boutons de changement de sous-onglets
        const tabsHtml = `
            <div style="display:flex; border-bottom:1px solid var(--border-color); background:var(--bg-card); padding:4px 8px; gap:4px; margin-bottom:12px;">
                <button class="btn-pill ${currentCorporatePreviewTab === 'checklist' ? 'active' : ''}" style="border:none; font-size:11px; padding:6px 12px;" onclick="switchCorporatePreviewTab('${app.id}', 'checklist')">
                    📋 Checklist MIPH (${items.length})
                </button>
                <button class="btn-pill ${currentCorporatePreviewTab === 'eform' ? 'active' : ''}" style="border:none; font-size:11px; padding:6px 12px;" onclick="switchCorporatePreviewTab('${app.id}', 'eform')">
                    📝 Formulaire e-Form
                </button>
                <button class="btn-pill ${currentCorporatePreviewTab === 'tabadol' ? 'active' : ''}" style="border:none; font-size:11px; padding:6px 12px;" onclick="switchCorporatePreviewTab('${app.id}', 'tabadol')">
                    🌐 Tabadol &amp; Visa PDT
                </button>
                <button class="btn-pill ${currentCorporatePreviewTab === 'miph' ? 'active' : ''}" style="border:none; font-size:11px; padding:6px 12px;" onclick="switchCorporatePreviewTab('${app.id}', 'miph')">
                    🏛️ Récépissé &amp; Arrêté
                </button>
            </div>
        `;

        let contentHtml = '';
        if (currentCorporatePreviewTab === 'checklist') {
            contentHtml = renderChecklistTabHtml(app);
        } else if (currentCorporatePreviewTab === 'eform') {
            contentHtml = renderEFormTabHtml(app);
        } else if (currentCorporatePreviewTab === 'tabadol') {
            contentHtml = renderTabadolTabHtml(app);
        } else if (currentCorporatePreviewTab === 'miph') {
            contentHtml = renderMIPHTabHtml(app);
        }

        preview.innerHTML = `
            <div style="padding:16px;">
                <!-- En-tête de la demande -->
                <div style="background:var(--bg-card); border:1px solid var(--border-color); border-radius:8px; padding:14px; margin-bottom:12px;">
                    <div style="display:flex; justify-content:space-between; align-items:flex-start; margin-bottom:8px;">
                        <div>
                            <div style="font-size:10px; font-weight:700; color:var(--text-dim); text-transform:uppercase; letter-spacing:0.5px;">
                                Demande d'Agrément MIPH • Décret 21-82
                            </div>
                            <h3 style="margin:2px 0 4px; font-size:15px; color:var(--text-main); font-weight:700;">
                                ${escapeHtml(app.title)}
                            </h3>
                            <div style="font-size:11px; font-family:var(--font-mono); color:var(--accent-blue);">
                                ${escapeHtml(app.application_number)}
                            </div>
                        </div>
                        <div style="text-align:right;">
                            <span class="status-tag status-ok" style="background:#0284c7; color:white; font-size:10px; font-weight:700;">
                                ${escapeHtml(app.status)}
                            </span>
                        </div>
                    </div>

                    <div style="display:grid; grid-template-columns: repeat(auto-fit, minmax(130px, 1fr)); gap:8px; margin-top:10px; font-size:11px;">
                        <div>
                            <span style="color:var(--text-dim);">Catégorie :</span>
                            <strong>${escapeHtml(app.category)}</strong>
                        </div>
                        <div>
                            <span style="color:var(--text-dim);">Nature :</span>
                            <strong>${escapeHtml(app.nature)}</strong>
                        </div>
                        <div>
                            <span style="color:var(--text-dim);">Établissement :</span>
                            <strong>${escapeHtml(app.establishment_name || 'Non rattaché')}</strong>
                        </div>
                        <div>
                            <span style="color:var(--text-dim);">Complétude :</span>
                            <strong style="color:var(--accent-emerald);">${pct}%</strong>
                        </div>
                    </div>
                </div>

                <!-- Navigation des sous-onglets -->
                ${tabsHtml}

                <!-- Contenu du sous-onglet sélectionné -->
                <div>
                    ${contentHtml}
                </div>
            </div>
        `;

    } catch (e) {
        preview.innerHTML = `<div style="text-align:center; padding:30px; color:var(--accent-red);">❌ Erreur: ${escapeHtml(e.message)}</div>`;
    }
}

function switchCorporatePreviewTab(appId, tab) {
    currentCorporatePreviewTab = tab;
    renderCorporateAppPreview(appId);
}

// ----------------------------------------------------------------------------
// A. Rendu Sous-onglet Checklist Décret 21-82
// ----------------------------------------------------------------------------
function renderChecklistTabHtml(app) {
    const items = app.checklist_items || [];
    if (items.length === 0) {
        return `<div style="text-align:center; padding:20px; color:var(--text-dim);">Aucune exigence configurée pour ce type d'établissement.</div>`;
    }

    // Regrouper par volet réglementaire Décret 21-82
    const sections = {};
    items.forEach(it => {
        const secKey = getCorporateSectionKey(it);
        if (!sections[secKey]) sections[secKey] = [];
        sections[secKey].push(it);
    });

    let html = `
        <div style="display:flex; justify-content:space-between; align-items:center; margin-bottom:10px;">
            <div style="font-size:12px; font-weight:700; color:var(--text-main);">
                Volets Réglementaires &amp; Exigences (Décret 21-82)
            </div>
            <button class="btn-table-icon btn-table-icon-primary" onclick="openTabadolAuditModal('${app.id}')" title="Évaluer la conformité Tabadol (Décret 21-82)">
                🔍
            </button>
        </div>
        <div style="display:flex; flex-direction:column; gap:10px;">
    `;

    const orderedKeys = CORPORATE_SECTIONS_ORDER.filter(k => sections[k] && sections[k].length > 0);
    Object.keys(sections).forEach(k => {
        if (!orderedKeys.includes(k)) orderedKeys.push(k);
    });

    orderedKeys.forEach(secKey => {
        const secItems = sections[secKey];
        const cfg = CORPORATE_SECTION_CONFIG[secKey] || {
            title: `Volet : ${secKey}`,
            shortTitle: secKey,
            badge: "Réglementaire",
            icon: "📁"
        };
        const isCollapsed = corporateDrawerCollapsedSections.has(secKey);
        const attachedCount = secItems.filter(i => i.ged_document_id).length;
        const totalCount = secItems.length;

        html += `
            <div style="background:var(--bg-card); border:1px solid var(--border-color); border-radius:6px; overflow:hidden;">
                <div onclick="toggleDrawerCorporateSection('${secKey}')" style="cursor:pointer; background:rgba(255,255,255,0.03); border-bottom:1px solid var(--border-color); padding:8px 10px; font-weight:700; font-size:11px; display:flex; justify-content:space-between; align-items:center; user-select:none;">
                    <span style="display:inline-flex; align-items:center; gap:6px;">
                        <span id="drawer-sec-ind-${secKey}" style="font-size:10px; color:var(--text-dim);">${isCollapsed ? '▶' : '▼'}</span>
                        <span>${cfg.icon} ${escapeHtml(cfg.shortTitle || cfg.title)}</span>
                    </span>
                    <span style="display:inline-flex; gap:6px; align-items:center;">
                        <span style="font-size:9px; background:rgba(59,130,246,0.15); color:var(--accent-blue); padding:1px 5px; border-radius:3px;">${escapeHtml(cfg.badge)}</span>
                        <span style="color:${attachedCount === totalCount ? 'var(--accent-green)' : 'var(--text-dim)'}; font-size:10px; font-weight:600;">${attachedCount}/${totalCount}</span>
                    </span>
                </div>
                <div id="drawer-sec-body-${secKey}" style="display:${isCollapsed ? 'none' : 'flex'}; flex-direction:column;">
        `;

        secItems.forEach(it => {
            // Badges de criticité
            let reqBadge = `<span class="status-tag" style="background:#64748b; color:white; font-size:9px;">[S] Standard</span>`;
            if (it.requirement_level === 'ELIMINATOIRE') {
                reqBadge = `<span class="status-tag status-req" style="background:#dc2626; color:white; font-size:9px; font-weight:800;">[E] Éliminatoire</span>`;
            } else if (it.requirement_level === 'COMPLEMENT') {
                reqBadge = `<span class="status-tag status-rev" style="background:#d97706; color:white; font-size:9px; font-weight:700;">[C] Complément</span>`;
            } else if (it.requirement_level === 'FACULTATIF') {
                reqBadge = `<span class="status-tag" style="background:#0284c7; color:white; font-size:9px;">[F] Facultatif</span>`;
            }

            // Statut de rattachement GED
            let gedInfo = '';
            if (it.ged_document_id) {
                gedInfo = `
                    <div style="display:flex; align-items:center; gap:6px; margin-top:4px;">
                        <span class="badge-icon-status" style="width:20px; height:20px; font-size:11px; background:rgba(5,150,105,0.12); border:1px solid rgba(5,150,105,0.3); flex-shrink:0;" title="Statut GED : Document rattaché">✅</span>
                        <a href="javascript:void(0)" onclick="openCorporateDocumentViewer('${it.ged_document_id}', '${escapeHtml(it.document_name)}', '${it.item_code}')" style="font-size:10px; color:var(--accent-blue); font-weight:700; text-decoration:underline; overflow:hidden; text-overflow:ellipsis; white-space:nowrap; max-width:180px;" title="${escapeHtml(it.document_name || 'Document GED')}">
                            ${escapeHtml(it.document_name || 'Document GED')}
                        </a>
                        <div style="margin-left:auto; display:inline-flex; gap:4px; align-items:center;">
                            <button class="btn-table-icon btn-table-icon-primary" onclick="openCorporateDocumentViewer('${it.ged_document_id}', '${escapeHtml(it.document_name)}', '${it.item_code}')" title="Visionner la pièce jointe">👁️</button>
                            <button class="btn-table-icon btn-table-icon-del" onclick="detachGEDDocumentFromItem('${app.id}', '${it.id}')" title="Détacher la pièce de l'exigence">❌</button>
                        </div>
                    </div>
                `;
            } else {
                gedInfo = `
                    <div style="display:flex; align-items:center; justify-content:space-between; margin-top:4px;">
                        <span style="display:inline-flex; align-items:center; gap:5px; font-size:10px; color:var(--text-dim); font-style:italic;">
                            <span class="badge-icon-status" style="width:20px; height:20px; font-size:11px; background:rgba(220,38,38,0.12); border:1px solid rgba(220,38,38,0.3); flex-shrink:0;" title="Statut GED : Document manquant">⚠️</span>
                            Aucune pièce rattachée
                        </span>
                        <button class="btn-table-icon" onclick="openGEDPickerForCorporateItem('${app.id}', '${it.id}')" title="Rattacher un document du coffre-fort GED">
                            📎
                        </button>
                    </div>
                `;
            }

            html += `
                <div style="padding:8px 10px; border-bottom:1px solid var(--border-color); font-size:11px;">
                    <div style="display:flex; justify-content:space-between; align-items:flex-start; gap:8px;">
                        <div>
                            <span style="font-family:var(--font-mono); font-weight:700; color:var(--accent-blue); margin-right:4px;">${escapeHtml(it.item_code)}</span>
                            <span style="font-weight:600; color:var(--text-main);">${escapeHtml(it.title)}</span>
                        </div>
                        <div>${reqBadge}</div>
                    </div>
                    ${gedInfo}
                </div>
            `;
        });

        html += `
                </div>
            </div>
        `;
    });

    html += `</div>`;
    return html;
}

// ----------------------------------------------------------------------------
// B. Rendu Sous-onglet Formulaire e-Form Réglementaire
// ----------------------------------------------------------------------------
function renderEFormTabHtml(app) {
    const fd = app.form_data || {};
    return `
        <div style="background:var(--bg-card); border:1px solid var(--border-color); border-radius:8px; padding:14px;">
            <div style="font-size:12px; font-weight:700; color:var(--text-main); margin-bottom:12px; border-bottom:1px solid var(--border-color); padding-bottom:6px;">
                📋 Formulaire Déclaratif e-Form (Décret 21-82 / Tabadol)
            </div>

            <div style="display:grid; grid-template-columns:1fr 1fr; gap:10px; font-size:11px;">
                <div>
                    <label style="display:block; color:var(--text-dim); margin-bottom:3px;">Raison Sociale / Dénomination :</label>
                    <input type="text" id="ef-corp-name" class="input-base" style="width:100%; font-size:11px; padding:6px;" value="${escapeHtml(fd.company_name || app.title || '')}">
                </div>
                <div>
                    <label style="display:block; color:var(--text-dim); margin-bottom:3px;">Forme Juridique :</label>
                    <select id="ef-corp-status" class="input-base" style="width:100%; font-size:11px; padding:6px;">
                        <option value="SARL" ${fd.legal_status === 'SARL' ? 'selected' : ''}>SARL</option>
                        <option value="SPA" ${fd.legal_status === 'SPA' ? 'selected' : ''}>SPA</option>
                        <option value="EURL" ${fd.legal_status === 'EURL' ? 'selected' : ''}>EURL</option>
                        <option value="SNC" ${fd.legal_status === 'SNC' ? 'selected' : ''}>SNC</option>
                    </select>
                </div>
                <div>
                    <label style="display:block; color:var(--text-dim); margin-bottom:3px;">Capital Social (DZD) :</label>
                    <input type="number" id="ef-corp-capital" class="input-base" style="width:100%; font-size:11px; padding:6px;" value="${escapeHtml(fd.capital_dzd || '50000000')}">
                </div>
                <div>
                    <label style="display:block; color:var(--text-dim); margin-bottom:3px;">N° Registre de Commerce (RC) :</label>
                    <input type="text" id="ef-corp-rc" class="input-base" style="width:100%; font-size:11px; padding:6px;" value="${escapeHtml(fd.rc_number || '')}">
                </div>
                <div style="grid-column: span 2;">
                    <label style="display:block; color:var(--text-dim); margin-bottom:3px;">Adresse du Siège Social :</label>
                    <input type="text" id="ef-corp-headquarters" class="input-base" style="width:100%; font-size:11px; padding:6px;" value="${escapeHtml(fd.headquarters_address || '')}">
                </div>
                <div style="grid-column: span 2;">
                    <label style="display:block; color:var(--text-dim); margin-bottom:3px;">Adresse du Site Opérationnel / Usine :</label>
                    <input type="text" id="ef-corp-site" class="input-base" style="width:100%; font-size:11px; padding:6px;" value="${escapeHtml(fd.site_address || '')}">
                </div>
                <div>
                    <label style="display:block; color:var(--text-dim); margin-bottom:3px;">Pharmacien Directeur Technique (PDT) :</label>
                    <input type="text" id="ef-corp-pdt" class="input-base" style="width:100%; font-size:11px; padding:6px;" value="${escapeHtml(fd.pdt_name || 'Dr. Pharmacien DT')}">
                </div>
                <div>
                    <label style="display:block; color:var(--text-dim); margin-bottom:3px;">N° Décision Nominative PDT (MIPH) :</label>
                    <input type="text" id="ef-corp-pdt-dec" class="input-base" style="width:100%; font-size:11px; padding:6px;" value="${escapeHtml(fd.pdt_decision_ref || '')}">
                </div>
                <div style="grid-column: span 2;">
                    <label style="display:block; color:var(--text-dim); margin-bottom:3px;">Lignes de Fabrication / Gammes Revendiquées :</label>
                    <textarea id="ef-corp-lines" class="input-base" rows="2" style="width:100%; font-size:11px; padding:6px;">${escapeHtml(fd.manufacturing_lines || 'Formes solides orales (comprimés, gélules), injectables stériles')}</textarea>
                </div>
            </div>

            <div style="margin-top:14px; text-align:right;">
                <button class="btn-pill" style="background:var(--accent-blue); color:white; border:none; padding:6px 14px; font-weight:700;" onclick="saveCorporateEForm('${app.id}')">
                    💾 Sauvegarder les Déclarations
                </button>
            </div>
        </div>
    `;
}

// ----------------------------------------------------------------------------
// C. Rendu Sous-onglet Tabadol & Visa PDT
// ----------------------------------------------------------------------------
function renderTabadolTabHtml(app) {
    const isPdtSigned = !!app.pdt_signed_at;
    const isSubmitted = !!app.tabadol_submitted_at;

    let pdtSealHtml = `
        <div style="background:rgba(239,68,68,0.06); border:1px dashed var(--accent-red); border-radius:6px; padding:10px; margin-bottom:12px;">
            <div style="font-weight:700; color:var(--accent-red); font-size:11px; margin-bottom:4px;">
                ⚠️ Visa du Pharmacien Directeur Technique requis
            </div>
            <div style="font-size:10px; color:var(--text-dim); line-height:1.4; margin-bottom:8px;">
                Conformément à l'Article 18 du Décret exécutif n° 21-82, le dossier d'agrément doit être obligatoirement certifié conforme par le Pharmacien DT avant toute télésoumission sur la plateforme Tabadol.
            </div>
            <button class="btn-pill" style="background:var(--accent-blue); color:white; border:none; font-size:10px; font-weight:700; padding:5px 12px; border-radius:4px;" onclick="openCorporatePDTSignModal('${app.id}')" title="Apposer le visa électronique PDT (21 CFR Part 11)">
                ⚖️ Apposer le Visa Électronique PDT (21 CFR Part 11)
            </button>
        </div>
    `;

    if (isPdtSigned) {
        pdtSealHtml = `
            <div style="background:rgba(16,185,129,0.06); border:1px solid var(--accent-emerald); border-radius:6px; padding:10px; margin-bottom:12px;">
                <div style="display:flex; justify-content:space-between; align-items:center;">
                    <div style="font-weight:700; color:var(--accent-emerald); font-size:11px;">
                        ✅ Dossier Certifié Conforme par le Pharmacien DT
                    </div>
                    <span style="font-size:10px; color:var(--text-dim);">${formatDate(app.pdt_signed_at)}</span>
                </div>
                <div style="font-size:9px; font-family:var(--font-mono); color:var(--text-muted); margin-top:4px; word-break:break-all;">
                    Sceau SHA-256 : ${escapeHtml(app.pdt_signature_seal || 'SHA256-PDT-SEAL-VALIDATED')}
                </div>
            </div>
        `;
    }

    let tabadolActionHtml = '';
    if (!isSubmitted) {
        tabadolActionHtml = `
            <div style="margin-top:12px;">
                <button class="btn-pill" style="background:var(--accent-emerald); color:white; border:none; padding:8px 14px; font-weight:700; width:100%; border-radius:4px;" onclick="openTabadolSubmitModal('${app.id}')" title="Télésoumettre à la Plateforme Tabadol MIPH">
                    🚀 Télésoumettre à la Plateforme Tabadol MIPH
                </button>
            </div>
        `;
    } else {
        tabadolActionHtml = `
            <div style="background:rgba(2,132,199,0.08); border:1px solid var(--accent-blue); border-radius:6px; padding:10px; margin-top:12px;">
                <div style="font-weight:700; color:var(--accent-blue); font-size:11px;">
                    🌐 Dossier Télésoumis avec succès sur Tabadol
                </div>
                <div style="font-size:11px; margin-top:4px;">
                    Réf. Télétransmission : <strong>${escapeHtml(app.tabadol_submission_ref)}</strong>
                </div>
                <div style="font-size:10px; color:var(--text-dim); margin-top:2px;">
                    Date d'enregistrement : ${formatDate(app.tabadol_submitted_at)}
                </div>
            </div>
        `;
    }

    return `
        <div style="background:var(--bg-card); border:1px solid var(--border-color); border-radius:8px; padding:14px;">
            <div style="font-size:12px; font-weight:700; color:var(--text-main); margin-bottom:10px;">
                Passerelle Réglementaire Tabadol (Ministère de l'Industrie Pharmaceutique)
            </div>

            ${pdtSealHtml}

            <div style="display:flex; gap:8px; margin-bottom:12px;">
                <button class="btn-pill" style="flex:1; background:var(--bg-input); border:1px solid var(--border-color); font-size:11px; font-weight:600;" onclick="openTabadolAuditModal('${app.id}')">
                    🔍 Lancer l'Audit Pré-soumission
                </button>
                <button class="btn-pill" style="flex:1; background:var(--bg-input); border:1px solid var(--border-color); font-size:11px; font-weight:600;" onclick="downloadTabadolPackage('${app.id}')">
                    📦 Télécharger Package (.zip)
                </button>
            </div>

            ${tabadolActionHtml}
        </div>
    `;
}

// ----------------------------------------------------------------------------
// D. Rendu Sous-onglet Récépissé & Arrêté MIPH
// ----------------------------------------------------------------------------
function renderMIPHTabHtml(app) {
    return `
        <div style="display:flex; flex-direction:column; gap:12px;">
            <!-- Récépissé Bureau d'Ordre -->
            <div style="background:var(--bg-card); border:1px solid var(--border-color); border-radius:8px; padding:14px;">
                <div style="font-size:12px; font-weight:700; color:var(--text-main); margin-bottom:8px;">
                    📥 Récépissé de Dépôt Physique (Bureau d'Ordre MIPH)
                </div>
                <div style="font-size:10px; color:var(--text-dim); margin-bottom:10px;">
                    L'accusé de réception physique confirme le dépôt officiel du dossier technique relié au Ministère.
                </div>
                <div style="display:grid; grid-template-columns:1fr 1fr; gap:8px; font-size:11px;">
                    <div>
                        <label style="display:block; color:var(--text-dim); margin-bottom:2px;">N° Récépissé BO :</label>
                        <input type="text" id="miph-receipt-ref" class="input-base" style="width:100%; font-size:11px; padding:6px;" value="${escapeHtml(app.miph_receipt_ref || '')}" placeholder="REC-MIPH-BO-2026-XXXX">
                    </div>
                    <div>
                        <label style="display:block; color:var(--text-dim); margin-bottom:2px;">Date de Dépôt :</label>
                        <input type="date" id="miph-receipt-date" class="input-base" style="width:100%; font-size:11px; padding:6px;" value="${escapeHtml(app.miph_receipt_date || '')}">
                    </div>
                </div>
                <div style="margin-top:10px; text-align:right;">
                    <button class="btn-pill" style="background:var(--accent-blue); color:white; border:none; font-size:11px; padding:5px 12px; font-weight:700;" onclick="saveCorporateReceipt('${app.id}')">
                        💾 Enregistrer Récépissé BO
                    </button>
                </div>
            </div>

            <!-- Décision Ministérielle & Arrêté d'Agrément -->
            <div style="background:var(--bg-card); border:1px solid var(--border-color); border-radius:8px; padding:14px;">
                <div style="font-size:12px; font-weight:700; color:var(--text-main); margin-bottom:8px;">
                    📜 Décision Ministérielle &amp; Arrêté d'Agrément (Décret 21-82)
                </div>
                <div style="font-size:10px; color:var(--text-dim); margin-bottom:10px;">
                    À l'issue de l'inspection de conformité, enregistrez l'arrêté ministériel pour activer l'agrément et inscrire le titre dans le registre avec compte à rebours quinquennal.
                </div>
                <div style="display:grid; grid-template-columns:1fr 1fr; gap:8px; font-size:11px;">
                    <div>
                        <label style="display:block; color:var(--text-dim); margin-bottom:2px;">N° Arrêté d'Agrément MIPH :</label>
                        <input type="text" id="miph-decision-ref" class="input-base" style="width:100%; font-size:11px; padding:6px;" value="${escapeHtml(app.decision_ref || '')}" placeholder="ARR-MIPH-AGR-2026-XXX">
                    </div>
                    <div>
                        <label style="display:block; color:var(--text-dim); margin-bottom:2px;">Date d'Octroi Ministériel :</label>
                        <input type="date" id="miph-decision-date" class="input-base" style="width:100%; font-size:11px; padding:6px;" value="${escapeHtml(app.decision_date || '')}">
                    </div>
                </div>
                <div style="margin-top:10px; text-align:right;">
                    <button class="btn-pill" style="background:var(--accent-emerald); color:white; border:none; font-size:11px; padding:5px 12px; font-weight:700;" onclick="saveCorporateLicense('${app.id}')">
                        🏆 Délivrer &amp; Inscrire l'Agrément au Registre
                    </button>
                </div>
            </div>
        </div>
    `;
}

// ============================================================================
// 6. ACTIONS MÉTIER & APPELS API
// ============================================================================

// A. Sauvegarde du formulaire e-Form
async function saveCorporateEForm(appId) {
    const payload = {
        title: document.getElementById("ef-corp-name") ? document.getElementById("ef-corp-name").value : "",
        form_data: {
            company_name: document.getElementById("ef-corp-name") ? document.getElementById("ef-corp-name").value : "",
            legal_status: document.getElementById("ef-corp-status") ? document.getElementById("ef-corp-status").value : "SARL",
            capital_dzd: document.getElementById("ef-corp-capital") ? document.getElementById("ef-corp-capital").value : "0",
            rc_number: document.getElementById("ef-corp-rc") ? document.getElementById("ef-corp-rc").value : "",
            headquarters_address: document.getElementById("ef-corp-headquarters") ? document.getElementById("ef-corp-headquarters").value : "",
            site_address: document.getElementById("ef-corp-site") ? document.getElementById("ef-corp-site").value : "",
            pdt_name: document.getElementById("ef-corp-pdt") ? document.getElementById("ef-corp-pdt").value : "",
            pdt_decision_ref: document.getElementById("ef-corp-pdt-dec") ? document.getElementById("ef-corp-pdt-dec").value : "",
            manufacturing_lines: document.getElementById("ef-corp-lines") ? document.getElementById("ef-corp-lines").value : ""
        }
    };

    try {
        const resp = await fetch('/api/v1/corporate/applications/' + appId, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });
        if (!resp.ok) {
            const err = await resp.json();
            throw new Error(err.message || "Échec de sauvegarde");
        }
        showNotifyModal("Succès", "Déclarations e-Form enregistrées avec succès.", "✅", "Fermer");
        renderCorporateAppPreview(appId);
    } catch (e) {
        showNotifyModal("Erreur", e.message, "❌", "Fermer");
    }
}

// B. Enregistrement Récépissé Bureau d'Ordre
async function saveCorporateReceipt(appId) {
    const ref = document.getElementById("miph-receipt-ref") ? document.getElementById("miph-receipt-ref").value.trim() : "";
    const date = document.getElementById("miph-receipt-date") ? document.getElementById("miph-receipt-date").value : "";

    if (!ref || !date) {
        showNotifyModal("Champ requis", "Veuillez renseigner le numéro et la date du récépissé.", "⚠️", "Fermer");
        return;
    }

    try {
        const resp = await fetch(`/api/v1/corporate/applications/${appId}/miph-receipt`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ miph_receipt_ref: ref, miph_receipt_date: date })
        });
        if (!resp.ok) {
            const err = await resp.json();
            throw new Error(err.message || "Échec d'enregistrement");
        }
        showNotifyModal("Récépissé Enregistré", "Le récépissé du Bureau d'Ordre a été enregistré.", "✅", "Fermer");
        renderCorporateView();
    } catch (e) {
        showNotifyModal("Erreur", e.message, "❌", "Fermer");
    }
}

// C. Octroi de l'Arrêté Ministériel d'Agrément
async function saveCorporateLicense(appId) {
    const ref = document.getElementById("miph-decision-ref") ? document.getElementById("miph-decision-ref").value.trim() : "";
    const date = document.getElementById("miph-decision-date") ? document.getElementById("miph-decision-date").value : "";

    if (!ref || !date) {
        showNotifyModal("Champ requis", "Veuillez renseigner le numéro d'arrêté et la date d'octroi ministériel.", "⚠️", "Fermer");
        return;
    }

    try {
        const resp = await fetch(`/api/v1/corporate/applications/${appId}/grant-license`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ decision_ref: ref, grant_date: date, validity_years: 5 })
        });
        if (!resp.ok) {
            const err = await resp.json();
            throw new Error(err.message || "Échec de l'octroi");
        }
        showNotifyModal("Agrément Octroyé !", "L'agrément a été délivré avec succès et inscrit dans le registre avec compte à rebours de validité quinquennale (Décret 21-82).", "🏆", "Fermer");
        selectCorporateFilter('LICENSES');
    } catch (e) {
        showNotifyModal("Erreur", e.message, "❌", "Fermer");
    }
}

// D. Téléchargement de l'archive Tabadol (.zip)
function downloadTabadolPackage(appId) {
    window.open(`/api/v1/corporate/applications/${appId}/tabadol/package`, '_blank');
}

// E. Détachement d'un document GED
async function detachGEDDocumentFromItem(appId, itemId) {
    if (!confirm("Détacher cette pièce GED de l'exigence réglementaire ?")) return;
    try {
        const resp = await fetch(`/api/v1/corporate/applications/${appId}/items/${itemId}/detach`, {
            method: 'POST'
        });
        if (!resp.ok) throw new Error("Échec du détachement");
        viewCorporateChecklist(appId);
    } catch (e) {
        showNotifyModal("Erreur", e.message, "❌", "Fermer");
    }
}

// F. Visualisation de la pièce jointe dans la modale redimensionnable
function openCorporateDocumentViewer(gedDocId, docTitle, itemCode) {
    if (!gedDocId) {
        showNotifyModal("Pièce Non Rattachée", "Aucun document GED n'est rattaché à cette exigence réglementaire.", "⚠️", "Fermer");
        return;
    }
    const fullTitle = (itemCode ? `[${itemCode}] ` : '') + (docTitle || "Pièce Jointe Réglementaire (Décret 21-82)");
    if (typeof openDocumentViewerModal === 'function') {
        openDocumentViewerModal(gedDocId, fullTitle);
    } else {
        window.open(`/api/v1/ged/documents/${gedDocId}/download`, '_blank');
    }
}

// ============================================================================
// 7. MODALS CORPORATE : AUDIT TABADOL, SIGNATURE PDT, SOUISSION & CRÉATION
// ============================================================================

// Modal Audit de Conformité Tabadol
async function openTabadolAuditModal(appId) {
    let modal = document.getElementById("modal-corporate-audit");
    if (!modal) {
        modal = document.createElement("div");
        modal.id = "modal-corporate-audit";
        modal.className = "modal-overlay";
        document.body.appendChild(modal);
    }

    modal.innerHTML = `
        <div class="modal" style="max-width:640px;">
            <div class="modal-header">
                <span class="modal-title">🔍 Audit de Conformité Tabadol (Décret 21-82)</span>
                <button class="modal-close" onclick="closeCorporateModal('modal-corporate-audit')">✕</button>
            </div>
            <div class="modal-body" id="corp-audit-body">
                <div style="text-align:center; padding:30px; color:var(--text-muted);"><span class="spin">⏳</span> Analyse des pièces et contrôle des critères éliminatoires...</div>
            </div>
            <div class="modal-footer">
                <button class="btn-pill" onclick="closeCorporateModal('modal-corporate-audit')">Fermer</button>
            </div>
        </div>
    `;

    modal.style.display = "flex";
    modal.classList.add("open");

    try {
        const resp = await fetch(`/api/v1/corporate/applications/${appId}/tabadol/audit`);
        if (!resp.ok) throw new Error("Erreur HTTP " + resp.status);
        const data = await resp.json();
        const rep = data.audit || data.report || {};
        const b = document.getElementById("corp-audit-body");

        const compPct = rep.completeness_percent != null ? rep.completeness_percent : Math.round((rep.completeness_ratio || 0) * 100);
        const missingCount = Array.isArray(rep.missing_mandatory) ? rep.missing_mandatory.length : (rep.missing_mandatory || 0);

        let statusBanner = '';
        if (rep.can_submit) {
            statusBanner = `
                <div style="background:rgba(16,185,129,0.1); border:1px solid var(--accent-emerald); border-radius:6px; padding:12px; margin-bottom:14px; text-align:center;">
                    <div style="font-size:24px; margin-bottom:4px;">🟢</div>
                    <div style="font-weight:700; color:var(--accent-emerald); font-size:13px;">Dossier Conforme pour Télésoumission Tabadol</div>
                    <div style="font-size:11px; color:var(--text-main); margin-top:4px;">Toutes les pièces éliminatoires [E] et obligatoires sont rattachées.</div>
                </div>
            `;
        } else {
            statusBanner = `
                <div style="background:rgba(239,68,68,0.1); border:1px solid var(--accent-red); border-radius:6px; padding:12px; margin-bottom:14px; text-align:center;">
                    <div style="font-size:24px; margin-bottom:4px;">🔴</div>
                    <div style="font-weight:700; color:var(--accent-red); font-size:13px;">Soumission Tabadol Bloquée</div>
                    <div style="font-size:11px; color:var(--text-main); margin-top:4px;">Des pièces éliminatoires [E] ou obligatoires sont manquantes ou non visées.</div>
                </div>
            `;
        }

        let blockingHtml = '';
        if (rep.blocking_errors && rep.blocking_errors.length > 0) {
            blockingHtml = `
                <div style="margin-top:12px;">
                    <div style="font-weight:700; font-size:11px; color:var(--accent-red); margin-bottom:6px;">
                        Motifs de Blocage Tabadol (${rep.blocking_errors.length}) :
                    </div>
                    <div style="display:flex; flex-direction:column; gap:4px; margin-bottom:10px;">
                        ${rep.blocking_errors.map(err => `
                            <div style="background:rgba(239,68,68,0.06); border-left:3px solid var(--accent-red); padding:6px 10px; font-size:11px; color:var(--text-main);">
                                ⚠️ ${escapeHtml(err)}
                            </div>
                        `).join('')}
                    </div>
                </div>
            `;
        }

        let missingHtml = '';
        const missingList = Array.isArray(rep.missing_mandatory) ? rep.missing_mandatory : (rep.missing_mandatory_items || []);
        if (missingList && missingList.length > 0) {
            missingHtml = `
                <div style="margin-top:8px;">
                    <div style="font-weight:700; font-size:11px; color:var(--accent-red); margin-bottom:6px;">
                        Exigences Obligatoires [E] Non Satisfaites :
                    </div>
                    <div style="display:flex; flex-direction:column; gap:4px; max-height:160px; overflow-y:auto;">
                        ${missingList.map(m => {
                            const txt = typeof m === 'string' ? m : `${m.item_code} : ${m.title}`;
                            return `
                                <div style="background:rgba(239,68,68,0.05); border:1px solid rgba(239,68,68,0.2); border-radius:4px; padding:6px 8px; font-size:10px; display:flex; justify-content:space-between;">
                                    <span>${escapeHtml(txt)}</span>
                                    <span class="status-tag status-req" style="font-size:8px;">[E] REQUIS</span>
                                </div>
                            `;
                        }).join('')}
                    </div>
                </div>
            `;
        }

        b.innerHTML = `
            ${statusBanner}

            <div style="display:grid; grid-template-columns: repeat(4, 1fr); gap:8px; text-align:center; margin-bottom:12px;">
                <div style="background:var(--bg-input); padding:8px; border-radius:6px;">
                    <div style="font-size:18px; font-weight:700; color:var(--text-main);">${rep.total_items || 0}</div>
                    <div style="font-size:10px; color:var(--text-dim);">Total Exigences</div>
                </div>
                <div style="background:var(--bg-input); padding:8px; border-radius:6px;">
                    <div style="font-size:18px; font-weight:700; color:var(--accent-emerald);">${rep.attached_items || 0}</div>
                    <div style="font-size:10px; color:var(--text-dim);">Pièces Rattachées</div>
                </div>
                <div style="background:var(--bg-input); padding:8px; border-radius:6px;">
                    <div style="font-size:18px; font-weight:700; color:var(--accent-red);">${missingCount}</div>
                    <div style="font-size:10px; color:var(--text-dim);">Critiques [E]</div>
                </div>
                <div style="background:var(--bg-input); padding:8px; border-radius:6px;">
                    <div style="font-size:18px; font-weight:700; color:var(--accent-blue);">${compPct}%</div>
                    <div style="font-size:10px; color:var(--text-dim);">Complétude</div>
                </div>
            </div>

            ${blockingHtml}
            ${missingHtml}
        `;

    } catch (e) {
        document.getElementById("corp-audit-body").innerHTML = `<div style="text-align:center; padding:20px; color:var(--accent-red);">❌ Erreur: ${escapeHtml(e.message)}</div>`;
    }
}

// Modal Visa PDT
function openCorporatePDTSignModal(appId) {
    let modal = document.getElementById("modal-corporate-pdt-sign");
    if (!modal) {
        modal = document.createElement("div");
        modal.id = "modal-corporate-pdt-sign";
        modal.className = "modal-overlay";
        document.body.appendChild(modal);
    }

    modal.innerHTML = `
        <div class="modal" style="max-width:520px;">
            <div class="modal-header">
                <span class="modal-title">⚖️ Certification Pharmacien Directeur Technique</span>
                <button class="modal-close" onclick="closeCorporateModal('modal-corporate-pdt-sign')">✕</button>
            </div>
            <div class="modal-body">
                <div style="font-size:11px; color:var(--text-main); line-height:1.5; margin-bottom:12px; background:rgba(2,132,199,0.06); border:1px solid var(--accent-blue); border-radius:6px; padding:10px;">
                    <strong>Déclaration solennelle d'engagement réglementaire (21 CFR Part 11) :</strong><br>
                    En qualité de Pharmacien Directeur Technique habilité de l'établissement, j'atteste sur l'honneur l'authenticité, l'intégrité et la conformité des pièces constitutives du présent dossier au regard du Décret exécutif n° 21-82 et des Bonnes Pratiques de Fabrication / Distribution pharmaceutiques.
                </div>
                <div>
                    <label style="display:block; font-size:11px; color:var(--text-dim); margin-bottom:4px;">Mention manuscrite ou observation du PDT :</label>
                    <input type="text" id="corp-pdt-statement" class="input-base" style="width:100%; font-size:11px; padding:6px;" value="Dossier vérifié et certifié conforme pour dépôt MIPH Tabadol.">
                </div>
            </div>
            <div class="modal-footer">
                <button class="btn-pill" onclick="closeCorporateModal('modal-corporate-pdt-sign')">Annuler</button>
                <button class="btn-pill" style="background:var(--accent-blue); color:white; border:none; font-weight:700;" onclick="submitCorporatePDTSign('${appId}')">
                    ⚖️ Signer &amp; Sceller le Dossier
                </button>
            </div>
        </div>
    `;

    modal.style.display = "flex";
    modal.classList.add("open");
}

async function submitCorporatePDTSign(appId) {
    const stmt = document.getElementById("corp-pdt-statement") ? document.getElementById("corp-pdt-statement").value : "";
    try {
        const resp = await fetch(`/api/v1/corporate/applications/${appId}/pdt-sign`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ signature_statement: stmt })
        });
        if (!resp.ok) {
            const err = await resp.json();
            throw new Error(err.message || "Échec de signature PDT");
        }
        closeCorporateModal('modal-corporate-pdt-sign');
        showNotifyModal("Visa PDT Apposé", "Le dossier a été certifié conforme et scellé cryptographiquement par le Pharmacien DT.", "✅", "Fermer");
        renderCorporateAppPreview(appId);
        renderCorporateView();
    } catch (e) {
        showNotifyModal("Erreur", e.message, "❌", "Fermer");
    }
}

// Modal Télésoumission Tabadol
function openTabadolSubmitModal(appId) {
    let modal = document.getElementById("modal-corporate-tabadol-submit");
    if (!modal) {
        modal = document.createElement("div");
        modal.id = "modal-corporate-tabadol-submit";
        modal.className = "modal-overlay";
        document.body.appendChild(modal);
    }

    const defaultRef = "TBD-" + new Date().getFullYear() + "-MIPH-" + Math.floor(10000 + Math.random() * 90000);

    modal.innerHTML = `
        <div class="modal" style="max-width:520px;">
            <div class="modal-header">
                <span class="modal-title">🌐 Télésoumission Passerelle Tabadol MIPH</span>
                <button class="modal-close" onclick="closeCorporateModal('modal-corporate-tabadol-submit')">✕</button>
            </div>
            <div class="modal-body">
                <div style="font-size:11px; color:var(--text-dim); margin-bottom:12px;">
                    Cette action transmet le dossier scellé à la passerelle ministérielle Tabadol et génère l'accusé de télétransmission officiel.
                </div>
                <div style="display:flex; flex-direction:column; gap:8px;">
                    <div>
                        <label style="display:block; font-size:11px; color:var(--text-dim); margin-bottom:3px;">Numéro d'enregistrement Tabadol :</label>
                        <input type="text" id="corp-tbd-ref" class="input-base" style="width:100%; font-size:11px; padding:6px;" value="${defaultRef}">
                    </div>
                    <div>
                        <label style="display:block; font-size:11px; color:var(--text-dim); margin-bottom:3px;">Notes de transmission ministérielle :</label>
                        <textarea id="corp-tbd-notes" class="input-base" rows="2" style="width:100%; font-size:11px; padding:6px;">Dépôt initial complet pour agrément établissement pharmaceutique.</textarea>
                    </div>
                </div>
            </div>
            <div class="modal-footer">
                <button class="btn-pill" onclick="closeCorporateModal('modal-corporate-tabadol-submit')">Annuler</button>
                <button class="btn-pill" style="background:var(--accent-emerald); color:white; border:none; font-weight:700;" onclick="submitCorporateTabadol('${appId}')">
                    🚀 Confirmer la Télésoumission
                </button>
            </div>
        </div>
    `;

    modal.style.display = "flex";
    modal.classList.add("open");
}

async function submitCorporateTabadol(appId) {
    const ref = document.getElementById("corp-tbd-ref") ? document.getElementById("corp-tbd-ref").value.trim() : "";
    const notes = document.getElementById("corp-tbd-notes") ? document.getElementById("corp-tbd-notes").value : "";

    try {
        const resp = await fetch(`/api/v1/corporate/applications/${appId}/tabadol/submit`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ tabadol_submission_ref: ref, submission_notes: notes })
        });
        if (!resp.ok) {
            const err = await resp.json();
            throw new Error(err.message || "Échec de télésoumission Tabadol");
        }
        closeCorporateModal('modal-corporate-tabadol-submit');
        showNotifyModal("Télésoumission Réussie", `Le dossier d'agrément a été transmis avec succès sur Tabadol (Réf : ${ref}).`, "✅", "Fermer");
        renderCorporateAppPreview(appId);
        renderCorporateView();
    } catch (e) {
        showNotifyModal("Erreur", e.message, "❌", "Fermer");
    }
}

// Modal Rattachement Document GED
async function openGEDPickerForCorporateItem(appId, itemId) {
    let modal = document.getElementById("modal-corporate-ged-picker");
    if (!modal) {
        modal = document.createElement("div");
        modal.id = "modal-corporate-ged-picker";
        modal.className = "modal-overlay";
        document.body.appendChild(modal);
    }

    modal.innerHTML = `
        <div class="modal" style="max-width:680px;">
            <div class="modal-header">
                <span class="modal-title">📎 Rattacher un Document GED Pharmaceutique</span>
                <button class="modal-close" onclick="closeCorporateModal('modal-corporate-ged-picker')">✕</button>
            </div>
            <div class="modal-body" id="corp-ged-picker-body">
                <div style="text-align:center; padding:30px; color:var(--text-muted);"><span class="spin">⏳</span> Chargement des documents approuvés du coffre-fort...</div>
            </div>
            <div class="modal-footer">
                <button class="btn-pill" onclick="closeCorporateModal('modal-corporate-ged-picker')">Annuler</button>
            </div>
        </div>
    `;

    modal.style.display = "flex";
    modal.classList.add("open");

    try {
        const resp = await fetch('/api/v1/ged/documents');
        if (!resp.ok) throw new Error("Erreur HTTP " + resp.status);
        const data = await resp.json();
        const docs = data.documents || [];
        const body = document.getElementById("corp-ged-picker-body");

        if (docs.length === 0) {
            body.innerHTML = `
                <div style="text-align:center; padding:30px; color:var(--text-muted);">
                    Aucun document GED disponible dans votre coffre-fort.<br>
                    Veuillez d'abord déposer et valider les pièces dans le module GED BPF.
                </div>
            `;
            return;
        }

        body.innerHTML = `
            <div style="font-size:11px; color:var(--text-dim); margin-bottom:8px;">
                Sélectionnez une pièce vérifiée pour satisfaire l'exigence réglementaire :
            </div>
            <div style="display:flex; flex-direction:column; gap:6px; max-height:300px; overflow-y:auto;">
                ${docs.map(d => `
                    <div style="display:flex; justify-content:space-between; align-items:center; padding:8px 10px; background:var(--bg-input); border-radius:6px; border:1px solid var(--border-color);">
                        <div>
                            <div style="font-weight:700; font-size:11px; color:var(--text-main);">${escapeHtml(d.title_fr || d.title || 'Document')}</div>
                            <div style="font-size:9px; color:var(--text-dim); font-family:var(--font-mono);">
                                Code: ${escapeHtml(d.document_code || '-')} • Ver: ${escapeHtml(d.current_version || 'v1.0')} • Statut: ${escapeHtml(d.status || 'approuvé')}
                            </div>
                        </div>
                        <button class="btn-pill" style="font-size:10px; padding:4px 10px; border:none; background:var(--accent-blue); color:white; font-weight:700;" onclick="attachGEDToCorporateItem('${appId}', '${itemId}', '${d.id}')">
                            Sélectionner
                        </button>
                    </div>
                `).join('')}
            </div>
        `;
    } catch (e) {
        document.getElementById("corp-ged-picker-body").innerHTML = `<div style="text-align:center; padding:20px; color:var(--accent-red);">❌ Erreur: ${escapeHtml(e.message)}</div>`;
    }
}

async function attachGEDToCorporateItem(appId, itemId, gedDocId) {
    try {
        const resp = await fetch(`/api/v1/corporate/applications/${appId}/items/${itemId}/attach`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ ged_document_id: gedDocId })
        });
        if (!resp.ok) {
            const err = await resp.json();
            throw new Error(err.message || "Échec de rattachement");
        }
        closeCorporateModal('modal-corporate-ged-picker');
        showNotifyModal("Pièce Rattachée", "Le document GED a été associé à l'exigence réglementaire avec succès.", "✅", "Fermer");
        viewCorporateChecklist(appId);
    } catch (e) {
        showNotifyModal("Erreur", e.message, "❌", "Fermer");
    }
}

// Modal Nouvelle Demande d'Agrément
async function openNewCorporateAppModal() {
    let modal = document.getElementById("modal-corporate-create");
    if (!modal) {
        modal = document.createElement("div");
        modal.id = "modal-corporate-create";
        modal.className = "modal-overlay";
        document.body.appendChild(modal);
    }

    // Charger les établissements existants
    let estOptions = `<option value="">-- Sélectionner un établissement existant --</option>`;
    try {
        const resp = await fetch('/api/v1/establishments');
        if (resp.ok) {
            const data = await resp.json();
            const ests = data.establishments || [];
            ests.forEach(e => {
                estOptions += `<option value="${e.id}">${escapeHtml(e.name_fr)} (${escapeHtml(e.miph_license_number || 'En cours')})</option>`;
            });
        }
    } catch (_) {}

    modal.innerHTML = `
        <div class="modal" style="max-width:580px;">
            <div class="modal-header">
                <span class="modal-title">🏛️ Nouvelle Demande d'Agrément MIPH (Décret 21-82)</span>
                <button class="modal-close" onclick="closeCorporateModal('modal-corporate-create')">✕</button>
            </div>
            <div class="modal-body">
                <div style="display:flex; flex-direction:column; gap:10px;">
                    <div>
                        <label style="display:block; font-size:11px; color:var(--text-dim); margin-bottom:3px;">Catégorie d'Établissement (Décret 21-82 Art. 4) :</label>
                        <select id="new-corp-category" class="input-base" style="width:100%; font-size:11px; padding:6px;">
                            <option value="FABRICATION">🏭 Fabrication de Produits Pharmaceutiques &amp; DM</option>
                            <option value="IMPORTATION">🚢 Importation de Produits Pharmaceutiques &amp; DM</option>
                            <option value="DISTRIBUTION">🚚 Distribution en Gros de Produits Pharmaceutiques</option>
                            <option value="EXPLOITATION">📑 Exploitation de Spécialités Pharmaceutiques</option>
                        </select>
                    </div>

                    <div>
                        <label style="display:block; font-size:11px; color:var(--text-dim); margin-bottom:3px;">Nature de la Demande :</label>
                        <select id="new-corp-nature" class="input-base" style="width:100%; font-size:11px; padding:6px;">
                            <option value="AGREMENT_INITIAL">Agrément Initial (Création)</option>
                            <option value="RENOUVELLEMENT">Renouvellement Quinquennal (J-180)</option>
                            <option value="MODIFICATION_SUBSTANTIELLE">Modification Substantielle (Locaux / Équipements)</option>
                            <option value="EXTENSION_LIGNES">Extension de Gammes / Lignes de Production</option>
                            <option value="TRANSFERT_SITE">Transfert de Site / Déménagement</option>
                        </select>
                    </div>

                    <div>
                        <label style="display:block; font-size:11px; color:var(--text-dim); margin-bottom:3px;">Dénomination &amp; Objet de la Demande :</label>
                        <input type="text" id="new-corp-title" class="input-base" style="width:100%; font-size:11px; padding:6px;" placeholder="ex: Agrément Initial Unité de Fabrication Formes Sèches SARL PHARMA">
                    </div>

                    <div>
                        <label style="display:block; font-size:11px; color:var(--text-dim); margin-bottom:3px;">Rattacher à un Établissement Référentiel (Optionnel) :</label>
                        <select id="new-corp-est" class="input-base" style="width:100%; font-size:11px; padding:6px;">
                            ${estOptions}
                        </select>
                    </div>

                    <div>
                        <label style="display:block; font-size:11px; color:var(--text-dim); margin-bottom:3px;">Référence Interne Dossier :</label>
                        <input type="text" id="new-corp-ref" class="input-base" style="width:100%; font-size:11px; padding:6px;" placeholder="ex: CORP-2026-FAB-001">
                    </div>
                </div>
            </div>
            <div class="modal-footer">
                <button class="btn-pill" onclick="closeCorporateModal('modal-corporate-create')">Annuler</button>
                <button class="btn-pill" style="background:var(--accent-blue); color:white; border:none; font-weight:700;" onclick="submitNewCorporateApp()">
                    ➕ Créer la Demande &amp; Générer la Checklist
                </button>
            </div>
        </div>
    `;

    modal.style.display = "flex";
    modal.classList.add("open");
}

async function submitNewCorporateApp() {
    const cat = document.getElementById("new-corp-category").value;
    const nature = document.getElementById("new-corp-nature").value;
    const title = document.getElementById("new-corp-title").value.trim();
    const estVal = document.getElementById("new-corp-est").value;
    const ref = document.getElementById("new-corp-ref").value.trim();

    if (!title) {
        showNotifyModal("Champ requis", "Veuillez renseigner le titre de la demande.", "⚠️", "Fermer");
        return;
    }

    const payload = {
        category: cat,
        nature: nature,
        title: title,
        internal_reference: ref,
        establishment_id: estVal ? estVal : null,
        form_data: {
            company_name: title,
            declared_at: new Date().toISOString()
        }
    };

    try {
        const resp = await fetch('/api/v1/corporate/applications', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });
        if (!resp.ok) {
            const err = await resp.json();
            throw new Error(err.message || "Échec de création");
        }
        const data = await resp.json();
        closeCorporateModal('modal-corporate-create');
        showNotifyModal("Demande Créée", `La demande d'agrément a été initialisée avec sa checklist réglementaire Décret 21-82 (${(data.application.checklist_items || []).length} exigences).`, "✅", "Fermer");
        selectedCorporateAppID = data.application.id;
        viewCorporateChecklist(data.application.id);
    } catch (e) {
        showNotifyModal("Erreur", e.message, "❌", "Fermer");
    }
}

function closeCorporateModal(id) {
    const m = document.getElementById(id);
    if (m) {
        m.style.display = "none";
        m.classList.remove("open");
    }
}

// Utilitaires locaux
function escapeHtml(str) {
    if (!str) return '';
    return String(str)
        .replace(/&/g, "&amp;")
        .replace(/</g, "&lt;")
        .replace(/>/g, "&gt;")
        .replace(/"/g, "&quot;")
        .replace(/'/g, "&#039;");
}

function formatDate(dateStr) {
    if (!dateStr) return '-';
    try {
        const d = new Date(dateStr);
        if (isNaN(d.getTime())) return dateStr;
        return d.toLocaleDateString('fr-FR', { year: 'numeric', month: '2-digit', day: '2-digit' });
    } catch (_) {
        return dateStr;
    }
}
