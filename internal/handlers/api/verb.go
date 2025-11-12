package api

import (
	"encoding/json"
	"fmt"
	"gaijin/internal/database"
	"net/http"
	"strings"
	"unicode/utf8"
)

type VerbHandler struct {
	db *database.Database
}

func NewVerbHandler(db *database.Database) *VerbHandler {
	return &VerbHandler{db: db}
}

// VerbType represents the type of Japanese verb
type VerbType int

const (
	Ichidan VerbType = iota
	Godan
	Irregular
)

// ConjugationRequest represents the incoming request
type ConjugationRequest struct {
	Verb     string `json:"verb"`
	Negative bool   `json:"negative"`
	Polite   bool   `json:"polite"`
}

// ConjugationResponse represents the full conjugation response
type ConjugationResponse struct {
	Valid        bool                   `json:"valid"`
	Error        string                 `json:"error,omitempty"`
	Verb         string                 `json:"verb"`
	VerbType     string                 `json:"verbType"`
	Conjugations map[string]interface{} `json:"conjugations"`
}

// ConjugationEntry represents a single conjugation
type ConjugationEntry struct {
	English  string   `json:"english"`
	Japanese string   `json:"japanese"`
	Alts     []string `json:"alts"`
}

// HandleConjugate handles verb conjugation requests
func (h *VerbHandler) HandleConjugate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req ConjugationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	verb := strings.TrimSpace(req.Verb)
	if verb == "" {
		json.NewEncoder(w).Encode(ConjugationResponse{
			Valid: false,
			Error: "Verb cannot be empty",
		})
		return
	}

	// Validate that input is Japanese
	if !isJapanese(verb) {
		json.NewEncoder(w).Encode(ConjugationResponse{
			Valid: false,
			Error: "Input must be in Japanese",
		})
		return
	}

	// Check if word exists in database and is a verb, get definition
	isValid, definition, err := h.validateVerbAndGetDefinition(verb)
	if err != nil {
		json.NewEncoder(w).Encode(ConjugationResponse{
			Valid: false,
			Error: "Failed to validate verb: " + err.Error(),
		})
		return
	}

	if !isValid {
		json.NewEncoder(w).Encode(ConjugationResponse{
			Valid: false,
			Error: "Word not found in database or is not a verb",
		})
		return
	}

	// Create English conjugator from definition
	engConjugator := NewEnglishConjugator(definition)

	// Debug logging
	if definition != "" {
		fmt.Printf("DEBUG: Verb '%s' has definition: '%s'\n", verb, definition)
		if engConjugator != nil {
			fmt.Printf("DEBUG: Extracted base verb: '%s'\n", engConjugator.GetBase())
		} else {
			fmt.Printf("DEBUG: Failed to create conjugator (engConjugator is nil)\n")
		}
	} else {
		fmt.Printf("DEBUG: Verb '%s' has no definition in database\n", verb)
	}

	// Determine verb type and conjugate
	verbType, verbTypeStr := determineVerbType(verb)
	conjugations := conjugateVerb(verb, verbType, engConjugator)

	// Apply negation and/or politeness transformations
	if req.Negative || req.Polite {
		conjugations = applyModifiers(conjugations, verb, verbType, req.Negative, req.Polite, engConjugator)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(ConjugationResponse{
		Valid:        true,
		Verb:         verb,
		VerbType:     verbTypeStr,
		Conjugations: conjugations,
	})
}

// validateVerbAndGetDefinition checks if the word exists in the database and is a verb, and returns its definition
func (h *VerbHandler) validateVerbAndGetDefinition(verb string) (bool, string, error) {
	if h.db == nil || h.db.DB == nil {
		// Database not available, but allow typical verb endings
		if endsWithVerbEnding(verb) {
			return true, "", nil
		}
		return false, "", fmt.Errorf("database not available")
	}

	query := `
		SELECT parts_of_speech, definitions FROM words 
		WHERE word = $1 OR furigana = $1
		LIMIT 1
	`
	var partsOfSpeech, definitions string
	err := h.db.DB.QueryRow(query, verb).Scan(&partsOfSpeech, &definitions)
	if err != nil {
		// Word not found - for now, we'll allow any Japanese input ending in る、う、く、ぐ、す、つ、ぬ、ぶ、む
		// This allows users to test with verbs not in our database
		if endsWithVerbEnding(verb) {
			return true, "", nil
		}
		return false, "", nil
	}

	// Check if parts of speech contains "verb"
	isVerb := strings.Contains(strings.ToLower(partsOfSpeech), "verb")
	return isVerb, definitions, nil
}

