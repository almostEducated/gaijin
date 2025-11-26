// Verb Grid Conjugation JavaScript

// Current verb being conjugated
let currentVerb = '食べる';
let currentVerbData = null;

// Current selected mood
let currentMood = 'plain';

// State for control buttons
const controls = {
    // Voice
    potential: false,
    passive: false,
    causative: false,
    // Aspect
    continuous: false,
    completion: false,
    resultant: false,
    // Tense
    simple: false,
    past: false,
    // Other
    negative: false
};

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

// Get active voice (potential or causative)
function getActiveVoice() {
    if (controls.potential) {
        return 'potential';
    } else if (controls.causative) {
        return 'causative';
    }
    return null;
}

// Update form labels based on active voice
function updateFormLabels() {
    const activeVoice = getActiveVoice();
    const formLabels = {
        'plain': 'Plain',
        'te': 'T Form',
        'volitional': 'Volitional',
        'conditional': 'Conditional',
        'desiderative': 'Desiderative',
        'deontic': 'Deontic',
        'imperative': 'Imperative'
    };
    
    const voiceNames = {
        'potential': 'Potential',
        'causative': 'Causative'
    };
    
    // Update the current mood label
    const labelElement = document.getElementById('currentMoodLabel');
    if (labelElement) {
        const baseLabel = formLabels[currentMood] || 'Plain';
        if (activeVoice) {
            labelElement.textContent = `${voiceNames[activeVoice]} ${baseLabel}`;
        } else {
            labelElement.textContent = baseLabel;
        }
    }
}

// Toggle mood selection
function toggleMood(mood) {
    currentMood = mood;
    
    // Update button states
    const moodButtons = ['plain', 'te', 'volitional', 'conditional', 'desiderative', 'deontic', 'imperative'];
    moodButtons.forEach(m => {
        const btn = document.getElementById(m + 'MoodBtn');
        if (btn) {
            if (m === mood) {
                btn.classList.add('active');
            } else {
                btn.classList.remove('active');
            }
        }
    });
    
    // Update labels and descriptions
    updateFormLabels();
    updateMoodDescription();
    
    // Update grid cells
    if (currentVerb && conjugateVerbGlobal) {
        updateGridCells();
    }
}

// Update mood description based on current mood
function updateMoodDescription() {
    const descriptions = {
        'plain': 'I verb',
        'te': 'be verb-ing',
        'volitional': 'Let\'s verb',
        'conditional': 'If I verb',
        'desiderative': 'I want to verb',
        'deontic': 'I must verb',
        'imperative': 'Verb!'
    };
    
    const descriptionElement = document.getElementById('englishDescription');
    if (descriptionElement) {
        descriptionElement.textContent = descriptions[currentMood] || 'I verb';
    }
    
    // Update cell onclick handlers
    const formalities = ['casual', 'standard', 'polite', 'formal'];
    formalities.forEach(formality => {
        const cell = document.getElementById(formality + 'Cell');
        if (cell) {
            cell.setAttribute('data-form', currentMood);
            cell.setAttribute('onclick', `openConjugationModal('${currentMood}', '${formality}')`);
        }
    });
}

// Toggle control button
function toggleControl(controlName) {
    // Handle mutual exclusivity for voice buttons
    if (controlName === 'potential' && controls.potential === false) {
        controls.causative = false; // Turn off causative if potential is being turned on
        const causativeBtn = document.getElementById('causativeBtn');
        if (causativeBtn) causativeBtn.classList.remove('active');
    } else if (controlName === 'causative' && controls.causative === false) {
        controls.potential = false; // Turn off potential if causative is being turned on
        const potentialBtn = document.getElementById('potentialBtn');
        if (potentialBtn) potentialBtn.classList.remove('active');
    }
    
    controls[controlName] = !controls[controlName];
    const btn = document.getElementById(controlName + 'Btn');
    if (controls[controlName]) {
        btn.classList.add('active');
    } else {
        btn.classList.remove('active');
    }
    
    // Update form labels
    updateFormLabels();
    
    // Re-conjugate the verb if we have one
    if (currentVerb && conjugateVerbGlobal) {
        conjugateVerbGlobal(currentVerb);
    }
}

