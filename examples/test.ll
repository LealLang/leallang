package app.main

func greet(name: string) -> bool:
    console.print_ln($"Hello, {name}")
    return name == "test"

myString: string = "test"
isTest: bool = greet(myString)

func main():
    console.print_ln(isTest)
