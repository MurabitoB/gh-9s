// Package demo is only here to give gh-9s something to review.
package demo

import (
	"fmt"
	"strings"
)

// Greeter says hello.
type Greeter struct {
	Name  string
	Loud  bool
	Count int
}

// Greet returns a greeting for who.
func (g *Greeter) Greet(who string) string {
	g.Count++
	msg := fmt.Sprintf("hello, %s", who)
	if g.Loud {
		msg = strings.ToUpper(msg)
	}
	return msg
}

// Reset forgets how many greetings were made.
func (g *Greeter) Reset() {
	g.Count = 0
}

// Retry runs fn up to three times.
func Retry(fn func() error) error {
	var err error
	for i := 0; i < 3; i++ {
		if err = fn(); err == nil {
			return nil
		}
	}
	return err
}

// Join joins words with a space.
func Join(words ...string) string {
	out := ""
	for i, w := range words {
		if i > 0 {
			out += " "
		}
		out += w
	}
	return out
}

// Shout makes s loud.
func Shout(s string) string {
	return strings.ToUpper(s) + "!"
}