// endsWithVerbEnding checks if the word ends with a typical verb ending
func endsWithVerbEnding(verb string) bool {
	if len(verb) == 0 {
		return false
	}

	lastChar := getLastRune(verb)
	// Common verb endings in hiragana
	verbEndings := []rune{'る', 'う', 'く', 'ぐ', 'す', 'つ', 'ぬ', 'ぶ', 'む'}

	for _, ending := range verbEndings {
		if lastChar == ending {
			return true
		}
	}
	return false
}

// isJapanese checks if the string contains Japanese characters
func isJapanese(s string) bool {
	for _, r := range s {
		if (r >= 0x3040 && r <= 0x309F) || // Hiragana
			(r >= 0x30A0 && r <= 0x30FF) || // Katakana
			(r >= 0x4E00 && r <= 0x9FAF) { // Kanji
			return true
		}
	}
	return false
}

// getLastRune returns the last rune in a string
func getLastRune(s string) rune {
	if s == "" {
		return 0
	}
	r, _ := utf8.DecodeLastRuneInString(s)
	return r
}

// getStem removes the last character from a verb to get the stem
func getStem(verb string) string {
	if verb == "" {
		return ""
	}
	runes := []rune(verb)
	if len(runes) == 0 {
		return ""
	}
	return string(runes[:len(runes)-1])
}

// determineVerbType determines if a verb is ichidan, godan, or irregular
func determineVerbType(verb string) (VerbType, string) {
	// Handle irregular verbs
	if verb == "する" || verb == "為る" {
		return Irregular, "irregular (する)"
	}
	if verb == "来る" || verb == "くる" {
		return Irregular, "irregular (来る)"
	}

	// Check for -eru or -iru ending (potential ichidan)
	if len(verb) < 2 {
		return Godan, "godan"
	}

	runes := []rune(verb)
	lastChar := runes[len(runes)-1]

	if lastChar != 'る' {
		return Godan, "godan"
	}

	// If ends in る, check the second-to-last character
	if len(runes) >= 2 {
		secondLast := runes[len(runes)-2]
		// Ichidan verbs end in -eru (え段 + る) or -iru (い段 + る)
		ichidanVowels := []rune{'い', 'き', 'ぎ', 'し', 'じ', 'ち', 'に', 'ひ', 'び', 'ぴ', 'み', 'り',
			'え', 'け', 'げ', 'せ', 'ぜ', 'て', 'で', 'ね', 'へ', 'べ', 'ぺ', 'め', 'れ'}

		for _, vowel := range ichidanVowels {
			if secondLast == vowel {
				return Ichidan, "ichidan"
			}
		}
	}

	return Godan, "godan"
}

