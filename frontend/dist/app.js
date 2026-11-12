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

const TABS = [['theme', 'テーマ'], ['color', '色'], ['font', 'フォント'], ['window', 'ウィンドウ'], ['advanced', '詳細設定']];

const CURATED = {
  color: [
    ['基本の色', ['background', 'foreground', 'bold-color', 'minimum-contrast']],
    ['カーソル・選択', ['cursor-color', 'cursor-text', 'selection-background', 'selection-foreground']],
    ['透明度・分割', ['background-opacity', 'unfocused-split-opacity', 'unfocused-split-fill', 'split-divider-color']],
  ],
  font: [
    ['サイズ・太さ', ['font-size', 'font-thicken', 'font-thicken-strength', 'font-synthetic-style']],
    ['文字の詰め', ['adjust-cell-height', 'adjust-cell-width', 'font-feature']],
  ],
  window: [
    ['余白', ['window-padding-x', 'window-padding-y', 'window-padding-balance', 'window-padding-color']],
    ['見た目', ['background-opacity', 'background-blur', 'window-decoration', 'window-theme', 'macos-titlebar-style']],
    ['カーソル', ['cursor-style', 'cursor-style-blink']],
    ['起動・終了', ['window-save-state', 'confirm-close-surface', 'window-inherit-working-directory']],
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
  const t = document.getElementById('toast');
  t.textContent = msg;
  t.className = 'show' + (err ? ' err' : '');
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => (t.className = ''), err ? 6000 : 2200);
}

function debounce(fn, ms) {
  let t;
  return (...a) => { clearTimeout(t); t = setTimeout(() => fn(...a), ms); };
}

