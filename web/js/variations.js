// ============================================================================
// MOTEUR DE VARIATIONS ANPP (Ligne Directrice LD-DEPP07/V01 & Décret 21-82)
// Référentiel des 34 Rubriques, Escalade Dynamique, Grouping & ES-REG-04
// ============================================================================

let currentVariationsList = [];
let selectedVariationID = null;
let masterRubriquesList = [];
let availableMedicinesList = [];

// Rubrique active dans le formulaire de création
let currentSelectedRubriqueDef = null;

// ============================================================================
// 1. CHARGEMENT ET AFFICHAGE DES VARIATIONS
// ============================================================================

async function renderVariationsTable() {
    if (window.innerWidth < 768 && typeof toggleMobileTree === 'function') {
        toggleMobileTree(false);
    }
    const thead = document.getElementById("table-head");
    const tbody = document.getElementById("doc-tbody");
    if (!tbody || !thead) return;

    thead.innerHTML = `
        <tr>
            <th style="width: 100px;">Classification</th>
            <th>N° Enregistrement &amp; Objet de la Variation (LD-DEPP07/V01)</th>
            <th>Spécialité Pharmaceutique</th>
            <th class="col-hide-md">D.C.I.</th>
            <th class="col-hide-sm">Redevance (BV2)</th>
            <th class="col-hide-lg">Grouping (Art. 09)</th>
            <th>Statut BPF</th>
            <th style="width: 150px; text-align:center;">Actions Réglementaires</th>
        </tr>
    `;

    tbody.innerHTML = `<tr><td colspan="8" style="text-align:center; padding:30px; color:var(--text-muted);"><span class="spin">⏳</span> Chargement des modifications post-AMM...</td></tr>`;

    try {
        const resp = await fetch('/api/v1/variations');
        if (!resp.ok) throw new Error("Erreur HTTP " + resp.status);
        const data = await resp.json();
        currentVariationsList = data.variations || [];

        if (currentVariationsList.length === 0) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="8" style="text-align:center; padding:40px; color:var(--text-muted);">
                        <div style="font-size:24px; margin-bottom:8px;">🔄</div>
                        <div style="font-weight:700; font-size:13px; color:var(--text-main);">Aucune demande de variation post-AMM enregistrée</div>
                        <div style="font-size:11px; margin-top:4px;">Utilisez le bouton <strong>➕ Déposer une Variation</strong> dans la barre supérieure pour initialiser un dossier selon la LD-DEPP07/V01.</div>
                    </td>
                </tr>
            `;
            const preview = document.getElementById("preview-body");
            if (preview) preview.innerHTML = `<div style="text-align:center; padding:40px; color:var(--text-muted);">Sélectionnez une variation pour visualiser l'Impact Tree et le dossier réglementaire.</div>`;
            return;
        }

        tbody.innerHTML = "";
        currentVariationsList.forEach((v) => {
            const tr = document.createElement("tr");
            if (v.id === selectedVariationID) tr.classList.add("selected");
            tr.onclick = (e) => {
                if (e.target.closest('button')) return;
                selectedVariationID = v.id;
                document.querySelectorAll('#doc-tbody tr').forEach(row => row.classList.remove('selected'));
                tr.classList.add('selected');
                renderVariationPreview(v.id);
                if (window.innerWidth < 1400 && typeof togglePreviewPane === 'function') togglePreviewPane(true);
            };

            // Badge de classification
            let classBadge = `<span class="status-tag status-req" style="background:#0284c7; color:white; font-weight:800;">${v.overall_classification}</span>`;
            if (v.overall_classification === 'Mmaj') {
                classBadge = `<span class="status-tag status-req" style="background:#dc2626; color:white; font-weight:800;">Mmaj</span>`;
            } else if (v.overall_classification === 'Mmod') {
                classBadge = `<span class="status-tag status-rev" style="background:#d97706; color:white; font-weight:800;">Mmod</span>`;
            } else if (v.overall_classification === 'Mmin') {
                classBadge = `<span class="status-tag status-ok" style="background:#059669; color:white; font-weight:800;">Mmin</span>`;
            }

            // Indicateur d'escalade
            let escTag = "";
            if (v.has_escalation) {
                escTag = `<span title="${escapeHTML(v.escalation_summary || 'Reclassé')}" style="background:#fee2e2; color:#b91c1c; font-size:9.5px; padding:2px 5px; border-radius:3px; font-weight:700; margin-left:4px; border:1px solid #fca5a5;">⚠️ Reclassé</span>`;
            }

            // Grouping tag
            let groupTag = `<span style="color:var(--text-muted); font-size:11px;">Unitaire</span>`;
            if (v.is_grouping) {
                groupTag = `<span style="background:rgba(126,34,206,0.12); color:#7e22ce; font-size:10px; padding:2px 6px; border-radius:4px; font-weight:700; border:1px solid rgba(126,34,206,0.25);">🔗 Grouping (${v.items ? v.items.length : 2} rubriques)</span>`;
            }

            // Statut BPF
            let statusBadge = `<span class="status-tag status-draft">Brouillon</span>`;
            if (v.status === 'pret_visa_dt') {
                statusBadge = `<span class="status-tag status-rev" style="background:#7e22ce; color:white; font-weight:700;">Prêt Visa DT</span>`;
            } else if (v.status === 'soumis_anpp') {
                statusBadge = `<span class="status-tag status-ok" style="background:#0284c7; color:white; font-weight:700;">Déposé ANPP</span>`;
            } else if (v.status === 'approuve') {
                statusBadge = `<span class="status-tag status-ok">Approuvé ANPP</span>`;
            }

            const feeFormatted = new Intl.NumberFormat('fr-DZ', { minimumFractionDigits: 2 }).format(v.fee_amount_dzd || 0) + " DZD";

            tr.innerHTML = `
                <td>${classBadge} ${escTag}</td>
                <td>
                    <div style="font-weight:700; color:var(--text-main); font-size:12px;">
                        <code style="color:var(--accent-blue);">${escapeHTML(v.tracking_number)}</code>
                    </div>
                    <div style="font-size:11px; color:var(--text-muted); margin-top:2px;">
                        ${escapeHTML(v.title)}
                    </div>
                </td>
                <td style="font-weight:600; color:var(--text-main);">
                    ${escapeHTML(v.medicine_name || 'Spécialité')}
                    <div style="font-size:10px; color:var(--text-dim);">${escapeHTML(v.registration_number || '')}</div>
                </td>
                <td class="col-hide-md" style="color:var(--text-muted); font-size:11px;">${escapeHTML(v.medicine_dci || '—')}</td>
                <td class="col-hide-sm" style="font-family:monospace; font-weight:700; color:#047857; font-size:11px;">${feeFormatted}</td>
                <td class="col-hide-lg">${groupTag}</td>
                <td>${statusBadge}</td>
                <td style="text-align:center;">
                    <div style="display:flex; gap:4px; justify-content:center;">
                        <button class="btn-pill" style="font-size:10.5px; padding:2px 7px;" onclick="openVariationDocsModal('${v.id}')" title="12 Pièces Justificatives (Section 5.3)">📁 12 Pièces</button>
                        <button class="btn-pill" style="font-size:10.5px; padding:2px 7px; background:var(--accent-blue); color:white; border:none;" onclick="printComparativeTable('${v.id}')" title="Imprimer le Tableau Comparatif ES-REG-04">🖨️ ES-04</button>
                    </div>
                </td>
            `;
            tbody.appendChild(tr);
        });

        // Sélectionner par défaut la première variation
        if (!selectedVariationID && currentVariationsList.length > 0) {
            selectedVariationID = currentVariationsList[0].id;
        }
        if (selectedVariationID) {
            const activeTr = Array.from(tbody.querySelectorAll('tr')).find((_, i) => currentVariationsList[i] && currentVariationsList[i].id === selectedVariationID);
            if (activeTr) activeTr.classList.add('selected');
            renderVariationPreview(selectedVariationID);
        }

    } catch (err) {
        tbody.innerHTML = `<tr><td colspan="8" style="text-align:center; padding:30px; color:var(--accent-red);">Erreur de chargement: ${escapeHTML(err.message)}</td></tr>`;
    }
}

