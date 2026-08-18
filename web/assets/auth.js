(function(){
  function el(id){ return document.getElementById(id); }
  function alertBox(msg){
    var a = el('alert');
    if(!msg){ a.classList.remove('show'); return; }
    a.textContent = msg; a.classList.add('show');
  }
  function setBtn(busy, label){
    var b = el('b');
    if(busy){ b.disabled = true; b.dataset.old = b.textContent; b.innerHTML = '<span class="spin"></span>' + (label||'处理中'); }
    else { b.disabled = false; b.textContent = b.dataset.old || label || '登 录'; }
  }
  function bindToggle(){
    var t = el('t'), p = el('p');
    if(t && p){ t.onclick = function(){ var s = p.type==='password'?'text':'password'; p.type=s; t.textContent = s==='password'?'显示':'隐藏'; }; }
  }
  function checkPassword(s){
    var lv = 0;
    if(s.length >= 8) lv++;
    if(/[a-z]/.test(s) && /[A-Z]/.test(s)) lv++;
    if(/\d/.test(s)) lv++;
    if(/[^A-Za-z0-9]/.test(s)) lv++;
    return lv;
  }

  function api(path, opts){
    return fetch(path, Object.assign({ headers:{'Content-Type':'application/json'}, credentials:'include' }, opts));
  }

  // 探测自助注册开关
  fetch('/api/register/enabled').then(function(r){ return r.json(); }).then(function(d){
    if(d && d.enabled){
      el('regLink').innerHTML = '没有账户？<a href="/register">立即注册</a>';
    }
  }).catch(function(){});

  window.initLogin = function(){
    bindToggle();
    el('f').onsubmit = function(e){
      e.preventDefault();
      alertBox('');
      var u = el('u').value.trim(), p = el('p').value;
      if(!u || !p){ alertBox('请输入用户名和密码'); return; }
      setBtn(true, '登录中');
      api('/api/login', { method:'POST', body: JSON.stringify({ username:u, password:p }) })
        .then(function(r){ return r.json().then(function(d){ return { ok:r.ok, d:d }; }); })
        .then(function(x){
          if(x.ok && x.d && !x.d.error){ window.location.href = '/'; }
          else { setBtn(false, '登 录'); alertBox((x.d && x.d.error) || '登录失败'); }
        })
        .catch(function(){ setBtn(false, '登 录'); alertBox('网络错误，请重试'); });
    };
  };

  window.initRegister = function(){
    bindToggle();
    var p = el('p'), pc = el('pc'), u = el('u');
    function paintStrength(){
      var lv = checkPassword(p.value);
      var bars = document.querySelectorAll('#strength .bar');
      bars.forEach(function(b, i){ b.className = 'bar' + (i < lv ? ' lv'+lv : ''); });
      var hint = el('pHint');
      if(!p.value){ hint.textContent=''; hint.className='hint'; }
      else if(lv < 3){ hint.textContent='密码强度较弱（建议 8 位以上，含大小写、数字、符号）'; hint.className='hint error'; }
      else { hint.textContent='密码强度良好'; hint.className='hint ok'; }
    }
    p.oninput = paintStrength;
    u.oninput = function(){
      var h = el('uHint');
      if(u.value.length > 0 && u.value.length < 3){ h.textContent='用户名至少 3 个字符'; h.className='hint error'; }
      else { h.textContent=''; h.className='hint'; }
    };
    el('f').onsubmit = function(e){
      e.preventDefault();
      alertBox('');
      var name = u.value.trim(), pw = p.value, pwc = pc.value;
      if(name.length < 3){ alertBox('用户名至少 3 个字符'); return; }
      if(checkPassword(pw) < 3){ alertBox('密码强度不足，请按提示设置'); return; }
      if(pw !== pwc){ alertBox('两次输入的密码不一致'); return; }
      setBtn(true, '注册中');
      api('/api/register', { method:'POST', body: JSON.stringify({ username:name, password:pw }) })
        .then(function(r){ return r.json().then(function(d){ return { ok:r.ok, d:d }; }); })
        .then(function(x){
          if(x.ok && x.d && !x.d.error){ window.location.href = '/login'; }
          else { setBtn(false, '注 册'); alertBox((x.d && x.d.error) || '注册失败'); }
        })
        .catch(function(){ setBtn(false, '注 册'); alertBox('网络错误，请重试'); });
    };
  };
})();
