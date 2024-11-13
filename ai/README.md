# 🤖 amsh

Ape Machine Shell, it's ehm, complicated...

## Agents Speak in Code

```boogie
; An example pipeline.
; in wraps the input, out is the output after execution of the pipeline.
; <= and => show the direction of the flow.
; <...> is a behavior, it sits between the process and the input, and shapes the output.
; Think of the | operator as error handling, it defines the order of prefered potential outcomes.
; Think of cancel as canceling the context.
out <= (
    ; The switch dictates how the internal process is directed.
    ; switch: from one step to the next.
    ; select: choose freely between any of the internal steps, repeat as needed
    switch <= (
        clean    => next | cancel
        validate => next | cancel
        enrich   => send | cancel
    ) <= switch[step2] <= (                              ; labels are for jumping.
        analyze<temporal>  => next | cancel              ; use analysis, with temporal behavior, or, perform temporal analysis.
        model              => next | back | cancel
        optimize           => send | back | cancel
    ) <= join <= (                                       ; join the two concurrent processes into an output.
        out <= select <= (
            reason               <= <5> => next | cancel ; iteration, maximum of 5 recursions.
            [plan <= step2.analyze.out] => next | cancel ; the second switch has been labeled, so we can refer to it.
        )                                                ; concurrent processing.
                                                         ; every closure is concurrent, you notice it when there are multiple under the same parent.
        out <= switch <= (
            format => next | back | cancel
            save   => send
        )                                                ; concurrent processing.
    ) <= match <= (
        success => send
        default => [step2.analyze.jump] ; nested labels chain together.
    ) <= switch[mylabel] <= (
        clean    => next | cancel
        validate => next | cancel
        enrich   => next | cancel
        
        out <= match <= (
            <5>     => send             ; on the 5th iteration, we send the output.
            default => [mylabel.jump]   ; otherwise, we jump back to the beginning of the switch.
        )
    )
) <= in                                 ; entrypoint, upon entry we jump to the top of the closure.
```
