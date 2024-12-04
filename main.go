package main

import (
	"image"
	"image/png"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/skip2/go-qrcode"
)

func main() {
	a := app.NewWithID("xyz.andy.fyqr")
	w := a.NewWindow("FyQR")

	in := widget.NewEntry()
	in.SetPlaceHolder("https://example.com")

	out := &canvas.Image{}
	out.ScaleMode = canvas.ImageScalePixels
	out.FillMode = canvas.ImageFillContain
	out.SetMinSize(fyne.NewSquareSize(256))

	save := widget.NewButtonWithIcon("", theme.DocumentSaveIcon(), func() {
		write(out.Image, w)
	})
	save.Disable()

	run := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
		img, err := gen(in.Text)
		if err != nil {
			dialog.ShowError(err, w)

			out.Image = nil
			out.Refresh()
			save.Disable()

			return
		}

		out.Image = img
		out.Refresh()
		save.Enable()
	})
	in.OnSubmitted = func(_ string) {
		run.OnTapped()
	}

	top := container.NewBorder(nil, nil, nil, container.NewHBox(run, save), in)
	ui := container.NewBorder(top, nil, nil, nil, out)

	w.SetContent(ui)
	w.ShowAndRun()
}

func gen(content string) (image.Image, error) {
	q, err := qrcode.New(content, qrcode.Medium)
	if err != nil {
		return nil, err
	}

	return q.Image(512), nil
}

func write(img image.Image, win fyne.Window) {
	dialog.ShowFileSave(func(w fyne.URIWriteCloser, err error) {
		if w == nil {
			return
		}
		if err != nil {
			dialog.ShowError(err, win)
			return
		}

		err = png.Encode(w, img)
		if err != nil {
			dialog.ShowError(err, win)
			return
		}
		_ = w.Close()
	}, win)
}
