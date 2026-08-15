package handler

// LoginHTML 是登录页（零外部依赖，TDesign 暗色风格）
const LoginHTML = `<!DOCTYPE html>
<html lang="zh-CN" data-theme="dark">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>登录 · PhiloFTP</title>
<style>
  :root {
    --brand: #0052d9; --brand-hover: #266fe8; --brand-active: #003cab;
    --success: #2ba471; --warning: #e37318; --danger: #d54941;
    --bg-page: #181818; --bg-card: #232324; --bg-input: #1a1a1a;
    --border: #424244; --text-1: rgba(255,255,255,.9); --text-2: rgba(255,255,255,.6); --text-3: rgba(255,255,255,.4);
    --radius: 12px;
  }
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif;
    background: radial-gradient(1200px 600px at 50% -10%, rgba(0,82,217,.18), transparent 60%), var(--bg-page);
    color: var(--text-1); min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 20px;
  }
  .login-card {
    width: 400px; max-width: 100%; background: var(--bg-card); border: 1px solid var(--border);
    border-radius: var(--radius); padding: 36px 34px; box-shadow: 0 12px 48px rgba(0,0,0,.45);
  }
  .brand { text-align: center; margin-bottom: 26px; }
  .brand .logo { font-size: 44px; }
  .brand h1 { font-size: 20px; font-weight: 600; margin-top: 8px; }
  .brand p { color: var(--text-3); font-size: 13px; margin-top: 4px; }
  .field { margin-bottom: 18px; }
  .field label { display: block; font-size: 13px; color: var(--text-2); margin-bottom: 7px; }
  .input {
    width: 100%; padding: 11px 14px; border-radius: 8px; border: 1px solid var(--border);
    background: var(--bg-input); color: var(--text-1); font-size: 14px; font-family: inherit; outline: none;
    transition: border-color .2s;
  }
  .input:focus { border-color: var(--brand); }
  .input-wrap { position: relative; }
  .input-wrap .toggle {
    position: absolute; right: 12px; top: 50%; transform: translateY(-50%); color: var(--text-3);
    cursor: pointer; font-size: 13px; user-select: none;
  }
  .btn-primary {
    width: 100%; padding: 12px; border: none; border-radius: 8px; background: var(--brand); color: #fff;
    font-size: 15px; cursor: pointer; font-family: inherit; transition: background .2s; margin-top: 4px;
  }
  .btn-primary:hover { background: var(--brand-hover); }
  .btn-primary:active { background: var(--brand-active); }
  .row-between { display: flex; justify-content: space-between; align-items: center; margin-top: 14px; font-size: 13px; }
  .row-between a { color: var(--brand); text-decoration: none; }
  .row-between a:hover { text-decoration: underline; }
  .alert {
    display: none; padding: 10px 14px; border-radius: 8px; font-size: 13px; margin-bottom: 16px;
    background: rgba(213,73,65,.12); color: var(--danger); border: 1px solid rgba(213,73,65,.4);
  }
  .alert.show { display: block; }
  .spinner {
    display: inline-block; width: 15px; height: 15px; border: 2px solid rgba(255,255,255,.35);
    border-top-color: #fff; border-radius: 50%; animation: spin .7s linear infinite; vertical-align: -2px; margin-right: 6px;
  }
  @keyframes spin { to { transform: rotate(360deg); } }
</style>
</head>
<body>
  <div class="login-card">
    <div class="brand">
      <div class="logo">📁</div>
      <h1>PhiloFTP 管理端</h1>
      <p>请登录以管理您的文件</p>
    </div>
    <div class="alert" id="alert"></div>
    <form id="loginForm" autocomplete="on">
      <div class="field">
        <label for="username">用户名</label>
        <input class="input" id="username" name="username" type="text" placeholder="输入用户名" required>
      </div>
      <div class="field">
        <label for="password">密码</label>
        <div class="input-wrap">
          <input class="input" id="password" name="password" type="password" placeholder="输入密码" required>
          <span class="toggle" id="togglePw">显示</span>
        </div>
      </div>
      <button class="btn-primary" id="submitBtn" type="submit">登 录</button>
    </form>
    <div class="row-between" id="regLink"></div>
  </div>

<script>
function showAlert(msg) {
  const a = document.getElementById('alert');
  a.textContent = msg; a.classList.add('show');
}
document.getElementById('togglePw').onclick = function() {
  const p = document.getElementById('password');
  const show = p.type === 'password';
  p.type = show ? 'text' : 'password';
  this.textContent = show ? '隐藏' : '显示';
};
// 是否显示注册入口
fetch('/api/config/public').then(r => r.json()).then(cfg => {
  if (cfg.allow_register) {
    document.getElementById('regLink').innerHTML = '<span style="color:var(--text-3)">还没有账号？</span><a href="/register">立即注册</a>';
  }
}).catch(() => {});

document.getElementById('loginForm').addEventListener('submit', async function(e) {
  e.preventDefault();
  const btn = document.getElementById('submitBtn');
  const u = document.getElementById('username').value.trim();
  const p = document.getElementById('password').value;
  if (!u || !p) { showAlert('请输入用户名和密码'); return; }
  btn.disabled = true; btn.innerHTML = '<span class="spinner"></span>登录中...';
  try {
    const res = await fetch('/api/login', {
      method: 'POST', headers: {'Content-Type': 'application/json'},
      credentials: 'same-origin',
      body: JSON.stringify({ username: u, password: p })
    });
    const data = await res.json().catch(() => ({}));
    if (res.ok) {
      window.location.href = '/';
    } else {
      showAlert(data.error || '登录失败');
      btn.disabled = false; btn.textContent = '登 录';
    }
  } catch (err) {
    showAlert('网络错误，请稍后重试');
    btn.disabled = false; btn.textContent = '登 录';
  }
});
</script>
</body>
</html>`

