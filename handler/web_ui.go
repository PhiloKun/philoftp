package handler

// DashboardHTML 是内置的 Web 管理页面（仪表盘）。
// 采用零外部依赖设计：单文件自包含，交互全部基于原生 HTML + TDesign 设计风格 CSS，
// 不引用任何 CDN，确保内网无外网环境也能完美运行与美化。
const DashboardHTML = `<!DOCTYPE html>
<html lang="zh-CN" data-theme="dark">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>PhiloFTP 管理端</title>
<style>
  :root {
    --brand: #0052d9; --brand-hover: #266fe8; --brand-active: #003cab;
    --success: #2ba471; --warning: #e37318; --danger: #d54941;
    --bg-page: #181818; --bg-container: #232324; --bg-container-hover: #2a2a2b;
    --bg-input: #1a1a1a; --border: #424244;
    --text-1: rgba(255,255,255,.9); --text-2: rgba(255,255,255,.6); --text-3: rgba(255,255,255,.4);
    --radius: 9px; --spacer: 16px; --shadow: 0 2px 8px rgba(0,0,0,.3);
  }
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif;
    background: var(--bg-page); color: var(--text-1); min-height: 100vh; font-size: 14px; line-height: 1.5;
  }
  .app-header {
    position: sticky; top: 0; z-index: 50; display: flex; align-items: center; justify-content: space-between;
    padding: 0 32px; height: 64px; background: var(--bg-container); border-bottom: 1px solid var(--border); box-shadow: var(--shadow);
  }
  .brand { display: flex; align-items: center; gap: 10px; font-size: 18px; font-weight: 600; }
  .brand .logo { font-size: 24px; }
  .header-right { display: flex; align-items: center; gap: 14px; }
  .user-chip { display: flex; align-items: center; gap: 8px; color: var(--text-2); font-size: 13px; }
  .user-chip .avatar { width: 28px; height: 28px; border-radius: 50%; background: var(--brand); color: #fff;
    display: inline-flex; align-items: center; justify-content: center; font-size: 13px; font-weight: 600; }
  .container { max-width: 1200px; margin: 24px auto; padding: 0 20px; }
  .page-title { font-size: 16px; font-weight: 600; margin: 4px 0 16px; color: var(--text-1); }

  .stat-grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: var(--spacer); margin-bottom: 24px; }
  .stat-card {
    background: var(--bg-container); border: 1px solid var(--border); border-radius: var(--radius);
    padding: 20px 22px; transition: border-color .2s, transform .2s;
  }
  .stat-card:hover { border-color: var(--brand); transform: translateY(-2px); }
  .stat-card .label { color: var(--text-2); font-size: 13px; margin-bottom: 8px; }
  .stat-card .value { font-size: 26px; font-weight: 700; color: var(--brand); word-break: break-all; }

  .panel { background: var(--bg-container); border: 1px solid var(--border); border-radius: var(--radius); padding: 22px; margin-bottom: 24px; }
  .panel-head { display: flex; align-items: center; justify-content: space-between; margin-bottom: 18px; flex-wrap: wrap; gap: 12px; }
  .panel-head h2 { font-size: 16px; font-weight: 600; display: flex; align-items: center; gap: 8px; }
  .panel-head h2::before { content: ""; width: 4px; height: 16px; border-radius: 2px; background: var(--brand); }
  .toolbar { display: flex; gap: 12px; align-items: center; flex-wrap: wrap; }

  .btn {
    border: none; border-radius: 6px; padding: 8px 16px; font-size: 14px; cursor: pointer;
    display: inline-flex; align-items: center; gap: 6px; line-height: 1; transition: background .2s, border-color .2s, color .2s; white-space: nowrap; font-family: inherit;
  }
  .btn-primary { background: var(--brand); color: #fff; }
  .btn-primary:hover { background: var(--brand-hover); }
  .btn-primary:active { background: var(--brand-active); }
  .btn-default { background: transparent; color: var(--text-1); border: 1px solid var(--border); }
  .btn-default:hover { border-color: var(--brand); color: var(--brand); }
  .btn-danger { background: transparent; color: var(--danger); border: 1px solid var(--border); }
  .btn-danger:hover { border-color: var(--danger); color: var(--danger); }
  .btn-sm { padding: 5px 12px; font-size: 13px; }

  .input, .select {
    padding: 8px 12px; border-radius: 6px; border: 1px solid var(--border); background: var(--bg-input);
    color: var(--text-1); font-size: 14px; font-family: inherit; outline: none; transition: border-color .2s;
  }
  .input:focus, .select:focus { border-color: var(--brand); }
  .input::placeholder { color: var(--text-3); }
  .toolbar .input { flex: 1; min-width: 180px; }
  .toolbar .select { min-width: 160px; }

  .tag { display: inline-block; padding: 2px 10px; border-radius: 10px; font-size: 12px; line-height: 18px; }
  .tag-success { color: var(--success); background: rgba(43,164,113,.15); }
  .tag-warning { color: var(--warning); background: rgba(227,115,24,.15); }
  .tag-default { color: var(--text-2); background: rgba(255,255,255,.08); }
  .tag-brand { color: var(--brand); background: rgba(0,82,217,.15); }

  table.grid { width: 100%; border-collapse: collapse; font-size: 14px; }
  table.grid th, table.grid td { text-align: left; padding: 12px 14px; border-bottom: 1px solid var(--border); }
  table.grid th { color: var(--text-2); font-weight: 500; font-size: 13px; background: var(--bg-page); }
  table.grid tbody tr { transition: background .15s; }
  table.grid tbody tr:hover { background: var(--bg-container-hover); }
  table.grid code { background: var(--bg-page); padding: 2px 6px; border-radius: 4px; font-size: 12px; color: var(--text-2); }
  .op-btns { display: flex; gap: 8px; }

  .path-bar {
    background: var(--bg-input); border: 1px solid var(--border); border-radius: 6px; padding: 10px 14px;
    margin-bottom: 16px; font-family: ui-monospace, monospace; font-size: 13px; color: var(--text-2); word-break: break-all;
    display: flex; align-items: center; gap: 8px;
  }
  .path-bar .crumb { cursor: pointer; }
  .path-bar .crumb:hover { color: var(--brand); }
  .file-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(160px, 1fr)); gap: 14px; }
  .file-item {
    background: var(--bg-input); border: 1px solid var(--border); border-radius: 8px; padding: 16px;
    transition: border-color .15s, transform .15s; cursor: pointer; position: relative;
  }
  .file-item:hover { border-color: var(--brand); transform: translateY(-2px); }
  .file-item .icon { font-size: 34px; margin-bottom: 8px; text-align: center; }
  .file-item .name { font-size: 13px; word-break: break-all; line-height: 1.4; text-align: center; }
  .file-item .meta { font-size: 11px; color: var(--text-3); margin-top: 4px; text-align: center; }
  .file-item .dl {
    position: absolute; top: 8px; right: 8px; width: 26px; height: 26px; border-radius: 6px;
    background: rgba(0,0,0,.35); border: 1px solid var(--border); color: var(--text-2);
    display: none; align-items: center; justify-content: center; cursor: pointer; font-size: 13px;
  }
  .file-item:hover .dl { display: inline-flex; }
  .file-item .dl:hover { color: #fff; border-color: var(--brand); background: var(--brand); }
  .empty { color: var(--text-3); text-align: center; padding: 48px 0; }

  .overlay { position: fixed; inset: 0; background: rgba(0,0,0,.55); z-index: 1000; display: flex; align-items: center; justify-content: center; padding: 20px; }
  .overlay.hidden { display: none; }
  .modal { background: var(--bg-container); border: 1px solid var(--border); border-radius: var(--radius); width: 440px; max-width: 100%; box-shadow: 0 6px 30px rgba(0,0,0,.4); overflow: hidden; }
  .modal-head { padding: 20px 22px 0; font-size: 16px; font-weight: 600; }
  .modal-body { padding: 20px 22px; }
  .modal-foot { padding: 0 22px 20px; display: flex; justify-content: flex-end; gap: 12px; }
  .form-row { margin-bottom: 16px; }
  .form-row > label { display: block; font-size: 13px; color: var(--text-2); margin-bottom: 6px; }
  .form-row .input { width: 100%; }
  .checkbox { display: flex; align-items: center; gap: 8px; cursor: pointer; color: var(--text-1); font-size: 14px; }
  .checkbox input { width: 16px; height: 16px; accent-color: var(--brand); }

  /* 下载进度条 */
  .dl-toast {
    position: fixed; right: 24px; bottom: 24px; width: 320px; max-width: calc(100vw - 48px); z-index: 2000;
    background: var(--bg-container); border: 1px solid var(--border); border-radius: var(--radius);
    padding: 14px 16px; box-shadow: 0 8px 30px rgba(0,0,0,.5);
  }
  .dl-toast .dl-name { font-size: 13px; color: var(--text-1); margin-bottom: 10px; word-break: break-all; display: flex; justify-content: space-between; gap: 10px; }
  .dl-toast .dl-pct { color: var(--brand); font-weight: 600; }
  .dl-bar { height: 7px; border-radius: 4px; background: var(--bg-input); overflow: hidden; }
  .dl-bar > i { display: block; height: 100%; width: 0; background: var(--brand); transition: width .2s; }
  .dl-toast .dl-meta { font-size: 11px; color: var(--text-3); margin-top: 6px; display: flex; justify-content: space-between; }

  #toast { position: fixed; bottom: 24px; left: 50%; transform: translateX(-50%); padding: 12px 22px; border-radius: 10px; font-weight: 600; z-index: 9999; color: #fff; background: var(--brand); opacity: 0; transition: opacity .25s; pointer-events: none; }

  @media (max-width: 640px) {
    .app-header { padding: 0 16px; }
    .container { padding: 0 12px; }
    .stat-grid { grid-template-columns: repeat(auto-fit, minmax(140px, 1fr)); }
    .file-grid { grid-template-columns: repeat(auto-fill, minmax(120px, 1fr)); }
    .toolbar { width: 100%; }
    .toolbar .input, .toolbar .select, .toolbar .btn { flex: 1 1 auto; }
  }
</style>
</head>
<body>

<header class="app-header">
  <div class="brand"><span class="logo">📁</span><span>PhiloFTP 管理端</span></div>
  <div class="header-right">
    <div class="user-chip"><span class="avatar" id="avatar">U</span><span id="curUser">...</span></div>
    <button class="btn btn-default btn-sm" id="logoutBtn">退出</button>
  </div>
</header>

<div class="container">
  <div class="page-title">系统概览</div>
  <div class="stat-grid" id="stats"></div>

  <div class="panel">
    <div class="panel-head">
      <h2>用户管理</h2>
      <div class="toolbar">
        <button class="btn btn-primary" id="addUserBtn"><span>＋</span> 新增用户</button>
        <input class="input" id="userSearch" placeholder="搜索用户名...">
      </div>
    </div>
    <table class="grid" id="userTable">
      <thead><tr><th>用户名</th><th>根目录</th><th>权限</th><th>状态</th><th>操作</th></tr></thead>
      <tbody></tbody>
    </table>
  </div>

  <div class="panel">
    <div class="panel-head">
      <h2>文件管理</h2>
      <div class="toolbar">
        <select class="select" id="fileUser"></select>
        <button class="btn btn-default" id="mkdirBtn">新建目录</button>
        <button class="btn btn-default" id="uploadBtn">上传</button>
        <button class="btn btn-default" id="batchUploadBtn">批量上传</button>
        <button class="btn btn-default" id="refreshBtn">刷新</button>
      </div>
    </div>
    <div class="path-bar" id="pathBar"></div>
    <div class="file-grid" id="fileGrid"></div>
  </div>
</div>

<!-- 用户编辑弹窗 -->
<div class="overlay hidden" id="userDialog">
  <div class="modal">
    <div class="modal-head" id="userDialogTitle">新增用户</div>
    <div class="modal-body">
      <div class="form-row"><label>用户名</label><input class="input" id="uName" placeholder="登录用户名"></div>
      <div class="form-row"><label>密码</label><input class="input" id="uPass" type="password" placeholder="登录密码"></div>
      <div class="form-row"><label>根目录（相对 data 目录）</label><input class="input" id="uHome" placeholder="例如 alice"></div>
      <div class="form-row"><label class="checkbox"><input type="checkbox" id="uRO"> 只读（不可上传/删除/建目录）</label></div>
      <div class="form-row"><label class="checkbox"><input type="checkbox" id="uEnabled" checked> 启用该用户</label></div>
    </div>
    <div class="modal-foot">
      <button class="btn btn-default" id="userCancel">取消</button>
      <button class="btn btn-primary" id="userSave">保存</button>
    </div>
  </div>
</div>

<!-- 新建目录弹窗 -->
<div class="overlay hidden" id="mkdirDialog">
  <div class="modal" style="width:420px">
    <div class="modal-head">新建目录</div>
    <div class="modal-body"><div class="form-row"><label>目录名</label><input class="input" id="dirName" placeholder="新目录名称"></div></div>
    <div class="modal-foot">
      <button class="btn btn-default" id="mkdirCancel">取消</button>
      <button class="btn btn-primary" id="mkdirSave">创建</button>
    </div>
  </div>
</div>

<input type="file" id="uploadInput" class="hidden">
<input type="file" id="batchInput" class="hidden" multiple>

<div id="toast"></div>

<script>
let currentPath = "";
let currentUser = "";
let editing = null;

function notify(msg, type) {
  type = type || 'success';
  const box = document.getElementById('toast');
  box.textContent = msg;
  box.style.background = type === 'error' ? 'var(--danger)' : (type === 'warning' ? 'var(--warning)' : 'var(--brand)');
  box.style.opacity = '1';
  clearTimeout(box._t);
  box._t = setTimeout(() => { box.style.opacity = '0'; }, 2200);
}

// ===== 登录守卫 =====
async function bootstrap() {
  try {
    const res = await fetch('/api/me');
    if (!res.ok) { window.location.href = '/login'; return; }
    const me = await res.json();
    document.getElementById('curUser').textContent = me.username + (me.read_only ? '（只读）' : '');
    document.getElementById('avatar').textContent = (me.username[0] || 'U').toUpperCase();
  } catch (e) { window.location.href = '/login'; }
}

document.getElementById('logoutBtn').onclick = async () => {
  await fetch('/api/logout', { method: 'POST' });
  window.location.href = '/login';
};

async function loadStats() {
  try {
    const [st, sys] = await Promise.all([
      fetch('/api/status').then(r => r.json()),
      fetch('/api/system').then(r => r.json())
    ]);
    const cards = [
      { label: 'FTP 端口', value: st.ftp_port },
      { label: 'Web 端口', value: st.web_port },
      { label: '用户数', value: st.user_count },
      { label: 'PASV 端口', value: st.pasv_ports },
      { label: 'FTPS', value: st.ftps ? '启用' : '禁用' },
      { label: '运行时长', value: st.uptime },
      { label: 'Go 版本', value: sys.go_version },
      { label: 'Goroutines', value: sys.goroutines },
    ];
    document.getElementById('stats').innerHTML = cards.map(c =>
      '<div class="stat-card"><div class="label">' + c.label + '</div><div class="value">' + c.value + '</div></div>'
    ).join('');
  } catch (e) { notify('加载状态失败', 'error'); }
}

async function loadUsers() {
  const q = document.getElementById('userSearch').value;
  const users = await (await fetch('/api/users')).json();
  const list = q ? users.filter(u => u.username.includes(q)) : users;
  const tb = document.querySelector('#userTable tbody');
  tb.innerHTML = list.map(u => {
    const perm = u.read_only ? '<span class="tag tag-warning">只读</span>' : '<span class="tag tag-success">可写</span>';
    const status = u.enabled ? '<span class="tag tag-success">启用</span>' : '<span class="tag tag-default">停用</span>';
    return '<tr>' +
      '<td>' + escapeHtml(u.username) + '</td>' +
      '<td><code>' + escapeHtml(u.home) + '</code></td>' +
      '<td>' + perm + '</td>' +
      '<td>' + status + '</td>' +
      '<td><div class="op-btns">' +
        '<button class="btn btn-default btn-sm" onclick="editUser(\'' + escapeAttr(u.username) + '\')">编辑</button> ' +
        '<button class="btn btn-danger btn-sm" onclick="delUser(\'' + escapeAttr(u.username) + '\')">删除</button>' +
      '</div></td>' +
      '</tr>';
  }).join('');
}

function escapeHtml(s) { return String(s).replace(/[&<>]/g, c => ({'&':'&amp;','<':'&lt;','>':'&gt;'}[c])); }
function escapeAttr(s) { return String(s).replace(/'/g, "\\'"); }

async function loadFiles() {
  if (!currentUser) return;
  const res = await fetch('/api/files?user=' + encodeURIComponent(currentUser) + '&path=' + encodeURIComponent(currentPath));
  const data = await res.json();
  renderPathBar();
  const grid = document.getElementById('fileGrid');
  const items = data.items || [];
  if (!items.length) { grid.innerHTML = '<div class="empty">空目录</div>'; return; }
  grid.innerHTML = items.map(it =>
    '<div class="file-item" ' + (it.is_dir ? 'ondblclick="enterDir(\'' + escapeAttr(it.name) + '\')"' : '') + '>' +
      (it.is_dir ? '' : '<div class="dl" title="下载" onclick="event.stopPropagation();downloadFile(\'' + escapeAttr(it.name) + '\')">⬇</div>') +
      '<div class="icon">' + (it.is_dir ? '📁' : iconFor(it.name)) + '</div>' +
      '<div class="name">' + escapeHtml(it.name) + '</div>' +
      '<div class="meta">' + (it.is_dir ? '目录' : formatSize(it.size)) + '</div>' +
    '</div>'
  ).join('');
}

function iconFor(name) {
  const ext = (name.split('.').pop() || '').toLowerCase();
  const map = {
    pdf: '📕', doc: '📘', docx: '📘', xls: '📗', xlsx: '📗', ppt: '📙', pptx: '📙',
    zip: '🗜️', rar: '🗜️', '7z': '🗜️', tar: '🗜️', gz: '🗜️',
    png: '🖼️', jpg: '🖼️', jpeg: '🖼️', gif: '🖼️', bmp: '🖼️', webp: '🖼️', svg: '🖼️',
    mp3: '🎵', wav: '🎵', mp4: '🎬', avi: '🎬', mov: '🎬', mkv: '🎬',
    txt: '📄', md: '📄', json: '📄', go: '📄', js: '📄', html: '📄', css: '📄',
    exe: '⚙️', apk: '📱', iso: '💿'
  };
  return map[ext] || '📄';
}

function formatSize(b) {
  if (b < 1024) return b + ' B';
  if (b < 1048576) return (b / 1024).toFixed(1) + ' KB';
  if (b < 1073741824) return (b / 1048576).toFixed(1) + ' MB';
  return (b / 1073741824).toFixed(2) + ' GB';
}

function renderPathBar() {
  const bar = document.getElementById('pathBar');
  const parts = currentPath ? currentPath.split('/') : [];
  let html = '<span class="crumb" onclick="gotoCrumb(-1)">🏠 ' + escapeHtml(currentUser) + '</span>';
  let acc = '';
  parts.forEach((p, i) => {
    acc += (acc ? '/' : '') + p;
    html += '<span> / </span><span class="crumb" onclick="gotoCrumb(' + i + ')">' + escapeHtml(p) + '</span>';
  });
  bar.innerHTML = html;
}
function gotoCrumb(idx) {
  const parts = currentPath ? currentPath.split('/') : [];
  if (idx < 0) { currentPath = ''; } else { currentPath = parts.slice(0, idx + 1).join('/'); }
  loadFiles();
}

function enterDir(name) { currentPath = currentPath ? currentPath + '/' + name : name; loadFiles(); }

// ===== 下载（带进度条）=====
async function downloadFile(name) {
  const fullPath = currentPath ? currentPath + '/' + name : name;
  const url = '/api/download?user=' + encodeURIComponent(currentUser) + '&path=' + encodeURIComponent(fullPath);
  const toast = document.createElement('div');
  toast.className = 'dl-toast';
  toast.innerHTML =
    '<div class="dl-name"><span>' + escapeHtml(name) + '</span><span class="dl-pct">0%</span></div>' +
    '<div class="dl-bar"><i></i></div>' +
    '<div class="dl-meta"><span class="dl-done">0 B</span><span class="dl-total">...</span></div>';
  document.body.appendChild(toast);
  const bar = toast.querySelector('.dl-bar > i');
  const pct = toast.querySelector('.dl-pct');
  const doneEl = toast.querySelector('.dl-done');
  const totalEl = toast.querySelector('.dl-total');

  try {
    const resp = await fetch(url);
    if (!resp.ok) { notify('下载失败：' + (resp.status === 404 ? '文件不存在' : resp.status), 'error'); toast.remove(); return; }
    const total = parseInt(resp.headers.get('Content-Length') || '0', 10);
    totalEl.textContent = total ? formatSize(total) : '';
    if (!resp.body) { // 不支持流式，直接下载
      const blob = await resp.blob();
      triggerBlob(blob, name);
      bar.style.width = '100%'; pct.textContent = '100%'; doneEl.textContent = formatSize(blob.size);
      setTimeout(() => toast.remove(), 1500);
      return;
    }
    const reader = resp.body.getReader();
    let received = 0;
    const chunks = [];
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      chunks.push(value);
      received += value.length;
      const p = total ? Math.floor((received / total) * 100) : -1;
      bar.style.width = (p < 0 ? 100 : p) + '%';
      pct.textContent = (p < 0 ? '处理中' : p + '%');
      doneEl.textContent = formatSize(received);
    }
    const blob = new Blob(chunks);
    triggerBlob(blob, name);
    bar.style.width = '100%'; pct.textContent = '完成';
    setTimeout(() => toast.remove(), 1500);
  } catch (e) {
    notify('下载出错', 'error');
    toast.remove();
  }
}
function triggerBlob(blob, name) {
  const a = document.createElement('a');
  a.href = URL.createObjectURL(blob);
  a.download = name;
  document.body.appendChild(a); a.click(); a.remove();
  setTimeout(() => URL.revokeObjectURL(a.href), 3000);
}

async function loadFileUsers() {
  const users = await (await fetch('/api/users')).json();
  const sel = document.getElementById('fileUser');
  sel.innerHTML = users.map(u => '<option value="' + escapeAttr(u.username) + '">' + escapeHtml(u.username) + '</option>').join('');
  if (users.length) { currentUser = users[0].username; }
}

// 用户弹窗
document.getElementById('addUserBtn').onclick = () => openUserModal(null);
document.getElementById('userCancel').onclick = () => closeOverlay('userDialog');
document.getElementById('userSave').onclick = async () => {
  const u = {
    username: document.getElementById('uName').value,
    password: document.getElementById('uPass').value,
    home: document.getElementById('uHome').value || document.getElementById('uName').value,
    read_only: document.getElementById('uRO').checked,
    enabled: document.getElementById('uEnabled').checked,
  };
  if (!u.username || !u.password) { notify('用户名和密码必填', 'warning'); return; }
  const res = await fetch('/api/users', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify(u) });
  if (res.ok) { notify('已保存'); closeOverlay('userDialog'); loadUsers(); loadFileUsers(); }
  else notify('保存失败', 'error');
};

async function editUser(name) {
  const users = await (await fetch('/api/users')).json();
  const u = users.find(x => x.username === name);
  document.getElementById('userDialogTitle').textContent = '编辑用户';
  document.getElementById('uName').value = u.username;
  document.getElementById('uName').setAttribute('disabled', '');
  document.getElementById('uPass').value = u.password;
  document.getElementById('uHome').value = u.home;
  document.getElementById('uRO').checked = u.read_only;
  document.getElementById('uEnabled').checked = u.enabled;
  openOverlay('userDialog');
}
async function delUser(name) {
  if (!confirm('确认删除用户 ' + name + '？')) return;
  const res = await fetch('/api/users/' + encodeURIComponent(name), { method: 'DELETE' });
  if (res.ok) { notify('已删除'); loadUsers(); } else notify('删除失败', 'error');
}

// 文件操作
document.getElementById('refreshBtn').onclick = loadFiles;
document.getElementById('fileUser').addEventListener('change', (e) => { currentUser = e.target.value; currentPath = ''; loadFiles(); });
document.getElementById('mkdirBtn').onclick = () => openOverlay('mkdirDialog');
document.getElementById('mkdirCancel').onclick = () => closeOverlay('mkdirDialog');
document.getElementById('mkdirSave').onclick = async () => {
  const name = document.getElementById('dirName').value;
  if (!name) return;
  const res = await fetch('/api/mkdir', { method: 'POST', headers: { 'Content-Type': 'application/json' }, body: JSON.stringify({ user: currentUser, path: currentPath, name }) });
  if (res.ok) { notify('已创建'); closeOverlay('mkdirDialog'); loadFiles(); }
  else notify('创建失败', 'error');
};
document.getElementById('uploadBtn').onclick = () => document.getElementById('uploadInput').click();
document.getElementById('uploadInput').onchange = async (e) => {
  const file = e.target.files[0]; if (!file) return;
  const fd = new FormData(); fd.append('user', currentUser); fd.append('path', currentPath); fd.append('file', file);
  const res = await fetch('/api/upload', { method: 'POST', body: fd });
  if (res.ok) { notify('已上传'); loadFiles(); } else notify('上传失败', 'error');
  e.target.value = '';
};
document.getElementById('batchUploadBtn').onclick = () => document.getElementById('batchInput').click();
document.getElementById('batchInput').onchange = async (e) => {
  const files = e.target.files; if (!files.length) return;
  const fd = new FormData(); fd.append('user', currentUser); fd.append('path', currentPath);
  for (const f of files) fd.append('files', f);
  const res = await fetch('/api/upload/batch', { method: 'POST', body: fd });
  if (res.ok) { notify('批量上传完成'); loadFiles(); } else notify('上传失败', 'error');
  e.target.value = '';
};

document.getElementById('userSearch').addEventListener('input', loadUsers);

function openOverlay(id) { document.getElementById(id).classList.remove('hidden'); }
function closeOverlay(id) { document.getElementById(id).classList.add('hidden'); }
document.querySelectorAll('.overlay').forEach(o => {
  o.addEventListener('click', (e) => { if (e.target === o) o.classList.add('hidden'); });
});

bootstrap();
loadStats(); loadUsers(); loadFileUsers().then(loadFiles);
setInterval(() => { loadStats(); loadFiles(); }, 10000);
</script>
</body>
</html>`
