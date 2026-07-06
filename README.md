# Multi Language Coding Bootcamp

A personalised coding bootcamp that teaches to code in various languages.
Currently it covers:
- Backend languages Go and Zig
- Frontend. HTML, CSS and Javascript

It uses AI to write a lessons.
Based on the content in the lesson, it will write challenges.
There are 4 levels of challenges. Basic, Intermediate, Advanced & GOAT
It will then provide feedback on the response to those challenges.
There is also a chatbot, where you can Ask it questions, where it has the context of the lesson and challenges provided when giving answers.

It requires an API key to run, and has been setup to work with both Anthropic and Deepseek models.
Store the API key is either `claude.key` or `deepseek.key`

If those files have a key, the user will automatically be presented with various options.

Each language covers the following:

**Fundamentals**
Covers the fundamentals skills to lean in each language

**Advanced Tracks**
Provised advanced training on various subjects, where each `Track` has multiple lessons and challenges

**Projects**
Based on the skills taught in previous sections, build a specific complete project step by step.

## Hardcoded and Dynamic content

The course structure is hard coded.
The lesson/challenge content is generated on the fly by whatever AI model has been chosen.
This content is cached, so it will survive a server restart.

# Security
This is not suitable for public access. While there is a user logon screen, there are no limits on who can sign up and how much they can use the system, so someone could use up an unlimited amount of API data.

This is really only designed to run locally where you can control or trust who has access and how much they will use it.

<img width="1245" height="1327" alt="Screenshot 2026-07-06 at 2 17 05 pm" src="https://github.com/user-attachments/assets/215888ca-6d01-4361-8c30-8d3b928f8845" />
