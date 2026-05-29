// reMarkable upload button for md-view.
// Adds a fixed-position button that uploads the current markdown file
// to a reMarkable device via the server's /upload-remarkable endpoint.
(function() {
    // SVG icons
    var tabletIcon = '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linecap="round" stroke-linejoin="round"><rect x="4" y="2" width="16" height="20" rx="2"/><line x1="12" y1="18" x2="12" y2="18.01"/></svg>';
    var spinnerIcon = '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 12a9 9 0 1 1-6.219-8.56"/></svg>';
    var checkIcon = '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><polyline points="20 6 9 17 4 12"/></svg>';
    var errorIcon = '<svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><line x1="18" y1="6" x2="6" y2="18"/><line x1="6" y1="6" x2="18" y2="18"/></svg>';

    // Find the file path from the URL
    var params = new URLSearchParams(window.location.search);
    var filePath = params.get('file');
    if (!filePath) return;

    // Find the theme toggle to position relative to it
    var themeToggle = document.querySelector('.md-view-theme-toggle');

    // Create button
    var btn = document.createElement('button');
    btn.className = 'md-view-remarkable-btn';
    btn.title = 'Upload to reMarkable';
    btn.setAttribute('aria-label', 'Upload to reMarkable');
    btn.innerHTML = tabletIcon;

    // Position: to the left of the theme toggle
    if (themeToggle) {
        themeToggle.parentNode.insertBefore(btn, themeToggle);
    } else {
        document.body.appendChild(btn);
    }

    // Create status toast (hidden by default)
    var toast = document.createElement('div');
    toast.className = 'md-view-remarkable-toast';
    toast.style.display = 'none';
    document.body.appendChild(toast);

    function showToast(html, duration, className) {
        toast.innerHTML = html;
        toast.className = 'md-view-remarkable-toast' + (className ? ' ' + className : '');
        toast.style.display = 'block';
        if (duration) {
            setTimeout(function() {
                toast.style.display = 'none';
            }, duration);
        }
    }

    function hideToast() {
        toast.style.display = 'none';
    }

    btn.addEventListener('click', function() {
        // Don't double-submit
        if (btn.disabled) return;
        btn.disabled = true;
        btn.innerHTML = spinnerIcon;
        btn.classList.add('md-view-remarkable-btn-loading');
        showToast('Uploading to reMarkable...', 0, 'md-view-remarkable-toast-loading');

        var url = '/upload-remarkable?file=' + encodeURIComponent(filePath);

        fetch(url, { method: 'POST' })
            .then(function(resp) {
                return resp.json().then(function(data) {
                    return { ok: resp.ok, status: resp.status, data: data };
                });
            })
            .then(function(result) {
                btn.disabled = false;
                btn.innerHTML = tabletIcon;
                btn.classList.remove('md-view-remarkable-btn-loading');

                if (result.ok && result.data.status === 'ok') {
                    btn.innerHTML = checkIcon;
                    btn.classList.add('md-view-remarkable-btn-success');
                    showToast('✓ ' + (result.data.message || 'Uploaded successfully'), 4000, 'md-view-remarkable-toast-success');
                    setTimeout(function() {
                        btn.innerHTML = tabletIcon;
                        btn.classList.remove('md-view-remarkable-btn-success');
                    }, 3000);
                } else {
                    btn.innerHTML = errorIcon;
                    btn.classList.add('md-view-remarkable-btn-error');
                    var errMsg = result.data.message || 'Upload failed';
                    showToast('✗ ' + errMsg, 6000, 'md-view-remarkable-toast-error');
                    setTimeout(function() {
                        btn.innerHTML = tabletIcon;
                        btn.classList.remove('md-view-remarkable-btn-error');
                    }, 5000);
                }
            })
            .catch(function(err) {
                btn.disabled = false;
                btn.innerHTML = errorIcon;
                btn.classList.add('md-view-remarkable-btn-error');
                showToast('✗ Network error: ' + err.message, 5000, 'md-view-remarkable-toast-error');
                setTimeout(function() {
                    btn.innerHTML = tabletIcon;
                    btn.classList.remove('md-view-remarkable-btn-error');
                }, 5000);
            });
    });
})();
