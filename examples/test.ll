package app.main

func handle_save():
    @Label[title].text = "Saved!"
    console.print_ln("Save button clicked")

ui func view():
    Window[main]:
        title = "LealLang App"
        w = 800
        h = 600

        Col[root]:
            gap = 12

            Label[title]:
                text = "Hello, LealLang!"

            Button[save_btn]:
                text = "Save"
                on click = handle_save

func main():
    console.print_ln("Starting LealLang UI...")
    view()
    console.print_ln("UI mounted successfully")
    handle_save()
