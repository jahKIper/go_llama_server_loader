/**
 * Llama Server Loader - Web UI JavaScript
 * Handles JSON-RPC 2.0 communication with the backend server.
 */

(function () {
    'use strict';

    // DOM Elements
    const scanBtn = document.getElementById('scan-btn');
    const modelsList = document.getElementById('models-list');
    const detailsSection = document.getElementById('details-section');
    const statusSection = document.getElementById('status-section');
    const actionsSection = document.getElementById('actions-section');
    const startBtn = document.getElementById('start-btn');
    const stopBtn = document.getElementById('stop-btn');
    const serverStatus = document.getElementById('server-status');
    const serverLog = document.getElementById('server-log');

    // State
    let models = [];
    let selectedModel = null;
    let statusInterval = null;

    /**
     * Send a JSON-RPC 2.0 request to the server.
     * @param {string} method - RPC method name
     * @param {*} params - Method parameters
     * @returns {Promise<*>} - Response result
     */
    async function rpcCall(method, params) {
        const id = Math.floor(Math.random() * 10000);
        const body = JSON.stringify({ method, params, id });

        try {
            const response = await fetch('/rpc', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: body
            });

            if (!response.ok) {
                throw new Error(`HTTP ${response.status}: ${response.statusText}`);
            }

            const data = await response.json();

            if (data.error) {
                throw new Error(`RPC Error: ${data.error.message}`);
            }

            return data.result;
        } catch (error) {
            logMessage(`RPC call failed: ${error.message}`);
            throw error;
        }
    }

    /**
     * Format file size to human-readable string.
     * @param {number} bytes - File size in bytes
     * @returns {string} Formatted size string
     */
    function formatSize(bytes) {
        if (bytes === 0) return '0 B';
        const units = ['B', 'KB', 'MB', 'GB', 'TB'];
        const k = 1024;
        const i = Math.floor(Math.log(bytes) / Math.log(k));
        return `${(bytes / Math.pow(k, i)).toFixed(2)} ${units[i]}`;
    }

    /**
     * Log a message to the server log area.
     * @param {string} msg - Message to log
     */
    function logMessage(msg) {
        const timestamp = new Date().toISOString();
        serverLog.textContent += `[${timestamp}] ${msg}\n`;
        serverLog.scrollTop = serverLog.scrollHeight;
    }

    /**
     * Update the status indicator display.
     * @param {string} status - Current status string
     */
    function updateStatus(status) {
        const parts = status.split(' ');
        const stateClass = parts[0];
        serverStatus.className = `status-indicator ${stateClass}`;

        let addrInfo = '';
        if (status.includes('port')) {
            const match = status.match(/port (\d+)/);
            if (match) addrInfo = ` on port ${match[1]}`;
        }
        serverStatus.textContent = `${status.charAt(0).toUpperCase() + status.slice(1)}${addrInfo}`;
    }

    /**
     * Render the list of models in the UI.
     */
    function renderModels() {
        if (models.length === 0) {
            modelsList.innerHTML = '<p class="placeholder">No models found</p>';
            return;
        }

        modelsList.innerHTML = '';

        models.forEach((model, index) => {
            const item = document.createElement('div');
            item.className = 'model-item';
            item.dataset.index = index;

            const nameSpan = document.createElement('span');
            nameSpan.className = 'model-name';
            nameSpan.textContent = model.name;

            const sizeSpan = document.createElement('span');
            sizeSpan.className = 'model-size';
            sizeSpan.textContent = formatSize(model.size);

            item.appendChild(nameSpan);
            item.appendChild(sizeSpan);

            item.addEventListener('click', () => selectModel(index));
            modelsList.appendChild(item);
        });
    }

    /**
     * Handle model selection.
     * @param {number} index - Model index
     */
    function selectModel(index) {
        // Remove previous selection
        const prev = modelsList.querySelector('.model-item.selected');
        if (prev) prev.classList.remove('selected');

        // Add selection to clicked item
        const items = modelsList.querySelectorAll('.model-item');
        items[index].classList.add('selected');

        selectedModel = models[index];
        detailsSection.classList.remove('hidden');
        actionsSection.classList.remove('hidden');
    }

    /**
     * Start polling server status.
     */
    function startStatusPolling() {
        if (statusInterval) clearInterval(statusInterval);

        // Initial fetch
        getStatus();

        statusInterval = setInterval(getStatus, 2000);
    }

    /**
     * Stop status polling.
     */
    function stopStatusPolling() {
        if (statusInterval) {
            clearInterval(statusInterval);
            statusInterval = null;
        }
    }

    /**
     * Fetch current server status via RPC.
     */
    async function getStatus() {
        try {
            const result = await rpcCall('getStatus', null);
            updateStatus(result.status);

            if (result.status === 'running' || result.status === 'shutting_down') {
                statusSection.classList.remove('hidden');
            }
        } catch (error) {
            logMessage(`Failed to get status: ${error.message}`);
        }
    }

    /**
     * Load models list via RPC.
     */
    async function loadModels() {
        try {
            const result = await rpcCall('getModels', null);
            models = result.models || [];
            renderModels();
            logMessage(`Loaded ${result.count} model(s)`);
        } catch (error) {
            logMessage(`Failed to load models: ${error.message}`);
        }
    }

    /**
     * Start the llama-server via RPC.
     */
    async function startServer() {
        if (!selectedModel) {
            logMessage('No model selected');
            return;
        }

        const threads = parseInt(document.getElementById('threads').value, 10);
        const temperature = parseFloat(document.getElementById('temperature').value);
        const port = parseInt(document.getElementById('port').value, 10);
        const mmprojOn = document.getElementById('mmproj-on').checked;

        const params = {
            model_path: selectedModel.path,
            threads: threads,
            temperature: temperature,
            port: port
        };

        if (selectedModel.mmproj_path && mmprojOn) {
            params.mmproj_path = selectedModel.mmproj_path;
        }

        try {
            const result = await rpcCall('startServer', params);
            logMessage(`Server response: ${result.message}`);
            statusSection.classList.remove('hidden');
            startBtn.classList.add('hidden');
            stopBtn.classList.remove('hidden');
            startStatusPolling();
        } catch (error) {
            logMessage(`Failed to start server: ${error.message}`);
        }
    }

    /**
     * Shutdown the server via RPC.
     */
    async function shutdownServer() {
        try {
            const result = await rpcCall('shutdown', null);
            logMessage(result.message);
            stopStatusPolling();
            startBtn.classList.remove('hidden');
            stopBtn.classList.add('hidden');
        } catch (error) {
            logMessage(`Shutdown failed: ${error.message}`);
        }
    }

    // Event Listeners
    scanBtn.addEventListener('click', loadModels);
    startBtn.addEventListener('click', startServer);
    stopBtn.addEventListener('click', shutdownServer);

    // Initial log
    logMessage('Web UI initialized');
})();
