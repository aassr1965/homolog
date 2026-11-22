// ============================================================================
// MODULE DISPOSITIFS MÉDICAUX (DDM) — JALON 2.6
// Homologation Approfondie 6 Sections, Recevabilité 13 Points & Régularisation
// Conforme au Décret exécutif n° 20-324 du 22/11/2020 et Arrêté du 10/05/2021
// ============================================================================

let currentDMTab = 'homologation'; // 'homologation' ou 'regularization'
let dmDevicesList = [];
let dmRegularizationsList = [];
let dmRegularizationStats = null;
let dmFilterClass = 'all';
let dmRegFilterStatus = 'all';
let activeDMForEdit = null;
let activeDMEval = null;

// Initialisation du tableau de bord DM
async function initDMDashboard() {
    updateDMToolbar();
    if (currentDMTab === 'homologation') {
        await loadDMDevices();
    } else {
        await loadDMRegularizations();
    }
}

// Mise à jour de la barre d'outils contextuelle
function updateDMToolbar() {
    const tb = document.getElementById("main-toolbar") || document.getElementById("toolbar-actions");
    if (!tb) return;

    if (currentDMTab === 'homologation') {
        tb.innerHTML = `
            <div style="display:flex; align-items:center; gap:8px;">
                <button class="btn-pill active" onclick="switchDMTab('homologation')">🩺 Portefeuille & 6 Sections</button>
                <button class="btn-pill" onclick="switchDMTab('regularization')">⏱️ Mesures Transitoires & Régularisation</button>
            </div>
            <div style="margin-left:auto; display:flex; gap:8px; align-items:center;">
                <button class="btn-pill" data-perm="REF_EDIT" style="background:var(--accent-blue); color:white; border:none;" onclick="openNewDMModal()">➕ Nouveau DM</button>
                <button class="btn-pill" onclick="loadDMDevices()">🔄 Actualiser</button>
                <button class="btn-pill" onclick="togglePreviewPane()">👁️ Volet Aperçu (Alt+P)</button>
            </div>
        `;
    } else {
        tb.innerHTML = `
            <div style="display:flex; align-items:center; gap:8px;">
                <button class="btn-pill" onclick="switchDMTab('homologation')">🩺 Portefeuille & 6 Sections</button>
                <button class="btn-pill active" onclick="switchDMTab('regularization')">⏱️ Mesures Transitoires & Régularisation</button>
            </div>
            <div style="margin-left:auto; display:flex; gap:8px; align-items:center;">
                <button class="btn-pill" data-perm="REF_EDIT" style="background:#059669; color:white; border:none;" onclick="openNewDMRegularizationModal()">➕ Nouveau Suivi Régularisation</button>
                <button class="btn-pill" style="background:#0284c7; color:white; border:none;" onclick="exportDMCanevasXLSX()">📊 Exporter Canevas Dérogatoire (.xlsx)</button>
                <button class="btn-pill" onclick="loadDMRegularizations()">🔄 Actualiser</button>
            </div>
        `;
    }
}

function switchDMTab(tab) {
    currentDMTab = tab;
    updateDMToolbar();
    if (tab === 'homologation') {
        loadDMDevices();
    } else {
        loadDMRegularizations();
    }
}

// ============================================================================
// 1. VOLET A : PORTEFEUILLE DM & HOMOLOGATION 6 SECTIONS
// ============================================================================

async function loadDMDevices() {
    const container = document.getElementById("table-container") || document.querySelector(".table-container");
    if (!container) return;

    container.innerHTML = `
        <div style="padding:40px; text-align:center; color:var(--text-muted);">
            <div class="spinner" style="margin-bottom:12px;"></div>
            Chargement des dispositifs médicaux et dossiers d'homologation DDM...
        </div>
    `;

    try {
        const resp = await fetch('/api/v1/medical-devices');
        if (!resp.ok) throw new Error("Erreur HTTP " + resp.status);
        const data = await resp.json();
        dmDevicesList = data.medical_devices || data.devices || [];
        renderDMDevicesTable();
    } catch (err) {
        container.innerHTML = `
            <div style="padding:20px; color:var(--accent-red); background:rgba(239,68,68,0.1); border-radius:6px; margin:16px;">
                ⚠️ Impossible de charger les dispositifs médicaux : ${err.message}
            </div>
        `;
    }
}

