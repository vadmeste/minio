// Copyright (c) 2015-2023 MinIO, Inc.
//
// This file is part of MinIO Object Storage stack
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package cmd

import (
	"time"

	"github.com/pkg/xattr"
)

//go:generate msgp -file=$GOFILE

type FileAttrs struct {
	ETag      string    `msgp:"et"`
	ModTime   time.Time `msgp"mt"`
	Parts     int       `msgp:"p"`
	Encrypted bool      `msgp:"enc"`
}

func getXAttrs(path string) (FileAttrs, error) {
	b, err := xattr.LGet(path, "user.m")
	if err != nil {
		return FileAttrs{}, err
	}

	var attrs FileAttrs
	_, err = attrs.UnmarshalMsg(b)
	if err != nil {
		return FileAttrs{}, err
	}
	return attrs, nil
}

func setXAttrs(path string, attrs FileAttrs) error {
	b, err := attrs.MarshalMsg(nil)
	if err != nil {
		return err
	}
	return xattr.LSet(path, "user.m", b)
}

func delXAttrs(path string) error {
	return xattr.LRemove(path, "user.m")
}