// ============================================================================
// 2. VOLET D'APERÇU CONTEXTUEL : FICHE LD-DEPP07 & CHANGE IMPACT TREE
// ============================================================================

async function renderVariationPreview(varID) {
    const preview = document.getElementById("preview-body");
    if (!preview) return;

    preview.innerHTML = `<div style="text-align:center; padding:30px; color:var(--text-muted);"><span class="spin">⏳</span> Chargement de l'arbre d'impact...</div>`;

    try {
        const resp = await fetch(`/api/v1/variations/${varID}`);
        if (!resp.ok) throw new Error("Variation introuvable");
        const data = await resp.json();
        const v = data.variation;

        const tree = v.change_impact_tree || { direct_red: [], ripple_orange: [], preserved_green: [], stats: {} };

        // Construction du graphe d'impact
        let redBadges = tree.direct_red.map(sec => `
            <div style="display:flex; justify-content:space-between; align-items:center; background:#fee2e2; border:1px solid #fca5a5; padding:4px 8px; border-radius:4px; font-size:10.5px;">
                <span style="font-weight:700; color:#b91c1c;">🔴 ${sec.section_code} — ${escapeHTML(sec.title)}</span>
                <span style="font-size:9.5px; color:#7f1d1d;">${escapeHTML(sec.reason)}</span>
            </div>
        `).join('') || '<div style="font-size:11px; color:var(--text-muted);">Aucune section en impact direct</div>';

        let orangeBadges = tree.ripple_orange.map(sec => `
            <div style="display:flex; justify-content:space-between; align-items:center; background:#fef3c7; border:1px solid #fde68a; padding:4px 8px; border-radius:4px; font-size:10.5px;">
                <span style="font-weight:700; color:#b45309;">🟠 ${sec.section_code} — ${escapeHTML(sec.title)}</span>
                <span style="font-size:9.5px; color:#78350f;">${escapeHTML(sec.reason)}</span>
            </div>
        `).join('') || '<div style="font-size:11px; color:var(--text-muted);">Aucun impact corollaire</div>';

        let greenBadges = tree.preserved_green.slice(0, 5).map(sec => `
            <span style="background:#d1fae5; color:#047857; border:1px solid #a7f3d0; padding:2px 6px; border-radius:3px; font-size:10px; font-weight:700;">
                🟢 ${sec.section_code} ${escapeHTML(sec.title)}
            </span>
        `).join('') + (tree.preserved_green.length > 5 ? ` <span style="font-size:10px; color:var(--text-dim);">+${tree.preserved_green.length - 5} autres...</span>` : '');

        // Liste des items de rubriques
        let itemsHTML = (v.items || []).map(it => `
            <div style="border:1px solid var(--border-color); border-radius:6px; padding:8px; margin-bottom:8px; background:var(--bg-card);">
                <div style="display:flex; justify-content:space-between; align-items:center;">
                    <div style="font-weight:700; font-size:11.5px; color:var(--text-main);">
                        <code style="color:var(--accent-blue);">Rubrique ${escapeHTML(it.rubrique_code)}</code> : ${escapeHTML(it.rubrique_title)}
                    </div>
                    <div>
                        <span class="status-tag status-req" style="background:${it.effective_classification === 'Mmaj' ? '#dc2626' : (it.effective_classification === 'Mmod' ? '#d97706' : '#059669')}; color:white; font-size:9.5px;">
                            ${it.effective_classification}
                        </span>
                        ${it.is_escalated ? '<span style="color:#dc2626; font-size:10px; font-weight:800; margin-left:4px;">⚠️ Reclassé</span>' : ''}
                    </div>
                </div>
                ${it.escalation_reason ? `<div style="font-size:10.5px; color:#dc2626; margin-top:4px; font-style:italic; background:#fff1f2; padding:4px; border-radius:3px;">${escapeHTML(it.escalation_reason)}</div>` : ''}
                <div style="display:grid; grid-template-columns:1fr 1fr; gap:8px; margin-top:6px; font-size:10.5px;">
                    <div style="background:var(--bg-app); padding:4px 6px; border-radius:4px;">
                        <span style="color:var(--text-dim); font-size:9.5px;">Approuvé :</span>
                        <div style="color:var(--text-main);">${escapeHTML(it.current_state || 'Non renseigné')}</div>
                    </div>
                    <div style="background:var(--bg-app); padding:4px 6px; border-radius:4px;">
                        <span style="color:var(--accent-blue); font-size:9.5px;">Proposé :</span>
                        <div style="color:var(--text-main);">${escapeHTML(it.proposed_state || 'Non renseigné')}</div>
                    </div>
                </div>
            </div>
        `).join('');

        // Visa DT
        let dtVisaBox = `
            <div style="background:rgba(126,34,206,0.06); border:1px dashed #7e22ce; border-radius:6px; padding:10px; margin-top:12px; display:flex; justify-content:space-between; align-items:center;">
                <div>
                    <div style="font-weight:700; font-size:11px; color:#7e22ce;">✍️ Visa du Pharmacien Directeur Technique (21 CFR Part 11)</div>
                    <div style="font-size:10.5px; color:var(--text-muted); margin-top:2px;">En attente de signature électronique légale</div>
                </div>
                <button class="btn-pill" style="background:#7e22ce; color:white; border:none; font-weight:700; font-size:11px;" onclick="openSealDTModal('${v.id}')">Apposer le Visa DT</button>
            </div>
        `;
        if (v.dt_signature_seal) {
            dtVisaBox = `
                <div style="background:rgba(5,150,105,0.08); border:1px solid #059669; border-radius:6px; padding:10px; margin-top:12px;">
                    <div style="display:flex; justify-content:space-between; align-items:center;">
                        <span style="font-weight:800; font-size:11.5px; color:#059669;">✔ Scellé par le Pharmacien Directeur Technique</span>
                        <code style="font-size:10px; color:#047857; background:#d1fae5; padding:2px 6px; border-radius:3px;">${escapeHTML(v.dt_signature_seal)}</code>
                    </div>
                    <div style="font-size:10px; color:var(--text-muted); margin-top:4px;">
                        Signé le ${new Date(v.dt_verified_at).toLocaleString('fr-FR')} • Non-modification des autres sections formellement attestée.
                    </div>
                </div>
            `;
        }

        preview.innerHTML = `
            <div style="padding:14px; display:flex; flex-direction:column; gap:12px;">
                <!-- En-tête Fiche Récapitulative -->
                <div style="background:linear-gradient(135deg, rgba(37,99,235,0.08), rgba(2,132,199,0.04)); border:1px solid rgba(37,99,235,0.25); border-radius:8px; padding:12px;">
                    <div style="display:flex; justify-content:space-between; align-items:flex-start;">
                        <div>
                            <span style="font-size:10px; font-weight:800; text-transform:uppercase; color:var(--accent-blue); letter-spacing:0.5px;">Directive LD-DEPP07/V01 • Fiche de Modification</span>
                            <div style="font-weight:900; font-size:14px; color:var(--text-main); margin-top:2px;">${escapeHTML(v.title)}</div>
                            <div style="font-size:11px; color:var(--text-muted); margin-top:2px;">
                                Spécialité : <strong>${escapeHTML(v.medicine_name)}</strong> (${escapeHTML(v.medicine_dci)}) • Décision AMM : <code>${escapeHTML(v.registration_number)}</code>
                            </div>
                        </div>
                        <div style="text-align:right;">
                            <span class="status-tag" style="background:#091e36; color:white; font-size:11px; font-weight:800;">${v.overall_classification}</span>
                            <div style="font-size:10.5px; color:#047857; font-weight:700; margin-top:4px;">${new Intl.NumberFormat('fr-DZ').format(v.fee_amount_dzd)} DZD</div>
                        </div>
                    </div>
                </div>

                <!-- Rubriques incluses -->
                <div>
                    <div style="font-size:11.5px; font-weight:800; color:var(--text-main); text-transform:uppercase; margin-bottom:6px; display:flex; justify-content:space-between;">
                        <span>Rubriques Sélectionnées (${(v.items || []).length})</span>
                        ${v.is_grouping ? '<span style="color:#7e22ce; font-size:10px; font-weight:700;">Procédure Groupée (Art. 09)</span>' : ''}
                    </div>
                    ${itemsHTML}
                </div>

                <!-- Change Impact Tree (Arbre d'impact tri-couleur CDC 5.5) -->
                <div style="background:var(--bg-panel); border:1px solid var(--border-color); border-radius:8px; padding:12px;">
                    <div style="display:flex; justify-content:space-between; align-items:center; border-bottom:1px solid var(--border-color); padding-bottom:6px; margin-bottom:8px;">
                        <span style="font-size:11.5px; font-weight:800; color:var(--text-main); text-transform:uppercase;">
                            🌳 Change Impact Tree (Arborescence des Impacts CTD)
                        </span>
                        <div style="display:flex; gap:6px; font-size:10px;">
                            <span style="color:#b91c1c; font-weight:700;">🔴 Direct : ${tree.stats.direct_red_count || 0}</span>
                            <span style="color:#b45309; font-weight:700;">🟠 Ripple : ${tree.stats.ripple_orange_count || 0}</span>
                            <span style="color:#047857; font-weight:700;">🟢 Préservé : ${tree.stats.preserved_green_count || 0}</span>
                        </div>
                    </div>

                    <div style="font-size:10.5px; font-weight:700; color:#b91c1c; margin-bottom:4px;">Sections CTD en Impact Direct (Remplacement page à page obligatoire) :</div>
                    <div style="display:flex; flex-direction:column; gap:4px; margin-bottom:8px;">
                        ${redBadges}
                    </div>

                    <div style="font-size:10.5px; font-weight:700; color:#b45309; margin-bottom:4px;">Sections en Effet Corollaire (Vérification de cohérence) :</div>
                    <div style="display:flex; flex-direction:column; gap:4px; margin-bottom:8px;">
                        ${orangeBadges}
                    </div>

                    <div style="font-size:10.5px; font-weight:700; color:#047857; margin-bottom:4px;">Sections Préservées (Couvertes par l'Attestation de Non-Modification) :</div>
                    <div style="display:flex; flex-wrap:wrap; gap:4px;">
                        ${greenBadges}
                    </div>
                </div>

                <!-- Justificatif technique -->
                <div style="background:var(--bg-card); border:1px solid var(--border-color); border-radius:6px; padding:10px;">
                    <div style="font-size:10px; font-weight:800; text-transform:uppercase; color:var(--text-dim); margin-bottom:4px;">
                        Justificatif Technico-Réglementaire (Section 5.3)
                    </div>
                    <div style="font-size:11px; line-height:1.5; color:var(--text-main);">
                        ${escapeHTML(v.technical_justification)}
                    </div>
                </div>

                <!-- Visa DT -->
                ${dtVisaBox}

                <!-- Boutons d'impression et export -->
                <div style="display:flex; gap:8px; margin-top:8px;">
                    <button class="btn-pill" style="flex:1; background:var(--accent-blue); color:white; border:none; font-weight:700; padding:8px;" onclick="printComparativeTable('${v.id}')">
                        🖨️ Tableau Comparatif Officiel (ES-REG-04)
                    </button>
                    <button class="btn-pill" style="padding:8px 14px;" onclick="openVariationDocsModal('${v.id}')">
                        📁 12 Pièces Justificatives
                    </button>
                </div>
            </div>
        `;
    } catch (err) {
        preview.innerHTML = `<div style="text-align:center; padding:30px; color:var(--accent-red);">Erreur: ${escapeHTML(err.message)}</div>`;
    }
}

