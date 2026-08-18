(function(){
  var state = { me:null, role:'user', cfg:null, curPath:'/' };
  function el(id){ return document.getElementById(id); }
  function q(sel,root){ return (root||document).querySelector(sel); }
  function qa(sel,root){ return Array.prototype.slice.call((root||document).querySelectorAll(sel)); }

  function api(path, opts){
    return fetch(path, Object.assign({ headers:{'Content-Type':'application/json'}, credentials:'include' }, opts))
      .then(function(r){ return r.json().then(function(d){ return { ok:r.ok, d:d }; }); });
  }
  function toast(msg, ok){
    var t = el('toast'); t.textContent = msg; t.className = 'toast show ' + (ok?'ok':'err');
    clearTimeout(t._t); t._t = setTimeout(function(){ t.className = 'toast'; }, 2600);
  }

  // 自定义确认弹窗：替换原生 confirm()，与「深空控制台」玻璃拟态风格统一
  function confirmDialog(opts){
    var type = opts.type || 'danger';
    var icoMap = {
      danger: { ico:'🗑', title:'确认删除', color:'var(--err)' },
      warn:   { ico:'⚠', title:'确认操作', color:'var(--warn)' },
      info:   { ico:'❔', title:'确认操作', color:'var(--cyan)' }
    };
    var m = icoMap[type] || icoMap.info;
    el('confirmIco').textContent = m.ico;
    el('confirmIco').style.color = m.color;
    el('confirmTitle').textContent = opts.title || m.title;
    el('confirmMsg').innerHTML = opts.message || '';
    var okBtn = el('confirmOk');
    okBtn.textContent = opts.okText || '确认';
    okBtn.className = 'btn btn-md ' + (type === 'danger' ? 'btn-danger' : 'btn-primary');
    el('confirm').classList.add('show');
    var done = false;
    function close(){ if(!done){ done = true; el('confirm').classList.remove('show'); cleanup(); } }
    function confirmOk(){ if(done) return; done = true; el('confirm').classList.remove('show'); cleanup(); if(opts.onOk) opts.onOk(); }
    function cleanup(){
      el('confirmCancel').removeEventListener('click', close);
      okBtn.removeEventListener('click', confirmOk);
      el('confirm').removeEventListener('click', overlayClick);
      document.removeEventListener('keydown', keyHandler);
    }
    function overlayClick(e){ if(e.target === el('confirm')) close(); }
    function keyHandler(e){ if(e.key === 'Escape') close(); if(e.key === 'Enter') confirmOk(); }
    el('confirmCancel').addEventListener('click', close);
    okBtn.addEventListener('click', confirmOk);
    el('confirm').addEventListener('click', overlayClick);
    document.addEventListener('keydown', keyHandler);
    okBtn.focus();
  }

  // 自定义输入弹窗：替换原生 prompt()，与玻璃拟态风格统一
  function promptDialog(opts){
    el('promptTitle').textContent = opts.title || '请输入';
    el('promptMsg').textContent = opts.message || '';
    el('promptIco').textContent = opts.icon || '📁';
    var input = el('promptInput');
    var hint = el('promptHint');
    input.value = opts.value || '';
    input.placeholder = opts.placeholder || '';
    input.type = opts.inputType || 'text';
    hint.textContent = '';
    hint.className = 'hint';
    hint.style.opacity = '0';
    el('promptOk').textContent = opts.okText || '确定';
    el('promptCancel').textContent = opts.cancelText || '取消';
    el('prompt').classList.add('show');
    var done = false;
    function close(){
      if(done) return; done = true;
      el('prompt').classList.remove('show');
      cleanup();
      if(opts.onCancel) opts.onCancel();
    }
    function submit(){
      if(done) return;
      var v = input.value;
      if(opts.validate && !opts.validate(v)) return;
      done = true;
      el('prompt').classList.remove('show');
      cleanup();
      if(opts.onOk) opts.onOk(v);
    }
    function cleanup(){
      el('promptCancel').removeEventListener('click', close);
      el('promptOk').removeEventListener('click', submit);
      el('prompt').removeEventListener('click', overlayClick);
      input.removeEventListener('keydown', onKey);
      input.removeEventListener('input', onInput);
    }
    function overlayClick(e){ if(e.target === el('prompt')) close(); }
    function onKey(e){ if(e.key === 'Escape') close(); if(e.key === 'Enter') submit(); }
    function onInput(){
      if(opts.onInput) opts.onInput(input.value, hint);
    }
    el('promptCancel').addEventListener('click', close);
    el('promptOk').addEventListener('click', submit);
    el('prompt').addEventListener('click', overlayClick);
    input.addEventListener('keydown', onKey);
    input.addEventListener('input', onInput);
    setTimeout(function(){ input.focus(); input.select(); }, 30);
  }
  function fmtSize(n){ if(n==null) return '—'; if(n<1024) return n+' B'; if(n<1048576) return (n/1024).toFixed(1)+' KB'; if(n<1073741824) return (n/1048576).toFixed(1)+' MB'; return (n/1073741824).toFixed(2)+' GB'; }
  function esc(s){ return String(s==null?'':s).replace(/[&<>"']/g, function(c){ return {'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[c]; }); }

  function checkAuth(){
    return api('/api/me').then(function(x){
      if(x.ok && x.d && x.d.username){
        state.me = x.d.username; state.role = x.d.role || 'user';
        return true;
      }
      window.location.href = '/login'; return false;
    }).catch(function(){ window.location.href = '/login'; return false; });
  }

  // ===== 导航 =====
  var NAV = [
    { id:'overview', ico:'◳', label:'概览', admin:false },
    { id:'files', ico:'🗂', label:'文件管理', admin:false },
    { id:'users', ico:'👥', label:'用户管理', admin:true },
    { id:'basic', ico:'🛠', label:'基础设置', admin:true },
    { id:'config', ico:'⚙', label:'系统配置', admin:true },
    { id:'system', ico:'📊', label:'系统信息', admin:false },
    { id:'about', ico:'ℹ', label:'关于', admin:false }
  ];
  function buildNav(){
    var isAdmin = state.role === 'admin';
    el('nav').innerHTML = NAV.map(function(n){
      if(n.admin && !isAdmin) return '';
      return '<div class="nav-item'+(n.id==='overview'?' active':'')+'" data-view="'+n.id+'"><span class="ico">'+n.ico+'</span>'+n.label+'</div>';
    }).join('');
    qa('.nav-item').forEach(function(item){
      item.onclick = function(){
        var v = item.getAttribute('data-view');
        if(NAV.filter(function(n){return n.id===v;})[0].admin && !isAdmin){ toast('权限不足：需要管理员角色', false); return; }
        showView(v);
      };
    });
  }
  function showView(id){
    qa('.view').forEach(function(v){ v.classList.toggle('active', v.id === 'v-'+id); });
    qa('.nav-item').forEach(function(n){ n.classList.toggle('active', n.getAttribute('data-view') === id); });
    var labels = { overview:'概览', files:'文件管理', users:'用户管理', basic:'基础设置', config:'系统配置', system:'系统信息', about:'关于' };
    el('title').textContent = labels[id] || '控制台';
    el('crumb').textContent = '控制台 · ' + (labels[id] || '');
    if(id==='overview') loadOverview();
    if(id==='files') loadFiles();
    if(id==='users') loadUsers();
    if(id==='basic') loadBasic();
    if(id==='config') loadConfig();
    if(id==='system') loadSystem();
    if(id==='about') loadAbout();
  }

  // ===== 概览 =====
  function loadOverview(){
    api('/api/status').then(function(x){ if(x.ok) el('ovStatus').innerHTML =
      'FTP：<b style="color:var(--cyan)">'+(x.d.ftp_port||'-')+'</b> · Web：<b style="color:var(--cyan)">'+(x.d.web_port||'-')+'</b><br>'+
      '运行时长：'+(x.d.uptime||'-')+'<br>FTP 连接：'+(x.d.ftp_conns||0)+' · 登录用户：'+(x.d.logged_in||0); });
    api('/api/about').then(function(x){ if(x.ok) el('ovAbout').innerHTML =
      'PhiloFTP '+esc(x.d.version||'')+'<br>Go '+esc(x.d.go_version||'')+(x.d.git_commit?' · '+esc(x.d.git_commit):''); });
    var cards = [
      { lbl:'📂 当前目录', big:'/', id:'c-files', go:'files' },
      { lbl:'⚡ 系统状态', big:'在线', id:'c-status', go:'system' }
    ];
    if(state.role === 'admin'){
      cards.push({ lbl:'👥 用户总数', big:'…', id:'c-users', go:'users' });
    }
    el('statCards').innerHTML = cards.map(function(c){
      return '<div class="card pad" data-go="'+c.go+'"><div class="lbl">'+c.lbl+'</div><div class="big" id="'+c.id+'">'+c.big+'</div></div>';
    }).join('');
    qa('#statCards .card').forEach(function(c){ c.onclick = function(){ showView(c.getAttribute('data-go')); }; });
    if(state.role === 'admin'){
      api('/api/users').then(function(x){ if(x.ok) el('c-users').textContent = (x.d.users||[]).length; });
    }
  }

  // ===== 文件 =====
  function loadFiles(path){
    if(path != null) state.curPath = path;
    var p = state.curPath;
    el('pathBar').textContent = '当前目录：' + p;
    el('upBtn').style.visibility = (p === '/' || p === '') ? 'hidden' : 'visible';
    renderBreadcrumb(p);
    api('/api/files?path=' + encodeURIComponent(p)).then(function(x){
      var tb = el('fileList'); el('fileEmpty').style.display = 'none';
      if(!x.ok){ toast(x.d.error||'加载失败', false); return; }
      var items = x.d.items || [];
      if(!items.length){ el('fileEmpty').style.display = 'block'; tb.innerHTML = ''; return; }
      tb.innerHTML = items.map(function(f){
        var ops;
        if(f.is_dir){
          ops = '<button class="btn btn-ghost btn-sm" data-open="'+esc(f.name)+'">打开</button>';
        } else {
          ops = '<button class="btn btn-ghost btn-sm" data-prev="'+esc(f.name)+'">预览</button>' +
                '<button class="btn btn-ghost btn-sm" data-dl="'+esc(f.name)+'">下载</button>';
        }
        ops += ' <button class="btn btn-danger btn-sm" data-del="'+esc(f.name)+'">删除</button>';
        var nameCell = '<span class="row-name'+(f.is_dir?' is-dir':'')+'" data-name="'+esc(f.name)+'" data-isdir="'+(f.is_dir?'1':'0')+'">'
          +(f.is_dir?'📁 ':'📄 ')+esc(f.name)+(f.is_dir?' ›':'')+'</span>';
        return '<tr><td>'+nameCell+'</td><td class="mono">'+fmtSize(f.size)+'</td><td class="mono">'+esc(f.mod_time||'')+'</td><td>'+ops+'</td></tr>';
      }).join('');
      qa('[data-open]', tb).forEach(function(b){ b.onclick = function(){ var np = p === '/' ? '/' + b.getAttribute('data-open') : p + '/' + b.getAttribute('data-open'); loadFiles(np); }; });
      qa('[data-prev]', tb).forEach(function(b){ b.onclick = function(){ preview(p, b.getAttribute('data-prev')); }; });
      qa('[data-dl]', tb).forEach(function(b){ b.onclick = function(){ download(p, b.getAttribute('data-dl')); }; });
      qa('[data-del]', tb).forEach(function(b){ b.onclick = function(){ delItem(p, b.getAttribute('data-del')); }; });
      qa('.row-name', tb).forEach(function(n){
        var isDir = n.getAttribute('data-isdir') === '1';
        var name = n.getAttribute('data-name');
        n.ondblclick = function(){
          if(isDir){
            var np = p === '/' ? '/' + name : p + '/' + name;
            loadFiles(np);
          } else {
            preview(p, name);
          }
        };
        // hover 提示：目录/文件分别提示双击动作
        n.title = isDir ? '双击进入' : '双击预览';
      });
    });
  }
  // 面包屑导航：将路径拆为逐级可点击段
  function renderBreadcrumb(p){
    var parts = (p || '/').split('/').filter(Boolean);
    var segs = ['<span class="crumb-seg" data-go="/">根目录</span>'];
    var acc = '';
    parts.forEach(function(s){
      acc += '/' + s;
      segs.push('<span class="crumb-sep">/</span><span class="crumb-seg" data-go="'+esc(acc)+'">'+esc(s)+'</span>');
    });
    var bc = el('breadcrumb');
    bc.innerHTML = segs.join('');
    qa('.crumb-seg', bc).forEach(function(s){ s.onclick = function(){ loadFiles(s.getAttribute('data-go')); }; });
  }
  // 返回上一级
  function goUp(){
    var p = state.curPath;
    if(!p || p === '/' || p === '') return;
    var idx = p.lastIndexOf('/');
    loadFiles(idx <= 0 ? '/' : p.substring(0, idx));
  }
  // 文件预览：按扩展名选择展示方式
  function preview(dir, name){
    var ext = (name.split('.').pop() || '').toLowerCase();
    var url = '/api/download?inline=1&path=' + encodeURIComponent((dir==='/'?'':dir) + '/' + name);
    var body = el('previewBody');
    el('previewName').textContent = name;
    var img = ['png','jpg','jpeg','gif','webp','bmp','svg','ico'];
    var txt = ['txt','md','log','csv','json','xml','yml','yaml','ini','conf','go','js','ts','css','html','sh','py','c','cpp','h','java','rs'];
    var vid = ['mp4','webm','ogg','mov'];
    var aud = ['mp3','wav','ogg','m4a','flac'];
    var html = '';
    if(img.indexOf(ext) >= 0){
      html = '<img class="preview-img" src="'+url+'" alt="'+esc(name)+'">';
    } else if(vid.indexOf(ext) >= 0){
      html = '<video class="preview-media" src="'+url+'" controls autoplay></video>';
    } else if(aud.indexOf(ext) >= 0){
      html = '<audio src="'+url+'" controls autoplay></audio>';
    } else if(txt.indexOf(ext) >= 0){
      html = '<iframe class="preview-frame" src="'+url+'"></iframe>';
    } else if(ext === 'pdf'){
      html = '<iframe class="preview-frame" src="'+url+'"></iframe>';
    } else {
      html = '<div class="preview-unsupported">该文件类型暂不支持在线预览。<br><button class="btn btn-primary btn-sm" id="prevDl">下载查看</button></div>';
    }
    body.innerHTML = html;
    var dl = el('prevDl');
    if(dl) dl.onclick = function(){ download(dir, name); };
    el('preview').classList.add('show');
  }
  function closePreview(){ el('preview').classList.remove('show'); el('previewBody').innerHTML = ''; }
  function download(dir, name){
    var url = '/api/download?path=' + encodeURIComponent((dir==='/'?'':dir) + '/' + name);
    var a = document.createElement('a'); a.href = url; a.download = name; document.body.appendChild(a); a.click(); a.remove();
  }
  function delItem(dir, name){
    confirmDialog({
      type:'danger',
      title:'删除文件',
      message:'确定要删除 <b style="color:var(--txt)">「'+esc(name)+'」</b> 吗？此操作将<b style="color:var(--err)">不可恢复</b>。',
      okText:'确认删除',
      onOk:function(){
        var full = (dir==='/'?'':dir) + '/' + name;
        api('/api/files?path=' + encodeURIComponent(full), { method:'DELETE' }).then(function(x){
          if(x.ok){ toast('已删除', true); loadFiles(); } else toast(x.d.error||'删除失败', false);
        });
      }
    });
  }
  function mkdir(){
    promptDialog({
      title:'新建目录',
      icon:'📁',
      message:'请输入新目录的名称，将创建在当前路径下：',
      placeholder:'例如 docs / 报告 / 2026',
      okText:'创建',
      validate:function(v){
        v = (v || '').trim();
        if(!v){ showPromptHint('请输入目录名称', 'error'); return false; }
        if(/[\\/:*?"<>|]/.test(v)){ showPromptHint('名称不能包含 \\ / : * ? " < > | 字符', 'error'); return false; }
        if(v === '.' || v === '..'){ showPromptHint('不能使用 . 或 ..', 'error'); return false; }
        if(v.length > 80){ showPromptHint('名称过长（最多 80 个字符）', 'error'); return false; }
        return true;
      },
      onOk:function(v){
        v = (v || '').trim();
        api('/api/mkdir', { method:'POST', body: JSON.stringify({ path: state.curPath, name: v }) }).then(function(x){
          if(x.ok){ toast('已创建目录「'+v+'」，可在文件列表中打开或返回上一级', true); loadFiles(); }
          else toast(x.d.error||'创建失败', false);
        });
      }
    });
  }
  function showPromptHint(msg, type){
    var h = el('promptHint');
    h.textContent = msg;
    h.className = 'hint visible' + (type === 'error' ? ' error' : type === 'ok' ? ' ok' : '');
    h.style.opacity = '1';
  }
  function upload(files){
    if(!files || !files.length) return;
    var fd = new FormData(); fd.append('path', state.curPath);
    qa('#fileInput').forEach(function(){ }); // noop
    for(var i=0;i<files.length;i++) fd.append('files', files[i]);
    el('mask').classList.add('show');
    var xhr = new XMLHttpRequest();
    xhr.open('POST', '/api/upload');
    xhr.withCredentials = true;
    xhr.upload.onprogress = function(e){ if(e.lengthComputable){ var pct = Math.round(e.loaded/e.total*100); el('pbar').style.width = pct+'%'; el('pmeta').textContent = pct+'%'; } };
    xhr.onload = function(){ el('mask').classList.remove('show'); el('pbar').style.width='0'; el('pmeta').textContent='0%';
      try{ var d = JSON.parse(xhr.responseText); if(d.ok || xhr.status===200){ toast('上传完成', true); loadFiles(); } else toast(d.error||'上传失败', false); }
      catch(e){ toast('上传失败', false); } };
    xhr.onerror = function(){ el('mask').classList.remove('show'); toast('网络错误', false); };
    xhr.send(fd);
  }

  // ===== 用户管理 =====
  function roleTag(role){ return role==='admin' ? '<span class="tag tag-adm">管理员</span>' : '<span class="tag tag-off">普通用户</span>'; }
  function loadUsers(){
    api('/api/users').then(function(x){
      var tb = el('userList');
      if(!x.ok){ toast(x.d.error||'加载失败', false); return; }
      var us = x.d.users || [];
      tb.innerHTML = us.map(function(u){
        var canDel = !(u.username === state.me) && !(u.role==='admin' && onlyAdmin(us));
        return '<tr><td class="mono">'+esc(u.username)+'</td><td>'+roleTag(u.role)+'</td><td>'+(u.enabled?'<span class="tag tag-ok">启用</span>':'<span class="tag tag-off">禁用</span>')+'</td><td class="mono">'+esc(u.home||'')+'</td><td>'+
          '<button class="btn btn-ghost btn-sm" data-edit="'+esc(u.username)+'">编辑</button> '+
          (canDel?'<button class="btn btn-danger btn-sm" data-del="'+esc(u.username)+'">删除</button>':'')+'</td></tr>';
      }).join('');
      qa('[data-edit]', tb).forEach(function(b){ b.onclick = function(){ openUserModal(b.getAttribute('data-edit')); }; });
      qa('[data-del]', tb).forEach(function(b){ b.onclick = function(){ delUser(b.getAttribute('data-del')); }; });
    });
  }
  function onlyAdmin(us){ return us.filter(function(u){ return u.role==='admin'; }).length <= 1; }
  function openUserModal(name){
    api('/api/users').then(function(x){
      var u = (x.d.users||[]).filter(function(z){ return z.username===name; })[0] || {};
      var isEdit = !!name;
      el('modalBody').innerHTML =
        '<h3>'+(isEdit?'编辑用户':'新增用户')+'</h3>'+
        (isEdit?'<div class="field"><label>用户名</label><input class="input" value="'+esc(u.username)+'" disabled></div>'
              :'<div class="field"><label>用户名</label><input class="input" id="uName" placeholder="至少 3 个字符"></div>')+
        '<div class="field"><label>密码'+(isEdit?'（留空则不修改）':'')+'</label><input class="input" id="uPass" type="password" placeholder="设置密码"></div>'+
        '<div class="field"><label>角色</label><select class="input" id="uRole"><option value="user"'+((u.role||'user')==='user'?' selected':'')+'>普通用户</option><option value="admin"'+(u.role==='admin'?' selected':'')+'>管理员</option></select></div>'+
        '<div class="field check"><input type="checkbox" id="uEnabled" '+(u.enabled!==false?'checked':'')+'><label for="uEnabled" style="cursor:pointer">账户启用</label></div>'+
        '<div style="display:flex;gap:10px;margin-top:18px"><button class="btn btn-primary" style="flex:1" id="uSave">保存</button><button class="btn btn-ghost" id="uCancel">取消</button></div>';
      el('modal').classList.add('show');
      el('uCancel').onclick = closeModal;
      el('uSave').onclick = saveUser;
    });
  }
  function saveUser(){
    var isEdit = el('uName') ? false : true;
    var name = isEdit ? q('#modalBody input.input').value.trim() : el('uName').value.trim();
    var pass = el('uPass').value, role = el('uRole').value, enabled = el('uEnabled').checked;
    if(!isEdit && name.length < 3){ toast('用户名至少 3 个字符', false); return; }
    var body = { role: role, enabled: enabled };
    if(!isEdit) body.username = name;
    if(pass) body.password = pass;
    var m = isEdit ? 'PUT' : 'POST';
    var path = isEdit ? '/api/users/' + encodeURIComponent(name) : '/api/users';
    api(path, { method:m, body: JSON.stringify(body) }).then(function(x){
      if(x.ok){ closeModal(); toast('已保存', true); loadUsers(); } else toast(x.d.error||'保存失败', false);
    });
  }
  function delUser(name){
    if(name === state.me){ toast('不能删除当前登录账户', false); return; }
    confirmDialog({
      type:'danger',
      title:'删除用户',
      message:'确定要删除用户 <b style="color:var(--txt)">「'+esc(name)+'」</b> 吗？该账户将<b style="color:var(--err)">永久失效</b>，此操作不可恢复。',
      okText:'确认删除',
      onOk:function(){
        api('/api/users/' + encodeURIComponent(name), { method:'DELETE' }).then(function(x){
          if(x.ok){ toast('已删除', true); loadUsers(); } else toast(x.d.error||'删除失败', false);
        });
      }
    });
  }

  // ===== 基础设置 / 系统配置 =====
  function field(id){ return el(id).value; }
  function loadBasic(){
    api('/api/config').then(function(x){
      if(!x.ok){ toast(x.d.error||'加载失败', false); return; }
      state.cfg = x.d; var c = x.d;
      el('cfgDataDir').value = c.data_dir || '';
      el('cfgRegister').checked = !!c.allow_register;
    });
  }
  function saveBasic(){
    var body = { allow_register: el('cfgRegister').checked, data_dir: el('cfgDataDir').value.trim() };
    api('/api/config', { method:'PUT', body: JSON.stringify(body) }).then(function(x){
      if(x.ok){ toast('基础设置已保存', true); loadBasic(); loadOverview(); } else toast(x.d.error||'保存失败', false);
    });
  }
  function loadConfig(){
    api('/api/config').then(function(x){
      if(!x.ok){ toast(x.d.error||'加载失败', false); return; }
      state.cfg = x.d; var c = x.d;
      el('cfgFtpPort').value = c.ftp_port || '';
      el('cfgWebPort').value = c.web_port || '';
      el('cfgPasv').value = c.pasv_port_range || '';
      el('cfgFtps').checked = !!c.enable_ftps;
    });
  }
  function saveCfg(){
    var body = {
      ftp_port: parseInt(el('cfgFtpPort').value,10),
      web_port: parseInt(el('cfgWebPort').value,10),
      pasv_port_range: el('cfgPasv').value.trim(),
      enable_ftps: el('cfgFtps').checked
    };
    api('/api/config', { method:'PUT', body: JSON.stringify(body) }).then(function(x){
      if(x.ok){ toast('系统配置已保存（FTP 热重载生效）', true); loadConfig(); } else toast(x.d.error||'保存失败', false);
    });
  }

  // ===== 系统信息 / 关于 =====
  function loadSystem(){
    api('/api/system').then(function(x){
      if(!x.ok){ toast(x.d.error||'加载失败', false); return; }
      var d = x.d, rows = '';
      function row(k,v){ return '<div style="display:flex;justify-content:space-between;padding:9px 0;border-bottom:1px solid rgba(30,42,64,.6)"><span style="color:var(--haze)">'+k+'</span><span class="mono">'+esc(v)+'</span></div>'; }
      rows += row('版本', d.version||'-');
      rows += row('Go 版本', d.go_version||'-');
      rows += row('Git 提交', d.git_commit||'-');
      rows += row('运行平台', d.platform||'-');
      rows += row('运行时长', d.uptime||'-');
      rows += row('FTP 端口', d.ftp_port||'-');
      rows += row('Web 端口', d.web_port||'-');
      rows += row('FTP 连接数', d.ftp_conns||0);
      rows += row('登录用户数', d.logged_in||0);
      el('sysInfo').innerHTML = rows;
    });
  }
  function loadAbout(){
    api('/api/about').then(function(x){
      var d = x.ok ? x.d : {};
      el('aboutInfo').innerHTML =
        '<div style="display:flex;align-items:center;gap:14px;margin-bottom:16px"><div style="width:46px;height:46px;border-radius:12px;background:linear-gradient(135deg,var(--cyan),var(--cyan2));color:#04121a;font-size:22px;font-weight:700;display:flex;align-items:center;justify-content:center">⬡</div><div><b style="font-size:18px">PhiloFTP</b><br><span style="color:var(--haze);font-size:13px">内网 FTP 控制中枢</span></div></div>'+
        '版本：'+esc(d.version||'-')+' · Go '+esc(d.go_version||'-')+(d.git_commit?' · '+esc(d.git_commit):'')+'<br><br>'+
        '基于 Go + Gin 构建，前端为纯静态资源（与后端完全分离）。';
    });
  }

  function closeModal(){ el('modal').classList.remove('show'); }
  function logout(){ api('/api/logout', { method:'POST' }).then(function(){ window.location.href = '/login'; }); }

  // ===== 启动 =====
  function init(){
    checkAuth().then(function(ok){
      if(!ok) return;
      el('meName').textContent = state.me + (state.role==='admin'?'（管理员）':'');
      el('layout').style.display = 'flex';
      buildNav();
      el('logout').onclick = logout;
      el('modal').onclick = function(e){ if(e.target === el('modal')) closeModal(); };
      el('newDirBtn').onclick = mkdir;
      el('refreshBtn').onclick = function(){ loadFiles(); };
      el('upBtn').onclick = goUp;
      el('previewClose').onclick = closePreview;
      el('preview').onclick = function(e){ if(e.target === el('preview')) closePreview(); };
      el('fileInput').onchange = function(e){ upload(e.target.files); e.target.value=''; };
      if(state.role === 'admin'){
        el('addUserBtn').onclick = function(){ openUserModal(''); };
        el('saveBasicBtn').onclick = saveBasic;
        el('saveCfgBtn').onclick = saveCfg;
      }
      loadOverview();
    });
  }
  if(document.readyState !== 'loading') init(); else document.addEventListener('DOMContentLoaded', init);
})();
