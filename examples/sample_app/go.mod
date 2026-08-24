module sample_app

go 1.25.12

require github.com/vickychhetri/nova v0.1.0

require (
	golang.org/x/image v0.39.0 // indirect
	golang.org/x/text v0.40.0 // indirect
)

replace github.com/vickychhetri/nova => ../
