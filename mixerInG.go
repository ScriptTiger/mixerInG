package mixerInG

import (
	"errors"
	"math"
	"os"

	"github.com/go-audio/audio"
	"github.com/go-audio/wav"
)

// Structure to hold information for a track
type TrackInfo struct {
	wavDec *wav.Decoder
	bitDepth int
	bufferSize int
	intBuffer *audio.IntBuffer
	floatBuffer *audio.FloatBuffer
}

// Function to read PCM data from a trackInfo wav decoder into its int PCM buffer
func (i *TrackInfo) ReadWavToBuffer() {
	var err error
	i.bufferSize, err = i.wavDec.PCMBuffer(i.intBuffer)
	if i.bufferSize == 0 || err != nil {return}
	i.floatBuffer = i.intBuffer.AsFloatBuffer()
}

// Function to create a new trackInfo
func NewTrackInfo(wavDec *wav.Decoder, bitDepth, bufferCap int) (newTrack *TrackInfo) {
	return &TrackInfo{
		wavDec: wavDec,
		bitDepth: bitDepth,
		intBuffer: &audio.IntBuffer{Format: wavDec.Format(), Data: make([]int, bufferCap)},
		floatBuffer: &audio.FloatBuffer{Format: wavDec.Format(), Data: make([]float64, bufferCap)},
	}
}

// Function to sum buffers
func SumFloatBuffers(trackBufDst, trackBufSrc *audio.FloatBuffer, bufferSize int) {
	for i := 0; i < bufferSize; i++ {trackBufDst.Data[i] = trackBufDst.Data[i]+trackBufSrc.Data[i]}
}

// Function to attenuate linearly to prevent clipping in real-time without knowing/using true peak, RMS, LUFS, etc.
func AttenuateFloatBuffer(trackBufDst *audio.FloatBuffer, numTracks, bufferSize int) {
	for i := 0; i < bufferSize; i++ {trackBufDst.Data[i] = trackBufDst.Data[i]/float64(numTracks)}
}

// Function to scale input bit depth to output bit depth
func ScaleFloatBuffer(trackBufDst *audio.FloatBuffer, srcBitDepth, dstBitDepth, bufferSize int) {
	for i := 0; i < bufferSize; i++ {trackBufDst.Data[i] = trackBufDst.Data[i]*math.Pow(2, float64(dstBitDepth-srcBitDepth))}
}

// Function to read wav decoders into buffers to be mixed and write mix to output
func MixWavDecoders(wavDecs []*wav.Decoder, wavOut *os.File, bitDepth int, attenuate bool) (error) {

	var (
		format *audio.Format
		sampleRate uint32
		numChans uint16
	)

	numTracks := len(wavDecs)
	index := make([]*TrackInfo, numTracks)
	bufferCap := 8000

	// Validate tracks, populate index and convenience variables
	for i, wavDec := range wavDecs {

		if !wavDec.IsValidFile() {return errors.New("Invalid file")}

		index[i] = NewTrackInfo(wavDec, int(wavDec.BitDepth), bufferCap)

		if i == 0 {
			format = wavDec.Format()
			sampleRate = wavDec.SampleRate
			numChans = wavDec.NumChans
		} else if sampleRate != wavDec.SampleRate {return errors.New("Sample rate mismatch")
		} else if numChans != wavDec.NumChans {return errors.New("Channel layout mismatch")}
	}

	// Initialize wav encoder
	wavEnc := wav.NewEncoder(
		wavOut,
		int(sampleRate),
		bitDepth,
		int(numChans),
		1,
	)
	defer wavEnc.Close()

	// Initialize mix buffer
	mixFloatBuffer := &audio.FloatBuffer{Format: format, Data: make([]float64, bufferCap)}

	// Systematically load buffers, scale if needed, sum all tracks to mix buffer, adjust gain if requested, and write to output
	for {
		var bufferSize int
		for _, track := range index {
			track.ReadWavToBuffer()
			if track.bufferSize != 0 {
				if track.bitDepth != bitDepth {ScaleFloatBuffer(track.floatBuffer, track.bitDepth, bitDepth, track.bufferSize)}
				if bufferSize == 0 {mixFloatBuffer = track.floatBuffer
				} else {SumFloatBuffers(mixFloatBuffer, track.floatBuffer, track.bufferSize)}
				if track.bufferSize > bufferSize {bufferSize = track.bufferSize}
			}
		}
		if bufferSize == 0 {break}
		if attenuate {AttenuateFloatBuffer(mixFloatBuffer, numTracks, bufferSize)}
		wavEnc.Write(mixFloatBuffer.AsIntBuffer())
	}

	return nil

}

// Function to mix wav files and write mix to output
func MixWavFiles(files []*string, outWavName *string, bitDepth int, attenuate bool) (error) {

	// Initialize wavDecs to number of tracks
	wavDecs := make([]*wav.Decoder, len(files))

	// Initialize slice for wave decoders
	for i, file := range files {
		wavFile, err := os.Open(*file)
		if err != nil {return err}
		defer wavFile.Close()
		wavDecs[i] = wav.NewDecoder(wavFile)
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
	err = MixWavDecoders(wavDecs, wavOut, bitDepth, attenuate)
	if err != nil {return err}

	return nil
}