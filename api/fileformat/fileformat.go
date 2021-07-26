package fileformat

import (
	"path"
	"strings"

	"github.com/twinj/uuid"
)

func UniqueFormat(fn string) string {
	filename := strings.TrimSuffix(fn, path.Ext(fn))
	extension := path.Ext(fn)
	u := uuid.NewV4()
	newFilename := filename + "-" + u.String() + extension

	return newFilename
}