// ============================================================================
// 3. MODALE DE CRÉATION D'UNE VARIATION (AVEC ARBRE DE DÉCISION & CONDITIONS)
// ============================================================================

async function openNewVariationModal() {
    const modal = document.getElementById("modal-new-variation");
    if (!modal) return;

    // Charger les produits et les 34 rubriques si pas encore fait
    try {
        if (availableMedicinesList.length === 0) {
            const resp = await fetch('/api/v1/medicines');
            if (resp.ok) {
                const data = await resp.json();
                availableMedicinesList = data.medicines || [];
            }
        }
        if (masterRubriquesList.length === 0) {
            const resp = await fetch('/api/v1/variations/rubriques');
            if (resp.ok) {
                const data = await resp.json();
                masterRubriquesList = data.rubriques || [];
            }
        }
    } catch (e) {
        console.warn("Erreur préchargement", e);
    }

    // Remplir le sélecteur de médicaments
    const medSelect = document.getElementById("new-var-medicine");
    if (medSelect) {
        medSelect.innerHTML = availableMedicinesList.map(m => `
            <option value="${m.id}">${escapeHTML(m.brand_name_fr)} (${escapeHTML(m.dci_fr)}) — DE: ${escapeHTML(m.registration_decision_number || 'Enregistré')}</option>
        `).join('') || '<option value="">Aucune spécialité disponible</option>';
    }

    // Remplir le sélecteur de rubriques (34 rubriques regroupées par catégorie)
    const rubSelect = document.getElementById("new-var-rubrique");
    if (rubSelect) {
        let currentCat = "";
        let optHTML = '<option value="">-- Sélectionnez l\'une des 34 rubriques LD-DEPP07 --</option>';
        masterRubriquesList.forEach(r => {
            if (r.category_label !== currentCat) {
                if (currentCat !== "") optHTML += '</optgroup>';
                currentCat = r.category_label;
                optHTML += `<optgroup label="${escapeHTML(currentCat)}">`;
            }
            optHTML += `<option value="${r.code}">Rubrique ${r.code} : ${escapeHTML(r.title)} [${r.default_classification}]</option>`;
        });
        if (currentCat !== "") optHTML += '</optgroup>';
        rubSelect.innerHTML = optHTML;
    }

    // Réinitialiser les champs
    document.getElementById("new-var-conditions-box").style.display = "none";
    document.getElementById("new-var-escalation-alert").style.display = "none";
    document.getElementById("new-var-fee-badge").innerText = "40 000,00 DZD";
    document.getElementById("new-var-class-badge").innerText = "Mmin";
    document.getElementById("new-var-current").value = "";
    document.getElementById("new-var-proposed").value = "";
    document.getElementById("new-var-justif").value = "";

    openModal("modal-new-variation");
}

