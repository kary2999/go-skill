// fs-local.js —— 网页版（GitHub Pages）通过 File System Access API 操作本地目录。
//
// 用户授权 ~/.claude 和 ~/.cursor 目录后，以下接口由本文件在浏览器内实现，
// 不再走静态快照 / 只读 stub（见 app.js 顶部 shim 的 __TS_FS__ 路由）：
//   GET  /api/installed                 已装清单（扫描授权目录）
//   POST /api/install                   一键安装（写 web-install-payload.json 里的文件）
//   POST /api/installed-skill-uninstall 按路径卸载
//
// 仅 Chrome / Edge 支持 showDirectoryPicker；目录句柄存 IndexedDB，刷新后可恢复
//（重新授权只需点一次确认，不用重选目录）。
(function () {
  'use strict';
  if (!window.__TS_STATIC__) return;
  const supported = typeof window.showDirectoryPicker === 'function';

  // ---------- IndexedDB 句柄持久化 ----------
  const DB = 'ts-fs-handles';
  function idb() {
    return new Promise((resolve, reject) => {
      const req = indexedDB.open(DB, 1);
      req.onupgradeneeded = () => req.result.createObjectStore('handles');
      req.onsuccess = () => resolve(req.result);
      req.onerror = () => reject(req.error);
    });
  }
  async function saveHandle(key, handle) {
    const db = await idb();
    return new Promise((resolve, reject) => {
      const tx = db.transaction('handles', 'readwrite');
      tx.objectStore('handles').put(handle, key);
      tx.oncomplete = resolve; tx.onerror = () => reject(tx.error);
    });
  }
  async function loadHandle(key) {
    const db = await idb();
    return new Promise((resolve) => {
      const req = db.transaction('handles').objectStore('handles').get(key);
      req.onsuccess = () => resolve(req.result || null);
      req.onerror = () => resolve(null);
    });
  }

  const roots = { claude: null, cursor: null, codex: null }; // DirectoryHandle

  async function ensurePermission(handle) {
    if (!handle) return false;
    const q = await handle.queryPermission({ mode: 'readwrite' });
    if (q === 'granted') return true;
    return (await handle.requestPermission({ mode: 'readwrite' })) === 'granted';
  }

  // ---------- 目录工具 ----------
  async function getDir(root, segments, create) {
    let h = root;
    for (const s of segments) {
      h = await h.getDirectoryHandle(s, { create: !!create });
    }
    return h;
  }
  async function dirExists(root, segments) {
    try { await getDir(root, segments, false); return true; } catch (_) { return false; }
  }
  async function writeFileAt(dir, name, data) {
    const fh = await dir.getFileHandle(name, { create: true });
    const w = await fh.createWritable();
    await w.write(data);
    await w.close();
  }
  async function walkSize(dir) {
    let total = 0, latest = 0;
    for await (const [, entry] of dir.entries()) {
      if (entry.kind === 'file') {
        const f = await entry.getFile();
        total += f.size;
        if (f.lastModified > latest) latest = f.lastModified;
      } else {
        const sub = await walkSize(entry);
        total += sub.total;
        if (sub.latest > latest) latest = sub.latest;
      }
    }
    return { total, latest };
  }
  function fmDescription(text) {
    const m = text.match(/^---\n([\s\S]*?)\n---/);
    if (!m) return '';
    const d = m[1].match(/^description:\s*["']?(.+?)["']?\s*$/m);
    return d ? d[1] : '';
  }

  // ---------- /api/installed ----------
  async function apiInstalled() {
    const resp = {
      claude_root: '~/.claude/skills', cursor_root: '~/.cursor/rules', codex_root: '~/.codex/skills',
      claude_skills: null, cursor_rules: null, codex_skills: null,
    };
    if (roots.claude && await dirExists(roots.claude, ['skills'])) {
      const skills = await getDir(roots.claude, ['skills']);
      const items = [];
      for await (const [name, entry] of skills.entries()) {
        if (entry.kind !== 'directory' || name.startsWith('.')) continue;
        const { total, latest } = await walkSize(entry);
        let summary = '';
        try {
          const f = await (await entry.getFileHandle('SKILL.md')).getFile();
          summary = fmDescription(await f.text());
        } catch (_) {}
        items.push({
          name, path: '~/.claude/skills/' + name, size: total,
          modified: new Date(latest || Date.now()).toISOString(),
          source: name === 'go-team-standards' ? 'team' : 'other', summary,
        });
      }
      items.sort((a, b) => ((a.source === 'team' ? '0' : '2') + a.name).localeCompare((b.source === 'team' ? '0' : '2') + b.name));
      resp.claude_skills = items;
    }
    if (roots.cursor && await dirExists(roots.cursor, ['rules'])) {
      const rules = await getDir(roots.cursor, ['rules']);
      const items = [];
      for await (const [name, entry] of rules.entries()) {
        if (entry.kind !== 'file' || !name.endsWith('.mdc')) continue;
        const f = await entry.getFile();
        const src = name.startsWith('custom-') ? 'custom'
          : (/^\d\d-/.test(name) ? 'team' : 'other');
        items.push({
          name, path: '~/.cursor/rules/' + name, size: f.size,
          modified: new Date(f.lastModified).toISOString(),
          source: src, summary: fmDescription(await f.text()),
        });
      }
      items.sort((a, b) => ((a.source === 'team' ? '0' : a.source === 'custom' ? '1' : '2') + a.name)
        .localeCompare((b.source === 'team' ? '0' : b.source === 'custom' ? '1' : '2') + b.name));
      resp.cursor_rules = items;
    }
    if (roots.codex && await dirExists(roots.codex, ['skills'])) {
      const skills = await getDir(roots.codex, ['skills']);
      const items = [];
      for await (const [name, entry] of skills.entries()) {
        if (entry.kind !== 'directory' || name.startsWith('.')) continue;
        const { total, latest } = await walkSize(entry);
        let summary = '';
        try {
          const f = await (await entry.getFileHandle('SKILL.md')).getFile();
          summary = fmDescription(await f.text());
        } catch (_) {}
        items.push({
          name, path: '~/.codex/skills/' + name, size: total,
          modified: new Date(latest || Date.now()).toISOString(),
          source: name === 'go-team-standards' ? 'team' : 'other', summary,
        });
      }
      items.sort((a, b) => ((a.source === 'team' ? '0' : '2') + a.name).localeCompare((b.source === 'team' ? '0' : '2') + b.name));
      resp.codex_skills = items;
    }
    return resp;
  }

  // ---------- /api/install（global scope）----------
  function decodePayload(pf) {
    if (pf.encoding === 'base64') {
      const bin = atob(pf.content);
      const buf = new Uint8Array(bin.length);
      for (let i = 0; i < bin.length; i++) buf[i] = bin.charCodeAt(i);
      return buf;
    }
    return pf.content;
  }

  async function apiInstall(body) {
    const req = body ? JSON.parse(body) : {};
    if (req.scope === 'project') {
      return { ok: false, error: '网页版暂只支持全局安装（项目级请用本地 App）' };
    }
    const payload = await (await fetch('api/web-install-payload.json')).json();
    const installed = [], warnings = [];
    // app.js 的 target 取值："claude" | "cursor" | "codex" | "all"
    const want = (t) => !req.target || req.target === 'all' || req.target === t;

    // 把负载文件树写进某个 skills 目录（先整树删除再写，与后端 installEmbeddedSkill 同构）
    async function writeSkillTree(skillsDir, files, topDir) {
      try { await skillsDir.removeEntry(topDir, { recursive: true }); } catch (_) {}
      for (const pf of files) {
        const segs = pf.path.split('/');
        const name = segs.pop();
        const dir = await getDir(skillsDir, segs, true);
        await writeFileAt(dir, name, decodePayload(pf));
      }
    }

    if (want('claude')) {
      if (roots.claude && await ensurePermission(roots.claude)) {
        const skills = await getDir(roots.claude, ['skills'], true);
        await writeSkillTree(skills, payload.claude, 'go-team-standards');
        installed.push('Claude: ~/.claude/skills/go-team-standards/（' + payload.claude.length + ' 个文件）');
      } else {
        warnings.push('Claude：未连接 ~/.claude 目录，跳过');
      }
    }
    if (want('cursor')) {
      if (roots.cursor && await ensurePermission(roots.cursor)) {
        const rules = await getDir(roots.cursor, ['rules'], true);
        for await (const [name, entry] of rules.entries()) {
          if (entry.kind === 'file' && /^\d\d-/.test(name) && name.endsWith('.mdc')) {
            try { await rules.removeEntry(name); } catch (_) {}
          }
        }
        for (const pf of payload.cursor) {
          await writeFileAt(rules, pf.path, decodePayload(pf));
        }
        installed.push('Cursor: ~/.cursor/rules/（' + payload.cursor.length + ' 条规则）');
      } else {
        warnings.push('Cursor：未连接 ~/.cursor 目录，跳过');
      }
    }
    if (want('codex')) {
      if (roots.codex && await ensurePermission(roots.codex)) {
        const skills = await getDir(roots.codex, ['skills'], true);
        await writeSkillTree(skills, payload.codex || [], 'go-team-standards');
        installed.push('Codex: ~/.codex/skills/go-team-standards/（' + (payload.codex || []).length + ' 个文件）');
      } else {
        warnings.push('Codex：未连接 ~/.codex 目录，跳过');
      }
    }
    return { ok: installed.length > 0, installed, warnings };
  }

  // ---------- /api/installed-skill-uninstall ----------
  async function apiUninstall(body) {
    const req = body ? JSON.parse(body) : {};
    const p = String(req.path || '');
    let m;
    if ((m = p.match(/\.claude\/skills\/([^/]+)$/)) && roots.claude) {
      if (!await ensurePermission(roots.claude)) return { ok: false, error: '未授权 ~/.claude' };
      const skills = await getDir(roots.claude, ['skills']);
      await skills.removeEntry(m[1], { recursive: true });
      return { ok: true, removed: [p] };
    }
    if ((m = p.match(/\.cursor\/rules\/([^/]+\.mdc)$/)) && roots.cursor) {
      if (!await ensurePermission(roots.cursor)) return { ok: false, error: '未授权 ~/.cursor' };
      const rules = await getDir(roots.cursor, ['rules']);
      await rules.removeEntry(m[1]);
      return { ok: true, removed: [p] };
    }
    if ((m = p.match(/\.codex\/skills\/([^/]+)$/)) && roots.codex) {
      if (!await ensurePermission(roots.codex)) return { ok: false, error: '未授权 ~/.codex' };
      const skills = await getDir(roots.codex, ['skills']);
      await skills.removeEntry(m[1], { recursive: true });
      return { ok: true, removed: [p] };
    }
    return { ok: false, error: '路径不在已授权目录内：' + p };
  }

  // ---------- fetch 路由（app.js shim 调用） ----------
  window.__TS_FS__ = {
    get connected() { return !!(roots.claude || roots.cursor); },
    supported,
    // 调试 / 测试注入口：允许塞入自定义 DirectoryHandle（如 OPFS、内存实现）
    _setRoot(key, handle) { roots[key] = handle; refreshBar(); },
    async route(method, url, body) {
      const path = new URL(url, location.href).pathname.replace(/^.*\/api\//, '/api/');
      if (method === 'GET' && path === '/api/installed' && this.connected) {
        return apiInstalled();
      }
      if (method === 'POST' && path === '/api/install' && this.connected) {
        return apiInstall(body);
      }
      if (method === 'POST' && path === '/api/installed-skill-uninstall' && this.connected) {
        return apiUninstall(body);
      }
      return null; // 未接管 → shim 走快照 / 只读 stub
    },
  };

  // ---------- 连接 UI（侧栏底部） ----------
  function refreshBar() {
    const el = document.getElementById('fsLocalStatus');
    if (!el) return;
    if (!supported) {
      el.textContent = '本浏览器不支持本地目录连接（需 Chrome / Edge）';
      return;
    }
    const parts = [];
    parts.push(roots.claude ? '✅ .claude' : '⚪ .claude');
    parts.push(roots.cursor ? '✅ .cursor' : '⚪ .cursor');
    parts.push(roots.codex ? '✅ .codex' : '⚪ .codex');
    el.textContent = parts.join('  ');
  }

  async function pick(key) {
    try {
      const handle = await window.showDirectoryPicker({ mode: 'readwrite', id: 'ts-' + key });
      if (handle.name !== '.' + key) {
        if (!confirm('选中的目录名是「' + handle.name + '」，期望是「.' + key + '」。确定用这个目录？\n\n（macOS 选择框里按 Cmd+Shift+. 可显示隐藏目录）')) return;
      }
      roots[key] = handle;
      await saveHandle(key, handle);
      refreshBar();
      // 连接后刷新已装清单 tab 的计数
      try { window.dispatchEvent(new Event('ts-fs-connected')); } catch (_) {}
    } catch (e) {
      if (e && e.name !== 'AbortError') alert('连接失败：' + e);
    }
  }

  async function restore() {
    for (const key of ['claude', 'cursor', 'codex']) {
      const h = await loadHandle(key);
      if (h && (await h.queryPermission({ mode: 'readwrite' })) === 'granted') {
        roots[key] = h;
      } else if (h) {
        roots[key] = h; // 句柄还在，权限等首次操作时 requestPermission（需用户手势）
      }
    }
    refreshBar();
  }

  function injectUI() {
    const aside = document.querySelector('aside, .sidebar, [class*=side]');
    const host = document.createElement('div');
    const installURL = new URL('install.sh', location.href).href;
    const cmd = 'curl -fsSL ' + installURL + ' | sh';
    host.style.cssText = 'padding:10px 14px; font-size:11.5px; line-height:1.7; border-top:0.5px solid var(--separator, #ddd); margin-top:8px;';
    host.innerHTML =
      '<div style="font-weight:600; margin-bottom:4px;">⚡ 终端一键安装（推荐）</div>' +
      '<div style="color:var(--label-3,#888); margin-bottom:4px;">三端规范（Claude / Cursor / Codex）一条命令装齐：</div>' +
      '<code id="fsCurlCmd" style="display:block; user-select:all; word-break:break-all; background:var(--bg-mute,rgba(0,0,0,0.05)); border-radius:6px; padding:6px 8px; font-size:10.5px; margin-bottom:6px;">' + cmd + '</code>' +
      '<button id="fsCopyCmd" class="btn btn-primary btn-sm" style="margin-bottom:10px;">📋 复制命令</button>' +
      '<div style="font-weight:600; margin-bottom:4px;">🔌 本地目录（高级）</div>' +
      '<div style="color:var(--label-3,#888); margin-bottom:4px;">⚠️ Chrome 安全策略禁止网页访问 .claude 等隐藏目录，多数情况连不上 —— 请优先用上面的命令安装。</div>' +
      '<div id="fsLocalStatus" style="color:var(--label-3,#888); margin-bottom:6px;"></div>' +
      (supported
        ? '<button id="fsConnectClaude" class="btn btn-ghost btn-sm" style="margin-right:6px;">连 .claude</button>' +
          '<button id="fsConnectCursor" class="btn btn-ghost btn-sm" style="margin-right:6px;">连 .cursor</button>' +
          '<button id="fsConnectCodex" class="btn btn-ghost btn-sm">连 .codex</button>'
        : '');
    (aside || document.body).appendChild(host);
    document.getElementById('fsCopyCmd').addEventListener('click', () => {
      navigator.clipboard.writeText(cmd).then(() => {
        const b = document.getElementById('fsCopyCmd');
        b.textContent = '✓ 已复制，去终端粘贴运行';
        setTimeout(() => { b.textContent = '📋 复制命令'; }, 2500);
      });
    });
    if (supported) {
      document.getElementById('fsConnectClaude').addEventListener('click', () => pick('claude'));
      document.getElementById('fsConnectCursor').addEventListener('click', () => pick('cursor'));
      document.getElementById('fsConnectCodex').addEventListener('click', () => pick('codex'));
    }
    refreshBar();
  }

  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => { injectUI(); restore(); });
  } else {
    injectUI(); restore();
  }
})();
