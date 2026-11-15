package domain_test

import (
	"testing"

	"github.com/google/uuid"
	"homolog/internal/domain"
)

func TestWorkflowCheckoutAndCheckinValidation(t *testing.T) {
	u1 := uuid.New()
	u2 := uuid.New()

	t.Run("Checkout sur item non verrouillé", func(t *testing.T) {
		err := domain.ValidateCheckout(false, nil, u1)
		if err != nil {
			t.Fatalf("checkout devrait être autorisé, erreur: %v", err)
		}
	})

	t.Run("Checkout par le même utilisateur qui détient le verrou", func(t *testing.T) {
		err := domain.ValidateCheckout(true, &u1, u1)
		if err != nil {
			t.Fatalf("checkout par le même utilisateur devrait être permis: %v", err)
		}
	})

	t.Run("Checkout conflictuel par un tiers", func(t *testing.T) {
		err := domain.ValidateCheckout(true, &u1, u2)
		if err == nil {
			t.Fatal("checkout par un tiers sur item verrouillé devrait être rejeté")
		}
	})

	t.Run("Checkin sur item non verrouillé", func(t *testing.T) {
		err := domain.ValidateCheckin(false, nil, u1)
		if err == nil {
			t.Fatal("checkin sur item non verrouillé devrait échouer")
		}
	})

	t.Run("Checkin par un utilisateur non détenteur du verrou", func(t *testing.T) {
		err := domain.ValidateCheckin(true, &u1, u2)
		if err == nil {
			t.Fatal("checkin par un tiers non détenteur du verrou devrait échouer")
		}
	})

	t.Run("Checkin légitime par le détenteur du verrou", func(t *testing.T) {
		err := domain.ValidateCheckin(true, &u1, u1)
		if err != nil {
			t.Fatalf("checkin par le détenteur du verrou devrait réussir: %v", err)
		}
	})
}

func TestNextMinorRevision(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", "v0.1"},
		{"v0.1", "v0.2"},
		{"v0.2", "v0.3"},
		{"v0.9", "v0.10"},
		{"v0.10", "v0.11"},
		{"invalid", "v0.1"},
	}

	for _, tc := range tests {
		got := domain.NextMinorRevision(tc.input)
		if got != tc.expected {
			t.Errorf("NextMinorRevision(%q) = %q, attendu %q", tc.input, got, tc.expected)
		}
	}
}
