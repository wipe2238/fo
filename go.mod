module github.com/wipe2238/fo

go 1.22.4

require (
	github.com/Jleagle/steam-go v0.0.0-20231027203227-3dc26c48c3d2
	github.com/shoenig/test v1.8.2
	github.com/spf13/cobra v1.8.1
	golang.org/x/sys v0.22.0
)

require (
	github.com/google/go-cmp v0.6.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.5 // indirect
)

replace github.com/inconshreveable/mousetrap => ./x/replace/mousetrap
