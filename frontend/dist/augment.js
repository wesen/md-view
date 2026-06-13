// augment.js — content augmentation for the md-view desktop frontend.
//
// In the daemon/HTTP model these ran once as IIFEs on page load (see
// pkg/renderer/static/{copy-button,mermaid-init}.js). In the Wails model the
// page chrome is stable and only #content is swapped per file / live reload,
// so augmentation must be RE-RUNNABLE against the current DOM.
//
// Exposes:
//   window.MDSAugmentPage()      — inject copy buttons + render mermaid blocks
//   window.MDSMermaidRerender(t) — re-render mermaid diagrams for a new theme
//
// All operations are idempotent (safe to call repeatedly after content swaps).
(function () {
    var clipboardIcon = '<svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><rect x="5" y="5" width="9" height="9" rx="1.5"/><path d="M3 11V3.5A1.5 1.5 0 014.5 2H11"/><path d="M5.5 5.5V3.5A1.5 1.5 0 017 2h5.5A1.5 1.5 0 0114 3.5v6A1.5 1.5 0 0112.5 11H11"/></svg>';
    var checkIcon = '<svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="3.5 8.5 6.5 11.5 12.5 4.5"/></svg>';

    function currentTheme() {
        return document.documentElement.getAttribute('data-theme') === 'dark' ? 'dark' : 'default';
    }

    // --- Copy-to-clipboard buttons for <pre><code> blocks (idempotent) ---
    function initCopyButtons() {
        var blocks = document.querySelectorAll('#content pre code');
        for (var i = 0; i < blocks.length; i++) {
            var codeBlock = blocks[i];
            var pre = codeBlock.parentElement;
            if (!pre || pre.tagName !== 'PRE') continue;
            // Mermaid blocks are handled by initMermaid, not as copyable code.
            if (codeBlock.classList.contains('language-mermaid')) continue;
            // Skip if already wrapped.
            if (pre.parentElement && pre.parentElement.classList.contains('md-view-code-container')) continue;

            var container = document.createElement('div');
            container.className = 'md-view-code-container';

            var button = document.createElement('button');
            button.className = 'md-view-copy-btn';
            button.title = 'Copy to clipboard';
            button.setAttribute('aria-label', 'Copy to clipboard');
            button.innerHTML = clipboardIcon;

            pre.parentNode.insertBefore(container, pre);
            container.appendChild(pre);
            container.appendChild(button);

            (function (codeEl, btn) {
                btn.addEventListener('click', function () {
                    var text = codeEl.textContent;
                    navigator.clipboard.writeText(text).then(function () {
                        btn.innerHTML = checkIcon;
                        btn.title = 'Copied!';
                        btn.classList.add('md-view-copy-btn-success');
                        setTimeout(function () {
                            btn.innerHTML = clipboardIcon;
                            btn.title = 'Copy to clipboard';
                            btn.classList.remove('md-view-copy-btn-success');
                        }, 2000);
                    });
                });
            })(codeBlock, button);
        }
    }

    // --- Mermaid: convert ```mermaid code blocks into rendered diagrams ---
    function initMermaid() {
        if (typeof mermaid === 'undefined') return;
        var codeBlocks = document.querySelectorAll('#content code.language-mermaid');
        var nodes = [];
        for (var i = 0; i < codeBlocks.length; i++) {
            var codeBlock = codeBlocks[i];
            var pre = codeBlock.parentElement;
            if (!pre || pre.tagName !== 'PRE') continue;
            // Skip if already converted.
            if (pre.parentElement && pre.parentElement.classList.contains('mermaid')) continue;

            var source = codeBlock.textContent;
            var div = document.createElement('div');
            div.className = 'mermaid';
            div.setAttribute('data-mermaid-source', source);
            div.textContent = source;
            pre.parentNode.replaceChild(div, pre);
            nodes.push(div);
        }
        if (nodes.length === 0) return;

        mermaid.initialize({
            startOnLoad: false,
            theme: currentTheme(),
            securityLevel: 'loose'
        });
        try { mermaid.run({ nodes: nodes }); } catch (e) { console.error('Mermaid render error:', e); }
    }

    // Re-render all mermaid diagrams for a new theme.
    window.MDSMermaidRerender = function (theme) {
        if (typeof mermaid === 'undefined') return;
        var mermaidTheme = theme === 'dark' ? 'dark' : 'default';
        mermaid.initialize({ startOnLoad: false, theme: mermaidTheme, securityLevel: 'loose' });
        var divs = document.querySelectorAll('#content .mermaid');
        for (var i = 0; i < divs.length; i++) {
            var div = divs[i];
            var source = div.getAttribute('data-mermaid-source');
            if (!source) continue;
            div.innerHTML = '';
            div.removeAttribute('data-processed');
            div.textContent = source;
        }
        try { mermaid.run({ nodes: Array.prototype.slice.call(divs) }); } catch (e) { console.error('Mermaid re-render error:', e); }
    };

    // Main entry point: call after every content swap.
    window.MDSAugmentPage = function () {
        initCopyButtons();
        initMermaid();
    };
})();
