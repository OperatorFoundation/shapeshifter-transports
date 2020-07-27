/*
	MIT License

	Copyright (c) 2020 Operator Foundation

	Permission is hereby granted, free of charge, to any person obtaining a copy
	of this software and associated documentation files (the "Software"), to deal
	in the Software without restriction, including without limitation the rights
	to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
	copies of the Software, and to permit persons to whom the Software is
	furnished to do so, subject to the following conditions:

	The above copyright notice and this permission notice shall be included in all
	copies or substantial portions of the Software.

	THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
	IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
	FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
	AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
	LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
	OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
	SOFTWARE.
*/

package toneburst

import (
	"net"
)

type ToneBurst interface {
	Perform(conn net.Conn) error
}

// func New(config *Config) ToneBurst {
// 	if config == nil {
// 		return nil
// 	}
// 	switch config.Selector {
// 	case "whalesong":
// 		if config.Whalesong == nil {
// 			return nil
// 		} else {
// 			return NewWhalesong(*config.Whalesong)
// 		}
// 	case "monolith":
// 		if config.Monotone == nil {
// 			return nil
// 		} else {
// 			return NewMonotone(*config.Monotone)
// 		}
// 	default:
// 		return nil
// 	}
// }
