package handler

// DashboardHTML 是内置管理后台（SPA：左侧菜单 + 右侧视图切换）。
// 零外部依赖（纯内联 CSS，离线可用），深空暗色「控制台风」界面。
var DashboardHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>PhiloFTP 控制台</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;700&family=Noto+Sans+SC:wght@400;500;700&display=swap" rel="stylesheet">
<style>
  :root{ color-scheme: dark; --line:#1e2a40; --haze:#8aa0c0; }
  html,body{height:100%;}
  body{ background:#070a12; color:#e6edf7; font-family:'Noto Sans SC',system-ui,sans-serif; margin:0; }
  .mono{ font-family:'JetBrains Mono','SFMono-Regular',Consolas,monospace; }

  /* 背景辉光 + 网格 */
  .aurora{ position:fixed; inset:0; z-index:-2; overflow:hidden; }
  .aurora::before{ content:""; position:absolute; inset:-20%;
    background:
      radial-gradient(40% 50% at 12% 8%, rgba(34,211,238,.18), transparent 60%),
      radial-gradient(45% 55% at 90% 20%, rgba(56,189,248,.14), transparent 60%),
      radial-gradient(50% 60% at 50% 110%, rgba(52,211,153,.10), transparent 60%);
    filter: blur(10px); animation: drift 18s ease-in-out infinite alternate; }
  .aurora::after{ content:""; position:absolute; inset:0; opacity:.5;
    background-image:
      linear-gradient(rgba(34,211,238,.05) 1px, transparent 1px),
      linear-gradient(90deg, rgba(34,211,238,.05) 1px, transparent 1px);
    background-size: 44px 44px; mask-image: radial-gradient(circle at 50% 30%, black, transparent 80%); }
  @keyframes drift{ from{transform:translate3d(-2%,-1%,0) scale(1.02);} to{transform:translate3d(2%,2%,0) scale(1.06);} }
  .glass{ background:rgba(13,19,32,.72); backdrop-filter:blur(14px); border:1px solid rgba(30,42,64,.9); }

  .layout{ display:flex; min-height:100vh; }
  .side{ width:236px; border-right:1px solid var(--line); padding:20px 13px; position:sticky; top:0; height:100vh;
    display:flex; flex-direction:column; }
  .brand{ display:flex; align-items:center; position:relative; font-size:19px; font-weight:700; padding:0 2px 20px; }
  .brand .logo{ display:inline-flex; align-items:center; justify-content:center; width:34px; height:34px; border-radius:10px;
    background:linear-gradient(135deg,#22d3ee,#38bdf8); color:#04121a; font-size:18px; margin-right:11px; }
  .brand b{ color:#22d3ee; text-shadow:0 0 18px rgba(34,211,238,.5); }
  .nav{ display:flex; flex-direction:column; gap:4px; flex:1; }
  .nav-item{ position:relative; padding:12px 14px; border-radius:12px; cursor:pointer; font-size:14px; color:#cdd9ec;
    border:1px solid transparent; transition:.18s; display:flex; align-items:center; gap:12px; }
  .nav-item .ico{ width:20px; text-align:center; color:var(--haze); transition:.18s; }
  .nav-item:hover{ background:rgba(34,211,238,.06); border-color:rgba(34,211,238,.15); }
  .nav-item.active{ background:linear-gradient(90deg, rgba(34,211,238,.18), rgba(34,211,238,.02));
    border-color:rgba(34,211,238,.35); box-shadow: inset 2px 0 0 #22d3ee; }
  .nav-item.active .ico{ color:#22d3ee; }
  .side-foot{ margin-top:12px; padding-top:16px; border-top:1px solid var(--line); }
  .side-foot .row{ display:flex; align-items:center; justify-content:space-between; padding:0 8px 12px; font-size:14px; color:var(--haze); }

  .main{ flex:1; min-width:0; display:flex; flex-direction:column; }
  .topbar{ display:flex; align-items:center; gap:12px; padding:0 24px; height:56px; position:sticky; top:0; z-index:5;
    border-bottom:1px solid var(--line); }
  .topbar h1{ font-size:16px; font-weight:600; margin:0; }
  .topbar .crumb{ font-size:12px; color:#6a7c98; }
  .topbar .live{ margin-left:auto; display:flex; align-items:center; gap:8px; font-size:12px; color:#6a7c98; }
  .topbar .dot{ width:8px; height:8px; border-radius:50%; background:#34d399; box-shadow:0 0 8px #34d399; display:inline-block; }
  .content{ padding:24px; flex:1; overflow:auto; }

  .view{ display:none; } .view.active{ display:block; animation: rise .35s cubic-bezier(.2,.7,.3,1) both; }
  @keyframes rise{ from{opacity:0; transform:translateY(10px);} to{opacity:1; transform:none;} }
  .stagger > *{ animation: rise .4s cubic-bezier(.2,.7,.3,1) both; }
  .stagger > *:nth-child(1){animation-delay:.03s} .stagger > *:nth-child(2){animation-delay:.07s}
  .stagger > *:nth-child(3){animation-delay:.11s} .stagger > *:nth-child(4){animation-delay:.15s}
  .stagger > *:nth-child(5){animation-delay:.19s} .stagger > *:nth-child(6){animation-delay:.23s}

  .grid{ display:grid; gap:16px; }
  .grid-3{ grid-template-columns:repeat(auto-fill, minmax(200px,1fr)); }
  .grid-2{ grid-template-columns:1fr 1fr; }
  .grid-2 > div:first-child, .grid-2 > div:last-child{ min-width:0; }
  @media(max-width:760px){ .grid-2{ grid-template-columns:1fr; } }

  .card{ background:rgba(18,26,43,.6); border:1px solid var(--line); border-radius:14px; position:relative; overflow:hidden; transition:.2s; }
  .card::after{ content:""; position:absolute; inset:0; border-radius:14px; pointer-events:none;
    background:linear-gradient(120deg, rgba(34,211,238,.06), transparent 40%); opacity:0; transition:.2s; }
  .card:hover{ transform:translateY(-3px); border-color:rgba(34,211,238,.4); box-shadow:0 12px 36px -14px rgba(34,211,238,.4); }
  .card:hover::after{ opacity:1; }
  .card.pad{ padding:20px; }
  .card .lbl{ font-size:12px; color:var(--haze); margin-bottom:12px; display:flex; align-items:center; gap:8px; }
  .card .big{ font-size:30px; font-weight:700; font-family:'JetBrains Mono',monospace; text-shadow:0 0 18px rgba(34,211,238,.4); }

  .btn{ font-family:'JetBrains Mono',monospace; border-radius:10px; padding:9px 16px; font-size:13px; cursor:pointer;
    transition:.18s; border:1px solid transparent; letter-spacing:.02em; }
  .btn-primary{ background:linear-gradient(135deg,#22d3ee,#38bdf8); color:#04121a; font-weight:700; box-shadow:0 6px 22px -8px rgba(34,211,238,.7); }
  .btn-primary:hover{ filter:brightness(1.08); transform:translateY(-1px); }
  .btn-ghost{ background:rgba(255,255,255,.03); border-color:var(--line); color:#cdd9ec; }
  .btn-ghost:hover{ border-color:rgba(34,211,238,.5); color:#fff; background:rgba(34,211,238,.06); }
  .btn-danger{ background:rgba(251,113,133,.14); border-color:rgba(251,113,133,.4); color:#fb7185; }
  .btn-danger:hover{ background:rgba(251,113,133,.22); }
  .btn-sm{ padding:5px 11px; font-size:12px; }
  .sec-head{ display:flex; align-items:center; justify-content:space-between; margin-bottom:16px; }
  .sec-head h2{ font-size:18px; font-weight:600; margin:0; }
  .toolbar{ display:flex; align-items:center; gap:12px; margin-bottom:16px; flex-wrap:wrap; }
  .pathbar{ flex:1; min-width:200px; padding:10px 16px; border-radius:12px; background:rgba(7,10,18,.6); border:1px solid var(--line); color:var(--haze); font-size:13px; }

  .field{ display:flex; flex-direction:column; gap:6px; margin-bottom:16px; max-width:440px; }
  .field label{ font-size:13px; color:var(--haze); }
  .input{ background:rgba(7,10,18,.6); border:1px solid var(--line); color:#e6edf7; padding:10px 13px; border-radius:10px;
    font-size:14px; width:100%; outline:none; font-family:'JetBrains Mono',monospace; transition:.15s; box-sizing:border-box; }
  .input:focus{ border-color:#22d3ee; box-shadow:0 0 0 3px rgba(34,211,238,.12); }
  .hint{ color:#6a7c98; font-size:12px; margin-top:2px; }
  .check{ display:flex; align-items:center; gap:12px; margin-bottom:8px; cursor:pointer; font-size:14px; color:#cdd9ec; }

  table{ width:100%; border-collapse:collapse; }
  th,td{ text-align:left; padding:13px 16px; font-size:14px; border-bottom:1px solid rgba(30,42,64,.7); }
  th{ color:var(--haze); font-weight:500; font-size:12px; text-transform:uppercase; letter-spacing:.06em; }
  tbody tr{ transition:.12s; } tbody tr:hover{ background:rgba(34,211,238,.04); }

  .tag{ display:inline-block; padding:3px 11px; border-radius:999px; font-size:12px; font-family:'JetBrains Mono',monospace; }
  .tag-ok{ background:rgba(52,211,153,.14); color:#34d399; border:1px solid rgba(52,211,153,.3); }
  .tag-off{ background:rgba(138,160,192,.12); color:var(--haze); border:1px solid rgba(138,160,192,.25); }
  .tag-ro{ background:rgba(251,191,36,.12); color:#fbbf24; border:1px solid rgba(251,191,36,.3); }

  .overlay{ position:fixed; inset:0; background:rgba(3,6,12,.66); backdrop-filter:blur(4px); display:none;
    align-items:center; justify-content:center; z-index:30; }
  .overlay.show{ display:flex; animation: fade .2s ease; }
  @keyframes fade{ from{opacity:0} to{opacity:1} }
  .modal{ background:#0d1320; border:1px solid rgba(34,211,238,.25); border-radius:16px; padding:26px; width:460px; max-width:92vw;
    box-shadow:0 30px 80px -20px rgba(0,0,0,.8), 0 0 0 1px rgba(34,211,238,.1); }
  .modal h3{ font-size:18px; font-weight:600; margin:0 0 20px; }

  .toast{ position:fixed; top:20px; right:20px; background:#0d1320; border:1px solid var(--line); padding:13px 19px;
    border-radius:12px; z-index:50; font-size:14px; opacity:0; transform:translateY(-10px); transition:.3s; max-width:360px;
    box-shadow:0 14px 40px -12px rgba(0,0,0,.7); }
  .toast.show{ opacity:1; transform:none; } .toast.ok{ border-color:rgba(52,211,153,.5); } .toast.err{ border-color:rgba(251,113,133,.5); }

  .mask{ position:fixed; inset:0; background:rgba(3,6,12,.6); display:none; align-items:center; justify-content:center; z-index:40; }
  .mask.show{ display:flex; }
  .pbar{ height:9px; background:rgba(7,10,18,.8); border-radius:6px; overflow:hidden; margin:14px 0; }
  .pbar > i{ display:block; height:100%; width:0; background:linear-gradient(90deg,#22d3ee,#38bdf8); transition:width .15s; }

  .spin{ display:inline-block; width:14px; height:14px; border:2px solid #8aa0c0; border-top-color:#22d3ee; border-radius:50%; animation: sp .7s linear infinite; vertical-align:-2px; margin-right:6px; }
  @keyframes sp{ to{ transform:rotate(360deg); } }
  .empty{ color:#6a7c98; text-align:center; padding:44px; font-size:14px; }
  .about-line{ color:var(--haze); font-size:14px; line-height:1.7; }

  ::-webkit-scrollbar{ width:10px; height:10px; } ::-webkit-scrollbar-thumb{ background:rgba(34,211,238,.2); border-radius:6px; }
  ::-webkit-scrollbar-track{ background:transparent; }

  @media (max-width:760px){
    .side{ width:100% !important; height:auto; flex-direction:row; flex-wrap:wrap; align-items:center; padding:10px; }
    .brand{ font-size:16px; }
    .nav{ flex-direction:row; flex-wrap:wrap; gap:6px; flex:1; }
    .nav-item{ padding:8px 12px; font-size:13px; }
    .nav-item.active{ box-shadow:none; }
    .side-foot{ width:100%; margin:0; padding-top:10px; }
    .content{ padding:16px; }
  }
</style>
</head>
<body>
<div class="aurora"></div>
<div class="layout">
  <aside class="side glass">
    <div class="brand"><span class="logo">P</span><span>Philo<b>FTP</b></span></div>
    <nav class="nav" id="nav">
      <div class="nav-item active" data-view="overview"><span class="ico">▤</span><span>概览</span></div>
      <div class="nav-item" data-view="users" data-admin="1"><span class="ico">◍</span><span>用户管理</span></div>
      <div class="nav-item" data-view="files"><span class="ico">⊞</span><span>文件管理</span></div>
      <div class="nav-item" data-view="basic" data-admin="1"><span class="ico">⚙</span><span>基础设置</span></div>
      <div class="nav-item" data-view="system" data-admin="1"><span class="ico">⌘</span><span>系统配置</span></div>
    </nav>
    <div class="side-foot">
      <div class="row"><span>当前 <b style="color:#e6edf7" id="curUser">…</b></span></div>
      <button class="btn btn-ghost" style="width:100%" id="logoutBtn">退出登录</button>
    </div>
  </aside>

  <main class="main">
    <div class="topbar glass">
      <h1 id="viewTitle">概览</h1>
      <span class="crumb" id="viewCrumb"></span>
      <span class="live"><span class="dot"></span> 运行中</span>
    </div>
    <div class="content">
      <section class="view active stagger" id="view-overview">
        <div class="grid grid-3" id="ovGrid"></div>
      </section>

      <section class="view" id="view-users">
        <div class="sec-head"><h2>用户列表</h2><button class="btn btn-primary" id="addUserBtn">+ 新增用户</button></div>
        <div class="card"><table id="userTable"><thead><tr><th>用户名</th><th>主目录</th><th>状态</th><th>权限</th><th>操作</th></tr></thead><tbody id="userBody"></tbody></table></div>
        <div class="empty" id="userEmpty" style="display:none">暂无用户</div>
      </section>

      <section class="view" id="view-files">
        <div class="toolbar">
          <div class="pathbar mono" id="filePath">/</div>
          <button class="btn btn-ghost btn-sm" id="mkdirBtn">新建目录</button>
          <label class="btn btn-ghost btn-sm" id="uploadLabel" style="margin:0;cursor:pointer;">上传<input type="file" id="uploadInput" multiple style="display:none"></label>
          <button class="btn btn-ghost btn-sm" id="refreshBtn">刷新</button>
        </div>
        <div class="card"><table id="fileTable"><thead><tr><th>名称</th><th>类型</th><th>大小</th><th>修改时间</th><th>操作</th></tr></thead><tbody id="fileBody"></tbody></table></div>
        <div class="empty" id="fileEmpty" style="display:none">目录为空</div>
      </section>

      <section class="view" id="view-basic">
        <div class="sec-head"><h2>基础信息设置</h2><button class="btn btn-primary" id="saveBasicBtn">保存配置</button></div>
        <div class="grid grid-2">
          <div>
            <div class="field"><label>FTP 控制端口</label><input class="input" id="cfgFtpPort" type="number"></div>
            <div class="field"><label>Web 管理端口</label><input class="input" id="cfgWebPort" type="number"></div>
            <div class="field"><label>PASV 被动端口（最小）</label><input class="input" id="cfgPasvMin" type="number"></div>
            <div class="field"><label>PASV 被动端口（最大）</label><input class="input" id="cfgPasvMax" type="number"></div>
          </div>
          <div>
            <div class="field"><label>数据根目录</label><input class="input" id="cfgDataDir" type="text"></div>
            <label class="check"><input type="checkbox" id="cfgFtps"> 启用 FTPS（需配置证书）</label>
            <div class="field"><label>TLS 证书路径</label><input class="input" id="cfgTlsCert" type="text" placeholder="cert.pem"><div class="hint">FTPS 启用时必填</div></div>
            <div class="field"><label>TLS 私钥路径</label><input class="input" id="cfgTlsKey" type="text" placeholder="key.pem"></div>
            <label class="check"><input type="checkbox" id="cfgRegister"> 允许 Web 端自助注册</label>
            <div class="hint">关闭后隐藏注册入口，仅管理员可创建账号</div>
          </div>
        </div>
        <div class="hint" style="border-top:1px dashed var(--line); padding-top:12px; margin-top:12px;">
          提示：保存后 FTP 服务（端口/PASV/FTPS）会即时热重载，数据目录、注册开关等实时生效；仅 Web 管理端口需重启服务进程后切换。
        </div>
      </section>

      <section class="view" id="view-system">
        <div class="sec-head"><h2>系统信息</h2><button class="btn btn-ghost btn-sm" id="refreshSysBtn">刷新</button></div>
        <div class="grid grid-3" id="sysGrid"></div>
        <h2 style="font-size:18px;font-weight:600;margin:32px 0 12px;">关于</h2>
        <div class="card pad" style="max-width:680px;">
          <div class="about-line" style="margin-bottom:8px;"><b style="color:#e6edf7">PhiloFTP</b> —— 轻量级内网 FTP 服务器，内置 Web 控制台。</div>
          <div class="about-line">支持用户管理、文件浏览/上传/下载、实时进度与断点续传，零外部依赖，单二进制部署。</div>
        </div>
      </section>
    </div>
  </main>
</div>

<div class="overlay" id="userOverlay">
  <div class="modal">
    <h3 id="userModalTitle">新增用户</h3>
    <div class="field"><label>用户名（3-32 位，字母/数字/下划线/连字符）</label><input class="input" id="uUsername" type="text"></div>
    <div class="field"><label>密码</label><input class="input" id="uPassword" type="password"><div class="hint" id="pwHint"></div></div>
    <div class="field"><label>主目录（留空则默认与用户名相同）</label><input class="input" id="uHome" type="text" placeholder="例如 alice"></div>
    <div class="field"><label>角色</label>
      <select class="input" id="uRole">
        <option value="user">普通用户（仅文件上传/下载/删除/建目录）</option>
        <option value="admin">管理员（全部权限）</option>
      </select>
    </div>
    <div style="display:flex;justify-content:flex-end;gap:12px;margin-top:8px;">
      <button class="btn btn-ghost" id="userCancel">取消</button>
      <button class="btn btn-primary" id="userSave">保存</button>
    </div>
  </div>
</div>

<div class="overlay" id="mkdirOverlay">
  <div class="modal">
    <h3>新建目录</h3>
    <div class="field"><label>目录名称</label><input class="input" id="dirName" type="text"></div>
    <div style="display:flex;justify-content:flex-end;gap:12px;">
      <button class="btn btn-ghost" id="mkdirCancel">取消</button>
      <button class="btn btn-primary" id="mkdirSave">创建</button>
    </div>
  </div>
</div>

<div class="toast" id="toast"></div>
<div class="mask" id="progressMask">
  <div class="card pad" style="width:360px;">
    <div class="mono" id="progressName" style="font-size:13px;margin-bottom:4px;">下载中…</div>
    <div class="pbar"><i id="progressFill"></i></div>
    <div style="display:flex;justify-content:space-between;font-size:12px;" class="mono"><span id="progressPct" style="color:var(--haze)">0%</span><span id="progressSpeed" style="color:var(--haze)"></span></div>
  </div>
</div>

<script>
const A=(p,o)=>fetch(p,Object.assign({headers:{'Content-Type':'application/json'}},o)).then(r=>r.json().then(d=>({ok:r.ok,data:d})));
const G=p=>A(p,{method:'GET'});
function toast(msg,ok){const t=document.getElementById('toast');t.textContent=msg;t.className='toast show '+(ok?'ok':'err');setTimeout(()=>t.className='toast',2600);}
function esc(s){return (s==null?'':String(s)).replace(/[&<>"']/g,c=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]));}
function fmtSize(n){if(n==null)return '-';const u=['B','KB','MB','GB','TB'];let i=0,v=n;while(v>=1024&&i<u.length-1){v/=1024;i++;}return v.toFixed(i?1:0)+' '+u[i];}
function showError(elId,msg){const el=document.getElementById(elId);if(el)el.innerHTML='<div class="empty" style="display:block">'+esc(msg)+'</div>';}
window.onerror=(msg,src,line)=>{toast('JS 错误: '+msg+' (line '+line+')',false);};

(async()=>{
  try{
    const r=await G('/api/me');
    if(!r.ok){location.href='/login';return;}
    document.getElementById('curUser').textContent=r.data.username||'-';
    // 角色分级：非管理员隐藏管理类入口，仅保留文件相关功能
    window.__isAdmin = !!r.data.is_admin;
    if(!window.__isAdmin){
      document.querySelectorAll('[data-admin="1"]').forEach(el=>el.style.display='none');
      const addBtn=document.getElementById('addUserBtn'); if(addBtn)addBtn.style.display='none';
    }
    await loadOverview();
  }catch(e){console.error(e);showError('ovGrid','页面初始化失败: '+e.message);}
})();

const titles={overview:'概览',users:'用户管理',files:'文件管理',basic:'基础设置',system:'系统配置'};
const crumbs={overview:'',users:'',files:'',basic:'修改端口 / FTPS / 注册等',system:'运行时信息'};
document.getElementById('nav').addEventListener('click',e=>{
  const it=e.target.closest('.nav-item'); if(!it) return;
  const v=it.dataset.view;
  if(it.dataset.admin==='1' && !window.__isAdmin){ toast('权限不足：需要管理员角色',false); return; }
  document.querySelectorAll('.nav-item').forEach(n=>n.classList.toggle('active',n===it));
  document.querySelectorAll('.view').forEach(s=>s.classList.toggle('active',s.id==='view-'+v));
  document.getElementById('viewTitle').textContent=titles[v];
  document.getElementById('viewCrumb').textContent=crumbs[v]||'';
  if(v==='overview')loadOverview();
  if(v==='users')loadUsers();
  if(v==='files')loadFiles('/');
  if(v==='basic')loadBasic();
  if(v==='system')loadSystem();
});

async function loadOverview(){
  try{
    const s=await G('/api/status'); if(!s.ok){showError('ovGrid','加载概览数据失败');return;}
    // 用户总数仅管理员可见
    let userCount='-';
    if(window.__isAdmin){ const u=await G('/api/users'); if(u.ok&&Array.isArray(u.data))userCount=u.data.length; }
    const rows=[
      {k:'FTP 端口',v:s.data.ftp_port,ico:'⚡'},{k:'Web 端口',v:s.data.web_port,ico:'🌐'},
      {k:'用户总数',v:userCount,ico:'◍'},
      {k:'PASV 端口',v:s.data.pasv_ports,ico:'📡'},{k:'FTPS',v:s.data.ftps?'已启用':'未启用',ico:'🔒'},
      {k:'运行时长',v:s.data.uptime,ico:'⏱'},
    ];
    document.getElementById('ovGrid').innerHTML=rows.map(r=>
      '<div class="card pad"><div class="lbl"><span style="font-size:18px">'+r.ico+'</span>'+esc(r.k)+'</div><div class="big">'+esc(String(r.v))+'</div></div>').join('');
  }catch(e){console.error(e);showError('ovGrid','加载概览失败: '+e.message);}
}

async function loadUsers(){
  try{
    const r=await G('/api/users'); if(!r.ok){showError('userBody','加载用户失败');return;}
    const list=r.data, body=document.getElementById('userBody');
    document.getElementById('userEmpty').style.display=list.length?'none':'block';
    body.innerHTML=list.map(u=>
      '<tr><td style="font-weight:500">'+esc(u.username)+'</td><td class="mono" style="color:var(--haze)">'+esc(u.home||'-')+'</td>'+
      '<td><span class="tag '+(u.enabled?'tag-ok':'tag-off')+'">'+(u.enabled?'启用':'禁用')+'</span></td>'+
      '<td><span class="tag '+(u.role==='admin'?'tag-adm':'tag-ok')+'">'+(u.role==='admin'?'管理员':'普通用户')+'</span></td>'+
      '<td><button class="btn btn-ghost btn-sm" onclick="delUser(\''+esc(u.username)+'\')">删除</button></td></tr>').join('');
  }catch(e){console.error(e);showError('userBody','加载用户失败: '+e.message);}
}
function openUserModal(){document.getElementById('userModalTitle').textContent='新增用户';
  ['uUsername','uPassword','uHome'].forEach(id=>document.getElementById(id).value='');
  document.getElementById('uRole').value='user';document.getElementById('pwHint').textContent='';
  document.getElementById('userOverlay').classList.add('show');}
document.getElementById('addUserBtn').addEventListener('click',openUserModal);
document.getElementById('userCancel').addEventListener('click',()=>document.getElementById('userOverlay').classList.remove('show'));
document.getElementById('uPassword').addEventListener('input',e=>{
  const v=e.target.value;let n=0;if(v.length>=8)n++;if(/[a-z]/.test(v)&&/[A-Z]/.test(v))n++;if(/\d/.test(v))n++;if(/[^A-Za-z0-9]/.test(v))n++;
  const map=['很弱','弱','中等','强','很强'];document.getElementById('pwHint').textContent=v?'强度：'+map[n]:'';});
document.getElementById('userSave').addEventListener('click',async()=>{
  const u={username:document.getElementById('uUsername').value.trim(),password:document.getElementById('uPassword').value,
    home:document.getElementById('uHome').value.trim(),role:document.getElementById('uRole').value};
  if(u.username.length<3){toast('用户名至少 3 个字符',false);return;}
  if(u.password.length<6){toast('密码至少 6 个字符',false);return;}
  const r=await A('/api/users',{method:'POST',body:JSON.stringify(u)});
  if(r.ok){toast('用户已保存',true);document.getElementById('userOverlay').classList.remove('show');loadUsers();}
  else toast(r.data.error||'保存失败',false);
});
async function delUser(name){if(!confirm('确认删除用户 '+name+' ？'))return;
  const r=await A('/api/users/'+encodeURIComponent(name),{method:'DELETE'});
  if(r.ok){toast('已删除',true);loadUsers();}else toast(r.data.error||'删除失败',false);}

let curPath='/';
async function loadFiles(path){
  try{
    curPath=path||'/';document.getElementById('filePath').textContent=curPath;
    const r=await G('/api/files?path='+encodeURIComponent(curPath)); if(!r.ok){showError('fileBody','加载文件列表失败');return;}
    const items=r.data.items||[], body=document.getElementById('fileBody');
    document.getElementById('fileEmpty').style.display=items.length?'none':'block';
    body.innerHTML=items.map(it=>{
      const type=it.is_dir?'文件夹':(it.name.split('.').pop().toUpperCase()||'文件');
      const openPath=curPath==='/'?('/'+it.name):(curPath+'/'+it.name);
      const act=it.is_dir?'<a class="btn btn-ghost btn-sm" onclick="loadFiles(\''+openPath+'\')">打开</a>'
        :'<a class="btn btn-ghost btn-sm" onclick="downloadFile(\''+it.name+'\')">下载</a>';
      return '<tr><td style="font-weight:500">'+esc(it.name)+'</td><td style="color:var(--haze)">'+type+'</td>'+
        '<td class="mono" style="color:var(--haze)">'+(it.is_dir?'-':fmtSize(it.size))+'</td>'+
        '<td class="mono" style="color:var(--haze);font-size:13px">'+esc(it.mod_time||'')+'</td><td>'+act+'</td></tr>';
    }).join('');
  }catch(e){console.error(e);showError('fileBody','加载文件列表失败: '+e.message);}
}
document.getElementById('refreshBtn').addEventListener('click',()=>loadFiles(curPath));
document.getElementById('uploadInput').addEventListener('change',async e=>{
  const files=e.target.files; if(!files.length)return;
  const fd=new FormData(); for(const f of files)fd.append('files',f); fd.append('path',curPath);
  const btn=document.getElementById('uploadLabel'); const old=btn.textContent; btn.innerHTML='<span class="spin"></span>上传中';
  let r; try{ r=await fetch('/api/upload/batch',{method:'POST',body:fd}).then(x=>x.json()); }catch(err){ r={}; }
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
  const fill=document.getElementById('progressFill'), pct=document.getElementById('progressPct'), spd=document.getElementById('progressSpeed');
  document.getElementById('progressName').textContent='下载：'+name;
  fill.style.width='0%'; pct.textContent='0%'; spd.textContent='';
  try{
    const resp=await fetch('/api/download?path='+encodeURIComponent(rel));
    if(!resp.ok){toast('下载失败',false);mask.classList.remove('show');return;}
    const total=+resp.headers.get('Content-Length')||0; const reader=resp.body.getReader();
    let rec=0,last=0,t0=Date.now(); const chunks=[];
    while(true){const {done,value}=await reader.read();if(done)break;chunks.push(value);rec+=value.length;
      const p=total?Math.floor(rec/total*100):0; fill.style.width=p+'%'; pct.textContent=p+'%';
      const dt=(Date.now()-t0)/1000; if(dt>0.5){spd.textContent=fmtSize((rec-last)/dt)+'/s'; last=rec; t0=Date.now();}}
    const a=document.createElement('a'); a.href=URL.createObjectURL(new Blob(chunks)); a.download=name; a.click(); URL.revokeObjectURL(a.href);
    toast('下载完成',true);
  }catch(err){toast('下载出错',false);}
  setTimeout(()=>mask.classList.remove('show'),400);
}

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
  }catch(e){console.error(e);toast('加载配置失败: '+e.message,false);}
}
document.getElementById('saveBasicBtn').addEventListener('click',async()=>{
  const body={
    ftp_port:+document.getElementById('cfgFtpPort').value, web_port:+document.getElementById('cfgWebPort').value,
    pasv_min_port:+document.getElementById('cfgPasvMin').value, pasv_max_port:+document.getElementById('cfgPasvMax').value,
    data_dir:document.getElementById('cfgDataDir').value.trim(), enable_ftps:document.getElementById('cfgFtps').checked,
    tls_cert:document.getElementById('cfgTlsCert').value.trim(), tls_key:document.getElementById('cfgTlsKey').value.trim(),
    allow_register:document.getElementById('cfgRegister').checked,
  };
  const r=await A('/api/config',{method:'PUT',body:JSON.stringify(body)});
  if(r.ok){toast(r.data.message||'已保存',true);
    try{await loadBasic();loadOverview();}catch(e){console.error(e);}
    if(r.data&&r.data.web_port_changed){toast('Web 端口已写入配置，请重启服务进程后于新端口访问',true);}
  }else toast(r.data.error||'保存失败',false);
});

async function loadSystem(){
  try{
    const r=await G('/api/system'); if(!r.ok){showError('sysGrid','加载系统信息失败');return;}
    const d=r.data;
    const rows=[{k:'Go 版本',v:d.go_version},{k:'运行时长',v:d.uptime},{k:'协程数',v:d.goroutines},
      {k:'数据根目录',v:d.data_dir},{k:'配置文件',v:d.config_path},{k:'用户文件数',v:d.user_count}];
    document.getElementById('sysGrid').innerHTML=rows.map(r=>
      '<div class="card pad"><div class="lbl">'+esc(r.k)+'</div><div style="font-size:14px" class="mono">'+esc(String(r.v))+'</div></div>').join('');
  }catch(e){console.error(e);showError('sysGrid','加载系统信息失败: '+e.message);}
}
document.getElementById('refreshSysBtn').addEventListener('click',loadSystem);
document.getElementById('logoutBtn').addEventListener('click',async()=>{await A('/api/logout',{method:'POST'});location.href='/login';});
</script>
</body>
</html>`
