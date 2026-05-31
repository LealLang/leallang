package app.main

import app.util as util

const MAX: int = 100

async func compute(text: string):
    loop i in 0..MAX, step 2, while i < 50:
        console.print_ln(text + $" i is {i}")

func main():
    loop i in 0..MAX, step 2, while i < 50:
        console.print_ln($"normal i is {i}")

    t1 = compute("first")
    t2 = compute("second")
    await t1
    await t2
