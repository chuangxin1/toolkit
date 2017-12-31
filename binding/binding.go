// Copyright 2014 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import (
	"net/http"

	validator "gopkg.in/go-playground/validator.v8"
)

const (
	MIMEJSON              = "application/json"
	MIMEHTML              = "text/html"
	MIMEXML               = "application/xml"
	MIMEXML2              = "text/xml"
	MIMEPlain             = "text/plain"
	MIMEPOSTForm          = "application/x-www-form-urlencoded"
	MIMEMultipartPOSTForm = "multipart/form-data"
	MIMEPROTOBUF          = "application/x-protobuf"
	MIMEMSGPACK           = "application/x-msgpack"
	MIMEMSGPACK2          = "application/msgpack"
)

// Binding bind http request params to struct
type Binding interface {
	Name() string
	Bind(*http.Request, interface{}) error
}

// StructValidator Validator
type StructValidator interface {
	// ValidateStruct can receive any kind of type and it should never panic, even if the configuration is not right.
	// If the received type is not a struct, any validation should be skipped and nil must be returned.
	// If the received type is a struct or pointer to a struct, the validation should be performed.
	// If the struct is not valid or the validation itself fails, a descriptive error should be returned.
	// Otherwise nil must be returned.
	ValidateStruct(interface{}) error

	// RegisterValidation adds a validation Func to a Validate's map of validators denoted by the key
	// NOTE: if the key already exists, the previous validation function will be replaced.
	// NOTE: this method is not thread-safe it is intended that these all be registered prior to any validation
	RegisterValidation(string, validator.Func) error
}

var Validator StructValidator = &defaultValidator{}

var (
	JSON          = jsonBinding{}
	XML           = xmlBinding{}
	Form          = formBinding{}
	Query         = queryBinding{}
	FormPost      = formPostBinding{}
	FormMultipart = formMultipartBinding{}
	ProtoBuf      = protobufBinding{}
	MsgPack       = msgpackBinding{}
)

// BindDefault default binding
func BindDefault(method, contentType string) Binding {
	if method == "GET" {
		return Form
	}

	switch contentType {
	case MIMEJSON:
		return JSON
	case MIMEXML, MIMEXML2:
		return XML
	case MIMEPROTOBUF:
		return ProtoBuf
	case MIMEMSGPACK, MIMEMSGPACK2:
		return MsgPack
	default: //case MIMEPOSTForm, MIMEMultipartPOSTForm:
		return Form
	}
}

func validate(obj interface{}) error {
	if Validator == nil {
		return nil
	}

	return Validator.ValidateStruct(obj)
}

func Bind(req *http.Request, obj interface{}) error {
	b := BindDefault(req.Method, ContentType(req))
	return MustBindWith(req, obj, b)
}

func MustBindWith(req *http.Request, obj interface{}, b Binding) (err error) {
	return b.Bind(req, obj)
}

func ContentType(req *http.Request) string {
	return filterFlags(req.Header.Get("Content-Type"))
}

func filterFlags(content string) string {
	for i, char := range content {
		if char == ' ' || char == ';' {
			return content[:i]
		}
	}
	return content
}
