// ============================================================================
// APPLICATION CORE, NAVIGATION & INTERFACE PHARMA EXPLORER
// ============================================================================

// Nettoyage esthétique de l'URL post-authentification / post-sas (?auth=... & ?gate=...)
(function cleanQueryAuthParam() {
    try {
        if (window.location.search && (window.location.search.includes('auth=') || window.location.search.includes('gate='))) {
            const url = new URL(window.location.href);
            url.searchParams.delete('auth');
            url.searchParams.delete('gate');
            const cleanSearch = url.searchParams.toString();
            const cleanUrl = url.pathname + (cleanSearch ? '?' + cleanSearch : '') + url.hash;
            window.history.replaceState({}, document.title, cleanUrl);
        }
    } catch (e) {
        // En cas d'environnement restreint sans Support URL/history
    }
})();

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

        // Mode Focus & Responsive Drawer Handlers
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
            if (!loc) return;
            if (loc.target && loc.target.startsWith("ctd-para-")) {
                const modNum = parseInt(loc.target.replace("ctd-para-", ""), 10) || 1;
                if (typeof openDossierModule === 'function') {
                    openDossierModule('parac', modNum);
                    return;
                }
            } else if (loc.target && loc.target.startsWith("dm-")) {
                const hint = loc.target.replace("dm-", "");
                if (typeof openDossierModule === 'function') {
                    openDossierModule(hint, 0);
                    return;
                }
            } else if (loc.target && loc.target.startsWith("ged-")) {
                if (typeof switchDomainModule === 'function') {
                    switchDomainModule('ged');
                    return;
                }
            }
            document.getElementById("bc-root").innerText = "GED";
            document.getElementById("bc-sec").innerText = "Affaires Réglementaires";
            document.getElementById("bc-prod").innerText = loc.title.split("—")[0].trim();
            document.getElementById("bc-mod").innerText = loc.title.includes("—") ? loc.title.split("—")[1].trim() : loc.title;
            renderTable();
        }

        let currentDomain = 'ref';
        var currentRefSection = 'establishments';

        function resetStandardTableContainer(force = false) {
            const container = document.getElementById("table-container") || document.querySelector('.table-container');
            if (!container) return;
            const table = document.getElementById("doc-table");
            const tbody = document.getElementById("doc-tbody");
            const thead = document.getElementById("table-head");
            // Si le conteneur contient des éléments parasites (Chemin critique CTD, E-Services, Compléments, DDM, tabs) ou n'a pas la table standard
            const hasParasites = container.querySelector('#dossier-gating-container') ||
                                 container.querySelector('#dossier-smart-planning-container') ||
                                 container.querySelector('#dossier-payments-container') ||
                                 container.querySelector('.deficiencies-dashboard-container') ||
                                 container.querySelector('.table-quick-filter-bar') ||
                                 container.querySelector('.kpi-gauge-bar') ||
                                 container.querySelector('.tabs-bar') ||
                                 container.children.length !== 1 ||
                                 (container.firstElementChild && container.firstElementChild.id !== 'doc-table');
            if (force || !table || !tbody || !thead || hasParasites) {
                container.innerHTML = `
                    <table id="doc-table">
                        <thead id="table-head"></thead>
                        <tbody id="doc-tbody">
                            <tr><td colspan="8" style="text-align:center; padding:30px; color:var(--text-muted);"><span class="spin">⏳</span> Chargement des données...</td></tr>
                        </tbody>
                    </table>
                `;
            }
        }
        window.resetStandardTableContainer = resetStandardTableContainer;

        function switchDomainModule(domain) {
            // Contrôle d'expiration de session réglementaire (BPF 15 min / 21 CFR Part 11)
            if (typeof isSessionTimedOut === 'function' && isSessionTimedOut()) {
                if (typeof showSessionTimeoutModal === 'function') {
                    showSessionTimeoutModal('navigation');
                }
                return;
            }
            // Contrôle d'habilitation préventif côté UI (CDC 7.2.2)
            if (domain === 'admin' && typeof hasPermission === 'function' && !hasPermission('USER_MANAGE') && !hasPermission('DOC_APPROVE_PDT')) {
                showNotifyModal("Accès Refusé (RBAC)", "Votre profil actif ne dispose pas des permissions requises (<code>USER_MANAGE</code> ou <code>DOC_APPROVE_PDT</code>) pour accéder aux Paramètres et à la Console d'Administration.", "🚫", "Fermer");
                return;
            }
            if (domain === 'ged' && typeof hasPermission === 'function' && !hasPermission('DOC_READ')) {
                showNotifyModal("Accès Refusé (Zero-Trust)", "Votre profil technique est assujetti aux règles Zero-Trust et ne dispose pas de l'habilitation <code>DOC_READ</code>.", "🚫", "Fermer");
                return;
            }
            // Garde-fou d'habilitation préventif côté UI (BPF / Need-to-Know)
            if ((domain === 'ctd' || domain === 'dm' || domain === 'deficiencies' || domain === 'variations') && typeof hasPermission === 'function' && !hasPermission('DOSSIER_VIEW')) {
                showNotifyModal(
                    "Accès Cloisonné (BPF / Need-to-Know)",
                    `<div style="display:flex; flex-direction:column; gap:8px;">
                        <div style="color:var(--accent-red); font-weight:700;">
                            🔒 Confidentialité Pharmaceutique & Secret Industriel
                        </div>
                        <div style="font-size:11px; color:var(--text-main); line-height:1.5;">
                            Conformément aux Bonnes Pratiques de Fabrication (BPF) et au principe du moindre privilège (<em>Need-to-Know</em>), l'accès aux dossiers techniques CTD et d'homologation des dispositifs médicaux requiert l'habilitation <code>DOSSIER_VIEW</code>.<br>
                            Votre profil administratif est strictement cantonné à la gestion des collaborateurs, des barèmes et de la gouvernance de l'établissement.
                        </div>
                    </div>`,
                    "🔒",
                    "Compris"
                );
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
                            Conformément à la politique de sécurité BPF et aux règles de cloisonnement des tiers, les consultants externes ont un accès strictement restreint à leur périmètre d'expertise contractuel.<br>
                            Votre mandat est strictement confiné au <strong>dossier d'enregistrement contractuellement assigné</strong>.
                        </div>
                    </div>`,
                    "🔒",
                    "Compris"
                );
                return;
            }

            currentDomain = domain;
            if (domain !== 'ctd' && domain !== 'deficiencies') {
                resetStandardTableContainer(true);
            }
            
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
                if (domain === 'corporate') {
                    icon = "🏛️";
                    text = isArabic ? "اعتمادات المؤسسات الصيدلانية MIPH" : "Agréments MIPH & Tabadol (Décret 21-82)";
                } else if (domain === 'ctd') {
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
                resetStandardTableContainer();
                loadRefSection(currentRefSection || 'establishments');
            } else if (domain === 'corporate') {
                document.getElementById("bc-root").innerText = "SANAD_CORPORATE";
                document.getElementById("bc-sec").innerText = "MIPH_Décret_21-82";
                document.getElementById("bc-prod").innerText = "Agréments & Tabadol";
                document.getElementById("bc-mod").innerText = "Établissements Pharmaceutiques";
                resetStandardTableContainer();
                if (typeof renderCorporateView === 'function') {
                    renderCorporateView(currentCorporateFilter || 'ALL');
                }
            } else if (domain === 'ctd') {
                document.getElementById("bc-root").innerText = "DOSSIERS_ANPP";
                document.getElementById("bc-sec").innerText = "Guide_CTD_2025";
                document.getElementById("bc-prod").innerText = "Catalogue des Dossiers";
                document.getElementById("bc-mod").innerText = "Checklists Réglementaires";
                resetStandardTableContainer();
                loadDossiers();
            } else if (domain === 'deficiencies') {
                document.getElementById("bc-root").innerText = "COMPLÉMENTS_ANPP";
                document.getElementById("bc-sec").innerText = "Veille_Réglementaire";
                document.getElementById("bc-prod").innerText = "Décret_20-325";
                document.getElementById("bc-mod").innerText = "Traçabilité_Délais_Légaux";
                const preview = document.getElementById("preview-body");
                if (preview) {
                    preview.innerHTML = `
                        <div style="padding:24px; text-align:center; color:var(--text-muted); font-size:12px;">
                            <div style="font-size:28px; margin-bottom:8px;">⏱️</div>
                            <strong style="color:var(--text-main); font-size:13px;">Studio de Compléments ANPP</strong>
                            <p style="margin-top:6px; font-size:11.5px; line-height:1.5;">Cliquez sur <strong>🔍 Ouvrir Studio</strong> sur une notification officielle pour décomposer les questions réglementaires et apposer le visa Pharmacien DT (1.A.7).</p>
                        </div>
                    `;
                }
                if (typeof loadDeficienciesDashboard === 'function') {
                    loadDeficienciesDashboard(typeof currentDeficiencyFilter !== 'undefined' ? currentDeficiencyFilter : 'all');
                }
            } else if (domain === 'variations') {
                document.getElementById("bc-root").innerText = "POST-AMM";
                document.getElementById("bc-sec").innerText = "Variations";
                document.getElementById("bc-prod").innerText = "LD-DEPP07/V01";
                document.getElementById("bc-mod").innerText = "12 Items Réglementaires";
                resetStandardTableContainer();
                const preview = document.getElementById("preview-body");
                if (preview && (typeof selectedVariationID === 'undefined' || !selectedVariationID)) {
                    preview.innerHTML = `<div style="text-align:center; padding:40px; color:var(--text-muted);">Sélectionnez une variation pour visualiser l'Impact Tree et le dossier réglementaire.</div>`;
                }
                if (typeof renderVariationsTable === 'function') {
                    renderVariationsTable(typeof currentVariationsFilter !== 'undefined' ? currentVariationsFilter : 'all');
                }
            } else if (domain === 'dm') {
                document.getElementById("bc-root").innerText = "DISPOSITIFS_MÉDICAUX";
                document.getElementById("bc-sec").innerText = "Décret_20-324";
                document.getElementById("bc-prod").innerText = "Homologation & Régularisation";
                document.getElementById("bc-mod").innerText = "Arrêté du 10 Mai 2021";
                if (typeof initDMDashboard === 'function') {
                    initDMDashboard();
                } else {
                    loadRefSection('devices');
                }
            } else if (domain === 'ged') {
                document.getElementById("bc-root").innerText = "COFFRE-FORT";
                document.getElementById("bc-sec").innerText = "GED_BPF";
                document.getElementById("bc-prod").innerText = "Taxonomie";
                document.getElementById("bc-mod").innerText = "Répertoire BPF";
                resetStandardTableContainer();
                currentGEDTaxonomyPath = '';
                loadGEDTaxonomy();
                loadGEDDocuments();
            } else if (domain === 'admin') {
                document.getElementById("bc-root").innerText = "ADMINISTRATION";
                document.getElementById("bc-sec").innerText = "Habilitations_RBAC";
                document.getElementById("bc-prod").innerText = "Collaborateurs & Matrice";
                document.getElementById("bc-mod").innerText = "Conformité PSSI & BPF";
                resetStandardTableContainer();
                if (typeof loadAdminSection === 'function') {
                    loadAdminSection(currentAdminSection || 'users');
                }
            } else if (domain === 'sys') {
                document.getElementById("bc-root").innerText = "INFRASTRUCTURE_IT";
                document.getElementById("bc-sec").innerText = "On-Premise_BPF";
                document.getElementById("bc-prod").innerText = "Registre & Coffre-fort GED";
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
                        Base Référentiel : <strong style="color:var(--accent-emerald);">Opérationnelle &amp; Conforme</strong>
                    </div>
                `;
            } else if (domain === 'corporate') {
                tb.innerHTML = `
                    <button class="btn-pill" style="background:var(--accent-blue); color:white; border:none; font-weight:700;" onclick="openNewCorporateAppModal()" title="Créer une nouvelle demande d'agrément (Décret 21-82)">➕ Nouvelle Demande</button>
                    <button class="btn-table-icon" onclick="selectCorporateFilter('ALL')" title="Toutes les Demandes MIPH">📁</button>
                    <button class="btn-table-icon" onclick="selectCorporateFilter('LICENSES')" title="Registre des Agréments Ministériels Actifs">📜</button>
                    <button class="btn-table-icon" onclick="renderCorporateView()" title="Actualiser la liste">🔄</button>
                    <button class="btn-table-icon" onclick="openAuditModal()" title="Journal d'Audit Trail Réglementaire">🛡️</button>
                    <button class="btn-table-icon" onclick="togglePreviewPane()" title="Afficher / Masquer le volet latéral (Alt+P)">👁️</button>
                    <div style="margin-left:auto; font-size:11px; color:var(--text-muted);" id="corporate-progress">
                        Passerelle Tabadol : <strong style="color:var(--accent-emerald);">Connectée &amp; Sécurisée</strong>
                    </div>
                `;
            } else if (domain === 'ctd') {
                tb.innerHTML = `
                    <button class="btn-pill" id="tb-btn-new-dossier" data-perm="DOSSIER_CREATE" style="background:var(--accent-blue); color:white; border:none;" onclick="openNewDossierModal()">➕ Nouveau Dossier</button>
                    <button class="btn-pill" id="tb-btn-members" onclick="openDossierMembersModal()">👥 Équipe &amp; Secret</button>
                    <button class="btn-pill" id="tb-btn-sign-pdt" onclick="openSignPDTModal()">✍️ Visa PDT (21 CFR Part 11)</button>
                    <button class="btn-pill" id="tb-btn-deposit-anpp" onclick="openRecordDepositModal()">🏛️ Enregistrer Dépôt ANPP</button>
                    <button class="btn-pill" id="tb-btn-eservices" style="background:linear-gradient(135deg, #2563eb, #10b981); color:white; border:none; font-weight:700;" onclick="openPreSubmissionAuditModal(selectedDossierId)">🚀 E-Services ANPP</button>
                    <button class="btn-pill" id="tb-btn-fees" onclick="openGenerateFeeSlipModal(selectedDossierId)">💰 Bordereau BV1/BV2</button>
                    <button class="btn-pill" id="tb-btn-schedules" onclick="openFeeSchedulesModal()">⚖️ Barème ANPP</button>
                    <button class="btn-pill" id="tb-btn-cost-analytics" style="background:rgba(37,99,235,0.1); color:var(--accent-blue); border:1px solid rgba(37,99,235,0.3); font-weight:700;" onclick="openCostAnalyticsDashboardModal()">📊 Coûts Consolidés</button>
                    <button class="btn-pill" id="tb-btn-deficiency" onclick="openNewDeficiencyLetterModal(selectedDossierId)">⏱️ Enregistrer Complément ANPP</button>
                    <button class="btn-pill" id="tb-btn-appointments" onclick="openAppointmentsModal()">📅 Séances ANPP</button>
                    <button class="btn-pill" id="tb-btn-reg-watch" onclick="openRegulatoryWatchModal()">📜 Veille Juridique</button>
                    <button class="btn-pill" id="tb-btn-export" data-perm="EXPORT_PACKAGE" onclick="simulateExportZip()">📦 Exporter CTD (ZIP)</button>
                    <button class="btn-pill" onclick="loadDossiers()">🔄 Actualiser</button>
                    <button class="btn-pill" data-perm="AUDIT_VIEW" onclick="openAuditModal()">🛡️ Audit Trail</button>
                    <button class="btn-pill" onclick="togglePreviewPane()">👁️ Volet Aperçu (Alt+P)</button>
                    <div style="margin-left:auto; font-size:11px; color:var(--text-muted);" id="dossier-progress">
                        Complétude : <strong id="kpi-dossier-completeness" style="color:var(--accent-blue);">0%</strong>
                    </div>
                `;
            } else if (domain === 'deficiencies') {
                tb.innerHTML = `
                    <button class="btn-pill" style="background:var(--accent-blue); color:white; border:none; font-weight:700;" onclick="openNewDeficiencyLetterModal()">➕ Enregistrer une Notification de Compléments ANPP</button>
                    <button class="btn-pill" onclick="loadDeficienciesDashboard(typeof currentDeficiencyFilter !== 'undefined' ? currentDeficiencyFilter : 'all')">🔄 Actualiser</button>
                    <button class="btn-pill" data-perm="AUDIT_VIEW" onclick="openAuditModal()">🛡️ Audit Trail</button>
                    <button class="btn-pill" onclick="togglePreviewPane()">👁️ Volet Aperçu (Alt+P)</button>
                    <div style="margin-left:auto; font-size:11px; color:var(--text-muted);" id="dossier-progress">
                        Délais Légaux : <strong style="color:var(--accent-blue);">Compte à Rebours J-X (Guide 2025)</strong>
                    </div>
                `;
            } else if (domain === 'variations') {
                tb.innerHTML = `
                    <button class="btn-pill" id="tb-btn-new-variation" data-perm="DOSSIER_CREATE" style="background:var(--accent-blue); color:white; border:none; font-weight:700;" onclick="openNewVariationModal()">➕ Déposer une Variation (LD-DEPP07)</button>
                    <button class="btn-pill" onclick="renderVariationsTable(typeof currentVariationsFilter !== 'undefined' ? currentVariationsFilter : 'all')">🔄 Actualiser</button>
                    <button class="btn-pill" data-perm="AUDIT_VIEW" onclick="openAuditModal()">🛡️ Audit Trail</button>
                    <button class="btn-pill" onclick="togglePreviewPane()">👁️ Volet Aperçu (Alt+P)</button>
                    <div style="margin-left:auto; font-size:11px; color:var(--text-muted);" id="dossier-progress">
                        Ligne Directrice : <strong style="color:var(--accent-blue);">ANPP LD-DEPP07/V01 (Avril 2026)</strong>
                    </div>
                `;
            } else if (domain === 'dm') {
                if (typeof updateDMToolbar === 'function') {
                    updateDMToolbar();
                } else {
                    tb.innerHTML = `
                        <button class="btn-pill" data-perm="REF_EDIT" style="background:var(--accent-blue); color:white; border:none;" onclick="document.getElementById('modal-new-dm').classList.add('open')">➕ Nouveau DM</button>
                        <button class="btn-pill" onclick="loadRefSection('devices')">🔄 Actualiser</button>
                        <button class="btn-pill" data-perm="AUDIT_VIEW" onclick="openAuditModal()">🛡️ Audit Trail</button>
                        <button class="btn-pill" onclick="togglePreviewPane()">👁️ Volet Aperçu (Alt+P)</button>
                        <div style="margin-left:auto; font-size:11px; color:var(--text-muted);" id="dossier-progress">
                            Cadre : <strong>Arrêté du 10 Mai 2021</strong>
                        </div>
                    `;
                }
            } else if (domain === 'ged') {
                tb.innerHTML = `
                    <button class="btn-pill" data-perm="DOC_DRAFT_MAKER" style="background:var(--accent-blue); color:white; border:none;" onclick="openNewGEDDocModal()">➕ Déposer Document</button>
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
                        Infrastructure : <strong style="color:var(--accent-emerald);">Coffre-fort Qualifié &amp; Empreinte SHA-256</strong>
                    </div>
                `;
            }
        }

        function renderTreeForDomain(domain) {
            const container = document.getElementById("tree-root");
            if (domain === 'corporate') {
                if (typeof renderCorporateTree === 'function') {
                    renderCorporateTree();
                }
                return;
            }
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
                const curF = (typeof currentDeficiencyFilter !== 'undefined') ? currentDeficiencyFilter : 'all';
                container.innerHTML = `
                    <div class="tree-node ${(!curF || curF === 'all') ? 'selected' : ''}" onclick="filterDeficienciesTree('all', this)">
                        <span class="chevron">⏱️</span>
                        <span>Toutes les Notifications ANPP</span>
                        <span class="count-badge req">Guide 2025</span>
                    </div>
                    <div class="tree-node ${curF === 'critical' ? 'selected' : ''}" style="padding-left:24px;" onclick="filterDeficienciesTree('critical', this)">
                        <span>🚨 Alertes Critiques (J ≤ 15)</span>
                    </div>
                    <div class="tree-node ${curF === 'urgent' ? 'selected' : ''}" style="padding-left:24px;" onclick="filterDeficienciesTree('urgent', this)">
                        <span>⚠️ Urgences Réglementaires (J ≤ 30)</span>
                    </div>
                    <div class="tree-node ${curF === 'validated' ? 'selected' : ''}" style="padding-left:24px;" onclick="filterDeficienciesTree('validated', this)">
                        <span>✍️ Visées Pharmacien DT (1.A.7)</span>
                    </div>
                `;
            } else if (domain === 'variations') {
                const curV = (typeof currentVariationsFilter !== 'undefined') ? currentVariationsFilter : 'all';
                container.innerHTML = `
                    <div class="tree-node ${(!curV || curV === 'all') ? 'selected' : ''}" onclick="filterVariationsTree('all', this)">
                        <span class="chevron">▼</span>
                        <span>🔄 Variations en Cours (ANPP)</span>
                        <span class="count-badge warn">3</span>
                    </div>
                    <div class="tree-sub open" id="sub-var-active">
                        <div class="tree-node ${curV === 'Mmin' ? 'selected' : ''}" style="padding-left:24px;" onclick="filterVariationsTree('Mmin', this)"><span>🟡 Mmin (Mineures - 7.b.1)</span><span class="count-badge ok">1</span></div>
                        <div class="tree-node ${curV === 'Mmod' ? 'selected' : ''}" style="padding-left:24px;" onclick="filterVariationsTree('Mmod', this)"><span>🟠 Mmod (Modérées - 12.a)</span><span class="count-badge ok">1</span></div>
                        <div class="tree-node ${curV === 'Mmaj' ? 'selected' : ''}" style="padding-left:24px;" onclick="filterVariationsTree('Mmaj', this)"><span>🔴 Mmaj (Majeures - 7.b.2)</span><span class="count-badge req">1</span></div>
                    </div>

                    <div class="tree-node" onclick="toggleSub('sub-var-history')">
                        <span class="chevron">▶</span>
                        <span>📅 Renouvellements Quinquennaux</span>
                        <span class="count-badge ok">2026-2027</span>
                    </div>
                `;
            } else if (domain === 'dm') {
                const isReg = typeof currentDMTab !== 'undefined' && currentDMTab === 'regularization';
                container.innerHTML = `
                    <div class="tree-node ${!isReg ? 'selected' : ''}" onclick="switchDMTab('homologation')">
                        <span class="chevron">🩺</span>
                        <span>Portefeuille DM (6 Sections)</span>
                        <span class="count-badge ok">Arrêté 10 Mai</span>
                    </div>
                    <div class="tree-node" style="padding-left:24px;" onclick="filterDMClass('all')"><span>📁 Tous les DM</span></div>
                    <div class="tree-node" style="padding-left:24px;" onclick="filterDMClass('I')"><span>📁 Classe I (Non invasifs)</span></div>
                    <div class="tree-node" style="padding-left:24px;" onclick="filterDMClass('IIa')"><span>📁 Classe IIa (Seringues, Cathéters)</span></div>
                    <div class="tree-node" style="padding-left:24px;" onclick="filterDMClass('IIb')"><span>📁 Classe IIb & III (Implants)</span></div>

                    <div class="tree-node ${isReg ? 'selected' : ''}" style="margin-top:8px;" onclick="switchDMTab('regularization')">
                        <span class="chevron">⏱️</span>
                        <span>Mesures Transitoires (Dérogatoires)</span>
                        <span class="count-badge req">J-22</span>
                    </div>
                    <div class="tree-node" style="padding-left:24px;" onclick="filterDMRegStatus('all')"><span>📑 Tous les DM Transitoires</span></div>
                    <div class="tree-node" style="padding-left:24px;" onclick="filterDMRegStatus('EN_COURS')"><span>⏳ En Cours de Régularisation</span></div>
                    <div class="tree-node" style="padding-left:24px;" onclick="filterDMRegStatus('REGULARISE')"><span>✅ Régularisés & Homologués</span></div>
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
                            <span class="count-badge req">Garde-fous</span>
                        </div>
                        <div class="tree-node ${currentAdminSection === 'settings' ? 'selected' : ''}" onclick="loadAdminSection('settings')">
                            <span class="chevron">⚙️</span>
                            <span>Paramètres & Redevances</span>
                            <span class="count-badge ok">ALCOA+</span>
                        </div>
                    `;
                }
            }
        }

        function filterDeficienciesTree(filter, el) {
            if (el) {
                document.querySelectorAll('#tree-root .tree-node').forEach(n => n.classList.remove('selected'));
                el.classList.add('selected');
            }
            if (typeof loadDeficienciesDashboard === 'function') {
                loadDeficienciesDashboard(filter);
            }
        }
        window.filterDeficienciesTree = filterDeficienciesTree;

        function filterVariationsTree(filter, el) {
            if (el) {
                document.querySelectorAll('#tree-root .tree-node').forEach(n => n.classList.remove('selected'));
                el.classList.add('selected');
            }
            if (typeof renderVariationsTable === 'function') {
                renderVariationsTable(filter);
            }
        }
        window.filterVariationsTree = filterVariationsTree;

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

            const gateTitle = document.getElementById("gate-title");
            const gateSub = document.getElementById("gate-subtitle");
            const gateLang = document.getElementById("gate-lang-label");

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

                if (gateTitle) gateTitle.innerText = "برمجية سند (SANAD) — المصادقة والتسجيل الصيدلاني المسبق";
                if (gateSub) gateSub.innerText = "فضاء تنظيمي خاص وسري لمخابر الأدوية والمستلزمات الطبية في الجزائر (ANPP / MIPH)";
                if (gateLang) gateLang.innerText = "🇫🇷";

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

                if (gateTitle) gateTitle.innerText = "Progiciel SANAD — Homologation ANPP & GED";
                if (gateSub) gateSub.innerText = "Cockpit réglementaire pour l'enregistrement des Médicaments (CTD), Dispositifs Médicaux (DDM) et Compléments Alimentaires en Algérie (ANPP / MIPH).";
                if (gateLang) gateLang.innerText = "🇩🇿";

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

        function openModal(id) {
            const el = document.getElementById(id);
            if (el) {
                el.classList.add("open");
                el.style.display = "flex";
            }
        }

        function toggleModalFullscreen(modalOverlayId) {
            const overlay = document.getElementById(modalOverlayId);
            if (!overlay) return;
            const modal = overlay.querySelector('.modal');
            if (!modal) return;

            const isFullscreen = modal.classList.toggle('modal-fullscreen');
            const btn = overlay.querySelector('.btn-modal-expand');
            if (btn) {
                const icon = btn.querySelector('.expand-icon');
                const lbl = btn.querySelector('.expand-label');
                if (isFullscreen) {
                    if (icon) icon.innerText = '🗗';
                    if (lbl) lbl.innerText = 'Réduire';
                    btn.title = 'Restaurer la taille normale';
                } else {
                    if (icon) icon.innerText = '⛶';
                    if (lbl) lbl.innerText = 'Agrandir';
                    btn.title = 'Agrandir en plein écran';
                }
            }
        }

        function closeAllModals() {
            document.querySelectorAll('.modal-overlay').forEach(m => {
                m.classList.remove('open');
                m.style.display = "";
                const modal = m.querySelector('.modal');
                if (modal) modal.classList.remove('modal-fullscreen');
            });
            const previewIframe = document.getElementById("doc-preview-iframe");
            if (previewIframe && previewIframe.src !== "about:blank") {
                previewIframe.src = "about:blank";
            }
        }

        function closeModal(id) {
            const el = document.getElementById(id);
            if (el) {
                el.classList.remove("open");
                el.style.display = "";
                const modal = el.querySelector('.modal');
                if (modal) modal.classList.remove('modal-fullscreen');
                const btn = el.querySelector('.btn-modal-expand');
                if (btn) {
                    const icon = btn.querySelector('.expand-icon');
                    const lbl = btn.querySelector('.expand-label');
                    if (icon) icon.innerText = '⛶';
                    if (lbl) lbl.innerText = 'Agrandir';
                }
            }
            if (id === 'modal-doc-preview') {
                const iframe = document.getElementById("doc-preview-iframe");
                if (iframe) iframe.src = "about:blank";
            }
            if (id === 'modal-dossier-reports') {
                const container = document.getElementById("report-iframe-container");
                if (container) container.innerHTML = "";
            }
        }

        // Initialiseur générique de modales redimensionnables et déplaçables (Façon Fenêtre OS)
        function initModalWindowResizer(dialogId, headerId, options) {
            const dialog = document.getElementById(dialogId);
            if (!dialog) return;
            if (dialog._resizerInitialized) return;
            dialog._resizerInitialized = true;

            const header = headerId ? document.getElementById(headerId) : dialog.querySelector('.modal-header');
            const minWidth = (options && options.minWidth) || 480;
            const minHeight = (options && options.minHeight) || 360;

            const edges = dialog.querySelectorAll('.modal-resize-edge');
            edges.forEach(edge => {
                edge.addEventListener('pointerdown', (e) => {
                    if (dialog.classList.contains('modal-fullscreen') || e.button !== 0) return;
                    e.preventDefault();
                    e.stopPropagation();

                    const dir = edge.dataset.direction;
                    const startRect = dialog.getBoundingClientRect();
                    const startX = e.clientX;
                    const startY = e.clientY;

                    dialog.style.position = 'absolute';
                    dialog.style.left = startRect.left + 'px';
                    dialog.style.top = startRect.top + 'px';
                    dialog.style.width = startRect.width + 'px';
                    dialog.style.height = startRect.height + 'px';
                    dialog.style.margin = '0';
                    dialog.style.maxWidth = 'none';
                    dialog.style.maxHeight = 'none';

                    edge.setPointerCapture(e.pointerId);

                    function onPointerMove(ev) {
                        const dx = ev.clientX - startX;
                        const dy = ev.clientY - startY;

                        let newW = startRect.width;
                        let newH = startRect.height;
                        let newL = startRect.left;
                        let newT = startRect.top;

                        if (dir.includes('e')) {
                            newW = Math.max(minWidth, startRect.width + dx);
                            if (newL + newW > window.innerWidth - 6) newW = window.innerWidth - 6 - newL;
                        }
                        if (dir.includes('w')) {
                            const maxL = startRect.right - minWidth;
                            let candL = startRect.left + dx;
                            if (candL > maxL) candL = maxL;
                            if (candL < 6) candL = 6;
                            newL = candL;
                            newW = startRect.right - newL;
                        }
                        if (dir.includes('s')) {
                            newH = Math.max(minHeight, startRect.height + dy);
                            if (newT + newH > window.innerHeight - 6) newH = window.innerHeight - 6 - newT;
                        }
                        if (dir.includes('n')) {
                            const maxT = startRect.bottom - minHeight;
                            let candT = startRect.top + dy;
                            if (candT > maxT) candT = maxT;
                            if (candT < 6) candT = 6;
                            newT = candT;
                            newH = startRect.bottom - newT;
                        }

                        dialog.style.left = newL + 'px';
                        dialog.style.top = newT + 'px';
                        dialog.style.width = newW + 'px';
                        dialog.style.height = newH + 'px';
                    }

                    function onPointerUp(ev) {
                        edge.removeEventListener('pointermove', onPointerMove);
                        edge.removeEventListener('pointerup', onPointerUp);
                        edge.removeEventListener('pointercancel', onPointerUp);
                        try { edge.releasePointerCapture(ev.pointerId); } catch (_) {}
                    }

                    edge.addEventListener('pointermove', onPointerMove);
                    edge.addEventListener('pointerup', onPointerUp);
                    edge.addEventListener('pointercancel', onPointerUp);
                });
            });

            if (header) {
                header.addEventListener('pointerdown', (e) => {
                    if (e.target.closest('button, a, input, select, textarea') || dialog.classList.contains('modal-fullscreen') || e.button !== 0) {
                        return;
                    }
                    e.preventDefault();
                    const startRect = dialog.getBoundingClientRect();
                    const startX = e.clientX;
                    const startY = e.clientY;

                    dialog.style.position = 'absolute';
                    dialog.style.left = startRect.left + 'px';
                    dialog.style.top = startRect.top + 'px';
                    dialog.style.width = startRect.width + 'px';
                    dialog.style.height = startRect.height + 'px';
                    dialog.style.margin = '0';
                    dialog.style.maxWidth = 'none';
                    dialog.style.maxHeight = 'none';

                    header.style.cursor = 'grabbing';
                    header.setPointerCapture(e.pointerId);

                    function onHeaderMove(ev) {
                        const dx = ev.clientX - startX;
                        const dy = ev.clientY - startY;
                        let candL = startRect.left + dx;
                        let candT = startRect.top + dy;
                        if (candL < 6) candL = 6;
                        if (candT < 6) candT = 6;
                        if (candL + startRect.width > window.innerWidth - 6) candL = window.innerWidth - 6 - startRect.width;
                        if (candT + 40 > window.innerHeight) candT = window.innerHeight - 40;
                        dialog.style.left = candL + 'px';
                        dialog.style.top = candT + 'px';
                    }

                    function onHeaderUp(ev) {
                        header.removeEventListener('pointermove', onHeaderMove);
                        header.removeEventListener('pointerup', onHeaderUp);
                        header.removeEventListener('pointercancel', onHeaderUp);
                        header.style.cursor = 'grab';
                        try { header.releasePointerCapture(ev.pointerId); } catch (_) {}
                    }

                    header.addEventListener('pointermove', onHeaderMove);
                    header.addEventListener('pointerup', onHeaderUp);
                    header.addEventListener('pointercancel', onHeaderUp);
                });
            }
        }
        window.initModalWindowResizer = initModalWindowResizer;

        // Visualiseur universel de documents & scans in-app (Agrandissable & Redimensionnable)
        function openDocumentViewerModal(docId, docTitle, versionId) {
            if (!docId) {
                if (typeof showNotifyModal === 'function') {
                    showNotifyModal("Document introuvable", "Aucun identifiant de pièce documentaire n'a été fourni.", "⚠️");
                } else {
                    alert("Aucun document à afficher");
                }
                return;
            }

            let url = `/api/v1/ged/documents/${docId}/download`;
            if (versionId) url += `?version_id=${encodeURIComponent(versionId)}`;

            const titleEl = document.getElementById("doc-preview-title");
            if (titleEl) {
                titleEl.innerText = docTitle || "Aperçu de la Pièce Numérisée";
                titleEl.title = docTitle || "Aperçu de la Pièce Numérisée";
            }

            const downloadBtn = document.getElementById("doc-preview-download-btn");
            if (downloadBtn) {
                downloadBtn.href = url;
            }

            const iframe = document.getElementById("doc-preview-iframe");
            if (iframe) {
                iframe.src = url;
            }

            openModal("modal-doc-preview");
            setTimeout(() => {
                initModalWindowResizer('modal-doc-preview-dialog', 'modal-doc-preview-header', { minWidth: 480, minHeight: 360 });
            }, 60);
        }

        function closeDocPreviewModal() {
            closeModal("modal-doc-preview");
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
            // Contrôle d'expiration de session réglementaire (BPF 15 min / 21 CFR Part 11)
            if (typeof isSessionTimedOut === 'function' && isSessionTimedOut()) {
                if (typeof showSessionTimeoutModal === 'function') {
                    showSessionTimeoutModal('navigation');
                }
                return;
            }
            if (targetKey && targetKey.startsWith("ctd-para-")) {
                const modNum = parseInt(targetKey.replace("ctd-para-", ""), 10) || 1;
                if (typeof openDossierModule === 'function') {
                    openDossierModule('parac', modNum);
                    return;
                }
            } else if (targetKey && targetKey.startsWith("dm-")) {
                const hint = targetKey.replace("dm-", "");
                if (typeof openDossierModule === 'function') {
                    openDossierModule(hint, 0);
                    return;
                }
            } else if (targetKey && targetKey.startsWith("ged-")) {
                if (typeof switchDomainModule === 'function') {
                    switchDomainModule('ged');
                    return;
                }
            }
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

        // ==============================================================================
        // GESTION DYNAMIQUE DU FUSEAU HORAIRE & HORLOGE RÉGLEMENTAIRE
        // ==============================================================================
        let appTimezone = localStorage.getItem('homolog_timezone') || 'Africa/Algiers';

        function getTimezoneDisplayName(tz) {
            if (tz === 'Africa/Algiers') return 'UTC+1 (Alger)';
            if (tz === 'UTC') return 'UTC';
            if (tz === 'LOCAL') {
                const offsetMin = -new Date().getTimezoneOffset();
                const sign = offsetMin >= 0 ? '+' : '-';
                const h = Math.floor(Math.abs(offsetMin) / 60);
                return `Local (UTC${sign}${h})`;
            }
            if (tz === 'Europe/Paris') return 'Paris';
            if (tz === 'Asia/Riyadh') return 'Riyad (UTC+3)';
            return tz.split('/')[1] || tz;
        }

        function updateClock() {
            const clockEl = document.getElementById("clock-utc");
            const badgeEl = document.getElementById("clock-tz-badge");
            const labelEl = document.getElementById("clock-label");
            if (!clockEl) return;

            const now = new Date();
            let timeStr = "";

            try {
                if (appTimezone === 'UTC') {
                    timeStr = now.toUTCString().split(" ")[4];
                    if (labelEl) labelEl.innerText = "Heure UTC :";
                } else if (appTimezone === 'LOCAL') {
                    timeStr = now.toLocaleTimeString('fr-FR', { hour12: false });
                    if (labelEl) labelEl.innerText = "Heure Locale :";
                } else {
                    timeStr = now.toLocaleTimeString('fr-FR', { timeZone: appTimezone, hour12: false });
                    if (labelEl) labelEl.innerText = "Heure :";
                }
            } catch(e) {
                timeStr = now.toLocaleTimeString('fr-FR', { hour12: false });
            }

            clockEl.innerText = timeStr;
            if (badgeEl) badgeEl.innerText = getTimezoneDisplayName(appTimezone);

            // Mettre à jour l'aperçu dans le modal s'il est ouvert
            const previewEl = document.getElementById("tz-current-time-preview");
            if (previewEl) {
                previewEl.innerText = `Heure affichée : ${timeStr} • ${getTimezoneDisplayName(appTimezone)}`;
            }
        }

        function openTimezoneModal() {
            const radios = document.querySelectorAll('input[name="app-tz-radio"]');
            radios.forEach(r => {
                r.checked = (r.value === appTimezone);
                const card = r.closest('.tz-option-card');
                if (card) {
                    if (r.checked) {
                        card.style.borderColor = 'var(--accent-blue)';
                        card.style.background = 'rgba(14,165,233,0.06)';
                    } else {
                        card.style.borderColor = 'var(--border-color)';
                        card.style.background = 'transparent';
                    }
                }
            });

            const localDesc = document.getElementById("tz-local-detected-desc");
            if (localDesc) {
                try {
                    const detected = Intl.DateTimeFormat().resolvedOptions().timeZone;
                    const offsetMin = -new Date().getTimezoneOffset();
                    const sign = offsetMin >= 0 ? '+' : '-';
                    const h = Math.floor(Math.abs(offsetMin) / 60);
                    localDesc.innerText = `Détecté : ${detected} (UTC${sign}${h})`;
                } catch(e) {}
            }

            if (typeof openModal === 'function') {
                openModal('modal-timezone-picker');
            }
        }

        function setTimezonePreference(tz) {
            appTimezone = tz;
            localStorage.setItem('homolog_timezone', tz);
            updateClock();

            const radios = document.querySelectorAll('input[name="app-tz-radio"]');
            radios.forEach(r => {
                const card = r.closest('.tz-option-card');
                if (card) {
                    if (r.value === tz) {
                        card.style.borderColor = 'var(--accent-blue)';
                        card.style.background = 'rgba(14,165,233,0.06)';
                    } else {
                        card.style.borderColor = 'var(--border-color)';
                        card.style.background = 'transparent';
                    }
                }
            });

            if (typeof showToast === 'function') {
                showToast(`Fuseau horaire réglé sur ${getTimezoneDisplayName(tz)}`, 'success', 2500);
            }
        }

        window.openTimezoneModal = openTimezoneModal;
        window.setTimezonePreference = setTimezonePreference;

        // Lancement immédiat puis rafraîchissement chaque seconde
        updateClock();
        setInterval(updateClock, 1000);

        // Initialisation des splitters glissants
        initSplitters();

        // Système global de notifications Toast (ALCOA+ / Feedback Utilisateur)
        function showToast(message, type = 'info', duration = 3800) {
            let container = document.getElementById('toast-container');
            if (!container) {
                container = document.createElement('div');
                container.id = 'toast-container';
                container.style.cssText = 'position:fixed; bottom:24px; right:24px; z-index:999999; display:flex; flex-direction:column; gap:10px; max-width:440px; pointer-events:none;';
                document.body.appendChild(container);
            }

            const toast = document.createElement('div');
            toast.className = `toast-item toast-${type}`;
            
            let bg = '#1e293b';
            let border = '#334155';
            let icon = 'ℹ️';
            if (type === 'success') { bg = '#064e3b'; border = '#059669'; icon = '✅'; }
            else if (type === 'error' || type === 'danger') { bg = '#7f1d1d'; border = '#dc2626'; icon = '⚠️'; }
            else if (type === 'warning') { bg = '#78350f'; border = '#d97706'; icon = '⚠️'; }

            toast.style.cssText = `
                background: ${bg};
                color: #ffffff;
                border: 1px solid ${border};
                box-shadow: 0 10px 25px -5px rgba(0, 0, 0, 0.4), 0 8px 10px -6px rgba(0, 0, 0, 0.3);
                border-radius: 8px;
                padding: 12px 16px;
                font-size: 13px;
                font-weight: 600;
                line-height: 1.4;
                display: flex;
                align-items: center;
                gap: 10px;
                pointer-events: auto;
                opacity: 0;
                transform: translateY(12px);
                transition: opacity 0.25s ease, transform 0.25s ease;
                cursor: pointer;
            `;

            toast.innerHTML = `<span style="font-size:16px;">${icon}</span><span style="flex:1;">${escapeHTML(message)}</span>`;
            container.appendChild(toast);

            requestAnimationFrame(() => {
                toast.style.opacity = '1';
                toast.style.transform = 'translateY(0)';
            });

            const removeToast = () => {
                toast.style.opacity = '0';
                toast.style.transform = 'translateY(12px)';
                setTimeout(() => { if (toast.parentNode) toast.parentNode.removeChild(toast); }, 250);
            };

            toast.addEventListener('click', removeToast);
            setTimeout(removeToast, duration);
        }
        window.showToast = showToast;