// conjugateVerb generates all conjugations for a verb
func conjugateVerb(verb string, verbType VerbType, engConjugator *EnglishConjugator) map[string]interface{} {
	conjugations := make(map[string]interface{})

	// Handle irregular verbs separately
	if verbType == Irregular {
		conjugations["tenses"] = conjugateIrregular(verb, engConjugator)
		conjugations["voice"] = voiceIrregular(verb, engConjugator)
		return conjugations
	}

	stem := getStem(verb)
	lastChar := getLastRune(verb)

	// Time tenses
	time := map[string]ConjugationEntry{
		"present": {
			English:  getEnglishForForm(engConjugator, "time", "present"),
			Japanese: verb,
			Alts:     []string{},
		},
		"past": {
			English:  getEnglishForForm(engConjugator, "time", "past"),
			Japanese: conjugatePast(verb, stem, lastChar, verbType),
			Alts:     []string{},
		},
		"future": {
			English:  getEnglishForForm(engConjugator, "time", "future"),
			Japanese: verb,
			Alts:     []string{},
		},
	}

	// Aspect
	aspect := map[string]ConjugationEntry{
		"simple": {
			English:  getEnglishForForm(engConjugator, "aspect", "simple"),
			Japanese: verb,
			Alts:     []string{},
		},
		"progressive": {
			English:  getEnglishForForm(engConjugator, "aspect", "progressive"),
			Japanese: conjugateTe(verb, stem, lastChar, verbType) + "いる",
			Alts:     []string{},
		},
		"perfect": {
			English:  getEnglishForForm(engConjugator, "aspect", "perfect"),
			Japanese: conjugatePast(verb, stem, lastChar, verbType) + "ばかり",
			Alts:     []string{},
		},
		"perfect_progressive": {
			English:  getEnglishForForm(engConjugator, "aspect", "perfect_progressive"),
			Japanese: conjugateTe(verb, stem, lastChar, verbType) + "いる",
			Alts:     []string{},
		},
	}

	// Mood
	mood := map[string]ConjugationEntry{
		"indicative": {
			English:  getEnglishForForm(engConjugator, "mood", "indicative"),
			Japanese: verb,
			Alts:     []string{},
		},
		"subjunctive": {
			English:  getEnglishForForm(engConjugator, "mood", "subjunctive"),
			Japanese: conjugatePast(verb, stem, lastChar, verbType) + "らいいのに",
			Alts:     []string{},
		},
		"conditional": {
			English:  getEnglishForForm(engConjugator, "mood", "conditional"),
			Japanese: conjugateConditional(verb, stem, lastChar, verbType),
			Alts:     []string{conjugatePast(verb, stem, lastChar, verbType) + "ら"},
		},
		"imperative": {
			English:  getEnglishForForm(engConjugator, "mood", "imperative"),
			Japanese: conjugateImperative(verb, stem, lastChar, verbType),
			Alts:     []string{},
		},
		"volitional": {
			English:  getEnglishForForm(engConjugator, "mood", "volitional"),
			Japanese: conjugateVolitional(verb, stem, lastChar, verbType),
			Alts:     []string{},
		},
	}

	// Modals
	modals := map[string]ConjugationEntry{
		"potential": {
			English:  getEnglishForForm(engConjugator, "modals", "potential"),
			Japanese: conjugatePotential(verb, stem, lastChar, verbType),
			Alts:     []string{},
		},
		"causative": {
			English:  getEnglishForForm(engConjugator, "modals", "causative"),
			Japanese: conjugateCausative(verb, stem, lastChar, verbType),
			Alts:     []string{},
		},
		"deontic": {
			English:  getEnglishForForm(engConjugator, "modals", "deontic"),
			Japanese: conjugateNegative(verb, stem, lastChar, verbType) + "ければならない",
			Alts:     []string{},
		},
	}

	// Desire
	desire := map[string]ConjugationEntry{
		"subject": {
			English:  getEnglishForForm(engConjugator, "desire", "subject"),
			Japanese: conjugateMasu(verb, stem, lastChar, verbType) + "たい",
			Alts:     []string{},
		},
	}

	// Voice
	voice := map[string]ConjugationEntry{
		"active": {
			English:  getEnglishForForm(engConjugator, "voice", "active"),
			Japanese: verb,
			Alts:     []string{},
		},
		"passive": {
			English:  getEnglishForForm(engConjugator, "voice", "passive"),
			Japanese: conjugatePassive(verb, stem, lastChar, verbType),
			Alts:     []string{},
		},
	}

	conjugations["tenses"] = map[string]interface{}{
		"time":   time,
		"aspect": aspect,
		"mood":   mood,
		"modals": modals,
		"desire": desire,
	}
	conjugations["voice"] = voice

	return conjugations
}

// Helper functions for conjugations

func getEnglishForForm(engConjugator *EnglishConjugator, category, form string) string {
	if engConjugator == nil {
		// Fallback for when we don't have a definition
		return getFallbackEnglish(category, form)
	}
	return engConjugator.ConjugateEnglishForTense(category, form)
}

func getFallbackEnglish(category, form string) string {
	// Fallback English phrases when we don't have a definition
	fallbacks := map[string]map[string]string{
		"time": {
			"present": "I do",
			"past":    "I did",
			"future":  "I will do",
		},
		"aspect": {
			"simple":              "I do",
			"progressive":         "I am doing",
			"perfect":             "I have done",
			"perfect_progressive": "I have been doing",
		},
		"mood": {
			"indicative":  "I do",
			"subjunctive": "I wish I did",
			"conditional": "If I do",
			"imperative":  "Do",
			"volitional":  "Let's do",
		},
		"modals": {
			"potential": "I can do",
			"causative": "I make you do",
			"deontic":   "I must do",
		},
		"desire": {
			"subject": "I want to do",
		},
		"voice": {
			"active":  "I do",
			"passive": "It is done by me",
		},
	}

	if cat, ok := fallbacks[category]; ok {
		if phrase, ok := cat[form]; ok {
			return phrase
		}
	}
	return "verb"
}

func conjugateMasu(verb, stem string, lastChar rune, verbType VerbType) string {
	if verbType == Ichidan {
		return stem
	}
	// Godan: convert to い-stem
	return stem + string(godanToI(lastChar))
}

func conjugatePast(verb, stem string, lastChar rune, verbType VerbType) string {
	if verbType == Ichidan {
		return stem + "た"
	}
	// Godan past tense
	switch lastChar {
	case 'う', 'つ', 'る':
		return stem + "った"
	case 'く':
		return stem + "いた"
	case 'ぐ':
		return stem + "いだ"
	case 'す':
		return stem + "した"
	case 'ぬ', 'ぶ', 'む':
		return stem + "んだ"
	default:
		return stem + "た"
	}
}

