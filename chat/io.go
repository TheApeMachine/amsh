package chat

func (model *Model) Read(p []byte) (n int, err error) {
	return model.reader.Read(p)
}

func (model *Model) Write(p []byte) (n int, err error) {
	return model.writer.Write(p)
}

func (model *Model) Close() error {
	return model.aiIO.Close()
}
