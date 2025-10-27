package api

import (
	"strings"
)

// IrregularVerb represents an irregular English verb with all its forms
type IrregularVerb struct {
	Base       string
	Past       string
	PastPart   string
	Present3rd string
	Gerund     string
}

// irregularVerbs is a lookup table for common irregular English verbs
var irregularVerbs = map[string]IrregularVerb{
	"be":         {Base: "be", Past: "was/were", PastPart: "been", Present3rd: "is", Gerund: "being"},
	"become":     {Base: "become", Past: "became", PastPart: "become", Present3rd: "becomes", Gerund: "becoming"},
	"begin":      {Base: "begin", Past: "began", PastPart: "begun", Present3rd: "begins", Gerund: "beginning"},
	"break":      {Base: "break", Past: "broke", PastPart: "broken", Present3rd: "breaks", Gerund: "breaking"},
	"bring":      {Base: "bring", Past: "brought", PastPart: "brought", Present3rd: "brings", Gerund: "bringing"},
	"build":      {Base: "build", Past: "built", PastPart: "built", Present3rd: "builds", Gerund: "building"},
	"buy":        {Base: "buy", Past: "bought", PastPart: "bought", Present3rd: "buys", Gerund: "buying"},
	"catch":      {Base: "catch", Past: "caught", PastPart: "caught", Present3rd: "catches", Gerund: "catching"},
	"choose":     {Base: "choose", Past: "chose", PastPart: "chosen", Present3rd: "chooses", Gerund: "choosing"},
	"come":       {Base: "come", Past: "came", PastPart: "come", Present3rd: "comes", Gerund: "coming"},
	"cost":       {Base: "cost", Past: "cost", PastPart: "cost", Present3rd: "costs", Gerund: "costing"},
	"cut":        {Base: "cut", Past: "cut", PastPart: "cut", Present3rd: "cuts", Gerund: "cutting"},
	"do":         {Base: "do", Past: "did", PastPart: "done", Present3rd: "does", Gerund: "doing"},
	"draw":       {Base: "draw", Past: "drew", PastPart: "drawn", Present3rd: "draws", Gerund: "drawing"},
	"drink":      {Base: "drink", Past: "drank", PastPart: "drunk", Present3rd: "drinks", Gerund: "drinking"},
	"drive":      {Base: "drive", Past: "drove", PastPart: "driven", Present3rd: "drives", Gerund: "driving"},
	"eat":        {Base: "eat", Past: "ate", PastPart: "eaten", Present3rd: "eats", Gerund: "eating"},
	"fall":       {Base: "fall", Past: "fell", PastPart: "fallen", Present3rd: "falls", Gerund: "falling"},
	"feel":       {Base: "feel", Past: "felt", PastPart: "felt", Present3rd: "feels", Gerund: "feeling"},
	"find":       {Base: "find", Past: "found", PastPart: "found", Present3rd: "finds", Gerund: "finding"},
	"fly":        {Base: "fly", Past: "flew", PastPart: "flown", Present3rd: "flies", Gerund: "flying"},
	"forget":     {Base: "forget", Past: "forgot", PastPart: "forgotten", Present3rd: "forgets", Gerund: "forgetting"},
	"get":        {Base: "get", Past: "got", PastPart: "gotten", Present3rd: "gets", Gerund: "getting"},
	"give":       {Base: "give", Past: "gave", PastPart: "given", Present3rd: "gives", Gerund: "giving"},
	"go":         {Base: "go", Past: "went", PastPart: "gone", Present3rd: "goes", Gerund: "going"},
	"grow":       {Base: "grow", Past: "grew", PastPart: "grown", Present3rd: "grows", Gerund: "growing"},
	"have":       {Base: "have", Past: "had", PastPart: "had", Present3rd: "has", Gerund: "having"},
	"hear":       {Base: "hear", Past: "heard", PastPart: "heard", Present3rd: "hears", Gerund: "hearing"},
	"hide":       {Base: "hide", Past: "hid", PastPart: "hidden", Present3rd: "hides", Gerund: "hiding"},
	"hit":        {Base: "hit", Past: "hit", PastPart: "hit", Present3rd: "hits", Gerund: "hitting"},
	"hold":       {Base: "hold", Past: "held", PastPart: "held", Present3rd: "holds", Gerund: "holding"},
	"keep":       {Base: "keep", Past: "kept", PastPart: "kept", Present3rd: "keeps", Gerund: "keeping"},
	"know":       {Base: "know", Past: "knew", PastPart: "known", Present3rd: "knows", Gerund: "knowing"},
	"leave":      {Base: "leave", Past: "left", PastPart: "left", Present3rd: "leaves", Gerund: "leaving"},
	"lend":       {Base: "lend", Past: "lent", PastPart: "lent", Present3rd: "lends", Gerund: "lending"},
	"let":        {Base: "let", Past: "let", PastPart: "let", Present3rd: "lets", Gerund: "letting"},
	"lose":       {Base: "lose", Past: "lost", PastPart: "lost", Present3rd: "loses", Gerund: "losing"},
	"make":       {Base: "make", Past: "made", PastPart: "made", Present3rd: "makes", Gerund: "making"},
	"mean":       {Base: "mean", Past: "meant", PastPart: "meant", Present3rd: "means", Gerund: "meaning"},
	"meet":       {Base: "meet", Past: "met", PastPart: "met", Present3rd: "meets", Gerund: "meeting"},
	"pay":        {Base: "pay", Past: "paid", PastPart: "paid", Present3rd: "pays", Gerund: "paying"},
	"put":        {Base: "put", Past: "put", PastPart: "put", Present3rd: "puts", Gerund: "putting"},
	"read":       {Base: "read", Past: "read", PastPart: "read", Present3rd: "reads", Gerund: "reading"},
	"ride":       {Base: "ride", Past: "rode", PastPart: "ridden", Present3rd: "rides", Gerund: "riding"},
	"ring":       {Base: "ring", Past: "rang", PastPart: "rung", Present3rd: "rings", Gerund: "ringing"},
	"rise":       {Base: "rise", Past: "rose", PastPart: "risen", Present3rd: "rises", Gerund: "rising"},
	"run":        {Base: "run", Past: "ran", PastPart: "run", Present3rd: "runs", Gerund: "running"},
	"say":        {Base: "say", Past: "said", PastPart: "said", Present3rd: "says", Gerund: "saying"},
	"see":        {Base: "see", Past: "saw", PastPart: "seen", Present3rd: "sees", Gerund: "seeing"},
	"sell":       {Base: "sell", Past: "sold", PastPart: "sold", Present3rd: "sells", Gerund: "selling"},
	"send":       {Base: "send", Past: "sent", PastPart: "sent", Present3rd: "sends", Gerund: "sending"},
	"set":        {Base: "set", Past: "set", PastPart: "set", Present3rd: "sets", Gerund: "setting"},
	"show":       {Base: "show", Past: "showed", PastPart: "shown", Present3rd: "shows", Gerund: "showing"},
	"shut":       {Base: "shut", Past: "shut", PastPart: "shut", Present3rd: "shuts", Gerund: "shutting"},
	"sing":       {Base: "sing", Past: "sang", PastPart: "sung", Present3rd: "sings", Gerund: "singing"},
	"sit":        {Base: "sit", Past: "sat", PastPart: "sat", Present3rd: "sits", Gerund: "sitting"},
	"sleep":      {Base: "sleep", Past: "slept", PastPart: "slept", Present3rd: "sleeps", Gerund: "sleeping"},
	"speak":      {Base: "speak", Past: "spoke", PastPart: "spoken", Present3rd: "speaks", Gerund: "speaking"},
	"spend":      {Base: "spend", Past: "spent", PastPart: "spent", Present3rd: "spends", Gerund: "spending"},
	"stand":      {Base: "stand", Past: "stood", PastPart: "stood", Present3rd: "stands", Gerund: "standing"},
	"swim":       {Base: "swim", Past: "swam", PastPart: "swum", Present3rd: "swims", Gerund: "swimming"},
	"take":       {Base: "take", Past: "took", PastPart: "taken", Present3rd: "takes", Gerund: "taking"},
	"teach":      {Base: "teach", Past: "taught", PastPart: "taught", Present3rd: "teaches", Gerund: "teaching"},
	"tear":       {Base: "tear", Past: "tore", PastPart: "torn", Present3rd: "tears", Gerund: "tearing"},
	"tell":       {Base: "tell", Past: "told", PastPart: "told", Present3rd: "tells", Gerund: "telling"},
	"think":      {Base: "think", Past: "thought", PastPart: "thought", Present3rd: "thinks", Gerund: "thinking"},
	"throw":      {Base: "throw", Past: "threw", PastPart: "thrown", Present3rd: "throws", Gerund: "throwing"},
	"understand": {Base: "understand", Past: "understood", PastPart: "understood", Present3rd: "understands", Gerund: "understanding"},
	"wake":       {Base: "wake", Past: "woke", PastPart: "woken", Present3rd: "wakes", Gerund: "waking"},
	"wear":       {Base: "wear", Past: "wore", PastPart: "worn", Present3rd: "wears", Gerund: "wearing"},
	"win":        {Base: "win", Past: "won", PastPart: "won", Present3rd: "wins", Gerund: "winning"},
	"write":      {Base: "write", Past: "wrote", PastPart: "written", Present3rd: "writes", Gerund: "writing"},
}