async function apply(promise, rerender = false) {
  const res = await promise;
  if (res.set) state.set = res.set;
  toast(res.message || (res.ok ? '保存しました' : '失敗しました'), !res.ok);
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
    return el('textarea', { readonly: true, title: '複数行の設定は試作品では編集できません' }, val);
  }
  if (o.kind === 'bool') {
    return el('input', { type: 'checkbox', checked: val === 'true', onchange: (e) => setOpt(o.key, e.target.checked) });
  }
  if (o.kind === 'enum') {
    return el('select', { onchange: (e) => setOpt(o.key, e.target.value) },
      o.choices.map((c) => el('option', { value: c, selected: c === val }, c)));
  }
  if (o.kind === 'color') {
    const text = el('input', { type: 'text', value: val, placeholder: 'テーマに従う', style: 'width:110px',
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
  return el('input', { type: 'text', value: val, placeholder: o.default || '(未設定)', onchange: (e) => setOpt(o.key, e.target.value) });
}

const RELOAD_LABEL = { newwindow: '新しいウィンドウから', restart: '再起動が必要' };

function optRow(o) {
  const changed = state.set[o.key] !== undefined;
  return el('div', { class: 'opt' + (changed ? ' changed' : ''), 'data-key': o.key },
    el('div', { class: 'row' },
      el('div', { class: 'info' },
        el('code', { class: 'key' }, o.key),
        RELOAD_LABEL[o.reload] ? el('span', { class: 'badge ' + o.reload }, RELOAD_LABEL[o.reload]) : null,
        el('div', { class: 'sum' }, o.summary),
        o.doc.length > o.summary.length + 20 ? el('details', {}, el('summary', {}, '詳しい説明'), el('pre', {}, o.doc)) : null),
      el('div', { class: 'ctl' },
        control(o),
        changed ? el('button', { class: 'reset', title: 'この行を設定ファイルから消して既定値に戻す', onclick: () => resetOpt(o.key) }, '既定に戻す') : null)));
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
    wrap.append(el('h3', {}, title), ...opts.map(optRow));
  }
  return wrap;
}

// ---------- theme tab ----------

function themeCard(t) {
  const on = state.set.theme === t.name;
  const p = t.palette;
  const line = (...parts) => el('div', {}, ...parts.map(([c, s]) => el('span', { style: `color:${c || t.foreground}` }, s)));
  return el('div', { class: 'th' + (on ? ' on' : ''), title: t.name, onclick: () => setOpt('theme', t.name, true) },
    el('div', { class: 'pv', style: `background:${t.background};color:${t.foreground}` },
      line([p[2], '➜ '], [p[6], '~/src '], [p[4], 'git:('], [p[1], 'main'], [p[4], ')']),
      line([null, 'ls '], [p[3], '-la']),
      line([p[4], 'docs/ '], [p[2], 'run.sh '], [p[5], 'img.png']),
      el('div', { class: 'pal' }, p.slice(0, 8).map((c) => el('i', { style: `background:${c}` })))),
    el('div', { class: 'nm' }, t.name));
}

function themePane() {
  const wrap = el('div', { class: 'scroll' });
  const overridden = ['background', 'foreground', 'palette', 'cursor-color', 'selection-background'].filter((k) => state.set[k] !== undefined);
  if (!themes) {
    wrap.append(el('p', { class: 'muted' }, 'テーマを読み込み中…'));
    api().Themes().then((t) => { themes = t || []; render(); });
    return wrap;
  }
  const grid = el('div', { class: 'themes' });
  const fill = () => {
    grid.replaceChildren();
    const q = themeQuery.toLowerCase();
    themes
      .filter((t) => (!q || t.name.toLowerCase().includes(q)) && (themeMode === 'all' || (themeMode === 'dark') === t.dark))
      .forEach((t) => grid.append(themeCard(t)));
  };
  wrap.append(...[
    el('div', { class: 'toolbar' },
      el('input', { type: 'search', placeholder: `テーマを検索（${themes.length} 件）`, value: themeQuery,
        oninput: (e) => { themeQuery = e.target.value; fill(); } }),
      el('select', { onchange: (e) => { themeMode = e.target.value; fill(); } },
        [['all', 'すべて'], ['dark', 'ダーク'], ['light', 'ライト']].map(([v, l]) => el('option', { value: v, selected: v === themeMode }, l))),
      el('span', { class: 'muted' }, '現在: ', el('b', {}, state.set.theme || '（既定）')),
      state.set.theme ? el('button', { class: 'reset', onclick: () => resetOpt('theme') }, '既定に戻す') : null),
    overridden.length ? el('div', { class: 'note' }, `色タブで ${overridden.join(', ')} を上書きしているため、テーマの一部が隠れます。`) : null,
    grid,
  ].filter(Boolean));
  fill();
  return wrap;
}

// ---------- font tab ----------

function fontPane() {
  const fam = state.set['font-family'] || '';
  const size = parseFloat(current(byKey['font-size'])) || 13;
  const sample = 'The quick brown fox jumps over the lazy dog\n0O 1lI |  => != === -> <= >= ~~>\nいろはにほへと　日本語の表示もここで確認';
  const preview = el('div', { class: 'preview' },
    el('div', { class: 'muted small' }, 'プレビュー（このアプリ内での表示です。実際の Ghostty にも同時に反映されます）'),
    el('div', { class: 'sample', style: `font-family:${fam ? `"${fam.split('\n')[0]}", ` : ''}"SF Mono", Menlo, monospace;font-size:${size}px` }, sample));
  const head = el('div', {},
    el('h3', {}, 'フォント'),
    el('div', { class: 'opt' + (fam ? ' changed' : '') },
      el('div', { class: 'row' },
        el('div', { class: 'info' }, el('code', { class: 'key' }, 'font-family'), el('div', { class: 'sum' }, byKey['font-family']?.summary)),
        el('div', { class: 'ctl' },
          fonts
            ? el('select', { onchange: (e) => setOpt('font-family', e.target.value, true) },
                el('option', { value: '' }, '（既定のフォント）'),
                fonts.map((f) => el('option', { value: f, selected: f === fam.split('\n')[0] }, f)))
            : el('span', { class: 'muted' }, 'フォント一覧を読み込み中…'),
          fam ? el('button', { class: 'reset', onclick: () => resetOpt('font-family') }, '既定に戻す') : null))),
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
      ? state.options.filter((o) => o.key.includes(q) || o.summary.toLowerCase().includes(q))
      : advCat === '__changed' ? state.options.filter((o) => state.set[o.key] !== undefined) : cats[advCat];
    list.replaceChildren(el('div', { class: 'scroll' },
      el('div', { class: 'muted small', style: 'margin-bottom:8px' }, `${opts.length} 項目`),
      opts.map(optRow)));
    list.scrollTop = 0;
  };
  const side = el('aside', {},
    el('input', { type: 'search', placeholder: `全 ${state.options.length} 項目を検索`, value: advQuery,
      oninput: (e) => { advQuery = e.target.value; fillList(); } }),
    el('div', { class: 'cat' + (advCat === '__changed' ? ' on' : ''), onclick: () => pick('__changed') },
      '変更済み', el('span', {}, Object.keys(state.set).length)),
    names.map((n) => el('div', { class: 'cat' + (n === advCat ? ' on' : ''), onclick: () => pick(n) }, n, el('span', {}, cats[n].length))));
  const pick = (n) => { advCat = n; advQuery = ''; render(); };
  fillList();
  return el('div', { class: 'adv' }, side, list);
}

// ---------- shell ----------

function render(keepScroll = false) {
  const pane = document.getElementById('pane');
  const scroll = keepScroll ? (pane.querySelector('.adv section') || pane).scrollTop : 0;
  document.getElementById('tabs').replaceChildren(...TABS.map(([id, label]) =>
    el('button', { class: id === tab ? 'on' : '', onclick: () => { tab = id; render(); } }, label)));
  let view;
  if (state.error) view = el('p', { class: 'pad' }, state.error);
  else if (tab === 'theme') view = themePane();
  else if (tab === 'font') view = fontPane();
  else if (tab === 'advanced') view = advancedPane();
  else view = curatedPane(CURATED[tab]);
  pane.replaceChildren(view);
  (pane.querySelector('.adv section') || pane).scrollTop = scroll;
}

async function boot() {
  state = await api().Load();
  state.options ||= [];
  state.set ||= {};
  byKey = Object.fromEntries(state.options.map((o) => [o.key, o]));
  document.getElementById('version').textContent = state.version || '';
  document.getElementById('path').textContent = state.configPath + (state.backupPath ? `\nバックアップ: ${state.backupPath}` : '');
  render();
}

window.addEventListener('DOMContentLoaded', () => {
  const wait = () => (window.go ? boot() : setTimeout(wait, 30));
  wait();
});
