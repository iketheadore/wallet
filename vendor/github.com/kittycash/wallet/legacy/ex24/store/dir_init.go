package store

import (
	"encoding/json"
	"errors"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"log"
	"os"
	"path"
	"strconv"
)

// InitDir initializes a directory.
func InitDir(rootDir string, count int) error {
	if e := os.RemoveAll(rootDir); e != nil {
		return e
	}
	if e := os.MkdirAll(rootDir, os.FileMode(0755)); e != nil {
		return e
	}
	var rect = image.NewRGBA(image.Rect(0, 0, 1200, 1200))
	for i := 0; i < count; i++ {

		var (
			iStr     = strconv.Itoa(i)
			jsonName = path.Join(rootDir, iStr+".json")
			pngName  = path.Join(rootDir, iStr+".png")
			raw, _   = json.MarshalIndent(
				KittyMeta{ID: uint64(i), Attributes: []string{}},
				"", "    ")
		)
		{
			f, e := os.Create(jsonName)
			if e != nil {
				return e
			}
			if _, e := f.Write(raw); e != nil {
				return e
			}
			f.Close()
		}
		{
			f, e := os.Create(pngName)
			if e != nil {
				return e
			}
			if e := png.Encode(f, rect); e != nil {
				return e
			}
			f.Close()
		}
	}
	return nil
}

// InitMetaStorerFromDir prepares a DB, loading a directory in.
func InitMetaStorerFromDir(rootDir string, kittyCount int, db MetaStorer) error {
	for i := 0; i < kittyCount; i++ {
		var (
			iStr     = strconv.Itoa(i)
			jsonName = path.Join(rootDir, iStr+".json")
			pngName  = path.Join(rootDir, iStr+".png")
			jsonObj  KittyMeta
			pngRaw   []byte
		)
		{
			f, e := os.Open(pngName)
			if e != nil {
				return e
			}
			pngRaw, e = ioutil.ReadAll(f)
			if e != nil {
				return e
			}
			f.Close()
		}
		{
			f, e := os.Open(jsonName)
			if e != nil {
				return e
			}
			if e := json.NewDecoder(f).Decode(&jsonObj); e != nil {
				return e
			}
			f.Close()
		}

		if kID, ok := db.Append(jsonObj, pngRaw); !ok {
			return errors.New(fmt.Sprintf(""))
		} else {
			log.Printf(
				"[%d] kitty added : id(%d)", i, kID)
		}
	}
	return nil
}
