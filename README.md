# Botgoram - State-based telegram bot framework, in Go

Botgoram is state-based telegram bot framework written in go. It is inspired by [tucnak/telebot](https://github.com/tucnak/telebot).

[![GoDoc](https://godoc.org/github.com/Patrolavia/botgoram?status.svg)](https://godoc.org/github.com/Patrolavia/botgoram)

Current it is still under development, not usable for now.

### State based

We think the work flow for bot is like a [Finite State Machine](https://en.wikipedia.org/wiki/Finite-state_machine): given current state, transit to next state acording to the input. We write code to choose right state, and to define what to to when entering/ leaving a state.

## Synopsis

See [example code on godoc.org](https://godoc.org/github.com/Patrolavia/botgoram#example-package).

## But how can I convert my business logic to a state machine

It depends. Draw a flowchart, especially a data flowchart, and treat each unit as a state might be a reasonable start. The "state pattern", "Automata-based programming" on wikipedia might also give you some thought.

## It looks so complicate!

[Yes, the code will be much longer.](https://en.wikipedia.org/wiki/Automata-based_programming#Automata-based_style_program) But it will also eliminates a number of control structures and function calls. And program can be faster if you apply certain optimization on your state map.

## License

Any version of MIT, GPL or LGPL.
