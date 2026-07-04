// ─────────────────────────────────────────────
//  Coding Bootcamp — app.js
// ─────────────────────────────────────────────

// ── State ─────────────────────────────────────
const S = {
  user:         null,    // logged-in username
  lang:         'go',    // active language ID
  langs:        [],      // flat [{id, name, icon, ...}], derived from cats
  cats:         [],      // [{id, name, languages:[...]}] from /api/languages
  topics:       [],      // [{id, name, completed, lessonCached}] for active lang
  activeId:     1,
  activeTab:    'lesson',
  authTab:      'login',
  difficulty:   'beginner', // active challenge tier: beginner|intermediate|advanced|goat

  // Caches keyed by "lang:topicId" or "lang:track:trackId:lessonId";
  // challenges/challengeRaw additionally get a ":difficulty" suffix
  // (see activeChallengeKey).
  lessons:      {},
  challenges:   {},
  challengeRaw: {},

  // Chat history: lang → cacheKey → [{role,content}]
  chatHistory:  {},

  streaming:    false,

  // ── Track state ───────────────────────────
  mode:              'fundamentals', // 'fundamentals' | 'track' | 'project'
  tracks:            [],             // [{id, title, icon, description, lessons:[...]}]
  activeTrackId:     null,
  activeTrackLesson: null,           // {id, title, summary, completed, lessonCached}
  expandedTracks:    {},             // { trackId: bool }

  // ── Project state ─────────────────────────
  // A project is a capstone: one generated brief plus per-milestone build
  // guidance. The brief is selected as a synthetic milestone with id 0
  // (isBrief:true); real milestones are 1..N.
  projects:                [],       // [{id, title, icon, description, briefCached, milestones:[...]}]
  activeProjectId:         null,
  activeProjectMilestone:  null,     // {id, title, summary, completed, guidanceCached} | {id:0, isBrief:true}
  expandedProjects:        {},       // { projectId: bool }
};

// ── DOM helpers ────────────────────────────────
const $  = id => document.getElementById(id);
const el = (tag, cls, html) => {
  const e = document.createElement(tag);
  if (cls)  e.className = cls;
  if (html) e.innerHTML = html;
  return e;
};

function activeTopic() {
  return S.topics.find(t => t.id === S.activeId)
      || { id: S.activeId, name: '', completed: false };
}

function cacheKey(topicId) {
  return `${S.lang}:${topicId}`;
}
function trackCacheKey(trackId, lessonId) {
  return `${S.lang}:track:${trackId}:${lessonId}`;
}
function projectCacheKey(projectId, milestoneId) {
  return `${S.lang}:project:${projectId}:${milestoneId}`;
}
function activeCacheKey() {
  if (S.mode === 'track' && S.activeTrackId && S.activeTrackLesson) {
    return trackCacheKey(S.activeTrackId, S.activeTrackLesson.id);
  }
  if (S.mode === 'project' && S.activeProjectId && S.activeProjectMilestone) {
    return projectCacheKey(S.activeProjectId, S.activeProjectMilestone.id);
  }
  return cacheKey(S.activeId);
}

// Key for the challenge caches: challenges are generated per difficulty tier,
// so the selection key gets a tier suffix. Projects are the exception — the
// milestone guidance doubles as the challenge and has no tiers, so the plain
// selection key is used (matching what loadProjectGuidance stores under).
function activeChallengeKey() {
  if (S.mode === 'project') return activeCacheKey();
  return `${activeCacheKey()}:${S.difficulty}`;
}

// True when the active project selection is the brief/overview (milestone 0).
function briefSelected() {
  return S.mode === 'project' && S.activeProjectMilestone?.id === 0;
}

// ── Mode-aware request helpers ─────────────────
// These collapse the old fundamentals/track function pairs into one each:
// the endpoint prefix and body shape are the only things that differed.
function endpoint(action) {
  if (S.mode === 'project') {
    // The "lesson" action maps to the brief for the overview entry, or the
    // milestone guidance for a real milestone. Everything else (evaluate, hint,
    // chat) keeps its name under the /api/project/ prefix.
    if (action === 'lesson') return briefSelected() ? '/api/project/brief' : '/api/project/milestone';
    return `/api/project/${action}`;
  }
  return S.mode === 'track' ? `/api/track/${action}` : `/api/${action}`;
}

function reqBody(extra = {}) {
  if (S.mode === 'project') {
    return { lang: S.lang, project_id: S.activeProjectId, milestone_id: S.activeProjectMilestone?.id, ...extra };
  }
  if (S.mode === 'track') {
    return { lang: S.lang, track_id: S.activeTrackId, lesson_id: S.activeTrackLesson?.id, ...extra };
  }
  const t = activeTopic();
  return { lang: S.lang, topic_id: t.id, topic_name: t.name, ...extra };
}

// True when there's a valid selection to act on for the current mode.
function hasSelection() {
  if (S.mode === 'project') return !!S.activeProjectMilestone;
  return S.mode === 'track' ? !!S.activeTrackLesson : true;
}

