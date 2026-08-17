package handler

// LoginHTML 是登录页（深空控制台风：纯内联 CSS，离线可用，零外部依赖）
const LoginHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>登录 · PhiloFTP</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;700&family=Noto+Sans+SC:wght@400;500;700&display=swap" rel="stylesheet">
<style>
  :root{ color-scheme: dark; --line:#1e2a40; --haze:#8aa0c0; }
  *{ box-sizing:border-box; margin:0; padding:0; }
  body{ background:#070a12; color:#e6edf7; font-family:'Noto Sans SC',system-ui,sans-serif; min-height:100vh;
    display:flex; align-items:center; justify-content:center; padding:20px; position:relative; overflow:hidden; }
  .mono{ font-family:'JetBrains Mono','SFMono-Regular',Consolas,monospace; }
  .aurora{ position:fixed; inset:0; z-index:-1; overflow:hidden; }
  .aurora::before{ content:""; position:absolute; inset:-20%;
    background:
      radial-gradient(40% 50% at 15% 10%, rgba(34,211,238,.18), transparent 60%),
      radial-gradient(45% 55% at 85% 25%, rgba(56,189,248,.16), transparent 60%),
      radial-gradient(50% 60% at 50% 110%, rgba(52,211,153,.10), transparent 60%);
    filter:blur(10px); animation: drift 18s ease-in-out infinite alternate; }
  @keyframes drift{ from{transform:translate3d(-2%,-1%,0) scale(1.02);} to{transform:translate3d(2%,2%,0) scale(1.06);} }
  .glass{ background:rgba(13,19,32,.72); backdrop-filter:blur(16px); border:1px solid rgba(34,211,238,.2); }
  .login-card{ width:400px; max-width:100%; border-radius:20px; padding:36px; box-shadow:0 30px 80px -20px rgba(0,0,0,.8), 0 0 0 1px rgba(34,211,238,.1); }
  .brand{ text-align:center; margin-bottom:28px; }
  .brand .logo{ width:54px; height:54px; border-radius:14px; background:linear-gradient(135deg,#22d3ee,#38bdf8); color:#04121a;
    font-size:26px; font-weight:700; display:inline-flex; align-items:center; justify-content:center; margin-bottom:12px; }
  .brand h1{ font-size:20px; font-weight:600; } .brand p{ color:#6a7c98; font-size:13px; margin-top:6px; }
  .brand b{ color:#22d3ee; text-shadow:0 0 18px rgba(34,211,238,.5); }
  .field{ margin-bottom:16px; }
  .field label{ display:block; font-size:13px; color:var(--haze); margin-bottom:7px; }
  .input{ width:100%; padding:11px 14px; border-radius:10px; border:1px solid var(--line); background:rgba(7,10,18,.6);
    color:#e6edf7; font-size:14px; font-family:'JetBrains Mono','SFMono-Regular',Consolas,monospace; outline:none; transition:.15s; }
  .input:focus{ border-color:#22d3ee; box-shadow:0 0 0 3px rgba(34,211,238,.12); }
  .input-wrap{ position:relative; }
  .toggle{ position:absolute; right:12px; top:50%; transform:translateY(-50%); color:#6a7c98; cursor:pointer; font-size:13px; user-select:none; }
  .btn-primary{ width:100%; padding:12px; border:none; border-radius:10px; cursor:pointer; font-size:15px; letter-spacing:.04em;
    font-family:'JetBrains Mono','SFMono-Regular',Consolas,monospace; font-weight:700;
    background:linear-gradient(135deg,#22d3ee,#38bdf8); color:#04121a; transition:.18s; box-shadow:0 8px 26px -10px rgba(34,211,238,.7); }
  .btn-primary:hover{ filter:brightness(1.08); transform:translateY(-1px); }
  .btn-primary:disabled{ opacity:.55; cursor:not-allowed; transform:none; }
  .alert{ display:none; padding:10px 14px; border-radius:10px; font-size:13px; margin-bottom:16px;
    background:rgba(251,113,133,.12); color:#fb7185; border:1px solid rgba(251,113,133,.4); }
  .alert.show{ display:block; }
  .spin{ display:inline-block; width:15px; height:15px; border:2px solid rgba(4,18,26,.35); border-top-color:#04121a;
    border-radius:50%; animation:sp .7s linear infinite; vertical-align:-2px; margin-right:6px; }
  @keyframes sp{ to{transform:rotate(360deg);} }
  .reg-link{ text-align:center; margin-top:16px; font-size:13px; color:#6a7c98; }
  .reg-link a{ color:#22d3ee; text-decoration:none; }
  ::-webkit-scrollbar{width:10px;height:10px} ::-webkit-scrollbar-thumb{background:rgba(34,211,238,.2);border-radius:6px}
  ::-webkit-scrollbar-track{background:transparent}
</style>
</head>
<body>
<div class="aurora"></div>
<div class="login-card glass">
  <div class="brand">
    <div class="logo">P</div>
    <h1>Philo<b>FTP</b> 控制台</h1>
    <p>登录以管理文件与用户</p>
  </div>
  <div class="alert" id="alert"></div>
  <form id="loginForm" autocomplete="on">
    <div class="field">
      <label for="username">用户名</label>
      <input class="input mono" id="username" name="username" type="text" placeholder="输入用户名" required>
    </div>
    <div class="field">
      <label for="password">密码</label>
      <div class="input-wrap">
        <input class="input mono" id="password" name="password" type="password" placeholder="输入密码" required>
        <span class="toggle" id="togglePw">显示</span>
      </div>
    </div>
    <button class="btn-primary" id="submitBtn" type="submit">登 录</button>
  </form>
  <div class="reg-link" id="regLink"></div>
</div>

<script>
function showAlert(msg){const a=document.getElementById('alert');a.textContent=msg;a.classList.add('show');}
document.getElementById('togglePw').onclick=function(){const p=document.getElementById('password');const show=p.type==='password';p.type=show?'text':'password';this.textContent=show?'隐藏':'显示';};
fetch('/api/config/public').then(r=>r.json()).then(cfg=>{
  if(cfg.allow_register){document.getElementById('regLink').innerHTML='还没有账号？ <a href="/register">立即注册</a>';}
}).catch(()=>{});
document.getElementById('loginForm').addEventListener('submit',async function(e){
  e.preventDefault();
  const btn=document.getElementById('submitBtn');
  const u=document.getElementById('username').value.trim();
  const p=document.getElementById('password').value;
  if(!u||!p){showAlert('请输入用户名和密码');return;}
  btn.disabled=true;btn.innerHTML='<span class="spin"></span>登录中...';
  try{
    const res=await fetch('/api/login',{method:'POST',headers:{'Content-Type':'application/json'},credentials:'same-origin',body:JSON.stringify({username:u,password:p})});
    const data=await res.json().catch(()=>({}));
    if(res.ok){window.location.href='/';}
    else{showAlert(data.error||'登录失败');btn.disabled=false;btn.textContent='登 录';}
  }catch(err){showAlert('网络错误，请稍后重试');btn.disabled=false;btn.textContent='登 录';}
});
</script>
</body>
</html>`

// RegisterHTML 是注册页（深空控制台风，含表单验证与密码强度提示）
const RegisterHTML = `<!DOCTYPE html>
<html lang="zh-CN">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>注册 · PhiloFTP</title>
<link rel="preconnect" href="https://fonts.googleapis.com">
<link href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500;700&family=Noto+Sans+SC:wght@400;500;700&display=swap" rel="stylesheet">
<style>
  :root{ color-scheme: dark; --line:#1e2a40; --haze:#8aa0c0; }
  *{ box-sizing:border-box; margin:0; padding:0; }
  body{ background:#070a12; color:#e6edf7; font-family:'Noto Sans SC',system-ui,sans-serif; min-height:100vh;
    display:flex; align-items:center; justify-content:center; padding:20px; position:relative; overflow:hidden; }
  .mono{ font-family:'JetBrains Mono','SFMono-Regular',Consolas,monospace; }
  .aurora{ position:fixed; inset:0; z-index:-1; overflow:hidden; }
  .aurora::before{ content:""; position:absolute; inset:-20%;
    background:
      radial-gradient(40% 50% at 15% 10%, rgba(34,211,238,.18), transparent 60%),
      radial-gradient(45% 55% at 85% 25%, rgba(56,189,248,.16), transparent 60%),
      radial-gradient(50% 60% at 50% 110%, rgba(52,211,153,.10), transparent 60%);
    filter:blur(10px); animation: drift 18s ease-in-out infinite alternate; }
  @keyframes drift{ from{transform:translate3d(-2%,-1%,0) scale(1.02);} to{transform:translate3d(2%,2%,0) scale(1.06);} }
  .glass{ background:rgba(13,19,32,.72); backdrop-filter:blur(16px); border:1px solid rgba(34,211,238,.2); }
  .reg-card{ width:440px; max-width:100%; border-radius:20px; padding:34px; box-shadow:0 30px 80px -20px rgba(0,0,0,.8), 0 0 0 1px rgba(34,211,238,.1); }
  .brand{ text-align:center; margin-bottom:24px; }
  .brand .logo{ width:52px; height:52px; border-radius:14px; background:linear-gradient(135deg,#22d3ee,#38bdf8); color:#04121a;
    font-size:24px; font-weight:700; display:inline-flex; align-items:center; justify-content:center; margin-bottom:12px; }
  .brand h1{ font-size:20px; font-weight:600; } .brand p{ color:#6a7c98; font-size:13px; margin-top:6px; }
  .field{ margin-bottom:16px; }
  .field label{ display:block; font-size:13px; color:var(--haze); margin-bottom:7px; }
  .input{ width:100%; padding:11px 14px; border-radius:10px; border:1px solid var(--line); background:rgba(7,10,18,.6);
    color:#e6edf7; font-size:14px; font-family:'JetBrains Mono','SFMono-Regular',Consolas,monospace; outline:none; transition:.15s; }
  .input:focus{ border-color:#22d3ee; box-shadow:0 0 0 3px rgba(34,211,238,.12); }
  .input.invalid{ border-color:#fb7185; } .input.valid{ border-color:#34d399; }
  .input-wrap{ position:relative; }
  .toggle{ position:absolute; right:12px; top:50%; transform:translateY(-50%); color:#6a7c98; cursor:pointer; font-size:13px; user-select:none; }
  .hint{ font-size:12px; margin-top:6px; min-height:16px; color:#6a7c98; }
  .hint.error{ color:#fb7185; } .hint.ok{ color:#34d399; }
  .strength{ display:flex; gap:6px; margin-top:8px; }
  .strength .bar{ flex:1; height:5px; border-radius:3px; background:var(--line); transition:background .25s; }
  .strength .bar.lv1{ background:#fb7185; } .strength .bar.lv2{ background:#fbbf24; }
  .strength .bar.lv3{ background:#eab308; } .strength .bar.lv4{ background:#34d399; }
  .btn-primary{ width:100%; padding:12px; border:none; border-radius:10px; cursor:pointer; font-size:15px; letter-spacing:.04em;
    font-family:'JetBrains Mono','SFMono-Regular',Consolas,monospace; font-weight:700;
    background:linear-gradient(135deg,#22d3ee,#38bdf8); color:#04121a; transition:.18s; box-shadow:0 8px 26px -10px rgba(34,211,238,.7); }
  .btn-primary:hover{ filter:brightness(1.08); transform:translateY(-1px); }
  .btn-primary:disabled{ opacity:.55; cursor:not-allowed; transform:none; }
  .alert{ display:none; padding:10px 14px; border-radius:10px; font-size:13px; margin-bottom:16px;
    background:rgba(251,113,133,.12); color:#fb7185; border:1px solid rgba(251,113,133,.4); }
  .alert.show{ display:block; }
  .spin{ display:inline-block; width:15px; height:15px; border:2px solid rgba(4,18,26,.35); border-top-color:#04121a;
    border-radius:50%; animation:sp .7s linear infinite; vertical-align:-2px; margin-right:6px; }
  @keyframes sp{ to{transform:rotate(360deg);} }
  .reg-link{ text-align:center; margin-top:16px; font-size:13px; color:#6a7c98; }
  .reg-link a{ color:#22d3ee; text-decoration:none; }
  ::-webkit-scrollbar{width:10px;height:10px} ::-webkit-scrollbar-thumb{background:rgba(34,211,238,.2);border-radius:6px}
  ::-webkit-scrollbar-track{background:transparent}
</style>
</head>
<body>
<div class="aurora"></div>
<div class="reg-card glass">
  <div class="brand">
    <div class="logo">✎</div>
    <h1>创建账号</h1>
    <p>注册后将获得一个独立文件空间</p>
  </div>
  <div class="alert" id="alert"></div>
  <form id="regForm" autocomplete="off" novalidate>
    <div class="field">
      <label for="username">用户名</label>
      <input class="input mono" id="username" name="username" type="text" placeholder="3-32 位，仅字母/数字/_/-" maxlength="32">
      <div class="hint" id="usernameHint"></div>
    </div>
    <div class="field">
      <label for="password">密码</label>
      <div class="input-wrap">
        <input class="input mono" id="password" name="password" type="password" placeholder="至少 8 位，含大小写/数字/符号中 3 类">
        <span class="toggle" id="togglePw">显示</span>
      </div>
      <div class="strength" id="strengthBars"><div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div></div>
      <div class="hint" id="strengthText">密码强度：—</div>
    </div>
    <div class="field">
      <label for="confirm">确认密码</label>
      <input class="input mono" id="confirm" name="confirm" type="password" placeholder="再次输入密码">
      <div class="hint" id="confirmHint"></div>
    </div>
    <button class="btn-primary" id="submitBtn" type="submit" disabled>注 册</button>
  </form>
  <div class="reg-link">已有账号？ <a href="/login">返回登录</a></div>
</div>

<script>
function showAlert(msg){const a=document.getElementById('alert');a.textContent=msg;a.classList.add('show');}
function setHint(el,msg,type){el.textContent=msg||'';el.className='hint'+(type?' '+type:'');}
document.getElementById('togglePw').onclick=function(){const p=document.getElementById('password');const show=p.type==='password';p.type=show?'text':'password';this.textContent=show?'隐藏':'显示';};
function calcStrength(pw){if(!pw)return 0;let score=0;if(pw.length>=8)score++;if(pw.length>=12)score++;let lower=/[a-z]/.test(pw),upper=/[A-Z]/.test(pw),digit=/[0-9]/.test(pw),sym=/[^A-Za-z0-9]/.test(pw);if(lower&&upper)score++;if(digit)score++;if(sym)score++;return Math.min(score,4);}
const lvText=['—','弱','较弱','中等','强'];
document.getElementById('password').addEventListener('input',function(){const lv=calcStrength(this.value);const bars=document.querySelectorAll('#strengthBars .bar');bars.forEach((b,i)=>{b.className='bar'+(i<lv?' lv'+lv:'');});document.getElementById('strengthText').textContent='密码强度：'+lvText[lv];validate();});
document.getElementById('username').addEventListener('input',function(){const v=this.value.trim();const el=document.getElementById('usernameHint');if(!v){setHint(el,'');this.classList.remove('valid','invalid');}else if(v.length<3||v.length>32){setHint(el,'用户名长度需 3-32 位','error');this.classList.add('invalid');this.classList.remove('valid');}else if(!/^[a-zA-Z0-9_\-]+$/.test(v)){setHint(el,'仅可包含字母、数字、下划线和连字符','error');this.classList.add('invalid');this.classList.remove('valid');}else{setHint(el,'可用','ok');this.classList.add('valid');this.classList.remove('invalid');}validate();});
document.getElementById('confirm').addEventListener('input',function(){const v=this.value;const pw=document.getElementById('password').value;const el=document.getElementById('confirmHint');if(!v){setHint(el,'');this.classList.remove('valid','invalid');}else if(v!==pw){setHint(el,'两次输入的密码不一致','error');this.classList.add('invalid');this.classList.remove('valid');}else{setHint(el,'一致','ok');this.classList.add('valid');this.classList.remove('invalid');}validate();});
function formValid(){const u=document.getElementById('username').value.trim();const p=document.getElementById('password').value;const c=document.getElementById('confirm').value;const uOk=u.length>=3&&u.length<=32&&/^[a-zA-Z0-9_\-]+$/.test(u);const pOk=calcStrength(p)>=3&&p.length>=8;const cOk=c===p&&c.length>0;return uOk&&pOk&&cOk;}
function validate(){document.getElementById('submitBtn').disabled=!formValid();}
document.getElementById('regForm').addEventListener('submit',async function(e){
  e.preventDefault();
  if(!formValid()){showAlert('请检查表单填写是否完整正确');return;}
  const btn=document.getElementById('submitBtn');
  const payload={username:document.getElementById('username').value.trim(),password:document.getElementById('password').value,home:document.getElementById('username').value.trim()};
  btn.disabled=true;btn.innerHTML='<span class="spin"></span>注册中...';
  try{
    const res=await fetch('/api/register',{method:'POST',headers:{'Content-Type':'application/json'},credentials:'same-origin',body:JSON.stringify(payload)});
    const data=await res.json().catch(()=>({}));
    if(res.ok){window.location.href='/';}
    else{showAlert(data.error||'注册失败');btn.disabled=false;btn.textContent='注 册';}
  }catch(err){showAlert('网络错误，请稍后重试');btn.disabled=false;btn.textContent='注 册';}
});
</script>
</body>
</html>`