function onRubriqueSelectionChanged() {
    const rubCode = document.getElementById("new-var-rubrique").value;
    const condBox = document.getElementById("new-var-conditions-box");
    const condList = document.getElementById("new-var-conditions-list");
    const alertBox = document.getElementById("new-var-escalation-alert");
    const classBadge = document.getElementById("new-var-class-badge");
    const feeBadge = document.getElementById("new-var-fee-badge");

    if (!rubCode) {
        condBox.style.display = "none";
        alertBox.style.display = "none";
        return;
    }

    currentSelectedRubriqueDef = masterRubriquesList.find(r => r.code === rubCode);
    if (!currentSelectedRubriqueDef) return;

    // Afficher les conditions
    condBox.style.display = "block";
    if (!currentSelectedRubriqueDef.conditions || currentSelectedRubriqueDef.conditions.length === 0) {
        condList.innerHTML = `<div style="font-size:11px; color:var(--text-muted); font-style:italic;">Aucune condition préalable restrictive pour cette rubrique administrative.</div>`;
    } else {
        condList.innerHTML = currentSelectedRubriqueDef.conditions.map(c => `
            <label style="display:flex; align-items:flex-start; gap:8px; font-size:11px; color:var(--text-main); margin-bottom:6px; cursor:pointer;">
                <input type="checkbox" id="cond-check-${c.id}" checked onchange="recalculateFormEscalation()" style="margin-top:2px;">
                <span><strong>[${c.id}]</strong> ${escapeHTML(c.description)}</span>
            </label>
        `).join('');
    }

    recalculateFormEscalation();
}