// Pull the last fenced code block out of a challenge to seed the editor.
// Everything from the "## Hints" section on is ignored: a near-spoiler hint
// can contain its own code block, and starter code always comes before hints.
function extractStarterCode(markdown) {
  const beforeHints = markdown.split(/\n##\s+Hints\b/i)[0];
  const blocks = beforeHints.match(/```[\w]*\s*([\s\S]*?)```/g);
  if (!blocks) return '';
  const last = blocks[blocks.length - 1];
  return last.replace(/```[\w]*\s*/, '').replace(/```$/, '').trim();
}

// ── marked setup ───────────────────────────────
if (typeof marked.use === 'function') {
  marked.use({ gfm: true, breaks: true });
}

function parseMarkdown(text) {
  const result = marked.parse(text || '');
  return typeof result === 'string' ? result : (text || '');
}

function applyHighlight(container) {
  container.querySelectorAll('pre code').forEach(b => hljs.highlightElement(b));
}

// ── Zig highlighting ───────────────────────────
// highlight.js core ships Go but not Zig, so register a compact Zig grammar
// here (no build step, no extra CDN script). Lessons use ```zig fences, which
// become <code class="language-zig"> for highlightElement to pick up.
function registerZigHighlighting() {
  if (typeof hljs === 'undefined' || hljs.getLanguage('zig')) return;
  hljs.registerLanguage('zig', function (hl) {
    return {
      name: 'Zig',
      aliases: ['zig'],
      keywords: {
        keyword:
          'const var fn return if else while for switch break continue defer ' +
          'errdefer try catch orelse unreachable async await suspend resume nosuspend ' +
          'comptime inline noinline pub usingnamespace test struct enum union opaque ' +
          'error and or threadlocal export extern packed align linksection callconv ' +
          'volatile allowzero anyframe asm',
        type:
          'i8 u8 i16 u16 i32 u32 i64 u64 i128 u128 isize usize c_short c_ushort c_int ' +
          'c_uint c_long c_ulong c_longlong c_ulonglong c_longdouble f16 f32 f64 f80 ' +
          'f128 bool void noreturn type anyerror anytype anyopaque comptime_int comptime_float',
        literal: 'true false null undefined',
      },
      contains: [
        hl.QUOTE_STRING_MODE,
        { className: 'string', begin: /'(\\.|[^\\'])'/ },     // char literal
        { className: 'string', begin: /\\\\/, end: /$/, relevance: 0 }, // multiline string
        hl.COMMENT('//', '$'),
        { className: 'built_in', begin: /@[a-zA-Z_]\w*/ },     // @import, @typeOf, ...
        {
          className: 'number',
          variants: [
            { begin: /\b0x[0-9a-fA-F][0-9a-fA-F_]*(\.[0-9a-fA-F_]+)?([pP][-+]?\d+)?/ },
            { begin: /\b0o[0-7][0-7_]*/ },
            { begin: /\b0b[01][01_]*/ },
            { begin: /\b\d[\d_]*(\.[\d_]+)?([eE][-+]?\d+)?/ },
          ],
          relevance: 0,
        },
      ],
    };
  });
}
registerZigHighlighting();

// ── Toast ──────────────────────────────────────
let toastTimer = null;
function showToast(msg, type = 'info') {
  const t = $('toast');
  t.textContent = msg;
  t.className = `toast toast-${type}`;
  clearTimeout(toastTimer);
  toastTimer = setTimeout(() => { t.className = 'toast hidden'; }, 3500);
}

// ── Streaming fetch ────────────────────────────
async function streamFetch(url, body, onChunk, onDone) {
  if (S.streaming) return;
  S.streaming = true;
  let accumulated = '';
  try {
    const resp = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    });
    if (!resp.ok) {
      if (resp.status === 401) {
        // Session expired mid-request. Fold a message into the panel and call
        // onDone so the caller cleans up its streaming state (the panel would
        // otherwise be stuck behind the login modal mid-"stream").
        showLoginModal();
        accumulated += '\n\n**Session expired** — sign in and try again.';
        onChunk(accumulated, accumulated);
        S.streaming = false;
        onDone(accumulated);
        return;
      }
      const err = await resp.text();
      onChunk(`\n\n**Error ${resp.status}**: ${err}`, err);
      onDone(err);
      S.streaming = false;
      return;
    }
    const reader  = resp.body.getReader();
    const decoder = new TextDecoder();
    let buffer = '';
    while (true) {
      const { done, value } = await reader.read();
      if (done) break;
      buffer += decoder.decode(value, { stream: true });
      const lines = buffer.split('\n');
      buffer = lines.pop();
      for (const line of lines) {
        if (!line.startsWith('data: ')) continue;
        const raw = line.slice(6).trim();
        if (raw === '[DONE]') { S.streaming = false; onDone(accumulated); return; }
        try {
          const parsed = JSON.parse(raw);
          if (parsed.error) {
            // Fold the error into the accumulated text so the final onDone
            // render shows it instead of overwriting it with empty content.
            accumulated += '\n\n**Error**: ' + parsed.error;
            onChunk('\n\n**Error**: ' + parsed.error, accumulated);
            S.streaming = false; onDone(accumulated); return;
          }
          if (parsed.text) {
            accumulated += parsed.text;
            onChunk(parsed.text, accumulated);
          }
        } catch (_) {}
      }
    }
  } catch (err) {
    accumulated += '\n\n**Network error**: ' + err.message;
    onChunk('\n\n**Network error**: ' + err.message, accumulated);
  }
  S.streaming = false;
  onDone(accumulated);
}

// ── Auth ───────────────────────────────────────
function showLoginModal() {
  $('login-modal').classList.remove('hidden');
}
function hideLoginModal() {
  $('login-modal').classList.add('hidden');
}

function setAuthTab(tab) {
  S.authTab = tab;
  $('tab-login').classList.toggle('active', tab === 'login');
  $('tab-register').classList.toggle('active', tab === 'register');
  $('auth-submit').textContent = tab === 'login' ? 'Sign In' : 'Register';
  $('auth-error').classList.add('hidden');
}

function showAuthError(msg) {
  const el = $('auth-error');
  el.textContent = msg;
  el.classList.remove('hidden');
}

async function submitAuth() {
  const username = $('auth-username').value.trim();
  const pin      = $('auth-pin').value.trim();
  if (!username || !pin) { showAuthError('Please fill in both fields.'); return; }

  const url = S.authTab === 'register' ? '/api/auth/register' : '/api/auth/login';
  try {
    const resp = await fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ username, pin }),
    });
    const data = await resp.json();
    if (!resp.ok) { showAuthError(data.error || 'Something went wrong.'); return; }
    S.user = data.username;
    hideLoginModal();
    $('auth-pin').value = '';
    await postAuthInit();
  } catch (err) {
    showAuthError('Network error: ' + err.message);
  }
}

async function logout() {
  await fetch('/api/auth/logout', { method: 'POST' });
  S.user = null;
  S.topics = [];
  S.lessons = {};
  S.challenges = {};
  S.challengeRaw = {};
  S.chatHistory = {};
  showLoginModal();
}

async function checkAuth() {
  try {
    const resp = await fetch('/api/auth/me');
    if (!resp.ok) return null;
    return await resp.json();
  } catch (_) { return null; }
}

// ── Languages ──────────────────────────────────

// langIconHTML renders a language's brand logo. Icons are SVG files under
// /static/icons (set in languages.go); the image scales to the container's
// font-size via the .lang-icon rule. Falls back to a graduation-cap emoji when
// the language is unknown (e.g. a chat avatar before languages have loaded).
function langIconHTML(lang) {
  if (!lang || !lang.icon) return '🎓';
  return `<img class="lang-icon" src="${lang.icon}" alt="${lang.name || ''}">`;
}

