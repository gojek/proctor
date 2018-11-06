package io

var originalPrinter Printer = nil

func SetupMockPrinter(mockPrinter Printer) {
	if originalPrinter == nil {
		originalPrinter = printerInstance
	}
	printerInstance = mockPrinter
}

func ResetPrinter() {
	if originalPrinter != nil {
		printerInstance = originalPrinter
		originalPrinter = nil
	}
}
