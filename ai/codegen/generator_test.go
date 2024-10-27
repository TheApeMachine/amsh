package codegen

import (
	"context"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestGenerator(t *testing.T) {
	Convey("Given a code generator", t, func() {
		gen, err := NewGenerator()
		So(err, ShouldBeNil)
		So(gen, ShouldNotBeNil)

		Convey("When generating a simple Go struct", func() {
			req := CodeRequest{
				Language:    "go",
				Name:        "user",
				Description: "User represents a system user",
				Imports:     []string{"time"},
				Structs: []StructSpec{
					{
						Name: "User",
						Fields: []FieldSpec{
							{
								Name:       "ID",
								Type:       "string",
								Tag:        `json:"id"`,
								Visibility: "public",
							},
							{
								Name:       "CreatedAt",
								Type:       "time.Time",
								Tag:        `json:"created_at"`,
								Visibility: "public",
							},
						},
						Doc: []string{
							"User represents a system user with basic attributes",
							"It implements standard interfaces for JSON marshaling",
						},
					},
				},
				Tests: true,
			}

			code, err := gen.Generate(context.Background(), req)

			Convey("Then it should generate valid Go code", func() {
				So(err, ShouldBeNil)
				So(code, ShouldContainSubstring, "type User struct")
				So(code, ShouldContainSubstring, "ID string `json:\"id\"`")
				So(code, ShouldContainSubstring, "CreatedAt time.Time `json:\"created_at\"`")
				So(code, ShouldContainSubstring, "func TestUser")
			})
		})

		Convey("When generating an interface with methods", func() {
			req := CodeRequest{
				Language:    "go",
				Name:        "storage",
				Description: "Storage interface for data persistence",
				Interfaces: []InterfaceSpec{
					{
						Name: "Storage",
						Methods: []MethodSpec{
							{
								Name: "Save",
								Params: []FieldSpec{
									{
										Name: "ctx",
										Type: "context.Context",
									},
									{
										Name: "data",
										Type: "interface{}",
									},
								},
								Returns: []FieldSpec{
									{
										Name: "",
										Type: "error",
									},
								},
								Visibility: "public",
							},
						},
						Doc: []string{
							"Storage defines the interface for data persistence operations",
						},
					},
				},
				Tests: true,
			}

			code, err := gen.Generate(context.Background(), req)

			Convey("Then it should generate a valid interface", func() {
				So(err, ShouldBeNil)
				So(code, ShouldContainSubstring, "type Storage interface")
				So(code, ShouldContainSubstring, "Save(ctx context.Context, data interface{}) error")
				So(code, ShouldContainSubstring, "func TestStorage")
			})
		})
	})
}
