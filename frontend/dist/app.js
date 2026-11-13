const api = () => window.go.main.App;

let state = null;
let byKey = {};
let themes = null;
let fonts = null;
let tab = 'theme';
let advCat = null;
let advQuery = '';
let themeQuery = '';
let themeMode = 'all';
let keys = null;      // KeybindState from the backend
let keyCat = 'tabs';
let keyQuery = '';
let recording = null; // row id, or 'new' while adding a shortcut
let adding = null;    // { action, param } while the add form is open

let lang = 'en';

const I18N = {
  en: {
    'tab.theme': 'Theme', 'tab.color': 'Colors', 'tab.font': 'Font', 'tab.window': 'Window', 'tab.keys': 'Keys', 'tab.advanced': 'Advanced',
    'kc.tabs': 'Tabs', 'kc.splits': 'Splits', 'kc.window': 'Window', 'kc.clipboard': 'Clipboard', 'kc.font': 'Font size', 'kc.scroll': 'Scrolling', 'kc.search': 'Search', 'kc.selection': 'Selection', 'kc.other': 'Other',
    keysHelp: 'Click a shortcut, then press the new keys. Esc cancels.', pressKeys: 'Press keys… (Esc to cancel)', loadingKeys: 'Loading shortcuts…',
    addShortcut: '+ Add shortcut', resetKeys: 'Reset all shortcuts', searchKeys: 'Search actions or keys', tagChanged: 'Changed', tagAdded: 'Added',
    actionPlaceholder: 'Action, e.g. toggle_tab_overview', paramPlaceholder: 'Parameter (optional)', recordKey: 'Record keys', remove: 'Remove', cancel: 'Cancel', noKey: 'None',
    keyConflict: '{0} was also assigned to “{1}”. It now runs the new action instead.', unknownAction: 'Choose an action from the list first.',
    'g.basicColors': 'Basic colors', 'g.cursorSelection': 'Cursor & selection', 'g.opacitySplits': 'Opacity & splits',
    'g.sizeWeight': 'Size & weight', 'g.spacing': 'Spacing', 'g.padding': 'Padding', 'g.look': 'Appearance',
    'g.cursor': 'Cursor', 'g.startQuit': 'Startup & quit',
    'cat.colors': 'Colors & theme', 'cat.font': 'Font', 'cat.cursor': 'Cursor', 'cat.window': 'Window',
    'cat.mouse': 'Mouse & clipboard', 'cat.keys': 'Keys', 'cat.shell': 'Shell & startup', 'cat.macos': 'macOS',
    'cat.linux': 'Linux (GTK)', 'cat.other': 'Other',
    saved: 'Saved', failed: 'Failed',
    repeatReadonly: 'Multi-line options cannot be edited yet',
    followTheme: 'From theme', unset: '(not set)',
    'reload.newwindow': 'New windows only', 'reload.restart': 'Restart required',
    moreDoc: 'Details', resetTitle: 'Remove this line from the config file and use the default', reset: 'Reset',
    loadingThemes: 'Loading themes…', searchThemes: 'Search themes ({0})', all: 'All', dark: 'Dark', light: 'Light',
    current: 'Current: ', defaultTheme: '(default)',
    themeOverridden: 'The Colors tab overrides {0}, so part of the theme is hidden.',
    fontPreview: 'Preview (shown in this app; Ghostty is updated at the same time)', font: 'Font',
    defaultFont: '(default font)', loadingFonts: 'Loading fonts…',
    fontHint: 'Hover to preview, click to apply', fontPreviewing: 'Previewing {0} — click it to apply',
    nItems: '{0} options', searchAll: 'Search all {0} options', changed: 'Changed',
    'g.palette': 'Palette (16 colors)', paletteHelp: 'Click a color to override the theme. Outlined colors are changed; right-click one to reset it.',
    resetPalette: 'Reset palette',
    'pal.0': 'Black', 'pal.1': 'Red', 'pal.2': 'Green', 'pal.3': 'Yellow', 'pal.4': 'Blue', 'pal.5': 'Magenta', 'pal.6': 'Cyan', 'pal.7': 'White',
    'pal.8': 'Bright black', 'pal.9': 'Bright red', 'pal.10': 'Bright green', 'pal.11': 'Bright yellow', 'pal.12': 'Bright blue', 'pal.13': 'Bright magenta', 'pal.14': 'Bright cyan', 'pal.15': 'Bright white',
    backup: 'Backup: ', language: 'Language', sampleCjk: 'CJK: いろはにほへと 日本語 中文 한국어',
  },
  ja: {
    'tab.theme': 'テーマ', 'tab.color': '色', 'tab.font': 'フォント', 'tab.window': 'ウィンドウ', 'tab.keys': 'キー操作', 'tab.advanced': '詳細設定',
    'kc.tabs': 'タブ', 'kc.splits': '分割', 'kc.window': 'ウィンドウ', 'kc.clipboard': 'クリップボード', 'kc.font': 'フォントサイズ', 'kc.scroll': 'スクロール', 'kc.search': '検索', 'kc.selection': '選択範囲', 'kc.other': 'その他',
    keysHelp: 'ショートカットをクリックして、新しいキーを押すと変更できます。Esc で取り消します。', pressKeys: 'キーを押してください…（Esc で取消）', loadingKeys: 'ショートカットを読み込み中…',
    addShortcut: '＋ ショートカットを追加', resetKeys: 'すべて既定に戻す', searchKeys: '動作やキーで検索', tagChanged: '変更', tagAdded: '追加',
    actionPlaceholder: '動作（例: toggle_tab_overview）', paramPlaceholder: '引数（任意）', recordKey: 'キーを記録', remove: '削除', cancel: 'キャンセル', noKey: 'なし',
    keyConflict: '{0} は「{1}」にも割り当てられていました。新しい動作に置き換えました。', unknownAction: '先に一覧から動作を選んでください。',
    'g.basicColors': '基本の色', 'g.cursorSelection': 'カーソル・選択', 'g.opacitySplits': '透明度・分割',
    'g.sizeWeight': 'サイズ・太さ', 'g.spacing': '文字の詰め', 'g.padding': '余白', 'g.look': '見た目',
    'g.cursor': 'カーソル', 'g.startQuit': '起動・終了',
    'cat.colors': '色・テーマ', 'cat.font': 'フォント', 'cat.cursor': 'カーソル', 'cat.window': 'ウィンドウ',
    'cat.mouse': 'マウス・クリップボード', 'cat.keys': 'キー操作', 'cat.shell': 'シェル・起動', 'cat.macos': 'macOS',
    'cat.linux': 'Linux (GTK)', 'cat.other': 'その他',
    saved: '保存しました', failed: '失敗しました',
    repeatReadonly: '複数行の設定はまだ編集できません',
    followTheme: 'テーマに従う', unset: '(未設定)',
    'reload.newwindow': '新しいウィンドウから', 'reload.restart': '再起動が必要',
    moreDoc: '詳しい説明', resetTitle: 'この行を設定ファイルから消して既定値に戻す', reset: '既定に戻す',
    loadingThemes: 'テーマを読み込み中…', searchThemes: 'テーマを検索（{0} 件）', all: 'すべて', dark: 'ダーク', light: 'ライト',
    current: '現在: ', defaultTheme: '（既定）',
    themeOverridden: '色タブで {0} を上書きしているため、テーマの一部が隠れます。',
    fontPreview: 'プレビュー（このアプリ内での表示です。実際の Ghostty にも同時に反映されます）', font: 'フォント',
    defaultFont: '（既定のフォント）', loadingFonts: 'フォント一覧を読み込み中…',
    fontHint: 'マウスを乗せると見本が切り替わり、クリックで適用します', fontPreviewing: '{0} をプレビュー中（クリックで適用）',
    nItems: '{0} 項目', searchAll: '全 {0} 項目を検索', changed: '変更済み',
    'g.palette': 'パレット（16 色）', paletteHelp: '色をクリックするとテーマの色を上書きできます。枠の付いた色は変更済みです（右クリックで既定に戻す）。',
    resetPalette: 'パレットを既定に戻す',
    'pal.0': '黒', 'pal.1': '赤', 'pal.2': '緑', 'pal.3': '黄', 'pal.4': '青', 'pal.5': 'マゼンタ', 'pal.6': 'シアン', 'pal.7': '白',
    'pal.8': '明るい黒', 'pal.9': '明るい赤', 'pal.10': '明るい緑', 'pal.11': '明るい黄', 'pal.12': '明るい青', 'pal.13': '明るいマゼンタ', 'pal.14': '明るいシアン', 'pal.15': '明るい白',
    backup: 'バックアップ: ', language: '言語', sampleCjk: 'いろはにほへと　日本語の表示もここで確認',
  },
};