// Get conjugation endings for a specific form
function getConjugationEndings(form, formality) {
    const endings = {
        plain: {
            casual: { 'い・え　る': '', 'る': '', 'う': '', 'く': '', 'す': '', 'つ': '', 'ぬ': '', 'ぶ': '', 'む': '', 'ぐ': '' },
            standard: { 'い・え　る': '', 'る': '', 'う': '', 'く': '', 'す': '', 'つ': '', 'ぬ': '', 'ぶ': '', 'む': '', 'ぐ': '' },
            polite: { 'い・え　る': 'ます', 'る': 'ます', 'う': 'います', 'く': 'きます', 'す': 'します', 'つ': 'ちます', 'ぬ': 'にます', 'ぶ': 'びます', 'む': 'みます', 'ぐ': 'ぎます' },
            formal: { 'い・え　る': 'ます', 'る': 'ます', 'う': 'います', 'く': 'きます', 'す': 'します', 'つ': 'ちます', 'ぬ': 'にます', 'ぶ': 'びます', 'む': 'みます', 'ぐ': 'ぎます' }
        },
        te: {
            casual: { 'い・え　る': 'て', 'る': 'って', 'う': 'って', 'く': 'いて', 'す': 'して', 'つ': 'って', 'ぬ': 'んで', 'ぶ': 'んで', 'む': 'んで', 'ぐ': 'いで' },
            standard: { 'い・え　る': 'て', 'る': 'って', 'う': 'って', 'く': 'いて', 'す': 'して', 'つ': 'って', 'ぬ': 'んで', 'ぶ': 'んで', 'む': 'んで', 'ぐ': 'いで' },
            polite: { 'い・え　る': 'て', 'る': 'って', 'う': 'って', 'く': 'いて', 'す': 'して', 'つ': 'って', 'ぬ': 'んで', 'ぶ': 'んで', 'む': 'んで', 'ぐ': 'いで' },
            formal: { 'い・え　る': 'て', 'る': 'って', 'う': 'って', 'く': 'いて', 'す': 'して', 'つ': 'って', 'ぬ': 'んで', 'ぶ': 'んで', 'む': 'んで', 'ぐ': 'いで' }
        },
        volitional: {
            casual: { 'い・え　る': 'よう', 'る': 'ろう', 'う': 'おう', 'く': 'こう', 'す': 'そう', 'つ': 'とう', 'ぬ': 'のう', 'ぶ': 'ぼう', 'む': 'もう', 'ぐ': 'ごう' },
            standard: { 'い・え　る': 'よう', 'る': 'ろう', 'う': 'おう', 'く': 'こう', 'す': 'そう', 'つ': 'とう', 'ぬ': 'のう', 'ぶ': 'ぼう', 'む': 'もう', 'ぐ': 'ごう' },
            polite: { 'い・え　る': 'ましょう', 'る': 'りましょう', 'う': 'いましょう', 'く': 'きましょう', 'す': 'しましょう', 'つ': 'ちましょう', 'ぬ': 'にましょう', 'ぶ': 'びましょう', 'む': 'みましょう', 'ぐ': 'ぎましょう' },
            formal: { 'い・え　る': 'ましょう', 'る': 'りましょう', 'う': 'いましょう', 'く': 'きましょう', 'す': 'しましょう', 'つ': 'ちましょう', 'ぬ': 'にましょう', 'ぶ': 'びましょう', 'む': 'みましょう', 'ぐ': 'ぎましょう' }
        },
        conditional: {
            casual: { 'い・え　る': 'たら', 'る': 'ったら', 'う': 'ったら', 'く': 'いたら', 'す': 'したら', 'つ': 'ったら', 'ぬ': 'んだら', 'ぶ': 'んだら', 'む': 'んだら', 'ぐ': 'いだら' },
            standard: { 'い・え　る': 'れば', 'る': 'れば', 'う': 'えば', 'く': 'けば', 'す': 'せば', 'つ': 'てば', 'ぬ': 'ねば', 'ぶ': 'べば', 'む': 'めば', 'ぐ': 'げば' },
            polite: { 'い・え　る': 'たら', 'る': 'ったら', 'う': 'ったら', 'く': 'いたら', 'す': 'したら', 'つ': 'ったら', 'ぬ': 'んだら', 'ぶ': 'んだら', 'む': 'んだら', 'ぐ': 'いだら' },
            formal: { 'い・え　る': 'れば', 'る': 'れば', 'う': 'えば', 'く': 'けば', 'す': 'せば', 'つ': 'てば', 'ぬ': 'ねば', 'ぶ': 'べば', 'む': 'めば', 'ぐ': 'げば' }
        },
        causative: {
            casual: { 'い・え　る': 'させる', 'る': 'らせる', 'う': 'わせる', 'く': 'かせる', 'す': 'させる', 'つ': 'たせる', 'ぬ': 'なせる', 'ぶ': 'ばせる', 'む': 'ませる', 'ぐ': 'がせる' },
            standard: { 'い・え　る': 'させる', 'る': 'らせる', 'う': 'わせる', 'く': 'かせる', 'す': 'させる', 'つ': 'たせる', 'ぬ': 'なせる', 'ぶ': 'ばせる', 'む': 'ませる', 'ぐ': 'がせる' },
            polite: { 'い・え　る': 'させます', 'る': 'らせます', 'う': 'わせます', 'く': 'かせます', 'す': 'させます', 'つ': 'たせます', 'ぬ': 'なせます', 'ぶ': 'ばせます', 'む': 'ませます', 'ぐ': 'がせます' },
            formal: { 'い・え　る': 'させます', 'る': 'らせます', 'う': 'わせます', 'く': 'かせます', 'す': 'させます', 'つ': 'たせます', 'ぬ': 'なせます', 'ぶ': 'ばせます', 'む': 'ませます', 'ぐ': 'がせます' }
        },
        potential: {
            casual: { 'い・え　る': 'られる', 'る': 'れる', 'う': 'える', 'く': 'ける', 'す': 'せる', 'つ': 'てる', 'ぬ': 'ねる', 'ぶ': 'べる', 'む': 'める', 'ぐ': 'げる' },
            standard: { 'い・え　る': 'られる', 'る': 'れる', 'う': 'える', 'く': 'ける', 'す': 'せる', 'つ': 'てる', 'ぬ': 'ねる', 'ぶ': 'べる', 'む': 'める', 'ぐ': 'げる' },
            polite: { 'い・え　る': 'られます', 'る': 'れます', 'う': 'えます', 'く': 'けます', 'す': 'せます', 'つ': 'てます', 'ぬ': 'ねます', 'ぶ': 'べます', 'む': 'めます', 'ぐ': 'げます' },
            formal: { 'い・え　る': 'られます', 'る': 'れます', 'う': 'えます', 'く': 'けます', 'す': 'せます', 'つ': 'てます', 'ぬ': 'ねます', 'ぶ': 'べます', 'む': 'めます', 'ぐ': 'げます' }
        },
        desiderative: {
            casual: { 'い・え　る': 'たい', 'る': 'りたい', 'う': 'いたい', 'く': 'きたい', 'す': 'したい', 'つ': 'ちたい', 'ぬ': 'にたい', 'ぶ': 'びたい', 'む': 'みたい', 'ぐ': 'ぎたい' },
            standard: { 'い・え　る': 'たい', 'る': 'りたい', 'う': 'いたい', 'く': 'きたい', 'す': 'したい', 'つ': 'ちたい', 'ぬ': 'にたい', 'ぶ': 'びたい', 'む': 'みたい', 'ぐ': 'ぎたい' },
            polite: { 'い・え　る': 'たいです', 'る': 'りたいです', 'う': 'いたいです', 'く': 'きたいです', 'す': 'したいです', 'つ': 'ちたいです', 'ぬ': 'にたいです', 'ぶ': 'びたいです', 'む': 'みたいです', 'ぐ': 'ぎたいです' },
            formal: { 'い・え　る': 'たいです', 'る': 'りたいです', 'う': 'いたいです', 'く': 'きたいです', 'す': 'したいです', 'つ': 'ちたいです', 'ぬ': 'にたいです', 'ぶ': 'びたいです', 'む': 'みたいです', 'ぐ': 'ぎたいです' }
        },
        deontic: {
            casual: { 'い・え　る': 'なければならない', 'る': 'らなければならない', 'う': 'わなければならない', 'く': 'かなければならない', 'す': 'さなければならない', 'つ': 'たなければならない', 'ぬ': 'ななければならない', 'ぶ': 'ばなければならない', 'む': 'まなければならない', 'ぐ': 'がなければならない' },
            standard: { 'い・え　る': 'なければならない', 'る': 'らなければならない', 'う': 'わなければならない', 'く': 'かなければならない', 'す': 'さなければならない', 'つ': 'たなければならない', 'ぬ': 'ななければならない', 'ぶ': 'ばなければならない', 'む': 'まなければならない', 'ぐ': 'がなければならない' },
            polite: { 'い・え　る': 'なければなりません', 'る': 'らなければなりません', 'う': 'わなければなりません', 'く': 'かければなりません', 'す': 'さなければなりません', 'つ': 'たなければなりません', 'ぬ': 'ななければなりません', 'ぶ': 'ばなければなりません', 'む': 'まなければなりません', 'ぐ': 'がなければなりません' },
            formal: { 'い・え　る': 'なければなりません', 'る': 'らなければなりません', 'う': 'わなければなりません', 'く': 'かければなりません', 'す': 'さなければなりません', 'つ': 'たなければなりません', 'ぬ': 'ななければなりません', 'ぶ': 'ばなければなりません', 'む': 'まなければなりません', 'ぐ': 'がなければなりません' }
        },
        imperative: {
            casual: { 'い・え　る': 'ろ', 'る': 'れ', 'う': 'え', 'く': 'け', 'す': 'せ', 'つ': 'て', 'ぬ': 'ね', 'ぶ': 'べ', 'む': 'め', 'ぐ': 'げ' },
            standard: { 'い・え　る': 'ろ', 'る': 'れ', 'う': 'え', 'く': 'け', 'す': 'せ', 'つ': 'て', 'ぬ': 'ね', 'ぶ': 'べ', 'む': 'め', 'ぐ': 'げ' },
            polite: { 'い・え　る': 'なさい', 'る': 'りなさい', 'う': 'いなさい', 'く': 'きなさい', 'す': 'しなさい', 'つ': 'ちなさい', 'ぬ': 'になさい', 'ぶ': 'びなさい', 'む': 'みなさい', 'ぐ': 'ぎなさい' },
            formal: { 'い・え　る': 'なさい', 'る': 'りなさい', 'う': 'いなさい', 'く': 'きなさい', 'す': 'しなさい', 'つ': 'ちなさい', 'ぬ': 'になさい', 'ぶ': 'びなさい', 'む': 'みなさい', 'ぐ': 'ぎなさい' }
        }
    };

    // Handle irregular verbs
    const irregularEndings = {
        'ある': { plain: { casual: 'ある', standard: 'ある', polite: 'あります', formal: 'あります' },
                  te: { casual: 'あって', standard: 'あって', polite: 'あって', formal: 'あって' },
                  volitional: { casual: 'あろう', standard: 'あろう', polite: 'ありましょう', formal: 'ありましょう' },
                  conditional: { casual: 'あったら', standard: 'あれば', polite: 'あったら', formal: 'あれば' },
                  causative: { casual: 'あらせる', standard: 'あらせる', polite: 'あらせます', formal: 'あらせます' },
                  potential: { casual: 'あれる', standard: 'あれる', polite: 'あれます', formal: 'あれます' },
                  desiderative: { casual: 'ありたい', standard: 'ありたい', polite: 'ありたいです', formal: 'ありたいです' },
                  deontic: { casual: 'あらなければならない', standard: 'あらなければならない', polite: 'あらなければなりません', formal: 'あらなければなりません' },
                  imperative: { casual: 'あれ', standard: 'あれ', polite: 'ありなさい', formal: 'ありなさい' } },
        'いる': { plain: { casual: 'いる', standard: 'いる', polite: 'います', formal: 'います' },
                  te: { casual: 'いて', standard: 'いて', polite: 'いて', formal: 'いて' },
                  volitional: { casual: 'いよう', standard: 'いよう', polite: 'いましょう', formal: 'いましょう' },
                  conditional: { casual: 'いたら', standard: 'いれば', polite: 'いたら', formal: 'いれば' },
                  causative: { casual: 'いさせる', standard: 'いさせる', polite: 'いさせます', formal: 'いさせます' },
                  potential: { casual: 'いられる', standard: 'いられる', polite: 'いられます', formal: 'いられます' },
                  desiderative: { casual: 'いたい', standard: 'いたい', polite: 'いたいです', formal: 'いたいです' },
                  deontic: { casual: 'いなければならない', standard: 'いなければならない', polite: 'いなければなりません', formal: 'いなければなりません' },
                  imperative: { casual: 'いろ', standard: 'いろ', polite: 'いなさい', formal: 'いなさい' } },
        '行く': { plain: { casual: '行く', standard: '行く', polite: '行きます', formal: '行きます' },
                  te: { casual: '行って', standard: '行って', polite: '行って', formal: '行って' },
                  volitional: { casual: '行こう', standard: '行こう', polite: '行きましょう', formal: '行きましょう' },
                  conditional: { casual: '行ったら', standard: '行けば', polite: '行ったら', formal: '行けば' },
                  causative: { casual: '行かせる', standard: '行かせる', polite: '行かせます', formal: '行かせます' },
                  potential: { casual: '行ける', standard: '行ける', polite: '行けます', formal: '行けます' },
                  desiderative: { casual: '行きたい', standard: '行きたい', polite: '行きたいです', formal: '行きたいです' },
                  deontic: { casual: '行かなければならない', standard: '行かなければならない', polite: '行かなければなりません', formal: '行かなければなりません' },
                  imperative: { casual: '行け', standard: '行け', polite: '行きなさい', formal: '行きなさい' } },
        'くる': { plain: { casual: 'くる', standard: 'くる', polite: 'きます', formal: 'きます' },
                  te: { casual: 'きて', standard: 'きて', polite: 'きて', formal: 'きて' },
                  volitional: { casual: 'こよう', standard: 'こよう', polite: 'きましょう', formal: 'きましょう' },
                  conditional: { casual: 'きたら', standard: 'くれば', polite: 'きたら', formal: 'くれば' },
                  causative: { casual: 'こさせる', standard: 'こさせる', polite: 'こさせます', formal: 'こさせます' },
                  potential: { casual: 'こられる', standard: 'こられる', polite: 'こられます', formal: 'こられます' },
                  desiderative: { casual: 'きたい', standard: 'きたい', polite: 'きたいです', formal: 'きたいです' },
                  deontic: { casual: 'こなければならない', standard: 'こなければならない', polite: 'こなければなりません', formal: 'こなければなりません' },
                  imperative: { casual: 'こい', standard: 'こい', polite: 'きなさい', formal: 'きなさい' } },
        'する': { plain: { casual: 'する', standard: 'する', polite: 'します', formal: 'します' },
                  te: { casual: 'して', standard: 'して', polite: 'して', formal: 'して' },
                  volitional: { casual: 'しよう', standard: 'しよう', polite: 'しましょう', formal: 'しましょう' },
                  conditional: { casual: 'したら', standard: 'すれば', polite: 'したら', formal: 'すれば' },
                  causative: { casual: 'させる', standard: 'させる', polite: 'させます', formal: 'させます' },
                  potential: { casual: 'できる', standard: 'できる', polite: 'できます', formal: 'できます' },
                  desiderative: { casual: 'したい', standard: 'したい', polite: 'したいです', formal: 'したいです' },
                  deontic: { casual: 'しなければならない', standard: 'しなければならない', polite: 'しなければなりません', formal: 'しなければなりません' },
                  imperative: { casual: 'しろ', standard: 'しろ', polite: 'しなさい', formal: 'しなさい' } }
    };

    return { endings: endings[form]?.[formality] || {}, irregularEndings };
}

