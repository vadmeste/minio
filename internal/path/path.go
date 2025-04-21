// Copyright (c) 2015-2025 MinIO, Inc.
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

package path

import (
	"path"

	"github.com/valyala/bytebufferpool"
)

// HasSuffixByte returns true if the last byte of s is 'suffix'
func HasSuffixByte(s string, suffix byte) bool {
	return len(s) > 0 && s[len(s)-1] == suffix
}

// SlashSeparator - slash separator.
const SlashSeparator = "/"

// SlashSeparatorChar - slash separator.
const SlashSeparatorChar = '/'

// Join - like stdlib path.Join() but retains trailing SlashSeparator of the
// last element
func Join(elem ...string) string {
	sb := bytebufferpool.Get()
	defer func() {
		sb.Reset()
		bytebufferpool.Put(sb)
	}()

	return JoinBuf(sb, elem...)
}

// JoinBuf - like stdlib path.Join() but retains trailing SlashSeparator of
// the last element. Provide a string builder to reduce allocation.
func JoinBuf(dst *bytebufferpool.ByteBuffer, elem ...string) string {
	trailingSlash := len(elem) > 0 && HasSuffixByte(elem[len(elem)-1], SlashSeparatorChar)
	dst.Reset()
	added := 0
	for _, e := range elem {
		if added > 0 || e != "" {
			if added > 0 {
				dst.WriteByte(SlashSeparatorChar)
			}
			dst.WriteString(e)
			added += len(e)
		}
	}

	if pathNeedsClean(dst.Bytes()) {
		s := path.Clean(dst.String())
		if trailingSlash {
			return s + SlashSeparator
		}
		return s
	}
	if trailingSlash {
		dst.WriteByte(SlashSeparatorChar)
	}
	return dst.String()
}

// pathNeedsClean returns whether path.Clean may change the path.
// Will detect all cases that will be cleaned,
// but may produce false positives on non-trivial paths.
func pathNeedsClean(path []byte) bool {
	if len(path) == 0 {
		return true
	}

	rooted := path[0] == '/'
	n := len(path)

	r, w := 0, 0
	if rooted {
		r, w = 1, 1
	}

	for r < n {
		switch {
		case path[r] > 127:
			// Non ascii.
			return true
		case path[r] == '/':
			// multiple / elements
			return true
		case path[r] == '.' && (r+1 == n || path[r+1] == '/'):
			// . element - assume it has to be cleaned.
			return true
		case path[r] == '.' && path[r+1] == '.' && (r+2 == n || path[r+2] == '/'):
			// .. element: remove to last / - assume it has to be cleaned.
			return true
		default:
			// real path element.
			// add slash if needed
			if rooted && w != 1 || !rooted && w != 0 {
				w++
			}
			// copy element
			for ; r < n && path[r] != '/'; r++ {
				w++
			}
			// allow one slash, not at end
			if r < n-1 && path[r] == '/' {
				r++
			}
		}
	}

	// Turn empty string into "."
	if w == 0 {
		return true
	}

	return false
}
