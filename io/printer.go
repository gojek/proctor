package io

import "github.com/fatih/color"

type Printer interface {
	Println(string, ...color.Attribute)
}

var printerInstance Printer

type printer struct{}

func GetPrinter() Printer {
	if printerInstance == nil {
		printerInstance = &printer{}
	}
	return printerInstance
}

func (p *printer) Println(msg string, attr ...color.Attribute) {
	color.New(attr...).Println(msg)
}
