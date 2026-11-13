// ============================================================================
// RÉFÉRENTIEL PHARMACEUTIQUE (Jalon 1.3 - Établissements, Sites, Produits ANPP)
// ============================================================================

let currentRefSection = 'establishments';
        let currentRefList = [];
        let selectedRefIndex = 0;

        function resetStandardTableContainer() {
            const container = document.querySelector('.table-container');
            if (!container) return;
            const table = document.getElementById("doc-table");
            const tbody = document.getElementById("doc-tbody");
            const thead = document.getElementById("table-head");
            // Si la table de base est absente ou polluée par une vue complexe (ex: CTD KPIs)
            if (!table || !tbody || !thead || container.querySelector('.kpi-dossier-bar') || container.querySelector('.tabs-bar')) {
                container.innerHTML = `
                    <table id="doc-table">
                        <thead id="table-head"></thead>
                        <tbody id="doc-tbody">
                            <tr><td colspan="7" style="text-align:center; padding:20px; color:var(--text-muted);">Chargement depuis PostgreSQL...</td></tr>
                        </tbody>
                    </table>
                `;
            }
        }

        async function loadRefSection(section) {
            currentRefSection = section || currentRefSection || 'establishments';
            resetStandardTableContainer();

            if (typeof renderTreeForDomain === 'function' && typeof currentDomain !== 'undefined') {
                renderTreeForDomain(currentDomain);
            }

            if (window.innerWidth < 768 && typeof toggleMobileTree === 'function') {
                toggleMobileTree(false);
            }
            
            const bcRoot = document.getElementById("bc-root");
            const bcSec = document.getElementById("bc-sec");
            const bcProd = document.getElementById("bc-prod");
            const bcMod = document.getElementById("bc-mod");
            if (bcRoot) bcRoot.innerText = "RÉFÉRENTIEL";
            if (bcSec) bcSec.innerText = "Pharmacopée";
            if (bcProd) bcProd.innerText = currentRefSection.toUpperCase();
            if (bcMod) bcMod.innerText = "Catalogue National";

            const tbody = document.getElementById("doc-tbody");
            if (tbody) {
                tbody.innerHTML = `<tr><td colspan="7" style="text-align:center; padding:20px; color:var(--text-muted);">Chargement depuis PostgreSQL...</td></tr>`;
            }

            let url = '/api/v1/establishments';
            if (currentRefSection === 'sites') url = '/api/v1/manufacturing-sites';
            else if (currentRefSection === 'medicines') url = '/api/v1/medicines';
            else if (currentRefSection === 'devices') url = '/api/v1/medical-devices';

            try {
                const res = await fetch(url);
                if (res.status === 401) {
                    if (typeof handleAuthExpired === 'function') handleAuthExpired();
                    return;
                }
                if (!res.ok) throw new Error("HTTP " + res.status);
                const data = await res.json();
                currentRefList = (data.items || data.establishments || data.sites || data.medicines || data.devices || []);
                selectedRefIndex = 0;
                renderRefTable(currentRefSection, currentRefList);
            } catch (e) {
                console.error("[Referential] Erreur chargement :", e);
                const tb = document.getElementById("doc-tbody");
                if (tb) {
                    tb.innerHTML = `<tr><td colspan="7" style="text-align:center; padding:20px; color:var(--accent-red);">Impossible de charger : ${e.message}</td></tr>`;
                }
            }
        }

        function renderRefTable(section, items) {
            const thead = document.getElementById("table-head");
            const tbody = document.getElementById("doc-tbody");
            tbody.innerHTML = "";

            if (section === 'establishments') {
                thead.innerHTML = `
                    <tr>
                        <th style="width: 90px;">Statut MIPH</th>
                        <th>Raison Sociale</th>
                        <th class="col-hide-md">Statut Juridique</th>
                        <th class="col-hide-sm">N° Agrément MIPH</th>
                        <th>Pharmacien DT (1.A.7)</th>
                        <th class="col-hide-lg">Décision Nominative</th>
                        <th class="col-hide-sm">Wilaya</th>
                    </tr>
                `;
                if (items.length === 0) {
                    tbody.innerHTML = `<tr><td colspan="7" style="text-align:center; padding:15px; color:var(--text-dim);">Aucun établissement enregistré. Cliquez sur '➕ Nouveau' pour en créer un.</td></tr>`;
                    return;
                }
                items.forEach((item, idx) => {
                    const tr = document.createElement("tr");
                    if (idx === selectedRefIndex) tr.classList.add("selected");
                    tr.onclick = () => {
                        selectedRefIndex = idx;
                        renderRefTable(section, items);
                        if (window.innerWidth < 1400) togglePreviewPane(true);
                    };
                    tr.innerHTML = `
                        <td><span class="status-tag status-ok">AGRÉÉ MIPH</span></td>
                        <td style="font-weight:600; color:var(--text-main);">${item.name_fr}</td>
                        <td class="col-hide-md">${item.legal_status || 'SPA'}</td>
                        <td class="col-hide-sm" style="font-family:monospace; color:var(--accent-blue);">${item.miph_license_number || '-'}</td>
                        <td style="font-weight:600;">${item.technical_director_name}</td>
                        <td class="col-hide-lg" style="font-family:monospace; color:var(--text-muted);">${item.technical_director_decision_ref || '-'}</td>
                        <td class="col-hide-sm">${item.wilaya || '-'}</td>
                    `;
                    tbody.appendChild(tr);
                });
            } else if (section === 'sites') {
                thead.innerHTML = `
                    <tr>
                        <th style="width: 90px;">Statut BPF</th>
                        <th>Site de Fabrication / Contrôle</th>
                        <th class="col-hide-sm">Rôle Réglementaire (1.H.8)</th>
                        <th>Pays</th>
                        <th class="col-hide-lg">Licence ML</th>
                        <th class="col-hide-md">Certificat BPF/GMP</th>
                        <th class="col-hide-lg">Autorité Émettrice</th>
                    </tr>
                `;
                if (items.length === 0) {
                    tbody.innerHTML = `<tr><td colspan="7" style="text-align:center; padding:15px; color:var(--text-dim);">Aucun site intervenant enregistré.</td></tr>`;
                    return;
                }
                items.forEach((item, idx) => {
                    const tr = document.createElement("tr");
                    if (idx === selectedRefIndex) tr.classList.add("selected");
                    tr.onclick = () => {
                        selectedRefIndex = idx;
                        renderRefTable(section, items);
                        if (window.innerWidth < 1400) togglePreviewPane(true);
                    };
                    tr.innerHTML = `
                        <td><span class="status-tag status-ok">CONFORME BPF</span></td>
                        <td style="font-weight:600; color:var(--text-main);">${item.site_name_fr}</td>
                        <td class="col-hide-sm"><span style="background:var(--badge-bg); padding:2px 6px; border-radius:3px; font-weight:600; font-size:10px;">${item.site_role}</span></td>
                        <td style="font-weight:700;">${item.country_code}</td>
                        <td class="col-hide-lg" style="font-family:monospace;">${item.manufacturing_license_number || '-'}</td>
                        <td class="col-hide-md" style="font-family:monospace; color:var(--accent-emerald);">${item.gmp_certificate_number || '-'}</td>
                        <td class="col-hide-lg">${item.gmp_issuing_authority || '-'}</td>
                    `;
                    tbody.appendChild(tr);
                });
            } else if (section === 'medicines') {
                thead.innerHTML = `
                    <tr>
                        <th style="width: 120px;">Typologie ANPP</th>
                        <th>Spécialité Commerciale</th>
                        <th>DCI (Substance Active)</th>
                        <th class="col-hide-sm">Forme Galénique</th>
                        <th>Dosage</th>
                        <th class="col-hide-lg">Dossier Économique</th>
                        <th class="col-hide-md">Code ATC</th>
                    </tr>
                `;
                if (items.length === 0) {
                    tbody.innerHTML = `<tr><td colspan="7" style="text-align:center; padding:15px; color:var(--text-dim);">Aucun médicament enregistré. Cliquez sur '➕ Nouveau' pour en créer un.</td></tr>`;
                    return;
                }
                items.forEach((item, idx) => {
                    const tr = document.createElement("tr");
                    if (idx === selectedRefIndex) tr.classList.add("selected");
                    tr.onclick = () => {
                        selectedRefIndex = idx;
                        renderRefTable(section, items);
                        if (window.innerWidth < 1400) togglePreviewPane(true);
                    };
                    let tagClass = "status-ok";
                    let tagLabel = "GÉNÉRIQUE LOCAL";
                    if (item.origin_type === 'princeps') {
                        tagClass = "status-rev";
                        tagLabel = "PRINCEPS RÉF.";
                    } else if (item.origin_type === 'generique_importe') {
                        tagClass = "status-req";
                        tagLabel = "GÉNÉRIQUE IMPORTÉ";
                    }
                    let ecoLabel = item.economic_file_type === 'pcsu' ? `PCSU : ${item.pcsu_amount || 0} DZD` : `FOB : ${item.fob_price || 0} ${item.fob_currency || 'EUR'}`;
                    tr.innerHTML = `
                        <td><span class="status-tag ${tagClass}">${tagLabel}</span></td>
                        <td style="font-weight:600; color:var(--text-main);">${item.brand_name_fr}</td>
                        <td style="color:var(--accent-blue);">${item.dci_fr}</td>
                        <td class="col-hide-sm">${item.pharmaceutical_form_fr}</td>
                        <td style="font-weight:600;">${item.dosage_fr}</td>
                        <td class="col-hide-lg" style="font-family:monospace; font-size:11px;">${ecoLabel}</td>
                        <td class="col-hide-md" style="font-family:monospace; color:var(--text-dim);">${item.atc_code || '-'}</td>
                    `;
                    tbody.appendChild(tr);
                });
            } else if (section === 'devices') {
                thead.innerHTML = `
                    <tr>
                        <th style="width: 80px;">Classe DM</th>
                        <th>Nom Commercial du DM</th>
                        <th class="col-hide-sm">Référence Fabricant</th>
                        <th class="col-hide-lg">Règle Arrêté 10/05/2021</th>
                        <th class="col-hide-md">Stérile</th>
                        <th class="col-hide-md">Usage Unique</th>
                        <th>Statut Homolog.</th>
                    </tr>
                `;
                if (items.length === 0) {
                    tbody.innerHTML = `<tr><td colspan="7" style="text-align:center; padding:15px; color:var(--text-dim);">Aucun dispositif médical enregistré.</td></tr>`;
                    return;
                }
                items.forEach((item, idx) => {
                    const tr = document.createElement("tr");
                    if (idx === selectedRefIndex) tr.classList.add("selected");
                    tr.onclick = () => {
                        selectedRefIndex = idx;
                        renderRefTable(section, items);
                        if (window.innerWidth < 1400) togglePreviewPane(true);
                    };
                    tr.innerHTML = `
                        <td><span class="status-tag status-ok">CLASSE ${item.dm_class}</span></td>
                        <td style="font-weight:600; color:var(--text-main);">${item.commercial_name_fr}</td>
                        <td class="col-hide-sm" style="font-family:monospace; color:var(--text-muted);">${item.manufacturer_reference}</td>
                        <td class="col-hide-lg" style="font-size:11px;">${item.classification_rule || '-'}</td>
                        <td class="col-hide-md">${item.is_sterile ? '✓ Oui' : 'Non'}</td>
                        <td class="col-hide-md">${item.is_single_use ? '✓ Oui' : 'Non'}</td>
                        <td><span style="color:var(--accent-emerald); font-weight:600;">EN COURS</span></td>
                    `;
                    tbody.appendChild(tr);
                });
            }
            renderRefPreview(section, items[selectedRefIndex]);
        }

        function renderRefPreview(section, item) {
            const preview = document.getElementById("preview-body");
            if (!item) {
                preview.innerHTML = `<div style="text-align:center; padding:30px; color:var(--text-dim);">Sélectionnez un élément dans le tableau pour afficher sa fiche légale.</div>`;
                return;
            }

            if (section === 'establishments') {
                preview.innerHTML = `
                    <div class="pdf-sim-page">
                        <div class="pdf-sim-header">
                            RÉPUBLIQUE ALGÉRIENNE DÉMOCRATIQUE ET POPULAIRE<br>
                            MINISTÈRE DE L'INDUSTRIE ET DE LA PRODUCTION PHARMACEUTIQUE
                        </div>
                        <div style="font-size:12px; font-weight:700; color:#0369a1; text-align:center; margin-bottom:10px;">
                            AGRÉMENT D'ÉTABLISSEMENT PHARMACEUTIQUE (DÉCRET 20-325)
                        </div>
                        <div style="font-size:11px; line-height:1.6;">
                            Il est certifié que l'établissement :<br>
                            <strong>${item.name_fr}</strong> (${item.legal_status || 'SPA'})<br>
                            Adresse : ${item.address_fr || 'Non renseignée'}, ${item.wilaya || 'Alger'}<br><br>
                            Est dûment agréé sous le numéro :<br>
                            <span style="font-size:13px; font-weight:700; color:var(--accent-blue); font-family:monospace;">${item.miph_license_number}</span><br><br>
                            <strong>Pharmacien Directeur Technique Titulaire (1.A.7) :</strong><br>
                            ${item.technical_director_name}<br>
                            Décision ministérielle nominative : <code>${item.technical_director_decision_ref || 'DEC-MIPH-DT-01'}</code>
                        </div>
                    </div>

                    <div class="meta-box">
                        <div class="meta-row"><span class="meta-label">Raison Sociale :</span><span class="meta-val">${item.name_fr}</span></div>
                        <div class="meta-row"><span class="meta-label">Agrément MIPH :</span><span class="meta-val">${item.miph_license_number}</span></div>
                        <div class="meta-row"><span class="meta-label">Pharmacien DT :</span><span class="meta-val">${item.technical_director_name}</span></div>
                        <div class="meta-row"><span class="meta-label">Wilaya :</span><span class="meta-val">${item.wilaya}</span></div>
                        <div class="meta-row"><span class="meta-label">Téléphone :</span><span class="meta-val">${item.phone || '-'}</span></div>
                        <div class="meta-row"><span class="meta-label">Email :</span><span class="meta-val">${item.email || '-'}</span></div>
                    </div>

                    <div style="margin-top:14px; padding-top:10px; border-top:1px solid var(--border-color);">
                        <button class="btn-pill" data-perm="REF_EDIT" style="background:var(--accent-blue); color:white; border:none; width:100%; justify-content:center; padding:7px 12px; font-weight:600;" onclick="openEditEstablishmentModal()">
                            ✏️ Mettre à jour l'Établissement (Agrément / PDT)
                        </button>
                    </div>
                `;
            } else if (section === 'sites') {
                preview.innerHTML = `
                    <div class="pdf-sim-page">
                        <div class="pdf-sim-header">
                            AGENCE NATIONALE DES PRODUITS PHARMACEUTIQUES (ANPP)<br>
                            CARTOGRAPHIE DES SITES INTERVENANTS (1.H.8 & 1.M)
                        </div>
                        <div style="font-size:12px; font-weight:700; color:#0369a1; text-align:center; margin-bottom:10px;">
                            CERTIFICAT DE BONNES PRATIQUES DE FABRICATION (BPF)
                        </div>
                        <div style="font-size:11px; line-height:1.6;">
                            Site : <strong>${item.site_name_fr}</strong><br>
                            Pays : <strong>${item.country_code}</strong> | Rôle : <code>${item.site_role}</code><br><br>
                            Certificat GMP/BPF : <strong>${item.gmp_certificate_number || '-'}</strong><br>
                            Autorité émettrice : ${item.gmp_issuing_authority || 'ANPP'}<br>
                            Licence de fabrication (ML) : ${item.manufacturing_license_number || '-'}
                        </div>
                    </div>

                    <div style="margin-top:14px; padding-top:10px; border-top:1px solid var(--border-color);">
                        <button class="btn-pill" data-perm="REF_EDIT" style="background:var(--accent-blue); color:white; border:none; width:100%; justify-content:center; padding:7px 12px; font-weight:600;" onclick="openEditManufacturingSiteModal()">
                            ✏️ Mettre à jour le Site (Certificat GMP / Licence ML)
                        </button>
                    </div>
                `;
            } else if (section === 'medicines') {
                preview.innerHTML = `
                    <div class="pdf-sim-page">
                        <div class="pdf-sim-header">
                            AGENCE NATIONALE DES PRODUITS PHARMACEUTIQUES (ANPP)<br>
                            FICHE TECHNIQUE RÉGLEMENTAIRE MÉDICAMENT
                        </div>
                        <div style="font-size:13px; font-weight:700; color:#0369a1; text-align:center; margin-bottom:10px;">
                            ${item.brand_name_fr}
                        </div>
                        <div style="font-size:11px; line-height:1.6;">
                            <strong>DCI :</strong> ${item.dci_fr}<br>
                            <strong>Forme galénique :</strong> ${item.pharmaceutical_form_fr}<br>
                            <strong>Dosage :</strong> ${item.dosage_fr} | <strong>Voie :</strong> ${item.administration_route_fr}<br>
                            <strong>Code ATC :</strong> <code>${item.atc_code || '-'}</code><br>
                            <strong>Typologie :</strong> <code>${item.origin_type}</code><br>
                            <strong>Péremption :</strong> ${item.shelf_life_months} mois<br>
                            <strong>Conservation :</strong> ${item.storage_conditions_fr || '-'}<br><br>
                            <strong>Volet Économique :</strong><br>
                            ${item.economic_file_type === 'pcsu' ? `PCSU Sortie Usine : <strong>${item.pcsu_amount} DZD</strong> (Fabrication Locale)` : `Prix FOB : <strong>${item.fob_price} ${item.fob_currency}</strong> (Importation avec Benchmark 1.H.5)`}
                        </div>
                    </div>
                    <div class="meta-box">
                        <div class="meta-row"><span class="meta-label">Spécialité :</span><span class="meta-val">${item.brand_name_fr}</span></div>
                        <div class="meta-row"><span class="meta-label">DCI :</span><span class="meta-val">${item.dci_fr}</span></div>
                        <div class="meta-row"><span class="meta-label">Typologie :</span><span class="meta-val">${item.origin_type}</span></div>
                        <div class="meta-row"><span class="meta-label">Monographie Ph. Eur. :</span><span class="meta-val">${item.has_pharmacopoeia_monograph ? 'Oui' : 'Non'}</span></div>
                        <div class="meta-row"><span class="meta-label">CEP EDQM :</span><span class="meta-val">${item.cep_number || 'Non requis'}</span></div>
                    </div>
                `;
            } else if (section === 'devices') {
                preview.innerHTML = `
                    <div class="pdf-sim-page">
                        <div class="pdf-sim-header">
                            AGENCE NATIONALE DES PRODUITS PHARMACEUTIQUES (ANPP)<br>
                            DOSSIER D'HOMOLOGATION DISPOSITIF MÉDICAL (DÉCRET 20-324)
                        </div>
                        <div style="font-size:13px; font-weight:700; color:#0369a1; text-align:center; margin-bottom:10px;">
                            ${item.commercial_name_fr}
                        </div>
                        <div style="font-size:11px; line-height:1.6;">
                            <strong>Classe de risque :</strong> Classe ${item.dm_class}<br>
                            <strong>Référence Fabricant :</strong> <code>${item.manufacturer_reference}</code><br>
                            <strong>Règle de classification :</strong> ${item.classification_rule || 'Arrêté 10/05/2021'}<br>
                            <strong>Stérilité :</strong> ${item.is_sterile ? '✓ Stérile' : 'Non stérile'}<br>
                            <strong>Usage unique :</strong> ${item.is_single_use ? '✓ Usage unique' : 'Réutilisable'}<br><br>
                            <strong>Indication Clinique :</strong><br>
                            ${item.intended_use_fr || 'Non renseignée'}
                        </div>
                    </div>
                `;
            }

            if (typeof applyPermissionsUI === 'function') {
                applyPermissionsUI();
            }
        }


        let editingEstID = null;
        let editingSiteID = null;

        async function populateEstablishmentsDropdown(selectedID) {
            const select = document.getElementById("site-est-id");
            if (!select) return;
            select.innerHTML = '<option value="">-- Aucun / Site Tiers Extérieur ou Façonnier --</option>';
            try {
                const res = await fetch('/api/v1/establishments');
                if (res.ok) {
                    const data = await res.json();
                    const list = data.establishments || data.items || [];
                    list.forEach(est => {
                        const opt = document.createElement("option");
                        opt.value = est.id;
                        opt.textContent = `${est.name_fr} (${est.miph_license_number || 'Agréé'})`;
                        if (selectedID && est.id === selectedID) {
                            opt.selected = true;
                        }
                        select.appendChild(opt);
                    });
                }
            } catch(e) {}
        }

        function openNewPharmaModal() {
            closeAllModals();
            if (currentRefSection === 'establishments') {
                editingEstID = null;
                document.getElementById("form-new-est").reset();
                const title = document.getElementById("modal-est-title");
                if (title) title.innerText = "Enregistrement d'un Établissement Pharmaceutique (MIPH / Décret 20-325)";
                const btn = document.getElementById("modal-est-submit-btn");
                if (btn) btn.innerText = "Enregistrer l'Établissement (Base SQL)";
                document.getElementById("modal-new-est").classList.add("open");
            } else if (currentRefSection === 'sites') {
                editingSiteID = null;
                document.getElementById("form-new-site").reset();
                const title = document.getElementById("modal-site-title");
                if (title) title.innerText = "Enregistrement d'un Site Intervenant (ANPP 1.H.8 & 1.M)";
                const btn = document.getElementById("modal-site-submit-btn");
                if (btn) btn.innerText = "Enregistrer le Site (Base SQL)";
                populateEstablishmentsDropdown();
                document.getElementById("modal-new-site").classList.add("open");
            } else if (currentRefSection === 'medicines') {
                document.getElementById("modal-new-med").classList.add("open");
            } else if (currentRefSection === 'devices') {
                document.getElementById("modal-new-dm").classList.add("open");
            } else {
                document.getElementById("modal-new-est").classList.add("open");
            }
        }

        function openEditEstablishmentModal() {
            const item = currentRefList[selectedRefIndex];
            if (!item) return;
            closeAllModals();
            editingEstID = item.id;
            document.getElementById("est-name-fr").value = item.name_fr || '';
            document.getElementById("est-legal-status").value = item.legal_status || 'SPA';
            document.getElementById("est-license-number").value = item.miph_license_number || '';
            document.getElementById("est-license-date").value = item.miph_license_grant_date ? item.miph_license_grant_date.slice(0, 10) : '';
            document.getElementById("est-dt-name").value = item.technical_director_name || '';
            document.getElementById("est-decision-ref").value = item.technical_director_decision_ref || '';
            document.getElementById("est-wilaya").value = item.wilaya || 'Alger';
            document.getElementById("est-phone").value = item.phone || '';
            document.getElementById("est-email").value = item.email || '';
            const title = document.getElementById("modal-est-title");
            if (title) title.innerText = "Mise à jour de l'Établissement Pharmaceutique";
            const btn = document.getElementById("modal-est-submit-btn");
            if (btn) btn.innerText = "Enregistrer les Modifications (Base SQL)";
            document.getElementById("modal-new-est").classList.add("open");
        }

        function openEditManufacturingSiteModal() {
            const item = currentRefList[selectedRefIndex];
            if (!item) return;
            closeAllModals();
            editingSiteID = item.id;
            document.getElementById("site-name-fr").value = item.site_name_fr || '';
            document.getElementById("site-country-code").value = item.country_code || 'DZ';
            document.getElementById("site-role").value = item.site_role || 'formulateur_pf';
            document.getElementById("site-ml-number").value = item.manufacturing_license_number || '';
            document.getElementById("site-gmp-number").value = item.gmp_certificate_number || '';
            document.getElementById("site-gmp-authority").value = item.gmp_issuing_authority || 'ANPP';
            document.getElementById("site-gmp-grant").value = item.gmp_grant_date ? item.gmp_grant_date.slice(0, 10) : '';
            document.getElementById("site-gmp-expiry").value = item.gmp_expiry_date ? item.gmp_expiry_date.slice(0, 10) : '';
            document.getElementById("site-address").value = item.address || '';
            populateEstablishmentsDropdown(item.establishment_id);
            const title = document.getElementById("modal-site-title");
            if (title) title.innerText = "Mise à jour du Site Intervenant (1.H.8 & 1.M)";
            const btn = document.getElementById("modal-site-submit-btn");
            if (btn) btn.innerText = "Enregistrer les Modifications (Base SQL)";
            document.getElementById("modal-new-site").classList.add("open");
        }

        async function handleCreateManufacturingSite(e) {
            if (e) e.preventDefault();
            const estVal = document.getElementById("site-est-id").value;
            const payload = {
                site_name_fr: document.getElementById("site-name-fr").value.trim(),
                country_code: document.getElementById("site-country-code").value.trim().toUpperCase(),
                site_role: document.getElementById("site-role").value,
                establishment_id: estVal ? estVal : null,
                manufacturing_license_number: document.getElementById("site-ml-number").value.trim(),
                gmp_certificate_number: document.getElementById("site-gmp-number").value.trim(),
                gmp_issuing_authority: document.getElementById("site-gmp-authority").value.trim(),
                gmp_grant_date: document.getElementById("site-gmp-grant").value ? new Date(document.getElementById("site-gmp-grant").value).toISOString() : null,
                gmp_expiry_date: document.getElementById("site-gmp-expiry").value ? new Date(document.getElementById("site-gmp-expiry").value).toISOString() : null,
                address: document.getElementById("site-address").value.trim(),
                is_active: true
            };

            const isEdit = !!editingSiteID;
            const url = isEdit ? `/api/v1/manufacturing-sites/${editingSiteID}` : '/api/v1/manufacturing-sites';
            const method = isEdit ? 'PUT' : 'POST';

            try {
                const res = await fetch(url, {
                    method: method,
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(payload)
                });
                if (res.ok) {
                    closeModal('modal-new-site');
                    const msg = isEdit ? "Le site de fabrication a été mis à jour avec succès." : "Le site de fabrication / contrôle a été inscrit au référentiel ANPP avec succès.";
                    showNotifyModal(isEdit ? "Site Mis à Jour" : "Site Enregistré", msg, "🏭");
                    editingSiteID = null;
                    loadRefSection('sites');
                } else {
                    const err = await res.json();
                    showNotifyModal("Erreur d'Enregistrement", err.error || "Échec de validation", "⚠️");
                }
            } catch(err) {
                showNotifyModal("Erreur Réseau", err.message, "⚠️");
            }
        }

        function toggleEcoFields() {
            const origin = document.getElementById("med-origin").value;
            const pcsuGroup = document.getElementById("eco-pcsu-group");
            const fobGroup = document.getElementById("eco-fob-group");
            const title = document.getElementById("eco-section-title");
            if (origin === 'generique_local') {
                pcsuGroup.style.display = 'flex';
                fobGroup.style.display = 'none';
                title.innerText = "💰 Volet Économique (1.H.4 - PCSU Fabrication Locale)";
            } else {
                pcsuGroup.style.display = 'none';
                fobGroup.style.display = 'flex';
                title.innerText = "💰 Volet Économique (1.H.4 & 1.H.5 - Prix FOB & Benchmark International)";
            }
        }

        async function handleCreateEstablishment(e) {
            if (e) e.preventDefault();
            const payload = {
                name_fr: document.getElementById("est-name-fr").value.trim(),
                legal_status: document.getElementById("est-legal-status").value,
                miph_license_number: document.getElementById("est-license-number").value.trim(),
                miph_license_grant_date: document.getElementById("est-license-date").value || null,
                technical_director_name: document.getElementById("est-dt-name").value.trim(),
                technical_director_decision_ref: document.getElementById("est-decision-ref").value.trim(),
                wilaya: document.getElementById("est-wilaya").value.trim(),
                phone: document.getElementById("est-phone").value.trim(),
                email: document.getElementById("est-email").value.trim(),
                is_active: true
            };

            const isEdit = !!editingEstID;
            const url = isEdit ? `/api/v1/establishments/${editingEstID}` : '/api/v1/establishments';
            const method = isEdit ? 'PUT' : 'POST';

            try {
                const res = await fetch(url, {
                    method: method,
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(payload)
                });
                if (res.ok) {
                    closeModal('modal-new-est');
                    const msg = isEdit ? "L'établissement a été mis à jour avec succès dans PostgreSQL." : "L'établissement pharmaceutique a été inscrit en base PostgreSQL et son agrément MIPH a été validé.";
                    showNotifyModal(isEdit ? "Établissement Mis à Jour" : "Établissement Enregistré", msg, "🏢");
                    editingEstID = null;
                    loadRefSection('establishments');
                } else {
                    const err = await res.json();
                    showNotifyModal("Erreur d'Enregistrement", err.error || "Échec de validation", "⚠️");
                }
            } catch(err) {
                showNotifyModal("Erreur Réseau", err.message, "⚠️");
            }
        }

        async function handleCreateMedicine(e) {
            if (e) e.preventDefault();
            const origin = document.getElementById("med-origin").value;
            const payload = {
                origin_type: origin,
                brand_name_fr: document.getElementById("med-brand-name").value.trim(),
                dci_fr: document.getElementById("med-dci").value.trim(),
                dosage_fr: document.getElementById("med-dosage").value.trim(),
                pharmaceutical_form_fr: document.getElementById("med-form").value.trim(),
                administration_route_fr: document.getElementById("med-route").value.trim(),
                atc_code: document.getElementById("med-atc").value.trim() || null,
                shelf_life_months: parseInt(document.getElementById("med-shelf-life").value) || 36,
                economic_file_type: origin === 'generique_local' ? 'pcsu' : 'fob_benchmark',
                pcsu_amount: origin === 'generique_local' ? parseFloat(document.getElementById("med-pcsu").value) || 0 : null,
                fob_price: origin !== 'generique_local' ? parseFloat(document.getElementById("med-fob").value) || 0 : null,
                fob_currency: document.getElementById("med-currency").value,
                has_chamber_of_commerce_visa: origin !== 'generique_local' ? document.getElementById("med-visa-cc").checked : false,
                status: 'en_preparation'
            };
            try {
                const res = await fetch('/api/v1/medicines', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(payload)
                });
                if (res.ok) {
                    closeModal('modal-new-med');
                    showNotifyModal("Médicament Enregistré", "La spécialité pharmaceutique a été enregistrée avec ses données médico-économiques selon la typologie ANPP 2025.", "💊");
                    loadRefSection('medicines');
                } else {
                    const err = await res.json();
                    showNotifyModal("Erreur d'Enregistrement", err.error || "Échec de validation", "⚠️");
                }
            } catch(err) {
                showNotifyModal("Erreur Réseau", err.message, "⚠️");
            }
        }

        async function handleCreateDevice(e) {
            if (e) e.preventDefault();
            const payload = {
                commercial_name_fr: document.getElementById("dm-name").value.trim(),
                manufacturer_reference: document.getElementById("dm-ref").value.trim(),
                dm_class: document.getElementById("dm-class").value,
                classification_rule: document.getElementById("dm-rule").value.trim(),
                is_sterile: document.getElementById("dm-sterile").checked,
                is_single_use: document.getElementById("dm-single-use").checked,
                intended_use_fr: document.getElementById("dm-intended-use").value.trim(),
                status: 'en_cours'
            };
            try {
                const res = await fetch('/api/v1/medical-devices', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify(payload)
                });
                if (res.ok) {
                    closeModal('modal-new-dm');
                    showNotifyModal("Dispositif Médical Enregistré", "Le dossier d'homologation DM a été créé conformément au Décret 20-324 et à l'Arrêté du 10 mai 2021.", "💉");
                    loadRefSection('devices');
                } else {
                    const err = await res.json();
                    showNotifyModal("Erreur d'Enregistrement", err.error || "Échec de validation", "⚠️");
                }
            } catch(err) {
                showNotifyModal("Erreur Réseau", err.message, "⚠️");
            }
        }
