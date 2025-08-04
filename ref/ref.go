package main

import (
	"math"
	"os"
	"strconv"
	"strings"

	"github.com/ScriptTiger/mixerInG"
)

// Function to display help text and exit
func help(err int) {
	os.Stdout.WriteString(
		"Usage: mixerInG [options...]\n"+
		" -i <file>        Input WAV file (must be used for each input)\n"+
		" -o <file>        Destination WAV file of mix\n"+
		" -bits <number>   Bit depth of mix WAV file (16|24|32)\n"+
		" -attenuate       Attenuate linearly to prevent clipping, dividing by number of tracks mixed\n"+
		" -nowrite         Do not write mix to file\n"+
		" -nostats         Do not collect stats\n"+
		" -buffer <number> Number of samples to buffer per track\n"+
		"\n"+
		" Input options (must precede target input):\n"+
		" -gain <number> Make gain adjustment\n"+
		" -invert        Invert polarity\n",

	)
	os.Exit(err)
}

func main() {

	// Ensure valid number of arguments
	if len(os.Args) < 2 {help(1)}

	// Declare argument variables, pointers, and other common variables
	var (
		files []*string
		wavOutName *string
		fx []*mixerInG.TrackFX
		gain *float64
		invert bool
		bitDepth int
		attenuate bool
		nowrite bool
		nostats bool
		stats []*mixerInG.TrackStats
		bufferCap *uint64
		err error
	)

	// Push arguments to pointers or set appropriate variables
	for i := 1; i < len(os.Args); i++ {
		if strings.HasPrefix(os.Args[i], "-") {
			switch strings.TrimPrefix(os.Args[i], "-") {
				case "gain":
					i++
					if gain != nil {help(2)}
					gain = new(float64)
					if strings.HasSuffix(os.Args[i], "dB") {
						*gain, err = strconv.ParseFloat(strings.TrimSuffix(os.Args[i], "dB"), 64)
						if err != nil {help(3)}
						*gain = math.Pow(10, *gain/20)
					} else {
						*gain, err = strconv.ParseFloat(os.Args[i], 64)
					}
					if err != nil {help(4)}
					continue
				case "invert":
					if invert {help(5)}
					invert = true
					continue
				case "i":
					i++
					files = append(files, &os.Args[i])
					if gain == nil {
						gain = new(float64)
						*gain = 1
					}
					fx = append(fx, &mixerInG.TrackFX{Gain: *gain, Invert: invert})
					gain = nil
					invert = false
					if !nostats {stats = append(stats, &mixerInG.TrackStats{})}
					continue
				case "o":
					if wavOutName != nil {help(6)}
					i++
					wavOutName = &os.Args[i]
					continue
				case "bits":
					if bitDepth != 0 {help(7)}
					i++
					bitDepth, err = strconv.Atoi(os.Args[i])
					if err != nil ||
					(bitDepth != 16 &&
					bitDepth != 24 &&
					bitDepth != 32) {help(8)}
					continue
				case "attenuate":
					if attenuate {help(9)}
					attenuate = true
					continue
				case "nowrite":
					if nowrite {help(10)}
					nowrite = true
					continue
				case "nostats":
					if nostats {help(11)}
					nostats = true
					stats = nil
					continue
				case "buffer":
					if bufferCap != nil {help(12)}
					i++
					bufferCap = new(uint64)
					*bufferCap, err = strconv.ParseUint(os.Args[i], 10, 64)
					if err != nil ||
					*bufferCap > 2147483647 {help(13)}
					continue
				case "":
					if wavOutName != nil {help(14)}
					wavOutName = &os.Args[i]
					continue
				default:
					help(15)
			}
		} else {
			if wavOutName != nil {help(16)}
			wavOutName = &os.Args[i]
			continue
		}
	}


	// Validate arguments
	if len(files) < 1 ||
	gain != nil ||
	invert ||
	(wavOutName != nil && nowrite) ||
	(nowrite && nostats) {help(17)}

	// Set default output as standard output if no output given as argument, and disable stats whenever output is standard ouput
	if !nowrite {
		if wavOutName == nil {
			wavOutName = new(string)
			*wavOutName = "-"
		}
		if *wavOutName == "-" {
			nostats = true
			stats = nil
		}
	}

	// Set default buffer size, 0 will default to 8000 samples
	if bufferCap == nil {bufferCap = new(uint64)}

	// Set default bit depth of mix if none specified
	if bitDepth == 0 {bitDepth = 24}

	// Create TrackStats for mix track if stats are enabled
	if !nostats {stats = append(stats, &mixerInG.TrackStats{})}

	// Mix files and write mix to output
	if nowrite {os.Stdout.WriteString("Processing mix without writing...\n")
	} else if *wavOutName != "-" {os.Stdout.WriteString("Writing mix to "+*wavOutName+"...\n")}
	err = mixerInG.MixWavFiles(files, wavOutName, fx, bitDepth, attenuate, stats, int(*bufferCap))
	if nowrite || *wavOutName != "-" {
		if err != nil {os.Stdout.WriteString(err.Error()+"\n")}
		if !nostats {
			for i, file := range files {
				os.Stdout.WriteString(
					"----- Stats for \""+*file+"\" -----\n"+
					strconv.FormatFloat(stats[i].RMSdB, 'f', -1, 64)+"dB RMS.\n"+
					strconv.FormatFloat(stats[i].PeakdB, 'f', -1, 64)+"dB peak.\n"+
					strconv.FormatUint(stats[i].ClippedCount, 10)+" clipped samples.\n"+
					strconv.FormatUint(stats[i].SampleCount, 10)+" total samples.\n",
				)
			}
			mix := len(stats)-1
			os.Stdout.WriteString(
				"----- Stats for mix -----\n"+
				strconv.FormatFloat(stats[mix].RMSdB, 'f', -1, 64)+"dB RMS.\n"+
				strconv.FormatFloat(stats[mix].PeakdB, 'f', -1, 64)+"dB peak.\n"+
				strconv.FormatUint(stats[mix].ClippedCount, 10)+" clipped samples.\n"+
				strconv.FormatUint(stats[mix].SampleCount, 10)+" total samples.\n",
			)
		}
	}
}