package models

// JLPTWord represents a single JLPT vocabulary word
type JLPTWord struct {
	Word     string `json:"word"`
	Meaning  string `json:"meaning"`
	Furigana string `json:"furigana"`
	Romaji   string `json:"romaji"`
	Level    int    `json:"level"`
}

// JLPTAPIResponse represents the paginated response from JLPT API
type JLPTAPIResponse struct {
	Words  []JLPTWord `json:"words"`
	Total  int        `json:"total"`
	Offset int        `json:"offset"`
	Limit  int        `json:"limit"`
}
