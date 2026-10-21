# 墨迹清单 Go 版

待办清单，数据文件 `%AppData%\todo-go\todo_taro_todos.json`，与 Taro / 小程序 JSON 互通。

## 构建

```bash
cd todo-go
go mod tidy
go build -o todo-go.exe .
```

## 图形 / UI

- **Fyne v2**（`fyne.io/fyne/v2`）：GLFW + OpenGL 渲染的桌面控件
