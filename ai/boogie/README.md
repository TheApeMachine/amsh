# Boogie 🎯

> Boogie is a domain-specific language designed for LLM agents to interact with systems in a highly flexible and dynamic way. It combines structured programming with autonomous decision-making, enabling agents to design their own workflows and processes on-the-fly.

---

## Core Concepts 🧠

To understand Boogie, you have to understand that it is meant to be both programmed by agents, as well as executed by agents.

This means there are two types of agents needed:

- **Programmer Agents**: Agents that write Boogie code. This provides a highly flexible, yet still structured schema.
- **Worker Agents**: Agents that execute Boogie code. Each instruction is executed by a short-lived worker agent that mutates the context.

> The benefit we gain from this is the lack of need for massive jsonschema definitions. Even though most LLM models have the context window to support
> that, the models struggle with understanding, or effective use of such schemas.
> By having a simple language, we can map instructions to single jsonschema definitions, which we can give to a worker agent.

---

### Context

The context starts with some input, either from a user request, for instance in a chat setup, or it could come from a webhook, etc.

Boogie instructions always operate on the context, it is what goes in, and what comes out, in most cases mutating along the way.

---

### Context Flow

- **Outer Scope**: Context flows bottom-to-top, right-to-left between closures
- **Inner Scope**: Context flows top-to-bottom, left-to-right within closures
- **Implicit Context**: The current context (`in`) is always present and implicit
- **Worker Agents**: Each operation is executed by a short-lived worker agent that mutates the context

---

## Language Constructs 🔨

### Comments `;`

Comments start with a semicolon `;` and extend to the end of the line.

```boogie
; This is a comment
; This makes it a multi-line comment
out <= () <= in ; This is a comment on the same line as code
```

### Closure `()`

The fundamental building block of Boogie. Closures are self-contained code blocks that process and mutate the current context.

> A Boogie program always has at least one closure, the outermost one.

```boogie
out <= () <= in ; The simplest possible valid Boogie program, which sends in to out, without mutation
```

> The value of `in` is always the current context, and `in` is always present, even if it is not explicitly mentioned.
> The way to visualize the program below would be:
> in -> analyze -{in}-> verify -{in}-> out

```boogie
out <= (
    analyze => next ; Analysis with error handling
    verify  => send ; Verification with multiple paths
) <= in
```

### Flow `<=` & `=>`

The flow operators `<=` and `=>` are used to define the flow of context between closures and operations.

> In a Boogie program, context between closures flows bottom-to-top, right-to-left, and within closures, it flows top-to-bottom, left-to-right.

```boogie
out <= (
    analyze => next ; Analysis without any behavior
    verify  => send ; Verification, and send up the chain
) <= in
```

> The above example shows the most simple method to flow the context through the program, where analyze sends the mutated context to verify, no matter what, and verify sends the context to out, no matter what.
> A more realistic example makes use of Boogie's fallback chaining feature.

```boogie
out <= (
    analyze => next | back | cancel ; Analysis with error handling
    verify  => send | back | cancel ; Verification, and send up the chain
) <= in
```

> Above is likely the most common version of using fallback chaining. The way to read this is:
> analyze tries to send the context to verify, unless there is an problem, in which case the context is sent back to analyze once, or if it was already sent back once, the operation is canceled.
> verify tries to send the context to out, unless there is an problem, in which case the context is sent back to verify once, or if it was already sent back once, the operation is canceled.

### Iteration `<= =>`

```boogie
out <= (
    analyze <= => next | back | cancel ; Analysis with error handling, and infinite iteration
    verify     => send | back | cancel ; Verification, and send up the chain
) <= in
```

> The above example shows the use of iteration, which is a special case of flow.
> The `<= =>` operator is used to define an iteration, which will repeat the operation until the stop condition is met.
> In the case of this example, which is technically infinite iterations, the worker agent would signal stop by saying it is done.
> In most cases you would add a behavior to this to limit the amount of iterations.

### Fallback Chaining `|`

Fallback chaining is used to define a sequence of operations that are tried in order, until one of them succeeds.

```boogie
out <= (
    analyze => next | back<3> | cancel ; Analysis with error handling
    verify  => send | back<3> | cancel ; Verification, and send up the chain
) <= in
```

> The above example adds a `behavior` to the back statement, which essentially says, try 3 times, before canceling.

### Conditionals `match`

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

Label are primarily used for referencing. The most common use-cases would be referencing data from another closure, or referencing a jump target.

```boogie
out <= (
    analyze[myLabel] => next ; Analysis without any behavior
    verify           => next ; Verification, and send up the chain
    match <= [myLabel] (
        ok    => send           ; If ok, send up the chain
        error => cancel         ; If error, cancel
    )
) <= in
```

> The example above references the state as it was after analyze, in the match statement.

```boogie
out <= (
    [myLabel] => (
        analyze => next ; Analysis without any behavior
        verify  => next ; Verification, and send up the chain
        match (
            ok    => send           ; If ok, send up the chain
            error => cancel         ; If error, cancel
            _     => [myLabel].jump ; In any other case, jump to the beginning of the closure
        )
    )
) <= in
```

> The example above shows how to perform a jump in a Boogie program.

---

## Concurrency 🔄

In the Boogie language, each closure is always running concurrently, next to other closures in the same parent scope.

You will only notice this if there are multiple closures in the same parent scope, otherwise there is only one concurrent closure being executed.

Operations inside a closure always run sequenctially.

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

> The above simple example of concurrency will run the two closures concurrently, and provide a join channel which streams the results of both closures up the chain.

### Parameters `[,]`

Especially used when controlling system integrations or tools.

```boogie
out <= (
    call<{
        search, 
        "some query"
    } => browser> => send ; Use a web browser to perform a search
) <= in
```

> In the above example we see how to pass paramters to a call operation.
> Essentially this says: perform the call operation, behaving like a browser instruction, using the paramters `search` and `"some query"`, and send the result up the chain.

---

### Behavior `<>`

A behavior is a modifier that alters the way an operation or statement "behaves".

```boogie
out <= (analyze<surface> => send) <= in
```

> A very simple example which would guide the worker agent executing the analyze operation using a `surface` (level) schema definition.

```boogie
out <= (
    analyze <= <3> => next ; Analysis without any behavior, but with an iteration instruction that behaves such that it limits iterations to 3
    verify         => next ; Verification, and send up the chain
) <= in
```

## Available Behaviors 🛠️

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
