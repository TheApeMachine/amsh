package types

/*
OpenChatMsg is sent when the chat window should be opened with the given context.
The context is typically highlighted text from the editor.
*/
type OpenChatMsg struct {
	Context string
}

type AISendMsg struct{}

type AIChunkMsg struct {
	Chunk string
}
type AIPromptMsg struct {
	Prompt string
}
type AIResponseMsg struct {
	Response string
}