async function loadLanguages() {
  try {
    const resp = await fetch('/api/languages');
    S.cats = await resp.json();
  } catch (_) {
    // If /api/languages is unreachable the server is down and nothing else
    // will work either — say so instead of carrying a stale duplicate of the
    // backend's language list here.
    S.cats = [];
    showToast('Could not load languages — is the server running?', 'error');
  }
  // Flatten into a single ordered list for lookups by id elsewhere.
  S.langs = S.cats.flatMap(c => c.languages);
  renderLangSwitcher();
}

function renderLangSwitcher() {
  const sw = $('lang-switcher');
  sw.innerHTML = '';
  S.cats.forEach(cat => {
    const group = el('div', 'lang-cat');
    const label = el('div', 'lang-cat-label', cat.name);
    const btns = el('div', 'lang-cat-btns');
    cat.languages.forEach(lang => {
      const btn = el('button', `lang-btn${lang.id === S.lang ? ' active' : ''}`);
      btn.dataset.lang = lang.id;
      btn.innerHTML = `${langIconHTML(lang)}<span>${lang.name}</span>`;
      btn.addEventListener('click', () => switchLang(lang.id));
      btns.appendChild(btn);
    });
    group.appendChild(label);
    group.appendChild(btns);
    sw.appendChild(group);
  });
}

function applyLangTheme(lang) {
  const root = document.documentElement;
  root.style.setProperty('--accent',      lang.accentColor);
  root.style.setProperty('--accent2',     lang.accentDark);
  root.style.setProperty('--accent-glow', lang.accentGlow);
  root.style.setProperty('--code-label',  `'${lang.codeLabel}'`);
  $('brand-icon').innerHTML    = langIconHTML(lang);
  $('brand-title').textContent = `${lang.name} Bootcamp`;
  $('brand-cmd').textContent   = lang.cmd;
  $('chat-intro-icon').innerHTML = langIconHTML(lang);
  document.querySelectorAll('.lang-btn').forEach(b => {
    b.classList.toggle('active', b.dataset.lang === lang.id);
  });
}

async function switchLang(id, reset = true) {
  const lang = S.langs.find(l => l.id === id);
  if (!lang) return;
  S.lang = id;
  S.mode = 'fundamentals';
  // Track and project selections don't carry across languages.
  S.activeTrackId = null;
  S.activeTrackLesson = null;
  S.activeProjectId = null;
  S.activeProjectMilestone = null;
  applyLangTheme(lang);
  await Promise.all([loadTopics(), loadTracks(), loadProjects()]);
  if (reset) {
    const firstIncomplete = S.topics.find(t => !t.completed);
    selectTopic(firstIncomplete ? firstIncomplete.id : 1, false);
    resetLessonPanel();
    resetChallengePanel();
    renderChat();
  }
}

// ── Progress bar ───────────────────────────────
function updateProgress() {
  const done  = S.topics.filter(t => t.completed).length;
  const total = S.topics.length || 1;
  const pct   = Math.round((done / total) * 100);
  $('progress-fill').style.width  = pct + '%';
  $('progress-label').textContent = `${done} / ${total} complete`;
}

// ── Sidebar topic list ─────────────────────────
function renderTopics() {
  const nav = $('topic-list');
  nav.innerHTML = '';
  S.topics.forEach(t => {
    const icon   = t.completed ? '✅' : (t.id === S.activeId ? '▶' : '○');
    const btn    = el('button', `topic-btn${t.id === S.activeId ? ' active' : ''}${t.completed ? ' done' : ''}`);
    btn.innerHTML = `<span class="topic-icon">${icon}</span><span class="topic-name">${t.name}</span><span class="topic-num">${t.id}</span>`;
    btn.addEventListener('click', () => selectTopic(t.id));
    nav.appendChild(btn);
  });
}

// ── Header ─────────────────────────────────────
function renderHeader() {
  // The difficulty bar lives in the challenge panel but depends on the mode,
  // and renderHeader runs on every selection change — update it here.
  renderDifficultyBar();

  // The brief hides the complete button; make sure it's visible again for
  // every other selection (the button is shared across all three modes).
  $('complete-btn').style.visibility = '';

  if (S.mode === 'project' && S.activeProjectId && S.activeProjectMilestone) {
    const project = S.projects.find(p => p.id === S.activeProjectId);
    const m = S.activeProjectMilestone;
    $('header-badge').textContent    = `${project?.icon || '🔗'} ${project?.title || 'Project'}`;
    $('header-title').textContent    = m.isBrief ? 'Project Brief' : m.title;
    $('chat-topic-name').textContent = m.isBrief ? (project?.title || 'this project') : m.title;
    if (m.isBrief) {
      // The overview isn't a completable build step.
      $('complete-btn').style.visibility = 'hidden';
      return;
    }
    const done = m.completed;
    $('complete-icon').textContent  = done ? '✅' : '○';
    $('complete-label').textContent = done ? 'Completed' : 'Mark complete';
    $('complete-btn').classList.toggle('done', done);
    return;
  }

  if (S.mode === 'track' && S.activeTrackId && S.activeTrackLesson) {
    const track = S.tracks.find(t => t.id === S.activeTrackId);
    const lesson = S.activeTrackLesson;
    $('header-badge').textContent    = `${track?.icon || '📘'} ${track?.title || 'Track'}`;
    $('header-title').textContent    = lesson.title;
    $('chat-topic-name').textContent = lesson.title;
    const done = lesson.completed;
    $('complete-icon').textContent   = done ? '✅' : '○';
    $('complete-label').textContent  = done ? 'Completed' : 'Mark complete';
    $('complete-btn').classList.toggle('done', done);
    return;
  }
  // Fundamentals
  const t = activeTopic();
  $('header-badge').textContent    = `Topic ${t.id}`;
  $('header-title').textContent    = t.name;
  $('chat-topic-name').textContent = t.name;
  const btn = $('complete-btn');
  if (t.completed) {
    $('complete-icon').textContent  = '✅';
    $('complete-label').textContent = 'Completed';
    btn.classList.add('done');
  } else {
    $('complete-icon').textContent  = '○';
    $('complete-label').textContent = 'Mark complete';
    btn.classList.remove('done');
  }
}

// ── Load topics ────────────────────────────────
async function loadTopics() {
  try {
    const resp = await fetch(`/api/topics?lang=${S.lang}`);
    if (resp.status === 401) { showLoginModal(); return; }
    S.topics = await resp.json();
  } catch (_) {
    S.topics = [];
  }
  renderTopics();
  updateProgress();
}

