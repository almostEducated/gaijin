// Verb Grid Conjugation JavaScript

// Current verb being conjugated
let currentVerb = '食べる';
let currentVerbData = null;

// Current selected mood
let currentMood = 'plain';

// Global reference to conjugateVerb function (set in DOMContentLoaded)
let conjugateVerbGlobal = null;

// State for control buttons
const controls = {
    // Person (mutually exclusive)
    first: false,
    second: false,
    third: false,
    // Voice (mutually exclusive: potential || passive, potential || causative)
    potential: false,
    passive: false,
    causative: false,
    // Aspect (can have multiple)
    continuous: false,
    completion: false,
    resultant: false,
    // Tense (mutually exclusive: simple || past)
    simple: false,
    past: false,
    // Mood (mutually exclusive: plain || te || volitional || imperative || deontic || desiderative)
    plain: false,
    te: false,
    volitional: false,
    imperative: false,
    deontic: false,
    desiderative: false,
    conditional: false,
    // Other
    negative: false
};

// Mode state: 'linguistic' or 'common'
let currentMode = 'Linguistic';

// Constants are now loaded from separate files:
// - verbConstants.js (verbExamples, fullVerbExamples, endingToIndex, ichidanRoots)
// - conjugationPatterns.js (conjugationPatterns)
// - conjugationTranslations.js (conjugationTranslations)
// - conjugationEndings.js (getConjugationEndingsData function)

// CSV-based conjugation patterns (from Japanese_FULL_map_reorganized.csv for casual)
// Structure: { construction: { base: string, root: string[], conjugation: string, alts: string[] } }
// Note: conjugationPatterns is now defined in conjugationPatterns.js
// conjugationPatterns is loaded from conjugationPatterns.js
    // Tier 1 - Basic forms

// Get verb ending index (handles ichidan vs godan ru)
function getVerbEndingIndex(verb) {
    const lastChar = verb[verb.length - 1];
    if (lastChar === 'る' && verb.length > 1) {
        const secondLast = verb[verb.length - 2];
        // Ichidan verbs end in いる or える sounds
        if (ichidanRoots[secondLast]) {
            return 9; // ichidan
        }
        return 2; // godan ru
    }
    return endingToIndex[lastChar] !== undefined ? endingToIndex[lastChar] : 2;
}

