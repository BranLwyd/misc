package main

import (
	"fmt"
	"os"
	"time"
)

//  ____
// |    |
// |    |
//  ____
// |    |
// |    |
//  ____

const digitHeight = 7

var (
	colon = [digitHeight]string{
		"      ",
		"      ",
		"   .  ",
		"      ",
		"   .  ",
		"      ",
		"      ",
	}
	aDigit = [digitHeight]string{
		" ____ ",
		"|    |",
		"|    |",
		" ____ ",
		"|    |",
		"|    |",
		"      ",
	}
	pDigit = [digitHeight]string{
		" ____ ",
		"|    |",
		"|    |",
		" ____ ",
		"|     ",
		"|     ",
		"      ",
	}

	digits = [10][digitHeight]string{
		{
			" ____ ",
			"|    |",
			"|    |",
			"      ",
			"|    |",
			"|    |",
			" ____ ",
		},
		{
			"      ",
			"     |",
			"     |",
			"      ",
			"     |",
			"     |",
			"      ",
		},
		{
			" ____ ",
			"     |",
			"     |",
			" ____ ",
			"|     ",
			"|     ",
			" ____ ",
		},
		{
			" ____ ",
			"     |",
			"     |",
			" ____ ",
			"     |",
			"     |",
			" ____ ",
		},
		{
			"      ",
			"|    |",
			"|    |",
			" ____ ",
			"     |",
			"     |",
			"      ",
		},
		{
			" ____ ",
			"|     ",
			"|     ",
			" ____ ",
			"     |",
			"     |",
			" ____ ",
		},
		{
			" ____ ",
			"|     ",
			"|     ",
			" ____ ",
			"|    |",
			"|    |",
			" ____ ",
		},
		{
			" ____ ",
			"     |",
			"     |",
			"      ",
			"     |",
			"     |",
			"      ",
		},
		{
			" ____ ",
			"|    |",
			"|    |",
			" ____ ",
			"|    |",
			"|    |",
			" ____ ",
		},
		{
			" ____ ",
			"|    |",
			"|    |",
			" ____ ",
			"     |",
			"     |",
			" ____ ",
		},
	}
)

const (
	civilianTimeFormat        = "3:04PM"
	civilianSecondsTimeFormat = "3:04:05PM"
	militaryTimeFormat        = "15:04"
	militarySecondsTimeFormat = "15:04:05"
)

func parseArgument(arg string) (_ time.Time, gotSeconds, isMilitary bool, _ error) {
	t, err := time.Parse(civilianTimeFormat, arg)
	if err == nil {
		return t, false, false, nil
	}

	t, err = time.Parse(civilianSecondsTimeFormat, arg)
	if err == nil {
		return t, true, false, nil
	}

	t, err = time.Parse(militaryTimeFormat, arg)
	if err == nil {
		return t, false, true, nil
	}

	t, err = time.Parse(militarySecondsTimeFormat, arg)
	if err == nil {
		return t, true, true, nil
	}

	return time.Time{}, false, false, fmt.Errorf("bad time format, expected 12:34[:56][PM], got %q", arg)
}

func toDigitalClockDigit(digit int) [digitHeight]string {
	if digit < 0 || digit > 9 {
		panic(fmt.Sprintf("Expected digit in range 0 through 9, got %d", digit))
	}
	return digits[digit]
}

func toDigitalClockDigits(number int) []string {
	if number < 0 || number > 99 {
		panic(fmt.Sprintf("Expected number in range 0 through 99, got %d", number))
	}
	tens, ones := number/10, number%10
	tensDigit, onesDigit := toDigitalClockDigit(tens), toDigitalClockDigit(ones)
	result := make([]string, len(tensDigit))
	for i := range tensDigit {
		result[i] = tensDigit[i] + " " + onesDigit[i]
	}
	return result
}

func die(format string, a ...interface{}) {
	fmt.Fprintf(os.Stderr, format+"\n", a...)
	os.Exit(1)
}

func toCivilianHour(hour int) (_ int, isAM bool) {
	switch {
	case hour == 0:
		return 12, true
	case hour < 12:
		return hour, true
	case hour == 12:
		return hour, false
	case hour < 24:
		return hour - 12, false
	default:
		panic(fmt.Sprintf("expected hour in range 0 through 23, got %d", hour))
	}
}

func main() {
	// Parse arguments.
	if len(os.Args) != 2 {
		die("Expected exactly one argument, e.g.: %s 12:34", os.Args[0]) // will there always be a first?
	}
	t, gotSeconds, isMilitary, err := parseArgument(os.Args[1])
	if err != nil {
		die("Could not parse argument: %v", err)
	}

	// Handle hour.
	var result [digitHeight]string
	hr := t.Hour()
	var isAM bool
	if !isMilitary {
		hr, isAM = toCivilianHour(t.Hour())
	}
	hrDigits := toDigitalClockDigits(hr)
	for i := range hrDigits {
		result[i] = hrDigits[i] + " " + colon[i]
	}

	// Handle minute.
	minDigits := toDigitalClockDigits(t.Minute())
	for i := range minDigits {
		result[i] = result[i] + " " + minDigits[i]
	}

	// Handle second, if specified.
	if gotSeconds {
		secDigits := toDigitalClockDigits(t.Second())
		for i := range secDigits {
			result[i] = result[i] + " " + colon[i] + " " + secDigits[i]
		}
	}

	// Handle AM/PM, if civilian time.
	if !isMilitary {
		if isAM {
			for i := range aDigit {
				result[i] = result[i] + " " + aDigit[i]
			}
		} else {
			for i := range pDigit {
				result[i] = result[i] + " " + pDigit[i]
			}
		}
	}

	for i := range result {
		fmt.Println(result[i])
	}
}
