# Getting Started with JavaScript

Welcome! Before the first lesson, it's worth knowing **where to actually run
JavaScript**. The editor built into this bootcamp doesn't execute your code —
your AI instructor reads it and gives feedback on the logic. Running your code
for real is where the learning sticks.

The good news: you already have a JavaScript engine installed. Every browser
ships one.

## Run JavaScript right now — no install

- **Your browser's console** — press **F12** (or `Cmd+Option+J` on a Mac,
  `Ctrl+Shift+J` on Windows/Linux) and click the *Console* tab. Type or paste
  JavaScript and press Enter; `console.log(...)` output appears right there.
  For one-liners and lesson snippets, nothing beats it.
- **[MDN Playground](https://developer.mozilla.org/en-US/play)** — a clean
  HTML/CSS/JS scratchpad from the people who write the web's documentation.
- **[codepen.io](https://codepen.io)** — the classic front-end playground.
  You'll want it anyway when the course reaches **The DOM** and **Events**,
  where your JavaScript needs an HTML page to work on.

## Install Node.js — JavaScript outside the browser

Node runs JavaScript files from your terminal, no browser needed. That's what
the `$ node app.js` under the sidebar logo means.

1. Download the **LTS** version from **[nodejs.org](https://nodejs.org)** and
   run the installer.
2. Check it worked:

   ```
   node --version
   ```

3. Create `app.js`:

   ```js
   const name = 'my own machine';
   console.log(`Hello from ${name}!`);
   ```

   And run it:

   ```
   node app.js
   ```

Note: browser-only features like `document` and `alert` don't exist in Node —
for the DOM topics you'll be back in the browser. Everything before those
(variables, loops, functions, arrays, objects…) runs perfectly in either.

## Set up a code editor

### VS Code

1. Install it from **[code.visualstudio.com](https://code.visualstudio.com)**.
2. That's it — JavaScript support (autocomplete, hover docs, error squiggles)
   is built in, no extension needed.
3. Optional, for the DOM/Events topics: the **Live Server** extension serves an
   HTML file and auto-reloads the browser every time you save.

### Zed

1. Install it from **[zed.dev](https://zed.dev)** — a newer, very fast editor.
2. JavaScript support is built in here too. Open a `.js` file and start typing.

Either is a great choice: VS Code has more tutorials and answers written about
it; Zed is lighter and faster. Pick one — don't agonise.

## Where to look things up

- **[MDN](https://developer.mozilla.org/en-US/docs/Web/JavaScript)** — *the*
  JavaScript reference, maintained by Mozilla. Every method with examples:
  searching "mdn array filter" or "mdn string slice" gets you exactly what
  those methods do. Trust MDN over random blog posts.
- **[javascript.info](https://javascript.info)** — a free, modern, in-depth
  tutorial. Good for a second, longer explanation of anything a lesson covers.
- **[nodejs.org/docs](https://nodejs.org/docs/)** — reference for Node's own
  APIs (files, servers…), useful once you're past the fundamentals.

## Suggested path

1. **Today:** open the browser console and run every example the lessons show
   you.
2. **Soon:** install Node so you can run `.js` files, and pick an editor.
3. **At The DOM topic:** switch to CodePen or Live Server, where your code has
   a page to manipulate.

Stuck on any step — an installer that won't run, a `command not found`, output
you didn't expect? Open the **💬 Ask** tab and describe what you're seeing.
