# 🤖 amsh

Ape Machine Shell, it's ehm, complicated...

## Agents Speak in Code

```boogie
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
    ) <= switch[mylabel] <= (
        clean    => next | cancel
        validate => next | cancel
        enrich   => next | cancel
        
        out <= match <= (
            <5>     => send           ; on the 5th iteration, we send the output.
            default => [mylabel.jump] ; otherwise, we jump back to the beginning of the switch.
        )
    )
) <= in
```
