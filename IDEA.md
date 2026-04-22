Perfetto, ti lascio una sintesi unica, pulita e pronta da incollare nel repo come base di lavoro.

---

## `define` — vision & scope (v0.1)

`define` è una CLI che estrae, modella e verifica la struttura di una codebase come un grafo di concetti, dipendenze e invocazioni.

Il tool verifica:

* **closure** (tutto è risolvibile)
* **reachability** (cosa è davvero usato dal root)
* **dead concepts** (cosa è dichiarato ma inutile)
* **coerenza strutturale**

e produce segnali utilizzabili come **quality gate prima della CI**.

---

## 🎯 Obiettivo

Non sostituire i test, ma:

> **ridurre spreco in CI e migliorare coerenza architetturale prima dell’esecuzione dei test**

`define` agisce come **preflight check strutturale**.

---

## 🧠 Modello

### Concetti base

* `define X` → simbolo
* `with A, B` → dipendenze richieste
* `{ ... }` → corpo (invocazioni)
* `X` → invocazione

### Esempio

```txt
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

---

## ✅ Closure

Un simbolo è **closed** se:

* è definito
* tutte le dipendenze (`with`) sono chiuse
* tutte le invocazioni esistono
* non ci sono cicli
* l’espansione termina

---

## ☠️ Dead concepts

Classificazioni:

* **dead by reference** → mai usato in `with`
* **dead by invocation** → mai invocato
* **unreachable** → non raggiungibile dal root

---

## 🔍 Output (CLI)

```bash
define Example1
```

```txt
Example1 is closed

Coverage:
- declared: 12
- reachable: 9
- invoked: 6

Dead concepts:
- never referenced:
  Debug
- never invoked:
  test
- unreachable:
  /routes/old
```

---

## ⚙️ Pipeline

### 1. Extraction (da codice reale)

```bash
define extract ./src --config define.yml --out model.def
```

### 2. Verification

```bash
define model.def
```

---

## 🧩 `define.yml` (projection layer)

Serve a **mappare codice → modello**.

Non è logica, ma configurazione di estrazione.

Esempio:

```yaml
language: typescript

sources:
  - src/**/*.ts

concepts:
  service:
    suffix: "Service"

  repository:
    suffix: "Repository"

relations:
  with:
    from: constructor_injection

  invoke:
    from: method_calls

ignore:
  - "**/*.spec.ts"
```

👉 `define.yml` = lente con cui osservi la codebase

---

## 🚀 Posizionamento

`define` è:

> **a structural CI gate for design closure, reachability and dead concept detection**

Non è:

* un linguaggio generale
* un sostituto dei test

È:

* un **verificatore strutturale**
* un **analyzer di impatto**
* un **effort-save gate**

---

## 💡 Valore pratico

Prima dei test, `define` può:

* bloccare build incoerenti
* segnalare design rotto
* identificare codice morto
* stimare impatto di una modifica
* suggerire livello di test necessario

---

## 🧪 Uso in CI

```txt
[ step 1 ] define extract
[ step 2 ] define check

if not closed → FAIL FAST
if low impact → reduce test scope
if high impact → full suite
```

---

## 🔧 MVP (v0.1)

Supporta:

* define / with / body
* invocazione simboli
* detection cicli
* closure
* reachability
* dead concepts

Non include ancora:

* semantica runtime
* test coverage reale
* analisi profonda linguaggi host

---

## 🧭 Visione

`define` evolve verso:

* design coverage
* impact analysis
* CI optimization
* drift detection tra codice e architettura

---

## 🧠 Sintesi finale

> `define` trasforma una codebase in un modello verificabile e usa quel modello per bloccare incoerenze, evidenziare sprechi e guidare decisioni intelligenti in CI.

---

Se vuoi, il prossimo passo è scrivere:

* CLI reale (`define`, `extract`, `check`)
* formato `.def` parser
* primo report funzionante in Go

Questo è già abbastanza per una demo forte.