function renderDMDevicesTable() {
    const container = document.getElementById("table-container") || document.querySelector(".table-container");
    if (!container) return;

    let filtered = dmDevicesList;
    if (dmFilterClass !== 'all') {
        filtered = dmDevicesList.filter(d => (d.dm_class || '').toLowerCase() === dmFilterClass.toLowerCase());
    }

    let html = `
        <div style="padding:16px 20px; border-bottom:1px solid var(--border-color); background:var(--bg-surface); display:flex; justify-content:space-between; align-items:center;">
            <div>
                <h2 style="font-size:14px; font-weight:800; color:var(--text-main); margin:0; display:flex; align-items:center; gap:8px;">
                    <span>🩺</span> Direction des Dispositifs Médicaux (DDM) — Homologation Approfondie
                    <span class="status-tag" style="background:#059669; color:white; font-size:10px;">Décret exécutif 20-324 & Arrêté 10 mai 2021</span>
                </h2>
                <div style="font-size:11px; color:var(--text-muted); margin-top:3px;">
                    Formulaire 6 sections (ES-REG-02-DM) & Grille de recevabilité en 13 points (ES-REG-03-DM)
                </div>
            </div>
            <div style="display:flex; gap:6px;">
                <button class="btn-pill ${dmFilterClass === 'all' ? 'active' : ''}" onclick="filterDMClass('all')">Tous (${dmDevicesList.length})</button>
                <button class="btn-pill ${dmFilterClass === 'I' ? 'active' : ''}" onclick="filterDMClass('I')">Classe I</button>
                <button class="btn-pill ${dmFilterClass === 'IIa' ? 'active' : ''}" onclick="filterDMClass('IIa')">Classe IIa</button>
                <button class="btn-pill ${dmFilterClass === 'IIb' ? 'active' : ''}" onclick="filterDMClass('IIb')">Classe IIb</button>
                <button class="btn-pill ${dmFilterClass === 'III' ? 'active' : ''}" onclick="filterDMClass('III')">Classe III</button>
            </div>
        </div>

        <div style="overflow-x:auto; padding:12px 16px;">
            <table class="dense-table" style="width:100%; font-size:11.5px;">
                <thead>
                    <tr>
                        <th style="width:25%;">Dispositif Médical</th>
                        <th style="width:12%;">Réf. Fabricant</th>
                        <th style="width:10%;">Classe</th>
                        <th style="width:12%;">Régime & Origine</th>
                        <th style="width:12%;">Stérilité</th>
                        <th style="width:11%;">Statut</th>
                        <th style="width:18%; text-align:right;">Actions Réglementaires</th>
                    </tr>
                </thead>
                <tbody>
    `;

    if (filtered.length === 0) {
        html += `
            <tr>
                <td colspan="7" style="text-align:center; padding:30px; color:var(--text-muted);">
                    Aucun dispositif médical trouvé pour ce filtre.
                </td>
            </tr>
        `;
    } else {
        filtered.forEach(d => {
            const isSterile = d.is_sterile ? `<span style="color:#059669; font-weight:700;">✓ Stérile</span> (${d.sterilization_method || 'ETO'})` : `<span style="color:var(--text-muted);">Non stérile</span>`;
            const isLocal = (d.origin_type || '').includes('local');
            const originTag = isLocal ? 
                `<span class="status-tag" style="background:#0284c7; color:white;">🏭 Fabrication Locale</span>` : 
                `<span class="status-tag" style="background:#64748b; color:white;">🚢 Importé (${d.legal_manufacturer_country || 'DE'})</span>`;

            html += `
                <tr>
                    <td>
                        <strong style="color:var(--text-main); font-size:12px;">${escapeHTML(d.commercial_name_fr || 'Sans nom')}</strong>
                        <div style="font-size:10px; color:var(--text-muted); font-family:'Traditional Arabic', Tahoma;">${escapeHTML(d.commercial_name_ar || '')}</div>
                        <div style="font-size:10px; color:var(--text-dim); margin-top:2px;">GMDN: <code>${d.gmdn_code || '34978'}</code></div>
                    </td>
                    <td><code>${escapeHTML(d.manufacturer_reference || 'N/A')}</code></td>
                    <td><span class="status-tag" style="background:#0f3b60; color:white; font-weight:bold;">Classe ${d.dm_class || 'IIa'}</span></td>
                    <td>${originTag}</td>
                    <td>${isSterile}</td>
                    <td><span class="status-tag status-ok">En Enregistrement</span></td>
                    <td style="text-align:right; white-space:nowrap;">
                        <button class="btn-pill" style="font-size:11px; padding:4px 8px; margin-right:4px;" onclick="openDMHomologationModal('${d.id}')" title="Éditer et vérifier les 6 sections techniques">📋 6 Sections</button>
                        <button class="btn-pill" style="font-size:11px; padding:4px 8px; margin-right:4px;" onclick="openDMRecevabiliteModal('${d.id}')" title="Grille des 13 points de recevabilité et signature PDT">📑 Recevabilité (13 pts)</button>
                        <button class="btn-pill" style="font-size:11px; padding:4px 8px;" onclick="previewDMReport('ES-REG-02-DM', '${d.id}')" title="Générer et prévisualiser ES-REG-02-DM">👁️ ES-02</button>
                    </td>
                </tr>
            `;
        });
    }

    html += `
                </tbody>
            </table>
        </div>
    `;

    container.innerHTML = html;
}

function filterDMClass(cl) {
    dmFilterClass = cl;
    renderDMDevicesTable();
}

// ============================================================================
// 2. MODALE DES 6 SECTIONS TECHNIQUES (ES-REG-02-DM)
// ============================================================================

