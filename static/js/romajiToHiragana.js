    // Comprehensive Romanji to Hiragana lookup table
    const romanjiToHiraganaMap = {
        // Vowels
        'a': 'あ', 'i': 'い', 'u': 'う', 'e': 'え', 'o': 'お',
        
        // K-line
        'ka': 'か', 'ki': 'き', 'ku': 'く', 'ke': 'け', 'ko': 'こ',
        'kya': 'きゃ', 'kyu': 'きゅ', 'kyo': 'きょ',
        
        // S-line
        'sa': 'さ', 'shi': 'し', 'si': 'し', 'su': 'す', 'se': 'せ', 'so': 'そ',
        'sha': 'しゃ', 'sya': 'しゃ', 'shya': 'しゃ',
        'shu': 'しゅ', 'syu': 'しゅ', 'shyu': 'しゅ',
        'sho': 'しょ', 'syo': 'しょ', 'shyo': 'しょ',
        
        // T-line
        'ta': 'た', 'chi': 'ち', 'ti': 'ち', 'tsu': 'つ', 'tu': 'つ', 'te': 'て', 'to': 'と',
        'cha': 'ちゃ', 'tya': 'ちゃ', 'chya': 'ちゃ',
        'chu': 'ちゅ', 'tyu': 'ちゅ', 'chyu': 'ちゅ',
        'cho': 'ちょ', 'tyo': 'ちょ', 'chyo': 'ちょ',
        
        // N-line
        'na': 'な', 'ni': 'に', 'nu': 'ぬ', 'ne': 'ね', 'no': 'の',
        'nya': 'にゃ', 'nyu': 'にゅ', 'nyo': 'にょ',
        
        // H-line
        'ha': 'は', 'hi': 'ひ', 'fu': 'ふ', 'hu': 'ふ', 'he': 'へ', 'ho': 'ほ',
        'hya': 'ひゃ', 'hyu': 'ひゅ', 'hyo': 'ひょ',
        
        // M-line
        'ma': 'ま', 'mi': 'み', 'mu': 'む', 'me': 'め', 'mo': 'も',
        'mya': 'みゃ', 'myu': 'みゅ', 'myo': 'みょ',
        
        // Y-line
        'ya': 'や', 'yu': 'ゆ', 'yo': 'よ',
        
        // R-line
        'ra': 'ら', 'ri': 'り', 'ru': 'る', 're': 'れ', 'ro': 'ろ',
        'rya': 'りゃ', 'ryu': 'りゅ', 'ryo': 'りょ',
        
        // W-line
        'wa': 'わ', 'wi': 'ゐ', 'we': 'ゑ', 'wo': 'を',
        
        // N
        'nn': 'ん',
        
        // G-line (dakuten)
        'ga': 'が', 'gi': 'ぎ', 'gu': 'ぐ', 'ge': 'げ', 'go': 'ご',
        'gya': 'ぎゃ', 'gyu': 'ぎゅ', 'gyo': 'ぎょ',
        
        // Z-line (dakuten)
        'za': 'ざ', 'ji': 'じ', 'zi': 'じ', 'zu': 'ず', 'ze': 'ぜ', 'zo': 'ぞ',
        'ja': 'じゃ', 'jya': 'じゃ', 'zya': 'じゃ', 'jia': 'じゃ',
        'ju': 'じゅ', 'jyu': 'じゅ', 'zyu': 'じゅ', 'jiu': 'じゅ',
        'jo': 'じょ', 'jyo': 'じょ', 'zyo': 'じょ', 'jio': 'じょ',
        
        // D-line (dakuten)
        'da': 'だ', 'di': 'ぢ', 'du': 'づ', 'de': 'で', 'do': 'ど',
        'dya': 'ぢゃ', 'dyu': 'ぢゅ', 'dyo': 'ぢょ',
        
        // B-line (dakuten)
        'ba': 'ば', 'bi': 'び', 'bu': 'ぶ', 'be': 'べ', 'bo': 'ぼ',
        'bya': 'びゃ', 'byu': 'びゅ', 'byo': 'びょ',
        
        // P-line (handakuten)
        'pa': 'ぱ', 'pi': 'ぴ', 'pu': 'ぷ', 'pe': 'ぺ', 'po': 'ぽ',
        'pya': 'ぴゃ', 'pyu': 'ぴゅ', 'pyo': 'ぴょ',
        
        // Additional combinations
        'kwa': 'くゎ', 'gwa': 'ぐゎ',
        'fa': 'ふぁ', 'fi': 'ふぃ', 'fe': 'ふぇ', 'fo': 'ふぉ',
        'va': 'ゔぁ', 'vi': 'ゔぃ', 'vu': 'ゔ', 've': 'ゔぇ', 'vo': 'ゔぉ',
    };

    function romanjiToHiragana(inputElement) {
        const romanji = inputElement.value.toLowerCase();
        let hiragana = '';
        let i = 0;
        
        // Store cursor position before conversion
        const cursorPos = inputElement.selectionStart;
        
        while (i < romanji.length) {
            // Handle small tsu (double consonants) - but NOT "nn" which becomes ん
            if (i < romanji.length - 1 && 
                romanji[i] === romanji[i + 1] && 
                romanji[i] !== 'n' &&  // Exclude 'nn' so it can become ん
                'bcdfghjklmnpqrstvwxyz'.includes(romanji[i])) {
                hiragana += 'っ';
                i++;
                continue;
            }
            
            // Try to match 3-character combinations first
            let matched = false;
            if (i <= romanji.length - 3) {
                const three = romanji.substring(i, i + 3);
                if (romanjiToHiraganaMap[three]) {
                    hiragana += romanjiToHiraganaMap[three];
                    i += 3;
                    matched = true;
                }
            }
            
            // Try 2-character combinations
            if (!matched && i <= romanji.length - 2) {
                const two = romanji.substring(i, i + 2);
                if (romanjiToHiraganaMap[two]) {
                    hiragana += romanjiToHiraganaMap[two];
                    i += 2;
                    matched = true;
                }
            }
            
            // Try single character
            if (!matched) {
                const one = romanji[i];
                if (romanjiToHiraganaMap[one]) {
                    hiragana += romanjiToHiraganaMap[one];
                    i++;
                    matched = true;
                }
            }
            
            // If no match, keep original character
            if (!matched) {
                hiragana += romanji[i];
                i++;
            }
        }
        
        // Update the input field value with hiragana
        inputElement.value = hiragana;
        
        // Try to maintain cursor position at the end
        inputElement.setSelectionRange(hiragana.length, hiragana.length);
    }