# Wordle Solver

This program aids in solving Wordles.

## Dependencies

This program currently requires [ripgrep](https://github.com/BurntSushi/ripgrep), `rg`, to be installed.

## How To Use

Execute the program.  After entering your guess on the website, enter your guess in the command line program.

- Green letters should be prefixed with a `+`.
- Yellow letters should be prefixed with a `-`.

`Ctrl-c` to quit the program.

### Example

Say you enter the word `saute`.  Wordle shows `a` and `e` as green, and `t` as
grey.  In this program, input: `s+au-t+e`.  It will then spit out the list of
possible words for your next guess.
