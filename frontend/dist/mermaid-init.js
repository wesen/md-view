// Mermaid initialization for md-view.
// Detects ```mermaid code blocks and renders them as SVG diagrams.
// Re-renders diagrams when the theme changes.
(function() {
    // Check if there are any mermaid code blocks on the page
    var mermaidBlocks = document.querySelectorAll('code.language-mermaid');
    if (mermaidBlocks.length === 0) return;

    // Wrap each <code class="language-mermaid"> in a <div class="mermaid">
    mermaidBlocks.forEach(function(codeBlock) {
        var pre = codeBlock.parentElement;
        if (pre && pre.tagName === 'PRE') {
            var div = document.createElement('div');
            div.className = 'mermaid';
            div.setAttribute('data-mermaid-source', codeBlock.textContent);
            div.textContent = codeBlock.textContent;
            pre.parentNode.replaceChild(div, pre);
        }
    });

    // Initialize mermaid with the current theme
    var isDark = document.documentElement.getAttribute('data-theme') === 'dark';
    mermaid.initialize({
        startOnLoad: true,
        theme: isDark ? 'dark' : 'default',
        securityLevel: 'loose'
    });

    // Re-render mermaid diagrams when theme toggles
    // We watch for data-theme attribute changes on <html>
    var observer = new MutationObserver(function(mutations) {
        mutations.forEach(function(mutation) {
            if (mutation.attributeName === 'data-theme') {
                var nowDark = document.documentElement.getAttribute('data-theme') === 'dark';
                var theme = nowDark ? 'dark' : 'default';

                // Re-render all mermaid diagrams with the new theme
                mermaid.initialize({ theme: theme, startOnLoad: false, securityLevel: 'loose' });

                document.querySelectorAll('.mermaid').forEach(function(div) {
                    var source = div.getAttribute('data-mermaid-source');
                    if (!source) return;

                    // Remove the old rendered SVG
                    div.innerHTML = '';
                    div.removeAttribute('data-processed');
                    div.textContent = source;

                    // Let mermaid re-render
                    try { mermaid.run({ nodes: [div] }); } catch(e) { console.error('Mermaid re-render error:', e); }
                });
            }
        });
    });

    observer.observe(document.documentElement, { attributes: true, attributeFilter: ['data-theme'] });
})();