// ── Select topic ───────────────────────────────
function selectTopic(id, reset = true) {
  S.mode = 'fundamentals'; // leaving any track/project selection
  S.activeId = id;
  renderTopics();
  renderHeader();
  if (reset) {
    const key   = cacheKey(id);
    const topic = S.topics.find(t => t.id === id);

    // Restore each panel: JS cache → server cache (auto-load) → empty state.
    // Only one stream runs at a time (streamFetch guards on S.streaming), so
    // the tab the user is actually looking at goes first and wins the stream;
    // the other tab loads now only if it doesn't need the stream (already in
    // the JS cache), otherwise it's picked up lazily on tab switch (see
    // switchTab). This keeps the lesson and challenge panels symmetric.
    const restoreLesson = () => {
      if (S.lessons[key])           showLessonContent(S.lessons[key]);
      else if (topic?.lessonCached) loadLesson();
      else                          resetLessonPanel();
    };
    const restoreChallenge = () => {
      const ckey = activeChallengeKey();
      if (S.challenges[ckey])           showChallengeContent(S.challenges[ckey]);
      else if (activeChallengeCached()) loadChallenge();
      else                              resetChallengePanel();
    };
    if (S.activeTab === 'challenge') { restoreChallenge(); restoreLesson(); }
    else                             { restoreLesson(); restoreChallenge(); }

    closeEval();
    $('code-editor').value = '';
    renderChat();
  }
}

// ── Tab switching ──────────────────────────────
function switchTab(tab) {
  S.activeTab = tab;
  document.querySelectorAll('.tab').forEach(b => {
    b.classList.toggle('active', b.dataset.tab === tab);
  });
  document.querySelectorAll('.panel').forEach(p => {
    const show = p.id === `panel-${tab}`;
    p.classList.toggle('active',  show);
    p.classList.toggle('hidden', !show);
  });

  // Opening a tab whose content is cached server-side but hasn't been shown
  // yet: fetch it instantly (served from the server cache as a single chunk —
  // no regeneration, no token cost). This mirrors the eager load in
  // selectTopic for whichever tab wasn't the visible one at selection time.
  if (!S.streaming && hasSelection()) {
    if (tab === 'lesson'
        && !S.lessons[activeCacheKey()] && activeLessonCached()) {
      loadLesson();
    } else if (tab === 'challenge'
        && !S.challenges[activeChallengeKey()] && activeChallengeCached()) {
      loadChallenge();
    }
  }
}

// True when the current selection has a lesson (or brief/guidance) cached
// server-side.
function activeLessonCached() {
  if (S.mode === 'project') {
    const m = S.activeProjectMilestone;
    if (!m) return false;
    if (m.isBrief) {
      const p = S.projects.find(x => x.id === S.activeProjectId);
      return !!p?.briefCached;
    }
    return !!m.guidanceCached;
  }
  return S.mode === 'track'
    ? !!S.activeTrackLesson?.lessonCached
    : !!activeTopic()?.lessonCached;
}

// True when the current selection has a challenge cached server-side at the
// active difficulty tier (challengeCached is a per-tier map). Projects have no
// separate challenge document — the milestone guidance doubles as the build
// spec — so this is always false in project mode.
function activeChallengeCached() {
  if (S.mode === 'project') return false;
  const cached = S.mode === 'track'
    ? S.activeTrackLesson?.challengeCached
    : activeTopic()?.challengeCached;
  return !!cached?.[S.difficulty];
}

// ── Lesson panel ───────────────────────────────
function resetLessonPanel() {
  $('lesson-empty').classList.remove('hidden');
  $('lesson-output').classList.add('hidden');
  $('lesson-output').innerHTML = '';
  $('lesson-footer').classList.add('hidden');
}

function showLessonContent(html) {
  $('lesson-empty').classList.add('hidden');
  const out = $('lesson-output');
  out.classList.remove('hidden', 'streaming');
  out.innerHTML = html;
  applyHighlight(out);
  $('lesson-footer').classList.remove('hidden');
}

function loadLesson(force = false) {
  if (S.streaming || !hasSelection()) return;
  if (S.mode === 'project') {
    if (briefSelected()) loadProjectBrief(force);
    else                 loadProjectGuidance(force);
    return;
  }
  const key = activeCacheKey();
  $('lesson-empty').classList.add('hidden');
  $('lesson-footer').classList.add('hidden');
  const out = $('lesson-output');
  out.innerHTML = '';
  out.classList.remove('hidden');
  out.classList.add('streaming');

  streamFetch(
    endpoint('lesson'),
    reqBody({ force }),
    (_, acc) => { out.innerHTML = parseMarkdown(acc); },
    (full) => {
      out.classList.remove('streaming');
      out.innerHTML = parseMarkdown(full);
      applyHighlight(out);
      $('lesson-footer').classList.remove('hidden');
      S.lessons[key] = out.innerHTML;
    }
  );
}

// ── Challenge panel ────────────────────────────
function resetChallengePanel() {
  $('challenge-empty').classList.remove('hidden');
  $('challenge-inner').classList.add('hidden');
  $('challenge-output').innerHTML = '';
  closeEval();
}

function showChallengeContent(html) {
  $('challenge-empty').classList.add('hidden');
  $('challenge-inner').classList.remove('hidden');
  const out = $('challenge-output');
  out.classList.remove('streaming');
  out.innerHTML = html;
  if (S.mode !== 'project') collapseHints(out);
  applyHighlight(out);
}

// ── Hidden hints ───────────────────────────────
// Generated challenges end with a "## Hints" section. It stays hidden behind
// a reveal button as part of the challenge: peeking is fine, but it counts as
// hint use, and the evaluation praises a pass earned without any hints.

// Wrap the Hints heading and everything after it in a hidden container with a
// reveal button in front. Runs on every streaming chunk too, so hints never
// flash on screen mid-stream. The button uses an inline onclick because the
// challenge panel is cached and restored via innerHTML, which would drop a
// listener attached with addEventListener.
function collapseHints(out) {
  if (out.querySelector('.hints-wrap')) return; // already collapsed
  const h = [...out.querySelectorAll('h2')]
    .find(x => /^hints\b/i.test(x.textContent.trim()));
  if (!h) return;
  const btn = el('button', 'btn-secondary hints-reveal-btn', '💡 Reveal hints');
  btn.setAttribute('onclick', 'revealHints(this)');
  btn.title = 'Counts as using a hint';
  const wrap = el('div', 'hints-wrap hints-hidden');
  h.parentNode.insertBefore(btn, h);
  while (btn.nextSibling) wrap.appendChild(btn.nextSibling);
  btn.after(wrap);
}

