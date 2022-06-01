package merger

import (
	"encoding/json"
	"sort"
	"strings"

	log "github.com/sirupsen/logrus"

	"github.com/vorlif/spreak/pkg/cldrplural"

	"github.com/vorlif/xspreak/encoder"
)

func MergeJSON(src []byte, dst []byte, cats []cldrplural.Category) []byte {
	if len(src) == 0 {
		log.Fatal("Source file is empty")
	}

	var sourceFile encoder.JSONFile
	var targetFile encoder.JSONFile
	if err := json.Unmarshal(src, &sourceFile); err != nil {
		log.WithError(err).Fatal("Source file could not be decoded")
	}

	if len(dst) == 0 {
		targetFile = make(encoder.JSONFile, len(sourceFile))
	} else {
		if err := json.Unmarshal(dst, &targetFile); err != nil {
			log.WithError(err).Fatal("Target file could not be decoded")
		}
	}

	newItems := make(map[string]encoder.JSONItem)

	isTargetCat := func(key string) bool {
		for _, cat := range cats {
			if catKey(cat) == key {
				return true
			}
		}
		return false
	}

	for _, srcItem := range sourceFile {
		msg := make(encoder.JSONMessage)

		if ctx, hasCtx := srcItem.Message["context"]; hasCtx {
			msg["context"] = ctx
		}

		keyCount := 0
		for k := range srcItem.Message {
			if isCategory(k) {
				keyCount++
			}
			if isTargetCat(k) {
				msg[k] = ""
			}
		}

		if keyCount > 1 {
			for _, cat := range cats {
				msg[catKey(cat)] = ""
			}
		}

		for k, v := range srcItem.Message {
			if _, ok := msg[k]; ok {
				msg[k] = v
			}
		}

		newItems[srcItem.Key] = encoder.JSONItem{
			Key:     srcItem.Key,
			Message: msg,
		}
	}

	for _, oldItem := range targetFile {
		newItem, ok := newItems[oldItem.Key]
		if !ok {
			continue
		}

		for k, v := range oldItem.Message {
			if _, ok = newItem.Message[k]; ok && v != "" {
				newItem.Message[k] = v
			}
		}
	}

	file := make(encoder.JSONFile, 0, len(newItems))
	for _, v := range newItems {
		file = append(file, v)
	}
	sort.Slice(file, func(i, j int) bool {
		return file[i].Key < file[j].Key
	})

	data, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		log.WithError(err).Fatal("Marshal failed")
	}
	return data
}

func catKey(cat cldrplural.Category) string {
	return strings.ToLower(cat.String())
}

func isCategory(key string) bool {
	for cat := range cldrplural.CategoryNames {
		if catKey(cat) == key {
			return true
		}
	}
	return false
}