// Get full verb conjugations
function getFullVerbConjugations(form, formality) {
    const stems = {
        'い・え　る': '食べ',  // ichidan verb
        'る': '作',           // godan ru
        'う': '買',           // godan u
        'く': '歩',           // godan ku
        'す': '話',           // godan su
        'つ': '持',           // godan tsu
        'ぬ': '死',           // godan nu
        'ぶ': '飛',           // godan bu
        'む': '飲',           // godan mu
        'ぐ': '急'            // godan gu
    };

    const conjugations = getConjugationEndings(form, formality);
    const fullVerbs = {};

    for (const [ending, stem] of Object.entries(stems)) {
        const endingText = conjugations.endings[ending] || '';
        fullVerbs[ending] = stem + endingText;
    }

    return fullVerbs;
}

// Open conjugation modal
function openConjugationModal(form, formality) {
    const modal = document.getElementById('conjugationModal');
    const modalTitle = document.getElementById('modalTitle');
    const tableBody = document.getElementById('conjugationTableBody');

    // Set title
    const formNames = {
        plain: 'Plain',
        te: 'T form',
        volitional: 'Volitional',
        conditional: 'Conditional',
        causative: 'Causative',
        potential: 'Potential',
        desiderative: 'Desiderative',
        deontic: 'Deontic',
        imperative: 'Imperative'
    };

    const formalityNames = {
        casual: 'Casual',
        standard: 'Standard',
        polite: 'Polite',
        formal: 'Formal'
    };

    modalTitle.textContent = `${formNames[form]} - ${formalityNames[formality]}`;

    // Clear table
    tableBody.innerHTML = '';

    // Get conjugations
    const conjugations = getConjugationEndings(form, formality);
    const fullVerbs = getFullVerbConjugations(form, formality);

    // Add irregular verbs first
    for (const verb of verbExamples.irregular) {
        const row = document.createElement('tr');
        const plainCell = document.createElement('td');
        const verbCell = document.createElement('td');
        const endingCell = document.createElement('td');
        const fullCell = document.createElement('td');

        plainCell.textContent = verb; // Plain form is the verb itself for irregular verbs
        verbCell.textContent = verb;
        
        if (conjugations.irregularEndings[verb] && conjugations.irregularEndings[verb][form]) {
            const fullForm = conjugations.irregularEndings[verb][form][formality] || '';
            // For irregular verbs, show the full form in both ending and full verb columns
            // since irregular verbs don't have a clear stem+ending pattern
            endingCell.textContent = fullForm;
            fullCell.textContent = fullForm;
        } else {
            endingCell.textContent = '';
            fullCell.textContent = verb;
        }

        row.appendChild(plainCell);
        row.appendChild(verbCell);
        row.appendChild(endingCell);
        row.appendChild(fullCell);
        tableBody.appendChild(row);
    }

    // Add gap row
    const gapRow = document.createElement('tr');
    gapRow.style.height = '20px';
    const gapCell = document.createElement('td');
    gapCell.colSpan = 4;
    gapRow.appendChild(gapCell);
    tableBody.appendChild(gapRow);

    // Add regular verb endings
    for (const ending of verbExamples.endings) {
        const row = document.createElement('tr');
        const plainCell = document.createElement('td');
        const endingCell = document.createElement('td');
        const conjugationCell = document.createElement('td');
        const fullCell = document.createElement('td');

        // Plain form is the dictionary form of the verb
        plainCell.textContent = fullVerbExamples[ending] || '';
        endingCell.textContent = ending;
        conjugationCell.textContent = conjugations.endings[ending] || '';
        fullCell.textContent = fullVerbs[ending] || '';

        row.appendChild(plainCell);
        row.appendChild(endingCell);
        row.appendChild(conjugationCell);
        row.appendChild(fullCell);
        tableBody.appendChild(row);
    }

    // Add gap row before adjectives
    const adjectiveGapRow = document.createElement('tr');
    adjectiveGapRow.style.height = '20px';
    const adjectiveGapCell = document.createElement('td');
    adjectiveGapCell.colSpan = 4;
    adjectiveGapRow.appendChild(adjectiveGapCell);
    tableBody.appendChild(adjectiveGapRow);

    // Add adjective rows
    // Get the verb ending from する (suru) conjugation
    // For adjectives, we use する as the verb, so we need the full する conjugation
    let verbEndingForAdjective = '';
    
    // Check if we have irregular する conjugation
    if (conjugations.irregularEndings && conjugations.irregularEndings['する'] && 
        conjugations.irregularEndings['する'][form]) {
        const suruForm = conjugations.irregularEndings['する'][form][formality] || '';
        // Use the full する form (e.g., したら, して, しよう)
        // For adjectives: 早くしたら = 早く + したら
        verbEndingForAdjective = suruForm;
    } else if (conjugations.endings['す']) {
        // Use the す ending pattern (for regular す verbs like 話す)
        const suEnding = conjugations.endings['す'];
        // For す verbs, conditional is "したら", te-form is "して", etc.
        // We need to construct the full する form
        if (suEnding.startsWith('し')) {
            // Already has し, use as is (e.g., したら)
            verbEndingForAdjective = suEnding;
        } else {
            // Add し prefix for する (e.g., て -> して)
            verbEndingForAdjective = 'し' + suEnding;
        }
    } else {
        // Fallback: construct based on form
        const formEndings = {
            'plain': 'する',
            'te': 'して',
            'volitional': 'しよう',
            'conditional': 'したら',
            'causative': 'させる',
            'potential': 'できる',
            'desiderative': 'したい',
            'deontic': 'しなければならない',
            'imperative': 'しろ'
        };
        verbEndingForAdjective = formEndings[form] || 'する';
    }

    // Row 1: 早い (i-adjective)
    const hayaiRow = document.createElement('tr');
    const hayaiPlainCell = document.createElement('td');
    const hayaiEndingCell = document.createElement('td');
    const hayaiConjugationCell = document.createElement('td');
    const hayaiFullCell = document.createElement('td');

    hayaiPlainCell.textContent = '早い';
    hayaiEndingCell.textContent = 'い';
    hayaiConjugationCell.textContent = 'く' + verbEndingForAdjective;
    hayaiFullCell.textContent = '早く' + verbEndingForAdjective;

    hayaiRow.appendChild(hayaiPlainCell);
    hayaiRow.appendChild(hayaiEndingCell);
    hayaiRow.appendChild(hayaiConjugationCell);
    hayaiRow.appendChild(hayaiFullCell);
    tableBody.appendChild(hayaiRow);

    // Row 2: 静かな (na-adjective)
    const shizukanaRow = document.createElement('tr');
    const shizukanaPlainCell = document.createElement('td');
    const shizukanaEndingCell = document.createElement('td');
    const shizukanaConjugationCell = document.createElement('td');
    const shizukanaFullCell = document.createElement('td');

    shizukanaPlainCell.textContent = '静かな';
    shizukanaEndingCell.textContent = 'な';
    shizukanaConjugationCell.textContent = 'に' + verbEndingForAdjective;
    shizukanaFullCell.textContent = '静かに' + verbEndingForAdjective;

    shizukanaRow.appendChild(shizukanaPlainCell);
    shizukanaRow.appendChild(shizukanaEndingCell);
    shizukanaRow.appendChild(shizukanaConjugationCell);
    shizukanaRow.appendChild(shizukanaFullCell);
    tableBody.appendChild(shizukanaRow);

    // Show modal
    modal.style.display = 'block';
}

