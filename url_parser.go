package main

import (
	"fmt"
	"regexp"
)

type URLParser struct {
	next handler
}

func (u *URLParser) execute(m *GenericMessage) {
	re, _ = regexp.Compile("foo")	
	// TODO - mitenköhän regexit toimikaan
	fmt.Println(regexp.MatchString(
		"(?i)\\b((?:https?://|www\\d{0,3}[.]|[a-z0-9.\\-]+[.][a-z]{2,4}/)(?:[^\s()<>]+|\(([^\s()<>]+|(\([^\s()<>]+\)))*\))+(?:\(([^\s()<>]+|(\([^\s()<>]+\)))*\)|[^\s`!()\[\]{};:'".,<>?«»“”‘’]))",
		"foobar123"))
    // let url_regex = Regex::new(r#"(?i)\b((?:https?://|www\d{0,3}[.]|[a-z0-9.\-]+[.][a-z]{2,4}/)(?:[^\s()<>]+|\(([^\s()<>]+|(\([^\s()<>]+\)))*\))+(?:\(([^\s()<>]+|(\([^\s()<>]+\)))*\)|[^\s`!()\[\]{};:'".,<>?«»“”‘’]))"#).unwrap();
    // url_regex
    //     .captures_iter(input)
    //     .map(|capture| capture[1].to_string())
    //     .collect()
}

func (u *URLParser) setNext(next handler) {
	u.next = next
}