function recalculateFormEscalation() {
    if (!currentSelectedRubriqueDef) return;

    const alertBox = document.getElementById("new-var-escalation-alert");
    const alertText = document.getElementById("new-var-escalation-text");
    const classBadge = document.getElementById("new-var-class-badge");
    const feeBadge = document.getElementById("new-var-fee-badge");

    const conditions = currentSelectedRubriqueDef.conditions || [];
    let allMet = true;
    let unmetList = [];

    conditions.forEach(c => {
        const chk = document.getElementById(`cond-check-${c.id}`);
        if (chk && !chk.checked) {
            allMet = false;
            unmetList.push(c.id);
        }
    });

    if (allMet) {
        alertBox.style.display = "none";
        const nominal = currentSelectedRubriqueDef.default_classification;
        classBadge.innerText = nominal;
        classBadge.style.background = nominal === 'Mmaj' ? '#dc2626' : (nominal === 'Mmod' ? '#d97706' : '#059669');
        feeBadge.innerText = nominal === 'Mmaj' ? '115 000,00 DZD' : (nominal === 'Mmod' ? '80 000,00 DZD' : '40 000,00 DZD');
    } else {
        // ESCALADE AUTOMATIQUE ENCLENCHÉE
        alertBox.style.display = "flex";
        const escalated = currentSelectedRubriqueDef.escalated_classification;
        alertText.innerHTML = `<strong>⚠️ ESCALADE AUTOMATIQUE ENCLENCHÉE :</strong> La condition <strong>[${unmetList.join(', ')}]</strong> n'étant pas satisfaite, la Ligne Directrice LD-DEPP07/V01 reclasse immédiatement cette demande de <strong>${currentSelectedRubriqueDef.default_classification}</strong> vers <strong>${escalated}</strong> avec ajustement du barème officiel.`;

        classBadge.innerText = escalated + " (Reclassé)";
        classBadge.style.background = escalated === 'Mmaj' ? '#dc2626' : '#d97706';
        feeBadge.innerText = escalated === 'Mmaj' ? '115 000,00 DZD' : '80 000,00 DZD';
    }
}

