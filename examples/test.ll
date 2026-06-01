package app.main

Window[main]:
    title = "LealLang App"
    w = 800
    h = 600

    ui func handle_save():
        @Label[title].text = "Saved!"
        console.print_ln("Save button clicked")
    
    ui func handle_resize():
        console.print_ln($"Window resized to: {@Window[main].w} {@Window[main].h}")

    func handle_hover():
        console.print_ln("Button hovered")

    Col[root]:
        gap = 12

        Label[title]:
            text = "Hello, LealLang!"

        Button[save_btn]:
            text = "Save"
            on click = handle_save
            on hover = handle_hover

func main():
    window.open(@Window[main])
    @Window[main].handle_save()
    @Window[main].handle_resize()
