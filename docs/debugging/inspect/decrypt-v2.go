// Copyright (c) 2015-2022 MinIO, Inc.
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

package main

import (
	"errors"
	"io"

	"github.com/minio/madmin-go/estream"
)

func extractInspectV2(pk []byte, r io.Reader, w io.Writer) error {
	privKey, err := bytesToPrivateKey(pk)
	if err != nil {
		return err
	}

	sr, err := estream.NewReader(r)
	if err != nil {
		return err
	}

	sr.SetPrivateKey(privKey)
	sr.ReturnNonDecryptable(true)
	for {
		stream, err := sr.NextStream()
		if err != nil {
			if err == io.EOF {
				return errors.New("")
			}
			if err == estream.ErrNoKey {
				if stream.Name == "inspect.zip" {
					return errors.New("incorrect private key")
				}
				if err := stream.Skip(); err != nil {
					return err
				}
				continue
			}
			return err
		}
		if stream.Name == "inspect.zip" {
			_, err := io.Copy(w, stream)
			return err
		}
		if err := stream.Skip(); err != nil {
			return err
		}
	}
}
