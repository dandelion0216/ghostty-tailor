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

let lang = 'en';

const I18N = {
  en: {
    'tab.theme': 'Theme', 'tab.color': 'Colors', 'tab.font': 'Font', 'tab.window': 'Window', 'tab.advanced': 'Advanced',
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
    nItems: '{0} options', searchAll: 'Search all {0} options', changed: 'Changed',
    backup: 'Backup: ', language: 'Language', sampleCjk: 'CJK: いろはにほへと 日本語 中文 한국어',
  },
  ja: {
    'tab.theme': 'テーマ', 'tab.color': '色', 'tab.font': 'フォント', 'tab.window': 'ウィンドウ', 'tab.advanced': '詳細設定',
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
    nItems: '{0} 項目', searchAll: '全 {0} 項目を検索', changed: '変更済み',
    backup: 'バックアップ: ', language: '言語', sampleCjk: 'いろはにほへと　日本語の表示もここで確認',
  },
};

const t = (id, ...args) => (I18N[lang][id] ?? I18N.en[id] ?? id).replace(/\{(\d)\}/g, (_, i) => args[i]);
const summaryOf = (o) => (lang === 'ja' && o.summaryJa) || o.summary;
const docOf = (o) => (lang === 'ja' && o.docJa) || o.doc;

const TABS = ['theme', 'color', 'font', 'window', 'advanced'];

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
    const pick = el('input', { type: 'color', value: toHex(val),
      oninput: debounce((e) => { text.value = e.target.value; setOpt(o.key, e.target.value); }, 250) });
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

// ---------- font tab ----------

function fontPane() {
  const fam = state.set['font-family'] || '';
  const size = parseFloat(current(byKey['font-size'])) || 13;
  const sample = 'The quick brown fox jumps over the lazy dog\n0O 1lI |  => != === -> <= >= ~~>\n' + t('sampleCjk');
  const preview = el('div', { class: 'preview' },
    el('div', { class: 'muted small' }, t('fontPreview')),
    el('div', { class: 'sample', style: `font-family:${fam ? `"${fam.split('\n')[0]}", ` : ''}"SF Mono", Menlo, monospace;font-size:${size}px` }, sample));
  const head = el('div', {},
    el('h3', {}, t('font')),
    el('div', { class: 'opt' + (fam ? ' changed' : '') },
      el('div', { class: 'row' },
        el('div', { class: 'info' }, el('code', { class: 'key' }, 'font-family'), el('div', { class: 'sum' }, byKey['font-family'] && summaryOf(byKey['font-family']))),
        el('div', { class: 'ctl' },
          fonts
            ? el('select', { onchange: (e) => setOpt('font-family', e.target.value, true) },
                el('option', { value: '' }, t('defaultFont')),
                fonts.map((f) => el('option', { value: f, selected: f === fam.split('\n')[0] }, f)))
            : el('span', { class: 'muted' }, t('loadingFonts')),
          fam ? el('button', { class: 'reset', onclick: () => resetOpt('font-family') }, t('reset')) : null))),
    preview);
  if (!fonts) api().Fonts().then((f) => { fonts = f || []; render(); });
  return curatedPane(CURATED.font, head);
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
  else if (tab === 'advanced') view = advancedPane();
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
