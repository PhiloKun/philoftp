package handler

// DashboardHTML 是内置的 Web 管理页面。
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
    --brand: #0052d9;
    --brand-hover: #266fe8;
    --brand-active: #003cab;
    --success: #2ba471;
    --warning: #e37318;
    --danger: #d54941;
    --bg-page: #181818;
    --bg-container: #232324;
    --bg-container-hover: #2a2a2b;
    --bg-input: #1a1a1a;
    --border: #424244;
    --text-1: rgba(255,255,255,.9);
    --text-2: rgba(255,255,255,.6);
    --text-3: rgba(255,255,255,.4);
    --radius: 9px;
    --spacer: 16px;
    --shadow: 0 2px 8px rgba(0,0,0,.3);
  }
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif;
    background: var(--bg-page);
    color: var(--text-1);
    min-height: 100vh;
    font-size: 14px;
    line-height: 1.5;
  }

  /* ===== 顶栏 ===== */
  .app-header {
    position: sticky; top: 0; z-index: 50;
    display: flex; align-items: center; justify-content: space-between;
    padding: 0 32px; height: 64px;
    background: var(--bg-container);
    border-bottom: 1px solid var(--border);
    box-shadow: var(--shadow);
  }
  .brand { display: flex; align-items: center; gap: 10px; font-size: 18px; font-weight: 600; }
  .brand .logo { font-size: 24px; }

  /* ===== 布局 ===== */
  .container { max-width: 1200px; margin: 24px auto; padding: 0 20px; }
  .page-title { font-size: 16px; font-weight: 600; margin: 4px 0 16px; color: var(--text-1); }

  /* ===== 状态卡片 ===== */
  .stat-grid {
    display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
    gap: var(--spacer); margin-bottom: 24px;
  }
  .stat-card {
    background: var(--bg-container);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 20px 22px;
    transition: border-color .2s, transform .2s;
  }
  .stat-card:hover { border-color: var(--brand); transform: translateY(-2px); }
  .stat-card .label { color: var(--text-2); font-size: 13px; margin-bottom: 8px; }
  .stat-card .value { font-size: 26px; font-weight: 700; color: var(--brand); word-break: break-all; }

  /* ===== 面板 ===== */
  .panel {
    background: var(--bg-container);
    border: 1px solid var(--border);
    border-radius: var(--radius);
    padding: 22px; margin-bottom: 24px;
  }
  .panel-head {
    display: flex; align-items: center; justify-content: space-between;
    margin-bottom: 18px; flex-wrap: wrap; gap: 12px;
  }
  .panel-head h2 { font-size: 16px; font-weight: 600; display: flex; align-items: center; gap: 8px; }
  .panel-head h2::before {
    content: ""; width: 4px; height: 16px; border-radius: 2px; background: var(--brand);
  }
  .toolbar { display: flex; gap: 12px; align-items: center; flex-wrap: wrap; }

  /* ===== 按钮（TDesign 风格） ===== */
  .btn {
    border: none; border-radius: 6px; padding: 8px 16px; font-size: 14px; cursor: pointer;
    display: inline-flex; align-items: center; gap: 6px; line-height: 1;
    transition: background .2s, border-color .2s, color .2s; white-space: nowrap;
    font-family: inherit;
  }
  .btn-primary { background: var(--brand); color: #fff; }
  .btn-primary:hover { background: var(--brand-hover); }
  .btn-primary:active { background: var(--brand-active); }
  .btn-default { background: transparent; color: var(--text-1); border: 1px solid var(--border); }
  .btn-default:hover { border-color: var(--brand); color: var(--brand); }
  .btn-danger { background: transparent; color: var(--danger); border: 1px solid var(--border); }
  .btn-danger:hover { border-color: var(--danger); color: var(--danger); }
  .btn-sm { padding: 5px 12px; font-size: 13px; }

  /* ===== 输入框 / 选择框 ===== */
  .input, .select {
    padding: 8px 12px; border-radius: 6px;
    border: 1px solid var(--border); background: var(--bg-input); color: var(--text-1);
    font-size: 14px; font-family: inherit; outline: none; transition: border-color .2s;
  }
  .input:focus, .select:focus { border-color: var(--brand); }
  .input::placeholder { color: var(--text-3); }
  .toolbar .input { flex: 1; min-width: 180px; }
  .toolbar .select { min-width: 160px; }

  /* ===== 标签 ===== */
  .tag {
    display: inline-block; padding: 2px 10px; border-radius: 10px; font-size: 12px; line-height: 18px;
  }
  .tag-success { color: var(--success); background: rgba(43,164,113,.15); }
  .tag-warning { color: var(--warning); background: rgba(227,115,24,.15); }
  .tag-default { color: var(--text-2); background: rgba(255,255,255,.08); }
  .tag-brand { color: var(--brand); background: rgba(0,82,217,.15); }

  /* ===== 用户表格 ===== */
  table.grid { width: 100%; border-collapse: collapse; font-size: 14px; }
  table.grid th, table.grid td { text-align: left; padding: 12px 14px; border-bottom: 1px solid var(--border); }
  table.grid th { color: var(--text-2); font-weight: 500; font-size: 13px; background: var(--bg-page); }
  table.grid tbody tr { transition: background .15s; }
  table.grid tbody tr:hover { background: var(--bg-container-hover); }
  table.grid code { background: var(--bg-page); padding: 2px 6px; border-radius: 4px; font-size: 12px; color: var(--text-2); }
  .op-btns { display: flex; gap: 8px; }

  /* ===== 文件网格 ===== */
  .path-bar {
    background: var(--bg-input); border: 1px solid var(--border); border-radius: 6px;
    padding: 10px 14px; margin-bottom: 16px; font-family: ui-monospace, monospace;
    font-size: 13px; color: var(--text-2); word-break: break-all;
  }
  .file-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(150px, 1fr)); gap: 14px; }
  .file-item {
    background: var(--bg-input); border: 1px solid var(--border); border-radius: 8px;
    padding: 16px; cursor: pointer; transition: border-color .15s, transform .15s; text-align: center;
  }
  .file-item:hover { border-color: var(--brand); transform: translateY(-2px); }
  .file-item .icon { font-size: 34px; margin-bottom: 8px; }
  .file-item .name { font-size: 13px; word-break: break-all; line-height: 1.4; }
  .file-item .meta { font-size: 11px; color: var(--text-3); margin-top: 4px; }
  .empty { color: var(--text-3); text-align: center; padding: 48px 0; }

  /* ===== 弹窗（原生 modal 模拟 t-dialog 风格） ===== */
  .overlay {
    position: fixed; inset: 0; background: rgba(0,0,0,.55); z-index: 1000;
    display: flex; align-items: center; justify-content: center; padding: 20px;
  }
  .overlay.hidden { display: none; }
  .modal {
    background: var(--bg-container); border: 1px solid var(--border); border-radius: var(--radius);
    width: 440px; max-width: 100%; box-shadow: 0 6px 30px rgba(0,0,0,.4); overflow: hidden;
  }
  .modal-head { padding: 20px 22px 0; font-size: 16px; font-weight: 600; }
  .modal-body { padding: 20px 22px; }
  .modal-foot { padding: 0 22px 20px; display: flex; justify-content: flex-end; gap: 12px; }
  .form-row { margin-bottom: 16px; }
  .form-row > label { display: block; font-size: 13px; color: var(--text-2); margin-bottom: 6px; }
  .form-row .input { width: 100%; }
  .checkbox { display: flex; align-items: center; gap: 8px; cursor: pointer; color: var(--text-1); font-size: 14px; }
  .checkbox input { width: 16px; height: 16px; accent-color: var(--brand); }

  /* ===== 轻提示 toast ===== */
  #toast {
    position: fixed; bottom: 24px; left: 50%; transform: translateX(-50%);
    padding: 12px 22px; border-radius: 10px; font-weight: 600; z-index: 9999;
    color: #fff; background: var(--brand); opacity: 0; transition: opacity .25s; pointer-events: none;
  }

  @media (max-width: 640px) {
    .app-header { padding: 0 16px; }
    .container { padding: 0 12px; }
  }