async function openDMHomologationModal(deviceID) {
    try {
        const resp = await fetch(`/api/v1/medical-devices/${deviceID}/homologation`);
        if (!resp.ok) throw new Error("Erreur lecture DM: " + resp.status);
        const data = await resp.json();
        activeDMForEdit = data.medical_device;
        populateDMHomologationModal(activeDMForEdit);
        document.getElementById("modal-dm-homologation").classList.add("open");
    } catch (e) {
        alert("Impossible de charger la fiche technique DM: " + e.message);
    }
}

function populateDMHomologationModal(dm) {
    document.getElementById("dm-edit-id").value = dm.id;
    document.getElementById("dm-edit-name-fr").value = dm.commercial_name_fr || '';
    document.getElementById("dm-edit-name-ar").value = dm.commercial_name_ar || '';
    document.getElementById("dm-edit-ref").value = dm.manufacturer_reference || '';
    document.getElementById("dm-edit-gmdn").value = dm.gmdn_code || '';
    document.getElementById("dm-edit-class").value = dm.dm_class || 'IIa';
    document.getElementById("dm-edit-rule").value = dm.classification_rule || '';
    document.getElementById("dm-edit-intended-use").value = dm.intended_use_fr || '';
    
    // Stérilité & Stérilisation
    document.getElementById("dm-edit-is-sterile").checked = !!dm.is_sterile;
    document.getElementById("dm-edit-steril-method").value = dm.sterilization_method || "Oxyde d'éthylène (ETO)";
    document.getElementById("dm-edit-is-single-use").checked = dm.is_single_use !== false;
    document.getElementById("dm-edit-steril-val").value = dm.sterilization_validation_ref || 'VAL-STERIL-ISO11135-2024';
    document.getElementById("dm-edit-pack-stab").value = dm.packaging_stability_iso11607_ref || 'STAB-ISO11607-36M-2025';

    // Composition & Dépistage
    document.getElementById("dm-edit-composition").value = dm.composition_details || 'Corps en Polyuréthane thermoplastique (PUR) biocompatible 85%, Canule acier 304L 10%, Raccord Luer-Lock Polypropylène 5%';
    document.getElementById("dm-edit-has-latex").checked = !!dm.has_latex;
    document.getElementById("dm-edit-has-phthalates").checked = !!dm.has_phthalates;
    document.getElementById("dm-edit-has-active-pharma").checked = !!dm.has_active_pharma;
    document.getElementById("dm-edit-has-biological").checked = !!dm.has_biological_origin;
    document.getElementById("dm-edit-has-cmr").checked = !!dm.has_cmr_substances;

    // Normes ISO & Sécurité
    document.getElementById("dm-edit-iso14971").value = dm.risk_management_iso14971_ref || 'PGR-ISO14971-2025-V01';
    document.getElementById("dm-edit-iso10993").value = dm.biocompatibility_iso10993_ref || 'BIO-ISO10993-GLP-2025-08';
    document.getElementById("dm-edit-iso14155").value = dm.clinical_eval_iso14155_ref || 'CER-ISO14155-2025-CLIN';

    // Échantillonnage
    document.getElementById("dm-edit-batch-pf").value = dm.sample_batch_pf || 'LOT-PF-2026-DM01';
    document.getElementById("dm-edit-batch-mp").value = dm.sample_batch_mp || 'LOT-MP-PUR-098';
    document.getElementById("dm-edit-sample-count").value = dm.sample_units_count || 3;
}

async function saveDMHomologationForm(event) {
    if (event) event.preventDefault();
    const id = document.getElementById("dm-edit-id").value;
    if (!id) return;

    const payload = {
        ...activeDMForEdit,
        commercial_name_fr: document.getElementById("dm-edit-name-fr").value,
        commercial_name_ar: document.getElementById("dm-edit-name-ar").value,
        manufacturer_reference: document.getElementById("dm-edit-ref").value,
        gmdn_code: document.getElementById("dm-edit-gmdn").value,
        dm_class: document.getElementById("dm-edit-class").value,
        classification_rule: document.getElementById("dm-edit-rule").value,
        intended_use_fr: document.getElementById("dm-edit-intended-use").value,
        is_sterile: document.getElementById("dm-edit-is-sterile").checked,
        sterilization_method: document.getElementById("dm-edit-steril-method").value,
        is_single_use: document.getElementById("dm-edit-is-single-use").checked,
        sterilization_validation_ref: document.getElementById("dm-edit-steril-val").value,
        packaging_stability_iso11607_ref: document.getElementById("dm-edit-pack-stab").value,
        composition_details: document.getElementById("dm-edit-composition").value,
        has_latex: document.getElementById("dm-edit-has-latex").checked,
        has_phthalates: document.getElementById("dm-edit-has-phthalates").checked,
        has_active_pharma: document.getElementById("dm-edit-has-active-pharma").checked,
        has_biological_origin: document.getElementById("dm-edit-has-biological").checked,
        has_cmr_substances: document.getElementById("dm-edit-has-cmr").checked,
        risk_management_iso14971_ref: document.getElementById("dm-edit-iso14971").value,
        biocompatibility_iso10993_ref: document.getElementById("dm-edit-iso10993").value,
        clinical_eval_iso14155_ref: document.getElementById("dm-edit-iso14155").value,
        sample_batch_pf: document.getElementById("dm-edit-batch-pf").value,
        sample_batch_mp: document.getElementById("dm-edit-batch-mp").value,
        sample_units_count: parseInt(document.getElementById("dm-edit-sample-count").value, 10) || 3,
    };

    try {
        const resp = await fetch(`/api/v1/medical-devices/${id}/homologation`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });
        if (!resp.ok) {
            const err = await resp.json();
            throw new Error(err.message || "Erreur sauvegarde");
        }
        closeModal('modal-dm-homologation');
        showNotifyModal("Fiche Technique DM Sauvegardée", "Les 6 sections d'homologation ont été enregistrées avec succès conformément à l'Arrêté du 10 mai 2021.", "✅", "OK");
        loadDMDevices();
    } catch (e) {
        alert("Erreur lors de la sauvegarde : " + e.message);
    }
}

