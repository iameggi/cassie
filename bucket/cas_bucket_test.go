package bucket

import (
	"bytes"
	"encoding/json"
	"io"
	"strconv"
	"testing"
	"time"
)

// Variant represents a product variant used in benchmark data.
type Variant struct {
	SKU   string `json:"sku"`
	Size  string `json:"size"`
	Color string `json:"color"`
	Stock int    `json:"stock"`
}

// Product represents a sample product structure for serialization benchmarks.
type Product struct {
	ID          string    `json:"id"`
	SKU         string    `json:"sku"`
	Name        string    `json:"name"`
	Price       float64   `json:"price"`
	Stock       int       `json:"stock"`
	Variants    []Variant `json:"variants"`
	Description string    `json:"description"`
}

var largeProductList []Product

// init prepares the test data with 10,000 products.
func init() {
	largeProductList = make([]Product, 10_000)
	for i := 0; i < len(largeProductList); i++ {
		largeProductList[i] = Product{
			ID:    "prod-" + strconv.Itoa(i),
			SKU:   "SKU-" + strconv.Itoa(i),
			Name:  "Product " + strconv.Itoa(i),
			Price: float64(i) * 1.23,
			Stock: i % 100,
			Variants: []Variant{
				{SKU: "V1-" + strconv.Itoa(i), Size: "M", Color: "Black", Stock: i % 10},
				{SKU: "V2-" + strconv.Itoa(i), Size: "L", Color: "Red", Stock: (i + 1) % 10},
			},
			Description: "Description for product " + strconv.Itoa(i),
		}
	}
}

// benchmarkWithLatency is a helper that records latency and allocations.
func benchmarkWithLatency(b *testing.B, fn func() error) {
	b.ReportAllocs()
	var totalTime time.Duration

	for i := 0; i < b.N; i++ {
		start := time.Now()
		if err := fn(); err != nil {
			b.Fatal(err)
		}
		totalTime += time.Since(start)
	}

	avgLatency := totalTime / time.Duration(b.N)
	b.Logf("Avg latency per op: %s", avgLatency)
}

// BenchmarkJSONEncoding_Standard benchmarks JSON encoding using standard allocation.
func BenchmarkJSONEncoding_Standard(b *testing.B) {
	benchmarkWithLatency(b, func() error {
		data, err := json.Marshal(largeProductList)
		if err != nil {
			return err
		}
		_, _ = io.Discard.Write(data)
		return nil
	})
}

// BenchmarkJSONEncoding_Cassie benchmarks JSON encoding using Cassie’s byte buffer pool.
func BenchmarkJSONEncoding_Cassie(b *testing.B) {
	benchmarkWithLatency(b, func() error {
		buf := ByteBucket.Get()
		buf.Reset()

		if err := json.NewEncoder(buf).Encode(largeProductList); err != nil {
			return err
		}
		_, _ = io.Discard.Write(buf.Bytes())
		ByteBucket.Put(buf)
		return nil
	})
}

// BenchmarkJSONEncoding_CassieWith benchmarks Cassie’s callback-style pooling helper.
func BenchmarkJSONEncoding_CassieWith(b *testing.B) {
	benchmarkWithLatency(b, func() error {
		return WithByteBufferErr(func(buf *bytes.Buffer) error {
			if err := json.NewEncoder(buf).Encode(largeProductList); err != nil {
				return err
			}
			_, err := io.Discard.Write(buf.Bytes())
			return err
		})
	})
}
