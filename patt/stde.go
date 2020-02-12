// Copyright (C) 2017, ccpaging <ccpaging@gmail.com>.  All rights reserved.

package patt

/** Global Encoders ***/

type Encoders struct {
	Level  LevelEncoder
	Time   TimeEncoder
	Caller CallerEncoder
	Fields FieldsEncoder
}

// Encoders is global encoders for external packages extending.
var stde *Encoders

func init() {
	stde = &Encoders{
		Level:  NewNopLevelEncoder(),
		Time:   NewTimeEncoder(),
		Caller: NewCallerEncoder(),
		Fields: NewFieldsEncoder(),
	}
}

func GetEncoders() *Encoders {
	return stde
}

func SetEncoders(e *Encoders) {
	stde = e
}