// ============================================================================
// 3. MODALE D'ÉVALUATION DE RECEVABILITÉ (13 POINTS — ES-REG-03-DM)
// ============================================================================

// Helper pour récupérer ou initialiser un dossier d'homologation pour un DM
async function getOrCreateDMDossierID(deviceID) {
    try {
        const dResp = await fetch('/api/v1/dossiers');
        if (dResp.ok) {
            const dData = await dResp.json();
            const dList = dData.dossiers || [];
            const found = dList.find(d => d.medical_device_id === deviceID);
            if (found) return found.id;
        }
        // Auto-création si non existant
        const dm = (dmDevicesList || []).find(d => d.id === deviceID);
        const refSuffix = (dm && dm.manufacturer_reference ? dm.manufacturer_reference : 'DM').replace(/[^a-zA-Z0-9_-]/g, '').slice(0, 12);
        const createResp = await fetch('/api/v1/dossiers', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                procedure_code: 'HOMOLOG_DM',
                product_type: 'dispositif_medical',
                internal_reference: `HOM-DM-${refSuffix}-${Date.now().toString().slice(-4)}`,
                medical_device_id: deviceID,
                status: 'brouillon'
            })
        });
        if (createResp.ok) {
            const cData = await createResp.json();
            return cData.dossier ? cData.dossier.id : (cData.id || null);
        }
    } catch (e) {
        console.error("Erreur getOrCreateDMDossierID:", e);
    }
    return null;
}

async function openDMRecevabiliteModal(deviceID) {
    try {
        const dossierID = await getOrCreateDMDossierID(deviceID);
        if (!dossierID) {
            alert("Impossible de localiser ou d'initialiser le dossier d'homologation pour ce DM.");
            return;
        }

        const resp = await fetch(`/api/v1/dossiers/${dossierID}/dm-recevabilite`);
        if (!resp.ok) throw new Error("Erreur lecture recevabilité: " + resp.status);
        const data = await resp.json();
        activeDMEval = data.evaluation;
        populateDMRecevabiliteModal(activeDMEval, dossierID);
        document.getElementById("modal-dm-recevabilite").classList.add("open");
    } catch (e) {
        alert("Impossible d'ouvrir la recevabilité: " + e.message);
    }
}

function populateDMRecevabiliteModal(evalData, dossierID) {
    document.getElementById("dm-rec-dossier-id").value = dossierID;
    document.getElementById("dm-rec-order-number").value = evalData.deposit_order_number || 'ANPP/DDM/REC/2026/0148';
    document.getElementById("dm-rec-agent").value = evalData.anpp_agent_name || "Dr. Amel Bouzid (Sous-Direction de l'Homologation DDM)";
    document.getElementById("dm-rec-copies").value = evalData.electronic_copies_count || 4;
    document.getElementById("dm-rec-reserves").value = evalData.reserves_notes || '';

    // Assignation des 13 points
    for (let i = 1; i <= 13; i++) {
        const sel = document.getElementById(`dm-rec-item-${i}`);
        if (sel) {
            sel.value = evalData[`item_${i}_status`] || 'conforme';
        }
    }

    updateRecevabiliteScoreUI();

    // Sceau PDT
    const sealZone = document.getElementById("dm-rec-seal-zone");
    if (evalData.pdt_signature_seal) {
        sealZone.innerHTML = `
            <div style="background:#f0fdf4; border:1px solid #059669; padding:8px 12px; border-radius:4px; font-size:11px; color:#065f46;">
                <strong>✔ Sceau 21 CFR Part 11 apposé :</strong> <code>${evalData.pdt_signature_seal}</code><br/>
                <span style="font-size:10px;">Signé par le Pharmacien DT le ${evalData.pdt_signed_at ? new Date(evalData.pdt_signed_at).toLocaleString() : ''}</span>
            </div>
        `;
        document.getElementById("btn-seal-dm-rec").style.display = "none";
    } else {
        sealZone.innerHTML = `
            <div style="background:#f8fafc; border:1px dashed #cbd5e1; padding:8px 12px; border-radius:4px; font-size:11px; color:#64748b;">
                En attente du visa solennel du Pharmacien Directeur Technique (Mandat 1.A.7).
            </div>
        `;
        document.getElementById("btn-seal-dm-rec").style.display = "inline-flex";
    }
}

