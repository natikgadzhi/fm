module github.com/natikgadzhi/fm

go 1.26.1

replace github.com/natikgadzhi/cli-kit => ../template/cli-kit

require (
	git.sr.ht/~rockorager/go-jmap v0.5.3
	github.com/natikgadzhi/cli-kit v0.0.0-00010101000000-000000000000
	github.com/spf13/cobra v1.10.2
	github.com/zalando/go-keyring v0.2.6
	golang.org/x/term v0.41.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	al.essio.dev/pkg/shellescape v1.5.1 // indirect
	github.com/danieljoos/wincred v1.2.2 // indirect
	github.com/godbus/dbus/v5 v5.1.0 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.9 // indirect
	golang.org/x/oauth2 v0.36.0 // indirect
	golang.org/x/sys v0.42.0 // indirect
)
