module example

go 1.22.2

replace github.com/mattiaswal/go-sysrepo => ../

require github.com/mattiaswal/go-sysrepo v0.0.0-20190703175107-33a398e21ee0

require github.com/mattiaswal/go-libyang v0.0.0-00010101000000-000000000000 // indirect

replace github.com/mattiaswal/go-libyang => ../../go-libyang