async function submitNewVariationForm(event) {
    event.preventDefault();

    const medID = document.getElementById("new-var-medicine").value;
    const rubCode = document.getElementById("new-var-rubrique").value;
    const currentText = document.getElementById("new-var-current").value.trim();
    const proposedText = document.getElementById("new-var-proposed").value.trim();
    const justifText = document.getElementById("new-var-justif").value.trim();
    const isGrouping = document.getElementById("new-var-is-grouping").checked;
    const nonModif = document.getElementById("new-var-non-modif").checked;

    if (!medID || !rubCode) {
        showNotifyModal("Champs Incomplets", "Veuillez sélectionner une spécialité et une rubrique officielle.", "⚠️");
        return;
    }
    if (!currentText || !proposedText || !justifText) {
        showNotifyModal("Champs Obligatoires", "La situation actuelle approuvée, la situation nouvelle proposée et le justificatif technique sont requis par la LD-DEPP07.", "⚠️");
        return;
    }

    // Récolter l'état des conditions
    const conditionsEvaluated = {};
    if (currentSelectedRubriqueDef && currentSelectedRubriqueDef.conditions) {
        currentSelectedRubriqueDef.conditions.forEach(c => {
            const chk = document.getElementById(`cond-check-${c.id}`);
            conditionsEvaluated[c.id] = chk ? chk.checked : true;
        });
    }

    try {
        // 1. Créer la demande
        const createResp = await fetch('/api/v1/variations', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                medicine_id: medID,
                title: `${currentSelectedRubriqueDef.title} (${rubCode})`,
                description: `Modification selon la rubrique officielle ${rubCode} de la LD-DEPP07/V01.`,
                is_grouping: isGrouping,
                current_approved_situation: currentText,
                proposed_situation: proposedText,
                technical_justification: justifText,
                non_modification_attested: nonModif
            })
        });

        if (!createResp.ok) {
            const err = await createResp.json();
            throw new Error(err.error || "Erreur création variation");
        }

        const createData = await createResp.json();
        const varID = createData.variation.id;

        // 2. Ajouter l'item rubrique
        await fetch(`/api/v1/variations/${varID}/items`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                rubrique_code: rubCode,
                conditions_evaluated: conditionsEvaluated,
                current_state: currentText,
                proposed_state: proposedText,
                notes: justifText
            })
        });

        closeModal("modal-new-variation");
        showNotifyModal(
            "Variation Initialisée avec Succès",
            `La demande a été créée sous le N° <code>${escapeHTML(createData.variation.tracking_number)}</code>.<br>Les 12 pièces obligatoires de la LD-DEPP07/V01 ont été automatiquement assignées.`,
            "✔"
        );
        selectedVariationID = varID;
        renderVariationsTable();

    } catch (err) {
        showNotifyModal("Erreur", err.message, "❌");
    }
}

