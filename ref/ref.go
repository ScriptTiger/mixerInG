package main

import (
	"os"
	"strconv"
	"strings"

	"github.com/ScriptTiger/mixerInG"
)

// Function to display help text and exit
func help(err int) {
	os.Stdout.WriteString(
		"Usage: mixerInG [options...]\n"+
		" -i <file>      Input WAV file (must be used for each input, for at least 2 inputs)\n"+
		" -o <file>      Destination WAV file of mix\n"+
		" -b <number>    Bit depth of mix WAV file (16|24|32)\n",

	)
	os.Exit(err)
}

func main() {

	// Ensure valid number of arguments
	if len(os.Args) < 4 {help(1)}

	// Declare argument variables, pointers, and other common variables
	var (
		files []*string
		wavOutName *string
		bitDepth int
		err error
	)

	// Push arguments to pointers or set appropriate variables
	for i := 1; i < len(os.Args); i++ {
		if strings.HasPrefix(os.Args[i], "-") {
			switch strings.TrimPrefix(os.Args[i], "-") {
				case "i":
					i++
					files = append(files, &os.Args[i])
					continue
				case "o":
					if wavOutName != nil {help(2)}
					i++
					wavOutName = &os.Args[i]
					continue
				case "b":
					if bitDepth != 0 {help(3)}
					i++
					bitDepth, err = strconv.Atoi(os.Args[i])
					if err != nil ||
					(bitDepth != 16 &&
					bitDepth != 24 &&
					bitDepth != 32) {help(4)}
					continue
				case "":
					if wavOutName != nil {help(5)}
					wavOutName = &os.Args[i]
					continue
				default:
					help(6)
			}
		} else {
			if wavOutName != nil {help(7)}
			wavOutName = &os.Args[i]
			continue
		}
	}

	// Ensure at least 2 inputs
	if len(files) < 2 {help(8)}

	// Set default output as standard output if no output given as argument
	if wavOutName == nil {
		wavOutName = new(string)
		*wavOutName = "-"
	}

	// Set default bit depth of mix if none specified
	if bitDepth == 0 {bitDepth = 24}

	// Mix files and write mix to output
	if *wavOutName != "-" {os.Stdout.WriteString("Writing mix to "+*wavOutName+"...\n")}
	err = mixerInG.MixWavFiles(files, wavOutName, bitDepth, false)
	if err != nil {os.Stdout.WriteString(err.Error()+"\n")}

}