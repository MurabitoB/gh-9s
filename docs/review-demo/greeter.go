// Package demo is only here to give gh-9s something to review.
package demo

import (
	"fmt"
	"strings"
	"time"
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
	msg := fmt.Sprintf("hello, %s (#%d)", who, g.Count)
	if g.Loud {
		return strings.ToUpper(msg) + "!"
	}
	return msg
}

// Retry runs fn up to n times, waiting a little longer after each failure.
func Retry(n int, fn func() error) error {
	var err error
	for i := 0; i < n; i++ {
		if err = fn(); err == nil {
			return nil
		}
		time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
	}
	return err
}

// Join joins words with a space.
func Join(words ...string) string {
	return strings.Join(words, " ")
}

// Shout makes s loud.
func Shout(s string) string {
	return strings.ToUpper(s) + "!"
}

// Whisper makes s quiet.
func Whisper(s string) string {
	return strings.ToLower(s) + "..."
}
