package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	"github.com/swqztech/todo-go/todostore"
)

type filterKind int

const (
	filterAll    filterKind = 0
	filterActive filterKind = 1
	filterDone   filterKind = 2
)

func main() {
	a := app.NewWithID("com.swqztech.todolist")
	w := a.NewWindow("墨迹清单")
	w.Resize(fyne.NewSize(480, 720))

	dataDir := filepath.Join(os.Getenv("APPDATA"), "todo-go")
	_ = os.MkdirAll(dataDir, 0o755)
	store := todostore.New(dataDir)

	var currentFilter filterKind = filterAll
	var selectedStoreIdx = -1

	counts := widget.NewLabel("")
	input := widget.NewEntry()
	input.SetPlaceHolder("新待办")

	type row struct {
		storeIdx int
		text     string
		done     bool
	}
	var rows []row

	var refresh func()
	list := widget.NewList(
		func() int { return len(rows) },
		func() fyne.CanvasObject {
			check := widget.NewCheck("", nil)
			label := widget.NewLabel("")
			return container.NewHBox(check, label)
		},
		func(i widget.ListItemID, o fyne.CanvasObject) {
			box := o.(*fyne.Container)
			check := box.Objects[0].(*widget.Check)
			label := box.Objects[1].(*widget.Label)
			r := rows[i]
			label.SetText(r.text)
			if r.done {
				label.TextStyle = fyne.TextStyle{Italic: true}
			} else {
				label.TextStyle = fyne.TextStyle{}
			}
			check.SetChecked(r.done)
			check.OnChanged = func(on bool) {
				store.SetDone(r.storeIdx, on)
				refresh()
			}
		},
	)
	list.OnSelected = func(id widget.ListItemID) {
		if id < 0 || int(id) >= len(rows) {
			selectedStoreIdx = -1
			return
		}
		selectedStoreIdx = rows[id].storeIdx
	}

	refresh = func() {
		items := store.Items()
		rows = nil
		for i, it := range items {
			if currentFilter == filterActive && it.Done {
				continue
			}
			if currentFilter == filterDone && !it.Done {
				continue
			}
			rows = append(rows, row{storeIdx: i, text: it.Text, done: it.Done})
		}
		counts.SetText(store.CountsLabel())
		list.UnselectAll()
		list.Refresh()
	}

	weekday := []string{"星期日", "星期一", "星期二", "星期三", "星期四", "星期五", "星期六"}
	now := time.Now()
	dateLabel := widget.NewLabelWithStyle(
		fmtDate(now, weekday),
		fyne.TextAlignLeading,
		fyne.TextStyle{Bold: true},
	)

	btnAdd := widget.NewButton("加入", func() {
		store.Add(input.Text)
		input.SetText("")
		selectedStoreIdx = -1
		refresh()
	})

	makeFilter := func(label string, f filterKind) *widget.Button {
		return widget.NewButton(label, func() {
			currentFilter = f
			selectedStoreIdx = -1
			refresh()
		})
	}

	btnExport := widget.NewButton("导出", func() {
		w.Clipboard().SetContent(store.ExportMarkdown())
		dialog.ShowInformation("墨迹清单", "已复制到剪贴板", w)
	})
	btnImport := widget.NewButton("导入", func() {
		dialog.ShowFileOpen(func(r fyne.URIReadCloser, err error) {
			if err != nil || r == nil {
				return
			}
			defer r.Close()
			b, err := os.ReadFile(r.URI().Path())
			if err != nil {
				return
			}
			path := r.URI().Path()
			if filepath.Ext(path) == ".json" {
				if err := store.ReplaceFromJSON(string(b)); err != nil {
					dialog.ShowError(err, w)
					return
				}
			} else {
				store.ImportMarkdown(string(b))
			}
			refresh()
			dialog.ShowInformation("墨迹清单", "导入完成", w)
		}, w)
	})
	btnPaste := widget.NewButton("粘贴导入", func() {
		store.ImportMarkdown(w.Clipboard().Content())
		refresh()
		dialog.ShowInformation("墨迹清单", "已从剪贴板导入", w)
	})

	btnUp := widget.NewButton("上移", func() {
		if selectedStoreIdx < 0 {
			dialog.ShowInformation("墨迹清单", "请先选中一条", w)
			return
		}
		if selectedStoreIdx == 0 {
			return
		}
		store.MoveItem(selectedStoreIdx, selectedStoreIdx-1)
		selectedStoreIdx--
		refresh()
	})
	btnDown := widget.NewButton("下移", func() {
		if selectedStoreIdx < 0 {
			dialog.ShowInformation("墨迹清单", "请先选中一条", w)
			return
		}
		if selectedStoreIdx >= len(store.Items())-1 {
			return
		}
		store.MoveItem(selectedStoreIdx, selectedStoreIdx+1)
		selectedStoreIdx++
		refresh()
	})
	btnEdit := widget.NewButton("编辑", func() {
		if selectedStoreIdx < 0 {
			dialog.ShowInformation("墨迹清单", "请先选中一条", w)
			return
		}
		items := store.Items()
		if selectedStoreIdx >= len(items) {
			return
		}
		old := items[selectedStoreIdx].Text
		entry := widget.NewEntry()
		entry.SetText(old)
		dialog.ShowCustomConfirm("编辑待办", "确定", "取消", entry, func(ok bool) {
			if !ok {
				return
			}
			store.UpdateText(selectedStoreIdx, entry.Text)
			refresh()
		}, w)
	})
	btnDel := widget.NewButton("删除", func() {
		if selectedStoreIdx < 0 {
			dialog.ShowInformation("墨迹清单", "请先选中一条", w)
			return
		}
		dialog.ShowConfirm("墨迹清单", "删除这条待办？", func(ok bool) {
			if !ok {
				return
			}
			store.RemoveAt(selectedStoreIdx)
			selectedStoreIdx = -1
			refresh()
		}, w)
	})

	refresh()

	w.SetContent(container.NewVBox(
		widget.NewLabelWithStyle("INK LIST", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		dateLabel,
		widget.NewLabel("写下来，划掉，继续。支持导出 / 导入 / 粘贴 Markdown。"),
		container.NewBorder(nil, nil, nil, btnAdd, input),
		container.NewGridWithColumns(3, makeFilter("全部", filterAll), makeFilter("未完成", filterActive), makeFilter("已完成", filterDone)),
		container.NewGridWithColumns(3, btnExport, btnImport, btnPaste),
		counts,
		list,
		container.NewGridWithColumns(4, btnUp, btnDown, btnEdit, btnDel),
	))
	w.ShowAndRun()
}

func fmtDate(t time.Time, weekday []string) string {
	return fmt.Sprintf("%d月%d日 %s", t.Month(), t.Day(), weekday[t.Weekday()])
}