func conjugateTe(verb, stem string, lastChar rune, verbType VerbType) string {
	if verbType == Ichidan {
		return stem + "て"
	}
	// Godan te-form
	switch lastChar {
	case 'う', 'つ', 'る':
		return stem + "って"
	case 'く':
		return stem + "いて"
	case 'ぐ':
		return stem + "いで"
	case 'す':
		return stem + "して"
	case 'ぬ', 'ぶ', 'む':
		return stem + "んで"
	default:
		return stem + "て"
	}
}

func conjugateNegative(verb, stem string, lastChar rune, verbType VerbType) string {
	if verbType == Ichidan {
		return stem + "な"
	}
	// Godan: convert to あ-stem
	return stem + string(godanToA(lastChar))
}

func conjugateConditional(verb, stem string, lastChar rune, verbType VerbType) string {
	if verbType == Ichidan {
		return stem + "れば"
	}
	// Godan: convert to え-stem + ば
	return stem + string(godanToE(lastChar)) + "ば"
}

func conjugateImperative(verb, stem string, lastChar rune, verbType VerbType) string {
	if verbType == Ichidan {
		return stem + "ろ"
	}
	// Godan: convert to え-stem
	return stem + string(godanToE(lastChar))
}

func conjugateVolitional(verb, stem string, lastChar rune, verbType VerbType) string {
	if verbType == Ichidan {
		return stem + "よう"
	}
	// Godan: convert to お-stem + う
	return stem + string(godanToO(lastChar)) + "う"
}

func conjugatePotential(verb, stem string, lastChar rune, verbType VerbType) string {
	if verbType == Ichidan {
		return stem + "られる"
	}
	// Godan: convert to え-stem + る
	return stem + string(godanToE(lastChar)) + "る"
}

func conjugateCausative(verb, stem string, lastChar rune, verbType VerbType) string {
	if verbType == Ichidan {
		return stem + "させる"
	}
	// Godan: convert to あ-stem + せる
	return stem + string(godanToA(lastChar)) + "せる"
}

func conjugatePassive(verb, stem string, lastChar rune, verbType VerbType) string {
	if verbType == Ichidan {
		return stem + "られた"
	}
	// Godan: convert to あ-stem + れた
	return stem + string(godanToA(lastChar)) + "れた"
}

// Godan vowel row converters
func godanToA(r rune) rune {
	switch r {
	case 'う':
		return 'わ'
	case 'く':
		return 'か'
	case 'ぐ':
		return 'が'
	case 'す':
		return 'さ'
	case 'つ':
		return 'た'
	case 'ぬ':
		return 'な'
	case 'ぶ':
		return 'ば'
	case 'む':
		return 'ま'
	case 'る':
		return 'ら'
	default:
		return 'あ'
	}
}

func godanToI(r rune) rune {
	switch r {
	case 'う':
		return 'い'
	case 'く':
		return 'き'
	case 'ぐ':
		return 'ぎ'
	case 'す':
		return 'し'
	case 'つ':
		return 'ち'
	case 'ぬ':
		return 'に'
	case 'ぶ':
		return 'び'
	case 'む':
		return 'み'
	case 'る':
		return 'り'
	default:
		return 'い'
	}
}

func godanToE(r rune) rune {
	switch r {
	case 'う':
		return 'え'
	case 'く':
		return 'け'
	case 'ぐ':
		return 'げ'
	case 'す':
		return 'せ'
	case 'つ':
		return 'て'
	case 'ぬ':
		return 'ね'
	case 'ぶ':
		return 'べ'
	case 'む':
		return 'め'
	case 'る':
		return 'れ'
	default:
		return 'え'
	}
}

func godanToO(r rune) rune {
	switch r {
	case 'う':
		return 'お'
	case 'く':
		return 'こ'
	case 'ぐ':
		return 'ご'
	case 'す':
		return 'そ'
	case 'つ':
		return 'と'
	case 'ぬ':
		return 'の'
	case 'ぶ':
		return 'ぼ'
	case 'む':
		return 'も'
	case 'る':
		return 'ろ'
	default:
		return 'お'
	}
}