function revealHints(btn) {
  const wrap = btn.nextElementSibling;
  if (wrap) wrap.classList.remove('hints-hidden');
  btn.remove();
  // Keep the revealed state when this tier is restored later in the session.
  S.challenges[activeChallengeKey()] = $('challenge-output').innerHTML;
  // Tell the server — revealing counts as using a hint, and the flag must
  // survive a page reload so the evaluation stays honest.
  fetch(endpoint('hints-viewed'), {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(reqBody({ difficulty: S.difficulty })),
  }).catch(() => {});
}

function loadChallenge(force = false) {
  if (S.streaming || !hasSelection()) return;
  if (S.mode === 'project') {
    // The overview/brief has no build step; everything else loads the
    // milestone guidance, which fills the challenge panel and the editor.
    if (briefSelected()) { showToast('Pick a milestone to start building', 'info'); return; }
    loadProjectGuidance(force);
    return;
  }
  const key = activeChallengeKey();
  $('challenge-empty').classList.add('hidden');
  $('challenge-inner').classList.remove('hidden');
  closeEval();
  const out = $('challenge-output');
  out.innerHTML = '';
  out.classList.add('streaming');

  streamFetch(
    endpoint('challenge'),
    reqBody({ force, difficulty: S.difficulty }),
    (_, acc) => { out.innerHTML = parseMarkdown(acc); collapseHints(out); },
    (full) => {
      out.classList.remove('streaming');
      S.challengeRaw[key] = full;
      out.innerHTML = parseMarkdown(full);
      collapseHints(out);
      applyHighlight(out);
      S.challenges[key] = out.innerHTML;
      // Pre-fill editor with starter code if empty
      const editor = $('code-editor');
      if (!editor.value.trim()) {
        const code = extractStarterCode(full);
        if (code) { editor.value = code; autoResize(editor); }
      }
    }
  );
}

// ── Challenge difficulty tiers ─────────────────
// Four tiers per lesson (Beginner → Intermediate → Advanced → GOAT), each
// generated and cached separately on the server. Switching tiers swaps the
// challenge panel to that tier's content without touching the others.

// Per-tier completion map for the active selection, e.g. {beginner: true}.
// Comes from /api/topics (fundamentals) or /api/tracks (tracks); projects have
// no challenge tiers, so this is always empty there.
function activeChallengeDone() {
  if (S.mode === 'project') return {};
  return (S.mode === 'track'
    ? S.activeTrackLesson?.challengeDone
    : activeTopic()?.challengeDone) || {};
}

function renderDifficultyBar() {
  // Projects have no separate challenge document, so tiers don't apply there.
  $('diff-bar').classList.toggle('hidden', S.mode === 'project');
  const done = activeChallengeDone();
  document.querySelectorAll('.diff-pill').forEach(b => {
    b.classList.toggle('active', b.dataset.diff === S.difficulty);
    b.classList.toggle('done', !!done[b.dataset.diff]);
  });
  renderChallengeCompleteBtn();
}

// The per-tier complete toggle in the challenge bar. Hidden for projects —
// milestones are completed with the header's Mark complete button instead.
function renderChallengeCompleteBtn() {
  const btn = $('challenge-complete-btn');
  if (S.mode === 'project') { btn.classList.add('hidden'); return; }
  btn.classList.remove('hidden');
  const isDone = !!activeChallengeDone()[S.difficulty];
  $('challenge-complete-icon').textContent  = isDone ? '✅' : '○';
  $('challenge-complete-label').textContent = isDone ? 'Challenge completed' : 'Mark challenge complete';
  btn.classList.toggle('done', isDone);
}

// Mark just the active difficulty tier's challenge complete (or undo it).
// Independent of the header's Mark complete: passing GOAT doesn't finish the
// topic, and finishing the topic doesn't tick any tier.
async function toggleChallengeComplete() {
  if (S.mode === 'project' || !hasSelection()) return;
  const next = !activeChallengeDone()[S.difficulty];
  try {
    await fetch('/api/progress', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(reqBody({ difficulty: S.difficulty, completed: next })),
    });
    // Update the object the done-map came from (a topic in S.topics, or the
    // active track lesson, which is the same object held in S.tracks).
    const target = S.mode === 'track' ? S.activeTrackLesson : activeTopic();
    if (target) {
      target.challengeDone = target.challengeDone || {};
      target.challengeDone[S.difficulty] = next;
    }
    renderDifficultyBar();
    showToast(next ? '✅ Challenge marked complete!' : 'Challenge marked incomplete', next ? 'success' : 'info');
  } catch (_) {
    showToast('Could not save progress', 'error');
  }
}

function setDifficulty(diff) {
  if (S.streaming || diff === S.difficulty) return;
  S.difficulty = diff;
  renderDifficultyBar();
  // Restore this tier's challenge: JS cache → server cache → empty state.
  const key = activeChallengeKey();
  if (S.challenges[key])            showChallengeContent(S.challenges[key]);
  else if (activeChallengeCached()) loadChallenge();
  else                              resetChallengePanel();
}

// ── Project content ────────────────────────────
// The brief is a project-level spec shown in the Lesson panel only (the
// overview entry has no build step). Milestone guidance, by contrast, doubles
// as the build spec: it fills BOTH the Lesson panel ("what to build") and the
// Challenge panel ("build it"), and seeds the editor with any starter code.

function loadProjectBrief(force = false) {
  const key = activeCacheKey();
  $('lesson-empty').classList.add('hidden');
  $('lesson-footer').classList.add('hidden');
  const out = $('lesson-output');
  out.innerHTML = '';
  out.classList.remove('hidden');
  out.classList.add('streaming');

  streamFetch(
    endpoint('lesson'),            // → /api/project/brief
    reqBody({ force }),
    (_, acc) => { out.innerHTML = parseMarkdown(acc); },
    (full) => {
      out.classList.remove('streaming');
      out.innerHTML = parseMarkdown(full);
      applyHighlight(out);
      $('lesson-footer').classList.remove('hidden');
      S.lessons[key] = out.innerHTML;
    }
  );
}

