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

  // 统一输入提示交互：聚焦显示 focus 提示，输入中按验证切换 ok/error，失焦且空则隐藏
  function bindHint(inputId, hintId, focusText, validator){
    var input = el(inputId), hint = el(hintId), field = input && input.closest('.field');
    if(!input || !hint) return;
    function set(cls, text){
      hint.className = 'hint visible' + (cls ? ' '+cls : '');
      if(text != null) hint.textContent = text;
    }
    function hide(){
      if(!input.value){ hint.className = 'hint'; }
    }
    input.addEventListener('focus', function(){
      if(field) field.classList.add('focused');
      if(!input.value){ set('focus', focusText); }
      else if(validator){ var r = validator(input.value); set(r.cls, r.text); }
      else { set('focus', focusText); }
    });
    input.addEventListener('blur', function(){
      if(field) field.classList.remove('focused');
      hide();
    });
    input.addEventListener('input', function(){
      if(validator){ var r = validator(input.value); set(r.cls, r.text); }
      else { set('focus', focusText); }
    });
    // 初始：空则隐藏，有值则校验
    hide();
    if(input.value && validator){ var r = validator(input.value); set(r.cls, r.text); }
  }

  function api(path, opts){
    return fetch(path, Object.assign({ headers:{'Content-Type':'application/json'}, credentials:'include' }, opts));
  }

  // 登录页默认显示注册入口；仅当后端明确关闭自助注册时隐藏该链接
  var regLink = el('regLink');
  fetch('/api/config/public').then(function(r){ return r.json(); }).then(function(d){
    if(d && d.allow_register === false && regLink){
      regLink.style.display = 'none';
    }
  }).catch(function(){});

  window.initLogin = function(){
    loadAccessInfo();
    bindToggle();
    bindHint('u', 'uHint', '请输入用户名，例如 admin', function(v){
      if(!v.trim()) return {cls:'focus', text:'请输入用户名，例如 admin'};
      return {cls:'ok', text:'用户名格式正确'};
    });
    bindHint('p', 'pHint', '请输入登录密码', function(v){
      if(!v) return {cls:'focus', text:'请输入登录密码'};
      return {cls:'ok', text:'已输入密码'};
    });
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

  // 加载并展示"局域网访问"信息（IP / mDNS 主机名 / 端口 / 二维码）
  function loadAccessInfo(){
    fetch('/api/access').then(function(r){ return r.json(); }).then(function(d){
      if(!d || !d.web_port) return;
      if(el('accessIP')) el('accessIP').textContent = d.ip || '—';
      if(el('accessHost')) el('accessHost').textContent = d.hostname || 'philoftp.local';
      if(el('accessPort')) el('accessPort').textContent = d.web_port;
      // 二维码图片（后端生成）
      if(el('accessQR')){
        var img = new Image();
        img.src = '/api/access/qr';
        img.alt = '访问二维码';
        img.onload = function(){ el('accessQR').innerHTML = ''; el('accessQR').appendChild(img); };
        img.onerror = function(){ el('accessQR').innerHTML = '<span style="color:var(--muted)">二维码加载失败</span>'; };
      }
    }).catch(function(){});
  };

  window.initRegister = function(){
    bindToggle();
    var p = el('p'), pc = el('pc'), u = el('u');
    function paintStrength(){
      var lv = checkPassword(p.value);
      var bars = document.querySelectorAll('#strength .bar');
      bars.forEach(function(b, i){ b.className = 'bar' + (i < lv ? ' lv'+lv : ''); });
      var hint = el('pHint');
      if(!p.value){ hint.className='hint'; hint.textContent=''; }
      else if(lv < 3){ hint.className='hint visible error'; hint.textContent='密码强度较弱（建议 8 位以上，含大小写、数字、符号）'; }
      else { hint.className='hint visible ok'; hint.textContent='密码强度良好'; }
    }
    p.addEventListener('input', paintStrength);
    p.addEventListener('focus', function(){ if(!p.value) el('pHint').className='hint visible focus'; el('pHint').textContent='8 位以上，包含大小写、数字、符号'; });
    p.addEventListener('blur', function(){ if(!p.value){ el('pHint').className='hint'; el('pHint').textContent=''; } });

    bindHint('u', 'uHint', '用户名至少 3 个字符，仅支持字母、数字与下划线', function(v){
      if(!v) return {cls:'focus', text:'用户名至少 3 个字符，仅支持字母、数字与下划线'};
      if(v.length < 3) return {cls:'error', text:'用户名太短，至少 3 个字符'};
      if(!/^[A-Za-z0-9_]+$/.test(v)) return {cls:'error', text:'用户名仅支持字母、数字与下划线'};
      return {cls:'ok', text:'用户名可用'};
    });
    bindHint('pc', 'pcHint', '请再次输入与上面相同的密码', function(v){
      if(!v) return {cls:'focus', text:'请再次输入与上面相同的密码'};
      if(v !== p.value) return {cls:'error', text:'两次输入的密码不一致'};
      return {cls:'ok', text:'两次输入一致'};
    });

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
