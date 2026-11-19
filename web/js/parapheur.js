/**
 * Sanad Pharma (Homolog) — Cockpit de Supervision & Parapheur Électronique Pharmacien DT
 * Conforme aux spécifications Section 7.4.6 du CDC & Normes 21 CFR Part 11 / GAMP 5
 */

(function () {
    let parapheurItems = [];
    let selectedItemIds = new Set();
    let currentDossierFilter = 'ALL';
    let currentModuleFilter = 'ALL';
    let currentDeptFilter = 'ALL';
    let currentGatingFilter = 'ALL';
    let currentSearchTerm = '';
    let isParapheurMaximized = false;
    let parapheurPrevStyles = null;

    // Initialisation au chargement de la page
    window.addEventListener('DOMContentLoaded', () => {
        // Rafraîchir le badge du parapheur si l'utilisateur est connecté
        setTimeout(() => {
            if (typeof currentUser !== 'undefined' && currentUser) {
                refreshParapheurBadge();
            }
        }, 1200);

        // Raccourci clavier Échap pour fermer
        window.addEventListener('keydown', (e) => {
            if (e.key === 'Escape') {
                const m = document.getElementById('modal-parapheur-dt');
                if (m && (m.classList.contains('active') || m.style.display === 'flex' || m.style.display === 'block')) {
                    const confirmModal = document.getElementById('modal-parapheur-batch-confirm');
                    if (confirmModal && (confirmModal.classList.contains('active') || confirmModal.style.display === 'flex')) {
                        closeBatchConfirmModal();
                    } else {
                        closeParapheurModal();
                    }
                }
            }
        });
    });

    /**
     * Bascule l'affichage plein écran / taille normale de la modale
     */
    window.toggleMaximizeParapheurModal = function () {
        const dialog = document.getElementById('modal-parapheur-dialog');
        const btn = document.getElementById('btn-parapheur-maximize');
        if (!dialog) return;

        if (!isParapheurMaximized) {
            // Sauvegarder les dimensions actuelles (y compris si l'utilisateur l'a redimensionnée manuellement)
            parapheurPrevStyles = {
                width: dialog.style.width,
                height: dialog.style.height,
                maxWidth: dialog.style.maxWidth,
                maxHeight: dialog.style.maxHeight
            };

            dialog.style.width = 'calc(100vw - 20px)';
            dialog.style.height = 'calc(100vh - 20px)';
            dialog.style.maxWidth = 'calc(100vw - 20px)';
            dialog.style.maxHeight = 'calc(100vh - 20px)';
            dialog.style.borderRadius = '6px';

            if (btn) {
                btn.innerHTML = '🗗';
                btn.title = "Restaurer la taille normale";
            }
            isParapheurMaximized = true;
        } else {
            // Restaurer les dimensions précédentes
            dialog.style.width = (parapheurPrevStyles && parapheurPrevStyles.width) || '1120px';
            dialog.style.height = (parapheurPrevStyles && parapheurPrevStyles.height) || '88vh';
            dialog.style.maxWidth = (parapheurPrevStyles && parapheurPrevStyles.maxWidth) || 'min(1120px, calc(100vw - 24px))';
            dialog.style.maxHeight = (parapheurPrevStyles && parapheurPrevStyles.maxHeight) || '96vh';
            dialog.style.borderRadius = '';

            if (btn) {
                btn.innerHTML = '⛶';
                btn.title = "Agrandir en plein écran";
            }
            isParapheurMaximized = false;
        }
    };

    /**
     * Ouvre la modale principale du Cockpit & Parapheur DT
     */
    window.openParapheurModal = function () {
        if (typeof openModal === 'function') {
            openModal('modal-parapheur-dt');
        } else {
            const m = document.getElementById('modal-parapheur-dt');
            if (m) m.classList.add('active');
        }
        loadParapheurItems();
    };

    /**
     * Ferme la modale du Parapheur DT
     */
    window.closeParapheurModal = function () {
        if (typeof closeModal === 'function') {
            closeModal('modal-parapheur-dt');
        } else {
            const m = document.getElementById('modal-parapheur-dt');
            if (m) m.classList.remove('active');
        }
    };

    /**
     * Actualise le compteur badge dans la barre latérale
     */
    window.refreshParapheurBadge = async function () {
        const badge = document.getElementById('badge-parapheur-count');
        const btn = document.getElementById('act-parapheur');
        const mobBtn = document.getElementById('mob-act-parapheur');
        if (!badge) return;

        // Contrôle de visibilité : Pharmacien DT ou Admin labo
        const isEligible = typeof hasPermission === 'function' 
            ? (hasPermission('DOC_APPROVE_PDT') || (typeof hasRole === 'function' && (hasRole('directeur_technique', 'admin_laboratoire'))))
            : true;

        if (btn) btn.style.display = isEligible ? '' : 'none';
        if (mobBtn) mobBtn.style.display = isEligible ? '' : 'none';

        if (!isEligible) return;

        try {
            const res = await fetch('/api/v1/parapheur/pending');
            if (!res.ok) return;
            const data = await res.json();
            const count = data.count || (data.items ? data.items.length : 0);
            if (count > 0) {
                badge.textContent = count;
                badge.style.display = 'inline-block';
            } else {
                badge.textContent = '0';
                badge.style.display = 'none';
            }
        } catch (e) {
            console.debug('Badge parapheur refresh check:', e);
        }
    };

    /**
     * Charge toutes les pièces au statut pending_approval
     */
    window.loadParapheurItems = async function () {
        const tbody = document.getElementById('parapheur-table-tbody');
        if (tbody) {
            tbody.innerHTML = `<tr><td colspan="9" style="text-align:center; padding:30px; color:var(--text-dim);"><div style="font-size:18px; margin-bottom:6px;">⏳</div>Chargement des pièces en attente de visa réglementaire...</td></tr>`;
        }

        selectedItemIds.clear();
        updateBatchBar();

        try {
            const res = await fetch('/api/v1/parapheur/pending');
            if (res.status === 401 && typeof handleAuthExpired === 'function') {
                handleAuthExpired();
                return;
            }
            if (!res.ok) throw new Error('HTTP ' + res.status);
            const data = await res.json();
            parapheurItems = data.items || [];

            // Mise à jour du badge global
            const badge = document.getElementById('badge-parapheur-count');
            if (badge) {
                if (parapheurItems.length > 0) {
                    badge.textContent = parapheurItems.length;
                    badge.style.display = 'inline-block';
                } else {
                    badge.textContent = '0';
                    badge.style.display = 'none';
                }
            }

            // Mettre à jour les sélecteurs de filtre
            populateDossierFilter(parapheurItems);

            // Mettre à jour les KPIs
            updateParapheurKPIs(parapheurItems);

            // Rendu de la table
            applyParapheurFilters();

        } catch (err) {
            console.error('Erreur chargement parapheur:', err);
            if (tbody) {
                tbody.innerHTML = `<tr><td colspan="9" style="text-align:center; padding:25px; color:#ef4444;">⚠️ Échec du chargement du parapheur : ${err.message}</td></tr>`;
            }
        }
    };

    /**
     * Remplit le menu déroulant des dossiers
     */
    function populateDossierFilter(items) {
        const select = document.getElementById('parapheur-filter-dossier');
        if (!select) return;

        const currentVal = select.value || 'ALL';
        const dossiersMap = new Map();
        items.forEach(it => {
            if (!dossiersMap.has(it.dossier_id)) {
                dossiersMap.set(it.dossier_id, {
                    code: it.dossier_code,
                    product: it.product_name
                });
            }
        });

        let html = `<option value="ALL">Tous les Dossiers (${dossiersMap.size})</option>`;
        dossiersMap.forEach((val, id) => {
            const label = `${val.code} — ${val.product || 'Dossier CTD'}`;
            html += `<option value="${id}">${escapeHtml(label)}</option>`;
        });
        select.innerHTML = html;
        select.value = dossiersMap.has(currentVal) ? currentVal : 'ALL';
    }

    /**
     * Met à jour les 4 blocs KPIs du Cockpit
     */
    function updateParapheurKPIs(items) {
        const totalCountEl = document.getElementById('kpi-parapheur-total');
        const dossiersCountEl = document.getElementById('kpi-parapheur-dossiers');
        const gatingCountEl = document.getElementById('kpi-parapheur-gating');
        const urgentCountEl = document.getElementById('kpi-parapheur-urgent');

        if (totalCountEl) totalCountEl.textContent = items.length;

        const uniqueDossiers = new Set(items.map(it => it.dossier_id));
        if (dossiersCountEl) dossiersCountEl.textContent = uniqueDossiers.size;

        const blockedCount = items.filter(it => it.is_gating_blocked).length;
        if (gatingCountEl) gatingCountEl.textContent = blockedCount;

        const urgentCount = items.filter(it => it.days_to_target_deposit !== undefined && it.days_to_target_deposit !== null && it.days_to_target_deposit <= 30).length;
        if (urgentCountEl) urgentCountEl.textContent = urgentCount;
    }

    /**
     * Applique les filtres actifs et restitue le tableau
     */
    window.applyParapheurFilters = function () {
        const dossierSelect = document.getElementById('parapheur-filter-dossier');
        const moduleSelect = document.getElementById('parapheur-filter-module');
        const deptSelect = document.getElementById('parapheur-filter-dept');
        const gatingSelect = document.getElementById('parapheur-filter-gating');
        const searchInput = document.getElementById('parapheur-search-input');

        currentDossierFilter = dossierSelect ? dossierSelect.value : 'ALL';
        currentModuleFilter = moduleSelect ? moduleSelect.value : 'ALL';
        currentDeptFilter = deptSelect ? deptSelect.value : 'ALL';
        currentGatingFilter = gatingSelect ? gatingSelect.value : 'ALL';
        currentSearchTerm = searchInput ? searchInput.value.trim().toLowerCase() : '';

        const filtered = parapheurItems.filter(it => {
            if (currentDossierFilter !== 'ALL' && it.dossier_id !== currentDossierFilter) return false;
            if (currentModuleFilter !== 'ALL') {
                if (currentModuleFilter === 'DM' && !it.item_code.startsWith('DM.')) return false;
                if (currentModuleFilter !== 'DM' && String(it.module_number) !== currentModuleFilter) return false;
            }
            if (currentDeptFilter !== 'ALL' && it.assigned_department !== currentDeptFilter) return false;
            if (currentGatingFilter === 'UNBLOCKED' && it.is_gating_blocked) return false;
            if (currentGatingFilter === 'BLOCKED' && !it.is_gating_blocked) return false;

            if (currentSearchTerm) {
                const searchCorpus = `${it.item_code} ${it.title_fr} ${it.dossier_code} ${it.product_name} ${it.checker_user_name || ''} ${it.maker_user_name || ''}`.toLowerCase();
                if (!searchCorpus.includes(currentSearchTerm)) return false;
            }

            return true;
        });

        renderParapheurTable(filtered);
    };

    /**
     * Rendu HTML du tableau haute densité du Parapheur
     */
    function renderParapheurTable(items) {
        const tbody = document.getElementById('parapheur-table-tbody');
        const filteredCountBadge = document.getElementById('parapheur-filtered-count');
        const selectAllCheckbox = document.getElementById('parapheur-select-all');

        if (filteredCountBadge) {
            filteredCountBadge.textContent = `${items.length} pièce(s)`;
        }

        if (!tbody) return;

        if (items.length === 0) {
            tbody.innerHTML = `
                <tr>
                    <td colspan="9" style="text-align:center; padding:40px; color:var(--text-dim);">
                        <div style="font-size:26px; margin-bottom:8px;">⚖️</div>
                        <div style="font-weight:700; font-size:13px; color:var(--text-main);">Aucune pièce en attente de visa</div>
                        <div style="font-size:11.5px; margin-top:4px;">Toutes les sections soumises ont été visées ou aucun document ne correspond aux filtres actuels.</div>
                    </td>
                </tr>
            `;
            if (selectAllCheckbox) {
                selectAllCheckbox.checked = false;
                selectAllCheckbox.disabled = true;
            }
            return;
        }

        if (selectAllCheckbox) selectAllCheckbox.disabled = false;

        const uid = (typeof currentUser !== 'undefined' && currentUser && currentUser.id) ? currentUser.id : '';

        tbody.innerHTML = items.map(it => {
            const isChecked = selectedItemIds.has(it.item_id);
            const isSelfMade = it.maker_user_id && uid && it.maker_user_id === uid;
            const isBlocked = it.is_gating_blocked;
            const isSelectable = !isSelfMade && !isBlocked;

            // Badge module
            let modBadge = `<span style="font-family:monospace; font-weight:800; font-size:10.5px; background:var(--bg-sub); padding:1px 5px; border-radius:3px;">M${it.module_number}</span>`;
            if (it.item_code.startsWith('DM.')) {
                modBadge = `<span style="font-family:monospace; font-weight:800; font-size:10.5px; background:rgba(14,165,233,0.1); color:#0284c7; padding:1px 5px; border-radius:3px;">DM</span>`;
            }

            // Statut Gating
            let gatingBadge = `<span style="color:#10b981; font-size:10.5px; font-weight:700; display:flex; align-items:center; gap:4px;"><span style="font-size:12px;">✅</span> Déverrouillé</span>`;
            if (isBlocked) {
                gatingBadge = `<span style="color:#ef4444; font-size:10.5px; font-weight:700; display:flex; align-items:center; gap:4px;" title="${escapeHtml(it.gating_block_reason || 'Prérequis bloquant non scellé')}"><span style="font-size:12px;">🔒</span> Verrou GxP</span>`;
            } else if (isSelfMade) {
                gatingBadge = `<span style="color:#f59e0b; font-size:10.5px; font-weight:700; display:flex; align-items:center; gap:4px;" title="Ségrégation des fonctions BPF : Vous êtes l'auteur de cette pièce"><span style="font-size:12px;">⚠️</span> SoD Auteur</span>`;
            }

            // Échéance / Smart Planning
            let targetDateDisplay = '-';
            if (it.target_deposit_date) {
                const dStr = it.target_deposit_date.substring(0, 10);
                if (it.days_to_target_deposit !== undefined && it.days_to_target_deposit !== null) {
                    const days = it.days_to_target_deposit;
                    let badgeClass = 'background:rgba(16,185,129,0.1); color:#059669;';
                    if (days < 0) {
                        badgeClass = 'background:rgba(239,68,68,0.12); color:#dc2626; font-weight:800;';
                    } else if (days <= 30) {
                        badgeClass = 'background:rgba(245,158,11,0.15); color:#d97706; font-weight:800;';
                    }
                    targetDateDisplay = `
                        <div>
                            <span style="font-size:11px; font-weight:600;">${dStr}</span><br>
                            <span style="font-size:9.5px; padding:1px 5px; border-radius:3px; ${badgeClass}">
                                ${days < 0 ? `DÉPASSÉ (${Math.abs(days)}j)` : (days === 0 ? "AUJOURD'HUI" : `J-${days}`)}
                            </span>
                        </div>
                    `;
                } else {
                    targetDateDisplay = `<span style="font-size:11px;">${dStr}</span>`;
                }
            }

            // Checker notes
            let checkerDisplay = `<span style="color:var(--text-dim); font-size:10.5px;">-</span>`;
            if (it.checker_user_name) {
                checkerDisplay = `
                    <div style="font-size:11px;">
                        <span style="font-weight:700; color:var(--text-main);">🔍 ${escapeHtml(it.checker_user_name)}</span>
                        ${it.checker_reviewed_at ? `<div style="font-size:9.5px; color:var(--text-dim);">${new Date(it.checker_reviewed_at).toLocaleDateString()}</div>` : ''}
                        ${it.checker_notes ? `<div style="font-size:10px; color:#0284c7; max-width:200px; white-space:nowrap; overflow:hidden; text-overflow:ellipsis;" title="${escapeHtml(it.checker_notes)}">💬 "${escapeHtml(it.checker_notes)}"</div>` : ''}
                    </div>
                `;
            }

            return `
                <tr style="border-bottom:1px solid var(--border-color); ${isChecked ? 'background:rgba(147,51,234,0.04);' : ''} transition:background 0.15s ease;">
                    <td style="padding:8px 10px; text-align:center;">
                        <input type="checkbox" class="parapheur-row-checkbox" value="${it.item_id}" 
                            ${isChecked ? 'checked' : ''} 
                            ${!isSelectable ? 'disabled' : ''}
                            onchange="onParapheurItemCheckboxChange('${it.item_id}', this.checked)"
                            style="cursor:${isSelectable ? 'pointer' : 'not-allowed'}; transform:scale(1.15);">
                    </td>
                    <td style="padding:8px 10px;">
                        <div style="font-weight:800; font-family:monospace; color:var(--accent-blue); font-size:11.5px;">
                            ${escapeHtml(it.dossier_code)}
                        </div>
                        <div style="font-size:11px; color:var(--text-main); font-weight:600; max-width:220px; white-space:nowrap; overflow:hidden; text-overflow:ellipsis;" title="${escapeHtml(it.product_name)}">
                            ${escapeHtml(it.product_name)}
                        </div>
                    </td>
                    <td style="padding:8px 10px; white-space:nowrap;">
                        <div style="display:flex; align-items:center; gap:5px;">
                            ${modBadge}
                            <span style="font-family:monospace; font-weight:800; color:#7e22ce; font-size:11.5px;">${escapeHtml(it.item_code)}</span>
                        </div>
                    </td>
                    <td style="padding:8px 10px;">
                        <div style="font-weight:700; color:var(--text-main); font-size:12px; margin-bottom:2px;">
                            ${escapeHtml(it.title_fr)}
                        </div>
                        <div style="display:flex; align-items:center; gap:6px;">
                            <span style="font-size:9.5px; background:var(--bg-sub); color:var(--text-dim); padding:1px 5px; border-radius:3px; font-weight:600;">Rév. ${escapeHtml(it.current_revision || 'v0.1')}</span>
                            ${it.assigned_department ? `<span style="font-size:9.5px; background:rgba(37,99,235,0.08); color:var(--accent-blue); padding:1px 5px; border-radius:3px; font-weight:700;">${escapeHtml(it.assigned_department)}</span>` : ''}
                            ${it.requirement_level === 'E' ? '<span style="font-size:9.5px; background:#fee2e2; color:#b91c1c; padding:1px 4px; border-radius:3px; font-weight:800;">[E] Exigé</span>' : ''}
                        </div>
                    </td>
                    <td style="padding:8px 10px; font-size:11px;">
                        ${it.maker_user_name ? `<span style="color:var(--text-main); font-weight:600;">👤 ${escapeHtml(it.maker_user_name)}</span>` : '<span style="color:var(--text-dim);">-</span>'}
                    </td>
                    <td style="padding:8px 10px;">
                        ${checkerDisplay}
                    </td>
                    <td style="padding:8px 10px;">
                        ${targetDateDisplay}
                    </td>
                    <td style="padding:8px 10px; white-space:nowrap;">
                        ${gatingBadge}
                    </td>
                    <td style="padding:8px 10px; text-align:center; white-space:nowrap;">
                        <div style="display:flex; gap:4px; justify-content:center; align-items:center;">
                            <button class="btn-pill" style="font-size:10.5px; padding:2px 7px;" title="Consulter dans le dossier" onclick="viewParapheurItemDetail('${it.dossier_id}', '${it.item_id}', '${escapeHtml(it.item_code)}')">
                                👁️
                            </button>
                            ${isSelectable ? `
                                <button class="btn-pill" style="font-size:10.5px; padding:2px 8px; background:#7e22ce; color:white; border:none; font-weight:700;" title="Signer cette pièce individuellement" onclick="signSingleParapheurItem('${it.dossier_id}', '${it.item_id}', '${escapeHtml(it.item_code)}', '${escapeHtml(it.title_fr)}')">
                                    ✍️
                                </button>
                            ` : ''}
                        </div>
                    </td>
                </tr>
            `;
        }).join('');

        updateSelectAllState(items);
        updateBatchBar();
    }

    /**
     * Gestion du clic sur "Tout sélectionner"
     */
    window.toggleSelectAllParapheur = function (checked) {
        const rowCheckboxes = document.querySelectorAll('.parapheur-row-checkbox:not(:disabled)');
        rowCheckboxes.forEach(cb => {
            cb.checked = checked;
            const id = cb.value;
            if (checked) {
                selectedItemIds.add(id);
            } else {
                selectedItemIds.delete(id);
            }
        });
        updateBatchBar();
    };

    /**
     * Gestion du changement sur une checkbox de ligne
     */
    window.onParapheurItemCheckboxChange = function (itemId, checked) {
        if (checked) {
            selectedItemIds.add(itemId);
        } else {
            selectedItemIds.delete(itemId);
        }
        updateSelectAllState(parapheurItems);
        updateBatchBar();
    };

    function updateSelectAllState(items) {
        const selectAllCheckbox = document.getElementById('parapheur-select-all');
        if (!selectAllCheckbox) return;

        const selectableItems = items.filter(it => !it.is_gating_blocked && (!currentUser || it.maker_user_id !== currentUser.id));
        if (selectableItems.length === 0) {
            selectAllCheckbox.checked = false;
            return;
        }

        const allSelected = selectableItems.every(it => selectedItemIds.has(it.item_id));
        selectAllCheckbox.checked = allSelected;
    }

    /**
     * Met à jour la barre d'action par lot
     */
    function updateBatchBar() {
        const bar = document.getElementById('parapheur-batch-bar');
        const countSpan = document.getElementById('parapheur-selected-count');
        const submitBtn = document.getElementById('btn-parapheur-batch-submit');

        const count = selectedItemIds.size;
        if (countSpan) countSpan.textContent = count;

        if (bar) {
            if (count > 0) {
                bar.style.display = 'flex';
            } else {
                bar.style.display = 'none';
            }
        }
        if (submitBtn) {
            submitBtn.disabled = count === 0;
        }
    }

    /**
     * Ouvre la modale de confirmation de signature par lot (21 CFR Part 11)
     */
    window.openBatchConfirmModal = function () {
        if (selectedItemIds.size === 0) {
            if (typeof showToast === 'function') showToast("Veuillez sélectionner au moins une pièce à signer.", "warning");
            return;
        }

        const selectedList = parapheurItems.filter(it => selectedItemIds.has(it.item_id));
        const listContainer = document.getElementById('parapheur-batch-confirm-list');
        const totalCountEl = document.getElementById('parapheur-batch-confirm-count');

        if (totalCountEl) totalCountEl.textContent = selectedList.length;

        if (listContainer) {
            listContainer.innerHTML = selectedList.map(it => `
                <div style="display:flex; justify-content:space-between; align-items:center; padding:6px 8px; border-bottom:1px solid var(--border-color); font-size:11.5px;">
                    <div>
                        <span style="font-family:monospace; font-weight:800; color:#7e22ce;">[${escapeHtml(it.item_code)}]</span>
                        <span style="font-weight:700; color:var(--text-main); margin-left:4px;">${escapeHtml(it.title_fr)}</span>
                        <div style="font-size:10px; color:var(--text-dim); margin-top:1px;">
                            ${escapeHtml(it.dossier_code)} • ${escapeHtml(it.product_name)} • Rév. ${escapeHtml(it.current_revision)}
                        </div>
                    </div>
                    <div style="font-size:10px; font-weight:700; background:rgba(16,185,129,0.1); color:#059669; padding:2px 6px; border-radius:3px;">
                        Prêt pour Visa
                    </div>
                </div>
            `).join('');
        }

        // Réinitialiser les champs
        const solemnCheck = document.getElementById('parapheur-batch-solemn-check');
        const totpInput = document.getElementById('parapheur-batch-totp');
        const errBox = document.getElementById('parapheur-batch-err');

        if (solemnCheck) solemnCheck.checked = false;
        if (totpInput) totpInput.value = '';
        if (errBox) errBox.style.display = 'none';

        if (typeof openModal === 'function') {
            openModal('modal-parapheur-batch-confirm');
        } else {
            const m = document.getElementById('modal-parapheur-batch-confirm');
            if (m) m.classList.add('active');
        }
    };

    /**
     * Ferme la modale de confirmation batch
     */
    window.closeBatchConfirmModal = function () {
        if (typeof closeModal === 'function') {
            closeModal('modal-parapheur-batch-confirm');
        } else {
            const m = document.getElementById('modal-parapheur-batch-confirm');
            if (m) m.classList.remove('active');
        }
    };

    /**
     * Soumission transactionnelle de la signature par lot (21 CFR Part 11)
     */
    window.submitBatchApproval = async function (e) {
        if (e) e.preventDefault();

        const solemnCheck = document.getElementById('parapheur-batch-solemn-check');
        const totpInput = document.getElementById('parapheur-batch-totp');
        const errBox = document.getElementById('parapheur-batch-err');
        const submitBtn = document.getElementById('parapheur-batch-confirm-btn');

        if (!solemnCheck || !solemnCheck.checked) {
            if (errBox) {
                errBox.textContent = "Vous devez accepter formellement la déclaration solennelle de responsabilité pharmaceutique légale.";
                errBox.style.display = 'block';
            }
            return;
        }

        const code = totpInput ? totpInput.value.trim() : '';

        if (submitBtn) {
            submitBtn.disabled = true;
            submitBtn.textContent = "Scellement cryptographique 21 CFR Part 11 en cours...";
        }

        try {
            const payload = {
                item_ids: Array.from(selectedItemIds),
                totp_code: code,
                password: code,
                solemn_declaration_accepted: true
            };

            const resp = await fetch('/api/v1/parapheur/batch-approve', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(payload)
            });

            const data = await resp.json();

            if (!resp.ok) {
                if (errBox) {
                    errBox.textContent = data.message || data.error || "Erreur lors de la signature par lot";
                    errBox.style.display = 'block';
                }
                if (submitBtn) {
                    submitBtn.disabled = false;
                    submitBtn.textContent = "✍️ Signer & Sceller la Liasse (21 CFR Part 11)";
                }
                return;
            }

            closeBatchConfirmModal();

            if (typeof showToast === 'function') {
                showToast(`✅ Signature par lot réussie : ${data.total_approved} pièce(s) scellée(s) v1.0 avec empreinte SHA-256`, "success", 5000);
            }

            // Rafraîchir le parapheur
            loadParapheurItems();

            // Rafraîchir le dossier en arrière-plan si ouvert
            if (typeof currentDossierId !== 'undefined' && currentDossierId && typeof loadDossierDetail === 'function') {
                loadDossierDetail(currentDossierId);
            }

        } catch (err) {
            console.error('Erreur Batch Approve:', err);
            if (errBox) {
                errBox.textContent = "Erreur de communication : " + err.message;
                errBox.style.display = 'block';
            }
        } finally {
            if (submitBtn) {
                submitBtn.disabled = false;
                submitBtn.textContent = "✍️ Signer & Sceller la Liasse (21 CFR Part 11)";
            }
        }
    };

    /**
     * Ouvre la vue détaillée d'un item dans son dossier
     */
    window.viewParapheurItemDetail = function (dossierId, itemId, itemCode) {
        closeParapheurModal();
        if (typeof switchDomainModule === 'function') {
            switchDomainModule('ctd');
        }
        if (typeof openDossierDetail === 'function') {
            openDossierDetail(dossierId, itemId);
        }
    };

    /**
     * Déclenche la signature unitaire via la modale existante
     */
    window.signSingleParapheurItem = function (dossierId, itemId, itemCode, titleFr) {
        if (typeof openPdtSealModal === 'function') {
            openPdtSealModal(dossierId, itemId, itemCode);
            // Fermer le parapheur pour donner le focus à la modale de visa
            closeParapheurModal();
        }
    };

    function escapeHtml(str) {
        if (!str) return '';
        return String(str)
            .replace(/&/g, '&amp;')
            .replace(/</g, '&lt;')
            .replace(/>/g, '&gt;')
            .replace(/"/g, '&quot;')
            .replace(/'/g, '&#039;');
    }
})();
