// ============================================================================
// COFFRE-FORT GED PHARMACEUTIQUE (BPF & ALCOA+)
// ============================================================================

// ============================================================================
        // COFFRE-FORT GED PHARMACEUTIQUE (BPF & ALCOA+)
        // ============================================================================
        const defaultGEDTaxonomy = [
            {
                code: "01",
                name_fr: "01. Affaires Réglementaires",
                name_ar: "01. الشؤون التنظيمية والصيدلانية",
                description: "Agréments d'exploitation, décisions d'enregistrement/homologation, dossiers CTD et correspondances officielles avec l'ANPP",
                total_docs: 0,
                folders: [
                    { path: "01_Affaires_Reglementaires/01_01_Agrements_Etablissement", code: "01_01", name_fr: "01.01 Agréments Établissement (MIPH)", name_ar: "اعتمادات المؤسسة الصيدلانية", document_count: 0 },
                    { path: "01_Affaires_Reglementaires/01_02_Decisions_Enregistrement", code: "01_02", name_fr: "01.02 Décisions d'Enregistrement (DE)", name_ar: "قرارات تسجيل الأدوية", document_count: 0 },
                    { path: "01_Affaires_Reglementaires/01_03_Decisions_Homologation", code: "01_03", name_fr: "01.03 Décisions d'Homologation DM", name_ar: "قرارات المصادقة على المستلزمات الطبية", document_count: 0 },
                    { path: "01_Affaires_Reglementaires/01_04_Dossiers_CTD_Reference", code: "01_04", name_fr: "01.04 Dossiers CTD de Référence", name_ar: "ملفات التسجيل التقنية المرجعية CTD", document_count: 0 },
                    { path: "01_Affaires_Reglementaires/01_05_Notices_Articles_Conditionnement", code: "01_05", name_fr: "01.05 Notices & Articles Conditionnement", name_ar: "النشرات ومواد التغليف والتوضيب", document_count: 0 },
                    { path: "01_Affaires_Reglementaires/01_06_Correspondances_ANPP", code: "01_06", name_fr: "01.06 Correspondances & Notifications ANPP", name_ar: "المراسلات والإخطارات مع الوكالة الوطنية", document_count: 0 }
                ]
            },
            {
                code: "02",
                name_fr: "02. Assurance Qualité & BPF",
                name_ar: "02. ضمان الجودة وحسن ممارسات الصنع",
                description: "Manuel qualité, SOPs, certificats BPF/GMP, rapports d'audits et CAPA",
                total_docs: 0,
                folders: [
                    { path: "02_Assurance_Qualite_BPF/02_01_Manuel_Qualite_Procedures", code: "02_01", name_fr: "02.01 Manuel Qualité & Procédures (SOP)", name_ar: "دليل الجودة وإجراءات العمل الموحدة", document_count: 0 },
                    { path: "02_Assurance_Qualite_BPF/02_02_Certificats_BPF_ISO", code: "02_02", name_fr: "02.02 Certificats BPF & ISO", name_ar: "شهادات حسن ممارسات الصنع وISO", document_count: 0 },
                    { path: "02_Assurance_Qualite_BPF/02_03_Rapports_Audit_Inspections", code: "02_03", name_fr: "02.03 Rapports d'Audit & Inspections ANPP", name_ar: "تقارير التدقيق والتفتيش الصيدلاني", document_count: 0 },
                    { path: "02_Assurance_Qualite_BPF/02_04_CAPA_Deviations", code: "02_04", name_fr: "02.04 CAPA & Gestion des Déviations", name_ar: "الإجراءات التصحيحية والوقائية والانحرافات", document_count: 0 }
                ]
            },
            {
                code: "03",
                name_fr: "03. Développement & Production",
                name_ar: "03. التطوير والإنتاج الصيدلاني",
                description: "Formules de fabrication, méthodes de contrôle, stabilités zone IVb et dossiers lots pilotes",
                total_docs: 0,
                folders: [
                    { path: "03_Developpement_Production/03_01_Formules_Fabrication_DMF", code: "03_01", name_fr: "03.01 Formules de Fabrication & DMF", name_ar: "صيغ التصنيع وملفات المواد الفعالة", document_count: 0 },
                    { path: "03_Developpement_Production/03_02_Methodes_Controle_Validation", code: "03_02", name_fr: "03.02 Méthodes de Contrôle & Validations", name_ar: "طرق المراقبة والتحقق التحليلي", document_count: 0 },
                    { path: "03_Developpement_Production/03_03_Etudes_Stabilite", code: "03_03", name_fr: "03.03 Études de Stabilité (Zone IVb)", name_ar: "دراسات الثبات وحفظ الأدوية", document_count: 0 },
                    { path: "03_Developpement_Production/03_04_Specifications_Matieres_PF", code: "03_04", name_fr: "03.04 Spécifications Matières & PF", name_ar: "دفاتر مواصفات المواد والمنتج النهائي", document_count: 0 },
                    { path: "03_Developpement_Production/03_05_Dossiers_Techniques_DM", code: "03_05", name_fr: "03.05 Dossiers Techniques DM", name_ar: "الملفات التقنية للمستلزمات الطبية", document_count: 0 }
                ]
            },
            {
                code: "04",
                name_fr: "04. Vigilance & Post-Market",
                name_ar: "04. اليقظة الصيدلانية ومتابعة ما بعد التسويق",
                description: "Pharmacovigilance, matériovigilance et gestion des retraits/rappels de lots",
                total_docs: 0,
                folders: [
                    { path: "04_Vigilance_Post_Market/04_01_Pharmacovigilance", code: "04_01", name_fr: "04.01 Pharmacovigilance & PSUR", name_ar: "اليقظة الدوائية وتقارير السلامة الدورية", document_count: 0 },
                    { path: "04_Vigilance_Post_Market/04_02_Materiovigilance", code: "04_02", name_fr: "04.02 Matériovigilance", name_ar: "اليقظة الخاصة بالمستلزمات الطبية", document_count: 0 },
                    { path: "04_Vigilance_Post_Market/04_03_Rappels_Lots", code: "04_03", name_fr: "04.03 Gestion des Rappels de Lots", name_ar: "إجراءات سحب واسترجاع الحصص الدوائية", document_count: 0 }
                ]
            },
            {
                code: "05",
                name_fr: "05. Contrats Tiers & Façonnage",
                name_ar: "05. العقود مع الغير والمناولة",
                description: "Accords qualité QTA, contrats de façonnage/sous-traitance et certificats CEP",
                total_docs: 0,
                folders: [
                    { path: "05_Contrats_Tiers/05_01_Accords_Qualite_QTA", code: "05_01", name_fr: "05.01 Accords Qualité (QTA)", name_ar: "اتفاقيات الجودة التقنية مع الشركاء", document_count: 0 },
                    { path: "05_Contrats_Tiers/05_02_Contrats_Faconnage_SousTraitance", code: "05_02", name_fr: "05.02 Contrats Façonnage & Sous-Traitance", name_ar: "عقود التصنيع بالمناولة والتحاليل الخارجية", document_count: 0 },
                    { path: "05_Contrats_Tiers/05_03_Certificats_Fournisseurs_APIs", code: "05_03", name_fr: "05.03 Certificats CEP & Accès APIs", name_ar: "شهادات المطابقة الأوروبية ورسائل الوصول للمواد", document_count: 0 }
                ]
            }
        ];

        let currentGEDTaxonomy = JSON.parse(JSON.stringify(defaultGEDTaxonomy));
        let currentGEDTaxonomyPath = '';
        let currentGEDStatusFilter = '';
        let currentGEDDocuments = [];
        let selectedGEDDocId = null;
        const openedGEDBranches = new Set(['01', '02', '03', '04', '05']);

        function openNewGEDDocModal(ctx) {
            const form = document.getElementById('form-new-ged-doc');
            if (form) form.reset();

            // Résolution automatique du contexte si non passé explicitement
            if (!ctx && typeof currentDossierDetail !== 'undefined' && currentDossierDetail && typeof selectedDossierItemId !== 'undefined' && selectedDossierItemId) {
                const items = currentDossierDetail.items || currentDossierDetail.checklist_data || [];
                const it = items.find(i => i.id === selectedDossierItemId);
                if (it) {
                    ctx = {
                        dossierId: currentDossierDetail.id,
                        dossierCode: currentDossierDetail.dossier_code || 'DOSSIER',
                        itemId: it.id,
                        itemCode: it.item_code,
                        itemTitle: it.title_fr,
                        itemTitleAr: it.title_ar,
                        department: it.assigned_department
                    };
                }
            }

            if (ctx && ctx.itemCode) {
                window.pendingGEDTargetItem = {
                    dossierId: ctx.dossierId,
                    itemId: ctx.itemId,
                    itemCode: ctx.itemCode
                };

                const code = ctx.itemCode || '';
                const title = ctx.itemTitle || '';
                const dCode = ctx.dossierCode || '';

                // 1. Taxonomie intelligente selon le module CTD
                const taxEl = document.getElementById('ged-doc-taxonomy');
                if (taxEl) {
                    if (code.startsWith('1.P.2') || /notice|étiquet|conditionnement/i.test(title)) {
                        taxEl.value = '01_Affaires_Reglementaires/01_05_Notices_Articles_Conditionnement';
                    } else if (code.startsWith('1.')) {
                        taxEl.value = '01_Affaires_Reglementaires/01_04_Dossiers_CTD_Reference';
                    } else if (code.startsWith('3.2.P.8') || /stabilit/i.test(title)) {
                        taxEl.value = '03_Developpement_Production/03_03_Etudes_Stabilite';
                    } else if (code.startsWith('3.2.S') || /substance|api|dmf/i.test(title)) {
                        taxEl.value = '03_Developpement_Production/03_01_Formules_Fabrication_DMF';
                    } else if (code.startsWith('3.2.P') || /validation|fabrication|procédé/i.test(title)) {
                        taxEl.value = '03_Developpement_Production/03_02_Methodes_Controle_Validation';
                    } else {
                        taxEl.value = '01_Affaires_Reglementaires/01_04_Dossiers_CTD_Reference';
                    }
                }

                // 2. Type de document
                const typeEl = document.getElementById('ged-doc-type');
                if (typeEl) {
                    if (/engagement|attestation|non-commercialisation/i.test(title) || code === '1.P.8') {
                        typeEl.value = 'lettre_engagement';
                    } else if (/validation|procédé|3 lots/i.test(title)) {
                        typeEl.value = 'rapport_validation';
                    } else if (/stabilit/i.test(title)) {
                        typeEl.value = 'rapport_stabilite';
                    } else if (/notice|maquette/i.test(title)) {
                        typeEl.value = 'notice_conditionnement';
                    } else if (/bpf|gmp/i.test(title)) {
                        typeEl.value = 'certificat_bpf';
                    }
                }

                // 3. Titre du document (FR)
                const titleFrEl = document.getElementById('ged-doc-title-fr');
                if (titleFrEl) {
                    if (code === '1.P.8' || /engagement.*3 lots/i.test(title)) {
                        titleFrEl.value = `Lettre d'Engagement de Non-Commercialisation & Validation 3 Lots [${dCode}]`;
                    } else {
                        titleFrEl.value = `[${code}] ${title} - ${dCode}`;
                    }
                }

                // 4. Code Référence BPF
                const codeEl = document.getElementById('ged-doc-code');
                if (codeEl) {
                    const cleanD = dCode.replace(/[^A-Za-z0-9]/g, '').substring(0, 10).toUpperCase();
                    if (code === '1.P.8') {
                        codeEl.value = `ENG-VAL-1P8-${cleanD || 'PARAC'}`;
                    } else {
                        const cleanC = code.replace(/\./g, '');
                        codeEl.value = `DOC-CTD-${cleanC}-${cleanD}`;
                    }
                }

                // 5. Intitulé en Arabe
                const titleArEl = document.getElementById('ged-doc-title-ar');
                if (titleArEl) {
                    if (code === '1.P.8') {
                        titleArEl.value = 'تعهد بعدم تسويق حصص الإنتاج قبل استكمال وتأكيد مطابقة المصادقة الصناعية';
                    } else if (ctx.itemTitleAr) {
                        titleArEl.value = ctx.itemTitleAr;
                    }
                }

                // 6. Motif BPF
                const sumEl = document.getElementById('ged-doc-summary');
                if (sumEl) {
                    sumEl.value = `Dépôt initial de la pièce justificative pour l'exigence CTD ${code} (${dCode})`;
                }

                // 7. Tags
                const tagsEl = document.getElementById('ged-doc-tags');
                if (tagsEl) {
                    tagsEl.value = `${code}, CTD, Validation, BPF, Non-commercialisation, ${dCode}`;
                }
            } else {
                window.pendingGEDTargetItem = null;
            }

            document.getElementById('modal-new-ged-doc').classList.add('open');
        }

        async function loadGEDTaxonomy() {
            try {
                const res = await fetch('/api/v1/ged/taxonomy');
                if (res.status === 401) {
                    handleAuthExpired();
                    return;
                }
                if (!res.ok) throw new Error("HTTP " + res.status);
                const data = await res.json();
                if (data.branches && data.branches.length > 0) {
                    currentGEDTaxonomy = data.branches;
                }
                renderGEDTree();
            } catch (e) {
                console.error("Erreur chargement taxonomie GED :", e);
                renderGEDTree();
            }
        }

        function toggleGEDBranchCollapse(branchCode, evt) {
            if (evt) evt.stopPropagation();
            if (openedGEDBranches.has(branchCode)) {
                openedGEDBranches.delete(branchCode);
            } else {
                openedGEDBranches.add(branchCode);
            }
            renderGEDTree();
        }

        function selectGEDBranch(branchCode, branchPrefix, branchName) {
            openedGEDBranches.add(branchCode);
            selectedGEDDocId = null;
            selectGEDFolder(branchPrefix, branchName);
        }

        function renderGEDTree() {
            const container = document.getElementById("tree-root");
            if (!container) return;

            let html = `
                <div class="tree-node ${currentGEDTaxonomyPath === '' ? 'selected' : ''}" onclick="selectGEDFolder('', '${isArabic ? "جميع وثائق المخزن الرقمي" : "Tous les Documents GED"}')">
                    <span class="chevron">📁</span>
                    <span><strong>${isArabic ? "جميع وثائق GED BPF" : "Tous les Documents GED"}</strong></span>
                    <span class="count-badge ok">Coffre-Fort BPF</span>
                </div>
            `;

            currentGEDTaxonomy.forEach(branch => {
                const bCode = branch.code || branch.Code || '';
                const bName = (isArabic ? branch.name_ar : branch.name_fr) || branch.name_fr || branch.NameFR || '';
                const bDocs = branch.total_docs ?? branch.TotalDocs ?? 0;
                const folders = branch.folders || branch.Folders || [];
                const bPrefix = folders[0] ? (folders[0].path || folders[0].Path || '').split('/')[0] : '';
                const isOpen = openedGEDBranches.has(bCode);
                const isBranchSelected = currentGEDTaxonomyPath === bPrefix;

                html += `
                    <div class="tree-node ${isBranchSelected ? 'selected' : ''}" onclick="selectGEDBranch('${bCode}', '${bPrefix}', '${bName.replace(/'/g, "\\'")}')">
                        <span class="chevron" onclick="toggleGEDBranchCollapse('${bCode}', event)" style="cursor:pointer; padding:2px;">${isOpen ? '▼' : '▶'}</span>
                        <span>📁 <strong>${bName}</strong></span>
                        <span class="count-badge ok">${bDocs}</span>
                    </div>
                    <div class="tree-sub ${isOpen ? 'open' : ''}" id="sub-branch-${bCode}">
                `;

                folders.forEach(folder => {
                    const fPath = folder.path || folder.Path || '';
                    const fCode = folder.code || folder.Code || '';
                    const fName = (isArabic ? folder.name_ar : folder.name_fr) || folder.name_fr || folder.NameFR || '';
                    const fCount = folder.document_count ?? folder.DocumentCount ?? 0;
                    const isFolderSelected = currentGEDTaxonomyPath === fPath;

                    html += `
                        <div class="tree-node ${isFolderSelected ? 'selected' : ''}" style="padding-left:24px;" onclick="selectGEDFolder('${fPath}', '${fName.replace(/'/g, "\\'")}')">
                            <span>📄 ${fName}</span>
                            <span class="count-badge ${fCount > 0 ? 'ok' : 'req'}">${fCount}</span>
                        </div>
                    `;
                });

                html += `</div>`;
            });

            container.innerHTML = html;
        }

        function selectGEDFolder(path, name) {
            currentGEDTaxonomyPath = path;
            selectedGEDDocId = null;
            renderGEDTree();

            document.getElementById("bc-root").innerText = "COFFRE-FORT";
            document.getElementById("bc-sec").innerText = "GED_BPF";
            document.getElementById("bc-prod").innerText = path ? (path.split('/')[0] || "TOUS") : "TOUS";
            document.getElementById("bc-mod").innerText = name || "Tous les dossiers";

            if (window.innerWidth < 768) {
                toggleMobileTree(false);
            }

            loadGEDDocuments(path, currentGEDStatusFilter);
        }

        function filterGEDByStatus(status) {
            currentGEDStatusFilter = status;
            selectedGEDDocId = null;
            loadGEDDocuments(currentGEDTaxonomyPath, status);
        }

        async function loadGEDDocuments(taxonomyPath = '', status = '', search = '') {
            const tbody = document.getElementById("doc-tbody");
            if (tbody) {
                tbody.innerHTML = `<tr><td colspan="8" style="text-align:center; padding:20px; color:var(--text-muted);">Chargement sécurisé du coffre-fort CAS...</td></tr>`;
            }

            let url = `/api/v1/ged/documents?limit=50`;
            if (taxonomyPath) url += `&taxonomy_path=${encodeURIComponent(taxonomyPath)}`;
            if (status) url += `&status=${encodeURIComponent(status)}`;
            if (search) url += `&search=${encodeURIComponent(search)}`;

            try {
                const res = await fetch(url);
                if (res.status === 401) {
                    handleAuthExpired();
                    return;
                }
                if (!res.ok) throw new Error("HTTP " + res.status);
                const data = await res.json();
                currentGEDDocuments = data.documents || data.items || [];
                const exists = currentGEDDocuments.some(d => d.id === selectedGEDDocId);
                if (!exists) {
                    selectedGEDDocId = currentGEDDocuments.length > 0 ? currentGEDDocuments[0].id : null;
                }
                renderGEDDocumentsTable(currentGEDDocuments);
            } catch (e) {
                if (tbody) {
                    tbody.innerHTML = `<tr><td colspan="8" style="text-align:center; padding:20px; color:var(--accent-red);">Erreur lors de la lecture des documents : ${e.message}</td></tr>`;
                }
            }
        }

        function selectGEDDocument(id) {
            selectedGEDDocId = id;
            document.querySelectorAll('#doc-tbody tr').forEach(tr => {
                if (tr.getAttribute('data-doc-id') === id) tr.classList.add('selected');
                else tr.classList.remove('selected');
            });
            const doc = currentGEDDocuments.find(d => d.id === id);
            if (doc) {
                renderGEDPreview(doc);
            }
        }

        function formatGEDStatusBadge(status) {
            switch (status) {
                case 'en_vigueur':
                    return '<span class="status-tag status-ok">EN VIGUEUR</span>';
                case 'en_approbation_dt':
                    return '<span class="status-tag" style="background:rgba(147,51,234,0.15); color:#a855f7; border:1px solid rgba(147,51,234,0.3);">EN APPROBATION DT</span>';
                case 'en_revue':
                    return '<span class="status-tag status-rev">EN REVUE</span>';
                case 'brouillon':
                    return '<span class="status-tag" style="background:rgba(148,163,184,0.15); color:#94a3b8; border:1px solid rgba(148,163,184,0.3);">BROUILLON</span>';
                case 'obsolete':
                    return '<span class="status-tag status-rej">OBSOLÈTE</span>';
                case 'archive':
                    return '<span class="status-tag" style="background:rgba(100,116,139,0.2); color:#64748b;">ARCHIVÉ</span>';
                default:
                    return `<span class="status-tag">${status}</span>`;
            }
        }

        function renderGEDDocumentsTable(docs) {
            const thead = document.getElementById("table-head");
            const tbody = document.getElementById("doc-tbody");
            if (!thead || !tbody) return;

            thead.innerHTML = `
                <tr>
                    <th style="width: 105px;">Statut BPF</th>
                    <th>Code &amp; Titre de la Pièce</th>
                    <th class="col-hide-sm">Dossier Taxonomie</th>
                    <th style="width: 65px;">Version</th>
                    <th class="col-hide-sm" style="width: 110px;">Verrouillage</th>
                    <th class="col-hide-md">Validation DT</th>
                    <th class="col-hide-lg">Stockage CAS (SHA-256)</th>
                    <th style="width: 130px; text-align:right;">Actions</th>
                </tr>
            `;

            tbody.innerHTML = "";
            if (docs.length === 0) {
                tbody.innerHTML = `<tr><td colspan="8" style="text-align:center; padding:25px; color:var(--text-dim);">Aucun document dans ce dossier. Cliquez sur '➕ Déposer Document' pour téléverser votre première pièce.</td></tr>`;
                renderGEDPreview(null);
                return;
            }

            docs.forEach(doc => {
                const tr = document.createElement("tr");
                tr.setAttribute('data-doc-id', doc.id);
                tr.style.cursor = 'pointer';
                if (doc.id === selectedGEDDocId) tr.classList.add("selected");
                tr.onclick = (e) => {
                    selectGEDDocument(doc.id);
                    if (window.innerWidth < 1400) togglePreviewPane(true);
                };

                const v = doc.latest_version;
                const folderShort = doc.taxonomy_path.split('/')[1] || doc.taxonomy_path;

                // Verrouillage
                let lockBadge = `<span style="color:var(--text-dim); font-size:11px;">🔓 Libre</span>`;
                if (doc.is_locked) {
                    lockBadge = `<span style="color:var(--accent-red); font-weight:600; font-size:11px;" title="Verrouillé par ${doc.locked_by_name || 'Utilisateur'}">🔒 Verrouillé</span>`;
                }

                // Signature DT
                let dtBadge = `<span style="color:var(--text-dim); font-size:11px;">⏳ En attente</span>`;
                if (v && v.is_signed_dt) {
                    dtBadge = `<span style="color:var(--accent-emerald); font-weight:600; font-size:11px;" title="Scellé par ${v.signed_by_dt_name}">✅ Scellé PDT</span>`;
                }

                // SHA-256
                const sha = v ? v.file_sha256 : "e3b0c44...";
                const sizeKB = v ? (v.file_size_bytes / 1024).toFixed(1) + " Ko" : "-";

                tr.innerHTML = `
                    <td>${formatGEDStatusBadge(doc.status)}</td>
                    <td>
                        <div style="font-weight:600; color:var(--text-main);">${doc.title_fr}</div>
                        <div style="font-size:10px; color:var(--accent-blue); font-family:monospace;">${doc.document_code}</div>
                    </td>
                    <td class="col-hide-sm" style="font-size:11px; color:var(--text-muted);">${folderShort}</td>
                    <td><strong style="color:var(--accent-emerald);">${doc.current_version}</strong></td>
                    <td class="col-hide-sm">${lockBadge}</td>
                    <td class="col-hide-md">${dtBadge}</td>
                    <td class="col-hide-lg">
                        <div style="font-family:monospace; font-size:10px; color:var(--text-dim);" title="${sha}">${sha.substring(0, 16)}...</div>
                        <div style="font-size:10px; color:var(--text-muted);">${sizeKB}</div>
                    </td>
                    <td style="text-align:right;">
                        <button class="btn-pill btn-zero-trust-pharma-doc" style="padding:2px 6px; font-size:10px;" onclick="selectGEDDocument('${doc.id}'); downloadGEDVersion('${doc.id}', '${v ? v.id : ''}')" title="Télécharger le fichier physique">📥</button>
                        <button class="btn-pill" style="padding:2px 6px; font-size:10px;" onclick="selectGEDDocument('${doc.id}'); openGEDVersionsModal('${doc.id}')" title="Historique & Check-in">📜</button>
                        ${doc.is_locked ? 
                            `<button class="btn-pill" style="padding:2px 6px; font-size:10px; color:var(--accent-red);" onclick="selectGEDDocument('${doc.id}'); handleCancelCheckOutGED('${doc.id}')" title="Annuler le verrouillage">🔓</button>` :
                            `<button class="btn-pill" style="padding:2px 6px; font-size:10px;" onclick="selectGEDDocument('${doc.id}'); handleCheckOutGED('${doc.id}')" title="Verrouiller pour édition (Check-out)">🔒</button>`
                        }
                        ${((doc.status === 'en_approbation_dt' || (v && !v.is_signed_dt)) && (typeof hasPermission !== 'function' || hasPermission('DOC_APPROVE_PDT'))) ?
                            `<button class="btn-pill" style="padding:2px 6px; font-size:10px; background:rgba(16,185,129,0.15); color:var(--accent-emerald);" onclick="selectGEDDocument('${doc.id}'); openSignGEDDTModal('${doc.id}', '${v ? v.id : ''}', '${doc.document_code}', '${doc.title_fr}', '${sha}')" title="Signer électroniquement (PDT)">✍️</button>` : ''
                        }
                    </td>
                `;

                tbody.appendChild(tr);
            });

            const selectedDoc = currentGEDDocuments.find(d => d.id === selectedGEDDocId) || currentGEDDocuments[0] || null;
            renderGEDPreview(selectedDoc);
        }

        function renderGEDPreview(doc) {
            const preview = document.getElementById("preview-body");
            if (!preview || !doc) return;

            const v = doc.latest_version;
            const sha = v ? v.file_sha256 : "e3b0c442...";
            const sizeFormatted = v ? (v.file_size_bytes > 1024*1024 ? (v.file_size_bytes/(1024*1024)).toFixed(2) + " Mo" : (v.file_size_bytes/1024).toFixed(1) + " Ko") : "-";

            preview.innerHTML = `
                <div class="pdf-sim-page">
                    <div class="pdf-sim-header" style="border-bottom:2px solid var(--accent-blue); padding-bottom:8px; margin-bottom:12px;">
                        <div style="font-size:10px; color:var(--accent-blue); font-weight:700;">RÉPUBLIQUE ALGÉRIENNE DÉMOCRATIQUE ET POPULAIRE</div>
                        <div style="font-size:12px; font-weight:800; color:var(--text-main); margin-top:2px;">COFFRE-FORT GED PHARMACEUTIQUE — BPF / ALCOA+</div>
                        <div style="font-size:10px; color:var(--text-muted); font-family:monospace;">CODE : ${doc.document_code} | VERSION : ${doc.current_version}</div>
                    </div>

                    <div style="font-size:13px; font-weight:700; color:var(--text-main); margin-bottom:6px;">
                        ${doc.title_fr}
                    </div>
                    ${doc.title_ar ? `<div style="font-size:12px; color:var(--text-muted); margin-bottom:12px; font-family:'Amiri', serif;" dir="rtl">${doc.title_ar}</div>` : ''}

                    <div style="background:var(--bg-surface); padding:8px; border-radius:4px; border:1px solid var(--border-color); font-size:11px; margin-bottom:12px;">
                        <div style="margin-bottom:4px;"><strong>Dossier BPF :</strong> ${doc.taxonomy_path}</div>
                        <div style="margin-bottom:4px;"><strong>Type documentaire :</strong> ${doc.document_type}</div>
                        <div style="margin-bottom:4px;"><strong>Département :</strong> ${doc.department || 'Non spécifié'}</div>
                        ${v ? `<div style="margin-bottom:4px;"><strong>Motif de révision :</strong> <em>${v.revision_summary}</em></div>` : ''}
                    </div>

                    ${v && v.is_signed_dt ? `
                        <div style="background:rgba(16,185,129,0.08); border:1px solid var(--accent-emerald); border-radius:4px; padding:10px; margin-bottom:12px;">
                            <div style="color:var(--accent-emerald); font-weight:700; font-size:11px; display:flex; align-items:center; gap:6px;">
                                <span>🛡️</span> <span>SCEAU LÉGAL DU DIRECTEUR TECHNIQUE (PDT)</span>
                            </div>
                            <div style="font-size:10px; color:var(--text-dim); margin-top:4px;">
                                Signataire : <strong>${v.signed_by_dt_name || 'Dr. Amine K.'}</strong><br>
                                Horodatage certifié : ${new Date(v.signed_at).toLocaleString()}<br>
                                Empreinte du scellement : <code style="word-break:break-all;">${v.signature_sha256}</code>
                            </div>
                        </div>
                    ` : ''}

                    <div style="margin-top:15px; padding-top:8px; border-top:1px dashed var(--border-color); font-size:10px; color:var(--text-dim);">
                        Stockage physique CAS : <code>${sha}</code><br>
                        Déduplication &amp; Intégrité ALCOA+ vérifiées à chaque accès.
                    </div>
                </div>

                <div class="meta-box" style="margin-top:15px;">
                    <div class="meta-row"><span class="meta-label">Statut actuel :</span><span class="meta-val">${formatGEDStatusBadge(doc.status)}</span></div>
                    <div class="meta-row"><span class="meta-label">Fichier original :</span><span class="meta-val" style="word-break:break-all;">${v ? v.original_filename : 'N/A'}</span></div>
                    <div class="meta-row"><span class="meta-label">Poids du binaire :</span><span class="meta-val">${sizeFormatted}</span></div>
                    <div class="meta-row"><span class="meta-label">Nombre de révisions :</span><span class="meta-val">${doc.versions_count || 1} version(s)</span></div>
                    <div class="meta-row"><span class="meta-label">État de verrou :</span><span class="meta-val">${doc.is_locked ? '🔒 En cours par ' + (doc.locked_by_name || 'utilisateur') : '🔓 Déverrouillé'}</span></div>
                </div>

                <div style="display:flex; flex-direction:column; gap:6px; margin-top:12px;">
                    <button class="btn-pill btn-zero-trust-pharma-doc" style="width:100%; background:var(--accent-blue); color:white; border:none;" onclick="downloadGEDVersion('${doc.id}', '${v ? v.id : ''}')">📥 Télécharger la Pièce (CAS)</button>
                    <button class="btn-pill" style="width:100%;" onclick="openGEDVersionsModal('${doc.id}')">📜 Consulter les Versions &amp; Check-in</button>
                    ${doc.status === 'brouillon' ? `<button class="btn-pill" style="width:100%;" onclick="handleStatusChangeGED('${doc.id}', 'en_revue')">🔄 Soumettre pour Revue</button>` : ''}
                    ${doc.status === 'en_revue' ? `<button class="btn-pill" style="width:100%;" onclick="handleStatusChangeGED('${doc.id}', 'en_approbation_dt')">📤 Transmettre pour Approbation PDT</button>` : ''}
                    ${doc.status === 'en_approbation_dt' && (typeof hasPermission !== 'function' || hasPermission('DOC_APPROVE_PDT')) ? `<button class="btn-pill" style="width:100%; background:var(--accent-emerald); color:white; border:none;" onclick="openSignGEDDTModal('${doc.id}', '${v ? v.id : ''}', '${doc.document_code}', '${doc.title_fr}', '${sha}')">✍️ Signer et Mettre en Vigueur (PDT)</button>` : ''}
                </div>
            `;
        }

        async function handleCreateGEDDocument(e) {
            e.preventDefault();
            const form = document.getElementById('form-new-ged-doc');
            const fileInput = document.getElementById('ged-doc-file');

            if (!fileInput.files || fileInput.files.length === 0) {
                showNotifyModal("Fichier Requis", "Veuillez sélectionner un fichier binaire à déposer dans le CAS.", "⚠️");
                return;
            }

            const formData = new FormData();
            formData.append('taxonomy_path', document.getElementById('ged-doc-taxonomy').value);
            formData.append('title_fr', document.getElementById('ged-doc-title-fr').value.trim());
            formData.append('title_ar', document.getElementById('ged-doc-title-ar').value.trim());
            formData.append('document_code', document.getElementById('ged-doc-code').value.trim());
            formData.append('document_type', document.getElementById('ged-doc-type').value);
            formData.append('department', 'affaires_reglementaires');
            formData.append('revision_summary', document.getElementById('ged-doc-summary').value.trim());
            formData.append('expiry_date', document.getElementById('ged-doc-expiry').value);
            formData.append('tags', document.getElementById('ged-doc-tags').value.trim());
            formData.append('file', fileInput.files[0]);

            try {
                const res = await fetch('/api/v1/ged/documents', {
                    method: 'POST',
                    body: formData
                });
                if (!res.ok) {
                    const err = await res.json();
                    throw new Error(err.error || "Échec du dépôt");
                }
                const created = await res.json();
                closeModal('modal-new-ged-doc');

                // Si le dépôt a été déclenché depuis un item de dossier CTD (contexte ciblé)
                if (window.pendingGEDTargetItem && window.pendingGEDTargetItem.dossierId && window.pendingGEDTargetItem.itemId) {
                    const target = window.pendingGEDTargetItem;
                    window.pendingGEDTargetItem = null;

                    try {
                        const attachRes = await fetch(`/api/v1/dossiers/${target.dossierId}/items/${target.itemId}/attach`, {
                            method: 'POST',
                            headers: { 'Content-Type': 'application/json' },
                            body: JSON.stringify({
                                ged_document_id: created.id,
                                notes: `Rattaché automatiquement lors du dépôt : [${created.document_code}] ${created.title_fr}`
                            })
                        });
                        if (attachRes.ok) {
                            showNotifyModal(
                                "Pièce Déposée & Rattachée",
                                `Le document <strong>[${created.document_code}] ${created.title_fr}</strong> a été stocké dans le CAS (SHA-256) et <strong>directement rattaché à l'exigence ${target.itemCode}</strong> !`,
                                "✅"
                            );
                            if (typeof selectDossier === 'function') await selectDossier(target.dossierId);
                            if (typeof selectDossierItem === 'function') selectDossierItem(target.itemId);
                            if (typeof switchPreviewSubTab === 'function') switchPreviewSubTab('doc');
                            return;
                        }
                    } catch (attachErr) {
                        console.warn("Échec du rattachement automatique :", attachErr);
                    }
                }

                showNotifyModal(
                    "Document Enregistré dans le Coffre-Fort",
                    `Le document <strong>[${created.document_code}] ${created.title_fr}</strong> a été stocké avec succès dans le Content-Addressable Storage (CAS) dédupliqué SHA-256.`,
                    "✅"
                );
                loadGEDTaxonomy();
                loadGEDDocuments(currentGEDTaxonomyPath);
            } catch (err) {
                showNotifyModal("Erreur de Dépôt GED", err.message, "⚠️");
            }
        }

        async function openGEDVersionsModal(docID) {
            selectGEDDocument(docID);
            document.getElementById('ver-doc-id').value = docID;
            document.getElementById('form-ged-new-version').reset();
            const container = document.getElementById('ged-history-table-container');
            container.innerHTML = `<div style="text-align:center; padding:15px; color:var(--text-dim);">Chargement des versions...</div>`;
            document.getElementById('modal-ged-versions').classList.add('open');

            try {
                const res = await fetch(`/api/v1/ged/documents/${docID}/versions`);
                if (!res.ok) throw new Error("HTTP " + res.status);
                const data = await res.json();
                const versions = data.versions || [];

                let html = `
                    <table class="table-main" style="width:100%; font-size:11px;">
                        <thead>
                            <tr>
                                <th style="width:60px;">Version</th>
                                <th>Date &amp; Auteur</th>
                                <th>Motif de Révision BPF</th>
                                <th>Signature PDT</th>
                                <th>SHA-256</th>
                                <th style="text-align:right;">Télécharger</th>
                            </tr>
                        </thead>
                        <tbody>
                `;

                versions.forEach(v => {
                    const sizeKB = (v.file_size_bytes / 1024).toFixed(1) + " Ko";
                    html += `
                        <tr>
                            <td><strong style="color:var(--accent-emerald); font-size:12px;">${v.version_label}</strong></td>
                            <td>
                                <div>${new Date(v.created_at).toLocaleDateString()} ${new Date(v.created_at).toLocaleTimeString().substring(0,5)}</div>
                                <div style="color:var(--text-muted); font-size:10px;">${v.author_name || 'Auteur BPF'}</div>
                            </td>
                            <td><em>${v.revision_summary}</em></td>
                            <td>
                                ${v.is_signed_dt ? 
                                    `<span style="color:var(--accent-emerald); font-weight:600;">✅ Scellé par ${v.signed_by_dt_name}</span>` : 
                                    `<span style="color:var(--text-dim);">Non signé</span>`
                                }
                            </td>
                            <td><code style="font-size:10px;" title="${v.file_sha256}">${v.file_sha256.substring(0, 16)}...</code></td>
                            <td style="text-align:right;">
                                <button class="btn-pill" style="padding:2px 8px; font-size:10px;" onclick="downloadGEDVersion('${docID}', '${v.id}')">📥 ${sizeKB}</button>
                            </td>
                        </tr>
                    `;
                });

                html += `</tbody></table>`;
                container.innerHTML = html;
            } catch (err) {
                container.innerHTML = `<div style="color:var(--accent-red); padding:10px;">Erreur de chargement : ${err.message}</div>`;
            }
        }

        async function handleUploadGEDVersion(e) {
            e.preventDefault();
            const docID = document.getElementById('ver-doc-id').value;
            const fileInput = document.getElementById('ver-file');

            if (!fileInput.files || fileInput.files.length === 0) {
                showNotifyModal("Fichier Requis", "Veuillez sélectionner un fichier révisé.", "⚠️");
                return;
            }

            const formData = new FormData();
            formData.append('revision_summary', document.getElementById('ver-summary').value.trim());
            formData.append('is_major', document.getElementById('ver-is-major').checked ? 'true' : 'false');
            formData.append('file', fileInput.files[0]);

            try {
                const res = await fetch(`/api/v1/ged/documents/${docID}/versions`, {
                    method: 'POST',
                    body: formData
                });
                if (!res.ok) {
                    const err = await res.json();
                    throw new Error(err.error || "Échec de téléversement");
                }
                const v = await res.json();
                showNotifyModal(
                    "Nouvelle Version Enregistrée (Check-in)",
                    `La version <strong>${v.version_label}</strong> a été archivée avec succès. Le document a été déverrouillé et l'aperçu ALCOA+ a été mis à jour.`,
                    "✅"
                );
                openGEDVersionsModal(docID);
                selectedGEDDocId = docID;
                await loadGEDDocuments(currentGEDTaxonomyPath);
                selectGEDDocument(docID);
            } catch (err) {
                showNotifyModal("Erreur de Versioning", err.message, "⚠️");
            }
        }

        async function handleCheckOutGED(docID) {
            selectGEDDocument(docID);
            try {
                const res = await fetch(`/api/v1/ged/documents/${docID}/checkout`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ reason: "Modification pour conformité BPF" })
                });
                if (!res.ok) {
                    const err = await res.json();
                    throw new Error(err.error || "Impossible de verrouiller");
                }
                await loadGEDDocuments(currentGEDTaxonomyPath);
                selectGEDDocument(docID);
            } catch (err) {
                showNotifyModal("Verrouillage Échoué", err.message, "⚠️");
            }
        }

        async function handleCancelCheckOutGED(docID) {
            selectGEDDocument(docID);
            try {
                const res = await fetch(`/api/v1/ged/documents/${docID}/cancel-checkout`, {
                    method: 'POST'
                });
                if (!res.ok) {
                    const err = await res.json();
                    throw new Error(err.error || "Impossible d'annuler");
                }
                await loadGEDDocuments(currentGEDTaxonomyPath);
                selectGEDDocument(docID);
            } catch (err) {
                showNotifyModal("Erreur", err.message, "⚠️");
            }
        }

        async function handleStatusChangeGED(docID, targetStatus) {
            selectGEDDocument(docID);
            try {
                const res = await fetch(`/api/v1/ged/documents/${docID}/status`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ target_status: targetStatus })
                });
                if (!res.ok) {
                    const err = await res.json();
                    throw new Error(err.error || "Transition refusée");
                }
                await loadGEDDocuments(currentGEDTaxonomyPath);
                selectGEDDocument(docID);
            } catch (err) {
                showNotifyModal("Transition Refusée", err.message, "⚠️");
            }
        }

        function openSignGEDDTModal(docID, versionID, docCode, title, sha) {
            // Guardrail client : Vérifier la permission DOC_APPROVE_PDT
            if (typeof hasPermission === 'function' && !hasPermission('DOC_APPROVE_PDT')) {
                showNotifyModal(
                    "Accès Refusé (Exclusivité 1.A.7)",
                    `<div style="display:flex; flex-direction:column; gap:8px;">
                        <div style="color:var(--accent-red); font-weight:700;">
                            🚫 Visa réservé exclusivement au Pharmacien Directeur Technique
                        </div>
                        <div style="font-size:11px; color:var(--text-main);">
                            Conformément à l'Arrêté ministériel du 10 décembre 2020 et au Décret 21-82, seul le Pharmacien DT titulaire de la décision nominative MIPH (1.A.7) peut apposer la signature d'approbation légale.
                        </div>
                    </div>`,
                    "💊",
                    "Fermer"
                );
                return;
            }

            selectGEDDocument(docID);
            document.getElementById('sign-ged-doc-id').value = docID;
            document.getElementById('sign-ged-version-id').value = versionID;
            document.getElementById('sign-ged-doc-name').value = `[${docCode}] ${title}`;
            document.getElementById('sign-ged-sha').value = sha;
            document.getElementById('sign-ged-password').value = '';
            document.getElementById('modal-sign-ged-dt').classList.add('open');
        }

        async function executeGEDSignDT() {
            const docID = document.getElementById('sign-ged-doc-id').value;
            const versionID = document.getElementById('sign-ged-version-id').value;
            const password = document.getElementById('sign-ged-password').value;

            if (!password) {
                showNotifyModal("Re-authentification Requise", "Veuillez saisir votre mot de passe pour sceller la pièce selon la norme 21 CFR Part 11.", "⚠️");
                return;
            }

            try {
                const res = await fetch(`/api/v1/ged/documents/${docID}/sign-dt`, {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({
                        version_id: versionID,
                        password: password,
                        declaration: "En qualité de Pharmacien Directeur Technique, j'atteste sur l'honneur l'exactitude, la véracité et la complétude des données du présent document conformément aux règles de Bonnes Pratiques de Fabrication (BPF) et à la réglementation pharmaceutique en vigueur."
                    })
                });

                if (!res.ok) {
                    const err = await res.json();
                    throw new Error(err.error || "Signature rejetée");
                }

                const data = await res.json();
                closeModal('modal-sign-ged-dt');
                showNotifyModal(
                    "Document Scellé et Signé Électroniquement",
                    `Le document a été scellé légalement par le Pharmacien Directeur Technique.<br>
                    Empreinte cryptographique de signature : <code style="word-break:break-all;">${data.signature_seal}</code>`,
                    "✅"
                );
                selectedGEDDocId = docID;
                await loadGEDDocuments(currentGEDTaxonomyPath);
                selectGEDDocument(docID);
            } catch (err) {
                showNotifyModal("Échec de Signature BPF", err.message, "⚠️");
            }
        }

        function downloadGEDVersion(docID, versionID) {
            // Guardrail client : Zero-Trust IT isolation pour les profils techniques
            if (typeof hasRole === 'function' && hasRole('admin_systeme_local', 'super_admin_editeur')) {
                showNotifyModal(
                    "Accès Refusé (PSSI Zero-Trust)",
                    `<div style="display:flex; flex-direction:column; gap:8px;">
                        <div style="color:var(--accent-red); font-weight:700;">
                            🚫 Téléchargement interdit pour les profils techniques IT
                        </div>
                        <div style="font-size:11px; color:var(--text-main);">
                            En vertu des règles de gouvernance Zero-Trust et de la protection des secrets industriels de formulation pharmaceutique, les comptes administrateurs système ne peuvent télécharger aucune pièce confidentielle.
                        </div>
                    </div>`,
                    "🛡️",
                    "Compris"
                );
                return;
            }

            if (typeof openDocumentViewerModal === 'function') {
                openDocumentViewerModal(docID, "Document GED", versionID);
            } else {
                let url = `/api/v1/ged/documents/${docID}/download`;
                if (versionID) url += `?version_id=${encodeURIComponent(versionID)}`;
                window.open(url, '_blank');
            }
        }
