package reporting

import (
	"fmt"
	"strings"

	"homolog/internal/domain"
)

// GenerateDDMHomologationDOCX pré-remplit le formulaire officiel ANPP « FORMULAIRE DE DEMANDE D’HOMOLOGATION.docx ».
func GenerateDDMHomologationDOCX(org *domain.Organization, dm *domain.MedicalDevice) ([]byte, error) {
	tpl, err := GetANPPTemplate("ddm_demande_homologation.docx")
	if err != nil {
		return nil, err
	}

	orgName := ""
	miphNum := ""
	pdtName := ""
	if org != nil {
		orgName = org.LegalName
		miphNum = org.MIPHLicenseNumber
		pdtName = org.TechnicalDirectorName
	}

	labels := map[string]string{
		"dénomination commerciale":               safeDMField(dm, func(d *domain.MedicalDevice) string { return d.CommercialNameFR }),
		"désignation":                            safeDMField(dm, func(d *domain.MedicalDevice) string { return d.IntendedUseFR }),
		"classification du risque":               safeDMField(dm, func(d *domain.MedicalDevice) string { return "Classe " + d.DMClass }),
		"etablissement pharmaceutique demandeur": orgName,
		"numéro et date d'agrément":              miphNum,
		"pays d'origine":                         safeDMField(dm, func(d *domain.MedicalDevice) string { return d.LegalManufacturerCountry }),
		"titulaire de l'amm":                     safeDMField(dm, func(d *domain.MedicalDevice) string { return d.LegalManufacturerName }),
		"nomenclature/codification du dm":        safeDMField(dm, func(d *domain.MedicalDevice) string { return d.GMDNCode }),
	}

	textRepl := map[string]string{}
	if dm != nil {
		switch strings.ToUpper(strings.TrimSpace(dm.DMClass)) {
		case "I":
			textRepl["□ Classe I"] = "☑ Classe I"
		case "IIA":
			textRepl["□ Classe IIa"] = "☑ Classe IIa"
		case "IIB":
			textRepl["□ Classe IIb"] = "☑ Classe IIb"
		case "III":
			textRepl["□ Classe III"] = "☑ Classe III"
		}
		if dm.IsSterile {
			textRepl["□ OUI"] = "☑ OUI"
		} else {
			textRepl["□ NON"] = "☑ NON"
		}
	}
	if pdtName != "" {
		textRepl["Nom, Prénom, Pharmacien Directeur Technique"] = pdtName + ", Pharmacien Directeur Technique"
	}

	return FillDOCXTemplate(tpl, DOCXFillOptions{
		LabelValues:      labels,
		TextReplacements: textRepl,
	})
}

// GenerateDDMRecepisseDOCX pré-remplit le gabarit officiel « RECEPISSE DE DEPOT DES DOSSIERS SOUMIS A L’HOMOLOGATION.docx ».
func GenerateDDMRecepisseDOCX(org *domain.Organization, dm *domain.MedicalDevice, eval *domain.DMRecevabiliteEvaluation) ([]byte, error) {
	tpl, err := GetANPPTemplate("ddm_recepisse_depot.docx")
	if err != nil {
		return nil, err
	}

	orgName, addr, phone, email := "", "", "", ""
	if org != nil {
		orgName = org.LegalName
		addr = org.Address
		phone = org.Phone
		email = org.Email
	}

	labels := map[string]string{
		"etablissement pharmaceutique demandeur": orgName,
		"adresse":                                addr,
		"tél/fax :":                              phone,
		"courriel":                               email,
		"dénomination commerciale/nom de marque": safeDMField(dm, func(d *domain.MedicalDevice) string { return d.CommercialNameFR }),
		"désignation/gamme":                      safeDMField(dm, func(d *domain.MedicalDevice) string { return d.IntendedUseFR }),
		"classe":                                 safeDMField(dm, func(d *domain.MedicalDevice) string { return "Classe " + d.DMClass }),
		"caractéristique/conditionnement":        safeDMField(dm, func(d *domain.MedicalDevice) string { return d.ManufacturerReference }),
	}

	var cellUpdates []TableCellCoord
	if eval != nil {
		statuses := []string{
			eval.Item1Status, eval.Item2Status, eval.Item3Status, eval.Item4Status,
			eval.Item5Status, eval.Item6Status, eval.Item7Status, eval.Item8Status,
			eval.Item9Status, eval.Item10Status, eval.Item11Status, eval.Item12Status,
			eval.Item13Status,
		}
		// Table 5, lignes 2 à 14 : col 2=Oui, col 3=Non, col 4=NA
		for idx, st := range statuses {
			rowIdx := 2 + idx
			switch strings.ToLower(strings.TrimSpace(st)) {
			case "oui", "conforme", "valide":
				cellUpdates = append(cellUpdates, TableCellCoord{TableIndex: 5, RowIndex: rowIdx, ColIndex: 2, Value: "☑"})
			case "non", "non_conforme":
				cellUpdates = append(cellUpdates, TableCellCoord{TableIndex: 5, RowIndex: rowIdx, ColIndex: 3, Value: "☑"})
			case "na", "non_applicable":
				cellUpdates = append(cellUpdates, TableCellCoord{TableIndex: 5, RowIndex: rowIdx, ColIndex: 4, Value: "☑"})
			}
		}
		if eval.DepositOrderNumber != "" {
			cellUpdates = append(cellUpdates, TableCellCoord{TableIndex: 4, RowIndex: 1, ColIndex: 1, Value: eval.ANPPAgentName})
			if org != nil {
				cellUpdates = append(cellUpdates, TableCellCoord{TableIndex: 4, RowIndex: 1, ColIndex: 2, Value: org.TechnicalDirectorName})
			}
		}
	}

	return FillDOCXTemplate(tpl, DOCXFillOptions{
		LabelValues: labels,
		CellUpdates: cellUpdates,
	})
}

