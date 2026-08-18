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
    { id:'settings', ico:'⚙', label:'系统设置', admin:true },
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
    var labels = { overview:'概览', files:'文件管理', users:'用户管理', settings:'系统设置', about:'关于' };
    el('title').textContent = labels[id] || '控制台';
    el('crumb').textContent = '控制台 · ' + (labels[id] || '');
    if(id==='overview') loadOverview();
    if(id==='files') loadFiles();
    if(id==='users') loadUsers();
    if(id==='settings') loadSettings();
    if(id==='about') loadAbout();
  }

  // ===== 概览 =====
  // 概览页：调用 /api/overview 一次性获取完整统计快照，
  // 渲染多区块仪表盘（KPI / 分布条形图 / 环形存储 / 活跃会话 / Top 文件等）
  function loadOverview(){
    // 头部
    el('ovNow').textContent = '同步中…';
    el('ovUptime').textContent = '—';
    el('ovKpiMain').innerHTML = skeletonKpi(4);
    el('ovKpiSub').innerHTML = skeletonKpi(3);

    api('/api/overview').then(function(x){
      if(!x.ok){ toast(x.d.error||'概览加载失败', false); return; }
      renderOverview(x.d);
    });
  }
  function skeletonKpi(n){
    var s = '';
    for(var i=0;i<n;i++) s += '<div class="ov-kpi-cell ov-skel"><div class="ov-skel-line w40"></div><div class="ov-skel-line w70"></div><div class="ov-skel-line w30"></div></div>';
    return s;
  }
  function fmtSize(n){
    if(n==null || n===0) return '0 B';
    if(n<1024) return n+' B';
    if(n<1048576) return (n/1024).toFixed(1)+' KB';
    if(n<1073741824) return (n/1048576).toFixed(1)+' MB';
    return (n/1073741824).toFixed(2)+' GB';
  }
  function pct(a, b){
    if(!b) return 0;
    return Math.max(0, Math.min(100, a*100/b));
  }
  function renderOverview(d){
    // 顶部时间与运行时长
    el('ovNow').textContent = '同步于 ' + (d.now || '—');
    el('ovUptime').textContent = d.uptime || '—';

    // 主 KPI 4 卡
    var fileSize = d.files.total_size || 0;
    var usedPct = d.storage.used_pct || 0;
    el('ovKpiMain').innerHTML = [
      kpiCell('kpi-users',   '👥 用户总数', d.users.total, '启用 ' + d.users.enabled + ' · 禁用 ' + d.users.disabled, 'cyan', 'users'),
      kpiCell('kpi-sessions','🟢 活跃会话', d.users.logged_in, '最近登录用户数', 'ok', 'users'),
      kpiCell('kpi-files',   '🗂 文件 / 目录', (d.files.file_count||0) + ' / ' + (d.files.dir_count||0), fmtSize(fileSize), 'cyan2', 'files'),
      kpiCell('kpi-storage', '💾 存储使用', (usedPct||0).toFixed(1) + '%', d.storage.total? (fmtSize(d.storage.used)+' / '+fmtSize(d.storage.total)) : '—', usedPct>85?'err':(usedPct>60?'warn':'ok'))
    ].join('');

    // 副 KPI 3 卡
    el('ovKpiSub').innerHTML = [
      kpiCell('kpi-admins',  '🛡 管理员', d.users.admins, '普通用户 ' + d.users.normal, 'cyan'),
      kpiCell('kpi-datadir', '📁 数据目录', d.server.data_dir || '—', 'FTP 数据根', 'cyan2', 'files'),
      kpiCell('kpi-go',      '⚙ Go 运行时', (d.load.go_version||'').replace('go',''), '协程 ' + d.load.goroutines, 'ok')
    ].join('');

    // 用户分布
    renderUserBars(d.users);
    // 文件类型分布
    renderExtBars(d.files.ext_dist || []);
    // Top 5 大文件
    renderTopFiles(d.files.top_files || []);
    // 服务器状态
    renderServer(d.server);
    // 活跃会话
    renderLogged(d.users);
    // 存储环
    renderStorage(d.storage);
    // 运行时负载
    renderLoad(d.load);
    // 最近更新
    el('ovLastUpdate').textContent = d.last_update || '—';
    el('ovLastSub').textContent = d.last_update && d.last_update !== '—' ? '数据目录中最近修改的文件' : '暂无文件修改记录';

    // 绑定跳转
    qa('[data-go]', el('v-overview')).forEach(function(b){
      b.onclick = function(){
        var v = b.getAttribute('data-go');
        if(!v) return;
        showView(v);
      };
    });

    // 启用概览页卡片拖拽自定义布局
    renderCustomCards();
    initOverviewDrag();
    applyOverviewLayout();
    applyCardVisibility();

    // 重置布局按钮
    var resetBtn = el('ovResetLayout');
    if(resetBtn){
      resetBtn.onclick = function(){
        localStorage.removeItem('philoftp-overview-layout');
        toast('已恢复默认布局', true);
        applyOverviewLayout(); // 无保存布局时自然恢复默认
      };
    }

    // 添加卡片按钮
    var addBtn = el('ovAddCard');
    if(addBtn) addBtn.onclick = openAddCard;

    // 卡片设置按钮
    var setBtn = el('ovSettingsBtn');
    if(setBtn) setBtn.onclick = openOverviewSettings;

    // 加载概览页"局域网访问"卡片的二维码与访问信息
    loadOverviewAccess();
  }

  // 加载概览页局域网访问卡片（二维码 + IP/主机名/端口 + 访问链接）
  function loadOverviewAccess(){
    var qrEl = el('ovAccessQR');
    if(!qrEl) return; // 卡片被隐藏或不存在时跳过
    api('/api/access').then(function(x){
      if(!x.ok) return;
      var d = x.d;
      if(el('ovAccessIP')) el('ovAccessIP').textContent = d.ip || '—';
      if(el('ovAccessHost')) el('ovAccessHost').textContent = d.hostname || 'philoftp.local';
      if(el('ovAccessPort')) el('ovAccessPort').textContent = d.web_port || '—';
      var linkEl = el('ovAccessLink');
      if(linkEl && d.ip && d.web_port){
        linkEl.href = 'http://' + d.ip + ':' + d.web_port;
        linkEl.textContent = '🔗 IP 访问 ' + d.ip + ':' + d.web_port;
      }
      var mdnsEl = el('ovAccessMdns');
      if(mdnsEl && d.hostname && d.web_port){
        mdnsEl.href = 'http://' + d.hostname + ':' + d.web_port;
        mdnsEl.textContent = '🌐 mDNS 访问 ' + d.hostname + ':' + d.web_port;
      }
    });
    var img = new Image();
    img.src = '/api/access/qr';
    img.alt = '访问二维码';
    img.onload = function(){ qrEl.innerHTML = ''; qrEl.appendChild(img); };
    img.onerror = function(){ qrEl.innerHTML = '<span style="font-size:12px;color:var(--muted)">二维码加载失败</span>'; };
  }

  // ===== 卡片配置模型（每张卡的 draggable / enabled / slot）=====
  function getCardConfig(){
    try {
      return JSON.parse(localStorage.getItem('philoftp-overview-cards') || '{}');
    } catch(e){ return {}; }
  }
  function setCardConfig(cfg){
    try { localStorage.setItem('philoftp-overview-cards', JSON.stringify(cfg)); } catch(e){}
  }
  function cardConf(id){
    var cfg = getCardConfig();
    var c = cfg[id] || {};
    return { draggable: c.draggable !== false, enabled: c.enabled !== false, custom: !!c.custom };
  }
  function setCardConf(id, patch){
    var cfg = getCardConfig();
    cfg[id] = Object.assign({}, cfg[id] || {}, patch);
    setCardConfig(cfg);
  }
  // 内置卡片名 → 中文显示名
  var BUILTIN_CARD_NAMES = {
    'kpi-users':'用户总数','kpi-sessions':'活跃会话','kpi-files':'文件/目录','kpi-storage':'存储使用',
    'kpi-admins':'管理员数','kpi-datadir':'数据目录','kpi-go':'Go 运行时',
    'user-dist':'用户分布','server-status':'服务器状态','active-sessions':'活跃会话列表',
    'last-update':'最近更新','storage-ring':'存储使用率','file-ext':'文件类型分布','top-files':'Top 5 大文件'
  };
  // 自定义卡片
  function getCustomCards(){
    try { return JSON.parse(localStorage.getItem('philoftp-overview-custom') || '{}'); }
    catch(e){ return {}; }
  }
  function setCustomCards(map){
    try { localStorage.setItem('philoftp-overview-custom', JSON.stringify(map)); } catch(e){}
  }
  function renderCustomCards(){
    var custom = getCustomCards();
    Object.keys(custom).forEach(function(id){
      var c = custom[id];
      var card = findOvCard(id);
      if(card){
        var content = card.querySelector('.ov-custom-content');
        if(content) content.textContent = c.content || '';
        var title = card.querySelector('.ov-custom-title');
        if(title) title.textContent = c.title || '备注';
        return;
      }
      // 尚未渲染 → 创建 DOM 并放入对应容器
      var div = document.createElement('div');
      div.className = 'card pad ov-card ov-custom';
      div.setAttribute('data-ov-card-id', id);
      div.setAttribute('draggable', 'true');
      div.innerHTML =
        '<div class="lbl">📋 <span class="ov-custom-title">'+esc(c.title||'备注')+'</span></div>'+
        '<div class="ov-custom-content">'+esc(c.content||'')+'</div>';
      var slot = c.slot || 'ovCol2';
      var container = el(slot) || el('ovCol2');
      container.appendChild(div);
    });
  }
  // 依据 draggable 配置设定卡片是否可拖
  function applyCardVisibility(){
    qa('[data-ov-card-id]').forEach(function(card){
      var id = card.getAttribute('data-ov-card-id');
      var conf = cardConf(id);
      card.style.display = conf.enabled ? '' : 'none';
      card.setAttribute('draggable', conf.draggable ? 'true' : 'false');
      var handle = card.querySelector('.ov-drag-handle');
      if(handle) handle.style.display = conf.draggable ? '' : 'none';
    });
  }

  // ===== 概览页卡片拖拽布局 =====
  var _ovDragSrc = null;
  function initOverviewDrag(){
    // 为主体 ov-card 补充拖拽手柄（KPI 卡片已在 kpiCell 中内联）
    qa('.ov-card[data-ov-card-id]').forEach(function(card){
      if(!card.querySelector('.ov-drag-handle')){
        var h = document.createElement('span');
        h.className = 'ov-drag-handle';
        h.title = '拖动调整卡片位置';
        h.textContent = '⋮⋮';
        card.insertBefore(h, card.firstChild);
      }
      card.setAttribute('draggable', 'true');
    });

    var draggables = qa('[data-ov-card-id]');
    draggables.forEach(function(item){
      item.ondragstart = onOvDragStart;
      item.ondragend = onOvDragEnd;
    });
    var containers = ['ovKpiMain','ovKpiSub','ovCol0','ovCol1','ovCol2'];
    containers.forEach(function(id){
      var c = el(id);
      if(!c) return;
      c.ondragover = onOvDragOver;
      c.ondragenter = onOvDragEnter;
      c.ondragleave = onOvDragLeave;
      c.ondrop = onOvDrop;
    });
  }
  function onOvDragStart(e){
    _ovDragSrc = this;
    this.classList.add('ov-dragging');
    e.dataTransfer.effectAllowed = 'move';
    e.dataTransfer.setData('text/plain', this.getAttribute('data-ov-card-id'));
  }
  function onOvDragEnd(e){
    this.classList.remove('ov-dragging');
    qa('.ov-drop-target').forEach(function(n){ n.classList.remove('ov-drop-target'); });
    _ovDragSrc = null;
    saveOverviewLayout();
  }
  function onOvDragEnter(e){
    e.preventDefault();
    if(this === _ovDragSrc || this === _ovDragSrc.parentNode) return;
    this.classList.add('ov-drop-target');
  }
  function onOvDragLeave(e){
    this.classList.remove('ov-drop-target');
  }
  function onOvDragOver(e){
    e.preventDefault();
    e.dataTransfer.dropEffect = 'move';
    var container = this;
    var afterEl = getOvDropAfter(container, e.clientY, e.clientX);
    if(_ovDragSrc && _ovDragSrc.parentNode === container && afterEl === _ovDragSrc.nextElementSibling){
      return;
    }
    // 移动源元素到目标位置
    if(_ovDragSrc){
      if(afterEl) container.insertBefore(_ovDragSrc, afterEl);
      else container.appendChild(_ovDragSrc);
    }
  }
  function onOvDrop(e){
    e.preventDefault();
    this.classList.remove('ov-drop-target');
    saveOverviewLayout();
  }
  // 计算容器内拖拽应插入到哪个子元素之前
  // KPI 容器是横向 grid，用 clientX；ov-col 是纵向 flex，用 clientY
  function getOvDropAfter(container, y, x){
    var draggable = qa('[data-ov-card-id]', container);
    var isRow = container.classList.contains('ov-kpi');
    return draggable.reduce(function(closest, child){
      if(child === _ovDragSrc) return closest;
      var box = child.getBoundingClientRect();
      var offset;
      if(isRow){
        offset = x - (box.left + box.width / 2);
        if(offset < 0 && offset > closest.offset){
          return { offset: offset, element: child };
        }
      } else {
        offset = y - (box.top + box.height / 2);
        if(offset < 0 && offset > closest.offset){
          return { offset: offset, element: child };
        }
      }
      return closest;
    }, { offset: Number.NEGATIVE_INFINITY }).element;
  }
  function saveOverviewLayout(){
    try {
      var layout = {};
      var containers = ['ovKpiMain','ovKpiSub','ovCol0','ovCol1','ovCol2'];
      containers.forEach(function(id){
        var c = el(id);
        if(!c) return;
        layout[id] = qa('[data-ov-card-id]', c).map(function(n){ return n.getAttribute('data-ov-card-id'); });
      });
      // 智能布局保护：三列主体（col0/col1/col2）不允许为空，空时从其他列移入第一张卡
      var cols = ['ovCol0','ovCol1','ovCol2'];
      for(var i=0;i<cols.length;i++){
        var colId = cols[i];
        if(layout[colId] && layout[colId].length) continue;
        // 该列空 → 从有卡的列借第一张卡
        var donor = cols.filter(function(x){ return layout[x] && layout[x].length; })[0];
        if(donor){
          var take = layout[donor].shift();
          layout[colId] = [take];
        }
      }
      // 同步自定义卡片的 slot
      var custom = getCustomCards();
      Object.keys(custom).forEach(function(id){
        for(var j=0;j<containers.length;j++){
          if(layout[containers[j]].indexOf(id) !== -1){ custom[id].slot = containers[j]; break; }
        }
      });
      setCustomCards(custom);
      localStorage.setItem('philoftp-overview-layout', JSON.stringify(layout));
    } catch(err){}
  }
  function applyOverviewLayout(){
    try {
      var raw = localStorage.getItem('philoftp-overview-layout');
      if(!raw) return;
      var layout = JSON.parse(raw);
      Object.keys(layout).forEach(function(id){
        var c = el(id);
        if(!c) return;
        var ids = layout[id];
        ids.forEach(function(cardId){
          var card = findOvCard(cardId);
          if(card) c.appendChild(card);
        });
      });
    } catch(err){}
  }
  function findOvCard(id){
    var found = null;
    ['ovKpiMain','ovKpiSub','ovCol0','ovCol1','ovCol2'].forEach(function(cid){
      var c = el(cid);
      if(!c) return;
      qa('[data-ov-card-id="'+id+'"]', c).forEach(function(n){ found = n; });
    });
    return found;
  }

  // ===== 添加自定义卡片弹窗 =====
  function openAddCard(){
    el('ovAddTitleInput').value = '';
    el('ovAddContent').value = '';
    el('ovAddHint').style.opacity = '0';
    el('ovAddSlot').value = 'ovCol2';
    el('ovAddCardModal').classList.add('show');
    var done = false;
    function close(){ if(done) return; done = true; el('ovAddCardModal').classList.remove('show'); cleanup(); }
    function submit(){
      if(done) return;
      var title = el('ovAddTitleInput').value.trim();
      var content = el('ovAddContent').value;
      if(!title){ el('ovAddHint').textContent = '请填写卡片标题'; el('ovAddHint').style.opacity = '1'; el('ovAddTitleInput').focus(); return; }
      done = true;
      el('ovAddCardModal').classList.remove('show');
      cleanup();
      var id = 'custom-' + Date.now().toString(36) + Math.random().toString(36).slice(2,7);
      var slot = el('ovAddSlot').value;
      var custom = getCustomCards();
      custom[id] = { title: title, content: content, slot: slot };
      setCustomCards(custom);
      setCardConf(id, { draggable: true, enabled: true, custom: true });
      // 立即渲染到对应容器
      renderCustomCards();
      initOverviewDrag();
      applyCardVisibility();
      saveOverviewLayout();
      toast('已添加卡片：' + title, true);
      loadOverview();
    }
    function cleanup(){
      el('ovAddCancel').removeEventListener('click', close);
      el('ovAddOk').removeEventListener('click', submit);
      el('ovAddCardModal').removeEventListener('click', overlayClick);
      el('ovAddTitleInput').removeEventListener('keydown', onKey);
    }
    function overlayClick(e){ if(e.target === el('ovAddCardModal')) close(); }
    function onKey(e){ if(e.key === 'Escape') close(); }
    el('ovAddCancel').addEventListener('click', close);
    el('ovAddOk').addEventListener('click', submit);
    el('ovAddCardModal').addEventListener('click', overlayClick);
    el('ovAddTitleInput').addEventListener('keydown', onKey);
    setTimeout(function(){ el('ovAddTitleInput').focus(); }, 30);
  }

  // ===== 卡片设置弹窗 =====
  function openOverviewSettings(){
    var listEl = el('ovSettingsList');
    listEl.innerHTML = '';
    var rows = [];
    // 内置 KPI 卡
    var kpiIds = ['kpi-users','kpi-sessions','kpi-files','kpi-storage','kpi-admins','kpi-datadir','kpi-go'];
    kpiIds.forEach(function(id){
      rows.push({ id:id, name:BUILTIN_CARD_NAMES[id], custom:false });
    });
    // 内置主体卡
    var bodyIds = ['user-dist','server-status','active-sessions','last-update','storage-ring','file-ext','top-files'];
    bodyIds.forEach(function(id){
      rows.push({ id:id, name:BUILTIN_CARD_NAMES[id], custom:false });
    });
    // 自定义卡
    var custom = getCustomCards();
    Object.keys(custom).forEach(function(id){
      rows.push({ id:id, name:custom[id].title || '备注', custom:true });
    });

    if(!rows.length){
      listEl.innerHTML = '<div class="ov-empty">暂无卡片</div>';
    }
    rows.forEach(function(r){
      var conf = cardConf(r.id);
      var slotName = slotLabel(getCustomSlot(r.id));
      var row = document.createElement('div');
      row.className = 'ov-set-row';
      row.innerHTML =
        '<div class="ov-set-info"><b>'+esc(r.name)+'</b>'+
        '<span class="ov-set-sub">'+slotName+(r.custom?' · 自定义':'')+'</span></div>'+
        '<div class="ov-set-ctrl">'+
          '<label class="ov-set-switch-wrap" title="开启后该卡片可被拖拽摆放">'+
            '<span>拖动</span><input type="checkbox" class="ov-set-switch" data-kind="drag" data-id="'+r.id+'"'+(conf.draggable?' checked':'')+'>'+
          '</label>'+
          '<label class="ov-set-switch-wrap" title="开启后该卡片在概览页显示">'+
            '<span>显示</span><input type="checkbox" class="ov-set-switch" data-kind="show" data-id="'+r.id+'"'+(conf.enabled?' checked':'')+'>'+
          '</label>'+
          (r.custom ? '<button class="btn btn-ghost btn-sm ov-set-del" data-id="'+r.id+'" title="删除此自定义卡片">🗑</button>' : '')+
        '</div>';
      listEl.appendChild(row);
    });

    // 事件：拖动/显示开关
    qa('.ov-set-switch', listEl).forEach(function(sw){
      sw.onchange = function(){
        var id = sw.getAttribute('data-id');
        var kind = sw.getAttribute('data-kind');
        if(kind === 'drag'){ setCardConf(id, { draggable: sw.checked }); }
        else { setCardConf(id, { enabled: sw.checked }); }
        applyCardVisibility();
      };
    });
    // 事件：删除自定义卡
    qa('.ov-set-del', listEl).forEach(function(btn){
      btn.onclick = function(){
        var id = btn.getAttribute('data-id');
        confirmDialog({
          title:'删除卡片',
          message:'确定删除自定义卡片「<b style="color:var(--cyan)">'+esc((getCustomCards()[id]||{}).title||'')+'</b>」吗？此操作不可恢复。',
          okText:'删除',
          onOk:function(){
            var custom = getCustomCards(); delete custom[id]; setCustomCards(custom);
            var cfg = getCardConfig(); delete cfg[id]; setCardConfig(cfg);
            var card = findOvCard(id); if(card && card.parentNode) card.parentNode.removeChild(card);
            saveOverviewLayout();
            openOverviewSettings();
            toast('已删除卡片', true);
          }
        });
      };
    });

    el('ovSettings').classList.add('show');
    el('ovSettingsClose').onclick = function(){ el('ovSettings').classList.remove('show'); };
    el('ovSettings').onclick = function(e){ if(e.target === el('ovSettings')) el('ovSettings').classList.remove('show'); };
  }
  function getCustomSlot(id){
    var custom = getCustomCards()[id];
    return custom && custom.slot ? custom.slot : '';
  }
  function slotLabel(id){
    var m = { ovKpiMain:'主 KPI 行', ovKpiSub:'副 KPI 行', ovCol0:'左列', ovCol1:'中列', ovCol2:'右列' };
    return m[id] || '概览';
  }

  function kpiCell(id, lbl, big, sub, tone, go){
    return '<div class="ov-kpi-cell tone-'+(tone||'cyan')+(go?' clickable':'')+'" draggable="true" data-ov-card-id="'+id+'"'+(go?' data-go="'+go+'"':'')+'>'+
      '<span class="ov-drag-handle" title="拖动调整卡片位置">⋮⋮</span>'+
      '<div class="ov-kpi-lbl">'+lbl+'</div>'+
      '<div class="ov-kpi-big">'+esc(big)+'</div>'+
      '<div class="ov-kpi-sub">'+esc(sub)+'</div>'+
    '</div>';
  }
  function renderUserBars(u){
    var total = Math.max(1, u.total);
    var rows = [
      { name:'管理员',   val:u.admins,    color:'var(--cyan)',   icon:'🛡' },
      { name:'普通用户', val:u.normal,    color:'var(--cyan2)',  icon:'👤' },
      { name:'已启用',   val:u.enabled,   color:'var(--ok)',     icon:'✓' },
      { name:'已禁用',   val:u.disabled,  color:'var(--err)',    icon:'⛔' },
      { name:'活跃会话', val:u.logged_in, color:'#a78bfa',       icon:'🟢' }
    ];
    el('ovUserBars').innerHTML = rows.map(function(r){
      var p = Math.round(r.val * 100 / total);
      return '<div class="ov-bar"><div class="ov-bar-lbl"><span>'+r.icon+' '+r.name+'</span><b>'+r.val+'</b></div>'+
        '<div class="ov-bar-track"><i class="ov-bar-fill" style="width:'+p+'%;background:'+r.color+'"></i></div></div>';
    }).join('');
    el('ovUserFoot').innerHTML = '<span>共 <b style="color:var(--cyan)">'+u.total+'</b> 个账户</span>';
  }
  function renderExtBars(list){
    if(!list.length){ el('ovExtBars').innerHTML = '<div class="ov-empty">数据目录为空，暂无文件类型</div>'; el('ovExtHint').textContent=''; return; }
    var top = list.slice(0, 6);
    var maxSize = top[0] ? top[0].size : 1;
    el('ovExtBars').innerHTML = top.map(function(e){
      var p = Math.round(e.size * 100 / maxSize);
      return '<div class="ov-bar mini"><div class="ov-bar-lbl"><span>.'+esc(e.ext)+'</span><b>'+fmtSize(e.size)+' · '+e.count+' 个</b></div>'+
        '<div class="ov-bar-track"><i class="ov-bar-fill" style="width:'+p+'%;background:linear-gradient(90deg,var(--cyan),var(--cyan2))"></i></div></div>';
    }).join('');
    el('ovExtHint').textContent = '共 ' + list.length + ' 种类型';
  }
  function renderTopFiles(list){
    if(!list.length){ el('ovTopFiles').innerHTML = '<div class="ov-empty">暂无文件</div>'; return; }
    el('ovTopFiles').innerHTML = list.map(function(f, i){
      return '<div class="ov-top-row"><span class="ov-top-i">#'+(i+1)+'</span>'+
        '<span class="ov-top-path mono" title="'+esc(f.path)+'">'+esc(f.path)+'</span>'+
        '<span class="ov-top-size mono">'+fmtSize(f.size)+'</span></div>';
    }).join('');
  }
  function renderServer(s){
    var rows = [
      ['FTP 控制端口',  s.ftp_port,  'cyan',  false],
      ['Web 管理端口',  s.web_port,  'cyan',  false],
      ['PASV 端口范围', s.pasv_ports,'cyan2', false],
      ['FTPS 加密',     s.ftps ? '已启用' : '未启用', s.ftps?'ok':'muted', false],
      ['数据目录',      s.data_dir,  '',      true]
    ];
    el('ovServer').innerHTML = rows.map(function(r){
      var cls = 'ov-kv-v' + (r[3] ? ' path' : ' mono') + (r[2] ? ' tone-'+r[2] : '');
      return '<div class="ov-kv-row"><span class="ov-kv-k">'+r[0]+'</span><span class="'+cls+'">'+esc(r[1])+'</span></div>';
    }).join('');
  }
  function renderLogged(u){
    var list = u.logged_list || [];
    el('ovLoggedHint').textContent = list.length + ' 个会话';
    if(!list.length){
      el('ovLogged').innerHTML = '<div class="ov-empty">当前无活跃会话</div>';
      return;
    }
    el('ovLogged').innerHTML = list.map(function(s){
      var tag = s.role === 'admin' ? '<span class="tag tag-adm">管理员</span>' : '<span class="tag tag-off">用户</span>';
      return '<div class="ov-sess"><div class="ov-sess-avatar">'+(s.username[0]||'?').toUpperCase()+'</div>'+
        '<div class="ov-sess-info"><div class="ov-sess-name">'+esc(s.username)+' '+tag+'</div>'+
        '<div class="ov-sess-time mono">'+esc(s.login_at)+' · '+esc(s.login_ago)+' 前</div></div></div>';
    }).join('');
  }
  function renderStorage(s){
    if(!s || !s.total){ el('ovUsedPct').textContent = '—'; el('ovStorage').innerHTML = '<div class="ov-empty">无法获取磁盘容量</div>'; return; }
    var usedPct = s.used_pct || 0;
    // 环形动画
    var c = el('ovRing');
    if(c){
      var len = 2 * Math.PI * 50; // ≈314.16
      c.setAttribute('stroke-dasharray', len.toFixed(2));
      c.setAttribute('stroke-dashoffset', (len * (1 - usedPct/100)).toFixed(2));
    }
    el('ovUsedPct').textContent = usedPct.toFixed(1);
    // 色调
    var ring = el('ovRing');
    if(ring){
      ring.style.stroke = usedPct > 85 ? 'var(--err)' : usedPct > 60 ? 'var(--warn)' : 'var(--cyan)';
    }
    el('ovStorage').innerHTML = [
      ['总容量',  fmtSize(s.total),       false],
      ['已使用',  fmtSize(s.used),        false],
      ['剩余',    fmtSize(s.free),        false],
      ['数据目录', s.path || '—',         true]
    ].map(function(r){
      var isPath = r[2];
      var cls = 'ov-kv-v' + (isPath ? ' path' : ' mono');
      if(!isPath){
        if(r[0]==='剩余') cls += ' tone-ok';
        else if(r[0]==='已使用') cls += ' tone-warn';
      }
      return '<div class="ov-kv-row"><span class="ov-kv-k">'+r[0]+'</span><span class="'+cls+'">'+esc(r[1])+'</span></div>';
    }).join('');
  }
  function renderLoad(l){
    // 运行时负载信息已并入 KPI 卡（kpi-go），无独立容器时静默跳过
    var target = el('ovLoad');
    if(!target) return;
    target.innerHTML = [
      ['协程数', l.goroutines],
      ['Go 版本', l.go_version],
      ['健康度', l.goroutines < 1000 ? '良好' : (l.goroutines < 5000 ? '正常' : '关注')]
    ].map(function(r){
      var tone = r[0]==='健康度' ? (r[1]==='良好'?'tone-ok':(r[1]==='正常'?'tone-warn':'tone-err')) : 'tone-cyan2';
      return '<div class="ov-kv-row"><span class="ov-kv-k">'+r[0]+'</span><span class="ov-kv-v mono '+tone+'">'+esc(r[1])+'</span></div>';
    }).join('');
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
    // 立即快照文件数组，避免异步回调期间 FileList 失效
    var fileArr = Array.prototype.slice.call(files);
    // 上传前先检查目标目录是否存在同名文件（非目录）
    api('/api/files?path=' + encodeURIComponent(state.curPath)).then(function(x){
      var existing = [];
      if(x.ok) existing = (x.d.items || []).filter(function(f){ return !f.is_dir; }).map(function(f){ return f.name; });
      var selected = fileArr.map(function(f){ return f.name; });
      // 找出同名冲突（去重）
      var conflicts = [];
      var seen = {};
      selected.forEach(function(n){
        if(existing.indexOf(n) !== -1 && !seen[n]){ seen[n] = 1; conflicts.push(n); }
      });
      if(conflicts.length){
        showUploadConflict(fileArr, conflicts);
      } else {
        doUpload(fileArr, 'rename');
      }
    });
  }

  // 展示上传同名冲突弹窗，让用户选择处理方式
  function showUploadConflict(files, conflicts){
    el('ucList').innerHTML = conflicts.map(function(n){ return '<li>'+esc(n)+'</li>'; }).join('');
    el('uploadConflict').classList.add('show');
    var done = false;
    function close(){ if(done) return; done = true; el('uploadConflict').classList.remove('show'); cleanup(); }
    function pick(mode){
      if(done) return; done = true;
      el('uploadConflict').classList.remove('show');
      cleanup();
      if(mode === 'cancel') return; // 取消上传
      doUpload(files, mode);
    }
    function cleanup(){
      el('ucOverwrite').removeEventListener('click', function(){});
      el('ucRename').removeEventListener('click', function(){});
      el('ucCancel').removeEventListener('click', function(){});
      el('uploadConflict').removeEventListener('click', overlayClose);
    }
    function overlayClose(e){ if(e.target === el('uploadConflict')) pick('cancel'); }
    el('ucOverwrite').onclick = function(){ pick('overwrite'); };
    el('ucRename').onclick = function(){ pick('rename'); };
    el('ucCancel').onclick = function(){ pick('cancel'); };
    el('uploadConflict').addEventListener('click', overlayClose);
  }

  // 执行上传（携带 mode）
  function doUpload(files, mode){
    var fd = new FormData(); fd.append('path', state.curPath); fd.append('mode', mode);
    qa('#fileInput').forEach(function(){ }); // noop
    for(var i=0;i<files.length;i++) fd.append('files', files[i]);
    el('mask').classList.add('show');
    var xhr = new XMLHttpRequest();
    xhr.open('POST', '/api/upload');
    xhr.withCredentials = true;
    xhr.upload.onprogress = function(e){ if(e.lengthComputable){ var pct = Math.round(e.loaded/e.total*100); el('pbar').style.width = pct+'%'; el('pmeta').textContent = pct+'%'; } };
    xhr.onload = function(){ el('mask').classList.remove('show'); el('pbar').style.width='0'; el('pmeta').textContent='0%';
      try{ var d = JSON.parse(xhr.responseText);
        if(d.ok || xhr.status===200){
          var msg = '上传完成';
          if(d.overwritten) msg += '（覆盖 ' + d.overwritten + ' 个文件）';
          else if(d.renamed && d.renamed.length) msg += '（重命名 ' + d.renamed.length + ' 个文件）';
          toast(msg, true); loadFiles();
        } else toast(d.error||'上传失败', false);
      }
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

  // ===== 系统设置（合并原基础设置 + 系统配置） =====
  function loadSettings(){
    api('/api/config').then(function(x){
      if(!x.ok){ toast(x.d.error||'加载失败', false); return; }
      state.cfg = x.d; var c = x.d;
      el('cfgFtpPort').value = c.ftp_port || '';
      el('cfgWebPort').value = c.web_port || '';
      // PASV 端口范围：后端返回 pasv_min_port/pasv_max_port 两个字段，前端组合为 "min-max"
      if(c.pasv_min_port && c.pasv_max_port){
        el('cfgPasv').value = c.pasv_min_port + '-' + c.pasv_max_port;
      } else {
        el('cfgPasv').value = '';
      }
      el('cfgFtps').checked = !!c.enable_ftps;
      el('cfgDataDir').value = c.data_dir || '';
      el('cfgRegister').checked = !!c.allow_register;
      el('cfgTLSCert').value = c.tls_cert || '';
      el('cfgTLSKey').value = c.tls_key || '';
    });
  }
  function saveSettings(){
    var body = {
      ftp_port: parseInt(el('cfgFtpPort').value,10),
      web_port: parseInt(el('cfgWebPort').value,10),
      enable_ftps: el('cfgFtps').checked,
      tls_cert: el('cfgTLSCert').value.trim(),
      tls_key: el('cfgTLSKey').value.trim(),
      data_dir: el('cfgDataDir').value.trim(),
      allow_register: el('cfgRegister').checked
    };
    // PASV 范围 "min-max" 拆分为两个端口字段提交
    var pasv = (el('cfgPasv').value || '').trim();
    if(pasv){
      var mm = pasv.split('-');
      var min = parseInt(mm[0], 10), max = parseInt(mm[1] != null ? mm[1] : mm[0], 10);
      if(min > 0 && max >= min){
        body.pasv_min_port = min;
        body.pasv_max_port = max;
      }
    }
    api('/api/config', { method:'PUT', body: JSON.stringify(body) }).then(function(x){
      if(x.ok){
        toast(x.d.message || '设置已保存并即时生效', true);
        loadSettings();
        loadOverview();
      } else toast(x.d.error||'保存失败', false);
    });
  }

  // ===== 关于 =====
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
        el('saveSettingsBtn').onclick = saveSettings;
      }
      loadOverview();
    });
  }
  if(document.readyState !== 'loading') init(); else document.addEventListener('DOMContentLoaded', init);
})();
