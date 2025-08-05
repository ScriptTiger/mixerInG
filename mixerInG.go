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
	BitDepth int
	BufferSize int
	IntBuffer *audio.IntBuffer
	FloatBuffer *audio.FloatBuffer
}

// Structure to hold FX to perform
type TrackFX struct {
	Gain float64
	Invert bool
}

// Structure to hold stats
type TrackStats struct {
	SampleCount uint64
	ClippedCount uint64
	Peak float64
	PeakdB float64
	SumOfSquares float64
	RMSdB float64
}

// Function to mix TrackInfo float buffers to a provided mix float buffer, performing FX, scaling, and attenuation operations as needed, and return length of longest buffer
func Mix(mixTrack *audio.FloatBuffer, sourceTracks []*TrackInfo, fx []*TrackFX, bitDepth int, attenuate bool, stats []*TrackStats) (mixBufferSize int) {
	ScaleFloatBuffers(sourceTracks, bitDepth)
	if fx != nil {FXFloatBuffers(sourceTracks, fx)}
	mixBufferSize = SumFloatBuffers(mixTrack, sourceTracks)
	if attenuate {AttenuateFloatBuffer(mixTrack, len(sourceTracks), mixBufferSize)}
	if mixBufferSize < cap(mixTrack.Data) {mixTrack.Data = mixTrack.Data[:mixBufferSize]}
	if stats != nil {updateTrackStats(stats, bitDepth, sourceTracks, mixTrack, mixBufferSize)}
	return mixBufferSize
}

// Function to sum TrackInfo float buffers to a mix float buffer and return buffer size of mix, equal to length of longest buffer
func SumFloatBuffers(mixTrack *audio.FloatBuffer, sourceTracks []*TrackInfo) (mixBufferSize int) {
	for _, sourceTrack := range sourceTracks {
		if sourceTrack.BufferSize == 0 {continue}
		if mixBufferSize == 0 {
			for m, _ := range mixTrack.Data {mixTrack.Data[m] = 0}
			for s, data := range sourceTrack.FloatBuffer.Data {mixTrack.Data[s] = data}
			mixBufferSize = sourceTrack.BufferSize
			continue
		}
		for s := 0; s < sourceTrack.BufferSize; s++ {
			mixTrack.Data[s] += +sourceTrack.FloatBuffer.Data[s]
		}
		if sourceTrack.BufferSize > mixBufferSize {mixBufferSize = sourceTrack.BufferSize}
	}
	return mixBufferSize
}

// Function to apply FX
func FXFloatBuffers(tracks []*TrackInfo, fx []*TrackFX) {
	for f, track := range tracks {
		if track.BufferSize != 0 {
			// Gain
			if fx[f].Gain != 1 {
				for i := 0; i < track.BufferSize; i++ {
					track.FloatBuffer.Data[i] *= fx[f].Gain
				}
			}
			// Invert polarity
			if fx[f].Invert {
				for i := 0; i < track.BufferSize; i++ {
					track.FloatBuffer.Data[i] *= -1
				}
			}
		}
	}
}

// Function to attenuate linearly to prevent clipping in real-time without knowing/using true peak, RMS, LUFS, etc.
func AttenuateFloatBuffer(mixTrack *audio.FloatBuffer, numTracks, bufferSize int) {
	for i := 0; i < bufferSize; i++ {mixTrack.Data[i] /= float64(numTracks)}
}

// Function to scale input bit depth to output bit depth
func ScaleFloatBuffers(tracks []*TrackInfo, bitDepth int) {
	for _, track := range tracks {
		if track.BufferSize != 0 && track.BitDepth != bitDepth {
			scaleFactor := math.Pow(2, float64(bitDepth-track.BitDepth))
			for i := 0; i < track.BufferSize; i++ {
				track.FloatBuffer.Data[i] *= scaleFactor
			}
		}
	}
}

// Function to read PCM data from a wav decoder into a TrackInfo's buffers, set bufferSize, and return length of longest buffer
func ReadWavsToBuffers(wavDecs []*wav.Decoder, tracks []*TrackInfo) (mixBufferSize int) {
	var err error
	for i, track := range tracks {
		if track.BufferSize == 0 {continue}
		if track.BufferSize != -1 && track.BufferSize < cap(track.IntBuffer.Data) {
			track.BufferSize = 0
			continue
		}
		track.BufferSize, err = wavDecs[i].PCMBuffer(track.IntBuffer)
		if track.BufferSize == 0 || err != nil {continue}
		if track.BufferSize < cap(track.IntBuffer.Data) {
			track.IntBuffer.Data = track.IntBuffer.Data[:track.BufferSize]
			track.FloatBuffer.Data = track.FloatBuffer.Data[:track.BufferSize]
		}
		track.FloatBuffer = track.IntBuffer.AsFloatBuffer()
		if track.BufferSize > mixBufferSize {mixBufferSize = track.BufferSize}
	}
	return mixBufferSize
}

