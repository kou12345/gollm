package utils

import "github.com/fatih/color"

var (
	ErrorColor   = color.New(color.FgRed).SprintFunc()
	SuccessColor = color.New(color.FgGreen).SprintFunc()
	UserColor    = color.New(color.FgCyan).SprintFunc()
	AIColor      = color.New(color.FgYellow).SprintFunc()
)
