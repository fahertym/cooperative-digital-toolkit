async function fetchProposals() {
  const res = await fetch('http://localhost:8080/api/proposals');
  if (!res.ok) throw new Error('Failed to fetch proposals');
  return await res.json();
}

async function createProposal(data) {
  const res = await fetch('http://localhost:8080/api/proposals', {
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

function render(app, proposals) {
  app.innerHTML = `
    <main style="max-width: 720px; margin: 2rem auto; font-family: system-ui, sans-serif;">
      <h1>Cooperative Digital Toolkit</h1>
      <section style="padding:1rem;border:1px solid #ddd;border-radius:12px;margin:1rem 0;">
        <h2>Create Proposal</h2>
        <form id="proposal-form">
          <div style="margin:.5rem 0;">
            <label>Title</label><br/>
            <input id="title" required style="width:100%;padding:.5rem"/>
          </div>
          <div style="margin:.5rem 0;">
            <label>Body</label><br/>
            <textarea id="body" rows="4" style="width:100%;padding:.5rem"></textarea>
          </div>
          <button type="submit">Submit</button>
        </form>
      </section>

      <section>
        <h2>Proposals</h2>
        <ul id="list" style="list-style:none;padding:0;"></ul>
      </section>
    </main>
  `;

  const list = document.querySelector('#list');
  list.innerHTML = proposals.map(p => `
    <li style="border:1px solid #eee;border-radius:12px;padding:1rem;margin:.5rem 0;">
      <strong>#${p.id}: ${p.title}</strong>
      <div style="color:#555; font-size:.9rem;">${new Date(p.created_at).toLocaleString()}</div>
      <p>${(p.body||'').replace(/</g,'&lt;')}</p>
    </li>
  `).join('');

  const form = document.querySelector('#proposal-form');
  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    const title = document.querySelector('#title').value.trim();
    const body = document.querySelector('#body').value.trim();
    try {
      await createProposal({ title, body });
      const updated = await fetchProposals();
      render(app, updated);
    } catch (err) {
      alert(err.message);
    }
  });
}

(async () => {
  const app = document.querySelector('#app');
  try {
    const proposals = await fetchProposals();
    render(app, proposals);
  } catch (err) {
    app.innerHTML = `<p style="color:red;">${err.message}</p>`;
  }
})();
