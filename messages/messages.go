package messages

type Mode string

const (
    NormalMode Mode = "NORMAL"
    InsertMode Mode = "INSERT"
)

type StatusUpdateMsg struct {
    Filename string
    Mode     Mode
}

type FileSelectedMsg struct {
    Path   string
    Width  int
    Height int
}

type OpenFileBrowserMsg struct {
    Width  int
    Height int
}