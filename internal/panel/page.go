package panel

func HTML() []byte {
	return []byte(panelHTML)
}

const panelHTML = `<!doctype html>
<html lang="en">
<head>
<meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>Codex Retry</title>
<style>
:root {
  color-scheme: light dark;
  --bg: #f7f8fb;
  --surface: #ffffff;
  --surface-soft: #eef2f6;
  --text: #17202a;
  --muted: #617083;
  --border: #dbe2ea;
  --accent: #0f766e;
  --accent-soft: #d7f3ee;
  --danger: #b42318;
  --warn: #a15c07;
  --shadow: 0 12px 32px rgba(18, 31, 46, 0.08);
}

:root.theme-dark,
:root[data-theme="dark"] {
  --bg: #101418;
  --surface: #171c21;
  --surface-soft: #202832;
  --text: #eef3f8;
  --muted: #aab6c3;
  --border: #303a45;
  --accent: #2dd4bf;
  --accent-soft: #123d3a;
  --danger: #ff8a80;
  --warn: #f6c36a;
  --shadow: none;
}

* { box-sizing: border-box; }

body {
  margin: 0;
  min-width: 320px;
  background: var(--bg);
  color: var(--text);
  font: 14px/1.5 Inter, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
}

button, select {
  font: inherit;
}

.shell {
  width: min(1120px, calc(100vw - 32px));
  margin: 0 auto;
  padding: 24px 0 32px;
}

.topbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 18px;
}

h1 {
  margin: 0;
  font-size: 24px;
  line-height: 1.2;
  letter-spacing: 0;
}

.subtitle {
  margin: 6px 0 0;
  color: var(--muted);
}

.actions {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.select, .button {
  min-height: 36px;
  border: 1px solid var(--border);
  background: var(--surface);
  color: var(--text);
  border-radius: 6px;
}

.select {
  min-width: 210px;
  padding: 0 10px;
}

.button {
  padding: 0 14px;
  cursor: pointer;
}

.button:hover {
  border-color: var(--accent);
}

.updated {
  width: 100%;
  color: var(--muted);
  font-size: 12px;
  text-align: right;
}

.updated[data-tone="danger"] {
  color: var(--danger);
}

.grid {
  display: grid;
  grid-template-columns: repeat(5, minmax(0, 1fr));
  gap: 12px;
  margin-bottom: 18px;
}

.metric, .panel {
  background: var(--surface);
  border: 1px solid var(--border);
  border-radius: 8px;
  box-shadow: var(--shadow);
}

.metric {
  min-height: 92px;
  padding: 14px;
}

.metric span {
  display: block;
  color: var(--muted);
  font-size: 12px;
}

.metric strong {
  display: block;
  margin-top: 10px;
  font-size: 24px;
  line-height: 1.1;
  letter-spacing: 0;
}

.metric[data-tone="accent"] strong { color: var(--accent); }
.metric[data-tone="danger"] strong { color: var(--danger); }

.panels {
  display: grid;
  grid-template-columns: minmax(0, 1.3fr) minmax(320px, 0.7fr);
  gap: 14px;
}

.panel {
  overflow: hidden;
}

.panelHeader {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 14px 16px;
  border-bottom: 1px solid var(--border);
}

.panelHeader h2 {
  margin: 0;
  font-size: 15px;
  letter-spacing: 0;
}

.panelHeader span {
  color: var(--muted);
  font-size: 12px;
}

table {
  width: 100%;
  border-collapse: collapse;
}

th, td {
  padding: 11px 16px;
  border-bottom: 1px solid var(--border);
  text-align: right;
  white-space: nowrap;
}

th:first-child, td:first-child {
  text-align: left;
  white-space: normal;
}

th {
  color: var(--muted);
  font-weight: 600;
  font-size: 12px;
  background: var(--surface-soft);
}

.modelName {
  font-weight: 650;
  word-break: break-word;
}

.recentList {
  display: grid;
  gap: 10px;
  padding: 14px;
}

.event {
  border: 1px solid var(--border);
  border-radius: 8px;
  padding: 12px;
  background: color-mix(in srgb, var(--surface) 82%, var(--surface-soft));
}

.eventHeader {
  display: flex;
  justify-content: space-between;
  gap: 10px;
  align-items: flex-start;
}

.eventModel {
  font-weight: 650;
  word-break: break-word;
}

.eventMeta, .empty {
  color: var(--muted);
  font-size: 12px;
}

.badge {
  display: inline-flex;
  align-items: center;
  min-height: 24px;
  padding: 2px 8px;
  border-radius: 999px;
  background: var(--surface-soft);
  color: var(--muted);
  font-size: 12px;
  white-space: nowrap;
}

.badge[data-tone="accent"] {
  background: var(--accent-soft);
  color: var(--accent);
}

.badge[data-tone="danger"] {
  background: color-mix(in srgb, var(--danger) 14%, transparent);
  color: var(--danger);
}

.empty {
  padding: 22px 16px;
}

@media (max-width: 860px) {
  .shell { width: min(100vw - 20px, 1120px); padding-top: 16px; }
  .topbar, .actions { display: grid; justify-content: stretch; }
  .select, .button { width: 100%; }
  .updated { text-align: left; }
  .grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
  .panels { grid-template-columns: minmax(0, 1fr); }
  table { min-width: 680px; }
  .tableWrap { overflow-x: auto; }
}
</style>
</head>
<body>
<main class="shell">
  <header class="topbar">
    <div>
      <h1 data-i18n="title">Codex Retry</h1>
      <p class="subtitle" data-i18n="subtitle">Guard activity since this plugin process started.</p>
    </div>
    <div class="actions">
      <select id="modelFilter" class="select" aria-label="Model"></select>
      <button id="refresh" class="button" type="button" data-i18n="refresh">Refresh</button>
      <div id="updated" class="updated"></div>
    </div>
  </header>

  <section class="grid" aria-label="summary">
    <div class="metric"><span data-i18n="total">Requests</span><strong id="total">0</strong></div>
    <div class="metric" data-tone="accent"><span data-i18n="intercepted">Intercepted</span><strong id="intercepted">0</strong></div>
    <div class="metric" data-tone="accent"><span data-i18n="ratio">Intercept Ratio</span><strong id="ratio">0%</strong></div>
    <div class="metric" data-tone="danger"><span data-i18n="blocked">Blocked</span><strong id="blocked">0</strong></div>
    <div class="metric"><span data-i18n="retries">Retries</span><strong id="retries">0</strong></div>
  </section>

  <section class="panels">
    <div class="panel">
      <div class="panelHeader">
        <h2 data-i18n="byModel">By Model</h2>
        <span id="modelCount"></span>
      </div>
      <div class="tableWrap">
        <table>
          <thead>
            <tr>
              <th data-i18n="model">Model</th>
              <th data-i18n="total">Requests</th>
              <th data-i18n="intercepted">Intercepted</th>
              <th data-i18n="ratio">Intercept Ratio</th>
              <th data-i18n="blocked">Blocked</th>
              <th data-i18n="retries">Retries</th>
            </tr>
          </thead>
          <tbody id="modelRows"></tbody>
        </table>
      </div>
      <div id="modelEmpty" class="empty" hidden data-i18n="empty">No guarded requests yet.</div>
    </div>

    <aside class="panel">
      <div class="panelHeader">
        <h2 data-i18n="recent">Recent Activity</h2>
        <span id="recentCount"></span>
      </div>
      <div id="recentList" class="recentList"></div>
      <div id="recentEmpty" class="empty" hidden data-i18n="empty">No guarded requests yet.</div>
    </aside>
  </section>
</main>

<script>
(function () {
  var STORAGE_KEY = 'cli-proxy-language';
  var messages = {
    en: {
      title: 'Codex Retry',
      subtitle: 'Guard activity since this plugin process started.',
      refresh: 'Refresh',
      allModels: 'All models',
      total: 'Requests',
      intercepted: 'Intercepted',
      ratio: 'Intercept Ratio',
      blocked: 'Blocked',
      retries: 'Retries',
      byModel: 'By Model',
      model: 'Model',
      recent: 'Recent Activity',
      empty: 'No guarded requests yet.',
      updated: 'Updated',
      stream: 'stream',
      nonStream: 'non-stream',
      blockedBadge: 'blocked',
      retriedBadge: 'retried',
      cleanBadge: 'clean',
      models: 'models',
      events: 'events',
      loadFailed: 'Failed to load metrics'
    },
    'zh-TW': {
      title: 'Codex Retry',
      subtitle: '此插件進程啟動以來的 guard 活動統計。',
      refresh: '重新整理',
      allModels: '全部模型',
      total: '請求數',
      intercepted: '攔截數',
      ratio: '攔截比例',
      blocked: '阻擋數',
      retries: '重試數',
      byModel: '依模型統計',
      model: '模型',
      recent: '最近活動',
      empty: '目前還沒有 guarded request。',
      updated: '更新',
      stream: '串流',
      nonStream: '非串流',
      blockedBadge: '已阻擋',
      retriedBadge: '已重試',
      cleanBadge: '通過',
      models: '個模型',
      events: '筆',
      loadFailed: '讀取統計失敗'
    },
    'zh-CN': {
      title: 'Codex Retry',
      subtitle: '此插件进程启动以来的 guard 活动统计。',
      refresh: '刷新',
      allModels: '全部模型',
      total: '请求数',
      intercepted: '拦截数',
      ratio: '拦截比例',
      blocked: '阻挡数',
      retries: '重试数',
      byModel: '按模型统计',
      model: '模型',
      recent: '最近活动',
      empty: '目前还没有 guarded request。',
      updated: '更新',
      stream: '流式',
      nonStream: '非流式',
      blockedBadge: '已阻挡',
      retriedBadge: '已重试',
      cleanBadge: '通过',
      models: '个模型',
      events: '条',
      loadFailed: '读取统计失败'
    }
  };

  var state = { lang: resolveLanguage(), snapshot: null, model: '' };
  var els = {
    modelFilter: document.getElementById('modelFilter'),
    refresh: document.getElementById('refresh'),
    updated: document.getElementById('updated'),
    total: document.getElementById('total'),
    intercepted: document.getElementById('intercepted'),
    ratio: document.getElementById('ratio'),
    blocked: document.getElementById('blocked'),
    retries: document.getElementById('retries'),
    modelRows: document.getElementById('modelRows'),
    modelEmpty: document.getElementById('modelEmpty'),
    modelCount: document.getElementById('modelCount'),
    recentList: document.getElementById('recentList'),
    recentEmpty: document.getElementById('recentEmpty'),
    recentCount: document.getElementById('recentCount')
  };

  function normalizeLanguage(value) {
    value = String(value || '').trim();
    var lower = value.toLowerCase();
    if (lower === 'zh-tw' || lower === 'zh-hk' || lower === 'zh-mo' || lower.indexOf('zh-hant') === 0) return 'zh-TW';
    if (lower === 'zh-cn' || lower.indexOf('zh') === 0) return 'zh-CN';
    if (lower.indexOf('en') === 0) return 'en';
    return '';
  }

  function parseStoredLanguage() {
    try {
      var raw = localStorage.getItem(STORAGE_KEY);
      if (!raw) return '';
      var parsed = JSON.parse(raw);
      return normalizeLanguage((parsed && parsed.state && parsed.state.language) || parsed.language || parsed);
    } catch (_) {
      try { return normalizeLanguage(localStorage.getItem(STORAGE_KEY)); } catch (_) { return ''; }
    }
  }

  function parentLanguage() {
    try {
      return normalizeLanguage(window.parent && window.parent.document && window.parent.document.documentElement.lang);
    } catch (_) {
      return '';
    }
  }

  function resolveLanguage() {
    var params = new URLSearchParams(location.search);
    return normalizeLanguage(params.get('lang')) ||
      parseStoredLanguage() ||
      parentLanguage() ||
      normalizeLanguage((navigator.languages && navigator.languages[0]) || navigator.language) ||
      'en';
  }

  function t(key) {
    return (messages[state.lang] && messages[state.lang][key]) || messages.en[key] || key;
  }

  function applyLanguage() {
    state.lang = resolveLanguage();
    document.documentElement.lang = state.lang;
    document.querySelectorAll('[data-i18n]').forEach(function (node) {
      node.textContent = t(node.getAttribute('data-i18n'));
    });
    render();
  }

  function number(value) {
    return new Intl.NumberFormat(state.lang).format(Number(value || 0));
  }

  function percent(value) {
    return new Intl.NumberFormat(state.lang, { style: 'percent', maximumFractionDigits: 1 }).format(Number(value || 0));
  }

  function formatTime(value) {
    if (!value) return '';
    var date = new Date(value);
    if (Number.isNaN(date.getTime())) return '';
    return date.toLocaleString(state.lang);
  }

  function selectedStats(snapshot) {
    if (!state.model) return snapshot;
    var found = (snapshot.models || []).find(function (item) { return item.model === state.model; });
    return found || { total_requests: 0, intercepted_requests: 0, blocked_requests: 0, retry_attempts: 0, intercept_ratio: 0 };
  }

  function syncModelFilter(models) {
    var previous = state.model;
    var exists = !previous || models.some(function (item) { return item.model === previous; });
    if (!exists) previous = '';
    state.model = previous;

    els.modelFilter.innerHTML = '';
    var all = document.createElement('option');
    all.value = '';
    all.textContent = t('allModels');
    els.modelFilter.appendChild(all);
    models.forEach(function (item) {
      var option = document.createElement('option');
      option.value = item.model;
      option.textContent = item.model;
      els.modelFilter.appendChild(option);
    });
    els.modelFilter.value = state.model;
  }

  function renderModels(models) {
    var filtered = state.model ? models.filter(function (item) { return item.model === state.model; }) : models;
    els.modelRows.innerHTML = '';
    filtered.forEach(function (item) {
      var row = document.createElement('tr');
      row.innerHTML =
        '<td><span class="modelName"></span></td>' +
        '<td>' + number(item.total_requests) + '</td>' +
        '<td>' + number(item.intercepted_requests) + '</td>' +
        '<td>' + percent(item.intercept_ratio) + '</td>' +
        '<td>' + number(item.blocked_requests) + '</td>' +
        '<td>' + number(item.retry_attempts) + '</td>';
      row.querySelector('.modelName').textContent = item.model;
      els.modelRows.appendChild(row);
    });
    els.modelEmpty.hidden = filtered.length > 0;
    els.modelCount.textContent = number(models.length) + ' ' + t('models');
  }

  function renderRecent(recent) {
    var filtered = state.model ? recent.filter(function (item) { return item.model === state.model; }) : recent;
    els.recentList.innerHTML = '';
    filtered.slice(0, 30).forEach(function (item) {
      var event = document.createElement('div');
      event.className = 'event';
      var badge = item.blocked ? t('blockedBadge') : (item.intercepted ? t('retriedBadge') : t('cleanBadge'));
      var tone = item.blocked ? 'danger' : (item.intercepted ? 'accent' : '');
      event.innerHTML =
        '<div class="eventHeader">' +
        '<div><div class="eventModel"></div><div class="eventMeta"></div></div>' +
        '<span class="badge"></span>' +
        '</div>';
      event.querySelector('.eventModel').textContent = item.model || '(unknown)';
      event.querySelector('.eventMeta').textContent = [
        formatTime(item.timestamp),
        item.stream ? t('stream') : t('nonStream'),
        item.reason || item.error_code || ''
      ].filter(Boolean).join(' / ');
      var badgeNode = event.querySelector('.badge');
      badgeNode.textContent = badge;
      if (tone) badgeNode.setAttribute('data-tone', tone);
      els.recentList.appendChild(event);
    });
    els.recentEmpty.hidden = filtered.length > 0;
    els.recentCount.textContent = number(filtered.length) + ' ' + t('events');
  }

  function render() {
    var snapshot = state.snapshot;
    if (!snapshot) {
      syncModelFilter([]);
      return;
    }
    var models = snapshot.models || [];
    syncModelFilter(models);
    var stats = selectedStats(snapshot);
    els.total.textContent = number(stats.total_requests);
    els.intercepted.textContent = number(stats.intercepted_requests);
    els.ratio.textContent = percent(stats.intercept_ratio);
    els.blocked.textContent = number(stats.blocked_requests);
    els.retries.textContent = number(stats.retry_attempts);
    els.updated.removeAttribute('data-tone');
    els.updated.textContent = t('updated') + ' ' + formatTime(snapshot.generated_at);
    renderModels(models);
    renderRecent(snapshot.recent || []);
  }

  async function load() {
    var url = new URL(location.href);
    url.searchParams.set('format', 'json');
    var response = await fetch(url.toString(), { cache: 'no-store' });
    if (!response.ok) throw new Error('HTTP ' + response.status);
    state.snapshot = await response.json();
    render();
  }

  function showError(error) {
    els.updated.setAttribute('data-tone', 'danger');
    els.updated.textContent = t('loadFailed') + ': ' + (error && error.message ? error.message : String(error));
  }

  els.refresh.addEventListener('click', function () { load().catch(showError); });
  els.modelFilter.addEventListener('change', function () {
    state.model = els.modelFilter.value;
    render();
  });

  applyLanguage();
  load().catch(showError);
  window.setInterval(function () { load().catch(showError); }, 5000);

  try {
    var parentElement = window.parent && window.parent.document && window.parent.document.documentElement;
    if (parentElement && window.MutationObserver) {
      new MutationObserver(applyLanguage).observe(parentElement, { attributes: true, attributeFilter: ['lang'] });
    }
  } catch (_) {}
})();
</script>
</body>
</html>`
