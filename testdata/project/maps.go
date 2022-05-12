package main

import (
	alias "github.com/vorlif/spreak/localize"

	"github.com/vorlif/testdata/sub"
)

var globalKeyMap = map[alias.Singular]string{
	"globalKeyMap-a": "a",
	"globalKeyMap-b": "b",
}

var globalValueMap = map[string]alias.Singular{
	"a": "globalValueMap-a",
	"b": "globalValueMap-b",
}

var globalMap = map[alias.MsgID]alias.Singular{
	"globalMap-ka": "globalMap-va",
	"globalMap-kb": "globalMap-vb",
}

func mapFunc() {
	_ = map[alias.Singular]string{
		"localKeyMap-a": "a",
		"localKeyMap-b": "b",
	}

	_ = map[string]alias.Singular{
		"a": "localValueMap-a",
		"b": "localValueMap-b",
	}

	_ = map[alias.MsgID]alias.Singular{
		"localMap-ka": "localMap-va",
		"localMap-kb": "localMap-vb",
	}

	_ = map[string]sub.Sub{
		"key": {
			Text:   "map struct msgid",
			Plural: "map struct plural",
		},
	}

	_ = map[string]*sub.Sub{
		"key": {
			Text:   "map pointer struct msgid",
			Plural: "map pointer struct plural",
		},
	}
}
