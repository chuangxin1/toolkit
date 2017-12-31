// Copyright 2017 Manu Martinez-Almeida.  All rights reserved.
// Use of this source code is governed by a MIT style
// license that can be found in the LICENSE file.

package binding

import (
	"bytes"
	"errors"
	"fmt"
	"reflect"
	"sync"

	"gopkg.in/go-playground/validator.v8"
)

type defaultValidator struct {
	once     sync.Once
	validate *validator.Validate
}

var _ StructValidator = &defaultValidator{}

func (v *defaultValidator) ValidateStruct(obj interface{}) error {
	if kindOfData(obj) == reflect.Struct {
		v.lazyinit()
		if err := v.validate.Struct(obj); err != nil {
			buf := bytes.NewBufferString("")
			for _, v := range err.(validator.ValidationErrors) {
				buf.WriteString(fmt.Sprintf("%s %s %s", v.Name, v.Tag, v.Param))
				buf.WriteString(";")
			}
			return errors.New(buf.String()) //error(err)
		}
	}
	return nil
}

func (v *defaultValidator) RegisterValidation(key string, fn validator.Func) error {
	v.lazyinit()
	return v.validate.RegisterValidation(key, fn)
}

func (v *defaultValidator) lazyinit() {
	v.once.Do(func() {
		config := &validator.Config{TagName: "validate"}
		v.validate = validator.New(config)
	})
}

func kindOfData(data interface{}) reflect.Kind {
	value := reflect.ValueOf(data)
	valueType := value.Kind()
	if valueType == reflect.Ptr {
		valueType = value.Elem().Kind()
	}
	return valueType
}