</style>
</head>
<body>

<header class="app-header">
  <div class="brand"><span class="logo">📁</span><span>PhiloFTP 管理端</span></div>
  <span class="tag tag-success" id="uptime">运行中</span>
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
    <div class="path-bar" id="pathBar">/</div>
    <div class="file-grid" id="fileGrid"></div>
  </div>
</div>

<!-- 用户编辑弹窗 -->
<div class="overlay hidden" id="userDialog">
  <div class="modal">
    <div class="modal-head" id="userDialogTitle">新增用户</div>
    <div class="modal-body">
      <div class="form-row">
        <label>用户名</label>
        <input class="input" id="uName" placeholder="登录用户名">
      </div>
      <div class="form-row">
        <label>密码</label>
        <input class="input" id="uPass" type="password" placeholder="登录密码">
      </div>
      <div class="form-row">
        <label>根目录（相对 data 目录）</label>
        <input class="input" id="uHome" placeholder="例如 alice">
      </div>
      <div class="form-row">
        <label class="checkbox"><input type="checkbox" id="uRO"> 只读（不可上传/删除/建目录）</label>
      </div>
      <div class="form-row">
        <label class="checkbox"><input type="checkbox" id="uEnabled" checked> 启用该用户</label>
      </div>
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
    <div class="modal-body">
      <div class="form-row">
        <label>目录名</label>
        <input class="input" id="dirName" placeholder="新目录名称">
      </div>
    </div>
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
    const up = document.getElementById('uptime');
    if (up) up.textContent = '运行 ' + st.uptime;
  } catch (e) { notify('加载状态失败', 'error'); }
}

