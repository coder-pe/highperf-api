/*
 * Copyright (C) 2025 Miguel Mamani <miguel.coder.per@gmail.com>
 *
 * This program is free software: you can redistribute it and/or modify
 * it under the terms of the GNU Affero General Public License as published
 * by the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
 * GNU Affero General Public License for more details.
 *
 * You should have received a copy of the GNU Affero General Public License
 * along with this program. If not, see <https://www.gnu.org/licenses/>.
 */

// internal/encoding/jsonx/pool.go
package jsonx

import (
	"bytes"
	"sync"
)

var bufPool = sync.Pool{
	New: func() any { return new(bytes.Buffer) },
}

func GetBuffer() *bytes.Buffer {
	b := bufPool.Get().(*bytes.Buffer)
	b.Reset()
	return b
}

func PutBuffer(b *bytes.Buffer) {
	if b.Cap() > 1<<20 { // evita pools gigantes (1MB)
		return
	}
	bufPool.Put(b)
}