// Close conjugation modal
function closeConjugationModal() {
    const modal = document.getElementById('conjugationModal');
    modal.style.display = 'none';
}

// Close modal when clicking outside
window.onclick = function(event) {
    const modal = document.getElementById('conjugationModal');
    if (event.target == modal) {
        modal.style.display = 'none';
    }
}

// Helper function to conjugate a verb form (treats voice forms as ichidan verbs)
function conjugateVoiceForm(voiceForm, form) {
    if (!voiceForm || !voiceForm.endsWith('る')) {
        return '';
    }
    
    // Voice forms (potential/causative) end in る, so treat as ichidan
    const stem = voiceForm.slice(0, -1); // Remove る
    
    switch (form) {
        case 'plain':
            return voiceForm;
        case 'te':
            return stem + 'て';
        case 'volitional':
            return stem + 'よう';
        case 'conditional':
            return stem + 'れば';
        case 'desiderative':
            return stem + 'たい';
        case 'deontic':
            return stem + 'なければならない';
        case 'imperative':
            return stem + 'ろ';
        default:
            return voiceForm;
    }
}

// Get conjugated form for a specific form and formality
function getConjugatedForm(form, formality) {
    if (!currentVerbData || !currentVerbData.conjugations) {
        return '';
    }

    const conjugations = currentVerbData.conjugations;
    const activeVoice = getActiveVoice();
    let result = '';

    // If a voice is active, we need to get the voice form first, then apply the conjugation
    let baseVoiceForm = null;
    if (activeVoice === 'potential') {
        const potential = conjugations.tenses?.modals?.potential;
        if (potential && potential.japanese) {
            baseVoiceForm = potential.japanese;
        }
    } else if (activeVoice === 'causative') {
        const causative = conjugations.tenses?.modals?.causative;
        if (causative && causative.japanese) {
            baseVoiceForm = causative.japanese;
        }
    }

    // Map our form names to the API response structure
    if (form === 'plain') {
        if (baseVoiceForm) {
            // Use the voice form as the plain form
            result = baseVoiceForm;
            // For polite/formal, convert to ます form
            if ((formality === 'polite' || formality === 'formal') && !result.endsWith('ます')) {
                if (result.endsWith('る')) {
                    result = result.slice(0, -1) + 'ます';
                } else if (result.endsWith('られる')) {
                    result = result.slice(0, -2) + 'られます';
                } else if (result.endsWith('せる')) {
                    result = result.slice(0, -2) + 'せます';
                } else if (result.endsWith('させる')) {
                    result = result.slice(0, -3) + 'させます';
                }
            }
        } else {
            // For polite/formal, use polite conjugations if available
            if ((formality === 'polite' || formality === 'formal') && currentVerbData.politeConjugations) {
                const present = currentVerbData.politeConjugations.tenses?.time?.present;
                if (present && present.japanese) {
                    result = present.japanese;
                }
            } else {
                const present = conjugations.tenses?.time?.present;
                if (present && present.japanese) {
                    result = present.japanese;
                }
            }
        }
    } else if (form === 'te') {
        if (baseVoiceForm) {
            // Conjugate the voice form
            result = conjugateVoiceForm(baseVoiceForm, 'te');
            // For polite/formal, add います
            if ((formality === 'polite' || formality === 'formal') && !result.endsWith('います')) {
                result = result + 'います';
            }
        } else {
            // Te-form is in progressive aspect (ている), extract just the te-form
            let progressive;
            if ((formality === 'polite' || formality === 'formal') && currentVerbData.politeConjugations) {
                progressive = currentVerbData.politeConjugations.tenses?.aspect?.progressive;
            } else {
                progressive = conjugations.tenses?.aspect?.progressive;
            }
            if (progressive && progressive.japanese) {
                result = progressive.japanese;
                // Remove いる from the end to get just the te-form
                if (result.endsWith('ている')) {
                    result = result.slice(0, -2); // Remove いる
                } else if (result.endsWith('います')) {
                    result = result.slice(0, -3); // Remove います
                } else if (result.endsWith('いません')) {
                    result = result.slice(0, -4); // Remove いません
                }
            }
        }
    } else if (form === 'volitional') {
        if (baseVoiceForm) {
            // Conjugate the voice form
            result = conjugateVoiceForm(baseVoiceForm, 'volitional');
            // For polite/formal, convert to ましょう
            if ((formality === 'polite' || formality === 'formal') && !result.endsWith('ましょう')) {
                if (result.endsWith('よう')) {
                    result = result.slice(0, -1) + 'ましょう';
                }
            }
        } else {
            const volitional = conjugations.tenses?.mood?.volitional;
            if (volitional && volitional.japanese) {
                result = volitional.japanese;
                // For polite/formal, convert to ましょう
                if ((formality === 'polite' || formality === 'formal') && !result.endsWith('ましょう')) {
                    // Convert volitional to polite form
                    if (result.endsWith('よう')) {
                        result = result.slice(0, -1) + 'ましょう';
                    } else if (result.endsWith('おう')) {
                        result = result.slice(0, -1) + 'いましょう';
                    } else if (result.endsWith('ろう')) {
                        result = result.slice(0, -1) + 'りましょう';
                    }
                }
            }
        }
    } else if (form === 'conditional') {
        if (baseVoiceForm) {
            // Conjugate the voice form
            result = conjugateVoiceForm(baseVoiceForm, 'conditional');
            // For standard/formal, use ば form (already in れば form)
            if ((formality === 'standard' || formality === 'formal')) {
                // Already in ば form (れば)
            } else {
                // For casual/polite, use たら form
                // Convert れば to たら by using past form
                const stem = baseVoiceForm.slice(0, -1);
                result = stem + 'たら';
            }
        } else {
            const conditional = conjugations.tenses?.mood?.conditional;
            if (conditional && conditional.japanese) {
                result = conditional.japanese;
                // For standard/formal, use ば form if available in alts
                if ((formality === 'standard' || formality === 'formal') && conditional.alts && conditional.alts.length > 0) {
                    // Check if there's a ば form in alternatives
                    const baForm = conditional.alts.find(alt => alt.includes('ば'));
                    if (baForm) {
                        result = baForm;
                    }
                }
            }
        }
    } else if (form === 'desiderative') {
        if (baseVoiceForm) {
            // Conjugate the voice form
            result = conjugateVoiceForm(baseVoiceForm, 'desiderative');
            // For polite/formal, add です
            if ((formality === 'polite' || formality === 'formal') && !result.endsWith('です')) {
                result = result + 'です';
            }
        } else {
            const desiderative = conjugations.tenses?.desire?.subject;
            if (desiderative && desiderative.japanese) {
                result = desiderative.japanese;
                // For polite/formal, add です
                if ((formality === 'polite' || formality === 'formal') && !result.endsWith('です')) {
                    result = result + 'です';
                }
            }
        }
    } else if (form === 'deontic') {
        if (baseVoiceForm) {
            // Conjugate the voice form
            result = conjugateVoiceForm(baseVoiceForm, 'deontic');
            // For polite/formal, convert ならない to なりません
            if ((formality === 'polite' || formality === 'formal') && result.includes('ならない')) {
                result = result.replace('ならない', 'なりません');
            }
        } else {
            const deontic = conjugations.tenses?.modals?.deontic;
            if (deontic && deontic.japanese) {
                result = deontic.japanese;
                // For polite/formal, convert ならない to なりません
                if ((formality === 'polite' || formality === 'formal') && result.includes('ならない')) {
                    result = result.replace('ならない', 'なりません');
                }
            }
        }
    } else if (form === 'imperative') {
        if (baseVoiceForm) {
            // Conjugate the voice form
            result = conjugateVoiceForm(baseVoiceForm, 'imperative');
            // For polite/formal, convert to なさい form
            if ((formality === 'polite' || formality === 'formal') && !result.endsWith('なさい')) {
                if (result.endsWith('ろ')) {
                    result = result.slice(0, -1) + 'りなさい';
                }
            }
        } else {
            const imperative = conjugations.tenses?.mood?.imperative;
            if (imperative && imperative.japanese) {
                result = imperative.japanese;
                // For polite/formal, convert to なさい form
                if ((formality === 'polite' || formality === 'formal') && !result.endsWith('なさい')) {
                    // Convert imperative to polite form
                    if (result.endsWith('ろ')) {
                        result = result.slice(0, -1) + 'りなさい';
                    } else if (result.endsWith('え')) {
                        result = result.slice(0, -1) + 'いなさい';
                    } else if (result.endsWith('け')) {
                        result = result.slice(0, -1) + 'きなさい';
                    } else if (result.endsWith('せ')) {
                        result = result.slice(0, -1) + 'しなさい';
                    } else if (result.endsWith('て')) {
                        result = result.slice(0, -1) + 'ちなさい';
                    } else if (result.endsWith('ね')) {
                        result = result.slice(0, -1) + 'になさい';
                    } else if (result.endsWith('べ')) {
                        result = result.slice(0, -1) + 'びなさい';
                    } else if (result.endsWith('め')) {
                        result = result.slice(0, -1) + 'みなさい';
                    } else if (result.endsWith('げ')) {
                        result = result.slice(0, -1) + 'ぎなさい';
                    }
                }
            }
        }
    }

    return result;
}

