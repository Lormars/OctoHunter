package common

type Opts struct {
	Hopper bool
	Target string
	File   string
}

type Atomic func(options *Opts)
