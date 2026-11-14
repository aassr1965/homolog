package domain_test

import (
	"strings"
	"testing"

	"homolog/internal/domain"
)

func TestConvertAmountDZDToWords(t *testing.T) {
	tests := []struct {
		amount     float64
		containsFR string
		containsAR string
	}{
		{
			amount:     350000.00,
			containsFR: "Trois cent cinquante mille dinars algériens",
			containsAR: "ثلاثمائة و خمسون ألف دينار جزائري",
		},
		{
			amount:     150000.00,
			containsFR: "Cent cinquante mille dinars algériens",
			containsAR: "مائة و خمسون ألف دينار جزائري",
		},
		{
			amount:     50000.00,
			containsFR: "Cinquante mille dinars algériens",
			containsAR: "خمسون ألف دينار جزائري",
		},
		{
			amount:     100000.00,
			containsFR: "Cent mille dinars algériens",
			containsAR: "مائة ألف دينار جزائري",
		},
	}

	for _, tc := range tests {
		fr, ar := domain.ConvertAmountDZDToWords(tc.amount)
		if !strings.Contains(strings.ToLower(fr), strings.ToLower(tc.containsFR)) {
			t.Errorf("Montant %.2f FR incorrect: obtenu '%s', attendu contient '%s'", tc.amount, fr, tc.containsFR)
		}
		if !strings.Contains(ar, "دينار جزائري") {
			t.Errorf("Montant %.2f AR doit contenir 'دينار جزائري', obtenu '%s'", tc.amount, ar)
		}
	}
}
