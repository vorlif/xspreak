package encoder

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"sort"
	"strings"
	"time"

	"github.com/vorlif/spreak/catalog/cldrplural"

	"github.com/vorlif/xspreak/extract"
	"github.com/vorlif/xspreak/extract/etype"
	"github.com/vorlif/xspreak/util"
)

type jsonEncoder struct {
	w *json.Encoder
}

func NewJSONEncoder(w io.Writer, ident string) Encoder {
	enc := json.NewEncoder(w)
	enc.SetIndent("", ident)

	return &jsonEncoder{w: enc}
}

func (e *jsonEncoder) Encode(issues []extract.Issue) error {
	util.TrackTime(time.Now(), "Build messages")

	items := make(map[string]JSONItem, len(issues))

	for _, iss := range issues {
		msg := make(JSONMessage)
		msg[catKey(cldrplural.Other)] = ""

		key := iss.MsgID

		if iss.Context != "" {
			msg["context"] = iss.Context
			key = fmt.Sprintf("%s_%s", iss.MsgID, iss.Context)
		}

		if iss.PluralID != "" {
			msg[catKey(cldrplural.One)] = ""

			if iss.IDToken != etype.PluralKey && iss.IDToken != etype.Key {
				msg[catKey(cldrplural.One)] = iss.MsgID
				msg[catKey(cldrplural.Other)] = iss.PluralID
			}
		} else {
			if iss.IDToken != etype.PluralKey && iss.IDToken != etype.Key {
				msg[catKey(cldrplural.Other)] = iss.MsgID
			}
		}

		items[key] = JSONItem{Key: key, Message: msg}
	}

	file := make(JSONFile, 0, len(issues))
	for _, item := range items {
		file = append(file, item)
	}

	sort.Slice(file, func(i, j int) bool {
		return file[i].Key < file[j].Key
	})

	return e.w.Encode(file)
}

type JSONItem struct {
	Key     string
	Message JSONMessage
}

type JSONFile []JSONItem

func (f JSONFile) MarshalJSON() ([]byte, error) {
	var buf bytes.Buffer
	buf.WriteString("{")

	for i, kv := range f {
		if i != 0 {
			buf.WriteString(",")
		}
		// marshal key
		key, err := json.Marshal(kv.Key)
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteString(":")
		// marshal value
		val, err := json.Marshal(&kv.Message)
		if err != nil {
			return nil, err
		}
		buf.Write(val)
	}

	buf.WriteString("}")
	return buf.Bytes(), nil
}

func (f *JSONFile) UnmarshalJSON(data []byte) error {
	messages := make(map[string]JSONMessage)
	if err := json.Unmarshal(data, &messages); err != nil {
		return err
	}

	for k, msg := range messages {
		*f = append(*f, JSONItem{Key: k, Message: msg})
	}

	return nil
}

type JSONMessage map[string]string

var messageKey = []string{"context", "zero", "one", "two", "few", "many", "other"}

func (m JSONMessage) MarshalJSON() ([]byte, error) {
	switch len(m) {
	case 0:
		return json.Marshal("")
	case 1:
		for _, v := range m {
			return json.Marshal(v)
		}
	}

	var buf bytes.Buffer
	buf.WriteString("{")

	keywordNum := 0
	for _, keyword := range messageKey {
		text, ok := m[keyword]
		if !ok {
			continue
		}

		if keywordNum != 0 {
			buf.WriteString(",")
		}
		keywordNum++

		// marshal key
		key, err := json.Marshal(keyword)
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteString(":")
		// marshal value
		val, err := json.Marshal(text)
		if err != nil {
			return nil, err
		}
		buf.Write(val)
	}

	buf.WriteString("}")
	return buf.Bytes(), nil
}

func (m *JSONMessage) UnmarshalJSON(data []byte) error {
	if (*m) == nil {
		*m = make(JSONMessage)
	}

	var other string
	if err := json.Unmarshal(data, &other); err == nil {
		(*m)[catKey(cldrplural.Other)] = other
		return nil
	}

	var mm map[string]string
	if err := json.Unmarshal(data, &mm); err != nil {
		return err
	}

	if len(mm) == 0 {
		(*m)[catKey(cldrplural.Other)] = ""
		return nil
	}

	for k, v := range mm {
		(*m)[k] = v
	}
	return nil
}

func catKey(cat cldrplural.Category) string {
	return strings.ToLower(cat.String())
}
