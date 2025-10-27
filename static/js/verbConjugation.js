// Verb Conjugation Chart JavaScript

// Global state for toggles
let isNegative = false;
let isPolite = false;

// Toggle functions
function toggleNegative() {
    isNegative = !isNegative;
    const btn = document.getElementById('negativeToggle');
    if (isNegative) {
        btn.style.background = '#dc3545';
        btn.style.color = 'white';
        btn.style.borderColor = '#dc3545';
    } else {
        btn.style.background = '#f8f9fa';
        btn.style.color = '#333';
        btn.style.borderColor = '#ddd';
    }
    // Re-conjugate if we have a verb
    const verbForm = document.getElementById('verbForm');
    const verbInput = document.getElementById('verbInput');
    if (verbInput.value.trim()) {
        verbForm.dispatchEvent(new Event('submit'));
    }
}

function togglePolite() {
    isPolite = !isPolite;
    const btn = document.getElementById('politeToggle');
    if (isPolite) {
        btn.style.background = '#28a745';
        btn.style.color = 'white';
        btn.style.borderColor = '#28a745';
    } else {
        btn.style.background = '#f8f9fa';
        btn.style.color = '#333';
        btn.style.borderColor = '#ddd';
    }
    // Re-conjugate if we have a verb
    const verbForm = document.getElementById('verbForm');
    const verbInput = document.getElementById('verbInput');
    if (verbInput.value.trim()) {
        verbForm.dispatchEvent(new Event('submit'));
    }
}

document.addEventListener('DOMContentLoaded', function() {
    const verbForm = document.getElementById('verbForm');
    const verbInput = document.getElementById('verbInput');
    const verbError = document.getElementById('verbError');
    const verbType = document.getElementById('verbType');
    const conjugationChart = document.getElementById('conjugationChart');

    verbForm.addEventListener('submit', async function(e) {
        e.preventDefault();
        
        const verb = verbInput.value.trim();
        if (!verb) {
            showError('Please enter a verb');
            return;
        }

        // Hide previous results
        hideError();
        hideChart();
        hideVerbType();

        try {
            const response = await fetch('/api/verb/conjugate', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ 
                    verb: verb,
                    negative: isNegative,
                    polite: isPolite
                })
            });

            const data = await response.json();

            if (!data.valid) {
                showError(data.error || 'Invalid verb');
                return;
            }

            // Show verb type with modifiers
            showVerbType(data.verb, data.verbType, isNegative, isPolite);

            // Populate conjugation chart
            populateChart(data.conjugations);
            
            // Show the chart
            showChart();

        } catch (error) {
            showError('Failed to conjugate verb. Please try again.');
            console.error('Error:', error);
        }
    });

    function showError(message) {
        verbError.textContent = message;
        verbError.style.display = 'block';
    }

    function hideError() {
        verbError.style.display = 'none';
    }

    function showVerbType(verb, type, negative, polite) {
        let modifiers = [];
        if (negative) modifiers.push('negative');
        if (polite) modifiers.push('polite');
        
        let text = `${verb} is a ${type} verb`;
        if (modifiers.length > 0) {
            text += ` (${modifiers.join(' + ')})`;
        }
        
        verbType.textContent = text;
        verbType.style.display = 'block';
    }

    function hideVerbType() {
        verbType.style.display = 'none';
    }

    function showChart() {
        conjugationChart.style.display = 'block';
    }

    function hideChart() {
        conjugationChart.style.display = 'none';
    }

    function populateChart(conjugations) {
        if (conjugations.tenses) {
            // Populate Time
            if (conjugations.tenses.time) {
                populateTable('timeConjugations', conjugations.tenses.time);
            }

            // Populate Aspect
            if (conjugations.tenses.aspect) {
                populateTable('aspectConjugations', conjugations.tenses.aspect);
            }

            // Populate Mood
            if (conjugations.tenses.mood) {
                populateTable('moodConjugations', conjugations.tenses.mood);
            }

            // Populate Modals
            if (conjugations.tenses.modals) {
                populateTable('modalsConjugations', conjugations.tenses.modals);
            }

            // Populate Desire
            if (conjugations.tenses.desire) {
                populateTable('desireConjugations', conjugations.tenses.desire);
            }
        }

        // Populate Voice
        if (conjugations.voice) {
            populateTable('voiceConjugations', conjugations.voice);
        }
    }

    function populateTable(tableId, conjugationData) {
        const tbody = document.getElementById(tableId);
        tbody.innerHTML = ''; // Clear existing content

        for (const [form, data] of Object.entries(conjugationData)) {
            const row = document.createElement('tr');
            
            // Form name
            const formCell = document.createElement('td');
            formCell.style.padding = '10px';
            formCell.style.border = '1px solid #ddd';
            formCell.textContent = capitalizeFirstLetter(form.replace(/_/g, ' '));
            
            // English
            const englishCell = document.createElement('td');
            englishCell.style.padding = '10px';
            englishCell.style.border = '1px solid #ddd';
            englishCell.textContent = data.english;
            
            // Japanese (with alts if available)
            const japaneseCell = document.createElement('td');
            japaneseCell.style.padding = '10px';
            japaneseCell.style.border = '1px solid #ddd';
            japaneseCell.style.fontSize = '20px';
            japaneseCell.style.fontWeight = 'bold';
            
            let japaneseText = data.japanese;
            if (data.alts && data.alts.length > 0) {
                japaneseText += ` (${data.alts.join(', ')})`;
            }
            japaneseCell.textContent = japaneseText;
            
            row.appendChild(formCell);
            row.appendChild(englishCell);
            row.appendChild(japaneseCell);
            
            tbody.appendChild(row);
        }
    }

    function capitalizeFirstLetter(string) {
        return string.charAt(0).toUpperCase() + string.slice(1);
    }

    // Load default verb on page load (optional)
    // Uncomment the lines below to load a default verb like 食べる
    
    verbInput.value = '食べる';
    verbForm.dispatchEvent(new Event('submit'));
    
});