// EnglishConjugator handles English verb conjugation
type EnglishConjugator struct {
	baseVerb    string
	isIrregular bool
	irregular   IrregularVerb
}

// NewEnglishConjugator creates a new English conjugator from a definition string
func NewEnglishConjugator(definition string) *EnglishConjugator {
	baseVerb := extractBaseVerb(definition)
	if baseVerb == "" {
		return nil
	}

	ec := &EnglishConjugator{
		baseVerb: baseVerb,
	}

	// Check if it's irregular
	if irregular, exists := irregularVerbs[baseVerb]; exists {
		ec.isIrregular = true
		ec.irregular = irregular
	}

	return ec
}

// extractBaseVerb extracts the base verb form from a definition
func extractBaseVerb(definition string) string {
	if definition == "" {
		return ""
	}

	// Split by semicolon and take the first definition
	parts := strings.Split(definition, ";")
	firstDef := strings.TrimSpace(parts[0])

	// Remove "to " prefix if present
	firstDef = strings.TrimPrefix(firstDef, "to ")
	firstDef = strings.TrimSpace(firstDef)

	// Remove parenthetical content
	if idx := strings.Index(firstDef, "("); idx != -1 {
		firstDef = strings.TrimSpace(firstDef[:idx])
	}

	// Remove common non-verb prefixes
	firstDef = strings.TrimPrefix(firstDef, "be ")
	firstDef = strings.TrimSpace(firstDef)

	// Take only the first word (base verb)
	words := strings.Fields(firstDef)
	if len(words) > 0 {
		baseVerb := strings.ToLower(words[0])
		// Only return if it looks like a verb (not a noun phrase, etc.)
		if baseVerb != "" && !strings.HasPrefix(baseVerb, "a") && !strings.HasPrefix(baseVerb, "the") {
			return baseVerb
		}
	}

	return ""
}