// applyModifiers applies negation and/or politeness to all conjugations
func applyModifiers(conjugations map[string]interface{}, verb string, verbType VerbType, negative bool, polite bool, engConjugator *EnglishConjugator) map[string]interface{} {
	// Helper function to modify a ConjugationEntry
	modifyEntry := func(entry ConjugationEntry, baseForm string) ConjugationEntry {
		modified := entry
		modified.Japanese = applyModifiersToForm(baseForm, verb, verbType, negative, polite)
		modified.English = modifyEnglish(entry.English, negative, polite)

		// Update alternatives
		if len(entry.Alts) > 0 {
			newAlts := make([]string, len(entry.Alts))
			for i, alt := range entry.Alts {
				newAlts[i] = applyModifiersToForm(alt, verb, verbType, negative, polite)
			}
			modified.Alts = newAlts
		}
		return modified
	}

	newConjugations := make(map[string]interface{})

	// Process tenses
	if tenses, ok := conjugations["tenses"].(map[string]interface{}); ok {
		newTenses := make(map[string]interface{})

		for category, forms := range tenses {
			if formsMap, ok := forms.(map[string]ConjugationEntry); ok {
				newForms := make(map[string]ConjugationEntry)
				for formName, entry := range formsMap {
					newForms[formName] = modifyEntry(entry, entry.Japanese)
				}
				newTenses[category] = newForms
			}
		}
		newConjugations["tenses"] = newTenses
	}

	// Process voice
	if voice, ok := conjugations["voice"].(map[string]ConjugationEntry); ok {
		newVoice := make(map[string]ConjugationEntry)
		for formName, entry := range voice {
			newVoice[formName] = modifyEntry(entry, entry.Japanese)
		}
		newConjugations["voice"] = newVoice
	}

	return newConjugations
}

// applyModifiersToForm applies negative and/or polite modifiers to a single Japanese form
func applyModifiersToForm(form string, originalVerb string, verbType VerbType, negative bool, polite bool) string {
	if form == "" {
		return form
	}

	// For irregular verbs, handle separately
	if verbType == Irregular {
		return applyModifiersIrregular(form, originalVerb, negative, polite)
	}

	stem := getStem(originalVerb)
	lastChar := getLastRune(originalVerb)

	// Detect what tense/form we're dealing with and apply appropriate transformation

	// Check if it's already a past tense form (ends in た/だ)
	if strings.HasSuffix(form, "た") || strings.HasSuffix(form, "だ") {
		// Past tense
		if negative && polite {
			// Negative + Polite past: ませんでした
			return conjugateMasu(originalVerb, stem, lastChar, verbType) + "ませんでした"
		} else if negative {
			// Negative past: なかった
			return conjugateNegative(originalVerb, stem, lastChar, verbType) + "かった"
		} else if polite {
			// Polite past: ました
			return conjugateMasu(originalVerb, stem, lastChar, verbType) + "ました"
		}
	}

	// Check if it's a te-form + いる (progressive)
	if strings.HasSuffix(form, "ている") {
		teForm := conjugateTe(originalVerb, stem, lastChar, verbType)
		if negative && polite {
			// Negative + Polite progressive: ていません
			return teForm + "いません"
		} else if negative {
			// Negative progressive: ていない
			return teForm + "いない"
		} else if polite {
			// Polite progressive: ています
			return teForm + "います"
		}
	}

	// Check if it's conditional form (ends in ば)
	if strings.HasSuffix(form, "ば") {
		if negative && polite {
			// Negative + Polite conditional: なければ (polite doesn't really apply here, use negative)
			return conjugateNegative(originalVerb, stem, lastChar, verbType) + "ければ"
		} else if negative {
			// Negative conditional: なければ
			return conjugateNegative(originalVerb, stem, lastChar, verbType) + "ければ"
		} else if polite {
			// Polite conditional - still use ば form (politeness shown elsewhere)
			return form
		}
	}

	// Check if it's volitional (ends in よう or おう)
	if strings.HasSuffix(form, "よう") || strings.HasSuffix(form, "おう") {
		if negative && polite {
			// Negative + Polite volitional: ましょう doesn't have negative, use ません
			return conjugateMasu(originalVerb, stem, lastChar, verbType) + "ません"
		} else if negative {
			// Negative volitional: use negative + よう isn't common, use ない
			return conjugateNegative(originalVerb, stem, lastChar, verbType) + "い"
		} else if polite {
			// Polite volitional: ましょう
			return conjugateMasu(originalVerb, stem, lastChar, verbType) + "ましょう"
		}
	}

	// Check if it's potential form (ends in られる or える)
	if strings.HasSuffix(form, "られる") || (strings.HasSuffix(form, "る") && form != originalVerb) {
		// For potential/causative/passive forms, they're verbs themselves, apply recursively
		if negative && polite {
			// Get the stem of the -ru verb and apply negative polite
			formStem := getStem(form)
			return formStem + "ません"
		} else if negative {
			// Negative: られない, えない
			formStem := getStem(form)
			return formStem + "ない"
		} else if polite {
			// Polite: られます, えます
			formStem := getStem(form)
			return formStem + "ます"
		}
	}

	// Check for various compound endings
	if strings.Contains(form, "ばかり") {
		// Perfect aspect (just done) - has past tense + ばかり
		if negative && polite {
			// Need negative past of the verb
			negPast := conjugateNegative(originalVerb, stem, lastChar, verbType) + "かった"
			return negPast + "ばかり"
		} else if negative {
			negPast := conjugateNegative(originalVerb, stem, lastChar, verbType) + "かった"
			return negPast + "ばかり"
		} else if polite {
			pastForm := conjugateMasu(originalVerb, stem, lastChar, verbType) + "ました"
			return pastForm + "ばかり"
		}
	}

	if strings.Contains(form, "ければならない") {
		// Deontic (must)
		if negative && polite {
			// Must not - polite: doesn't have standard form, use negative + polite markers
			return conjugateNegative(originalVerb, stem, lastChar, verbType) + "くてもいいです"
		} else if negative {
			// Don't have to / need not
			return conjugateNegative(originalVerb, stem, lastChar, verbType) + "くてもいい"
		} else if polite {
			return conjugateNegative(originalVerb, stem, lastChar, verbType) + "ければなりません"
		}
	}

	if strings.Contains(form, "たい") {
		// Desire form
		if negative && polite {
			return conjugateMasu(originalVerb, stem, lastChar, verbType) + "たくないです"
		} else if negative {
			return conjugateMasu(originalVerb, stem, lastChar, verbType) + "たくない"
		} else if polite {
			return conjugateMasu(originalVerb, stem, lastChar, verbType) + "たいです"
		}
	}

	// Default: present tense (dictionary form or imperative/other)
	if negative && polite {
		// Negative + Polite present: ません
		return conjugateMasu(originalVerb, stem, lastChar, verbType) + "ません"
	} else if negative {
		// Negative present: ない
		return conjugateNegative(originalVerb, stem, lastChar, verbType) + "い"
	} else if polite {
		// Polite present: ます
		return conjugateMasu(originalVerb, stem, lastChar, verbType) + "ます"
	}

	return form
}

