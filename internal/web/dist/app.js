// GoZen Skill Management UI
(function () {
  'use strict';

  const API = '/api/v1/bot/skills';

  // --- Tab navigation ---
  document.querySelectorAll('.tab').forEach(tab => {
    tab.addEventListener('click', () => {
      document.querySelectorAll('.tab').forEach(t => t.classList.remove('active'));
      document.querySelectorAll('.panel').forEach(p => p.classList.remove('active'));
      tab.classList.add('active');
      document.getElementById(tab.dataset.tab + '-panel').classList.add('active');
    });
  });

  // --- Skills list ---
  async function loadSkills() {
    const list = document.getElementById('skill-list');
    list.innerHTML = '<p style="color:var(--text-muted)">Loading...</p>';
    try {
      const res = await fetch(API);
      const skills = await res.json();
      if (!skills || skills.length === 0) {
        list.innerHTML = '<p style="color:var(--text-muted)">No skills registered.</p>';
        return;
      }
      list.innerHTML = skills.map(s => renderSkillCard(s)).join('');
    } catch (e) {
      list.innerHTML = '<p style="color:var(--error)">Failed to load skills: ' + e.message + '</p>';
    }
  }

  function renderSkillCard(s) {
    const badge = s.builtin
      ? '<span class="badge badge-builtin">builtin</span>'
      : '<span class="badge badge-custom">custom</span>';
    const keywords = Object.entries(s.keywords || {})
      .map(([lang, kws]) => kws.map(k => '<span>' + lang + ':' + esc(k) + '</span>').join(''))
      .join('');
    const actions = s.builtin
      ? ''
      : '<div class="actions">' +
        '<button class="btn" onclick="window._editSkill(\'' + esc(s.name) + '\')">Edit</button>' +
        '<button class="btn btn-danger" onclick="window._deleteSkill(\'' + esc(s.name) + '\')">Delete</button>' +
        '</div>';
    return '<div class="skill-card">' +
      '<h3>' + esc(s.name) + ' ' + badge + '</h3>' +
      '<div class="meta">Intent: ' + esc(s.intent) + ' | Priority: ' + (s.priority || '-') + '</div>' +
      (s.description ? '<div style="margin-bottom:0.5rem;font-size:0.85rem">' + esc(s.description) + '</div>' : '') +
      '<div class="keywords">' + keywords + '</div>' +
      actions +
      '</div>';
  }

  function esc(s) {
    if (!s) return '';
    const d = document.createElement('div');
    d.textContent = String(s);
    return d.innerHTML;
  }

  // --- Add / Edit skill ---
  const dialog = document.getElementById('skill-dialog');
  const form = document.getElementById('skill-form');
  let editingSkill = null;

  document.getElementById('btn-add-skill').addEventListener('click', () => {
    editingSkill = null;
    document.getElementById('dialog-title').textContent = 'Add Skill';
    form.reset();
    document.getElementById('skill-name').disabled = false;
    dialog.showModal();
  });

  document.getElementById('btn-cancel-skill').addEventListener('click', () => {
    dialog.close();
  });

  window._editSkill = async function (name) {
    try {
      const res = await fetch(API + '/' + encodeURIComponent(name));
      if (!res.ok) throw new Error('Not found');
      const s = await res.json();
      editingSkill = name;
      document.getElementById('dialog-title').textContent = 'Edit Skill';
      document.getElementById('skill-name').value = s.name;
      document.getElementById('skill-name').disabled = true;
      document.getElementById('skill-desc').value = s.description || '';
      document.getElementById('skill-intent').value = s.intent || '';
      document.getElementById('skill-priority').value = s.priority || 50;
      document.getElementById('skill-keywords-en').value = (s.keywords && s.keywords.en) ? s.keywords.en.join(', ') : '';
      document.getElementById('skill-keywords-zh').value = (s.keywords && s.keywords.zh) ? s.keywords.zh.join(', ') : '';
      dialog.showModal();
    } catch (e) {
      alert('Failed to load skill: ' + e.message);
    }
  };

  window._deleteSkill = async function (name) {
    if (!confirm('Delete skill "' + name + '"?')) return;
    try {
      const res = await fetch(API + '/' + encodeURIComponent(name), { method: 'DELETE' });
      if (!res.ok) throw new Error(await res.text());
      loadSkills();
    } catch (e) {
      alert('Failed to delete: ' + e.message);
    }
  };

  form.addEventListener('submit', async (e) => {
    e.preventDefault();
    const keywords = {};
    const enKw = document.getElementById('skill-keywords-en').value.trim();
    const zhKw = document.getElementById('skill-keywords-zh').value.trim();
    if (enKw) keywords.en = enKw.split(',').map(k => k.trim()).filter(Boolean);
    if (zhKw) keywords.zh = zhKw.split(',').map(k => k.trim()).filter(Boolean);

    const body = {
      name: document.getElementById('skill-name').value.trim(),
      description: document.getElementById('skill-desc').value.trim(),
      intent: document.getElementById('skill-intent').value.trim(),
      priority: parseInt(document.getElementById('skill-priority').value) || 50,
      keywords: keywords,
    };

    try {
      const url = editingSkill ? API + '/' + encodeURIComponent(editingSkill) : API;
      const method = editingSkill ? 'PUT' : 'POST';
      const res = await fetch(url, {
        method: method,
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });
      if (!res.ok) {
        const err = await res.json().catch(() => ({ error: 'Unknown error' }));
        throw new Error(err.error || res.statusText);
      }
      dialog.close();
      loadSkills();
    } catch (e) {
      alert('Failed to save: ' + e.message);
    }
  });

  // --- Test matching ---
  document.getElementById('test-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const msg = document.getElementById('test-message').value.trim();
    if (!msg) return;
    const resultDiv = document.getElementById('test-result');
    resultDiv.className = '';
    resultDiv.innerHTML = '<p style="color:var(--text-muted)">Testing...</p>';

    try {
      const res = await fetch(API + '/test', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ message: msg }),
      });
      const data = await res.json();
      if (data.matched) {
        resultDiv.className = 'result-matched';
        resultDiv.innerHTML =
          '<p><strong>Matched!</strong></p>' +
          '<p>Skill: <strong>' + esc(data.skill) + '</strong></p>' +
          '<p>Intent: ' + esc(data.intent) + '</p>' +
          '<p>Confidence: ' + (data.confidence * 100).toFixed(1) + '%</p>' +
          '<p>Method: ' + esc(data.method) + '</p>';
      } else {
        resultDiv.className = 'result-no-match';
        resultDiv.innerHTML = '<p>No match for: <em>' + esc(msg) + '</em></p>';
      }
    } catch (e) {
      resultDiv.innerHTML = '<p style="color:var(--error)">Error: ' + e.message + '</p>';
    }
  });

  // --- Logs ---
  async function loadLogs() {
    const list = document.getElementById('log-list');
    list.innerHTML = '<p style="color:var(--text-muted)">Loading...</p>';
    try {
      const res = await fetch(API + '/logs?limit=50');
      const logs = await res.json();
      if (!logs || logs.length === 0) {
        list.innerHTML = '<p style="color:var(--text-muted)">No match logs yet.</p>';
        return;
      }
      list.innerHTML = logs.map(l => {
        const matched = l.result ? 'Matched: ' + esc(l.result.skill) + ' (' + esc(l.result.intent) + ')' : 'No match';
        const ts = l.timestamp ? new Date(l.timestamp).toLocaleTimeString() : '';
        return '<div class="log-entry">' +
          '<span class="log-input">' + esc(l.input) + '</span> ' +
          '<span class="log-meta">→ ' + matched + ' | LLM: ' + (l.llm_used ? 'Yes' : 'No') + ' | ' + ts + '</span>' +
          '</div>';
      }).join('');
    } catch (e) {
      list.innerHTML = '<p style="color:var(--error)">Failed to load logs: ' + e.message + '</p>';
    }
  }

  document.getElementById('btn-refresh-logs').addEventListener('click', loadLogs);

  // --- Config ---
  async function loadConfig() {
    const container = document.getElementById('config-form-container');
    try {
      const res = await fetch(API + '/config');
      const cfg = await res.json();
      container.innerHTML =
        '<label>Enabled: <input type="checkbox" id="cfg-enabled" ' + (cfg.enabled ? 'checked' : '') + '></label>' +
        '<label>Confidence Threshold: <input type="number" id="cfg-threshold" step="0.05" min="0" max="1" value="' + cfg.confidence_threshold + '"></label>' +
        '<label>LLM Fallback: <input type="checkbox" id="cfg-llm" ' + (cfg.llm_fallback ? 'checked' : '') + '></label>' +
        '<label>Log Buffer Size: <input type="number" id="cfg-bufsize" min="0" value="' + cfg.log_buffer_size + '"></label>' +
        '<button class="btn btn-primary" id="btn-save-config" style="margin-top:1rem">Save Config</button>';

      document.getElementById('btn-save-config').addEventListener('click', saveConfig);
    } catch (e) {
      container.innerHTML = '<p style="color:var(--error)">Failed to load config: ' + e.message + '</p>';
    }
  }

  async function saveConfig() {
    const body = {
      enabled: document.getElementById('cfg-enabled').checked,
      confidence_threshold: parseFloat(document.getElementById('cfg-threshold').value),
      llm_fallback: document.getElementById('cfg-llm').checked,
      log_buffer_size: parseInt(document.getElementById('cfg-bufsize').value),
    };
    try {
      const res = await fetch(API + '/config', {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      });
      if (!res.ok) throw new Error(await res.text());
      alert('Config saved!');
    } catch (e) {
      alert('Failed to save config: ' + e.message);
    }
  }

  // --- Init ---
  loadSkills();
  loadConfig();

  // Auto-load logs when switching to logs tab
  document.querySelector('[data-tab="logs"]').addEventListener('click', loadLogs);
})();
