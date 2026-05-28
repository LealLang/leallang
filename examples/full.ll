package app.main

import app.util as util

const MAX: int = 100

type Point:
    X: int
    Y: int
    constructor(x: int, y: int):
        X = x
        Y = y
    func Greet() -> string:
        return $"Point({X}, {Y})"

func main():
    p = Point(x: 10, y: 20)
    
    result: string = switch p.X:
        10 -> "ten"
        _  -> "other"

    loop i in 0..MAX, step 2, while i < 50:
        console.print_ln($"i is {i}")
