// Top-bar utility buttons for md-view.
// Adds "Copy path" and "Download markdown" buttons next to the reMarkable button.
(function() {
    var params = new URLSearchParams(window.location.search);
    var filePath = params.get('file');
    if (!filePath) return;

    var checkIcon = '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>';
    var clipboardIcon = '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><rect x="8" y="2" width="12" height="14" rx="2"/><path d="M16 4H6a2 2 0 0 0-2 2v14a2 2 0 0 0 2 2h10"/></svg>';
    var downloadIcon = '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><polyline points="7 10 12 15 17 10"/><line x1="12" y1="15" x2="12" y2="3"/></svg>';

    // Copy path button
    var copyBtn = document.createElement('button');
    copyBtn.className = 'md-view-toolbar-btn md-view-copy-path-btn';
    copyBtn.title = filePath;
    copyBtn.setAttribute('aria-label', 'Copy file path to clipboard');
    copyBtn.innerHTML = clipboardIcon;

    copyBtn.addEventListener('click', function() {
        navigator.clipboard.writeText(filePath).then(function() {
            copyBtn.innerHTML = checkIcon;
            copyBtn.classList.add('md-view-toolbar-btn-success');
            copyBtn.title = 'Copied!';
            setTimeout(function() {
                copyBtn.innerHTML = clipboardIcon;
                copyBtn.classList.remove('md-view-toolbar-btn-success');
                copyBtn.title = filePath;
            }, 2000);
        });
    });

    // Download markdown button
    var dlBtn = document.createElement('button');
    dlBtn.className = 'md-view-toolbar-btn md-view-download-btn';
    dlBtn.title = 'Download markdown';
    dlBtn.setAttribute('aria-label', 'Download markdown file');

    // Use an <a> inside the button so the download attribute works
    var dlLink = document.createElement('a');
    dlLink.href = '/raw?file=' + encodeURIComponent(filePath);
    var fileName = filePath.split('/').pop();
    dlLink.download = fileName;
    dlLink.innerHTML = downloadIcon;
    dlLink.style.cssText = 'display:flex;align-items:center;color:inherit;text-decoration:none;';
    dlBtn.appendChild(dlLink);

    // Insert both buttons before the reMarkable button (or theme toggle)
    var target = document.querySelector('.md-view-remarkable-btn') ||
                 document.querySelector('.md-view-theme-toggle');
    if (target && target.parentNode) {
        target.parentNode.insertBefore(dlBtn, target);
        target.parentNode.insertBefore(copyBtn, dlBtn);
    } else {
        document.body.appendChild(copyBtn);
        document.body.appendChild(dlBtn);
    }
})();