const t = (id, ...args) => (I18N[lang][id] ?? I18N.en[id] ?? id).replace(/\{(\d)\}/g, (_, i) => args[i]);
const summaryOf = (o) => (lang === 'ja' && o.summaryJa) || o.summary;
const docOf = (o) => (lang === 'ja' && o.docJa) || o.doc;

const TABS = ['theme', 'color', 'font', 'window', 'keys', 'advanced'];

const CURATED = {
  color: [
    ['g.basicColors', ['background', 'foreground', 'bold-color', 'minimum-contrast']],
    ['g.cursorSelection', ['cursor-color', 'cursor-text', 'selection-background', 'selection-foreground']],
    ['g.opacitySplits', ['background-opacity', 'unfocused-split-opacity', 'unfocused-split-fill', 'split-divider-color']],
  ],
  font: [
    ['g.sizeWeight', ['font-size', 'font-thicken', 'font-thicken-strength', 'font-synthetic-style']],
    ['g.spacing', ['adjust-cell-height', 'adjust-cell-width', 'font-feature']],
  ],
  window: [
    ['g.padding', ['window-padding-x', 'window-padding-y', 'window-padding-balance', 'window-padding-color']],
    ['g.look', ['background-opacity', 'background-blur', 'window-decoration', 'window-theme', 'macos-titlebar-style']],
    ['g.cursor', ['cursor-style', 'cursor-style-blink']],
    ['g.startQuit', ['window-save-state', 'confirm-close-surface', 'window-inherit-working-directory']],
  ],
};

const SLIDERS = {
  'font-size': [6, 36, 0.5],
  'background-opacity': [0, 1, 0.05],
  'unfocused-split-opacity': [0.15, 1, 0.05],
  'minimum-contrast': [1, 21, 0.1],
  'window-padding-x': [0, 64, 1],
  'window-padding-y': [0, 64, 1],
  'font-thicken-strength': [0, 255, 1],
};


function el(tag, attrs = {}, ...kids) {
  const e = document.createElement(tag);
  for (const [k, v] of Object.entries(attrs)) {
    if (k === 'class') e.className = v;
    else if (k.startsWith('on')) e.addEventListener(k.slice(2), v);
    else if (v !== undefined && v !== null && v !== false) e.setAttribute(k, v === true ? '' : v);
  }
  for (const kid of kids.flat()) if (kid !== null && kid !== undefined) e.append(kid);
  return e;
}

let toastTimer;
function toast(msg, err = false) {
  const box = document.getElementById('toast');
  box.textContent = msg;
  box.className = 'show' + (err ? ' err' : '');
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => (box.className = ''), err ? 6000 : 2200);
}

