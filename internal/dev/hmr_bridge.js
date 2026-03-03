// Golem HMR (Hot Module Replacement) Client
// This script connects to the dev server via WebSocket and handles
// live reload and module-level updates.

(function() {
    'use strict';

    var wsUrl = 'ws://' + window.location.host + '/golem-hmr';
    var ws = null;
    var reconnectAttempts = 0;
    var maxReconnectAttempts = 10;
    var reconnectDelay = 1000;

    function connect() {
        ws = new WebSocket(wsUrl);

        ws.onopen = function() {
            console.log('[Golem HMR] Connected');
            reconnectAttempts = 0;
        };

        ws.onmessage = function(event) {
            var msg;
            try {
                msg = JSON.parse(event.data);
            } catch (e) {
                console.warn('[Golem HMR] Invalid message:', event.data);
                return;
            }

            switch (msg.type) {
                case 'reload':
                    console.log('[Golem HMR] Full reload requested');
                    window.location.reload();
                    break;

                case 'module-update':
                    console.log('[Golem HMR] Module update:', msg.module);
                    handleModuleUpdate(msg.module);
                    break;

                case 'error':
                    console.error('[Golem HMR] Build error:', msg.error);
                    showErrorOverlay(msg.error);
                    break;

                default:
                    console.warn('[Golem HMR] Unknown message type:', msg.type);
            }
        };

        ws.onclose = function() {
            console.log('[Golem HMR] Disconnected');
            if (reconnectAttempts < maxReconnectAttempts) {
                reconnectAttempts++;
                var delay = reconnectDelay * Math.pow(2, reconnectAttempts - 1);
                console.log('[Golem HMR] Reconnecting in ' + delay + 'ms...');
                setTimeout(connect, delay);
            }
        };

        ws.onerror = function(err) {
            console.error('[Golem HMR] WebSocket error:', err);
        };
    }

    function handleModuleUpdate(modulePath) {
        // For WASM-based apps, we need to re-fetch and re-instantiate the
        // affected module. For now, we do a targeted reload by fetching the
        // new WASM module and swapping it in.
        var cacheBuster = '?t=' + Date.now();
        var wasmUrl = modulePath + '.wasm' + cacheBuster;

        // If Golem runtime is available, use its module swap API
        if (window.__golem_hmr && window.__golem_hmr.swapModule) {
            window.__golem_hmr.swapModule(modulePath, wasmUrl);
        } else {
            // Fallback: full reload for module updates
            console.log('[Golem HMR] No module swap API available, falling back to full reload');
            window.location.reload();
        }
    }

    function showErrorOverlay(errorMsg) {
        var existing = document.getElementById('golem-hmr-error');
        if (existing) {
            existing.remove();
        }

        var overlay = document.createElement('div');
        overlay.id = 'golem-hmr-error';
        overlay.style.cssText = 'position:fixed;top:0;left:0;right:0;bottom:0;background:rgba(0,0,0,0.85);color:#ff5555;font-family:monospace;font-size:14px;padding:40px;z-index:99999;overflow:auto;white-space:pre-wrap;';

        var close = document.createElement('button');
        close.textContent = 'X';
        close.style.cssText = 'position:fixed;top:10px;right:20px;background:none;border:2px solid #ff5555;color:#ff5555;font-size:18px;cursor:pointer;padding:4px 10px;z-index:100000;';
        close.onclick = function() { overlay.remove(); };

        var title = document.createElement('h2');
        title.textContent = 'Build Error';
        title.style.cssText = 'color:#ff5555;margin:0 0 20px 0;';

        var content = document.createElement('pre');
        content.textContent = errorMsg;

        overlay.appendChild(close);
        overlay.appendChild(title);
        overlay.appendChild(content);
        document.body.appendChild(overlay);
    }

    // Start connection
    connect();
})();