// Build construction name from active controls
// CSV column 2 order: tense → negation → voice → aspect → mood
function getConstructionName() {
    const parts = [];
    
    // Build construction name in the order it appears in CSV column 2:
    // tense → negation → voice → aspect → mood
    
    // 1. TENSE (past or simple present - simple present is default/implied)
    const isPast = controls.past;
    
    // 2. NEGATION
    const isNegative = controls.negative;
    
    // 3. VOICE (potential, passive, causative)
    const hasPotential = controls.potential;
    const hasPassive = controls.passive;
    const hasCausative = controls.causative;
    
    // 4. ASPECT (continuous, completion, resultant)
    const hasContinuous = controls.continuous;
    const hasCompletion = controls.completion;
    const hasResultant = controls.resultant;
    
    // 5. MOOD (conditional, desiderative, deontic)
    const hasConditional = controls.conditional;
    const hasDesiderative = controls.desiderative;
    const hasDeontic = controls.deontic;
    
    // Build name in CSV order: tense → negation → voice → aspect → mood
    
    // 1. Start with tense
    if (isPast) parts.push('past');
    
    // 2. Add negation
    if (isNegative) parts.push('negative');
    
    // 3. Add voice (mutually exclusive in practice, but handle combinations from CSV)
    // Standardize to "causative passive" order when both are present
    if (hasPotential && hasPassive) {
        parts.push('potential');
        parts.push('passive');
    } else if (hasPotential && hasCausative) {
        parts.push('potential');
        parts.push('causative');
    } else if (hasCausative && hasPassive) {
        // Always use "causative passive" order for consistency
        parts.push('causative');
        parts.push('passive');
    } else if (hasPotential) {
        parts.push('potential');
    } else if (hasPassive) {
        parts.push('passive');
    } else if (hasCausative) {
        parts.push('causative');
    }
    
    // 4. Add aspect (can have multiple, but CSV shows specific combinations)
    // Note: continuous, completion, resultant use T Form base
    if (hasCompletion && hasContinuous) {
        parts.push('completion');
        parts.push('continuous');
    } else if (hasCompletion && hasResultant) {
        parts.push('completion');
        parts.push('resultant');
    } else if (hasContinuous && hasResultant) {
        parts.push('continuous');
        parts.push('resultant');
    } else if (hasContinuous) {
        parts.push('continuous');
    } else if (hasCompletion) {
        parts.push('completion');
    } else if (hasResultant) {
        parts.push('resultant');
    }
    
    // 5. Add mood (conditional, desiderative, deontic are modifiers)
    // Note: volitional and imperative are handled separately as base constructions
    if (hasConditional && hasDesiderative) {
        parts.push('conditional');
        parts.push('desiderative');
    } else if (hasConditional && hasDeontic) {
        parts.push('conditional');
        parts.push('deontic');
    } else if (hasDesiderative) {
        parts.push('desiderative');
    } else if (hasDeontic) {
        parts.push('deontic');
    } else if (hasConditional) {
        parts.push('conditional');
    }
    
    // If no modifiers, return basic form
    if (parts.length === 0) return 'simple present';
    
    // Join parts
    const constructionName = parts.join(' ');
    
    // Check if this exact pattern exists, if not try simpler combinations
    if (conjugationPatterns[constructionName]) {
        return constructionName;
    }
    
    // Try to find closest match or build from components
    // For now, return what we have - we'll handle missing patterns in applyConjugationPattern
    return constructionName;
}

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
    // Handle mutual exclusivity: plain || te || volitional || imperative || deontic || desiderative || conditional
    const moodButtons = ['plain', 'te', 'volitional', 'conditional', 'desiderative', 'deontic', 'imperative'];
    
    // Ensure at least one mood is always on (self-check: clicking the only active one keeps it on)
    const activeMoodCount = moodButtons.reduce((count, m) => count + (controls[m] ? 1 : 0), 0);
    const isCurrentlyActive = controls[mood];
    
    // If this is the only active mood and it's being clicked, keep it on (don't toggle)
    if (isCurrentlyActive && activeMoodCount === 1) {
        // Don't toggle - keep it on
        return;
    }
    
    // If this mood is already active (but not the only one), toggle it off
    if (controls[mood]) {
        controls[mood] = false;
        currentMood = null;
    } else {
        // Turn off all other moods
        moodButtons.forEach(m => {
            controls[m] = false;
            const btn = document.getElementById(m + 'MoodBtn');
            if (btn) btn.classList.remove('active');
        });
        // Turn on selected mood
        controls[mood] = true;
        currentMood = mood;
        
        // Special logic for T form: reset to first person, simple present, and clear voice and aspect
        if (mood === 'te') {
            // Reset to first person
            controls.first = true;
            controls.second = false;
            controls.third = false;
            const firstBtn = document.getElementById('firstPersonBtn');
            const secondBtn = document.getElementById('secondPersonBtn');
            const thirdBtn = document.getElementById('thirdPersonBtn');
            if (firstBtn) firstBtn.classList.add('active');
            if (secondBtn) secondBtn.classList.remove('active');
            if (thirdBtn) thirdBtn.classList.remove('active');
            
            // Reset to simple present
            controls.simple = true;
            controls.past = false;
            const simpleBtn = document.getElementById('simpleBtn');
            const pastBtn = document.getElementById('pastBtn');
            if (simpleBtn) simpleBtn.classList.add('active');
            if (pastBtn) pastBtn.classList.remove('active');
            
            // Clear voice (potential, passive, causative)
            controls.potential = false;
            controls.passive = false;
            controls.causative = false;
            const potentialBtn = document.getElementById('potentialBtn');
            const passiveBtn = document.getElementById('passiveBtn');
            const causativeBtn = document.getElementById('causativeBtn');
            if (potentialBtn) potentialBtn.classList.remove('active');
            if (passiveBtn) passiveBtn.classList.remove('active');
            if (causativeBtn) causativeBtn.classList.remove('active');
            
            // Clear aspect (continuous, completion, resultant)
            controls.continuous = false;
            controls.completion = false;
            controls.resultant = false;
            const continuousBtn = document.getElementById('continuousBtn');
            const completionBtn = document.getElementById('completionBtn');
            const resultantBtn = document.getElementById('resultantBtn');
            if (continuousBtn) continuousBtn.classList.remove('active');
            if (completionBtn) completionBtn.classList.remove('active');
            if (resultantBtn) resultantBtn.classList.remove('active');
        }
    }
    
    // Update button states
    moodButtons.forEach(m => {
        const btn = document.getElementById(m + 'MoodBtn');
        if (btn) {
            if (controls[m]) {
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
        conjugateVerbGlobal(currentVerb);
    } else if (currentVerb) {
        updateGridCells();
    }
}

// CSV-based translation mappings (from Japanese Conjugation List Translations.csv)
// Ordered alphabetically
// Note: conjugationTranslations is now defined in conjugationTranslations.js
// conjugationTranslations is loaded from conjugationTranslations.js

// Map button combinations to CSV construction names
function getCSVConstructionName() {
    const parts = [];
    
    // Determine tense
    const isPast = controls.past;
    const isNegative = controls.negative;
    
    // Determine voice
    const hasPotential = controls.potential;
    const hasPassive = controls.passive;
    const hasCausative = controls.causative;
    
    // Determine aspect
    const hasContinuous = controls.continuous;
    const hasCompletion = controls.completion;
    const hasResultant = controls.resultant;
    
    // Determine mood
    const hasConditional = controls.conditional;
    const hasDesiderative = controls.desiderative;
    const hasDeontic = controls.deontic;
    const hasVolitional = controls.volitional;
    const hasImperative = controls.imperative;
    const hasTe = controls.te;
    
    // Handle special moods first (imperative, volitional, te-form)
    if (hasImperative && !hasPotential && !hasPassive && !hasCausative && !hasContinuous && !hasCompletion && !hasResultant && !hasConditional && !hasDesiderative && !hasDeontic) {
        if (isNegative) {
            return 'Negative Volitional'; // Note: CSV doesn't have "Negative Imperative", using closest match
        }
        return 'Imperative';
    }
    
    if (hasVolitional && !hasPotential && !hasPassive && !hasCausative && !hasContinuous && !hasCompletion && !hasResultant && !hasConditional && !hasDesiderative && !hasDeontic) {
        if (isNegative) {
            return 'Negative Volitional';
        }
        return 'Volitional';
    }
    
    if (hasTe && !hasContinuous && !hasCompletion && !hasResultant && !hasPotential && !hasPassive && !hasCausative && !hasConditional && !hasDesiderative && !hasDeontic) {
        return 'T-Form';
    }
    
    // Build construction name in order: tense → negation → voice → aspect → mood
    // 1. Tense
    if (isPast) parts.push('Past');
    
    // 2. Negation
    if (isNegative) parts.push('Negative');
    
    // 3. Voice (mutually exclusive in practice, but handle combinations)
    // Standardize to "Causative Passive" order when both are present
    if (hasCausative && hasPassive) {
        parts.push('Causative');
        parts.push('Passive');
    } else if (hasCausative) {
        parts.push('Causative');
    } else if (hasPassive) {
        parts.push('Passive');
    } else if (hasPotential) {
        parts.push('Potential');
    }
    
    // 4. Aspect
    if (hasCompletion && hasContinuous) {
        parts.push('Completion');
        parts.push('Continuous');
    } else if (hasCompletion) {
        parts.push('Completion');
    } else if (hasContinuous) {
        parts.push('Continuous');
    } else if (hasResultant) {
        parts.push('Resultant');
    }
    
    // 5. Mood
    if (hasConditional) {
        // CSV has typo "Conditional" but we'll use "Conditional" and search will handle it
        parts.push('Conditional');
    } else if (hasDesiderative) {
        parts.push('Desiderative');
    } else if (hasDeontic) {
        parts.push('Deontic');
    }
    
    // If no parts, check if we have a basic mood selected
    if (parts.length === 0) {
        // If plain mood with no modifiers, check tense
        if (currentMood === 'plain' || !currentMood) {
            if (isPast) {
                return 'Past';
            }
            // No CSV entry for simple present - return empty string to use fallback
            return '';
        }
    }
    
    // Join parts with space
    let constructionName = parts.join(' ');
    
    // Handle special cases and typos in CSV
    // Check for exact match first
    if (conjugationTranslations[constructionName]) {
        return constructionName;
    }
    
    // Try variations
    // Handle "Past Negative" -> "Past-Negative"
    if (constructionName === 'Past Negative' && !hasPotential && !hasPassive && !hasCausative && !hasContinuous && !hasCompletion && !hasResultant && !hasConditional && !hasDesiderative && !hasDeontic) {
        return 'Past-Negative';
    }
    
    // Try with "Completion Past Potential" -> "Completion Past Potential"
    if (constructionName.includes('Completion') && constructionName.includes('Past') && constructionName.includes('Potential')) {
        return 'Completion Past Potential';
    }
    
    // Fallback: try to find closest match
    // Remove spaces and try to match
    const normalizedName = constructionName.replace(/\s+/g, ' ').trim();
    for (const key in conjugationTranslations) {
        if (key.replace(/\s+/g, ' ').trim() === normalizedName) {
            return key;
        }
    }
    
    // If still no match, return the constructed name (might not be in CSV)
    return constructionName;
}

// Update mood description based on current mood
function updateMoodDescription() {
    const descriptionElement = document.getElementById('englishDescription');
    if (descriptionElement) {
        // Get CSV construction name from current button state
        const constructionName = getCSVConstructionName();
        
        // If empty construction name (plain present with no modifiers), use fallback
        if (!constructionName || constructionName === '') {
            descriptionElement.textContent = 'I verb';
            // Update cell data-form and return
            const formalities = ['casual', 'standard', 'polite', 'formal'];
            formalities.forEach(formality => {
                const cell = document.getElementById(formality + 'Cell');
                if (cell && currentMood) {
                    cell.setAttribute('data-form', currentMood);
                }
            });
            return;
        }
        
        // Get translation from CSV mapping
        let translation = conjugationTranslations[constructionName];
        
        // If no translation found, try to clean up the construction name and search again
        if (!translation) {
            // Try trimmed version
            translation = conjugationTranslations[constructionName.trim()];
        }
        
        // If still no translation, try case-insensitive and normalized search
        if (!translation) {
            const normalizedName = constructionName.toLowerCase().trim().replace(/\s+/g, ' ');
            for (const key in conjugationTranslations) {
                const normalizedKey = key.toLowerCase().trim().replace(/\s+/g, ' ');
                // Also try replacing "conditional" with "conitional" to handle CSV typo
                const normalizedNameAlt = normalizedName.replace(/conditional/g, 'conitional');
                const normalizedKeyAlt = normalizedKey.replace(/conditional/g, 'conitional');
                if (normalizedKey === normalizedName || normalizedKeyAlt === normalizedNameAlt) {
                    translation = conjugationTranslations[key];
                    break;
                }
            }
        }
        
        // Clean up the translation (remove extra quotes if present)
        if (translation) {
            if (translation.startsWith('"') && translation.endsWith('"')) {
                translation = translation.slice(1, -1);
            }
            
            // If translation contains multiple options (separated by ", "), show the first one
            if (translation.includes('", "')) {
                translation = translation.split('", "')[0].replace(/^"/, '').replace(/"$/, '');
            }
        }
        
        // Use translation or fallback
        descriptionElement.textContent = translation || 'I verb';
    }
    
    // Update cell data-form
    const formalities = ['casual', 'standard', 'polite', 'formal'];
    formalities.forEach(formality => {
        const cell = document.getElementById(formality + 'Cell');
        if (cell && currentMood) {
            cell.setAttribute('data-form', currentMood);
        }
    });
}

// Get current person for common mode labels
function getCurrentPerson() {
    if (controls.first) return 'I';
    if (controls.second) return 'You';
    if (controls.third) return 'They';
    return 'I'; // default
}

// Update button labels based on mode
function updateButtonLabels() {
    if (currentMode === 'Common') {
        const person = getCurrentPerson();
        const isNegative = controls.negative;
        const isPast = controls.past;
        
        // Person buttons
        const firstBtn = document.getElementById('firstPersonBtn');
        const secondBtn = document.getElementById('secondPersonBtn');
        const thirdBtn = document.getElementById('thirdPersonBtn');
        if (firstBtn) firstBtn.textContent = 'I';
        if (secondBtn) secondBtn.textContent = 'You';
        if (thirdBtn) thirdBtn.textContent = 'They';
        
        // Voice buttons
        const potentialBtn = document.getElementById('potentialBtn');
        const passiveBtn = document.getElementById('passiveBtn');
        const causativeBtn = document.getElementById('causativeBtn');
        if (potentialBtn) {
            if (isPast) {
                potentialBtn.textContent = isNegative ? `[${person}] could not` : `[${person}] could`;
            } else {
                potentialBtn.textContent = isNegative ? `[${person}] can not` : `[${person}] can`;
            }
        }
        if (passiveBtn) {
            if (isPast) {
                passiveBtn.textContent = isNegative ? 'it did not verb' : 'it verbed';
            } else {
                passiveBtn.textContent = isNegative ? 'it does not verb' : 'it verbs';
            }
        }
        if (causativeBtn) {
            if (isPast) {
                causativeBtn.textContent = isNegative ? `[${person}] didn't make someone` : `[${person}] made someone`;
            } else {
                causativeBtn.textContent = isNegative ? `[${person}] don't make someone` : `[${person}] make someone`;
            }
        }
        
        // Aspect buttons
        const continuousBtn = document.getElementById('continuousBtn');
        const completionBtn = document.getElementById('completionBtn');
        const resultantBtn = document.getElementById('resultantBtn');
        if (continuousBtn) {
            if (isPast) {
                continuousBtn.textContent = isNegative ? `[${person}] didn't continue -ing` : `[${person}] continued -ing`;
            } else {
                continuousBtn.textContent = isNegative ? `[${person}] don't continue -ing` : `[${person}] continue -ing`;
            }
        }
        if (completionBtn) {
            if (isPast) {
                completionBtn.textContent = isNegative ? `[${person}] didn't happen to` : `[${person}] happened to`;
            } else {
                completionBtn.textContent = isNegative ? `[${person}] happen to not` : `[${person}] happen to`;
            }
        }
        // Resultant uses appropriate verb form based on person, negative, and tense
        if (resultantBtn) {
            if (isPast) {
                if (isNegative) {
                    if (person === 'I') {
                        resultantBtn.textContent = '[I was] still not -ing';
                    } else if (person === 'You') {
                        resultantBtn.textContent = '[You were] still not -ing';
                    } else {
                        resultantBtn.textContent = '[They were] still not -ing';
                    }
                } else {
                    if (person === 'I') {
                        resultantBtn.textContent = '[I was] still -ing';
                    } else if (person === 'You') {
                        resultantBtn.textContent = '[You were] still -ing';
                    } else {
                        resultantBtn.textContent = '[They were] still -ing';
                    }
                }
            } else {
                if (isNegative) {
                    if (person === 'I') {
                        resultantBtn.textContent = '[I am] still not -ing';
                    } else if (person === 'You') {
                        resultantBtn.textContent = '[You are] still not -ing';
                    } else {
                        resultantBtn.textContent = '[They are] still not -ing';
                    }
                } else {
                    if (person === 'I') {
                        resultantBtn.textContent = '[I am] still -ing';
                    } else if (person === 'You') {
                        resultantBtn.textContent = '[You are] still -ing';
                    } else {
                        resultantBtn.textContent = '[They are] still -ing';
                    }
                }
            }
        }
        
        // Mood buttons
        const plainBtn = document.getElementById('plainMoodBtn');
        const teBtn = document.getElementById('teMoodBtn');
        const volitionalBtn = document.getElementById('volitionalMoodBtn');
        const conditionalBtn = document.getElementById('conditionalMoodBtn');
        const desiderativeBtn = document.getElementById('desiderativeMoodBtn');
        const deonticBtn = document.getElementById('deonticMoodBtn');
        const imperativeBtn = document.getElementById('imperativeMoodBtn');
        if (plainBtn) {
            if (isPast) {
                plainBtn.textContent = isNegative ? `[${person}] didn't verb` : `[${person}] verbed`;
            } else {
                plainBtn.textContent = isNegative ? `[${person}] don't verb` : `[${person}] verb`;
            }
        }
        if (teBtn) {
            if (isPast) {
                teBtn.textContent = isNegative ? `[${person}] didn't verb and...` : `[${person}] verbed and...`;
            } else {
                teBtn.textContent = isNegative ? `[${person}] don't verb and...` : `[${person}] verb and...`;
            }
        }
        if (volitionalBtn) {
            volitionalBtn.textContent = isNegative ? 'let\'s not verb' : 'let\'s verb';
        }
        if (conditionalBtn) {
            if (isPast) {
                conditionalBtn.textContent = isNegative ? `if [${person}] didn't verb` : `if [${person}] verbed`;
            } else {
                conditionalBtn.textContent = isNegative ? `if [${person}] don't verb` : `if [${person}] verb`;
            }
        }
        if (desiderativeBtn) {
            if (isPast) {
                desiderativeBtn.textContent = isNegative ? `[${person}] didn't want to verb` : `[${person}] wanted to verb`;
            } else {
                desiderativeBtn.textContent = isNegative ? `[${person}] don't want to verb` : `[${person}] want to verb`;
            }
        }
        if (deonticBtn) {
            deonticBtn.textContent = isNegative ? `[${person}] must not verb` : `[${person}] must verb`;
        }
        if (imperativeBtn) {
            imperativeBtn.textContent = isNegative ? 'don\'t verb!' : 'verb!';
        }
    } else {
        // Linguistic mode - use original labels
        const firstBtn = document.getElementById('firstPersonBtn');
        const secondBtn = document.getElementById('secondPersonBtn');
        const thirdBtn = document.getElementById('thirdPersonBtn');
        if (firstBtn) firstBtn.textContent = 'First';
        if (secondBtn) secondBtn.textContent = 'Second';
        if (thirdBtn) thirdBtn.textContent = 'Third';
        
        const potentialBtn = document.getElementById('potentialBtn');
        const passiveBtn = document.getElementById('passiveBtn');
        const causativeBtn = document.getElementById('causativeBtn');
        if (potentialBtn) potentialBtn.textContent = 'Potential';
        if (passiveBtn) passiveBtn.textContent = 'Passive';
        if (causativeBtn) causativeBtn.textContent = 'Causative';
        
        const continuousBtn = document.getElementById('continuousBtn');
        const completionBtn = document.getElementById('completionBtn');
        const resultantBtn = document.getElementById('resultantBtn');
        if (continuousBtn) continuousBtn.textContent = 'Continuous';
        if (completionBtn) completionBtn.textContent = 'Completion';
        if (resultantBtn) resultantBtn.textContent = 'Resultant';
        
        const plainBtn = document.getElementById('plainMoodBtn');
        const teBtn = document.getElementById('teMoodBtn');
        const volitionalBtn = document.getElementById('volitionalMoodBtn');
        const conditionalBtn = document.getElementById('conditionalMoodBtn');
        const desiderativeBtn = document.getElementById('desiderativeMoodBtn');
        const deonticBtn = document.getElementById('deonticMoodBtn');
        const imperativeBtn = document.getElementById('imperativeMoodBtn');
        if (plainBtn) plainBtn.textContent = 'Plain';
        if (teBtn) teBtn.textContent = 'T Form';
        if (volitionalBtn) volitionalBtn.textContent = 'Volitional';
        if (conditionalBtn) conditionalBtn.textContent = 'Conditional';
        if (desiderativeBtn) desiderativeBtn.textContent = 'Desiderative';
        if (deonticBtn) deonticBtn.textContent = 'Deontic';
        if (imperativeBtn) imperativeBtn.textContent = 'Imperative';
    }
}

// Toggle between common and linguistic mode
function toggleMode() {
    currentMode = currentMode === 'Linguistic' ? 'Common' : 'Linguistic';
    const modeBtn = document.getElementById('modeToggleBtn');
    if (modeBtn) {
        modeBtn.textContent = currentMode;
    }
    updateButtonLabels();
}

// Validate and enforce mutually exclusive combinations
// Invalid combinations (order matters):
// 1. Potential + Continuous
// 2. Completion + Passive + Continuous
// 3. Completion + Causative-Passive + Continuous
// 4. Completion + Passive
// 5. Completion + Causative
// 6. Completion + Potential + Continuous + Deontic
function validateControlCombinations() {
    // Helper to disable a control and update its button
    function disableControl(controlName) {
        controls[controlName] = false;
        const btn = document.getElementById(controlName + 'Btn') || 
                    document.getElementById(controlName + 'PersonBtn') ||
                    document.getElementById(controlName + 'MoodBtn');
        if (btn) {
            btn.classList.remove('active');
        }
    }
    
    // Check rules in order of priority
    
    // 4. Completion + Passive (not allowed) - check this first
    // If both are on, disable passive (completion takes priority)
    if (controls.completion && controls.passive) {
        disableControl('passive');
    }
    
    // 5. Completion + Causative (not allowed)
    // If both are on, disable causative (completion takes priority)
    if (controls.completion && controls.causative) {
        disableControl('causative');
    }
    
    // 1. Potential + Continuous (not allowed)
    // Note: This is handled at the toggle level with swap behavior, but kept as a safety net
    // If both are on, disable continuous (potential takes priority)
    if (controls.potential && controls.continuous) {
        disableControl('continuous');
    }
    
    // 2. Completion + Passive + Continuous (not allowed)
    // If all three are on, disable continuous
    // Note: This handles the case even if passive was just disabled above
    if (controls.completion && controls.passive && controls.continuous) {
        disableControl('continuous');
    }
    
    // 3. Completion + Causative-Passive + Continuous (not allowed)
    // Causative-Passive means both causative and passive are active
    // If completion + causative + passive + continuous are all on, disable continuous
    if (controls.completion && controls.causative && controls.passive && controls.continuous) {
        disableControl('continuous');
    }
    
    // 6. Completion + Potential + Continuous + Deontic (not allowed)
    // If all four are on, disable deontic
    // Note: This can only happen if rule #1 didn't catch it, so we check it
    if (controls.completion && controls.potential && controls.continuous && controls.deontic) {
        disableControl('deontic');
    }
}

// Toggle control button
function toggleControl(controlName) {
    // Handle mutual exclusivity groups
    if (controlName === 'simple' || controlName === 'past') {
        // Tense: simple || past
        // Ensure at least one is always on (self-check: clicking the only active one keeps it on)
        const activeTenseCount = (controls.simple ? 1 : 0) + (controls.past ? 1 : 0);
        const isCurrentlyActive = controls[controlName];
        
        // If this is the only active button and it's being clicked, keep it on (don't toggle)
        if (isCurrentlyActive && activeTenseCount === 1) {
            // Don't toggle - keep it on
            return;
        }
        
        // Otherwise, normal toggle behavior
        if (controlName === 'simple' && !controls.simple) {
            controls.past = false;
            const pastBtn = document.getElementById('pastBtn');
            if (pastBtn) pastBtn.classList.remove('active');
        } else if (controlName === 'past' && !controls.past) {
            controls.simple = false;
            const simpleBtn = document.getElementById('simpleBtn');
            if (simpleBtn) simpleBtn.classList.remove('active');
        }
    } else if (controlName === 'first' || controlName === 'second' || controlName === 'third') {
        // Person: first || second || third
        // Ensure at least one is always on (self-check: clicking the only active one keeps it on)
        const activePersonCount = (controls.first ? 1 : 0) + (controls.second ? 1 : 0) + (controls.third ? 1 : 0);
        const isCurrentlyActive = controls[controlName];
        
        // If this is the only active button and it's being clicked, keep it on (don't toggle)
        if (isCurrentlyActive && activePersonCount === 1) {
            // Don't toggle - keep it on
            return;
        }
        
        // Otherwise, normal toggle behavior
        if (controlName === 'first' && !controls.first) {
            controls.second = false;
            controls.third = false;
            const secondBtn = document.getElementById('secondPersonBtn');
            const thirdBtn = document.getElementById('thirdPersonBtn');
            if (secondBtn) secondBtn.classList.remove('active');
            if (thirdBtn) thirdBtn.classList.remove('active');
        } else if (controlName === 'second' && !controls.second) {
            controls.first = false;
            controls.third = false;
            const firstBtn = document.getElementById('firstPersonBtn');
            const thirdBtn = document.getElementById('thirdPersonBtn');
            if (firstBtn) firstBtn.classList.remove('active');
            if (thirdBtn) thirdBtn.classList.remove('active');
        } else if (controlName === 'third' && !controls.third) {
            controls.first = false;
            controls.second = false;
            const firstBtn = document.getElementById('firstPersonBtn');
            const secondBtn = document.getElementById('secondPersonBtn');
            if (firstBtn) firstBtn.classList.remove('active');
            if (secondBtn) secondBtn.classList.remove('active');
        }
    } else if (controlName === 'potential') {
        // potential || passive, potential || causative, potential || continuous
        // Swap behavior: if turning potential on and continuous is active, swap them
        if (!controls.potential) {
            // Turn off mutually exclusive voice buttons
            controls.passive = false;
            controls.causative = false;
            const passiveBtn = document.getElementById('passiveBtn');
            const causativeBtn = document.getElementById('causativeBtn');
            if (passiveBtn) passiveBtn.classList.remove('active');
            if (causativeBtn) causativeBtn.classList.remove('active');
            
            // Swap with continuous if it's active
            if (controls.continuous) {
                controls.continuous = false;
                const continuousBtn = document.getElementById('continuousBtn');
                if (continuousBtn) continuousBtn.classList.remove('active');
            }
        }
    } else if (controlName === 'passive') {
        // potential || passive
        // Swap behavior: if turning passive on and potential is active, swap them
        if (!controls.passive) {
            controls.potential = false;
            const potentialBtn = document.getElementById('potentialBtn');
            if (potentialBtn) potentialBtn.classList.remove('active');
        }
    } else if (controlName === 'causative') {
        // potential || causative
        // Swap behavior: if turning causative on and potential is active, swap them
        if (!controls.causative) {
            controls.potential = false;
            const potentialBtn = document.getElementById('potentialBtn');
            if (potentialBtn) potentialBtn.classList.remove('active');
        }
    } else if (controlName === 'continuous') {
        // continuous || resultant, continuous || potential
        // Swap behavior: if turning continuous on and potential is active, swap them
        if (!controls.continuous) {
            // Turn off mutually exclusive aspect button
            controls.resultant = false;
            const resultantBtn = document.getElementById('resultantBtn');
            if (resultantBtn) resultantBtn.classList.remove('active');
            
            // Swap with potential if it's active
            if (controls.potential) {
                controls.potential = false;
                const potentialBtn = document.getElementById('potentialBtn');
                if (potentialBtn) potentialBtn.classList.remove('active');
            }
        }
    } else if (controlName === 'resultant') {
        // Aspect: continuous || resultant (mutually exclusive, but both can combine with completion)
        // Swap behavior: if turning resultant on and continuous is active, swap them
        if (!controls.resultant) {
            controls.continuous = false;
            const continuousBtn = document.getElementById('continuousBtn');
            if (continuousBtn) continuousBtn.classList.remove('active');
        }
    }
    
    // If T form is active and user is changing any other control (not the te mood button itself),
    // reset mood to plain form
    if (controls.te && controlName !== 'te') {
        // Reset to plain form
        controls.te = false;
        controls.plain = true;
        currentMood = 'plain';
        
        // Update mood button states
        const teBtn = document.getElementById('teMoodBtn');
        const plainBtn = document.getElementById('plainMoodBtn');
        if (teBtn) teBtn.classList.remove('active');
        if (plainBtn) plainBtn.classList.add('active');
        
        // Update mood description
        updateMoodDescription();
    }
    
    // Toggle the control
    controls[controlName] = !controls[controlName];
    const btn = document.getElementById(controlName + 'Btn') || 
                document.getElementById(controlName + 'PersonBtn') ||
                document.getElementById(controlName + 'MoodBtn');
    if (btn) {
        if (controls[controlName]) {
            btn.classList.add('active');
        } else {
            btn.classList.remove('active');
        }
    }
    
    // Validate and enforce mutually exclusive combinations
    validateControlCombinations();
    
    // Update form labels
    updateFormLabels();
    
    // Update English description from CSV
    updateMoodDescription();
    
    // Update button labels if person, negative, or tense changed and in common mode
    if (((controlName === 'first' || controlName === 'second' || controlName === 'third' || controlName === 'negative' || controlName === 'simple' || controlName === 'past') && currentMode === 'Common')) {
        updateButtonLabels();
    }
    
    // Re-conjugate the verb if we have one
    if (currentVerb) {
        if (conjugateVerbGlobal) {
            conjugateVerbGlobal(currentVerb);
        } else {
            updateGridCells();
        }
    }
}

// Get conjugation endings for a specific form
// Note: endings and irregularEndings are now loaded from conjugationEndings.js
function getConjugationEndings(form, formality) {
    // Get data from conjugationEndings.js
    const endingsData = getConjugationEndingsData();
    const endings = endingsData.endings;
    const irregularEndings = endingsData.irregularEndings;
    
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


// Apply conjugation pattern to a verb
// Japanese construction order: stem → voice → aspect → tense → mood
// The CSV patterns encode the full transformation, so we apply them directly
function applyConjugationPattern(verb, pattern, verbEndingIndex) {
    if (!pattern) {
        // If pattern not found, try to build from components
        console.warn('Pattern not found, using fallback');
        return verb;
    }
    
    let stem = '';
    const lastChar = verb[verb.length - 1];
    
    // Get base form based on pattern.base
    // This extracts the stem and applies the appropriate base transformation
    if (pattern.base === 'Dictionary') {
        // Use verb as-is (for deontic, simple present)
        stem = verb;
    } else if (pattern.base === 'Continuative / Stem') {
        // Get stem form
        if (verbEndingIndex === 9) {
            // Ichidan: remove る
            stem = verb.slice(0, -1);
        } else {
            // Godan: change last character to stem form
            const stemEndings = ['い', 'ち', 'り', 'き', 'ぎ', 'み', 'に', 'び', 'し', ''];
            stem = verb.slice(0, -1) + stemEndings[verbEndingIndex];
        }
    } else if (pattern.base === 'Irrealis / Imperfective') {
        // Get negative stem (Irrealis form)
        const rootEnding = pattern.root[verbEndingIndex];
        if (verbEndingIndex === 9) {
            // Ichidan: remove る, then add root if not ''
            stem = verb.slice(0, -1);
            if (rootEnding !== '') {
                stem = stem + rootEnding;
            }
        } else {
            // Godan: use root from pattern
            if (rootEnding === '') {
                stem = verb.slice(0, -1); // ichidan
            } else {
                stem = verb.slice(0, -1) + rootEnding;
            }
        }
    } else if (pattern.base === 'Hypothetical') {
        // Get hypothetical form
        if (verbEndingIndex === 9) {
            // Ichidan: remove る, add hypothetical
            stem = verb.slice(0, -1) + 'れ';
        } else {
            // Godan: use root from pattern
            const rootEnding = pattern.root[verbEndingIndex];
            stem = verb.slice(0, -1) + rootEnding;
        }
    } else if (pattern.base === 'T Form') {
        // Get te-form
        if (verbEndingIndex === 9) {
            // Ichidan: remove る, add て
            stem = verb.slice(0, -1) + 'て';
        } else {
            // Godan: use root from pattern
            const rootEnding = pattern.root[verbEndingIndex];
            stem = verb.slice(0, -1) + rootEnding;
        }
    } else if (pattern.base === 'Modified T') {
        // Past form (ta-form)
        if (verbEndingIndex === 9) {
            // Ichidan: remove る, add た
            stem = verb.slice(0, -1) + 'た';
        } else {
            // Godan: use root from pattern
            const rootEnding = pattern.root[verbEndingIndex];
            stem = verb.slice(0, -1) + rootEnding;
        }
    } else if (pattern.base === 'Imperative') {
        // Imperative form
        if (verbEndingIndex === 9) {
            // Ichidan: remove る, add ろ
            stem = verb.slice(0, -1) + 'ろ';
        } else {
            // Godan: use root from pattern
            const rootEnding = pattern.root[verbEndingIndex];
            stem = verb.slice(0, -1) + rootEnding;
        }
    }
    
    // Add conjugation ending
    return stem + pattern.conjugation;
}

// Build full construction name including base mood when appropriate
function getFullConstructionName() {
    const parts = [];
    
    // 1. TENSE (past or simple present - simple present is default/implied)
    if (controls.past) parts.push('past');
    
    // 2. NEGATION
    if (controls.negative) parts.push('negative');
    
    // 3. VOICE (potential, passive, causative)
    if (controls.potential && controls.passive) {
        parts.push('potential');
        parts.push('passive');
    } else if (controls.potential && controls.causative) {
        parts.push('potential');
        parts.push('causative');
    } else if (controls.causative && controls.passive) {
        parts.push('causative');
        parts.push('passive');
    } else if (controls.potential) {
        parts.push('potential');
    } else if (controls.passive) {
        parts.push('passive');
    } else if (controls.causative) {
        parts.push('causative');
    }
    
    // 4. ASPECT (continuous, completion, resultant)
    if (controls.completion && controls.continuous) {
        parts.push('completion');
        parts.push('continuous');
    } else if (controls.completion && controls.resultant) {
        parts.push('completion');
        parts.push('resultant');
    } else if (controls.continuous && controls.resultant) {
        parts.push('continuous');
        parts.push('resultant');
    } else if (controls.continuous) {
        parts.push('continuous');
    } else if (controls.completion) {
        parts.push('completion');
    } else if (controls.resultant) {
        parts.push('resultant');
    }
    
    // 5. MOOD modifiers (conditional, desiderative, deontic)
    if (controls.conditional && controls.desiderative) {
        parts.push('conditional');
        parts.push('desiderative');
    } else if (controls.conditional && controls.deontic) {
        parts.push('conditional');
        parts.push('deontic');
    } else if (controls.desiderative) {
        parts.push('desiderative');
    } else if (controls.deontic) {
        parts.push('deontic');
    } else if (controls.conditional) {
        parts.push('conditional');
    }
    
    // Handle base moods (volitional, imperative) - these are standalone patterns
    // Note: te-form is handled specially as it has no pattern when standalone
    if (controls.volitional && parts.length === 0) {
        return 'volitional';
    }
    if (controls.imperative && parts.length === 0) {
        return 'imperative';
    }
    
    // If no parts, return basic form
    if (parts.length === 0) return 'simple present';
    
    // Join parts
    return parts.join(' ');
}

// Get conjugated form for a specific form and formality
function getConjugatedForm(form, formality) {
    if (!currentVerb) {
        return '';
    }

    const verbEndingIndex = getVerbEndingIndex(currentVerb);
    
    // Handle standalone te-form (no modifiers) - build te-form directly
    if (controls.te && !controls.continuous && !controls.completion && !controls.resultant && 
        !controls.negative && !controls.past && !controls.potential && !controls.passive && 
        !controls.causative && !controls.conditional && !controls.desiderative && !controls.deontic) {
        // Build te-form directly using T Form roots
        if (verbEndingIndex === 9) {
            // Ichidan: remove る, add て
            return currentVerb.slice(0, -1) + 'て';
        } else {
            // Godan: use T Form roots
            const teRoots = ['って', 'って', 'って', 'いて', 'いで', 'んで', 'んで', 'んで', 'して', 'て'];
            return currentVerb.slice(0, -1) + teRoots[verbEndingIndex];
        }
    }
    
    // Build construction name from active buttons
    const constructionName = getFullConstructionName();
    
    // Select pattern set based on formality
    const patternSet = (formality === 'polite' || formality === 'formal') 
        ? conjugationPatternsPolite 
        : conjugationPatterns;
    
    // Look up pattern
    const pattern = patternSet[constructionName];
    
    if (pattern) {
        // Apply pattern: verb stem + root[verbEndingIndex] + conjugation
        return applyConjugationPattern(currentVerb, pattern, verbEndingIndex);
    }
    
    // Fallback: return empty string if pattern not found
    console.warn(`Pattern not found for construction: ${constructionName}, formality: ${formality}`);
    return '';
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

    // Initialize default controls
    // Set defaults: first person, simple/present, plain mood
    controls.first = true;
    controls.third = false;
    controls.simple = true;
    controls.plain = true;
    currentMood = 'plain';
    
    // Update button states to reflect defaults
    const firstBtn = document.getElementById('firstPersonBtn');
    const simpleBtn = document.getElementById('simpleBtn');
    const plainBtn = document.getElementById('plainMoodBtn');
    if (firstBtn) firstBtn.classList.add('active');
    if (simpleBtn) simpleBtn.classList.add('active');
    if (plainBtn) plainBtn.classList.add('active');
    
    // Initialize button labels
    updateButtonLabels();
    
    // Initialize English description from CSV
    updateMoodDescription();
    
    // Load default verb on page load
    if (verbInput.value) {
        // Initialize mood selection
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
    window.toggleMode = toggleMode;
});

