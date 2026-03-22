# Agent Info

Generally speaking, you should browse the codebase to figure out what is going on.

We have a few "philosophies" I want to make sure we honor throughout development:

### 1. Performance above all else

When in doubt, do the thing that makes the app feel the fastest to use.

This includes things like

- Optimistic updates
- Avoiding waterfalls in anything from js to file fetching

### 2. Good defaults

Users should expect things to behave well by default. Less config is best.

### 3. Security

We want to make things convenient, but we don't want to be insecure. Be thoughtful about how things are implemented.

## Project Overview
CLI AI assistant (GoCo - Go Conventional) that generates Conventional Commit messages using Google Gemini AI. Built with Cobra for CLI framework and Charm Bracelet libraries for CLI components.

## Code Style
- **Go version**: 1.24.6+

## Project Structure