// ============================================================================
// 4. MODALE DES 12 PIÈCES JUSTIFICATIVES OBLIGATOIRES (SECTION 5.3)
// ============================================================================

let currentDocsModalVarID = null;

async function openVariationDocsModal(varID) {
    currentDocsModalVarID = varID;
    const modal = document.getElementById("modal-variation-docs");
    const tbody = document.getElementById("var-docs-tbody");
    if (!modal || !tbody) return;

    tbody.innerHTML = `<tr><td colspan="5" style="text-align:center; padding:20px; color:var(--text-muted);"><span class="spin">⏳</span> Chargement des 12 pièces...</td></tr>`;
    openModal("modal-variation-docs");

    try {
        const resp = await fetch(`/api/v1/variations/${varID}`);
        if (!resp.ok) throw new Error("Variation introuvable");
        const data = await resp.json();
        const v = data.variation;

        document.getElementById("var-docs-title").innerText = `${v.tracking_number} — ${v.medicine_name}`;

        tbody.innerHTML = "";
        (v.documents || []).forEach(doc => {
            const tr = document.createElement("tr");

            let statusHTML = `<span style="color:#dc2626; font-weight:700; font-size:11px;">🔴 En Attente</span>`;
            let fileInfo = `<span style="color:var(--text-dim); font-size:10.5px;">Aucune pièce rattachée</span>`;
            if (doc.is_attached) {
                statusHTML = `<span style="color:#047857; font-weight:800; font-size:11px;">✔ Conforme</span>`;
                fileInfo = `<code style="font-size:10px; color:#0369a1;">${escapeHTML(doc.ged_file_name || 'Scan_GED.pdf')}</code>`;
            }

            tr.innerHTML = `
                <td style="font-weight:800; text-align:center; width:30px;">${doc.doc_index}</td>
                <td>
                    <div style="font-weight:700; font-size:11.5px; color:var(--text-main);">${escapeHTML(doc.title)}</div>
                    <div style="font-size:10px; color:var(--text-dim); margin-top:2px;">${escapeHTML(doc.description)}</div>
                </td>
                <td style="font-size:11px;">${fileInfo}</td>
                <td style="text-align:center;">${statusHTML}</td>
                <td style="text-align:center;">
                    <button class="btn-pill" style="font-size:10px; padding:2px 8px;" onclick="promptAttachGEDDoc('${doc.id}')">📎 Joindre GED</button>
                </td>
            `;
            tbody.appendChild(tr);
        });

    } catch (err) {
        tbody.innerHTML = `<tr><td colspan="5" style="text-align:center; padding:20px; color:var(--accent-red);">Erreur: ${escapeHTML(err.message)}</td></tr>`;
    }
}

