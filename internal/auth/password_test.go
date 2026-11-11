package auth

import (
	"testing"
)

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  error
	}{
		{
			name:     "Mot de passe conforme BPF",
			password: "P@ssw0rdPharma2026!",
			wantErr:  nil,
		},
		{
			name:     "Trop court (< 12 car.)",
			password: "Short123!",
			wantErr:  ErrPasswordTooShort,
		},
		{
			name:     "Manque majuscule",
			password: "password123456!@",
			wantErr:  ErrPasswordNoUpper,
		},
		{
			name:     "Manque minuscule",
			password: "PASSWORD123456!@",
			wantErr:  ErrPasswordNoLower,
		},
		{
			name:     "Manque chiffre",
			password: "PasswordSpecialOnly!@#",
			wantErr:  ErrPasswordNoDigit,
		},
		{
			name:     "Manque caractère spécial",
			password: "Password12345678",
			wantErr:  ErrPasswordNoSpecial,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePassword(tt.password)
			if err != tt.wantErr {
				t.Errorf("ValidatePassword() error = %v, wantErr = %v", err, tt.wantErr)
			}
		})
	}
}

func TestArgon2idHashAndVerify(t *testing.T) {
	password := "S3cur3PharmaPass!2026"

	// Hachage avec paramètres allégés pour exécution rapide des tests unitaires
	testParams := &Argon2Params{
		Memory:      16 * 1024, // 16 Mo pour les tests
		Iterations:  1,
		Parallelism: 2,
		SaltLength:  16,
		KeyLength:   32,
	}

	hash, err := HashPassword(password, testParams)
	if err != nil {
		t.Fatalf("HashPassword() échec: %v", err)
	}

	if len(hash) == 0 {
		t.Fatal("HashPassword() a retourné une chaîne vide")
	}

	// 1. Vérification avec mot de passe correct
	valid, err := VerifyPassword(password, hash)
	if err != nil {
		t.Fatalf("VerifyPassword() erreur inattendue: %v", err)
	}
	if !valid {
		t.Error("VerifyPassword() attendu true pour mot de passe correct, reçu false")
	}

	// 2. Vérification avec mot de passe incorrect
	invalid, err := VerifyPassword("MauvaisMotDePasse!123", hash)
	if err != nil {
		t.Fatalf("VerifyPassword() erreur inattendue: %v", err)
	}
	if invalid {
		t.Error("VerifyPassword() attendu false pour mot de passe incorrect, reçu true")
	}

	// 3. Format invalide
	_, err = VerifyPassword(password, "invalid$format")
	if err == nil {
		t.Error("VerifyPassword() attendu erreur pour hash malformé, reçu nil")
	}
}
