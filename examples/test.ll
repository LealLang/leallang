package app.main

func greet(name: string) -> bool:
    console.print_ln($"Hello, {name}")
    return name == "test"

myString: string = "0"
isTest: bool = greet(myString)

func main():
    console.print_ln(isTest)
