package main

import (
	"os"
	"strings"

	"github.com/ScriptTiger/mixerInG"
)

// Function to display help text and exit
func help(err int) {
	os.Stdout.WriteString(
		"Usage: mixerInG [options...]\n"+
		" -i <file>      Input WAV file (must be used for each input, for at least 2 inputs)\n"+
		" -o <file>      Destination WAV file\n",
	)
	os.Exit(err)
}

func main() {

	// Ensure valid number of arguments
	if len(os.Args) < 4 {help(1)}

	// Declare argument pointers
	var (
		files []*string
		wavOutName *string
	)

	// Push arguments to pointers
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
				default:
					help(3)
			}
		} else {help(4)}
	}

	// Ensure at least 2 inputs
	if len(files) < 2 {help(4)}

	// Set default output as standard output if no output given as argument
	if wavOutName == nil {
		wavOutName = new(string)
		*wavOutName = "-"
	}

	// Mix files and write mix to output
	err := mixerInG.MixWavFiles(files, wavOutName, false)
	if err != nil {panic(err)}

}