// Function to mix wav decoders and write mix to output
func MixWavDecoders(wavDecs []*wav.Decoder, wavOut *os.File, fx []*TrackFX, bitDepth int, attenuate bool, stats []*TrackStats, bufferCap int) (error) {

	var (
		format *audio.Format
		sampleRate int
		numChans int
	)

	index := make([]*TrackInfo, len(wavDecs))
	if bufferCap == 0 {bufferCap = 8000}

	// Validate tracks, populate index and convenience variables
	for i, wavDec := range wavDecs {

		if !wavDec.IsValidFile() {return errors.New("Invalid file")}

		if wavDec.WavAudioFormat == 3 {return errors.New("IEEE float is not currently supported")}
		if wavDec.WavAudioFormat == 6 {return errors.New("A-law is not currently supported")}
		if wavDec.WavAudioFormat == 7 {return errors.New("µ-law is not currently supported")}
		if wavDec.WavAudioFormat != 1 && wavDec.WavAudioFormat != 0xFFFE {return errors.New("Only signed PCM is currently supported")}

		if i == 0 {
			format = wavDec.Format()
			sampleRate = int(wavDec.SampleRate)
			numChans = int(wavDec.NumChans)
		} else if sampleRate != int(wavDec.SampleRate) {return errors.New("Sample rate mismatch")
		} else if numChans != int(wavDec.NumChans) {return errors.New("Channel layout mismatch")}

		index[i] = newTrack(format, int(wavDec.BitDepth), bufferCap)
	}

	// Initialize wav encoder
	var wavEnc *wav.Encoder
	if wavOut != nil {
		wavEnc = wav.NewEncoder(
			wavOut,
			sampleRate,
			bitDepth,
			numChans,
			1,
		)
		defer wavEnc.Close()
	}

	// Initialize mix buffer
	mixFloatBuffer := &audio.FloatBuffer{Format: format, Data: make([]float64, bufferCap)}

	// Systematically load buffers, scale if needed, sum all tracks to mix buffer, adjust gain if requested, and write to output
	for {
		mixBufferSize := ReadWavsToBuffers(wavDecs, index)
		if mixBufferSize == 0 {break}
		Mix(mixFloatBuffer, index, fx, bitDepth, attenuate, stats)
		if wavEnc != nil {wavEnc.Write(mixFloatBuffer.AsIntBuffer())}
		if mixBufferSize < bufferCap {break}
	}

	return nil

}

// Function to mix wav files and write mix to output
func MixWavFiles(files []*string, outWavName *string, fx []*TrackFX, bitDepth int, attenuate bool, stats []*TrackStats, bufferCap int) (error) {

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
	if outWavName != nil {
		if *outWavName == "-" {wavOut = os.Stdout
		} else {
			wavOut, err = os.Create(*outWavName)
			if err != nil {return err}
			defer wavOut.Close()
		}
	}

	// Mix decoders
	err = MixWavDecoders(wavDecs, wavOut, fx, bitDepth, attenuate, stats, bufferCap)
	if err != nil {return err}

	return nil
}

// Function to create a new TrackInfo
func newTrack(format *audio.Format, bitDepth, bufferCap int) (newTrack *TrackInfo) {
	return &TrackInfo{
		BitDepth: bitDepth,
		BufferSize: -1,
		IntBuffer: &audio.IntBuffer{Format: format, Data: make([]int, bufferCap)},
		FloatBuffer: &audio.FloatBuffer{Format: format, Data: make([]float64, bufferCap)},
	}
}

// Function to update stats
func updateTrackStats(stats []*TrackStats, bitDepth int, sourceTracks []*TrackInfo, mixTrack *audio.FloatBuffer, mixBufferSize int) {
	var max float64
	var min float64
	if bitDepth == 16 {
		max = 32767
		min = -32768
	} else if bitDepth == 24 {
		max = 8388607
		min = -8388608
	} else if bitDepth == 32 {
		max = 2147483647
		min = -2147483648
	}
	for t, track := range sourceTracks {
		if track.BufferSize != 0 {
			for i := 0; i < track.BufferSize; i++ {
				sample := track.FloatBuffer.Data[i]
				sampleAbs := math.Abs(sample)
				if sampleAbs > (*stats[t]).Peak {(*stats[t]).Peak = sampleAbs}
				if sample > max || sample < min {(*stats[t]).ClippedCount++}
				(*stats[t]).SumOfSquares += sample*sample
			}
			(*stats[t]).PeakdB = 20*math.Log10(stats[t].Peak/max)
			(*stats[t]).SampleCount += uint64(track.BufferSize)
			(*stats[t]).RMSdB = 20*math.Log10(math.Sqrt((*stats[t]).SumOfSquares/float64((*stats[t]).SampleCount))/max)
		}
	}
	mix := len(stats)-1
	for i := 0; i < mixBufferSize; i++ {
		sample := mixTrack.Data[i]
		sampleAbs := math.Abs(sample)
		if sampleAbs > (*stats[mix]).Peak {(*stats[mix]).Peak = sampleAbs}
		if sample > max || sample < min {(*stats[mix]).ClippedCount++}
		(*stats[mix]).SumOfSquares += sample*sample
	}
	(*stats[mix]).PeakdB = 20*math.Log10(stats[mix].Peak/max)
	(*stats[mix]).SampleCount += uint64(mixBufferSize)
	(*stats[mix]).RMSdB = 20*math.Log10(math.Sqrt((*stats[mix]).SumOfSquares/float64((*stats[mix]).SampleCount))/max)
}