function loadProjectGuidance(force = false) {
  const key = activeCacheKey();
  // Lesson panel
  $('lesson-empty').classList.add('hidden');
  $('lesson-footer').classList.add('hidden');
  const lout = $('lesson-output');
  lout.innerHTML = '';
  lout.classList.remove('hidden');
  lout.classList.add('streaming');
  // Challenge panel (same guidance, shown alongside the editor)
  $('challenge-empty').classList.add('hidden');
  $('challenge-inner').classList.remove('hidden');
  closeEval();
  const cout = $('challenge-output');
  cout.innerHTML = '';
  cout.classList.add('streaming');

  streamFetch(
    endpoint('lesson'),            // → /api/project/milestone
    reqBody({ force }),
    (_, acc) => {
      const html = parseMarkdown(acc);
      lout.innerHTML = html;
      cout.innerHTML = html;
    },
    (full) => {
      const html = parseMarkdown(full);
      lout.classList.remove('streaming');
      lout.innerHTML = html;
      applyHighlight(lout);
      $('lesson-footer').classList.remove('hidden');
      S.lessons[key] = lout.innerHTML;

      cout.classList.remove('streaming');
      cout.innerHTML = html;
      applyHighlight(cout);
      S.challenges[key] = cout.innerHTML;
      // The guidance is what submit/hint evaluate against.
      S.challengeRaw[key] = full;

      const editor = $('code-editor');
      if (!editor.value.trim()) {
        const code = extractStarterCode(full);
        if (code) { editor.value = code; autoResize(editor); }
      }
    }
  );
}

// ── Submit / Hint / Eval ───────────────────────
function submitCode() {
  if (S.streaming || !hasSelection()) return;
  const code = $('code-editor').value.trim();
  const key  = activeChallengeKey();
  if (!code)                { showToast('Write some code first!', 'error'); return; }
  if (!S.challengeRaw[key]) { showToast('Load a challenge first!', 'error'); return; }

  const panel = $('eval-panel');
  const out   = $('eval-output');
  panel.classList.remove('hidden');
  out.innerHTML = '';
  out.classList.add('streaming');

  streamFetch(
    endpoint('evaluate'),
    // difficulty lets the server check hint use for this exact tier and
    // recognise a no-hints solve in the feedback.
    reqBody({ code, challenge: S.challengeRaw[key], difficulty: S.difficulty }),
    (_, acc) => { out.innerHTML = parseMarkdown(acc); },
    (full) => {
      out.classList.remove('streaming');
      out.innerHTML = parseMarkdown(full);
      applyHighlight(out);
      // Judge only the verdict line (the first line containing ✅ or ❌) —
      // scanning the whole response false-positives when a ❌ verdict's body
      // happens to mention "pass" or use a ✅ elsewhere.
      const verdictLine = full.split('\n').find(l => l.includes('✅') || l.includes('❌')) || '';
      if (verdictLine.includes('✅') && verdictLine.toLowerCase().includes('pass')) {
        showToast('✅ Challenge passed! Great work!', 'success');
      }
    }
  );
}

function getHint() {
  if (S.streaming || !hasSelection()) return;
  const key = activeChallengeKey();
  if (!S.challengeRaw[key]) { showToast('Load a challenge first!', 'error'); return; }

  const panel = $('eval-panel');
  const out   = $('eval-output');
  panel.classList.remove('hidden');
  out.innerHTML = '<p><strong>💡 Hint</strong></p>';
  out.classList.add('streaming');

  streamFetch(
    endpoint('hint'),
    // difficulty lets the server record hint use against this exact tier.
    reqBody({ challenge: S.challengeRaw[key], code: $('code-editor').value, difficulty: S.difficulty }),
    (_, acc) => { out.innerHTML = '<p><strong>💡 Hint</strong></p>' + parseMarkdown(acc); },
    (full) => {
      out.classList.remove('streaming');
      out.innerHTML = '<p><strong>💡 Hint</strong></p>' + parseMarkdown(full);
    }
  );
}

function closeEval()    { $('eval-panel').classList.add('hidden'); $('eval-output').innerHTML = ''; }
function clearEditor()  { $('code-editor').value = ''; closeEval(); }

// ── Chat ───────────────────────────────────────
function getChatHistory() {
  const key = activeCacheKey();
  if (!S.chatHistory[key]) S.chatHistory[key] = [];
  return S.chatHistory[key];
}

function renderChat() {
  const msgs = getChatHistory();
  const box  = $('chat-messages');
  box.querySelectorAll('.chat-msg').forEach(m => m.remove());
  msgs.forEach(m => box.appendChild(createChatBubble(m.role, m.content)));
  box.scrollTop = box.scrollHeight;
}

function createChatBubble(role, content) {
  const wrap   = el('div', `chat-msg chat-msg-${role}`);
  const avatar = el('div', 'chat-avatar', role === 'user' ? '🧑‍💻' : langIconHTML(S.langs.find(l => l.id === S.lang)));
  const bubble = el('div', 'chat-bubble');
  if (role === 'assistant') {
    bubble.innerHTML = parseMarkdown(content);
    applyHighlight(bubble);
  } else {
    bubble.textContent = content;
  }
  wrap.appendChild(avatar);
  wrap.appendChild(bubble);
  return wrap;
}

async function sendChat() {
  if (S.streaming || !hasSelection()) return;
  const input = $('chat-input');
  const text  = input.value.trim();
  if (!text) return;

  const history = getChatHistory();
  history.push({ role: 'user', content: text });
  input.value = '';
  autoResize(input);

  const box = $('chat-messages');
  box.appendChild(createChatBubble('user', text));

  const assistantWrap   = el('div', 'chat-msg chat-msg-assistant');
  const assistantAvatar = el('div', 'chat-avatar', langIconHTML(S.langs.find(l => l.id === S.lang)));
  const assistantBubble = el('div', 'chat-bubble streaming');
  assistantWrap.appendChild(assistantAvatar);
  assistantWrap.appendChild(assistantBubble);
  box.appendChild(assistantWrap);
  box.scrollTop = box.scrollHeight;

  streamFetch(
    endpoint('chat'),
    // difficulty lets the server include the challenge tier the student is
    // actually looking at in the chat context (ignored by project chat).
    reqBody({ messages: history, difficulty: S.difficulty }),
    (_, acc) => { assistantBubble.innerHTML = parseMarkdown(acc); box.scrollTop = box.scrollHeight; },
    (full) => {
      assistantBubble.classList.remove('streaming');
      assistantBubble.innerHTML = parseMarkdown(full);
      applyHighlight(assistantBubble);
      history.push({ role: 'assistant', content: full });
      box.scrollTop = box.scrollHeight;
    }
  );
}

