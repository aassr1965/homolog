package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pquerna/otp"
	"github.com/pquerna/otp/totp"
)

// Erreurs liées au 2FA TOTP
var (
	ErrInvalidTOTPCode   = errors.New("code d'authentification 2FA invalide ou expiré")
	ErrInvalidBackupCode = errors.New("code de secours invalide ou déjà consommé")
)

// TOTPSetupData contient les informations requises pour l'enrôlement 2FA hors-ligne
type TOTPSetupData struct {
	Secret            string   `json:"secret"`             // Clé secrète Base32 (pour saisie manuelle)
	AuthURL           string   `json:"auth_url"`           // URL standard otpauth://
	BackupCodes       []string `json:"backup_codes"`       // 10 codes de secours en clair (affichés une seule fois)
	HashedBackupCodes []string `json:"-"`                  // Empreintes SHA-256 à persister en base de données
}

// GenerateTOTPSetup génère un secret RFC 6238 et 10 codes de secours 100% hors-ligne
func GenerateTOTPSetup(accountEmail string) (*TOTPSetupData, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "Homolog.dz",
		AccountName: accountEmail,
		Period:      30,
		Digits:      otp.DigitsSix,
		Algorithm:   otp.AlgorithmSHA1,
	})
	if err != nil {
		return nil, fmt.Errorf("impossible de générer la clé TOTP: %w", err)
	}

	backupCodes, hashedCodes, err := generateBackupCodes(10)
	if err != nil {
		return nil, fmt.Errorf("impossible de générer les codes de secours: %w", err)
	}

	return &TOTPSetupData{
		Secret:            key.Secret(),
		AuthURL:           key.URL(),
		BackupCodes:       backupCodes,
		HashedBackupCodes: hashedCodes,
	}, nil
}

// ValidateTOTPCode valide un code à 6 chiffres avec tolérance de dérive d'horloge (±30s)
func ValidateTOTPCode(passcode, secret string) bool {
	cleanPasscode := strings.TrimSpace(passcode)
	if len(cleanPasscode) != 6 {
		return false
	}

	// Valide avec une fenêtre de dérive de 1 période (±30 secondes) pour absorber
	// l'éventuelle dérive d'horloge d'un smartphone hors-ligne
	opts := totp.ValidateOpts{
		Period:    30,
		Skew:      1,
		Digits:    otp.DigitsSix,
		Algorithm: otp.AlgorithmSHA1,
	}

	valid, err := totp.ValidateCustom(cleanPasscode, secret, time.Now().UTC(), opts)
	return err == nil && valid
}

// ValidateAndConsumeBackupCode vérifie un code de secours et le retire de la liste si valide
func ValidateAndConsumeBackupCode(rawCode string, hashedCodes []string) (bool, []string) {
	cleanCode := strings.ToUpper(strings.TrimSpace(strings.ReplaceAll(rawCode, "-", "")))
	if cleanCode == "" {
		return false, hashedCodes
	}

	// Hachage SHA-256 du code saisi
	hasher := sha256.New()
	hasher.Write([]byte(cleanCode))
	candidateHash := hex.EncodeToString(hasher.Sum(nil))

	for i, expectedHash := range hashedCodes {
		if subtle.ConstantTimeCompare([]byte(candidateHash), []byte(expectedHash)) == 1 {
			// Code valide trouvé -> on le retire (usage unique)
			remaining := append(hashedCodes[:i], hashedCodes[i+1:]...)
			return true, remaining
		}
	}

	return false, hashedCodes
}

// generateBackupCodes génère N codes de secours alphanumériques (format XXXX-XXXX)
func generateBackupCodes(count int) (rawCodes []string, hashedCodes []string, err error) {
	const charset = "23456789ABCDEFGHJKLMNPQRSTUVWXYZ" // Évite les ambiguïtés visuelles (0, O, 1, I)

	rawCodes = make([]string, count)
	hashedCodes = make([]string, count)

	for i := 0; i < count; i++ {
		randomBytes := make([]byte, 8)
		if _, err := rand.Read(randomBytes); err != nil {
			return nil, nil, err
		}

		var sb strings.Builder
		for j, b := range randomBytes {
			if j == 4 {
				sb.WriteString("-")
			}
			idx := int(b) % len(charset)
			sb.WriteByte(charset[idx])
		}

		formattedCode := sb.String()
		rawCodes[i] = formattedCode

		// Stockage uniquement du SHA-256 du code normalisé (sans tiret)
		normalized := strings.ReplaceAll(formattedCode, "-", "")
		hasher := sha256.New()
		hasher.Write([]byte(normalized))
		hashedCodes[i] = hex.EncodeToString(hasher.Sum(nil))
	}

	return rawCodes, hashedCodes, nil
}
