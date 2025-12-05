//go:build ignore

package main

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"

	"gaijin/internal/database"
)

// KanaEntry represents a kana character with its romaji and category
type KanaEntry struct {
	Character string
	Romaji    string
	Category  string
}

// All hiragana characters organized by category
var hiraganaData = []KanaEntry{
	// Vowels (gojūon)
	{"あ", "a", "vowel"},
	{"い", "i", "vowel"},
	{"う", "u", "vowel"},
	{"え", "e", "vowel"},
	{"お", "o", "vowel"},

	// K-row
	{"か", "ka", "k-row"},
	{"き", "ki", "k-row"},
	{"く", "ku", "k-row"},
	{"け", "ke", "k-row"},
	{"こ", "ko", "k-row"},

	// S-row
	{"さ", "sa", "s-row"},
	{"し", "shi", "s-row"},
	{"す", "su", "s-row"},
	{"せ", "se", "s-row"},
	{"そ", "so", "s-row"},

	// T-row
	{"た", "ta", "t-row"},
	{"ち", "chi", "t-row"},
	{"つ", "tsu", "t-row"},
	{"て", "te", "t-row"},
	{"と", "to", "t-row"},

	// N-row
	{"な", "na", "n-row"},
	{"に", "ni", "n-row"},
	{"ぬ", "nu", "n-row"},
	{"ね", "ne", "n-row"},
	{"の", "no", "n-row"},

	// H-row
	{"は", "ha", "h-row"},
	{"ひ", "hi", "h-row"},
	{"ふ", "fu", "h-row"},
	{"へ", "he", "h-row"},
	{"ほ", "ho", "h-row"},

	// M-row
	{"ま", "ma", "m-row"},
	{"み", "mi", "m-row"},
	{"む", "mu", "m-row"},
	{"め", "me", "m-row"},
	{"も", "mo", "m-row"},

	// Y-row
	{"や", "ya", "y-row"},
	{"ゆ", "yu", "y-row"},
	{"よ", "yo", "y-row"},

	// R-row
	{"ら", "ra", "r-row"},
	{"り", "ri", "r-row"},
	{"る", "ru", "r-row"},
	{"れ", "re", "r-row"},
	{"ろ", "ro", "r-row"},

	// W-row
	{"わ", "wa", "w-row"},
	{"を", "wo", "w-row"},

	// N
	{"ん", "n", "n-standalone"},

	// Dakuten (voiced) - G-row
	{"が", "ga", "g-row (dakuten)"},
	{"ぎ", "gi", "g-row (dakuten)"},
	{"ぐ", "gu", "g-row (dakuten)"},
	{"げ", "ge", "g-row (dakuten)"},
	{"ご", "go", "g-row (dakuten)"},

	// Dakuten - Z-row
	{"ざ", "za", "z-row (dakuten)"},
	{"じ", "ji", "z-row (dakuten)"},
	{"ず", "zu", "z-row (dakuten)"},
	{"ぜ", "ze", "z-row (dakuten)"},
	{"ぞ", "zo", "z-row (dakuten)"},

	// Dakuten - D-row
	{"だ", "da", "d-row (dakuten)"},
	{"ぢ", "ji", "d-row (dakuten)"},
	{"づ", "zu", "d-row (dakuten)"},
	{"で", "de", "d-row (dakuten)"},
	{"ど", "do", "d-row (dakuten)"},

	// Dakuten - B-row
	{"ば", "ba", "b-row (dakuten)"},
	{"び", "bi", "b-row (dakuten)"},
	{"ぶ", "bu", "b-row (dakuten)"},
	{"べ", "be", "b-row (dakuten)"},
	{"ぼ", "bo", "b-row (dakuten)"},

	// Handakuten - P-row
	{"ぱ", "pa", "p-row (handakuten)"},
	{"ぴ", "pi", "p-row (handakuten)"},
	{"ぷ", "pu", "p-row (handakuten)"},
	{"ぺ", "pe", "p-row (handakuten)"},
	{"ぽ", "po", "p-row (handakuten)"},

	// Yōon (combination sounds) - K
	{"きゃ", "kya", "yōon (k)"},
	{"きゅ", "kyu", "yōon (k)"},
	{"きょ", "kyo", "yōon (k)"},

	// Yōon - S
	{"しゃ", "sha", "yōon (s)"},
	{"しゅ", "shu", "yōon (s)"},
	{"しょ", "sho", "yōon (s)"},

	// Yōon - T
	{"ちゃ", "cha", "yōon (t)"},
	{"ちゅ", "chu", "yōon (t)"},
	{"ちょ", "cho", "yōon (t)"},

	// Yōon - N
	{"にゃ", "nya", "yōon (n)"},
	{"にゅ", "nyu", "yōon (n)"},
	{"にょ", "nyo", "yōon (n)"},

	// Yōon - H
	{"ひゃ", "hya", "yōon (h)"},
	{"ひゅ", "hyu", "yōon (h)"},
	{"ひょ", "hyo", "yōon (h)"},

	// Yōon - M
	{"みゃ", "mya", "yōon (m)"},
	{"みゅ", "myu", "yōon (m)"},
	{"みょ", "myo", "yōon (m)"},

	// Yōon - R
	{"りゃ", "rya", "yōon (r)"},
	{"りゅ", "ryu", "yōon (r)"},
	{"りょ", "ryo", "yōon (r)"},

	// Yōon - G (dakuten)
	{"ぎゃ", "gya", "yōon (g)"},
	{"ぎゅ", "gyu", "yōon (g)"},
	{"ぎょ", "gyo", "yōon (g)"},

	// Yōon - J (dakuten)
	{"じゃ", "ja", "yōon (j)"},
	{"じゅ", "ju", "yōon (j)"},
	{"じょ", "jo", "yōon (j)"},

	// Yōon - B (dakuten)
	{"びゃ", "bya", "yōon (b)"},
	{"びゅ", "byu", "yōon (b)"},
	{"びょ", "byo", "yōon (b)"},

	// Yōon - P (handakuten)
	{"ぴゃ", "pya", "yōon (p)"},
	{"ぴゅ", "pyu", "yōon (p)"},
	{"ぴょ", "pyo", "yōon (p)"},
}