function debounce(fn, ms) {
  let timer;
  return (...a) => { clearTimeout(timer); timer = setTimeout(() => fn(...a), ms); };
}

async function apply(promise, rerender = false) {
  const res = await promise;
  if (res.set) state.set = res.set;
  toast(res.message || t(res.ok ? 'saved' : 'failed'), !res.ok);
  if (rerender || !res.ok) render(true);
  else refreshRows();
}
const setOpt = (key, value, rerender) => apply(api().SetOption(key, String(value)), rerender);
const resetOpt = (key) => apply(api().ResetOption(key), true);

// ---------- color picking ----------

// WebKit fires "change" on every pick in the macOS color panel, and re-rendering
// removes the <input> and closes the panel. So a picked color only updates the
// preview; it is saved once the panel is closed.
let pendingColor = null;
let colorPoll = null;

function holdColor(save) {
  pendingColor = save;
  if (colorPoll) return;
  colorPoll = setInterval(async () => {
    let open = false;
    try { open = await api().ColorPanelOpen(); } catch (_) { /* treat as closed */ }
    if (!open) flushColor();
  }, 300);
}

function flushColor() {
  clearInterval(colorPoll);
  colorPoll = null;
  const save = pendingColor;
  pendingColor = null;
  if (save) save();
}

// ---------- option rows ----------

function current(o) {
  return state.set[o.key] !== undefined ? state.set[o.key] : o.default;
}

