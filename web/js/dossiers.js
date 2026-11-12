// =========================================================================
// JALON 1.5 : MOTEUR DE DOSSIERS ANPP & CHECKLISTS DYNAMIQUES CTD
// =========================================================================

// =========================================================================
        // JALON 1.5 : MOTEUR DE DOSSIERS ANPP & CHECKLISTS DYNAMIQUES CTD
        // =========================================================================

        function escapeHTML(str) {
            if (str === null || str === undefined) return '';
            return String(str)
                .replace(/&/g, '&amp;')
                .replace(/</g, '&lt;')
                .replace(/>/g, '&gt;')
                .replace(/"/g, '&quot;')
                .replace(/'/g, '&#39;');
        }

        let currentDossiers = [];
        let selectedDossierId = null;
        let currentDossierDetail = null;
        let currentDossierFilterProc = '';
        let currentDossierModuleFilter = 'all'; // 'all', 1, 2, 3, 4, 5, 0
        let currentDossierSubModuleFilter = 'all'; // 'all', '1.A', '1.H', '1.M', '1.P', 'DM.A', 'DM.Q', 'DM.T', 'DM.S', 'DM.I'
        let currentDossierStatusFilter = 'all'; // 'all', 'missing', 'attached', 'mandatory'
        let currentDossierSearchQuery = '';
        let selectedDossierItemId = null;
        let expandedDossierTreeIds = new Set();
        let collapsedSections = new Set();
        let currentPreviewSubTab = 'doc'; // 'doc' ou 'anpp'

        function formatFileSize(bytes) {
            if (!bytes || bytes === 0) return '0 B';
            const k = 1024;
            const sizes = ['B', 'Ko', 'Mo', 'Go'];
            const i = Math.floor(Math.log(bytes) / Math.log(k));
            return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
        }

        function getItemSectionInfo(item) {
            const code = (item && item.item_code) ? item.item_code.trim() : '';
            if (code.startsWith('1.A')) return { key: '1.A', title: '1.A Renseignements Administratifs & Demandeur', order: 10 };
            if (code.startsWith('1.H')) return { key: '1.H', title: '1.H Données Fabricants & Certificats BPF', order: 20 };
            if (code.startsWith('1.M')) return { key: '1.M', title: '1.M Informations sur le Produit & Échantillons', order: 30 };
            if (code.startsWith('1.P')) return { key: '1.P', title: '1.P Documents Spécifiques & Prix Réglementé', order: 40 };
            if (code.startsWith('2.')) return { key: '2.S', title: 'Module 2 : Résumés CTD & Overviews Réglementaires', order: 50 };
            if (code.startsWith('3.2.S')) return { key: '3.2.S', title: '3.2.S Substance Active (API / Principe Actif)', order: 60 };
            if (code.startsWith('3.2.P')) return { key: '3.2.P', title: '3.2.P Produit Fini (Drug Product)', order: 70 };
            if (code.startsWith('3.2.A')) return { key: '3.2.A', title: '3.2.A Installations & Équipements', order: 80 };
            if (code.startsWith('3.2.R')) return { key: '3.2.R', title: '3.2.R Documents Régionaux Qualité', order: 90 };
            if (code.startsWith('3.')) return { key: '3.X', title: 'Module 3 : Qualité CMC', order: 95 };
            if (code.startsWith('4.')) return { key: '4.X', title: 'Module 4 : Rapports d\'Études Non-Cliniques', order: 100 };
            if (code.startsWith('5.')) return { key: '5.X', title: 'Module 5 : Rapports d\'Études Cliniques & BE', order: 110 };
            if (code.startsWith('DM.A')) return { key: 'DM.A', title: 'DM.A Dossier Administratif & Statutaire', order: 200 };
            if (code.startsWith('DM.Q')) return { key: 'DM.Q', title: 'DM.Q Système Qualité (SMQ) ISO 13485', order: 210 };
            if (code.startsWith('DM.T')) return { key: 'DM.T', title: 'DM.T Dossier Technique & Évaluation de Conformité', order: 220 };
            if (code.startsWith('DM.S')) return { key: 'DM.S', title: 'DM.S Sécurité Préclinique & Évaluation Clinique', order: 230 };
            if (code.startsWith('DM.I')) return { key: 'DM.I', title: 'DM.I Étiquetage, Notice & Informations Fabricant', order: 240 };
            return { key: 'AUTRE', title: 'Autres exigences réglementaires', order: 999 };
        }

        async function loadDossiers(procFilter = '') {
            currentDossierFilterProc = procFilter;
            const treeRoot = document.getElementById("tree-root");
            if (treeRoot && currentDomain === 'ctd') {
                treeRoot.innerHTML = `<div style="padding:20px; text-align:center; color:var(--text-muted); font-size:12px;">Chargement des dossiers ANPP...</div>`;
            }

            let url = '/api/v1/dossiers';
            if (procFilter) url += `?procedure_code=${encodeURIComponent(procFilter)}`;

            try {
                const res = await fetch(url);
                if (res.status === 401) {
                    handleAuthExpired();
                    return;
                }
                if (!res.ok) throw new Error("HTTP " + res.status);
                const data = await res.json();
                currentDossiers = data.dossiers || [];

                if (currentDossiers.length > 0) {
                    const exists = currentDossiers.some(d => d.id === selectedDossierId);
                    if (!exists) {
                        selectedDossierId = currentDossiers[0].id;
                    }
                    expandedDossierTreeIds.add(selectedDossierId);
                    renderDossiersTree();
                    await selectDossier(selectedDossierId);
                } else {
                    selectedDossierId = null;
                    currentDossierDetail = null;
                    renderDossiersTree();
                    renderEmptyDossierView();
                }
            } catch (err) {
                if (treeRoot) {
                    treeRoot.innerHTML = `<div style="padding:20px; text-align:center; color:var(--accent-red); font-size:12px;">Erreur de chargement : ${err.message}</div>`;
                }
            }
        }

        function toggleDossierTreeItem(event, dossierId) {
            if (event) event.stopPropagation();
            if (expandedDossierTreeIds.has(dossierId)) {
                expandedDossierTreeIds.delete(dossierId);
            } else {
                expandedDossierTreeIds.add(dossierId);
            }
            renderDossiersTree();
        }

        function renderDossiersTree() {
            const container = document.getElementById("tree-root");
            if (!container) return;

            let html = `
                <div style="padding:10px; border-bottom:1px solid var(--border-color); display:flex; flex-direction:column; gap:8px;">
                    <button class="btn-pill" style="width:100%; background:var(--accent-blue); color:white; border:none; justify-content:center; font-weight:600;" onclick="openNewDossierModal()">
                        ➕ Nouveau Dossier ANPP
                    </button>
                    <div style="display:flex; gap:6px; overflow-x:auto; padding-bottom:4px; -webkit-overflow-scrolling:touch;">
                        <button class="btn-pill ${currentDossierFilterProc === '' ? 'active' : ''}" style="height:26px; font-size:11px; font-weight:700; padding:2px 8px; flex-shrink:0;" onclick="loadDossiers('')">Tous (${currentDossiers.length})</button>
                        <button class="btn-pill ${currentDossierFilterProc === 'NI' ? 'active' : ''}" style="height:26px; font-size:11px; font-weight:700; padding:2px 8px; flex-shrink:0;" onclick="loadDossiers('NI')">NI</button>
                        <button class="btn-pill ${currentDossierFilterProc === 'NF' ? 'active' : ''}" style="height:26px; font-size:11px; font-weight:700; padding:2px 8px; flex-shrink:0;" onclick="loadDossiers('NF')">NF</button>
                        <button class="btn-pill ${currentDossierFilterProc === 'RI' ? 'active' : ''}" style="height:26px; font-size:11px; font-weight:700; padding:2px 8px; flex-shrink:0;" onclick="loadDossiers('RI')">RI/RF</button>
                        <button class="btn-pill ${currentDossierFilterProc === 'HOMOLOG_DM' ? 'active' : ''}" style="height:26px; font-size:11px; font-weight:700; padding:2px 8px; flex-shrink:0;" onclick="loadDossiers('HOMOLOG_DM')">DM</button>
                    </div>
                </div>
                <div style="padding:8px; overflow-y:auto; flex:1;">
            `;

            if (currentDossiers.length === 0) {
                html += `<div style="padding:20px; text-align:center; color:var(--text-muted); font-size:11px;">Aucun dossier trouvé dans cette sélection.</div>`;
            } else {
                currentDossiers.forEach(d => {
                    const isSel = d.id === selectedDossierId;
                    const isExp = expandedDossierTreeIds.has(d.id);
                    const pCode = (d.procedure_code || 'NI').toLowerCase();
                    const pBadgeClass = `badge-proc badge-proc-${pCode}`;
                    
                    let prodName = "Produit non lié";
                    if (d.medicine && d.medicine.brand_name_fr) {
                        prodName = escapeHTML(d.medicine.brand_name_fr);
                    } else if (d.medical_device && d.medical_device.commercial_name_fr) {
                        prodName = escapeHTML(d.medical_device.commercial_name_fr);
                    }

                    const pct = d.completeness_percent || 0;
                    let gaugeClass = "low";
                    if (pct >= 100) gaugeClass = "high";
                    else if (pct >= 50) gaugeClass = "mid";

                    let statusLabel = formatDossierStatusBadge(d.status);

                    html += `
                        <div class="tree-dossier-root ${isSel ? 'selected' : ''} ${isExp ? 'expanded' : ''}" id="tree-dossier-${d.id}">
                            <div class="tree-dossier-header" onclick="selectDossier('${d.id}')">
                                <div class="tree-dossier-toprow">
                                    <div class="tree-dossier-code">
                                        <span onclick="toggleDossierTreeItem(event, '${d.id}')" style="cursor:pointer; display:inline-block; width:14px; text-align:center; color:var(--text-dim);" title="${isExp ? 'Replier l\'arbre' : 'Déplier l\'arbre'}">${isExp ? '▼' : '▶'}</span>
                                        <span class="${pBadgeClass}">${escapeHTML(d.procedure_code)}</span>
                                        <span>${escapeHTML(d.dossier_code)}</span>
                                    </div>
                                    ${statusLabel}
                                </div>
                                <div class="tree-dossier-product" title="${prodName}">
                                    ${prodName}
                                </div>
                                <div style="display:flex; justify-content:space-between; font-size:10px; color:var(--text-dim); margin-top:2px;">
                                    <span>Complétude</span>
                                    <strong style="color:var(--text-main);">${pct}%</strong>
                                </div>
                                <div class="kpi-gauge-bar" style="margin-top:2px; height:4px;">
                                    <div class="kpi-gauge-fill ${gaugeClass}" style="width:${pct}%;"></div>
                                </div>
                            </div>
                            <div class="tree-dossier-branches">
                                ${renderTreeBranchesForDossier(d)}
                            </div>
                        </div>
                    `;
                });
            }

            html += `</div>`;
            container.innerHTML = html;
        }

        function renderTreeBranchesForDossier(d) {
            const isMedicine = d.product_type === 'medicament';
            const isSelDossier = d.id === selectedDossierId;
            const items = (isSelDossier && currentDossierDetail && currentDossierDetail.items) ? currentDossierDetail.items : (d.items || []);

            // Calcul des ratios avec pastille haute visibilité
            function getCounts(modNum, subCode) {
                if (!items || items.length === 0) return '';
                let subset = items;
                if (modNum !== null) subset = subset.filter(it => it.module_number === modNum);
                if (subCode) subset = subset.filter(it => (it.item_code || '').startsWith(subCode));
                const total = subset.length;
                if (total === 0) return '';
                const linked = subset.filter(it => it.ged_document_id).length;
                const isComplete = linked === total && total > 0;
                return ` <span class="tree-count-pill ${isComplete ? 'complete' : ''}">${linked}/${total}</span>`;
            }

            let branchesHTML = '';

            if (isMedicine) {
                const isMod1Active = isSelDossier && (currentDossierModuleFilter === 1 || currentDossierModuleFilter === '1');
                const isMod2Active = isSelDossier && (currentDossierModuleFilter === 2 || currentDossierModuleFilter === '2');
                const isMod3Active = isSelDossier && (currentDossierModuleFilter === 3 || currentDossierModuleFilter === '3');
                const isMod4Active = isSelDossier && (currentDossierModuleFilter === 4 || currentDossierModuleFilter === '4');
                const isMod5Active = isSelDossier && (currentDossierModuleFilter === 5 || currentDossierModuleFilter === '5');

                branchesHTML += `
                    <div class="tree-module-item ${isMod1Active && currentDossierSubModuleFilter === 'all' ? 'active' : ''}" onclick="selectDossierModuleTree('${d.id}', 1, 'all')">
                        <span>📁 M1 : Spécificités Algérie</span>
                        ${getCounts(1, '')}
                    </div>
                    <div class="tree-submodule-item ${isMod1Active && currentDossierSubModuleFilter === '1.A' ? 'active' : ''}" onclick="selectDossierModuleTree('${d.id}', 1, '1.A')">
                        <span>└ 1.A Demandeur &amp; Admin</span>
                        ${getCounts(1, '1.A')}
                    </div>
                    <div class="tree-submodule-item ${isMod1Active && currentDossierSubModuleFilter === '1.H' ? 'active' : ''}" onclick="selectDossierModuleTree('${d.id}', 1, '1.H')">
                        <span>└ 1.H Fabricant &amp; BPF</span>
                        ${getCounts(1, '1.H')}
                    </div>
                    <div class="tree-submodule-item ${isMod1Active && currentDossierSubModuleFilter === '1.M' ? 'active' : ''}" onclick="selectDossierModuleTree('${d.id}', 1, '1.M')">
                        <span>└ 1.M Info Produit &amp; Éch.</span>
                        ${getCounts(1, '1.M')}
                    </div>
                    <div class="tree-submodule-item ${isMod1Active && currentDossierSubModuleFilter === '1.P' ? 'active' : ''}" onclick="selectDossierModuleTree('${d.id}', 1, '1.P')">
                        <span>└ 1.P Docs Spécifiques Algérie</span>
                        ${getCounts(1, '1.P')}
                    </div>

                    <div class="tree-module-item ${isMod2Active ? 'active' : ''}" onclick="selectDossierModuleTree('${d.id}', 2, 'all')">
                        <span>📁 M2 : Résumés CTD</span>
                        ${getCounts(2, '')}
                    </div>

                    <div class="tree-module-item ${isMod3Active && currentDossierSubModuleFilter === 'all' ? 'active' : ''}" onclick="selectDossierModuleTree('${d.id}', 3, 'all')">
                        <span>📁 M3 : Qualité CMC</span>
                        ${getCounts(3, '')}
                    </div>
                    <div class="tree-submodule-item ${isMod3Active && currentDossierSubModuleFilter === '3.2.S' ? 'active' : ''}" onclick="selectDossierModuleTree('${d.id}', 3, '3.2.S')">
                        <span>└ 3.2.S Substance Active (API)</span>
                        ${getCounts(3, '3.2.S')}
                    </div>
                    <div class="tree-submodule-item ${isMod3Active && currentDossierSubModuleFilter === '3.2.P' ? 'active' : ''}" onclick="selectDossierModuleTree('${d.id}', 3, '3.2.P')">
                        <span>└ 3.2.P Produit Fini</span>
                        ${getCounts(3, '3.2.P')}
                    </div>

                    <div class="tree-module-item ${isMod4Active ? 'active' : ''}" onclick="selectDossierModuleTree('${d.id}', 4, 'all')">
                        <span>📁 M4 : Non-Clinique</span>
                        ${getCounts(4, '')}
                    </div>

                    <div class="tree-module-item ${isMod5Active ? 'active' : ''}" onclick="selectDossierModuleTree('${d.id}', 5, 'all')">
                        <span>📁 M5 : Clinique &amp; BE</span>
                        ${getCounts(5, '')}
                    </div>
                `;
            } else {
                const isDMActive = isSelDossier && (currentDossierModuleFilter === 0 || currentDossierModuleFilter === '0' || currentDossierModuleFilter === 'all');
                branchesHTML += `
                    <div class="tree-module-item ${isDMActive && currentDossierSubModuleFilter === 'all' ? 'active' : ''}" onclick="selectDossierModuleTree('${d.id}', 0, 'all')">
                        <span>📁 Homologation DM</span>
                        ${getCounts(0, '')}
                    </div>
                    <div class="tree-submodule-item ${isDMActive && currentDossierSubModuleFilter === 'DM.A' ? 'active' : ''}" onclick="selectDossierModuleTree('${d.id}', 0, 'DM.A')">
                        <span>└ DM.A Administratif</span>
                        ${getCounts(0, 'DM.A')}
                    </div>
                    <div class="tree-submodule-item ${isDMActive && currentDossierSubModuleFilter === 'DM.Q' ? 'active' : ''}" onclick="selectDossierModuleTree('${d.id}', 0, 'DM.Q')">
                        <span>└ DM.Q SMQ ISO 13485</span>
                        ${getCounts(0, 'DM.Q')}
                    </div>
                    <div class="tree-submodule-item ${isDMActive && currentDossierSubModuleFilter === 'DM.T' ? 'active' : ''}" onclick="selectDossierModuleTree('${d.id}', 0, 'DM.T')">
                        <span>└ DM.T Dossier Technique</span>
                        ${getCounts(0, 'DM.T')}
                    </div>
                    <div class="tree-submodule-item ${isDMActive && currentDossierSubModuleFilter === 'DM.S' ? 'active' : ''}" onclick="selectDossierModuleTree('${d.id}', 0, 'DM.S')">
                        <span>└ DM.S Sécurité &amp; Clinique</span>
                        ${getCounts(0, 'DM.S')}
                    </div>
                    <div class="tree-submodule-item ${isDMActive && currentDossierSubModuleFilter === 'DM.I' ? 'active' : ''}" onclick="selectDossierModuleTree('${d.id}', 0, 'DM.I')">
                        <span>└ DM.I Étiquetage &amp; IFU</span>
                        ${getCounts(0, 'DM.I')}
                    </div>
                `;
            }

            return branchesHTML;
        }

        async function selectDossierModuleTree(dossierId, modNumber, subModCode = 'all') {
            currentDossierModuleFilter = modNumber;
            currentDossierSubModuleFilter = subModCode;

            if (selectedDossierId !== dossierId) {
                await selectDossier(dossierId);
            } else {
                renderDossiersTree();
                if (currentDossierDetail) {
                    renderDossierCenterView(currentDossierDetail);
                }
            }
        }

        function formatDossierStatusBadge(st) {
            switch(st) {
                case 'brouillon':
                    return '<span class="status-tag" style="background:rgba(148,163,184,0.15); color:#94a3b8;">BROUILLON</span>';
                case 'en_preparation':
                    return '<span class="status-tag status-rev">EN PRÉPARATION</span>';
                case 'en_revue':
                    return '<span class="status-tag" style="background:rgba(245,158,11,0.15); color:#f59e0b;">EN REVUE</span>';
                case 'valide_interne':
                    return '<span class="status-tag" style="background:rgba(16,185,129,0.15); color:#10b981; border:1px solid rgba(16,185,129,0.3);">✅ VALIDÉ PDT</span>';
                case 'depose_anpp':
                    return '<span class="status-tag status-ok">🏛️ DÉPOSÉ ANPP</span>';
                case 'approuve_anpp':
                    return '<span class="status-tag" style="background:rgba(2,132,199,0.15); color:#0284c7; border:1px solid rgba(2,132,199,0.3);">DECISION ENREGISTRÉ</span>';
                case 'rejete_anpp':
                    return '<span class="status-tag status-rej">REFUSÉ ANPP</span>';
                default:
                    return `<span class="status-tag">${st || 'INCONNU'}</span>`;
            }
        }

        async function selectDossier(id) {
            selectedDossierId = id;
            expandedDossierTreeIds.add(id);
            renderDossiersTree();

            if (window.innerWidth < 768) {
                toggleMobileTree(false);
            }

            try {
                const res = await fetch(`/api/v1/dossiers/${id}`);
                if (!res.ok) throw new Error("HTTP " + res.status);
                currentDossierDetail = await res.json();
                
                // Mettre à jour fil d'ariane
                document.getElementById("bc-root").innerText = "DOSSIERS_ANPP";
                document.getElementById("bc-sec").innerText = currentDossierDetail.procedure_code || "PROC";
                document.getElementById("bc-prod").innerText = currentDossierDetail.dossier_code;
                
                let prodTitle = "Produit";
                if (currentDossierDetail.medicine) {
                    prodTitle = currentDossierDetail.medicine.brand_name_fr;
                } else if (currentDossierDetail.medical_device) {
                    prodTitle = currentDossierDetail.medical_device.commercial_name_fr;
                }
                document.getElementById("bc-mod").innerText = prodTitle;

                // KPI toolbar
                const kpiEl = document.getElementById("kpi-dossier-completeness");
                if (kpiEl) {
                    kpiEl.innerText = `${currentDossierDetail.completeness_percent}%`;
                    kpiEl.style.color = currentDossierDetail.completeness_percent === 100 ? 'var(--accent-emerald)' : 'var(--accent-blue)';
                }

                // Ajustement des boutons de la toolbar
                const btnSignPDT = document.getElementById("tb-btn-sign-pdt");
                if (btnSignPDT) {
                    if (currentDossierDetail.pdt_signed_at) {
                        btnSignPDT.innerHTML = "✅ Visa PDT Apposé";
                        btnSignPDT.style.background = "rgba(16,185,129,0.15)";
                        btnSignPDT.style.color = "var(--accent-emerald)";
                        btnSignPDT.style.borderColor = "rgba(16,185,129,0.4)";
                        btnSignPDT.disabled = true;
                    } else if (currentDossierDetail.completeness_percent === 100) {
                        btnSignPDT.innerHTML = "✍️ Apposer Visa PDT (100%)";
                        btnSignPDT.style.background = "var(--accent-emerald)";
                        btnSignPDT.style.color = "white";
                        btnSignPDT.style.border = "none";
                        btnSignPDT.disabled = false;
                    } else {
                        btnSignPDT.innerHTML = "✍️ Visa PDT (Incomplet)";
                        btnSignPDT.style.background = "transparent";
                        btnSignPDT.style.color = "var(--text-muted)";
                        btnSignPDT.disabled = false;
                    }
                }

                // Mettre à jour l'arbre avec les compteurs réels
                renderDossiersTree();
                renderDossierCenterView(currentDossierDetail);
            } catch(e) {
                console.error("selectDossier error:", e);
            }
        }

        function handleChecklistSearch(query) {
            currentDossierSearchQuery = (query || '').trim().toLowerCase();
            if (currentDossierDetail) {
                const items = currentDossierDetail.items || currentDossierDetail.checklist_data || [];
                renderDossierTableRows(items);
            }
        }

        function setQuickStatusFilter(status) {
            currentDossierStatusFilter = status;
            document.querySelectorAll('.pill-filter').forEach(p => p.classList.remove('active'));
            const activePill = document.getElementById(`pill-filter-${status}`);
            if (activePill) activePill.classList.add('active');

            if (currentDossierDetail) {
                const items = currentDossierDetail.items || currentDossierDetail.checklist_data || [];
                renderDossierTableRows(items);
            }
        }

        function toggleSectionCollapse(sectionKey) {
            if (collapsedSections.has(sectionKey)) {
                collapsedSections.delete(sectionKey);
            } else {
                collapsedSections.add(sectionKey);
            }
            if (currentDossierDetail) {
                const items = currentDossierDetail.items || currentDossierDetail.checklist_data || [];
                renderDossierTableRows(items);
            }
        }

        function filterDossierModule(mod) {
            currentDossierModuleFilter = mod;
            currentDossierSubModuleFilter = 'all';
            
            document.querySelectorAll('.module-tab-btn').forEach(btn => btn.classList.remove('active'));
            if (event && event.currentTarget) {
                event.currentTarget.classList.add('active');
            }

            renderDossiersTree();
            if (currentDossierDetail) {
                const items = currentDossierDetail.items || currentDossierDetail.checklist_data || [];
                renderDossierTableRows(items);
            }
        }

        function renderDossierCenterView(dossier) {
            const tableContainer = document.querySelector('.table-container');
            if (!tableContainer) return;

            let prodName = "N/A";
            let prodSub = "";
            let isMedicine = dossier.product_type === 'medicament';

            if (dossier.medicine) {
                prodName = escapeHTML(dossier.medicine.brand_name_fr);
                prodSub = `DCI: <strong>${escapeHTML(dossier.medicine.dci_fr || 'N/A')}</strong> | Dosage: <strong>${escapeHTML(dossier.medicine.dosage_fr || '')}</strong> | Forme: <strong>${escapeHTML(dossier.medicine.pharmaceutical_form_fr || '')}</strong>`;
            } else if (dossier.medical_device) {
                prodName = escapeHTML(dossier.medical_device.commercial_name_fr);
                prodSub = `Classe: <strong>Classe ${escapeHTML(dossier.medical_device.dm_class)}</strong> | Stérile: <strong>${dossier.medical_device.is_sterile ? 'Oui (ISO 11135)' : 'Non'}</strong> | Réf: <strong>${escapeHTML(dossier.medical_device.manufacturer_reference || '')}</strong>`;
            }

            const pCode = (dossier.procedure_code || 'NI').toLowerCase();
            const pBadgeClass = `badge-proc badge-proc-${pCode}`;
            const pct = dossier.completeness_percent || 0;

            const items = dossier.items || dossier.checklist_data || [];
            let totalMandatory = 0;
            let fulfilledMandatory = 0;
            items.forEach(it => {
                if (it.is_mandatory && it.requirement_level === 'E') {
                    totalMandatory++;
                    if (it.status === 'rattache' || it.status === 'valide_qualite' || it.status === 'approuve_dt') {
                        fulfilledMandatory++;
                    }
                }
            });

            // Décompte par module
            const countByModule = {};
            items.forEach(it => {
                const k = it.module_number;
                countByModule[k] = (countByModule[k] || 0) + 1;
            });

            let tabsHTML = `
                <div class="module-tab-bar">
                    <button class="module-tab-btn ${currentDossierModuleFilter === 'all' ? 'active' : ''}" onclick="filterDossierModule('all')">
                        Toutes les exigences (${items.length})
                    </button>
            `;

            if (isMedicine) {
                tabsHTML += `
                    <button class="module-tab-btn ${currentDossierModuleFilter === 1 || currentDossierModuleFilter === '1' ? 'active' : ''}" onclick="filterDossierModule(1)">
                        Module 1 : Spécificités Algérie (${countByModule[1] || 0})
                    </button>
                    <button class="module-tab-btn ${currentDossierModuleFilter === 2 || currentDossierModuleFilter === '2' ? 'active' : ''}" onclick="filterDossierModule(2)">
                        Module 2 : Résumés CTD (${countByModule[2] || 0})
                    </button>
                    <button class="module-tab-btn ${currentDossierModuleFilter === 3 || currentDossierModuleFilter === '3' ? 'active' : ''}" onclick="filterDossierModule(3)">
                        Module 3 : Qualité CMC (${countByModule[3] || 0})
                    </button>
                    <button class="module-tab-btn ${currentDossierModuleFilter === 4 || currentDossierModuleFilter === '4' ? 'active' : ''}" onclick="filterDossierModule(4)">
                        Module 4 : Non-Clinique (${countByModule[4] || 0})
                    </button>
                    <button class="module-tab-btn ${currentDossierModuleFilter === 5 || currentDossierModuleFilter === '5' ? 'active' : ''}" onclick="filterDossierModule(5)">
                        Module 5 : Clinique &amp; BE (${countByModule[5] || 0})
                    </button>
                `;
            } else {
                tabsHTML += `
                    <button class="module-tab-btn ${currentDossierModuleFilter === 0 || currentDossierModuleFilter === '0' ? 'active' : ''}" onclick="filterDossierModule(0)">
                        Exigences Homologation DM (${countByModule[0] || 0})
                    </button>
                `;
            }
            tabsHTML += `</div>`;

            // Barre d'actions & filtres rapides
            let quickFilterBarHTML = `
                <div class="table-quick-filter-bar">
                    <div class="search-box-checklist">
                        <span>🔍</span>
                        <input type="text" id="checklist-search-input" placeholder="Filtrer par code (ex: 1.A.1), titre, mot-clé..." value="${escapeHTML(currentDossierSearchQuery)}" oninput="handleChecklistSearch(this.value)">
                    </div>
                    <div class="quick-status-pills">
                        <span style="font-size:10px; color:var(--text-dim); text-transform:uppercase; font-weight:700; margin-right:2px;">Filtre Statut :</span>
                        <div class="pill-filter ${currentDossierStatusFilter === 'all' ? 'active' : ''}" id="pill-filter-all" onclick="setQuickStatusFilter('all')">Tous</div>
                        <div class="pill-filter pill-missing ${currentDossierStatusFilter === 'missing' ? 'active' : ''}" id="pill-filter-missing" onclick="setQuickStatusFilter('missing')">🔴 Manquants</div>
                        <div class="pill-filter pill-attached ${currentDossierStatusFilter === 'attached' ? 'active' : ''}" id="pill-filter-attached" onclick="setQuickStatusFilter('attached')">🟢 Rattachés</div>
                        <div class="pill-filter pill-mandatory ${currentDossierStatusFilter === 'mandatory' ? 'active' : ''}" id="pill-filter-mandatory" onclick="setQuickStatusFilter('mandatory')">⚠️ Obligatoires [E]</div>
                    </div>
                </div>
            `;

            // En-tête KPI Dossier
            let headerKPI = `
                <div style="background:var(--bg-panel); border-bottom:1px solid var(--border-color); padding:12px 16px;">
                    <div style="display:flex; justify-content:space-between; align-items:flex-start; margin-bottom:8px; flex-wrap:wrap; gap:12px;">
                        <div>
                            <div style="display:flex; align-items:center; gap:8px; margin-bottom:4px; flex-wrap:wrap;">
                                <span class="${pBadgeClass}" style="font-size:11px; padding:3px 8px;">${escapeHTML(dossier.procedure_code)}</span>
                                <span style="font-family:monospace; font-size:15px; font-weight:700; color:var(--text-main);">${escapeHTML(dossier.dossier_code)}</span>
                                <span style="font-size:12px; color:var(--text-dim);">(Réf : ${escapeHTML(dossier.internal_reference || 'N/A')})</span>
                                ${formatDossierStatusBadge(dossier.status)}
                            </div>
                            <div style="font-size:13px; font-weight:600; color:var(--accent-blue); margin-bottom:2px;">
                                ${prodName}
                            </div>
                            <div style="font-size:11px; color:var(--text-muted);">
                                ${prodSub}
                            </div>
                        </div>
                        <div style="text-align:right;">
                            <div style="font-size:11px; color:var(--text-muted); margin-bottom:2px;">Complétude Réglementaire BPF</div>
                            <div style="display:flex; align-items:baseline; justify-content:flex-end; gap:6px;">
                                <span style="font-size:20px; font-weight:800; color:${pct === 100 ? 'var(--accent-emerald)' : (pct >= 50 ? 'var(--accent-amber)' : 'var(--accent-red)')};">${pct}%</span>
                                <span style="font-size:11px; color:var(--text-dim);">(${fulfilledMandatory} / ${totalMandatory} exigées)</span>
                            </div>
                            <div style="font-size:10px; color:var(--text-dim); margin-top:2px;">
                                Cible Dépôt : <strong>${dossier.target_deposit_date ? dossier.target_deposit_date.substring(0, 10) : 'Non fixée'}</strong>
                            </div>
                        </div>
                    </div>

                    <!-- Barre de progression -->
                    <div class="kpi-gauge-bar" style="height:6px;">
                        <div class="kpi-gauge-fill ${pct === 100 ? 'high' : (pct >= 50 ? 'mid' : 'low')}" style="width:${pct}%;"></div>
                    </div>

                    ${dossier.pdt_signed_at ? `
                        <div style="margin-top:8px; background:rgba(16,185,129,0.08); border:1px solid rgba(16,185,129,0.3); border-radius:4px; padding:6px 10px; display:flex; justify-content:space-between; align-items:center; font-size:11px;">
                            <div style="display:flex; align-items:center; gap:6px; color:var(--accent-emerald); font-weight:600;">
                                <span>✍️ Visa Officiel Pharmacien Directeur Technique apposé</span>
                                <span style="font-size:10px; color:var(--text-dim); font-weight:normal;">le ${new Date(dossier.pdt_signed_at).toLocaleString()}</span>
                            </div>
                            <div style="font-family:monospace; font-size:10px; color:var(--text-dim);" title="${dossier.pdt_signature_seal}">
                                Sceau SHA-256 : ${dossier.pdt_signature_seal ? dossier.pdt_signature_seal.substring(0, 16) + '...' : ''}
                            </div>
                        </div>
                    ` : ''}

                    ${dossier.deposit_receipt_ref ? `
                        <div style="margin-top:6px; background:rgba(2,132,199,0.08); border:1px solid rgba(2,132,199,0.3); border-radius:4px; padding:6px 10px; display:flex; justify-content:space-between; align-items:center; font-size:11px;">
                            <div style="color:var(--accent-blue); font-weight:600;">
                                🏛️ Récépissé Dépôt ANPP : <code style="color:var(--text-main);">${escapeHTML(dossier.deposit_receipt_ref)}</code> (${dossier.deposit_receipt_date ? escapeHTML(dossier.deposit_receipt_date.substring(0, 10)) : ''})
                            </div>
                            ${dossier.anpp_official_reference ? `<div style="font-size:10px; color:var(--text-dim);">N° Dossier ANPP : <strong>${escapeHTML(dossier.anpp_official_reference)}</strong></div>` : ''}
                        </div>
                    ` : ''}
                </div>
            `;

            tableContainer.innerHTML = `
                ${headerKPI}
                ${tabsHTML}
                ${quickFilterBarHTML}
                <div style="overflow-y:auto; flex:1;">
                    <table id="doc-table">
                        <thead id="table-head">
                            <tr>
                                <th class="col-hide-xs" style="width: 75px;">Module</th>
                                <th class="col-hide-sm" style="width: 80px;">Exigence</th>
                                <th style="width: 90px;">Statut</th>
                                <th>Code &amp; Intitulé Réglementaire (Guide ANPP)</th>
                                <th>Document GED Rattaché (Coffre-Fort BPF)</th>
                                <th style="width: 140px; text-align:center;">Actions</th>
                            </tr>
                        </thead>
                        <tbody id="doc-tbody">
                            <!-- Lignes remplies par renderDossierTableRows -->
                        </tbody>
                    </table>
                </div>
            `;

            renderDossierTableRows(items);
        }

        function renderDossierTableRows(items) {
            const tbody = document.getElementById("doc-tbody");
            if (!tbody) return;

            // 1. Filtrage par module
            let filtered = items;
            if (currentDossierModuleFilter !== 'all') {
                const targetMod = Number(currentDossierModuleFilter);
                filtered = filtered.filter(it => it.module_number === targetMod);
            }

            // 2. Filtrage par sous-module (si spécifié dans l'arbre)
            if (currentDossierSubModuleFilter !== 'all') {
                filtered = filtered.filter(it => (it.item_code || '').startsWith(currentDossierSubModuleFilter));
            }

            // 3. Filtrage par statut rapide
            if (currentDossierStatusFilter === 'missing') {
                filtered = filtered.filter(it => !it.ged_document_id);
            } else if (currentDossierStatusFilter === 'attached') {
                filtered = filtered.filter(it => !!it.ged_document_id);
            } else if (currentDossierStatusFilter === 'mandatory') {
                filtered = filtered.filter(it => it.is_mandatory && it.requirement_level === 'E');
            }

            // 4. Filtrage par recherche textuelle
            if (currentDossierSearchQuery) {
                const q = currentDossierSearchQuery;
                filtered = filtered.filter(it => {
                    const c = (it.item_code || '').toLowerCase();
                    const tf = (it.title_fr || '').toLowerCase();
                    const df = (it.description_fr || '').toLowerCase();
                    const ta = (it.title_ar || '').toLowerCase();
                    return c.includes(q) || tf.includes(q) || df.includes(q) || ta.includes(q);
                });
            }

            if (filtered.length === 0) {
                tbody.innerHTML = `<tr><td colspan="6" style="text-align:center; padding:30px; color:var(--text-muted);">
                    <div style="font-size:20px; margin-bottom:6px;">🔍</div>
                    <div>Aucun élément ne correspond aux filtres appliqués.</div>
                    <div style="margin-top:8px;">
                        <button class="btn-pill" style="font-size:11px;" onclick="setQuickStatusFilter('all'); handleChecklistSearch('');">Réinitialiser les filtres</button>
                    </div>
                </td></tr>`;
                return;
            }

            // Regroupement par section pour les accordéons
            const sectionsMap = new Map();
            filtered.forEach(it => {
                const sec = getItemSectionInfo(it);
                if (!sectionsMap.has(sec.key)) {
                    sectionsMap.set(sec.key, {
                        key: sec.key,
                        title: sec.title,
                        order: sec.order,
                        items: []
                    });
                }
                sectionsMap.get(sec.key).items.push(it);
            });

            // Tri des sections
            const sortedSections = Array.from(sectionsMap.values()).sort((a, b) => a.order - b.order);

            tbody.innerHTML = "";

            sortedSections.forEach(section => {
                const isCollapsed = collapsedSections.has(section.key);
                const totalInSec = section.items.length;
                const linkedInSec = section.items.filter(it => it.ged_document_id).length;
                const mandatoryInSec = section.items.filter(it => it.is_mandatory && it.requirement_level === 'E').length;

                // Ligne d'en-tête de section (accordéon)
                const trHeader = document.createElement("tr");
                trHeader.className = "section-accordion-header";
                trHeader.onclick = () => toggleSectionCollapse(section.key);
                trHeader.title = isCollapsed ? "Cliquer pour déplier cette section" : "Cliquer pour replier cette section";

                trHeader.innerHTML = `
                    <td colspan="6">
                        <div class="sec-banner">
                            <span style="font-size:11px; color:var(--text-dim);">${isCollapsed ? '▶' : '▼'}</span>
                            <span style="color:var(--text-main); font-weight:700;">${section.title}</span>
                            <div class="sec-counts">
                                <span style="color:${linkedInSec === totalInSec ? 'var(--accent-emerald)' : 'var(--text-muted)'};">
                                    ${linkedInSec} / ${totalInSec} pièce(s) liée(s)
                                </span>
                                ${mandatoryInSec > 0 ? `<span style="margin-left:8px; color:var(--accent-amber);">[${mandatoryInSec} E]</span>` : ''}
                            </div>
                        </div>
                    </td>
                `;
                tbody.appendChild(trHeader);

                // Si la section n'est pas repliée, afficher ses items
                if (!isCollapsed) {
                    section.items.forEach(it => {
                        const tr = document.createElement("tr");
                        if (it.id === selectedDossierItemId) tr.classList.add("selected");
                        tr.onclick = () => selectDossierItem(it.id);

                        let reqBadge = `<span class="badge-req-e">[E] Exigé</span>`;
                        if (it.requirement_level === 'F') reqBadge = `<span class="badge-req-f">[F] Facultatif</span>`;
                        else if (it.requirement_level === 'P') reqBadge = `<span class="badge-req-p">[P] Partiel</span>`;
                        else if (it.requirement_level === 'NA') reqBadge = `<span class="badge-req-na">[NA] Non Appl.</span>`;

                        let statusBadge = '<span class="status-tag status-rej">MANQUANT</span>';
                        if (it.status === 'rattache') statusBadge = '<span class="status-tag status-rev">RATTACHÉ</span>';
                        else if (it.status === 'valide_qualite') statusBadge = '<span class="status-tag status-ok">VALIDÉ QUALITÉ</span>';
                        else if (it.status === 'approuve_dt') statusBadge = '<span class="status-tag" style="background:rgba(147,51,234,0.15); color:#7e22ce; border:1.5px solid #c084fc; font-weight:800;">APPROUVÉ PDT</span>';

                        let modLabel = `M${it.module_number}`;
                        if (it.module_number === 0) modLabel = `DM`;

                        let gedInfo = `<span style="color:var(--text-dim); font-size:11px; font-weight:600; font-style:italic;">— Non rattaché —</span>`;
                        if (it.ged_document) {
                            gedInfo = `
                                <div style="display:flex; align-items:center; gap:6px;">
                                    <span style="font-size:13px;">📄</span>
                                    <div>
                                        <div style="font-weight:700; color:var(--accent-blue); font-size:11.5px;">${escapeHTML(it.ged_document.title_fr)}</div>
                                        <div style="font-family:monospace; font-size:10.5px; font-weight:600; color:var(--text-dim);">${escapeHTML(it.ged_document.document_code || '')} (${escapeHTML(it.ged_document.current_version)})</div>
                                    </div>
                                </div>
                            `;
                        }

                        let actionsHTML = '';
                        const isLocked = currentDossierDetail && currentDossierDetail.is_locked_for_deposit;

                        if (!it.ged_document_id) {
                            actionsHTML = `
                                <button class="btn-pill" style="height:24px; padding:2px 8px; font-size:10.5px; font-weight:700; background:var(--accent-blue); color:white; border:none;" 
                                        onclick="event.stopPropagation(); openAttachGEDModal('${it.id}', '${it.item_code}', '${escapeHTML(it.title_fr)}', '${it.requirement_level}')"
                                        ${isLocked ? 'disabled title="Dossier verrouillé"' : ''}>
                                    📎 Rattacher GED
                                </button>
                            `;
                        } else {
                            actionsHTML = `
                                <div style="display:flex; gap:4px; justify-content:center;">
                                    <button class="btn-pill" style="height:24px; padding:2px 6px; font-size:10.5px; font-weight:700;" title="Aperçu document" onclick="event.stopPropagation(); previewGEDItemDoc('${it.ged_document_id}')">
                                        👁️
                                    </button>
                                    <button class="btn-pill" style="height:24px; padding:2px 6px; font-size:10.5px; font-weight:700; color:var(--accent-red); background:#fff5f5; border-color:#fca5a5;" title="Détacher" 
                                            onclick="event.stopPropagation(); executeDetachGED('${it.id}', '${it.item_code}')"
                                            ${isLocked ? 'disabled title="Dossier verrouillé"' : ''}>
                                        ✕
                                    </button>
                                </div>
                            `;
                        }

                        tr.innerHTML = `
                            <td class="col-hide-xs" style="font-weight:800; color:var(--text-dim); font-size:11.5px; font-family:monospace; white-space:nowrap;">${modLabel}</td>
                            <td class="col-hide-sm" style="white-space:nowrap;">${reqBadge}</td>
                            <td style="white-space:nowrap;">${statusBadge}</td>
                            <td style="min-width:240px;">
                                <div style="display:flex; align-items:center; gap:6px; flex-wrap:wrap;">
                                    <span style="font-family:monospace; font-weight:800; color:var(--accent-blue); font-size:12px;">${escapeHTML(it.item_code)}</span>
                                    <span style="font-weight:700; color:var(--text-main); font-size:12px;">${escapeHTML(it.title_fr)}</span>
                                </div>
                                ${it.title_ar ? `<div style="font-size:11px; font-weight:600; color:var(--text-main); font-family:tahoma,sans-serif; margin-top:2px;" dir="rtl">${escapeHTML(it.title_ar)}</div>` : ''}
                                ${it.description_fr ? `<div style="font-size:11px; font-weight:500; color:var(--text-muted); margin-top:3px; line-height:1.4;">${escapeHTML(it.description_fr)}</div>` : ''}
                            </td>
                            <td style="min-width:180px;">${gedInfo}</td>
                            <td style="text-align:center; white-space:nowrap;">${actionsHTML}</td>
                        `;
                        tbody.appendChild(tr);
                    });
                }
            });

            // Si aucun item sélectionné dans la liste visible, sélectionner le premier
            if (filtered.length > 0) {
                const currentStillExists = filtered.some(it => it.id === selectedDossierItemId);
                if (!currentStillExists) {
                    selectDossierItem(filtered[0].id);
                }
            }
        }

        function selectDossierItem(itemId) {
            selectedDossierItemId = itemId;
            document.querySelectorAll('#doc-tbody tr').forEach(tr => tr.classList.remove('selected'));
            if (typeof event !== 'undefined' && event && event.currentTarget) {
                event.currentTarget.classList.add('selected');
                if (window.innerWidth < 1400) togglePreviewPane(true);
            }

            if (!currentDossierDetail) return;
            const items = currentDossierDetail.items || currentDossierDetail.checklist_data || [];
            const item = items.find(it => it.id === itemId);
            if (item) {
                renderDossierItemPreview(item);
            }
        }

        function switchPreviewSubTab(tab) {
            currentPreviewSubTab = tab;
            const btnDoc = document.getElementById("pvt-btn-doc");
            const btnAnpp = document.getElementById("pvt-btn-anpp");
            if (btnDoc && btnAnpp) {
                btnDoc.classList.toggle("active", tab === 'doc');
                btnAnpp.classList.toggle("active", tab === 'anpp');
            }

            if (currentDossierDetail && selectedDossierItemId) {
                const items = currentDossierDetail.items || currentDossierDetail.checklist_data || [];
                const item = items.find(it => it.id === selectedDossierItemId);
                if (item) {
                    renderDossierItemPreview(item);
                }
            }
        }

        function renderDossierItemPreview(item) {
            const preview = document.getElementById("preview-body");
            if (!preview) return;

            let reqDesc = "Exigence obligatoire pour le dépôt ANPP.";
            if (item.requirement_level === 'E') reqDesc = "Exigé [E] : La pièce doit être obligatoirement présente dans le dossier sous peine de rejet d'irrecevabilité.";
            else if (item.requirement_level === 'F') reqDesc = "Facultatif [F] : L'Agence Nationale des Produits Pharmaceutiques peut exiger cette pièce à tout moment en cours d'évaluation.";
            else if (item.requirement_level === 'P') reqDesc = "Partiel [P] : Requis selon les caractéristiques du produit (ex: étude bioéquivalence ou justification biowaiver).";

            const isLocked = currentDossierDetail && currentDossierDetail.is_locked_for_deposit;

            if (currentPreviewSubTab === 'doc') {
                // ONGLET 1 : Aperçu Pièce GED (Visualiseur BPF / Coffre-fort certifié)
                if (item.ged_document) {
                    const doc = item.ged_document;
                    preview.innerHTML = `
                        <div class="pdf-sim-page">
                            <div class="pdf-sim-header" style="display:flex; justify-content:space-between; align-items:center;">
                                <span>📄 Pièce GED Certifiée — [${escapeHTML(item.item_code)}]</span>
                                <span class="status-tag status-ok">COFFRE-FORT SÉCURISÉ</span>
                            </div>

                            <div style="margin-top:10px; margin-bottom:12px;">
                                <div style="font-size:14px; font-weight:800; color:var(--text-main); margin-bottom:4px;">
                                    ${escapeHTML(doc.title_fr)}
                                </div>
                                <div style="display:flex; gap:8px; align-items:center; font-family:monospace; font-size:11px; font-weight:700; color:var(--accent-blue);">
                                    <span>Code : ${escapeHTML(doc.document_code || 'N/A')}</span>
                                    <span>•</span>
                                    <span>Version : ${escapeHTML(doc.current_version || 'v1.0')}</span>
                                    <span>•</span>
                                    <span>Statut : ${escapeHTML(doc.status || 'validé')}</span>
                                </div>
                            </div>

                            <!-- Cartouche Intégrité CAS & Horodatage BPF -->
                            <div style="background:var(--bg-surface); border:1.5px solid var(--border-color); border-radius:6px; padding:10px; margin-bottom:14px;">
                                <div style="font-size:10.5px; font-weight:800; text-transform:uppercase; color:var(--text-muted); margin-bottom:6px;">
                                    Empreinte Cryptographique d'Origine (CAS SHA-256) :
                                </div>
                                <div style="font-family:monospace; font-size:10.5px; font-weight:700; color:#064e3b; background:#ecfdf5; padding:6px 8px; border-radius:4px; word-break:break-all; border:1.5px solid #6ee7b7;">
                                    ${escapeHTML(doc.content_hash || 'e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855')}
                                </div>
                                <div style="display:flex; justify-content:space-between; margin-top:8px; font-size:11px; color:var(--text-main); font-weight:600;">
                                    <span>Emplacement : <code style="background:#e2e8f0; padding:1px 4px; border-radius:3px;">${escapeHTML(doc.taxonomy_path || 'GED/CTD')}</code></span>
                                    <span>Taille : <strong>${formatFileSize(doc.file_size_bytes || 245000)}</strong></span>
                                </div>
                            </div>

                            <!-- Visualiseur du document (Simulateur BPF Haute Fidélité) -->
                            <div style="flex:1; background:var(--bg-app); border:1.5px solid var(--border-color); border-radius:6px; padding:12px; margin-bottom:14px; display:flex; flex-direction:column; gap:8px;">
                                <div style="display:flex; justify-content:space-between; align-items:center; border-bottom:1.5px solid var(--border-color); padding-bottom:6px; font-size:11.5px;">
                                    <span style="font-weight:700; color:var(--text-main);">Certificat de Contrôle Qualité &amp; Conformité</span>
                                    <span style="color:var(--text-dim); font-size:10.5px; font-weight:600;">Page 1 sur 1</span>
                                </div>
                                <div style="font-size:11.5px; color:var(--text-muted); font-weight:500; line-height:1.5; padding:6px 0;">
                                    Document numérisé conforme aux spécifications du Guide CTD ANPP 2025. L'intégrité du contenu a été scellée lors de la signature électronique qualifiée.
                                </div>
                                <div style="margin-top:auto; display:flex; gap:8px;">
                                    <button class="btn-pill" style="flex:1; justify-content:center; background:var(--accent-blue); color:white; border:none; height:30px; font-weight:700;" onclick="downloadGEDVersion('${item.ged_document_id}')">
                                        📥 Télécharger le Fichier Original
                                    </button>
                                    <button class="btn-pill" style="color:var(--accent-red); background:#fff5f5; border-color:#fca5a5; height:30px; font-weight:700;" title="Détacher la pièce du dossier" onclick="executeDetachGED('${item.id}', '${item.item_code}')" ${isLocked ? 'disabled' : ''}>
                                        ✕ Détacher
                                    </button>
                                </div>
                            </div>

                            <div style="font-size:10.5px; color:var(--text-muted); font-weight:500; line-height:1.4; border-top:1.5px dashed var(--border-color); padding-top:8px;">
                                🔒 <strong>Validation 21 CFR Part 11 :</strong> Toute modification apportée à cette association est irrévocablement consignée dans la piste d'audit BPF avec horodatage UTC.
                            </div>
                        </div>
                    `;
                } else {
                    // Aucune pièce rattachée
                    preview.innerHTML = `
                        <div class="pdf-sim-page" style="text-align:center; display:flex; flex-direction:column; justify-content:center; padding:30px 20px;">
                            <div style="font-size:36px; margin-bottom:12px;">📁</div>
                            <div style="font-size:14px; font-weight:800; color:var(--text-main); margin-bottom:6px;">
                                Aucune Pièce GED Rattachée
                            </div>
                            <div style="font-size:11.5px; color:var(--text-muted); font-weight:500; line-height:1.5; margin-bottom:16px;">
                                L'exigence réglementaire <strong>${item.item_code}</strong> est actuellement vacante dans le coffre-fort CTD.
                            </div>
                            <button class="btn-pill" style="background:var(--accent-blue); color:white; border:none; padding:8px 16px; font-size:11.5px; font-weight:700; align-self:center;" 
                                    onclick="openAttachGEDModal('${item.id}', '${item.item_code}', '${escapeHTML(item.title_fr)}', '${item.requirement_level}')"
                                    ${isLocked ? 'disabled title="Dossier verrouillé"' : ''}>
                                📎 Rattacher une Pièce du Coffre-Fort GED
                            </button>
                            <div style="margin-top:20px; font-size:11px; color:var(--text-dim); font-weight:600;">
                                Basculez sur l'onglet <strong>📋 Fiche ANPP</strong> pour consulter les textes réglementaires applicables.
                            </div>
                        </div>
                    `;
                }
            } else {
                // ONGLET 2 : Fiche Réglementaire ANPP
                preview.innerHTML = `
                    <div class="pdf-sim-page">
                        <div class="pdf-sim-header" style="display:flex; justify-content:space-between; align-items:center;">
                            <span>Exigence ANPP — [${escapeHTML(item.item_code)}]</span>
                            <span class="badge-req-e">${reqDesc.startsWith('Exigé') ? '[E] OBLIGATOIRE' : (reqDesc.startsWith('Facultatif') ? '[F] FACULTATIF' : '[P] PARTIEL')}</span>
                        </div>
                        <div style="font-size:14px; font-weight:800; color:var(--text-main); margin-top:10px; margin-bottom:4px;">
                            ${escapeHTML(item.title_fr)}
                        </div>
                        ${item.title_ar ? `<div style="font-size:12px; font-weight:600; color:var(--text-main); font-family:tahoma,sans-serif; margin-bottom:8px;" dir="rtl">${escapeHTML(item.title_ar)}</div>` : ''}
                        
                        <div style="font-size:11.5px; color:var(--text-main); font-weight:500; line-height:1.5; margin-bottom:12px; background:var(--bg-surface); padding:10px; border-radius:4px; border:1.5px solid var(--border-color); border-left:4px solid var(--accent-blue);">
                            ${escapeHTML(item.description_fr || 'Exigence documentaire selon le Guide CTD ANPP Juillet 2025.')}
                        </div>

                        <div style="margin-bottom:12px;">
                            <div style="font-size:10.5px; font-weight:800; text-transform:uppercase; color:var(--text-muted); margin-bottom:4px;">Niveau d'Exigence Réglementaire :</div>
                            <div style="font-size:11.5px; color:var(--text-main); font-weight:600; line-height:1.4;">${reqDesc}</div>
                        </div>

                        <div style="margin-bottom:12px;">
                            <div style="font-size:10.5px; font-weight:800; text-transform:uppercase; color:var(--text-muted); margin-bottom:4px;">Module &amp; Section CTD :</div>
                            <div style="font-size:11.5px; color:var(--text-main); font-family:monospace; font-weight:700;">Module ${item.module_number === 0 ? 'DM' : item.module_number} — ${getItemSectionInfo(item).title}</div>
                        </div>

                        <div style="margin-bottom:14px;">
                            <div style="font-size:10.5px; font-weight:800; text-transform:uppercase; color:var(--text-muted); margin-bottom:6px;">État de rattachement dans le dossier :</div>
                            ${item.ged_document ? `
                                <div style="background:#dcfce7; border:1.5px solid #4ade80; border-radius:4px; padding:8px 10px; font-size:11.5px; color:#14532d; font-weight:700;">
                                    ✓ Pièce liée : <strong>${escapeHTML(item.ged_document.title_fr)}</strong> (${escapeHTML(item.ged_document.current_version)})
                                </div>
                            ` : `
                                <div style="background:#fee2e2; border:1.5px solid #f87171; border-radius:4px; padding:8px 10px; font-size:11.5px; color:#7f1d1d; font-weight:700; display:flex; justify-content:space-between; align-items:center;">
                                    <span>⚠️ Document manquant</span>
                                    <button class="btn-pill" style="height:22px; padding:1px 8px; font-size:10px; font-weight:700; background:var(--accent-blue); color:white; border:none;" 
                                            onclick="openAttachGEDModal('${item.id}', '${item.item_code}', '${escapeHTML(item.title_fr)}', '${item.requirement_level}')"
                                            ${isLocked ? 'disabled' : ''}>
                                        Rattacher
                                    </button>
                                </div>
                            `}
                        </div>

                        ${item.validation_notes ? `
                            <div style="background:rgba(3,105,161,0.08); border:1.5px solid rgba(3,105,161,0.3); border-radius:4px; padding:8px; font-size:11.5px; color:var(--text-main); margin-bottom:12px; font-weight:500;">
                                <strong>Notes de conformité réglementaire :</strong> ${escapeHTML(item.validation_notes)}
                            </div>
                        ` : ''}

                        <div style="margin-top:auto; padding-top:10px; border-top:1.5px dashed var(--border-color); font-size:10.5px; color:var(--text-muted); font-weight:500;">
                            🏛️ <strong>Référence Officielle :</strong> Agence Nationale des Produits Pharmaceutiques (ANPP) — Décision d'homologation et Guide CTD Édition Juillet 2025.
                        </div>
                    </div>
                `;
            }
        }

        function previewGEDItemDoc(docID) {
            switchDomainModule('ged');
            setTimeout(() => {
                selectGEDDocument(docID);
            }, 300);
        }

        // --- Modals & Actions Dossiers ---

        async function openNewDossierModal() {
            const selectEl = document.getElementById("nd-product-select");
            selectEl.innerHTML = '<option value="">Chargement des produits...</option>';

            try {
                const resMed = await fetch('/api/v1/medicines');
                const dataMed = await resMed.json();
                const meds = dataMed.medicines || [];

                const resDM = await fetch('/api/v1/medical-devices');
                const dataDM = await resDM.json();
                const dms = dataDM.devices || dataDM.medical_devices || [];

                window._cachedMeds = meds;
                window._cachedDMs = dms;

                updateProductSelectOptions();

                const year = new Date().getFullYear();
                const rand = Math.floor(100 + Math.random() * 900);
                document.getElementById("nd-dossier-code").value = `DOS-${year}-ANPP-${rand}`;
                document.getElementById("nd-internal-ref").value = `REF-SANAD-${year}-${rand}`;

                document.getElementById("modal-new-dossier").classList.add("open");
            } catch(e) {
                showNotifyModal("Erreur", "Impossible de charger le référentiel des produits : " + e.message, "⚠️");
            }
        }

        function onNewDossierProcChange() {
            const proc = document.getElementById("nd-procedure-code").value;
            const pTypeEl = document.getElementById("nd-product-type");
            if (proc === 'HOMOLOG_DM') {
                pTypeEl.value = 'dispositif_medical';
            } else {
                pTypeEl.value = 'medicament';
            }
            updateProductSelectOptions();
        }

        function onNewDossierProductTypeChange() {
            updateProductSelectOptions();
        }

        function updateProductSelectOptions() {
            const pType = document.getElementById("nd-product-type").value;
            const selectEl = document.getElementById("nd-product-select");
            selectEl.innerHTML = '';

            if (pType === 'medicament') {
                const meds = window._cachedMeds || [];
                if (meds.length === 0) {
                    selectEl.innerHTML = '<option value="">Aucun médicament disponible dans le référentiel</option>';
                    return;
                }
                meds.forEach(m => {
                    const opt = document.createElement("option");
                    opt.value = m.id;
                    opt.textContent = `${m.brand_name_fr} (${m.dci_fr || 'DCI'} - ${m.dosage_fr || ''} - ${m.origin_type || 'local'})`;
                    selectEl.appendChild(opt);
                });
            } else {
                const dms = window._cachedDMs || [];
                if (dms.length === 0) {
                    selectEl.innerHTML = '<option value="">Aucun dispositif médical disponible</option>';
                    return;
                }
                dms.forEach(dm => {
                    const opt = document.createElement("option");
                    opt.value = dm.id;
                    opt.textContent = `${dm.commercial_name_fr} (Classe ${dm.dm_class} - ${dm.is_sterile ? 'Stérile' : 'Non stérile'})`;
                    selectEl.appendChild(opt);
                });
            }
        }

        async function executeCreateDossier() {
            const code = document.getElementById("nd-dossier-code").value.trim();
            const proc = document.getElementById("nd-procedure-code").value;
            const pType = document.getElementById("nd-product-type").value;
            const productID = document.getElementById("nd-product-select").value;
            const internalRef = document.getElementById("nd-internal-ref").value.trim();
            const targetDate = document.getElementById("nd-target-date").value;
            const notes = document.getElementById("nd-notes").value.trim();

            if (!code) {
                showNotifyModal("Champ Requis", "Le code unique du dossier est obligatoire.", "⚠️");
                return;
            }
            if (!productID) {
                showNotifyModal("Champ Requis", "Veuillez sélectionner un produit du référentiel.", "⚠️");
                return;
            }

            const payload = {
                dossier_code: code,
                procedure_code: proc,
                product_type: pType,
                internal_reference: internalRef || code,
                submission_notes: notes || null
            };

            if (pType === 'medicament') {
                payload.medicine_id = productID;
            } else {
                payload.medical_device_id = productID;
            }

            if (targetDate) {
                payload.target_deposit_date = targetDate;
            }

            try {
                const res = await fetch('/api/v1/dossiers', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(payload)
                });

                if (!res.ok) {
                    const err = await res.json();
                    throw new Error(err.error || "Échec de création");
                }

                const created = await res.json();
                closeModal('modal-new-dossier');
                showNotifyModal(
                    "Dossier ANPP Généré avec Succès",
                    `Le dossier <strong>${created.dossier_code}</strong> a été créé avec sa checklist CTD complète.<br>
                    Nombre d'exigences générées : <strong>${created.items ? created.items.length : '40+'}</strong> pièces.`,
                    "📋"
                );

                selectedDossierId = created.id;
                await loadDossiers();
            } catch(e) {
                showNotifyModal("Erreur de Création", e.message, "⚠️");
            }
        }

        async function openAttachGEDModal(itemId, itemCode, itemTitle, reqLvl) {
            document.getElementById("attach-item-id").value = itemId;
            document.getElementById("attach-item-code").innerText = itemCode;
            document.getElementById("attach-item-title").innerText = itemTitle;
            
            const badgeEl = document.getElementById("attach-item-badge");
            badgeEl.className = reqLvl === 'E' ? 'badge-req-e' : (reqLvl === 'F' ? 'badge-req-f' : 'badge-req-p');
            badgeEl.innerText = `[${reqLvl}] ${reqLvl === 'E' ? 'Exigé' : 'Facultatif'}`;

            const selectEl = document.getElementById("attach-ged-doc-select");
            selectEl.innerHTML = '<option value="">Chargement des pièces de la GED...</option>';

            document.getElementById("modal-attach-ged").classList.add("open");

            try {
                const res = await fetch('/api/v1/ged/documents?limit=100');
                const data = await res.json();
                const docs = data.documents || data.items || [];
                
                selectEl.innerHTML = '<option value="">-- Sélectionner un document validé --</option>';
                docs.forEach(d => {
                    const opt = document.createElement("option");
                    opt.value = d.id;
                    opt.textContent = `[${d.document_code || 'SANS-CODE'}] ${d.title_fr} (${d.current_version} - ${d.status})`;
                    selectEl.appendChild(opt);
                });
            } catch(e) {
                selectEl.innerHTML = '<option value="">Erreur de chargement des documents</option>';
            }
        }

        async function executeAttachGED() {
            const itemId = document.getElementById("attach-item-id").value;
            const gedDocId = document.getElementById("attach-ged-doc-select").value;
            const notes = document.getElementById("attach-notes").value.trim();

            if (!gedDocId) {
                showNotifyModal("Document Requis", "Veuillez sélectionner un document de la GED à rattacher.", "⚠️");
                return;
            }

            try {
                const res = await fetch(`/api/v1/dossiers/${selectedDossierId}/items/${itemId}/attach`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        ged_document_id: gedDocId,
                        notes: notes || null
                    })
                });

                if (!res.ok) {
                    const err = await res.json();
                    throw new Error(err.error || "Rattachement impossible");
                }

                closeModal('modal-attach-ged');
                await selectDossier(selectedDossierId);
                renderDossiersTree();
            } catch(e) {
                showNotifyModal("Échec de Rattachement", e.message, "⚠️");
            }
        }

        async function executeDetachGED(itemId, itemCode) {
            if (!confirm(`Détacher la pièce associée à l'exigence [${itemCode}] ?`)) {
                return;
            }

            try {
                const res = await fetch(`/api/v1/dossiers/${selectedDossierId}/items/${itemId}/detach`, {
                    method: 'POST'
                });

                if (!res.ok) {
                    const err = await res.json();
                    throw new Error(err.error || "Détachement impossible");
                }

                await selectDossier(selectedDossierId);
                renderDossiersTree();
            } catch(e) {
                showNotifyModal("Échec de Détachement", e.message, "⚠️");
            }
        }

        function openSignPDTModal() {
            if (!currentDossierDetail) return;
            if (currentDossierDetail.pdt_signed_at) {
                showNotifyModal("Dossier Déjà Signé", "Ce dossier a déjà reçu le visa officiel du Pharmacien Directeur Technique.", "ℹ️");
                return;
            }
            if (currentDossierDetail.completeness_percent < 100) {
                showNotifyModal(
                    "Verrouillage de Conformité BPF",
                    `Le visa du Pharmacien Directeur Technique exige une complétude stricte de <strong>100%</strong>.<br>
                    Complétude actuelle : <strong>${currentDossierDetail.completeness_percent}%</strong>.<br>
                    Toutes les pièces obligatoires [E] doivent être rattachées avant signature.`,
                    "⚠️"
                );
                return;
            }

            document.getElementById("sign-pdt-dossier-code").innerText = currentDossierDetail.dossier_code;
            document.getElementById("sign-pdt-completeness").innerText = `${currentDossierDetail.completeness_percent}% (Conforme BPF)`;
            document.getElementById("sign-pdt-password").value = '';
            document.getElementById("modal-sign-dossier-pdt").classList.add("open");
        }

        async function executeSignDossierPDT() {
            const password = document.getElementById("sign-pdt-password").value;
            const totp = document.getElementById("sign-pdt-totp").value.trim();

            if (!password) {
                showNotifyModal("Re-authentification Obligatoire", "Veuillez saisir votre mot de passe pour apposer le visa légal 21 CFR Part 11.", "⚠️");
                return;
            }

            try {
                const res = await fetch(`/api/v1/dossiers/${selectedDossierId}/sign-pdt`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        password: password,
                        totp_code: totp || null
                    })
                });

                if (!res.ok) {
                    const err = await res.json();
                    throw new Error(err.error || "Signature rejetée");
                }

                const signed = await res.json();
                closeModal('modal-sign-dossier-pdt');

                showNotifyModal(
                    "Visa PDT Apposé — Dossier Verrouillé pour Dépôt ANPP",
                    `Le dossier <strong>${signed.dossier_code}</strong> a reçu le visa officiel de conformité du Pharmacien Directeur Technique.<br>
                    Empreinte cryptographique inaltérable (21 CFR Part 11) :<br>
                    <code style="word-break:break-all; font-size:10px;">${signed.pdt_signature_seal}</code>`,
                    "✅"
                );

                await selectDossier(selectedDossierId);
                await loadDossiers();
            } catch(e) {
                showNotifyModal("Échec du Visa PDT", e.message, "⚠️");
            }
        }

        function openRecordDepositModal() {
            if (!currentDossierDetail) return;
            if (!currentDossierDetail.pdt_signed_at) {
                showNotifyModal("Visa PDT Préalable Requis", "Le dossier doit obligatoirement avoir reçu le visa officiel du Pharmacien Directeur Technique avant d'enregistrer le dépôt auprès de l'ANPP.", "⚠️");
                return;
            }

            document.getElementById("dep-receipt-ref").value = '';
            document.getElementById("dep-receipt-date").value = new Date().toISOString().substring(0, 10);
            document.getElementById("dep-official-ref").value = '';
            document.getElementById("dep-notes").value = '';

            document.getElementById("modal-record-deposit").classList.add("open");
        }

        async function executeRecordDeposit() {
            const receiptRef = document.getElementById("dep-receipt-ref").value.trim();
            const receiptDate = document.getElementById("dep-receipt-date").value;
            const officialRef = document.getElementById("dep-official-ref").value.trim();
            const notes = document.getElementById("dep-notes").value.trim();

            if (!receiptRef) {
                showNotifyModal("Champ Requis", "Le numéro de récépissé ou quittance de dépôt ANPP est obligatoire.", "⚠️");
                return;
            }

            try {
                const res = await fetch(`/api/v1/dossiers/${selectedDossierId}/status`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        status: 'depose_anpp',
                        deposit_receipt_ref: receiptRef,
                        deposit_receipt_date: receiptDate || null,
                        anpp_official_reference: officialRef || null,
                        submission_notes: notes || null
                    })
                });

                if (!res.ok) {
                    const err = await res.json();
                    throw new Error(err.error || "Mise à jour impossible");
                }

                closeModal('modal-record-deposit');
                showNotifyModal(
                    "Dépôt ANPP Enregistré",
                    `Le dossier est officiellement marqué comme <strong>DÉPOSÉ AUPRÈS DE L'ANPP</strong>.<br>
                    Récépissé : <code>${receiptRef}</code>`,
                    "🏛️"
                );

                await selectDossier(selectedDossierId);
                await loadDossiers();
            } catch(e) {
                showNotifyModal("Erreur Enregistrement Dépôt", e.message, "⚠️");
            }
        }

        function renderEmptyDossierView() {
            const tableContainer = document.querySelector('.table-container');
            if (tableContainer) {
                tableContainer.innerHTML = `
                    <div style="padding:40px; text-align:center; color:var(--text-muted);">
                        <div style="font-size:36px; margin-bottom:12px;">📋</div>
                        <div style="font-size:14px; font-weight:600; color:var(--text-main); margin-bottom:6px;">Aucun dossier réglementaire trouvé</div>
                        <div style="font-size:12px; margin-bottom:16px;">Créez un nouveau dossier pour générer la checklist officielle ANPP (Guide CTD Juillet 2025).</div>
                        <button class="btn-pill" style="background:var(--accent-blue); color:white; border:none; padding:8px 16px; font-size:12px;" onclick="openNewDossierModal()">
                            ➕ Créer un Dossier ANPP
                        </button>
                    </div>
                `;
            }
        }
