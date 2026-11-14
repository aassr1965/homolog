// ============================================================================
// APPLICATION CORE, NAVIGATION & INTERFACE PHARMA EXPLORER
// ============================================================================

// Cibles de navigation pour la barre d'adresse (Ctrl+K)
        const addressLocations = [
            { icon: "💊", title: "Paracétamol 500mg Comprimé — Module 3 : Qualité CMC", path: "📁 GED > 01_Affaires_Reglementaires > CTD_Paracetamol > Module 3", target: "ctd-para-3" },
            { icon: "💊", title: "Paracétamol 500mg Comprimé — Module 1 : Administratif Algérie", path: "📁 GED > 01_Affaires_Reglementaires > CTD_Paracetamol > Module 1", target: "ctd-para-1" },
            { icon: "💊", title: "Paracétamol 500mg Comprimé — Module 2 : Résumés CTD", path: "📁 GED > 01_Affaires_Reglementaires > CTD_Paracetamol > Module 2", target: "ctd-para-2" },
            { icon: "💊", title: "Paracétamol 500mg Comprimé — Module 4 : Non-Clinique", path: "📁 GED > 01_Affaires_Reglementaires > CTD_Paracetamol > Module 4", target: "ctd-para-4" },
            { icon: "💊", title: "Paracétamol 500mg Comprimé — Module 5 : Clinique & BE", path: "📁 GED > 01_Affaires_Reglementaires > CTD_Paracetamol > Module 5", target: "ctd-para-5" },
            { icon: "💉", title: "Seringue 3 pièces stérile 5ml — Dossier Technique (Arrêté 10 mai 2021)", path: "📁 Dossiers_ANPP > DM_Classe_IIa > Seringue_3p_5ml", target: "dm-seringue" },
            { icon: "🩹", title: "Cathéter IV avec ailette — Régularisation Décret 20-324", path: "📁 Dossiers_ANPP > Regularisation_DM > Catheter_IV", target: "dm-catheter" },
            { icon: "📁", title: "GED BPF : 01. Affaires Réglementaires & Agréments MIPH", path: "📁 GED_Pharmaceutique > 01_Affaires_Reglementaires", target: "ged-01" },
            { icon: "📁", title: "GED BPF : 02. Assurance Qualité & BPF (Certificats, Audits)", path: "📁 GED_Pharmaceutique > 02_Assurance_Qualite_BPF", target: "ged-02" },
            { icon: "📁", title: "GED BPF : 03. Développement & Production (DMF, Méthodes)", path: "📁 GED_Pharmaceutique > 03_Developpement_Production", target: "ged-03" },
            { icon: "📁", title: "GED BPF : 04. Vigilance & Post-Market (PV, Matériovigilance)", path: "📁 GED_Pharmaceutique > 04_Vigilance_Post_Market", target: "ged-04" },
            { icon: "📁", title: "GED BPF : 05. Contrats & Tiers (Accords Qualité QTA)", path: "📁 GED_Pharmaceutique > 05_Contrats_Tiers", target: "ged-05" }
        ];

        // Documents d'exemple riches
        const mockDocuments = [
            {
                id: 1,
                name: "3.2.S.1_Nomenclature_et_Structure.pdf",
                code: "CTD-MOD3-S1",
                version: "v1.0",
                expiry: "31/12/2028",
                status: "status-ok",
                statusLabel: "VALIDÉ BPF",
                validDT: "Dr. Amine K. (10/11/2026)",
                sha: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
                size: "2.4 Mo",
                previewTitle: "MODULE 3.2.S.1 : NOMENCLATURE",
                previewContent: "Substance Active : PARACÉTAMOL (DCI)<br>Formule brute : C8H9NO2<br>Masse molaire : 151.16 g/mol<br>Fabricant d'API : Phalgaphar Site de Blida<br>Conforme Ph. Eur. 11ème édition."
            },
            {
                id: 2,
                name: "3.2.S.2_Procede_Fabrication_et_Controles.pdf",
                code: "CTD-MOD3-S2",
                version: "v1.0",
                expiry: "15/09/2027",
                status: "status-ok",
                statusLabel: "VALIDÉ BPF",
                validDT: "Dr. Amine K. (08/11/2026)",
                sha: "8f434346648f6b96df89dda901c5176b10a6d83961dd3c1ac88b59b2dc327aa4",
                size: "8.1 Mo",
                previewTitle: "MODULE 3.2.S.2 : SYNTHÈSE ET CONTRÔLES",
                previewContent: "Schéma de synthèse 4 étapes.<br>Paramètres critiques de contrôle en cours de fabrication (IPC).<br>Contrôle des solvants résiduels par CPG conforme ICH Q3C."
            },
            {
                id: 3,
                name: "3.2.P.1_Description_et_Composition_Comprime.pdf",
                code: "CTD-MOD3-P1",
                version: "v0.2",
                expiry: "-",
                status: "status-rev",
                statusLabel: "EN REVUE DT",
                validDT: "En attente signature",
                sha: "ca978112ca1bbdcafac231b39a23dc4da786eff8147c4e72b9807785afee48bb",
                size: "1.9 Mo",
                previewTitle: "MODULE 3.2.P.1 : FORMULE DU PRODUIT FINI",
                previewContent: "Dosage unitaire : 500 mg Paracétamol.<br>Excipients : Amidon de maïs, Povidone K30, Stéarate de magnésium, Talc purifié.<br>Comprimé sécable blanc, arête biseautée."
            },
            {
                id: 4,
                name: "3.2.P.5_Specifications_Controle_Produit_Fini.pdf",
                code: "CTD-MOD3-P5",
                version: "-",
                expiry: "-",
                status: "status-req",
                statusLabel: "MANQUANT REQUIS",
                validDT: "Non déposé",
                sha: "----------------------------------------------------------------",
                size: "0 Mo",
                previewTitle: "PIÈCE MANQUANTE OBLIGATOIRE",
                previewContent: "⚠️ Ce document est requis par l'arrêté MIPH de décembre 2020 pour l'enregistrement ANPP.<br><br>Veuillez téléverser les spécifications de libération et de péremption du produit fini."
            },
            {
                id: 5,
                name: "Certificat_BPF_Site_Fabrication_MIPH.pdf",
                code: "QUAL-BPF-2026",
                version: "v2.0",
                expiry: "20/06/2027",
                status: "status-ok",
                statusLabel: "VALIDÉ BPF",
                validDT: "Dr. Amine K. (01/10/2026)",
                sha: "4b227777d4dd1fc61c6f884f48641d02b4d121d3fd328cb08b5531fcacdabf8a",
                size: "4.7 Mo",
                previewTitle: "RÉPUBLIQUE ALGÉRIENNE - MINISTÈRE DE L'INDUSTRIE PHARMACEUTIQUE",
                previewContent: "CERTIFICAT DE BONNES PRATIQUES DE FABRICATION (BPF)<br>N° Ag/MIPH/2024/0982<br>Formes sèches (Comprimés, Gélules). Conforme Décret Exécutif 22-247."
            }
        ];

        let selectedDocId = 1;
        let isArabic = false;
        let paletteSelectedIndex = 0;
        let currentTheme = localStorage.getItem("sanad-theme") || "light";

        // Application initiale du thème
        applyTheme(currentTheme);

        function toggleTheme() {
            currentTheme = (currentTheme === "dark") ? "light" : "dark";
            localStorage.setItem("sanad-theme", currentTheme);
            applyTheme(currentTheme);
        }

        function applyTheme(theme) {
            document.documentElement.setAttribute("data-theme", theme);
            const icon = document.getElementById("theme-icon");
            const text = document.getElementById("theme-text");
            const loginIcon = document.getElementById("login-theme-icon");
            if (theme === "light") {
                if (icon) icon.innerText = "🌙";
                if (text) text.innerText = "Sombre";
                if (loginIcon) loginIcon.innerText = "🌙";
            } else {
                if (icon) icon.innerText = "☀️";
                if (text) text.innerText = "Clair";
                if (loginIcon) loginIcon.innerText = "☀️";
            }
        }

        // Mode Focus & Responsive Drawer Handlers (Jalon 1.6)
        function toggleFocusMode() {
            document.body.classList.toggle('focus-mode');
            const isFocus = document.body.classList.contains('focus-mode');
            const btn = document.getElementById('btn-focus-mode');
            if (btn) {
                btn.innerHTML = isFocus ? '✕ Quitter Focus' : '🎯 Focus';
            }
            localStorage.setItem('sanad-focus-mode', isFocus ? '1' : '0');
        }

        function toggleMobileTree(force) {
            const tree = document.getElementById('pane-tree');
            const backdrop = document.getElementById('tree-backdrop');
            if (!tree || !backdrop) return;
            const open = (typeof force === 'boolean') ? force : !tree.classList.contains('drawer-open');
            tree.classList.toggle('drawer-open', open);
            backdrop.classList.toggle('active', open);
        }

        function togglePreviewPane(force) {
            const preview = document.getElementById('pane-preview');
            const backdrop = document.getElementById('preview-backdrop');
            if (!preview) return;
            
            // Sur desktop et laptop >= 992px (ancré dans le flux, sans superposition ni backdrop bloquant)
            if (window.innerWidth >= 992) {
                const hidden = (typeof force === 'boolean') ? !force : !preview.classList.contains('hidden');
                preview.classList.toggle('hidden', hidden);
                if (backdrop) backdrop.classList.remove('active');
            } else {
                // Sur tablettes / mobiles (< 992px) : mode slide-over drawer
                const open = (typeof force === 'boolean') ? force : !preview.classList.contains('slide-open');
                preview.classList.toggle('slide-open', open);
                if (backdrop) backdrop.classList.toggle('active', open);
            }
        }

        // Gestionnaire universel des raccourcis clavier
        window.addEventListener('keydown', function(e) {
            // Ctrl + K ou Ctrl + L ou Alt + D : Ouvrir la Barre d'Adresse In-Place
            if ((e.ctrlKey || e.metaKey) && (e.key === 'k' || e.key === 'K' || e.key === 'l' || e.key === 'L')) {
                e.preventDefault();
                openAddressBar();
                return;
            }
            if (e.altKey && (e.key === 'd' || e.key === 'D')) {
                e.preventDefault();
                openAddressBar();
                return;
            }

            // Alt + P : Basculer le Volet de Prévisualisation (Preview Pane)
            if (e.altKey && (e.key === 'p' || e.key === 'P')) {
                e.preventDefault();
                togglePreviewPane();
                return;
            }

            // Ctrl + F : Focus sur le filtre de tableau
            if ((e.ctrlKey || e.metaKey) && (e.key === 'f' || e.key === 'F')) {
                e.preventDefault();
                closeAddressBar();
                const sb = document.getElementById('search-box');
                if (sb) { sb.focus(); sb.select(); }
                return;
            }

            // Échap : Fermer la barre d'adresse et toutes les modales
            if (e.key === 'Escape') {
                closeAddressBar();
                closeAllModals();
            }
        });

        // Ouvre la barre d'adresse directement dans l'en-tête (façon Windows Explorer)
        function openAddressBar() {
            closeAllModals();

            const display = document.getElementById("breadcrumb-display");
            const editBox = document.getElementById("address-input-box");
            const dropdown = document.getElementById("address-dropdown");
            const input = document.getElementById("address-input");

            display.style.display = "none";
            editBox.classList.add("active");
            dropdown.classList.add("open");

            input.value = "GED\\01_Affaires_Reglementaires\\CTD_Paracetamol\\Module_3_Qualite";
            filterAddressItems();

            setTimeout(() => {
                input.focus();
                input.select();
            }, 30);
        }

        function closeAddressBar() {
            const display = document.getElementById("breadcrumb-display");
            const editBox = document.getElementById("address-input-box");
            const dropdown = document.getElementById("address-dropdown");

            display.style.display = "flex";
            editBox.classList.remove("active");
            dropdown.classList.remove("open");
        }

        function filterAddressItems() {
            const query = document.getElementById("address-input").value.toLowerCase();
            const dropdown = document.getElementById("address-dropdown");
            dropdown.innerHTML = "";

            const filtered = addressLocations.filter(loc => 
                loc.title.toLowerCase().includes(query) || loc.path.toLowerCase().includes(query)
            );

            paletteSelectedIndex = 0;

            if (filtered.length === 0) {
                dropdown.innerHTML = `<div style="padding:12px; text-align:center; color:var(--text-dim); font-size:11px;">Aucun dossier ou chemin correspondant.</div>`;
                return;
            }

            filtered.forEach((loc, idx) => {
                const item = document.createElement("div");
                item.className = "addr-item" + (idx === 0 ? " selected" : "");
                item.id = `addr-item-${idx}`;
                item.onclick = (e) => {
                    e.stopPropagation();
                    selectAddressLocation(loc);
                };
                item.innerHTML = `
                    <span style="font-size:15px;">${loc.icon}</span>
                    <span style="font-weight:600;">${loc.title}</span>
                    <span class="item-path">${loc.path}</span>
                `;
                dropdown.appendChild(item);
            });
        }

        function handleAddressKeyDown(e) {
            const items = document.querySelectorAll(".addr-item");
            if (items.length === 0) {
                if (e.key === "Escape") closeAddressBar();
                return;
            }

            if (e.key === "ArrowDown") {
                e.preventDefault();
                items[paletteSelectedIndex].classList.remove("selected");
                paletteSelectedIndex = (paletteSelectedIndex + 1) % items.length;
                items[paletteSelectedIndex].classList.add("selected");
                items[paletteSelectedIndex].scrollIntoView({ block: "nearest" });
            } else if (e.key === "ArrowUp") {
                e.preventDefault();
                items[paletteSelectedIndex].classList.remove("selected");
                paletteSelectedIndex = (paletteSelectedIndex - 1 + items.length) % items.length;
                items[paletteSelectedIndex].classList.add("selected");
                items[paletteSelectedIndex].scrollIntoView({ block: "nearest" });
            } else if (e.key === "Enter") {
                e.preventDefault();
                items[paletteSelectedIndex].click();
            } else if (e.key === "Escape") {
                closeAddressBar();
            }
        }

        function selectAddressLocation(loc) {
            closeAddressBar();
            document.getElementById("bc-root").innerText = "GED";
            document.getElementById("bc-sec").innerText = "Affaires Réglementaires";
            document.getElementById("bc-prod").innerText = loc.title.split("—")[0].trim();
            document.getElementById("bc-mod").innerText = loc.title.includes("—") ? loc.title.split("—")[1].trim() : loc.title;
            renderTable();
        }

        let currentDomain = 'ref';
        var currentRefSection = 'establishments';

        function switchDomainModule(domain) {
            // Contrôle d'habilitation préventif côté UI (Jalon 1.2 & CDC 7.2.2)
            if (domain === 'admin' && typeof hasPermission === 'function' && !hasPermission('USER_MANAGE')) {
                showNotifyModal("Accès Refusé (RBAC)", "Votre profil actif ne dispose pas de la permission <code>USER_MANAGE</code> pour accéder à la Console d'Administration.", "🚫", "Fermer");
                return;
            }
            if (domain === 'ged' && typeof hasPermission === 'function' && !hasPermission('DOC_READ')) {
                showNotifyModal("Accès Refusé (Zero-Trust)", "Votre profil technique est assujetti aux règles Zero-Trust et ne dispose pas de l'habilitation <code>DOC_READ</code>.", "🚫", "Fermer");
                return;
            }
            // Cloisonnement strict pour Consultant Externe (CDC Section 7.2.2) : Seul l'univers CTD assigné est accessible
            if (typeof hasRole === 'function' && hasRole('consultant_externe', 'consultant_cro_externe') && domain !== 'ctd') {
                showNotifyModal(
                    "Accès Cloisonné Tiers (CRO / Expert)",
                    `<div style="display:flex; flex-direction:column; gap:8px;">
                        <div style="color:var(--accent-blue); font-weight:700;">
                            🤝 Périmètre d'intervention strictement restreint
                        </div>
                        <div style="font-size:11px; color:var(--text-main); line-height:1.5;">
                            Conformément à la politique de sécurité BPF et à l'article 7.2.2 du Cahier des Charges, les consultants externes n'ont pas accès à la GED transverse de l'entreprise, au Référentiel, ni aux Variations.<br>
                            Votre mandat est strictement confiné au <strong>dossier d'enregistrement contractuellement assigné</strong>.
                        </div>
                    </div>`,
                    "🔒",
                    "Compris"
                );
                return;
            }

            currentDomain = domain;
            
            // Mise à jour de l'état actif dans l'Activity Bar
            document.querySelectorAll('.act-btn').forEach(b => b.classList.remove('active'));
            const btn = document.getElementById('act-' + domain);
            if (btn) btn.classList.add('active');

            // Mise à jour de l'état actif dans la barre mobile tiroir
            document.querySelectorAll('.mob-dom-btn').forEach(b => b.classList.remove('active'));
            const mobBtn = document.getElementById('mob-act-' + domain);
            if (mobBtn) mobBtn.classList.add('active');

            // Mise à jour de l'état actif dans la Bottom Navigation Bar mobile
            document.querySelectorAll('.mob-nav-btn').forEach(b => b.classList.remove('active'));
            const mobNavBtn = document.getElementById('mobnav-' + domain);
            if (mobNavBtn) mobNavBtn.classList.add('active');

            // Fermeture auto de l'arbre sur mobile si ouvert
            if (window.innerWidth < 768) {
                toggleMobileTree(false);
            }

            // Titre contextuel de l'arborescence
            const titleEl = document.getElementById("tree-section-title");
            if (titleEl) {
                let icon = "🏢";
                let text = isArabic ? "المرجعية الصيدلانية" : "Référentiel Pharmaceutique";
                if (domain === 'ctd') {
                    icon = "📋";
                    text = isArabic ? "ملفات التسجيل ANPP (CTD)" : "Dossiers d'Enregistrement (CTD)";
                } else if (domain === 'deficiencies') {
                    icon = "⏱️";
                    text = isArabic ? "استوديو الاستكمالات ANPP" : "Studio de Compléments ANPP & Délais";
                } else if (domain === 'variations') {
                    icon = "🔄";
                    text = isArabic ? "تعديلات ما بعد التسجيل" : "Variations & Post-AMM (12 Items)";
                } else if (domain === 'dm') {
                    icon = "💉";
                    text = isArabic ? "المستلزمات الطبية (20-324)" : "Dispositifs Médicaux (Décret 20-324)";
                } else if (domain === 'ged') {
                    icon = "📁";
                    text = isArabic ? "التوثيق الصيدلاني BPF" : "GED Pharmaceutique & BPF";
                } else if (domain === 'admin') {
                    icon = "⚙️";
                    text = isArabic ? "إدارة المخبر والمصفوفة" : "Administration Labo & Matrice RBAC";
                } else if (domain === 'sys') {
                    icon = "🖥️";
                    text = isArabic ? "لوحة تحكم البنية التحتية" : "Cockpit IT & Infrastructure PSSI";
                }
                titleEl.innerHTML = `<span>${icon}</span> <span>${text}</span>`;
            }

            updateToolbarForDomain(domain);
            renderTreeForDomain(domain);

            // Chargement des données métier correspondantes
            if (domain === 'ref') {
                loadRefSection(currentRefSection || 'establishments');
            } else if (domain === 'ctd') {
                document.getElementById("bc-root").innerText = "DOSSIERS_ANPP";
                document.getElementById("bc-sec").innerText = "Guide_CTD_2025";
                document.getElementById("bc-prod").innerText = "Catalogue des Dossiers";
                document.getElementById("bc-mod").innerText = "Checklists Réglementaires";
                loadDossiers();
            } else if (domain === 'deficiencies') {
                document.getElementById("bc-root").innerText = "COMPLÉMENTS_ANPP";
                document.getElementById("bc-sec").innerText = "Veille_Réglementaire";
                document.getElementById("bc-prod").innerText = "Décret_20-325";
                document.getElementById("bc-mod").innerText = "Traçabilité_Délais_Légaux";
                if (typeof loadDeficienciesDashboard === 'function') {
                    loadDeficienciesDashboard();
                }
            } else if (domain === 'variations') {
                document.getElementById("bc-root").innerText = "POST-AMM";
                document.getElementById("bc-sec").innerText = "Variations";
                document.getElementById("bc-prod").innerText = "LD-DEPP07/V01";
                document.getElementById("bc-mod").innerText = "12 Items Réglementaires";
                renderVariationsTable();
            } else if (domain === 'dm') {
                loadRefSection('devices');
            } else if (domain === 'ged') {
                document.getElementById("bc-root").innerText = "COFFRE-FORT";
                document.getElementById("bc-sec").innerText = "GED_BPF";
                document.getElementById("bc-prod").innerText = "Taxonomie";
                document.getElementById("bc-mod").innerText = "Section 8 du CDC";
                currentGEDTaxonomyPath = '';
                loadGEDTaxonomy();
                loadGEDDocuments();
            } else if (domain === 'admin') {
                document.getElementById("bc-root").innerText = "ADMINISTRATION";
                document.getElementById("bc-sec").innerText = "Habilitations_RBAC";
                document.getElementById("bc-prod").innerText = "Collaborateurs & Matrice";
                document.getElementById("bc-mod").innerText = "Conformité PSSI & BPF";
                if (typeof loadAdminSection === 'function') {
                    loadAdminSection(currentAdminSection || 'users');
                }
            } else if (domain === 'sys') {
                document.getElementById("bc-root").innerText = "INFRASTRUCTURE_IT";
                document.getElementById("bc-sec").innerText = "On-Premise_BPF";
                document.getElementById("bc-prod").innerText = "PostgreSQL & CAS";
                document.getElementById("bc-mod").innerText = "Haute Disponibilité";
                if (typeof loadSystemSection === 'function') {
                    loadSystemSection(currentSystemSection || 'status');
                }
            }

            // Réapplication dynamique des permissions UI sur les nouveaux boutons injectés
            if (typeof applyPermissionsUI === 'function') {
                applyPermissionsUI();
            }
        }

        function updateToolbarForDomain(domain) {
            const tb = document.getElementById("main-toolbar");
            if (!tb) return;
            if (domain === 'ref') {
                tb.innerHTML = `
                    <button class="btn-pill" data-perm="REF_EDIT" style="background:var(--accent-blue); color:white; border:none;" onclick="openNewPharmaModal()">➕ Nouveau</button>
                    <button class="btn-pill" onclick="loadRefSection(currentRefSection)">🔄 Actualiser</button>
                    <button class="btn-pill" data-perm="AUDIT_VIEW" onclick="openAuditModal()">🛡️ Audit Trail</button>
                    <button class="btn-pill" onclick="togglePreviewPane()">👁️ Volet Aperçu (Alt+P)</button>
                    <div style="margin-left:auto; font-size:11px; color:var(--text-muted);" id="dossier-progress">
                        Base Référentiel : <strong style="color:var(--accent-emerald);">PostgreSQL Connecté</strong>
                    </div>
                `;
            } else if (domain === 'ctd') {
                tb.innerHTML = `
                    <button class="btn-pill" id="tb-btn-new-dossier" data-perm="DOSSIER_EDIT" style="background:var(--accent-blue); color:white; border:none;" onclick="openNewDossierModal()">➕ Nouveau Dossier</button>
                    <button class="btn-pill" id="tb-btn-members" onclick="openDossierMembersModal()">👥 Équipe &amp; Secret</button>
                    <button class="btn-pill" id="tb-btn-sign-pdt" onclick="openSignPDTModal()">✍️ Visa PDT (21 CFR Part 11)</button>
                    <button class="btn-pill" id="tb-btn-deposit-anpp" onclick="openRecordDepositModal()">🏛️ Enregistrer Dépôt ANPP</button>
                    <button class="btn-pill" id="tb-btn-fees" onclick="openGenerateFeeSlipModal(selectedDossierId)">💰 Bordereau BV1/BV2</button>
                    <button class="btn-pill" id="tb-btn-schedules" onclick="openFeeSchedulesModal()">⚖️ Barème ANPP</button>
                    <button class="btn-pill" id="tb-btn-deficiency" onclick="openNewDeficiencyLetterModal(selectedDossierId)">⏱️ Notifier Complément ANPP</button>
                    <button class="btn-pill" id="tb-btn-export" data-perm="EXPORT_LEGAL" onclick="simulateExportZip()">📦 Exporter CTD (ZIP)</button>
                    <button class="btn-pill" onclick="loadDossiers()">🔄 Actualiser</button>
                    <button class="btn-pill" data-perm="AUDIT_VIEW" onclick="openAuditModal()">🛡️ Audit Trail</button>
                    <button class="btn-pill" onclick="togglePreviewPane()">👁️ Volet Aperçu (Alt+P)</button>
                    <div style="margin-left:auto; font-size:11px; color:var(--text-muted);" id="dossier-progress">
                        Complétude : <strong id="kpi-dossier-completeness" style="color:var(--accent-blue);">0%</strong>
                    </div>
                `;
            } else if (domain === 'deficiencies') {
                tb.innerHTML = `
                    <button class="btn-pill" data-perm="DOSSIER_EDIT" style="background:var(--accent-blue); color:white; border:none;" onclick="openNewDeficiencyLetterModal()">➕ Enregistrer Complément</button>
                    <button class="btn-pill" onclick="loadDeficienciesDashboard()">🔄 Actualiser</button>
                    <button class="btn-pill" data-perm="AUDIT_VIEW" onclick="openAuditModal()">🛡️ Audit Trail</button>
                    <button class="btn-pill" onclick="togglePreviewPane()">👁️ Volet Aperçu (Alt+P)</button>
                    <div style="margin-left:auto; font-size:11px; color:var(--text-muted);" id="dossier-progress">
                        Délais Légaux : <strong style="color:var(--accent-blue);">Compte à Rebours J-X (Guide 2025)</strong>
                    </div>
                `;
            } else if (domain === 'variations') {
                tb.innerHTML = `
                    <button class="btn-pill" data-perm="VARIATION_MANAGE" style="background:var(--accent-amber); color:white; border:none;" onclick="showNotifyModal('Dépôt de Variation ANPP', 'Le formulaire de dépôt selon les 12 items de la directive LD-DEPP07/V01 permet d\\'enregistrer les modifications Mmin, Mmod ou Mmaj avec pièces justificatives régionales 3.2.R.', '🔄')">➕ Déposer une Variation</button>
                    <button class="btn-pill" onclick="renderVariationsTable()">🔄 Actualiser</button>
                    <button class="btn-pill" data-perm="AUDIT_VIEW" onclick="openAuditModal()">🛡️ Audit Trail</button>
                    <button class="btn-pill" onclick="togglePreviewPane()">👁️ Volet Aperçu (Alt+P)</button>
                    <div style="margin-left:auto; font-size:11px; color:var(--text-muted);" id="dossier-progress">
                        Directive : <strong>ANPP LD-DEPP07/V01</strong>
                    </div>
                `;
            } else if (domain === 'dm') {
                tb.innerHTML = `
                    <button class="btn-pill" data-perm="REF_EDIT" style="background:var(--accent-blue); color:white; border:none;" onclick="document.getElementById('modal-new-dm').classList.add('open')">➕ Nouveau DM</button>
                    <button class="btn-pill" onclick="loadRefSection('devices')">🔄 Actualiser</button>
                    <button class="btn-pill" data-perm="AUDIT_VIEW" onclick="openAuditModal()">🛡️ Audit Trail</button>
                    <button class="btn-pill" onclick="togglePreviewPane()">👁️ Volet Aperçu (Alt+P)</button>
                    <div style="margin-left:auto; font-size:11px; color:var(--text-muted);" id="dossier-progress">
                        Cadre : <strong>Arrêté du 10 Mai 2021</strong>
                    </div>
                `;
            } else if (domain === 'ged') {
                tb.innerHTML = `
                    <button class="btn-pill" data-perm="DOC_WRITE" style="background:var(--accent-blue); color:white; border:none;" onclick="openNewGEDDocModal()">➕ Déposer Document</button>
                    <select id="ged-filter-status" onchange="filterGEDByStatus(this.value)" class="btn-pill" style="background:var(--bg-surface); color:var(--text-main); border:1px solid var(--border-color); font-size:11px;">
                        <option value="">Tous les statuts</option>
                        <option value="en_vigueur">En Vigueur (Approuvé)</option>
                        <option value="en_approbation_dt">En Approbation DT</option>
                        <option value="en_revue">En Revue</option>
                        <option value="brouillon">Brouillon</option>
                        <option value="obsolete">Obsolète</option>
                    </select>
                    <button class="btn-pill" onclick="loadGEDDocuments(currentGEDTaxonomyPath)">🔄 Actualiser</button>
                    <button class="btn-pill" data-perm="AUDIT_VIEW" onclick="openAuditModal()">🛡️ Audit Trail</button>
                    <button class="btn-pill" onclick="togglePreviewPane()">👁️ Volet Aperçu (Alt+P)</button>
                    <div style="margin-left:auto; font-size:11px; color:var(--text-muted);" id="dossier-progress">
                        Coffre-Fort CAS : <strong style="color:var(--accent-emerald);">Stockage Immuable SHA-256</strong>
                    </div>
                `;
            } else if (domain === 'admin') {
                tb.innerHTML = `
                    <button class="btn-pill" data-perm="USER_MANAGE" style="background:var(--accent-blue); color:white; border:none;" onclick="openNewUserModal()">➕ Inviter Collaborateur</button>
                    <button class="btn-pill" onclick="loadAdminSection(currentAdminSection)">🔄 Actualiser</button>
                    <button class="btn-pill" data-perm="USER_MANAGE" onclick="exportRBACMatrixCSV()">📊 Exporter Matrice des Droits</button>
                    <button class="btn-pill" data-perm="AUDIT_VIEW" onclick="openAuditModal()">🛡️ Audit Trail</button>
                    <button class="btn-pill" onclick="togglePreviewPane()">👁️ Volet Aperçu (Alt+P)</button>
                    <div style="margin-left:auto; font-size:11px; color:var(--text-muted);" id="dossier-progress">
                        PSSI Labo : <strong style="color:var(--accent-emerald);">Contrôle d'Accès ALCOA+ & SoD</strong>
                    </div>
                `;
            } else if (domain === 'sys') {
                tb.innerHTML = `
                    <button class="btn-pill" style="background:var(--accent-blue); color:white; border:none;" onclick="loadSystemMetrics()">🔄 Actualiser Métriques</button>
                    <button class="btn-pill" style="background:var(--accent-emerald); color:white; border:none;" onclick="triggerSystemBackup()">💾 Instantané Sauvegarde</button>
                    <button class="btn-pill" onclick="triggerCASIntegrityCheck()">🔍 Contrôle Intégrité CAS</button>
                    <button class="btn-pill" style="border-color:#ef4444; color:#ef4444;" onclick="openEmergencyLockdownModal()">🚨 Verrouillage Urgence</button>
                    <button class="btn-pill" onclick="togglePreviewPane()">👁️ Volet Aperçu (Alt+P)</button>
                    <div style="margin-left:auto; font-size:11px; color:var(--text-muted);" id="dossier-progress">
                        Infrastructure : <strong style="color:var(--accent-emerald);">PostgreSQL 16+ &amp; CAS SHA-256</strong>
                    </div>
                `;
            }
        }

        function renderTreeForDomain(domain) {
            const container = document.getElementById("tree-root");
            if (domain === 'sys') {
                if (typeof renderSystemTree === 'function') {
                    renderSystemTree();
                }
                return;
            }
            if (domain === 'ref') {
                container.innerHTML = `
                    <div class="tree-node ${currentRefSection === 'establishments' ? 'selected' : ''}" onclick="loadRefSection('establishments')">
                        <span class="chevron">🏢</span>
                        <span>Établissements Agréés (MIPH)</span>
                        <span class="count-badge ok">Décret 20-325</span>
                    </div>
                    <div class="tree-node ${currentRefSection === 'sites' ? 'selected' : ''}" onclick="loadRefSection('sites')">
                        <span class="chevron">🏭</span>
                        <span>Sites Intervenants (1.H.8)</span>
                        <span class="count-badge ok">BPF / GMP</span>
                    </div>
                    <div class="tree-node ${currentRefSection === 'medicines' ? 'selected' : ''}" onclick="loadRefSection('medicines')">
                        <span class="chevron">💊</span>
                        <span>Catalogue Médicaments</span>
                        <span class="count-badge warn">Princeps/Génér.</span>
                    </div>
                    <div class="tree-node ${currentRefSection === 'devices' ? 'selected' : ''}" onclick="loadRefSection('devices')">
                        <span class="chevron">💉</span>
                        <span>Répertoire Dispositifs Médicaux</span>
                        <span class="count-badge req">Arrêté 10 Mai</span>
                    </div>
                `;
            } else if (domain === 'ctd') {
                renderDossiersTree();
            } else if (domain === 'deficiencies') {
                container.innerHTML = `
                    <div class="tree-node selected" onclick="loadDeficienciesDashboard()">
                        <span class="chevron">⏱️</span>
                        <span>Toutes les Notifications ANPP</span>
                        <span class="count-badge req">Guide 2025</span>
                    </div>
                    <div class="tree-node" style="padding-left:24px;" onclick="loadDeficienciesDashboard()">
                        <span>🚨 Alertes Critiques (J ≤ 15)</span>
                    </div>
                    <div class="tree-node" style="padding-left:24px;" onclick="loadDeficienciesDashboard()">
                        <span>⚠️ Urgences Réglementaires (J ≤ 30)</span>
                    </div>
                    <div class="tree-node" style="padding-left:24px;" onclick="loadDeficienciesDashboard()">
                        <span>✍️ Visées Pharmacien DT (1.A.7)</span>
                    </div>
                `;
            } else if (domain === 'variations') {
                container.innerHTML = `
                    <div class="tree-node selected" onclick="toggleSub('sub-var-active')">
                        <span class="chevron">▼</span>
                        <span>🔄 Variations en Cours (ANPP)</span>
                        <span class="count-badge warn">3</span>
                    </div>
                    <div class="tree-sub open" id="sub-var-active">
                        <div class="tree-node" style="padding-left:24px;" onclick="renderVariationsTable()"><span>🟡 Mmin (Mineures - 7.b.1)</span><span class="count-badge ok">1</span></div>
                        <div class="tree-node" style="padding-left:24px;" onclick="renderVariationsTable()"><span>🟠 Mmod (Modérées - 12.a)</span><span class="count-badge ok">1</span></div>
                        <div class="tree-node" style="padding-left:24px;" onclick="renderVariationsTable()"><span>🔴 Mmaj (Majeures - 7.b.2)</span><span class="count-badge req">1</span></div>
                    </div>

                    <div class="tree-node" onclick="toggleSub('sub-var-history')">
                        <span class="chevron">▶</span>
                        <span>📅 Renouvellements Quinquennaux</span>
                        <span class="count-badge ok">2026-2027</span>
                    </div>
                `;
            } else if (domain === 'dm') {
                container.innerHTML = `
                    <div class="tree-node selected" onclick="loadRefSection('devices')">
                        <span class="chevron">💉</span>
                        <span>Tous les Dispositifs Homologués</span>
                        <span class="count-badge ok">Décret 20-324</span>
                    </div>
                    <div class="tree-node" style="padding-left:24px;" onclick="loadRefSection('devices')"><span>📁 Classe I (Non invasifs)</span></div>
                    <div class="tree-node selected" style="padding-left:24px;" onclick="loadRefSection('devices')"><span>📁 Classe IIa (Seringues, Canules)</span></div>
                    <div class="tree-node" style="padding-left:24px;" onclick="loadRefSection('devices')"><span>📁 Classe IIb & III (Implants)</span></div>
                `;
            } else if (domain === 'ged') {
                renderGEDTree();
            } else if (domain === 'admin') {
                if (typeof renderAdminTree === 'function') {
                    renderAdminTree();
                } else {
                    container.innerHTML = `
                        <div class="tree-node ${currentAdminSection === 'users' ? 'selected' : ''}" onclick="loadAdminSection('users')">
                            <span class="chevron">👥</span>
                            <span>Collaborateurs & Rôles</span>
                            <span class="count-badge ok">Actifs</span>
                        </div>
                        <div class="tree-node ${currentAdminSection === 'matrix' ? 'selected' : ''}" onclick="loadAdminSection('matrix')">
                            <span class="chevron">🔐</span>
                            <span>Matrice des 14 Permissions</span>
                            <span class="count-badge ok">Tableau BPF</span>
                        </div>
                        <div class="tree-node ${currentAdminSection === 'policies' ? 'selected' : ''}" onclick="loadAdminSection('policies')">
                            <span class="chevron">🛡️</span>
                            <span>Règles SoD & Zero-Trust</span>
                            <span class="count-badge req">Guardrails</span>
                        </div>
                    `;
                }
            }
        }


        function toggleSub(id) {
            const el = document.getElementById(id);
            if (el) el.classList.toggle("open");
        }

        function renderTable(docs = mockDocuments) {
            const thead = document.getElementById("table-head");
            if (thead) {
                thead.innerHTML = `
                    <tr>
                        <th style="width: 80px;">Statut</th>
                        <th>Nom de la pièce</th>
                        <th>Code Référence</th>
                        <th style="width: 70px;">Version</th>
                        <th>Échéance</th>
                        <th>Validation DT</th>
                        <th>Empreinte SHA-256</th>
                    </tr>
                `;
            }
            const tbody = document.getElementById("doc-tbody");
            tbody.innerHTML = "";
            docs.forEach(doc => {
                const tr = document.createElement("tr");
                if (doc.id === selectedDocId) tr.classList.add("selected");
                tr.onclick = () => selectDocument(doc.id);
                tr.innerHTML = `
                    <td><span class="status-tag ${doc.status}">${doc.statusLabel}</span></td>
                    <td style="font-weight:600; color:var(--text-main);">${doc.name}</td>
                    <td style="color:var(--text-muted); font-family:monospace;">${doc.code}</td>
                    <td style="font-weight:600;">${doc.version}</td>
                    <td style="color:${doc.expiry !== '-' ? 'var(--accent-amber)' : 'var(--text-dim)'};">${doc.expiry}</td>
                    <td>${doc.validDT}</td>
                    <td style="font-family:monospace; color:var(--text-dim); font-size:10px;">${doc.sha.substring(0, 16)}...</td>
                `;
                tbody.appendChild(tr);
            });
            renderPreview();
        }

        function selectDocument(id) {
            selectedDocId = id;
            renderTable();
        }

        function renderPreview() {
            const doc = mockDocuments.find(d => d.id === selectedDocId) || mockDocuments[0];
            const preview = document.getElementById("preview-body");
            const isNeedSign = doc.status === "status-rev";

            preview.innerHTML = `
                <div class="pdf-sim-page">
                    <div class="pdf-sim-header">
                        ${doc.previewTitle}
                    </div>
                    <div style="font-size:11px; margin-bottom:12px;">
                        ${doc.previewContent}
                    </div>
                    <div style="margin-top:20px; padding-top:8px; border-top:1px dashed var(--border-color); font-size:10px; color:var(--text-dim);">
                        Document certifié conforme pour l'Agence Nationale des Produits Pharmaceutiques (ANPP).<br>
                        Empreinte cryptographique : <code>${doc.sha.substring(0, 24)}...</code>
                    </div>
                </div>

                <div class="meta-box">
                    <div class="meta-row"><span class="meta-label">Nom du fichier :</span><span class="meta-val" style="word-break:break-all;">${doc.name}</span></div>
                    <div class="meta-row"><span class="meta-label">Version actuelle :</span><span class="meta-val">${doc.version}</span></div>
                    <div class="meta-row"><span class="meta-label">Taille :</span><span class="meta-val">${doc.size}</span></div>
                    <div class="meta-row"><span class="meta-label">Validité légale :</span><span class="meta-val">${doc.expiry}</span></div>
                    <div class="meta-row"><span class="meta-label">Statut BPF :</span><span class="meta-val">${doc.statusLabel}</span></div>
                    <div class="meta-row"><span class="meta-label">SHA-256 complet :</span><span class="meta-val" style="font-size:9px; word-break:break-all;">${doc.sha}</span></div>
                </div>

                ${isNeedSign ? `
                    <button class="btn-sign-dt" onclick="openSignModal('${doc.name}')">
                        ✍️ Signer et Approuver en qualité de DT (BPF)
                    </button>
                ` : `
                    <div style="text-align:center; padding:8px; background:var(--tag-ok-bg); border:1px solid var(--tag-ok-border); border-radius:4px; color:var(--tag-ok-text); font-size:11px; font-weight:600;">
                        ✓ Document visé et scellé par le Directeur Technique
                    </div>
                `}
            `;
        }

        function filterTable() {
            const query = document.getElementById("search-box").value.toLowerCase();
            if (currentDomain === 'ged') {
                const filtered = currentGEDDocuments.filter(d => 
                    (d.title_fr && d.title_fr.toLowerCase().includes(query)) || 
                    (d.title_ar && d.title_ar.toLowerCase().includes(query)) ||
                    (d.document_code && d.document_code.toLowerCase().includes(query)) ||
                    (d.status && d.status.toLowerCase().includes(query))
                );
                renderGEDDocumentsTable(filtered);
                return;
            }
            const filtered = mockDocuments.filter(d => 
                d.name.toLowerCase().includes(query) || 
                d.code.toLowerCase().includes(query) ||
                d.statusLabel.toLowerCase().includes(query)
            );
            renderTable(filtered);
        }

        function toggleLanguage() {
            isArabic = !isArabic;
            const html = document.documentElement;

            const loginTitle = document.getElementById("login-title");
            const loginSub = document.getElementById("login-subtitle");
            const lblEmail = document.getElementById("lbl-login-email");
            const lblPwd = document.getElementById("lbl-login-password");
            const btnLogin = document.getElementById("btn-login-submit");
            const loginLang = document.getElementById("login-lang-label");

            if (isArabic) {
                html.setAttribute("dir", "rtl");
                html.setAttribute("lang", "ar");
                document.getElementById("lang-btn-label").innerText = "🇫🇷 Français (LTR)";
                document.getElementById("txt-app-title").innerText = "لوحة القيادة ANPP وإدارة الوثائق الصيدلانية";
                document.getElementById("txt-urgent-alerts").innerText = "🚨 2 مواعيد حرجة (أقل من 30 يوماً)";
                document.getElementById("user-badge").innerText = currentUser ? `${currentUser.first_name} ${currentUser.last_name} (${formatRole(currentUser.role)})` : "د. أمين ك. (صيدلي مدير تقني)";
                document.getElementById("txt-goto-btn").innerText = "انتقال إلى...";

                if (loginTitle) loginTitle.innerText = "بوابة الولوج التنظيمي والمصادقة";
                if (loginSub) loginSub.innerText = "لوحة القيادة ANPP وإدارة الوثائق الصيدلانية BPF";
                if (lblEmail) lblEmail.innerText = "البريد الإلكتروني الاسمي :";
                if (lblPwd) lblPwd.innerText = "كلمة المرور المشفرة (Argon2id) :";
                if (btnLogin) btnLogin.innerText = "تسجيل الدخول إلى النظام ➔";
                if (loginLang) loginLang.innerText = "🇫🇷";

                const lRef = document.getElementById("lbl-act-ref");
                const lCtd = document.getElementById("lbl-act-ctd");
                const lVar = document.getElementById("lbl-act-var");
                const lDm = document.getElementById("lbl-act-dm");
                const lGed = document.getElementById("lbl-act-ged");
                const lAud = document.getElementById("lbl-act-audit");
                if (lRef) lRef.innerText = "المرجعية";
                if (lCtd) lCtd.innerText = "ملفات CTD";
                if (lVar) lVar.innerText = "التعديلات";
                if (lDm) lDm.innerText = "مستلزمات";
                if (lGed) lGed.innerText = "وثائق BPF";
                if (lAud) lAud.innerText = "التدقيق";
            } else {
                html.setAttribute("dir", "ltr");
                html.setAttribute("lang", "fr");
                document.getElementById("lang-btn-label").innerText = "🇩🇿 العربية (RTL)";
                document.getElementById("txt-app-title").innerText = "Cockpit ANPP & GED BPF";
                document.getElementById("txt-urgent-alerts").innerText = "🚨 2 Échéances (J < 30j)";
                document.getElementById("user-badge").innerText = currentUser ? `${currentUser.first_name} ${currentUser.last_name} (${formatRole(currentUser.role)})` : "Dr. Amine K. (Pharmacien DT)";
                document.getElementById("txt-goto-btn").innerText = "Aller à...";

                if (loginTitle) loginTitle.innerText = "Portail d'Accès Réglementaire";
                if (loginSub) loginSub.innerText = "Cockpit ANPP & Coffre-Fort GED Pharmaceutique Souverain";
                if (lblEmail) lblEmail.innerText = "Adresse e-mail nominative (Habilitation BPF) :";
                if (lblPwd) lblPwd.innerText = "Mot de passe sécurisé (Argon2id) :";
                if (btnLogin) btnLogin.innerText = "Se connecter au cockpit ➔";
                if (loginLang) loginLang.innerText = "🇩🇿";

                const lRef = document.getElementById("lbl-act-ref");
                const lCtd = document.getElementById("lbl-act-ctd");
                const lVar = document.getElementById("lbl-act-var");
                const lDm = document.getElementById("lbl-act-dm");
                const lGed = document.getElementById("lbl-act-ged");
                const lAud = document.getElementById("lbl-act-audit");
                if (lRef) lRef.innerText = "Référentiel";
                if (lCtd) lCtd.innerText = "Dossiers CTD";
                if (lVar) lVar.innerText = "Variations";
                if (lDm) lDm.innerText = "Disp. Méd.";
                if (lGed) lGed.innerText = "GED BPF";
                if (lAud) lAud.innerText = "Audit";
            }
            switchDomainModule(currentDomain);
            applyTheme(currentTheme); // Met à jour le texte du bouton de thème dans la bonne langue
        }

        // Piste d'audit trail ALCOA+ en mémoire

        function closeAllModals() {
            document.querySelectorAll('.modal-overlay').forEach(m => m.classList.remove('open'));
        }

        function closeModal(id) {
            const el = document.getElementById(id);
            if (el) el.classList.remove("open");
        }

        function showNotifyModal(title, contentHtml, icon = "ℹ️", btnText = "Compris") {
            closeAllModals();
            document.getElementById("notify-title").innerText = title;
            document.getElementById("notify-icon").innerText = icon;
            document.getElementById("notify-body").innerHTML = contentHtml;
            document.getElementById("notify-btn-ok").innerText = btnText;
            document.getElementById("modal-notify").classList.add("open");
        }


        function goToLocation(targetKey) {
            const loc = addressLocations.find(l => l.target === targetKey);
            if (loc) {
                selectAddressLocation(loc);
            }
        }

        function resetToRoot() {
            closeAddressBar();
            closeAllModals();
            selectAddressLocation(addressLocations[0]);
        }

        function initSplitters() {
            const splitLeft = document.getElementById("splitter-left");
            const splitRight = document.getElementById("splitter-right");
            const paneTree = document.getElementById("pane-tree");
            const panePreview = document.getElementById("pane-preview");

            // Rétablir largeurs mémorisées
            const savedLeft = localStorage.getItem("agy_splitter_left_width");
            if (savedLeft && paneTree) paneTree.style.width = savedLeft;

            const savedRight = localStorage.getItem("agy_splitter_right_width");
            if (savedRight && panePreview) panePreview.style.width = savedRight;

            // Splitter Gauche
            if (splitLeft && paneTree) {
                let isDraggingLeft = false;
                splitLeft.addEventListener("mousedown", (e) => {
                    isDraggingLeft = true;
                    document.body.style.cursor = "col-resize";
                    document.body.style.userSelect = "none";
                });

                splitLeft.addEventListener("dblclick", () => {
                    paneTree.style.width = "340px";
                    localStorage.removeItem("agy_splitter_left_width");
                });

                window.addEventListener("mousemove", (e) => {
                    if (!isDraggingLeft) return;
                    const activityBarWidth = 82;
                    const newWidth = Math.max(240, Math.min(600, e.clientX - activityBarWidth));
                    paneTree.style.width = `${newWidth}px`;
                    localStorage.setItem("agy_splitter_left_width", `${newWidth}px`);
                });

                window.addEventListener("mouseup", () => {
                    if (isDraggingLeft) {
                        isDraggingLeft = false;
                        document.body.style.cursor = "";
                        document.body.style.userSelect = "";
                    }
                });
            }

            // Splitter Droit
            if (splitRight && panePreview) {
                let isDraggingRight = false;
                splitRight.addEventListener("mousedown", (e) => {
                    isDraggingRight = true;
                    document.body.style.cursor = "col-resize";
                    document.body.style.userSelect = "none";
                });

                splitRight.addEventListener("dblclick", () => {
                    panePreview.style.width = "380px";
                    localStorage.removeItem("agy_splitter_right_width");
                });

                window.addEventListener("mousemove", (e) => {
                    if (!isDraggingRight) return;
                    const newWidth = Math.max(280, Math.min(700, window.innerWidth - e.clientX));
                    panePreview.style.width = `${newWidth}px`;
                    localStorage.setItem("agy_splitter_right_width", `${newWidth}px`);
                });

                window.addEventListener("mouseup", () => {
                    if (isDraggingRight) {
                        isDraggingRight = false;
                        document.body.style.cursor = "";
                        document.body.style.userSelect = "";
                    }
                });
            }
        }

        // Horloge UTC
        setInterval(() => {
            document.getElementById("clock-utc").innerText = new Date().toUTCString().split(" ")[4];
        }, 1000);

        // Initialisation des splitters glissants
        initSplitters();