function updateRecevabiliteScoreUI() {
    let score = 0;
    let hasReserve = false;
    let hasNonConforme = false;

    for (let i = 1; i <= 13; i++) {
        const sel = document.getElementById(`dm-rec-item-${i}`);
        if (sel) {
            if (sel.value === 'conforme') score++;
            else if (sel.value === 'avec_reserve') hasReserve = true;
            else hasNonConforme = true;
        }
    }

    const scoreEl = document.getElementById("dm-rec-score-badge");
    const decisionEl = document.getElementById("dm-rec-decision-text");

    if (scoreEl) {
        scoreEl.textContent = `${score} / 13 Points Conformes`;
        scoreEl.style.background = score === 13 ? '#059669' : (hasNonConforme ? '#dc2626' : '#d97706');
    }

    if (decisionEl) {
        if (hasNonConforme || score < 10) {
            decisionEl.innerHTML = `<span style="color:#dc2626; font-weight:800;">⛔ IRRECEVABLE</span> — Manquements majeurs constatés`;
        } else if (hasReserve || score < 13) {
            decisionEl.innerHTML = `<span style="color:#d97706; font-weight:800;">⚠️ RECEVABLE AVEC RÉSERVES</span> — Compléments requis sous 30 jours`;
        } else {
            decisionEl.innerHTML = `<span style="color:#059669; font-weight:800;">✓ COMPLET ET RECEVABLE</span> — Conforme pour instruction DDM`;
        }
    }
}

async function saveDMRecevabiliteForm() {
    const dossierID = document.getElementById("dm-rec-dossier-id").value;
    if (!dossierID) return;

    const payload = {
        deposit_order_number: document.getElementById("dm-rec-order-number").value,
        anpp_agent_name: document.getElementById("dm-rec-agent").value,
        electronic_copies_count: parseInt(document.getElementById("dm-rec-copies").value, 10) || 4,
        reserves_notes: document.getElementById("dm-rec-reserves").value,
    };

    for (let i = 1; i <= 13; i++) {
        const sel = document.getElementById(`dm-rec-item-${i}`);
        if (sel) payload[`item_${i}_status`] = sel.value;
    }

    try {
        const resp = await fetch(`/api/v1/dossiers/${dossierID}/dm-recevabilite`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });
        if (!resp.ok) {
            const err = await resp.json();
            throw new Error(err.message || "Erreur sauvegarde recevabilité");
        }
        showNotifyModal("Recevabilité Enregistrée", "La grille en 13 points a été mise à jour avec succès.", "✅", "OK");
    } catch (e) {
        alert("Erreur: " + e.message);
    }
}

async function sealDMRecevabiliteDT() {
    const dossierID = document.getElementById("dm-rec-dossier-id").value;
    if (!dossierID) return;

    const confirmed = confirm("En qualité de Pharmacien Directeur Technique, confirmez-vous solennellement la complétude et l'exactitude des pièces soumises sous le N° d'ordre DDM ?");
    if (!confirmed) return;

    try {
        const resp = await fetch(`/api/v1/dossiers/${dossierID}/dm-recevabilite/seal`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                solemn_declaration: true,
                decision_miph: "DEC-MIPH-PDT-2022-7714"
            })
        });
        if (!resp.ok) {
            const err = await resp.json();
            throw new Error(err.message || "Erreur scellement");
        }
        const data = await resp.json();
        closeModal('modal-dm-recevabilite');
        showNotifyModal("Récépissé Scellé (21 CFR Part 11)", `L'acte officiel ES-REG-03-DM a été scellé par le Pharmacien DT avec l'empreinte :<br><code>${data.seal}</code>`, "🔏", "OK");
        loadDMDevices();
    } catch (e) {
        alert("Erreur scellement: " + e.message);
    }
}

// ============================================================================
// 4. VOLET B : MESURES TRANSITOIRES & RÉGULARISATION DES DM NON HOMOLOGUÉS
// ============================================================================

async function loadDMRegularizations() {
    const container = document.getElementById("table-container") || document.querySelector(".table-container");
    if (!container) return;

    container.innerHTML = `
        <div style="padding:40px; text-align:center; color:var(--text-muted);">
            <div class="spinner" style="margin-bottom:12px;"></div>
            Chargement des dossiers sous mesures transitoires et déclarations dérogatoires...
        </div>
    `;

    try {
        const [statsResp, listResp] = await Promise.all([
            fetch('/api/v1/dm-regularizations/stats'),
            fetch(`/api/v1/dm-regularizations?status=${dmRegFilterStatus}`)
        ]);

        if (statsResp.ok) {
            const sData = await statsResp.json();
            dmRegularizationStats = sData.stats;
        }

        if (listResp.ok) {
            const lData = await listResp.json();
            dmRegularizationsList = lData.regularizations || [];
        }

        renderDMRegularizationsDashboard();
    } catch (e) {
        container.innerHTML = `
            <div style="padding:20px; color:var(--accent-red); background:rgba(239,68,68,0.1); border-radius:6px; margin:16px;">
                ⚠️ Impossible de charger les mesures transitoires DM: ${e.message}
            </div>
        `;
    }
}

