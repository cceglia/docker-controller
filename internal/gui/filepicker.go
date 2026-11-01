package gui

import (
	"context"
	"errors"

	"github.com/diamondburned/gotk4/pkg/core/gerror"
	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
)

const (
	// pickFilesText is the title of the file picker and composeFilterText the
	// name of its filter.
	pickFilesText     = "Scegli i File compose"
	composeFilterText = "File compose"
)

// pickComposeFiles shows the picker of one or more File compose on parent and
// hands the chosen paths to onPicked. A picker the user closed calls onPicked
// with nothing, so the caller keeps what it had. onError reports a picker that
// did not open.
func pickComposeFiles(ctx context.Context, parent *gtk.Window, onPicked func([]string), onError func(error)) {
	picker := gtk.NewFileDialog()
	picker.SetTitle(pickFilesText)
	picker.SetModal(true)

	filter := gtk.NewFileFilter()
	filter.SetName(composeFilterText)
	filter.AddPattern("*.yml")
	filter.AddPattern("*.yaml")
	picker.SetDefaultFilter(filter)
	picker.SetFilters(filterList(filter))

	if ctx == nil {
		ctx = context.Background()
	}
	picker.OpenMultiple(ctx, parent, func(res gio.AsyncResulter) {
		model, err := picker.OpenMultipleFinish(res)
		if err != nil {
			if !dismissed(err) {
				onError(err)
			}
			return
		}
		paths := make([]string, 0, model.NItems())
		for i := uint(0); i < model.NItems(); i++ {
			file := &gio.File{Object: model.Item(i)}
			paths = append(paths, file.Path())
		}
		onPicked(paths)
	})
}

// filterList builds the list of filters of the file picker. A GTK filter list
// is a store of GtkFileFilter objects.
func filterList(filters ...*gtk.FileFilter) *gio.ListStore {
	store := gio.NewListStore(glib.TypeFromName("GtkFileFilter"))
	for _, filter := range filters {
		store.Append(filter.Object)
	}
	return store
}

// dismissed reports whether the picker ended because the user closed it. Any
// other error means the picker did not open, and the form says so.
func dismissed(err error) bool {
	var gerr *gerror.GError
	if !errors.As(err, &gerr) {
		return false
	}
	return gerr.Quark() == uint32(gtk.DialogErrorQuark()) &&
		gerr.ErrorCode() == int(gtk.DialogErrorDismissed)
}
