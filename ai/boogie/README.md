# Boogie 🎯

> Boogie is a domain-specific language designed for LLM agents to interact with systems in a highly flexible and dynamic way. It combines structured programming with autonomous decision-making, enabling agents to design their own workflows and processes on-the-fly.

## Formal Language Definition 📝

### Grammar Rules

```ebnf
program     ::= context '<=' closure '<=' context
closure     ::= '(' statement* ')'
statement   ::= operation flow_chain comment?
flow_chain  ::= '=>' target ('|' target)*
operation   ::= identifier behavior? params?
behavior    ::= '<' identifier '>'
params      ::= '{' value (',' value)* '}'
target      ::= identifier | label | closure
label       ::= '[' identifier ']'
comment     ::= ';' [^\n]*
```

## Core Concepts 🧠

### Dual-Agent Architecture

Boogie operates with two types of agents:

- **Programmer Agents**: Write Boogie code to design workflows
- **Worker Agents**: Execute individual instructions, each as a short-lived agent that mutates context

The key benefit is simplification of schema handling - instead of requiring agents to understand massive JSON schemas, each worker agent only needs to understand the schema for its specific behavior.

### Context and Flow

Context is the fundamental concept in Boogie:

- Always implicitly present as `in`
- Flows through operations that mutate it
- Follows consistent flow patterns:
  - Between closures: Bottom-to-top, right-to-left
  - Within closures: Top-to-bottom, left-to-right

## Language Constructs 🔨

### Basic Structure

The simplest valid Boogie program:

```boogie
out <= () <= in ; Empty closure, no mutation
```

### Closures `()`

Closures are self-contained blocks that process context:

```boogie
out <= (
    analyze => next ; Analysis with error handling
    verify  => send ; Verification with multiple paths
) <= in
```

### Flow Operations

Context can flow to:

- `next`: Continue to next operation
- `send`: Send context upward
- `back`: Return to previous operation
- `cancel`: Terminate the flow

Basic flow example:

```boogie
out <= (
    analyze => next ; Analysis without any behavior
    verify  => send ; Verification, and send up the chain
) <= in
```

### Fallback Chaining `|`

Provides error handling and retry logic:

```boogie
out <= (
    analyze => next | back | cancel ; Analysis with error handling
    verify  => send | back | cancel ; Verification, and send up the chain
) <= in
```

With retry count:

```boogie
out <= (
    analyze => next | back<3> | cancel ; Analysis with error handling
    verify  => send | back<3> | cancel ; Verification, and send up the chain
) <= in
```

### Match Blocks

Conditional flow control based on context state:

```boogie
out <= (
    analyze => next ; Analysis without any behavior
    verify  => next ; Verification, and send up the chain
    match (
        ok    => send   ; If ok, send up the chain
        error => cancel ; If error, cancel
        _     => back   ; In any other case, go back to verify
    )
) <= in
```

### Labels `[]`

Reference points for context state or flow:

```boogie
out <= (
    analyze[myLabel] => next ; Analysis without any behavior
    verify           => next ; Verification, and send up the chain
    match <= [myLabel] (
        ok    => send   ; If ok, send up the chain
        error => cancel ; If error, cancel
    )
) <= in
```

With jump targets:

```boogie
out <= (
    [myLabel] => (
        analyze => next ; Analysis without any behavior
        verify  => next ; Verification, and send up the chain
        match (
            ok    => send           ; If ok, send up the chain
            error => cancel         ; If error, cancel
            _     => [myLabel].jump ; In any other case, jump to the beginning
        )
    )
) <= in
```

### Iteration `<= =>`

Repeating operations:

```boogie
out <= (
    analyze <= => next | back | cancel ; Analysis with error handling and (infinite) iteration
    verify     => send | back | cancel ; Verification, and send up the chain
) <= in
```

With iteration limit:

```boogie
out <= (
    analyze <= <3> => next ; Analysis with 3 iterations maximum
    verify         => send ; Verification, and send up the chain
) <= in
```

### Concurrency

Parallel execution with join:

```boogie
out <= (
    join <= (
        analyze => next ; Analysis without any behavior
        verify  => send ; Verification, and send up the chain
    ) (
        analyze => next ; Analysis without any behavior
        verify  => send ; Verification, and send up the chain
    )
) <= in
```