function renderDMRegularizationsDashboard() {
    const container = document.getElementById("table-container") || document.querySelector(".table-container");
    if (!container) return;

    const stats = dmRegularizationStats || {
        total_devices_count: 0,
        derogatoire_count: 0,
        depose_ddm_count: 0,
        en_instruction_count: 0,
        regularise_count: 0,
        critical_deadline_count: 0,
        urgent_deadline_count: 0,
        alert_deadline_count: 0,
        total_volume_distributed: 0,
    };

    let html = `
        <!-- EN-TÊTE DU COCKPIT RÉGULARISATION -->
        <div style="padding:16px 20px; border-bottom:1px solid var(--border-color); background:var(--bg-surface);">
            <div style="display:flex; justify-content:space-between; align-items:center;">
                <div>
                    <h2 style="font-size:14px; font-weight:800; color:var(--text-main); margin:0; display:flex; align-items:center; gap:8px;">
                        <span>⏱️</span> Suivi des Mesures Transitoires & Régularisation des DM non homologués
                        <span class="status-tag" style="background:#b45309; color:white; font-size:10px;">Décret exécutif 20-324</span>
                    </h2>
                    <div style="font-size:11px; color:var(--text-muted); margin-top:3px;">
                        Contrôle des autorisations dérogatoires, compte à rebours légal DDM et canevas officiel de déclaration journalière
                    </div>
                </div>
                <div style="display:flex; gap:8px;">
                    <button class="btn-pill" style="background:#0284c7; color:white; border:none; font-size:11px;" onclick="exportDMCanevasXLSX()">📊 Télécharger Canevas Officiel (.xlsx)</button>
                </div>
            </div>

            <!-- 4 CARTES KPIS -->
            <div style="display:grid; grid-template-columns:repeat(4, 1fr); gap:12px; margin-top:14px;">
                <div style="background:var(--bg-card); border:1px solid var(--border-color); border-radius:6px; padding:10px 14px;">
                    <div style="font-size:10.5px; text-transform:uppercase; font-weight:700; color:var(--text-muted);">DM sous Dérogation</div>
                    <div style="font-size:22px; font-weight:900; color:var(--text-main); margin-top:2px;">${stats.total_devices_count}</div>
                    <div style="font-size:10px; color:var(--text-dim); margin-top:2px;">${stats.derogatoire_count} dérogations actives</div>
                </div>
                <div style="background:var(--bg-card); border:1px solid var(--border-color); border-radius:6px; padding:10px 14px;">
                    <div style="font-size:10.5px; text-transform:uppercase; font-weight:700; color:#d97706;">Dossiers Déposés DDM</div>
                    <div style="font-size:22px; font-weight:900; color:#d97706; margin-top:2px;">${stats.depose_ddm_count + stats.en_instruction_count}</div>
                    <div style="font-size:10px; color:var(--text-dim); margin-top:2px;">${stats.en_instruction_count} en instruction technique</div>
                </div>
                <div style="background:var(--bg-card); border:1px solid var(--border-color); border-radius:6px; padding:10px 14px;">
                    <div style="font-size:10.5px; text-transform:uppercase; font-weight:700; color:#dc2626;">Échéances Critiques ≤ 30j</div>
                    <div style="font-size:22px; font-weight:900; color:#dc2626; margin-top:2px;">${stats.critical_deadline_count}</div>
                    <div style="font-size:10px; color:var(--text-dim); margin-top:2px;">+ ${stats.urgent_deadline_count} à J≤60 (Urgent)</div>
                </div>
                <div style="background:var(--bg-card); border:1px solid var(--border-color); border-radius:6px; padding:10px 14px;">
                    <div style="font-size:10.5px; text-transform:uppercase; font-weight:700; color:#059669;">Volume Distribué Cumulé</div>
                    <div style="font-size:22px; font-weight:900; color:#059669; margin-top:2px;">${stats.total_volume_distributed.toLocaleString()}</div>
                    <div style="font-size:10px; color:var(--text-dim); margin-top:2px;">Unités déclarées à l'ANPP</div>
                </div>
            </div>

            <!-- FILTRES STATUT -->
            <div style="display:flex; gap:6px; margin-top:14px;">
                <button class="btn-pill ${dmRegFilterStatus === 'all' ? 'active' : ''}" onclick="filterDMRegStatus('all')">Tous</button>
                <button class="btn-pill ${dmRegFilterStatus === 'commercialise_derogatoire' ? 'active' : ''}" onclick="filterDMRegStatus('commercialise_derogatoire')">Commercialisé Dérogatoire</button>
                <button class="btn-pill ${dmRegFilterStatus === 'dossier_depose_ddm' ? 'active' : ''}" onclick="filterDMRegStatus('dossier_depose_ddm')">Dossier Déposé DDM</button>
                <button class="btn-pill ${dmRegFilterStatus === 'en_instruction' ? 'active' : ''}" onclick="filterDMRegStatus('en_instruction')">En Instruction</button>
                <button class="btn-pill ${dmRegFilterStatus === 'homologue_regularise' ? 'active' : ''}" onclick="filterDMRegStatus('homologue_regularise')">Homologué & Régularisé</button>
            </div>
        </div>

        <!-- TABLEAU DES DOSSIERS TRANSITOIRES -->
        <div style="overflow-x:auto; padding:12px 16px;">
            <table class="dense-table" style="width:100%; font-size:11px;">
                <thead>
                    <tr>
                        <th style="width:12%;">Code Dossier</th>
                        <th style="width:22%;">Dispositif Médical</th>
                        <th style="width:14%;">Dérogation Ministérielle</th>
                        <th style="width:14%;">Statut Régularisation</th>
                        <th style="width:14%;">Compte à Rebours Légal</th>
                        <th style="width:12%;">Distribution</th>
                        <th style="width:12%; text-align:right;">Actions</th>
                    </tr>
                </thead>
                <tbody>
    `;

    if (dmRegularizationsList.length === 0) {
        html += `
            <tr>
                <td colspan="7" style="text-align:center; padding:30px; color:var(--text-muted);">
                    Aucun dossier de régularisation sous ce filtre.
                </td>
            </tr>
        `;
    } else {
        dmRegularizationsList.forEach(r => {
            const dm = r.medical_device || {};
            let statusTag = '';
            switch (r.status) {
                case 'commercialise_derogatoire':
                    statusTag = `<span class="status-tag" style="background:#b45309; color:white;">Commercialisé (Dérogation)</span>`;
                    break;
                case 'dossier_depose_ddm':
                    statusTag = `<span class="status-tag" style="background:#0284c7; color:white;">Déposé au Guichet DDM</span>`;
                    break;
                case 'en_instruction':
                    statusTag = `<span class="status-tag" style="background:#6366f1; color:white;">En Instruction Technique</span>`;
                    break;
                case 'homologue_regularise':
                    statusTag = `<span class="status-tag" style="background:#059669; color:white;">✔ Homologué & Régularisé</span>`;
                    break;
                default:
                    statusTag = `<span class="status-tag">${escapeHTML(r.status)}</span>`;
            }

            let urgencyBadge = '';
            if (r.status === 'homologue_regularise') {
                urgencyBadge = `<span class="status-tag status-ok">Dossier Clos</span>`;
            } else if (r.days_remaining <= 0) {
                urgencyBadge = `<span class="status-tag" style="background:#dc2626; color:white; font-weight:bold;">⛔ EXPIRÉ (${Math.abs(r.days_remaining)}j)</span>`;
            } else if (r.days_remaining <= 30) {
                urgencyBadge = `<span class="status-tag" style="background:#dc2626; color:white; font-weight:bold;">🔴 J-${r.days_remaining} (Critique)</span>`;
            } else if (r.days_remaining <= 60) {
                urgencyBadge = `<span class="status-tag" style="background:#ea580c; color:white; font-weight:bold;">🟠 J-${r.days_remaining} (Urgent)</span>`;
            } else if (r.days_remaining <= 90) {
                urgencyBadge = `<span class="status-tag" style="background:#d97706; color:white; font-weight:bold;">🟡 J-${r.days_remaining} (Alerte)</span>`;
            } else {
                urgencyBadge = `<span class="status-tag" style="background:#059669; color:white;">🟢 J-${r.days_remaining} (Dans les délais)</span>`;
            }

            html += `
                <tr>
                    <td>
                        <strong style="color:var(--text-main);">${escapeHTML(r.regularization_code)}</strong>
                        <div style="font-size:9.5px; color:var(--text-dim); margin-top:2px;">Créé le ${new Date(r.created_at).toLocaleDateString()}</div>
                    </td>
                    <td>
                        <strong style="color:var(--text-main);">${escapeHTML(dm.commercial_name_fr || 'DM')}</strong>
                        <div style="font-size:10px; color:var(--text-muted);">Classe ${dm.dm_class || 'IIa'} • Réf: <code>${escapeHTML(dm.manufacturer_reference || '')}</code></div>
                    </td>
                    <td>
                        <code>${escapeHTML(r.derogation_number || 'N/A')}</code>
                        <div style="font-size:9.5px; color:var(--text-dim); margin-top:2px;">Sur marché: ${new Date(r.initial_market_date).toLocaleDateString()}</div>
                    </td>
                    <td>
                        ${statusTag}
                        ${r.ddm_acknowledgement_number ? `<div style="font-size:9.5px; color:var(--text-muted); margin-top:2px;">N° Rec: <code>${escapeHTML(r.ddm_acknowledgement_number)}</code></div>` : ''}
                    </td>
                    <td>
                        ${urgencyBadge}
                        <div style="font-size:9.5px; color:var(--text-muted); margin-top:2px;">Butoir: <strong>${new Date(r.legal_deadline).toLocaleDateString()}</strong></div>
                    </td>
                    <td>
                        <div><strong>${(r.quarterly_volume_distributed || 0).toLocaleString()}</strong> unités</div>
                        <div style="font-size:9.5px; color:var(--text-dim);">${r.hospital_distribution_share}% Hôpital / ${r.retail_distribution_share}% Officine</div>
                    </td>
                    <td style="text-align:right;">
                        <button class="btn-pill" style="font-size:10.5px; padding:3px 8px;" onclick="openEditDMRegModal('${r.id}')">✏️ Modifier</button>
                    </td>
                </tr>
            `;
        });
    }

    html += `
                </tbody>
            </table>
        </div>
    `;

    container.innerHTML = html;
}