// All katakana characters organized by category
var katakanaData = []KanaEntry{
	// Vowels (gojūon)
	{"ア", "a", "vowel"},
	{"イ", "i", "vowel"},
	{"ウ", "u", "vowel"},
	{"エ", "e", "vowel"},
	{"オ", "o", "vowel"},

	// K-row
	{"カ", "ka", "k-row"},
	{"キ", "ki", "k-row"},
	{"ク", "ku", "k-row"},
	{"ケ", "ke", "k-row"},
	{"コ", "ko", "k-row"},

	// S-row
	{"サ", "sa", "s-row"},
	{"シ", "shi", "s-row"},
	{"ス", "su", "s-row"},
	{"セ", "se", "s-row"},
	{"ソ", "so", "s-row"},

	// T-row
	{"タ", "ta", "t-row"},
	{"チ", "chi", "t-row"},
	{"ツ", "tsu", "t-row"},
	{"テ", "te", "t-row"},
	{"ト", "to", "t-row"},

	// N-row
	{"ナ", "na", "n-row"},
	{"ニ", "ni", "n-row"},
	{"ヌ", "nu", "n-row"},
	{"ネ", "ne", "n-row"},
	{"ノ", "no", "n-row"},

	// H-row
	{"ハ", "ha", "h-row"},
	{"ヒ", "hi", "h-row"},
	{"フ", "fu", "h-row"},
	{"ヘ", "he", "h-row"},
	{"ホ", "ho", "h-row"},

	// M-row
	{"マ", "ma", "m-row"},
	{"ミ", "mi", "m-row"},
	{"ム", "mu", "m-row"},
	{"メ", "me", "m-row"},
	{"モ", "mo", "m-row"},

	// Y-row
	{"ヤ", "ya", "y-row"},
	{"ユ", "yu", "y-row"},
	{"ヨ", "yo", "y-row"},

	// R-row
	{"ラ", "ra", "r-row"},
	{"リ", "ri", "r-row"},
	{"ル", "ru", "r-row"},
	{"レ", "re", "r-row"},
	{"ロ", "ro", "r-row"},

	// W-row
	{"ワ", "wa", "w-row"},
	{"ヲ", "wo", "w-row"},

	// N
	{"ン", "n", "n-standalone"},

	// Dakuten (voiced) - G-row
	{"ガ", "ga", "g-row (dakuten)"},
	{"ギ", "gi", "g-row (dakuten)"},
	{"グ", "gu", "g-row (dakuten)"},
	{"ゲ", "ge", "g-row (dakuten)"},
	{"ゴ", "go", "g-row (dakuten)"},

	// Dakuten - Z-row
	{"ザ", "za", "z-row (dakuten)"},
	{"ジ", "ji", "z-row (dakuten)"},
	{"ズ", "zu", "z-row (dakuten)"},
	{"ゼ", "ze", "z-row (dakuten)"},
	{"ゾ", "zo", "z-row (dakuten)"},

	// Dakuten - D-row
	{"ダ", "da", "d-row (dakuten)"},
	{"ヂ", "ji", "d-row (dakuten)"},
	{"ヅ", "zu", "d-row (dakuten)"},
	{"デ", "de", "d-row (dakuten)"},
	{"ド", "do", "d-row (dakuten)"},

	// Dakuten - B-row
	{"バ", "ba", "b-row (dakuten)"},
	{"ビ", "bi", "b-row (dakuten)"},
	{"ブ", "bu", "b-row (dakuten)"},
	{"ベ", "be", "b-row (dakuten)"},
	{"ボ", "bo", "b-row (dakuten)"},

	// Handakuten - P-row
	{"パ", "pa", "p-row (handakuten)"},
	{"ピ", "pi", "p-row (handakuten)"},
	{"プ", "pu", "p-row (handakuten)"},
	{"ペ", "pe", "p-row (handakuten)"},
	{"ポ", "po", "p-row (handakuten)"},

	// Yōon (combination sounds) - K
	{"キャ", "kya", "yōon (k)"},
	{"キュ", "kyu", "yōon (k)"},
	{"キョ", "kyo", "yōon (k)"},

	// Yōon - S
	{"シャ", "sha", "yōon (s)"},
	{"シュ", "shu", "yōon (s)"},
	{"ショ", "sho", "yōon (s)"},

	// Yōon - T
	{"チャ", "cha", "yōon (t)"},
	{"チュ", "chu", "yōon (t)"},
	{"チョ", "cho", "yōon (t)"},

	// Yōon - N
	{"ニャ", "nya", "yōon (n)"},
	{"ニュ", "nyu", "yōon (n)"},
	{"ニョ", "nyo", "yōon (n)"},

	// Yōon - H
	{"ヒャ", "hya", "yōon (h)"},
	{"ヒュ", "hyu", "yōon (h)"},
	{"ヒョ", "hyo", "yōon (h)"},

	// Yōon - M
	{"ミャ", "mya", "yōon (m)"},
	{"ミュ", "myu", "yōon (m)"},
	{"ミョ", "myo", "yōon (m)"},

	// Yōon - R
	{"リャ", "rya", "yōon (r)"},
	{"リュ", "ryu", "yōon (r)"},
	{"リョ", "ryo", "yōon (r)"},

	// Yōon - G (dakuten)
	{"ギャ", "gya", "yōon (g)"},
	{"ギュ", "gyu", "yōon (g)"},
	{"ギョ", "gyo", "yōon (g)"},

	// Yōon - J (dakuten)
	{"ジャ", "ja", "yōon (j)"},
	{"ジュ", "ju", "yōon (j)"},
	{"ジョ", "jo", "yōon (j)"},

	// Yōon - B (dakuten)
	{"ビャ", "bya", "yōon (b)"},
	{"ビュ", "byu", "yōon (b)"},
	{"ビョ", "byo", "yōon (b)"},

	// Yōon - P (handakuten)
	{"ピャ", "pya", "yōon (p)"},
	{"ピュ", "pyu", "yōon (p)"},
	{"ピョ", "pyo", "yōon (p)"},

	// Extended katakana for foreign sounds
	{"ティ", "ti", "extended"},
	{"ディ", "di", "extended"},
	{"トゥ", "tu", "extended"},
	{"ドゥ", "du", "extended"},
	{"ファ", "fa", "extended"},
	{"フィ", "fi", "extended"},
	{"フェ", "fe", "extended"},
	{"フォ", "fo", "extended"},
	{"ヴァ", "va", "extended"},
	{"ヴィ", "vi", "extended"},
	{"ヴ", "vu", "extended"},
	{"ヴェ", "ve", "extended"},
	{"ヴォ", "vo", "extended"},
	{"ウィ", "wi", "extended"},
	{"ウェ", "we", "extended"},
	{"ウォ", "wo", "extended"},
}

