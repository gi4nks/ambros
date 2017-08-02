package utils

import (
	"crypto/rand"
	"encoding/json"

	"github.com/gi4nks/quant"
)

type Utilities struct {
	parrot *quant.Parrot
}

func NewUtilities(p quant.Parrot) *Utilities {
	return &Utilities{parrot: &p}
}

func (u *Utilities) AsJson(o interface{}) string {
	b, err := json.Marshal(o)
	if err != nil {
		u.parrot.Error("Warning", err)
		return "{}"
	}
	return string(b)
}

func (u *Utilities) Random() string {

	var dictionary = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"

	var bytes = make([]byte, 12)
	rand.Read(bytes)
	for k, v := range bytes {
		bytes[k] = dictionary[v%byte(len(dictionary))]
	}
	return string(bytes)
}

func (u *Utilities) Tail(a []string) []string {
	if len(a) >= 2 {
		return []string(a)[1:]
	}
	return []string{}
}

func (u *Utilities) Check(e error) {
	if e != nil {
		u.parrot.Error("Error...", e)
		return
	}
}

func (u *Utilities) Fatal(e error) {
	if e != nil {
		u.parrot.Error("Fatal...", e)
		panic(e)
	}
}
