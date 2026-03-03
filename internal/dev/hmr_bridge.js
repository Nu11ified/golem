(function() {
    'use strict';

    // HMR Bridge for Golem Dev Server
    var ws = new WebSocket('ws://' + window.location.host + '/ws');

    ws.onopen = function() {
        console.log('[Golem HMR] Connected to dev server');
    };

    ws.onmessage = function(event) {
        var data = event.data;

        if (data === 'reload') {
            console.log('[Golem HMR] Full reload');
            window.location.reload();
            return;
        }

        try {
            var msg = JSON.parse(data);

            if (msg.type === 'error') {
                showErrorOverlay(msg.message);
                return;
            }

            if (msg.type === 'module_update') {
                handleModuleUpdate(msg.module, msg.url);
                return;
            }
        } catch(e) {
            console.warn('[Golem HMR] Unknown message:', data);
        }
    };

    ws.onclose = function() {
        console.log('[Golem HMR] Connection lost, attempting reconnect...');
        setTimeout(function() { window.location.reload(); }, 2000);
    };

    // Module hot-swapping
    async function handleModuleUpdate(moduleName, moduleUrl) {
        console.log('[Golem HMR] Updating module:', moduleName);
        try {
            var response = await fetch(moduleUrl);
            if (!response.ok) {
                throw new Error('Failed to fetch module: ' + response.status);
            }
            var wasmBytes = await response.arrayBuffer();
            var go = new Go();
            var result = await WebAssembly.instantiate(wasmBytes, go.importObject);

            // Notify shell to swap page content if hook is registered
            if (window.__golem_hmr_swap) {
                window.__golem_hmr_swap(moduleName, result.instance);
            }

            go.run(result.instance);
            hideErrorOverlay();
            console.log('[Golem HMR] Module updated successfully:', moduleName);
        } catch(err) {
            console.error('[Golem HMR] Module update failed:', err);
            showErrorOverlay('Module update failed: ' + err.message);
        }
    }

    // Error overlay
    function showErrorOverlay(message) {
        hideErrorOverlay();

        var overlay = document.createElement('div');
        overlay.id = 'golem-error-overlay';
        overlay.style.cssText = [
            'position: fixed',
            'top: 0',
            'left: 0',
            'width: 100%',
            'height: 100%',
            'background: rgba(0, 0, 0, 0.85)',
            'color: #ff6b6b',
            'z-index: 99999',
            'display: flex',
            'flex-direction: column',
            'align-items: center',
            'justify-content: center',
            'font-family: monospace',
            'font-size: 14px',
            'padding: 20px',
            'box-sizing: border-box'
        ].join(';');

        var title = document.createElement('div');
        title.style.cssText = 'font-size: 24px; font-weight: bold; margin-bottom: 20px; color: #ff4444;';
        title.textContent = 'Golem Build Error';

        var content = document.createElement('pre');
        content.style.cssText = [
            'background: #1a1a2e',
            'color: #e0e0e0',
            'padding: 20px',
            'border-radius: 8px',
            'max-width: 80%',
            'max-height: 60%',
            'overflow: auto',
            'white-space: pre-wrap',
            'word-wrap: break-word',
            'border: 1px solid #ff4444'
        ].join(';');
        content.textContent = message;

        var closeBtn = document.createElement('button');
        closeBtn.textContent = 'Close';
        closeBtn.style.cssText = [
            'margin-top: 20px',
            'padding: 10px 30px',
            'background: #ff4444',
            'color: white',
            'border: none',
            'border-radius: 4px',
            'cursor: pointer',
            'font-size: 14px',
            'font-family: monospace'
        ].join(';');
        closeBtn.onclick = function() {
            hideErrorOverlay();
        };

        overlay.appendChild(title);
        overlay.appendChild(content);
        overlay.appendChild(closeBtn);
        document.body.appendChild(overlay);
    }

    function hideErrorOverlay() {
        var existing = document.getElementById('golem-error-overlay');
        if (existing) {
            existing.parentNode.removeChild(existing);
        }
    }
})();
