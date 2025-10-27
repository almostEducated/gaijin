// SR Timer - tracks how long user takes to answer
let srStartTime = null;

document.addEventListener('DOMContentLoaded', () => {
    srStartTime = Date.now();
});

function getSRElapsedTime() {
    if (!srStartTime) {
        return 0;
    }
    return Math.floor(Date.now() - srStartTime); // Return milliseconds as integer
}

// Update hidden time input before form submission
function updateTimeBeforeSubmit(form) {
    const timeInput = form.querySelector('input[name="time"]');
    if (timeInput) {
        timeInput.value = getSRElapsedTime(); // Send as integer milliseconds
    }
    return true; // Allow form to submit
}