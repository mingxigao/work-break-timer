package main

import (
	"errors"
	"fmt"
	"log"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/data/binding"
	"fyne.io/fyne/v2/widget"
)

type Pref struct {
	workMinutes              int
	breakMinutes             int
	forceWindowFocusDuration int
}

func Load() Pref {
	app := fyne.CurrentApp()
	workMinutes := app.Preferences().IntWithFallback("workMinutes", 25)
	breakMinutes := app.Preferences().IntWithFallback("breakMinutes", 5)
	forceWindowFocusDuration := app.Preferences().IntWithFallback("forceWindowFocusDuration", 60)

	return Pref{
		workMinutes,
		breakMinutes,
		forceWindowFocusDuration,
	}
}

func Save(pref Pref) {
	app := fyne.CurrentApp()
	app.Preferences().SetInt("workMinutes", pref.workMinutes)
	app.Preferences().SetInt("breakMinutes", pref.breakMinutes)
	app.Preferences().SetInt("forceWindowFocusDuration", pref.forceWindowFocusDuration)
}

type Settings interface {
	Show()
	SetOnSubmit(callback func())
	SetOnClose(callback func())
}

// type validation
var _ Settings = (*settings)(nil)

type settings struct {
	win  *fyne.Window
	form *widget.Form
}

func NewSettings() Settings {
	win := fyne.CurrentApp().NewWindow("Settings")
	f := makeForm()
	f.OnCancel = func() {
		win.Close()
	}
	// Need to "refresh" to make the Submit and Cancel buttons appears
	f.Refresh()

	win.SetContent(f)
	return &settings{win: &win, form: f}
}

func (s *settings) Show() {
	(*s.win).Show()
}

func (s *settings) SetOnSubmit(callback func()) {
	formSubmit := s.form.OnSubmit
	s.form.OnSubmit = func() {
		formSubmit()
		(*s.win).Close()
		callback()
	}
	(*s.form).Refresh()
}

func (s *settings) SetOnClose(callback func()) {
	(*s.win).SetOnClosed(callback)
}

func makeForm() *widget.Form {
	myPref := Load()
	form := widget.NewForm()

	workMinutesBinding := binding.NewInt()
	if err := workMinutesBinding.Set(myPref.workMinutes); err != nil {
		log.Println("Error setting workMinutesBinding: ", err)
	}
	form.AppendItem(newIntegerFormItem(workMinutesBinding, "Work duration in minutes", "Default is: %d minutes.  ", NewRangeValidator(0, 999)))

	breakMinutesBinding := binding.NewInt()
	if err := breakMinutesBinding.Set(myPref.breakMinutes); err != nil {
		log.Println("Error setting breakMinutesBinding: ", err)
	}
	form.AppendItem(newIntegerFormItem(breakMinutesBinding, "Break duration in minutes", "Default is: %d minutes.  ", NewRangeValidator(0, 999)))

	forceWindowFocusDurationBinding := binding.NewInt()
	if err := forceWindowFocusDurationBinding.Set(myPref.forceWindowFocusDuration); err != nil {
		log.Println("Error setting forceWindowFocusDurationBinding: ", err)
	}
	form.AppendItem(newIntegerFormItem(forceWindowFocusDurationBinding, "Force Window Focus in seconds", "Default is: %d seconds.  ", NewRangeValidator(0, 999)))

	label := widget.NewLabel("")
	formItem := widget.NewFormItem("Left Time Duration in seconds", label)
	label.Bind(leftTimeDurationBinding)
	form.AppendItem(formItem)

	form.OnSubmit = func() {
		disableAllTimers()
		workMinutes, _ := workMinutesBinding.Get()
		breakMinutes, _ := breakMinutesBinding.Get()
		forceWindowFocusDuration, _ := forceWindowFocusDurationBinding.Get()
		newPref := Pref{
			workMinutes,
			breakMinutes,
			forceWindowFocusDuration,
		}
		Save(newPref)
		time.Sleep(time.Second)
		loadSetting()
		go startWorkTimer()
	}

	return form
}

func newIntegerFormItem(bind binding.Int, entryText string, hintText string, validator fyne.StringValidator) *widget.FormItem {
	value, err := bind.Get()
	if err != nil {
		log.Println("Error getting value: ", err)
	}
	entryData := binding.IntToString(bind)
	entry := newIntegerEntryWithData(entryData)
	// it should be a fyne bug
	go func() {
		time.Sleep(100 * time.Millisecond)
		entry.Validator = validator
	}()
	formItem := widget.NewFormItem(entryText, entry)
	formItem.HintText = fmt.Sprintf(hintText, value)
	return formItem
}

type integerEntry struct {
	widget.Entry
}

func newIntegerEntryWithData(data binding.String) *integerEntry {
	entry := &integerEntry{}
	entry.ExtendBaseWidget(entry)
	entry.Bind(data)
	return entry
}

func (e *integerEntry) TypedRune(r rune) {
	switch r {
	case '0', '1', '2', '3', '4', '5', '6', '7', '8', '9':
		e.Entry.TypedRune(r)
	}
}

func (e *integerEntry) TypedShortcut(shortcut fyne.Shortcut) {
	paste, ok := shortcut.(*fyne.ShortcutPaste)
	if !ok {
		e.Entry.TypedShortcut(shortcut)
		return
	}

	content := paste.Clipboard.Content()
	if _, err := strconv.ParseInt(content, 10, 64); err == nil {
		e.Entry.TypedShortcut(shortcut)
	}
}

func NewRangeValidator(min int64, max int64) fyne.StringValidator {
	return func(text string) error {
		v, err := strconv.ParseInt(text, 10, 64)
		if err != nil {
			return errors.New("not a valid number")
		}
		if v < min {
			return fmt.Errorf("must be greater than %d", min)
		}
		if v > max {
			return fmt.Errorf("must be lesser than %d", max)
		}

		return nil // Nothing to validate with, same as having no validator.
	}
}
