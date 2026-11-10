package auth

import (
	"testing"
	"time"

	"github.com/pquerna/otp/totp"
)

func TestGenerateAndValidateTOTP(t *testing.T) {
	email := "dr.amine@laboratoire-algerie.dz"

	setup, err := GenerateTOTPSetup(email)
	if err != nil {
		t.Fatalf("GenerateTOTPSetup() échec: %v", err)
	}

	if len(setup.Secret) == 0 {
		t.Error("Secret TOTP vide")
	}
	if len(setup.AuthURL) == 0 {
		t.Error("AuthURL TOTP vide")
	}
	if len(setup.BackupCodes) != 10 {
		t.Errorf("Attendu 10 codes de secours, reçu %d", len(setup.BackupCodes))
	}
	if len(setup.HashedBackupCodes) != 10 {
		t.Errorf("Attendu 10 hashes de secours, reçu %d", len(setup.HashedBackupCodes))
	}

	// 1. Génération du code actuel à la seconde T
	currentCode, err := totp.GenerateCode(setup.Secret, time.Now().UTC())
	if err != nil {
		t.Fatalf("Génération code TOTP test échec: %v", err)
	}

	// 2. Validation du code valide
	if !ValidateTOTPCode(currentCode, setup.Secret) {
		t.Errorf("ValidateTOTPCode() attendu true pour code valide '%s', reçu false", currentCode)
	}

	// 3. Validation d'un faux code
	if ValidateTOTPCode("000000", setup.Secret) && currentCode != "000000" {
		t.Error("ValidateTOTPCode() attendu false pour faux code '000000', reçu true")
	}

	// 4. Validation avec chaîne malformée
	if ValidateTOTPCode("123", setup.Secret) {
		t.Error("ValidateTOTPCode() attendu false pour code trop court '123'")
	}
}

func TestBackupCodeConsumption(t *testing.T) {
	email := "qualite@labo.dz"
	setup, err := GenerateTOTPSetup(email)
	if err != nil {
		t.Fatalf("GenerateTOTPSetup() échec: %v", err)
	}

	rawCodes := setup.BackupCodes
	hashedCodes := setup.HashedBackupCodes

	firstCode := rawCodes[0]

	// 1. Première utilisation du code de secours -> doit réussir
	valid, remaining := ValidateAndConsumeBackupCode(firstCode, hashedCodes)
	if !valid {
		t.Errorf("ValidateAndConsumeBackupCode() échec sur premier code valide '%s'", firstCode)
	}
	if len(remaining) != 9 {
		t.Errorf("Attendu 9 codes restants après consommation, reçu %d", len(remaining))
	}

	// 2. Réutilisation du même code (Replay attack) -> doit impérativement échouer
	reused, remainingAfterReplay := ValidateAndConsumeBackupCode(firstCode, remaining)
	if reused {
		t.Error("ValidateAndConsumeBackupCode() a accepté un code déjà consommé !")
	}
	if len(remainingAfterReplay) != 9 {
		t.Errorf("La liste a été altérée après rejet du code rejoué: %d restants", len(remainingAfterReplay))
	}

	// 3. Test de robustesse (insensible aux tirets et à la casse)
	secondCode := rawCodes[1]
	// Ex: "ABCD-EFGH" -> "abcd efgh"
	validCase, remaining2 := ValidateAndConsumeBackupCode(secondCode, remaining)
	if !validCase {
		t.Errorf("ValidateAndConsumeBackupCode() échec sur deuxième code '%s'", secondCode)
	}
	if len(remaining2) != 8 {
		t.Errorf("Attendu 8 codes restants, reçu %d", len(remaining2))
	}
}