async function promptAttachGEDDoc(docID) {
    try {
        const resp = await fetch('/api/v1/ged/documents');
        if (!resp.ok) throw new Error("Erreur GED");
        const data = await resp.json();
        const docs = data.documents || [];

        if (docs.length === 0) {
            showNotifyModal("Coffre-fort GED", "Aucun document n'est disponible dans le coffre-fort. Téléversez d'abord une pièce.", "ℹ️");
            return;
        }

        // Sélectionner le premier ou demander confirmation
        const selectedDoc = docs[0];
        const attachResp = await fetch(`/api/v1/variations/${currentDocsModalVarID}/documents/${docID}/attach`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ ged_document_id: selectedDoc.id })
        });

        if (attachResp.ok) {
            openVariationDocsModal(currentDocsModalVarID);
        } else {
            showNotifyModal("Erreur", "Échec du rattachement", "❌");
        }
    } catch (e) {
        showNotifyModal("Erreur GED", e.message, "❌");
    }
}

// ============================================================================
// 5. SCELLEMENT PARAPHEUR DT (21 CFR PART 11)
// ============================================================================

let currentSealVarID = null;

function openSealDTModal(varID) {
    currentSealVarID = varID;
    document.getElementById("seal-dt-solemn-check").checked = false;
    document.getElementById("seal-dt-notes").value = "";
    openModal("modal-seal-variation-dt");
}

async function submitSealVariationDT(event) {
    event.preventDefault();
    const solemn = document.getElementById("seal-dt-solemn-check").checked;
    const notes = document.getElementById("seal-dt-notes").value.trim();

    if (!solemn) {
        showNotifyModal("Engagement Obligatoire", "La déclaration solennelle du Pharmacien Directeur Technique est requise par la réglementation.", "⚠️");
        return;
    }

    try {
        const resp = await fetch(`/api/v1/variations/${currentSealVarID}/seal`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                solemn_declaration: true,
                notes: notes
            })
        });

        if (!resp.ok) {
            const err = await resp.json();
            throw new Error(err.error || "Erreur scellement");
        }

        const data = await resp.json();
        closeModal("modal-seal-variation-dt");
        showNotifyModal(
            "Demande de Variation Scellée (21 CFR Part 11)",
            `Le dossier a été officiellement scellé par le Pharmacien DT.<br>Sceau cryptographique : <code>${escapeHTML(data.seal)}</code>.<br>L'état officiel <strong>ES-REG-04</strong> est désormais généré sans filigrane brouillon.`,
            "✔"
        );
        renderVariationsTable();

    } catch (e) {
        showNotifyModal("Erreur", e.message, "❌");
    }
}

// ============================================================================
// 6. VISIONNEUSE IN-APP & IMPRESSION TABLEAU COMPARATIF ES-REG-04
// ============================================================================

let activeVariationReportUrl = null;

function openComparativeTableModal(varID) {
    if (!varID) return;
    activeVariationReportUrl = `/api/v1/variations/${varID}/comparative-table`;
    
    const iframe = document.getElementById("variation-report-iframe");
    if (iframe) {
        iframe.src = activeVariationReportUrl;
    }
    
    const subTitle = document.getElementById("variation-report-subtitle");
    if (subTitle) {
        subTitle.textContent = `Ligne Directrice ANPP LD-DEPP07/V01 • Document Opposable A4 (300+ DPI)`;
    }
    
    if (typeof openModal === 'function') {
        openModal("modal-variation-report-viewer");
    } else {
        const el = document.getElementById("modal-variation-report-viewer");
        if (el) {
            el.classList.add("open");
            el.style.display = "flex";
        }
    }
}

// Alias de compatibilité pour tous les boutons existants
function printComparativeTable(varID) {
    openComparativeTableModal(varID);
}

function printActiveVariationReport() {
    const iframe = document.getElementById("variation-report-iframe");
    if (iframe && iframe.contentWindow) {
        iframe.contentWindow.focus();
        iframe.contentWindow.print();
    } else if (activeVariationReportUrl) {
        window.open(activeVariationReportUrl, '_blank');
    }
}

function openActiveVariationReportInNewTab() {
    if (activeVariationReportUrl) {
        window.open(activeVariationReportUrl, '_blank');
    }
}

function closeVariationReportModal() {
    if (typeof closeModal === 'function') {
        closeModal("modal-variation-report-viewer");
    } else {
        const el = document.getElementById("modal-variation-report-viewer");
        if (el) {
            el.classList.remove("open");
            el.style.display = "none";
        }
    }
    const iframe = document.getElementById("variation-report-iframe");
    if (iframe) {
        iframe.src = "about:blank";
    }
}