// applyModifiersIrregular handles modifiers for irregular verbs
func applyModifiersIrregular(form string, originalVerb string, negative bool, polite bool) string {
	if originalVerb == "する" || originalVerb == "為る" {
		// Check for past tense
		if strings.HasSuffix(form, "した") {
			if negative && polite {
				return "しませんでした"
			} else if negative {
				return "しなかった"
			} else if polite {
				return "しました"
			}
		}
		// Check for progressive
		if strings.HasSuffix(form, "している") {
			if negative && polite {
				return "していません"
			} else if negative {
				return "していない"
			} else if polite {
				return "しています"
			}
		}
		// Check for conditional
		if strings.HasSuffix(form, "すれば") {
			if negative && polite {
				return "しなければ"
			} else if negative {
				return "しなければ"
			} else if polite {
				return form // Conditional doesn't really change with politeness
			}
		}
		// Check for volitional
		if strings.HasSuffix(form, "しよう") {
			if negative && polite {
				return "しません"
			} else if negative {
				return "しない"
			} else if polite {
				return "しましょう"
			}
		}
		// Check for various compound forms
		if strings.Contains(form, "ばかり") {
			if negative && polite {
				return "しなかったばかり"
			} else if negative {
				return "しなかったばかり"
			} else if polite {
				return "しましたばかり"
			}
		}
		if strings.Contains(form, "たい") {
			if negative && polite {
				return "したくないです"
			} else if negative {
				return "したくない"
			} else if polite {
				return "したいです"
			}
		}
		if strings.Contains(form, "ければならない") {
			if negative && polite {
				return "しなくてもいいです"
			} else if negative {
				return "しなくてもいい"
			} else if polite {
				return "しなければなりません"
			}
		}
		// Default: present tense
		if negative && polite {
			return "しません"
		} else if negative {
			return "しない"
		} else if polite {
			return "します"
		}
	} else if originalVerb == "来る" || originalVerb == "くる" {
		// Check for past tense
		if strings.Contains(form, "来た") || strings.Contains(form, "きた") {
			if negative && polite {
				return "来ませんでした"
			} else if negative {
				return "来なかった"
			} else if polite {
				return "来ました"
			}
		}
		// Check for progressive
		if strings.Contains(form, "来ている") || strings.Contains(form, "きている") {
			if negative && polite {
				return "来ていません"
			} else if negative {
				return "来ていない"
			} else if polite {
				return "来ています"
			}
		}
		// Check for conditional
		if strings.Contains(form, "来れば") {
			if negative && polite {
				return "来なければ"
			} else if negative {
				return "来なければ"
			} else if polite {
				return form
			}
		}
		// Check for volitional
		if strings.Contains(form, "来よう") {
			if negative && polite {
				return "来ません"
			} else if negative {
				return "来ない"
			} else if polite {
				return "来ましょう"
			}
		}
		// Check for various compound forms
		if strings.Contains(form, "ばかり") {
			if negative && polite {
				return "来なかったばかり"
			} else if negative {
				return "来なかったばかり"
			} else if polite {
				return "来ましたばかり"
			}
		}
		if strings.Contains(form, "たい") {
			if negative && polite {
				return "来たくないです"
			} else if negative {
				return "来たくない"
			} else if polite {
				return "来たいです"
			}
		}
		if strings.Contains(form, "ければならない") {
			if negative && polite {
				return "来なくてもいいです"
			} else if negative {
				return "来なくてもいい"
			} else if polite {
				return "来なければなりません"
			}
		}
		// Default: present tense
		if negative && polite {
			return "来ません"
		} else if negative {
			return "来ない"
		} else if polite {
			return "来ます"
		}
	}
	return form
}