function toHex(v) {
  if (/^#?[0-9a-f]{6}$/i.test(v)) return v.startsWith('#') ? v : '#' + v;
  return '#000000';
}

function control(o) {
  const val = current(o);
  const range = SLIDERS[o.key];
  if (o.kind === 'repeat') {
    return el('textarea', { readonly: true, title: t('repeatReadonly') }, val);
  }
  if (o.kind === 'bool') {
    return el('input', { type: 'checkbox', checked: val === 'true', onchange: (e) => setOpt(o.key, e.target.checked) });
  }
  if (o.kind === 'enum') {
    return el('select', { onchange: (e) => setOpt(o.key, e.target.value) },
      o.choices.map((c) => el('option', { value: c, selected: c === val }, c)));
  }
  if (o.kind === 'color') {
    const text = el('input', { type: 'text', value: val, placeholder: t('followTheme'), style: 'width:110px',
      onchange: (e) => setOpt(o.key, e.target.value) });
    const follow = (e) => { text.value = e.target.value; holdColor(() => setOpt(o.key, e.target.value)); };
    const pick = el('input', { type: 'color', value: toHex(val), oninput: follow, onchange: follow });
    return el('div', { class: 'ctl' }, pick, text);
  }
  if (range || o.kind === 'number') {
    const [min, max, step] = range || [undefined, undefined, 'any'];
    const num = el('input', { type: 'number', value: val, min, max, step, onchange: (e) => setOpt(o.key, e.target.value) });
    if (!range) return num;
    const slider = el('input', { type: 'range', min, max, step, value: parseFloat(val) || 0,
      oninput: (e) => { num.value = e.target.value; send(e.target.value); } });
    const send = debounce((v) => setOpt(o.key, v), 250);
    return el('div', { class: 'ctl' }, slider, num);
  }
  return el('input', { type: 'text', value: val, placeholder: o.default || t('unset'), onchange: (e) => setOpt(o.key, e.target.value) });
}

const HAS_RELOAD_LABEL = { newwindow: true, restart: true };

function optRow(o) {
  const changed = state.set[o.key] !== undefined;
  return el('div', { class: 'opt' + (changed ? ' changed' : ''), 'data-key': o.key },
    el('div', { class: 'row' },
      el('div', { class: 'info' },
        el('code', { class: 'key' }, o.key),
        HAS_RELOAD_LABEL[o.reload] ? el('span', { class: 'badge ' + o.reload }, t('reload.' + o.reload)) : null,
        el('div', { class: 'sum' }, summaryOf(o)),
        docOf(o).length > summaryOf(o).length + 20 ? el('details', {}, el('summary', {}, t('moreDoc')), el('pre', {}, docOf(o))) : null),
      el('div', { class: 'ctl' },
        control(o),
        changed ? el('button', { class: 'reset', title: t('resetTitle'), onclick: () => resetOpt(o.key) }, t('reset')) : null)));
}

function refreshRows() {
  document.querySelectorAll('.opt[data-key]').forEach((row) => {
    const o = byKey[row.dataset.key];
    const changed = state.set[o.key] !== undefined;
    if (changed !== row.classList.contains('changed')) row.replaceWith(optRow(o));
  });
}

function curatedPane(groups, head) {
  const wrap = el('div', { class: 'scroll' });
  if (head) wrap.append(head);
  for (const [title, keys] of groups) {
    const opts = keys.map((k) => byKey[k]).filter(Boolean);
    if (!opts.length) continue;
    wrap.append(el('h3', {}, t(title)), ...opts.map(optRow));
  }
  return wrap;
}

// ---------- theme tab ----------

function themeCard(th) {
  const on = state.set.theme === th.name;
  const p = th.palette;
  const line = (...parts) => el('div', {}, ...parts.map(([c, s]) => el('span', { style: `color:${c || th.foreground}` }, s)));
  return el('div', { class: 'th' + (on ? ' on' : ''), title: th.name, onclick: () => setOpt('theme', th.name, true) },
    el('div', { class: 'pv', style: `background:${th.background};color:${th.foreground}` },
      line([p[2], '➜ '], [p[6], '~/src '], [p[4], 'git:('], [p[1], 'main'], [p[4], ')']),
      line([null, 'ls '], [p[3], '-la']),
      line([p[4], 'docs/ '], [p[2], 'run.sh '], [p[5], 'img.png']),
      el('div', { class: 'pal' }, p.slice(0, 8).map((c) => el('i', { style: `background:${c}` })))),
    el('div', { class: 'nm' }, th.name));
}

function themePane() {
  const wrap = el('div', { class: 'scroll' });
  const overridden = ['background', 'foreground', 'palette', 'cursor-color', 'selection-background'].filter((k) => state.set[k] !== undefined);
  if (!themes) {
    wrap.append(el('p', { class: 'muted' }, t('loadingThemes')));
    api().Themes().then((list) => { themes = list || []; render(); });
    return wrap;
  }
  const grid = el('div', { class: 'themes' });
  const fill = () => {
    grid.replaceChildren();
    const q = themeQuery.toLowerCase();
    themes
      .filter((th) => (!q || th.name.toLowerCase().includes(q)) && (themeMode === 'all' || (themeMode === 'dark') === th.dark))
      .forEach((th) => grid.append(themeCard(th)));
  };
  wrap.append(...[
    el('div', { class: 'toolbar' },
      el('input', { type: 'search', placeholder: t('searchThemes', themes.length), value: themeQuery,
        oninput: (e) => { themeQuery = e.target.value; fill(); } }),
      el('select', { onchange: (e) => { themeMode = e.target.value; fill(); } },
        ['all', 'dark', 'light'].map((v) => el('option', { value: v, selected: v === themeMode }, t(v)))),
      el('span', { class: 'muted' }, t('current'), el('b', {}, state.set.theme || t('defaultTheme'))),
      state.set.theme ? el('button', { class: 'reset', onclick: () => resetOpt('theme') }, t('reset')) : null),
    overridden.length ? el('div', { class: 'note' }, t('themeOverridden', overridden.join(', '))) : null,
    grid,
  ].filter(Boolean));
  fill();
  return wrap;
}

// ---------- palette (colors tab) ----------

// "0=#1d1f21\n1=#cc6666" -> {0: '#1d1f21', 1: '#cc6666'}
function paletteMap(text) {
  const m = {};
  for (const line of (text || '').split('\n')) {
    const i = line.indexOf('=');
    const n = parseInt(line.slice(0, i), 10);
    if (i > 0 && !Number.isNaN(n)) m[n] = line.slice(i + 1).trim();
  }
  return m;
}

function pickPalette(e, i, user) {
  const v = e.target.value;
  e.target.parentElement.querySelector('i').style.background = v;
  document.querySelectorAll(`.termpv [data-pal="${i}"]`).forEach((n) => { n.style.color = v; });
  holdColor(() => savePalette({ ...user, [i]: v }));
}

function savePalette(map) {
  const values = Object.keys(map).map(Number).sort((a, b) => a - b).map((i) => `${i}=${map[i]}`);
  apply(api().SetRepeated('palette', values), true);
}

function palettePane() {
  if (!themes) api().Themes().then((list) => { themes = list || []; render(true); });
  const user = paletteMap(state.set.palette);
  const theme = (themes || []).find((th) => th.name === state.set.theme);
  const defaults = paletteMap(byKey.palette && byKey.palette.default);
  const color = (i) => user[i] || (theme && theme.palette[i]) || defaults[i] || '#000000';
  const swatches = [];
  for (let i = 0; i < 16; i++) {
    swatches.push(el('label', { class: 'sw' + (user[i] ? ' changed' : ''), title: `palette = ${i}=${color(i)}`,
      oncontextmenu: (e) => { e.preventDefault(); if (user[i]) { const m = { ...user }; delete m[i]; savePalette(m); } } },
      el('i', { style: `background:${color(i)}` }),
      el('div', {}, el('span', {}, String(i)), el('span', {}, t('pal.' + i))),
      el('input', { type: 'color', value: toHex(color(i)),
        oninput: (e) => pickPalette(e, i, user), onchange: (e) => pickPalette(e, i, user) })));
  }
  const bg = state.set.background || (theme && theme.background) || (byKey.background && byKey.background.default);
  const fg = state.set.foreground || (theme && theme.foreground) || (byKey.foreground && byKey.foreground.default);
  const span = (i, text) => el('span', { style: `color:${color(i)}`, 'data-pal': i }, text);
  const preview = el('div', { class: 'termpv', style: `background:${toHex(bg)};color:${toHex(fg)}` },
    el('div', {}, span(2, '➜ '), span(6, '~/src '), span(4, 'git:('), span(1, 'main'), span(4, ') '), span(7, 'ls -la')),
    el('div', {}, span(4, 'docs/ '), span(2, 'run.sh '), span(5, 'img.png '), span(9, 'error.log '), span(3, 'notes.txt')));
  return el('div', {},
    el('div', { class: 'row-head' }, el('h3', {}, t('g.palette')),
      Object.keys(user).length ? el('button', { class: 'reset', onclick: () => apply(api().ResetOption('palette'), true) }, t('resetPalette')) : null),
    el('div', { class: 'muted small' }, t('paletteHelp')),
    el('div', { class: 'palgrid' }, swatches),
    preview);
}

// ---------- font tab ----------

function fontPane() {
  const fam = state.set['font-family'] || '';
  const size = parseFloat(current(byKey['font-size'])) || 13;
  const sample = 'The quick brown fox jumps over the lazy dog\n0O 1lI |  => != === -> <= >= ~~>\n' + t('sampleCjk');
  const cur = fam.split('\n')[0];
  const stack = (f) => `${f ? `"${f}", ` : ''}"SF Mono", Menlo, monospace`;
  const sampleBox = el('div', { class: 'sample', style: `font-family:${stack(cur)};font-size:${size}px` }, sample);
  const previewLabel = el('div', { class: 'muted small' }, t('fontPreview'));
  const preview = el('div', { class: 'preview' }, previewLabel, sampleBox);
  // Hovering a font only changes the preview; clicking applies it.
  const show = (f) => {
    sampleBox.style.fontFamily = stack(f);
    previewLabel.textContent = f === cur ? t('fontPreview') : t('fontPreviewing', f || t('defaultFont'));
  };
  const fontList = fonts && el('div', { class: 'fontlist', onmouseleave: () => show(cur) },
    ['', ...fonts].map((f) => el('button', { class: 'fontitem' + (f === cur ? ' on' : ''), style: `font-family:${stack(f)}`,
      onmouseenter: () => show(f), onfocus: () => show(f),
      onclick: () => (f === cur ? null : f ? setOpt('font-family', f, true) : resetOpt('font-family')) },
    el('span', { class: 'fname' }, f || t('defaultFont')), el('span', { class: 'fsample' }, 'AaBb 0O 1lI => 日本語'))));
  const head = el('div', {},
    el('h3', {}, t('font')),
    el('div', { class: 'opt' + (fam ? ' changed' : '') },
      el('div', { class: 'row' },
        el('div', { class: 'info' }, el('code', { class: 'key' }, 'font-family'), el('div', { class: 'sum' }, byKey['font-family'] && summaryOf(byKey['font-family']))),
        el('div', { class: 'ctl' },
          fonts ? el('span', { class: 'muted small' }, t('fontHint')) : el('span', { class: 'muted' }, t('loadingFonts')),
          fam ? el('button', { class: 'reset', onclick: () => resetOpt('font-family') }, t('reset')) : null)),
      fontList),
    preview);
  if (!fonts) api().Fonts().then((f) => { fonts = f || []; render(); });
  return curatedPane(CURATED.font, head);
}

// ---------- keys tab ----------

// action -> [category, English, Japanese]
const ACTIONS = {
  new_tab: ['tabs', 'New tab', '新しいタブ'], close_tab: ['tabs', 'Close tab', 'タブを閉じる'],
  previous_tab: ['tabs', 'Previous tab', '前のタブ'], next_tab: ['tabs', 'Next tab', '次のタブ'], last_tab: ['tabs', 'Last tab', '最後のタブへ'],
  goto_tab: ['tabs', 'Go to tab', 'タブへ移動'], move_tab: ['tabs', 'Move tab', 'タブを移動'], toggle_tab_overview: ['tabs', 'Tab overview', 'タブ一覧を開く'],
  move_tab_to_new_window: ['tabs', 'Move tab to new window', 'タブを新しいウィンドウへ'],
  new_split: ['splits', 'New split', '分割'], goto_split: ['splits', 'Go to split', '分割へ移動'], resize_split: ['splits', 'Resize split', '分割の大きさを変える'],
  equalize_splits: ['splits', 'Equalize splits', '分割の大きさをそろえる'], toggle_split_zoom: ['splits', 'Zoom split', '分割を拡大／元に戻す'],
  new_window: ['window', 'New window', '新しいウィンドウ'], close_window: ['window', 'Close window', 'ウィンドウを閉じる'],
  close_all_windows: ['window', 'Close all windows', 'すべてのウィンドウを閉じる'], close_surface: ['window', 'Close terminal', 'ターミナルを閉じる'],
  toggle_fullscreen: ['window', 'Toggle full screen', 'フルスクリーンの切り替え'], toggle_maximize: ['window', 'Toggle maximize', '最大化の切り替え'],
  toggle_quick_terminal: ['window', 'Quick terminal', 'クイックターミナル'], toggle_visibility: ['window', 'Show / hide Ghostty', 'Ghostty の表示切り替え'],
  toggle_window_float_on_top: ['window', 'Float on top', '最前面に固定'], reset_window_size: ['window', 'Reset window size', 'ウィンドウの大きさを戻す'],
  quit: ['window', 'Quit Ghostty', 'Ghostty を終了'],
  copy_to_clipboard: ['clipboard', 'Copy', 'コピー'], paste_from_clipboard: ['clipboard', 'Paste', 'ペースト'],
  paste_from_selection: ['clipboard', 'Paste selection', '選択範囲をペースト'], copy_url_to_clipboard: ['clipboard', 'Copy URL', 'URL をコピー'],
  copy_title_to_clipboard: ['clipboard', 'Copy title', 'タイトルをコピー'],
  increase_font_size: ['font', 'Increase font size', '文字を大きく'], decrease_font_size: ['font', 'Decrease font size', '文字を小さく'],
  reset_font_size: ['font', 'Reset font size', '文字の大きさを元に戻す'], set_font_size: ['font', 'Set font size', '文字の大きさを指定'],
  scroll_to_top: ['scroll', 'Scroll to top', '一番上へスクロール'], scroll_to_bottom: ['scroll', 'Scroll to bottom', '一番下へスクロール'],
  scroll_page_up: ['scroll', 'Page up', '1 ページ上へ'], scroll_page_down: ['scroll', 'Page down', '1 ページ下へ'],
  scroll_to_selection: ['scroll', 'Scroll to selection', '選択範囲までスクロール'], jump_to_prompt: ['scroll', 'Jump to prompt', 'プロンプトへ移動'],
  clear_screen: ['scroll', 'Clear screen', '画面を消去'],
  start_search: ['search', 'Find', '検索'], search_selection: ['search', 'Find selection', '選択範囲で検索'], end_search: ['search', 'End search', '検索を終了'],
  navigate_search: ['search', 'Find next / previous', '検索結果を移動'],
  adjust_selection: ['selection', 'Extend selection', '選択範囲を広げる'], select_all: ['selection', 'Select all', 'すべて選択'],
  reload_config: ['other', 'Reload config', '設定を再読み込み'], open_config: ['other', 'Open config file', '設定ファイルを開く'],
  toggle_command_palette: ['other', 'Command palette', 'コマンドパレット'], undo: ['other', 'Undo', '元に戻す'], redo: ['other', 'Redo', 'やり直す'],
  inspector: ['other', 'Terminal inspector', 'インスペクタ'], write_screen_file: ['other', 'Write screen to file', '画面の内容をファイルへ'],
  text: ['other', 'Send text', 'テキストを送信'], esc: ['other', 'Send Esc sequence', 'Esc シーケンスを送信'], csi: ['other', 'Send CSI sequence', 'CSI シーケンスを送信'],
};
const PARAMS = {
  left: ['left', '左'], right: ['right', '右'], up: ['up', '上'], down: ['down', '下'], previous: ['previous', '前'], next: ['next', '次'],
  page_up: ['page up', '1 ページ上'], page_down: ['page down', '1 ページ下'], home: ['to start', '先頭まで'], end: ['to end', '末尾まで'],
  this: ['this tab', 'このタブ'], toggle: ['toggle', '切り替え'],
};
const KEY_CATS = ['tabs', 'splits', 'window', 'clipboard', 'font', 'scroll', 'search', 'selection', 'other'];

const actionBase = (a) => a.split(':')[0];
const actionCat = (a) => (ACTIONS[actionBase(a)] || ['other'])[0];
function actionLabel(a) {
  const [base, param] = [actionBase(a), a.slice(actionBase(a).length + 1)];
  const def = ACTIONS[base];
  if (!def) return a;
  const name = lang === 'ja' ? def[2] : def[1];
  if (base === 'goto_tab' && /^\d+$/.test(param)) return lang === 'ja' ? `${param} 番目のタブへ` : `Go to tab ${param}`;
  if (!param) return name;
  const first = param.split(',')[0];
  const p = PARAMS[first] ? (lang === 'ja' ? PARAMS[first][1] : PARAMS[first][0]) : param;
  return lang === 'ja' ? `${name}（${p}）` : `${name} (${p})`;
}

const isMac = /Mac/.test(navigator.platform || navigator.userAgent);
const MOD_SYM = { ctrl: '⌃', alt: '⌥', shift: '⇧', super: '⌘' };
const MOD_TXT = { ctrl: 'Ctrl', alt: 'Alt', shift: 'Shift', super: 'Super' };
const MOD_ALIAS = { cmd: 'super', command: 'super', super: 'super', opt: 'alt', option: 'alt', alt: 'alt', control: 'ctrl', ctrl: 'ctrl', shift: 'shift' };
const MOD_ORDER = ['ctrl', 'alt', 'shift', 'super'];
const KEY_SYM = { arrow_up: '↑', arrow_down: '↓', arrow_left: '←', arrow_right: '→', enter: '↩', tab: '⇥', escape: '⎋', backspace: '⌫',
  delete: '⌦', space: 'Space', page_up: '⇞', page_down: '⇟', home: '↖', end: '↘', backslash: '\\', grave_accent: '`' };

// "shift+super+t" -> {prefix: '', chords: [{mods: ['shift', 'super'], key: 't'}]}
function parseTrigger(trig) {
  let prefix = '';
  let rest = trig;
  for (let m; (m = /^([a-z_]+):/i.exec(rest)); rest = rest.slice(m[0].length)) prefix += m[0].toLowerCase();
  const chords = rest.split('>').map((chord) => {
    const parts = chord.endsWith('++') ? [...chord.slice(0, -2).split('+'), '+'] : chord.split('+');
    const mods = [];
    let key = '';
    parts.forEach((p, i) => {
      const m = MOD_ALIAS[p.toLowerCase()];
      if (m && i < parts.length - 1) mods.push(m); else key = p === '+' ? '+' : p.toLowerCase();
    });
    mods.sort((a, b) => MOD_ORDER.indexOf(a) - MOD_ORDER.indexOf(b));
    return { mods, key };
  });
  return { prefix, chords };
}
const normTrig = (trig) => { const p = parseTrigger(trig); return p.prefix + p.chords.map((c) => [...c.mods, c.key].join('+')).join('>'); };
function formatTrigger(trig) {
  if (!trig) return t('noKey');
  const { prefix, chords } = parseTrigger(trig);
  const keyName = (k) => KEY_SYM[k] || (/^digit_\d$/.test(k) ? k.slice(6) : /^f\d+$/.test(k) ? k.toUpperCase() : k.length === 1 ? k.toUpperCase() : k);
  const text = chords.map((c) => isMac
    ? c.mods.map((m) => MOD_SYM[m]).join('') + keyName(c.key)
    : [...c.mods.map((m) => MOD_TXT[m]), keyName(c.key)].join('+')).join(' → ');
  return (prefix ? prefix.replace(/:/g, ' ').trim() + ' ' : '') + text;
}

// KeyboardEvent.code -> Ghostty key name
function keyFromEvent(e) {
  const c = e.code;
  if (/^Key[A-Z]$/.test(c)) return c.slice(3).toLowerCase();
  if (/^Digit\d$/.test(c)) return c.slice(5);
  if (/^F\d+$/.test(c)) return c.toLowerCase();
  const map = { Minus: '-', Equal: '=', BracketLeft: '[', BracketRight: ']', Backslash: 'backslash', Semicolon: ';', Quote: "'",
    Comma: ',', Period: '.', Slash: '/', Backquote: 'grave_accent', ArrowUp: 'arrow_up', ArrowDown: 'arrow_down', ArrowLeft: 'arrow_left',
    ArrowRight: 'arrow_right', Enter: 'enter', Tab: 'tab', Space: 'space', Backspace: 'backspace', Delete: 'delete', Home: 'home',
    End: 'end', PageUp: 'page_up', PageDown: 'page_down' };
  return map[c] || null;
}

// Ghostty binds some defaults twice (e.g. super+1 and super+digit_1) and they
// look identical, so rows with the same action and the same default key are
// shown and edited as one.
function keyGroups(rows) {
  const groups = new Map();
  for (const r of rows) {
    const k = r.origin === 'added' ? r.id : `${r.action}|${formatTrigger(r.defaultTrigger)}`;
    if (!groups.has(k)) groups.set(k, []);
    groups.get(k).push(r);
  }
  return [...groups.values()].map((list) => {
    const main = list.find((r) => r.enabled && r.origin === 'changed') || list.find((r) => r.enabled) || list[0];
    return { id: main.id, ids: list.map((r) => r.id), action: main.action, trigger: main.trigger, enabled: list.some((r) => r.enabled),
      origin: list.some((r) => r.origin === 'added') ? 'added' : list.some((r) => r.origin === 'changed') ? 'changed' : 'default',
      triggers: list.map((r) => r.trigger) };
  });
}

async function saveKeys(rows) {
  await apply(api().SetKeybinds(rows, keys.extra || []), true);
  keys = null;
  render(true);
}

function finishRecording(trig) {
  const id = recording;
  recording = null;
  if (!trig) { render(true); return; }
  const rows = keys.rows.map((r) => ({ ...r }));
  let target;
  if (id === 'new') {
    const action = adding.param ? `${adding.action}:${adding.param}` : adding.action;
    target = { id: 'new', action, trigger: trig, defaultTrigger: '', enabled: true, origin: 'added' };
    rows.push(target);
    adding = null;
  } else {
    const group = keyGroups(keys.rows).find((g) => g.id === id);
    for (const r of rows) {
      if (!group.ids.includes(r.id)) continue;
      if (r.id === id) { r.trigger = trig; r.enabled = true; } else r.enabled = false;
    }
    target = rows.find((r) => r.id === id);
  }
  const others = rows.filter((r) => r !== target && r.enabled && normTrig(r.trigger) === normTrig(trig));
  others.forEach((r) => { r.enabled = false; });
  if (others.length) setTimeout(() => toast(t('keyConflict', formatTrigger(trig), actionLabel(others[0].action))), 2300);
  saveKeys(rows);
}

window.addEventListener('keydown', (e) => {
  if (!recording) return;
  e.preventDefault();
  e.stopPropagation();
  if (e.key === 'Escape') { finishRecording(null); return; }
  const key = keyFromEvent(e);
  if (!key) return; // a modifier on its own, or a key we cannot name
  const mods = [e.ctrlKey && 'ctrl', e.altKey && 'alt', e.shiftKey && 'shift', e.metaKey && 'super'].filter(Boolean);
  finishRecording([...mods, key].join('+'));
}, true);

function keyRow(r) {
  const rec = recording === r.id;
  return el('div', { class: 'kb-row' + (r.origin !== 'default' ? ' changed' : '') + (r.enabled ? '' : ' off') },
    r.origin === 'added'
      ? el('button', { class: 'kb-del', title: t('remove'), onclick: () => saveKeys(keys.rows.filter((x) => x.id !== r.id)) }, '×')
      : el('input', { type: 'checkbox', checked: r.enabled,
          onchange: (e) => saveKeys(keys.rows.map((x) => (r.ids.includes(x.id) ? { ...x, enabled: e.target.checked } : x))) }),
    el('div', { class: 'kb-name' }, actionLabel(r.action),
      r.origin !== 'default' ? el('span', { class: 'tag' }, t(r.origin === 'added' ? 'tagAdded' : 'tagChanged')) : null,
      el('small', {}, r.action)),
    el('button', { class: 'kb-key' + (rec ? ' rec' : ''), title: t('keysHelp'),
      onclick: () => { recording = r.id; render(true); } }, rec ? t('pressKeys') : formatTrigger(r.trigger)));
}

function addForm() {
  const list = el('datalist', { id: 'kb-actions' }, (keys.actions || []).map((a) => el('option', { value: a }, actionLabel(a))));
  const rec = recording === 'new';
  return el('div', { class: 'kb-add' }, list,
    el('input', { type: 'text', list: 'kb-actions', placeholder: t('actionPlaceholder'), value: adding.action,
      oninput: (e) => { adding.action = e.target.value.trim(); } }),
    el('input', { type: 'text', placeholder: t('paramPlaceholder'), value: adding.param, style: 'width:150px',
      oninput: (e) => { adding.param = e.target.value.trim(); } }),
    el('button', { class: 'kb-key' + (rec ? ' rec' : ''), onclick: () => {
      if (!(keys.actions || []).includes(adding.action)) { toast(t('unknownAction'), true); return; }
      recording = 'new'; render(true);
    } }, rec ? t('pressKeys') : t('recordKey')),
    el('button', { class: 'reset', onclick: () => { adding = null; recording = null; render(true); } }, t('cancel')));
}

function keysPane() {
  if (!keys) {
    api().Keybinds().then((k) => { keys = k || { rows: [], actions: [], extra: [] }; keys.rows ||= []; render(true); });
    return el('p', { class: 'muted pad' }, t('loadingKeys'));
  }
  const groups = keyGroups(keys.rows);
  const changed = (g) => g.origin !== 'default' || !g.enabled;
  const counts = Object.fromEntries(KEY_CATS.map((c) => [c, groups.filter((g) => actionCat(g.action) === c).length]));
  const q = keyQuery.toLowerCase();
  const rows = q
    ? groups.filter((g) => [actionLabel(g.action), g.action, formatTrigger(g.trigger), ...g.triggers].some((s) => s.toLowerCase().includes(q)))
    : groups.filter((g) => (keyCat === '__changed' ? changed(g) : actionCat(g.action) === keyCat));
  const side = el('aside', {},
    el('input', { type: 'search', placeholder: t('searchKeys'), value: keyQuery,
      oninput: debounce((e) => {
        keyQuery = e.target.value;
        render(true);
        const box = document.querySelector('.adv aside input[type=search]');
        box.focus();
        box.setSelectionRange(keyQuery.length, keyQuery.length);
      }, 200) }),
    el('div', { class: 'cat' + (keyCat === '__changed' && !q ? ' on' : ''), onclick: () => { keyCat = '__changed'; keyQuery = ''; render(); } },
      t('changed'), el('span', {}, groups.filter(changed).length)),
    KEY_CATS.filter((c) => counts[c]).map((c) => el('div', { class: 'cat' + (c === keyCat && !q ? ' on' : ''),
      onclick: () => { keyCat = c; keyQuery = ''; render(); } }, t('kc.' + c), el('span', {}, counts[c]))));
  const title = q ? t('searchKeys') : keyCat === '__changed' ? t('changed') : t('kc.' + keyCat);
  const list = el('section', {}, el('div', { class: 'kb-list' },
    el('div', { class: 'kb-head' }, el('h3', {}, title), el('span', { class: 'muted small' }, t('keysHelp'))),
    adding ? addForm() : null,
    el('div', { class: 'kb-table' }, rows.map(keyRow)),
    el('div', { class: 'kb-foot' },
      adding ? el('span') : el('button', { class: 'btn', onclick: () => { adding = { action: '', param: '' }; render(true); } }, t('addShortcut')),
      groups.some(changed) || (keys.extra || []).length
        ? el('button', { class: 'btn', onclick: async () => { await apply(api().ResetOption('keybind'), true); keys = null; render(true); } }, t('resetKeys'))
        : null)));
  return el('div', { class: 'adv' }, side, list);
}

// ---------- advanced tab ----------

function advancedPane() {
  const cats = {};
  for (const o of state.options) (cats[o.category] ||= []).push(o);
  const names = Object.keys(cats);
  if (!advCat) advCat = names[0];
  const list = el('section', {});
  const fillList = () => {
    const q = advQuery.toLowerCase();
    const opts = q
      ? state.options.filter((o) => o.key.includes(q) || o.summary.toLowerCase().includes(q) || (o.summaryJa || '').toLowerCase().includes(q))
      : advCat === '__changed' ? state.options.filter((o) => state.set[o.key] !== undefined) : cats[advCat];
    list.replaceChildren(el('div', { class: 'scroll' },
      el('div', { class: 'muted small', style: 'margin-bottom:8px' }, t('nItems', opts.length)),
      opts.map(optRow)));
    list.scrollTop = 0;
  };
  const side = el('aside', {},
    el('input', { type: 'search', placeholder: t('searchAll', state.options.length), value: advQuery,
      oninput: (e) => { advQuery = e.target.value; fillList(); } }),
    el('div', { class: 'cat' + (advCat === '__changed' ? ' on' : ''), onclick: () => pick('__changed') },
      t('changed'), el('span', {}, Object.keys(state.set).length)),
    names.map((n) => el('div', { class: 'cat' + (n === advCat ? ' on' : ''), onclick: () => pick(n) }, t('cat.' + n), el('span', {}, cats[n].length))));
  const pick = (n) => { advCat = n; advQuery = ''; render(); };
  fillList();
  return el('div', { class: 'adv' }, side, list);
}

// ---------- shell ----------

function render(keepScroll = false) {
  if (pendingColor) flushColor(); // e.g. switching tabs with the panel still open
  const pane = document.getElementById('pane');
  const scroll = keepScroll ? (pane.querySelector('.adv section') || pane).scrollTop : 0;
  document.documentElement.lang = lang;
  document.getElementById('tabs').replaceChildren(...TABS.map((id) =>
    el('button', { class: id === tab ? 'on' : '', onclick: () => { tab = id; render(); } }, t('tab.' + id))));
  document.getElementById('path').textContent = state.configPath + (state.backupPath ? `　${t('backup')}${state.backupPath}` : '');
  langPicker();
  let view;
  if (state.error) view = el('p', { class: 'pad' }, state.error);
  else if (tab === 'theme') view = themePane();
  else if (tab === 'font') view = fontPane();
  else if (tab === 'keys') view = keysPane();
  else if (tab === 'advanced') view = advancedPane();
  else if (tab === 'color') view = curatedPane(CURATED.color, palettePane());
  else view = curatedPane(CURATED[tab]);
  pane.replaceChildren(view);
  (pane.querySelector('.adv section') || pane).scrollTop = scroll;
}

function langPicker() {
  const box = document.getElementById('lang');
  box.replaceChildren(el('select', { title: t('language'), 'aria-label': t('language'),
    onchange: async (e) => {
      lang = e.target.value;
      render(true);
      const res = await api().SetLang(lang);
      if (!res.ok) toast(res.message, true);
    } },
    [['en', 'English'], ['ja', '日本語']].map(([v, l]) => el('option', { value: v, selected: v === lang }, l))));
}

async function boot() {
  state = await api().Load();
  state.options ||= [];
  state.set ||= {};
  byKey = Object.fromEntries(state.options.map((o) => [o.key, o]));
  document.getElementById('version').textContent = state.version || '';
  lang = state.lang === 'ja' ? 'ja' : 'en';
  render();
}

window.addEventListener('DOMContentLoaded', () => {
  const wait = () => (window.go ? boot() : setTimeout(wait, 30));
  wait();
});
