// ---------- 静态模式（GitHub Pages PWA 版）----------
// 本地 App 跑在 127.0.0.1，有 Go 后端；部署到 GitHub Pages 后没有后端：
//   GET /api/xxx?a=b  →  改读构建期快照 api/xxx__a_b.json（映射规则见 export_static.go）
//   写操作（POST/DELETE）→  返回统一的"网页版只读"提示，不发请求
(function () {
  // __TS_STATIC__ 由 Pages 构建流程注入 index.html（见 .github/workflows/pages.yml）
  if (!window.__TS_STATIC__) return;

  const origFetch = window.fetch.bind(window);
  const stub = (msg) => Promise.resolve(new Response(
    JSON.stringify({ ok: false, static: true, error: msg }),
    { status: 200, headers: { 'Content-Type': 'application/json' } }
  ));

  const jsonResponse = (obj) => new Response(JSON.stringify(obj),
    { status: 200, headers: { 'Content-Type': 'application/json' } });

  window.fetch = function (input, init) {
    const raw = typeof input === 'string' ? input : input.url;
    if (!raw.startsWith('/api/')) return origFetch(input, init);
    const method = ((init && init.method) || 'GET').toUpperCase();

    // 已通过 File System Access API 连接本地目录 → 部分接口由 fs-local.js 在浏览器内实现
    const fs = window.__TS_FS__;
    if (fs && fs.connected) {
      const routed = fs.route(method, raw, init && init.body);
      if (routed) {
        return routed.then((d) => d === null
          ? (method === 'GET' ? origFetch(input, init) : stub('网页预览版为只读，安装 / 写入操作请下载本地 App 使用'))
          : jsonResponse(d))
          .catch((e) => jsonResponse({ ok: false, error: '本地目录操作失败：' + e }));
      }
    }

    if (method !== 'GET') {
      // 安装类操作 → 直接给出终端一键安装命令（浏览器禁止网页写 ~/.claude 等隐藏目录）
      const u0 = new URL(raw, location.href);
      if (/\/api\/(install|unit-test\/install|commands\/install|orangecat\/install|code-review\/install|dev-dna\/install)$/.test(u0.pathname)) {
        const cmd = 'curl -fsSL ' + new URL('install.sh', location.href).href + ' | sh';
        return stub('浏览器无法写入本地目录。请复制下面命令到终端运行（一条装齐 Claude/Cursor/Codex 三端规范）：\n\n' + cmd);
      }
      return stub('网页预览版为只读，此操作请使用本地 App');
    }
    const u = new URL(raw, location.href);
    let key = u.pathname.replace(/^\/api\//, '');
    if (u.search) key += '__' + u.search.slice(1).replace(/[^A-Za-z0-9._-]/g, '_');
    return origFetch('api/' + key + '.json').then((r) =>
      r.ok ? r : stub('该数据在网页版不可用（仅本地 App 提供）')
    );
  };
})();

(function () {
  'use strict';

  // ---------- 键盘快捷键（WKWebView 默认不转发 Cmd+C/V/X/A，手动加）----------
  document.addEventListener('keydown', function (e) {
    const isCmd = e.metaKey || e.ctrlKey;
    if (!isCmd) return;
    const k = e.key.toLowerCase();
    const el = document.activeElement;
    const inField = el && (el.tagName === 'INPUT' || el.tagName === 'TEXTAREA' || el.isContentEditable);

    if (k === 'a') {
      if (inField) { el.select(); e.preventDefault(); }
      return;
    }
    if (k === 'c') {
      if (inField && el.selectionStart !== el.selectionEnd) {
        const text = el.value.substring(el.selectionStart, el.selectionEnd);
        (navigator.clipboard?.writeText(text) || Promise.reject()).catch(() => {
          document.execCommand('copy');
        });
        e.preventDefault();
      }
      return;
    }
    if (k === 'x') {
      if (inField && el.selectionStart !== el.selectionEnd) {
        const text = el.value.substring(el.selectionStart, el.selectionEnd);
        (navigator.clipboard?.writeText(text) || Promise.reject()).catch(() => {});
        el.value = el.value.substring(0, el.selectionStart) + el.value.substring(el.selectionEnd);
        el.dispatchEvent(new Event('input'));
        e.preventDefault();
      }
      return;
    }
    if (k === 'v') {
      if (inField && navigator.clipboard?.readText) {
        navigator.clipboard.readText().then((text) => {
          const start = el.selectionStart, end = el.selectionEnd;
          el.value = el.value.substring(0, start) + text + el.value.substring(end);
          el.selectionStart = el.selectionEnd = start + text.length;
          el.dispatchEvent(new Event('input'));
        }).catch(() => { document.execCommand('paste'); });
        e.preventDefault();
      }
      return;
    }
    if (k === 'z') {
      if (inField) { document.execCommand(e.shiftKey ? 'redo' : 'undo'); e.preventDefault(); }
      return;
    }
    // v1.7.24: Cmd+Q 退出 App
    // 后端 main.go 已 w.Bind("appQuit", ...)，这里只缺触发
    // 不论焦点在哪都触发（macOS 标准行为：Cmd+Q 全局退出）
    if (k === 'q') {
      e.preventDefault();
      if (typeof window.appQuit === 'function') {
        window.appQuit();
      }
    }
  });


  let skills = [];

  // ---------- 导航切换 ----------
  document.querySelectorAll('.navitem').forEach((btn) => {
    btn.addEventListener('click', () => {
      document.querySelectorAll('.navitem').forEach((b) => b.classList.remove('active'));
      document.querySelectorAll('.tab-panel').forEach((p) => p.classList.remove('active'));
      btn.classList.add('active');
      document.getElementById('tab-' + btn.dataset.tab).classList.add('active');
      document.querySelector('.content').scrollTop = 0;
      if (btn.dataset.tab === 'custom') { loadCustomRules(); loadPresets(); }
    });
  });

  // ---------- Scope 切换 ----------
  document.querySelectorAll('input[name="scope"]').forEach((r) => {
    r.addEventListener('change', (e) => {
      document.getElementById('projectPathRow').style.display =
        e.target.value === 'project' ? 'block' : 'none';
    });
  });

  // 项目目录选择（v1.7.45：调 macOS 原生 osascript choose folder）
  const projectPathBrowseBtn = document.getElementById('projectPathBrowseBtn');
  if (projectPathBrowseBtn) {
    projectPathBrowseBtn.addEventListener('click', () => {
      projectPathBrowseBtn.disabled = true;
      projectPathBrowseBtn.textContent = '⏳';
      fetch('/api/scaffold/pick-folder?prompt=' + encodeURIComponent('选择要装到的项目根目录'), {
        method: 'POST',
      }).then((r) => r.json()).then((d) => {
        projectPathBrowseBtn.disabled = false;
        projectPathBrowseBtn.textContent = '📁 选目录';
        if (d.path) {
          document.getElementById('projectPath').value = d.path;
        }
        // 用户取消 → path 为空 → 啥也不做
      }).catch((e) => {
        projectPathBrowseBtn.disabled = false;
        projectPathBrowseBtn.textContent = '📁 选目录';
        alert('选目录失败：' + e);
      });
    });
  }

  // ---------- Segmented sub-tab 切换（安装页内部） ----------
  document.querySelectorAll('.seg').forEach((seg) => {
    const items = seg.querySelectorAll('.seg-item');
    items.forEach((it) => {
      it.addEventListener('click', () => {
        const target = it.dataset.sub;
        // 仅切换同一 tab-panel 内的 subpage
        const panel = it.closest('.tab-panel');
        if (!panel) return;
        items.forEach((x) => x.classList.remove('active'));
        it.classList.add('active');
        panel.querySelectorAll('.subpage').forEach((sp) => {
          sp.classList.toggle('active', sp.dataset.sub === target);
        });
        document.querySelector('.content').scrollTop = 0;
      });
    });
  });

  // ---------- 安装页 Hero + mini-tiles 数据填充 ----------
  function refreshInstallHero() {
    const heroNum   = document.getElementById('installHeroNum');
    const heroTitle = document.getElementById('installHeroTitle');
    const heroSub   = document.getElementById('installHeroSub');
    const skillsNum = document.getElementById('miniSkillsNum');
    const skillsDelta = document.getElementById('miniSkillsDelta');
    const verNum    = document.getElementById('miniVerNum');
    const customNum = document.getElementById('miniCustomNum');
    const customDelta = document.getElementById('miniCustomDelta');
    const cmdsNum   = document.getElementById('miniCmdsNum');
    const cmdsDelta = document.getElementById('miniCmdsDelta');
    if (!heroNum) return;

    // Skill 数量来自已加载的 catalog
    const skillCount = skills.length || 0;
    if (skillsNum) skillsNum.textContent = skillCount;
    if (skillsDelta) skillsDelta.textContent = skillCount > 0 ? '已就绪' : '未加载';

    // 版本：从 #version 拿
    const verEl = document.getElementById('version');
    if (verNum && verEl) verNum.textContent = verEl.textContent || 'v…';

    // 自定义规则数：直接 fetch /api/custom
    fetch('/api/custom').then((r) => r.json()).then((d) => {
      const n = (d.rules || []).length;
      if (customNum) customNum.textContent = n;
      if (customDelta) customDelta.textContent = n > 0 ? n + ' 条已激活' : '无自定义';
    }).catch(() => {});

    // 快捷命令数：fetch /api/commands/list
    fetch('/api/commands/list').then((r) => r.json()).then((d) => {
      const list = d.commands || d.list || [];
      const n = Array.isArray(list) ? list.length : 0;
      if (cmdsNum) cmdsNum.textContent = n;
      if (cmdsDelta) cmdsDelta.textContent = n > 0 ? n + ' 条可用' : '未安装';
    }).catch(() => {});

    // Hero：基于 skill 数 + 自定义数给一个总览
    if (heroNum && heroTitle && heroSub) {
      heroNum.textContent = skillCount > 0 ? '✓' : '–';
      heroTitle.textContent = skillCount > 0
        ? `${skillCount} 个 Skill · ${skills.reduce((a, s) => a + (s.references || []).length, 0) || 17} 个 references 都已就绪`
        : '尚未加载规范目录';
      heroSub.textContent = '点击下方"立即安装"覆盖到 Claude Code / Cursor。已装的 custom-* 文件不会丢。';
    }
  }
  // 首次加载完 catalog 后填一次；每次切到安装 tab 也刷一次
  setTimeout(refreshInstallHero, 400);
  document.querySelectorAll('.navitem[data-tab="install"]').forEach((b) => {
    b.addEventListener('click', () => setTimeout(refreshInstallHero, 50));
  });

  // ---------- 加载规范目录 ----------
  fetch('/api/catalog')
    .then((r) => r.json())
    .then((data) => {
      skills = data.skills || [];
      const verStr = 'v' + data.version;
      document.getElementById('version').textContent = verStr;
      document.getElementById('skillsCount').textContent = skills.length;
      renderSkills(skills);
      // 欢迎页：版本拿到后才能比对
      showWelcomeIfNeeded(verStr);
    })
    .catch((e) => {
      document.getElementById('skillsList').textContent = '加载失败：' + e;
    });

  // 把 catalog 里 string 形式的 trigger 推断为 pill 类型
  // 启发式：
  //   含 *.xxx / glob 形 → 文件
  //   含 / 开头（如 /tech-design）→ 主动
  //   scope === always → 始终
  //   其它（关键字 / 短语）→ 被动
  function classifyTrigger(t, scope) {
    const s = String(t).trim();
    if (/^\/\w/.test(s) || /^\/[a-z-]+$/i.test(s))                 return { cls: 'tg-active',  text: '主动 · ' + s };
    if (/[*\.]/.test(s) && /(\.\w+|\*)/.test(s) && s.length < 40)  return { cls: 'tg-glob',    text: '文件 · ' + s };
    return { cls: 'tg-passive', text: '被动 · ' + s };
  }

  function renderSkills(list) {
    const c = document.getElementById('skillsList');
    c.innerHTML = '';
    list.forEach((s) => {
      const card = document.createElement('div');
      card.className = 'card';
      const scopeLabel =
        s.scope === 'always' ? '永远生效' : s.scope === 'manual' ? '手动触发' : '按条件';

      // 触发 pill：scope=always 必加"始终注入"；其余按 trigger 字串分类
      const tags = [];
      if (s.scope === 'always') {
        tags.push({ cls: 'tg-always', text: '始终注入' });
      }
      (s.triggers || []).forEach((t) => tags.push(classifyTrigger(t, s.scope)));

      const tagsHTML = tags.length
        ? '<div class="trigger-tags">' +
          tags.map((t) =>
            `<span class="trigger-tag ${t.cls}"><span class="dot"></span>${escapeHTML(t.text)}</span>`
          ).join('') +
          '</div>'
        : '';

      // v1.7.43：规范模块只读，🗑 卸载入口移到「已装清单」tab
      card.innerHTML = `
        <div class="card-head">
          <span class="card-title">${escapeHTML(s.title)}</span>
          <span class="card-scope ${s.scope} legacy-hidden">${scopeLabel}</span>
        </div>
        <div class="card-desc">${escapeHTML(s.description)}</div>
        ${tagsHTML}
      `;
      card.addEventListener('click', () => openSkillModal(s));
      c.appendChild(card);
    });
  }

  // ---------- WKWebView confirm() 替代方案：两步式按钮确认 ----------
  // WKWebView 屏蔽原生 confirm/alert，调用直接返回 false，导致所有确认操作静默失败。
  // twoStep(btn, confirmLabel, cb)：
  //   第一次点击 → 按钮变红色提示文字，3s 超时自动复原
  //   第二次点击 → 执行 cb()
  const _twoStepState = new WeakMap();
  function twoStep(btn, confirmLabel, cb) {
    if (_twoStepState.get(btn)) {
      // 第二次点击 → 确认执行
      _twoStepState.set(btn, false);
      btn.textContent = btn._origText;
      btn.classList.remove('danger-active');
      clearTimeout(btn._twoStepTimer);
      cb();
      return;
    }
    // 第一次点击 → 进入确认态
    btn._origText = btn._origText || btn.textContent;
    _twoStepState.set(btn, true);
    btn.textContent = confirmLabel || '再点一次确认';
    btn.classList.add('danger-active');
    btn._twoStepTimer = setTimeout(() => {
      _twoStepState.set(btn, false);
      btn.textContent = btn._origText;
      btn.classList.remove('danger-active');
    }, 3000);
  }

  // ---------- 单 skill 卸载（v1.7.41 起，v1.7.43 移到「已装清单」用）----------
  function uninstallSkill(s) {
    if (!confirm(`卸载 ${s.title}？\n\n` +
                 `Source: ${s.source}\n` +
                 `ID: ${s.id}\n\n` +
                 `gsd-framework 会删整个 ~/.claude/skills/${s.id}/\n` +
                 `ref-auto 会删 references/${s.id}.md\n\n` +
                 `Cmd+Q 重启 Claude / Cursor 后生效。`)) return;

    fetch('/api/skill-uninstall', {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id: s.id, source: s.source }),
    }).then((r) => r.json()).then((res) => {
      let msg = res.ok ? '✓ 卸载完成\n' : '⚠ 部分失败\n';
      (res.removed || []).forEach((p) => msg += '  · 删 ' + p + '\n');
      (res.skipped || []).forEach((p) => msg += '  · ' + p + '\n');
      (res.failed || []).forEach((p) => msg += '  ✗ ' + p + '\n');
      if (res.hint) msg += '\n' + res.hint;
      if (res.error) msg += '\n✗ ' + res.error;
      alert(msg);
      // 刷新规范模块卡片列表
      fetch('/api/catalog').then((r) => r.json()).then((d) => {
        skills = d.skills || [];
        document.getElementById('skillsCount').textContent = skills.length;
        renderSkills(skills);
      });
    });
  }

  // ---------- 规范详情弹窗 ----------
  function openSkillModal(skill) {
    document.getElementById('modalTitle').textContent = skill.title;
    document.getElementById('modalContent').textContent = '加载中…';
    document.getElementById('modalOverlay').classList.add('show');

    // 三种 source 分别走不同 endpoint：
    //   hardcoded / ref-auto → /api/reference?file=<basename>（embedded 或同步过来的）
    //   gsd-framework        → /api/skill-disk-file?path=<absolute path>（磁盘上）
    let fetchURL;
    if (skill.source === 'gsd-framework') {
      document.getElementById('modalSub').textContent = skill.reference;
      fetchURL = '/api/skill-disk-file?path=' + encodeURIComponent(skill.reference);
    } else {
      document.getElementById('modalSub').textContent = 'references/' + skill.reference;
      fetchURL = '/api/reference?file=' + encodeURIComponent(skill.reference);
    }

    fetch(fetchURL)
      .then((r) => r.json())
      .then((d) => {
        document.getElementById('modalContent').textContent = d.content || d.error || '(无内容)';
        document.querySelector('#modalContent').scrollTop = 0;
      })
      .catch((e) => {
        document.getElementById('modalContent').textContent = '加载失败：' + e;
      });
  }

  function closeModal() { document.getElementById('modalOverlay').classList.remove('show'); }
  document.getElementById('modalClose').addEventListener('click', closeModal);
  document.getElementById('modalOverlay').addEventListener('click', (e) => {
    if (e.target === document.getElementById('modalOverlay')) closeModal();
  });
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') closeModal();
  });

  // ---------- 安装 ----------
  document.getElementById('installBtn').addEventListener('click', () => {
    const scope = document.querySelector('input[name="scope"]:checked').value;
    const target = document.querySelector('input[name="target"]:checked').value;
    const path = document.getElementById('projectPath').value.trim();

    if (scope === 'project' && !path) {
      showLog('installLog', '✗ 请先填写项目绝对路径', 'err');
      return;
    }

    const btn = document.getElementById('installBtn');
    btn.disabled = true;
    btn.textContent = '安装中…';
    showLog('installLog', '正在安装…', '');

    fetch('/api/install', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ scope, target, path }),
    })
      .then((r) => r.json())
      .then((res) => {
        btn.disabled = false;
        btn.textContent = '⚡ 立即安装（覆盖已有）';
        if (!res.ok) {
          showLog('installLog', '✗ 安装失败：' + (res.error || '未知错误'), 'err');
          return;
        }
        let msg = '✓ 安装完成（v' + res.version + '）\n';
        res.installed.forEach((line) => (msg += '  · ' + line + '\n'));
        if (res.warnings && res.warnings.length > 0) {
          msg += '\n⚠ 提示：\n';
          res.warnings.forEach((w) => (msg += '  · ' + w + '\n'));
        }
        msg += '\n下一步：';
        if (res.claude_dir) msg += '\n  · Claude Code 重启后 Skill 自动加载';
        if (res.cursor_dir) msg += '\n  · Cursor 重启后 rules 自动生效';
        if (res.codex_dir)  msg += '\n  · Codex 重启后 Skill 自动加载';
        showLog('installLog', msg, 'ok');
      })
      .catch((e) => {
        btn.disabled = false;
        btn.textContent = '⚡ 立即安装（覆盖已有）';
        showLog('installLog', '✗ 请求失败：' + e, 'err');
      });
  });

  // ---------- 生成分发 zip 到 Downloads ----------
  let lastSavedZip = null;
  // ---------- AI 人格 ----------
  function loadPersona() {
    fetch('/api/persona')
      .then((r) => r.json())
      .then((d) => {
        document.getElementById('personaContent').value = d.content || '';
        document.getElementById('personaContent').dataset.default = d.default || '';
        const ins = d.installed || {};
        const claudeEl = document.getElementById('personaClaudeStatus');
        const cursorEl = document.getElementById('personaCursorStatus');
        claudeEl.textContent = ins.claude_md ? '✓ 已配置' : '— 未配置';
        claudeEl.className = ins.claude_md ? 'ok' : 'no';
        cursorEl.textContent = ins.cursor_mdc ? '✓ 已配置' : '— 未配置';
        cursorEl.className = ins.cursor_mdc ? 'ok' : 'no';
        const personaCountEl = document.getElementById('personaCount');
        const activeCount = (ins.claude_md ? 1 : 0) + (ins.cursor_mdc ? 1 : 0);
        personaCountEl.textContent = activeCount;
      });
  }

  document.getElementById('personaSaveBtn').addEventListener('click', () => {
    const content = document.getElementById('personaContent').value;
    if (!content.trim()) { showLog('personaLog', '✗ 内容为空', 'err'); return; }
    showLog('personaLog', '正在写入...', '');
    fetch('/api/persona', {
      method: 'POST', headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ content }),
    })
      .then((r) => r.json())
      .then((res) => {
        if (!res.ok) { showLog('personaLog', '✗ ' + (res.error || '失败'), 'err'); return; }
        let msg = '✓ 已保存并同步\n';
        (res.installed || []).forEach((l) => msg += '  · ' + l + '\n');
        if (res.warnings && res.warnings.length) {
          msg += '\n⚠ ' + res.warnings.join('\n  ');
        }
        msg += '\n\n下一步：彻底重启 Claude Code / Cursor（Cmd+Q 后再开）';
        showLog('personaLog', msg, 'ok');
        loadPersona();
      });
  });
  document.getElementById('personaResetBtn').addEventListener('click', () => {
    const def = document.getElementById('personaContent').dataset.default;
    if (def) document.getElementById('personaContent').value = def;
    showLog('personaLog', '已填入默认内容，点保存生效', '');
  });

  // 两步式卸载（WKWebView 没 confirm）
  let personaRmConfirm = false;
  document.getElementById('personaRemoveBtn').addEventListener('click', (e) => {
    const btn = e.currentTarget;
    if (!personaRmConfirm) {
      personaRmConfirm = true;
      btn.textContent = '再点一次确认卸载';
      btn.classList.add('danger-active');
      setTimeout(() => {
        personaRmConfirm = false;
        btn.textContent = '🗑 全部卸载';
        btn.classList.remove('danger-active');
      }, 3000);
      return;
    }
    fetch('/api/persona', { method: 'DELETE' })
      .then((r) => r.json())
      .then((res) => {
        showLog('personaLog', '✓ 已卸载：\n' + (res.removed || []).map((x) => '  · ' + x).join('\n'), 'ok');
        personaRmConfirm = false;
        btn.textContent = '🗑 全部卸载';
        btn.classList.remove('danger-active');
        loadPersona();
      });
  });

  // 切到 persona tab 加载数据
  document.querySelectorAll('.navitem').forEach((btn) => {
    if (btn.dataset.tab === 'persona') btn.addEventListener('click', loadPersona);
  });
  loadPersona();


  // ---------- 已装清单 ----------
  function loadInstalled() {
    fetch('/api/installed')
      .then((r) => r.json())
      .then((d) => {
        document.getElementById('claudePath').textContent = d.claude_root;
        document.getElementById('cursorPath').textContent = d.cursor_root;
        const claudeList = document.getElementById('claudeList');
        const cursorList = document.getElementById('cursorList');
        const codexList  = document.getElementById('codexList');
        claudeList.innerHTML = '';
        cursorList.innerHTML = '';
        if (codexList) codexList.innerHTML = '';
        const total = (d.claude_skills || []).length + (d.cursor_rules || []).length + (d.codex_skills || []).length;
        document.getElementById('installedCount').textContent = total;

        if (!d.claude_skills || !d.claude_skills.length) {
          claudeList.innerHTML = '<div class="installed-empty">~/.claude/skills/ 下还没有装任何 Skill</div>';
        } else {
          d.claude_skills.forEach((item) => claudeList.appendChild(renderItem(item)));
        }
        if (!d.cursor_rules || !d.cursor_rules.length) {
          cursorList.innerHTML = '<div class="installed-empty">~/.cursor/rules/ 下还没有装任何规则</div>';
        } else {
          d.cursor_rules.forEach((item) => cursorList.appendChild(renderItem(item)));
        }
        if (codexList) {
          if (!d.codex_skills || !d.codex_skills.length) {
            codexList.innerHTML = '<div class="installed-empty">~/.codex/skills/ 下还没有装任何 Skill</div>';
          } else {
            d.codex_skills.forEach((item) => codexList.appendChild(renderItem(item)));
          }
        }
      });
  }

  function renderItem(item) {
    const box = document.createElement('div');
    box.className = 'installed-item ' + item.source;
    const size = formatSize(item.size);
    const when = new Date(item.modified).toLocaleString('zh-CN', { hour12: false });
    const sourceLabel = { team: '团队', custom: '自定义', other: '其他' }[item.source] || item.source;
    const desc = item.summary ? '<div class="installed-item-desc">' + escapeHTML(item.summary) + '</div>' : '';

    // v1.7.43: 加 🗑 卸载按钮（团队 hardcoded 的禁卸 —— 装在 DMG 里）
    // 团队 skill 走 source=team，本质是 ~/.claude/skills/{go-team-standards,orangecat,...}
    // 这些 4 个团队 skill 是 hardcoded，不允许这里直接卸（要在「⚡ 安装」对应卡片走）
    const teamHardcoded = ['go-team-standards', 'orangecat', 'dev-dna', 'code-review'];
    const canUninstall = !teamHardcoded.includes(item.name);

    const uninstallBtn = canUninstall
      ? `<button class="btn btn-ghost btn-sm danger" data-uninstall="${escapeHTML(item.name)}" data-path="${escapeHTML(item.path)}">🗑 卸载</button>`
      : '';

    box.innerHTML = `
      <div class="installed-item-head">
        <span class="installed-item-name">${escapeHTML(item.name)}</span>
        <span class="tag ${item.source}">${sourceLabel}</span>
      </div>
      ${desc}
      <div class="installed-item-meta">${size} · ${when}</div>
      <div class="installed-item-actions">
        <button class="btn btn-ghost btn-sm" data-reveal="${escapeHTML(item.path)}">🔍 在 Finder 中显示</button>
        ${uninstallBtn}
      </div>
    `;
    box.querySelector('button[data-reveal]').addEventListener('click', () => {
      fetch('/api/reveal', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: item.path }),
      });
    });
    const rmBtn = box.querySelector('button[data-uninstall]');
    if (rmBtn) {
      rmBtn.addEventListener('click', () => {
        const name = rmBtn.dataset.uninstall;
        const path = rmBtn.dataset.path;
        twoStep(rmBtn, `确认卸载「${name}」`, () => {
        fetch('/api/installed-skill-uninstall', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ path }),
        }).then((r) => r.json()).then((res) => {
          if (res.ok) {
            showLog('installedLog', '✓ 已卸载：' + name + '　Cmd+Q 重启 Claude / Cursor 生效', 'ok');
            loadInstalled();
          } else {
            const errMsg = res.error || '未知';
            const detail = res.detail ? '\n详情: ' + res.detail : '';
            showLog('installedLog', '✗ 卸载失败：' + errMsg + detail, 'err');
            console.error('uninstall failed:', res);
          }
        }).catch((e) => {
          showLog('installedLog', '✗ 网络/请求失败：' + e, 'err');
          console.error('uninstall fetch error:', e);
        });
        }); // twoStep end
      });
    }
    return box;
  }

  function formatSize(n) {
    if (n < 1024) return n + ' B';
    if (n < 1024 * 1024) return (n / 1024).toFixed(1) + ' KB';
    return (n / 1048576).toFixed(2) + ' MB';
  }

  document.getElementById('refreshInstalledBtn').addEventListener('click', loadInstalled);
  // 首次进入 Tab 加载
  document.querySelectorAll('.navitem').forEach((btn) => {
    if (btn.dataset.tab === 'installed') {
      btn.addEventListener('click', loadInstalled);
    }
  });
  // 启动时就加载一次，好给侧栏计数
  loadInstalled();

  // ---------- 提示词 Tab ----------
  // 一键复制：点卡片或复制按钮均触发
  document.querySelectorAll('.prompt-card').forEach((card) => {
    const copyBtn = card.querySelector('.prompt-copy-btn');
    const text = card.dataset.prompt || '';

    function doCopy() {
      navigator.clipboard.writeText(text).then(() => {
        card.classList.add('copied');
        if (copyBtn) copyBtn.textContent = '✓ 已复制';
        setTimeout(() => {
          card.classList.remove('copied');
          if (copyBtn) copyBtn.textContent = '复制';
        }, 1800);
      }).catch(() => {
        // fallback
        const ta = document.createElement('textarea');
        ta.value = text; ta.style.position = 'fixed'; ta.style.opacity = '0';
        document.body.appendChild(ta); ta.select();
        document.execCommand('copy');
        document.body.removeChild(ta);
        card.classList.add('copied');
        if (copyBtn) copyBtn.textContent = '✓ 已复制';
        setTimeout(() => { card.classList.remove('copied'); if (copyBtn) copyBtn.textContent = '复制'; }, 1800);
      });
    }

    card.addEventListener('click', (e) => {
      if (e.target === copyBtn) return; // 让按钮自己处理
      doCopy();
    });
    if (copyBtn) copyBtn.addEventListener('click', (e) => { e.stopPropagation(); doCopy(); });
  });

  // AGENTS.md 生成：选目录 + 写文件
  const agentsInitBrowse = document.getElementById('agentsInitBrowse');
  const agentsInitBtn    = document.getElementById('agentsInitBtn');
  const agentsInitPath   = document.getElementById('agentsInitPath');
  const agentsInitLog    = document.getElementById('agentsInitLog');

  if (agentsInitBrowse) {
    agentsInitBrowse.addEventListener('click', () => {
      agentsInitBrowse.disabled = true; agentsInitBrowse.textContent = '⏳';
      fetch('/api/scaffold/pick-folder?prompt=' + encodeURIComponent('选择项目根目录'), { method: 'POST' })
        .then((r) => r.json()).then((d) => {
          agentsInitBrowse.disabled = false; agentsInitBrowse.textContent = '📁 选目录';
          if (d.path) agentsInitPath.value = d.path;
        }).catch(() => { agentsInitBrowse.disabled = false; agentsInitBrowse.textContent = '📁 选目录'; });
    });
  }

  if (agentsInitBtn) {
    agentsInitBtn.addEventListener('click', () => {
      const p = agentsInitPath ? agentsInitPath.value.trim() : '';
      if (!p) { showLog('agentsInitLog', '✗ 请先选择项目根目录', 'err'); return; }
      agentsInitBtn.disabled = true; agentsInitBtn.textContent = '⏳';
      fetch('/api/agents-init', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: p }),
      }).then((r) => r.json()).then((res) => {
        agentsInitBtn.disabled = false; agentsInitBtn.textContent = '生成 AGENTS.md';
        showLog('agentsInitLog', res.ok
          ? (res.created ? '✓ 已创建：' + res.file : '⚠ 已存在（未覆盖）：' + res.file)
          : '✗ ' + (res.error || '未知错误'),
          res.ok ? 'ok' : 'err');
      }).catch((e) => {
        agentsInitBtn.disabled = false; agentsInitBtn.textContent = '生成 AGENTS.md';
        showLog('agentsInitLog', '✗ 请求失败：' + e, 'err');
      });
    });
  }

  // 通用 Agent 适配：选目录 + 一键生成各家 rules 入口文件
  const universalBrowse = document.getElementById('universalBrowse');
  const universalBtn    = document.getElementById('universalBtn');
  const universalPath   = document.getElementById('universalPath');

  if (universalBrowse) {
    universalBrowse.addEventListener('click', () => {
      universalBrowse.disabled = true; universalBrowse.textContent = '⏳';
      fetch('/api/scaffold/pick-folder?prompt=' + encodeURIComponent('选择项目根目录'), { method: 'POST' })
        .then((r) => r.json()).then((d) => {
          universalBrowse.disabled = false; universalBrowse.textContent = '📁 选目录';
          if (d.path) universalPath.value = d.path;
        }).catch(() => { universalBrowse.disabled = false; universalBrowse.textContent = '📁 选目录'; });
    });
  }

  if (universalBtn) {
    universalBtn.addEventListener('click', () => {
      const p = universalPath ? universalPath.value.trim() : '';
      if (!p) { showLog('universalLog', '✗ 请先选择项目根目录', 'err'); return; }
      universalBtn.disabled = true; universalBtn.textContent = '⏳';
      fetch('/api/universal/install', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path: p }),
      }).then((r) => r.json()).then((res) => {
        universalBtn.disabled = false; universalBtn.textContent = '一键适配全部 Agent';
        let msg = '';
        (res.created || []).forEach((f) => msg += '✓ 创建 ' + f + '\n');
        (res.skipped || []).forEach((f) => msg += '⚠ 跳过 ' + f + '\n');
        (res.failed || []).forEach((f) => msg += '✗ 失败 ' + f + '\n');
        if (res.error) msg += '✗ ' + res.error;
        showLog('universalLog', msg.trim() || '（无变化）', res.ok === false ? 'err' : 'ok');
      }).catch((e) => {
        universalBtn.disabled = false; universalBtn.textContent = '一键适配全部 Agent';
        showLog('universalLog', '✗ 请求失败：' + e, 'err');
      });
    });
  }

  // ---------- 脚手架 ----------
  const $$ = (id) => document.getElementById(id);

  function initScaffold() {
    // 项目名：强制 kebab-case 小写 —— 只允许 [a-z0-9-]，大写自动转小写，其他字符直接过滤
    $$('sfName').addEventListener('input', (e) => {
      const el = e.target;
      const caret = el.selectionStart;
      const cleaned = el.value.toLowerCase().replace(/[^a-z0-9-]/g, '');
      if (cleaned !== el.value) {
        el.value = cleaned;
        try { el.setSelectionRange(caret, caret); } catch (_) {}
      }
      const v = cleaned.trim();
      if (!v) return;
      const short = v.replace(/-(service|svc)$/, '');
      const biz = short.replace(/[-_]/g, '');
      if (!$$('sfBiz').dataset.manual) $$('sfBiz').value = biz;
      if (!$$('sfModule').dataset.manual) $$('sfModule').value = 'github.com/company/' + v;
    });

    // 业务域：更严格 —— 只允许 [a-z0-9]（无短横线），后端要求纯字母数字
    $$('sfBiz').addEventListener('input', (e) => {
      const el = e.target;
      const caret = el.selectionStart;
      const cleaned = el.value.toLowerCase().replace(/[^a-z0-9]/g, '');
      if (cleaned !== el.value) {
        el.value = cleaned;
        try { el.setSelectionRange(caret, caret); } catch (_) {}
      }
      el.dataset.manual = '1';
    });

    $$('sfModule').addEventListener('input', () => ($$('sfModule').dataset.manual = '1'));

    // 模板类型切换
    document.querySelectorAll('input[name="sfTplType"]').forEach((r) => {
      r.addEventListener('change', (e) => {
        const isGit = e.target.value === 'git';
        $$('sfTplLocalRow').style.display = isGit ? 'none' : '';
        $$('sfTplGitRow').style.display = isGit ? '' : 'none';
        if (isGit && !$$('sfRemoteURL').dataset.manual) {
          deriveRemote();
        }
      });
    });

    // 远端 URL 自动推导（根据模板 Git URL + 项目名）
    function deriveRemote() {
      if ($$('sfRemoteURL').dataset.manual) return;
      const tpl = $$('sfTplGitURL').value.trim();
      const name = $$('sfName').value.trim();
      if (!tpl || !name) return;
      fetch('/api/scaffold/derive-remote?template_url=' + encodeURIComponent(tpl) +
            '&project_name=' + encodeURIComponent(name))
        .then((r) => r.json())
        .then((d) => {
          if (d.remote_url && !$$('sfRemoteURL').dataset.manual) {
            $$('sfRemoteURL').value = d.remote_url;
          }
        });
    }
    $$('sfTplGitURL').addEventListener('input', deriveRemote);
    $$('sfName').addEventListener('input', () => {
      // 项目名变了也刷新一下（在现有 listener 之外）
      setTimeout(deriveRemote, 0);
    });
    $$('sfRemoteURL').addEventListener('input', () => ($$('sfRemoteURL').dataset.manual = '1'));

    $$('sfPickDir').addEventListener('click', () => {
      $$('sfPickDir').disabled = true;
      $$('sfPickDir').textContent = '选择中…';
      fetch('/api/scaffold/pick-folder', { method: 'POST' })
        .then((r) => r.json())
        .then((d) => {
          $$('sfPickDir').disabled = false;
          $$('sfPickDir').textContent = '📂 选择目录';
          if (d.path) $$('sfOutDir').value = d.path;
        })
        .catch(() => {
          $$('sfPickDir').disabled = false;
          $$('sfPickDir').textContent = '📂 选择目录';
        });
    });

    function checkTemplate() {
      const tpl = $$('sfTplPath').value.trim();
      const el = $$('sfTplStatus');
      el.className = 'tpl-status';
      el.textContent = '检查中…';
      fetch('/api/scaffold/check-template?path=' + encodeURIComponent(tpl))
        .then((r) => r.json())
        .then((d) => {
          if (d.ok) {
            el.className = 'tpl-status ok';
            el.textContent = '✓ 模板有效，module = ' + d.module;
          } else {
            el.className = 'tpl-status err';
            el.textContent = '✗ ' + d.msg;
          }
        });
    }
    $$('sfTplPath').addEventListener('blur', checkTemplate);

    // 模板路径浏览按钮
    const tplPickBtn = $$('sfTplPickBtn');
    if (tplPickBtn) {
      tplPickBtn.addEventListener('click', () => {
        tplPickBtn.disabled = true;
        tplPickBtn.textContent = '选择中…';
        fetch('/api/scaffold/pick-folder?prompt=' + encodeURIComponent('选择模板目录（mask-go-sample-service 或 fork 版本）'), { method: 'POST' })
          .then((r) => r.json())
          .then((d) => {
            tplPickBtn.disabled = false;
            tplPickBtn.textContent = '📁 浏览…';
            if (d.path) {
              $$('sfTplPath').value = d.path;
              checkTemplate();
            }
          })
          .catch(() => {
            tplPickBtn.disabled = false;
            tplPickBtn.textContent = '📁 浏览…';
          });
      });
    }

    // Skill 测试用例勾选时展开说明
    const genSkill = $$('sfGenSkillTest');
    const genSkillDetail = $$('sfGenSkillTestDetail');
    if (genSkill && genSkillDetail) {
      const syncDetail = () => { genSkillDetail.style.display = genSkill.checked ? 'block' : 'none'; };
      genSkill.addEventListener('change', syncDetail);
      syncDetail();
    }

    document.querySelectorAll('.navitem').forEach((btn) => {
      if (btn.dataset.tab === 'scaffold') {
        btn.addEventListener('click', () => {
          if (!$$('sfTplStatus').dataset.checked) {
            $$('sfTplStatus').dataset.checked = '1';
            checkTemplate();
          }
        });
      }
    });

    // 组件依赖联动：勾 Alarm → 强制 Kafka 勾 + 锁；勾 Cron → 强制 PG 勾 + 锁
    function updateComponentLocks() {
      const alarm = $$('compAlarm').checked;
      const cron = $$('compCron').checked;
      const kafkaEl = $$('compKafka');
      const pgEl = $$('compPG');
      const kafkaLabel = kafkaEl.closest('.component-item');
      const pgLabel = pgEl.closest('.component-item');

      if (alarm) {
        kafkaEl.checked = true; kafkaEl.disabled = true;
        kafkaLabel.classList.add('locked');
      } else {
        kafkaEl.disabled = false;
        kafkaLabel.classList.remove('locked');
      }
      if (cron) {
        pgEl.checked = true; pgEl.disabled = true;
        pgLabel.classList.add('locked');
      } else {
        pgEl.disabled = false;
        pgLabel.classList.remove('locked');
      }
    }
    ['compPG', 'compRedis', 'compKafka', 'compAlarm', 'compCron'].forEach((id) => {
      const el = $$(id);
      if (el) el.addEventListener('change', updateComponentLocks);
    });
    updateComponentLocks();

    $$('sfCreateBtn').addEventListener('click', () => {
      const tplType = document.querySelector('input[name="sfTplType"]:checked').value;
      const body = {
        project_name: $$('sfName').value.trim(),
        biz: $$('sfBiz').value.trim(),
        module: $$('sfModule').value.trim(),
        http_port: parseInt($$('sfHTTP').value, 10) || 8000,
        grpc_port: parseInt($$('sfGRPC').value, 10) || 9000,
        author: $$('sfAuthor').value.trim(),
        output_dir: $$('sfOutDir').value.trim(),
        template_type: tplType,
        template_path: $$('sfTplPath').value.trim(),
        template_git_url: $$('sfTplGitURL').value.trim(),
        template_branch: $$('sfTplBranch').value.trim(),
        remote_git_url: $$('sfRemoteURL').value.trim(),
        gitlab_api_url: $$('sfGitLabAPI').value.trim(),
        gitlab_group: $$('sfGitLabGroup').value.trim(),
        gitlab_token: $$('sfGitLabToken').value.trim(),
        skeleton_only: $$('sfSkeleton').checked,
        init_git: $$('sfGit').checked,
        run_tidy: $$('sfTidy').checked,
        gen_skill_test_cases: $$('sfGenSkillTest') ? $$('sfGenSkillTest').checked : false,
        components: {
          postgres: $$('compPG').checked,
          redis: $$('compRedis').checked,
          kafka: $$('compKafka').checked,
          alarm: $$('compAlarm').checked,
          cron: $$('compCron').checked,
        },
      };
      if (!body.project_name || !body.module || !body.biz || !body.output_dir) {
        showLog('sfLog', '✗ 项目名 / Module / 业务域 / 输出目录 不能为空', 'err');
        return;
      }
      const btn = $$('sfCreateBtn');
      btn.disabled = true;
      btn.textContent = '生成中（30-60 秒）…';
      showLog('sfLog', '正在复制模板 + 替换占位符…', '');
      fetch('/api/scaffold/create', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      })
        .then((r) => r.json())
        .then((res) => {
          btn.disabled = false;
          btn.textContent = '🚀 开始生成';
          if (!res.ok) {
            showLog('sfLog', '✗ 生成失败：' + (res.error || '未知错误'), 'err');
            return;
          }
          let msg = '✓ 项目已生成：\n  ' + res.dest_path + '\n\n日志：\n' +
            (res.logs || []).map((l) => '  ' + l).join('\n');
          if (res.warnings && res.warnings.length) {
            msg += '\n\n⚠ 需注意：\n' + res.warnings.map((w) => '  ' + w).join('\n');
          }
          msg += '\n\n下一步：看项目根目录的「下一步.md」';
          const el = $$('sfLog');
          el.classList.remove('err');
          el.classList.add('ok', 'show');
          el.innerHTML = escapeHTML(msg).replace(/\n/g, '<br>') +
            '<br><br><button class="btn btn-ghost btn-sm" id="sfRevealBtn">🔍 在 Finder 中显示</button>';
          $$('sfRevealBtn').addEventListener('click', () => {
            fetch('/api/reveal', {
              method: 'POST',
              headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ path: res.dest_path }),
            });
          });
        })
        .catch((e) => {
          btn.disabled = false;
          btn.textContent = '🚀 开始生成';
          showLog('sfLog', '✗ 请求失败：' + e, 'err');
        });
    });
  }
  initScaffold();

  // ---------- 下载分发包（DMG 推荐 / zip 兼容）----------
  function savePackage(kind, btn, origText) {
    btn.disabled = true;
    btn.textContent = '打包中（约 10 秒）…';
    const kindLabel = kind === 'dmg' ? 'DMG 磁盘镜像' : 'zip 压缩包';
    showLog('downloadLog', '正在打包 ' + kindLabel + '…', '');
    const endpoint = kind === 'dmg' ? '/api/save-dmg'
                    : kind === 'shell' ? '/api/save-shell'
                    : '/api/save-zip';
    fetch(endpoint, { method: 'POST' })
      .then((r) => r.json())
      .then((res) => {
        btn.disabled = false;
        btn.textContent = origText;
        if (!res.ok) {
          showLog('downloadLog', '✗ 打包失败：' + (res.error || '未知错误'), 'err');
          return;
        }
        const sizeMB = (res.size / 1048576).toFixed(1);
        const el = document.getElementById('downloadLog');
        el.classList.remove('err');
        el.classList.add('ok', 'show');
        const steps = kind === 'dmg'
          ? '  1. 双击 .dmg 挂载\n  2. 拖 🦕 Team Standards.app 到 Applications 图标\n  3. 弹出 DMG\n  4. 在 /Applications 双击启动 → 点「立即安装」'
          : '  1. 解压 zip\n  2. 双击 Team Standards.app\n  3. 点「立即安装」';
        el.innerHTML =
          '✓ ' + kindLabel + ' 打包完成（' + sizeMB + ' MB）\n\n' +
          '文件名：<strong>' + escapeHTML(res.name) + '</strong>\n' +
          '位置：<code>' + escapeHTML(res.path) + '</code>\n\n' +
          '<button class="btn btn-ghost btn-sm" id="revealPkgBtn">🔍 在 Finder 中显示</button>\n\n' +
          '对方收到后：\n' + steps;
        document.getElementById('revealPkgBtn').addEventListener('click', () => {
          fetch('/api/reveal', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ path: res.path }),
          });
        });
      })
      .catch((e) => {
        btn.disabled = false;
        btn.textContent = origText;
        showLog('downloadLog', '✗ 请求失败：' + e, 'err');
      });
  }
  document.getElementById('downloadDMGBtn').addEventListener('click', (e) => {
    savePackage('dmg', e.currentTarget, '💽 生成 DMG（推荐）');
  });
  document.getElementById('downloadZipBtn').addEventListener('click', (e) => {
    savePackage('zip', e.currentTarget, '📦 生成 zip（兼容）');
  });
  document.getElementById('downloadShellBtn').addEventListener('click', (e) => {
    savePackage('shell', e.currentTarget, '📜 生成 Shell（终极）');
  });

  // ---------- OrangeCat 提测文档 Skill 独立管理 ----------
  function loadOrangecatStatus() {
    fetch('/api/orangecat/status')
      .then((r) => r.json())
      .then((d) => {
        const el = document.getElementById('orangecatStatus');
        const c = d.claude_installed, cu = d.cursor_installed;
        const custom = d.custom_template_used ? '（使用自定义模板）' : '';
        if (c && cu) {
          el.className = 'tixuebj-status installed';
          el.textContent = '✓ 已装（Claude + Cursor 双侧）· ' + (d.version || '未知版本').trim() + ' ' + custom;
        } else if (c || cu) {
          el.className = 'tixuebj-status partial';
          el.textContent = '⚠ 只装了一侧：' + (c ? 'Claude ✓' : 'Claude ✗') + ' / ' + (cu ? 'Cursor ✓' : 'Cursor ✗');
        } else {
          el.className = 'tixuebj-status missing';
          el.textContent = '○ 未安装（点下面"独立安装"即可）';
        }
      });
  }

  document.getElementById('orangecatInstallBtn').addEventListener('click', () => {
    const btn = document.getElementById('orangecatInstallBtn');
    btn.disabled = true; btn.textContent = '安装中…';
    fetch('/api/orangecat/install', { method: 'POST' })
      .then((r) => r.json())
      .then((res) => {
        btn.disabled = false; btn.textContent = '📥 独立安装';
        let msg = res.ok ? '✓ 安装完成\n' : '✗ 部分失败\n';
        (res.installed || []).forEach((l) => msg += '  · ' + l + '\n');
        (res.warnings || []).forEach((w) => msg += '  ⚠ ' + w + '\n');
        msg += '\n⚠️ 必做：Cmd+Q 彻底重启 Claude Code / Cursor 才生效';
        showLog('orangecatLog', msg, res.ok ? 'ok' : 'err');
        loadOrangecatStatus();
        showRestartBanner(null);
      });
  });

  let orangecatRmConfirm = false;
  document.getElementById('orangecatUninstallBtn').addEventListener('click', (e) => {
    const btn = e.currentTarget;
    if (!orangecatRmConfirm) {
      orangecatRmConfirm = true;
      btn.textContent = '再点一次确认卸载';
      btn.classList.add('danger-active');
      setTimeout(() => {
        orangecatRmConfirm = false;
        btn.textContent = '🗑 卸载';
        btn.classList.remove('danger-active');
      }, 3000);
      return;
    }
    fetch('/api/orangecat/uninstall', { method: 'POST' })
      .then((r) => r.json())
      .then((res) => {
        orangecatRmConfirm = false;
        btn.textContent = '🗑 卸载';
        btn.classList.remove('danger-active');
        showLog('orangecatLog', '✓ 已卸载\n' + (res.removed || []).map((x) => '  · ' + x).join('\n'), 'ok');
        loadOrangecatStatus();
      });
  });

  // v1.7.22 通用辅助：调 save-zip 落盘 + 在指定 logEl 渲染 path + Finder/Claude 按钮
  function doSkillSaveZipFlow(skillName, logElId, btn) {
    const orig = btn ? btn.textContent : null;
    if (btn) { btn.disabled = true; btn.textContent = '导出中…'; }
    showLog(logElId, '正在打包 ' + skillName + '…', '');
    fetch('/api/claude-desktop/save-zip?name=' + encodeURIComponent(skillName), { method: 'POST' })
      .then((r) => r.json())
      .then((res) => {
        if (btn) { btn.disabled = false; btn.textContent = orig; }
        if (!res.ok) {
          showLog(logElId, '✗ 失败：' + (res.error || res.detail || '未知'), 'err');
          return;
        }
        const sizeKB = (res.size / 1024).toFixed(1);
        const log = document.getElementById(logElId);
        log.classList.remove('err'); log.classList.add('ok', 'show');
        log.innerHTML =
          '✓ 已保存 ' + escapeHTML(res.name) + '（' + sizeKB + ' KB · 含 INSTALL.md 安装说明）\n\n' +
          '<strong>📂 文件位置</strong>：\n' +
          '<code style="user-select:all; padding:2px 6px; background:rgba(255,255,255,0.06); border-radius:3px;">' + escapeHTML(res.path) + '</code>\n\n' +
          '<button class="btn btn-primary btn-sm cd-reveal-btn" data-path="' + escapeHTML(res.path) + '">📂 在 Finder 中显示</button> ' +
          '<button class="btn btn-ghost btn-sm cd-open-claude-btn">🚀 打开 Claude Desktop</button> ' +
          '<button class="btn btn-ghost btn-sm cd-copy-path-btn" data-path="' + escapeHTML(res.path) + '">📋 复制路径</button>\n\n' +
          '收件人安装方法：解压 zip → 看里面的 INSTALL.md（含 Claude Code / Cursor / Claude Desktop 三种装法）';
        // bind 子按钮
        log.querySelectorAll('.cd-reveal-btn').forEach((b) => {
          b.addEventListener('click', () => {
            fetch('/api/reveal', {
              method: 'POST', headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ path: b.dataset.path }),
            });
          });
        });
        log.querySelectorAll('.cd-open-claude-btn').forEach((b) => {
          b.addEventListener('click', () => {
            fetch('/api/open-app', {
              method: 'POST', headers: { 'Content-Type': 'application/json' },
              body: JSON.stringify({ app: 'Claude' }),
            }).catch(() => {});
          });
        });
        log.querySelectorAll('.cd-copy-path-btn').forEach((b) => {
          b.addEventListener('click', () => {
            navigator.clipboard?.writeText(b.dataset.path).then(() => {
              b.textContent = '✓ 已复制';
              setTimeout(() => { b.textContent = '📋 复制路径'; }, 1500);
            });
          });
        });
      })
      .catch((e) => {
        if (btn) { btn.disabled = false; btn.textContent = orig; }
        showLog(logElId, '✗ 网络错误：' + e, 'err');
      });
  }

  document.getElementById('orangecatExportZipBtn').addEventListener('click', (e) => {
    doSkillSaveZipFlow('orangecat', 'orangecatLog', e.currentTarget);
  });
  document.getElementById('orangecatExportMdBtn').addEventListener('click', () => {
    fetch('/api/claude-desktop/skill-md?name=orangecat')
      .then((r) => r.json())
      .then((d) => {
        if (d.error) { showLog('orangecatLog', '✗ ' + d.error + '（先"独立安装"再导出）', 'err'); return; }
        const text = d.content || '';
        const doCopy = () => {
          if (navigator.clipboard?.writeText) return navigator.clipboard.writeText(text).then(() => true, () => false);
          return Promise.resolve(false);
        };
        doCopy().then((ok) => {
          showLog('orangecatLog', ok
            ? '✓ SKILL.md 已复制到剪贴板（' + d.bytes + ' 字节）'
            : '✗ 剪贴板写入失败', ok ? 'ok' : 'err');
        });
      });
  });

  // 模板编辑器（QA 版 + 开发版双 Tab，各自缓存未保存内容）
  let currentTplWhich = 'qa';
  const tplDraft = { qa: null, dev: null }; // 缓存未保存的内容，切换 Tab 不丢

  function loadOrangecatTemplate() {
    fetch('/api/orangecat/template?which=' + currentTplWhich)
      .then((r) => r.json())
      .then((d) => {
        const src = document.getElementById('orangecatTplSource');
        const ta = document.getElementById('orangecatTplEditor');
        const label = currentTplWhich === 'qa' ? 'QA 版' : '开发版';
        src.textContent = d.source === 'custom'
          ? '⚙️ 当前使用：自定义 ' + label + ' (' + d.path + ')'
          : '📦 当前使用：内置 ' + label + ' 默认模板';
        // 如果有 draft 就用 draft，否则用后端返回
        ta.value = (tplDraft[currentTplWhich] !== null) ? tplDraft[currentTplWhich] : (d.content || '');
      });
  }

  // 切 Tab：先保存当前 textarea 到 draft，再切
  document.querySelectorAll('.orangecat-tpl-tab').forEach((btn) => {
    btn.addEventListener('click', () => {
      tplDraft[currentTplWhich] = document.getElementById('orangecatTplEditor').value;
      currentTplWhich = btn.dataset.which;
      document.querySelectorAll('.orangecat-tpl-tab').forEach((b) => {
        if (b.dataset.which === currentTplWhich) { b.classList.remove('btn-ghost'); b.classList.add('active'); }
        else { b.classList.add('btn-ghost'); b.classList.remove('active'); }
      });
      loadOrangecatTemplate();
    });
  });

  document.getElementById('orangecatTplSaveBtn').addEventListener('click', () => {
    const text = document.getElementById('orangecatTplEditor').value;
    if (text.length < 10) { showLog('orangecatLog', '✗ 模板太短', 'err'); return; }
    fetch('/api/orangecat/template?which=' + currentTplWhich, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ content: text }),
    }).then((r) => r.json()).then((d) => {
      tplDraft[currentTplWhich] = null; // 已保存，清 draft
      showLog('orangecatLog', d.ok
        ? '✓ 已保存 ' + (currentTplWhich === 'qa' ? 'QA 版' : '开发版') + ' → ' + d.path + '\n下次点「独立安装」才会写入 skill'
        : '✗ 保存失败：' + (d.error || ''),
        d.ok ? 'ok' : 'err');
      loadOrangecatTemplate();
      loadOrangecatStatus();
    });
  });
  document.getElementById('orangecatTplResetBtn').addEventListener('click', (e) => {
    const label = currentTplWhich === 'qa' ? 'QA 版' : '开发版';
    twoStep(e.currentTarget, '确认恢复内置 ' + label + '？', () => {
    fetch('/api/orangecat/template?which=' + currentTplWhich, { method: 'DELETE' })
      .then((r) => r.json())
      .then(() => {
        tplDraft[currentTplWhich] = null;
        showLog('orangecatLog', '✓ 已恢复内置 ' + label, 'ok');
        loadOrangecatTemplate();
        loadOrangecatStatus();
      });
    }); // twoStep end
  });
  document.getElementById('orangecatTplReloadBtn').addEventListener('click', () => {
    tplDraft[currentTplWhich] = null;
    loadOrangecatTemplate();
  });

  // ---------- Go Unit Test Skill 模块化安装 ----------
  // v1.7.6: 卡片在 UI 里已删除。下面所有逻辑用 if (unittestCardPresent) 门禁。
  const unittestCardPresent = !!document.getElementById('unittestStatus');
  let unittestManifest = { modules: [], selected: [] };

  function renderUnittestLists() {
    if (!unittestCardPresent) return;
    const refs = document.getElementById('unittestReferencesList');
    const assets = document.getElementById('unittestAssetsList');
    refs.innerHTML = '';
    assets.innerHTML = '';
    const selectedSet = new Set(unittestManifest.selected || []);
    (unittestManifest.modules || []).forEach((m) => {
      const row = document.createElement('label');
      row.style.display = 'flex';
      row.style.gap = '8px';
      row.style.padding = '4px 0';
      row.style.cursor = 'pointer';
      row.innerHTML = `
        <input type="checkbox" data-mod-id="${m.id}" ${selectedSet.has(m.id) ? 'checked' : ''}>
        <div>
          <div style="font-weight:600; font-size:13px;">${m.title}</div>
          <div style="font-size:12px; color:#888;">${m.desc}</div>
        </div>`;
      (m.group === 'references' ? refs : assets).appendChild(row);
    });
  }

  function loadUnittest() {
    if (!unittestCardPresent) return;
    Promise.all([
      fetch('/api/unit-test/manifest').then((r) => r.json()),
      fetch('/api/unit-test/status').then((r) => r.json()),
    ]).then(([manifest, status]) => {
      unittestManifest = manifest;
      // 用实际已装状态覆盖 selected（以便打开 App 时反映真实安装情况）
      if (status.installed_modules) {
        unittestManifest.selected = Object.keys(status.installed_modules).filter((k) => status.installed_modules[k]);
      }
      renderUnittestLists();

      const el = document.getElementById('unittestStatus');
      const c = status.claude_installed, cu = status.cursor_installed;
      const n = unittestManifest.selected.length;
      if (c && cu) {
        el.className = 'tixuebj-status installed';
        el.textContent = `✓ 已装（Claude + Cursor 双侧）· ${(status.version || '?').trim()} · ${n} 个可选模块`;
      } else if (c || cu) {
        el.className = 'tixuebj-status partial';
        el.textContent = '⚠ 只装了一侧：' + (c ? 'Claude ✓' : 'Claude ✗') + ' / ' + (cu ? 'Cursor ✓' : 'Cursor ✗');
      } else {
        el.className = 'tixuebj-status missing';
        el.textContent = '○ 未安装（勾选模块 → 点下面「应用选择并安装」）';
      }
    });
  }

  function collectSelectedModuleIds() {
    return Array.from(document.querySelectorAll('#unittestReferencesList input[type=checkbox]:checked, #unittestAssetsList input[type=checkbox]:checked'))
      .map((cb) => cb.dataset.modId);
  }

  if (unittestCardPresent) {
  document.getElementById('unittestInstallBtn').addEventListener('click', () => {
    const btn = document.getElementById('unittestInstallBtn');
    btn.disabled = true; btn.textContent = '安装中…';
    const ids = collectSelectedModuleIds();
    fetch('/api/unit-test/install', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ modules: ids }),
    }).then((r) => r.json()).then((res) => {
      btn.disabled = false; btn.textContent = '📥 应用选择并安装';
      let msg = res.ok ? '✓ 安装完成\n' : '✗ 部分失败\n';
      (res.installed || []).forEach((l) => msg += '  · ' + l + '\n');
      (res.warnings || []).forEach((w) => msg += '  ⚠ ' + w + '\n');
      msg += '\n⚠️ 必做：Cmd+Q 彻底重启 Claude Code / Cursor 才生效';
      showLog('unittestLog', msg, res.ok ? 'ok' : 'err');
      loadUnittest();
      showRestartBanner(null);
    });
  });

  let unittestRmConfirm = false;
  document.getElementById('unittestUninstallBtn').addEventListener('click', (e) => {
    const btn = e.currentTarget;
    if (!unittestRmConfirm) {
      unittestRmConfirm = true;
      btn.textContent = '再点一次确认卸载';
      btn.classList.add('danger-active');
      setTimeout(() => {
        unittestRmConfirm = false;
        btn.textContent = '🗑 卸载';
        btn.classList.remove('danger-active');
      }, 3000);
      return;
    }
    fetch('/api/unit-test/uninstall', { method: 'POST' })
      .then((r) => r.json())
      .then((res) => {
        unittestRmConfirm = false;
        btn.textContent = '🗑 卸载';
        btn.classList.remove('danger-active');
        showLog('unittestLog', '✓ 已卸载\n' + (res.removed || []).map((x) => '  · ' + x).join('\n'), 'ok');
        loadUnittest();
      });
  });

  document.getElementById('unittestSelectAllBtn').addEventListener('click', () => {
    document.querySelectorAll('#unittestReferencesList input[type=checkbox], #unittestAssetsList input[type=checkbox]').forEach((cb) => cb.checked = true);
  });
  document.getElementById('unittestClearBtn').addEventListener('click', () => {
    document.querySelectorAll('#unittestReferencesList input[type=checkbox], #unittestAssetsList input[type=checkbox]').forEach((cb) => cb.checked = false);
  });
  } // end if (unittestCardPresent)


  // ---------- 规范覆盖检查 ----------
  function renderCoverage(data) {
    const root = document.getElementById('coverageResult');
    const icon = (st) => ({ ok: '✅', outdated: '🟡', missing: '❌', custom: '🔷' }[st] || '⚪');
    const color = (st) => ({
      ok: '#3dd68c', outdated: '#e0a800', missing: '#ff6b6b', custom: '#5aa7ff',
    }[st] || '#888');

    const total = data.total || {};
    let html = '<div style="margin-bottom:12px; padding:10px; border-radius:6px; background:rgba(255,255,255,0.04);">';
    html += '<strong>总计</strong>：';
    html += ' ✅ ' + (total.ok || 0) + ' 已装最新';
    html += ' · 🟡 ' + (total.outdated || 0) + ' 过时';
    html += ' · ❌ ' + (total.missing || 0) + ' 缺失';
    html += ' · 🔷 ' + (total.custom || 0) + ' 自定义';
    html += '</div>';

    (data.bundles || []).forEach((b) => {
      html += '<details open style="margin-bottom:10px; border:1px solid rgba(255,255,255,0.08); border-radius:6px;">';
      html += '<summary style="cursor:pointer; padding:8px 12px; font-weight:600;">' + b.label;
      html += ' <span style="font-weight:400; color:#888; font-size:12px;">' + b.base_dir + '</span></summary>';
      html += '<div style="padding:8px 12px; font-family:ui-monospace,Menlo,monospace; font-size:12px;">';
      (b.items || []).forEach((it) => {
        html += '<div style="padding:2px 0; color:' + color(it.status) + '; display:flex; gap:8px; align-items:center;">';
        html += '<span style="flex:0 0 auto;">' + icon(it.status) + '</span>';
        html += '<span style="flex:1; min-width:0;">' + it.rel_path + '</span>';
        // 版本/日期标签（v1.7.15 起）
        let meta = '';
        if (it.installed_version || it.installed_modified) {
          meta = `<span style="flex:0 0 auto; font-size:11px; color:#888;">v${it.installed_version || '?'} (${it.installed_modified || '-'})</span>`;
        }
        if (it.status === 'outdated' && (it.embed_version || it.embed_modified)) {
          meta += `<span style="flex:0 0 auto; font-size:11px; color:#e0a800;"> → v${it.embed_version || '?'} (${it.embed_modified || '-'})</span>`;
        } else if (it.status === 'missing' && (it.embed_version || it.embed_modified)) {
          meta = `<span style="flex:0 0 auto; font-size:11px; color:#ff6b6b;">缺失 (最新: v${it.embed_version || '?'} ${it.embed_modified || ''})</span>`;
        }
        html += meta;
        html += '</div>';
      });
      if (!b.items || b.items.length === 0) {
        html += '<div style="color:#888;">（无可扫描文件）</div>';
      }
      html += '</div></details>';
    });

    root.innerHTML = html;

    // 如果有缺失或过时，启用"一键补齐"
    const fixBtn = document.getElementById('coverageFixBtn');
    const hasFixable = (total.missing || 0) + (total.outdated || 0) > 0;
    fixBtn.disabled = !hasFixable;
    if (hasFixable) {
      fixBtn.textContent = '📥 一键补齐（' + ((total.missing || 0) + (total.outdated || 0)) + ' 项）';
    } else {
      fixBtn.textContent = '📥 一键补齐（缺失/过时）';
    }
  }

  document.getElementById('coverageCheckBtn').addEventListener('click', () => {
    const btn = document.getElementById('coverageCheckBtn');
    btn.disabled = true; btn.textContent = '扫描中…';
    fetch('/api/coverage/check')
      .then((r) => r.json())
      .then((data) => {
        renderCoverage(data);
        btn.disabled = false; btn.textContent = '🔄 重新检查';
      })
      .catch((err) => {
        btn.disabled = false; btn.textContent = '🔍 开始检查';
        document.getElementById('coverageResult').innerHTML = '<div style="color:#ff6b6b;">扫描失败：' + err + '</div>';
      });
  });

  // 🔄 一键更新规范到最新（v1.7.26：全部 4 skill + Cursor rules + Slash commands）
  document.getElementById('updateAllBtn').addEventListener('click', () => {
    const btn = document.getElementById('updateAllBtn');
    btn.disabled = true; btn.textContent = '⏳ 更新中…';
    Promise.all([
      fetch('/api/install', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ scope: 'global', target: 'all' }),
      }).then((r) => r.json()).catch((e) => ({ ok: false, error: String(e) })),
      fetch('/api/orangecat/install',  { method: 'POST' }).then((r) => r.json()).catch((e) => ({ ok: false, error: String(e) })),
      fetch('/api/dev-dna/install',    { method: 'POST' }).then((r) => r.json()).catch((e) => ({ ok: false, error: String(e) })),
      fetch('/api/code-review/install',{ method: 'POST' }).then((r) => r.json()).catch((e) => ({ ok: false, error: String(e) })),
      fetch('/api/commands/install',   { method: 'POST' }).then((r) => r.json()).catch((e) => ({ ok: false, error: String(e) })),
    ]).then(([mainRes, ocRes, dnaRes, crRes, cmdRes]) => {
      btn.disabled = false; btn.textContent = '🔄 一键更新';
      const lines = [];
      const append = (label, res) => {
        if (res && res.ok) {
          lines.push('✓ ' + label + ' 已更新');
          (res.installed || []).forEach((l) => lines.push('  · ' + l));
        } else {
          lines.push('✗ ' + label + ' 失败：' + ((res && res.error) || '未知'));
        }
      };
      append('go-team-standards + Cursor rules', mainRes);
      append('orangecat',     ocRes);
      append('dev-dna',       dnaRes);
      append('code-review',   crRes);
      append('Slash commands (/tech-design /design-table /api-doc /review /tixuebj)', cmdRes);
      lines.push('');
      lines.push('💡 GSD 框架（gsd-build/get-shit-done）独立装 → 安装页「💪 GSD 框架」卡片点 npx 一键装');
      lines.push('');
      lines.push('⚠️ 必做：Cmd+Q 彻底重启 Claude Code / Cursor 才生效');
      const allOK = mainRes.ok && ocRes.ok && dnaRes.ok && crRes.ok && cmdRes.ok;
      showLog('updateAllLog', lines.join('\n'), allOK ? 'ok' : 'err');
      setTimeout(() => document.getElementById('coverageCheckBtn').click(), 300);
    });
  });

  document.getElementById('updateAllCheckBtn').addEventListener('click', () => {
    // 就是覆盖检查的快捷入口
    document.getElementById('coverageCheckBtn').click();
    // 滚动到 coverageResult
    document.getElementById('coverageResult').scrollIntoView({ behavior: 'smooth', block: 'center' });
  });

  document.getElementById('coverageFixBtn').addEventListener('click', () => {
    // 直接触发一次主安装（scope=global, target=all），会把 go-team-standards + cursor rules 重写
    // 注意：orangecat / go-unit-test 需要各自卡片的"独立安装"按钮补
    const btn = document.getElementById('coverageFixBtn');
    btn.disabled = true; btn.textContent = '补齐中…';
    fetch('/api/install', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ scope: 'global', target: 'all' }),
    }).then((r) => r.json()).then((res) => {
      btn.textContent = '📥 一键补齐（缺失/过时）';
      const note = document.createElement('div');
      note.style.cssText = 'margin-top:8px; padding:8px; border-radius:4px;';
      if (res.ok) {
        note.style.background = 'rgba(61,214,140,0.15)';
        note.style.color = '#3dd68c';
        note.textContent = '✓ 已触发 go-team-standards + Cursor rules 重装。如果 orangecat / go-unit-test 也有缺失/过时，请在对应卡片点"独立安装"。';
      } else {
        note.style.background = 'rgba(255,107,107,0.15)';
        note.style.color = '#ff6b6b';
        note.textContent = '✗ 重装失败：' + (res.error || '');
      }
      document.getElementById('coverageResult').prepend(note);
      // 自动重新扫描
      setTimeout(() => document.getElementById('coverageCheckBtn').click(), 500);
    });
  });

  // ---------- ☁️ 规范云端同步 (v1.7.30) ----------
  function fmtTimeAgo(iso) {
    if (!iso || iso === '0001-01-01T00:00:00Z') return '从未';
    const d = new Date(iso);
    const ms = Date.now() - d.getTime();
    const pad = n => String(n).padStart(2, '0');
    const fmtAbsolute = () => {
      const now = new Date();
      const sameDay = now.toDateString() === d.toDateString();
      if (sameDay) return pad(d.getHours()) + ':' + pad(d.getMinutes());
      return (d.getMonth() + 1) + '-' + pad(d.getDate()) + ' ' + pad(d.getHours()) + ':' + pad(d.getMinutes());
    };
    const m = Math.floor(ms / 60000);
    if (m < 60) return fmtAbsolute();
    const h = Math.floor(m / 60);
    if (h < 24) return h + ' 小时前';
    return Math.floor(h / 24) + ' 天前';
  }

  function shortSHA(s) { return s ? s.substring(0, 7) : '–'; }

  function loadStandardsSyncStatus() {
    const statusEl = document.getElementById('standardsSyncStatus');
    if (!statusEl) return;

    fetch('/api/standards-sync/config').then((r) => r.json()).then((cfg) => {
      if (!cfg.configured) {
        statusEl.className = 'tixuebj-status missing';
        statusEl.textContent = '⚠ 未配置 — 点开下面「⚙️ 配置」填 GitLab 镜像 URL';
        document.getElementById('standardsSyncConfigDetails').open = true;
        return;
      }
      // 已配置，显示已知信息
      document.getElementById('standardsSyncRepoURL').value = cfg.repo_url || '';
      document.getElementById('standardsSyncBranch').value = cfg.branch || 'main';

      let msg = `📡 ${cfg.repo_url}`;
      if (cfg.last_synced_sha) {
        msg += ` · 上次同步 ${shortSHA(cfg.last_synced_sha)} (${fmtTimeAgo(cfg.last_synced_at)})`;
      } else {
        msg += ' · 从未拉取';
      }
      if (cfg.last_check_error) {
        msg += `\n⚠ 上次检查失败：${cfg.last_check_error}`;
        statusEl.className = 'tixuebj-status partial';
      } else {
        statusEl.className = 'tixuebj-status installed';
      }
      statusEl.textContent = msg;

      // 异步检查是否有新版本
      fetch('/api/standards-sync/check').then((r) => r.json()).then((chk) => {
        if (!chk.reachable) return;
        if (chk.has_update) {
          statusEl.className = 'tixuebj-status partial';
          const synced = cfg.last_synced_at
            ? ` · 上次同步 ${fmtTimeAgo(cfg.last_synced_at)}`
            : ' · 从未拉取';
          statusEl.textContent = `🆕 有新版本 ${shortSHA(chk.latest_sha)}（当前 ${shortSHA(chk.current_sha)}${synced}）—— 点「📥 拉取最新规范」应用`;
        }
      }).catch(() => {});
    }).catch(() => {
      statusEl.textContent = '加载配置失败';
    });
  }

  // 代理配置加载 + 保存
  if (document.getElementById('proxyEnabled')) {
    fetch('/api/proxy/config').then((r) => r.json()).then((c) => {
      document.getElementById('proxyEnabled').checked = !!c.enabled;
      document.getElementById('proxyURL').value = c.proxy_url || '';
    });
    document.getElementById('proxySaveBtn').addEventListener('click', () => {
      const enabled = document.getElementById('proxyEnabled').checked;
      const proxy_url = document.getElementById('proxyURL').value.trim();
      const btn = document.getElementById('proxySaveBtn');
      btn.disabled = true; btn.textContent = '⏳';
      fetch('/api/proxy/config', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ enabled, proxy_url }),
      }).then((r) => r.json()).then((res) => {
        btn.disabled = false; btn.textContent = '💾 保存';
        if (res.ok) {
          btn.textContent = '✓ 已保存';
          setTimeout(() => btn.textContent = '💾 保存', 1500);
        } else {
          showLog('standardsSyncLog', '✗ ' + (res.error || '保存失败'), 'err');
        }
      });
    });
  }

  if (document.getElementById('standardsSyncStatus')) {
    loadStandardsSyncStatus();

    document.getElementById('standardsSyncSaveBtn').addEventListener('click', () => {
      const btn = document.getElementById('standardsSyncSaveBtn');
      const url = document.getElementById('standardsSyncRepoURL').value.trim();
      const branch = document.getElementById('standardsSyncBranch').value.trim() || 'main';
      if (!url) {
        showLog('standardsSyncLog', '✗ 仓库 URL 不能为空', 'err');
        return;
      }
      btn.disabled = true; btn.textContent = '⏳ 测试中…';
      fetch('/api/standards-sync/config', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ repo_url: url, branch }),
      }).then((r) => r.json()).then((res) => {
        btn.disabled = false; btn.textContent = '💾 保存并测试连通';
        if (res.ok) {
          showLog('standardsSyncLog', `✓ 配置保存成功 · 远端最新 ${shortSHA(res.latest_sha)}`, 'ok');
          loadStandardsSyncStatus();
        } else {
          showLog('standardsSyncLog', `✗ ${res.error || '失败'}\n${res.detail || ''}`, 'err');
        }
      }).catch((e) => {
        btn.disabled = false; btn.textContent = '💾 保存并测试连通';
        showLog('standardsSyncLog', '✗ 请求失败：' + e, 'err');
      });
    });

    document.getElementById('standardsSyncCheckBtn').addEventListener('click', () => {
      const btn = document.getElementById('standardsSyncCheckBtn');
      btn.disabled = true; btn.textContent = '⏳ 检查中…';
      fetch('/api/standards-sync/check').then((r) => r.json()).then((res) => {
        btn.disabled = false; btn.textContent = '🔍 检查更新';
        if (!res.configured) {
          showLog('standardsSyncLog', '✗ 未配置仓库', 'err');
          return;
        }
        if (!res.reachable) {
          showLog('standardsSyncLog', '✗ 连不上远端：' + (res.error || ''), 'err');
          return;
        }
        if (res.has_update) {
          showLog('standardsSyncLog', `🆕 有新版本！\n  当前：${shortSHA(res.current_sha)}\n  远端：${shortSHA(res.latest_sha)}\n  点「📥 拉取最新规范」应用`, 'ok');
        } else {
          showLog('standardsSyncLog', `✓ 已是最新（${shortSHA(res.latest_sha)}）`, 'ok');
        }
        loadStandardsSyncStatus();
      });
    });

    document.getElementById('standardsSyncPullBtn').addEventListener('click', () => {
      const btn = document.getElementById('standardsSyncPullBtn');
      btn.disabled = true; btn.textContent = '⏳ 拉取中…';
      fetch('/api/standards-sync/pull', { method: 'POST' }).then((r) => r.json()).then((res) => {
        btn.disabled = false; btn.textContent = '📥 拉取最新规范';
        if (res.ok) {
          let msg = `✓ ${res.message}\n  最新 SHA：${shortSHA(res.latest_sha)}\n`;
          if (res.files_changed > 0) {
            msg += '  变更文件：\n';
            (res.changed || []).slice(0, 20).forEach((f) => msg += '    · ' + f + '\n');
            // 拉取有变更 → 立即刷新规范模块卡片
            fetch('/api/catalog').then((r) => r.json()).then((d) => {
              skills = d.skills || [];
              renderSkills(skills);
              document.getElementById('skillsCount').textContent = skills.length;
            });
          } else {
            msg += '  无文件需更新（哈希一致）';
          }
          showLog('standardsSyncLog', msg, 'ok');
        } else {
          showLog('standardsSyncLog', `✗ ${res.error || '失败'}\n${res.detail || ''}`, 'err');
        }
        loadStandardsSyncStatus();
      });
    });
  }

  // ---------- 📁 项目级 skill 同步 (v1.7.21) ----------
  const projScanBtn = document.getElementById('projScanBtn');
  if (projScanBtn) {
    let projScanData = null;

    function renderProjScanResult(d) {
      const root = document.getElementById('projScanResult');
      if (!d.projects || d.projects.length === 0) {
        root.innerHTML = '<div style="color:#888; padding:8px;">在 ' + escapeHTML(d.root) + ' 下没找到含 .claude/.cursor 的项目</div>';
        return;
      }
      let html = '<div style="margin-bottom:8px; color:#888; font-size:12px;">扫到 ' + d.count + ' 个项目（root: ' + escapeHTML(d.root) + '）</div>';
      html += '<div style="border:1px solid rgba(255,255,255,0.1); border-radius:6px; padding:8px;">';
      d.projects.forEach((p, i) => {
        html += '<label style="display:flex; gap:8px; padding:6px 4px; border-bottom:1px solid rgba(255,255,255,0.05); cursor:pointer;">';
        html += '<input type="checkbox" class="proj-check" data-path="' + escapeHTML(p.project_path) + '" checked>';
        html += '<div style="flex:1;">';
        html += '<div style="font-family:ui-monospace,Menlo,monospace; font-size:13px;">' + escapeHTML(p.project_path) + '</div>';
        html += '<div style="font-size:12px; color:#888; margin-top:2px;">';
        const tags = [];
        if (p.has_claude_skills) tags.push('🤖 .claude/skills (' + (p.claude_skill_names || []).length + ')');
        if (p.has_cursor_skills) tags.push('✂️ .cursor/skills-cursor (' + (p.cursor_skill_names || []).length + ')');
        if (p.has_cursor_rules) tags.push('📜 .cursor/rules (' + p.cursor_rules_count + ' .mdc)');
        html += tags.join(' · ');
        html += '</div></div></label>';
      });
      html += '</div>';
      html += '<div class="dist-buttons" style="margin-top:10px;">';
      html += '<button id="projSyncSelectedBtn" class="btn btn-primary">🔄 把规范同步到勾选的项目</button>';
      html += '<button id="projSelectAllBtn" class="btn btn-ghost btn-sm">全选</button>';
      html += '<button id="projSelectNoneBtn" class="btn btn-ghost btn-sm">全不选</button>';
      html += '</div>';
      html += '<div class="install-log" id="projSyncLog"></div>';
      root.innerHTML = html;

      // bind
      document.getElementById('projSelectAllBtn').addEventListener('click', () => {
        document.querySelectorAll('.proj-check').forEach((cb) => cb.checked = true);
      });
      document.getElementById('projSelectNoneBtn').addEventListener('click', () => {
        document.querySelectorAll('.proj-check').forEach((cb) => cb.checked = false);
      });
      document.getElementById('projSyncSelectedBtn').addEventListener('click', () => {
        const selected = Array.from(document.querySelectorAll('.proj-check:checked'))
          .map((cb) => cb.dataset.path);
        if (selected.length === 0) {
          showLog('projSyncLog', '✗ 没勾选任何项目', 'err');
          return;
        }
        const btn = document.getElementById('projSyncSelectedBtn');
        btn.disabled = true; btn.textContent = '⏳ 同步中…';
        fetch('/api/project-skills/sync', {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ projects: selected }),
        }).then((r) => r.json()).then((res) => {
          btn.disabled = false; btn.textContent = '🔄 把规范同步到勾选的项目';
          let msg = '';
          (res.results || []).forEach((rr) => {
            msg += (rr.ok ? '✓ ' : '⚠ ') + rr.project_path + '\n';
            (rr.installed || []).forEach((l) => msg += '    ' + l + '\n');
            (rr.warnings || []).forEach((w) => msg += '    ⚠ ' + w + '\n');
          });
          msg += '\n⚠️ 提醒：项目级 skill 也要 Cmd+Q 重启 Claude/Cursor 才生效';
          showLog('projSyncLog', msg, res.ok ? 'ok' : 'err');
        });
      });
    }

    projScanBtn.addEventListener('click', () => {
      const root = document.getElementById('projScanRoot').value.trim();
      projScanBtn.disabled = true; projScanBtn.textContent = '扫描中…';
      fetch('/api/project-skills/scan?root=' + encodeURIComponent(root))
        .then((r) => r.json()).then((d) => {
          projScanBtn.disabled = false; projScanBtn.textContent = '🔍 扫描';
          if (d.error) {
            document.getElementById('projScanResult').innerHTML =
              '<div style="color:#ff6b6b;">✗ ' + escapeHTML(d.error) + '</div>';
            return;
          }
          projScanData = d;
          renderProjScanResult(d);
        }).catch((e) => {
          projScanBtn.disabled = false; projScanBtn.textContent = '🔍 扫描';
          document.getElementById('projScanResult').innerHTML =
            '<div style="color:#ff6b6b;">✗ 网络错误：' + e + '</div>';
        });
    });
  }

  // ---------- ⌨️ Slash Commands (v1.7.23) ----------
  function loadCommandsList() {
    const root = document.getElementById('commandsList');
    if (!root) return;
    fetch('/api/commands/list').then((r) => r.json()).then((d) => {
      let html = '<div style="border:1px solid rgba(255,255,255,0.1); border-radius:6px; overflow:hidden;">';
      (d.commands || []).forEach((c) => {
        const c1 = c.claude_installed ? '✅' : '○';
        const c2 = c.cursor_installed ? '✅' : '○';
        html += '<div style="padding:10px; border-bottom:1px solid rgba(255,255,255,0.05);">';
        html += '<div style="display:flex; gap:8px; align-items:start;">';
        html += '<div style="flex:1;">';
        html += '<div><strong>' + escapeHTML(c.title) + '</strong> ';
        html += '<code style="font-size:12px; color:#a855f7;">/' + escapeHTML(c.id) + '</code></div>';
        html += '<div style="font-size:12px; color:#888; margin-top:2px;">' + escapeHTML(c.description) + '</div>';
        html += '<div style="font-size:11px; color:#666; margin-top:4px; font-family:ui-monospace,Menlo,monospace;">用法：' + escapeHTML(c.usage_hint) + '</div>';
        html += '</div>';
        html += '<div style="font-size:11px; color:#888; white-space:nowrap;">Claude ' + c1 + ' / Cursor ' + c2 + '</div>';
        html += '</div></div>';
      });
      html += '</div>';
      html += '<div style="font-size:12px; color:#888; margin-top:6px;">装到：<code>' + escapeHTML(d.claude_dir) + '</code> + <code>' + escapeHTML(d.cursor_dir) + '</code></div>';
      root.innerHTML = html;
    });
  }

  const cmdInstallBtn = document.getElementById('commandsInstallBtn');
  if (cmdInstallBtn) {
    cmdInstallBtn.addEventListener('click', () => {
      cmdInstallBtn.disabled = true; cmdInstallBtn.textContent = '安装中…';
      fetch('/api/commands/install', { method: 'POST' })
        .then((r) => r.json()).then((res) => {
          cmdInstallBtn.disabled = false; cmdInstallBtn.textContent = '📥 安装全部 (Claude + Cursor)';
          let msg = res.ok ? '✓ 已装 ' + res.count + ' 个文件\n' : '✗ 失败\n';
          (res.installed || []).slice(0, 12).forEach((l) => msg += '  · ' + l + '\n');
          (res.warnings || []).forEach((w) => msg += '  ⚠ ' + w + '\n');
          msg += '\n用法：在 Claude Code / Cursor 里直接敲 /tech-design 或 /design-table 等';
          msg += '\n⚠️ 必做：Cmd+Q 重启 Claude/Cursor 后命令才显示在 / 列表';
          showLog('commandsLog', msg, res.ok ? 'ok' : 'err');
          loadCommandsList();
        });
    });

    let cmdRmConfirm = false;
    document.getElementById('commandsUninstallBtn').addEventListener('click', (e) => {
      const btn = e.currentTarget;
      if (!cmdRmConfirm) {
        cmdRmConfirm = true;
        btn.textContent = '再点一次确认卸载';
        btn.classList.add('danger-active');
        setTimeout(() => { cmdRmConfirm = false; btn.textContent = '🗑 卸载'; btn.classList.remove('danger-active'); }, 3000);
        return;
      }
      fetch('/api/commands/uninstall', { method: 'POST' }).then((r) => r.json()).then((res) => {
        cmdRmConfirm = false; btn.textContent = '🗑 卸载'; btn.classList.remove('danger-active');
        showLog('commandsLog', '✓ 已删 ' + (res.removed || []).length + ' 个文件', 'ok');
        loadCommandsList();
      });
    });

    loadCommandsList();
  }

  // ---------- 🧬 dev-dna · 个人开发档案 (v1.7.16) ----------
  function loadDevDnaStatus() {
    fetch('/api/dev-dna/status').then((r) => r.json()).then((d) => {
      const el = document.getElementById('devDnaStatus');
      if (!el) return;
      const c = d.claude_installed, cu = d.cursor_installed;
      const custom = d.custom_profile_used ? '（使用自定义 profile）' : '（使用内置默认 profile）';
      if (c && cu) {
        el.className = 'tixuebj-status installed';
        el.textContent = '✓ 已装（Claude + Cursor 双侧）· ' + (d.version || '?').trim() + ' ' + custom;
      } else if (c || cu) {
        el.className = 'tixuebj-status partial';
        el.textContent = '⚠ 只装了一侧：' + (c ? 'Claude ✓' : 'Claude ✗') + ' / ' + (cu ? 'Cursor ✓' : 'Cursor ✗');
      } else {
        el.className = 'tixuebj-status missing';
        el.textContent = '○ 未安装（点下面"独立安装"即可）';
      }
    });
  }

  function loadDevDnaProfile() {
    fetch('/api/dev-dna/profile').then((r) => r.json()).then((d) => {
      const src = document.getElementById('devDnaProfileSource');
      const ta = document.getElementById('devDnaProfileEditor');
      if (!src || !ta) return;
      src.textContent = d.source === 'custom'
        ? '⚙️ 当前使用：自定义 profile (' + d.path + ')'
        : '📦 当前使用：内置默认 profile（点保存自动切到自定义）';
      ta.value = d.content || '';
    });
  }

  const devDnaInstallBtn = document.getElementById('devDnaInstallBtn');
  if (devDnaInstallBtn) {
    devDnaInstallBtn.addEventListener('click', () => {
      devDnaInstallBtn.disabled = true; devDnaInstallBtn.textContent = '安装中…';
      fetch('/api/dev-dna/install', { method: 'POST' }).then((r) => r.json()).then((res) => {
        devDnaInstallBtn.disabled = false; devDnaInstallBtn.textContent = '📥 独立安装';
        let msg = res.ok ? '✓ 安装完成\n' : '✗ 部分失败\n';
        (res.installed || []).forEach((l) => msg += '  · ' + l + '\n');
        (res.warnings || []).forEach((w) => msg += '  ⚠ ' + w + '\n');
        msg += '\n⚠️ 必做：Cmd+Q 彻底重启 Claude Code / Cursor 才生效';
        showLog('devDnaLog', msg, res.ok ? 'ok' : 'err');
        loadDevDnaStatus();
      });
    });

    let devDnaRmConfirm = false;
    document.getElementById('devDnaUninstallBtn').addEventListener('click', (e) => {
      const btn = e.currentTarget;
      if (!devDnaRmConfirm) {
        devDnaRmConfirm = true;
        btn.textContent = '再点一次确认卸载';
        btn.classList.add('danger-active');
        setTimeout(() => { devDnaRmConfirm = false; btn.textContent = '🗑 卸载'; btn.classList.remove('danger-active'); }, 3000);
        return;
      }
      fetch('/api/dev-dna/uninstall', { method: 'POST' }).then((r) => r.json()).then((res) => {
        devDnaRmConfirm = false;
        btn.textContent = '🗑 卸载';
        btn.classList.remove('danger-active');
        showLog('devDnaLog', '✓ 已卸载\n' + (res.removed || []).map((x) => '  · ' + x).join('\n'), 'ok');
        loadDevDnaStatus();
      });
    });

    document.getElementById('devDnaExportZipBtn').addEventListener('click', (e) => {
      doSkillSaveZipFlow('dev-dna', 'devDnaLog', e.currentTarget);
    });

    document.getElementById('devDnaProfileSaveBtn').addEventListener('click', () => {
      const text = document.getElementById('devDnaProfileEditor').value;
      if (text.length < 50) { showLog('devDnaLog', '✗ 档案太短（至少 50 字符）', 'err'); return; }
      fetch('/api/dev-dna/profile', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ content: text }),
      }).then((r) => r.json()).then((d) => {
        showLog('devDnaLog', d.ok
          ? '✓ 已保存自定义档案 → ' + d.path + '\n下次点「独立安装」才会写入 skill'
          : '✗ 保存失败：' + (d.error || ''),
          d.ok ? 'ok' : 'err');
        loadDevDnaProfile();
        loadDevDnaStatus();
      });
    });

    document.getElementById('devDnaProfileResetBtn').addEventListener('click', (e) => {
      twoStep(e.currentTarget, '确认恢复内置默认档案？', () => {
        fetch('/api/dev-dna/profile', { method: 'DELETE' }).then((r) => r.json()).then(() => {
          showLog('devDnaLog', '✓ 已恢复内置默认 profile', 'ok');
          loadDevDnaProfile(); loadDevDnaStatus();
        });
      });
    });

    document.getElementById('devDnaProfileReloadBtn').addEventListener('click', loadDevDnaProfile);
  }

  // 启动 + 切到 Install Tab 时都刷状态
  loadOrangecatStatus();
  loadOrangecatTemplate();
  loadUnittest();
  loadDevDnaStatus();
  loadDevDnaProfile();

  // ---------- 💪 GSD 框架 v2 · gsd-build/get-shit-done (v1.7.39) ----------
  function loadGSDStatus() {
    const el = document.getElementById('gsdStatus');
    if (!el) return;
    fetch('/api/gsd/status').then((r) => r.json()).then((d) => {
      const lines = [];
      if (!d.npx_available) {
        el.className = 'tixuebj-status missing';
        el.textContent = '⚠ 找不到 npx（请先装 Node.js: brew install node）';
        return;
      }
      lines.push(`✓ Node ${d.node_version || '(未知)'}`);

      if (d.framework_ready) {
        lines.push(`✓ GSD 框架已装 · ${d.installed_count} 个 gsd-* skill（核心 ${d.core_installed}/${d.core_total}）`);
        el.className = 'tixuebj-status installed';
      } else if (d.installed_count > 0) {
        lines.push(`⚠ 部分装好 · ${d.installed_count} 个 gsd-* skill（核心 ${d.core_installed}/${d.core_total}）—— 建议重跑一键装`);
        el.className = 'tixuebj-status partial';
      } else {
        lines.push('❌ 未装 GSD 框架');
        el.className = 'tixuebj-status missing';
      }
      if (d.has_old_simple) {
        lines.push('⚠ 检测到老版任务清单 GTD 简版 (~/.claude/skills/gsd/) —— 点「🗑 卸载」一并清理');
      }
      el.textContent = lines.join('\n');
    }).catch(() => {
      el.textContent = '加载失败';
    });
  }

  const gsdInstallBtn = document.getElementById('gsdInstallBtn');
  if (gsdInstallBtn) {
    gsdInstallBtn.addEventListener('click', () => {
      if (!confirm('将执行：\n  npx --yes get-shit-done-cc@latest --claude --global\n\n首次可能下载 50-100MB 包，最多等 8 分钟。\n点确定开始？')) return;
      gsdInstallBtn.disabled = true; gsdInstallBtn.textContent = '⏳ 安装中（最多 8 分钟）…';
      fetch('/api/gsd/install', { method: 'POST' }).then((r) => r.json()).then((res) => {
        gsdInstallBtn.disabled = false; gsdInstallBtn.textContent = '🚀 一键装（npx）';
        let msg = res.ok ? '✓ 安装完成（exit 0）\n' : '✗ 失败 (exit ' + (res.exit_code || '?') + ')\n';
        if (res.error) msg += '错误: ' + res.error + '\n';
        if (res.hint) msg += '提示: ' + res.hint + '\n';
        msg += '\n--- npx 输出 ---\n' + (res.output || '(空)');
        showLog('gsdLog', msg, res.ok ? 'ok' : 'err');
        loadGSDStatus();
      });
    });

    document.getElementById('gsdCopyCmdBtn').addEventListener('click', (e) => {
      const cmd = 'npx --yes get-shit-done-cc@latest --claude --global';
      navigator.clipboard.writeText(cmd).then(() => {
        const btn = e.currentTarget;
        const orig = btn.textContent;
        btn.textContent = '✓ 已复制';
        setTimeout(() => { btn.textContent = orig; }, 1500);
      });
    });

    let gsdRmConfirm = false;
    document.getElementById('gsdUninstallBtn').addEventListener('click', (e) => {
      const btn = e.currentTarget;
      if (!gsdRmConfirm) {
        gsdRmConfirm = true;
        btn.classList.add('danger-active');
        btn.textContent = '⚠ 再点确认（删所有 gsd-* + 老版 gsd/）';
        setTimeout(() => {
          gsdRmConfirm = false;
          btn.classList.remove('danger-active');
          btn.textContent = '🗑 卸载全部 gsd-*';
        }, 4000);
        return;
      }
      gsdRmConfirm = false;
      btn.classList.remove('danger-active');
      btn.textContent = '卸载中…';
      fetch('/api/gsd/uninstall', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ include_old_simple: true }),
      }).then((r) => r.json()).then((res) => {
        btn.textContent = '🗑 卸载全部 gsd-*';
        let msg = res.ok ? '✓ 已卸载\n' : '✗ 失败\n';
        (res.removed || []).forEach((d) => msg += '  · 删 ' + d + '\n');
        (res.failed || []).forEach((d) => msg += '  ⚠ ' + d + '\n');
        if (res.note) msg += '\n' + res.note;
        showLog('gsdLog', msg, res.ok ? 'ok' : 'err');
        loadGSDStatus();
      });
    });
  }

  loadGSDStatus();

  // ---------- ⚡ Superpowers tab（v1.8.7）----------
  let spTarget = 'claude'; // 'claude' | 'cursor' | 'project'
  let spProjectPath = '';

  const SP_COMMANDS = {
    claude:  '/plugin install superpowers@claude-plugins-official',
    cursor:  '/add-plugin superpowers',
    project: '/plugin install superpowers@claude-plugins-official --scope project',
  };

  function updateSpCmdDisplay() {
    const el = document.getElementById('spInstallCmd');
    if (el) el.textContent = SP_COMMANDS[spTarget] || '';
    const pathRow = document.getElementById('spProjectPathRow');
    if (pathRow) pathRow.style.display = spTarget === 'project' ? 'flex' : 'none';
  }

  function loadSuperpowersStatus() {
    const statusEl = document.getElementById('spStatusPanel');
    const countEl  = document.getElementById('superpowersCount');
    if (!statusEl) return;
    statusEl.className = 'tixuebj-status';
    statusEl.textContent = '检查中…';

    const params = new URLSearchParams({ target: spTarget });
    if (spTarget === 'project' && spProjectPath) params.set('path', spProjectPath);

    fetch('/api/superpowers/status?' + params).then((r) => r.json()).then((d) => {
      if (d.installed) {
        statusEl.className = 'tixuebj-status installed';
        statusEl.textContent = `✓ 已安装 · ${d.skill_count || 8} 个 skill · 路径：${d.install_path || '~/.claude/skills/'}`;
      } else {
        statusEl.className = 'tixuebj-status missing';
        statusEl.textContent = '❌ 未安装 · 点下方「一键安装」';
      }
      if (countEl) countEl.textContent = d.installed ? (d.skill_count || '✓') : '–';
    }).catch((e) => {
      statusEl.className = 'tixuebj-status missing';
      statusEl.textContent = '检查失败：' + e;
    });
  }

  // 目标切换
  document.querySelectorAll('#spTargetBar .sp-target-btn').forEach((btn) => {
    btn.addEventListener('click', () => {
      document.querySelectorAll('#spTargetBar .sp-target-btn').forEach((b) => b.classList.remove('active'));
      btn.classList.add('active');
      spTarget = btn.dataset.target;
      updateSpCmdDisplay();
      loadSuperpowersStatus();
    });
  });

  const spPathBrowse = document.getElementById('spProjectPathBrowse');
  if (spPathBrowse) {
    spPathBrowse.addEventListener('click', () => {
      spPathBrowse.disabled = true; spPathBrowse.textContent = '⏳';
      fetch('/api/scaffold/pick-folder?prompt=' + encodeURIComponent('选择项目根目录'), { method: 'POST' })
        .then((r) => r.json()).then((res) => {
          spPathBrowse.disabled = false; spPathBrowse.textContent = '📁 选目录';
          if (res.path) {
            document.getElementById('spProjectPathInput').value = res.path;
            spProjectPath = res.path;
            loadSuperpowersStatus();
          }
        }).catch(() => { spPathBrowse.disabled = false; spPathBrowse.textContent = '📁 选目录'; });
    });
  }

  const spPathConfirm = document.getElementById('spProjectPathConfirm');
  if (spPathConfirm) {
    spPathConfirm.addEventListener('click', () => {
      const val = document.getElementById('spProjectPathInput').value.trim();
      if (val) { spProjectPath = val; loadSuperpowersStatus(); }
    });
  }

  // 复制命令
  const spCopyBtn = document.getElementById('spCopyCmdBtn');
  if (spCopyBtn) {
    spCopyBtn.addEventListener('click', () => {
      const cmd = SP_COMMANDS[spTarget] || '';
      navigator.clipboard.writeText(cmd).then(() => {
        const orig = spCopyBtn.textContent;
        spCopyBtn.textContent = '✓ 已复制';
        setTimeout(() => spCopyBtn.textContent = orig, 1500);
      }).catch(() => {
        prompt('手工复制：', cmd);
      });
    });
  }

  // 安装
  const spInstallBtn = document.getElementById('spInstallBtn');
  if (spInstallBtn) {
    spInstallBtn.addEventListener('click', () => {
      const body = { target: spTarget };
      if (spTarget === 'project') body.path = spProjectPath;
      spInstallBtn.disabled = true; spInstallBtn.textContent = '⏳ 安装中…';
      fetch('/api/superpowers/install', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(body),
      }).then((r) => r.json()).then((res) => {
        spInstallBtn.disabled = false; spInstallBtn.textContent = '🚀 一键安装';
        const msg = res.ok
          ? '✓ 安装成功\n' + (res.detail || '')
          : '✗ 安装失败：' + (res.error || '未知') + '\n' + (res.detail || '');
        showLog('spLog', msg, res.ok ? 'ok' : 'err');
        loadSuperpowersStatus();
      }).catch((e) => {
        spInstallBtn.disabled = false; spInstallBtn.textContent = '🚀 一键安装';
        showLog('spLog', '✗ 请求失败：' + e, 'err');
      });
    });
  }

  // 卸载
  const spUninstallBtn = document.getElementById('spUninstallBtn');
  if (spUninstallBtn) {
    spUninstallBtn.addEventListener('click', (e) => {
      twoStep(e.currentTarget, '确认卸载 Superpowers？', () => {
        const body = { target: spTarget };
        if (spTarget === 'project') body.path = spProjectPath;
        fetch('/api/superpowers/uninstall', {
          method: 'POST', headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify(body),
        }).then((r) => r.json()).then((res) => {
          let msg = res.ok ? '✓ 已卸载\n' : '✗ 失败\n';
          (res.removed || []).forEach((p) => msg += '  · 删 ' + p + '\n');
          if (res.error) msg += '\n✗ ' + res.error;
          showLog('spLog', msg, res.ok ? 'ok' : 'err');
          loadSuperpowersStatus();
        }).catch((e) => showLog('spLog', '✗ 请求失败：' + e, 'err'));
      });
    });
  }

  // 刷新按钮
  const spRefreshBtn = document.getElementById('spRefreshBtn');
  if (spRefreshBtn) spRefreshBtn.addEventListener('click', loadSuperpowersStatus);

  // 进 superpowers tab 时加载
  document.querySelectorAll('.navitem').forEach((btn) => {
    if (btn.dataset.tab === 'superpowers') btn.addEventListener('click', loadSuperpowersStatus);
  });

  // 初始化
  updateSpCmdDisplay();
  loadSuperpowersStatus();

  document.querySelectorAll('.navitem').forEach((btn) => {
    if (btn.dataset.tab === 'install') btn.addEventListener('click', () => {
      loadOrangecatStatus();
      loadUnittest();
      loadDevDnaStatus();
      loadGSDStatus();
    });
  });

  // ---------- Claude Desktop 导出（v1.7.18：落盘 + 显示路径 + Finder 定位）----------
  document.querySelectorAll('.cd-zip-btn').forEach((btn) => {
    btn.addEventListener('click', () => {
      const name = btn.dataset.skill;
      const orig = btn.textContent;
      btn.disabled = true; btn.textContent = '导出中…';
      showLog('cdLog', '正在打包 ' + name + ' 并保存到磁盘…', '');

      fetch('/api/claude-desktop/save-zip?name=' + encodeURIComponent(name), { method: 'POST' })
        .then((r) => r.json())
        .then((res) => {
          btn.disabled = false; btn.textContent = orig;
          if (!res.ok) {
            showLog('cdLog', '✗ 失败：' + (res.error || res.detail || '未知'), 'err');
            return;
          }
          const sizeKB = (res.size / 1024).toFixed(1);
          // 渲染含按钮的 HTML 到 cdLog（用 innerHTML 而非 textContent）
          const log = document.getElementById('cdLog');
          log.classList.remove('err'); log.classList.add('ok', 'show');
          log.innerHTML =
            '✓ 已保存 ' + escapeHTML(res.name) + '（' + sizeKB + ' KB）\n\n' +
            '<strong>📂 文件位置</strong>：\n' +
            '<code style="user-select:all; padding:2px 6px; background:rgba(255,255,255,0.06); border-radius:3px;">' + escapeHTML(res.path) + '</code>\n\n' +
            '<button class="btn btn-primary btn-sm cd-reveal-btn" data-path="' + escapeHTML(res.path) + '">📂 在 Finder 中显示</button> ' +
            '<button class="btn btn-ghost btn-sm cd-open-claude-btn">🚀 打开 Claude Desktop</button> ' +
            '<button class="btn btn-ghost btn-sm cd-copy-path-btn" data-path="' + escapeHTML(res.path) + '">📋 复制路径</button>\n\n' +
            '导入步骤：\n' +
            '  1. 点上方「📂 在 Finder 中显示」定位到 ' + escapeHTML(res.name) + '\n' +
            '  2. 点上方「🚀 打开 Claude Desktop」（或自己 Cmd+Space → "Claude"）\n' +
            '  3. Claude Desktop → Settings → Skills → + → 上传 zip → 把刚才那个文件拖进去';

          // bind 子按钮
          log.querySelectorAll('.cd-reveal-btn').forEach((b) => {
            b.addEventListener('click', () => {
              fetch('/api/reveal', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ path: b.dataset.path }),
              });
            });
          });
          log.querySelectorAll('.cd-open-claude-btn').forEach((b) => {
            b.addEventListener('click', () => {
              fetch('/api/open-app', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ app: 'Claude' }),
              }).catch(() => {});
            });
          });
          log.querySelectorAll('.cd-copy-path-btn').forEach((b) => {
            b.addEventListener('click', () => {
              navigator.clipboard?.writeText(b.dataset.path).then(() => {
                b.textContent = '✓ 已复制';
                setTimeout(() => { b.textContent = '📋 复制路径'; }, 1500);
              });
            });
          });
        })
        .catch((e) => {
          btn.disabled = false; btn.textContent = orig;
          showLog('cdLog', '✗ 网络错误：' + e, 'err');
        });
    });
  });

  document.querySelectorAll('.cd-md-btn').forEach((btn) => {
    btn.addEventListener('click', () => {
      const name = btn.dataset.skill;
      fetch('/api/claude-desktop/skill-md?name=' + encodeURIComponent(name))
        .then((r) => r.json())
        .then((d) => {
          if (d.error) { showLog('cdLog', '✗ ' + d.error, 'err'); return; }
          const text = d.content || '';
          const fallbackCopy = () => {
            try {
              const ta = document.createElement('textarea');
              ta.value = text;
              ta.style.position = 'fixed'; ta.style.left = '-9999px';
              document.body.appendChild(ta);
              ta.select();
              document.execCommand('copy');
              document.body.removeChild(ta);
              return true;
            } catch (_) { return false; }
          };
          const finish = (ok) => {
            if (ok) {
              showLog('cdLog',
                '✓ ' + name + ' 的 SKILL.md 已复制到剪贴板（' + d.bytes + ' 字节）\n\n' +
                '粘贴：Claude Desktop → Skills → + → Editor (</>) → Cmd+V\n\n' +
                '⚠️ 仅主 SKILL.md，附带文件（references / templates 等）会丢。要完整版用 📦 zip 按钮。',
                'ok');
            } else {
              showLog('cdLog', '✗ 剪贴板写入失败', 'err');
            }
          };
          if (navigator.clipboard?.writeText) {
            navigator.clipboard.writeText(text).then(() => finish(true), () => finish(fallbackCopy()));
          } else {
            finish(fallbackCopy());
          }
        })
        .catch((e) => showLog('cdLog', '✗ ' + e, 'err'));
    });
  });

  // ---------- 日志控制台 ----------
  let logsData = [];
  let logsExpanded = new Set(); // time+op 作 key 展开状态
  let logsAutoTimer = null;

  function loadLogs() {
    fetch('/api/logs')
      .then((r) => r.json())
      .then((d) => {
        logsData = (d.entries || []).slice().reverse(); // 新的在前
        document.getElementById('logsCount').textContent = logsData.length;
        renderLogs();
      });
  }

  function renderLogs() {
    const kinds = new Set();
    document.querySelectorAll('.logs-filters input[data-kind]').forEach((el) => {
      if (el.checked) kinds.add(el.dataset.kind);
    });
    const onlyErr = document.getElementById('logsOnlyErr').checked;

    const filtered = logsData.filter((e) => {
      if (!kinds.has(e.kind)) return false;
      if (onlyErr && e.status !== 'err') return false;
      return true;
    });

    // 统计栏
    document.getElementById('logsTotal').textContent = logsData.length;
    document.getElementById('logsErrs').textContent = logsData.filter((e) => e.status === 'err').length;
    document.getElementById('logsShell').textContent = logsData.filter((e) => e.kind === 'shell').length;
    document.getElementById('logsApi').textContent = logsData.filter((e) => e.kind === 'api').length;
    document.getElementById('logsHttp').textContent = logsData.filter((e) => e.kind === 'http').length;

    const list = document.getElementById('logsList');
    if (filtered.length === 0) {
      list.innerHTML = '<div class="logs-empty">没有符合筛选条件的日志</div>';
      return;
    }
    list.innerHTML = '';
    filtered.forEach((e, idx) => {
      const key = e.time + '|' + e.op + '|' + idx;
      const t = new Date(e.time);
      const hh = String(t.getHours()).padStart(2, '0');
      const mm = String(t.getMinutes()).padStart(2, '0');
      const ss = String(t.getSeconds()).padStart(2, '0');
      const ms = String(t.getMilliseconds()).padStart(3, '0');
      const timeStr = `${hh}:${mm}:${ss}.${ms}`;
      const statusIcon = e.status === 'err' ? '✗' : '✓';
      const dur = e.duration_ms ? e.duration_ms + 'ms' : '—';

      const row = document.createElement('div');
      row.className = 'log-row' + (e.status === 'err' ? ' err' : '');
      row.innerHTML = `
        <span class="log-time">${timeStr}</span>
        <span class="log-kind ${e.kind}">${e.kind}</span>
        <span class="log-op-detail">
          <span class="log-op">${escapeHTML(e.op)}</span>
          <span class="log-detail">${escapeHTML(e.detail || '')}</span>
        </span>
        <span class="log-dur">${dur}</span>
        <span class="log-status ${e.status}">${statusIcon}</span>
      `;
      row.addEventListener('click', () => {
        if (logsExpanded.has(key)) {
          logsExpanded.delete(key);
        } else {
          logsExpanded.add(key);
        }
        renderLogs();
      });
      list.appendChild(row);

      if (logsExpanded.has(key)) {
        const exp = document.createElement('div');
        exp.className = 'log-row-expanded';
        let body = `时间：${t.toLocaleString('zh-CN', { hour12: false })}.${ms}\n`;
        body += `类型：${e.kind}\n操作：${e.op}\n`;
        if (e.detail) body += `详情：${e.detail}\n`;
        body += `耗时：${dur}\n状态：${e.status}`;
        exp.textContent = body;
        if (e.error) {
          const errEl = document.createElement('div');
          errEl.className = 'err-line';
          errEl.textContent = '错误：' + e.error;
          exp.appendChild(errEl);
        }
        list.appendChild(exp);
      }
    });
  }

  document.getElementById('logsRefreshBtn').addEventListener('click', loadLogs);
  document.getElementById('logsClearBtn').addEventListener('click', () => {
    fetch('/api/logs/clear', { method: 'POST' }).then(() => {
      logsExpanded.clear();
      loadLogs();
    });
  });
  document.querySelectorAll('.logs-filters input').forEach((el) => {
    el.addEventListener('change', renderLogs);
  });
  document.getElementById('logsAuto').addEventListener('change', (e) => {
    if (logsAutoTimer) { clearInterval(logsAutoTimer); logsAutoTimer = null; }
    if (e.target.checked) {
      logsAutoTimer = setInterval(loadLogs, 2000);
    }
  });

  // Tab 切进来时启动自动刷新，离开时停止
  document.querySelectorAll('.navitem').forEach((btn) => {
    if (btn.dataset.tab === 'logs') {
      btn.addEventListener('click', () => {
        loadLogs();
        if (document.getElementById('logsAuto').checked && !logsAutoTimer) {
          logsAutoTimer = setInterval(loadLogs, 2000);
        }
      });
    } else {
      btn.addEventListener('click', () => {
        if (logsAutoTimer) { clearInterval(logsAutoTimer); logsAutoTimer = null; }
      });
    }
  });
  // 启动时加载一次（给侧栏计数）
  loadLogs();

  // ── 强制约束 Hooks 管理 ──────────────────────────────────────
  let hooksScope = 'global';        // 当前作用域
  let hooksProjectPath = '';        // scope=project 时的项目路径

  // 作用域切换按钮
  document.querySelectorAll('#hooksGuardScopeBar .scope-btn').forEach((btn) => {
    btn.addEventListener('click', () => {
      document.querySelectorAll('#hooksGuardScopeBar .scope-btn').forEach((b) => b.classList.remove('active'));
      btn.classList.add('active');
      hooksScope = btn.dataset.scope;
      const pathRow  = document.getElementById('hooksGuardProjectPath');
      const hint     = document.getElementById('hooksGuardScopeHint');
      if (hooksScope === 'project') {
        pathRow.style.display = '';
        hint.innerHTML = `安装到 <code>&lt;项目&gt;/.claude/hooks/</code>，仅对当前项目生效`;
        // 路径未设定前不加载列表
        if (!hooksProjectPath) return;
      } else {
        pathRow.style.display = 'none';
        hint.innerHTML = `安装到 <code>~/.claude/hooks/</code>，对所有项目生效`;
      }
      loadHooksGuard();
    });
  });

  // 项目路径浏览
  const hooksGuardBrowseBtn = document.getElementById('hooksGuardPathBrowse');
  if (hooksGuardBrowseBtn) {
    hooksGuardBrowseBtn.addEventListener('click', () => {
      hooksGuardBrowseBtn.disabled = true; hooksGuardBrowseBtn.textContent = '⏳';
      fetch('/api/scaffold/pick-folder?prompt=' + encodeURIComponent('选择项目根目录'), { method: 'POST' })
        .then((r) => r.json()).then((d) => {
          hooksGuardBrowseBtn.disabled = false; hooksGuardBrowseBtn.textContent = '📁 选目录';
          if (d.path) {
            document.getElementById('hooksGuardPathInput').value = d.path;
            hooksProjectPath = d.path;
            loadHooksGuard();
          }
        }).catch(() => { hooksGuardBrowseBtn.disabled = false; hooksGuardBrowseBtn.textContent = '📁 选目录'; });
    });
  }

  // 项目路径确认
  document.getElementById('hooksGuardPathConfirm').addEventListener('click', () => {
    const val = document.getElementById('hooksGuardPathInput').value.trim();
    if (!val) return;
    hooksProjectPath = val;
    loadHooksGuard();
  });

  const HOOK_TYPE_LABEL = {
    UserPromptSubmit: '🟣 提问前触发',
    PostToolUse:      '🟡 写文件后触发',
    PreToolUse:       '🔴 执行命令前触发',
    'UserPromptSubmit+PreToolUse+PostToolUse': '⚡ 全时机触发',
  };

  function hookBody(extra) {
    return JSON.stringify({ scope: hooksScope, path: hooksProjectPath, ...extra });
  }

  function renderHookCard(h) {
    const typeLabel  = HOOK_TYPE_LABEL[h.hook_type] || h.hook_type;
    // 图标从 name 中提取第一个 emoji（格式：emoji 名称）
    const iconMatch = h.name.match(/^(\p{Emoji_Presentation}|\p{Emoji}️)/u);
    const icon = iconMatch ? iconMatch[0] : '🔒';
    const plainName = h.name.replace(/^(\p{Emoji_Presentation}|\p{Emoji}️)\s*/u, '');

    // hook_type → CSS class（badge 用）
    const typeClass = h.hook_type === 'PreToolUse' ? 'pre'
      : h.hook_type === 'UserPromptSubmit' ? 'prompt'
      : '';  // PostToolUse = default indigo
    // icon 背景 class
    const iconClass = h.hook_type === 'PreToolUse' ? 'type-pre'
      : h.hook_type === 'UserPromptSubmit' ? 'type-prompt'
      : h.hook_type.includes('+') ? 'type-all'
      : '';  // PostToolUse = default indigo

    // 状态 badge
    let badgeClass, badgeText;
    if (!h.installed)        { badgeClass = 'missing'; badgeText = '● 未安装'; }
    else if (h.enabled)      { badgeClass = 'on';      badgeText = '● 运行中'; }
    else                     { badgeClass = 'off';     badgeText = '● 已暂停'; }

    const card = document.createElement('div');
    card.className = 'hook-card';
    card.dataset.id = h.id;
    card.innerHTML = `
      <div class="hook-card-icon ${iconClass}">${icon}</div>
      <div class="hook-card-body">
        <div class="hook-card-name">${escapeHTML(plainName)}</div>
        <span class="hook-card-type ${typeClass}">${typeLabel}</span>
        <p class="hook-card-desc">${escapeHTML(h.description)}</p>
      </div>
      <div class="hook-card-right">
        <span class="hook-status-badge ${badgeClass}">${badgeText}</span>
        <div class="hook-card-actions">
          ${!h.installed ? `<button class="btn btn-primary btn-sm hook-install-btn">安装</button>` : ''}
          ${h.installed && h.enabled  ? `<button class="btn btn-ghost btn-sm hook-toggle-btn">暂停</button>` : ''}
          ${h.installed && !h.enabled ? `<button class="btn btn-primary btn-sm hook-toggle-btn">启用</button>` : ''}
          ${h.installed ? `<button class="btn btn-ghost btn-sm danger hook-uninstall-btn">卸载</button>` : ''}
        </div>
      </div>
    `;

    const installBtn   = card.querySelector('.hook-install-btn');
    const toggleBtn    = card.querySelector('.hook-toggle-btn');
    const uninstallBtn = card.querySelector('.hook-uninstall-btn');

    if (installBtn) {
      installBtn.addEventListener('click', () => {
        installBtn.disabled = true; installBtn.textContent = '⏳';
        fetch('/api/hooks/install', {
          method: 'POST', headers: {'Content-Type': 'application/json'},
          body: hookBody({id: h.id}),
        }).then((r) => r.json()).then((res) => {
          if (res.ok) {
            let msg = `✓ 已安装：${h.name}\n路径：${res.path}`;
            if (res.settings_note) msg += `\n\n${res.settings_note}`;
            showLog('hooksGuardLog', msg, 'ok');
            loadHooksGuard();
          } else {
            showLog('hooksGuardLog', `✗ 安装失败：${res.error || '未知'}`, 'err');
            installBtn.disabled = false; installBtn.textContent = '安装';
          }
        }).catch((e) => { showLog('hooksGuardLog', '✗ 请求失败：' + e, 'err'); installBtn.disabled = false; installBtn.textContent = '安装'; });
      });
    }

    if (toggleBtn) {
      toggleBtn.addEventListener('click', () => {
        toggleBtn.disabled = true;
        fetch('/api/hooks/toggle', {
          method: 'POST', headers: {'Content-Type': 'application/json'},
          body: hookBody({id: h.id}),
        }).then((r) => r.json()).then((res) => {
          if (res.ok) {
            showLog('hooksGuardLog', res.enabled ? `✓ 已启用：${h.name}` : `✓ 已禁用：${h.name}`, 'ok');
            loadHooksGuard();
          } else {
            showLog('hooksGuardLog', `✗ 切换失败：${res.error || '未知'}`, 'err');
            toggleBtn.disabled = false;
          }
        }).catch((e) => { showLog('hooksGuardLog', '✗ 请求失败：' + e, 'err'); toggleBtn.disabled = false; });
      });
    }

    if (uninstallBtn) {
      twoStep(uninstallBtn, `确认卸载 ${h.name}？`, () => {
        fetch('/api/hooks/uninstall', {
          method: 'POST', headers: {'Content-Type': 'application/json'},
          body: hookBody({id: h.id}),
        }).then((r) => r.json()).then((res) => {
          showLog('hooksGuardLog', res.ok ? `✓ 已卸载：${h.name}` : `✗ 卸载失败：${res.error}`, res.ok ? 'ok' : 'err');
          loadHooksGuard();
        }).catch((e) => showLog('hooksGuardLog', '✗ 请求失败：' + e, 'err'));
      });
    }

    return card;
  }

  function loadHooksGuard() {
    const qs = hooksScope === 'project' && hooksProjectPath
      ? `?scope=project&path=${encodeURIComponent(hooksProjectPath)}`
      : '?scope=global';
    fetch('/api/hooks' + qs).then((r) => r.json()).then((res) => {
      const list = document.getElementById('hooksGuardList');
      if (!list) return;
      list.innerHTML = '';
      const hooks = res.hooks || [];
      hooks.forEach((h) => list.appendChild(renderHookCard(h)));

      // 侧栏徽章：已启用数量
      const enabledCount = hooks.filter((h) => h.enabled).length;
      const badge = document.getElementById('hooksEnabledCount');
      if (badge) {
        if (enabledCount > 0) {
          badge.textContent = enabledCount;
          badge.style.display = '';
        } else {
          badge.style.display = 'none';
        }
      }
    }).catch((e) => showLog('hooksGuardLog', '✗ 加载失败：' + e, 'err'));
  }

  // 进入 tab 时加载
  document.querySelectorAll('.navitem').forEach((btn) => {
    if (btn.dataset.tab === 'hooks-guard') {
      btn.addEventListener('click', loadHooksGuard);
    }
  });
  // 启动时加载一次（更新侧栏徽章）
  loadHooksGuard();

  // ---------- 自定义 Skill ----------
  let editingRuleId = null;

  function loadCustomRules() {
    fetch('/api/custom')
      .then((r) => r.json())
      .then((d) => {
        const rules = d.rules || [];
        document.getElementById('customCount').textContent = rules.length;
        const list = document.getElementById('customList');
        if (rules.length === 0) {
          list.innerHTML = '<div class="empty">尚未添加 Skill。从下方预设模板开始，或点"+ 新建"。</div>';
          return;
        }
        list.innerHTML = '';
        rules.forEach((r) => {
          const box = document.createElement('div');
          box.className = 'custom-item';
          // 缩减 scope 文案：permanent/按文件，不重复把 glob 塞到胶囊里
          const scopeShort = r.apply_to === 'always'
            ? '永远'
            : (r.globs && r.globs.length > 24
                ? r.globs.split(',')[0].trim() + '…'   // 多 glob 只显示第一条
                : (r.globs || '**/*.go'));
          const scopeFull = r.apply_to === 'always' ? '永远生效' : 'Globs: ' + (r.globs || '**/*.go');
          box.innerHTML = `
            <div class="custom-item-head">
              <strong title="${escapeHTML(r.title)}">${escapeHTML(r.title)}</strong>
              <span class="custom-scope" title="${escapeHTML(scopeFull)}">${escapeHTML(scopeShort)}</span>
            </div>
            <div class="custom-item-desc" title="${escapeHTML(r.description || '')}">${escapeHTML(r.description || '')}</div>
            <div class="custom-item-actions">
              <button class="btn btn-ghost btn-sm" data-act="edit" data-id="${r.id}">编辑</button>
              <button class="btn btn-ghost btn-sm danger" data-act="del" data-id="${r.id}">删除</button>
            </div>
          `;
          list.appendChild(box);
        });
        list.querySelectorAll('button[data-act]').forEach((btn) => {
          btn.addEventListener('click', () => {
            const id = btn.dataset.id;
            if (btn.dataset.act === 'edit') {
              const target = rules.find((x) => x.id === id);
              if (target) fillForm(target);
              return;
            }
            if (btn.dataset.act === 'del') {
              // WKWebView 里 window.confirm() 不工作，改"再点一次确认"两步式
              if (btn.dataset.confirming !== '1') {
                btn.dataset.confirming = '1';
                btn.dataset.origText = btn.textContent;
                btn.textContent = '再点一次确认';
                btn.classList.add('danger-active');
                const timer = setTimeout(() => {
                  delete btn.dataset.confirming;
                  btn.textContent = btn.dataset.origText || '删除';
                  btn.classList.remove('danger-active');
                }, 3000);
                btn.dataset.timer = timer;
                return;
              }
              clearTimeout(parseInt(btn.dataset.timer, 10));
              deleteRule(id);
            }
          });
        });
      });
  }

  function loadPresets() {
    fetch('/api/custom/presets')
      .then((r) => r.json())
      .then((d) => {
        const list = document.getElementById('presetList');
        list.innerHTML = '';
        (d.presets || []).forEach((p) => {
          const box = document.createElement('div');
          box.className = 'preset-item';
          box.innerHTML = `
            <div class="preset-title" title="${escapeHTML(p.title)}">${escapeHTML(p.title)}</div>
            <div class="preset-desc" title="${escapeHTML(p.description)}">${escapeHTML(p.description)}</div>
            <button class="btn btn-ghost btn-sm">👉 套用</button>
          `;
          box.querySelector('button').addEventListener('click', () => {
            fillForm({ ...p, id: '' });
            showLog('customLog', '已套用预设，修改后点保存', '');
          });
          list.appendChild(box);
        });
      });
  }

  function fillForm(rule) {
    editingRuleId = rule.id || null;
    document.getElementById('formTitle').textContent = editingRuleId ? '编辑 Skill' : '新建 Skill';
    document.getElementById('ruleTitle').value = rule.title || '';
    document.getElementById('ruleDesc').value = rule.description || '';
    document.getElementById('ruleReason').value = rule.reason || '';
    document.getElementById('ruleContent').value = rule.content || '';
    document.getElementById('ruleGlobs').value = rule.globs || '**/*.go';
    const applyTo = rule.apply_to || 'globs';
    document.querySelectorAll('input[name="applyTo"]').forEach((r) => (r.checked = r.value === applyTo));
    document.getElementById('globsCol').style.display = applyTo === 'always' ? 'none' : '';
    document.getElementById('cancelRuleBtn').style.display = editingRuleId ? '' : 'none';
  }

  function resetForm() {
    editingRuleId = null;
    document.getElementById('formTitle').textContent = '新建 Skill';
    ['ruleTitle', 'ruleDesc', 'ruleReason', 'ruleContent'].forEach((id) => (document.getElementById(id).value = ''));
    document.getElementById('ruleGlobs').value = '**/*.go';
    document.querySelectorAll('input[name="applyTo"]').forEach((r) => (r.checked = r.value === 'globs'));
    document.getElementById('globsCol').style.display = '';
    document.getElementById('cancelRuleBtn').style.display = 'none';
    document.getElementById('customLog').classList.remove('show');
  }

  document.getElementById('newRuleBtn').addEventListener('click', resetForm);
  document.getElementById('cancelRuleBtn').addEventListener('click', resetForm);

  document.querySelectorAll('input[name="applyTo"]').forEach((r) => {
    r.addEventListener('change', (e) => {
      document.getElementById('globsCol').style.display = e.target.value === 'always' ? 'none' : '';
    });
  });

  document.getElementById('saveRuleBtn').addEventListener('click', () => {
    const body = {
      id: editingRuleId || '',
      title: document.getElementById('ruleTitle').value.trim(),
      description: document.getElementById('ruleDesc').value.trim(),
      reason: document.getElementById('ruleReason').value.trim(),
      content: document.getElementById('ruleContent').value.trim(),
      apply_to: document.querySelector('input[name="applyTo"]:checked').value,
      globs: document.getElementById('ruleGlobs').value.trim(),
    };
    if (!body.title || !body.content || !body.description) {
      showLog('customLog', '✗ 规则名 / 简介 / 内容 都不能为空', 'err');
      return;
    }
    fetch('/api/custom', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })
      .then((r) => r.json())
      .then((res) => {
        if (res.error) { showLog('customLog', '✗ ' + res.error, 'err'); return; }
        let msg = '✓ 已保存：' + res.rule.title + '\n';
        if (res.synced && res.synced.length) {
          msg += '\n已同步到：\n' + res.synced.map((x) => '  · ' + x).join('\n');
        } else {
          msg += '\n（Claude / Cursor 尚未安装，先去"安装"Tab 装一次即可生效）';
        }
        showLog('customLog', msg, 'ok');
        resetForm();
        loadCustomRules();
      });
  });

  function deleteRule(id) {
    fetch('/api/custom?id=' + encodeURIComponent(id), { method: 'DELETE' })
      .then((r) => r.json())
      .then((res) => {
        if (res.error) {
          showLog('customLog', '✗ 删除失败：' + res.error, 'err');
          return;
        }
        let msg = '✓ 已删除并同步';
        if (res.synced && res.synced.length) {
          msg += '\n已清理：\n' + res.synced.map((x) => '  · ' + x).join('\n');
        }
        showLog('customLog', msg, 'ok');
        loadCustomRules();
      })
      .catch((e) => {
        showLog('customLog', '✗ 请求失败：' + e, 'err');
      });
  }

  // ---------- 通用 ----------
  function showLog(elId, text, level) {
    const el = document.getElementById(elId);
    el.classList.remove('ok', 'err');
    if (level) el.classList.add(level);
    el.classList.add('show');
    el.textContent = text;
  }

  function escapeHTML(s) {
    return String(s).replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
  }

  // ---------- 📋 提交规范化（v1.7.37）----------
  function commitGuardCheckStatus() {
    const path = document.getElementById('commitGuardPath').value.trim();
    if (!path) return;

    fetch('/api/commit-guard/status?path=' + encodeURIComponent(path))
      .then((r) => r.json())
      .then((d) => {
        const statusEl = document.getElementById('commitGuardStatus');
        statusEl.style.display = 'block';
        if (!d.valid) {
          statusEl.className = 'tixuebj-status missing';
          statusEl.textContent = '✗ ' + (d.error || '路径不合法');
          document.getElementById('commitGuardInstallBtn').disabled = true;
          document.getElementById('commitGuardCheckBtn').disabled = true;
          document.getElementById('commitGuardUninstallBtn').disabled = true;
          return;
        }
        let msg = '✓ 是 git 仓库 · ';
        msg += d.is_team_hook ? '已装 Team hook ✓' :
               d.hook_exists ? '有其它 pre-commit hook（装时会备份）' : '无 pre-commit hook';
        msg += ' · ' + (d.script_exists ? 'scripts/team-standards-check.sh ✓' : '无检查脚本');
        statusEl.className = 'tixuebj-status ' + (d.is_team_hook && d.script_exists ? 'installed' : 'partial');
        statusEl.textContent = msg;

        document.getElementById('commitGuardInstallBtn').disabled = false;
        document.getElementById('commitGuardCheckBtn').disabled = !d.script_exists;
        document.getElementById('commitGuardUninstallBtn').disabled = !(d.is_team_hook || d.script_exists);
      })
      .catch((e) => {
        const statusEl = document.getElementById('commitGuardStatus');
        statusEl.style.display = 'block';
        statusEl.className = 'tixuebj-status missing';
        statusEl.textContent = '✗ 检查失败：' + e;
      });
  }

  // 加载 CI snippet 一次（缓存）
  let commitGuardScriptsCache = null;
  function loadCommitGuardScripts() {
    fetch('/api/commit-guard/scripts').then((r) => r.json()).then((scripts) => {
      commitGuardScriptsCache = scripts;
      const ciEl = document.getElementById('commitGuardCISnippet');
      if (ciEl && scripts['gitlab-ci-snippet.yml']) {
        ciEl.textContent = scripts['gitlab-ci-snippet.yml'];
      }
    });
  }

  if (document.getElementById('commitGuardPath')) {
    loadCommitGuardScripts();

    document.getElementById('commitGuardCheckPathBtn').addEventListener('click', commitGuardCheckStatus);

    document.getElementById('commitGuardPath').addEventListener('keypress', (e) => {
      if (e.key === 'Enter') {
        e.preventDefault();
        commitGuardCheckStatus();
      }
    });

    document.getElementById('commitGuardInstallBtn').addEventListener('click', () => {
      const path = document.getElementById('commitGuardPath').value.trim();
      const btn = document.getElementById('commitGuardInstallBtn');
      btn.disabled = true; btn.textContent = '⏳ 安装中…';
      fetch('/api/commit-guard/install', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path }),
      }).then((r) => r.json()).then((res) => {
        btn.disabled = false; btn.textContent = '📥 装 pre-commit hook';
        let msg = res.ok ? '✓ 装好了\n' : '✗ 失败\n';
        if (res.hook_path) msg += '  hook: ' + res.hook_path + '\n';
        if (res.script_path) msg += '  script: ' + res.script_path + '\n';
        if (res.backup_path) msg += '  ⚠ 原 pre-commit hook 已备份到: ' + res.backup_path + '\n';
        if (res.message) msg += '\n' + res.message;
        if (res.error) msg += '\n✗ ' + res.error + '\n' + (res.detail || '');
        showLog('commitGuardLog', msg, res.ok ? 'ok' : 'err');
        commitGuardCheckStatus();
      });
    });

    document.getElementById('commitGuardCheckBtn').addEventListener('click', () => {
      const path = document.getElementById('commitGuardPath').value.trim();
      const btn = document.getElementById('commitGuardCheckBtn');
      btn.disabled = true; btn.textContent = '⏳ 扫描中…';
      fetch('/api/commit-guard/check', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path, all: true }),
      }).then((r) => r.json()).then((res) => {
        btn.disabled = false; btn.textContent = '🧪 试跑一次（--all 扫整个仓库）';
        let msg = '退出码: ' + res.exit_code + '\n\n' + (res.output || '(空)');
        if (res.error) msg = '✗ ' + res.error + '\n' + (res.detail || '');
        showLog('commitGuardLog', msg, res.exit_code === 0 ? 'ok' : 'err');
      });
    });

    document.getElementById('commitGuardUninstallBtn').addEventListener('click', () => {
      const path = document.getElementById('commitGuardPath').value.trim();
      const btn = document.getElementById('commitGuardUninstallBtn');
      btn.disabled = true; btn.textContent = '⏳ 卸载中…';
      fetch('/api/commit-guard/uninstall', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ path }),
      }).then((r) => r.json()).then((res) => {
        btn.disabled = false; btn.textContent = '🗑 卸载';
        let msg = res.ok ? '✓ 已卸载\n' : '✗ 失败\n';
        (res.removed || []).forEach((p) => msg += '  · 删 ' + p + '\n');
        if (res.note) msg += '\n' + res.note;
        showLog('commitGuardLog', msg, res.ok ? 'ok' : 'err');
        commitGuardCheckStatus();
      });
    });

    document.getElementById('commitGuardCopyCIBtn').addEventListener('click', (e) => {
      const txt = document.getElementById('commitGuardCISnippet').textContent;
      navigator.clipboard.writeText(txt).then(() => {
        const btn = e.currentTarget;
        const orig = btn.textContent;
        btn.textContent = '✓ 已复制';
        setTimeout(() => { btn.textContent = orig; }, 1500);
      });
    });
  }

  // ---------- 📡 Claude Desktop / Cowork 直装（v1.7.35）----------
  let coworkProbeData = null;

  function renderCoworkProbe(d) {
    const root = document.getElementById('coworkProbeResult');
    coworkProbeData = d;

    let html = '<div style="font-size: 12.5px; color: var(--label-2); margin-bottom: 10px;">勾选要装的目标路径（推荐勾「父目录 ✓」+「可写 ✓」的）：</div>';

    // 候选路径
    html += '<div style="background: var(--bg-card-2); border: 0.5px solid var(--separator); border-radius: 8px; padding: 10px;">';
    (d.candidates || []).forEach((c, i) => {
      const score = (c.parent_exists ? 2 : 0) + (c.writable ? 1 : 0) + (c.exists ? 1 : 0);
      const recommend = c.parent_exists && c.writable;
      const sourceTag = c.source === 'claude-desktop'
        ? '<span style="background:rgba(0,122,255,0.12);color:#0040DD;padding:1px 6px;border-radius:4px;font-size:10px;">Claude Desktop</span>'
        : '<span style="background:rgba(175,82,222,0.12);color:#7E2EAB;padding:1px 6px;border-radius:4px;font-size:10px;">Cowork</span>';
      html += `
        <label style="display: grid; grid-template-columns: 18px 1fr auto; gap: 8px; padding: 6px 4px; border-bottom: 0.5px solid var(--separator); cursor: pointer; align-items: center;">
          <input type="checkbox" class="cowork-path-check" data-path="${escapeHTML(c.path)}" ${recommend ? 'checked' : ''} style="margin: 0;">
          <div style="min-width: 0;">
            <div style="font-size: 12.5px;">${sourceTag} <strong>${escapeHTML(c.name)}</strong></div>
            <div style="font-family: var(--mono); font-size: 11px; color: var(--label-3); margin-top: 2px; white-space: nowrap; overflow: hidden; text-overflow: ellipsis;">${escapeHTML(c.path)}</div>
            <div style="font-size: 11px; color: var(--label-3); margin-top: 2px;">${escapeHTML(c.note)}</div>
          </div>
          <div style="font-size: 11px; white-space: nowrap;">
            ${c.parent_exists ? '<span style="color:var(--sys-green);">父目录 ✓</span>' : '<span style="color:var(--label-4);">父目录 ✗</span>'}
            ${c.exists ? ' · <span style="color:var(--sys-green);">已有 skills/ ✓</span>' : ''}
            ${c.writable ? ' · <span style="color:var(--sys-green);">可写 ✓</span>' : ''}
          </div>
        </label>
      `;
    });
    html += '</div>';

    // 扫到的 Anthropic 相关目录
    if (d.discovered_dirs && d.discovered_dirs.length > 0) {
      html += '<details style="margin-top: 10px;"><summary style="cursor:pointer; font-size: 12.5px; color: var(--label-2);">🔍 扫到的所有 Anthropic / Claude / Cowork 相关目录（' + d.discovered_dirs.length + ' 个）</summary>';
      html += '<div style="margin-top: 6px; padding: 8px 10px; background: rgba(0,0,0,0.03); border-radius: 6px; font-family: var(--mono); font-size: 11.5px;">';
      d.discovered_dirs.forEach((dir) => {
        html += '<div style="padding: 2px 0;">' + escapeHTML(dir.path);
        if (dir.has_skill_subdir) html += ' <span style="color: var(--sys-green);">[含 skills/]</span>';
        html += '</div>';
      });
      html += '</div></details>';
    }

    // 自定义路径
    html += `
      <details style="margin-top: 10px;">
        <summary style="cursor:pointer; font-size: 12.5px; color: var(--label-2);">➕ 自定义路径（手填，必须在 $HOME 内）</summary>
        <div style="margin-top: 8px; display: flex; gap: 6px;">
          <input type="text" id="coworkCustomPath" class="input"
                 placeholder="${escapeHTML(d.home)}/Library/Application Support/Cowork/skills"
                 style="flex: 1; font-family: var(--mono); font-size: 12px;">
          <button id="coworkAddCustomBtn" class="btn btn-sm">添加</button>
        </div>
        <div id="coworkCustomList" style="margin-top: 6px; font-family: var(--mono); font-size: 11.5px; color: var(--label-2);"></div>
      </details>
    `;

    root.innerHTML = html;

    // 勾选变化时刷新安装按钮状态
    function refreshInstallBtn() {
      const checked = document.querySelectorAll('.cowork-path-check:checked').length
                    + document.querySelectorAll('#coworkCustomList .custom-item').length;
      document.getElementById('coworkInstallBtn').disabled = checked === 0;
    }
    document.querySelectorAll('.cowork-path-check').forEach((cb) =>
      cb.addEventListener('change', refreshInstallBtn));
    refreshInstallBtn();

    // 自定义路径添加
    document.getElementById('coworkAddCustomBtn').addEventListener('click', () => {
      const input = document.getElementById('coworkCustomPath');
      const v = input.value.trim();
      if (!v) return;
      const list = document.getElementById('coworkCustomList');
      const div = document.createElement('div');
      div.className = 'custom-item';
      div.style.cssText = 'padding: 4px 8px; margin-top: 4px; background: rgba(175,82,222,0.10); border-radius: 4px; display: flex; justify-content: space-between;';
      div.innerHTML = `<span data-path="${escapeHTML(v)}">${escapeHTML(v)}</span><button class="btn btn-sm" style="padding: 1px 8px;">删</button>`;
      div.querySelector('button').addEventListener('click', () => {
        div.remove();
        refreshInstallBtn();
      });
      list.appendChild(div);
      input.value = '';
      refreshInstallBtn();
    });
  }

  if (document.getElementById('coworkProbeBtn')) {
    document.getElementById('coworkProbeBtn').addEventListener('click', () => {
      const btn = document.getElementById('coworkProbeBtn');
      btn.disabled = true; btn.textContent = '探测中…';
      fetch('/api/cowork/probe').then((r) => r.json()).then((d) => {
        btn.disabled = false; btn.textContent = '🔍 重新探测';
        renderCoworkProbe(d);
      });
    });

    document.getElementById('coworkInstallBtn').addEventListener('click', () => {
      const paths = Array.from(document.querySelectorAll('.cowork-path-check:checked'))
                       .map((cb) => cb.dataset.path);
      const customPaths = Array.from(document.querySelectorAll('#coworkCustomList .custom-item span'))
                            .map((s) => s.dataset.path);
      const allPaths = paths.concat(customPaths);
      if (allPaths.length === 0) return;

      const btn = document.getElementById('coworkInstallBtn');
      btn.disabled = true; btn.textContent = '⏳ 安装中…';
      fetch('/api/cowork/install', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ paths: allPaths }),
      }).then((r) => r.json()).then((res) => {
        btn.disabled = false; btn.textContent = '📥 安装到勾选路径';
        let msg = res.ok ? '✓ 安装完成\n' : '✗ 部分失败\n';
        (res.installed || []).forEach((l) => msg += '  · ' + l + '\n');
        (res.warnings || []).forEach((w) => msg += '  ⚠ ' + w + '\n');
        if (res.note) msg += '\n' + res.note;
        showLog('coworkInstallLog', msg, res.ok ? 'ok' : 'err');
      });
    });
  }

  // ---------- 🆕 App 自动更新检查（v1.7.34）----------
  let updateInfo = null;

  function checkAppUpdate() {
    fetch('/api/app-update/check').then((r) => r.json()).then((d) => {
      updateInfo = d;
      if (!d.reachable) {
        // 网络不通——静默不弹（用户不在乎）
        return;
      }
      if (!d.has_update) return;             // 已是最新
      if (d.is_skipped) return;              // 用户跳过过这个版本

      // 显示顶部 banner
      const banner = document.getElementById('updateBanner');
      const text = document.getElementById('updateBannerText');
      text.innerHTML = `<b>Team Standards ${escapeHTML(d.latest_version)}</b> 可用 · 当前 ${escapeHTML(d.current_version)}`;
      banner.classList.add('show');
    }).catch(() => {
      // 静默失败
    });
  }

  // banner 操作
  if (document.getElementById('updateBanner')) {
    document.getElementById('updateBannerView').addEventListener('click', showUpdateModal);
    document.getElementById('updateBannerLater').addEventListener('click', () => {
      document.getElementById('updateBanner').classList.remove('show');
    });
    document.getElementById('updateBannerSkip').addEventListener('click', () => {
      if (!updateInfo) return;
      fetch('/api/app-update/skip', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ version: updateInfo.latest_version }),
      }).then(() => {
        document.getElementById('updateBanner').classList.remove('show');
      });
    });
  }

  // modal 操作
  function showUpdateModal() {
    if (!updateInfo) return;
    document.getElementById('updateModalTitle').textContent =
      `Team Standards ${updateInfo.latest_version} 可用`;
    document.getElementById('updateModalSub').textContent =
      `当前 ${updateInfo.current_version} · 发布于 ${new Date(updateInfo.released_at).toLocaleString()}`;
    document.getElementById('updateModalNotes').textContent =
      updateInfo.release_notes || '(无 release notes)';
    document.getElementById('updateModal').classList.add('show');
  }

  if (document.getElementById('updateModal')) {
    document.getElementById('updateModalClose').addEventListener('click', () => {
      document.getElementById('updateModal').classList.remove('show');
    });

    // 🚀 一键升级（v1.7.38）
    document.getElementById('updateAutoInstallBtn').addEventListener('click', () => {
      // 关闭详情 modal + banner
      document.getElementById('updateModal').classList.remove('show');
      document.getElementById('updateBanner').classList.remove('show');

      // 弹进度 modal
      const progModal = document.getElementById('upgradeProgressModal');
      progModal.classList.add('show');
      document.getElementById('upgradeProgressPhase').textContent = '启动…';
      document.getElementById('upgradeProgressMsg').textContent = '请求开始升级…';
      document.getElementById('upgradeProgressBar').style.width = '0%';
      document.getElementById('upgradeProgressDetail').textContent = '';
      document.getElementById('upgradeProgressNote').style.display = 'none';
      document.getElementById('upgradeProgressActions').style.display = 'none';

      fetch('/api/app-update/auto-install', { method: 'POST' })
        .then((r) => r.json())
        .then((res) => {
          if (!res.ok) {
            showUpgradeFailed(res.error || '启动失败');
            return;
          }
          // 进入 polling
          pollUpgradeProgress();
        }).catch((e) => showUpgradeFailed('网络错误：' + e));
    });

    let pollUpgradeTimer = null;
    function pollUpgradeProgress() {
      fetch('/api/app-update/progress').then((r) => r.json()).then((p) => {
        const phaseEl = document.getElementById('upgradeProgressPhase');
        const msgEl = document.getElementById('upgradeProgressMsg');
        const barEl = document.getElementById('upgradeProgressBar');
        const detailEl = document.getElementById('upgradeProgressDetail');

        const phaseTxt = {
          'downloading': '⏬ 下载中',
          'ready': '✓ 准备完成',
          'failed': '✗ 失败',
        }[p.phase] || p.phase;

        phaseEl.textContent = phaseTxt;
        msgEl.textContent = p.message || '';

        if (p.bytes_total > 0) {
          const mbDone = (p.bytes_done / 1024 / 1024).toFixed(1);
          const mbTotal = (p.bytes_total / 1024 / 1024).toFixed(1);
          barEl.style.width = p.percent + '%';
          detailEl.textContent = `${mbDone} MB / ${mbTotal} MB · ${p.percent}%`;
        }

        if (p.phase === 'ready') {
          // 准备好了 → 提示用户 + 倒计时退出
          if (pollUpgradeTimer) { clearTimeout(pollUpgradeTimer); pollUpgradeTimer = null; }
          barEl.style.width = '100%';
          const note = document.getElementById('upgradeProgressNote');
          note.style.display = 'block';
          note.innerHTML = '<strong>✓ 下载完成。3 秒后 App 自动退出，helper 接手安装。</strong><br>helper 会：备份当前 → 装新版 → 解 Gatekeeper → 自动重启 App<br>整个过程约 10 秒。';
          let cd = 3;
          const tick = () => {
            cd--;
            if (cd > 0) {
              note.querySelector('strong').textContent = `✓ 下载完成。${cd} 秒后 App 自动退出，helper 接手安装。`;
              setTimeout(tick, 1000);
            } else {
              // call appQuit
              if (typeof window.appQuit === 'function') {
                window.appQuit();
              } else {
                note.innerHTML = '<strong>✗ appQuit 桥不可用</strong>，请手动 Cmd+Q 退出 App。<br>helper 已经 spawn，等你退出后接手。';
              }
            }
          };
          setTimeout(tick, 1000);
          return;
        }

        if (p.phase === 'failed') {
          showUpgradeFailed(p.error);
          return;
        }

        // 继续 polling
        pollUpgradeTimer = setTimeout(pollUpgradeProgress, 500);
      }).catch((e) => {
        showUpgradeFailed('查询进度失败：' + e);
      });
    }

    function showUpgradeFailed(err) {
      if (pollUpgradeTimer) { clearTimeout(pollUpgradeTimer); pollUpgradeTimer = null; }
      document.getElementById('upgradeProgressPhase').textContent = '✗ 升级失败';
      document.getElementById('upgradeProgressMsg').textContent = err;
      const note = document.getElementById('upgradeProgressNote');
      note.style.display = 'block';
      note.innerHTML = `<strong>升级未完成</strong>　原因：${escapeHTML(err)}<br><br>`
        + `<button id="failedDirectDownloadBtn" class="btn btn-primary" style="margin-top:4px;">💽 立即下载 DMG 手动安装</button>`;
      document.getElementById('upgradeProgressActions').style.display = 'flex';
      // 绑定直接下载
      const dlBtn = document.getElementById('failedDirectDownloadBtn');
      if (dlBtn) {
        dlBtn.addEventListener('click', () => {
          if (!updateInfo || !updateInfo.dmg_url) {
            alert('找不到下载链接，请去 Release 页手动下载：\n' + (updateInfo?.release_url || ''));
            return;
          }
          fetch('/api/app-update/open-download', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ url: updateInfo.dmg_url }),
          });
        });
      }
    }

    document.getElementById('upgradeCloseBtn').addEventListener('click', () => {
      document.getElementById('upgradeProgressModal').classList.remove('show');
    });
    document.getElementById('upgradeViewLogBtn').addEventListener('click', () => {
      fetch('/api/app-update/progress').then((r) => r.json()).then((p) => {
        if (p.log_path) {
          fetch('/api/open-app', {
            method: 'POST', headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ path: p.log_path }),
          });
        }
      });
    });

    document.getElementById('updateDownloadBtn').addEventListener('click', () => {
      if (!updateInfo || !updateInfo.dmg_url) {
        alert('找不到 DMG 下载链接，请去 release 页手工下载：\n' + (updateInfo && updateInfo.release_url || ''));
        return;
      }
      fetch('/api/app-update/open-download', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ url: updateInfo.dmg_url }),
      }).then(() => {
        // 关闭 modal + banner（用户已经知道要下载了）
        document.getElementById('updateModal').classList.remove('show');
        document.getElementById('updateBanner').classList.remove('show');
      });
    });

    document.getElementById('updateCopyXattrBtn').addEventListener('click', (e) => {
      const cmd = 'xattr -cr /Applications/Team\\ Standards.app';
      navigator.clipboard.writeText(cmd).then(() => {
        const btn = e.currentTarget;
        const orig = btn.textContent;
        btn.textContent = '✓ 已复制';
        setTimeout(() => { btn.textContent = orig; }, 1500);
      });
    });

    document.getElementById('updateModalSkipBtn').addEventListener('click', () => {
      if (!updateInfo) return;
      fetch('/api/app-update/skip', {
        method: 'POST', headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ version: updateInfo.latest_version }),
      }).then(() => {
        document.getElementById('updateModal').classList.remove('show');
        document.getElementById('updateBanner').classList.remove('show');
      });
    });
  }

  // 启动后 1 秒异步 check（不阻塞 UI 加载）
  setTimeout(checkAppUpdate, 1000);

  // ---------- 🆕 安装页 App 版本更新卡片 ----------
  function renderAppUpdateCard(d) {
    const curEl = document.getElementById('appCurrentVersion');
    const latEl = document.getElementById('appLatestVersion');
    const installBtn = document.getElementById('appUpdateInstallBtn');
    const log = document.getElementById('appUpdateLog');
    if (!curEl) return;
    curEl.textContent = d.current_version || '-';
    if (!d.reachable) {
      latEl.textContent = '无法连接（网络不通）';
      return;
    }
    latEl.textContent = d.latest_version || '-';
    if (d.has_update && !d.is_skipped) {
      latEl.style.color = 'var(--sys-green)';
      if (installBtn) installBtn.style.display = '';
      if (log) showLog('appUpdateLog', `🆕 发现新版本 ${d.latest_version}，点「立即升级」自动下载安装。`, 'ok');
    } else {
      latEl.style.color = '';
      if (log) showLog('appUpdateLog', '✓ 已是最新版本', 'ok');
    }
  }

  const appUpdateCheckBtn = document.getElementById('appUpdateCheckBtn');
  const appUpdateInstallBtn2 = document.getElementById('appUpdateInstallBtn');
  if (appUpdateCheckBtn) {
    appUpdateCheckBtn.addEventListener('click', () => {
      appUpdateCheckBtn.disabled = true;
      appUpdateCheckBtn.textContent = '检查中…';
      fetch('/api/app-update/check').then((r) => r.json()).then((d) => {
        updateInfo = d;
        renderAppUpdateCard(d);
        appUpdateCheckBtn.disabled = false;
        appUpdateCheckBtn.textContent = '🔍 检查更新';
      }).catch(() => {
        appUpdateCheckBtn.disabled = false;
        appUpdateCheckBtn.textContent = '🔍 检查更新';
      });
    });
    // 切到「更新与同步」tab 时自动刷新
    document.querySelectorAll('.seg-item[data-sub="update"]').forEach((el) => {
      el.addEventListener('click', () => {
        fetch('/api/app-update/check').then((r) => r.json()).then((d) => {
          updateInfo = d;
          renderAppUpdateCard(d);
        });
      });
    });
  }
  if (appUpdateInstallBtn2) {
    appUpdateInstallBtn2.addEventListener('click', () => showUpdateModal());
  }

  // ── 欢迎页：大版本升级后 / 首次安装时显示一次 ──
  // 由 catalog 回调在版本确定后调用，避免版本号还是 "v…" 时就触发
  function showWelcomeIfNeeded(ver) {
    const WELCOME_KEY = 'ts_welcomed_version';
    if (!ver || ver === 'v…') return;
    const seen = localStorage.getItem(WELCOME_KEY) || '';
    if (seen === ver) return;

    const overlay = document.getElementById('welcomeOverlay');
    const note    = document.getElementById('welcomeVersionNote');
    const btn     = document.getElementById('welcomeCloseBtn');
    if (!overlay) return;

    note.textContent = seen
      ? `${seen} → ${ver} 升级完成，所有规范和工具已更新到最新版本`
      : `${ver} · 首次启动，欢迎加入团队规范体系`;

    overlay.style.display = 'flex';

    function closeWelcome() {
      overlay.style.display = 'none';
      localStorage.setItem(WELCOME_KEY, ver);
    }
    btn.addEventListener('click', closeWelcome);
    // 3 秒后自动关闭
    setTimeout(closeWelcome, 3000);
  }

})();