### System Integration

Tool operations with parameters:

```boogie
out <= (
    call<{
        search, 
        "some query"
    } => browser> => send ; Browser search operation
) <= in
```

### Behaviors `<>`

Operation modifiers:

```boogie
out <= (
    analyze<surface> => send
) <= in
```

## Available Behaviors 🛠️

These are not hard-coded into the language and always subject to change, so they do not have a true representation as part of the language. However, they are documented here for reference.

### 🔍 Analysis Behaviors

| Behavior        | Description               |
|-----------------|---------------------------|
| `<surface>`     | Surface-level analysis    |
| `<temporal>`    | Temporal analysis         |
| `<pattern>`     | Pattern-matching analysis |
| `<quantum>`     | Quantum-layer analysis    |
| `<fractal>`     | Fractal analysis          |
| `<holographic>` | Holographic analysis      |
| `<tensor>`      | Tensor analysis           |
| `<narrative>`   | Narrative analysis        |
| `<analogy>`     | Analogy analysis          |
| `<practical>`   | Practical analysis        |
| `<contextual>`  | Contextual analysis       |

### 🤔 Reasoning Behaviors

| Behavior             | Description                  |
|----------------------|------------------------------|
| `<chainofthought>`   | Chain-of-thought reasoning   |
| `<treeofthought>`    | Tree-of-thought reasoning    |
| `<selfcritique>`     | Self-critique approach       |
| `<selfassessment>`   | Self-assessment approach     |
| `<roleplay>`         | Roleplay-based reasoning     |
| `<metacognition>`    | Metacognitive reasoning      |
| `<hypothesis>`       | Hypothesis-based reasoning   |
| `<validation>`       | Validation-focused reasoning |
| `<devideandconquer>` | Divide and conquer strategy  |
| `<analogical>`       | Analogical reasoning         |
| `<probabilistic>`    | Probabilistic reasoning      |
| `<deductive>`        | Deductive reasoning          |
| `<inductive>`        | Inductive reasoning          |
| `<abductive>`        | Abductive reasoning          |

### 💡 Generation Behaviors

| Behavior     | Description                    |
|--------------|--------------------------------|
| `<moonshot>` | Innovative, ambitious ideas    |
| `<sensible>` | Practical, grounded ideas      |
| `<catalyst>` | Transformative ideas           |
| `<guardian>` | Protective, conservative ideas |
| `<metrics>`  | Metric generation              |
| `<code>`     | Code generation                |
| `<plan>`     | Plan generation                |

### 🔌 System Integration Behaviors

| Behavior        | Description                  |
|-----------------|------------------------------|
| `<browser>`     | Chrome browser operations    |
| `<github>`      | GitHub integration           |
| `<environment>` | Linux environment operations |
| `<memory>`      | Vector/graph store access    |
| `<helpdesk>`    | Helpdesk integration         |
| `<slack>`       | Slack communication          |
| `<wiki>`        | Wiki information access      |
| `<boards>`      | Project management           |
| `<recruit>`     | Team formation               |

## Complex Examples 🌟

### Multi-Stage Analysis

```boogie
out <= (
    call<{search, "topic"} => browser> => next                 ; Initial research
    analyze<pattern>                   => next | back | cancel ; Pattern analysis
    analyze<temporal>                  => next | back | cancel ; Timeline analysis
    match (
        ok     => send   ; Send if analysis complete
        error  => cancel ; Cancel on errors
        _      => back   ; Otherwise retry analysis
    )
) <= in
```

### Iterative Refinement

```boogie
out <= (
    [refine] => (
        analyze<surface>   => next ; Initial analysis
        analyze<practical> => next ; Practical check
        verify<validation> => next ; Validation
        match (
            ok    => send          ; Accept if valid
            error => [refine].jump ; Jump back if needs work
        )
    )
) <= in
```

### Parallel Processing

```boogie
out <= (
    match (
        complete => send   ; Send combined results
        error    => cancel ; Cancel on error
        _        => back   ; Retry if incomplete
    ) <= join <= (
        analyze<pattern> => next ; Pattern analysis path
        verify           => send
    ) (
        call<{query} => wiki> => next ; Wiki lookup path
        analyze<practical>    => send
    )
) <= in
```
