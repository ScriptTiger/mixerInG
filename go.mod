module github.com/ScriptTiger/mixerInG

go 1.23.0

replace github.com/go-audio/wav => github.com/ScriptTiger/wav v0.0.0-20241009130152-f2b055a7031c

require (
	github.com/go-audio/audio v1.0.0
	github.com/go-audio/wav v0.0.0-20181013172942-de841e69b884
)

require github.com/go-audio/riff v1.0.0 // indirect
