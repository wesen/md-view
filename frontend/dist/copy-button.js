// Copy-to-clipboard button for <pre><code> blocks.
// Adds a small clipboard icon button to every code block that copies
// the code text to the clipboard with visual feedback.
(function() {
    var preBlocks = document.querySelectorAll('pre code');
    if (preBlocks.length === 0) return;

    // SVG icons
    var clipboardIcon = '<svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><rect x="5" y="5" width="9" height="9" rx="1.5"/><path d="M3 11V3.5A1.5 1.5 0 014.5 2H11"/><path d="M5.5 5.5V3.5A1.5 1.5 0 017 2h5.5A1.5 1.5 0 0114 3.5v6A1.5 1.5 0 0112.5 11H11"/></svg>';
    var checkIcon = '<svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="3.5 8.5 6.5 11.5 12.5 4.5"/></svg>';

    preBlocks.forEach(function(codeBlock) {
        var pre = codeBlock.parentElement;
        if (!pre || pre.tagName !== 'PRE') return;

        // Create wrapper container
        var container = document.createElement('div');
        container.className = 'md-view-code-container';

        // Create copy button
        var button = document.createElement('button');
        button.className = 'md-view-copy-btn';
        button.title = 'Copy to clipboard';
        button.setAttribute('aria-label', 'Copy to clipboard');
        button.innerHTML = clipboardIcon;

        // Insert wrapper before pre, move pre into wrapper, add button
        pre.parentNode.insertBefore(container, pre);
        container.appendChild(pre);
        container.appendChild(button);

        button.addEventListener('click', function() {
            var text = codeBlock.textContent;
            navigator.clipboard.writeText(text).then(function() {
                button.innerHTML = checkIcon;
                button.title = 'Copied!';
                button.classList.add('md-view-copy-btn-success');
                setTimeout(function() {
                    button.innerHTML = clipboardIcon;
                    button.title = 'Copy to clipboard';
                    button.classList.remove('md-view-copy-btn-success');
                }, 2000);
            }).catch(function() {
                // Fallback: select text
                var range = document.createRange();
                range.selectNodeContents(codeBlock);
                var sel = window.getSelection();
                sel.removeAllRanges();
                sel.addRange(range);
            });
        });
    });
})();
