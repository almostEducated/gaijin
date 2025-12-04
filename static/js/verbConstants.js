// Verb Constants
// Basic verb examples and mapping constants

// Verb examples for the table
const verbExamples = {
    irregular: ['ある', 'いる', '行く', 'くる', 'する'],
    endings: ['い・え　る', 'る', 'う', 'く', 'す', 'つ', 'ぬ', 'ぶ', 'む', 'ぐ']
};

// Full verb examples matching the endings
const fullVerbExamples = {
    'い・え　る': '食べる',
    'る': '作る',
    'う': '買う',
    'く': '歩く',
    'す': '話す',
    'つ': '持つ',
    'ぬ': '死ぬ',
    'ぶ': '飛ぶ',
    'む': '飲む',
    'ぐ': '急ぐ'
};

// Map verb endings to root indices
// Order: う, つ, る (godan), く, ぐ, む, ぬ, ぶ, す, る (ichidan)
const endingToIndex = {
    'う': 0, 'つ': 1, 'る': 2, // godan ru
    'く': 3, 'ぐ': 4, 'む': 5, 'ぬ': 6, 'ぶ': 7, 'す': 8,
    'る': 9 // ichidan (い・え　る)
};

const ichidanRoots = {
    'い': 'い',
    'き': 'き',
    'し': 'し',
    'ち': 'ち',
    'に': 'に',
    'び': 'び',
    'み': 'み',
    'え': 'え',
    'け': 'け',
    'せ': 'せ',
    'て': 'て',
    'ね': 'ね',
    'べ': 'べ',
    'め': 'め',
};

