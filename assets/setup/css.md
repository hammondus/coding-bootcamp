# Getting Started with CSS

Welcome! Before the first lesson, it's worth knowing **where to actually see
your CSS work**. The editor built into this bootcamp doesn't render styles —
your AI instructor reads them and gives feedback. Watching your styles change
a real page is where the learning sticks.

Like HTML, CSS needs no install: **a text editor and a browser are the whole
toolchain.** (CSS always styles *something*, so these lessons assume a little
basic HTML — the HTML track's first topics cover everything you need.)

## See CSS working right now — no install

- **[codepen.io](https://codepen.io)** — the classic front-end playground, and
  it's *made* for CSS experiments: HTML in one panel, CSS in another, live
  result below. Most CSS tricks you'll find online are shared as CodePens.
- **[MDN Playground](https://developer.mozilla.org/en-US/play)** — a cleaner,
  simpler scratchpad from the people who write the web's documentation.

## The secret weapon: DevTools

Press **F12** (or `Cmd+Option+I` on a Mac), click the *Elements* tab, and
select any element. The *Styles* pane shows every CSS rule affecting it — and
you can **edit the values live**: change a color, tick properties on and off,
nudge padding with the arrow keys, and watch the page update instantly.

This works on *any* website, and it's the single best way to learn CSS: when a
lesson mentions `flex` or `margin`, open a real page and play with it. (Your
edits vanish on refresh — you're changing your browser's copy, not the site.)

## The local workflow

1. Make a folder with two files. `index.html`:

   ```html
   <!DOCTYPE html>
   <html lang="en">
   <head>
     <meta charset="UTF-8">
     <title>CSS Practice</title>
     <link rel="stylesheet" href="style.css">
   </head>
   <body>
     <h1 class="title">Hello, CSS!</h1>
   </body>
   </html>
   ```

   and `style.css`:

   ```css
   .title {
     color: steelblue;
     font-family: sans-serif;
   }
   ```

2. Open `index.html` in your browser (double-click it — that's the
   `$ open index.html` under the sidebar logo).
3. Edit the CSS, save, refresh. That loop is the whole workflow.

## Set up a code editor

### VS Code

1. Install it from **[code.visualstudio.com](https://code.visualstudio.com)**.
2. CSS support is built in: autocomplete for properties and values, color
   swatches next to every color, and a color picker when you hover them.
3. Optional: the **Live Server** extension auto-refreshes the browser on every
   save — especially pleasant for CSS tinkering.

### Zed

1. Install it from **[zed.dev](https://zed.dev)** — a newer, very fast editor.
2. CSS completion works out of the box. Open a `.css` file and start typing.

## Where to look things up

- **[MDN's CSS reference](https://developer.mozilla.org/en-US/docs/Web/CSS)**
  — every property, every value, with live examples. Searching "mdn flexbox"
  or "mdn border-radius" gets the definitive answer. Trust MDN over random
  blog posts.
- **[CSS-Tricks' flexbox guide](https://css-tricks.com/snippets/css/a-guide-to-flexbox/)**
  and **[grid guide](https://css-tricks.com/snippets/css/complete-guide-grid/)**
  — the two most-bookmarked pages in CSS, for good reason. Keep them open
  during the Flexbox and Grid topics.
- **[caniuse.com](https://caniuse.com)** — tells you which browsers support a
  feature. Everything this course teaches is safely supported, but you'll want
  this the moment you explore beyond it.
- **[web.dev/learn/css](https://web.dev/learn/css)** — Google's free CSS
  course; a good second angle on any topic here.

## Suggested path

1. **Today:** open DevTools on a site you like and change its styles live.
   Then do your first lesson with CodePen open beside it.
2. **Soon:** set up the two-file folder above and pick an editor.
3. **Habit:** when a property confuses you, look it up on MDN and poke at it
   in DevTools before asking — you'll often answer your own question, which is
   the best way to remember it.

Stuck on any step — a stylesheet that isn't applying, a selector that won't
match? Open the **💬 Ask** tab and describe what you're seeing.
