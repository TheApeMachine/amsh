# AMSH Language Specification
Version 1.0

## 1. Overview
AMSH (Ape Machine Shell) is a pipeline-oriented language designed for orchestrating multi-agent AI systems. It emphasizes declarative flow control, error handling, and concurrent processing capabilities.

## 2. Lexical Structure

### 2.1 Tokens
```ebnf
token ::= operator | identifier | number | delimiter

operator ::= '<=' | '=>' | '|'

identifier ::= [a-zA-Z][a-zA-Z0-9_]*

number ::= [0-9]+

delimiter ::= '(' | ')' | '[' | ']' | '<' | '>'
```

### 2.2 Keywords
```
switch    - Sequential flow control
select    - Non-deterministic choice
join      - Parallel execution
match     - Pattern matching
next      - Continue to next step
back      - Return to previous step
cancel    - Terminate current context
timeout   - Time-based termination
default   - Default case in pattern matching
in        - Input marker
out       - Output marker
jump      - Control flow transfer
```

## 3. Syntax

### 3.1 Program Structure
```ebnf
program ::= pipeline

pipeline ::= out_marker '<=' block '<=' in_marker

block ::= '(' statement+ ')'

out_marker ::= 'out'
in_marker ::= 'in'
```

### 3.2 Control Structures

#### Switch Statement
```ebnf
switch_stmt ::= 'switch' ['[' label ']'] '<=' '(' step+ ')'

step ::= identifier '=>' outcome ('|' outcome)*

outcome ::= 'next' | 'back' | 'cancel' | 'timeout' | 'send' | jump_expr
```

#### Select Statement
```ebnf
select_stmt ::= 'select' '<=' '(' selection+ ')'

selection ::= identifier ['<' number '>'] '=>' outcome ('|' outcome)*
```

#### Join Statement
```ebnf
join_stmt ::= 'join' '<=' '(' concurrent_block+ ')'

concurrent_block ::= 'out' '<=' (switch_stmt | select_stmt)
```

#### Match Statement
```ebnf
match_stmt ::= 'match' '<=' '(' match_case+ ')'

match_case ::= ('success' | 'default' | '<' number '>') '=>' (outcome | jump_expr)
```

### 3.3 References and Labels
```ebnf
label ::= identifier

jump_expr ::= '[' (label '.')? 'jump' ']'

block_ref ::= '[' label '.' identifier '.' 'out' ']'
```

## 4. Operational Semantics

### 4.1 Pipeline Execution
- Pipelines execute from right to left (`in` to `out`)
- Each block processes its input and produces output for the next block
- The `<=` operator indicates data flow direction

### 4.2 Error Handling
- The `|` operator defines fallback paths
- Execution follows the leftmost successful path
- `cancel` terminates the current context and may trigger cleanup
- `timeout` indicates time-based termination

### 4.3 Concurrency
- `join` blocks execute their internal pipelines concurrently
- All concurrent pipelines must complete before proceeding
- Resources are shared between concurrent pipelines

### 4.4 State Management
- Each block maintains its own state
- State can be referenced using labeled blocks
- State transitions are logged automatically

## 5. Built-in Behaviors

### 5.1 Iteration Control
```
<number>        - Limits iteration count
reason<5>       - Limits reasoning steps to 5 iterations
```

### 5.2 Flow Control
```
next            - Proceed to next step
back            - Return to previous step
cancel          - Terminate current context
send            - Output results
jump            - Transfer control to labeled location
```

## 6. Logging and Tracing

### 6.1 Automatic Logging
- All messages and state transitions are logged
- Logs are stored in specified storage (e.g., S3)
- Each operation generates a trace ID

### 6.2 Trace Format
```json
{
  "trace_id": "string",
  "timestamp": "datetime",
  "operation": "string",
  "block_label": "string?",
  "input_state": "any",
  "output_state": "any",
  "error": "string?"
}
```

## 7. Example Patterns

### 7.1 Basic Pipeline
```
out <= (
    clean    => next | cancel
    validate => next | cancel
    enrich   => send | cancel
) <= in
```

### 7.2 Iterative Reasoning
```
out <= select <= (
    reason<5> => next | cancel
    analyze   => next | back | cancel
) <= in
```

### 7.3 Parallel Processing with Fallback
```
out <= join <= (
    out <= switch <= (
        process1 => next | cancel
        process2 => send | back
    )
    out <= switch <= (
        process3 => next | cancel
        process4 => send | back
    )
) <= match <= (
    success => send
    default => cancel
) <= in
```

## 8. Type System

### 8.1 Basic Types
```
Block     - Pipeline processing unit
State     - Internal state object
Label     - Block identifier
Outcome   - Processing result
```

### 8.2 Type Rules
1. Labels must be unique within their scope
2. Referenced labels must exist
3. Outcome paths must be valid for the block type
4. Iteration limits must be positive integers

## 9. Best Practices

