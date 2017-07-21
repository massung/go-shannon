# Shannon-Fano Text Encoding

Huffman is a Go package that implements [Shannon-Fano coding][1] for text strings in [Go][2]. It's coded with the following goals in mind:

* Simple to create Shannon-Fano tables using multiple methods.
* Table can be written to an `io.Writer` and read from an `io.Reader`.
* Encode *and* decode strings to/from `[]uint32` bit vectors.

## Install

Simply use `go get` to download the package into your `$GOPATH`:

```
go get github.com/massung/go-shannon
```

## Documentation

Documentation can be found on [GoDoc](https://godoc.org/github.com/massung/go-shannon).

## Quickstart

Here's a quick example that shows common use and how to serialize the table (to memory, disk, network, etc.) as well:

```go
package main

import (
    "bytes"
    "encoding/gob"

    "github.com/massung/go-shannon"
)

var (
    loremIpsum = "Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip ex ea commodo consequat. Duis aute irure dolor in reprehenderit in voluptate velit esse cillum dolore eu fugiat nulla pariatur. Excepteur sint occaecat cupidatat non proident, sunt in culpa qui officia deserunt mollit anim id est laborum."
)

func main() {
    var network bytes.Buffer
    var encodeTable shannon.Table
    var decodeTable shannon.Table

    // construct a new table of shannon codes from a source string
    encodeTable = shannon.BuildTableFromString(loremIpsum)
    
    // encode the source text using the table
    bitVec, size, err := encodeTable.Encode(loremIpsum)
    if err != nil {
        panic(err)
    }

    // save the table to memory (could be disk, network, etc.)
    if err := gob.NewEncoder(&network).Encode(&encodeTable); err != nil {
        panic(err)
    }

    // load the table from memory
    r := bytes.NewReader(network.Bytes())
    if err := gob.NewDecoder(r).Decode(&decodeTable); err != nil {
        panic(err)
    }
    
    // decode the bit vector back into a string
    s, err := decodeTable.Decode(bitVec, size)
    if err != nil {
        panic(err)
    }
    
    // should print: true 445 244
    println(s == loremIpsum, len(loremIpsum), len(bitVec)*4)
}
```

## Building a Table

There are three methods provided for building a `shannon.Table`:

#### BuildTable(probabilityMap map[rune]float64) shannon.Table

Provide a probability map of `rune` to `float64`. There's no requirement for the probabilities to add up to 1.0, but none of the values should be negative. The other two function below both build a probability map and call this method.

#### BuildTableFromString(sourceString string) shannon.Table

Give a source string. Typically, this string would be what you intend to encode, and is probably a one-time use: the same table will not be used to encode many things.

#### BuildTableFromOrderedRunes(runes []rune) shannon.Table

Pass in a list of runes in the order of probability. The first rune in the slice occurs most frequently, and the last rune the least. The probability is of each rune is evenly distributed so that combined they add up to 1.0. For example, if the `[]rune{'a','b','c'}` is passed in, the probability map created will be:

```go
map[rune]float64{
    'a': 0.500, // 3/6
    'b': 0.333, // 2/6
    'c': 0.166, // 1/6
}
```

## Encoding and Decoding

Once you have a `shannon.Table`, you can use it to `Encode` a string and `Decode` a bit vector (`[]uint32`). Keep in mind that the *same table must be used to decode a bit vector that was used to encode!*

The encoded bit vectors are ordered MSB first. That means if `Encode` returns a `[]uint32` of length 1, and a size of 17, then the most-significant 17 bits of the first `uint32` are what contain valid bits to be decoded. Typically, this isn't that important as you just serialize the `[]uint32` returned. But, if you wanted to do anything with the bits, it would matter.

*Note: do not forget to serialize the size returned from encoding as well as the bit vector, as both are needed to decode!*

## Performance

Encoding is extremely fast as it's a simple matter of looking up each rune and appending the matching code bits to the bit vector.

I haven't done any work to optimizing decoding. 

I wanted to keep the code extremely small and simple, and unless there are hundreds of runes in the table, the list is small enough that a linear search should be quite quick. The lorem ipsum text above has only 28 unique runes, and if the most typical use case is ASCII characters, then the limit is ~94 (32-126).

## That's all folks!

If you find it useful, have a comment, or find a bug, let me know or open an issue!

[1]: https://en.wikipedia.org/wiki/Shannon–Fano_coding
[2]: https://golang.org