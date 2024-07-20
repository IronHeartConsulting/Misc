//  clock  with text styling
//
//   updated to use canvas text
//   can se ttet attributes - color, size, etc

package main

import (
	"fyne.io/fyne/v2/app"
	// "fyne.io/fyne/v2/widget"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2"
	"image/color"

	"time"
)


func updateTime(clockT *canvas.Text, clockD *canvas.Text) {

	formattedT := time.Now().Format("Time: 03:04:05")
	// formattedD := time.Now().Format("Date: Mon Jan 2 MST 2006")
	clockT.Text = formattedT
	clockT.Refresh()
	clockD.Text = time.Now().Format("Date: Mon Jan 2 MST 2006")
	clockD.Refresh()
	// clock.SetText(formatted)
}

func main()  {

	green := color.NRGBA{R: 0, G: 180, B: 0, A: 255}
	red := color.NRGBA{R: 180, G: 0, B: 0, A: 255}

	a := app.New()
	w := a.NewWindow("Clock container") //widget
	// c := w.Canvas()
	clockT := canvas.NewText("", red)
	clockT.TextStyle.Bold = true

	formatted := time.Now().Format("Time: 03:04:05")
	clockT.Text = formatted
	clockT.TextSize = 100.0
	clockT.Alignment = fyne.TextAlignCenter

	clockD := canvas.NewText("", green)
	clockD.Text = time.Now().Format("Date: Mon Jan 2 MST 2006")
	clockD.Alignment = fyne.TextAlignCenter
	clockD.TextSize = 50.0
	// clockD.Move(fyne.NewPos(20,20))
	ctnr := container.New(layout.NewVBoxLayout(), clockT, clockD)

//  
//   clock is a widget (special canvas) set to Label
	// clock := widget.NewLabelWithStyle("", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

// set window content to the canvas.text 
	w.SetContent(ctnr)
	//c.SetContent(clockc)

	go func() {
		for range time.Tick(time.Second) {
			updateTime(clockT, clockD)
		}
	} ()

	w.ShowAndRun()

}
