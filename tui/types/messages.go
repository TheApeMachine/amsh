package types

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

type LoadFileMsg struct {
	Filepath string
}