### 9.1 Structure
1. Keep pipelines focused on single responsibilities
2. Use meaningful labels for blocks
3. Limit nesting depth for clarity
4. Prefer `select` over `switch` for non-sequential operations

### 9.2 Error Handling
1. Always provide fallback paths
2. Use `cancel` judiciously
3. Handle timeouts explicitly
4. Log error conditions with context

### 9.3 State Management
1. Minimize shared state
2. Use labeled blocks for clear state references
3. Validate state transitions
4. Clean up resources on cancellation

## 10. Implementation Considerations

### 10.1 Runtime Requirements
1. Concurrent execution capability
2. State isolation between blocks
3. Reliable logging infrastructure
4. Error recovery mechanisms

### 10.2 Performance
1. Minimize state copying
2. Optimize concurrent execution
3. Efficient logging implementation
4. Smart resource allocation

## 11. Security Considerations

### 11.1 Execution Context
1. Isolate pipeline executions
2. Validate input data
3. Control resource usage
4. Monitor execution time

### 11.2 Data Handling
1. Secure state storage
2. Encrypted logging
3. Access control
4. Audit trail

## 12. Agent-Specific Features

### 12.1 Agent Declaration
```ebnf
agent_decl ::= 'agent' identifier '<=' '(' capability+ ')'

capability ::= identifier ':' (primitive_cap | compound_cap)

primitive_cap ::= 'reason' | 'perceive' | 'act' | 'communicate'

compound_cap ::= '{' capability_expr '}'

capability_expr ::= capability_term (('&' | '|') capability_term)*

capability_term ::= identifier | '(' capability_expr ')'
```

### 12.2 Agent Interaction Patterns

#### Message Passing
```
out <= switch <= (
    send_message[agent1] => next | retry | cancel
    await_response<timeout=300> => next | cancel
    process_response => send | back
) <= in
```

#### Collaborative Reasoning
```
out <= join <= (
    out <= agent1.reason<depth=3> <= (
        analyze => next | cancel
        synthesize => send
    )
    out <= agent2.validate <= (
        check => next | back
        approve => send | reject
    )
) <= match <= (
    success => send
    reject => [retry.jump]
    default => cancel
) <= in
```

### 12.3 Agent State Management

#### State Visibility
```ebnf
visibility ::= 'private' | 'shared' | 'public'

state_decl ::= visibility 'state' identifier '<=' expression
```

#### State Transitions
```
state_transition ::= current_state '=>' '[' condition ']' '=>' next_state

condition ::= boolean_expr | 'timeout' | 'error' | 'success'
```

### 12.4 Agent Capabilities

#### Reasoning Patterns
- Depth-limited exploration: `reason<depth=N>`
- Time-bounded analysis: `analyze<timeout=T>`
- Confidence-based decisions: `decide<confidence=0.9>`

#### Learning Integration
```
out <= learn <= (
    observe => next | cancel
    hypothesis => next | back
    validate => (
        confidence>0.8 => store | back
        default => [hypothesis.jump]
    )
) <= in
```

## 13. Implementation Patterns

### 13.1 Common Agent Architectures

#### Belief-Desire-Intention (BDI)
```
out <= agent[bdi] <= (
    update_beliefs <= (
        perceive => next | cancel
        integrate => next | back
        validate => send | back
    )
    select_intention <= (
        filter_desires => next | cancel
        prioritize => next | back
        commit => send | reconsider
    )
    execute_plan <= (
        decompose => next | replan
        act => next | retry
        monitor => send | [update_beliefs.jump]
    )
) <= in
```

#### Hierarchical Task Network (HTN)
```
out <= agent[htn] <= (
    task_decomposition <= (
        analyze => next | cancel
        decompose => next | backtrack
        validate => send | backtrack
    )
    primitive_execution <= (
        sequence => next | replan
        parallel => next | serialize
        monitor => send | [decompose.jump]
    )
) <= in
```

### 13.2 Multi-Agent Coordination

#### Contract Net Protocol
```
out <= contract_net <= (
    broadcast_task => next | retry | cancel
    collect_bids<timeout=200> => next | cancel
    evaluate_bids => (
        found => award | rebroadcast
        default => cancel
    )
    monitor_execution <= (
        track => next | intervene
        complete => send | [evaluate_bids.jump]
    )
) <= in
```

#### Blackboard Architecture
```
out <= blackboard <= (
    register_agents => next | cancel
    monitor_knowledge <= (
        update => next | clean
        trigger => next | wait
        notify => send | retry
    )
    control_agents <= (
        select => next | wait
        activate => next | cancel
        synchronize => send | [monitor_knowledge.jump]
    )
) <= in
```

## 14. Advanced Features

### 14.1 Meta-Programming
- Runtime pipeline modification
- Dynamic agent creation
- Capability composition

### 14.2 Formal Verification
- Deadlock detection
- Liveness properties
- Safety constraints

### 14.3 Performance Optimization
- Pipeline parallelization
- State caching
- Message batching

### 14.4 Debugging and Monitoring
- Step-by-step execution
- State inspection
- Performance profiling