// modifyEnglish updates English conjugations for negative and polite forms
func modifyEnglish(english string, negative bool, polite bool) string {
	if english == "" {
		return english
	}

	// Simple transformation - prepend modifiers
	var modifiers []string
	if polite {
		modifiers = append(modifiers, "[Polite]")
	}
	if negative {
		// Add "not" to the English
		// This is a simple implementation - a more sophisticated one would parse the sentence
		if strings.Contains(strings.ToLower(english), " do") {
			english = strings.Replace(english, " do", " don't", 1)
		} else if strings.Contains(strings.ToLower(english), " did") {
			english = strings.Replace(english, " did", " didn't", 1)
		} else if strings.Contains(strings.ToLower(english), " will") {
			english = strings.Replace(english, " will", " won't", 1)
		} else if strings.Contains(strings.ToLower(english), " am") {
			english = strings.Replace(english, " am", " am not", 1)
		} else if strings.Contains(strings.ToLower(english), " have") {
			english = strings.Replace(english, " have", " haven't", 1)
		} else if strings.Contains(strings.ToLower(english), " can") {
			english = strings.Replace(english, " can", " can't", 1)
		} else if strings.Contains(strings.ToLower(english), " must") {
			english = strings.Replace(english, " must", " must not", 1)
		} else if strings.HasPrefix(strings.ToLower(english), "i ") {
			// Handle "I verb" pattern
			parts := strings.SplitN(english, " ", 2)
			if len(parts) == 2 {
				english = parts[0] + " don't " + parts[1]
			}
		} else {
			// Default: just add [Negative] marker
			modifiers = append([]string{"[Negative]"}, modifiers...)
		}
	}

	if len(modifiers) > 0 {
		return strings.Join(modifiers, " ") + " " + english
	}
	return english
}

