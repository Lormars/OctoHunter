package common

type Opts struct {
	Hopper bool
	Dork   bool
	Target string
	File   string
}

type Atomic func(options *Opts)