// Convert verb to masu form (simplified - uses API response when available)
function convertToMasuForm(verb) {
    // This is a simplified conversion - in practice, we'd use the API response
    // For now, just return the verb as-is if it doesn't already end with ます
    if (verb.endsWith('ます') || verb.endsWith('です')) {
        return verb;
    }
    // Try to get from API response if available
    if (currentVerbData && currentVerbData.conjugations && 
        currentVerbData.conjugations.tenses && 
        currentVerbData.conjugations.tenses.time && 
        currentVerbData.conjugations.tenses.time.present) {
        // Check if there's a polite version
        const present = currentVerbData.conjugations.tenses.time.present;
        if (present.japanese && present.japanese.endsWith('ます')) {
            return present.japanese;
        }
    }
    return verb;
}

// Convert to mashou form (volitional polite)
function convertToMashouForm(verb) {
    if (verb.endsWith('ましょう')) {
        return verb;
    }
    // Try to get from API response
    if (currentVerbData && currentVerbData.conjugations && 
        currentVerbData.conjugations.tenses && 
        currentVerbData.conjugations.tenses.mood && 
        currentVerbData.conjugations.tenses.mood.volitional) {
        const volitional = currentVerbData.conjugations.tenses.mood.volitional;
        if (volitional.japanese && volitional.japanese.endsWith('ましょう')) {
            return volitional.japanese;
        }
    }
    return verb;
}

