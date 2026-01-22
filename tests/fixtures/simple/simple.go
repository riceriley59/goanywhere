// Copyright 2026 Riley Rice
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package simple

// Point represents a 2D point
type Point struct {
	X int
	Y int
}

// Add adds two integers
func Add(a, b int) int {
	return a + b
}

// Greet returns a greeting
func Greet(name string) string {
	return "Hello, " + name
}

// Divide divides two numbers, returns error if divisor is zero
func Divide(a, b float64) (float64, error) {
	if b == 0 {
		return 0, nil
	}
	return a / b, nil
}

// Sum returns sum of variadic ints
func Sum(nums ...int) int {
	total := 0
	for _, n := range nums {
		total += n
	}
	return total
}

// NewPoint creates a new Point
func NewPoint(x, y int) *Point {
	return &Point{X: x, Y: y}
}

// Distance calculates distance from origin
func (p *Point) Distance() float64 {
	return 0
}

// Scale multiplies point by factor
func (p *Point) Scale(factor int) {
	p.X *= factor
	p.Y *= factor
}

// unexported should be skipped
func unexported() {} //nolint:unused // test fixture for parser

// unexportedStruct should be skipped
type unexportedStruct struct{} //nolint:unused // test fixture for parser