// RegisterHTML 是注册页（含表单验证与密码强度提示）
const RegisterHTML = `<!DOCTYPE html>
<html lang="zh-CN" data-theme="dark">
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<title>注册 · PhiloFTP</title>
<style>
  :root {
    --brand: #0052d9; --brand-hover: #266fe8; --brand-active: #003cab;
    --success: #2ba471; --warning: #e37318; --danger: #d54941;
    --bg-page: #181818; --bg-card: #232324; --bg-input: #1a1a1a;
    --border: #424244; --text-1: rgba(255,255,255,.9); --text-2: rgba(255,255,255,.6); --text-3: rgba(255,255,255,.4);
    --radius: 12px;
  }
  * { box-sizing: border-box; margin: 0; padding: 0; }
  body {
    font-family: -apple-system, BlinkMacSystemFont, "Segoe UI", "PingFang SC", "Microsoft YaHei", sans-serif;
    background: radial-gradient(1200px 600px at 50% -10%, rgba(0,82,217,.18), transparent 60%), var(--bg-page);
    color: var(--text-1); min-height: 100vh; display: flex; align-items: center; justify-content: center; padding: 20px;
  }
  .reg-card {
    width: 440px; max-width: 100%; background: var(--bg-card); border: 1px solid var(--border);
    border-radius: var(--radius); padding: 34px 34px; box-shadow: 0 12px 48px rgba(0,0,0,.45);
  }
  .brand { text-align: center; margin-bottom: 24px; }
  .brand .logo { font-size: 40px; }
  .brand h1 { font-size: 20px; font-weight: 600; margin-top: 8px; }
  .field { margin-bottom: 16px; }
  .field label { display: block; font-size: 13px; color: var(--text-2); margin-bottom: 7px; }
  .input {
    width: 100%; padding: 11px 14px; border-radius: 8px; border: 1px solid var(--border);
    background: var(--bg-input); color: var(--text-1); font-size: 14px; font-family: inherit; outline: none;
    transition: border-color .2s;
  }
  .input:focus { border-color: var(--brand); }
  .input.invalid { border-color: var(--danger); }
  .input.valid { border-color: var(--success); }
  .input-wrap { position: relative; }
  .input-wrap .toggle {
    position: absolute; right: 12px; top: 50%; transform: translateY(-50%); color: var(--text-3);
    cursor: pointer; font-size: 13px; user-select: none;
  }
  .hint { font-size: 12px; margin-top: 6px; min-height: 16px; color: var(--text-3); }
  .hint.error { color: var(--danger); }
  .hint.ok { color: var(--success); }
  /* 密码强度条 */
  .strength { display: flex; gap: 6px; margin-top: 8px; }
  .strength .bar { flex: 1; height: 5px; border-radius: 3px; background: var(--border); transition: background .25s; }
  .strength .bar.lv1 { background: var(--danger); }
  .strength .bar.lv2 { background: var(--warning); }
  .strength .bar.lv3 { background: #cfb312; }
  .strength .bar.lv4 { background: var(--success); }
  .strength-text { font-size: 12px; margin-top: 5px; color: var(--text-2); }
  .btn-primary {
    width: 100%; padding: 12px; border: none; border-radius: 8px; background: var(--brand); color: #fff;
    font-size: 15px; cursor: pointer; font-family: inherit; transition: background .2s; margin-top: 6px;
  }
  .btn-primary:hover { background: var(--brand-hover); }
  .btn-primary:active { background: var(--brand-active); }
  .btn-primary:disabled { opacity: .6; cursor: not-allowed; }
  .row-between { text-align: center; margin-top: 16px; font-size: 13px; color: var(--text-3); }
  .row-between a { color: var(--brand); text-decoration: none; }
  .alert {
    display: none; padding: 10px 14px; border-radius: 8px; font-size: 13px; margin-bottom: 16px;
    background: rgba(213,73,65,.12); color: var(--danger); border: 1px solid rgba(213,73,65,.4);
  }
  .alert.show { display: block; }
  .spinner {
    display: inline-block; width: 15px; height: 15px; border: 2px solid rgba(255,255,255,.35);
    border-top-color: #fff; border-radius: 50%; animation: spin .7s linear infinite; vertical-align: -2px; margin-right: 6px;
  }
  @keyframes spin { to { transform: rotate(360deg); } }
</style>
</head>
<body>
  <div class="reg-card">
    <div class="brand">
      <div class="logo">📝</div>
      <h1>创建账号</h1>
      <p style="color:var(--text-3);font-size:13px;margin-top:4px">注册后将获得一个独立文件空间</p>
    </div>
    <div class="alert" id="alert"></div>
    <form id="regForm" autocomplete="off" novalidate>
      <div class="field">
        <label for="username">用户名</label>
        <input class="input" id="username" name="username" type="text" placeholder="3-32 位，仅字母/数字/_/-" maxlength="32">
        <div class="hint" id="usernameHint"></div>
      </div>
      <div class="field">
        <label for="password">密码</label>
        <div class="input-wrap">
          <input class="input" id="password" name="password" type="password" placeholder="至少 8 位，含大小写/数字/符号中 3 类">
          <span class="toggle" id="togglePw">显示</span>
        </div>
        <div class="strength" id="strengthBars">
          <div class="bar"></div><div class="bar"></div><div class="bar"></div><div class="bar"></div>
        </div>
        <div class="strength-text" id="strengthText">密码强度：—</div>
      </div>
      <div class="field">
        <label for="confirm">确认密码</label>
        <input class="input" id="confirm" name="confirm" type="password" placeholder="再次输入密码">
        <div class="hint" id="confirmHint"></div>
      </div>
      <button class="btn-primary" id="submitBtn" type="submit" disabled>注 册</button>
    </form>
    <div class="row-between">已有账号？ <a href="/login">返回登录</a></div>
  </div>

<script>
function showAlert(msg) {
  const a = document.getElementById('alert');
  a.textContent = msg; a.classList.add('show');
}
function setHint(el, msg, type) {
  el.textContent = msg || '';
  el.className = 'hint' + (type ? ' ' + type : '');
}

document.getElementById('togglePw').onclick = function() {
  const p = document.getElementById('password');
  const show = p.type === 'password';
  p.type = show ? 'text' : 'password';
  this.textContent = show ? '隐藏' : '显示';
};

// 密码强度计算（前端）
function calcStrength(pw) {
  if (!pw) return 0;
  let score = 0;
  if (pw.length >= 8) score++;
  if (pw.length >= 12) score++;
  let lower = /[a-z]/.test(pw), upper = /[A-Z]/.test(pw), digit = /[0-9]/.test(pw), sym = /[^A-Za-z0-9]/.test(pw);
  if (lower && upper) score++;
  if (digit) score++;
  if (sym) score++;
  return Math.min(score, 4);
}
const lvText = ['—', '弱', '较弱', '中等', '强'];
document.getElementById('password').addEventListener('input', function() {
  const lv = calcStrength(this.value);
  const bars = document.querySelectorAll('#strengthBars .bar');
  bars.forEach((b, i) => {
    b.className = 'bar' + (i < lv ? ' lv' + lv : '');
  });
  document.getElementById('strengthText').textContent = '密码强度：' + lvText[lv];
  validate();
});

document.getElementById('username').addEventListener('input', function() {
  const v = this.value.trim();
  const el = document.getElementById('usernameHint');
  const input = this;
  if (!v) { setHint(el, ''); input.classList.remove('valid','invalid'); }
  else if (v.length < 3 || v.length > 32) { setHint(el, '用户名长度需 3-32 位', 'error'); input.classList.add('invalid'); input.classList.remove('valid'); }
  else if (!/^[a-zA-Z0-9_\-]+$/.test(v)) { setHint(el, '仅可包含字母、数字、下划线和连字符', 'error'); input.classList.add('invalid'); input.classList.remove('valid'); }
  else { setHint(el, '可用', 'ok'); input.classList.add('valid'); input.classList.remove('invalid'); }
  validate();
});

document.getElementById('confirm').addEventListener('input', function() {
  const v = this.value;
  const pw = document.getElementById('password').value;
  const el = document.getElementById('confirmHint');
  if (!v) { setHint(el, ''); this.classList.remove('valid','invalid'); }
  else if (v !== pw) { setHint(el, '两次输入的密码不一致', 'error'); this.classList.add('invalid'); this.classList.remove('valid'); }
  else { setHint(el, '一致', 'ok'); this.classList.add('valid'); this.classList.remove('invalid'); }
  validate();
});

function formValid() {
  const u = document.getElementById('username').value.trim();
  const p = document.getElementById('password').value;
  const c = document.getElementById('confirm').value;
  const uOk = u.length >= 3 && u.length <= 32 && /^[a-zA-Z0-9_\-]+$/.test(u);
  const pOk = calcStrength(p) >= 3 && p.length >= 8;
  const cOk = c === p && c.length > 0;
  return uOk && pOk && cOk;
}
function validate() {
  document.getElementById('submitBtn').disabled = !formValid();
}

document.getElementById('regForm').addEventListener('submit', async function(e) {
  e.preventDefault();
  if (!formValid()) { showAlert('请检查表单填写是否完整正确'); return; }
  const btn = document.getElementById('submitBtn');
  const payload = {
    username: document.getElementById('username').value.trim(),
    password: document.getElementById('password').value,
    home: document.getElementById('username').value.trim()
  };
  btn.disabled = true; btn.innerHTML = '<span class="spinner"></span>注册中...';
  try {
    const res = await fetch('/api/register', {
      method: 'POST', headers: {'Content-Type': 'application/json'}, credentials: 'same-origin',
      body: JSON.stringify(payload)
    });
    const data = await res.json().catch(() => ({}));
    if (res.ok) {
      window.location.href = '/';
    } else {
      showAlert(data.error || '注册失败');
      btn.disabled = false; btn.textContent = '注 册';
    }
  } catch (err) {
    showAlert('网络错误，请稍后重试');
    btn.disabled = false; btn.textContent = '注 册';
  }
});
</script>
</body>
</html>`
