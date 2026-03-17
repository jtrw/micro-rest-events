// Modal management
function showModal() {
    document.getElementById('modal-backdrop').classList.add('show');
}

function hideModal() {
    document.getElementById('modal-backdrop').classList.remove('show');
    document.getElementById('modal-content').innerHTML = '';
}

// Close modal on escape
document.addEventListener('keydown', function (e) {
    if (e.key === 'Escape') hideModal();
});

// Close modal after successful HTMX swap into events-table
document.body.addEventListener('htmx:afterSwap', function (e) {
    if (e.detail.target && e.detail.target.id === 'events-table') {
        hideModal();
    }
});

// Theme
const savedTheme = document.cookie.match(/theme=([^;]+)/);
if (savedTheme) {
    document.documentElement.setAttribute('data-theme', savedTheme[1]);
} else if (window.matchMedia('(prefers-color-scheme: dark)').matches) {
    document.documentElement.setAttribute('data-theme', 'dark');
}

function toggleTheme() {
    const current = document.documentElement.getAttribute('data-theme') || '';
    const next = current === 'dark' ? '' : 'dark';
    if (next === '') {
        document.documentElement.removeAttribute('data-theme');
    } else {
        document.documentElement.setAttribute('data-theme', next);
    }
    document.cookie = 'theme=' + next + '; path=/; max-age=31536000';
}