// GenerateDMCanevasNonHomologuesXLSX pré-remplit le fichier Excel officiel « CANEVAS DECLARATION DM NON HOMOLOGUES.xlsx ».
func GenerateDMCanevasNonHomologuesXLSX(org *domain.Organization, regs []domain.DMRegularization) ([]byte, error) {
	tpl, err := GetANPPTemplate("ddm_canevas_non_homologues.xlsx")
	if err != nil {
		return nil, err
	}

	var updates []XLSXCellUpdate
	if org != nil {
		updates = append(updates,
			XLSXCellUpdate{CellRef: "C4", Value: org.LegalName},
			XLSXCellUpdate{CellRef: "D4", Value: org.Address},
			XLSXCellUpdate{CellRef: "E4", Value: org.MIPHLicenseNumber},
		)
	}

	for i, reg := range regs {
		rowNum := 11 + i
		brand := ""
		intended := ""
		form := ""
		if reg.MedicalDevice != nil {
			brand = reg.MedicalDevice.CommercialNameFR
			intended = reg.MedicalDevice.IntendedUseFR
			form = "Classe " + reg.MedicalDevice.DMClass
		}
		updates = append(updates,
			XLSXCellUpdate{CellRef: fmt.Sprintf("A%d", rowNum), Value: reg.LastDeclarationDate.Format("2006-01-02")},
			XLSXCellUpdate{CellRef: fmt.Sprintf("B%d", rowNum), Value: brand},
			XLSXCellUpdate{CellRef: fmt.Sprintf("C%d", rowNum), Value: intended},
			XLSXCellUpdate{CellRef: fmt.Sprintf("D%d", rowNum), Value: form},
			XLSXCellUpdate{CellRef: fmt.Sprintf("F%d", rowNum), Value: reg.Status},
			XLSXCellUpdate{CellRef: fmt.Sprintf("G%d", rowNum), Value: reg.DDMAcknowledgementNumber},
			XLSXCellUpdate{CellRef: fmt.Sprintf("H%d", rowNum), Value: reg.RegularizationCode},
			XLSXCellUpdate{CellRef: fmt.Sprintf("K%d", rowNum), Value: fmt.Sprintf("%.0f", reg.QuarterlyVolumeImported), IsNumber: true},
		)
	}

	return FillXLSXTemplate(tpl, updates)
}

