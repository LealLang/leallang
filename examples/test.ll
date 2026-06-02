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
            on_click = handle_save
            on_hover = handle_hover

func main():
    window.open(@Window[main])
    @Window[main].handle_save()
    @Window[main].handle_resize()

    my_list: List<int> = [1, 2, 3, 4, 5]
    loop num in 0..<my_list.count():
        console.print_ln($"Number: {my_list[num]}")