async function loadUsers() {
  const q = document.getElementById('userSearch').value;
  const users = await (await fetch('/api/users')).json();
  const list = q ? users.filter(u => u.username.includes(q)) : users;
  const tb = document.querySelector('#userTable tbody');
  tb.innerHTML = list.map(u => {
    const perm = u.read_only
      ? '<span class="tag tag-warning">只读</span>'
      : '<span class="tag tag-success">可写</span>';
    const status = u.enabled
      ? '<span class="tag tag-success">启用</span>'
      : '<span class="tag tag-default">停用</span>';
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
  document.getElementById('pathBar').textContent = '/' + currentUser + (currentPath ? '/' + currentPath : '');
  const grid = document.getElementById('fileGrid');
  const items = data.items || [];
  if (!items.length) { grid.innerHTML = '<div class="empty">空目录</div>'; return; }
  grid.innerHTML = items.map(it =>
    '<div class="file-item" ' + (it.is_dir ? 'ondblclick="enterDir(\'' + escapeAttr(it.name) + '\')"' : '') + '>' +
      '<div class="icon">' + (it.is_dir ? '📁' : '📄') + '</div>' +
      '<div class="name">' + escapeHtml(it.name) + '</div>' +
      '<div class="meta">' + (it.is_dir ? '目录' : formatSize(it.size)) + '</div>' +
    '</div>'
  ).join('');
}

function formatSize(b) {
  if (b < 1024) return b + ' B';
  if (b < 1048576) return (b / 1024).toFixed(1) + ' KB';
  return (b / 1048576).toFixed(1) + ' MB';
}

function enterDir(name) { currentPath = currentPath ? currentPath + '/' + name : name; loadFiles(); }

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
  if (res.ok) {
    notify('已保存');
    closeOverlay('userDialog');
    loadUsers(); loadFileUsers();
  } else notify('保存失败', 'error');
};

async function editUser(name) {
  const users = await (await fetch('/api/users')).json();
  const u = users.find(x => x.username === name);
  editing = name;
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

// 弹窗显隐（原生 overlay + 点击遮罩关闭）
function openOverlay(id) { document.getElementById(id).classList.remove('hidden'); }
function closeOverlay(id) { document.getElementById(id).classList.add('hidden'); }
document.querySelectorAll('.overlay').forEach(o => {
  o.addEventListener('click', (e) => { if (e.target === o) o.classList.add('hidden'); });
});

loadStats(); loadUsers(); loadFileUsers().then(loadFiles);
setInterval(() => { loadStats(); loadFiles(); }, 10000);
</script>
</body>
</html>`