// GenerateDDMBordereauDOCX pré-remplit les bordereaux officiels DM (modification, renouvellement, transfert).
func GenerateDDMBordereauDOCX(kind string, org *domain.Organization, dm *domain.MedicalDevice, slipRef, desc string) ([]byte, error) {
	tplName := "ddm_bv_modification.docx"
	switch strings.ToLower(kind) {
	case "renouvellement":
		tplName = "ddm_bv_renouvellement.docx"
	case "transfert":
		tplName = "ddm_bv_transfert.docx"
	}
	tpl, err := GetANPPTemplate(tplName)
	if err != nil {
		return nil, err
	}

	orgName, pdtName, phone, email := "", "", "", ""
	if org != nil {
		orgName = org.LegalName
		pdtName = org.TechnicalDirectorName
		phone = org.Phone
		email = org.Email
	}
	dmName := safeDMField(dm, func(d *domain.MedicalDevice) string { return d.CommercialNameFR })
	dmUse := safeDMField(dm, func(d *domain.MedicalDevice) string { return d.IntendedUseFR })
	dmClass := safeDMField(dm, func(d *domain.MedicalDevice) string { return "Classe " + d.DMClass })

	textRepl := map[string]string{
		"N°.…":                              "N° " + slipRef,
		"Dénomination commercial :":         "Dénomination commerciale : " + dmName,
		"Dénomination commercial\u00a0:":    "Dénomination commerciale : " + dmName,
		"Désignation :":                     "Désignation : " + dmUse,
		"Désignation\u00a0:":                "Désignation : " + dmUse,
		"Classification :":                  "Classification : " + dmClass,
		"Classification\u00a0:":             "Classification : " + dmClass,
		"Descriptif de modification :":      "Descriptif de modification : " + desc,
		"Descriptif de modification\u00a0:": "Descriptif de modification : " + desc,
		"Tél. Fixe :":                       "Tél. Fixe : " + phone,
		"Tél. Fixe\u00a0:":                  "Tél. Fixe : " + phone,
		"E-mail :":                          "E-mail : " + email,
		"E-mail\u00a0:":                     "E-mail : " + email,
	}
	if orgName != "" || pdtName != "" {
		textRepl["Désignation du redevable de la redevance (Demandeur):"] = fmt.Sprintf("Désignation du redevable de la redevance (Demandeur): %s — PDT: %s", orgName, pdtName)
	}

	return FillDOCXTemplate(tpl, DOCXFillOptions{
		TextReplacements: textRepl,
	})
}

// GenerateDEPPFormDOCX pré-remplit les formulaires officiels DEPP02-A/B/C/D/E/F-V02 pour un dossier CTD.
func GenerateDEPPFormDOCX(kind string, dossier *domain.Dossier, org *domain.Organization) ([]byte, error) {
	tplMap := map[string]string{
		"enregistrement": "depp02_a_enregistrement.docx",
		"modification":   "depp02_b_modification.docx",
		"renouvellement": "depp02_c_renouvellement.docx",
		"transfert":      "depp02_d_transfert.docx",
		"recepisse":      "depp02_e_recepisse.docx",
		"echantillons":   "depp02_f_echantillons.docx",
	}
	tplName, ok := tplMap[strings.ToLower(strings.TrimSpace(kind))]
	if !ok {
		return nil, fmt.Errorf("type de formulaire DEPP inconnu: %s", kind)
	}
	tpl, err := GetANPPTemplate(tplName)
	if err != nil {
		return nil, err
	}

	brand, dci, dosage, form, _, _ := extractProductDetails(dossier)
	orgName := ""
	if org != nil {
		orgName = org.LegalName
	}

	labels := map[string]string{
		"dénomination commerciale":            brand,
		"dénomination commune internationale": dci,
		"commune internationale":              dci,
		"forme pharmaceutique":                form,
		"dosage":                              dosage,
		"nom/raison sociale":                  orgName,
		"dénomination commerciale /dci":       strings.TrimSpace(brand + " / " + dci),
		"forme/dosage/présentation":           strings.TrimSpace(form + " " + dosage),
	}

	textRepl := map[string]string{}
	if dossier != nil {
		proc := strings.ToUpper(string(dossier.ProcedureCode))
		switch proc {
		case "NF", "RF", "MF":
			textRepl["Fabrication locale ☐"] = "Fabrication locale ☑"
			textRepl["Fabrication☐"] = "Fabrication☑"
		case "NI", "RI", "MI":
			textRepl["Importation ☐"] = "Importation ☑"
			textRepl["Importation\u00a0  ☐"] = "Importation ☑"
			textRepl["Importation☐"] = "Importation☑"
		}
		if proc == "NI" || proc == "NF" {
			textRepl["Nouvel Enregistrement☐"] = "Nouvel Enregistrement☑"
		} else if proc == "RI" || proc == "RF" {
			textRepl["Renouvellement☐"] = "Renouvellement☑"
		} else if proc == "MI" || proc == "MF" {
			textRepl["Modification☐"] = "Modification☑"
		}
		if dossier.DossierCode != "" {
			textRepl["Code dossier : …………………………………………………………………"] = "Code dossier : " + dossier.DossierCode
			textRepl["Code dossier\u00a0: …………………………………………………………………"] = "Code dossier : " + dossier.DossierCode
		}
	}

	return FillDOCXTemplate(tpl, DOCXFillOptions{
		LabelValues:      labels,
		TextReplacements: textRepl,
	})
}

