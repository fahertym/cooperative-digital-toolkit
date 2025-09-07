// frontend/src/main.js

const API_BASE = 'http://localhost:8080';

// --- simple admin mode (no auth yet) ---
// Toggle in UI or set in browser console: localStorage.setItem('isAdmin','1')
function isAdmin() {
  return localStorage.getItem('isAdmin') === '1';
}

async function fetchProposals() {
  const res = await fetch(`${API_BASE}/api/proposals`);
  if (!res.ok) throw new Error('Failed to fetch proposals');
  return await res.json();
}

async function createProposal(data) {
  const res = await fetch(`${API_BASE}/api/proposals`, {
    method: 'POST',
    headers: {'Content-Type': 'application/json'},
    body: JSON.stringify(data),
  });
  if (!res.ok) {
    const txt = await res.text();
    throw new Error(`Create failed: ${res.status} ${txt}`);
  }
  return await res.json();
}

async function closeProposal(id) {
  const res = await fetch(`${API_BASE}/api/proposals/${id}/close`, {
    method: 'POST'
  });
  if (!res.ok) {
    const txt = await res.text();
    let msg = `Close failed: ${res.status}`;
    if (res.status === 404) msg = 'Proposal not found';
    if (res.status === 409) msg = 'Proposal is not open';
    throw new Error(`${msg}${txt ? ` — ${txt}` : ''}`);
  }
  return await res.json();
}

function statusChip(status) {
  return `<span class="badge badge-${status}">${status}</span>`;
}

function render(app, proposals) {
  app.innerHTML = `
    <style>
      :root { --radius: 12px; --muted:#f6f6f6; --border:#e5e5e5; }
      main { max-width: 820px; margin: 2rem auto; font-family: system-ui, sans-serif; line-height: 1.35; }
      .card { padding:1rem;border:1px solid var(--border);border-radius:var(--radius);background:#fff; }
      .row { display:flex; align-items:center; gap:.5rem; justify-content:space-between; }
      .stack { display:grid; gap:.25rem; }
      .badge { display:inline-block; padding:.15rem .5rem; border-radius:999px; font-size:.8rem; text-transform:uppercase; letter-spacing:.02em; border:1px solid var(--border); }
      .badge-open { background:#eefbf0; }
      .badge-closed { background:#f3f4f6; color:#444; }
      .badge-archived { background:#f8f0ff; }
      .btn { padding:.5rem .75rem; border:1px solid var(--border); background:#fff; border-radius:10px; cursor:pointer; }
      .btn[disabled] { opacity:.6; cursor:not-allowed; }
      ul{ list-style:none; padding:0; margin:0; }
      li{ border:1px solid var(--border); border-radius:var(--radius); padding:1rem; margin:.5rem 0; }
      small{ color:#555; }
      textarea, input[type="text"] { width:100%; padding:.5rem; border:1px solid var(--border); border-radius:8px; }
      .toolbar { display:flex; align-items:center; gap:.75rem; }
      .danger { border-color:#ffd3d3; background:#fff6f6; }
    </style>

    <main>
      <header class="row" style="margin-bottom:1rem;">
        <h1>Cooperative Digital Toolkit</h1>
        <label class="toolbar">
          <input id="adminToggle" type="checkbox" ${isAdmin() ? 'checked' : ''}/>
          <span>Admin mode</span>
        </label>
      </header>

      <section class="card" style="margin:1rem 0;">
        <h2>Create Proposal</h2>
        <form id="proposal-form">
          <div class="stack">
            <label>Title</label>
            <input id="title" type="text" required />
          </div>
          <div class="stack" style="margin-top:.5rem;">
            <label>Body</label>
            <textarea id="body" rows="4"></textarea>
          </div>
          <div style="margin-top:.75rem;">
            <button class="btn" type="submit">Submit</button>
          </div>
        </form>
      </section>

      <section>
        <div class="row">
          <h2>Proposals</h2>
          <small>${proposals.length} total</small>
        </div>
        <ul id="list"></ul>
      </section>
    </main>
  `;

  const list = document.querySelector('#list');
  list.innerHTML = proposals.map(p => {
    const created = new Date(p.created_at || Date.now()).toLocaleString();
    const canClose = isAdmin() && (p.status || 'open') === 'open';
    return `
      <li data-id="${p.id}">
        <div class="row">
          <div class="stack">
            <strong>#${p.id}: ${escapeHtml(p.title)}</strong>
            <small>${created}</small>
          </div>
          <div class="toolbar">
            ${statusChip(p.status || 'open')}
            ${canClose ? `<button class="btn danger close-btn" data-id="${p.id}">Close</button>` : ''}
          </div>
        </div>
        ${p.body ? `<p style="margin:.5rem 0 0;">${escapeHtml(p.body)}</p>` : ''}
      </li>
    `;
  }).join('');

  // Admin toggle
  const toggle = document.querySelector('#adminToggle');
  toggle.addEventListener('change', () => {
    if (toggle.checked) localStorage.setItem('isAdmin','1');
    else localStorage.removeItem('isAdmin');
    render(app, proposals); // re-render to show/hide buttons
  });

  // Create proposal form
  const form = document.querySelector('#proposal-form');
  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    const title = document.querySelector('#title').value.trim();
    const body  = document.querySelector('#body').value.trim();
    if (!title) return alert('Title is required.');
    try {
      await createProposal({ title, body });
      const updated = await fetchProposals();
      render(app, updated);
    } catch (err) {
      alert(err.message);
    }
  });

  // Event delegation for Close buttons
  list.addEventListener('click', async (e) => {
    const btn = e.target.closest('.close-btn');
    if (!btn) return;
    const id = btn.getAttribute('data-id');
    btn.disabled = true;
    try {
      await closeProposal(id);
      const updated = await fetchProposals();
      render(app, updated);
    } catch (err) {
      alert(err.message);
      btn.disabled = false;
    }
  });
}

function escapeHtml(str) {
  return String(str).replace(/[&<>"']/g, (m) =>
    ({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;','\'':'&#39;'}[m])
  );
}

// bootstrap
(async () => {
  const app = document.querySelector('#app');
  try {
    const proposals = await fetchProposals();
    render(app, proposals);
  } catch (err) {
    app.innerHTML = `<p style="color:red;">${escapeHtml(err.message)}</p>`;
  }
})();
