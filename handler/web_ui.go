package handler

// DashboardHTML 是内置管理后台（SPA：左侧菜单 + 右侧视图切换）
var DashboardHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>PhiloFTP 管理后台</title>
<style>
  :root{
    --bg:#0b0e14; --panel:#161b27; --panel2:#1d2433; --border:#2a3346;
    --text:#e6ebf2; --muted:#8b97a8; --brand:#0052d9; --brand2:#366ef4;
    --success:#00a870; --danger:#d54941; --warn:#ed7b2f; --radius:10px;
  }
  *{box-sizing:border-box;}
  html,body{margin:0;height:100%;}
  body{
    font-family:-apple-system,BlinkMacSystemFont,"Segoe UI","PingFang SC","Microsoft YaHei",sans-serif;
    background:var(--bg); color:var(--text); display:flex; min-height:100vh;
  }
  /* 左侧菜单 */
  .sidebar{
    width:230px; flex-shrink:0; background:var(--panel); border-right:1px solid var(--border);
    display:flex; flex-direction:column; padding:18px 12px; position:sticky; top:0; height:100vh;
  }
  .brand{display:flex; align-items:center; gap:10px; padding:6px 10px 18px; font-size:18px; font-weight:700;}
  .brand .logo{width:30px;height:30px;border-radius:8px;background:linear-gradient(135deg,var(--brand),var(--brand2));
    display:flex;align-items:center;justify-content:center;font-size:16px;color:#fff;}
  .nav{display:flex; flex-direction:column; gap:4px; flex:1;}
  .nav-item{
    display:flex; align-items:center; gap:10px; padding:11px 13px; border-radius:8px;
    color:var(--muted); cursor:pointer; font-size:14px; user-select:none; transition:.15s;
  }
  .nav-item:hover{background:var(--panel2); color:var(--text);}
  .nav-item.active{background:var(--brand); color:#fff;}
  .nav-item .ico{width:18px; text-align:center; font-size:15px;}
  .sidebar-foot{margin-top:12px; padding-top:14px; border-top:1px solid var(--border);}
  .user-line{font-size:13px; color:var(--muted); padding:0 10px 10px; display:flex; justify-content:space-between; align-items:center;}
  .user-line b{color:var(--text);}
  .btn-logout{background:transparent;border:1px solid var(--border);color:var(--muted);
    padding:6px 12px;border-radius:8px;cursor:pointer;font-size:13px;}
  .btn-logout:hover{color:var(--text);border-color:var(--muted);}
  /* 右侧内容 */
  .main{flex:1; min-width:0; display:flex; flex-direction:column;}
  .topbar{height:56px;border-bottom:1px solid var(--border);display:flex;align-items:center;
    padding:0 24px;gap:12px;background:var(--panel);position:sticky;top:0;z-index:5;}
  .topbar h1{font-size:16px;margin:0;font-weight:600;}
  .topbar .crumb{color:var(--muted);font-size:13px;}
  .content{padding:24px; overflow:auto; flex:1;}
  .view{display:none; animation:fade .2s ease;}
  .view.active{display:block;}
  @keyframes fade{from{opacity:0;transform:translateY(6px);}to{opacity:1;transform:none;}}
  /* 卡片/网格 */
  .grid{display:grid; grid-template-columns:repeat(auto-fill,minmax(220px,1fr)); gap:16px; margin-top:18px;}
  .card{background:var(--panel);border:1px solid var(--border);border-radius:var(--radius);padding:18px;}
  .card .k{color:var(--muted);font-size:13px;margin-bottom:8px;}
  .card .v{font-size:24px;font-weight:700;}
  .stat-ico{font-size:20px;margin-right:8px;}
  .section-title{font-size:15px;font-weight:600;margin:4px 0 14px;display:flex;align-items:center;justify-content:space-between;}
  /* 表格 */
  table{width:100%;border-collapse:collapse;background:var(--panel);border-radius:var(--radius);overflow:hidden;}
  th,td{padding:12px 14px;text-align:left;font-size:14px;border-bottom:1px solid var(--border);}
  th{color:var(--muted);font-weight:500;background:var(--panel2);}
  tr:last-child td{border-bottom:none;}
  .tag{display:inline-block;padding:2px 9px;border-radius:20px;font-size:12px;}
  .tag.ok{background:rgba(0,168,112,.16);color:var(--success);}
  .tag.off{background:rgba(139,151,168,.16);color:var(--muted);}
  .tag.ro{background:rgba(237,123,47,.16);color:var(--warn);}
  /* 按钮 */
  .btn{background:var(--brand);color:#fff;border:none;padding:9px 16px;border-radius:8px;cursor:pointer;font-size:14px;}
  .btn:hover{background:var(--brand2);}
  .btn.ghost{background:transparent;border:1px solid var(--border);color:var(--text);}
  .btn.ghost:hover{border-color:var(--muted);}
  .btn.danger{background:var(--danger);}
  .btn.sm{padding:5px 11px;font-size:13px;}
  /* 表单 */
  .form-row{display:flex;flex-direction:column;gap:7px;margin-bottom:18px;max-width:420px;}
  .form-row label{font-size:13px;color:var(--muted);}
  .input,select.input,textarea.input{background:var(--panel2);border:1px solid var(--border);color:var(--text);
    padding:10px 12px;border-radius:8px;font-size:14px;width:100%;outline:none;}
  .input:focus{border-color:var(--brand);}
  .switch{display:inline-flex;align-items:center;gap:10px;cursor:pointer;}
  .switch input{width:18px;height:18px;}
  /* 弹层 */
  .overlay{position:fixed;inset:0;background:rgba(0,0,0,.55);display:none;align-items:center;justify-content:center;z-index:20;}
  .overlay.show{display:flex;}
  .modal{background:var(--panel);border:1px solid var(--border);border-radius:14px;padding:24px;width:440px;max-width:92vw;}
  .modal h3{margin:0 0 18px;font-size:17px;}
  .form-actions{display:flex;gap:10px;justify-content:flex-end;margin-top:8px;}
  /* 提示条 */
  .toast{position:fixed;top:18px;right:18px;background:var(--panel2);border:1px solid var(--border);
    padding:12px 18px;border-radius:10px;z-index:40;font-size:14px;opacity:0;transform:translateY(-8px);
    transition:.25s;max-width:340px;}
  .toast.show{opacity:1;transform:none;}
  .toast.ok{border-color:var(--success);}
  .toast.err{border-color:var(--danger);}
  .hint{color:var(--muted);font-size:12px;margin-top:6px;}
  .empty{color:var(--muted);text-align:center;padding:40px;font-size:14px;}
  /* 文件区 */
  .filebar{display:flex;gap:10px;align-items:center;margin-bottom:14px;flex-wrap:wrap;}
  .path-box{background:var(--panel2);border:1px solid var(--border);border-radius:8px;padding:8px 12px;font-size:13px;flex:1;min-width:200px;color:var(--muted);}
  .progress-mask{position:fixed;inset:0;background:rgba(0,0,0,.5);display:none;align-items:center;justify-content:center;z-index:50;}
  .progress-mask.show{display:flex;}
  .progress-box{background:var(--panel);border:1px solid var(--border);border-radius:12px;padding:22px;width:360px;}
  .progress-bar{height:10px;background:var(--panel2);border-radius:6px;overflow:hidden;margin:14px 0;}
  .progress-bar > i{display:block;height:100%;width:0;background:linear-gradient(90deg,var(--brand),var(--brand2));transition:width .15s;}
  .progress-text{font-size:13px;color:var(--muted);display:flex;justify-content:space-between;}
  .spin{display:inline-block;width:14px;height:14px;border:2px solid var(--muted);border-top-color:var(--brand);
    border-radius:50%;animation:sp .7s linear infinite;vertical-align:-2px;margin-right:6px;}
  @keyframes sp{to{transform:rotate(360deg);}}
  /* 响应式 */
  @media (max-width:760px){
    body{flex-direction:column;}
    .sidebar{width:100%;height:auto;flex-direction:row;flex-wrap:wrap;align-items:center;padding:10px;position:static;}
    .brand{padding:6px 10px;font-size:16px;}
    .nav{flex-direction:row;flex-wrap:wrap;gap:6px;flex:1;margin-top:0;}
    .nav-item{padding:8px 12px;font-size:13px;}
    .nav-item .label{display:inline;}
    .sidebar-foot{width:100%;margin-top:0;border-top:none;padding-top:0;}
    .user-line{width:100%;border-top:1px solid var(--border);padding-top:10px;}
    .topbar{padding:0 14px;}
    .content{padding:16px;}
    .grid{grid-template-columns:1fr 1fr;}
  }
  @media (max-width:480px){.grid{grid-template-columns:1fr;}}
</style>
</head>
<body>
  <aside class="sidebar">
    <div class="brand"><span class="logo">📁</span><span>PhiloFTP</span></div>
    <nav class="nav" id="nav">
      <div class="nav-item active" data-view="overview"><span class="ico">📊</span><span class="label">概览</span></div>
      <div class="nav-item" data-view="users"><span class="ico">👤</span><span class="label">用户管理</span></div>
      <div class="nav-item" data-view="files"><span class="ico">📂</span><span class="label">文件管理</span></div>
      <div class="nav-item" data-view="basic"><span class="ico">⚙️</span><span class="label">基础设置</span></div>
      <div class="nav-item" data-view="system"><span class="ico">🖥️</span><span class="label">系统配置</span></div>
    </nav>
    <div class="sidebar-foot">
      <div class="user-line"><span>当前用户 <b id="curUser">...</b></span></div>
      <button class="btn-logout" id="logoutBtn">退出登录</button>
    </div>
  </aside>

  <div class="main">
    <div class="topbar"><h1 id="viewTitle">概览</h1><span class="crumb" id="viewCrumb"></span></div>
    <div class="content">

      <!-- 概览 -->
      <section class="view active" id="view-overview">
        <div class="grid" id="ovGrid"></div>
      </section>

      <!-- 用户管理 -->
      <section class="view" id="view-users">
        <div class="section-title">
          <span>用户列表</span>
          <button class="btn" id="addUserBtn">+ 新增用户</button>
        </div>
        <table id="userTable">
          <thead><tr><th>用户名</th><th>主目录</th><th>状态</th><th>权限</th><th>操作</th></tr></thead>
          <tbody id="userBody"></tbody>
        </table>
        <div class="empty" id="userEmpty" style="display:none">暂无用户</div>
      </section>

      <!-- 文件管理 -->
      <section class="view" id="view-files">
        <div class="filebar">
          <div class="path-box" id="filePath">/</div>
          <button class="btn ghost sm" id="mkdirBtn">新建目录</button>
          <label class="btn sm" style="margin:0">上传<input type="file" id="uploadInput" multiple style="display:none"></label>
          <button class="btn ghost sm" id="refreshBtn">刷新</button>
        </div>
        <table id="fileTable">
          <thead><tr><th>名称</th><th>类型</th><th>大小</th><th>修改时间</th><th>操作</th></tr></thead>
          <tbody id="fileBody"></tbody>
        </table>
        <div class="empty" id="fileEmpty" style="display:none">目录为空</div>
      </section>

      <!-- 基础设置 -->
      <section class="view" id="view-basic">
        <div class="section-title"><span>基础信息设置</span><button class="btn" id="saveBasicBtn">保存配置</button></div>
        <div class="form-row"><label>FTP 控制端口</label><input class="input" id="cfgFtpPort" type="number"></div>
        <div class="form-row"><label>Web 管理端口</label><input class="input" id="cfgWebPort" type="number"></div>
        <div class="form-row"><label>PASV 被动端口范围（最小）</label><input class="input" id="cfgPasvMin" type="number"></div>
        <div class="form-row"><label>PASV 被动端口范围（最大）</label><input class="input" id="cfgPasvMax" type="number"></div>
        <div class="form-row"><label>数据根目录</label><input class="input" id="cfgDataDir" type="text"></div>
        <div class="form-row"><label class="switch"><input type="checkbox" id="cfgFtps"> 启用 FTPS（需配置证书）</label></div>
        <div class="form-row"><label>TLS 证书路径配置</label><input class="input" id="cfgTlsCert" type="text" placeholder="cert.pem"><div class="hint">证书与私钥路径（FTPS 启用时必填）</div></div>
        <div class="form-row"><input class="input" id="cfgTlsKey" type="text" placeholder="key.pem"></div>
        <div class="form-row"><label class="switch"><input type="checkbox" id="cfgRegister"> 允许 Web 端自助注册</label><div class="hint">关闭后将隐藏注册入口，仅管理员可创建账号</div></div>
        <div class="hint" style="margin-top:8px">注：保存后 FTP 服务（端口/PASV/FTPS）会即时热重载，数据目录、注册开关等实时生效；仅 Web 管理端口需重启服务进程后切换。</div>
      </section>

      <!-- 系统配置 -->
      <section class="view" id="view-system">
        <div class="section-title"><span>系统信息</span><button class="btn ghost sm" id="refreshSysBtn">刷新</button></div>
        <div class="grid" id="sysGrid"></div>
        <div class="section-title" style="margin-top:26px"><span>关于</span></div>
        <div class="card" style="max-width:520px;line-height:1.8;color:var(--muted);font-size:14px">
          <div><b style="color:var(--text)">PhiloFTP</b> —— 轻量级内网 FTP 服务器，内置 Web 管理后台。</div>
          <div>支持用户管理、文件浏览/上传/下载、实时进度与断点续传，零外部依赖，单二进制部署。</div>
        </div>
      </section>

    </div>
  </div>

  <!-- 用户弹层 -->
  <div class="overlay" id="userOverlay">
    <div class="modal">
      <h3 id="userModalTitle">新增用户</h3>
      <div class="form-row"><label>用户名（3-32 位，字母/数字/下划线/连字符）</label><input class="input" id="uUsername" type="text"></div>
      <div class="form-row"><label>密码</label><input class="input" id="uPassword" type="password"><div class="hint" id="pwHint"></div></div>
      <div class="form-row"><label>主目录（留空则默认与用户名相同）</label><input class="input" id="uHome" type="text" placeholder="例如 alice"></div>
      <label class="switch" style="margin-bottom:18px"><input type="checkbox" id="uReadOnly"> 只读用户（禁止上传/删除）</label>
      <div class="form-actions">
        <button class="btn ghost" id="userCancel">取消</button>
        <button class="btn" id="userSave">保存</button>
      </div>
    </div>
  </div>

  <!-- 新建目录弹层 -->
  <div class="overlay" id="mkdirOverlay">
    <div class="modal">
      <h3>新建目录</h3>
      <div class="form-row"><label>目录名称</label><input class="input" id="dirName" type="text"></div>
      <div class="form-actions">
        <button class="btn ghost" id="mkdirCancel">取消</button>
        <button class="btn" id="mkdirSave">创建</button>
      </div>
    </div>
  </div>

  <div class="toast" id="toast"></div>
  <div class="progress-mask" id="progressMask">
    <div class="progress-box">
      <div id="progressName">下载中…</div>
      <div class="progress-bar"><i id="progressFill"></i></div>
      <div class="progress-text"><span id="progressPct">0%</span><span id="progressSpeed"></span></div>
    </div>
  </div>

<script>
const A = (p,o)=>fetch(p,Object.assign({headers:{'Content-Type':'application/json'}},o)).then(r=>r.json().then(d=>({ok:r.ok,data:d})));
const G = p => A(p,{method:'GET'});
const QS = s => new URLSearchParams(s).toString();
function toast(msg, ok){const t=document.getElementById('toast');t.textContent=msg;t.className='toast show '+(ok?'ok':'err');setTimeout(()=>t.className='toast',2600);}
function esc(s){return (s||'').replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));}
function fmtSize(n){if(n==null)return '-';const u=['B','KB','MB','GB','TB'];let i=0;let v=n;while(v>=1024&&i<u.length-1){v/=1024;i++;}return v.toFixed(i?1:0)+' '+u[i];}
function showError(elId, msg){const el=document.getElementById(elId); if(el) el.innerHTML='<div class="empty" style="display:block">'+esc(msg)+'</div>';}

// 全局错误兜底，防止 API 抛错导致页面空白
window.onerror=(msg,src,line)=>{toast('JS 错误: '+msg+' (line '+line+')',false);};

// 登录守卫
(async()=>{
  try{
    const r=await G('/api/me');
    if(!r.ok){location.href='/login';return;}
    document.getElementById('curUser').textContent=r.data.username||'-';
    await loadOverview();
  }catch(e){console.error(e); showError('ovGrid','页面初始化失败: '+e.message);}
})();

// —— 菜单切换 ——
const titles={overview:'概览',users:'用户管理',files:'文件管理',basic:'基础设置',system:'系统配置'};
const crumbs={overview:'',users:'',files:'',basic:'修改端口/FTPS/注册等',system:'运行时信息'};
document.getElementById('nav').addEventListener('click',e=>{
  const it=e.target.closest('.nav-item'); if(!it) return;
  const v=it.dataset.view;
  document.querySelectorAll('.nav-item').forEach(n=>n.classList.toggle('active',n===it));
  document.querySelectorAll('.view').forEach(s=>s.classList.toggle('active',s.id==='view-'+v));
  document.getElementById('viewTitle').textContent=titles[v];
  document.getElementById('viewCrumb').textContent=crumbs[v]||'';
  if(v==='overview') loadOverview();
  if(v==='users') loadUsers();
  if(v==='files') loadFiles('/');
  if(v==='basic') loadBasic();
  if(v==='system') loadSystem();
});

// —— 概览 ——
async function loadOverview(){
  try{
    const [s,u]=await Promise.all([G('/api/status'),G('/api/users')]);
    if(!s.ok||!u.ok){showError('ovGrid','加载概览数据失败'); return;}
    const rows=[
      {k:'FTP 端口',v:s.data.ftp_port,ico:'🔌'},
      {k:'Web 端口',v:s.data.web_port,ico:'🌐'},
      {k:'用户总数',v:Array.isArray(u.data)?u.data.length:s.data.user_count,ico:'👥'},
      {k:'PASV 端口',v:s.data.pasv_ports,ico:'📡'},
      {k:'FTPS',v:s.data.ftps?'已启用':'未启用',ico:'🔒'},
      {k:'运行时长',v:s.data.uptime,ico:'⏱️'},
    ];
    document.getElementById('ovGrid').innerHTML=rows.map(r=>
      '<div class="card"><div class="k">'+r.ico+' '+r.k+'</div><div class="v">'+esc(String(r.v))+'</div></div>').join('');
  }catch(e){console.error(e); showError('ovGrid','加载概览失败: '+e.message);}
}

// —— 用户管理 ——
async function loadUsers(){
  try{
    const r=await G('/api/users'); if(!r.ok){showError('userBody','加载用户失败'); return;}
    const list=r.data; const body=document.getElementById('userBody');
    document.getElementById('userEmpty').style.display=list.length?'none':'block';
    body.innerHTML=list.map(u=>
      '<tr><td>'+esc(u.username)+'</td><td>'+esc(u.home||'-')+'</td>'+
      '<td><span class="tag '+(u.enabled?'ok':'off')+'">'+(u.enabled?'启用':'禁用')+'</span></td>'+
      '<td><span class="tag '+(u.read_only?'ro':'ok')+'">'+(u.read_only?'只读':'可读写')+'</span></td>'+
      '<td><button class="btn ghost sm" onclick="delUser(\''+esc(u.username)+'\')">删除</button></td></tr>').join('');
  }catch(e){console.error(e); showError('userBody','加载用户失败: '+e.message);}
}
function openUserModal(){document.getElementById('userModalTitle').textContent='新增用户';
  ['uUsername','uPassword','uHome'].forEach(id=>document.getElementById(id).value='');
  document.getElementById('uReadOnly').checked=false;document.getElementById('pwHint').textContent='';
  document.getElementById('userOverlay').classList.add('show');}
document.getElementById('addUserBtn').addEventListener('click',openUserModal);
document.getElementById('userCancel').addEventListener('click',()=>document.getElementById('userOverlay').classList.remove('show'));
document.getElementById('uPassword').addEventListener('input',e=>{
  const v=e.target.value;let n=0;if(v.length>=8)n++;if(/[a-z]/.test(v)&&/[A-Z]/.test(v))n++;
  if(/\d/.test(v))n++;if(/[^A-Za-z0-9]/.test(v))n++;
  const map=['很弱','弱','中等','强','很强'];document.getElementById('pwHint').textContent=v?'强度：'+map[n]:'';});
document.getElementById('userSave').addEventListener('click',async()=>{
  const u={username:document.getElementById('uUsername').value.trim(),
    password:document.getElementById('uPassword').value,
    home:document.getElementById('uHome').value.trim(),
    read_only:document.getElementById('uReadOnly').checked};
  if(u.username.length<3){toast('用户名至少 3 个字符',false);return;}
  if(u.password.length<6){toast('密码至少 6 个字符',false);return;}
  const r=await A('/api/users',{method:'POST',body:JSON.stringify(u)});
  if(r.ok){toast('用户已保存',true);document.getElementById('userOverlay').classList.remove('show');loadUsers();}
  else toast(r.data.error||'保存失败',false);
});
async function delUser(name){
  if(!confirm('确认删除用户 '+name+' ？'))return;
  const r=await A('/api/users/'+encodeURIComponent(name),{method:'DELETE'});
  if(r.ok){toast('已删除',true);loadUsers();}else toast(r.data.error||'删除失败',false);
}

// —— 文件管理 ——
let curPath='/';
async function loadFiles(path){
  try{
    curPath=path||'/'; document.getElementById('filePath').textContent=curPath;
    const r=await G('/api/files?path='+encodeURIComponent(curPath)); if(!r.ok){showError('fileBody','加载文件列表失败'); return;}
    const items=r.data.items||[]; const body=document.getElementById('fileBody');
    document.getElementById('fileEmpty').style.display=items.length?'none':'block';
    body.innerHTML=items.map(it=>{
      const type=it.is_dir?'文件夹':(it.name.split('.').pop().toUpperCase()||'文件');
      const openPath=curPath==='/'?('/'+it.name):(curPath+'/'+it.name);
      const act=it.is_dir
        ? '<a class="btn ghost sm" onclick="loadFiles(\''+openPath+'\')">打开</a>'
        : '<a class="btn ghost sm" onclick="downloadFile(\''+it.name+'\')">下载</a>';
      return '<tr><td>'+esc(it.name)+'</td><td>'+type+'</td><td>'+(it.is_dir?'-':fmtSize(it.size))+'</td><td>'+esc(it.mod_time||'')+'</td><td>'+act+'</td></tr>';
    }).join('');
  }catch(e){console.error(e); showError('fileBody','加载文件列表失败: '+e.message);}
}
document.getElementById('refreshBtn').addEventListener('click',()=>loadFiles(curPath));
document.getElementById('uploadInput').addEventListener('change',async e=>{
  const files=e.target.files; if(!files.length)return;
  const fd=new FormData();
  for(const f of files)fd.append('files',f);
  fd.append('path',curPath);
  const btn=e.target.previousElementSibling; const old=btn.textContent; btn.innerHTML='<span class="spin"></span>上传中';
  const r=await fetch('/api/upload/batch',{method:'POST',body:fd}).then(x=>x.json());
  btn.textContent=old; e.target.value='';
  if(r.ok){toast('上传成功（'+r.count+' 个文件）',true);loadFiles(curPath);}else toast(r.error||'上传失败',false);
});
document.getElementById('mkdirBtn').addEventListener('click',()=>document.getElementById('mkdirOverlay').classList.add('show'));
document.getElementById('mkdirCancel').addEventListener('click',()=>document.getElementById('mkdirOverlay').classList.remove('show'));
document.getElementById('mkdirSave').addEventListener('click',async()=>{
  const name=document.getElementById('dirName').value.trim(); if(!name){toast('名称不能为空',false);return;}
  const r=await A('/api/mkdir',{method:'POST',body:JSON.stringify({path:curPath,name})});
  if(r.ok){toast('已创建',true);document.getElementById('mkdirOverlay').classList.remove('show');document.getElementById('dirName').value='';loadFiles(curPath);}
  else toast(r.data.error||'创建失败',false);
});
async function downloadFile(name){
  const rel=curPath==='/'?('/'+name):(curPath+'/'+name);
  const mask=document.getElementById('progressMask'); mask.classList.add('show');
  const fill=document.getElementById('progressFill'); const pct=document.getElementById('progressPct');
  const spd=document.getElementById('progressSpeed'); document.getElementById('progressName').textContent='下载：'+name;
  fill.style.width='0%'; pct.textContent='0%'; spd.textContent='';
  try{
    const resp=await fetch('/api/download?path='+encodeURIComponent(rel));
    if(!resp.ok){toast('下载失败',false);mask.classList.remove('show');return;}
    const total=+resp.headers.get('Content-Length')||0;
    const reader=resp.body.getReader(); let rec=0; let last=rec; let t0=Date.now();
    const chunks=[];
    while(true){const {done,value}=await reader.read();if(done)break;chunks.push(value);rec+=value.length;
      const p=total?Math.floor(rec/total*100):0; fill.style.width=p+'%'; pct.textContent=p+'%';
      const dt=(Date.now()-t0)/1000; if(dt>0.5){spd.textContent=fmtSize((rec-last)/dt)+'/s'; last=rec; t0=Date.now();}}
    const blob=new Blob(chunks);
    const a=document.createElement('a'); a.href=URL.createObjectURL(blob); a.download=name; a.click();
    URL.revokeObjectURL(a.href); toast('下载完成',true);
  }catch(err){toast('下载出错',false);}
  setTimeout(()=>mask.classList.remove('show'),400);
}

// —— 基础设置 ——
async function loadBasic(){
  try{
    const r=await G('/api/config'); if(!r.ok){toast('加载配置失败',false);return;}
    const c=r.data;
    document.getElementById('cfgFtpPort').value=c.ftp_port||'';
    document.getElementById('cfgWebPort').value=c.web_port||'';
    document.getElementById('cfgPasvMin').value=c.pasv_min_port||'';
    document.getElementById('cfgPasvMax').value=c.pasv_max_port||'';
    document.getElementById('cfgDataDir').value=c.data_dir||'';
    document.getElementById('cfgFtps').checked=!!c.enable_ftps;
    document.getElementById('cfgTlsCert').value=c.tls_cert||'';
    document.getElementById('cfgTlsKey').value=c.tls_key||'';
    document.getElementById('cfgRegister').checked=c.allow_register!==false;
  }catch(e){console.error(e); toast('加载配置失败: '+e.message,false);}
}
document.getElementById('saveBasicBtn').addEventListener('click',async()=>{
  const body={
    ftp_port:+document.getElementById('cfgFtpPort').value,
    web_port:+document.getElementById('cfgWebPort').value,
    pasv_min_port:+document.getElementById('cfgPasvMin').value,
    pasv_max_port:+document.getElementById('cfgPasvMax').value,
    data_dir:document.getElementById('cfgDataDir').value.trim(),
    enable_ftps:document.getElementById('cfgFtps').checked,
    tls_cert:document.getElementById('cfgTlsCert').value.trim(),
    tls_key:document.getElementById('cfgTlsKey').value.trim(),
    allow_register:document.getElementById('cfgRegister').checked,
  };
  const r=await A('/api/config',{method:'PUT',body:JSON.stringify(body)});
  if(r.ok){
    toast(r.data.message||'已保存',true);
    // 保存成功后立即重新拉取最新配置与状态，刷新界面（无需刷新页面）
    try{ await loadBasic(); loadOverview(); }catch(e){ console.error(e); }
    // Web 端口变更需重启进程方可切换监听，提示用户
    if(r.data&&r.data.web_port_changed){
      toast('Web 端口已写入配置，请重启服务进程后于新端口访问',true);
    }
  } else {
    toast(r.data.error||'保存失败',false);
  }
});

// —— 系统配置 ——
async function loadSystem(){
  try{
    const r=await G('/api/system'); if(!r.ok){showError('sysGrid','加载系统信息失败'); return;}
    const d=r.data;
    const rows=[
      {k:'Go 版本',v:d.go_version},{k:'运行时长',v:d.uptime},{k:'协程数',v:d.goroutines},
      {k:'数据根目录',v:d.data_dir},{k:'配置文件',v:d.config_path},{k:'用户文件数',v:d.user_count},
    ];
    document.getElementById('sysGrid').innerHTML=rows.map(r=>
      '<div class="card"><div class="k">'+r.k+'</div><div class="v" style="font-size:16px;word-break:break-all">'+esc(String(r.v))+'</div></div>').join('');
  }catch(e){console.error(e); showError('sysGrid','加载系统信息失败: '+e.message);}
}
document.getElementById('refreshSysBtn').addEventListener('click',loadSystem);

// —— 退出 ——
document.getElementById('logoutBtn').addEventListener('click',async()=>{
  await A('/api/logout',{method:'POST'}); location.href='/login';
});
</script>
</body>
</html>`
