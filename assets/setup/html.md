# Getting Started with HTML

Welcome! Before the first lesson, it's worth knowing **where to actually see
your HTML render**. The editor built into this bootcamp doesn't display your
pages — your AI instructor reads the markup and gives feedback on it. Seeing
your pages in a real browser is where the learning sticks, and for HTML that
takes about thirty seconds to set up.

HTML needs no compiler, no runtime, no install: **a text editor and a browser
are the whole toolchain.**

## See a page right now — no install

- **[MDN Playground](https://developer.mozilla.org/en-US/play)** — a clean
  HTML/CSS/JS scratchpad from the people who write the web's documentation.
  Paste in a lesson example and see it render instantly.
- **[codepen.io](https://codepen.io)** — the classic front-end playground.
  Type in the HTML panel, see the result live below.

## The local workflow — a file and a browser

1. Create a file called `index.html` anywhere on your computer, containing:

   ```html
   <!DOCTYPE html>
   <html lang="en">
   <head>
     <meta charset="UTF-8">
     <title>My First Page</title>
   </head>
   <body>
     <h1>Hello from my own machine!</h1>
   </body>
   </html>
   ```

2. Open it in your browser — double-click the file, or drag it onto a browser
   window. That's what `$ open index.html` under the sidebar logo means.
3. Edit the file, save, and **refresh the browser** to see the change. That
   edit–save–refresh loop is the whole workflow.

One more essential habit: press **F12** (or `Cmd+Option+I` on a Mac) to open
the browser's *DevTools* and click the *Elements* tab. It shows the live
structure of any page — including how the browser understood (or corrected)
your markup.

## Set up a code editor

### VS Code

1. Install it from **[code.visualstudio.com](https://code.visualstudio.com)**.
2. HTML support is built in, including **Emmet**: in an empty `.html` file,
   type `!` and press Tab — a full page skeleton appears.
3. Optional but lovely: the **Live Server** extension. Click "Go Live" in the
   status bar and the browser refreshes itself every time you save — no more
   manual reloading.

### Zed

1. Install it from **[zed.dev](https://zed.dev)** — a newer, very fast editor.
2. Install the **HTML** extension from the extensions list (`zed: extensions`
   in the command palette) for full completion and formatting support.

Either is a great choice: VS Code has more tutorials and answers written about
it (and Live Server is genuinely handy for HTML work); Zed is lighter and
faster.

## Where to look things up

- **[MDN's HTML reference](https://developer.mozilla.org/en-US/docs/Web/HTML/Element)**
  — every element that exists, what it's for, and its attributes. Searching
  "mdn table" or "mdn form" gets the definitive answer. Trust MDN over random
  blog posts.
- **[validator.w3.org](https://validator.w3.org)** — paste in your HTML and
  it lists every structural mistake. Browsers silently forgive broken markup,
  so the validator is how you find out what you actually wrote.
- **[web.dev/learn/html](https://web.dev/learn/html)** — Google's free HTML
  course; a good second angle on any topic here.

## Suggested path

1. **Today:** create `index.html`, open it in your browser, and keep DevTools'
   Elements tab open while you read the first lessons.
2. **Soon:** pick an editor. Try Emmet's `!` + Tab trick once — you'll never
   type a doctype by hand again.
3. **Habit:** run anything you build through the validator before submitting
   it. Catching your own mistakes beats being told about them.

Stuck on any step — a file that opens as text instead of a page, an element
that won't render? Open the **💬 Ask** tab and describe what you're seeing.