func conjugateIrregular(verb string, engConjugator *EnglishConjugator) map[string]interface{} {
	if verb == "する" || verb == "為る" {
		return map[string]interface{}{
			"time": map[string]ConjugationEntry{
				"present": {English: getEnglishForForm(engConjugator, "time", "present"), Japanese: "する", Alts: []string{}},
				"past":    {English: getEnglishForForm(engConjugator, "time", "past"), Japanese: "した", Alts: []string{}},
				"future":  {English: getEnglishForForm(engConjugator, "time", "future"), Japanese: "する", Alts: []string{}},
			},
			"aspect": map[string]ConjugationEntry{
				"simple":              {English: getEnglishForForm(engConjugator, "aspect", "simple"), Japanese: "する", Alts: []string{}},
				"progressive":         {English: getEnglishForForm(engConjugator, "aspect", "progressive"), Japanese: "している", Alts: []string{}},
				"perfect":             {English: getEnglishForForm(engConjugator, "aspect", "perfect"), Japanese: "したばかり", Alts: []string{}},
				"perfect_progressive": {English: getEnglishForForm(engConjugator, "aspect", "perfect_progressive"), Japanese: "している", Alts: []string{}},
			},
			"mood": map[string]ConjugationEntry{
				"indicative":  {English: getEnglishForForm(engConjugator, "mood", "indicative"), Japanese: "する", Alts: []string{}},
				"subjunctive": {English: getEnglishForForm(engConjugator, "mood", "subjunctive"), Japanese: "したらいいのに", Alts: []string{}},
				"conditional": {English: getEnglishForForm(engConjugator, "mood", "conditional"), Japanese: "すれば", Alts: []string{"したら"}},
				"imperative":  {English: getEnglishForForm(engConjugator, "mood", "imperative"), Japanese: "しろ", Alts: []string{"せよ"}},
				"volitional":  {English: getEnglishForForm(engConjugator, "mood", "volitional"), Japanese: "しよう", Alts: []string{}},
			},
			"modals": map[string]ConjugationEntry{
				"potential": {English: getEnglishForForm(engConjugator, "modals", "potential"), Japanese: "できる", Alts: []string{}},
				"causative": {English: getEnglishForForm(engConjugator, "modals", "causative"), Japanese: "させる", Alts: []string{}},
				"deontic":   {English: getEnglishForForm(engConjugator, "modals", "deontic"), Japanese: "しなければならない", Alts: []string{}},
			},
			"desire": map[string]ConjugationEntry{
				"subject": {English: getEnglishForForm(engConjugator, "desire", "subject"), Japanese: "したい", Alts: []string{}},
			},
		}
	} else if verb == "来る" || verb == "くる" {
		return map[string]interface{}{
			"time": map[string]ConjugationEntry{
				"present": {English: getEnglishForForm(engConjugator, "time", "present"), Japanese: "来る", Alts: []string{"くる"}},
				"past":    {English: getEnglishForForm(engConjugator, "time", "past"), Japanese: "来た", Alts: []string{"きた"}},
				"future":  {English: getEnglishForForm(engConjugator, "time", "future"), Japanese: "来る", Alts: []string{"くる"}},
			},
			"aspect": map[string]ConjugationEntry{
				"simple":              {English: getEnglishForForm(engConjugator, "aspect", "simple"), Japanese: "来る", Alts: []string{"くる"}},
				"progressive":         {English: getEnglishForForm(engConjugator, "aspect", "progressive"), Japanese: "来ている", Alts: []string{"きている"}},
				"perfect":             {English: getEnglishForForm(engConjugator, "aspect", "perfect"), Japanese: "来たばかり", Alts: []string{}},
				"perfect_progressive": {English: getEnglishForForm(engConjugator, "aspect", "perfect_progressive"), Japanese: "来ている", Alts: []string{}},
			},
			"mood": map[string]ConjugationEntry{
				"indicative":  {English: getEnglishForForm(engConjugator, "mood", "indicative"), Japanese: "来る", Alts: []string{}},
				"subjunctive": {English: getEnglishForForm(engConjugator, "mood", "subjunctive"), Japanese: "来たらいいのに", Alts: []string{}},
				"conditional": {English: getEnglishForForm(engConjugator, "mood", "conditional"), Japanese: "来れば", Alts: []string{"来たら"}},
				"imperative":  {English: getEnglishForForm(engConjugator, "mood", "imperative"), Japanese: "来い", Alts: []string{}},
				"volitional":  {English: getEnglishForForm(engConjugator, "mood", "volitional"), Japanese: "来よう", Alts: []string{}},
			},
			"modals": map[string]ConjugationEntry{
				"potential": {English: getEnglishForForm(engConjugator, "modals", "potential"), Japanese: "来られる", Alts: []string{}},
				"causative": {English: getEnglishForForm(engConjugator, "modals", "causative"), Japanese: "来させる", Alts: []string{}},
				"deontic":   {English: getEnglishForForm(engConjugator, "modals", "deontic"), Japanese: "来なければならない", Alts: []string{}},
			},
			"desire": map[string]ConjugationEntry{
				"subject": {English: getEnglishForForm(engConjugator, "desire", "subject"), Japanese: "来たい", Alts: []string{}},
			},
		}
	}
	return nil
}

func voiceIrregular(verb string, engConjugator *EnglishConjugator) map[string]ConjugationEntry {
	if verb == "する" || verb == "為る" {
		return map[string]ConjugationEntry{
			"active":  {English: getEnglishForForm(engConjugator, "voice", "active"), Japanese: "する", Alts: []string{}},
			"passive": {English: getEnglishForForm(engConjugator, "voice", "passive"), Japanese: "された", Alts: []string{}},
		}
	} else if verb == "来る" || verb == "くる" {
		return map[string]ConjugationEntry{
			"active":  {English: getEnglishForForm(engConjugator, "voice", "active"), Japanese: "来る", Alts: []string{}},
			"passive": {English: getEnglishForForm(engConjugator, "voice", "passive"), Japanese: "来られた", Alts: []string{}},
		}
	}
	return nil
}
