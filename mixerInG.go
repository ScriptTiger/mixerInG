package mixerInG

import (
	"errors"
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

// Function to sum buffers
func SumFloatBuffers(trackBufDst, trackBufSrc *audio.FloatBuffer, bufferSize int) {
	for i := 0; i < bufferSize; i++ {trackBufDst.Data[i] = trackBufDst.Data[i]+trackBufSrc.Data[i]}
}

// Function to attenuate linearly to prevent clipping in real-time without knowing/using true peak, RMS, LUFS, etc.
func attenuateFloatBuffer(trackBufDst *audio.FloatBuffer, numTracks float64, bufferSize int) {
	for i := 0; i < bufferSize; i++ {trackBufDst.Data[i] = trackBufDst.Data[i]/numTracks}
}

// Function to read wav decoders into buffers to be mixed and write mix to output
func MixWavDecoders(trackDecs []*wav.Decoder, wavOut *os.File, attenuate bool) (error) {

	var (
		bufferSize int
		format *audio.Format
		sampleRate uint32
		bitDepth uint16
		numChans uint16
		duration float64
		err error
	)

	// Buffer capacity
	bufferCap := 8000

	// Validate tracks and populate format properties
	for i, trackDec := range trackDecs {

		if !trackDec.IsValidFile() {return errors.New("Invalid file")}

		durationTime, _ := trackDec.Duration()
		if i == 0 {
			format = trackDec.Format()
			sampleRate = trackDec.SampleRate
			numChans = trackDec.NumChans
			bitDepth = trackDec.BitDepth
			duration = durationTime.Seconds()
		} else if sampleRate != trackDec.SampleRate {return errors.New("Sample rate mismatch")
		} else if numChans != trackDec.NumChans {return errors.New("Channel layout mismatch")
		} else if bitDepth != trackDec.BitDepth {return errors.New("Bit depth mismatch")
		} else if duration != durationTime.Seconds() {return errors.New("Duration mismatch")}
	}

	// Initialize wav encoder
	wavEnc := wav.NewEncoder(
		wavOut,
		int(sampleRate),
		int(bitDepth),
		int(numChans),
		1,
	)
	defer wavEnc.Close()

	// Initialize buffers
	trackBufDstInt := &audio.IntBuffer{Format: format, Data: make([]int, bufferCap)}
	trackBufSrcInt := &audio.IntBuffer{Format: format, Data: make([]int, bufferCap)}
	trackBufDstFloat := &audio.FloatBuffer{Format: format, Data: make([]float64, bufferCap)}

	// Systematically load buffers and sum track 0 buffer with all other track buffers, adjust gain, and write to output
	for {
		for i, trackDec := range trackDecs {
			if i == 0 {
				bufferSize, err = trackDec.PCMBuffer(trackBufDstInt)
				if bufferSize == 0 || err != nil {break}
				trackBufDstFloat = trackBufDstInt.AsFloatBuffer()
			} else {
				bufferSize, err = trackDec.PCMBuffer(trackBufSrcInt)
				if bufferSize == 0 || err != nil {break}
				SumFloatBuffers(trackBufDstFloat, trackBufSrcInt.AsFloatBuffer(), bufferSize)
			}
		}
		if bufferSize == 0 || err != nil {break}
		if attenuate {attenuateFloatBuffer(trackBufDstFloat, float64(len(trackDecs)), bufferSize)}
		wavEnc.Write(trackBufDstFloat.AsIntBuffer())
	}

	return nil

}

// Function to mix wav files and write mix to output
func MixWavFiles(files []*string, outWavName *string, attenuate bool) (error) {

	// Initialize trackDecs to number of tracks
	trackDecs := make([]*wav.Decoder, len(files))

	// Initialize slice for wave decoders
	for i, file := range files {
		wavFile, err := os.Open(*file)
		if err != nil {return err}
		defer wavFile.Close()
		trackDecs[i] = wav.NewDecoder(wavFile)
	}

	// Initialize output
	var wavOut *os.File
	var err error
	if *outWavName == "-" {wavOut = os.Stdout
	} else {
		wavOut, err = os.Create(*outWavName)
		if err != nil {return err}
		defer wavOut.Close()
	}

	// Mix decoders
	err = MixWavDecoders(trackDecs, wavOut, attenuate)
	if err != nil {return err}

	return nil
}