// GetPastSimple returns the simple past form
func (ec *EnglishConjugator) GetPastSimple() string {
	if ec.isIrregular {
		return ec.irregular.Past
	}
	return ec.addRegularPastEnding()
}

// GetPastParticiple returns the past participle form
func (ec *EnglishConjugator) GetPastParticiple() string {
	if ec.isIrregular {
		return ec.irregular.PastPart
	}
	return ec.addRegularPastEnding()
}

// GetPresent3rdPerson returns the 3rd person singular present form
func (ec *EnglishConjugator) GetPresent3rdPerson() string {
	if ec.isIrregular {
		return ec.irregular.Present3rd
	}
	return ec.addRegularPresentEnding()
}

// GetGerund returns the gerund/present participle form (-ing)
func (ec *EnglishConjugator) GetGerund() string {
	if ec.isIrregular {
		return ec.irregular.Gerund
	}
	return ec.addRegularGerundEnding()
}

// GetBase returns the base form
func (ec *EnglishConjugator) GetBase() string {
	return ec.baseVerb
}

// addRegularPastEnding adds -ed to regular verbs
func (ec *EnglishConjugator) addRegularPastEnding() string {
	verb := ec.baseVerb

	// If ends in 'e', just add 'd'
	if strings.HasSuffix(verb, "e") {
		return verb + "d"
	}

	// If ends in consonant + y, change y to i and add ed
	if len(verb) >= 2 && strings.HasSuffix(verb, "y") {
		secondLast := verb[len(verb)-2]
		if !isVowel(secondLast) {
			return verb[:len(verb)-1] + "ied"
		}
	}

	// If ends in single vowel + consonant, double the consonant
	if len(verb) >= 2 {
		lastChar := verb[len(verb)-1]
		secondLast := verb[len(verb)-2]
		if !isVowel(lastChar) && isVowel(secondLast) && shouldDoubleConsonant(verb) {
			return verb + string(lastChar) + "ed"
		}
	}

	// Default: just add 'ed'
	return verb + "ed"
}