func safeDMField(dm *domain.MedicalDevice, fn func(*domain.MedicalDevice) string) string {
	if dm == nil {
		return ""
	}
	return fn(dm)
}

func formatLotDates(lot *domain.ImportLotRelease) string {
	if lot == nil {
		return ""
	}
	parts := []string{}
	if lot.LotNumber != "" {
		parts = append(parts, "Lot: "+lot.LotNumber)
	}
	if lot.SerialNumber != "" {
		parts = append(parts, "Série: "+lot.SerialNumber)
	}
	if lot.DDF != nil {
		parts = append(parts, "DDF: "+lot.DDF.Format("2006-01-02"))
	}
	if lot.DDP != nil {
		parts = append(parts, "DDP: "+lot.DDP.Format("2006-01-02"))
	}
	return strings.Join(parts, " / ")
}

func checkYesNo(v bool) string {
	if v {
		return "Oui ☑                                  Non ☐"
	}
	return "Oui ☐                                  Non ☑"
}

// GenerateDCEDMLotReceptionDOCX pré-remplit le gabarit officiel « Formulaire de réception des DM importés soumis au contrôle de libération de lot ».
func GenerateDCEDMLotReceptionDOCX(org *domain.Organization, lot *domain.ImportLotRelease) ([]byte, error) {
	tpl, err := GetANPPTemplate("dce_reception_dm_importes.docx")
	if err != nil {
		return nil, err
	}

	orgAddr, email, phone, pdtName := "", "", "", ""
	if org != nil {
		orgAddr = strings.TrimSpace(org.LegalName + " — " + org.Address)
		email = org.Email
		phone = org.Phone
		pdtName = org.TechnicalDirectorName
	}

	var cellUpdates []TableCellCoord
	cellUpdates = append(cellUpdates,
		TableCellCoord{TableIndex: 0, RowIndex: 0, ColIndex: 1, Value: orgAddr},
		TableCellCoord{TableIndex: 0, RowIndex: 1, ColIndex: 1, Value: fmt.Sprintf("E-mail : %s   N° tél : %s", email, phone)},
	)
	if lot != nil {
		customsInv := strings.TrimSpace(strings.Join([]string{lot.CustomsDeclNumber, lot.InvoiceNumber}, " / "))
		customsInv = strings.Trim(customsInv, " /")
		cellUpdates = append(cellUpdates,
			TableCellCoord{TableIndex: 1, RowIndex: 1, ColIndex: 1, Value: lot.CommercialName},
			TableCellCoord{TableIndex: 1, RowIndex: 2, ColIndex: 1, Value: lot.Designation},
			TableCellCoord{TableIndex: 1, RowIndex: 3, ColIndex: 1, Value: lot.FormAspect},
			TableCellCoord{TableIndex: 1, RowIndex: 4, ColIndex: 1, Value: formatLotDates(lot)},
			TableCellCoord{TableIndex: 1, RowIndex: 5, ColIndex: 1, Value: lot.SalesUnits},
			TableCellCoord{TableIndex: 1, RowIndex: 6, ColIndex: 1, Value: lot.HomologationOrDERef},
			TableCellCoord{TableIndex: 1, RowIndex: 8, ColIndex: 1, Value: lot.HolderOrExploiter},
			TableCellCoord{TableIndex: 1, RowIndex: 9, ColIndex: 1, Value: checkYesNo(lot.ImportProgramAttached)},
			TableCellCoord{TableIndex: 1, RowIndex: 10, ColIndex: 1, Value: lot.ProductionSite},
			TableCellCoord{TableIndex: 1, RowIndex: 11, ColIndex: 1, Value: lot.QCSite},
			TableCellCoord{TableIndex: 1, RowIndex: 12, ColIndex: 1, Value: lot.ReleaseSite},
			TableCellCoord{TableIndex: 1, RowIndex: 13, ColIndex: 1, Value: customsInv},
			TableCellCoord{TableIndex: 1, RowIndex: 14, ColIndex: 1, Value: checkYesNo(lot.OriginCertAttached)},
			TableCellCoord{TableIndex: 1, RowIndex: 15, ColIndex: 1, Value: checkYesNo(lot.TechFormAttached)},
			TableCellCoord{TableIndex: 1, RowIndex: 16, ColIndex: 1, Value: checkYesNo(lot.SamplingPlanAttached)},
		)
	}
	if pdtName != "" {
		cellUpdates = append(cellUpdates, TableCellCoord{TableIndex: 2, RowIndex: 1, ColIndex: 0, Value: pdtName})
	}

	textRepl := map[string]string{}
	if lot != nil {
		switch strings.ToUpper(strings.TrimSpace(lot.DMClass)) {
		case "I":
			textRepl["I  ☐"] = "I  ☑"
		case "IIA":
			textRepl["IIa  ☐"] = "IIa  ☑"
		case "IIB":
			textRepl["IIb  ☐"] = "IIb  ☑"
		case "III":
			textRepl["III  ☐"] = "III  ☑"
		}
	}

	return FillDOCXTemplate(tpl, DOCXFillOptions{
		CellUpdates:      cellUpdates,
		TextReplacements: textRepl,
	})
}

