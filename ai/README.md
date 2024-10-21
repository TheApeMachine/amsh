# AI VM

We are starting a new experimental approach where we lightly model the AI system as a virtual machine, including a primitive coding language.

Let's start with some conceptual sketches of what a language could look like.

```
; Input is processed through a series of sequential workers.
; Each worker either sends its output to the next worker, sends the output to a previous worker, cancels the process, or sends the final output.
out <= (
    switch <= (
        reason  => next || cancel
        plan    => next || back || cancel
        execute => next || back || cancel
        verify  => back || continue
        report  => next || back * 2 || cancel
        answer  => send
    )
) <= in
```

```
; Define a pipeline that processes input data and handles errors.
; Error handling is done by means of the || and since everything is logged, we already have built-in traces.
; We can just store every message, every generation inside an S3 bucket.
; With this mindset, we can think of the | operator as error handling, as it defines the order of prefered potential outcomes.
; Think of cancel as canceling the context as you would do in Go.
out <= (
    ; The arrow direction shows that this "closure" outputs the state of its internal processing.
    ; Having pre-processing and finalize seems like an arbitrary division, or I can not see the benefits.
    ; The switch in this case dictates how the internal process is directed.
    ; switch: from one step to the next.
    ; select: choose freely between any of the internal steps, repeat as needed
    ; working only with single word statements forces us to keep the language minimalistic, which is better
    ; for the user, and for the LLMs.
    switch <= (
        clean    => next | cancel
        validate => next | cancel
        enrich   => send | cancel
    ) <= switch[step2] <= (
        analyze  => next | cancel | timeout
        model    => next | back | cancel
        optimize => send | back | cancel
    ) <= join <= (
        ; concurrent processing.
        out <= select <= (
            reason<5>                   => next | cancel        ; reason is limited to 5 iterations, or selects, after that it is consumed.
            [plan <= step2.analyze.out] => next | back | cancel ; the second switch has been labeled, so we can refer to it.
        )

        ; concurrent processing.
        out <= switch <= (
            format => next | back | cancel
            save   => send
        )
    ) <= match <= (
        success => send
        default => [step2.analyze.jump]
    ) <= switch[mylabel]<5> <= (
        clean    => next | cancel
        validate => next | cancel
        enrich   => next | cancel
        
        out <= match <= (
            done    => send           ; on the 5th iteration, we send the output.
            default => [mylabel.jump] ; otherwise, we jump back to the beginning of the switch.
        )
    )
) <= in
```

## Artifact

A Cap'n Proto type has been defined called `Artifact` that acts as a wrapper around any type of
payload data, and adds metadata to it.

Artifact should be seen as the ultimate primitive type inside this system, and thus everything takes
an `Artifact` as input and returns an `Artifact` as output.

```
struct Artifact {
  id @0 :Text;
  checksum @1 :Data;
  pubkey @2 :Data;
  version @3 :Text;
  type @4 :Text;
  timestamp @5 :UInt64;
  origin @6 :Text;
  role @7 :Text;
  scope @8 :Text;
  attributes @9 :List(Attribute);
  payload @10 :Data;
}

struct Attribute {
  key @0 :Text;
  value @1 :Text;
}
```

This metadata is used to derive behavior, but notably also defines the prefix within the S3 bucket where the artifact is stored.

```
{
    "id":         "<uuid>",
    "version":    "v0.0.1",
    "type":       "application/json",
    "origin":     "<worker-id>",
    "role":       "<worker-role>",
    "scope":      "<worker-scope>",
    "timestamp":  "<unix-timestamp>",
    "attributes": [],
    "payload":    "<[]byte>"
}

PREFIX: <origin>/<role>/<scope>/<timestamp>/<type>/<version>/<id>.json
```

## Virtual Machine Structure

```
Pool (CPU)
    Worker (Process)
        Job (Executable)
            Artifact<metadata> (Program)
Registers (Memory)
    Artifact<payload> (Local state)
    VectorStore<local> (Long term private memory)
    VectorStore<global> (Shared memory)
    GraphDatabase<local> (Long term relational memory)
    GraphDatabase<global> (Shared relational memory)
Queue (IO)
    Inputs
        UI User Prompt
        Webhook
        Event
    Outputs
        Internal Messaging
        UI Response
```