// addRegularPresentEnding adds -s/-es for 3rd person singular
func (ec *EnglishConjugator) addRegularPresentEnding() string {
	verb := ec.baseVerb

	// If ends in s, z, x, ch, sh, add 'es'
	if strings.HasSuffix(verb, "s") || strings.HasSuffix(verb, "z") ||
		strings.HasSuffix(verb, "x") || strings.HasSuffix(verb, "ch") ||
		strings.HasSuffix(verb, "sh") {
		return verb + "es"
	}

	// If ends in consonant + y, change y to ies
	if len(verb) >= 2 && strings.HasSuffix(verb, "y") {
		secondLast := verb[len(verb)-2]
		if !isVowel(secondLast) {
			return verb[:len(verb)-1] + "ies"
		}
	}

	// If ends in consonant + o, add 'es'
	if len(verb) >= 2 && strings.HasSuffix(verb, "o") {
		secondLast := verb[len(verb)-2]
		if !isVowel(secondLast) {
			return verb + "es"
		}
	}

	// Default: just add 's'
	return verb + "s"
}

// addRegularGerundEnding adds -ing for gerund/present participle
func (ec *EnglishConjugator) addRegularGerundEnding() string {
	verb := ec.baseVerb

	// If ends in 'ie', change to 'ying'
	if strings.HasSuffix(verb, "ie") {
		return verb[:len(verb)-2] + "ying"
	}

	// If ends in 'e' (but not 'ee' or 'ye' or 'oe'), drop the 'e'
	if strings.HasSuffix(verb, "e") && !strings.HasSuffix(verb, "ee") &&
		!strings.HasSuffix(verb, "ye") && !strings.HasSuffix(verb, "oe") {
		return verb[:len(verb)-1] + "ing"
	}

	// If ends in single vowel + consonant, double the consonant
	if len(verb) >= 2 {
		lastChar := verb[len(verb)-1]
		secondLast := verb[len(verb)-2]
		if !isVowel(lastChar) && isVowel(secondLast) && shouldDoubleConsonant(verb) {
			return verb + string(lastChar) + "ing"
		}
	}

	// Default: just add 'ing'
	return verb + "ing"
}

// isVowel checks if a byte is a vowel
func isVowel(c byte) bool {
	c = c | 0x20 // Convert to lowercase
	return c == 'a' || c == 'e' || c == 'i' || c == 'o' || c == 'u'
}

// shouldDoubleConsonant determines if the final consonant should be doubled
func shouldDoubleConsonant(verb string) bool {
	if len(verb) < 2 {
		return false
	}

	lastChar := verb[len(verb)-1]

	// Don't double w, x, y
	if lastChar == 'w' || lastChar == 'x' || lastChar == 'y' {
		return false
	}

	// For one-syllable words or stressed final syllable, double
	// This is a simplified rule - proper implementation would need syllable counting
	// For now, we'll double for short words (3-4 letters) ending in single vowel + consonant
	if len(verb) <= 4 {
		return true
	}

	return false
}

// ConjugateEnglishForTense conjugates English verb for specific tense/mood/aspect
func (ec *EnglishConjugator) ConjugateEnglishForTense(category, form string) string {
	if ec == nil {
		return ""
	}

	base := ec.GetBase()
	past := ec.GetPastSimple()
	pastPart := ec.GetPastParticiple()
	gerund := ec.GetGerund()

	switch category {
	case "time":
		switch form {
		case "present":
			return "I " + base
		case "past":
			return "I " + past
		case "future":
			return "I will " + base
		}
	case "aspect":
		switch form {
		case "simple":
			return "I " + base
		case "progressive":
			return "I am " + gerund
		case "perfect":
			return "I have " + pastPart
		case "perfect_progressive":
			return "I have been " + gerund
		}
	case "mood":
		switch form {
		case "indicative":
			return "I " + base
		case "subjunctive":
			return "I wish I " + past
		case "conditional":
			return "If I " + base
		case "imperative":
			return strings.Title(base)
		case "volitional":
			return "Let's " + base
		}
	case "modals":
		switch form {
		case "potential":
			return "I can " + base
		case "causative":
			return "I make you " + base
		case "deontic":
			return "I must " + base
		}
	case "desire":
		switch form {
		case "subject":
			return "I want to " + base
		}
	case "voice":
		switch form {
		case "active":
			return "I " + base
		case "passive":
			return "It is " + pastPart + " by me"
		}
	}

	return ""
}