// GenerateDCEPPLotReceptionDOCX pré-remplit le gabarit officiel « Formulaire de réception des PP importés soumis au contrôle de libération de lot ».
func GenerateDCEPPLotReceptionDOCX(org *domain.Organization, lot *domain.ImportLotRelease) ([]byte, error) {
	tpl, err := GetANPPTemplate("dce_reception_pp_importes.docx")
	if err != nil {
		return nil, err
	}

	orgAddr, email, phone, pdtName := "", "", "", ""
	if org != nil {
		orgAddr = strings.TrimSpace(org.LegalName + " — " + org.Address)
		email = org.Email
		phone = org.Phone
		pdtName = org.TechnicalDirectorName
	}

	var cellUpdates []TableCellCoord
	cellUpdates = append(cellUpdates,
		TableCellCoord{TableIndex: 0, RowIndex: 0, ColIndex: 1, Value: orgAddr},
		TableCellCoord{TableIndex: 0, RowIndex: 1, ColIndex: 1, Value: fmt.Sprintf("E-mail : %s   N° tél : %s", email, phone)},
	)
	if lot != nil {
		customsInv := strings.Trim(strings.Join([]string{lot.CustomsDeclNumber, lot.InvoiceNumber}, " / "), " /")
		cellUpdates = append(cellUpdates,
			TableCellCoord{TableIndex: 1, RowIndex: 1, ColIndex: 1, Value: lot.CommercialName},
			TableCellCoord{TableIndex: 1, RowIndex: 2, ColIndex: 1, Value: lot.Designation},
			TableCellCoord{TableIndex: 1, RowIndex: 3, ColIndex: 1, Value: lot.FormAspect},
			TableCellCoord{TableIndex: 1, RowIndex: 4, ColIndex: 1, Value: formatLotDates(lot)},
			TableCellCoord{TableIndex: 1, RowIndex: 5, ColIndex: 1, Value: lot.SalesUnits},
			TableCellCoord{TableIndex: 1, RowIndex: 6, ColIndex: 1, Value: lot.HomologationOrDERef},
			TableCellCoord{TableIndex: 1, RowIndex: 7, ColIndex: 1, Value: lot.HolderOrExploiter},
			TableCellCoord{TableIndex: 1, RowIndex: 8, ColIndex: 1, Value: checkYesNo(lot.ImportProgramAttached)},
			TableCellCoord{TableIndex: 1, RowIndex: 9, ColIndex: 1, Value: lot.ProductionSite},
			TableCellCoord{TableIndex: 1, RowIndex: 10, ColIndex: 1, Value: lot.QCSite},
			TableCellCoord{TableIndex: 1, RowIndex: 11, ColIndex: 1, Value: lot.ReleaseSite},
			TableCellCoord{TableIndex: 1, RowIndex: 12, ColIndex: 1, Value: customsInv},
			TableCellCoord{TableIndex: 1, RowIndex: 15, ColIndex: 1, Value: checkYesNo(lot.TechFormAttached)},
			TableCellCoord{TableIndex: 1, RowIndex: 17, ColIndex: 1, Value: checkYesNo(lot.SamplingPlanAttached)},
		)
	}
	if pdtName != "" {
		cellUpdates = append(cellUpdates, TableCellCoord{TableIndex: 2, RowIndex: 1, ColIndex: 0, Value: pdtName})
	}

	return FillDOCXTemplate(tpl, DOCXFillOptions{CellUpdates: cellUpdates})
}

