package io

import "github.com/fatih/color"

type Printer interface {
	Println(string, ...color.Attribute)
}

type printer struct{}

func NewPrinter() Printer {
	return &printer{}
}

func (p *printer) Println(msg string, attr ...color.Attribute) {
	color.New(attr...).Println(msg)
}
