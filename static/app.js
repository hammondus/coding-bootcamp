// ─────────────────────────────────────────────
//  Coding Bootcamp — app.js
// ─────────────────────────────────────────────

// ── State ─────────────────────────────────────
const S = {
  user:         null,    // logged-in username
  lang:         'go',    // active language ID
  langs:        [],      // [{id, name, icon, ...}] from /api/languages
  topics:       [],      // [{id, name, completed, lessonCached}] for active lang
  activeId:     1,
  activeTab:    'lesson',
  authTab:      'login',

  // Caches keyed by "lang:topicId" or "lang:track:trackId:lessonId"
  lessons:      {},
  challenges:   {},
  challengeRaw: {},

  // Chat history: lang → cacheKey → [{role,content}]
  chatHistory:  {},

  streaming:    false,

  // ── Track state ───────────────────────────
  mode:              'fundamentals', // 'fundamentals' | 'track'
  tracks:            [],             // [{id, title, icon, description, lessons:[...]}]
  activeTrackId:     null,
  activeTrackLesson: null,           // {id, title, summary, completed, lessonCached}
  expandedTracks:    {},             // { trackId: bool }
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
function activeCacheKey() {
  if (S.mode === 'track' && S.activeTrackId && S.activeTrackLesson) {
    return trackCacheKey(S.activeTrackId, S.activeTrackLesson.id);
  }
  return cacheKey(S.activeId);
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
      if (resp.status === 401) { showLoginModal(); S.streaming = false; return; }
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
            onChunk('\n\n**Error**: ' + parsed.error, parsed.error);
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
    onChunk('\n\n**Network error**: ' + err.message, err.message);
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
async function loadLanguages() {
  try {
    const resp = await fetch('/api/languages');
    S.langs = await resp.json();
  } catch (_) {
    S.langs = [
      { id: 'go',  name: 'Go',  icon: '🐹', cmd: '$ go run .', accentColor: '#00ADD8', accentDark: '#007fa0', accentGlow: 'rgba(0,173,216,0.15)', codeLabel: 'GO' },
      { id: 'zig', name: 'Zig', icon: '⚡', cmd: '$ zig build run', accentColor: '#F7A41D', accentDark: '#C47D0A', accentGlow: 'rgba(247,164,29,0.15)', codeLabel: 'ZIG' },
    ];
  }
  renderLangSwitcher();
}

function renderLangSwitcher() {
  const sw = $('lang-switcher');
  sw.innerHTML = '';
  S.langs.forEach(lang => {
    const btn = el('button', `lang-btn${lang.id === S.lang ? ' active' : ''}`);
    btn.dataset.lang = lang.id;
    btn.innerHTML = `<span>${lang.icon}</span><span>${lang.name}</span>`;
    btn.addEventListener('click', () => switchLang(lang.id));
    sw.appendChild(btn);
  });
}

function applyLangTheme(lang) {
  const root = document.documentElement;
  root.style.setProperty('--accent',      lang.accentColor);
  root.style.setProperty('--accent2',     lang.accentDark);
  root.style.setProperty('--accent-glow', lang.accentGlow);
  root.style.setProperty('--code-label',  `'${lang.codeLabel}'`);
  $('brand-icon').textContent  = lang.icon;
  $('brand-title').textContent = `${lang.name} Bootcamp`;
  $('brand-cmd').textContent   = lang.cmd;
  $('chat-intro-icon').textContent = lang.icon;
  document.querySelectorAll('.lang-btn').forEach(b => {
    b.classList.toggle('active', b.dataset.lang === lang.id);
  });
}

async function switchLang(id, reset = true) {
  const lang = S.langs.find(l => l.id === id);
  if (!lang) return;
  S.lang = id;
  S.mode = 'fundamentals';
  applyLangTheme(lang);
  await Promise.all([loadTopics(), loadTracks()]);
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
  S.activeId = id;
  renderTopics();
  renderHeader();
  if (reset) {
    const key   = cacheKey(id);
    const topic = S.topics.find(t => t.id === id);

    // Lesson: JS cache → server cache (auto-load) → empty state
    if (S.lessons[key]) {
      showLessonContent(S.lessons[key]);
    } else if (topic?.lessonCached) {
      loadLesson(); // server has it — fetch instantly, no button click needed
    } else {
      resetLessonPanel();
    }

    if (!S.challenges[key]) resetChallengePanel();
    else                    showChallengeContent(S.challenges[key]);

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
  if (S.mode === 'track') { loadTrackLesson(force); return; }
  if (S.streaming) return;
  const t = activeTopic();
  $('lesson-empty').classList.add('hidden');
  $('lesson-footer').classList.add('hidden');
  const out = $('lesson-output');
  out.innerHTML = '';
  out.classList.remove('hidden');
  out.classList.add('streaming');

  streamFetch(
    '/api/lesson',
    { lang: S.lang, topic_id: t.id, topic_name: t.name, force },
    (_, acc) => { out.innerHTML = parseMarkdown(acc); },
    (full) => {
      out.classList.remove('streaming');
      out.innerHTML = parseMarkdown(full);
      applyHighlight(out);
      $('lesson-footer').classList.remove('hidden');
      S.lessons[cacheKey(t.id)] = out.innerHTML;
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
  applyHighlight(out);
}

function loadChallenge() {
  if (S.mode === 'track') { loadTrackChallenge(); return; }
  if (S.streaming) return;
  const t = activeTopic();
  $('challenge-empty').classList.add('hidden');
  $('challenge-inner').classList.remove('hidden');
  closeEval();
  const out = $('challenge-output');
  out.innerHTML = '';
  out.classList.add('streaming');

  streamFetch(
    '/api/challenge',
    { lang: S.lang, topic_id: t.id, topic_name: t.name },
    (_, acc) => { out.innerHTML = parseMarkdown(acc); },
    (full) => {
      out.classList.remove('streaming');
      S.challengeRaw[cacheKey(t.id)] = full;
      out.innerHTML = parseMarkdown(full);
      applyHighlight(out);
      S.challenges[cacheKey(t.id)] = out.innerHTML;
      // Pre-fill editor with starter code if empty
      const editor = $('code-editor');
      if (!editor.value.trim()) {
        const match = full.match(/```[\w]*\s*([\s\S]*?)```/g);
        if (match) {
          const last = match[match.length - 1];
          const code = last.replace(/```[\w]*\s*/, '').replace(/```$/, '').trim();
          editor.value = code;
          autoResize(editor);
        }
      }
    }
  );
}

// ── Submit / Hint / Eval ───────────────────────
function submitCode() {
  if (S.mode === 'track') { submitTrackCode(); return; }
  if (S.streaming) return;
  const t    = activeTopic();
  const code = $('code-editor').value.trim();
  const key  = cacheKey(t.id);
  if (!code)              { showToast('Write some code first!', 'error'); return; }
  if (!S.challengeRaw[key]) { showToast('Load a challenge first!', 'error'); return; }

  const panel = $('eval-panel');
  const out   = $('eval-output');
  panel.classList.remove('hidden');
  out.innerHTML = '';
  out.classList.add('streaming');

  streamFetch(
    '/api/evaluate',
    { lang: S.lang, topic_id: t.id, topic_name: t.name, code, challenge: S.challengeRaw[key] },
    (_, acc) => { out.innerHTML = parseMarkdown(acc); },
    (full) => {
      out.classList.remove('streaming');
      out.innerHTML = parseMarkdown(full);
      applyHighlight(out);
      const lower = full.toLowerCase();
      if (full.includes('✅') && (lower.includes('pass') || lower.includes('correct') || lower.includes('well done'))) {
        showToast('✅ Challenge passed! Great work!', 'success');
      }
    }
  );
}

function getHint() {
  if (S.mode === 'track') { getTrackHint(); return; }
  if (S.streaming) return;
  const t   = activeTopic();
  const key = cacheKey(t.id);
  if (!S.challengeRaw[key]) { showToast('Load a challenge first!', 'error'); return; }

  const panel = $('eval-panel');
  const out   = $('eval-output');
  panel.classList.remove('hidden');
  out.innerHTML = '<p><strong>💡 Hint</strong></p>';
  out.classList.add('streaming');

  streamFetch(
    '/api/hint',
    { lang: S.lang, topic_id: t.id, topic_name: t.name, challenge: S.challengeRaw[key], code: $('code-editor').value },
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
  const avatar = el('div', 'chat-avatar', role === 'user' ? '🧑‍💻' : (S.langs.find(l => l.id === S.lang)?.icon || '🎓'));
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
  if (S.mode === 'track') { await sendTrackChat(); return; }
  if (S.streaming) return;
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
  const assistantAvatar = el('div', 'chat-avatar', S.langs.find(l => l.id === S.lang)?.icon || '🎓');
  const assistantBubble = el('div', 'chat-bubble streaming');
  assistantWrap.appendChild(assistantAvatar);
  assistantWrap.appendChild(assistantBubble);
  box.appendChild(assistantWrap);
  box.scrollTop = box.scrollHeight;

  const t = activeTopic();
  streamFetch(
    '/api/chat',
    { lang: S.lang, topic_id: t.id, topic_name: t.name, messages: history },
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

function onChatKey(e) {
  if (e.key === 'Enter' && !e.shiftKey) { e.preventDefault(); sendChat(); }
}

// ── Mark complete ──────────────────────────────
async function toggleComplete() {
  if (S.mode === 'track') { await toggleTrackComplete(); return; }
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
    loadTrackLesson();
  } else {
    resetLessonPanel();
  }

  switchTab('lesson');
}

// ── Track content functions ────────────────────
function trackBody() {
  return {
    lang:      S.lang,
    track_id:  S.activeTrackId,
    lesson_id: S.activeTrackLesson?.id,
  };
}

function loadTrackLesson(force = false) {
  if (S.streaming || !S.activeTrackLesson) return;
  const key = trackCacheKey(S.activeTrackId, S.activeTrackLesson.id);

  resetLessonPanel();
  $('lesson-empty').classList.add('hidden');
  const out = $('lesson-output');
  out.innerHTML = '';
  out.classList.remove('hidden');
  out.classList.add('streaming');
  $('lesson-footer').classList.add('hidden');

  streamFetch('/api/track/lesson', { ...trackBody(), force },
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

function loadTrackChallenge() {
  if (S.streaming || !S.activeTrackLesson) return;
  const key = trackCacheKey(S.activeTrackId, S.activeTrackLesson.id);

  $('challenge-empty').classList.add('hidden');
  $('challenge-inner').classList.remove('hidden');
  closeEval();
  const out = $('challenge-output');
  out.innerHTML = '';
  out.classList.add('streaming');

  streamFetch('/api/track/challenge', trackBody(),
    (_, acc) => { out.innerHTML = parseMarkdown(acc); },
    (full) => {
      out.classList.remove('streaming');
      S.challengeRaw[key] = full;
      out.innerHTML = parseMarkdown(full);
      applyHighlight(out);
      S.challenges[key] = out.innerHTML;
      const editor = $('code-editor');
      if (!editor.value.trim()) {
        const match = full.match(/```[\w]*\s*([\s\S]*?)```/g);
        if (match) {
          const last = match[match.length - 1];
          editor.value = last.replace(/```[\w]*\s*/, '').replace(/```$/, '').trim();
          autoResize(editor);
        }
      }
    }
  );
}

function submitTrackCode() {
  if (S.streaming || !S.activeTrackLesson) return;
  const code = $('code-editor').value.trim();
  const key  = trackCacheKey(S.activeTrackId, S.activeTrackLesson.id);
  if (!code)               { showToast('Write some code first!', 'error'); return; }
  if (!S.challengeRaw[key]) { showToast('Load a challenge first!', 'error'); return; }

  const panel = $('eval-panel');
  const out   = $('eval-output');
  panel.classList.remove('hidden');
  out.innerHTML = '';
  out.classList.add('streaming');

  streamFetch('/api/track/evaluate', { ...trackBody(), code, challenge: S.challengeRaw[key] },
    (_, acc) => { out.innerHTML = parseMarkdown(acc); },
    (full) => {
      out.classList.remove('streaming');
      out.innerHTML = parseMarkdown(full);
      applyHighlight(out);
      const lower = full.toLowerCase();
      if (full.includes('✅') && (lower.includes('pass') || lower.includes('correct') || lower.includes('well done'))) {
        showToast('✅ Challenge passed!', 'success');
      }
    }
  );
}

function getTrackHint() {
  if (S.streaming || !S.activeTrackLesson) return;
  const key = trackCacheKey(S.activeTrackId, S.activeTrackLesson.id);
  if (!S.challengeRaw[key]) { showToast('Load a challenge first!', 'error'); return; }

  const panel = $('eval-panel');
  const out   = $('eval-output');
  panel.classList.remove('hidden');
  out.innerHTML = '<p><strong>💡 Hint</strong></p>';
  out.classList.add('streaming');

  streamFetch('/api/track/hint', { ...trackBody(), challenge: S.challengeRaw[key], code: $('code-editor').value },
    (_, acc) => { out.innerHTML = '<p><strong>💡 Hint</strong></p>' + parseMarkdown(acc); },
    (full)   => { out.classList.remove('streaming'); out.innerHTML = '<p><strong>💡 Hint</strong></p>' + parseMarkdown(full); }
  );
}

async function sendTrackChat() {
  if (S.streaming || !S.activeTrackLesson) return;
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
  const assistantAvatar = el('div', 'chat-avatar', S.langs.find(l => l.id === S.lang)?.icon || '🎓');
  const assistantBubble = el('div', 'chat-bubble streaming');
  assistantWrap.appendChild(assistantAvatar);
  assistantWrap.appendChild(assistantBubble);
  box.appendChild(assistantWrap);
  box.scrollTop = box.scrollHeight;

  streamFetch('/api/track/chat', { ...trackBody(), messages: history },
    (_, acc) => { assistantBubble.innerHTML = parseMarkdown(acc); box.scrollTop = box.scrollHeight; },
    (full)   => {
      assistantBubble.classList.remove('streaming');
      assistantBubble.innerHTML = parseMarkdown(full);
      applyHighlight(assistantBubble);
      history.push({ role: 'assistant', content: full });
      box.scrollTop = box.scrollHeight;
    }
  );
}

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

// ── Post-auth init ─────────────────────────────
async function postAuthInit() {
  $('username-display').textContent = S.user;

  const defaultLang = S.langs[0]?.id || 'go';
  await switchLang(defaultLang, false); // loads topics + tracks

  // Start on first incomplete topic; reset=true triggers cache-aware lesson logic
  const firstIncomplete = S.topics.find(t => !t.completed);
  selectTopic(firstIncomplete ? firstIncomplete.id : 1, true);

  switchTab('lesson');
  setupEditorTab();
}

// ── Init ───────────────────────────────────────
async function init() {
  await loadLanguages();

  const me = await checkAuth();
  if (!me) {
    showLoginModal();
    return; // submitAuth() calls postAuthInit() on success
  }
  S.user = me.username;
  await postAuthInit();
}

init();
