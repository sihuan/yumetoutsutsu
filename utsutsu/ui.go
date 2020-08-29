package main

import (
	"fyne.io/fyne/app"
	"fyne.io/fyne/widget"
)

type Ui struct {
	parent *Utsutsu
}

func NewApp(parent *Utsutsu)  {
	u := new(Ui)
	u.parent = parent
	a := app.New()

	w := a.NewWindow("Hello")
	w.SetContent(widget.NewVBox(
		widget.NewLabel("Hello Fyne!"),
		widget.NewButton("Quit", func() {
			a.Quit()
		}),
	))

	w.ShowAndRun()
}

//func (u *Ui) Start() {
//	u.app
//}