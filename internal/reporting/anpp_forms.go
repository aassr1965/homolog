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

// GenerateDDMBordereauDOCX pré-remplit les bordereaux officiels DM (homologation, modification, renouvellement, transfert).
func GenerateDDMBordereauDOCX(kind string, org *domain.Organization, dm *domain.MedicalDevice, slipRef, desc string) ([]byte, error) {
	tplName := "ddm_bv_modification.docx"
	switch strings.ToLower(kind) {
	case "homologation":
		tplName = "ddm_bv_homologation.docx"
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
		"Dénomination commerciale :":        "Dénomination commerciale : " + dmName,
		"Dénomination commerciale\u00a0:":   "Dénomination commerciale : " + dmName,
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

// GenerateDEPPFormDOCX pré-remplit les formulaires officiels DEPP (.docx) pour un dossier CTD.
func GenerateDEPPFormDOCX(kind string, dossier *domain.Dossier, org *domain.Organization) ([]byte, error) {
	k := strings.ToLower(strings.TrimSpace(kind))
	tplMap := map[string]string{
		"enregistrement": "depp02_a_enregistrement.docx",
		"modification":   "depp02_b_modification.docx",
		"renouvellement": "depp02_c_renouvellement.docx",
		"transfert":      "depp02_d_transfert.docx",
		"recepisse":      "depp02_e_recepisse.docx",
		"echantillons":   "depp02_f_echantillons.docx",
		"pre-soumission": "depp_pre_soumission.docx",
		"structure-prix": "depp_modele_structure.docx",
		"recepisse-pp":   "depp_recepisse_depot_pp.docx",
	}
	tplName, ok := tplMap[k]
	if !ok {
		return nil, fmt.Errorf("type de formulaire DEPP inconnu: %s", kind)
	}
	tpl, err := GetANPPTemplate(tplName)
	if err != nil {
		return nil, err
	}

	brand, dci, dosage, form, _, _ := extractProductDetails(dossier)
	orgName, orgAddr, orgLicense, pdtName := "", "", "", ""
	if org != nil {
		orgName = org.LegalName
		orgAddr = org.Address
		orgLicense = org.MIPHLicenseNumber
		pdtName = org.TechnicalDirectorName
	}

	labels := map[string]string{
		"dénomination commerciale":                        brand,
		"dénomination commune internationale":             dci,
		"commune internationale":                          dci,
		"forme pharmaceutique":                            form,
		"dosage":                                          dosage,
		"nom/raison sociale":                              orgName,
		"établissement pharmaceutique demandeur":          orgName,
		"nom et prénom du pharmacien directeur technique": pdtName,
		"dénomination commerciale /dci":                   strings.TrimSpace(brand + " / " + dci),
		"forme/dosage/présentation":                       strings.TrimSpace(form + " " + dosage),
	}

	textRepl := map[string]string{}
	var strikeTerms []string
	var cellUpdates []TableCellCoord

	if k == "pre-soumission" {
		if brand != "" {
			textRepl["Dénomination commerciale\u00a0:"] = "Dénomination commerciale : " + brand
		}
		if dci != "" {
			textRepl["Dénomination commune internationale (DCI)\u00a0:"] = "Dénomination commune internationale (DCI) : " + dci
		}
		if form != "" {
			textRepl["Forme pharmaceutique :"] = "Forme pharmaceutique : " + form
		}
		if dosage != "" {
			textRepl["Dosage\u00a0:"] = "Dosage : " + dosage
		}
		if orgName != "" || orgAddr != "" {
			textRepl["Nom et adresse de l’établissement pharmaceutique demandeur:"] = "Nom et adresse de l’établissement pharmaceutique demandeur : " + strings.TrimSpace(orgName+" — "+orgAddr)
		}
		if orgLicense != "" {
			textRepl["Numéro et date de l’agrément de l’établissement pharmaceutique demandeur\u00a0:"] = "Numéro et date de l’agrément de l’établissement pharmaceutique demandeur : " + orgLicense
		}
		if pdtName != "" {
			textRepl["Nom et prénom du pharmacien directeur technique\u00a0:"] = "Nom et prénom du pharmacien directeur technique : " + pdtName
		}
	}

	if k == "recepisse-pp" {
		cellUpdates = []TableCellCoord{
			{TableIndex: 3, RowIndex: 1, ColIndex: 1, Value: brand},
			{TableIndex: 3, RowIndex: 2, ColIndex: 1, Value: dci},
			{TableIndex: 3, RowIndex: 3, ColIndex: 1, Value: form},
			{TableIndex: 3, RowIndex: 4, ColIndex: 1, Value: dosage},
			{TableIndex: 3, RowIndex: 6, ColIndex: 1, Value: orgName},
			{TableIndex: 3, RowIndex: 7, ColIndex: 1, Value: pdtName},
		}
	}

	if dossier != nil {
		proc := strings.ToUpper(string(dossier.ProcedureCode))
		switch proc {
		case "NF", "RF", "MF":
			textRepl["Fabrication locale ☐"] = "Fabrication locale ☑"
			textRepl["Fabrication☐"] = "Fabrication☑"
			if k == "recepisse-pp" {
				strikeTerms = append(strikeTerms, "Importation")
			}
		case "NI", "RI", "MI":
			textRepl["Importation ☐"] = "Importation ☑"
			textRepl["Importation\u00a0  ☐"] = "Importation ☑"
			textRepl["Importation☐"] = "Importation☑"
			if k == "recepisse-pp" {
				strikeTerms = append(strikeTerms, "Fabrication locale")
			}
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
		StrikeTerms:      strikeTerms,
		CellUpdates:      cellUpdates,
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
func GenerateDVTIVBasePDTQPPVDOCX(org *domain.Organization, pdt *domain.User, assistant *domain.User, qppv *domain.User, marketedCount int, hasPSMF, hasPSMFNational bool, estTypes ...string) ([]byte, error) {
	return GenerateDVTIVBasePDTQPPVDOCXWithDraft(org, pdt, assistant, qppv, nil, marketedCount, hasPSMF, hasPSMFNational, estTypes...)
}

// GenerateDVTIVBasePDTQPPVDOCXWithDraft pré-remplit le formulaire DVTIV en fusionnant un éventuel brouillon QPPV/PSMF.
func GenerateDVTIVBasePDTQPPVDOCXWithDraft(org *domain.Organization, pdt *domain.User, assistant *domain.User, qppv *domain.User, qppvDraft *domain.DVTIVQPPVDraft, marketedCount int, hasPSMF, hasPSMFNational bool, estTypes ...string) ([]byte, error) {
	tpl, err := GetANPPTemplate("dvtiv_base_pdt_qppv.docx")
	if err != nil {
		return nil, err
	}

	orgName, orgAddr, pdtName, pdtEmail, pdtPhone, pdtOrdre := "", "", "", "", "", ""
	if pdt != nil {
		pdtName = strings.TrimSpace(pdt.FirstName + " " + pdt.LastName)
		pdtEmail = pdt.Email
	}
	if org != nil {
		orgName = org.LegalName
		orgAddr = org.Address
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
	qppvName, qppvEmail, qppvPhone, qppvDeputy := "", "", "", ""
	if qppv != nil {
		qppvName = strings.TrimSpace(qppv.FirstName + " " + qppv.LastName)
		qppvEmail = qppv.Email
	}
	if qppvDraft != nil {
		if strings.TrimSpace(qppvDraft.QPPVName) != "" {
			qppvName = strings.TrimSpace(qppvDraft.QPPVName)
		}
		if strings.TrimSpace(qppvDraft.QPPVEmail) != "" {
			qppvEmail = strings.TrimSpace(qppvDraft.QPPVEmail)
		}
		if strings.TrimSpace(qppvDraft.QPPVPhone) != "" {
			qppvPhone = strings.TrimSpace(qppvDraft.QPPVPhone)
		}
		if strings.TrimSpace(qppvDraft.QPPVDeputy) != "" {
			qppvDeputy = strings.TrimSpace(qppvDraft.QPPVDeputy)
		}
		if qppvDraft.HasPSMF {
			hasPSMF = true
		}
		if qppvDraft.HasPSMFNational {
			hasPSMFNational = true
		}
	}

	cellUpdates := []TableCellCoord{
		{TableIndex: 0, RowIndex: 1, ColIndex: 1, Value: pdtName},
		{TableIndex: 0, RowIndex: 2, ColIndex: 1, Value: pdtOrdre},
		{TableIndex: 0, RowIndex: 3, ColIndex: 1, Value: pdtEmail},
		{TableIndex: 0, RowIndex: 4, ColIndex: 1, Value: pdtPhone},
		{TableIndex: 0, RowIndex: 5, ColIndex: 1, Value: asstName},
		{TableIndex: 1, RowIndex: 1, ColIndex: 1, Value: qppvName},
		{TableIndex: 1, RowIndex: 2, ColIndex: 1, Value: qppvEmail},
		{TableIndex: 1, RowIndex: 3, ColIndex: 1, Value: qppvPhone},
		{TableIndex: 1, RowIndex: 4, ColIndex: 1, Value: qppvDeputy},
	}

	textRepl := map[string]string{}
	if orgName != "" {
		textRepl["Nom de l’établissement pharmaceutique :"] = "Nom de l’établissement pharmaceutique : " + orgName
		textRepl["Nom de l’établissement pharmaceutique\u00a0:"] = "Nom de l’établissement pharmaceutique : " + orgName
	}
	if orgAddr != "" {
		textRepl["Adresse :"] = "Adresse : " + orgAddr
		textRepl["Adresse\u00a0:"] = "Adresse : " + orgAddr
	}
	if marketedCount > 0 {
		textRepl["Le nombre de produits pharmaceutiques commercialisés en Algérie : …."] = fmt.Sprintf("Le nombre de produits pharmaceutiques commercialisés en Algérie : %d", marketedCount)
		textRepl["Le nombre de produits pharmaceutiques commercialisés en Algérie\u00a0: …."] = fmt.Sprintf("Le nombre de produits pharmaceutiques commercialisés en Algérie : %d", marketedCount)
	}
	var strikeTerms []string
	joinedTypes := strings.ToLower(strings.Join(estTypes, " "))
	if strings.TrimSpace(joinedTypes) != "" {
		hasDet := strings.Contains(joinedTypes, "detenteur")
		hasExp := strings.Contains(joinedTypes, "exploit")
		hasFab := strings.Contains(joinedTypes, "fabric")
		hasImp := strings.Contains(joinedTypes, "import")
		if hasDet || hasExp || hasFab || hasImp {
			if !hasDet {
				strikeTerms = append(strikeTerms, "Détenteur de DE")
			}
			if !hasExp {
				strikeTerms = append(strikeTerms, "Exploitant de DE")
			}
			if !hasFab {
				strikeTerms = append(strikeTerms, "Fabricant")
			}
			if !hasImp {
				strikeTerms = append(strikeTerms, "Importateur")
			}
		}
	}

	cellStrikeTerms := map[[3]int][]string{
		{2, 1, 1}: {"Présence"},
		{2, 2, 1}: {"Présence"},
	}
	if hasPSMF {
		cellStrikeTerms[[3]int{2, 1, 1}] = []string{"Absence"}
	}
	if hasPSMFNational {
		cellStrikeTerms[[3]int{2, 2, 1}] = []string{"Absence"}
	}

	return FillDOCXTemplate(tpl, DOCXFillOptions{
		CellUpdates:      cellUpdates,
		TextReplacements: textRepl,
		StrikeTerms:      strikeTerms,
		CellStrikeTerms:  cellStrikeTerms,
	})
}

// GenerateDCECompositionDossierDOCX retourne la fiche officielle ANPP « composition-du-dossier-de-liberation-de-lot.docx ».
func GenerateDCECompositionDossierDOCX() ([]byte, error) {
	tpl, err := GetANPPTemplate("dce_composition_dossier_liberation.docx")
	if err != nil {
		return nil, err
	}
	return FillDOCXTemplate(tpl, DOCXFillOptions{})
}

// GenerateDCECanevasCertificatLiberationDOCX pré-remplit le canevas générique « canevas-de-certificat-de-liberation-de-lot.docx ».
func GenerateDCECanevasCertificatLiberationDOCX(org *domain.Organization, lot *domain.ImportLotRelease) ([]byte, error) {
	tpl, err := GetANPPTemplate("dce_canevas_certificat_liberation.docx")
	if err != nil {
		return nil, err
	}
	var cellUpdates []TableCellCoord
	textRepl := map[string]string{}
	if lot != nil {
		ddfStr, ddpStr := "", ""
		if lot.DDF != nil {
			ddfStr = lot.DDF.Format("2006-01-02")
		}
		if lot.DDP != nil {
			ddpStr = lot.DDP.Format("2006-01-02")
		}
		cellUpdates = []TableCellCoord{
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
		if lot.ProductionSite != "" {
			textRepl["………………………………………………………………………………………………."] = lot.ProductionSite
		}
		if lot.QCSite != "" {
			textRepl["………………………………………………………………………………………………"] = lot.QCSite
		}
		if lot.ReleaseSite != "" {
			textRepl["…………………………………………………………………………………………………"] = lot.ReleaseSite
		}
	} else if org != nil && org.LegalName != "" {
		textRepl["…………………………………………………………………………………………………"] = fmt.Sprintf("%s — %s (Agrément %s)", org.LegalName, org.Address, org.MIPHLicenseNumber)
	}
	return FillDOCXTemplate(tpl, DOCXFillOptions{
		CellUpdates:      cellUpdates,
		TextReplacements: textRepl,
	})
}

// GenerateDCEDeclarationQuotidienneDOCX pré-remplit les déclarations quotidiennes DCE (.docx) pour PP ou DM.
func GenerateDCEDeclarationQuotidienneDOCX(kind string, org *domain.Organization, lots []*domain.ImportLotRelease) ([]byte, error) {
	return GenerateDCEDeclarationQuotidienneWithDraftDOCX(kind, org, lots, nil)
}

// GenerateDCEDeclarationQuotidienneWithDraftDOCX pré-remplit les déclarations quotidiennes DCE (.docx) avec le complément SA.
func GenerateDCEDeclarationQuotidienneWithDraftDOCX(kind string, org *domain.Organization, lots []*domain.ImportLotRelease, saDraft *domain.DCESAComplementDraft) ([]byte, error) {
	isDM := strings.EqualFold(strings.TrimSpace(kind), "dm")
	tplName := "dce_decl_quotidienne_pp.docx"
	if isDM {
		tplName = "dce_decl_quotidienne_dm.docx"
	}
	tpl, err := GetANPPTemplate(tplName)
	if err != nil {
		return nil, err
	}

	var cellUpdates []TableCellCoord
	if org != nil {
		header := strings.TrimSpace(fmt.Sprintf("%s | %s | Tél: %s | Agrément MIPH: %s", org.LegalName, org.Address, org.Phone, org.MIPHLicenseNumber))
		cellUpdates = append(cellUpdates, TableCellCoord{TableIndex: 0, RowIndex: 0, ColIndex: 0, Value: header})
	}

	rowIdx := 1
	for _, lot := range lots {
		if lot == nil || rowIdx > 6 {
			continue
		}
		if isDM && lot.ProductKind != "dm" {
			continue
		}
		if !isDM && lot.ProductKind != "pp" {
			continue
		}
		if rowIdx == 1 {
			cellUpdates = append(cellUpdates,
				TableCellCoord{TableIndex: 1, RowIndex: 1, ColIndex: 1, Value: lot.ProductionSite},
				TableCellCoord{TableIndex: 1, RowIndex: 1, ColIndex: 2, Value: lot.QCSite},
				TableCellCoord{TableIndex: 1, RowIndex: 3, ColIndex: 1, Value: lot.ReleaseSite},
			)
		}
		ddfStr, ddpStr := "", ""
		if lot.DDF != nil {
			ddfStr = lot.DDF.Format("2006-01-02")
		}
		if lot.DDP != nil {
			ddpStr = lot.DDP.Format("2006-01-02")
		}
		relDate := lot.CreatedAt.Format("2006-01-02")
		if isDM {
			cellUpdates = append(cellUpdates,
				TableCellCoord{TableIndex: 2, RowIndex: rowIdx, ColIndex: 0, Value: relDate},
				TableCellCoord{TableIndex: 2, RowIndex: rowIdx, ColIndex: 1, Value: lot.CommercialName},
				TableCellCoord{TableIndex: 2, RowIndex: rowIdx, ColIndex: 2, Value: lot.Designation},
				TableCellCoord{TableIndex: 2, RowIndex: rowIdx, ColIndex: 3, Value: lot.FormAspect},
				TableCellCoord{TableIndex: 2, RowIndex: rowIdx, ColIndex: 4, Value: lot.HomologationOrDERef},
				TableCellCoord{TableIndex: 2, RowIndex: rowIdx, ColIndex: 5, Value: lot.LotNumber},
				TableCellCoord{TableIndex: 2, RowIndex: rowIdx, ColIndex: 6, Value: ddfStr},
				TableCellCoord{TableIndex: 2, RowIndex: rowIdx, ColIndex: 7, Value: ddpStr},
				TableCellCoord{TableIndex: 2, RowIndex: rowIdx, ColIndex: 8, Value: lot.SalesUnits},
			)
			if saDraft != nil {
				cellUpdates = append(cellUpdates,
					TableCellCoord{TableIndex: 2, RowIndex: rowIdx, ColIndex: 9, Value: saDraft.SALotNumber},
					TableCellCoord{TableIndex: 2, RowIndex: rowIdx, ColIndex: 10, Value: strings.TrimSpace(saDraft.SADDF + " / " + saDraft.SADDP)},
					TableCellCoord{TableIndex: 2, RowIndex: rowIdx, ColIndex: 11, Value: saDraft.SAManufacturer},
				)
			}
		} else {
			cellUpdates = append(cellUpdates,
				TableCellCoord{TableIndex: 2, RowIndex: rowIdx, ColIndex: 0, Value: relDate},
				TableCellCoord{TableIndex: 2, RowIndex: rowIdx, ColIndex: 1, Value: lot.CommercialName},
				TableCellCoord{TableIndex: 2, RowIndex: rowIdx, ColIndex: 2, Value: lot.Designation},
				TableCellCoord{TableIndex: 2, RowIndex: rowIdx, ColIndex: 3, Value: lot.FormAspect},
				TableCellCoord{TableIndex: 2, RowIndex: rowIdx, ColIndex: 5, Value: lot.HomologationOrDERef},
				TableCellCoord{TableIndex: 2, RowIndex: rowIdx, ColIndex: 6, Value: lot.LotNumber},
				TableCellCoord{TableIndex: 2, RowIndex: rowIdx, ColIndex: 7, Value: ddfStr},
				TableCellCoord{TableIndex: 2, RowIndex: rowIdx, ColIndex: 8, Value: ddpStr},
				TableCellCoord{TableIndex: 2, RowIndex: rowIdx, ColIndex: 9, Value: lot.SalesUnits},
			)
			if saDraft != nil {
				cellUpdates = append(cellUpdates,
					TableCellCoord{TableIndex: 2, RowIndex: rowIdx, ColIndex: 10, Value: saDraft.SALotNumber},
					TableCellCoord{TableIndex: 2, RowIndex: rowIdx, ColIndex: 11, Value: strings.TrimSpace(saDraft.SADDF + " / " + saDraft.SADDP)},
					TableCellCoord{TableIndex: 2, RowIndex: rowIdx, ColIndex: 12, Value: saDraft.SAManufacturer},
				)
			}
		}
		rowIdx++
	}

	return FillDOCXTemplate(tpl, DOCXFillOptions{CellUpdates: cellUpdates})
}

// GenerateDCEDeclarationJournaliereXLSX pré-remplit les canevas Excel de déclaration journalière (« pp », « dm » ou « pp_dm »).
func GenerateDCEDeclarationJournaliereXLSX(kind string, org *domain.Organization, lots []*domain.ImportLotRelease) ([]byte, error) {
	return GenerateDCEDeclarationJournaliereWithDraftXLSX(kind, org, lots, nil)
}

// GenerateDCEDeclarationJournaliereWithDraftXLSX pré-remplit les canevas Excel avec le complément SA.
func GenerateDCEDeclarationJournaliereWithDraftXLSX(kind string, org *domain.Organization, lots []*domain.ImportLotRelease, saDraft *domain.DCESAComplementDraft) ([]byte, error) {
	k := strings.ToLower(strings.TrimSpace(kind))
	isCombined := k == "pp_dm"
	isDMOnly := k == "dm"
	tplName := "dce_decl_journaliere_pp.xlsx"
	startRow := 14
	if isCombined {
		tplName = "dce_decl_journaliere_pp_dm.xlsx"
		startRow = 15
	} else if isDMOnly {
		tplName = "ddm_decl_journaliere_dm.xlsx"
		startRow = 14
	}
	tpl, err := GetANPPTemplate(tplName)
	if err != nil {
		return nil, err
	}

	var updates []XLSXCellUpdate
	if org != nil {
		header := strings.TrimSpace(fmt.Sprintf("%s — %s — Agrément %s", org.LegalName, org.Address, org.MIPHLicenseNumber))
		updates = append(updates,
			XLSXCellUpdate{SheetPath: "xl/worksheets/sheet1.xml", CellRef: "A1", Value: header},
			XLSXCellUpdate{SheetPath: "xl/worksheets/sheet1.xml", CellRef: "C4", Value: org.LegalName},
			XLSXCellUpdate{SheetPath: "xl/worksheets/sheet1.xml", CellRef: "D4", Value: org.Address},
			XLSXCellUpdate{SheetPath: "xl/worksheets/sheet1.xml", CellRef: "E4", Value: org.MIPHLicenseNumber},
		)
		if isCombined {
			updates = append(updates,
				XLSXCellUpdate{SheetPath: "xl/worksheets/sheet2.xml", CellRef: "A1", Value: header},
				XLSXCellUpdate{SheetPath: "xl/worksheets/sheet2.xml", CellRef: "C4", Value: org.LegalName},
				XLSXCellUpdate{SheetPath: "xl/worksheets/sheet2.xml", CellRef: "D4", Value: org.Address},
				XLSXCellUpdate{SheetPath: "xl/worksheets/sheet2.xml", CellRef: "E4", Value: org.MIPHLicenseNumber},
			)
		}
	}

	ppRow, dmRow := startRow, startRow
	for _, lot := range lots {
		if lot == nil {
			continue
		}
		ddfStr, ddpStr := "", ""
		if lot.DDF != nil {
			ddfStr = lot.DDF.Format("2006-01-02")
		}
		if lot.DDP != nil {
			ddpStr = lot.DDP.Format("2006-01-02")
		}
		relDate := lot.CreatedAt.Format("2006-01-02")

		if lot.ProductKind == "dm" {
			if !isCombined && !isDMOnly {
				continue
			}
			sheet := "xl/worksheets/sheet2.xml"
			if isDMOnly {
				sheet = "xl/worksheets/sheet1.xml"
			}
			r := dmRow
			updates = append(updates,
				XLSXCellUpdate{SheetPath: sheet, CellRef: fmt.Sprintf("A%d", r), Value: relDate},
				XLSXCellUpdate{SheetPath: sheet, CellRef: fmt.Sprintf("B%d", r), Value: lot.CommercialName},
				XLSXCellUpdate{SheetPath: sheet, CellRef: fmt.Sprintf("C%d", r), Value: lot.Designation},
				XLSXCellUpdate{SheetPath: sheet, CellRef: fmt.Sprintf("D%d", r), Value: lot.FormAspect},
				XLSXCellUpdate{SheetPath: sheet, CellRef: fmt.Sprintf("F%d", r), Value: lot.HomologationOrDERef},
				XLSXCellUpdate{SheetPath: sheet, CellRef: fmt.Sprintf("G%d", r), Value: lot.DMClass},
				XLSXCellUpdate{SheetPath: sheet, CellRef: fmt.Sprintf("H%d", r), Value: lot.LotNumber},
				XLSXCellUpdate{SheetPath: sheet, CellRef: fmt.Sprintf("I%d", r), Value: ddfStr},
				XLSXCellUpdate{SheetPath: sheet, CellRef: fmt.Sprintf("J%d", r), Value: ddpStr},
				XLSXCellUpdate{SheetPath: sheet, CellRef: fmt.Sprintf("K%d", r), Value: lot.SalesUnits},
			)
			if saDraft != nil {
				updates = append(updates,
					XLSXCellUpdate{SheetPath: sheet, CellRef: fmt.Sprintf("L%d", r), Value: saDraft.SAName},
					XLSXCellUpdate{SheetPath: sheet, CellRef: fmt.Sprintf("M%d", r), Value: saDraft.SALotNumber},
					XLSXCellUpdate{SheetPath: sheet, CellRef: fmt.Sprintf("N%d", r), Value: saDraft.SADDF},
					XLSXCellUpdate{SheetPath: sheet, CellRef: fmt.Sprintf("O%d", r), Value: saDraft.SADDP},
					XLSXCellUpdate{SheetPath: sheet, CellRef: fmt.Sprintf("P%d", r), Value: saDraft.SAManufacturer},
				)
			}
			dmRow++
		} else if !isDMOnly {
			r := ppRow
			updates = append(updates,
				XLSXCellUpdate{SheetPath: "xl/worksheets/sheet1.xml", CellRef: fmt.Sprintf("A%d", r), Value: relDate},
				XLSXCellUpdate{SheetPath: "xl/worksheets/sheet1.xml", CellRef: fmt.Sprintf("B%d", r), Value: lot.CommercialName},
				XLSXCellUpdate{SheetPath: "xl/worksheets/sheet1.xml", CellRef: fmt.Sprintf("C%d", r), Value: lot.Designation},
				XLSXCellUpdate{SheetPath: "xl/worksheets/sheet1.xml", CellRef: fmt.Sprintf("D%d", r), Value: lot.FormAspect},
				XLSXCellUpdate{SheetPath: "xl/worksheets/sheet1.xml", CellRef: fmt.Sprintf("F%d", r), Value: lot.HomologationOrDERef},
				XLSXCellUpdate{SheetPath: "xl/worksheets/sheet1.xml", CellRef: fmt.Sprintf("H%d", r), Value: lot.LotNumber},
				XLSXCellUpdate{SheetPath: "xl/worksheets/sheet1.xml", CellRef: fmt.Sprintf("I%d", r), Value: ddfStr},
				XLSXCellUpdate{SheetPath: "xl/worksheets/sheet1.xml", CellRef: fmt.Sprintf("J%d", r), Value: ddpStr},
				XLSXCellUpdate{SheetPath: "xl/worksheets/sheet1.xml", CellRef: fmt.Sprintf("K%d", r), Value: lot.SalesUnits},
			)
			if saDraft != nil {
				firstLot := "NON"
				if saDraft.FirstMarketedLot {
					firstLot = "OUI"
				}
				updates = append(updates,
					XLSXCellUpdate{SheetPath: "xl/worksheets/sheet1.xml", CellRef: fmt.Sprintf("G%d", r), Value: saDraft.ChannelAllocation},
					XLSXCellUpdate{SheetPath: "xl/worksheets/sheet1.xml", CellRef: fmt.Sprintf("L%d", r), Value: saDraft.SALotNumber},
					XLSXCellUpdate{SheetPath: "xl/worksheets/sheet1.xml", CellRef: fmt.Sprintf("M%d", r), Value: saDraft.SADDF},
					XLSXCellUpdate{SheetPath: "xl/worksheets/sheet1.xml", CellRef: fmt.Sprintf("N%d", r), Value: saDraft.SADDP},
					XLSXCellUpdate{SheetPath: "xl/worksheets/sheet1.xml", CellRef: fmt.Sprintf("O%d", r), Value: saDraft.SAManufacturer},
					XLSXCellUpdate{SheetPath: "xl/worksheets/sheet1.xml", CellRef: fmt.Sprintf("P%d", r), Value: firstLot},
				)
			}
			ppRow++
		}
	}

	return FillXLSXTemplate(tpl, updates)
}

// GenerateDCETemplateCASRSecondaireDOCX pré-remplit le prototype Certificat d'analyse SR secondaire.
func GenerateDCETemplateCASRSecondaireDOCX(org *domain.Organization) ([]byte, error) {
	return GenerateDCETemplateCASRSecondaireWithDraftDOCX(org, nil)
}

// GenerateDCETemplateCASRSecondaireWithDraftDOCX pré-remplit le prototype Certificat d'analyse SR secondaire depuis un brouillon.
func GenerateDCETemplateCASRSecondaireWithDraftDOCX(org *domain.Organization, draft *domain.CASRSecondaireDraft) ([]byte, error) {
	tpl, err := GetANPPTemplate("dce_template_ca_sr_secondaire.docx")
	if err != nil {
		return nil, err
	}
	textRepl := map[string]string{}
	var cellUpdates []TableCellCoord
	if org != nil && strings.TrimSpace(org.LegalName) != "" {
		textRepl["Nom du fabricant/fournisseur :"] = "Nom du fabricant/fournisseur : " + strings.TrimSpace(org.LegalName)
	}
	if draft != nil {
		if draft.StandardName != "" {
			textRepl["Nom du standard de référence secondaire :"] = "Nom du standard de référence secondaire : " + draft.StandardName
		}
		if draft.LotCode != "" {
			textRepl["Numéro de lot/code d’identification interne :"] = "Numéro de lot/code d’identification interne : " + draft.LotCode
		}
		if draft.ExpiryDate != "" {
			textRepl["Date d’expiration/de re-test :"] = "Date d’expiration/de re-test : " + draft.ExpiryDate
		}
		if draft.Manufacturer != "" {
			textRepl["Nom du fabricant/fournisseur :"] = "Nom du fabricant/fournisseur : " + draft.Manufacturer
		}
		if draft.MethodRef != "" {
			textRepl["Référence de la méthode de standardisation\u00a0:"] = "Référence de la méthode de standardisation : " + draft.MethodRef
			textRepl["Référence de la méthode de standardisation :"] = "Référence de la méthode de standardisation : " + draft.MethodRef
		}
		if draft.Weighing != "" {
			textRepl["Pesée :"] = "Pesée : " + draft.Weighing
		}
		if draft.StorageConditions != "" {
			textRepl["Conditions de conservation :"] = "Conditions de conservation : " + draft.StorageConditions
		}
		if draft.SafetyInstructions != "" {
			textRepl["Consignes de sécurité (joindre la fiche des données de sécurités FDS)\u00a0:"] = "Consignes de sécurité (FDS jointe) : " + draft.SafetyInstructions
			textRepl["Consignes de sécurité (joindre la fiche des données de sécurités FDS) :"] = "Consignes de sécurité (FDS jointe) : " + draft.SafetyInstructions
		}
		if draft.PrimaryLotNumber != "" {
			textRepl["□   Ce standard de référence est traçable au SR primaire, N°lot\u00a0: ........"] = "☑   Ce standard de référence est traçable au SR primaire, N°lot : " + draft.PrimaryLotNumber
		}
		if draft.StabilityValidated {
			textRepl["□   La durée de stabilité du standard est validée\u00a0;"] = "☑   La durée de stabilité du standard est validée ;"
		}
		if draft.EquipQualified {
			textRepl["□   Les équipements utilisés pour la standardisation sont qualifiés"] = "☑   Les équipements utilisés pour la standardisation sont qualifiés"
		}
		cellUpdates = append(cellUpdates,
			TableCellCoord{TableIndex: 0, RowIndex: 1, ColIndex: 1, Value: draft.DescSpec},
			TableCellCoord{TableIndex: 0, RowIndex: 1, ColIndex: 2, Value: draft.DescResult},
			TableCellCoord{TableIndex: 0, RowIndex: 2, ColIndex: 1, Value: draft.IdentSpec},
			TableCellCoord{TableIndex: 0, RowIndex: 2, ColIndex: 2, Value: draft.IdentResult},
			TableCellCoord{TableIndex: 0, RowIndex: 3, ColIndex: 1, Value: draft.AssaySpec},
			TableCellCoord{TableIndex: 0, RowIndex: 3, ColIndex: 2, Value: draft.AssayResult},
			TableCellCoord{TableIndex: 0, RowIndex: 4, ColIndex: 1, Value: draft.PuritySpec},
			TableCellCoord{TableIndex: 0, RowIndex: 4, ColIndex: 2, Value: draft.PurityResult},
			TableCellCoord{TableIndex: 0, RowIndex: 5, ColIndex: 1, Value: draft.WaterSpec},
			TableCellCoord{TableIndex: 0, RowIndex: 5, ColIndex: 2, Value: draft.WaterResult},
			TableCellCoord{TableIndex: 1, RowIndex: 1, ColIndex: 1, Value: draft.PreparedBy},
			TableCellCoord{TableIndex: 1, RowIndex: 1, ColIndex: 2, Value: draft.VerifiedBy},
			TableCellCoord{TableIndex: 1, RowIndex: 1, ColIndex: 3, Value: draft.ApprovedBy},
		)
	}
	return FillDOCXTemplate(tpl, DOCXFillOptions{
		TextReplacements: textRepl,
		CellUpdates:      cellUpdates,
	})
}

// GenerateDEPPStructurePrixWithDraftDOCX pré-remplit le modèle Structure de prix à la fabrication locale depuis un brouillon.
func GenerateDEPPStructurePrixWithDraftDOCX(draft *domain.StructurePrixDraft) ([]byte, error) {
	tpl, err := GetANPPTemplate("depp_modele_structure.docx")
	if err != nil {
		return nil, err
	}
	var cellUpdates []TableCellCoord
	textRepl := map[string]string{}
	if draft != nil {
		cellUpdates = []TableCellCoord{
			{TableIndex: 0, RowIndex: 1, ColIndex: 1, Value: draft.SubstanceActive},
			{TableIndex: 0, RowIndex: 2, ColIndex: 1, Value: draft.MatieresImportees},
			{TableIndex: 0, RowIndex: 3, ColIndex: 1, Value: draft.MatieresLocales},
			{TableIndex: 0, RowIndex: 4, ColIndex: 1, Value: draft.ServicesImportes},
			{TableIndex: 0, RowIndex: 5, ColIndex: 1, Value: draft.ServicesLocaux},
			{TableIndex: 0, RowIndex: 6, ColIndex: 1, Value: draft.FraisPersonnel},
			{TableIndex: 0, RowIndex: 7, ColIndex: 1, Value: draft.Taxes},
			{TableIndex: 0, RowIndex: 8, ColIndex: 1, Value: draft.FraisFinanciers},
			{TableIndex: 0, RowIndex: 9, ColIndex: 1, Value: draft.Amortissements},
			{TableIndex: 0, RowIndex: 10, ColIndex: 1, Value: draft.AutresCharges},
			{TableIndex: 0, RowIndex: 11, ColIndex: 1, Value: draft.PrixRevient},
			{TableIndex: 0, RowIndex: 12, ColIndex: 1, Value: draft.PCSU},
			{TableIndex: 0, RowIndex: 13, ColIndex: 1, Value: draft.TauxExportation},
			{TableIndex: 0, RowIndex: 14, ColIndex: 1, Value: draft.TauxIntegration},
		}
		if strings.TrimSpace(draft.TauxChange) != "" {
			textRepl["NB\u00a0: Mentionnez le taux de change utilisé."] = "NB : Taux de change utilisé : " + strings.TrimSpace(draft.TauxChange)
			textRepl["NB : Mentionnez le taux de change utilisé."] = "NB : Taux de change utilisé : " + strings.TrimSpace(draft.TauxChange)
		}
	}
	return FillDOCXTemplate(tpl, DOCXFillOptions{
		CellUpdates:      cellUpdates,
		TextReplacements: textRepl,
	})
}
