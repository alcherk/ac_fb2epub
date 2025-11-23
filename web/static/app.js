const API_BASE = '/api/v1';
let currentJobId = null;
let statusCheckInterval = null;

// DOM elements
const uploadSection = document.getElementById('uploadSection');
const processingSection = document.getElementById('processingSection');
const resultSection = document.getElementById('resultSection');
const errorSection = document.getElementById('errorSection');
const fileInput = document.getElementById('fileInput');
const uploadArea = document.getElementById('uploadArea');
const statusMessage = document.getElementById('statusMessage');
const progressFill = document.getElementById('progressFill');
const errorMessage = document.getElementById('errorMessage');
const downloadBtn = document.getElementById('downloadBtn');
const convertAnotherBtn = document.getElementById('convertAnotherBtn');
const retryBtn = document.getElementById('retryBtn');

// File input change
fileInput.addEventListener('change', (e) => {
    if (e.target.files.length > 0) {
        handleFile(e.target.files[0]);
    }
});

// Drag and drop
uploadArea.addEventListener('dragover', (e) => {
    e.preventDefault();
    uploadArea.classList.add('dragover');
});

uploadArea.addEventListener('dragleave', () => {
    uploadArea.classList.remove('dragover');
});

uploadArea.addEventListener('drop', (e) => {
    e.preventDefault();
    uploadArea.classList.remove('dragover');
    
    if (e.dataTransfer.files.length > 0) {
        const file = e.dataTransfer.files[0];
        if (file.name.endsWith('.fb2') || file.name.endsWith('.xml')) {
            handleFile(file);
        } else {
            showError('Please select a .fb2 or .xml file');
        }
    }
});

// Click to upload
uploadArea.addEventListener('click', () => {
    fileInput.click();
});

// Handle file upload and conversion
async function handleFile(file) {
    // Validate file
    if (!file.name.endsWith('.fb2') && !file.name.endsWith('.xml')) {
        showError('Invalid file type. Please select a .fb2 or .xml file.');
        return;
    }

    // Show processing section
    showProcessing();

    // Create form data
    const formData = new FormData();
    formData.append('file', file);

    try {
        // Upload and start conversion
        const response = await fetch(`${API_BASE}/convert`, {
            method: 'POST',
            body: formData
        });

        if (!response.ok) {
            const error = await response.json();
            throw new Error(error.error || 'Failed to upload file');
        }

        const data = await response.json();
        currentJobId = data.job_id;

        // Start polling for status
        startStatusPolling();
    } catch (error) {
        showError(error.message);
    }
}

// Poll for conversion status
function startStatusPolling() {
    let attempts = 0;
    const maxAttempts = 60; // 60 seconds max

    statusCheckInterval = setInterval(async () => {
        attempts++;
        updateProgress(Math.min((attempts / maxAttempts) * 90, 90));

        try {
            const response = await fetch(`${API_BASE}/status/${currentJobId}`);
            const data = await response.json();

            if (data.status === 'completed') {
                clearInterval(statusCheckInterval);
                updateProgress(100);
                setTimeout(() => showResult(), 500);
            } else if (data.status === 'failed') {
                clearInterval(statusCheckInterval);
                showError(data.error || 'Conversion failed');
            } else if (attempts >= maxAttempts) {
                clearInterval(statusCheckInterval);
                showError('Conversion timeout. Please try again.');
            }
        } catch (error) {
            clearInterval(statusCheckInterval);
            showError('Failed to check conversion status');
        }
    }, 1000);
}

// Update progress bar
function updateProgress(percent) {
    progressFill.style.width = `${percent}%`;
}

// Show processing section
function showProcessing() {
    uploadSection.style.display = 'none';
    processingSection.style.display = 'block';
    resultSection.style.display = 'none';
    errorSection.style.display = 'none';
    updateProgress(10);
}

// Show result section
function showResult() {
    uploadSection.style.display = 'none';
    processingSection.style.display = 'none';
    resultSection.style.display = 'block';
    errorSection.style.display = 'none';
}

// Show error section
function showError(message) {
    if (statusCheckInterval) {
        clearInterval(statusCheckInterval);
    }
    uploadSection.style.display = 'none';
    processingSection.style.display = 'none';
    resultSection.style.display = 'none';
    errorSection.style.display = 'block';
    errorMessage.textContent = message;
}

// Download EPUB
downloadBtn.addEventListener('click', () => {
    if (currentJobId) {
        window.location.href = `${API_BASE}/download/${currentJobId}`;
    }
});

// Convert another file
convertAnotherBtn.addEventListener('click', () => {
    resetUI();
});

// Retry
retryBtn.addEventListener('click', () => {
    resetUI();
});

// Reset UI
function resetUI() {
    currentJobId = null;
    if (statusCheckInterval) {
        clearInterval(statusCheckInterval);
    }
    fileInput.value = '';
    uploadSection.style.display = 'block';
    processingSection.style.display = 'none';
    resultSection.style.display = 'none';
    errorSection.style.display = 'none';
    updateProgress(0);
}

