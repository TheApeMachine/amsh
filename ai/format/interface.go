package format

type ResponseFormat interface {
	Format() ResponseFormat
	String() string
}
