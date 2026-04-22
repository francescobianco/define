# define

> **Predict what your tests will find — before they run.**

---

## The idea

Tests are essential. They will always be the source of truth for runtime correctness.  
But tests are *reactive* by nature: they tell you something is broken only after executing.

`define` is *proactive*: it reads the structure of your codebase and tells you, in advance, what the tests are going to find.

> If `define` says a dependency is missing, the tests will say it too.  
> You just know it 2 seconds into the pipeline instead of 10 minutes in.

This is not about skipping tests. It is about **moving the answer earlier** — from the end of the pipeline to the beginning.

---

## What `define` anticipates

`define` extracts a **concept graph** from your codebase — packages, classes, modules — and checks its structural integrity:

| Signal | What it predicts |
|--------|-----------------|
| **Closure** | "these tests will fail: a dependency is undefined" |
| **Reachability** | "this module will never be exercised from the entry point" |
| **Dead concepts** | "these symbols are declared but nothing reaches them" |
| **Cycle detection** | "this circular dependency will cause a runtime or import error" |

When `define` flags something, it is giving you the test result early — derived from structure alone, with no execution cost.

---

## Anticipate. Don't wait.

```
[ step 1 ]  define extract ./src --out model.def    # ~2 seconds
[ step 2 ]  define check model.def                  # ~1 second

→ closed, low impact   → run only the tests likely affected
→ closed, high impact  → run the full suite, you know why
→ NOT closed           → the tests will fail — fix the design first
```

The pipeline does not skip tests. It knows *before the tests run* whether they will pass structurally, and it acts on that knowledge.

### The numbers

A typical mid-size project spends **8–15 minutes per CI run** in test execution.  
`define` adds **3 seconds** of structural anticipation. In exchange:

- **Broken-import failures** are predicted before any compilation starts
- **Dead modules** are surfaced before they accumulate in the codebase
- **Change impact** is visible — a small graph change can be confirmed safe without running the full suite

In a team running 50 CI builds/day, anticipating even 2 broken builds per day saves **25+ minutes of compute** — and the context-switch cost of a developer waiting for a result they could have known in seconds.

---

## Installation

```bash
git clone https://github.com/francescobianco/define
cd define
make install        # installs to ~/.local/bin/define
```

**Requirements:** Go 1.22+

---

## Quick start

### 1. Write a `.def` model by hand

```
define Example1 with Example2, Routes

define Example2 with load, test

define load
define test

define Routes with /routes/a

define /routes/a {
    load;
    Example2 test;
}
```

```bash
define model.def
```

```
Example1 is closed

Coverage:
  - declared:  6
  - reachable: 6
  - invoked:   3

No dead concepts found.
```

### 2. Extract from a real codebase (Go)

```bash
define extract ./src --out model.def
define check model.def
```

With a custom profile (external to the codebase):

```bash
define extract ./src --config profiles/go-generic.yml --out model.def
define check model.def MyApp/EntryPoint
```

---

## The `.def` language

| Syntax | Meaning |
|--------|---------|
| `define X` | declare concept X |
| `define X with A, B` | X depends on A and B |
| `define X { Y; Z; }` | X's body invokes Y and Z |
| `# comment` | line comment |

Concept names can be identifiers (`MyService`) or paths (`/routes/api`).

---

## The library pattern — a real example

`define` reveals something tests cannot: **the difference between what a library declares and what it consumes itself**.

We ran `define` on **PHPUnit** — one of the most widely used PHP testing frameworks:

### Without tests (`src/` only)

```
declared:   991 classes
reachable:  362 from TextUI/Application (the CLI entry point)
```

629 classes are unreachable from the entry point. They look dead.  
But they are not dead — they are the **public API** that consumers use.

The tool also detects real structural issues:
```
cycle detected: [Framework/Constraint/Constraint → Framework/Assert → Framework/Constraint/Constraint]
cycle detected: [Event/Facade → Runner/DeprecationCollector/Facade → Event/Facade]
```

These are real circular dependencies in PHPUnit's `src/` — visible in 2 seconds of static analysis.

### With tests (`src/` + `tests/`)

```
declared:  2426 classes (1435 test classes added)
reachable:  401 from TextUI/Application
```

When the test suite is included, 39 previously "unreachable" library classes become reachable — they are consumed by the tests.  
30 classes that appeared "never referenced" in the library are now clearly used.

**The insight:** a library's public API is only meaningful when analyzed *together with its consumers*.  
`define` makes this visible in a single command.

---

## CI integration

```yaml
# .github/workflows/ci.yml
jobs:
  preflight:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - name: Install define
        run: |
          git clone https://github.com/francescobianco/define /tmp/define
          cd /tmp/define && make install
      - name: Structural check
        run: |
          define extract ./src --out model.def
          define check model.def          # exits 1 if not closed → blocks next jobs

  test:
    needs: preflight        # only runs if structure is valid
    runs-on: ubuntu-latest
    steps:
      - run: go test ./...
```

With `needs: preflight`, test jobs never start if the design is broken. This alone saves compute on every broken-architecture push.

---

## Commands

| Command | Description |
|---------|-------------|
| `define <file.def>` | verify model from first declared symbol |
| `define <file.def> <Symbol>` | verify from a specific root symbol |
| `define check <file.def> [Symbol]` | explicit check subcommand |
| `define extract <dir>` | extract model from Go source (auto-detects language) |
| `define extract <dir> --config <profile.yml>` | extract with explicit projection profile |
| `define extract <dir> --check` | extract and verify in one step |

---

## Projection profiles (`define.yml`)

The profile lives **outside** the codebase being analyzed. You can experiment freely without touching upstream code.

```yaml
# labs/profiles/go-generic.yml
language: go

sources:
  - "**/*.go"

ignore:
  - "**/*_test.go"
  - vendor/**

concepts:
  package:
    # each Go package directory = one concept

relations:
  with:
    from: imports

  invoke:
    from: function_calls   # v0.2
```

### Supported languages

| Language | Status |
|----------|--------|
| Go | stable |
| PHP | beta (class-level extraction) |
| TypeScript | planned (v0.2) |

---

## Labs

The `labs/` directory contains a test bench for running `define` against real open-source projects:

```bash
make labs-fetch          # clone repos listed in labs/repos.txt
make labs-analyze        # extract + check all repos
make labs-phpunit-demo   # compare PHPUnit with and without test suite
```

Profiles for each codebase live in `labs/profiles/` — separate from the cloned code.

---

## What `define` is not

- Not a linter — it does not check style or formatting
- Not a test runner — it does not execute code
- Not a type checker — it does not verify types or signatures
- Not a replacement for tests — it is a *complement*, run before tests

---

## Roadmap

| Version | Focus |
|---------|-------|
| v0.1 | Go extraction, closure/reachability/dead-concept detection (done) |
| v0.2 | Body invocation tracking, TypeScript extractor, impact scoring |
| v0.3 | CI impact analysis: "how many tests should this change trigger?" |
| v1.0 | GitHub Action, design coverage reports, drift detection |