// GenerateDCECertificatLiberationPPDOCX pré-remplit le « certificat de libération de lot du produit pharmaceutique.docx ».
func GenerateDCECertificatLiberationPPDOCX(lot *domain.ImportLotRelease) ([]byte, error) {
	tpl, err := GetANPPTemplate("dce_certificat_liberation_pp.docx")
	if err != nil {
		return nil, err
	}
	if lot == nil {
		return tpl, nil
	}
	ddfStr, ddpStr := "", ""
	if lot.DDF != nil {
		ddfStr = lot.DDF.Format("2006-01-02")
	}
	if lot.DDP != nil {
		ddpStr = lot.DDP.Format("2006-01-02")
	}
	cellUpdates := []TableCellCoord{
		{TableIndex: 0, RowIndex: 1, ColIndex: 0, Value: lot.CommercialName},
		{TableIndex: 0, RowIndex: 1, ColIndex: 1, Value: lot.Designation},
		{TableIndex: 0, RowIndex: 1, ColIndex: 2, Value: lot.FormAspect},
		{TableIndex: 0, RowIndex: 1, ColIndex: 4, Value: lot.SalesUnits},
		{TableIndex: 1, RowIndex: 1, ColIndex: 0, Value: lot.LotNumber},
		{TableIndex: 1, RowIndex: 1, ColIndex: 1, Value: ddfStr},
		{TableIndex: 1, RowIndex: 1, ColIndex: 2, Value: ddpStr},
		{TableIndex: 1, RowIndex: 1, ColIndex: 3, Value: lot.HomologationOrDERef},
		{TableIndex: 1, RowIndex: 1, ColIndex: 4, Value: lot.SalesUnits},
	}
	return FillDOCXTemplate(tpl, DOCXFillOptions{CellUpdates: cellUpdates})
}

// GenerateDVTIVBasePDTQPPVDOCX pré-remplit le formulaire DVTIV « FORMULAIRE-DRAFT- Base de données centralisées des PDT et QPPV.docx ».
func GenerateDVTIVBasePDTQPPVDOCX(org *domain.Organization, pdt *domain.User, assistant *domain.User, qppv *domain.User) ([]byte, error) {
	tpl, err := GetANPPTemplate("dvtiv_base_pdt_qppv.docx")
	if err != nil {
		return nil, err
	}

	pdtName, pdtEmail, pdtPhone, pdtOrdre := "", "", "", ""
	if pdt != nil {
		pdtName = strings.TrimSpace(pdt.FirstName + " " + pdt.LastName)
		pdtEmail = pdt.Email
	}
	if org != nil {
		if pdtName == "" {
			pdtName = org.TechnicalDirectorName
		}
		if pdtEmail == "" {
			pdtEmail = org.Email
		}
		pdtPhone = org.Phone
		pdtOrdre = org.MIPHLicenseNumber
	}
	asstName := ""
	if assistant != nil {
		asstName = strings.TrimSpace(assistant.FirstName + " " + assistant.LastName)
	}
	qppvName, qppvEmail := "", ""
	if qppv != nil {
		qppvName = strings.TrimSpace(qppv.FirstName + " " + qppv.LastName)
		qppvEmail = qppv.Email
	}

	cellUpdates := []TableCellCoord{
		{TableIndex: 0, RowIndex: 1, ColIndex: 1, Value: pdtName},
		{TableIndex: 0, RowIndex: 2, ColIndex: 1, Value: pdtOrdre},
		{TableIndex: 0, RowIndex: 3, ColIndex: 1, Value: pdtEmail},
		{TableIndex: 0, RowIndex: 4, ColIndex: 1, Value: pdtPhone},
		{TableIndex: 0, RowIndex: 5, ColIndex: 1, Value: asstName},
		{TableIndex: 1, RowIndex: 1, ColIndex: 1, Value: qppvName},
		{TableIndex: 1, RowIndex: 2, ColIndex: 1, Value: qppvEmail},
	}
	return FillDOCXTemplate(tpl, DOCXFillOptions{CellUpdates: cellUpdates})
}
