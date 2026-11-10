package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	"golang.org/x/crypto/argon2"
)

// Erreurs de politique de mots de passe (Conformité BPF / 21 CFR Part 11)
var (
	ErrPasswordTooShort     = errors.New("le mot de passe doit comporter au moins 12 caractères")
	ErrPasswordTooLong      = errors.New("le mot de passe ne doit pas dépasser 128 caractères")
	ErrPasswordNoUpper      = errors.New("le mot de passe doit contenir au moins une lettre majuscule")
	ErrPasswordNoLower      = errors.New("le mot de passe doit contenir au moins une lettre minuscule")
	ErrPasswordNoDigit      = errors.New("le mot de passe doit contenir au moins un chiffre")
	ErrPasswordNoSpecial    = errors.New("le mot de passe doit contenir au moins un caractère spécial (@$!%*?&#...)")
	ErrInvalidHashFormat    = errors.New("format de hachage Argon2id invalide")
	ErrIncompatibleArgonVer = errors.New("version Argon2 non supportée")
)

// Paramètres de hachage recommandés par l'OWASP pour Argon2id
type Argon2Params struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultArgon2Params : Configuration équilibrée sécurité / performance enterprise
var DefaultArgon2Params = &Argon2Params{
	Memory:      64 * 1024, // 64 Mo
	Iterations:  3,
	Parallelism: 4,
	SaltLength:  16,
	KeyLength:   32,
}

// ValidatePassword vérifie la conformité du mot de passe aux règles de sécurité BPF
func ValidatePassword(password string) error {
	if len(password) < 12 {
		return ErrPasswordTooShort
	}
	if len(password) > 128 {
		return ErrPasswordTooLong
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool

	for _, ch := range password {
		switch {
		case unicode.IsUpper(ch):
			hasUpper = true
		case unicode.IsLower(ch):
			hasLower = true
		case unicode.IsDigit(ch):
			hasDigit = true
		case unicode.IsPunct(ch) || unicode.IsSymbol(ch) || strings.ContainsRune("@$!%*?&#_-+=~", ch):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return ErrPasswordNoUpper
	}
	if !hasLower {
		return ErrPasswordNoLower
	}
	if !hasDigit {
		return ErrPasswordNoDigit
	}
	if !hasSpecial {
		return ErrPasswordNoSpecial
	}

	return nil
}

// HashPassword calcule le hachage sécurisé d'un mot de passe via Argon2id
func HashPassword(password string, params *Argon2Params) (string, error) {
	if params == nil {
		params = DefaultArgon2Params
	}

	salt := make([]byte, params.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("impossible de générer le sel cryptographique: %w", err)
	}

	hash := argon2.IDKey([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLength)

	b64Salt := base64.RawStdEncoding.EncodeToString(salt)
	b64Hash := base64.RawStdEncoding.EncodeToString(hash)

	// Format standard : $argon2id$v=19$m=65536,t=3,p=4$<b64Salt>$<b64Hash>
	encoded := fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version, params.Memory, params.Iterations, params.Parallelism, b64Salt, b64Hash)

	return encoded, nil
}

// VerifyPassword vérifie un mot de passe en temps constant contre son empreinte Argon2id
func VerifyPassword(password, encodedHash string) (bool, error) {
	parts := strings.Split(encodedHash, "$")
	if len(parts) != 6 {
		return false, ErrInvalidHashFormat
	}

	if parts[1] != "argon2id" {
		return false, fmt.Errorf("algorithme non supporté: %s (seul argon2id est autorisé)", parts[1])
	}

	var version int
	_, err := fmt.Sscanf(parts[2], "v=%d", &version)
	if err != nil || version != argon2.Version {
		return false, ErrIncompatibleArgonVer
	}

	params := &Argon2Params{}
	_, err = fmt.Sscanf(parts[3], "m=%d,t=%d,p=%d", &params.Memory, &params.Iterations, &params.Parallelism)
	if err != nil {
		return false, ErrInvalidHashFormat
	}

	salt, err := base64.RawStdEncoding.DecodeString(parts[4])
	if err != nil {
		return false, ErrInvalidHashFormat
	}
	params.SaltLength = uint32(len(salt))

	expectedHash, err := base64.RawStdEncoding.DecodeString(parts[5])
	if err != nil {
		return false, ErrInvalidHashFormat
	}
	params.KeyLength = uint32(len(expectedHash))

	calculatedHash := argon2.IDKey([]byte(password), salt, params.Iterations, params.Memory, params.Parallelism, params.KeyLength)

	// Comparaison en temps constant pour immunité contre les attaques par canal auxiliaire (timing attacks)
	if subtle.ConstantTimeCompare(expectedHash, calculatedHash) == 1 {
		return true, nil
	}

	return false, nil
}

// GenerateSecureToken génère un jeton cryptographique aléatoire de n octets (hex-encodé)
func GenerateSecureToken(byteLength int) (string, error) {
	bytes := make([]byte, byteLength)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", bytes), nil
}

// Helper pour parser des entiers simples
func parseUint(s string) (uint64, error) {
	return strconv.ParseUint(s, 10, 64)
}
