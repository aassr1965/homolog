package domain

import (
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestDossierValidation(t *testing.T) {
	orgID := uuid.New()
	medID := uuid.New()
	dmID := uuid.New()

	t.Run("Dossier Médicament valide (Procédure NI)", func(t *testing.T) {
		d := &Dossier{
			OrganizationID:    orgID,
			DossierCode:       "DOS-2026-AMOX-NI",
			ProcedureCode:     ProcedureNI,
			ProductType:       ProductTypeMedicine,
			MedicineID:        &medID,
			InternalReference: "REF-INT-001",
		}
		if err := d.ValidateDossier(); err != nil {
			t.Fatalf("dossier valide rejeté: %v", err)
		}
	})

	t.Run("Dossier DM valide (Homologation)", func(t *testing.T) {
		d := &Dossier{
			OrganizationID:    orgID,
			DossierCode:       "DOS-2026-SERINGUE-DM",
			ProcedureCode:     ProcedureDM,
			ProductType:       ProductTypeMedicalDevice,
			MedicalDeviceID:   &dmID,
			InternalReference: "REF-INT-002",
		}
		if err := d.ValidateDossier(); err != nil {
			t.Fatalf("dossier DM valide rejeté: %v", err)
		}
	})

	t.Run("Dossier sans code rejeté", func(t *testing.T) {
		d := &Dossier{
			OrganizationID:    orgID,
			DossierCode:       "",
			ProcedureCode:     ProcedureNF,
			ProductType:       ProductTypeMedicine,
			MedicineID:        &medID,
			InternalReference: "REF-003",
		}
		if err := d.ValidateDossier(); err == nil {
			t.Fatal("dossier sans code aurait dû être rejeté")
		}
	})

	t.Run("Procédure inconnue rejetée", func(t *testing.T) {
		d := &Dossier{
			OrganizationID:    orgID,
			DossierCode:       "DOS-INVALIDE",
			ProcedureCode:     "PROC_INCONNUE",
			ProductType:       ProductTypeMedicine,
			MedicineID:        &medID,
			InternalReference: "REF-004",
		}
		if err := d.ValidateDossier(); err == nil {
			t.Fatal("procédure inconnue aurait dû être rejetée")
		}
	})

	t.Run("Dossier Médicament sans MedicineID rejeté", func(t *testing.T) {
		d := &Dossier{
			OrganizationID:    orgID,
			DossierCode:       "DOS-NO-MED",
			ProcedureCode:     ProcedureNI,
			ProductType:       ProductTypeMedicine,
			MedicineID:        nil,
			InternalReference: "REF-005",
		}
		if err := d.ValidateDossier(); err == nil {
			t.Fatal("dossier médicament sans medicine_id aurait dû être rejeté")
		}
	})
}

func TestGenerateCTDChecklist_Specificites(t *testing.T) {
	t.Run("Fabrication Locale Générique (NF) : Exigences 1.P.8 et 1.P.9 requises", func(t *testing.T) {
		items := GenerateCTDChecklist(ProcedureNF, OriginGeneriqueLocal, false)

		var item1P8, item1P9, item1H4, item32P35 *ChecklistItem
		for i := range items {
			switch items[i].ItemCode {
			case "1.P.8":
				item1P8 = &items[i]
			case "1.P.9":
				item1P9 = &items[i]
			case "1.H.4":
				item1H4 = &items[i]
			case "3.2.P.3.5":
				item32P35 = &items[i]
			}
		}

		if item1P8 == nil || item1P8.RequirementLevel != RequirementExiged || !item1P8.IsMandatory {
			t.Fatalf("1.P.8 (Lettre non-commercialisation lots validation) doit être obligatoirement exigée [E] pour NF, reçu: %+v", item1P8)
		}
		if item1P9 == nil || item1P9.RequirementLevel != RequirementExiged || !item1P9.IsMandatory {
			t.Fatalf("1.P.9 (Engagement stabilités Zone IVa/IVb) doit être obligatoirement exigée [E] pour NF, reçu: %+v", item1P9)
		}
		if item1H4 == nil || item1H4.RequirementLevel != RequirementExiged {
			t.Fatalf("1.H.4 (Structure de prix PCSU) doit être obligatoirement exigée [E] pour NF, reçu: %+v", item1H4)
		}
		if item32P35 == nil || item32P35.RequirementLevel != RequirementExiged {
			t.Fatalf("3.2.P.3.5 (Lots de validation sur site national) doit être exigé [E] pour fabrication locale, reçu: %+v", item32P35)
		}
	})

	t.Run("Importation Générique (NI) : Benchmark 1.H.5 exigé, 1.P.8 non applicable", func(t *testing.T) {
		items := GenerateCTDChecklist(ProcedureNI, OriginGeneriqueImporte, false)

		var item1H5, item1P8 *ChecklistItem
		for i := range items {
			switch items[i].ItemCode {
			case "1.H.5":
				item1H5 = &items[i]
			case "1.P.8":
				item1P8 = &items[i]
			}
		}

		if item1H5 == nil || item1H5.RequirementLevel != RequirementExiged || !item1H5.IsMandatory {
			t.Fatalf("1.H.5 (Benchmark prix international) doit être obligatoire pour importation NI, reçu: %+v", item1H5)
		}
		if item1P8 == nil || item1P8.RequirementLevel != RequirementNotApplicable || item1P8.IsMandatory {
			t.Fatalf("1.P.8 ne doit pas être exigé pour un produit importé, reçu: %+v", item1P8)
		}
	})

	t.Run("Renouvellement Importé (RI) : Déclaration non-modification 1.H.10 exigée", func(t *testing.T) {
		items := GenerateCTDChecklist(ProcedureRI, OriginGeneriqueImporte, false)

		var item1H10, item1A9 *ChecklistItem
		for i := range items {
			switch items[i].ItemCode {
			case "1.H.10":
				item1H10 = &items[i]
			case "1.A.9":
				item1A9 = &items[i]
			}
		}

		if item1H10 == nil || item1H10.RequirementLevel != RequirementExiged || !item1H10.IsMandatory {
			t.Fatalf("1.H.10 (Déclaration de non-modification) doit être exigée pour renouvellement RI, reçu: %+v", item1H10)
		}
		if item1A9 == nil || item1A9.RequirementLevel != RequirementExiged {
			t.Fatalf("1.A.9 (Historique DE et anciens dépôts) doit être exigée pour RI, reçu: %+v", item1A9)
		}
	})

	t.Run("Générique avec Bioéquivalence 5.3.1.2 et 5.3.1.4 obligatoires", func(t *testing.T) {
		items := GenerateCTDChecklist(ProcedureNF, OriginGeneriqueLocal, false)

		found5312 := false
		found5314 := false
		for _, it := range items {
			if it.ItemCode == "5.3.1.2" && it.RequirementLevel == RequirementExiged && it.IsMandatory {
				found5312 = true
			}
			if it.ItemCode == "5.3.1.4" && it.RequirementLevel == RequirementExiged && it.IsMandatory {
				found5314 = true
			}
		}

		if !found5312 {
			t.Fatal("5.3.1.2 (Étude comparative BE in vivo) manquante dans la checklist générique")
		}
		if !found5314 {
			t.Fatal("5.3.1.4 (Méthodes bioanalytiques humaines) manquante dans la checklist générique")
		}
	})

	t.Run("Princeps : Modules 2 et 4 exigés obligatoirement [E]", func(t *testing.T) {
		items := GenerateCTDChecklist(ProcedureNI, OriginPrinceps, true) // Princeps = true

		var m2, m4 *ChecklistItem
		for i := range items {
			switch items[i].ItemCode {
			case "2.Overview":
				m2 = &items[i]
			case "4.Safety":
				m4 = &items[i]
			}
		}

		if m2 == nil || m2.RequirementLevel != RequirementExiged || !m2.IsMandatory {
			t.Fatalf("Module 2 doit être obligatoirement exigé [E] pour Princeps, reçu: %+v", m2)
		}
		if m4 == nil || m4.RequirementLevel != RequirementExiged || !m4.IsMandatory {
			t.Fatalf("Module 4 (Toxicologie/Sécurité) doit être obligatoirement exigé [E] pour Princeps, reçu: %+v", m4)
		}
	})
}

func TestGenerateDMChecklist(t *testing.T) {
	t.Run("Dispositif Médical Stérile Classe IIa", func(t *testing.T) {
		items := GenerateDMChecklist("IIa", true)

		hasSterileValidation := false
		hasISO13485 := false
		for _, it := range items {
			if it.ItemCode == "DM.S.1" && it.IsMandatory {
				hasSterileValidation = true
			}
			if it.ItemCode == "DM.Q.1" && it.IsMandatory {
				hasISO13485 = true
			}
		}

		if !hasSterileValidation {
			t.Fatal("DM.S.1 (Validation du procédé de stérilisation) doit être exigé pour un DM stérile")
		}
		if !hasISO13485 {
			t.Fatal("DM.Q.1 (Certificat ISO 13485) doit être présent et obligatoire")
		}
	})

	t.Run("Dispositif Médical Non-Stérile Classe I : Pas d'exigence de stérilisation", func(t *testing.T) {
		items := GenerateDMChecklist("I", false)

		for _, it := range items {
			if it.ItemCode == "DM.S.1" || it.ItemCode == "DM.S.2" {
				t.Fatalf("Les exigences de stérilisation ne doivent pas être présentes pour un DM non-stérile, reçu: %s", it.ItemCode)
			}
		}
	})

	t.Run("Dispositif Médical Classe III : Évaluation clinique obligatoire DM.C.1", func(t *testing.T) {
		items := GenerateDMChecklist("III", true)

		hasClinical := false
		for _, it := range items {
			if it.ItemCode == "DM.C.1" && it.IsMandatory {
				hasClinical = true
			}
		}

		if !hasClinical {
			t.Fatal("DM.C.1 (Rapport d'évaluation clinique) obligatoire pour Classe III")
		}
	})
}

func TestCalculateCompletenessAndLock(t *testing.T) {
	dossierID := uuid.New()
	docID := uuid.New()
	dtID := uuid.New()

	items := []ChecklistItem{
		{DossierID: dossierID, ItemCode: "1.A.1", IsMandatory: true, RequirementLevel: RequirementExiged, Status: ItemStatusMissing},
		{DossierID: dossierID, ItemCode: "1.A.2", IsMandatory: true, RequirementLevel: RequirementExiged, Status: ItemStatusMissing},
		{DossierID: dossierID, ItemCode: "1.A.3", IsMandatory: true, RequirementLevel: RequirementExiged, Status: ItemStatusMissing},
		{DossierID: dossierID, ItemCode: "1.A.4", IsMandatory: true, RequirementLevel: RequirementExiged, Status: ItemStatusMissing},
		{DossierID: dossierID, ItemCode: "1.A.9", IsMandatory: false, RequirementLevel: RequirementNotApplicable, Status: ItemStatusMissing},
	}

	// 0 / 4 obligatoires attachées = 0%
	if score := CalculateCompleteness(items); score != 0 {
		t.Fatalf("attendu 0%%, reçu: %d%%", score)
	}

	// 2 / 4 obligatoires attachées = 50%
	items[0].Status = ItemStatusAttached
	items[1].Status = ItemStatusValidated
	if score := CalculateCompleteness(items); score != 50 {
		t.Fatalf("attendu 50%%, reçu: %d%%", score)
	}

	// 4 / 4 obligatoires attachées = 100%
	items[2].Status = ItemStatusAttached
	items[3].Status = ItemStatusAttached
	if score := CalculateCompleteness(items); score != 100 {
		t.Fatalf("attendu 100%%, reçu: %d%%", score)
	}

	d := &Dossier{
		ID:                  dossierID,
		Status:              DossierStatusInPreparation,
		CompletenessPercent: 100,
	}

	// Complétude 100% mais SANS visa PDT -> Doit être refusé par verrou BPF
	items[0].GEDDocumentID = &docID
	items[1].GEDDocumentID = &docID
	items[2].GEDDocumentID = &docID
	items[3].GEDDocumentID = &docID

	if err := CanValidateInternal(d, items); err == nil {
		t.Fatal("validation interne aurait dû être refusée car absence de visa légal du Directeur Technique")
	}

	// Avec signature légale PDT -> Validé avec succès
	now := time.Now()
	seal := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
	role := "Pharmacien Directeur Technique (PDT)"
	d.PDTSignedAt = &now
	d.PDTSignedBy = &dtID
	d.PDTSignatureSeal = &seal
	d.PDTSignatureRole = &role

	if err := CanValidateInternal(d, items); err != nil {
		t.Fatalf("la validation interne aurait dû réussir avec 100%% de complétude et signature PDT, reçu: %v", err)
	}
}