function filterDMRegStatus(st) {
    dmRegFilterStatus = st;
    loadDMRegularizations();
}

// ============================================================================
// 5. EXPORT DU CANEVAS DÉROGATOIRE (.XLSX / CSV FORMATTE)
// ============================================================================

function exportDMCanevasXLSX() {
    window.location.href = '/api/v1/dm-regularizations/export/canevas';
}

// ============================================================================
// 6. PRÉVISUALISATION IN-APP DES ÉTATS OFFICIELS (ES-REG-02-DM & ES-REG-03-DM)
// ============================================================================

async function previewDMReport(reportCode, deviceID) {
    try {
        const dossierID = await getOrCreateDMDossierID(deviceID);
        if (!dossierID) {
            alert("Impossible d'ouvrir l'état officiel : aucun dossier d'homologation disponible.");
            return;
        }

        const reportURL = `/api/v1/dossiers/${dossierID}/reports/${reportCode}?format=html`;
        
        // Utiliser la modale in-app universelle
        const modal = document.getElementById("modal-variation-report-viewer");
        const iframe = document.getElementById("variation-report-iframe") || document.getElementById("variation-report-frame");
        const titleEl = document.getElementById("variation-report-title") || (modal ? modal.querySelector(".modal-title") : null);
        const subTitleEl = document.getElementById("variation-report-subtitle");

        if (modal && iframe) {
            if (titleEl) {
                titleEl.textContent = reportCode === 'ES-REG-02-DM' ? 
                    "ES-REG-02-DM • Formulaire Officiel d'Homologation des Dispositifs Médicaux (6 Sections)" :
                    "ES-REG-03-DM • Récépissé Officiel de Dépôt & Recevabilité DM (13 Points Arrêté 10 Mai 2021)";
            }
            if (subTitleEl) {
                subTitleEl.textContent = reportCode === 'ES-REG-02-DM' ?
                    "Arrêté du 10 mai 2021 & Décret 20-324 • Dossier Technique DDM Opposable A4" :
                    "Arrêté du 10 mai 2021 (Art. 05 & 06) • Évaluation de Recevabilité & Scellement PDT (21 CFR Part 11)";
            }
            iframe.src = reportURL;
            window.activeVariationReportURL = reportURL;
            modal.classList.add("open");
        } else {
            window.open(reportURL, '_blank');
        }
    } catch (e) {
        alert("Erreur ouverture état officiel : " + e.message);
    }
}

