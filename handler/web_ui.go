package handler

// DashboardHTML 是内置的 Web 管理页面（纯前端单页，调用 /api/*）
const DashboardHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>PhiloFTP 管理端</title>
<style>
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body { font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif; background: #0f172a; color: #e2e8f0; min-height: 100vh; }
  .header { background: linear-gradient(90deg, #2563eb, #1d4ed8); padding: 20px 32px; display: flex; align-items: center; justify-content: space-between; box-shadow: 0 2px 12px rgba(0,0,0,.3); }
  .header h1 { font-size: 22px; font-weight: 700; letter-spacing: .5px; }
  .header .badge { background: rgba(255,255,255,.15); padding: 6px 14px; border-radius: 20px; font-size: 13px; }
  .container { max-width: 1180px; margin: 28px auto; padding: 0 20px; }
  .grid { display: grid; grid-template-columns: repeat(auto-fit, minmax(220px, 1fr)); gap: 16px; margin-bottom: 28px; }
  .card { background: #1e293b; border: 1px solid #334155; border-radius: 14px; padding: 20px; transition: transform .15s; }
  .card:hover { transform: translateY(-3px); border-color: #3b82f6; }
  .card .label { color: #94a3b8; font-size: 13px; margin-bottom: 8px; }
  .card .value { font-size: 24px; font-weight: 700; color: #60a5fa; }
  .section { background: #1e293b; border: 1px solid #334155; border-radius: 14px; padding: 22px; margin-bottom: 24px; }
  .section h2 { font-size: 17px; margin-bottom: 16px; color: #cbd5e1; border-left: 4px solid #3b82f6; padding-left: 10px; }
  .toolbar { display: flex; gap: 10px; flex-wrap: wrap; margin-bottom: 16px; align-items: center; }
  button { background: #2563eb; color: #fff; border: none; padding: 9px 16px; border-radius: 8px; cursor: pointer; font-size: 14px; transition: background .15s; }
  button:hover { background: #1d4ed8; }
  button.danger { background: #dc2626; }
  button.danger:hover { background: #b91c1c; }
  button.ghost { background: #334155; }
  button.ghost:hover { background: #475569; }
  input, select { background: #0f172a; border: 1px solid #334155; color: #e2e8f0; padding: 9px 12px; border-radius: 8px; font-size: 14px; outline: none; }
  input:focus, select:focus { border-color: #3b82f6; }
  table { width: 100%; border-collapse: collapse; font-size: 14px; }
  th, td { text-align: left; padding: 11px 12px; border-bottom: 1px solid #334155; }
  th { color: #94a3b8; font-weight: 600; font-size: 13px; }
  tr:hover td { background: #0f172a; }
  .tag { display: inline-block; padding: 2px 9px; border-radius: 10px; font-size: 12px; }
  .tag.ro { background: #7c2d12; color: #fdba74; }
  .tag.rw { background: #064e3b; color: #6ee7b7; }
  .tag.on { background: #1e3a8a; color: #93c5fd; }
  .tag.off { background: #475569; color: #cbd5e1; }
  .path-bar { background: #0f172a; border: 1px solid #334155; border-radius: 8px; padding: 10px 12px; margin-bottom: 14px; font-family: monospace; font-size: /13px; color: #94a3b8; word-break: break-all; }
  .file-grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(150px, 1fr)); gap: 12px; }
  .file-item { background: #0f172a; border: 1px solid #334155; border-radius: 10px; padding: 14px; cursor: pointer; transition: border-color .15s; }
  .file-item:hover { border-color: #3b82f6; }
  .file-item .icon { font-size: 30px; margin-bottom: 8px; }
  .file-item .name { font-size: 13px; word-break: break-all; }
  .file-item .meta { font-size: 11px; color: #64748b; margin-top: 4px; }
  .modal { position: fixed; inset: 0; background: rgba(0,0,0,.6); display: none; align-items: center; justify-content: center; z-index: 100; }
  .modal.open { display: flex; }
  .modal-box { background: #1e293b; border: 1px solid #334155; border-radius: 14px; padding: 24px; width: 420px; max-width: 92vw; }
  .modal-box h3 { margin-bottom: 16px; }
  .modal-box .row { display: flex; flex-direction: column; gap: 6px; margin-bottom: 12px; }
  .modal-box .actions { display: flex; justify-content: flex-end; gap: 10px; margin-top: 8px; }
  .toast { position: fixed; bottom: 24px; left: 50%; transform: translateX(-50%); background: #22c55e; color: #04210f; padding: 12px 22px; border-radius: 10px; font-weight: 600; opacity: 0; transition: opacity .3s; z-index: 200; }
  .toast.show { opacity: 1; }
  .toast.err { background: #ef4444; color: #fff; }
  .empty { color: #64748b; text-align: center; padding: 40px 0; }
  .hidden { display: none; }
</style>
</head>
<body>
<div class="header">
  <h1>📁 PhiloFTP 管理端</h1>
  <div class="badge" id="uptime">运行中</div>
</div>
<div class="container">
  <div class="grid" id="stats"></div>

  <div class="section">
    <h2>用户管理</h2>
    <div class="toolbar">
      <button id="addUserBtn">+ 新增用户</button>
      <input id="userSearch" placeholder="搜索用户名..." style="flex:1; min-width:160px;">
    </div>
    <table id="userTable">
      <thead><tr><th>用户名</th><th>根目录</th><th>权限</th><th>状态</th><th>操作</th></tr></thead>
      <tbody></tbody>
    </table>
  </div>

  <div class="section">
    <h2>文件管理</h2>
    <div class="toolbar">
      <select id="fileUser"></select>
      <button class="ghost" id="mkdirBtn">新建目录</button>
      <button class="ghost" id="uploadBtn">上传</button>
      <button class="ghost" id="batchUploadBtn">批量上传</button>
      <button class="ghost" id="refreshBtn">刷新</button>
    </div>
    <div class="path-bar" id="pathBar">/</div>
    <div class="file-grid" id="fileGrid"></div>
  </div>
</div>

<div class="modal" id="userModal">
  <div class="modal-box">
    <h3 id="userModalTitle">新增用户</h3>
    <div class="row"><label>用户名</label><input id="uName"></div>
    <div class="row"><label>密码</label><input id="uPass" type="password"></div>
    <div class="row"><label>根目录（相对 data 目录）</label><input id="uHome" placeholder="例如 alice"></div>
    <div class="row"><label><input type="checkbox" id="uRO"> 只读</label></div>
    <div class="row"><label><input type="checkbox" id="uEnabled" checked> 启用</label></div>
    <div class="actions">
      <button class="ghost" id="userCancel">取消</button>
      <button id="userSave">保存</button>
    </div>
  </div>
</div>

<div class="modal" id="mkdirModal">
  <div class="modal-box">
    <h3>新建目录</h3>
    <div class="row"><label>目录名</label><input id="dirName"></div>
    <div class="actions">
      <button class="ghost" id="mkdirCancel">取消</button>
      <button id="mkdirSave">创建</button>
    </div>
  </div>
</div>

<input type="file" id="uploadInput" class="hidden">
<input type="file" id="batchInput" class="hidden" multiple>

<div class="toast" id="toast"></div>

<script>
let currentPath = "";
let currentUser = "";

function toast(msg, isErr) {
  const t = document.getElementById('toast');
  t.textContent = msg;
  t.className = 'toast show' + (isErr ? ' err' : '');
  setTimeout(() => t.className = 'toast', 2200);
}

async function loadStats() {
  try {
    const [st, sys] = await Promise.all([fetch('/api/status').then(r=>r.json()), fetch('/api/system').then(r=>r.json())]);
    const cards = [
      {label:'FTP 端口', value: st.ftp_port},
      {label:'Web 端口', value: st.web_port},
      {label:'用户数', value: st.user_count},
      {label:'PASV 端口', value: st.pasv_ports},
      {label:'FTPS', value: st.ftps ? '启用' : '禁用'},
      {label:'运行时长', value: st.uptime},
      {label:'Go 版本', value: sys.go_version},
      {label:'Goroutines', value: sys.goroutines},
    ];
    document.getElementById('stats').innerHTML = cards.map(c=>
      '<div class="card"><div class="label">'+c.label+'</div><div class="value">'+c.value+'</div></div>'
    ).join('');
  } catch(e) { toast('加载状态失败', true); }
}

async function loadUsers() {
  const q = document.getElementById('userSearch').value;
  const res = await fetch('/api/users');
  let users = await res.json();
  if (q) users = users.filter(u => u.username.includes(q));
  const tb = document.querySelector('#userTable tbody');
  tb.innerHTML = users.map(u =>
    '<tr>'+
    '<td>'+u.username+'</td>'+
    '<td><code>'+u.home+'</code></td>'+
    '<td>'+(u.read_only ? '<span class="tag ro">只读</span>' : '<span class="tag rw">可写</span>')+'</td>'+
    '<td>'+(u.enabled ? '<span class="tag on">启用</span>' : '<span class="tag off">停用</span>')+'</td>'+
    '<td><button class="ghost" onclick="editUser(\''+u.username+'\')">编辑</button> '+
    '<button class="danger" onclick="delUser(\''+u.username+'\')">删除</button></td>'+
    '</tr>'
  ).join('');
}

async function loadFiles() {
  if (!currentUser) return;
  const res = await fetch('/api/files?user='+encodeURIComponent(currentUser)+'&path='+encodeURIComponent(currentPath));
  const data = await res.json();
  document.getElementById('pathBar').textContent = '/' + currentUser + (currentPath ? '/' + currentPath : '');
  const grid = document.getElementById('fileGrid');
  const items = data.items || [];
  if (!items.length) { grid.innerHTML = '<div class="empty">空目录</div>'; return; }
  grid.innerHTML = items.map(it =>
    '<div class="file-item" '+(it.is_dir ? 'ondblclick="enterDir(\''+it.name+'\')"' : '')+'>'+
    '<div class="icon">'+(it.is_dir ? '📁' : '📄')+'</div>'+
    '<div class="name">'+it.name+'</div>'+
    '<div class="meta">'+(it.is_dir ? '目录' : formatSize(it.size))+'</div>'+
    '</div>'
  ).join('');
}

function formatSize(b) {
  if (b < 1024) return b + ' B';
  if (b < 1048576) return (b/1024).toFixed(1) + ' KB';
  return (b/1048576).toFixed(1) + ' MB';
}

function enterDir(name) { currentPath = currentPath ? currentPath + '/' + name : name; loadFiles(); }

async function loadFileUsers() {
  const res = await fetch('/api/users');
  const users = await res.json();
  const sel = document.getElementById('fileUser');
  sel.innerHTML = users.map(u => '<option value="'+u.username+'">'+u.username+'</option>').join('');
  if (users.length) { currentUser = users[0].username; }
}

// 用户弹窗
let editing = null;
document.getElementById('addUserBtn').onclick = () => openUserModal(null);
document.getElementById('userCancel').onclick = () => document.getElementById('userModal').classList.remove('open');
document.getElementById('userSave').onclick = async () => {
  const u = {
    username: document.getElementById('uName').value,
    password: document.getElementById('uPass').value,
    home: document.getElementById('uHome').value || document.getElementById('uName').value,
    read_only: document.getElementById('uRO').checked,
    enabled: document.getElementById('uEnabled').checked,
  };
  if (!u.username || !u.password) { toast('用户名和密码必填', true); return; }
  const res = await fetch('/api/users', {method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify(u)});
  if (res.ok) { toast('已保存'); document.getElementById('userModal').classList.remove('open'); loadUsers(); loadFileUsers(); }
  else toast('保存失败', true);
};

async function editUser(name) {
  const users = await (await fetch('/api/users')).json();
  const u = users.find(x => x.username === name);
  editing = name;
  document.getElementById('userModalTitle').textContent = '编辑用户';
  document.getElementById('uName').value = u.username; document.getElementById('uName').disabled = true;
  document.getElementById('uPass').value = u.password;
  document.getElementById('uHome').value = u.home;
  document.getElementById('uRO').checked = u.read_only;
  document.getElementById('uEnabled').checked = u.enabled;
  document.getElementById('userModal').classList.add('open');
}
async function delUser(name) {
  if (!confirm('确认删除用户 '+name+'？')) return;
  const res = await fetch('/api/users/'+encodeURIComponent(name), {method:'DELETE'});
  if (res.ok) { toast('已删除'); loadUsers(); } else toast('删除失败', true);
}

// 文件操作
document.getElementById('refreshBtn').onclick = loadFiles;
document.getElementById('fileUser').onchange = (e) => { currentUser = e.target.value; currentPath=''; loadFiles(); };
document.getElementById('mkdirBtn').onclick = () => document.getElementById('mkdirModal').classList.add('open');
document.getElementById('mkdirCancel').onclick = () => document.getElementById('mkdirModal').classList.remove('open');
document.getElementById('mkdirSave').onclick = async () => {
  const name = document.getElementById('dirName').value;
  if (!name) return;
  const res = await fetch('/api/mkdir', {method:'POST', headers:{'Content-Type':'application/json'},
    body: JSON.stringify({user:currentUser, path:currentPath, name})});
  if (res.ok) { toast('已创建'); document.getElementById('mkdirModal').classList.remove('open'); loadFiles(); }
  else toast('创建失败', true);
};
document.getElementById('uploadBtn').onclick = () => document.getElementById('uploadInput').click();
document.getElementById('uploadInput').onchange = async (e) => {
  const file = e.target.files[0]; if (!file) return;
  const fd = new FormData(); fd.append('user', currentUser); fd.append('path', currentPath); fd.append('file', file);
  const res = await fetch('/api/upload', {method:'POST', body: fd});
  if (res.ok) { toast('已上传'); loadFiles(); } else toast('上传失败', true);
  e.target.value = '';
};
document.getElementById('batchUploadBtn').onclick = () => document.getElementById('batchInput').click();
document.getElementById('batchInput').onchange = async (e) => {
  const files = e.target.files; if (!files.length) return;
  const fd = new FormData(); fd.append('user', currentUser); fd.append('path', currentPath);
  for (const f of files) fd.append('files', f);
  const res = await fetch('/api/upload/batch', {method:'POST', body: fd});
  if (res.ok) { toast('批量上传完成'); loadFiles(); } else toast('上传失败', true);
  e.target.value = '';
};

document.getElementById('userSearch').oninput = loadUsers;

loadStats(); loadUsers(); loadFileUsers().then(loadFiles);
setInterval(() => { loadStats(); loadFiles(); }, 10000);
</script>
</body>
</html>`