// Update all grid cells with conjugated forms
function updateGridCells() {
    const formalities = ['casual', 'standard', 'polite', 'formal'];

    formalities.forEach(formality => {
        const cell = document.getElementById(formality + 'Cell');
        if (cell) {
            const conjugated = getConjugatedForm(currentMood, formality);
            cell.textContent = conjugated || '';
        }
    });
}

// Handle verb input form submission
document.addEventListener('DOMContentLoaded', function() {
    const verbForm = document.getElementById('verbInputForm');
    const verbInput = document.getElementById('verbInput');
    const verbError = document.getElementById('verbError');

    // Load default verb on page load
    if (verbInput.value) {
        // Initialize mood selection
        toggleMood('plain');
        updateFormLabels(); // Initialize form labels
        conjugateVerb(verbInput.value);
    }

    verbForm.addEventListener('submit', async function(e) {
        e.preventDefault();
        
        const verb = verbInput.value.trim();
        if (!verb) {
            showVerbError('Please enter a verb');
            return;
        }

        hideVerbError();
        await conjugateVerb(verb);
    });

    function showVerbError(message) {
        verbError.textContent = message;
        verbError.style.display = 'block';
    }

    function hideVerbError() {
        verbError.style.display = 'none';
    }

    async function conjugateVerb(verb) {
        try {
            // Call API without polite flag first to get casual forms
            const response = await fetch('/api/verb/conjugate', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ 
                    verb: verb,
                    negative: controls.negative,
                    polite: false
                })
            });

            const data = await response.json();

            if (!data.valid) {
                showVerbError(data.error || 'Invalid verb');
                return;
            }

            // Also get polite forms
            const politeResponse = await fetch('/api/verb/conjugate', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ 
                    verb: verb,
                    negative: controls.negative,
                    polite: true
                })
            });

            const politeData = await politeResponse.json();
            
            // Merge polite forms into the data
            if (politeData.valid && politeData.conjugations) {
                data.politeConjugations = politeData.conjugations;
            }

            currentVerb = verb;
            currentVerbData = data;
            updateGridCells();

        } catch (error) {
            showVerbError('Failed to conjugate verb. Please try again.');
            console.error('Error:', error);
        }
    }
    
    // Make conjugateVerb available globally for toggleControl
    conjugateVerbGlobal = conjugateVerb;
    window.conjugateVerb = conjugateVerb;
});