// ============================================================================
// 7. CRÉATION D'UN NOUVEAU SUIVI DE RÉGULARISATION
// ============================================================================

function openNewDMRegularizationModal() {
    // Remplir le sélecteur de DM disponibles
    const sel = document.getElementById("new-dm-reg-device-id");
    if (sel) {
        sel.innerHTML = dmDevicesList.map(d => `
            <option value="${d.id}">${escapeHTML(d.commercial_name_fr)} (Classe ${d.dm_class})</option>
        `).join('');
    }
    const codeInput = document.getElementById("new-dm-reg-code");
    if (codeInput) {
        codeInput.value = `REG-DM-2026-${String(dmRegularizationsList.length + 1).padStart(3, '0')}`;
    }
    document.getElementById("modal-dm-regularization-new").classList.add("open");
}

async function submitNewDMRegularization(event) {
    if (event) event.preventDefault();

    const payload = {
        medical_device_id: document.getElementById("new-dm-reg-device-id").value,
        regularization_code: document.getElementById("new-dm-reg-code").value,
        derogation_number: document.getElementById("new-dm-reg-derogation").value,
        status: document.getElementById("new-dm-reg-status").value,
        quarterly_volume_imported: parseFloat(document.getElementById("new-dm-reg-vol-imp").value) || 0,
        quarterly_volume_distributed: parseFloat(document.getElementById("new-dm-reg-vol-dist").value) || 0,
        hospital_distribution_share: parseInt(document.getElementById("new-dm-reg-hosp-share").value, 10) || 70,
        retail_distribution_share: parseInt(document.getElementById("new-dm-reg-ret-share").value, 10) || 30,
        observations: document.getElementById("new-dm-reg-obs").value,
    };

    try {
        const resp = await fetch('/api/v1/dm-regularizations', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(payload)
        });
        if (!resp.ok) {
            const err = await resp.json();
            throw new Error(err.message || "Erreur création régularisation");
        }
        closeModal('modal-dm-regularization-new');
        showNotifyModal("Régularisation Enregistrée", "Le dispositif médical a été inscrit au registre des mesures transitoires DDM.", "✅", "OK");
        loadDMRegularizations();
    } catch (e) {
        alert("Erreur: " + e.message);
    }
}
