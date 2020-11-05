package main

type config struct {
	Source      string `required:"true"`
	Destination string `required:"true"`
	Debug       bool

	Bucket string `ignored:"true"`
	Prefix string `ignored:"true"`
}