// ── Mark complete ──────────────────────────────
async function toggleComplete() {
  if (S.mode === 'track')   { await toggleTrackComplete();   return; }
  if (S.mode === 'project') { await toggleProjectComplete(); return; }
  const t    = activeTopic();
  const next = !t.completed;
  try {
    await fetch('/api/progress', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ lang: S.lang, topic_id: t.id, completed: next }),
    });
    const idx = S.topics.findIndex(x => x.id === t.id);
    if (idx !== -1) S.topics[idx].completed = next;
    renderTopics();
    renderHeader();
    updateProgress();
    showToast(next ? '✅ Topic marked complete!' : 'Marked incomplete', next ? 'success' : 'info');
  } catch (_) {
    showToast('Could not save progress', 'error');
  }
}

// ── Helpers ────────────────────────────────────
function autoResize(el) {
  el.style.height = 'auto';
  el.style.height = Math.min(el.scrollHeight, 160) + 'px';
}

function setupEditorTab() {
  $('code-editor').addEventListener('keydown', e => {
    if (e.key === 'Tab') {
      e.preventDefault();
      const ta = e.target, s = ta.selectionStart, en = ta.selectionEnd;
      ta.value = ta.value.slice(0, s) + '\t' + ta.value.slice(en);
      ta.selectionStart = ta.selectionEnd = s + 1;
    }
  });
}

// Let the user drag the top edge of the Feedback panel to resize it.
// The code editor above has flex:1, so whatever height we give the panel
// is taken from the editor automatically.
function setupEvalResize() {
  const handle = $('eval-resize');
  const panel  = $('eval-panel');
  if (!handle) return;

  handle.addEventListener('mousedown', e => {
    e.preventDefault();
    const col       = panel.parentElement;          // .editor-col
    const startY    = e.clientY;
    const startH    = panel.offsetHeight;
    // Keep at least this much room for the editor above; cap so the panel
    // can't swallow the whole column.
    const maxH      = Math.max(120, col.clientHeight - 140);

    function onMove(ev) {
      // Dragging up (smaller clientY) grows the panel.
      let h = startH + (startY - ev.clientY);
      h = Math.max(120, Math.min(h, maxH));
      panel.style.maxHeight = 'none';               // override the CSS cap
      panel.style.height = h + 'px';
    }
    function onUp() {
      document.removeEventListener('mousemove', onMove);
      document.removeEventListener('mouseup', onUp);
      document.body.style.userSelect = '';
    }
    document.body.style.userSelect = 'none';         // no text selection while dragging
    document.addEventListener('mousemove', onMove);
    document.addEventListener('mouseup', onUp);
  });
}

// ── Track sidebar ──────────────────────────────
async function loadTracks() {
  try {
    const resp = await fetch(`/api/tracks?lang=${S.lang}`);
    if (resp.status === 401) { showLoginModal(); return; }
    S.tracks = await resp.json();
  } catch (_) {
    S.tracks = [];
  }
  renderTrackList();
}

function renderTrackList() {
  const list = $('track-list');
  list.innerHTML = '';
  S.tracks.forEach(track => {
    const done    = track.lessons.filter(l => l.completed).length;
    const isOpen  = !!S.expandedTracks[track.id];
    const isActive = S.mode === 'track' && S.activeTrackId === track.id;
    const item = document.createElement('div');
    item.className = `track-item${isOpen ? ' open' : ''}`;
    item.innerHTML = `
      <button class="track-header" onclick="toggleTrack('${track.id}')">
        <span class="track-icon">${track.icon}</span>
        <span class="track-title">${track.title}</span>
        <span class="track-progress">${done}/${track.lessons.length}</span>
        <span class="track-chevron">▶</span>
      </button>
      <ul class="track-lessons" id="tl-${track.id}"></ul>`;
    const ul = item.querySelector('.track-lessons');
    track.lessons.forEach(lesson => {
      const isLessonActive = isActive && S.activeTrackLesson?.id === lesson.id;
      const li = document.createElement('li');
      const btn = document.createElement('button');
      btn.className = `track-lesson-btn${isLessonActive ? ' active' : ''}${lesson.completed ? ' done' : ''}`;
      btn.innerHTML = `<span class="tl-num">${lesson.id}</span><span class="tl-title">${lesson.title}</span><span class="tl-done">${lesson.completed ? '✅' : ''}</span>`;
      btn.addEventListener('click', () => selectTrackLesson(track.id, lesson.id));
      li.appendChild(btn);
      ul.appendChild(li);
    });
    list.appendChild(item);
  });
}

function toggleTrack(trackId) {
  S.expandedTracks[trackId] = !S.expandedTracks[trackId];
  renderTrackList();
}

// ── Select track lesson ────────────────────────
function selectTrackLesson(trackId, lessonId) {
  const track  = S.tracks.find(t => t.id === trackId);
  const lesson = track?.lessons.find(l => l.id === lessonId);
  if (!track || !lesson) return;

  S.mode              = 'track';
  S.activeTrackId     = trackId;
  S.activeTrackLesson = lesson;
  S.expandedTracks[trackId] = true;

  renderTrackList();
  renderHeader();
  resetChallengePanel();
  closeEval();
  $('code-editor').value = '';
  renderChat();

  const key = trackCacheKey(trackId, lessonId);
  if (S.lessons[key]) {
    showLessonContent(S.lessons[key]);
  } else if (lesson.lessonCached) {
    loadLesson();
  } else {
    resetLessonPanel();
  }

  switchTab('lesson');
}

// ── Track content functions ────────────────────
// (loadLesson/loadChallenge/submitCode/getHint/sendChat are now mode-aware
//  and handle tracks too — see the unified versions above.)

