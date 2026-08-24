package matching

// The two stopword lists are transcribed from
// docs/ingredient-matching-algorithm.md section 2, which is the source of
// truth. They are kept separate and named separately because they have
// different provenance and will change for different reasons: the measurement
// list was transcribed from this product's existing measurement vocabulary,
// the descriptor list is tuned against real ingredient lines.

// MeasurementStopwords are the tokens of the measurement vocabulary, split
// into single tokens because normalization has already tokenized by the time
// stopwords are applied.
var MeasurementStopwords = map[string]struct{}{
	// Volume
	"tsp": {}, "t": {}, "teaspoon": {}, "teaspoons": {},
	"tbsp": {}, "tablespoon": {}, "tablespoons": {}, "tbs": {},
	"fl": {}, "floz": {}, "fluid": {}, "ounce": {}, "ounces": {},
	"cup": {}, "cups": {}, "c": {},
	"pint": {}, "pints": {}, "pt": {},
	"quart": {}, "quarts": {}, "qt": {},
	"gallon": {}, "gallons": {}, "gal": {},
	"ml": {}, "milliliter": {}, "milliliters": {}, "millilitre": {}, "millilitres": {},
	"l": {}, "liter": {}, "liters": {}, "litre": {}, "litres": {},

	// Mass
	"g": {}, "gram": {}, "grams": {},
	"kg": {}, "kilogram": {}, "kilograms": {},
	"oz": {},
	"lb": {}, "lbs": {}, "pound": {}, "pounds": {},

	// Count
	"piece": {}, "pieces": {}, "whole": {},
	"clove": {}, "cloves": {},
	"can": {}, "cans": {}, "tin": {}, "tins": {},
	"slice": {}, "slices": {},

	// Special
	"pinch": {}, "pinches": {},
	"dash": {}, "dashes": {},
	"handful": {}, "handfuls": {},
	"to": {}, "taste": {},
}

// DescriptorStopwords are preparation, condition and size words that describe
// how a food arrives rather than what it is, plus the function words that
// appear inside ingredient lines.
var DescriptorStopwords = map[string]struct{}{
	// Preparation
	"chopped": {}, "diced": {}, "minced": {}, "sliced": {}, "shredded": {},
	"grated": {}, "crushed": {}, "ground": {}, "mashed": {}, "cubed": {},
	"julienned": {}, "halved": {}, "quartered": {}, "trimmed": {}, "peeled": {},
	"seeded": {}, "pitted": {}, "stemmed": {}, "rinsed": {}, "drained": {},
	"softened": {}, "melted": {}, "beaten": {}, "packed": {}, "sifted": {},
	"toasted": {}, "roasted": {}, "cooked": {}, "uncooked": {}, "raw": {},
	"boneless": {}, "skinless": {},

	// Condition and quality
	"fresh": {}, "freshly": {}, "frozen": {}, "dried": {}, "ripe": {},
	"unsalted": {}, "unsweetened": {}, "plain": {}, "low": {}, "reduced": {},
	"room": {}, "temperature": {}, "warm": {}, "cold": {}, "hot": {}, "boiling": {},

	// Size and amount
	"large": {}, "small": {}, "medium": {}, "extra": {}, "thin": {},
	"thick": {}, "finely": {}, "coarsely": {}, "thinly": {}, "thickly": {},
	"lightly": {}, "heaping": {}, "scant": {},

	// Line noise
	"optional": {}, "divided": {}, "plus": {}, "more": {}, "needed": {},
	"garnish": {}, "and": {}, "or": {}, "of": {}, "the": {}, "a": {},
	"an": {}, "for": {}, "with": {}, "into": {}, "in": {}, "on": {},
	"about": {}, "approximately": {},
}

func isStopword(token string) bool {
	if _, ok := MeasurementStopwords[token]; ok {
		return true
	}
	_, ok := DescriptorStopwords[token]
	return ok
}
