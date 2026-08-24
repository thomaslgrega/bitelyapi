package matching

import (
	"testing"
)

// Rows transcribed from docs/ingredient-matching-algorithm.md section 1.
func TestNormalize(t *testing.T) {
	tests := []struct {
		raw    string
		tokens []string
	}{
		{"Tomatoes", []string{"tomatoes"}},
		{"  ToMaTo  ", []string{"tomato"}},
		{"2 Yellow Onions", []string{"onions", "yellow"}},
		{"Boneless, skinless chicken breasts", []string{"breasts", "chicken"}},
		{"1 1/2 cups all-purpose flour", []string{"all", "flour", "purpose"}},
		{"Salt & pepper, to taste", []string{"pepper", "salt"}},
		{"freshly chopped", nil},
		{"", nil},
		{"   ", nil},
		{",.-/()", nil},
		{"2 1/2", nil},
		{"to taste", nil},
		{"½ cup ⅓ milk", []string{"milk"}},
		{"500g Beef", []string{"beef"}},
		{"fl.oz olive oil", []string{"oil", "olive"}},
		{"Onion onion ONION", []string{"onion"}},
	}

	for _, test := range tests {
		t.Run(test.raw, func(t *testing.T) {
			if got := Normalize(test.raw).Tokens(); !sameTokens(got, test.tokens) {
				t.Fatalf("Normalize(%q) = %v, want %v", test.raw, got, test.tokens)
			}
		})
	}
}

func TestTermIsEmpty(t *testing.T) {
	if !Normalize("freshly chopped").IsEmpty() {
		t.Fatal("a term of only stopwords should be empty")
	}
	if Normalize("tomato").IsEmpty() {
		t.Fatal("a term with a token should not be empty")
	}
}