async function toggleTrackComplete() {
  if (!S.activeTrackLesson) return;
  const next = !S.activeTrackLesson.completed;
  try {
    await fetch('/api/progress', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ lang: S.lang, track_id: S.activeTrackId, lesson_id: S.activeTrackLesson.id, completed: next }),
    });
    S.activeTrackLesson.completed = next;
    // Also update in S.tracks
    const track = S.tracks.find(t => t.id === S.activeTrackId);
    if (track) {
      const lesson = track.lessons.find(l => l.id === S.activeTrackLesson.id);
      if (lesson) lesson.completed = next;
    }
    renderHeader();
    renderTrackList();
    showToast(next ? '✅ Lesson marked complete!' : 'Marked incomplete', next ? 'success' : 'info');
  } catch (_) {
    showToast('Could not save progress', 'error');
  }
}

// ── Project sidebar ────────────────────────────
async function loadProjects() {
  try {
    const resp = await fetch(`/api/projects?lang=${S.lang}`);
    if (resp.status === 401) { showLoginModal(); return; }
    S.projects = await resp.json();
  } catch (_) {
    S.projects = [];
  }
  renderProjectList();
}

function renderProjectList() {
  const list = $('project-list');
  list.innerHTML = '';
  S.projects.forEach(project => {
    const done     = project.milestones.filter(m => m.completed).length;
    const isOpen   = !!S.expandedProjects[project.id];
    const isActive = S.mode === 'project' && S.activeProjectId === project.id;
    const item = document.createElement('div');
    item.className = `track-item${isOpen ? ' open' : ''}`;
    item.innerHTML = `
      <button class="track-header" onclick="toggleProject('${project.id}')">
        <span class="track-icon">${project.icon}</span>
        <span class="track-title">${project.title}</span>
        <span class="track-progress">${done}/${project.milestones.length}</span>
        <span class="track-chevron">▶</span>
      </button>
      <ul class="track-lessons" id="pl-${project.id}"></ul>`;
    const ul = item.querySelector('.track-lessons');

    // Overview / brief entry (synthetic milestone id 0).
    const briefActive = isActive && S.activeProjectMilestone?.id === 0;
    const bli  = document.createElement('li');
    const bbtn = document.createElement('button');
    bbtn.className = `track-lesson-btn${briefActive ? ' active' : ''}`;
    bbtn.innerHTML = `<span class="tl-num">📋</span><span class="tl-title">Overview</span><span class="tl-done"></span>`;
    bbtn.addEventListener('click', () => selectProjectMilestone(project.id, 0));
    bli.appendChild(bbtn);
    ul.appendChild(bli);

    // Build milestones.
    project.milestones.forEach(m => {
      const mActive = isActive && S.activeProjectMilestone?.id === m.id;
      const li  = document.createElement('li');
      const btn = document.createElement('button');
      btn.className = `track-lesson-btn${mActive ? ' active' : ''}${m.completed ? ' done' : ''}`;
      btn.innerHTML = `<span class="tl-num">${m.id}</span><span class="tl-title">${m.title}</span><span class="tl-done">${m.completed ? '✅' : ''}</span>`;
      btn.addEventListener('click', () => selectProjectMilestone(project.id, m.id));
      li.appendChild(btn);
      ul.appendChild(li);
    });
    list.appendChild(item);
  });
}

function toggleProject(projectId) {
  S.expandedProjects[projectId] = !S.expandedProjects[projectId];
  renderProjectList();
}

// ── Select project milestone ───────────────────
function selectProjectMilestone(projectId, milestoneId) {
  const project = S.projects.find(p => p.id === projectId);
  if (!project) return;

  let milestone;
  if (milestoneId === 0) {
    milestone = { id: 0, title: 'Overview', summary: '', isBrief: true };
  } else {
    milestone = project.milestones.find(m => m.id === milestoneId);
  }
  if (!milestone) return;

  S.mode                   = 'project';
  S.activeProjectId        = projectId;
  S.activeProjectMilestone = milestone;
  S.expandedProjects[projectId] = true;

  renderProjectList();
  renderHeader();
  resetChallengePanel();
  closeEval();
  $('code-editor').value = '';
  renderChat();

  const key = activeCacheKey();

  // Restore the lesson/brief panel: JS cache → server cache → empty state.
  if (S.lessons[key]) {
    showLessonContent(S.lessons[key]);
  } else if (activeLessonCached()) {
    loadLesson();
  } else {
    resetLessonPanel();
  }

  // For a real milestone, restore the challenge panel from the JS cache too so
  // the build view is ready without a refetch. (loadLesson above fills both
  // panels when it has to fetch.)
  if (!milestone.isBrief && S.challenges[key]) {
    showChallengeContent(S.challenges[key]);
  }

  switchTab('lesson');
}

async function toggleProjectComplete() {
  const m = S.activeProjectMilestone;
  if (!m || m.isBrief) return;
  const next = !m.completed;
  try {
    await fetch('/api/progress', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ lang: S.lang, project_id: S.activeProjectId, milestone_id: m.id, completed: next }),
    });
    m.completed = next;
    // Also update the copy held in S.projects.
    const project = S.projects.find(p => p.id === S.activeProjectId);
    if (project) {
      const ms = project.milestones.find(x => x.id === m.id);
      if (ms) ms.completed = next;
    }
    renderHeader();
    renderProjectList();
    showToast(next ? '✅ Milestone marked complete!' : 'Marked incomplete', next ? 'success' : 'info');
  } catch (_) {
    showToast('Could not save progress', 'error');
  }
}

// ── Post-auth init ─────────────────────────────
async function postAuthInit() {
  // A fresh login can follow an expired session whose failed loads left error
  // text in the JS content caches — start clean so it can't resurface.
  S.lessons      = {};
  S.challenges   = {};
  S.challengeRaw = {};
  S.chatHistory  = {};

  $('username-display').textContent = S.user;

  const defaultLang = S.langs[0]?.id || 'go';
  await switchLang(defaultLang, false); // loads topics + tracks

  // Start on first incomplete topic; reset=true triggers cache-aware lesson logic
  const firstIncomplete = S.topics.find(t => !t.completed);
  selectTopic(firstIncomplete ? firstIncomplete.id : 1, true);

  switchTab('lesson');
  setupEditorTab();
  setupEvalResize();
}

// ── Init ───────────────────────────────────────
async function init() {
  await loadLanguages();

  const me = await checkAuth();
  if (!me) {
    showLoginModal();
    return; // submitAuth() calls postAuthInit() on success
  }
  // Already authenticated (an existing session, or -dev auto-login). The modal
  // is visible by default, so hide it here — only submitAuth() would otherwise.
  hideLoginModal();
  S.user = me.username;
  await postAuthInit();
}

init();