func main() {
	// Load environment variables
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	// Connect to database
	db, err := database.ConnectDB()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}
	defer db.Close()

	// Initialize tables (in case they don't exist)
	err = db.InitializeTables()
	if err != nil {
		log.Fatal("Failed to initialize tables:", err)
	}

	// Seed hiragana
	hiraganaCount := 0
	for _, h := range hiraganaData {
		query := `
			INSERT INTO hiragana (character, romaji, category)
			VALUES ($1, $2, $3)
			ON CONFLICT (character) DO NOTHING
		`
		result, err := db.DB.Exec(query, h.Character, h.Romaji, h.Category)
		if err != nil {
			log.Printf("Failed to insert hiragana %s: %v", h.Character, err)
			continue
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			hiraganaCount++
		}
	}
	fmt.Printf("✅ Inserted %d new hiragana characters (total in list: %d)\n", hiraganaCount, len(hiraganaData))

	// Seed katakana
	katakanaCount := 0
	for _, k := range katakanaData {
		query := `
			INSERT INTO katakana (character, romaji, category)
			VALUES ($1, $2, $3)
			ON CONFLICT (character) DO NOTHING
		`
		result, err := db.DB.Exec(query, k.Character, k.Romaji, k.Category)
		if err != nil {
			log.Printf("Failed to insert katakana %s: %v", k.Character, err)
			continue
		}
		rowsAffected, _ := result.RowsAffected()
		if rowsAffected > 0 {
			katakanaCount++
		}
	}
	fmt.Printf("✅ Inserted %d new katakana characters (total in list: %d)\n", katakanaCount, len(katakanaData))

	// Print summary
	var hCount, kCount int
	db.DB.QueryRow("SELECT COUNT(*) FROM hiragana").Scan(&hCount)
	db.DB.QueryRow("SELECT COUNT(*) FROM katakana").Scan(&kCount)
	fmt.Printf("\n📊 Database totals:\n")
	fmt.Printf("   Hiragana: %d characters\n", hCount)
	fmt.Printf("   Katakana: %d characters\n", kCount)

	// Verify by printing some samples
	fmt.Printf("\n🔍 Sample hiragana:\n")
	rows, _ := db.DB.Query("SELECT character, romaji, category FROM hiragana LIMIT 5")
	for rows.Next() {
		var char, romaji, category string
		rows.Scan(&char, &romaji, &category)
		fmt.Printf("   %s → %s (%s)\n", char, romaji, category)
	}
	rows.Close()

	fmt.Printf("\n🔍 Sample katakana:\n")
	rows, _ = db.DB.Query("SELECT character, romaji, category FROM katakana LIMIT 5")
	for rows.Next() {
		var char, romaji, category string
		rows.Scan(&char, &romaji, &category)
		fmt.Printf("   %s → %s (%s)\n", char, romaji, category)
	}
	rows.Close()

	fmt.Println("\n✅ Kana seeding complete!")

	// Check if run with --verify flag
	if len(os.Args) > 1 && os.Args[1] == "--verify" {
		fmt.Println("\n📝 Full hiragana list:")
		rows, _ := db.DB.Query("SELECT character, romaji, category FROM hiragana ORDER BY id")
		for rows.Next() {
			var char, romaji, category string
			rows.Scan(&char, &romaji, &category)
			fmt.Printf("   %s → %s (%s)\n", char, romaji, category)
		}
		rows.Close()

		fmt.Println("\n📝 Full katakana list:")
		rows, _ = db.DB.Query("SELECT character, romaji, category FROM katakana ORDER BY id")
		for rows.Next() {
			var char, romaji, category string
			rows.Scan(&char, &romaji, &category)
			fmt.Printf("   %s → %s (%s)\n", char, romaji, category)
		}
		rows.Close()
	}
}
