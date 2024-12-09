package main

import (
	"errors"
	"image"
	"image/png"
	"net/url"
	"strconv"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"

	"github.com/skip2/go-qrcode"
)

type qrInfo struct {
	content  string
	extra    string
	extra2   string
	mode     int
	security string
	hidden   bool
}

func main() {
	a := app.NewWithID("xyz.andy.fyqr")
	w := a.NewWindow("FyQR")

	types := widget.NewSelect([]string{"URL/Text", "Phone", "Email", "SMS", "WhatsApp", "Wifi", "FaceTime"}, nil)
	types.SetSelectedIndex(0)

	in := widget.NewEntry()

	in2 := widget.NewEntry()
	in2.Hide()

	in3 := widget.NewEntry()
	in3.Hide()

	wifiEnc := widget.NewSelect([]string{"WPA", "WEP", "None"}, nil)
	wifiEnc.SetSelectedIndex(0)
	wifiEnc.Hide()

	in2.Validator = func(s string) error {
		//Only check for the password if the network isn't open
		if wifiEnc.SelectedIndex() != 2 && s != "" {
			return nil
		}
		return errors.New("wifi password required")
	}

	WifiHidden := widget.NewCheck("Hidden", nil)
	WifiHidden.Hide()

	in.SetPlaceHolder("https://example.com")
	in.Validator = func(s string) error {
		//Initialize with nil value for error
		var err error
		err = nil

		//No text
		if s == "" {
			return errors.New("cannot be empty")
		}

		switch types.SelectedIndex() {
		case 1, 3, 4:
			//Phone, SMS & Whatsapp: only the phone number needed
			// Can we convert the string to int?
			_, err = strconv.Atoi(s)
		case 2:
			//For email, we only need the address
			// Does the string contain an @?
			if !strings.Contains(s, "@") {
				err = errors.New("invalid email")
			}
			//FaceTime only needs the recipient
		}
		return err
	}

	types.OnChanged = func(_ string) {
		i := types.SelectedIndex()
		switch i {
		case 0, 1, 6:
			placeholder := "https://example.com"
			if i == 1 {
				placeholder = "1234567890"
			}
			if i == 6 {
				placeholder = "phone or email"
			}
			in.SetPlaceHolder(placeholder)
			in2.Hide()
			in3.Hide()
			wifiEnc.Hide()
			WifiHidden.Hide()
		case 2:
			in.SetPlaceHolder("mail@example.com")
			in2.Show()
			in2.SetPlaceHolder("Subject")
			in3.Show()
			in3.SetPlaceHolder("Body")
			wifiEnc.Hide()
			WifiHidden.Hide()
		case 3, 4:
			in.SetPlaceHolder("1234567890")
			in2.Show()
			in2.SetPlaceHolder("Message")
			in3.Hide()
			wifiEnc.Hide()
			WifiHidden.Hide()
		case 5:
			in.SetPlaceHolder("SSID")
			in2.Show()
			in2.SetPlaceHolder("Password")
			in3.Hide()
			wifiEnc.Show()
			WifiHidden.Show()
		}
		if i != 5 {
			//Reset the wifi encryption to WPA, else in2.Validator will throw an error :)
			wifiEnc.SetSelectedIndex(0)
		}
	}

	out := &canvas.Image{}
	out.ScaleMode = canvas.ImageScalePixels
	out.FillMode = canvas.ImageFillContain
	out.SetMinSize(fyne.NewSquareSize(256))

	save := widget.NewButtonWithIcon("", theme.DocumentSaveIcon(), func() {
		write(out.Image, w)
	})
	save.Disable()
	save.Importance = widget.SuccessImportance

	run := widget.NewButtonWithIcon("", theme.NavigateNextIcon(), func() {
		err := in.Validate()
		if err != nil {
			dialog.ShowError(err, w)

			out.Image = nil
			out.Refresh()
			save.Disable()

			return
		}
		img, err := gen(qrInfo{content: in.Text, extra: in2.Text, extra2: in3.Text, mode: types.SelectedIndex(), security: wifiEnc.Selected, hidden: WifiHidden.Checked})
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
	run.Importance = widget.HighImportance
	in.OnSubmitted = func(_ string) {
		run.OnTapped()
	}

	top := container.NewBorder(nil, nil, nil, container.NewHBox(run, save), container.NewVBox(in, types, in2, WifiHidden, in3, wifiEnc))
	ui := container.NewBorder(top, nil, nil, nil, out)

	w.SetContent(ui)
	w.ShowAndRun()
}

func gen(data qrInfo) (image.Image, error) {
	var content string
	switch data.mode {
	case 0:
		content = data.content
	case 1:
		content = "tel:" + data.content
	case 2:
		content = "mailto:" + data.content + "?subject=" + url.QueryEscape(data.extra) + "&body=" + url.QueryEscape(data.extra2)
	case 3:
		content = "smsto:" + data.content + ":" + data.extra
	case 4:
		content = "https://wa.me/" + data.content + "?text=" + url.QueryEscape(data.extra)
	case 5:
		content = "WIFI:S:" + data.content + ";T:" + data.security + ";P:" + data.extra + ";H:" + strconv.FormatBool(data.hidden) + ";"
	case 6:
		content = "facetime:" + data.content
	}
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
